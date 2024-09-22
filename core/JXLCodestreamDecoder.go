package core

import (
	"errors"
	"fmt"
	"io"

	"github.com/kpfaulkner/jxl-go/bundle"
	"github.com/kpfaulkner/jxl-go/color"
	"github.com/kpfaulkner/jxl-go/jxlio"
	"github.com/kpfaulkner/jxl-go/util"
)

var (

	// signature for JXL container
	CONTAINER_SIGNATURE = []byte{0x00, 0x00, 0x00, 0x0C, 'J', 'X', 'L', ' ', 0x0D, 0x0A, 0x87, 0x0A}
)

type JXLOptions struct {
	debug           bool
	parseOnly       bool
	renderVarblocks bool
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
		invisibleFrames := int64(0)
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

			err = frame.initializeNoise(int64(visibleFrames<<32) | invisibleFrames)
			if err != nil {
				return err
			}
			err = frame.upsample()
			if err != nil {
				return err
			}

			if save && header.saveBeforeCT {
				reference[header.saveAsReference] = frame.buffer
			}

			err = jxl.computePatches(reference, frame)
			if err != nil {
				return err
			}

			err = frame.renderSplines()
			if err != nil {
				return err
			}

			err = frame.synthesizeNoise()
			if err != nil {
				return err
			}

			err = jxl.performColourTransforms(matrix, frame)
			if err != nil {
				return err
			}

			if header.encoding == VARDCT && jxl.options.renderVarblocks {
				panic("VARDCT not implemented yet")
			}

			if header.frameType == REGULAR_FRAME || header.frameType == SKIP_PROGRESSIVE {
				found := false
				for i := uint32(0); i < 4; i++ {
					if util.Matrix3Equal(reference[i], canvas) && i != header.saveAsReference {
						found = true
						break
					}
				}
				if found {
					canvas = util.DeepCopy3[float32](canvas)
				}
				err = jxl.blendFrame(canvas, reference, frame)
				if err != nil {
					return err
				}
			}

			if save && !header.saveBeforeCT {
				reference[header.saveAsReference] = canvas
			}

			if header.isLast {
				break
			}
		}

		err = jxl.bitReader.ZeroPadToByte()
		if err != nil {
			return err
		}

		// TOOD(kpfaulkner) unsure if need to perform similar drain cache functionality here. Don't think we do.

		if jxl.options.parseOnly {
			return nil
		}

		orientation := imageHeader.orientation
		orientedCanvas := util.MakeMatrix3D[float32](len(canvas), 0, 0)
		for i := 0; i < len(orientedCanvas); i++ {
			orientedCanvas[i], err = jxl.transposeBuffer(canvas[i], orientation)
			if err != nil {
				return err
			}
		}
	}

	panic("make JXL image here?")
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

func (jxl *JXLCodestreamDecoder) computePatches(references [][][][]float32, frame *Frame) error {

	header := frame.header
	frameBuffer := frame.buffer
	colourChannels := jxl.imageHeader.getColourChannelCount()
	extraChannels := len(jxl.imageHeader.extraChannelInfo)
	patches := frame.lfGlobal.patches
	for i := 0; i < len(patches); i++ {
		patch := patches[i]
		if patch.ref > 3 {
			return errors.New("patch out of range")
		}
		refBuffer := references[patch.ref]
		if refBuffer == nil || len(refBuffer) == 0 {
			continue
		}
		if patch.height+int32(patch.origin.Y) > int32(len(refBuffer[0])) || patch.width+int32(patch.origin.X) > int32(len(refBuffer[0][0])) {
			return errors.New("patch too large")
		}
		for j := 0; i < len(patch.positions); j++ {
			x0 := patch.positions[j].X
			y0 := patch.positions[j].Y
			if x0 < 0 || y0 < 0 {
				return errors.New("patch size out of bounds")
			}

			if patch.height+int32(y0) > int32(header.height) || patch.width+int32(x0) > int32(header.width) {
				return errors.New("patch size out of bounds")
			}

			for d := int32(0); d < int32(colourChannels)+int32(extraChannels); d++ {
				var c int32
				if d < int32(colourChannels) {
					c = 0
				} else {
					c = d - int32(colourChannels) + 1
				}
				info := patch.blendingInfos[j][c]
				if info.mode == 0 {
					continue
				}
				var premult bool
				if jxl.imageHeader.hasAlpha() {
					premult = jxl.imageHeader.extraChannelInfo[info.alphaChannel].alphaAssociated
				} else {
					premult = true
				}
				isAlpha := c > 0 && jxl.imageHeader.extraChannelInfo[c-1].ecType == bundle.ALPHA
				if info.mode > 0 && header.upsampling > 1 && c > 0 && header.ecUpsampling[c-1]<<jxl.imageHeader.extraChannelInfo[c-1].dimShift != header.upsampling {
					return errors.New("Alpha channel upsampling mismatch during patches")
				}
				for y := int32(0); y < patch.height; y++ {
					for x := int32(0); x < patch.width; x++ {
						oldX := x + int32(x0)
						oldY := y + int32(y0)
						newX := x + int32(patch.origin.X)
						newY := y + int32(patch.origin.Y)
						oldSample := frameBuffer[d][oldY][oldX]
						newSample := refBuffer[d][newY][newX]
						alpha := float32(0.0)
						newAlpha := float32(0.0)
						oldAlpha := float32(0.0)
						if info.mode > 3 {
							if jxl.imageHeader.hasAlpha() {
								oldAlpha = frameBuffer[uint32(colourChannels)+info.alphaChannel][oldY][oldX]
							} else {
								oldAlpha = 1.0
							}
							if jxl.imageHeader.hasAlpha() {
								newAlpha = refBuffer[uint32(colourChannels)+info.alphaChannel][newY][newX]
							} else {
								newAlpha = 1.0
							}
							if info.clamp {
								newAlpha = util.Clamp3Float32(newAlpha, 0.0, 1.0)
							}
							if info.mode < 6 || !isAlpha || !premult {
								alpha = oldAlpha + newAlpha*(1-oldAlpha)
							}

							var sample float32
							switch info.mode {
							case 0:
								sample = oldSample
								break
							case 1:
								sample = newSample
								break
							case 2:
								sample = oldSample + newSample
								break
							case 3:
								sample = oldSample * newSample
								break
							case 4:
								if isAlpha {
									sample = float32(alpha)
								} else {
									if premult {
										sample = newSample + oldSample*(1-newAlpha)
									} else {
										sample = (newSample*newAlpha + oldSample*oldAlpha*(1-newAlpha)) / float32(alpha)
									}
								}
								break
							case 5:
								if isAlpha {
									sample = float32(alpha)
								} else {
									if premult {
										sample = oldSample + newSample*(1-newAlpha)
									} else {
										sample = (oldSample*newAlpha + newSample*oldAlpha*(1-newAlpha)) / float32(alpha)
									}
								}
								break
							case 6:
								if isAlpha {
									sample = newAlpha
								} else {
									sample = oldSample + float32(alpha)*newSample
								}
								break
							case 7:
								if isAlpha {
									sample = oldAlpha
								} else {
									sample = newSample + float32(alpha)*oldSample
								}
								break
							default:
								return errors.New("Challenge complete how did we get here")
							}
							frameBuffer[d][oldY][oldX] = sample
						}
					}
				}

			}
		}
	}
	return nil
}

func (jxl *JXLCodestreamDecoder) performColourTransforms(matrix *color.OpsinInverseMatrix, frame *Frame) error {
	frameBuffer := frame.buffer
	if matrix != nil {
		err := matrix.InvertXYB(frameBuffer, jxl.imageHeader.toneMapping.GetIntensityTarget())
		if err != nil {
			return err
		}
	}

	if frame.header.doYCbCr {
		size, err := frame.getPaddedFrameSize()
		if err != nil {
			return err
		}
		for y := uint32(0); y < size.Y; y++ {
			for x := uint32(0); x < size.X; x++ {
				cb := frameBuffer[0][y][x]
				yh := frameBuffer[1][y][x] + 0.50196078431372549019
				cr := frameBuffer[2][y][x]
				frameBuffer[0][y][x] = yh + 1.402*cr
				frameBuffer[1][y][x] = yh - 0.34413628620102214650*cb - 0.71413628620102214650*cr
				frameBuffer[2][y][x] = yh + 1.772*cb
			}
		}
	}
	return nil
}

func (jxl *JXLCodestreamDecoder) blendFrame(canvas [][][]float32, reference [][][][]float32, frame *Frame) error {

	width := jxl.imageHeader.size.width
	height := jxl.imageHeader.size.height
	header := frame.header
	frameStart := header.origin.Max(util.ZERO)
	frameSize := util.NewIntPointWithXY(width, height).Min(header.origin.Plus(util.NewIntPointWithXY(header.width, header.height))).Minus(frameStart)
	frameColours := frame.getColorChannelCount()
	imageColours := jxl.imageHeader.getColourChannelCount()
	for c := int32(0); c < int32(len(canvas)); c++ {
		var frameC int32
		if frameColours != imageColours {
			if c == 0 {
				frameC = 1
			} else {
				frameC = c + 2
			}
		} else {
			frameC = c
		}
		frameBuffer := frame.buffer[frameC]
		var info *BlendingInfo
		if frameC < int32(frameColours) {
			info = frame.header.blendingInfo
		} else {
			info = &frame.header.ecBlendingInfo[frameC-int32(frameColours)]
		}
		isAlpha := c >= int32(imageColours) && jxl.imageHeader.extraChannelInfo[c-int32(imageColours)].ecType == bundle.ALPHA
		var premult bool
		if jxl.imageHeader.hasAlpha() {
			premult = jxl.imageHeader.extraChannelInfo[info.alphaChannel].alphaAssociated
		} else {
			premult = true
		}
		refBuffer := reference[info.source]
		if info.mode == BLEND_REPLACE || refBuffer == nil && info.mode == BLEND_ADD {
			jxl.copyToCanvas(canvas[c], frameStart, frameStart.Minus(header.origin), frameSize, frameBuffer)
			continue
		}
		ref := refBuffer[c]
		switch info.mode {
		case BLEND_ADD:
			for y := uint32(0); y < frameSize.Y; y++ {
				cy := y + frameStart.Y
				for x := uint32(0); x < frameSize.X; x++ {
					cx := x + frameStart.X
					canvas[c][cy][cx] = ref[cy][cx] + frameBuffer[y][x]
				}
			}
			break
		case BLEND_MULT:
			for y := uint32(0); y < frameSize.Y; y++ {
				cy := y + frameStart.Y
				if ref != nil {
					for x := uint32(0); x < frameSize.X; x++ {
						cx := x + frameStart.X
						newSample := frameBuffer[y][x]
						if info.clamp {
							newSample = util.Clamp3Float32(newSample, 0.0, 1.0)
						}
						canvas[c][cy][cx] = newSample * ref[cy][cx]
					}
				} else {
					util.FillFloat32(canvas[c][cy], frameStart.X, frameSize.X, 0.0)
				}
			}
			break
		case BLEND_BLEND:
			for cy := frameStart.Y; cy < frameSize.Y+frameStart.Y; cy++ {
				for cx := frameStart.X; cx < frameSize.X+frameStart.X; cx++ {
					var oldAlpha float32

					if jxl.imageHeader.hasAlpha() {
						oldAlpha = 1.0
					} else {
						if ref != nil {
							oldAlpha = refBuffer[imageColours+int(info.alphaChannel)][cy][cx]
						} else {
							oldAlpha = 0.0
						}
					}
					var newAlpha float32
					if jxl.imageHeader.hasAlpha() {
						newAlpha = 1.0
					} else {
						newAlpha = frame.getImageSample(uint32(frameColours)+info.alphaChannel, cx, cy)
					}

					if info.clamp {
						newAlpha = util.Clamp3Float32(newAlpha, 0.0, 1.0)
					}
					var alpha = float32(1.0)
					var oldSample float32
					if ref != nil {
						oldSample = ref[cy][cx]
					} else {
						oldSample = 0.0
					}
					newSample := frame.getImageSample(uint32(frameC), cx, cy)
					if isAlpha || !premult {
						alpha = oldAlpha + newAlpha*(1-oldAlpha)
					}
					if isAlpha {
						canvas[c][cy][cx] = alpha
					} else if premult {
						canvas[c][cy][cx] = newSample + oldSample*(1-newAlpha)
					} else {
						canvas[c][cy][cx] = (newSample*newAlpha + oldSample*oldAlpha*(1-newAlpha)) / alpha
					}
				}
			}
			break
		case BLEND_MULADD:
			for cy := frameStart.Y; cy < frameSize.Y+frameStart.Y; cy++ {
				for cx := frameStart.X; cx < frameSize.X+frameStart.X; cx++ {
					var oldAlpha float32
					if !jxl.imageHeader.hasAlpha() {
						oldAlpha = 1.0
					} else {
						if ref != nil {
							oldAlpha = refBuffer[imageColours+int(info.alphaChannel)][cy][cx]
						} else {
							oldAlpha = 0.0
						}
					}
					var newAlpha float32
					if !jxl.imageHeader.hasAlpha() {
						newAlpha = 1.0
					} else {
						newAlpha = frame.getImageSample(uint32(frameColours)+info.alphaChannel, cx, cy)
					}

					if info.clamp {
						newAlpha = util.Clamp3Float32(newAlpha, 0.0, 1.0)
					}
					var oldSample float32
					if ref != nil {
						oldSample = ref[cy][cx]
					} else {
						oldSample = 0.0
					}
					newSample := frame.getImageSample(uint32(frameC), cx, cy)
					if isAlpha {
						canvas[c][cy][cx] = oldAlpha
					} else {
						canvas[c][cy][cx] = oldSample + newSample*newAlpha
					}

				}
			}
			break
		default:
			return errors.New("Illegal blend mode")
		}
	}
	return nil
}

// FIXME(kpfaulkner) really unsure about this
func (jxl *JXLCodestreamDecoder) copyToCanvas(canvas [][]float32, start util.IntPoint, off util.IntPoint, size util.IntPoint, frameBuffer [][]float32) {
	for y := uint32(0); y < size.Y; y++ {
		copy(canvas[y+start.X][off.X:], frameBuffer[y+off.Y][off.X:off.X+size.X])
	}
}

func (jxl *JXLCodestreamDecoder) transposeBuffer(src [][]float32, orientation uint32) ([][]float32, error) {

	size := util.IntPointSizeOf(src)
	var dest [][]float32
	if orientation > 4 {
		dest = util.MakeMatrix2D[float32](int(size.X), int(size.Y))
	} else if orientation > 1 {
		dest = util.MakeMatrix2D[float32](int(size.Y), int(size.X))
	} else {
		dest = nil
	}

	switch orientation {
	case 1:
		return src, nil
	case 2:
		transposeIter := util.RangeIteratorWithIntPoint(size)
		for {
			p, err := transposeIter()
			if err == io.EOF {
				break
			}
			dest[p.Y][size.X-1-p.X] = src[p.X][p.Y]
		}
	case 3:
		transposeIter := util.RangeIteratorWithIntPoint(size)
		for {
			p, err := transposeIter()
			if err == io.EOF {
				break
			}
			dest[size.Y-1-p.Y][size.X-1-p.X] = src[p.X][p.Y]
		}
	case 4:
		transposeIter := util.RangeIteratorWithIntPoint(size)
		for {
			p, err := transposeIter()
			if err == io.EOF {
				break
			}
			dest[size.Y-1-p.Y][p.X] = src[p.X][p.Y]
		}
	case 5:
		transposeIter := util.RangeIteratorWithIntPoint(size)
		for {
			p, err := transposeIter()
			if err == io.EOF {
				break
			}
			dest[p.X][p.Y] = src[p.Y][p.X]
		}
	case 6:
		transposeIter := util.RangeIteratorWithIntPoint(size)
		for {
			p, err := transposeIter()
			if err == io.EOF {
				break
			}
			dest[p.X][size.Y-1-p.Y] = src[p.Y][p.X]
		}
	case 7:
		transposeIter := util.RangeIteratorWithIntPoint(size)
		for {
			p, err := transposeIter()
			if err == io.EOF {
				break
			}
			dest[size.X-1-p.X][size.Y-1-p.Y] = src[p.Y][p.X]
		}
	case 8:
		transposeIter := util.RangeIteratorWithIntPoint(size)
		for {
			p, err := transposeIter()
			if err == io.EOF {
				break
			}
			dest[size.X-1-p.X][p.Y] = src[p.Y][p.X]
		}
	default:
		return nil, errors.New("Invalid orientation")

	}

	return nil, nil
}
