package core

import (
	"bytes"
	"image"
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

func ReadFileIntoMemory() JXLDecoderOption {
	return func(jxl *JXLDecoder) error {
		jxl.readFileIntoMemory = true
		return nil
	}
}

// JXLDecoder decodes the JXL image
type JXLDecoder struct {

	// input filename to read from.
	filename string

	// read entire JXL file into memory before processing
	readFileIntoMemory bool

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

	if jxl.readFileIntoMemory {
		f, err := os.ReadFile(jxl.filename)
		if err != nil {
			log.Errorf("Error opening file: %v\n", err)
			return nil
		}
		jxl.in = bytes.NewReader(f)
	} else {
		jxl.in = f
	}
	jxl.decoder = NewJXLCodestreamDecoder(jxl.in, nil)
	return jxl
}

func (jxl *JXLDecoder) Decode() (image.Image, error) {

	jxlImage, err := jxl.decoder.decode()
	if err != nil {
		return nil, err
	}

	return jxlImage, nil

}
