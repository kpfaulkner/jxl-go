package frame

import (
	"errors"
	"fmt"

	"github.com/kpfaulkner/jxl-go/jxlio"
	"github.com/kpfaulkner/jxl-go/util"
)

type HFMetadata struct {
	nbBlocks       uint64
	dctSelect      [][]TransformType
	hfMultiplier   [][]int32
	hfStreamBuffer [][][]int32
	parent         *LFGroup
	blockList      []util.Point
}

func NewHFMetadataWithReader(reader *jxlio.Bitreader, parent *LFGroup, frame *Frame) (*HFMetadata, error) {
	hf := &HFMetadata{
		parent: parent,
	}

	n := util.CeilLog2(parent.size.Height * parent.size.Width)
	nbBlocks, err := reader.ReadBits(n)
	if err != nil {
		return nil, err
	}
	hf.nbBlocks = nbBlocks + 1

	correlationHeight := int32((parent.size.Height + 7) / 8)
	correlationWidth := int32((parent.size.Width + 7) / 8)
	xFromY := NewModularChannelWithAllParams(correlationWidth, correlationHeight, 0, 0, false)
	bFromY := NewModularChannelWithAllParams(correlationWidth, correlationHeight, 0, 0, false)
	blockInfo := NewModularChannelWithAllParams(2, int32(hf.nbBlocks), 0, 0, false)
	sharpness := NewModularChannelWithAllParams(int32(parent.size.Height), int32(parent.size.Width), 0, 0, false)
	hfStream, err := NewModularStreamWithStreamIndex(reader, frame, 1+2*int(frame.numLFGroups)+int(parent.lfGroupID), []ModularChannel{*xFromY, *bFromY, *blockInfo, *sharpness})
	if err != nil {
		return nil, err
	}
	err = hfStream.decodeChannels(reader, false)
	if err != nil {
		return nil, err
	}

	hf.hfStreamBuffer = hfStream.getDecodedBuffer()
	hf.dctSelect = util.MakeMatrix2D[TransformType](parent.size.Height, parent.size.Width)
	hf.hfMultiplier = util.MakeMatrix2D[int32](parent.size.Height, parent.size.Width)
	blockInfoBuffer := hf.hfStreamBuffer[2]
	lastBlock := util.Point{X: 0, Y: 0}
	tta := allDCT
	hf.blockList = make([]util.Point, hf.nbBlocks)
	for i := uint64(0); i < hf.nbBlocks; i++ {
		t := blockInfoBuffer[0][i]
		if t > 26 || t < 0 {
			return nil, errors.New(fmt.Sprintf("Invalid Transform Type %d", t))
		}
		tt := tta[t]
		pos := hf.placeBlock(lastBlock, tt, 1+blockInfoBuffer[1][i])
		lastBlock = util.Point{
			X: pos.X,
			Y: pos.Y,
		}
	}
	return hf, nil
}

func (m HFMetadata) placeBlock(block util.Point, tt TransformType, i int32) util.Point {

	panic("not implemented")
}
