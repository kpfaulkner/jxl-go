package core

import "github.com/kpfaulkner/jxl-go/jxlio"

type LFChannelCorrelation struct {
	colorFactor      uint32
	baseCorrelationX float32
	baseCorrelationB float32
	xFactorLF        uint32
	bFactorLF        uint32
}

func NewLFChannelCorrelation() (*LFChannelCorrelation, error) {
	NewLFChannelCorrelationWithReaderAndDefault(nil, true)
}

func NewLFChannelCorrelationWithReaderAndDefault(reader *jxlio.Bitreader, allDefault bool) (*LFChannelCorrelation, error) {
	lf := &LFChannelCorrelation{}

	if allDefault {
		lf.colorFactor = 84
		lf.baseCorrelationX = 0.0
		lf.baseCorrelationB = 1.0
		lf.xFactorLF = 128
		lf.bFactorLF = 128
	} else {
		panic("NewLFChannelCorrelationWithReaderAndDefault not implemented yet")
	}
	return lf, nil
}

func NewLFChannelCorrelationWithReader(reader *jxlio.Bitreader) (*LFChannelCorrelation, error) {
	return NewLFChannelCorrelationWithReaderAndDefault(reader, reader.MustReadBool())
}
