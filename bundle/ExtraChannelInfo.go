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
	var err error
	var dAlpha bool
	if dAlpha, err = reader.ReadBool(); err != nil {
		return nil, err
	}
	if !dAlpha {
		if eci.EcType, err = reader.ReadEnum(); err != nil {
			return nil, err
		}
		if !ValidateExtraChannel(eci.EcType) {
			return nil, errors.New("Illegal extra channel type")
		}
		if bitDepth, err := NewBitDepthHeaderWithReader(reader); err != nil {
			return nil, err
		} else {
			eci.BitDepth = *bitDepth
		}

		if dimShift, err := reader.ReadU32(0, 0, 3, 0, 4, 0, 1, 3); err != nil {
			return nil, err
		} else {
			eci.DimShift = int32(dimShift)
		}
		var nameLen uint32
		var err error
		if nameLen, err = reader.ReadU32(0, 0, 0, 4, 16, 5, 48, 10); err != nil {
			return nil, err
		}

		nameBuffer := make([]byte, nameLen)
		for i := uint32(0); i < nameLen; i++ {
			nameBuffer[i] = byte(reader.MustReadBits(8))
		}
		eci.name = string(nameBuffer)

		var alphaBool bool
		if alphaBool, err = reader.ReadBool(); err != nil {
			return nil, err
		}
		eci.AlphaAssociated = (eci.EcType == ALPHA) && alphaBool
	} else {
		eci.EcType = ALPHA
		eci.BitDepth = *NewBitDepthHeader()
		eci.DimShift = 0
		eci.name = ""
		eci.AlphaAssociated = false
	}

	if eci.EcType == SPOT_COLOR {
		var err error
		if eci.red, err = reader.ReadF16(); err != nil {
			return nil, err
		}
		if eci.green, err = reader.ReadF16(); err != nil {
			return nil, err
		}
		if eci.blue, err = reader.ReadF16(); err != nil {
			return nil, err
		}
		if eci.solidity, err = reader.ReadF16(); err != nil {
			return nil, err
		}
	} else {
		eci.red = 0
		eci.green = 0
		eci.blue = 0
		eci.solidity = 0
	}

	if eci.EcType == COLOR_FILTER_ARRAY {
		if cfaIndex, err := reader.ReadU32(1, 0, 0, 2, 3, 4, 19, 8); err != nil {
			return nil, err
		} else {
			eci.cfaIndex = int32(cfaIndex)
		}
	} else {
		eci.cfaIndex = 1
	}
	return eci, nil
}
