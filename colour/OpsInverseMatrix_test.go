package colour

import (
	"fmt"
	"math"
	"testing"

	"github.com/kpfaulkner/jxl-go/testcommon"
	"github.com/kpfaulkner/jxl-go/util"
)

func TestNewOpsinInverseMatrix(t *testing.T) {

	for _, tc := range []struct {
		name           string
		expectedResult *OpsinInverseMatrix
	}{
		{
			name: "success",
			expectedResult: &OpsinInverseMatrix{
				Matrix:        [][]float32{[]float32{11.031567, -9.866944, -0.16462299}, []float32{-3.2541473, 4.4187703, -0.16462299}, []float32{-3.6588514, 2.712923, 1.9459282}},
				OpsinBias:     []float32{-0.0037930734, -0.0037930734, -0.0037930734},
				QuantBias:     []float32{0.94534993, 0.9299455, 0.9500649},
				CbrtOpsinBias: []float32{-0.1559542, -0.1559542, -0.1559542},
				Primaries: CIEPrimaries{
					Red:   &CIEXY{X: 0.6399987, Y: 0.33001015},
					Green: &CIEXY{X: 0.3000038, Y: 0.60000336},
					Blue:  &CIEXY{X: 0.15000205, Y: 0.059997205},
				},
				WhitePoint:         CIEXY{X: 0.3127, Y: 0.329},
				QuantBiasNumerator: 0.145,
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {

			matrix := NewOpsinInverseMatrix()

			if !matrix.Matches(*tc.expectedResult) {
				t.Errorf("NewOpsinInverseMatrix() expected %v, got %v", *tc.expectedResult, matrix)
			}
		})
	}
}

func TestNewOpsinInverseMatrixWithReader(t *testing.T) {

	for _, tc := range []struct {
		name           string
		expectedResult *OpsinInverseMatrix
		readBoolData   []bool
		readF16Data    []float32
		expectErr      bool
	}{
		{
			name:         "Use default",
			readBoolData: []bool{true},
			expectErr:    false,
			expectedResult: &OpsinInverseMatrix{
				Matrix:        [][]float32{[]float32{11.031567, -9.866944, -0.16462299}, []float32{-3.2541473, 4.4187703, -0.16462299}, []float32{-3.6588514, 2.712923, 1.9459282}},
				OpsinBias:     []float32{-0.0037930734, -0.0037930734, -0.0037930734},
				QuantBias:     []float32{0.94534993, 0.9299455, 0.9500649},
				CbrtOpsinBias: []float32{-0.1559542, -0.1559542, -0.1559542},
				Primaries: CIEPrimaries{
					Red:   &CIEXY{X: 0.6399987, Y: 0.33001015},
					Green: &CIEXY{X: 0.3000038, Y: 0.60000336},
					Blue:  &CIEXY{X: 0.15000205, Y: 0.059997205},
				},
				WhitePoint:         CIEXY{X: 0.3127, Y: 0.329},
				QuantBiasNumerator: 0.145,
			},
		},
		{
			name:         "success",
			readBoolData: []bool{false},
			readF16Data:  []float32{1, 2, 3, 4, 5, 6, 7, 8, 9, 11, 12, 13, 21, 22, 23, 31},
			expectErr:    false,
			expectedResult: &OpsinInverseMatrix{
				Matrix:        [][]float32{[]float32{1, 2, 3}, []float32{4, 5, 6}, []float32{7, 8, 9}},
				OpsinBias:     []float32{11, 12, 13},
				QuantBias:     []float32{21, 22, 23},
				CbrtOpsinBias: []float32{2.2239802, 2.2894285, 2.3513348},
				Primaries: CIEPrimaries{
					Red:   &CIEXY{X: 0.6399987, Y: 0.33001015},
					Green: &CIEXY{X: 0.3000038, Y: 0.60000336},
					Blue:  &CIEXY{X: 0.15000205, Y: 0.059997205},
				},
				WhitePoint:         CIEXY{X: 0.3127, Y: 0.329},
				QuantBiasNumerator: 31,
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {

			bitReader := testcommon.NewFakeBitReader()
			bitReader.ReadF16Data = tc.readF16Data
			bitReader.ReadBoolData = tc.readBoolData

			matrix, err := NewOpsinInverseMatrixWithReader(bitReader)

			fmt.Printf("XXXXXX %#v\n", *matrix)
			if tc.expectErr && err == nil {
				t.Errorf("expected error but got none")
			}

			if !tc.expectErr && err != nil {
				t.Errorf("got error when none was expected : %v", err)
			}

			if tc.expectErr && err != nil {
				return
			}

			if !matrix.Matches(*tc.expectedResult) {
				t.Errorf("NewOpsinInverseMatrix() expected %v, got %v", *tc.expectedResult, matrix)
			}
		})
	}
}

func TestInvertXYB(t *testing.T) {

	matrix := NewOpsinInverseMatrix()

	expectedResult := [][][]float32{[][]float32{[]float32{1.5267061e+06, 3.7549045e+06, 6.785943e+06}, []float32{1.0742343e+07, 1.5746626e+07, 2.1921318e+07}, []float32{2.9388938e+07, 3.827201e+07, 4.8693052e+07}}, [][]float32{[]float32{-529389.94, -1.2439688e+06, -2.20135e+06}, []float32{-3.4379725e+06, -4.9902745e+06, -6.8946945e+06}, []float32{-9.187672e+06, -1.1905647e+07, -1.5085057e+07}}, [][]float32{[]float32{2.4760028e+06, 2.2452708e+06, 1.8012834e+06}, []float32{1.1060344e+06, 121518.45, -1.1902722e+06}, []float32{-2.8673455e+06, -4.9477045e+06, -7.4693575e+06}}}

	// dummy data..
	buf := util.MakeMatrix3D[float32](3, 3, 3)
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			for k := 0; k < 3; k++ {
				buf[i][j][k] = float32(i*9 + j*3 + k + 1)
			}
		}
	}

	err := matrix.InvertXYB(buf, 1.1)

	if err != nil {
		t.Errorf("InvertXYB() unexpected error: %v", err)
	}

	if util.CompareMatrix3D[float32](expectedResult, buf, func(a, b float32) bool {

		// just check they're close enough..
		return math.Abs(float64(a-b)) < 0.00001
	}) {

	}

}
