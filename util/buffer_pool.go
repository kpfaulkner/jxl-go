package util

import (
	"sync"
	"sync/atomic"
)

type poolKey3D struct {
	d, h, w int
}

// Matrix3DPool provides pooling for 3D matrices to reduce allocations
type Matrix3DPool[T any] struct {
	pools map[poolKey3D]*sync.Pool
	mu    sync.RWMutex

	// Metrics
	hits   atomic.Int64
	misses atomic.Int64
}

var (
	float32Pool3D = &Matrix3DPool[float32]{pools: make(map[poolKey3D]*sync.Pool)}
	int32Pool3D   = &Matrix3DPool[int32]{pools: make(map[poolKey3D]*sync.Pool)}

	float32Pool2D = &Matrix2DPool[float32]{pools: make(map[poolKey2D]*sync.Pool)}
	int32Pool2D   = &Matrix2DPool[int32]{pools: make(map[poolKey2D]*sync.Pool)}
)

// Get retrieves a matrix from the pool or creates a new one
func (p *Matrix3DPool[T]) Get(depth, height, width int) [][][]T {
	if depth == 0 || height == 0 || width == 0 {
		return MakeMatrix3D[T](depth, height, width)
	}

	key := poolKey3D{d: depth, h: height, w: width}

	// Fast path: read lock
	p.mu.RLock()
	pool, exists := p.pools[key]
	p.mu.RUnlock()

	if !exists {
		// Slow path: create new pool
		p.mu.Lock()
		// Double-check after acquiring write lock
		pool, exists = p.pools[key]
		if !exists {
			pool = &sync.Pool{
				New: func() interface{} {
					m := MakeMatrix3D[T](depth, height, width)
					return &m
				},
			}
			p.pools[key] = pool
		}
		p.mu.Unlock()
	}

	if matrix := pool.Get(); matrix != nil {
		p.hits.Add(1)
		return *(matrix.(*[][][]T))
	}

	p.misses.Add(1)
	return MakeMatrix3D[T](depth, height, width)
}

// Put returns a matrix to the pool after clearing it
func (p *Matrix3DPool[T]) Put(matrix [][][]T) {
	if len(matrix) == 0 {
		return
	}

	depth := len(matrix)
	height := len(matrix[0])
	if height == 0 {
		return
	}

	width := len(matrix[0][0])
	key := poolKey3D{d: depth, h: height, w: width}

	p.mu.RLock()
	pool, exists := p.pools[key]
	p.mu.RUnlock()

	if exists {
		// Clear the matrix before returning to pool
		var zero T
		for i := range matrix {
			for j := range matrix[i] {
				for k := range matrix[i][j] {
					matrix[i][j][k] = zero
				}
			}
		}
		pool.Put(&matrix)
	}
}

// GetMetrics returns pool usage statistics
func (p *Matrix3DPool[T]) GetMetrics() (hits, misses int64) {
	return p.hits.Load(), p.misses.Load()
}

type poolKey2D struct {
	h, w int
}

// Matrix2DPool provides pooling for 2D matrices
type Matrix2DPool[T any] struct {
	pools  map[poolKey2D]*sync.Pool
	mu     sync.RWMutex
	hits   atomic.Int64
	misses atomic.Int64
}

// Get retrieves a 2D matrix from the pool or creates a new one
func (p *Matrix2DPool[T]) Get(height, width int) [][]T {
	if height == 0 || width == 0 {
		return MakeMatrix2D[T](height, width)
	}

	key := poolKey2D{h: height, w: width}

	p.mu.RLock()
	pool, exists := p.pools[key]
	p.mu.RUnlock()

	if !exists {
		p.mu.Lock()
		pool, exists = p.pools[key]
		if !exists {
			pool = &sync.Pool{
				New: func() interface{} {
					m := MakeMatrix2D[T](height, width)
					return &m
				},
			}
			p.pools[key] = pool
		}
		p.mu.Unlock()
	}

	if matrix := pool.Get(); matrix != nil {
		p.hits.Add(1)
		return *(matrix.(*[][]T))
	}

	p.misses.Add(1)
	return MakeMatrix2D[T](height, width)
}

// Put returns a 2D matrix to the pool after clearing it
func (p *Matrix2DPool[T]) Put(matrix [][]T) {
	if len(matrix) == 0 {
		return
	}

	height := len(matrix)
	width := len(matrix[0])
	key := poolKey2D{h: height, w: width}

	p.mu.RLock()
	pool, exists := p.pools[key]
	p.mu.RUnlock()

	if exists {
		var zero T
		for i := range matrix {
			for j := range matrix[i] {
				matrix[i][j] = zero
			}
		}
		pool.Put(&matrix)
	}
}

// GetMetrics returns pool usage statistics
func (p *Matrix2DPool[T]) GetMetrics() (hits, misses int64) {
	return p.hits.Load(), p.misses.Load()
}

// Public API functions

// MakeMatrix3DPooled creates or retrieves a 3D matrix from the pool
func MakeMatrix3DPooled[T any](depth, height, width int) [][][]T {
	var zero T
	switch any(zero).(type) {
	case float32:
		return any(float32Pool3D.Get(depth, height, width)).([][][]T)
	case int32:
		return any(int32Pool3D.Get(depth, height, width)).([][][]T)
	default:
		// Fallback for unsupported types
		return MakeMatrix3D[T](depth, height, width)
	}
}

// ReturnMatrix3DToPool returns a 3D matrix to the pool
func ReturnMatrix3DToPool[T any](matrix [][][]T) {
	if len(matrix) == 0 {
		return
	}

	var zero T
	switch any(zero).(type) {
	case float32:
		float32Pool3D.Put(any(matrix).([][][]float32))
	case int32:
		int32Pool3D.Put(any(matrix).([][][]int32))
	}
}

// MakeMatrix2DPooled creates or retrieves a 2D matrix from the pool
func MakeMatrix2DPooled[T any](height, width int) [][]T {
	var zero T
	switch any(zero).(type) {
	case float32:
		return any(float32Pool2D.Get(height, width)).([][]T)
	case int32:
		return any(int32Pool2D.Get(height, width)).([][]T)
	default:
		return MakeMatrix2D[T](height, width)
	}
}

// ReturnMatrix2DToPool returns a 2D matrix to the pool
func ReturnMatrix2DToPool[T any](matrix [][]T) {
	if len(matrix) == 0 {
		return
	}

	var zero T
	switch any(zero).(type) {
	case float32:
		float32Pool2D.Put(any(matrix).([][]float32))
	case int32:
		int32Pool2D.Put(any(matrix).([][]int32))
	}
}

// GetPoolMetrics returns metrics for all pools
func GetPoolMetrics() map[string]map[string]int64 {
	f32_3d_hits, f32_3d_misses := float32Pool3D.GetMetrics()
	i32_3d_hits, i32_3d_misses := int32Pool3D.GetMetrics()
	f32_2d_hits, f32_2d_misses := float32Pool2D.GetMetrics()
	i32_2d_hits, i32_2d_misses := int32Pool2D.GetMetrics()

	return map[string]map[string]int64{
		"float32_3d": {
			"hits":   f32_3d_hits,
			"misses": f32_3d_misses,
		},
		"int32_3d": {
			"hits":   i32_3d_hits,
			"misses": i32_3d_misses,
		},
		"float32_2d": {
			"hits":   f32_2d_hits,
			"misses": f32_2d_misses,
		},
		"int32_2d": {
			"hits":   i32_2d_hits,
			"misses": i32_2d_misses,
		},
	}
}
