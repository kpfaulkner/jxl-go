package core

import (
	"fmt"
	"image"

	"github.com/kpfaulkner/jxl-go/color"
	"github.com/kpfaulkner/jxl-go/util"
)

// JXLImage contains the core information about the JXL image.
type JXLImage struct {
	Width                int
	Height               int
	Buffer               [][]float32
	ColorEncoding        int32
	alphaIndex           int32
	imageHeader          ImageHeader
	primaries            int32
	whitePoint           int32
	primariesXY          *color.CIEPrimaries
	whiteXY              *color.CIEXY
	transfer             int32
	iccProfile           []byte
	taggedTransfer       int32
	alphaIsPremultiplied bool
}

func NewJXLImageWithBuffer(buffer [][][]float32, header ImageHeader) (*JXLImage, error) {
	jxl := &JXLImage{}
	jxl.Width = len(buffer[0][0])
	jxl.Height = len(buffer[0])

	channels := header.getTotalChannelCount()
	jxl.Buffer = util.MakeMatrix2D[float32](channels, jxl.Width*jxl.Height)
	for c := 0; c < channels; c++ {
		for y := 0; y < jxl.Height; y++ {
			copy(jxl.Buffer[c][y*jxl.Width:], buffer[c][y])
		}
	}

	bundle := header.colorEncoding
	jxl.ColorEncoding = bundle.ColorEncoding
	if header.hasAlpha() {
		jxl.alphaIndex = header.alphaIndices[0]
	} else {
		jxl.alphaIndex = -1
	}
	jxl.imageHeader = header
	jxl.primaries = bundle.Primaries
	jxl.whitePoint = bundle.WhitePoint
	jxl.primariesXY = bundle.Prim
	jxl.whiteXY = bundle.White

	if jxl.imageHeader.xybEncoded {
		jxl.transfer = color.TF_LINEAR
		jxl.iccProfile = nil
	} else {
		jxl.transfer = bundle.Tf
		jxl.iccProfile = header.getDecodedICC()
	}
	jxl.taggedTransfer = bundle.Tf
	jxl.alphaIsPremultiplied = jxl.imageHeader.hasAlpha() && jxl.imageHeader.extraChannelInfo[jxl.alphaIndex].alphaAssociated
	return jxl, nil
}

// NewImage generates a standard go image.Image from the JXL image.
func NewImage(buffer [][][]float32, header ImageHeader) (image.Image, error) {

	// default to NRGBA for now. Will need to detect properly later.
	// TODO(kpfaulkner) get right image type
	jxl := image.NewRGBA(image.Rect(0, 0, len(buffer[0][0]), len(buffer[0])))

	//channels := header.getTotalChannelCount()
	//jxl.Buffer = util.MakeMatrix2D[float32](channels, jxl.Width*jxl.Height)
	//for c := 0; c < channels; c++ {
	//	for y := 0; y < jxl.Height; y++ {
	//		copy(jxl.Buffer[c][y*jxl.Width:], buffer[c][y])
	//	}
	//}

	pix := jxl.Pix
	dx := jxl.Bounds().Dx()
	dy := jxl.Bounds().Dy()
	pos := 0
	//for c := 0; c < channels; c++ {
	for y := 0; y < dy; y++ {

		if y == 612 {
			fmt.Printf("snoop\n")
		}
		for x := 0; x < dx; x++ {
			//pixPos := 4*(y*jxl.Stride+x) + c
			pix[pos] = uint8(buffer[0][y][x] * 255)
			pos++
			pix[pos] = uint8(buffer[1][y][x] * 255)
			pos++

			pix[pos] = uint8(buffer[2][y][x] * 255)
			pos++
			pix[pos] = 255
			pos++
			//pos++
			//val := uint8(buffer[c][y][x] * 255)
			//pix[pixPos] = val
			//jxl.Set(x, y, buffer[c][y][x])
		}
	}
	//}

	//bundle := header.colorEncoding
	//jxl.ColorEncoding = bundle.ColorEncoding
	//if header.hasAlpha() {
	//	jxl.alphaIndex = header.alphaIndices[0]
	//} else {
	//	jxl.alphaIndex = -1
	//}
	//jxl.imageHeader = header
	//jxl.primaries = bundle.Primaries
	//jxl.whitePoint = bundle.WhitePoint
	//jxl.primariesXY = bundle.Prim
	//jxl.whiteXY = bundle.White
	//
	//if jxl.imageHeader.xybEncoded {
	//	jxl.transfer = color.TF_LINEAR
	//	jxl.iccProfile = nil
	//} else {
	//	jxl.transfer = bundle.Tf
	//	jxl.iccProfile = header.getDecodedICC()
	//}
	//jxl.taggedTransfer = bundle.Tf
	//jxl.alphaIsPremultiplied = jxl.imageHeader.hasAlpha() && jxl.imageHeader.extraChannelInfo[jxl.alphaIndex].alphaAssociated
	//return jxl, nil

	return jxl, nil
}
