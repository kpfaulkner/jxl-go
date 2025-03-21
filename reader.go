package jxl_go

import (
	"errors"
	"image"
	color2 "image/color"
	"io"

	"github.com/kpfaulkner/jxl-go/colour"
	"github.com/kpfaulkner/jxl-go/core"
)

const jxlHeader = "\x00\x00\x00\x0C\x4A\x58\x4C\x20\x0D\x0A\x87\x0A"

func init() {
	image.RegisterFormat("jxl", jxlHeader, Decode, DecodeConfig)
}

func Decode(r io.Reader) (image.Image, error) {

	var rs io.ReadSeeker
	var ok bool

	if rs, ok = r.(io.ReadSeeker); !ok {
		return nil, errors.New("required ReadSeeker")
	}

	jxl := core.NewJXLDecoder(rs, nil)

	var img image.Image
	var jxlImg *core.JXLImage
	var err error
	if jxlImg, err = jxl.Decode(); err != nil {
		return nil, err
	}

	if img, err = jxlImg.ToImage(); err != nil {
		return nil, err
	}
	return img, nil
}

func DecodeConfig(r io.Reader) (image.Config, error) {
	var rs io.ReadSeeker
	var ok bool

	if rs, ok = r.(io.ReadSeeker); !ok {
		return image.Config{}, errors.New("required ReadSeeker")
	}

	jxl := core.NewJXLDecoder(rs, nil)
	header, err := jxl.GetImageHeader()
	if err != nil {
		return image.Config{}, err
	}

	var colourModel color2.Model

	switch header.GetColourModel() {
	case colour.CE_RGB:
		colourModel = color2.RGBAModel
	case colour.CE_XYB:
		colourModel = color2.GrayModel // unsure how to deal with XYB
	case colour.CE_GRAY:
		colourModel = color2.GrayModel
	default:
		colourModel = color2.RGBAModel
	}

	dim := header.GetSize()
	return image.Config{
		ColorModel: colourModel,
		Width:      int(dim.Width),
		Height:     int(dim.Height),
	}, nil

}
