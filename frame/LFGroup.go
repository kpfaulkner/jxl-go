package frame

import (
	"github.com/kpfaulkner/jxl-go/jxlio"
	"github.com/kpfaulkner/jxl-go/util"
)

type LFGroup struct {
	//lfCoeff        LFCoefficients
	//hfMetadata     HFMetadata

	// lossless doesn't use lfCoeff or hfMetadata afaik... so not implementing them.
	lfCoeff        any
	hfMetadata     any
	lfGroupID      int32
	frame          *Frame
	size           util.Dimension
	modularLFGroup *ModularStream
}

func NewLFGroup(reader *jxlio.Bitreader, parent *Frame, index int32, replaced []ModularChannel, lfBuffer [][][]float32) (*LFGroup, error) {
	lfg := &LFGroup{
		frame:     parent,
		lfGroupID: index,
	}

	pixelSize, err := lfg.frame.getLFGroupSize(lfg.lfGroupID)
	if err != nil {
		return nil, err
	}

	lfg.size = util.Dimension{
		Height: pixelSize.Height >> 3,
		Width:  pixelSize.Width >> 3,
	}

	if parent.Header.Encoding == VARDCT {
		panic("VARDCT not implemented")
	} else {
		lfg.lfCoeff = nil
	}
	stream, err := NewModularStreamWithStreamIndex(reader, parent, 1+int(parent.numLFGroups+uint32(lfg.lfGroupID)), replaced)
	if err != nil {
		return nil, err
	}

	lfg.modularLFGroup = stream
	err = stream.decodeChannels(reader, false)
	if err != nil {
		return nil, err
	}

	if parent.Header.Encoding == VARDCT {
		panic("VARDCT not implemented")
	} else {
		lfg.hfMetadata = nil
	}
	return lfg, nil
}
