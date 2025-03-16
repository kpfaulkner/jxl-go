package util

import (
	"golang.org/x/exp/constraints"
)

// Make 1D slice appear as 2D slice and helper functions
// Trying out generic approach... if too slow will make type specific versions

type Matrix[T constraints.Ordered] struct {
	Width  int32
	Height int32
	Data   []T
}

// New2DMatrix creates a new 2D matrix with the given dimensions
// Note height is the first dimension, width is the second
func New2DMatrix[T constraints.Ordered](height int32, width int32) *Matrix[T] {
	matrix := make([]T, width*height)
	return &Matrix[T]{Width: width, Height: height, Data: matrix}
}

func New2DMatrixWithContents[T constraints.Ordered](height int32, width int32, initialData [][]T) *Matrix[T] {
	matrix := New2DMatrix[T](width, height)
	for h := int32(0); h < height; h++ {
		copy(matrix.Data[h*width:(h+1)*width], initialData[h])
	}
	return matrix
}

// Note y is first param...  just for compatibility
func (s *Matrix[T]) Get(y int32, x int32) T {
	return s.Data[y*s.Width+x]
}

func (s *Matrix[T]) Set(y int32, x int32, value T) {
	s.Data[y*s.Width+x] = value
}

func (s *Matrix[T]) IncrementBy(y int32, x int32, value T) {
	s.Data[y*s.Width+x] += value
}

func (s *Matrix[T]) GetRow(y int32) []T {
	return s.Data[y*s.Width : (y+1)*s.Width]
}

func (s *Matrix[T]) SetRow(y int32, data []T) {
	copy(s.Data[y*s.Width:(y+1)*s.Width], data)
}

// actually generate 2D slice...  just to have this as a temporary step until
// other code refactored to use Matrix struct
func (s *Matrix[T]) GetAs2DSlice() [][]T {
	a := make([][]T, s.Height)
	for height := 0; height < int(s.Height); height++ {
		a[height] = s.GetRow(int32(height))
	}
	return a
}
