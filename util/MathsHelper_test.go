package util

import (
	"math"
	"testing"
)

// SignedPow tests
func TestSignedPow(t *testing.T) {
	tests := []struct {
		base     float32
		exponent float32
		expected float32
	}{
		{2.0, 2.0, 4.0},
		{2.0, 3.0, 8.0},
		{-2.0, 2.0, -4.0},
		{-2.0, 3.0, -8.0},
		{0.0, 2.0, 0.0},
		{4.0, 0.5, 2.0},
		{-8.0, 1.0 / 3.0, -2.0},
	}

	for _, tt := range tests {
		result := SignedPow(tt.base, tt.exponent)
		if math.Abs(float64(result-tt.expected)) > 0.0001 {
			t.Errorf("SignedPow(%f, %f) = %f; want %f", tt.base, tt.exponent, result, tt.expected)
		}
	}
}

// CeilLog1p tests
func TestCeilLog1p(t *testing.T) {
	tests := []struct {
		input    int
		expected int
	}{
		{0, 0},
		{1, 1},
		{2, 2},
		{3, 2},
		{4, 3},
		{7, 3},
		{8, 4},
		{15, 4},
		{16, 5},
		{255, 8},
		{256, 9},
	}

	for _, tt := range tests {
		result := CeilLog1p(tt.input)
		if result != tt.expected {
			t.Errorf("CeilLog1p(%d) = %d; want %d", tt.input, result, tt.expected)
		}
	}
}

func TestCeilLog1pUint64(t *testing.T) {
	tests := []struct {
		input    uint64
		expected int
	}{
		{0, 0},
		{1, 1},
		{255, 8},
		{256, 9},
		{1<<32 - 1, 32},
		{1 << 32, 33},
	}

	for _, tt := range tests {
		result := CeilLog1pUint64(tt.input)
		if result != tt.expected {
			t.Errorf("CeilLog1pUint64(%d) = %d; want %d", tt.input, result, tt.expected)
		}
	}
}

func TestFloorLog1p(t *testing.T) {
	// FloorLog1p computes floor(log2(x+1))
	tests := []struct {
		input    int
		expected int64
	}{
		{0, 0},  // floor(log2(1)) = 0
		{1, 1},  // floor(log2(2)) = 1
		{2, 1},  // floor(log2(3)) = 1
		{3, 2},  // floor(log2(4)) = 2
		{4, 2},  // floor(log2(5)) = 2
		{7, 3},  // floor(log2(8)) = 3
		{8, 3},  // floor(log2(9)) = 3
		{15, 4}, // floor(log2(16)) = 4
		{16, 4}, // floor(log2(17)) = 4
	}

	for _, tt := range tests {
		result := FloorLog1p(tt.input)
		if result != tt.expected {
			t.Errorf("FloorLog1p(%d) = %d; want %d", tt.input, result, tt.expected)
		}
	}
}

func TestFloorLog1pUint64(t *testing.T) {
	tests := []struct {
		input    uint64
		expected int64
	}{
		{0, 0},
		{1, 1},
		{3, 2},
		{7, 3},
		{15, 4},
	}

	for _, tt := range tests {
		result := FloorLog1pUint64(tt.input)
		if result != tt.expected {
			t.Errorf("FloorLog1pUint64(%d) = %d; want %d", tt.input, result, tt.expected)
		}
	}
}

func TestCeilLog2(t *testing.T) {
	tests := []struct {
		input    int
		expected int
	}{
		{1, 0},
		{2, 1},
		{3, 2},
		{4, 2},
		{5, 3},
		{8, 3},
		{9, 4},
		{16, 4},
		{17, 5},
	}

	for _, tt := range tests {
		result := CeilLog2(tt.input)
		if result != tt.expected {
			t.Errorf("CeilLog2(%d) = %d; want %d", tt.input, result, tt.expected)
		}
	}
}

// Max/Min tests
func TestMax(t *testing.T) {
	if Max(1, 2, 3) != 3 {
		t.Error("Max(1,2,3) should be 3")
	}
	if Max(3, 2, 1) != 3 {
		t.Error("Max(3,2,1) should be 3")
	}
	if Max(-1, -2, -3) != -1 {
		t.Error("Max(-1,-2,-3) should be -1")
	}
	if Max(5) != 5 {
		t.Error("Max(5) should be 5")
	}
	if Max(1.5, 2.5, 0.5) != 2.5 {
		t.Error("Max(1.5,2.5,0.5) should be 2.5")
	}
}

func TestMaxEmpty(t *testing.T) {
	result := Max[int]()
	if result != 0 {
		t.Errorf("Max() should return zero value, got %d", result)
	}
}

func TestMaxNaN(t *testing.T) {
	nan := float64(math.NaN())
	result := Max(nan, 1.0, 2.0)
	if !math.IsNaN(result) {
		t.Error("Max with NaN first should return NaN")
	}
	result = Max(1.0, nan, 2.0)
	if !math.IsNaN(result) {
		t.Error("Max with NaN in middle should return NaN")
	}
}

func TestMin(t *testing.T) {
	if Min(1, 2, 3) != 1 {
		t.Error("Min(1,2,3) should be 1")
	}
	if Min(3, 2, 1) != 1 {
		t.Error("Min(3,2,1) should be 1")
	}
	if Min(-1, -2, -3) != -3 {
		t.Error("Min(-1,-2,-3) should be -3")
	}
	if Min(5) != 5 {
		t.Error("Min(5) should be 5")
	}
}

func TestMinEmpty(t *testing.T) {
	result := Min[int]()
	if result != 0 {
		t.Errorf("Min() should return zero value, got %d", result)
	}
}

func TestMinNaN(t *testing.T) {
	nan := float64(math.NaN())
	result := Min(nan, 1.0, 2.0)
	if !math.IsNaN(result) {
		t.Error("Min with NaN first should return NaN")
	}
}

// Clamp tests
func TestClamp3(t *testing.T) {
	tests := []struct {
		v, a, b  int32
		expected int32
	}{
		{5, 0, 10, 5},   // in range
		{-5, 0, 10, 0},  // below range
		{15, 0, 10, 10}, // above range
		{5, 10, 0, 5},   // reversed bounds, in range
		{-5, 10, 0, 0},  // reversed bounds, below
		{15, 10, 0, 10}, // reversed bounds, above
	}

	for _, tt := range tests {
		result := Clamp3(tt.v, tt.a, tt.b)
		if result != tt.expected {
			t.Errorf("Clamp3(%d, %d, %d) = %d; want %d", tt.v, tt.a, tt.b, result, tt.expected)
		}
	}
}

func TestClamp3Float32(t *testing.T) {
	tests := []struct {
		v, a, b  float32
		expected float32
	}{
		{0.5, 0.0, 1.0, 0.5},
		{-0.5, 0.0, 1.0, 0.0},
		{1.5, 0.0, 1.0, 1.0},
		{0.5, 1.0, 0.0, 0.5}, // reversed bounds
	}

	for _, tt := range tests {
		result := Clamp3Float32(tt.v, tt.a, tt.b)
		if result != tt.expected {
			t.Errorf("Clamp3Float32(%f, %f, %f) = %f; want %f", tt.v, tt.a, tt.b, result, tt.expected)
		}
	}
}

func TestClamp(t *testing.T) {
	// Clamp(v, a, b, c) clamps v to [min(a,b), max(a,b)] with c potentially narrowing bounds
	// If lower >= c, lower becomes c; if upper <= c, upper becomes c
	tests := []struct {
		v, a, b, c int32
		expected   int32
	}{
		{5, 0, 10, 5, 5},   // v=5 in [0,10], stays 5
		{0, 0, 10, 5, 0},   // v=0 in [0,10], stays 0
		{10, 0, 10, 5, 10}, // v=10 in [0,10], stays 10
		{-5, 0, 10, 5, 0},  // v=-5 clamped to lower=0
		{15, 0, 10, 5, 10}, // v=15 clamped to upper=10
		{5, 10, 20, 15, 10}, // clamp 5 to [10,20] = 10
		{5, 10, 20, 5, 5},   // lower=10>=5 true->lower=5, clamp 5 to [5,20]=5
	}

	for _, tt := range tests {
		result := Clamp(tt.v, tt.a, tt.b, tt.c)
		if result != tt.expected {
			t.Errorf("Clamp(%d, %d, %d, %d) = %d; want %d", tt.v, tt.a, tt.b, tt.c, result, tt.expected)
		}
	}
}

// Abs tests
func TestAbs(t *testing.T) {
	if Abs(int32(5)) != 5 {
		t.Error("Abs(5) should be 5")
	}
	if Abs(int32(-5)) != 5 {
		t.Error("Abs(-5) should be 5")
	}
	if Abs(int32(0)) != 0 {
		t.Error("Abs(0) should be 0")
	}
	if Abs(int64(-100)) != 100 {
		t.Error("Abs(-100) should be 100")
	}
}

// MakeSliceWithDefault tests
func TestMakeSliceWithDefault(t *testing.T) {
	slice := MakeSliceWithDefault(5, 42)
	if len(slice) != 5 {
		t.Errorf("Expected length 5, got %d", len(slice))
	}
	for i, v := range slice {
		if v != 42 {
			t.Errorf("slice[%d] = %d; want 42", i, v)
		}
	}
}

func TestMakeSliceWithDefaultNegative(t *testing.T) {
	slice := MakeSliceWithDefault(-1, 42)
	if slice != nil {
		t.Error("Expected nil for negative length")
	}
}

// Matrix tests
func TestMatrixIdentity(t *testing.T) {
	m := MatrixIdentity(3)
	expected := [][]float32{
		{1, 0, 0},
		{0, 1, 0},
		{0, 0, 1},
	}

	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			if m[i][j] != expected[i][j] {
				t.Errorf("MatrixIdentity(3)[%d][%d] = %f; want %f", i, j, m[i][j], expected[i][j])
			}
		}
	}
}

func TestMatrixVectorMultiply(t *testing.T) {
	matrix := [][]float32{
		{1, 2, 3},
		{4, 5, 6},
	}
	vector := []float32{1, 2, 3}

	result, err := MatrixVectorMultiply(matrix, vector)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expected := []float32{14, 32} // 1*1+2*2+3*3=14, 4*1+5*2+6*3=32
	for i, v := range expected {
		if result[i] != v {
			t.Errorf("result[%d] = %f; want %f", i, result[i], v)
		}
	}
}

func TestMatrixVectorMultiplyEmpty(t *testing.T) {
	result, err := MatrixVectorMultiply([][]float32{}, []float32{1, 2, 3})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(result) != 3 {
		t.Error("Empty matrix should return original vector")
	}
}

func TestMatrixVectorMultiplyInvalid(t *testing.T) {
	matrix := [][]float32{{1, 2, 3}}
	vector := []float32{1, 2} // too short

	_, err := MatrixVectorMultiply(matrix, vector)
	if err == nil {
		t.Error("Expected error for mismatched dimensions")
	}
}

func TestMatrixVectorMultiplyEmptyVector(t *testing.T) {
	matrix := [][]float32{{1, 2, 3}}
	vector := []float32{}

	_, err := MatrixVectorMultiply(matrix, vector)
	if err == nil {
		t.Error("Expected error for empty vector")
	}
}

func TestMatrixVectorMultiplyExtra(t *testing.T) {
	matrix := [][]float32{
		{1, 0},
		{0, 1},
	}
	vector := []float32{2, 3, 4, 5} // extra elements

	result, err := MatrixVectorMultiply(matrix, vector)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Result should be [2, 3, 4, 5] (first 2 transformed, last 2 copied)
	if len(result) != 4 {
		t.Errorf("Expected length 4, got %d", len(result))
	}
	if result[0] != 2 || result[1] != 3 || result[2] != 4 || result[3] != 5 {
		t.Errorf("Unexpected result: %v", result)
	}
}

func TestMatrixMatrixMultiply(t *testing.T) {
	left := [][]float32{
		{1, 2},
		{3, 4},
	}
	right := [][]float32{
		{5, 6},
		{7, 8},
	}

	result, err := MatrixMatrixMultiply(left, right)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expected := [][]float32{
		{19, 22},  // 1*5+2*7, 1*6+2*8
		{43, 50},  // 3*5+4*7, 3*6+4*8
	}

	for i := 0; i < 2; i++ {
		for j := 0; j < 2; j++ {
			if result[i][j] != expected[i][j] {
				t.Errorf("result[%d][%d] = %f; want %f", i, j, result[i][j], expected[i][j])
			}
		}
	}
}

func TestMatrixMatrixMultiplyNil(t *testing.T) {
	matrix := [][]float32{{1, 2}, {3, 4}}

	result, err := MatrixMatrixMultiply(nil, matrix)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if result[0][0] != 1 {
		t.Error("nil left should return right")
	}

	result, err = MatrixMatrixMultiply(matrix, nil)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if result[0][0] != 1 {
		t.Error("nil right should return left")
	}
}

func TestMatrixMatrixMultiplyInvalid(t *testing.T) {
	left := [][]float32{{1, 2, 3}}
	right := [][]float32{{1, 2}}

	_, err := MatrixMatrixMultiply(left, right)
	if err == nil {
		t.Error("Expected error for mismatched dimensions")
	}
}

func TestMatrixMultiply(t *testing.T) {
	a := [][]float32{{1, 2}, {3, 4}}
	b := [][]float32{{1, 0}, {0, 1}}
	c := [][]float32{{2, 0}, {0, 2}}

	result, err := MatrixMultiply(a, b, c)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// a * identity * 2*identity = 2*a
	expected := [][]float32{{2, 4}, {6, 8}}
	for i := 0; i < 2; i++ {
		for j := 0; j < 2; j++ {
			if result[i][j] != expected[i][j] {
				t.Errorf("result[%d][%d] = %f; want %f", i, j, result[i][j], expected[i][j])
			}
		}
	}
}

func TestInvertMatrix3x3(t *testing.T) {
	// Test with identity matrix
	identity := [][]float32{
		{1, 0, 0},
		{0, 1, 0},
		{0, 0, 1},
	}
	result := InvertMatrix3x3(identity)

	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			if math.Abs(float64(result[i][j]-identity[i][j])) > 0.0001 {
				t.Errorf("Inverse of identity[%d][%d] = %f; want %f", i, j, result[i][j], identity[i][j])
			}
		}
	}
}

func TestInvertMatrix3x3NonIdentity(t *testing.T) {
	matrix := [][]float32{
		{1, 2, 3},
		{0, 1, 4},
		{5, 6, 0},
	}
	inverse := InvertMatrix3x3(matrix)

	// Multiply matrix by its inverse, should get identity
	product, _ := MatrixMatrixMultiply(matrix, inverse)

	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			expected := float32(0)
			if i == j {
				expected = 1
			}
			if math.Abs(float64(product[i][j]-expected)) > 0.0001 {
				t.Errorf("product[%d][%d] = %f; want %f", i, j, product[i][j], expected)
			}
		}
	}
}

func TestInvertMatrix3x3Singular(t *testing.T) {
	// Singular matrix (det = 0)
	singular := [][]float32{
		{1, 2, 3},
		{4, 5, 6},
		{7, 8, 9},
	}
	result := InvertMatrix3x3(singular)
	if result != nil {
		t.Error("Expected nil for singular matrix")
	}
}

// Transpose tests
func TestTransposeMatrix(t *testing.T) {
	matrix := [][]float32{
		{1, 2, 3},
		{4, 5, 6},
	}
	result := TransposeMatrix(matrix, Point{X: 3, Y: 2})

	expected := [][]float32{
		{1, 4},
		{2, 5},
		{3, 6},
	}

	for i := 0; i < 3; i++ {
		for j := 0; j < 2; j++ {
			if result[i][j] != expected[i][j] {
				t.Errorf("result[%d][%d] = %f; want %f", i, j, result[i][j], expected[i][j])
			}
		}
	}
}

func TestTransposeMatrixEmpty(t *testing.T) {
	result := TransposeMatrix([][]float32{}, Point{X: 0, Y: 0})
	if result != nil {
		t.Error("Expected nil for zero size")
	}
}

func TestTransposeMatrixInto(t *testing.T) {
	src := [][]float32{
		{1, 2},
		{3, 4},
	}
	dest := MakeMatrix2D[float32](2, 2)

	TransposeMatrixInto(src, dest, ZERO, ZERO, Point{X: 2, Y: 2})

	if dest[0][0] != 1 || dest[0][1] != 3 || dest[1][0] != 2 || dest[1][1] != 4 {
		t.Errorf("Unexpected transpose result: %v", dest)
	}
}

// Matrix3Equal tests
func TestMatrix3Equal(t *testing.T) {
	a := [][][]int{{{1, 2}, {3, 4}}}
	b := [][][]int{{{1, 2}, {3, 4}}}
	c := [][][]int{{{1, 2}, {3, 5}}}

	if !Matrix3Equal(a, b) {
		t.Error("Equal matrices should return true")
	}
	if Matrix3Equal(a, c) {
		t.Error("Different matrices should return false")
	}
}

func TestMatrix3EqualDifferentLengths(t *testing.T) {
	a := [][][]int{{{1, 2}}}
	b := [][][]int{{{1, 2}}, {{3, 4}}}

	if Matrix3Equal(a, b) {
		t.Error("Different length matrices should return false")
	}
}

// DCT tests
func TestInverseDCT2D(t *testing.T) {
	// Simple 8x8 DCT test
	src := MakeMatrix2D[float32](8, 8)
	dest := MakeMatrix2D[float32](8, 8)
	scratch0 := MakeMatrix2D[float32](8, 8)
	scratch1 := MakeMatrix2D[float32](8, 8)

	// Set DC component only
	src[0][0] = 1.0

	err := InverseDCT2D(src, dest, ZERO, ZERO, Dimension{Width: 8, Height: 8}, scratch0, scratch1, false)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// This implementation produces uniform output of 1.0 for DC=1.0 input
	// (non-normalized DCT, consistent with forward/inverse round-trip)
	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			if math.Abs(float64(dest[y][x]-1.0)) > 0.0001 {
				t.Errorf("dest[%d][%d] = %f; want ~1.0", y, x, dest[y][x])
			}
		}
	}
}

func TestInverseDCT2DTransposed(t *testing.T) {
	src := MakeMatrix2D[float32](8, 8)
	dest := MakeMatrix2D[float32](8, 8)
	scratch0 := MakeMatrix2D[float32](8, 8)
	scratch1 := MakeMatrix2D[float32](8, 8)

	src[0][0] = 1.0

	err := InverseDCT2D(src, dest, ZERO, ZERO, Dimension{Width: 8, Height: 8}, scratch0, scratch1, true)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Transposed mode is used for specific transform ordering
	// Just verify no error and output is populated
	hasNonZero := false
	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			if dest[y][x] != 0 {
				hasNonZero = true
				break
			}
		}
	}
	if !hasNonZero {
		t.Error("Expected non-zero output from inverse DCT")
	}
}

func TestForwardDCT2D(t *testing.T) {
	src := MakeMatrix2D[float32](8, 8)
	dest := MakeMatrix2D[float32](8, 8)
	scratch0 := MakeMatrix2D[float32](8, 8)
	scratch1 := MakeMatrix2D[float32](8, 8)

	// Uniform input
	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			src[y][x] = 1.0
		}
	}

	err := ForwardDCT2D(src, dest, ZERO, ZERO, Dimension{Width: 8, Height: 8}, scratch0, scratch1, false)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// DC component should be 1.0, all AC components should be 0
	if math.Abs(float64(dest[0][0]-1.0)) > 0.0001 {
		t.Errorf("DC component = %f; want 1.0", dest[0][0])
	}

	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			if y == 0 && x == 0 {
				continue
			}
			if math.Abs(float64(dest[y][x])) > 0.0001 {
				t.Errorf("AC component[%d][%d] = %f; want 0", y, x, dest[y][x])
			}
		}
	}
}

func TestForwardInverseDCTRoundTrip(t *testing.T) {
	original := MakeMatrix2D[float32](8, 8)
	freq := MakeMatrix2D[float32](8, 8)
	reconstructed := MakeMatrix2D[float32](8, 8)
	scratch0 := MakeMatrix2D[float32](8, 8)
	scratch1 := MakeMatrix2D[float32](8, 8)

	// Create a simple pattern
	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			original[y][x] = float32(x + y)
		}
	}

	// Forward DCT
	err := ForwardDCT2D(original, freq, ZERO, ZERO, Dimension{Width: 8, Height: 8}, scratch0, scratch1, false)
	if err != nil {
		t.Fatalf("Forward DCT error: %v", err)
	}

	// Inverse DCT
	err = InverseDCT2D(freq, reconstructed, ZERO, ZERO, Dimension{Width: 8, Height: 8}, scratch0, scratch1, false)
	if err != nil {
		t.Fatalf("Inverse DCT error: %v", err)
	}

	// Compare
	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			if math.Abs(float64(reconstructed[y][x]-original[y][x])) > 0.001 {
				t.Errorf("Mismatch at [%d][%d]: got %f, want %f", y, x, reconstructed[y][x], original[y][x])
			}
		}
	}
}

// MirrorCoordinate tests
func TestMirrorCoordinate(t *testing.T) {
	tests := []struct {
		coord    int32
		size     int32
		expected int32
	}{
		{0, 10, 0},
		{5, 10, 5},
		{9, 10, 9},
		{10, 10, 9},  // beyond size, mirrors back
		{-1, 10, 0},  // negative, mirrors
		{-2, 10, 1},
		{11, 10, 8},
	}

	for _, tt := range tests {
		result := MirrorCoordinate(tt.coord, tt.size)
		if result != tt.expected {
			t.Errorf("MirrorCoordinate(%d, %d) = %d; want %d", tt.coord, tt.size, result, tt.expected)
		}
	}
}

// CeilDiv tests
func TestCeilDiv(t *testing.T) {
	tests := []struct {
		num, denom uint32
		expected   uint32
	}{
		{10, 3, 4},
		{9, 3, 3},
		{1, 1, 1},
		{0, 1, 0},
		{7, 4, 2},
	}

	for _, tt := range tests {
		result := CeilDiv(tt.num, tt.denom)
		if result != tt.expected {
			t.Errorf("CeilDiv(%d, %d) = %d; want %d", tt.num, tt.denom, result, tt.expected)
		}
	}
}

// Benchmarks
func BenchmarkMatrixMatrixMultiply3x3(b *testing.B) {
	left := [][]float32{{1, 2, 3}, {4, 5, 6}, {7, 8, 9}}
	right := [][]float32{{9, 8, 7}, {6, 5, 4}, {3, 2, 1}}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		MatrixMatrixMultiply(left, right)
	}
}

func BenchmarkMatrixMatrixMultiply8x8(b *testing.B) {
	left := MakeMatrix2D[float32](8, 8)
	right := MakeMatrix2D[float32](8, 8)
	for i := 0; i < 8; i++ {
		for j := 0; j < 8; j++ {
			left[i][j] = float32(i * j)
			right[i][j] = float32(i + j)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		MatrixMatrixMultiply(left, right)
	}
}

func BenchmarkInvertMatrix3x3(b *testing.B) {
	matrix := [][]float32{{1, 2, 3}, {0, 1, 4}, {5, 6, 0}}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		InvertMatrix3x3(matrix)
	}
}

func BenchmarkInverseDCT8x8(b *testing.B) {
	src := MakeMatrix2D[float32](8, 8)
	dest := MakeMatrix2D[float32](8, 8)
	scratch0 := MakeMatrix2D[float32](8, 8)
	scratch1 := MakeMatrix2D[float32](8, 8)
	src[0][0] = 1.0

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		InverseDCT2D(src, dest, ZERO, ZERO, Dimension{Width: 8, Height: 8}, scratch0, scratch1, false)
	}
}

func BenchmarkForwardDCT8x8(b *testing.B) {
	src := MakeMatrix2D[float32](8, 8)
	dest := MakeMatrix2D[float32](8, 8)
	scratch0 := MakeMatrix2D[float32](8, 8)
	scratch1 := MakeMatrix2D[float32](8, 8)
	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			src[y][x] = float32(x + y)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ForwardDCT2D(src, dest, ZERO, ZERO, Dimension{Width: 8, Height: 8}, scratch0, scratch1, false)
	}
}
