package core

import (
	"github.com/kpfaulkner/jxl-go/jxlio"
	"github.com/kpfaulkner/jxl-go/util"
)

type Pass struct {
	minShift         uint32
	maxShift         uint32
	replacedChannels []ModularChannelInfo
	hfPass           *HFPass
}

func NewPassWithReader(reader *jxlio.Bitreader, frame *Frame, passIndex uint32, prevMinShift uint32) (Pass, error) {
	p := Pass{}

	if passIndex > 0 {
		p.maxShift = prevMinShift
	} else {
		p.maxShift = 3
	}

	n := -1
	passes := frame.header.passes
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

	globalModular := frame.lfGlobal.gModular
	p.replacedChannels = make([]ModularChannelInfo, len(globalModular.stream.channels))
	for i := 0; i < len(globalModular.stream.channels); i++ {
		ch, ok := globalModular.stream.channels[i].(*ModularChannel)
		if !ok {
			return p, nil
		}
		if !ch.decoded {
			m := uint32(min(ch.hshift, ch.vshift))
			if p.minShift <= m && m < p.maxShift {
				p.replacedChannels[i] = *NewModularChannelInfoFromInfo((*ch).ModularChannelInfo)
			}
		}
	}
	var err error
	if frame.header.encoding == VARDCT {
		p.hfPass, err = NewHFPassWithReader(reader, frame, passIndex)
		if err != nil {
			return Pass{}, err
		}
	} else {
		p.hfPass = nil
	}

	return p, nil

}