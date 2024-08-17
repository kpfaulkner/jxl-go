package core

import (
	"errors"

	"github.com/kpfaulkner/jxl-go/bundle"
	"github.com/kpfaulkner/jxl-go/color"
	"github.com/kpfaulkner/jxl-go/jxlio"
)

const (
	CODESTREAM_HEADER uint32 = 0x0AFF
)

type ImageHeader struct {
	level           int32
	size            *SizeHeader
	orientation     uint32
	intrinsicSize   *SizeHeader
	previewHeader   *PreviewHeader
	animationHeader *AnimationHeader
	bitDepth        *BitDepthHeader

	orientedWidth       uint32
	orientedHeight      uint32
	modular16BitBuffers bool

	extraChannelInfo []ExtraChannelInfo
	xybEncoded       bool
	colorEncoding    *color.ColorEncodingBundle
	alphaIndices     []int32

	toneMapping *color.ToneMapping
	extensions  *bundle.Extensions
}

func NewImageHeader() *ImageHeader {
	ih := &ImageHeader{}
	return ih
}

func ParseImageHeader(reader *jxlio.Bitreader, level int32) (*ImageHeader, error) {
	header := NewImageHeader()

	if reader.MustReadBits(16) != CODESTREAM_HEADER {
		return nil, errors.New("Not a JXL codestream: 0xFF0A majoc mismatch")
	}

	err := header.setLevel(level)
	if err != nil {
		return nil, err
	}

	header.size, err = NewSizeHeader(reader, level)
	if err != nil {
		return nil, err
	}

	allDefault := reader.MustReadBool()
	extraFields := false
	if !allDefault {
		extraFields = reader.MustReadBool()
	}

	if extraFields {
		header.orientation = 1 + reader.MustReadBits(3)
		if reader.MustReadBool() {
			header.intrinsicSize, err = NewSizeHeader(reader, level)
			if err != nil {
				return nil, err
			}
		}
		if reader.MustReadBool() {
			header.previewHeader, err = NewPreviewHeader(reader)
			if err != nil {
				return nil, err
			}
		}
		if reader.MustReadBool() {
			header.animationHeader, err = NewAnimationHeader(reader)
			if err != nil {
				return nil, err
			}
		}
	} else {
		header.orientation = 1
	}

	if header.orientation > 4 {
		header.orientedWidth = header.size.height
		header.orientedHeight = header.size.width
	} else {
		header.orientedWidth = header.size.width
		header.orientedHeight = header.size.height
	}

	if allDefault {
		header.bitDepth = NewBitDepthHeader()
		header.modular16BitBuffers = true
		header.extraChannelInfo = []ExtraChannelInfo{}
		header.xybEncoded = true
		header.colorEncoding, err = color.NewColorEncodingBundle()
		if err != nil {
			return nil, err
		}
	} else {
		header.bitDepth = NewBitDepthHeaderWithReader(reader)
		header.modular16BitBuffers = reader.MustReadBool()
		extraChannelCount := reader.MustReadU32(0, 0, 1, 0, 2, 4, 1, 12)
		header.extraChannelInfo = make([]ExtraChannelInfo, extraChannelCount)
		alphaIndicies := make([]int32, extraChannelCount)
		numAlphaChannels := 0

		for i := 0; i < int(extraChannelCount); i++ {
			eci, err := NewExtraChannelInfoWithReader(reader)
			if err != nil {
				return nil, err
			}
			header.extraChannelInfo[i] = *eci

			if header.extraChannelInfo[i].ecType == bundle.ALPHA {
				alphaIndicies[numAlphaChannels] = int32(i)
				numAlphaChannels++
			}
		}
		header.alphaIndices = make([]int32, numAlphaChannels)
		copy(header.alphaIndices, alphaIndicies[:numAlphaChannels])
		header.xybEncoded = reader.MustReadBool()
		header.colorEncoding, err = color.NewColorEncodingBundleWithReader(reader)
		if err != nil {
			return nil, err
		}
	}

	if extraFields {
		header.toneMapping, err = color.NewToneMappingWithReader(reader)
		if err != nil {
			return nil, err
		}
	} else {
		header.toneMapping = color.NewToneMapping()
	}

	if allDefault {
		header.extensions = bundle.NewExtensions()
	} else {
		header.extensions, err = bundle.NewExtensionsWithReader(reader)
		if err != nil {
			return nil, err
		}
	}
	
	return header, nil
}

func (h *ImageHeader) setLevel(level int32) error {
	if level != 5 && level != 10 {
		return errors.New("invalid bitstream")
	}
	h.level = level
	return nil
}
