package bundle

import "github.com/kpfaulkner/jxl-go/jxlio"

type BitDepthHeader struct {
	UsesFloatSamples bool
	BitsPerSample    uint32
	ExpBits          uint32
}

func NewBitDepthHeader() *BitDepthHeader {
	bh := &BitDepthHeader{}
	bh.UsesFloatSamples = false
	bh.BitsPerSample = 8
	bh.ExpBits = 0
	return bh
}

func NewBitDepthHeaderWithReader(reader *jxlio.Bitreader) *BitDepthHeader {
	bh := &BitDepthHeader{}
	bh.UsesFloatSamples = reader.MustReadBool()
	if bh.UsesFloatSamples {
		bh.BitsPerSample = reader.MustReadU32(32, 0, 16, 0, 24, 0, 1, 6)
		bh.ExpBits = 1 + uint32(reader.MustReadBits(4))
	} else {
		bh.BitsPerSample = reader.MustReadU32(8, 0, 10, 0, 12, 0, 1, 6)
		bh.ExpBits = 0
	}
	return bh
}
