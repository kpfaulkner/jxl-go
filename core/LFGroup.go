package core

import "github.com/kpfaulkner/jxl-go/jxlio"

type LFGroup struct {
	lfCoeff        LFCoefficients
	hfMetadata     HFMetadata
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

	pixelSize := lfg.frame.getLFGroupSize(lfg.lfGroupID)

	lfGroup.size = Dimension{
		X: parent.header.width >> parent.header.lfGroupDim,
		Y: parent.header.height >> parent.header.lfGroupDim,
	}

	stream, err := NewModularStreamWithStreamIndex(reader, parent, int(18+3*parent.numLFGroups+index), replaced)
	if err != nil {
		return nil, err
	}

	lfGroup.modularLFGroup = stream
	err = stream.decodeChannels(reader, false)
	if err != nil {
		return nil, err
	}

	panic("boom")
	return lfGroup, nil
}
