package util

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
