package core

import (
	"errors"

	"github.com/kpfaulkner/jxl-go/bundle"
	"github.com/kpfaulkner/jxl-go/jxlio"
)

type ExtraChannelInfo struct {
	ecType                     int32
	bitDepth                   BitDepthHeader
	dimShift                   int32
	name                       string
	alphaAssociated            bool
	red, green, blue, solidity float32
	cfaIndex                   int32
}

func NewExtraChannelInfoWithReader(reader *jxlio.Bitreader) (*ExtraChannelInfo, error) {
	eci := &ExtraChannelInfo{}

	dAlpha := reader.TryReadBool()
	if dAlpha {
		eci.ecType = reader.MustReadEnum()
		if !bundle.ValidateExtraChannel(eci.ecType) {
			return nil, errors.New("Illegal extra channel type")
		}
		eci.bitDepth = *NewBitDepthHeaderWithReader(reader)
		eci.dimShift = int32(reader.MustReadU32(0, 0, 3, 0, 4, 0, 1, 3))
		nameLen := reader.MustReadU32(0, 0, 0, 4, 16, 5, 48, 10)
		nameBuffer := make([]byte, nameLen)
		for i := uint32(0); i < nameLen; i++ {
			nameBuffer[i] = byte(reader.TryReadBits(8))
		}
		eci.name = string(nameBuffer)
		eci.alphaAssociated = (eci.ecType == bundle.ALPHA) && reader.TryReadBool()
	} else {
		eci.ecType = bundle.ALPHA
		eci.bitDepth = *NewBitDepthHeader()
		eci.dimShift = 0
		eci.name = ""
		eci.alphaAssociated = false
	}

	if eci.ecType == bundle.SPOT_COLOR {
		eci.red = reader.MustReadF16()
		eci.green = reader.MustReadF16()
		eci.blue = reader.MustReadF16()
		eci.solidity = reader.MustReadF16()
	} else {
		eci.red = 0
		eci.green = 0
		eci.blue = 0
		eci.solidity = 0
	}

	if eci.ecType == bundle.COLOR_FILTER_ARRAY {
		eci.cfaIndex = int32(reader.MustReadU32(1, 0, 0, 2, 3, 4, 19, 8))
	} else {
		eci.cfaIndex = 1
	}
	return eci, nil
}
