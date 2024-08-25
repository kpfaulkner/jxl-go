package core

import (
	"github.com/kpfaulkner/jxl-go/jxlio"
)

type AnimationHeader struct {
	haveTimeCodes bool
}

func NewAnimationHeader(reader *jxlio.Bitreader) (*AnimationHeader, error) {
	ah := &AnimationHeader{}

	panic("Animation Header not implemented")
	return ah, nil
}
