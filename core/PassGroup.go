package core

import (
	"fmt"

	"github.com/kpfaulkner/jxl-go/jxlio"
)

type PassGroup struct {
	modularPassGroupBuffer [][][]int32
	modularStream          *ModularStream
	frame                  *Frame
	groupID                uint32
	passID                 uint32
	hfCoefficients         []HFCoefficients
	lfg                    *LFGroup
}

func NewPassGroupWithReader(reader *jxlio.Bitreader, frame *Frame, pass uint32, group uint32, replacedChannels []ModularChannel) (*PassGroup, error) {

	fmt.Printf("PassGroup reader pos %d\n", reader.BitsRead())
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
	pg.modularStream = stream
	err = stream.decodeChannels(reader, false)
	if err != nil {
		return nil, err
	}

	return pg, nil
}
