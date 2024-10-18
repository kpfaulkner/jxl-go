package frame

import (
	"github.com/kpfaulkner/jxl-go/entropy"
	"github.com/kpfaulkner/jxl-go/jxlio"
	"github.com/kpfaulkner/jxl-go/util"
)

type Patch struct {
	Width         int32
	Height        int32
	Bounds        util.Rectangle
	Ref           int32
	Origin        util.IntPoint
	Positions     []util.IntPoint
	BlendingInfos [][]BlendingInfo
}

func NewPatchWithStreamAndReader(stream *entropy.EntropyStream, reader *jxlio.Bitreader, extraChannelCount int, alphaChannelCount int) (Patch, error) {

	return Patch{}, nil
}
