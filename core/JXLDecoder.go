package core

import (
	"image"
	"io"

	"github.com/kpfaulkner/jxl-go/bundle"
	"github.com/kpfaulkner/jxl-go/jxlio"
)

// JXLDecoder decodes the JXL image
type JXLDecoder struct {

	// input Stream
	in io.ReadSeeker

	// decoder
	decoder *JXLCodestreamDecoder
}

func NewJXLDecoder(in io.ReadSeeker) *JXLDecoder {
	jxl := &JXLDecoder{
		in: in,
	}

	br := jxlio.NewBitreader(in)
	jxl.decoder = NewJXLCodestreamDecoder(br, nil)
	return jxl
}

func (jxl *JXLDecoder) Decode() (image.Image, error) {

	jxlImage, err := jxl.decoder.decode()
	if err != nil {
		return nil, err
	}

	// convert to regular Go image.Image
	img, err := NewImageFromJXLImage(jxlImage)
	if err != nil {
		return nil, err
	}
	return img, nil

}

func (jxl *JXLDecoder) GetImageHeader() (*bundle.ImageHeader, error) {

	header, err := jxl.decoder.GetImageHeader()
	if err != nil {
		return nil, err
	}

	return header, nil

}
