package core

import (
	"image"
	"io"
)

// JXLDecoder decodes the JXL image
type JXLDecoder struct {

	// input stream
	in io.ReadSeeker

	// decoder
	decoder *JXLCodestreamDecoder
}

func NewJXLDecoder(in io.ReadSeeker) *JXLDecoder {
	jxl := &JXLDecoder{
		in: in,
	}

	jxl.decoder = NewJXLCodestreamDecoder(jxl.in, nil)
	return jxl
}

func (jxl *JXLDecoder) Decode() (image.Image, error) {

	jxlImage, err := jxl.decoder.decode()
	if err != nil {
		return nil, err
	}

	return jxlImage, nil

}

func (jxl *JXLDecoder) GetImageHeader() (*ImageHeader, error) {

	header, err := jxl.decoder.GetImageHeader()
	if err != nil {
		return nil, err
	}

	return header, nil

}
