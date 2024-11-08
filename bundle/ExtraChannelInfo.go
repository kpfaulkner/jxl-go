package bundle

import (
	"errors"

	"github.com/kpfaulkner/jxl-go/jxlio"
)

type ExtraChannelInfo struct {
	EcType                     int32
	BitDepth                   BitDepthHeader
	DimShift                   int32
	name                       string
	AlphaAssociated            bool
	red, green, blue, solidity float32
	cfaIndex                   int32
}

func NewExtraChannelInfoWithReader(reader *jxlio.Bitreader) (*ExtraChannelInfo, error) {
	eci := &ExtraChannelInfo{}

	dAlpha := reader.MustReadBool()
	if !dAlpha {
		eci.EcType = reader.MustReadEnum()
		if !ValidateExtraChannel(eci.EcType) {
			return nil, errors.New("Illegal extra channel type")
		}
		eci.BitDepth = *NewBitDepthHeaderWithReader(reader)
		eci.DimShift = int32(reader.MustReadU32(0, 0, 3, 0, 4, 0, 1, 3))
		nameLen := reader.MustReadU32(0, 0, 0, 4, 16, 5, 48, 10)
		nameBuffer := make([]byte, nameLen)
		for i := uint32(0); i < nameLen; i++ {
			nameBuffer[i] = byte(reader.MustReadBits(8))
		}
		eci.name = string(nameBuffer)
		eci.AlphaAssociated = (eci.EcType == ALPHA) && reader.MustReadBool()
	} else {
		eci.EcType = ALPHA
		eci.BitDepth = *NewBitDepthHeader()
		eci.DimShift = 0
		eci.name = ""
		eci.AlphaAssociated = false
	}

	if eci.EcType == SPOT_COLOR {
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

	if eci.EcType == COLOR_FILTER_ARRAY {
		eci.cfaIndex = int32(reader.MustReadU32(1, 0, 0, 2, 3, 4, 19, 8))
	} else {
		eci.cfaIndex = 1
	}
	return eci, nil
}
