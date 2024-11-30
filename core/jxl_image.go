package core

import (
	"fmt"
	"image"

	"github.com/kpfaulkner/jxl-go/bundle"
	"github.com/kpfaulkner/jxl-go/color"
	image2 "github.com/kpfaulkner/jxl-go/image"
	"github.com/kpfaulkner/jxl-go/util"
)

const (

	// unsure if will use these yet.
	PEAK_DETECT_AUTO = -1
	PEAK_DETECT_ON   = 1
	PEAK_DETECT_OFF  = 2
)

// JXLImage contains the core information about the JXL image.
type JXLImage struct {
	Width                uint32
	Height               uint32
	Buffer               []image2.ImageBuffer
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
	bitDepths            []uint32
}

// NewJXLImageWithBuffer creates a new JXLImage with the given buffer and header.
func NewJXLImageWithBuffer(buffer []image2.ImageBuffer, header bundle.ImageHeader) (*JXLImage, error) {
	jxl := &JXLImage{}

	jxl.imageHeader = header
	jxl.Width = header.OrientedWidth
	jxl.Height = header.OrientedHeight
	jxl.Buffer = buffer
	bundle := header.ColorEncoding
	jxl.ColorEncoding = bundle.ColorEncoding
	if header.HasAlpha() {
		jxl.alphaIndex = header.AlphaIndices[0]
	} else {
		jxl.alphaIndex = -1
	}
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
	jxl.bitDepths = make([]uint32, len(buffer))
	colours := util.IfThenElse(jxl.ColorEncoding == color.CE_GRAY, 1, 3)
	for c := 0; c < len(jxl.bitDepths); c++ {
		if c < colours {
			jxl.bitDepths[c] = header.BitDepth.BitsPerSample
		} else {
			jxl.bitDepths[c] = header.ExtraChannelInfo[c-colours].BitDepth.BitsPerSample
		}
	}
	return jxl, nil
}

// GetFloatChannelData will return the floating point image data for a channel.
// The underlying image MAY not have any floating point data (this is all image dependant).
func (jxl *JXLImage) GetFloatChannelData(c int) ([][]float32, error) {
	if c < 0 || c >= len(jxl.Buffer) {
		return nil, fmt.Errorf("Invalid channel index %d", c)
	}
	return jxl.Buffer[c].FloatBuffer, nil
}

// GetIntChannelData will return the integer image data for a channel.
// The underlying image MAY not have any integer point data (this is all image dependant).
func (jxl *JXLImage) GetIntChannelData(c int) ([][]int32, error) {
	if c < 0 || c >= len(jxl.Buffer) {
		return nil, fmt.Errorf("Invalid channel index %d", c)
	}
	return jxl.Buffer[c].IntBuffer, nil
}

// IsIntBased returns true if underlying data related to image is integer based.
func (jxl *JXLImage) IsIntBased() bool {
	return jxl.Buffer[0].IsInt()
}

// IsFloatBased returns true if underlying data related to image is float based.
func (jxl *JXLImage) IsFloatBased() bool {
	return jxl.Buffer[0].IsFloat()
}

func (jxl *JXLImage) HasAlpha() bool {
	return jxl.imageHeader.HasAlpha()
}

func (jxl *JXLImage) NumExtraChannels() int {
	return len(jxl.imageHeader.ExtraChannelInfo)
}

// ChannelToImage converts a single channel to grayscale Go image.Image interface.
// Can be used for any channel (R,G,B, alpha, depth..... etc) but is really expected to be
// used for NON "regular" channels (ie depth etc)
// TODO(kpfaulkner) still need to confirm if generating an 8 bit grayscale image is appropriate
// for this, or if Gray16... or even RGBA with generating the RGB values itself is the way to
// proceed.
func (jxl *JXLImage) ChannelToImage(channelNo int) (image.Image, error) {
	buffer := jxl.getBuffer(true)
	if channelNo < 0 || channelNo >= len(buffer) {
		return nil, fmt.Errorf("Invalid channel index %d", channelNo)
	}
	//img := image.NewRGBA(image.Rect(0, 0, int(buffer[0].Width), int(buffer[0].Height)))
	img := image.NewGray(image.Rect(0, 0, int(buffer[0].Width), int(buffer[0].Height)))
	pix := img.Pix
	dx := img.Bounds().Dx()
	dy := img.Bounds().Dy()
	pos := 0
	if buffer[0].IsFloat() {
		for y := 0; y < dy; y++ {
			for x := 0; x < dx; x++ {
				pix[pos] = uint8(buffer[channelNo].FloatBuffer[y][x] * 255)
				pos++
			}
		}
	} else {

		for y := 0; y < dy; y++ {
			for x := 0; x < dx; x++ {
				pix[pos] = uint8(buffer[channelNo].IntBuffer[y][x])
				pos++
			}
		}
	}
	return img, nil
}

// ToImage converts to standard Go image.Image RGBA format for the R,G,B and alpha channels
func (jxl *JXLImage) ToImage() (image.Image, error) {

	var bitDepth int32
	if jxl.imageHeader.BitDepth.BitsPerSample > 8 {
		bitDepth = 16
	} else {
		bitDepth = 8
	}

	//gray := jxlImage.ColorEncoding == color.CE_GRAY
	primaries := color.CM_PRI_SRGB
	tf := color.TF_SRGB
	if jxl.isHDR() {
		primaries = color.CM_PRI_BT2100
		tf = color.TF_PQ
	}
	whitePoint := color.CM_WP_D65
	iccProfile := jxl.iccProfile
	var err error
	if iccProfile == nil {

		// transforms in place
		err = jxl.transform(primaries, whitePoint, tf, PEAK_DETECT_AUTO)
		if err != nil {
			return nil, err
		}
	}
	maxValue := int32(^(^0 << bitDepth))
	coerce := jxl.alphaIsPremultiplied
	buffer := jxl.getBuffer(true)
	if !coerce {
		for c := 0; c < len(buffer); c++ {
			if buffer[c].IsInt() && jxl.bitDepths[c] != uint32(bitDepth) {
				coerce = true
				break
			}
		}
	}
	if coerce {
		for c := 0; c < len(buffer); c++ {
			if err := buffer[c].CastToFloatIfInt(^(^0 << jxl.bitDepths[c])); err != nil {
				return nil, err
			}
		}
	}
	if jxl.alphaIsPremultiplied {
		panic("not implemented")
	}
	for c := 0; c < len(buffer); c++ {
		if buffer[c].IsInt() && jxl.bitDepths[c] == uint32(bitDepth) {
			if err := buffer[c].Clamp(maxValue); err != nil {
				return nil, err
			}
		} else {
			if err := buffer[c].CastToIntIfFloat(maxValue); err != nil {
				return nil, err
			}
		}
	}

	img := image.NewRGBA(image.Rect(0, 0, int(buffer[0].Width), int(buffer[0].Height)))
	pix := img.Pix
	dx := img.Bounds().Dx()
	dy := img.Bounds().Dy()
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

				if jxl.imageHeader.HasAlpha() {
					pix[pos] = uint8(buffer[3].IntBuffer[y][x])
					pos++
				} else {
					pix[pos] = 255
					pos++
				}
			}
		}
	}
	return img, nil
}

func (jxl *JXLImage) isHDR() bool {
	if jxl.taggedTransfer == color.TF_PQ || jxl.taggedTransfer == color.TF_HLG ||
		jxl.taggedTransfer == color.TF_LINEAR {
		return true
	}
	colour := jxl.imageHeader.ColorEncoding
	return !colour.Prim.Matches(color.CM_PRI_SRGB) &&
		!colour.Prim.Matches(color.CM_PRI_P3)
}

func (jxl *JXLImage) transform(primaries *color.CIEPrimaries, whitePoint *color.CIEXY, transfer int32, peakDetect int32) error {

	if primaries.Matches(jxl.primariesXY) && whitePoint.Matches(jxl.whiteXY) {
		return jxl.transferImage(transfer, peakDetect)
	}

	if err := jxl.linearize(); err != nil {
		return err
	}

	if err := jxl.toneMapLinear(); err != nil {
		return err
	}

	if err := jxl.transferImage(transfer, peakDetect); err != nil {
		return err
	}

	return nil
}

// getBuffer gets a copy of the buffer... making a copy if required
// REALLY not optimised but will fix later.
func (jxl *JXLImage) getBuffer(makeCopy bool) []image2.ImageBuffer {
	if !makeCopy {
		return jxl.Buffer
	}

	buffer := make([]image2.ImageBuffer, len(jxl.Buffer))
	for c := 0; c < len(jxl.Buffer); c++ {
		buffer[c] = *image2.NewImageBuffer(jxl.Buffer[c].BufferType, jxl.Buffer[c].Height, jxl.Buffer[c].Width)

		// very stupid... will optimise later TODO(kpfaulkner)
		if buffer[c].IsInt() {
			for y := 0; y < int(jxl.Buffer[c].Height); y++ {
				for x := 0; x < int(jxl.Buffer[c].Width); x++ {
					buffer[c].IntBuffer[y][x] = jxl.Buffer[c].IntBuffer[y][x]
				}
			}
		} else {
			for y := 0; y < int(jxl.Buffer[c].Height); y++ {
				for x := 0; x < int(jxl.Buffer[c].Width); x++ {
					buffer[c].FloatBuffer[y][x] = jxl.Buffer[c].FloatBuffer[y][x]
				}
			}
		}
	}
	return buffer
}

func (jxl *JXLImage) transferImage(transfer int32, peakDetect int32) error {
	if transfer == jxl.transfer {
		return nil
	}
	if err := jxl.linearize(); err != nil {
		return err
	}
	if jxl.taggedTransfer == color.TF_PQ &&
		(peakDetect == PEAK_DETECT_AUTO || peakDetect == PEAK_DETECT_ON) {
		panic("not implemented")
	}

	transferFunction, err := color.GetTransferFunction(transfer)
	if err != nil {
		return err
	}
	if err := jxl.transferInPlace(transferFunction.FromLinear); err != nil {
		return err
	}

	return nil
}

func (jxl *JXLImage) linearize() error {

	if jxl.transfer == color.TF_LINEAR {
		return nil
	}

	panic("not implemented")
}

func (jxl *JXLImage) toneMapLinear() error {
	return nil
}

func (jxl *JXLImage) transferInPlace(transferFunction func(float64) float64) error {
	colours := 3
	if jxl.ColorEncoding == color.CE_GRAY {
		colours = 1
	}

	buffers := util.MakeMatrix3D[float32](colours, 0, 0)
	for c := 0; c < colours; c++ {
		jxl.Buffer[c].CastToFloatIfInt(int32(jxl.bitDepths[c]))
		buffers[c] = jxl.Buffer[c].FloatBuffer
	}

	for c := 0; c < colours; c++ {
		for y := 0; y < int(jxl.Height); y++ {
			for x := 0; x < int(jxl.Width); x++ {
				buffers[c][y][x] = float32(transferFunction(float64(buffers[c][y][x])))
			}
		}
	}
	return nil
}
