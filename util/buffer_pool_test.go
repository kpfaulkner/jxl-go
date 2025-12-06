package util

import (
	"runtime"
	"testing"
)

func TestMatrix3DPool(t *testing.T) {
	// Test basic get and put
	matrix := MakeMatrix3DPooled[float32](3, 256, 256)
	if len(matrix) != 3 || len(matrix[0]) != 256 || len(matrix[0][0]) != 256 {
		t.Errorf("Matrix dimensions incorrect")
	}

	// Modify matrix
	matrix[0][0][0] = 42.0

	// Return to pool
	ReturnMatrix3DToPool(matrix)

	// Get again - should be cleared
	matrix2 := MakeMatrix3DPooled[float32](3, 256, 256)
	if matrix2[0][0][0] != 0.0 {
		t.Errorf("Matrix not cleared after return to pool: got %f", matrix2[0][0][0])
	}

	ReturnMatrix3DToPool(matrix2)
}

func TestMatrix2DPool(t *testing.T) {
	matrix := MakeMatrix2DPooled[int32](100, 100)
	if len(matrix) != 100 || len(matrix[0]) != 100 {
		t.Errorf("Matrix dimensions incorrect")
	}

	matrix[0][0] = 42

	ReturnMatrix2DToPool(matrix)

	matrix2 := MakeMatrix2DPooled[int32](100, 100)
	if matrix2[0][0] != 0 {
		t.Errorf("Matrix not cleared after return to pool")
	}

	ReturnMatrix2DToPool(matrix2)
}

func TestPoolMetrics(t *testing.T) {
	// Clear any previous state by getting fresh matrices
	for i := 0; i < 5; i++ {
		m := MakeMatrix3DPooled[float32](2, 2, 2)
		ReturnMatrix3DToPool(m)
	}

	metrics := GetPoolMetrics()
	if metrics["float32_3d"]["hits"] == 0 && metrics["float32_3d"]["misses"] == 0 {
		t.Errorf("No metrics recorded")
	}

	t.Logf("Metrics: %+v", metrics)
}

func TestDifferentSizes(t *testing.T) {
	sizes := []struct {
		d, h, w int
	}{
		{3, 256, 256},
		{5, 256, 256},
		{3, 512, 512},
		{1, 100, 100},
	}

	for _, size := range sizes {
		matrix := MakeMatrix3DPooled[float32](size.d, size.h, size.w)
		if len(matrix) != size.d {
			t.Errorf("Size %v: incorrect depth", size)
		}
		ReturnMatrix3DToPool(matrix)
	}
}

func TestZeroSize(t *testing.T) {
	// Should not panic
	matrix := MakeMatrix3DPooled[float32](0, 0, 0)
	if matrix != nil && len(matrix) != 0 {
		t.Errorf("Expected empty matrix for zero size")
	}
	ReturnMatrix3DToPool(matrix) // Should not panic
}

func TestConcurrentAccess(t *testing.T) {
	const goroutines = 10
	const iterations = 100

	done := make(chan bool, goroutines)

	for g := 0; g < goroutines; g++ {
		go func() {
			for i := 0; i < iterations; i++ {
				matrix := MakeMatrix3DPooled[float32](3, 64, 64)
				matrix[0][0][0] = float32(i)
				ReturnMatrix3DToPool(matrix)
			}
			done <- true
		}()
	}

	for g := 0; g < goroutines; g++ {
		<-done
	}
}

// Benchmarks

func BenchmarkMatrix3DPooled(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		matrix := MakeMatrix3DPooled[float32](3, 256, 256)
		ReturnMatrix3DToPool(matrix)
	}
}

func BenchmarkMatrix3DDirect(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = MakeMatrix3D[float32](3, 256, 256)
	}
}

func BenchmarkMatrix2DPooled(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		matrix := MakeMatrix2DPooled[float32](256, 256)
		ReturnMatrix2DToPool(matrix)
	}
}

func BenchmarkMatrix2DDirect(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = MakeMatrix2D[float32](256, 256)
	}
}

// Memory pressure benchmark
func BenchmarkMemoryPressurePooled(b *testing.B) {
	b.ReportAllocs()

	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	before := m.TotalAlloc

	for i := 0; i < b.N; i++ {
		// Simulate typical usage in decoder
		scratch := MakeMatrix3DPooled[float32](5, 256, 256)
		coeffs := MakeMatrix3DPooled[float32](3, 512, 512)

		// Simulate work
		scratch[0][0][0] = float32(i)
		coeffs[0][0][0] = float32(i)

		ReturnMatrix3DToPool(scratch)
		ReturnMatrix3DToPool(coeffs)
	}

	runtime.ReadMemStats(&m)
	after := m.TotalAlloc
	b.ReportMetric(float64(after-before)/float64(b.N), "B/op")
}

func BenchmarkMemoryPressureDirect(b *testing.B) {
	b.ReportAllocs()

	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	before := m.TotalAlloc

	for i := 0; i < b.N; i++ {
		// Simulate typical usage without pooling
		scratch := MakeMatrix3D[float32](5, 256, 256)
		coeffs := MakeMatrix3D[float32](3, 512, 512)

		// Simulate work
		scratch[0][0][0] = float32(i)
		coeffs[0][0][0] = float32(i)
	}

	runtime.ReadMemStats(&m)
	after := m.TotalAlloc
	b.ReportMetric(float64(after-before)/float64(b.N), "B/op")
}

func BenchmarkConcurrentPoolAccess(b *testing.B) {
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			matrix := MakeMatrix3DPooled[float32](3, 128, 128)
			matrix[0][0][0] = 1.0
			ReturnMatrix3DToPool(matrix)
		}
	})
}