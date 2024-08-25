package core

import (
	"github.com/kpfaulkner/jxl-go/jxlio"
	"github.com/kpfaulkner/jxl-go/util"
)

const (
	REGULAR_FRAME    = 0
	LF_FRAME         = 1
	REFERENCE_ONLY   = 2
	SKIP_PROGRESSIVE = 3

	VARDCT  = 0
	MODULAR = 1

	USE_LF_FRAME = 32

	BLEND_REPLACE = 0
	BLEND_ADD     = 1
	BLEND_BLEND   = 2
	BLEND_MULADD  = 3
	BLEND_MULT    = 4
)

type FrameHeader struct {
	frameType      uint32
	width          uint32
	height         uint32
	upsampling     uint32
	lfLevel        uint32
	groupDim       uint32
	passes         *PassesInfo
	encoding       uint32
	flags          uint64
	doYCbCr        bool
	jpegUpsampling []util.IntPoint
	ecUpsampling   []uint32
	groupSizeShift uint32
	lfGroupDim     uint32
	logGroupDim    uint32
	logLFGroupDIM  uint32
	xqmScale       uint32
	bqmScale       uint32
	haveCrop       bool
	origin         util.IntPoint
	ecBlendingInfo []BlendingInfo
	blendingInfo   *BlendingInfo
}

func NewFrameHeaderWithReader(reader *jxlio.Bitreader, parent *ImageHeader) (*FrameHeader, error) {
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
	} else {
		fh.jpegUpsampling = []util.IntPoint{util.ZERO, util.ZERO, util.ZERO}
	}

	fh.ecUpsampling = make([]uint32, len(parent.extraChannelInfo))
	if !allDefault && (fh.flags&USE_LF_FRAME) == 0 {
		fh.upsampling = 1 << reader.MustReadBits(2)
		for i := 0; i < len(fh.ecUpsampling); i++ {
			fh.ecUpsampling[i] = 1 << reader.MustReadBits(2)
		}
	} else {
		fh.upsampling = 1
		fh.ecUpsampling = []uint32{1}
	}

	if fh.encoding == MODULAR {
		fh.groupSizeShift = reader.MustReadBits(2)
	} else {
		fh.groupSizeShift = 1
	}
	fh.groupDim = 128 << fh.groupSizeShift
	fh.lfGroupDim = fh.groupDim << 3
	fh.logGroupDim = uint32(util.CeilLog2(int64(fh.groupDim)))
	fh.logLFGroupDIM = uint32(util.CeilLog2(int64(fh.lfGroupDim)))
	if parent.xybEncoded && fh.encoding == VARDCT {
		panic("VARDCT not implemented")
	} else {
		fh.xqmScale = 2
		fh.bqmScale = 2
	}

	var err error
	if !allDefault && fh.frameType != REFERENCE_ONLY {
		fh.passes, err = NewPassesInfoWithReader(reader)
		if err != nil {
			return nil, err
		}
	} else {
		fh.passes = NewPassesInfo()
	}

	if fh.frameType == LF_FRAME {
		fh.lfLevel = reader.MustReadBits(2)
	} else {
		fh.lfLevel = 0
	}
	if !allDefault && fh.frameType != LF_FRAME {
		fh.haveCrop = reader.MustReadBool()
	} else {
		fh.haveCrop = false
	}

	if fh.haveCrop && fh.frameType != REFERENCE_ONLY {
		x0 := reader.MustReadU32(0, 8, 256, 11, 2304, 14, 18688, 30)
		y0 := reader.MustReadU32(0, 8, 256, 11, 2304, 14, 18688, 30)
		x0Signed := jxlio.UnpackSigned(x0)
		y0Signed := jxlio.UnpackSigned(y0)
		fh.origin = util.NewIntPointWithXY(uint32(x0Signed), uint32(y0Signed))

	}

	if fh.haveCrop {
		fh.width = reader.MustReadU32(0, 8, 256, 11, 2304, 14, 18688, 30)
		fh.height = reader.MustReadU32(0, 8, 256, 11, 2304, 14, 18688, 30)
	} else {
		fh.width = parent.size.width
		fh.height = parent.size.height
	}

	normalFrame := !allDefault && (fh.frameType == REGULAR_FRAME || fh.frameType == SKIP_PROGRESSIVE)
	fullFrame := fh.origin.X <= 0 && fh.origin.Y <= 0 && (fh.width+fh.origin.X) >= parent.size.width && (fh.height+fh.origin.Y) >= parent.size.height
	fh.ecBlendingInfo = make([]BlendingInfo, len(parent.extraChannelInfo))
	if normalFrame {
		fh.blendingInfo, err = NewBlendingInfoWithReader(reader, len(fh.ecBlendingInfo) > 0, fullFrame)
		if err != nil {
			return nil, err
		}
		for i := 0; i < len(fh.ecBlendingInfo); i++ {
			bi, err := NewBlendingInfoWithReader(reader, true, fullFrame)
			if err != nil {
				return nil, err
			}
			// store value not pointer. TODO(kpfaulkner) check this is fine.
			fh.ecBlendingInfo[i] = *bi
		}
	} else {
		fh.blendingInfo = NewBlendingInfo()
		for i := 0; i < len(fh.ecBlendingInfo); i++ {
			fh.ecBlendingInfo[i] = *fh.blendingInfo
		}
	}

	// TODO(kpfaulkner) to continue......

	return fh
}
