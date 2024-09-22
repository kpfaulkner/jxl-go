package core

import "github.com/kpfaulkner/jxl-go/jxlio"

type LFGroup struct {
	//lfCoeff        LFCoefficients
	//hfMetadata     HFMetadata

	// lossless doesn't use lfCoeff or hfMetadata afaik... so not implementing them.
	lfCoeff        any
	hfMetadata     any
	lfGroupID      int32
	frame          *Frame
	size           Dimension
	modularLFGroup *ModularStream
}

func NewLFGroup(reader *jxlio.Bitreader, parent *Frame, index int32, replaced []ModularChannel, lfBuffer [][][]float32) (*LFGroup, error) {
	lfg := &LFGroup{
		frame:     parent,
		lfGroupID: index,
	}

	//pixelSize := lfg.frame.getLFGroupSize(lfg.lfGroupID)

	lfg.size = Dimension{
		height: parent.header.height >> 3,
		width:  parent.header.width >> 3,
	}

	stream, err := NewModularStreamWithStreamIndex(reader, parent, int(18+3*parent.numLFGroups+uint32(index)), replaced)
	if err != nil {
		return nil, err
	}

	lfg.modularLFGroup = stream
	err = stream.decodeChannels(reader, false)
	if err != nil {
		return nil, err
	}

	panic("boom")
	return lfg, nil
}
