package frame

import (
	"errors"
	"fmt"
	"math"

	"github.com/kpfaulkner/jxl-go/util"
)

const (
	MODE_LIBRARY = 0
	MODE_HORNUSS = 1
	MODE_DCT2    = 2
	MODE_DCT4    = 3
	MODE_DCT4_8  = 4
	MODE_AFV     = 5
	MODE_DCT     = 6
	MODE_RAW     = 7

	METHOD_DCT     = 0
	METHOD_DCT2    = 1
	METHOD_DCT4    = 2
	METHOD_HORNUSS = 3
	METHOD_DCT8_4  = 4
	METHOD_DCT4_8  = 5
	METHOD_AFV     = 6
)

type TransformType struct {
	parameterIndex  int32
	dctSelectWidth  int32
	dctSelectHeight int32
	name            string
	ttType          int32
	blockWidth      int32
	blockHeight     int32
	matrixWidth     int32
	matrixHeight    int32
	orderID         int32
	transformMethod int32
	llfScale        [][]float32
}

var (
	DCT8       = NewTransformType("DCT 8x8", 0, 0, 1, 1, 8, 8, 8, 8, 0, 0)
	HORNUSS    = NewTransformType("Hornuss", 1, 1, 1, 1, 8, 8, 8, 8, 1, 3)
	DCT2       = NewTransformType("DCT 2x2", 2, 2, 1, 1, 8, 8, 8, 8, 1, 1)
	DCT4       = NewTransformType("DCT 4x4", 3, 3, 1, 1, 8, 8, 8, 8, 1, 2)
	DCT16      = NewTransformType("DCT 16x16", 4, 4, 2, 2, 16, 16, 16, 16, 2, 0)
	DCT32      = NewTransformType("DCT 32x32", 5, 5, 4, 4, 32, 32, 32, 32, 3, 0)
	DCT16_8    = NewTransformType("DCT 16x8", 6, 6, 2, 1, 16, 8, 8, 16, 4, 0)
	DCT8_16    = NewTransformType("DCT 8x16", 7, 6, 1, 2, 8, 16, 8, 16, 4, 0)
	DCT32_8    = NewTransformType("DCT 32x8", 8, 7, 4, 1, 32, 8, 8, 32, 5, 0)
	DCT8_32    = NewTransformType("DCT 8x32", 9, 7, 1, 4, 8, 32, 8, 32, 5, 0)
	DCT32_16   = NewTransformType("DCT 32x16", 10, 8, 4, 2, 32, 16, 16, 32, 6, 0)
	DCT16_32   = NewTransformType("DCT 16x32", 11, 8, 2, 4, 16, 32, 16, 32, 6, 0)
	DCT4_8     = NewTransformType("DCT 4x8", 12, 9, 1, 1, 8, 8, 8, 8, 1, 5)
	DCT8_4     = NewTransformType("DCT 8x4", 13, 9, 1, 1, 8, 8, 8, 8, 1, 4)
	AFV0       = NewTransformType("AFV0", 14, 10, 1, 1, 8, 8, 8, 8, 1, 6)
	AFV1       = NewTransformType("AFV1", 15, 10, 1, 1, 8, 8, 8, 8, 1, 6)
	AFV2       = NewTransformType("AFV2", 16, 10, 1, 1, 8, 8, 8, 8, 1, 6)
	AFV3       = NewTransformType("AFV3", 17, 10, 1, 1, 8, 8, 8, 8, 1, 6)
	DCT64      = NewTransformType("DCT 64x64", 18, 11, 8, 8, 64, 64, 64, 64, 7, 0)
	DCT64_32   = NewTransformType("DCT 64x32", 19, 12, 8, 4, 64, 32, 32, 64, 8, 0)
	DCT32_64   = NewTransformType("DCT 32x64", 20, 12, 4, 8, 32, 64, 32, 64, 8, 0)
	DCT128     = NewTransformType("DCT 128x128", 21, 13, 16, 16, 128, 128, 128, 128, 9, 0)
	DCT128_64  = NewTransformType("DCT 128x64", 22, 14, 16, 8, 128, 64, 64, 128, 10, 0)
	DCT64_128  = NewTransformType("DCT 64x128", 23, 14, 8, 16, 64, 128, 64, 128, 10, 0)
	DCT256     = NewTransformType("DCT 256x256", 24, 15, 32, 32, 256, 256, 256, 256, 11, 0)
	DCT256_128 = NewTransformType("DCT 256x128", 25, 16, 32, 16, 256, 128, 128, 256, 12, 0)
	DCT128_256 = NewTransformType("DCT 128x256", 26, 16, 16, 32, 128, 256, 128, 256, 12, 0)

	allDCT = []TransformType{*DCT8, *HORNUSS, *DCT2, *DCT4, *DCT16, *DCT32, *DCT16_8, *DCT8_16, *DCT32_8, *DCT8_32, *DCT32_16, *DCT16_32, *DCT4_8, *DCT8_4, *AFV0, *AFV1, *AFV2, *AFV3, *DCT64, *DCT64_32, *DCT32_64, *DCT128, *DCT128_64, *DCT64_128, *DCT256, *DCT256_128, *DCT128_256}
)

func NewTransformType(name string, transType int32, parameterIndex int32, dctSelectHeight int32, dctSelectWidth int32,
	blockHeight int32, blockWidth int32, matrixHeight int32, matrixWidth int32, orderID int32, transformMethod int32) *TransformType {

	tt := &TransformType{
		parameterIndex:  parameterIndex,
		dctSelectWidth:  dctSelectWidth,
		dctSelectHeight: dctSelectHeight,
		name:            name,
		ttType:          transType,
		blockWidth:      blockWidth,
		blockHeight:     blockHeight,
		matrixWidth:     matrixWidth,
		matrixHeight:    matrixHeight,
		orderID:         orderID,
		transformMethod: transformMethod,
		llfScale:        util.MakeMatrix2D[float32](dctSelectHeight, dctSelectWidth),
	}
	for y := int32(0); y < dctSelectHeight; y++ {
		for x := int32(0); x < dctSelectWidth; x++ {
			tt.llfScale[y][x] = float32(scaleF(float64(y), float64(dctSelectHeight)) * scaleF(float64(x), float64(dctSelectWidth)))
		}
	}

	return tt
}

func (tt TransformType) isVertical() bool {
	return tt.blockHeight > tt.blockWidth
}

func (tt TransformType) flip() bool {
	return tt.blockHeight > tt.blockWidth || tt.transformMethod == METHOD_DCT && tt.blockHeight == tt.blockWidth
}

func (tt TransformType) getPixelSize() util.Dimension {
	return util.Dimension{
		Width:  uint32(tt.blockWidth),
		Height: uint32(tt.blockHeight),
	}

}

func scaleF(c float64, b float64) float64 {
	piSize := math.Pi * c
	return (1.0 / math.Cos(piSize/(2*b)) * math.Cos(piSize/b) * math.Cos(2.0*piSize/(2.0*b)))
}

func validateIndex(index int32, mode int32) (bool, error) {
	if mode < 0 || mode > 7 {
		return false, errors.New("Invalid mode")
	}
	if mode == MODE_LIBRARY || mode == MODE_DCT || mode == MODE_RAW {
		return true, nil
	}

	if index >= 0 && index <= 3 || index == 9 || index == 10 {
		return true, nil
	}

	return false, errors.New(fmt.Sprintf("Invalid index %d for mode %d", index, mode))

}

func getHorizontalTransformType(index int32) (*TransformType, error) {

	for _, tt := range allDCT {
		if tt.parameterIndex == index && !tt.isVertical() {
			return &tt, nil
		}
	}

	return nil, errors.New("Unable to find horizontal transform type")
}

func getByOrderID(orderID int32) (*TransformType, error) {
	for _, tt := range allDCT {
		if tt.orderID == orderID {
			return &tt, nil
		}
	}
	return nil, errors.New("Unable to find transform type for orderID")
}
