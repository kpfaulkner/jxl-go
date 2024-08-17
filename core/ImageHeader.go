package core

import (
	"errors"

	"github.com/kpfaulkner/jxl-go/jxlio"
)

const (
	CODESTREAM_HEADER int32 = 0x0AFF
)

type ImageHeader struct {
	level int32
	size  *SizeHeader
}

func NewImageHeader() *ImageHeader {
	ih := &ImageHeader{}
	return ih
}

func ParseImageHeader(reader *jxlio.Bitreader, level int32) (*ImageHeader, error) {
	header := NewImageHeader()

	if reader.MustReadBits(16) != CODESTREAM_HEADER {
		return nil, errors.New("Not a JXL codestream: 0xFF0A majoc mismatch")
	}

	err := header.setLevel(level)
	if err != nil {
		return nil, err
	}

	header.size = NewSizeHeader(reader, level)
	return header, nil
}

func (h *ImageHeader) setLevel(level int32) error {
	if level != 5 && level != 10 {
		return errors.New("invalid bitstream")
	}
	h.level = level
	return nil
}
