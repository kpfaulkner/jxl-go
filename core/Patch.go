package core

import (
	"github.com/kpfaulkner/jxl-go/entropy"
	"github.com/kpfaulkner/jxl-go/jxlio"
)

type Patch struct {
}

func NewPatchWithStreamAndReader(stream *entropy.EntropyStream, reader *jxlio.Bitreader, extraChannelCount int, alphaChannelCount int) (Patch, error) {

	return Patch{}, nil
}
