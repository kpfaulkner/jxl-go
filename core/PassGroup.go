package core

import (
	"errors"
	"fmt"

	"github.com/kpfaulkner/jxl-go/jxlio"
)

type PassGroup struct {
	modularPassGroupBuffer [][][]int32
	modularPassGroupInfo   []ModularChannelInfo
	frame                  *Frame
	groupID                uint32
	passID                 uint32
	hfCoefficients         []HFCoefficients
}

func NewPassGroupWithReader(reader *jxlio.Bitreader, frame *Frame, pass uint32, group uint32, replacedChannels []ModularChannelInfo) (*PassGroup, error) {
	pg := &PassGroup{}
	pg.frame = frame
	pg.groupID = group
	pg.passID = pass
	if frame.header.encoding == VARDCT {
		panic("VARDCT not implemented")
	} else {
		pg.hfCoefficients = nil
	}

	fmt.Printf("Passgroup pass %d group %d\n", pass, group)
	stream, err := NewModularStreamWithStreamIndex(reader, frame, int(18+3*frame.numLFGroups+frame.numGroups*pass+group), replacedChannels)
	if err != nil {
		return nil, err
	}

	if pass == 0 && group == 4 {
		fmt.Printf("boom\n")
	}
	err = stream.decodeChannels(reader, false)
	if err != nil {
		return nil, err
	}

	pg.modularPassGroupBuffer = stream.getDecodedBuffer()
	pg.modularPassGroupInfo = make([]ModularChannelInfo, len(pg.modularPassGroupBuffer))
	for c := 0; c < len(pg.modularPassGroupBuffer); c++ {
		ci, ok := stream.channels[c].(*ModularChannelInfo)
		if !ok {
			return nil, errors.New("ModularChannelInfo not found")
		}

		pg.modularPassGroupInfo[c] = *ci
	}

	return pg, nil
}
