package frame

import (
	"testing"
)

func copyArray(a [][]float64) [][]float64 {
	b := make([][]float64, len(a))
	for i := range a {
		b[i] = make([]float64, len(a[i]))
		copy(b[i], a[i])
	}
	return b
}

func generateTestDCTParams(fn func(dct DCTParam) DCTParam) DCTParam {
	d := NewDCTParam()
	d.dctParam = copyArray(dct4x8params)
	d.param = [][]float32{
		{3072, 3072, 256, 256, 256, 414, 0.0, 0.0, 0.0},
		{1024, 1024, 50.0, 50.0, 50.0, 58, 0.0, 0.0, 0.0},
		{384, 384, 12.0, 12.0, 12.0, 22, -0.25, -0.25, -0.25}}
	d.mode = MODE_AFV
	d.denominator = 1
	d.params4x4 = copyArray(dct4x4params)

	d2 := fn(*d)
	return d2
}

func TestDctParamsEquals(t *testing.T) {

	for _, tc := range []struct {
		name          string
		other         DCTParam
		expectSuccess bool
	}{
		{
			name:          "success",
			other:         generateTestDCTParams(func(dct DCTParam) DCTParam { return dct }),
			expectSuccess: true,
		},
		{
			name: "fail different mode",
			other: generateTestDCTParams(func(dct DCTParam) DCTParam {
				dct.mode = MODE_DCT
				return dct
			}),
			expectSuccess: false,
		},
		{
			name: "fail different dctParams",
			other: generateTestDCTParams(func(dct DCTParam) DCTParam {
				dct.dctParam[0][0] = 123
				return dct
			}),
			expectSuccess: false,
		},
		{
			name: "fail different param",
			other: generateTestDCTParams(func(dct DCTParam) DCTParam {
				dct.param[0][0] = 123
				return dct
			}),
			expectSuccess: false,
		},
		{
			name: "fail different denominator",
			other: generateTestDCTParams(func(dct DCTParam) DCTParam {
				dct.denominator = 2
				return dct
			}),
			expectSuccess: false,
		},
		{
			name: "fail different params4x4",
			other: generateTestDCTParams(func(dct DCTParam) DCTParam {
				dct.params4x4[0][0] = 123
				return dct
			}),
			expectSuccess: false,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {

			orig := generateTestDCTParams(func(dct DCTParam) DCTParam { return dct })

			if orig.Equals(tc.other) != tc.expectSuccess {
				if !tc.expectSuccess {
					t.Errorf("expected Equals to return %v", tc.expectSuccess)
				}
			}
		})
	}
}
