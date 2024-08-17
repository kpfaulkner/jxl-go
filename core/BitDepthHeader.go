package core

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
