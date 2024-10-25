package core

import (
	"image"

	"github.com/kpfaulkner/jxl-go/bundle"
	"github.com/kpfaulkner/jxl-go/color"
	image2 "github.com/kpfaulkner/jxl-go/image"
	"github.com/kpfaulkner/jxl-go/util"
)

// JXLImage contains the core information about the JXL image.
type JXLImage struct {
	Width                int
	Height               int
	Buffer               [][]float32
	ColorEncoding        int32
	alphaIndex           int32
	imageHeader          bundle.ImageHeader
	primaries            int32
	whitePoint           int32
	primariesXY          *color.CIEPrimaries
	whiteXY              *color.CIEXY
	transfer             int32
	iccProfile           []byte
	taggedTransfer       int32
	alphaIsPremultiplied bool
}

func NewJXLImageWithBuffer(buffer [][][]float32, header bundle.ImageHeader) (*JXLImage, error) {
	jxl := &JXLImage{}
	jxl.Width = len(buffer[0][0])
	jxl.Height = len(buffer[0])

	channels := header.GetTotalChannelCount()
	jxl.Buffer = util.MakeMatrix2D[float32](channels, jxl.Width*jxl.Height)
	for c := 0; c < channels; c++ {
		for y := 0; y < jxl.Height; y++ {
			copy(jxl.Buffer[c][y*jxl.Width:], buffer[c][y])
		}
	}

	bundle := header.ColorEncoding
	jxl.ColorEncoding = bundle.ColorEncoding
	if header.HasAlpha() {
		jxl.alphaIndex = header.AlphaIndices[0]
	} else {
		jxl.alphaIndex = -1
	}
	jxl.imageHeader = header
	jxl.primaries = bundle.Primaries
	jxl.whitePoint = bundle.WhitePoint
	jxl.primariesXY = bundle.Prim
	jxl.whiteXY = bundle.White

	if jxl.imageHeader.XybEncoded {
		jxl.transfer = color.TF_LINEAR
		jxl.iccProfile = nil
	} else {
		jxl.transfer = bundle.Tf
		jxl.iccProfile = header.GetDecodedICC()
	}
	jxl.taggedTransfer = bundle.Tf
	jxl.alphaIsPremultiplied = jxl.imageHeader.HasAlpha() && jxl.imageHeader.ExtraChannelInfo[jxl.alphaIndex].AlphaAssociated
	return jxl, nil
}

// NewImage generates a standard go image.Image from the JXL image.
func NewImage(buffer []image2.ImageBuffer, header bundle.ImageHeader) (image.Image, error) {

	// default to NRGBA for now. Will need to detect properly later.
	// TODO(kpfaulkner) get right image type
	jxl := image.NewRGBA(image.Rect(0, 0, int(buffer[0].Width), int(buffer[0].Height)))

	// case to int image..
	for c := 0; c < len(buffer); c++ {
		if buffer[c].IsInt() {
			// convert to float
			buffer[c].CastToFloatIfInt(255)
		}
	}
	pix := jxl.Pix
	dx := jxl.Bounds().Dx()
	dy := jxl.Bounds().Dy()
	pos := 0
	if buffer[0].IsFloat() {
		for y := 0; y < dy; y++ {
			for x := 0; x < dx; x++ {
				pix[pos] = uint8(buffer[0].FloatBuffer[y][x] * 255)
				pos++
				pix[pos] = uint8(buffer[1].FloatBuffer[y][x] * 255)
				pos++
				pix[pos] = uint8(buffer[2].FloatBuffer[y][x] * 255)
				pos++

				// FIXME(kpfaulkner) deal with alpha channels properly
				pix[pos] = 255 // uint8(buffer[3].FloatBuffer[y][x] * 255)
				pos++
			}
		}
	} else {
		for y := 0; y < dy; y++ {
			for x := 0; x < dx; x++ {
				pix[pos] = uint8(buffer[0].IntBuffer[y][x])
				pos++
				pix[pos] = uint8(buffer[1].IntBuffer[y][x])
				pos++
				pix[pos] = uint8(buffer[2].IntBuffer[y][x])
				pos++

				// FIXME(kpfaulkner) deal with alpha channels properly
				pix[pos] = uint8(buffer[3].IntBuffer[y][x])
				pos++
			}
		}
	}

	return jxl, nil
}
