package core

import (
	"errors"
	"fmt"
	"io"

	"github.com/kpfaulkner/jxl-go/color"
	"github.com/kpfaulkner/jxl-go/jxlio"
	"github.com/kpfaulkner/jxl-go/util"
	log "github.com/sirupsen/logrus"
)

var (

	// signature for JXL container
	CONTAINER_SIGNATURE = []byte{0x00, 0x00, 0x00, 0x0C, 'J', 'X', 'L', ' ', 0x0D, 0x0A, 0x87, 0x0A}
)

type JXLOptions struct {
	debug     bool
	parseOnly bool
}

func NewJXLOptions(options *JXLOptions) *JXLOptions {

	opt := &JXLOptions{}
	if options != nil {
		opt.debug = options.debug
		opt.parseOnly = options.parseOnly
	}
	return opt
}

// Box information (not sure what this is yet)
type BoxInfo struct {
	boxSize   uint32
	posInBox  uint32
	container bool
}

// JXLCodestreamDecoder decodes the JXL image
type JXLCodestreamDecoder struct {
	in io.ReadSeeker

	// bit reader... the actual thing that will read the bits/U16/U32/U64 etc.
	bitReader *jxlio.Bitreader

	foundSignature bool
	boxHeaders     []ContainerBoxHeader
	level          int
	imageHeader    *ImageHeader
	options        JXLOptions
}

func NewJXLCodestreamDecoder(in io.ReadSeeker, options *JXLOptions) *JXLCodestreamDecoder {
	jxl := &JXLCodestreamDecoder{}
	jxl.in = in
	jxl.bitReader = jxlio.NewBitreader(jxl.in, true)
	jxl.foundSignature = false

	if options != nil {
		jxl.options = *options
	}
	return jxl
}

func (jxl *JXLCodestreamDecoder) atEnd() bool {
	if jxl.bitReader != nil {
		return jxl.bitReader.AtEnd()
	}
	return false
}

func (jxl *JXLCodestreamDecoder) decode() error {

	// read header to get signature
	_, err := jxl.readSignatureAndBoxes()
	if err != nil {
		return err
	}

	// loop through each box.
	// first thing is to set the BitReader to the beginning of the data for that box.
	for _, box := range jxl.boxHeaders {
		_, err := jxl.bitReader.Seek(box.Offset, io.SeekStart)
		if err != nil {
			return err
		}

		if jxl.atEnd() {
			return nil
		}

		// Read the actual data to process.

		//b, _ := jxl.bitReader.ReadByteArray(10)
		//fmt.Printf("b is %x\n", b)

		sb, err := jxl.bitReader.ShowBits(16)
		if err != nil {
			return err
		}

		fmt.Printf("show bits is %d\n", sb)

		level := int32(jxl.level)
		imageHeader, err := ParseImageHeader(jxl.bitReader, level)
		if err != nil {
			return err
		}

		jxl.imageHeader = imageHeader
		fmt.Printf("imageheader %+v\n", *imageHeader)
		//gray := imageHeader.getColourChannelCount() < 3
		//alpha := imageHeader.hasAlpha()
		//ce := imageHeader.colorEncoding

		if imageHeader.animationHeader != nil {
			panic("dont care about animation for now")
		}

		if imageHeader.previewHeader != nil {
			previewOptions := NewJXLOptions(&jxl.options)
			previewOptions.parseOnly = true
			frame := NewFrameWithReader(jxl.bitReader, jxl.imageHeader, previewOptions)
			frame.readFrameHeader()
			panic("not implemented previewheader yet")
		}

		frameCount := 0
		reference := make([][][][]float32, 4)
		header := FrameHeader{}
		lfBuffer := make([][][][]float32, 5)

		var matrix *color.OpsinInverseMatrix
		if imageHeader.xybEncoded {
			bundle := imageHeader.colorEncoding
			matrix, err = imageHeader.opsinInverseMatrix.GetMatrix(bundle.Prim, bundle.White)
			if err != nil {
				return err
			}
		}

		var canvas [][][]float32
		if !jxl.options.parseOnly {
			canvas = util.MakeMatrix3D[float32](imageHeader.getColourChannelCount()+len(imageHeader.extraChannelInfo), int(imageHeader.size.height), int(imageHeader.size.width))
		}
		invisibleFrames := 0
		visibleFrames := 0

		// XXXXXXXXXXX JXLCodestreamDecoder line 337
		for {
			frame := NewFrameWithReader(jxl.bitReader, jxl.imageHeader, &jxl.options)
			header, err = frame.readFrameHeader()
			if err != nil {
				return err
			}
			frameCount++

			if lfBuffer[header.lfLevel] == nil && header.flags&USE_LF_FRAME != 0 {
				return errors.New("LF level too large")
			}
			if jxl.options.parseOnly {
				frame.skipFrameData()
				continue
			}
			err := frame.decodeFrame(lfBuffer[header.lfLevel])
			if err != nil {
				return err
			}

			if header.lfLevel > 0 {
				lfBuffer[header.lfLevel-1] = frame.buffer
			}
			save := (header.saveAsReference != 0 || header.duration == 0) && !header.isLast && header.frameType != LF_FRAME
			if frame.isVisible() {
				visibleFrames++
				invisibleFrames = 0
			} else {
				invisibleFrames++
			}

			// TODO(kpfaulkner)
			panic("not implemented yet")

		}

		fmt.Printf("XXXX %v %v %v %v %v %v\n", reference, header, lfBuffer, matrix, canvas, invisibleFrames, visibleFrames)

	}

	bits := jxl.bitReader.MustShowBits(16)
	log.Debugf("Initial bits %016b\n", bits)

	return nil
}

// Read signature
// See Demuxer.java supplyExceptionally()
func (jxl *JXLCodestreamDecoder) readSignatureAndBoxes() ([]byte, error) {

	br := NewBoxReader(jxl.bitReader)
	boxHeaders, err := br.ReadBoxHeader()
	if err != nil {
		return nil, err
	}

	jxl.boxHeaders = boxHeaders
	jxl.level = br.level
	//if !jxl.foundSignature {
	//	signature := make([]byte, 12)
	//	remaining, err := jxlio.ReadFully(jxl.in, signature)
	//	if err != nil {
	//		return nil, err
	//	}
	//
	//	// if not equall... just return the dodgy signature
	//	if bytes.Compare(signature, CONTAINER_SIGNATURE) != 0 {
	//		if remaining != 0 {
	//			signature = signature[:len(signature)-remaining]
	//		}
	//		jxl.boxInfo.boxSize = 0
	//		jxl.boxInfo.posInBox = len(signature)
	//		jxl.boxInfo.container = false
	//		return signature, err
	//	} else {
	//		jxl.boxInfo.boxSize = 12
	//		jxl.boxInfo.posInBox = 12
	//		jxl.boxInfo.container = true
	//	}
	//}

	//if !jxl.boxInfo.container || jxl.boxInfo.boxSize > 0 && jxl.boxInfo.posInBox < jxl.boxInfo.boxSize || jxl.boxInfo.boxSize == 0 {
	//	l := uint32(4096)
	//
	//	if jxl.boxInfo.boxSize > 0 && jxl.boxInfo.boxSize-jxl.boxInfo.posInBox < l {
	//		l = min(math.MaxUint32, uint32(jxl.boxInfo.boxSize-jxl.boxInfo.posInBox))
	//	}
	//	buf := make([]byte, l)
	//	remaining, err := jxlio.ReadFully(jxl.in, buf)
	//	if err != nil {
	//		return nil, err
	//	}
	//	jxl.boxInfo.posInBox += l - uint32(remaining)
	//	if remaining > 0 {
	//		if uint32(remaining) == l {
	//			return []byte{}, nil
	//		}
	//
	//	}
	//
	//	return nil, nil
	//}

	return nil, nil
}
