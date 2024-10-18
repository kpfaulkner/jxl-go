package color

import (
	"errors"

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
	matrix             [][]float32
	opsinBias          []float32
	QuantBias          []float32
	QuantBiasNumerator float32
	primaries          CIEPrimaries
	whitePoint         CIEXY
	cbrtOpsinBias      []float32
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
	oim.matrix = matrix
	oim.opsinBias = opsinBias
	oim.QuantBias = quantBias
	oim.QuantBiasNumerator = quantBiasNumerator
	oim.primaries = primaries
	oim.whitePoint = whitePoint
	oim.bakeCbrtBias()
	return oim
}

func NewOpsinInverseMatrixWithReader(reader *jxlio.Bitreader) *OpsinInverseMatrix {
	oim := &OpsinInverseMatrix{}

	if reader.MustReadBool() {
		oim.matrix = DEFAULT_MATRIX
		oim.opsinBias = DEFAULT_OPSIN_BIAS
		oim.QuantBias = DEFAULT_QUANT_BIAS
		oim.QuantBiasNumerator = DEFAULT_QBIAS_NUMERATOR
	} else {
		oim.matrix = util.MakeMatrix2D[float32](3, 3)
		for i := 0; i < 3; i++ {
			for j := 0; j < 3; j++ {
				oim.matrix[i][j] = reader.MustReadF16()
			}
		}
		oim.opsinBias = make([]float32, 3)
		for i := 0; i < 3; i++ {
			oim.opsinBias[i] = reader.MustReadF16()
		}
		oim.QuantBias = make([]float32, 3)
		for i := 0; i < 3; i++ {
			oim.QuantBias[i] = reader.MustReadF16()
		}
		oim.QuantBiasNumerator = reader.MustReadF16()
	}
	oim.primaries = *CM_PRI_SRGB
	oim.whitePoint = *CM_WP_D65
	oim.bakeCbrtBias()

	return oim
}

func (oim *OpsinInverseMatrix) bakeCbrtBias() {
	oim.cbrtOpsinBias = make([]float32, 3)
	for c := 0; c < 3; c++ {
		oim.cbrtOpsinBias[c] = util.SignedPow(oim.opsinBias[c], 1.0/3.0)
	}
}

func (oim *OpsinInverseMatrix) GetMatrix(prim *CIEPrimaries, white *CIEXY) (*OpsinInverseMatrix, error) {
	conversion, err := GetConversionMatrix(*prim, *white, oim.primaries, oim.whitePoint)
	if err != nil {
		return nil, err
	}
	matrix, err := util.MatrixMultiply(conversion, oim.matrix)
	if err != nil {
		return nil, err
	}

	return NewOpsinInverseMatrixAllParams(*prim, *white, matrix, oim.opsinBias, oim.QuantBias, oim.QuantBiasNumerator), nil
}

func (oim *OpsinInverseMatrix) InvertXYB(buffer [][][]float32, intensityTarget float32) error {

	if len(buffer) < 3 {
		return errors.New("Can only XYB on 3 channels")
	}
	itScale := 255.0 / intensityTarget
	for y := 0; y < len(buffer[0]); y++ {
		for x := 0; x < len(buffer[0][y]); x++ {
			gammaL := buffer[1][y][x] + buffer[0][y][x] - oim.cbrtOpsinBias[0]
			gammaM := buffer[1][y][x] - buffer[0][y][x] - oim.cbrtOpsinBias[1]
			gammaS := buffer[2][y][x] - oim.cbrtOpsinBias[2]
			mixL := gammaL*gammaL*gammaL + oim.opsinBias[0]
			mixM := gammaM*gammaM*gammaM + oim.opsinBias[1]
			mixS := gammaS*gammaS*gammaS + oim.opsinBias[2]
			for c := 0; c < 3; c++ {
				buffer[c][y][x] = (mixL*oim.matrix[c][0] + mixM*oim.matrix[c][1] + mixS*oim.matrix[c][2]) * itScale
			}
		}
	}
	return nil
}
