package frame

import (
	"github.com/kpfaulkner/jxl-go/jxlio"
	"github.com/kpfaulkner/jxl-go/util"
)

type Pass struct {
	minShift         uint32
	maxShift         uint32
	replacedChannels []*ModularChannel
	hfPass           *HFPass
}

// TODO(kpfaulkner) 20241025 hfPass wrong after a few attempts....
func NewPassWithReader(reader *jxlio.Bitreader, frame *Frame, passIndex uint32, prevMinShift uint32) (Pass, error) {
	p := Pass{}

	if passIndex > 0 {
		p.maxShift = prevMinShift
	} else {
		p.maxShift = 3
	}

	n := -1
	passes := frame.Header.passes
	for i := 0; i < len(passes.lastPass); i++ {
		if passes.lastPass[i] == passIndex {
			n = i
			break
		}
	}

	if n >= 0 {
		p.minShift = uint32(util.CeilLog1p(int64(passes.downSample[n] - 1)))
	} else {
		p.minShift = p.maxShift
	}

	globalModular := frame.LfGlobal.gModular
	p.replacedChannels = make([]*ModularChannel, len(globalModular.Stream.channels))
	for i := 0; i < len(globalModular.Stream.channels); i++ {
		ch := globalModular.Stream.channels[i]
		if !ch.decoded {
			m := uint32(min(ch.hshift, ch.vshift))
			if p.minShift <= m && m < p.maxShift {
				p.replacedChannels[i] = NewModularChannelFromChannel(*ch)
			}
		}
	}
	var err error
	if frame.Header.Encoding == VARDCT {
		p.hfPass, err = NewHFPassWithReader(reader, frame, passIndex)
		if err != nil {
			return Pass{}, err
		}
	} else {
		p.hfPass = nil
	}

	return p, nil

}
