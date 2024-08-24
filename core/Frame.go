package core

import (
	"github.com/kpfaulkner/jxl-go/jxlio"
)

type Frame struct {
	globalMetadata *ImageHeader
	options        *JXLOptions
	reader         *jxlio.Bitreader
	header         *FrameHeader
}

func (f *Frame) readFrameHeader() {
	f.reader.ZeroPadToByte()
	f.header = NewFrameHeaderWithReader(f.reader, f.globalMetadata)
}

func NewFrameWithReader(reader *jxlio.Bitreader, imageHeader *ImageHeader, options *JXLOptions) *Frame {

	frame := &Frame{
		globalMetadata: imageHeader,
		options:        options,
		reader:         reader,
	}

	return frame
}
