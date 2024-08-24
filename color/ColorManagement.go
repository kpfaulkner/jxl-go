package color

import (
	"errors"

	"github.com/kpfaulkner/jxl-go/util"
)

var (
	CM_PRI_SRGB = GetPrimaries(PRI_SRGB)

	CM_WP_D65 = GetWhitePoint(WP_D65)
	CM_WP_D50 = GetWhitePoint(WP_D50)

	BRADFORD = [][]float32{
		{0.8951, 0.2664, -0.1614},
		{-0.7502, 1.7135, 0.0367},
		{0.0389, -0.0685, 1.0296},
	}

	BRADFORD_INVERSE = util.InvertMatrix3x3(BRADFORD)
)

func GetConversionMatrix(targetPrim CIEPrimaries, targetWP CIEXY, currentPrim CIEPrimaries, currentWP CIEXY) ([][]float32, error) {
	if targetPrim.Matches(&currentPrim) && targetWP.Matches(&currentWP) {
		return util.MatrixIdentity(3), nil
	}

	var whitePointConv [][]float32
	var err error
	if !targetWP.Matches(&currentWP) {
		whitePointConv, err = AdaptWhitePoint(&targetWP, &currentWP)
		if err != nil {
			return nil, err
		}
	}
	forward, err := primariesToXYZ(&currentPrim, &currentWP)
	if err != nil {
		return nil, err
	}

	t, err := primariesToXYZ(&targetPrim, &targetWP)
	if err != nil {
		return nil, err
	}
	reverse := util.InvertMatrix3x3(t)
	res, err := util.MatrixMultiply(reverse, whitePointConv, forward)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func primariesToXYZ(primaries *CIEPrimaries, wp *CIEXY) ([][]float32, error) {
	if primaries == nil {
		return nil, nil
	}

	if wp == nil {
		wp = CM_WP_D50
	}
	if wp.x < 0 || wp.x > 1 || wp.y <= 0 || wp.y > 1 {
		return nil, errors.New("invalid argument")
	}
	r, errR := GetXYZ(*primaries.red)
	g, errG := GetXYZ(*primaries.green)
	b, errB := GetXYZ(*primaries.blue)
	if errR != nil || errG != nil || errB != nil {
		return nil, errors.New("invalid argument")
	}
	primariesTr := [][]float32{r, g, b}
	primariesMatrix := util.TransposeMatrix(primariesTr, util.NewIntPoint(3))
	inversePrimaries := util.InvertMatrix3x3(primariesMatrix)
	w, err := GetXYZ(*wp)
	if err != nil {
		return nil, err
	}
	xyz, err := util.MatrixVectorMultiply(inversePrimaries, w)
	if err != nil {
		return nil, err
	}
	a := [][]float32{{xyz[0], 0, 0}, {0, xyz[1], 0}, {0, 0, xyz[2]}}
	res, err := util.MatrixMatrixMultiply(primariesMatrix, a)
	if err != nil {
		return nil, err
	}
	return res, nil

}

func validateXY(xy CIEXY) error {
	if xy.x < 0 || xy.x > 1 || xy.y <= 0 || xy.y > 1 {
		return errors.New("Invalid argument")
	}
	return nil
}

func GetXYZ(xy CIEXY) ([]float32, error) {
	if err := validateXY(xy); err != nil {
		return nil, err
	}
	invY := 1.0 / xy.y
	return []float32{xy.x * invY, 1.0, (1.0 - xy.x - xy.y) * invY}, nil
}
