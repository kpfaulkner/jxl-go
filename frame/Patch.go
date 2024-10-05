package frame

import (
	"github.com/kpfaulkner/jxl-go/entropy"
	"github.com/kpfaulkner/jxl-go/jxlio"
	"github.com/kpfaulkner/jxl-go/util"
)

type Patch struct {
	width         int32
	height        int32
	ref           int32
	origin        util.IntPoint
	positions     []util.IntPoint
	blendingInfos [][]BlendingInfo
}

func NewPatchWithStreamAndReader(stream *entropy.EntropyStream, reader *jxlio.Bitreader, extraChannelCount int, alphaChannelCount int) (Patch, error) {

	return Patch{}, nil
}
