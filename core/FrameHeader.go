package core

import (
	"github.com/kpfaulkner/jxl-go/jxlio"
	"github.com/kpfaulkner/jxl-go/util"
)

const (
	REGULAR_FRAME = 0

	VARDCT  = 0
	MODULAR = 1

	USE_LF_FRAME = 32
)

type FrameHeader struct {
	frameType      uint32
	width          uint32
	height         uint32
	upsampling     uint32
	lfLevel        uint32
	groupDim       uint32
	passes         PassesInfo
	encoding       uint32
	flags          uint64
	doYCbCr        bool
	jpegUpsampling []util.IntPoint
}

func NewFrameHeaderWithReader(reader *jxlio.Bitreader, parent *ImageHeader) *FrameHeader {
	fh := &FrameHeader{}

	allDefault := reader.MustReadBool()
	if allDefault {
		fh.frameType = REGULAR_FRAME
		fh.encoding = VARDCT
		fh.flags = 0
	} else {
		fh.frameType = reader.MustReadBits(2)
		fh.encoding = reader.MustReadBits(1)
		fh.flags = reader.MustReadU64()
	}

	if !allDefault && !parent.xybEncoded {
		fh.doYCbCr = reader.MustReadBool()
	} else {
		fh.doYCbCr = false
	}
	fh.jpegUpsampling = make([]util.IntPoint, 3)
	if fh.doYCbCr && (fh.flags&USE_LF_FRAME) == 0 {
		for i := 0; i < 3; i++ {
			y := reader.MustReadBits(1)
			x := reader.MustReadBits(1)
			fh.jpegUpsampling[i] = util.NewIntPointWithXY(x^y, y)
		}
	}

	return fh
}
