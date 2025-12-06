# Memory Pool Optimization Recommendations for jxl-go

## Executive Summary
This document outlines comprehensive memory pool optimizations to reduce GC pressure and improve performance in the JXL decoder. The optimizations focus on frequently allocated large buffers in hot code paths.

## Current Status
- ✅ PassGroup.invertVarDCT: scratchBlock pooled (5x256x256)
- ⚠️ Frame.copyFloatBuffers: Non-pooled allocations in hot path
- ⚠️ HFCoefficients: Large matrix allocations per group
- ⚠️ ModularChannel: Buffer allocations per channel
- ⚠️ LFCoefficients: Dequant coefficient matrices

## Priority 1: Critical Hot Path Optimizations

### 1. HFCoefficients Matrix Allocations
**Location:** `frame/HFCoefficients.go:65-74`
**Impact:** HIGH - Allocated per pass group (can be hundreds)
**Current Code:**
```go
nonZeros := util.MakeMatrix3D[int32](3, 32, 32)
hf.quantizedCoeffs = util.MakeMatrix3D[int32](3, 0, 0)
hf.dequantHFCoeff = util.MakeMatrix3D[float32](3, 0, 0)

for c := 0; c < 3; c++ {
    sY := size.Height >> header.jpegUpsamplingY[c]
    sX := size.Width >> header.jpegUpsamplingX[c]
    hf.quantizedCoeffs[c] = util.MakeMatrix2D[int32](sY, sX)
    hf.dequantHFCoeff[c] = util.MakeMatrix2D[float32](sY, sX)
}
```

**Optimization:**
- `nonZeros`: Pool for 3x32x32 int32 matrices (3KB each)
- `quantizedCoeffs` and `dequantHFCoeff`: Variable size, but typically similar dimensions
- Create pools for common sizes (256x256, 512x512) with fallback to heap

**Estimated Savings:** ~3-5MB per frame for typical images

### 2. Frame.copyFloatBuffers (EPF Filter)
**Location:** `frame/Frame.go:1043-1051`
**Impact:** HIGH - Called 3 times per EPF iteration, allocates full frame buffers
**Current Code:**
```go
func copyFloatBuffers(buffer []image.ImageBuffer, colours int32) [][][]float32 {
    data := util.MakeMatrix3D[float32](int(colours), int(buffer[0].Height), int(buffer[0].Width))
    for c := int32(0); c < colours; c++ {
        for y := int32(0); y < buffer[c].Height; y++ {
            copy(data[c][y], buffer[c].FloatBuffer[y])
        }
    }
    return data
}
```

**Optimization:**
```go
func copyFloatBuffers(buffer []image.ImageBuffer, colours int32) [][][]float32 {
    data := util.MakeMatrix3DPooled[float32](int(colours), int(buffer[0].Height), int(buffer[0].Width))
    for c := int32(0); c < colours; c++ {
        for y := int32(0); y < buffer[c].Height; y++ {
            copy(data[c][y], buffer[c].FloatBuffer[y])
        }
    }
    return data
}
```

**Critical:** Must call `util.ReturnMatrix3DToPool()` after use (currently missing at line 994-996)

**Estimated Savings:** ~20-50MB per frame (3 allocations × frame size)

### 3. Frame.decodePassGroupsConcurrent - Frame Buffers
**Location:** `frame/Frame.go:666`
**Impact:** MEDIUM - One allocation per frame
**Current Code:**
```go
buffers := util.MakeMatrix3D[float32](3, 0, 0)
```

**Optimization:**
```go
buffers := util.MakeMatrix3DPooled[float32](3, 0, 0)
defer util.ReturnMatrix3DToPool(buffers)
```

**Estimated Savings:** Frame-dependent, typically 5-20MB

### 4. LFCoefficients Dequantization Buffers
**Location:** `frame/lfcoefficients.go` (check line ~170-180)
**Impact:** MEDIUM - One per LF group
**Pattern to Look For:**
```go
dequantLFCoeff := util.MakeMatrix3D[float32](3, 0, 0)
```

**Optimization:** Pool these 3D float32 matrices

## Priority 2: Moderate Impact Optimizations

### 5. ModularChannel Buffer Allocations
**Location:** `frame/ModularChannel.go:63-73`
**Impact:** MEDIUM - Many allocations, but smaller size
**Current Code:**
```go
func (mc *ModularChannel) allocate() {
    if len(mc.buffer) != 0 {
        return
    }
    if mc.size.Height == 0 || mc.size.Width == 0 {
        mc.buffer = util.MakeMatrix2D[int32](0, 0)
    } else {
        mc.buffer = util.MakeMatrix2D[int32](int(mc.size.Height), int(mc.size.Width))
    }
}
```

**Optimization Strategy:**
- Create separate pools for common sizes
- Track allocation/deallocation with lifecycle management
- Consider adding a `deallocate()` method to return to pool

**Challenge:** Buffers are stored in struct fields and have variable lifetimes

### 6. Frame.performUpsampling Buffer
**Location:** `frame/Frame.go:1130`
**Impact:** LOW-MEDIUM - Per channel upsampling
**Current Code:**
```go
newBuffer := util.MakeMatrix2D[float32](len(buffer)*int(k), 0)
for y := 0; y < len(buffer); y++ {
    newBuffer[y*int(k)+ky] = make([]float32, len(buffer[y])*int(k))
```

**Optimization:** Pool 2D float32 matrices for upsampling

### 7. Scratch Blocks in DCT Operations
**Location:** `util/dct.go` or `util/InverseDCT2D` calls
**Impact:** MEDIUM - Called frequently in VarDCT path

**Optimization:** Pass pooled scratch buffers to DCT functions instead of allocating internally

## Priority 3: Infrastructure Improvements

### 8. Buffer Pool Implementation
**File:** `util/buffer_pool.go` (create)

**Recommended Design:**
```go
package util

import "sync"

type Matrix3DPool[T any] struct {
    pools map[string]*sync.Pool // key: "depth_height_width"
}

func (p *Matrix3DPool[T]) Get(depth, height, width int) [][][]T {
    key := fmt.Sprintf("%d_%d_%d", depth, height, width)
    pool, exists := p.pools[key]
    if !exists {
        pool = &sync.Pool{
            New: func() interface{} {
                return MakeMatrix3D[T](depth, height, width)
            },
        }
        p.pools[key] = pool
    }
    return pool.Get().([][][]T)
}

func (p *Matrix3DPool[T]) Put(matrix [][][]T) {
    if len(matrix) == 0 {
        return
    }
    key := fmt.Sprintf("%d_%d_%d", len(matrix), len(matrix[0]), len(matrix[0][0]))
    if pool, exists := p.pools[key]; exists {
        // Clear matrix before returning to pool
        for i := range matrix {
            for j := range matrix[i] {
                for k := range matrix[i][j] {
                    matrix[i][j][k] = *new(T)
                }
            }
        }
        pool.Put(matrix)
    }
}
```

**Features:**
- Size-specific pools using sync.Pool
- Automatic clearing before return
- Thread-safe for concurrent decoding

### 9. Pool Metrics and Monitoring
**Purpose:** Track pool effectiveness

```go
type PoolMetrics struct {
    Hits   int64
    Misses int64
    Size   int64
}

func (p *Matrix3DPool[T]) GetMetrics() PoolMetrics {
    // Return metrics for monitoring
}
```

### 10. Slice Pooling for Small Allocations
**Locations:** Various `make([]T, n)` calls

**Examples:**
- `frame/Frame.go:806-808`: normGab slices (per frame)
- `frame/Frame.go:950`: sumChannels (per pixel in hot loop)
- `frame/FrameHeader.go:106-107`: jpegUpsampling slices

**Optimization:** Use sync.Pool for frequently allocated slices

## Implementation Strategy

### Phase 1: Critical Path (Week 1)
1. ✅ Implement basic Matrix3DPool
2. ✅ Add MakeMatrix3DPooled/ReturnMatrix3DToPool helpers
3. Update copyFloatBuffers and add return calls
4. Update Frame.decodePassGroupsConcurrent

### Phase 2: Medium Impact (Week 2)
5. Pool HFCoefficients matrices
6. Pool LFCoefficients buffers
7. Add Matrix2DPool for 2D allocations

### Phase 3: Polish (Week 3)
8. Add metrics tracking
9. ModularChannel lifecycle management
10. Performance testing and tuning

## Performance Testing Checklist

- [ ] Benchmark memory allocation rate (allocations/second)
- [ ] Measure GC pause time reduction
- [ ] Profile with `pprof` heap/allocs
- [ ] Test with various image sizes (small/medium/large)
- [ ] Verify no memory leaks (pool growth)
- [ ] Compare decode time before/after

## Expected Results

**Memory Allocation Reduction:** 60-80%
**GC Pressure Reduction:** 50-70%
**Performance Improvement:** 10-25% faster decode
**Peak Memory Usage:** 30-50% reduction

## Risk Mitigation

1. **Memory Leaks:** Ensure all pooled buffers are returned
   - Add `defer` statements immediately after allocation
   - Lint rules to check for missing returns

2. **Concurrency Issues:** Use sync.Pool for thread-safety
   - Each goroutine gets independent buffer from pool
   - No shared mutable state

3. **Size Explosion:** Limit pool size per bucket
   - Max 10 buffers per size
   - LRU eviction for uncommon sizes

4. **Incorrect Reuse:** Clear buffers before returning
   - Zero out matrices in Return function
   - Add debug mode to verify clearing

## Quick Wins (< 1 hour each)

1. ✅ Frame.go:666 - Add pool to buffers allocation
2. Frame.go:1044 - Convert copyFloatBuffers to pooled
3. Frame.go:994-996 - Add ReturnMatrix3DToPool calls
4. PassGroup.go - Verify existing pool usage is correct

## Code Review Checklist

For each pooled allocation:
- [ ] Is there a matching `defer Return()` or explicit return?
- [ ] Is the buffer cleared before return?
- [ ] Is the size predictable or variable?
- [ ] Is this in a hot path (called frequently)?
- [ ] Are there any data races with concurrent access?

## References

- Go sync.Pool documentation: https://pkg.go.dev/sync#Pool
- Buffer pool patterns: https://go.dev/blog/pprof
- Memory profiling: `go test -memprofile mem.prof -bench .`

---
*Generated: 2025-12-06*
*Status: Planning Document*
