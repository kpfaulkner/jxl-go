package frame

import (
	"github.com/kpfaulkner/jxl-go/jxlio"
)

type BlendingInfo struct {
	Mode         uint32
	AlphaChannel uint32
	Clamp        bool
	Source       uint32
}

func NewBlendingInfo() *BlendingInfo {
	bi := &BlendingInfo{}
	bi.Mode = BLEND_REPLACE
	bi.AlphaChannel = 0
	bi.Clamp = false
	bi.Source = 0
	return bi
}

func NewBlendingInfoWithReader(reader *jxlio.Bitreader, extra bool, fullFrame bool) (*BlendingInfo, error) {

	bi := &BlendingInfo{}
	bi.Mode = reader.MustReadU32(0, 0, 1, 0, 2, 0, 3, 2)
	//if bi.Mode == 0 {
	//	return bi, nil
	//}

	if extra && (bi.Mode == BLEND_BLEND || bi.Mode == BLEND_MULADD) {
		bi.AlphaChannel = reader.MustReadU32(0, 0, 1, 0, 2, 0, 3, 3)
	} else {
		bi.AlphaChannel = 0
	}

	var err error
	if extra && (bi.Mode == BLEND_BLEND ||
		bi.Mode == BLEND_MULT ||
		bi.Mode == BLEND_MULADD) {
		if bi.Clamp, err = reader.ReadBool(); err != nil {
			return nil, err
		}
	} else {
		bi.Clamp = false
	}

	if bi.Mode != BLEND_REPLACE || !fullFrame {
		bi.Source = uint32(reader.MustReadBits(2))
	} else {
		bi.Source = 0
	}

	return bi, nil
}
