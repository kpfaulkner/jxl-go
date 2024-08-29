package core

import "github.com/kpfaulkner/jxl-go/jxlio"

type BitDepthHeader struct {
	usesFloatSamples bool
	bitsPerSample    uint32
	expBits          uint32
}

func NewBitDepthHeader() *BitDepthHeader {
	bh := &BitDepthHeader{}
	bh.usesFloatSamples = false
	bh.bitsPerSample = 8
	bh.expBits = 0
	return bh
}

func NewBitDepthHeaderWithReader(reader *jxlio.Bitreader) *BitDepthHeader {
	bh := &BitDepthHeader{}
	bh.usesFloatSamples = reader.TryReadBool()
	if bh.usesFloatSamples {
		bh.bitsPerSample = reader.MustReadU32(32, 0, 16, 0, 24, 0, 1, 6)
		bh.expBits = 1 + uint32(reader.TryReadBits(4))
	} else {
		bh.bitsPerSample = reader.MustReadU32(8, 0, 10, 0, 12, 0, 1, 6)
		bh.expBits = 0
	}
	return bh
}
