package core

import (
	"github.com/kpfaulkner/jxl-go/color"
	"github.com/kpfaulkner/jxl-go/jxlio"
)

type GlobalModular struct {
	frame      *Frame
	globalTree *MATree
	stream     *ModularStream
}

func NewGlobalModularWithReader(reader *jxlio.Bitreader, parent *Frame) (*GlobalModular, error) {
	gm := &GlobalModular{}
	gm.frame = parent

	var err error
	hasGlobalTree := reader.MustReadBool()
	if hasGlobalTree {
		gm.globalTree, err = NewMATreeWithReader(reader)
		if err != nil {
			return nil, err
		}
	} else {
		gm.globalTree = nil
	}

	gm.frame.globalTree = gm.globalTree
	subModularChannelCount := len(gm.frame.globalMetadata.extraChannelInfo)
	header := gm.frame.header
	ecStart := 0
	if header.encoding == MODULAR {
		if !header.doYCbCr && !gm.frame.globalMetadata.xybEncoded && gm.frame.globalMetadata.colorEncoding.ColorEncoding == color.CE_GRAY {
			ecStart = 1
		} else {
			ecStart = 3
		}
	}
	subModularChannelCount += ecStart
	gm.stream, err = NewModularStreamWithReader(reader, gm.frame, 0, subModularChannelCount, ecStart)
	if err != nil {
		return nil, err
	}

	err = gm.stream.decodeChannels(reader, true)
	if err != nil {
		return nil, err
	}

	return gm, nil
}
