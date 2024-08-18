package core

import (
	"io"

	"github.com/kpfaulkner/jxl-go/jxlio"
	log "github.com/sirupsen/logrus"
)

// JXLCodestreamDecoder decodes the JXL image
type JXLCodestreamDecoder struct {
	in io.ReadSeeker

	// bit reader... the actual thing that will read the bits/U16/U32/U64 etc.
	bitReader *jxlio.Bitreader
}

func NewJXLCodestreamDecoder(in io.ReadSeeker) *JXLCodestreamDecoder {
	jxl := &JXLCodestreamDecoder{}
	jxl.in = in
	jxl.bitReader = jxlio.NewBitreader(jxl.in)
	return jxl
}

func (jxl *JXLCodestreamDecoder) atEnd() bool {
	if jxl.bitReader != nil {
		return jxl.bitReader.AtEnd()
	}
	return false
}

func (jxl *JXLCodestreamDecoder) decode() error {

	if jxl.atEnd() {
		return nil
	}

	bits := jxl.bitReader.MustShowBits(16)
	log.Debugf("Initial bits %016b\n", bits)

	return nil
}
