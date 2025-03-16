package frame

import "github.com/kpfaulkner/jxl-go/util"

type DCTParam struct {
	dctParam    *util.Matrix[float64]
	param       *util.Matrix[float32]
	mode        int32
	denominator float32
	params4x4   *util.Matrix[float64]
}

// Equals compares 2 DCTParams structs. Slightly concerned about having to
// compare all the multi-dimensional slices. Investigated slices.EqualFunc, Equal, compare etc.
// but unsure about getting those working for multi-dimensional. So will just do naively for now
// and measure later.
// TODO(kpfaulkner) do some measuring around performance here.
func (dct DCTParam) Equals(other DCTParam) bool {
	if dct.mode != other.mode {
		return false
	}
	if dct.denominator != other.denominator {
		return false
	}
	if dct.dctParam.Height != other.dctParam.Height {
		return false
	}
	if dct.param.Height != other.param.Height {
		return false
	}

	// FIXME(kpfaulkner) not keen on == for float64...  need to double check
	if !util.CompareMatrix2D(dct.dctParam, other.dctParam, func(a float64, b float64) bool {
		return a == b
	}) {
		return false
	}

	if !util.CompareMatrix2D(dct.param, other.param, func(a float32, b float32) bool {
		return a == b
	}) {
		return false
	}

	if !util.CompareMatrix2D(dct.params4x4, other.params4x4, func(a float64, b float64) bool {
		return a == b
	}) {
		return false
	}
	return true
}

func NewDCTParam() *DCTParam {
	return &DCTParam{
		dctParam:    util.New2DMatrix[float64](0, 0),
		param:       util.New2DMatrix[float32](0, 0),
		params4x4:   util.New2DMatrix[float64](0, 0),
		mode:        0,
		denominator: 0,
	}
}
