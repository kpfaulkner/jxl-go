package frame

import (
	"github.com/kpfaulkner/jxl-go/colour"
	"github.com/kpfaulkner/jxl-go/jxlio"
)

type GlobalModular struct {
	frame      *Frame
	globalTree *MATree
	Stream     *ModularStream
}

func NewGlobalModularWithReader(reader jxlio.BitReader, parent *Frame) (*GlobalModular, error) {
	gm := &GlobalModular{}
	gm.frame = parent

	// Have confirmed that java and Go MATrees are identical (comparing property/value/offset etc).
	var err error
	var hasGlobalTree bool
	if hasGlobalTree, err = reader.ReadBool(); err != nil {
		return nil, err
	}
	if hasGlobalTree {
		gm.globalTree, err = NewMATreeWithReader(reader)
		if err != nil {
			return nil, err
		}
	} else {
		gm.globalTree = nil
	}

	gm.frame.globalTree = gm.globalTree
	subModularChannelCount := len(gm.frame.globalMetadata.ExtraChannelInfo)
	header := gm.frame.Header
	ecStart := 0
	if header.Encoding == MODULAR {
		if !header.DoYCbCr && !gm.frame.globalMetadata.XybEncoded && gm.frame.globalMetadata.ColourEncoding.ColourEncoding == colour.CE_GRAY {
			ecStart = 1
		} else {
			ecStart = 3
		}
	}
	subModularChannelCount += ecStart
	gm.Stream, err = NewModularStreamWithReader(reader, gm.frame, 0, subModularChannelCount, ecStart)
	if err != nil {
		return nil, err
	}

	err = gm.Stream.decodeChannels(reader, true)
	if err != nil {
		return nil, err
	}

	return gm, nil
}
