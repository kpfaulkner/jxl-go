package core

import (
	"github.com/kpfaulkner/jxl-go/bundle"
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
	frameType         uint32
	width             uint32
	height            uint32
	upsampling        uint32
	lfLevel           uint32
	groupDim          uint32
	passes            *PassesInfo
	encoding          uint32
	flags             uint64
	doYCbCr           bool
	jpegUpsampling    []util.IntPoint
	ecUpsampling      []uint32
	groupSizeShift    uint32
	lfGroupDim        uint32
	logGroupDim       uint32
	logLFGroupDIM     uint32
	xqmScale          uint32
	bqmScale          uint32
	haveCrop          bool
	origin            util.IntPoint
	ecBlendingInfo    []BlendingInfo
	blendingInfo      *BlendingInfo
	isLast            bool
	duration          uint32
	timecode          uint32
	saveAsReference   uint32
	saveBeforeCT      bool
	name              string
	restorationFilter *RestorationFilter
	extensions        *bundle.Extensions
}

func NewFrameHeaderWithReader(reader *jxlio.Bitreader, parent *ImageHeader) (*FrameHeader, error) {
	fh := &FrameHeader{}

	allDefault := reader.MustReadBool()
	if allDefault {
		fh.frameType = REGULAR_FRAME
		fh.encoding = VARDCT
		fh.flags = 0
	} else {
		fh.frameType = uint32(reader.MustReadBits(2))
		fh.encoding = uint32(reader.MustReadBits(1))
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
			fh.jpegUpsampling[i] = util.NewIntPointWithXY(uint32(x^y), uint32(y))
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
		fh.groupSizeShift = uint32(reader.MustReadBits(2))
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
		fh.lfLevel = uint32(reader.MustReadBits(2))
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

	if normalFrame && parent.animationHeader != nil {
		// dont care about animation
		panic("animation")
	} else {
		fh.duration = 0
	}
	if normalFrame && parent.animationHeader != nil && parent.animationHeader.haveTimeCodes {
		// dont care about animation
		panic("animation")
	} else {
		fh.timecode = 0
	}

	if normalFrame {
		fh.isLast = reader.MustReadBool()
	} else {
		fh.isLast = fh.frameType == REGULAR_FRAME
	}

	if !allDefault && fh.frameType != LF_FRAME && !fh.isLast {
		fh.saveAsReference = uint32(reader.MustReadBits(2))
	} else {
		fh.saveAsReference = 0
	}

	if !allDefault && (fh.frameType == REFERENCE_ONLY || fullFrame &&
		(fh.frameType == REGULAR_FRAME || fh.frameType == SKIP_PROGRESSIVE) &&
		(fh.duration == 0 || fh.saveAsReference != 0) &&
		!fh.isLast && fh.blendingInfo.mode == BLEND_REPLACE) {
		fh.saveBeforeCT = reader.MustReadBool()
	} else {
		fh.saveBeforeCT = false
	}

	if allDefault {
		fh.name = ""
	} else {
		nameLen := reader.MustReadU32(0, 0, 0, 4, 16, 5, 48, 10)
		buffer := make([]byte, nameLen)
		for i := 0; i < int(nameLen); i++ {
			buffer[i], err = reader.ReadByte()
			if err != nil {
				return nil, err
			}
			fh.name = string(buffer)
		}
	}
	if allDefault {
		fh.restorationFilter = NewRestorationFilter()
	} else {
		fh.restorationFilter, err = NewRestorationFilterWithReader(reader, fh.encoding)
		if err != nil {
			return nil, err
		}
	}

	if allDefault {
		fh.extensions = bundle.NewExtensions()
	} else {
		fh.extensions, err = bundle.NewExtensionsWithReader(reader)
		if err != nil {
			return nil, err
		}
	}

	// jxlatte has int maxJPX = Stream.of(jpegUpsampling).mapToInt(p -> p.x).reduce(Math::max).getAsInt();
	// which I take it means get the maximum X value from jpegUpsampling. Ditto for Y value.
	maxJPX := 0
	maxJPY := 0
	for _, j := range fh.jpegUpsampling {
		if int(j.X) > maxJPX {
			maxJPX = int(j.X)
		}
		if int(j.Y) > maxJPY {
			maxJPY = int(j.Y)
		}
	}
	fh.width = util.CeilDiv(fh.width, 1<<maxJPX) << maxJPX
	fh.height = util.CeilDiv(fh.height, 1<<maxJPY) << maxJPY

	for i := 0; i < 3; i++ {
		fh.jpegUpsampling[i] = util.NewIntPointWithXY(uint32(maxJPX), uint32(maxJPY)).Minus(fh.jpegUpsampling[i])
	}

	return fh, nil
}
