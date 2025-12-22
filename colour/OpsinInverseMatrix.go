package colour

import (
	"errors"
	"slices"

	"github.com/kpfaulkner/jxl-go/jxlio"
	"github.com/kpfaulkner/jxl-go/util"
)

var (
	DEFAULT_MATRIX = [][]float32{
		{11.031566901960783, -9.866943921568629, -0.16462299647058826},
		{-3.254147380392157, 4.418770392156863, -0.16462299647058826},
		{-3.6588512862745097, 2.7129230470588235, 1.9459282392156863}}
	DEFAULT_OPSIN_BIAS              = []float32{-0.0037930732552754493, -0.0037930732552754493, -0.0037930732552754493}
	DEFAULT_QUANT_BIAS              = []float32{1.0 - 0.05465007330715401, 1.0 - 0.07005449891748593, 1.0 - 0.049935103337343655}
	DEFAULT_QBIAS_NUMERATOR float32 = 0.145
)

type OpsinInverseMatrix struct {
	Matrix             [][]float32
	OpsinBias          []float32
	QuantBias          []float32
	CbrtOpsinBias      []float32
	Primaries          CIEPrimaries
	WhitePoint         CIEXY
	QuantBiasNumerator float32
}

func NewOpsinInverseMatrix() *OpsinInverseMatrix {
	return NewOpsinInverseMatrixAllParams(*CM_PRI_SRGB, *CM_WP_D65, DEFAULT_MATRIX, DEFAULT_OPSIN_BIAS, DEFAULT_QUANT_BIAS, DEFAULT_QBIAS_NUMERATOR)
}

func NewOpsinInverseMatrixAllParams(
	primaries CIEPrimaries,
	whitePoint CIEXY,
	matrix [][]float32,
	opsinBias []float32,
	quantBias []float32,
	quantBiasNumerator float32) *OpsinInverseMatrix {

	oim := &OpsinInverseMatrix{}
	oim.Matrix = matrix
	oim.OpsinBias = opsinBias
	oim.QuantBias = quantBias
	oim.QuantBiasNumerator = quantBiasNumerator
	oim.Primaries = primaries
	oim.WhitePoint = whitePoint
	oim.bakeCbrtBias()
	return oim
}

func NewOpsinInverseMatrixWithReader(reader jxlio.BitReader) (*OpsinInverseMatrix, error) {
	oim := &OpsinInverseMatrix{}
	var err error
	var useMatrix bool
	if useMatrix, err = reader.ReadBool(); err != nil {
		return nil, err
	}
	if useMatrix {
		oim.Matrix = DEFAULT_MATRIX
		oim.OpsinBias = DEFAULT_OPSIN_BIAS
		oim.QuantBias = DEFAULT_QUANT_BIAS
		oim.QuantBiasNumerator = DEFAULT_QBIAS_NUMERATOR
	} else {
		oim.Matrix = util.MakeMatrix2D[float32](3, 3)
		for i := 0; i < 3; i++ {
			for j := 0; j < 3; j++ {
				if oim.Matrix[i][j], err = reader.ReadF16(); err != nil {
					return nil, err
				}
			}
		}
		oim.OpsinBias = make([]float32, 3)
		for i := 0; i < 3; i++ {
			if oim.OpsinBias[i], err = reader.ReadF16(); err != nil {
				return nil, err
			}
		}
		oim.QuantBias = make([]float32, 3)
		for i := 0; i < 3; i++ {
			if oim.QuantBias[i], err = reader.ReadF16(); err != nil {
				return nil, err
			}
		}
		if oim.QuantBiasNumerator, err = reader.ReadF16(); err != nil {
			return nil, err
		}
	}
	oim.Primaries = *CM_PRI_SRGB
	oim.WhitePoint = *CM_WP_D65
	oim.bakeCbrtBias()

	return oim, nil
}

func (oim *OpsinInverseMatrix) bakeCbrtBias() {
	oim.CbrtOpsinBias = make([]float32, 3)
	for c := 0; c < 3; c++ {
		oim.CbrtOpsinBias[c] = util.SignedPow(oim.OpsinBias[c], 1.0/3.0)
	}
}

func (oim *OpsinInverseMatrix) GetMatrix(prim *CIEPrimaries, white *CIEXY) (*OpsinInverseMatrix, error) {
	conversion, err := GetConversionMatrix(*prim, *white, oim.Primaries, oim.WhitePoint)
	if err != nil {
		return nil, err
	}
	matrix, err := util.MatrixMultiply(conversion, oim.Matrix)
	if err != nil {
		return nil, err
	}

	return NewOpsinInverseMatrixAllParams(*prim, *white, matrix, oim.OpsinBias, oim.QuantBias, oim.QuantBiasNumerator), nil
}

func (oim *OpsinInverseMatrix) InvertXYB(buffer [][][]float32, intensityTarget float32) error {

	if len(buffer) < 3 {
		return errors.New("Can only XYB on 3 channels")
	}

	// Cache struct fields locally to avoid repeated lookups in hot loop
	itScale := 255.0 / intensityTarget
	cbrtBias0 := oim.CbrtOpsinBias[0]
	cbrtBias1 := oim.CbrtOpsinBias[1]
	cbrtBias2 := oim.CbrtOpsinBias[2]
	opsinBias0 := oim.OpsinBias[0]
	opsinBias1 := oim.OpsinBias[1]
	opsinBias2 := oim.OpsinBias[2]

	// Cache matrix coefficients (unrolled for 3 channels)
	m00 := oim.Matrix[0][0]
	m01 := oim.Matrix[0][1]
	m02 := oim.Matrix[0][2]
	m10 := oim.Matrix[1][0]
	m11 := oim.Matrix[1][1]
	m12 := oim.Matrix[1][2]
	m20 := oim.Matrix[2][0]
	m21 := oim.Matrix[2][1]
	m22 := oim.Matrix[2][2]

	buf0 := buffer[0]
	buf1 := buffer[1]
	buf2 := buffer[2]

	for y := 0; y < len(buf0); y++ {
		// Cache row pointers to avoid repeated slice lookups in inner loop
		row0 := buf0[y]
		row1 := buf1[y]
		row2 := buf2[y]
		rowLen := len(row0)

		for x := 0; x < rowLen; x++ {
			b0 := row0[x]
			b1 := row1[x]
			b2 := row2[x]

			gammaL := b1 + b0 - cbrtBias0
			gammaM := b1 - b0 - cbrtBias1
			gammaS := b2 - cbrtBias2

			mixL := gammaL*gammaL*gammaL + opsinBias0
			mixM := gammaM*gammaM*gammaM + opsinBias1
			mixS := gammaS*gammaS*gammaS + opsinBias2

			// Unrolled loop for 3 channels
			row0[x] = (mixL*m00 + mixM*m01 + mixS*m02) * itScale
			row1[x] = (mixL*m10 + mixM*m11 + mixS*m12) * itScale
			row2[x] = (mixL*m20 + mixM*m21 + mixS*m22) * itScale
		}
	}
	return nil
}

// Matches determines if values are equal. Simplistic but will do for now.
func (oim *OpsinInverseMatrix) Matches(other OpsinInverseMatrix) bool {

	if !util.CompareMatrix2D(oim.Matrix, other.Matrix, func(a float32, b float32) bool { return a == b }) {
		return false
	}

	if slices.Compare(oim.OpsinBias, other.OpsinBias) != 0 {
		return false
	}

	if slices.Compare(oim.QuantBias, other.QuantBias) != 0 {
		return false
	}

	if oim.QuantBiasNumerator != other.QuantBiasNumerator {
		return false
	}

	if !oim.Primaries.Red.Matches(other.Primaries.Red) {
		return false
	}

	if !oim.Primaries.Green.Matches(other.Primaries.Green) {
		return false
	}

	if !oim.Primaries.Blue.Matches(other.Primaries.Blue) {
		return false
	}

	if !oim.WhitePoint.Matches(&other.WhitePoint) {
		return false
	}

	return true
}
