package util

type Matrix[T any] struct {
	data []T
	a    int
	b    int
	c    int
}

func NewMatrix3D[T any](a int, b int, c int) *Matrix[T] {
	m := &Matrix[T]{
		data: make([]T, a*b*c),
		a:    a,
		b:    b,
		c:    c,
	}
	return m
}

func (m *Matrix[T]) Get(i int, j int, k int) T {
	return m.data[i*m.b*m.c+j*m.c+k]
}

func (m *Matrix[T]) GetDimension1(i int) []T {

	first := 0
	if i > 0 {
		first = (i - 1) * m.b * m.c
	}

	last := i * m.b * m.c
	return m.data[first:last]
}

func (m *Matrix[T]) Set(i int, j int, k int, val T) {
	m.data[i*m.b*m.c+j*m.c+k] = val
}

// single dimension len
func (m *Matrix[T]) Len() int {
	return m.a
}

func IfThenElse[T any](condition bool, a T, b T) T {
	if condition {
		return a
	}
	return b
}

func MakeMatrix2D[T any](a int, b int) [][]T {
	matrix := make([][]T, a)
	for i, _ := range matrix {
		matrix[i] = make([]T, b)
	}
	return matrix
}

func MakeMatrix3D[T any](a int, b int, c int) [][][]T {
	matrix := make([][][]T, a)
	for i, _ := range matrix {
		matrix[i] = make([][]T, b)
		for j, _ := range matrix[i] {
			matrix[i][j] = make([]T, c)
		}
	}
	return matrix
}

func MakeMatrix4D[T any](a int, b int, c int, d int) [][][][]T {
	matrix := make([][][][]T, a)
	for i, _ := range matrix {
		matrix[i] = make([][][]T, b)
		for j, _ := range matrix[i] {
			matrix[i][j] = make([][]T, c)
			for k, _ := range matrix[i][j] {
				matrix[i][j][k] = make([]T, d)
			}
		}
	}
	return matrix
}

func FillFloat32(a []float32, fromIndex uint32, toIndex uint32, val float32) {
	for i := fromIndex; i < toIndex; i++ {
		a[i] = val
	}
}

func Add[T any](slice []T, index int, elem T) []T {
	newSlice := append(slice[:index], elem)
	newSlice = append(newSlice, slice[index:]...)
	return newSlice
}
