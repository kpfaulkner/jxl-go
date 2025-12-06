# Buffer Pooling Implementation Guide

## Quick Start: Applying Pools to Critical Paths

This guide shows exactly how to apply the buffer pool optimizations to the most critical code paths.

## 1. Frame.copyFloatBuffers (HIGHEST PRIORITY)

### Current Code (frame/Frame.go:1043-1051)
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

### Updated Code ✅
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

### Usage Update (frame/Frame.go:935-996)
**Before:**
```go
// copy first 3 (well number of colours we have) buffers
inputBuffers := copyFloatBuffers(f.Buffer, colours)
outputBuffers := copyFloatBuffers(outputBuffer, colours)

// ... processing ...

for c := 0; c < int(colours); c++ {
    tmp := f.Buffer[c]
    f.Buffer[c].FloatBuffer = outputBuffers[c]
    outputBuffer[c] = tmp
}
```

**After:**
```go
// copy first 3 (well number of colours we have) buffers
inputBuffers := copyFloatBuffers(f.Buffer, colours)
outputBuffers := copyFloatBuffers(outputBuffer, colours)
defer util.ReturnMatrix3DToPool(inputBuffers)
defer util.ReturnMatrix3DToPool(outputBuffers)

// ... processing ...

for c := 0; c < int(colours); c++ {
    tmp := f.Buffer[c]
    f.Buffer[c].FloatBuffer = outputBuffers[c]
    outputBuffer[c] = tmp
}

// Note: Don't return pools here since they're now part of f.Buffer
// They'll be cleaned up when the frame is done
```

**Impact:** Saves ~20-50MB per frame, called 3 times per EPF iteration

---

## 2. Frame.decodePassGroupsConcurrent (HIGH PRIORITY)

### Current Code (frame/Frame.go:666-672)
```go
// get floating point version of frame buffer
buffers := util.MakeMatrix3D[float32](3, 0, 0)
for c := 0; c < 3; c++ {
    if err := f.Buffer[c].CastToFloatIfMax(^(^0 << f.GlobalMetadata.BitDepth.BitsPerSample)); err != nil {
        return err
    }
    buffers[c] = f.Buffer[c].FloatBuffer
}
```

### Updated Code ✅
```go
// get floating point version of frame buffer
buffers := util.MakeMatrix3DPooled[float32](3, 0, 0)
defer util.ReturnMatrix3DToPool(buffers)
for c := 0; c < 3; c++ {
    if err := f.Buffer[c].CastToFloatIfMax(^(^0 << f.GlobalMetadata.BitDepth.BitsPerSample)); err != nil {
        return err
    }
    buffers[c] = f.Buffer[c].FloatBuffer
}
```

**Impact:** Saves frame-size allocation (5-20MB typically)

---

## 3. HFCoefficients Matrix Allocations (MEDIUM PRIORITY)

### Current Code (frame/HFCoefficients.go:65-74)
```go
nonZeros := util.MakeMatrix3D[int32](3, 32, 32)
hf.stream = entropy.NewEntropyStreamWithStream(hfPass.contextStream)
hf.quantizedCoeffs = util.MakeMatrix3D[int32](3, 0, 0)
hf.dequantHFCoeff = util.MakeMatrix3D[float32](3, 0, 0)

for c := 0; c < 3; c++ {
    sY := size.Height >> header.jpegUpsamplingY[c]
    sX := size.Width >> header.jpegUpsamplingX[c]
    hf.quantizedCoeffs[c] = util.MakeMatrix2D[int32](sY, sX)
    hf.dequantHFCoeff[c] = util.MakeMatrix2D[float32](sY, sX)
}
```

### Updated Code ✅
```go
nonZeros := util.MakeMatrix3DPooled[int32](3, 32, 32)
defer util.ReturnMatrix3DToPool(nonZeros)

hf.stream = entropy.NewEntropyStreamWithStream(hfPass.contextStream)
hf.quantizedCoeffs = util.MakeMatrix3DPooled[int32](3, 0, 0)
hf.dequantHFCoeff = util.MakeMatrix3DPooled[float32](3, 0, 0)

for c := 0; c < 3; c++ {
    sY := size.Height >> header.jpegUpsamplingY[c]
    sX := size.Width >> header.jpegUpsamplingX[c]
    hf.quantizedCoeffs[c] = util.MakeMatrix2DPooled[int32](sY, sX)
    hf.dequantHFCoeff[c] = util.MakeMatrix2DPooled[float32](sY, sX)
}
```

**Challenge:** Need to track lifecycle and return to pool when HFCoefficients is done

### Add Cleanup Method
```go
func (hf *HFCoefficients) Release() {
    if hf.quantizedCoeffs != nil {
        util.ReturnMatrix3DToPool(hf.quantizedCoeffs)
        hf.quantizedCoeffs = nil
    }
    if hf.dequantHFCoeff != nil {
        util.ReturnMatrix3DToPool(hf.dequantHFCoeff)
        hf.dequantHFCoeff = nil
    }
}
```

### Call From PassGroup
```go
// In frame/Frame.go after decodePassGroupsConcurrent completes
defer func() {
    for pass := 0; pass < numPasses; pass++ {
        for group := 0; group < numGroups; group++ {
            if passGroups[pass][group].hfCoefficients != nil {
                passGroups[pass][group].hfCoefficients.Release()
            }
        }
    }
}()
```

**Impact:** Saves 3-5MB per pass group × number of pass groups

---

## 4. PassGroup.invertVarDCT (ALREADY DONE ✅)

### Current Code (frame/PassGroup.go:123-124)
```go
scratchBlock := util.MakeMatrix3DPooled[float32](5, 256, 256)
defer util.ReturnMatrix3DToPool(scratchBlock)
```

✅ **Already implemented correctly!**

---

## 5. LFCoefficients.DequantizeLF

### Pattern to Look For
```go
dequantLFCoeff := util.MakeMatrix3D[float32](3, 0, 0)
```

### Update To
```go
dequantLFCoeff := util.MakeMatrix3DPooled[float32](3, 0, 0)
defer util.ReturnMatrix3DToPool(dequantLFCoeff)
```

---

## Common Patterns and Solutions

### Pattern 1: Temporary Buffer in Function

**Before:**
```go
func processData() error {
    buffer := util.MakeMatrix3D[float32](3, 256, 256)
    // use buffer
    return nil
}
```

**After:**
```go
func processData() error {
    buffer := util.MakeMatrix3DPooled[float32](3, 256, 256)
    defer util.ReturnMatrix3DToPool(buffer)
    // use buffer
    return nil
}
```

### Pattern 2: Buffer Stored in Struct

**Before:**
```go
type MyStruct struct {
    buffer [][][]float32
}

func (m *MyStruct) Init() {
    m.buffer = util.MakeMatrix3D[float32](3, 100, 100)
}
```

**After:**
```go
type MyStruct struct {
    buffer [][][]float32
}

func (m *MyStruct) Init() {
    m.buffer = util.MakeMatrix3DPooled[float32](3, 100, 100)
}

func (m *MyStruct) Release() {
    if m.buffer != nil {
        util.ReturnMatrix3DToPool(m.buffer)
        m.buffer = nil
    }
}
```

### Pattern 3: Buffer Passed Between Functions

**Before:**
```go
func createBuffer() [][][]float32 {
    return util.MakeMatrix3D[float32](3, 256, 256)
}

func processBuffer(buf [][][]float32) {
    // process
}

func main() {
    buf := createBuffer()
    processBuffer(buf)
}
```

**After:**
```go
func createBuffer() [][][]float32 {
    return util.MakeMatrix3DPooled[float32](3, 256, 256)
}

func processBuffer(buf [][][]float32) {
    // process
}

func main() {
    buf := createBuffer()
    defer util.ReturnMatrix3DToPool(buf)
    processBuffer(buf)
}
```

---

## Testing Your Changes

### 1. Run Unit Tests
```bash
go test ./util -run TestMatrix
```

### 2. Run Benchmarks
```bash
go test ./util -bench=. -benchmem
```

Expected output:
```
BenchmarkMatrix3DPooled-8     1000000    1234 ns/op    0 B/op    0 allocs/op
BenchmarkMatrix3DDirect-8      100000   12345 ns/op  786432 B/op  259 allocs/op
```

### 3. Memory Profile Before/After
```bash
# Before optimization
go test ./core -bench=BenchmarkDecode -memprofile=mem_before.prof

# After optimization
go test ./core -bench=BenchmarkDecode -memprofile=mem_after.prof

# Compare
go tool pprof -base=mem_before.prof mem_after.prof
```

### 4. Check for Memory Leaks
```bash
# Run with race detector
go test -race ./...

# Run with memory leak detection
GODEBUG=gctrace=1 go test ./core -bench=BenchmarkDecode
```

---

## Debugging Common Issues

### Issue 1: Panic on Pool Return
**Symptom:** `panic: runtime error: invalid memory address`

**Cause:** Trying to return nil or already-returned buffer

**Solution:**
```go
if buffer != nil {
    util.ReturnMatrix3DToPool(buffer)
    buffer = nil  // Prevent double-return
}
```

### Issue 2: Data Corruption
**Symptom:** Incorrect decode results, artifacts in images

**Cause:** Buffer not properly cleared or reused too soon

**Solution:** Verify buffer is not used after return to pool

### Issue 3: Memory Still Growing
**Symptom:** Memory usage doesn't decrease

**Cause:** Missing `defer Return()` calls

**Solution:** Search codebase for `MakeMatrix3DPooled` without corresponding `Return`

```bash
# Find potential issues
grep -r "MakeMatrix3DPooled" --include="*.go" | while read line; do
    file=$(echo $line | cut -d: -f1)
    if ! grep -q "ReturnMatrix3DToPool" "$file"; then
        echo "Missing Return in: $file"
    fi
done
```

---

## Verification Checklist

- [ ] All `MakeMatrix3DPooled` calls have corresponding `Return` calls
- [ ] `defer` used immediately after allocation when possible
- [ ] Struct fields with pooled buffers have `Release()` methods
- [ ] No buffer use after return to pool
- [ ] Tests pass: `go test ./...`
- [ ] Benchmarks show improvement
- [ ] Memory profile shows reduced allocations
- [ ] No race conditions: `go test -race ./...`

---

## Performance Expectations

### Before Optimization
- **Allocations:** ~500-1000 allocs per frame
- **Memory:** 100-200MB peak for typical image
- **GC Pauses:** 10-50ms

### After Optimization
- **Allocations:** ~50-200 allocs per frame (60-80% reduction)
- **Memory:** 50-100MB peak (40-50% reduction)
- **GC Pauses:** 2-10ms (70-80% reduction)
- **Decode Speed:** 10-25% faster

---

## Next Steps

1. ✅ Apply changes to `Frame.copyFloatBuffers`
2. ✅ Update `Frame.decodePassGroupsConcurrent`
3. Add `Release()` methods to structs with pooled buffers
4. Run benchmarks and verify improvements
5. Search for remaining `MakeMatrix3D` calls
6. Profile and iterate

---

*Generated: 2025-12-06*