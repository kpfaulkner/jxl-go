package frame

import (
	"github.com/kpfaulkner/jxl-go/jxlio"
)

type BlendingInfo struct {
	Mode         uint32
	alphaChannel uint32
	clamp        bool
	source       uint32
}

func NewBlendingInfo() *BlendingInfo {
	bi := &BlendingInfo{}
	bi.Mode = BLEND_REPLACE
	bi.alphaChannel = 0
	bi.clamp = false
	bi.source = 0
	return bi
}

func NewBlendingInfoWithReader(reader *jxlio.Bitreader, extra bool, fullFrame bool) (*BlendingInfo, error) {

	bi := &BlendingInfo{}
	bi.Mode = reader.MustReadU32(0, 0, 1, 0, 2, 0, 3, 2)
	//if bi.Mode == 0 {
	//	return bi, nil
	//}

	if extra && (bi.Mode == BLEND_BLEND || bi.Mode == BLEND_MULADD) {
		bi.alphaChannel = reader.MustReadU32(0, 0, 1, 0, 2, 0, 3, 3)
	} else {
		bi.alphaChannel = 0
	}

	if extra && (bi.Mode == BLEND_BLEND ||
		bi.Mode == BLEND_MULT ||
		bi.Mode == BLEND_MULADD) {
		bi.clamp = reader.MustReadBool()
	} else {
		bi.clamp = false
	}

	if bi.Mode != BLEND_REPLACE || !fullFrame {
		bi.source = uint32(reader.MustReadBits(2))
	} else {
		bi.source = 0
	}

	return bi, nil
}
