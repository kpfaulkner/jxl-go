package frame

import (
	"github.com/kpfaulkner/jxl-go/jxlio"
)

type PassGroup struct {
	modularPassGroupBuffer [][][]int32
	modularStream          *ModularStream
	frame                  *Frame
	groupID                uint32
	passID                 uint32
	hfCoefficients         *HFCoefficients
	lfg                    *LFGroup
}

func NewPassGroupWithReader(reader *jxlio.Bitreader, frame *Frame, pass uint32, group uint32, replacedChannels []ModularChannel) (*PassGroup, error) {

	pg := &PassGroup{}
	pg.frame = frame
	pg.groupID = group
	pg.passID = pass
	if frame.Header.Encoding == VARDCT {
		coeff, err := NewHFCoefficientsWithReader(reader, frame, pass, group)
		if err != nil {
			return nil, err
		}
		pg.hfCoefficients = coeff
		panic("VARDCT not implemented")
	} else {
		pg.hfCoefficients = nil
	}

	stream, err := NewModularStreamWithStreamIndex(reader, frame, int(18+3*frame.numLFGroups+frame.numGroups*pass+group), replacedChannels)
	if err != nil {
		return nil, err
	}

	pg.modularStream = stream
	err = stream.decodeChannels(reader, false)
	if err != nil {
		return nil, err
	}

	pg.lfg = frame.getLFGroupForGroup(int32(group))

	return pg, nil
}
