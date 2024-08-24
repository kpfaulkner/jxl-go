package core

import (
	"github.com/kpfaulkner/jxl-go/jxlio"
)

const (
	REGULAR_FRAME = 0
)

type FrameHeader struct {
	frameType uint32
}

func NewFrameHeaderWithReader(reader *jxlio.Bitreader, imageHeader *ImageHeader) *FrameHeader {
	fh := &FrameHeader{}

	allDefault := reader.MustReadBool()
	if allDefault {
		fh.frameType = REGULAR_FRAME
	} else {
		fh.frameType = reader.MustReadBits(2)
	}

	return fh
}
