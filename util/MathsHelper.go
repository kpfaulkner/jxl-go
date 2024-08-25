package util

import (
	"cmp"
	"errors"
	"math"
	"math/bits"
)

func SignedPow(base float32, exponent float32) float32 {
	if base < 0 {
		return float32(math.Pow(float64(-base), float64(exponent)))
	}
	return float32(math.Pow(float64(base), float64(exponent)))
}

func CeilLog1p(x int64) int {
	xx := bits.LeadingZeros64(uint64(x))
	return 64 - xx
}

func CeilLog2(x int64) int {
	return CeilLog1p(x - 1)
}

func Max[T cmp.Ordered](args ...T) T {
	if len(args) == 0 {
		return *new(T)
	}

	if isNan(args[0]) {
		return args[0]
	}

	max := args[0]
	for _, arg := range args[1:] {

		if isNan(arg) {
			return arg
		}

		if arg > max {
			max = arg
		}
	}
	return max
}

func Min[T cmp.Ordered](args ...T) T {
	if len(args) == 0 {
		return *new(T)
	}

	if isNan(args[0]) {
		return args[0]
	}

	min := args[0]
	for _, arg := range args[1:] {

		if isNan(arg) {
			return arg
		}

		if arg < min {
			min = arg
		}
	}
	return min
}

func isNan[T comparable](arg T) bool {
	return arg != arg
}

func MatrixIdentity(i int) [][]float32 {
	matrix := make([][]float32, i)
	for j := 0; j < i; j++ {
		matrix[j] = make([]float32, i)
		matrix[j][j] = 1
	}
	return matrix
}

func MatrixVectorMultiply(matrix [][]float32, columnVector []float32) ([]float32, error) {

	if len(matrix) == 0 {
		return columnVector, nil
	}

	if len(matrix[0]) > len(columnVector) || len(columnVector) == 0 {
		return nil, errors.New("Invalid argument")
	}
	extra := len(columnVector) - len(matrix[0])
	total := make([]float32, len(matrix)+extra)

	for y := 0; y < len(matrix); y++ {
		row := matrix[y]

		for x := 0; x < len(row); x++ {
			total[y] += row[x] * columnVector[x]
		}
	}
	if extra != 0 {
		copy(total[len(matrix):], columnVector[len(matrix[0]):])
	}

	return total, nil
}

// multiply any number of matrices
func MatrixMultiply(matrix ...[][]float32) ([][]float32, error) {

	var err error
	left := matrix[0]
	for i := 1; i < len(matrix); i++ {
		right := matrix[i]
		left, err = MatrixMatrixMultiply(left, right)
		if err != nil {
			return nil, err
		}
	}
	return left, nil
}

func MatrixMatrixMultiply(left [][]float32, right [][]float32) ([][]float32, error) {
	if len(left[0]) != len(right) {
		return nil, errors.New("Invalid argument")
	}

	result := make([][]float32, len(left))
	for i := 0; i < len(left); i++ {
		result[i] = make([]float32, len(right[0]))
	}

	for i := 0; i < len(left); i++ {
		for j := 0; j < len(right[0]); j++ {
			for k := 0; k < len(right); k++ {
				result[i][j] += left[i][k] * right[k][j]
			}
		}
	}
	return result, nil
}

func InvertMatrix3x3(matrix [][]float32) [][]float32 {
	det := matrix[0][0]*matrix[1][1]*matrix[2][2] + matrix[0][1]*matrix[1][2]*matrix[2][0] + matrix[0][2]*matrix[1][0]*matrix[2][1] - matrix[0][2]*matrix[1][1]*matrix[2][0] - matrix[0][1]*matrix[1][0]*matrix[2][2] - matrix[0][0]*matrix[1][2]*matrix[2][1]
	if det == 0 {
		return nil
	}
	invDet := 1.0 / det
	return [][]float32{
		{(matrix[1][1]*matrix[2][2] - matrix[1][2]*matrix[2][1]) * invDet, (matrix[0][2]*matrix[2][1] - matrix[0][1]*matrix[2][2]) * invDet, (matrix[0][1]*matrix[1][2] - matrix[0][2]*matrix[1][1]) * invDet},
		{(matrix[1][2]*matrix[2][0] - matrix[1][0]*matrix[2][2]) * invDet, (matrix[0][0]*matrix[2][2] - matrix[0][2]*matrix[2][0]) * invDet, (matrix[0][2]*matrix[1][0] - matrix[0][0]*matrix[1][2]) * invDet},
		{(matrix[1][0]*matrix[2][1] - matrix[1][1]*matrix[2][0]) * invDet, (matrix[0][1]*matrix[2][0] - matrix[0][0]*matrix[2][1]) * invDet, (matrix[0][0]*matrix[1][1] - matrix[0][1]*matrix[1][0]) * invDet},
	}
}

func CeilDiv(numerator uint32, denominator uint32) uint32 {
	return ((numerator - 1) / denominator) + 1
}

func TransposeMatrix(matrix [][]float32, inSize IntPoint) [][]float32 {
	if inSize.X == 0 || inSize.y == 0 {
		return nil
	}
	dest := make([][]float32, inSize.X)
	transposeMatrixInto(matrix, dest, ZERO, ZERO, inSize)
	return dest
}

func transposeMatrixInto(src [][]float32, dest [][]float32, srcStart IntPoint, destStart IntPoint, srcSize IntPoint) {
	for y := uint32(0); y < srcSize.y; y++ {
		srcY := src[y+srcStart.y]
		for x := uint32(0); x < srcSize.X; x++ {
			dest[destStart.y+x][destStart.X+y] = srcY[srcStart.X+x]
		}
	}
}
