package core

import "github.com/kpfaulkner/jxl-go/jxlio"

type BlendingInfo struct {
	mode         uint32
	alphaChannel uint32
	clamp        bool
	source       uint32
}

func NewBlendingInfo() *BlendingInfo {
	bi := &BlendingInfo{}
	bi.mode = BLEND_REPLACE
	bi.alphaChannel = 0
	bi.clamp = false
	bi.source = 0
	return bi
}

func NewBlendingInfoWithReader(reader *jxlio.Bitreader, extra bool, fullFrame bool) (*BlendingInfo, error) {

	bi := &BlendingInfo{}
	bi.mode = reader.MustReadU32()
	if bi.mode == 0 {
		return bi, nil
	}

	bi.alphaChannel = reader.MustReadU32(0, 0, 1, 0, 2, 0, 3, 2)
	if extra && (bi.mode == BLEND_BLEND || bi.mode == BLEND_MULADD) {
		bi.alphaChannel = reader.MustReadU32(0, 0, 1, 0, 2, 0, 3, 3)
	} else {
		bi.alphaChannel = 0
	}

	if extra && (bi.mode == BLEND_BLEND ||
		bi.mode == BLEND_MULT ||
		bi.mode == BLEND_MULADD) {
		bi.clamp = reader.MustReadBool()
	} else {
		bi.clamp = false
	}

	if bi.mode != BLEND_REPLACE || !fullFrame {
		bi.source = reader.MustReadBits(2)
	} else {
		bi.source = 0
	}

	return bi, nil
}
