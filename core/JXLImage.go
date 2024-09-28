package core

import (
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
