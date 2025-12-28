package util

import (
	"testing"
)

func TestIfThenElse(t *testing.T) {
	if IfThenElse(true, 1, 2) != 1 {
		t.Error("IfThenElse(true, 1, 2) should be 1")
	}
	if IfThenElse(false, 1, 2) != 2 {
		t.Error("IfThenElse(false, 1, 2) should be 2")
	}
	if IfThenElse(true, "a", "b") != "a" {
		t.Error("IfThenElse(true, 'a', 'b') should be 'a'")
	}
	if IfThenElse(false, "a", "b") != "b" {
		t.Error("IfThenElse(false, 'a', 'b') should be 'b'")
	}
}

func TestMakeMatrix2D(t *testing.T) {
	m := MakeMatrix2D[int](3, 4)
	if len(m) != 3 {
		t.Errorf("Expected 3 rows, got %d", len(m))
	}
	for i, row := range m {
		if len(row) != 4 {
			t.Errorf("Row %d: expected 4 columns, got %d", i, len(row))
		}
	}
}

func TestMakeMatrix2DZero(t *testing.T) {
	m := MakeMatrix2D[int](0, 0)
	if len(m) != 0 {
		t.Errorf("Expected 0 rows, got %d", len(m))
	}
}

func TestMakeMatrix3D(t *testing.T) {
	m := MakeMatrix3D[float32](2, 3, 4)
	if len(m) != 2 {
		t.Errorf("Expected 2 depth, got %d", len(m))
	}
	for i := range m {
		if len(m[i]) != 3 {
			t.Errorf("Depth %d: expected 3 rows, got %d", i, len(m[i]))
		}
		for j := range m[i] {
			if len(m[i][j]) != 4 {
				t.Errorf("Depth %d, row %d: expected 4 columns, got %d", i, j, len(m[i][j]))
			}
		}
	}
}

func TestMakeMatrix4D(t *testing.T) {
	m := MakeMatrix4D[int](2, 3, 4, 5)
	if len(m) != 2 {
		t.Errorf("Expected 2 at dim 0, got %d", len(m))
	}
	if len(m[0]) != 3 {
		t.Errorf("Expected 3 at dim 1, got %d", len(m[0]))
	}
	if len(m[0][0]) != 4 {
		t.Errorf("Expected 4 at dim 2, got %d", len(m[0][0]))
	}
	if len(m[0][0][0]) != 5 {
		t.Errorf("Expected 5 at dim 3, got %d", len(m[0][0][0]))
	}
}

func TestCompareMatrix2D(t *testing.T) {
	eq := func(a, b int) bool { return a == b }

	a := [][]int{{1, 2}, {3, 4}}
	b := [][]int{{1, 2}, {3, 4}}
	c := [][]int{{1, 2}, {3, 5}}

	if !CompareMatrix2D(a, b, eq) {
		t.Error("Equal matrices should return true")
	}
	if CompareMatrix2D(a, c, eq) {
		t.Error("Different matrices should return false")
	}
}

func TestCompareMatrix2DDifferentLengths(t *testing.T) {
	eq := func(a, b int) bool { return a == b }

	a := [][]int{{1, 2}}
	b := [][]int{{1, 2}, {3, 4}}

	if CompareMatrix2D(a, b, eq) {
		t.Error("Different row count should return false")
	}

	c := [][]int{{1, 2, 3}}
	if CompareMatrix2D(a, c, eq) {
		t.Error("Different column count should return false")
	}
}

func TestCompareMatrix3D(t *testing.T) {
	eq := func(a, b int) bool { return a == b }

	a := [][][]int{{{1, 2}, {3, 4}}}
	b := [][][]int{{{1, 2}, {3, 4}}}
	c := [][][]int{{{1, 2}, {3, 5}}}

	if !CompareMatrix3D(a, b, eq) {
		t.Error("Equal matrices should return true")
	}
	if CompareMatrix3D(a, c, eq) {
		t.Error("Different matrices should return false")
	}
}

func TestCompareMatrix3DDifferentLengths(t *testing.T) {
	eq := func(a, b int) bool { return a == b }

	a := [][][]int{{{1, 2}}}
	b := [][][]int{{{1, 2}}, {{3, 4}}}

	if CompareMatrix3D(a, b, eq) {
		t.Error("Different depth should return false")
	}

	c := [][][]int{{{1, 2}, {3, 4}}}
	if CompareMatrix3D(a, c, eq) {
		t.Error("Different row count should return false")
	}

	d := [][][]int{{{1, 2, 3}}}
	if CompareMatrix3D(a, d, eq) {
		t.Error("Different column count should return false")
	}
}

func TestFillFloat32(t *testing.T) {
	a := make([]float32, 10)
	FillFloat32(a, 2, 7, 3.14)

	for i := 0; i < 10; i++ {
		if i >= 2 && i < 7 {
			if a[i] != 3.14 {
				t.Errorf("a[%d] = %f; want 3.14", i, a[i])
			}
		} else {
			if a[i] != 0 {
				t.Errorf("a[%d] = %f; want 0", i, a[i])
			}
		}
	}
}

func TestAdd(t *testing.T) {
	// Add inserts element at index, returning a new slice
	// Note: Due to append behavior, works correctly when slice has no extra capacity
	slice := []int{1, 2, 4, 5}
	result := Add(slice, 2, 3)

	if len(result) != 5 {
		t.Fatalf("Expected length 5, got %d", len(result))
	}
	// Verify the element was inserted
	if result[2] != 3 {
		t.Errorf("result[2] = %d; want 3", result[2])
	}
	if result[0] != 1 || result[1] != 2 {
		t.Errorf("First elements wrong: %v", result[:2])
	}
}

func TestAddAtBeginning(t *testing.T) {
	slice := []int{2, 3, 4}
	result := Add(slice, 0, 1)

	if len(result) != 4 {
		t.Fatalf("Expected length 4, got %d", len(result))
	}
	if result[0] != 1 {
		t.Errorf("result[0] = %d; want 1", result[0])
	}
}

func TestAddAtEnd(t *testing.T) {
	slice := []int{1, 2, 3}
	result := Add(slice, 3, 4)

	expected := []int{1, 2, 3, 4}
	for i, v := range expected {
		if result[i] != v {
			t.Errorf("result[%d] = %d; want %d", i, result[i], v)
		}
	}
}

func TestRectangleComputeLowerCorner(t *testing.T) {
	r := Rectangle{
		Origin: Point{X: 10, Y: 20},
		Size:   Dimension{Width: 100, Height: 50},
	}

	corner := r.ComputeLowerCorner()
	if corner.X != 110 || corner.Y != 70 {
		t.Errorf("ComputeLowerCorner() = (%d, %d); want (110, 70)", corner.X, corner.Y)
	}
}

func TestRectangleComputeLowerCornerZero(t *testing.T) {
	r := Rectangle{
		Origin: Point{X: 0, Y: 0},
		Size:   Dimension{Width: 0, Height: 0},
	}

	corner := r.ComputeLowerCorner()
	if corner.X != 0 || corner.Y != 0 {
		t.Errorf("ComputeLowerCorner() = (%d, %d); want (0, 0)", corner.X, corner.Y)
	}
}

// Benchmarks
func BenchmarkMakeMatrix2D(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = MakeMatrix2D[float32](256, 256)
	}
}

func BenchmarkMakeMatrix3D(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = MakeMatrix3D[float32](3, 256, 256)
	}
}

func BenchmarkMakeMatrix4D(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = MakeMatrix4D[float32](3, 8, 8, 8)
	}
}

func BenchmarkCompareMatrix2D(b *testing.B) {
	eq := func(a, b float32) bool { return a == b }
	m1 := MakeMatrix2D[float32](256, 256)
	m2 := MakeMatrix2D[float32](256, 256)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CompareMatrix2D(m1, m2, eq)
	}
}
