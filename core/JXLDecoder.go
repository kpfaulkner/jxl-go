package core

import (
	"io"
	"os"

	log "github.com/sirupsen/logrus"
)

type JXLDecoderOption func(c *JXLDecoder) error

func WithInputFilename(fn string) JXLDecoderOption {
	return func(jxl *JXLDecoder) error {
		jxl.filename = fn
		return nil
	}
}

// JXLDecoder decodes the JXL image
type JXLDecoder struct {

	// input filename to read from.
	filename string

	// input stream
	in io.ReadSeeker

	// decoder
	decoder *JXLCodestreamDecoder
}

func NewJXLDecoder(opts ...JXLDecoderOption) *JXLDecoder {
	jxl := &JXLDecoder{}

	for _, opt := range opts {
		if err := opt(jxl); err != nil {
			panic("Error applying option to JXLDecoder: " + err.Error())
		}
	}

	f, err := os.Open(jxl.filename)
	if err != nil {
		log.Errorf("Error opening file: %v\n", err)
		return nil
	}

	jxl.in = f
	jxl.decoder = NewJXLCodestreamDecoder(jxl.in)
	return jxl
}

func (jxl *JXLDecoder) decode() error {

	return jxl.decoder.decode()
}
