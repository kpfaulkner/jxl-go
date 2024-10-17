package core

import (
	"errors"
	"fmt"
	"image"
	"io"

	"github.com/kpfaulkner/jxl-go/bundle"
	"github.com/kpfaulkner/jxl-go/color"
	"github.com/kpfaulkner/jxl-go/frame"
	image2 "github.com/kpfaulkner/jxl-go/image"
	"github.com/kpfaulkner/jxl-go/jxlio"
	"github.com/kpfaulkner/jxl-go/options"
	"github.com/kpfaulkner/jxl-go/util"
)

// Box information (not sure what this is yet)
type BoxInfo struct {
	boxSize   uint32
	posInBox  uint32
	container bool
}

// JXLCodestreamDecoder decodes the JXL image
type JXLCodestreamDecoder struct {
	// bit reader... the actual thing that will read the bits/U16/U32/U64 etc.
	bitReader *jxlio.Bitreader

	foundSignature bool
	boxHeaders     []ContainerBoxHeader
	level          int
	imageHeader    *bundle.ImageHeader
	options        options.JXLOptions
	reference      [][]image2.ImageBuffer
	lfBuffer       [][]image2.ImageBuffer
	canvas         []image2.ImageBuffer
}

func NewJXLCodestreamDecoder(br *jxlio.Bitreader, options *options.JXLOptions) *JXLCodestreamDecoder {
	jxl := &JXLCodestreamDecoder{}
	jxl.bitReader = br
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

// GetImageHeader just duplicates the first chunk of code from decode. This is so we can get the image size
// and colour model.
func (jxl *JXLCodestreamDecoder) GetImageHeader() (*bundle.ImageHeader, error) {

	// read header to get signature
	err := jxl.readSignatureAndBoxes()
	if err != nil {
		return nil, err
	}

	for _, box := range jxl.boxHeaders {
		_, err := jxl.bitReader.Seek(box.Offset, io.SeekStart)
		if err != nil {
			return nil, err
		}

		if jxl.atEnd() {
			return nil, nil
		}

		level := int32(jxl.level)
		imageHeader, err := bundle.ParseImageHeader(jxl.bitReader, level)
		if err != nil {
			return nil, err
		}

		return imageHeader, nil
	}
	return nil, errors.New("unable to find image header")
}

func (jxl *JXLCodestreamDecoder) decode() (image.Image, error) {

	// read header to get signature
	err := jxl.readSignatureAndBoxes()
	if err != nil {
		return nil, err
	}

	// loop through each box.
	// first thing is to set the BitReader to the beginning of the data for that box.
	// FIXME(kpfaulkner) need to figure out how to handle multiple boxes.
	for _, box := range jxl.boxHeaders {
		_, err := jxl.bitReader.Seek(box.Offset, io.SeekStart)
		if err != nil {
			return nil, err
		}

		if jxl.atEnd() {
			return nil, nil
		}

		var show uint64
		show, err = jxl.bitReader.ShowBits(32)
		fmt.Printf("show %d\n", show)
		if err != nil {
			return nil, err
		}

		level := int32(jxl.level)
		imageHeader, err := bundle.ParseImageHeader(jxl.bitReader, level)
		if err != nil {
			return nil, err
		}

		jxl.imageHeader = imageHeader
		if imageHeader.AnimationHeader != nil {
			panic("dont care about animation for now")
		}

		if imageHeader.PreviewSize != nil {
			previewOptions := options.NewJXLOptions(&jxl.options)
			previewOptions.ParseOnly = true
			frame := frame.NewFrameWithReader(jxl.bitReader, jxl.imageHeader, previewOptions)
			frame.ReadFrameHeader()
			panic("not implemented previewheader yet")
		}

		frameCount := 0
		//reference := make([][][][]float32, 4)
		header := frame.FrameHeader{}
		//lfBuffer := make([][][][]float32, 5)

		var matrix *color.OpsinInverseMatrix
		if imageHeader.XybEncoded {
			bundle := imageHeader.ColorEncoding
			matrix, err = imageHeader.OpsinInverseMatrix.GetMatrix(bundle.Prim, bundle.White)
			if err != nil {
				return nil, err
			}
		}

		var canvas []image2.ImageBuffer
		if !jxl.options.ParseOnly {
			canvas = util.MakeMatrix3D[float32](imageHeader.GetColourChannelCount()+len(imageHeader.ExtraChannelInfo), int(imageHeader.Size.Height), int(imageHeader.Size.Width))
		}
		invisibleFrames := int64(0)
		visibleFrames := 0

		for {
			imgFrame := frame.NewFrameWithReader(jxl.bitReader, jxl.imageHeader, &jxl.options)
			header, err = imgFrame.ReadFrameHeader()
			if err != nil {
				return nil, err
			}
			frameCount++

			if jxl.lfBuffer[header.LfLevel] == nil && header.Flags&frame.USE_LF_FRAME != 0 {
				return nil, errors.New("LF level too large")
			}

			err := imgFrame.ReadTOC()
			if err != nil {
				return nil, err
			}

			if jxl.options.ParseOnly {
				imgFrame.SkipFrameData()
				continue
			}
			err = imgFrame.DecodeFrame(jxl.lfBuffer[header.LfLevel])
			if err != nil {
				return nil, err
			}

			if header.LfLevel > 0 {
				jxl.lfBuffer[header.LfLevel-1] = imgFrame.Buffer
			}
			save := (header.SaveAsReference != 0 || header.Duration == 0) && !header.IsLast && header.FrameType != frame.LF_FRAME
			if imgFrame.IsVisible() {
				visibleFrames++
				invisibleFrames = 0
			} else {
				invisibleFrames++
			}

			err = imgFrame.InitializeNoise(int64(visibleFrames<<32) | invisibleFrames)
			if err != nil {
				return nil, err
			}
			err = imgFrame.Upsample()
			if err != nil {
				return nil, err
			}

			if save && header.SaveBeforeCT {
				jxl.reference[header.SaveAsReference] = imgFrame.Buffer
			}

			err = jxl.computePatches(imgFrame)
			if err != nil {
				return nil, err
			}

			err = imgFrame.RenderSplines()
			if err != nil {
				return nil, err
			}

			err = imgFrame.SynthesizeNoise()
			if err != nil {
				return nil, err
			}

			err = jxl.performColourTransforms(matrix, imgFrame)
			if err != nil {
				return nil, err
			}

			if header.Encoding == frame.VARDCT && jxl.options.RenderVarblocks {
				panic("VARDCT not implemented yet")
			}

			if header.FrameType == frame.REGULAR_FRAME || header.FrameType == frame.SKIP_PROGRESSIVE {
				found := false
				for i := uint32(0); i < 4; i++ {
					if image2.ImageBufferEquals(jxl.reference[i], canvas) && i != header.SaveAsReference {
						found = true
						break
					}
				}
				if found {
					canvas = util.DeepCopy3[float32](canvas)
				}
				err = jxl.blendFrame(canvas, reference, imgFrame)
				if err != nil {
					return nil, err
				}
			}

			if save && !header.SaveBeforeCT {
				reference[header.SaveAsReference] = canvas
			}

			if header.IsLast {
				break
			}
		}

		err = jxl.bitReader.ZeroPadToByte()
		if err != nil {
			return nil, err
		}

		// TOOD(kpfaulkner) unsure if need to perform similar drain cache functionality here. Don't think we do.
		if jxl.options.ParseOnly {
			return nil, nil
		}

		orientation := imageHeader.Orientation
		orientedCanvas := util.MakeMatrix3D[float32](len(canvas), 0, 0)
		for i := 0; i < len(orientedCanvas); i++ {
			orientedCanvas[i], err = jxl.transposeBuffer(canvas[i], orientation)
			if err != nil {
				return nil, err
			}
		}

		// generate image and return.
		img, err := NewImage(orientedCanvas, *imageHeader)
		if err != nil {
			return nil, err
		}

		return img, nil
	}

	panic("make JXL image here?")
	return nil, nil
}

// Read signature
func (jxl *JXLCodestreamDecoder) readSignatureAndBoxes() error {

	br := NewBoxReader(jxl.bitReader)
	boxHeaders, err := br.ReadBoxHeader()
	if err != nil {
		return err
	}

	jxl.boxHeaders = boxHeaders
	jxl.level = br.level
	return nil
}

func (jxl *JXLCodestreamDecoder) computePatches(frame *frame.Frame) error {

	header := frame.Header
	frameBuffer := frame.Buffer
	colourChannels := jxl.imageHeader.GetColourChannelCount()
	extraChannels := len(jxl.imageHeader.ExtraChannelInfo)
	patches := frame.LfGlobal.Patches
	for i := 0; i < len(patches); i++ {
		patch := patches[i]
		if patch.Ref > 3 {
			return errors.New("patch out of range")
		}
		refBuffer := jxl.reference[patch.Ref]
		if refBuffer == nil || len(refBuffer) == 0 {
			continue
		}
		if patch.Height+int32(patch.Origin.Y) > int32(len(refBuffer[0])) || patch.Width+int32(patch.Origin.X) > int32(len(refBuffer[0][0])) {
			return errors.New("patch too large")
		}
		for j := 0; i < len(patch.Positions); j++ {
			x0 := patch.Positions[j].X
			y0 := patch.Positions[j].Y
			if x0 < 0 || y0 < 0 {
				return errors.New("patch size out of bounds")
			}

			if patch.Height+int32(y0) > int32(header.Height) || patch.Width+int32(x0) > int32(header.Width) {
				return errors.New("patch size out of bounds")
			}

			for d := int32(0); d < int32(colourChannels)+int32(extraChannels); d++ {
				var c int32
				if d < int32(colourChannels) {
					c = 0
				} else {
					c = d - int32(colourChannels) + 1
				}
				info := patch.BlendingInfos[j][c]
				if info.Mode == 0 {
					continue
				}
				var premult bool
				if jxl.imageHeader.HasAlpha() {
					premult = jxl.imageHeader.ExtraChannelInfo[info.AlphaChannel].AlphaAssociated
				} else {
					premult = true
				}
				isAlpha := c > 0 && jxl.imageHeader.ExtraChannelInfo[c-1].EcType == bundle.ALPHA
				if info.Mode > 0 && header.Upsampling > 1 && c > 0 && header.EcUpsampling[c-1]<<jxl.imageHeader.ExtraChannelInfo[c-1].DimShift != header.Upsampling {
					return errors.New("Alpha channel upsampling mismatch during patches")
				}
				for y := int32(0); y < patch.Height; y++ {
					for x := int32(0); x < patch.Width; x++ {
						oldX := x + int32(x0)
						oldY := y + int32(y0)
						newX := x + int32(patch.Origin.X)
						newY := y + int32(patch.Origin.Y)
						oldSample := frameBuffer[d][oldY][oldX]
						newSample := refBuffer[d][newY][newX]
						alpha := float32(0.0)
						newAlpha := float32(0.0)
						oldAlpha := float32(0.0)
						if info.Mode > 3 {
							if jxl.imageHeader.HasAlpha() {
								oldAlpha = frameBuffer[uint32(colourChannels)+info.AlphaChannel][oldY][oldX]
							} else {
								oldAlpha = 1.0
							}
							if jxl.imageHeader.HasAlpha() {
								newAlpha = refBuffer[uint32(colourChannels)+info.AlphaChannel][newY][newX]
							} else {
								newAlpha = 1.0
							}
							if info.Clamp {
								newAlpha = util.Clamp3Float32(newAlpha, 0.0, 1.0)
							}
							if info.Mode < 6 || !isAlpha || !premult {
								alpha = oldAlpha + newAlpha*(1-oldAlpha)
							}

							var sample float32
							switch info.Mode {
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

func (jxl *JXLCodestreamDecoder) performColourTransforms(matrix *color.OpsinInverseMatrix, frame *frame.Frame) error {
	frameBuffer := frame.Buffer
	if matrix != nil {
		err := matrix.InvertXYB(frameBuffer, jxl.imageHeader.ToneMapping.GetIntensityTarget())
		if err != nil {
			return err
		}
	}

	if frame.Header.DoYCbCr {
		size, err := frame.GetPaddedFrameSize()
		if err != nil {
			return err
		}
		for y := uint32(0); y < size.Height; y++ {
			for x := uint32(0); x < size.Width; x++ {
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

func (jxl *JXLCodestreamDecoder) blendFrame(canvas [][][]float32, reference []image2.ImageBuffer, imgFrame *frame.Frame) error {

	width := jxl.imageHeader.Size.Width
	height := jxl.imageHeader.Size.Height
	header := imgFrame.Header
	frameStartY := int32(0)
	if header.Bounds.Origin.X >= 0 {
		frameStartY = header.Bounds.Origin.Y
	}
	frameStartX := int32(0)
	if header.Bounds.Origin.X >= 0 {
		frameStartX = header.Bounds.Origin.X
	}
	lowerCorner := header.Bounds.ComputeLowerCorner()
	frameHeight := util.Min(lowerCorner.Y, int32(height))
	frameWidth := util.Min(lowerCorner.X, int32(width))

	frameColours := imgFrame.GetColorChannelCount()
	imageColours := jxl.imageHeader.GetColourChannelCount()
	hasAlpha := jxl.imageHeader.HasAlpha()
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
		frameBuffer := imgFrame.Buffer[frameC]
		var info *frame.BlendingInfo
		if frameC < int32(frameColours) {
			info = imgFrame.Header.BlendingInfo
		} else {
			info = &imgFrame.Header.EcBlendingInfo[frameC-int32(frameColours)]
		}
		isAlpha := c >= int32(imageColours) && jxl.imageHeader.ExtraChannelInfo[c-int32(imageColours)].EcType == bundle.ALPHA
		premult := hasAlpha && jxl.imageHeader.ExtraChannelInfo[info.AlphaChannel].AlphaAssociated

		refBuffer := reference[info.Source]
		if info.Mode == frame.BLEND_REPLACE || refBuffer == nil && info.Mode == frame.BLEND_ADD {
			offY := frameStartY - header.Bounds.Origin.Y
			offX := frameStartX - header.Bounds.Origin.X
			jxl.copyToCanvas(canvas[c], util.Point{Y: frameStartY, X: frameStartX}, util.Point{X: offX, Y: offY},
				util.Dimension{Width: uint32(frameWidth), Height: uint32(frameHeight)}, frameBuffer)
			continue
		}
		ref := refBuffer[c]
		switch info.Mode {
		case frame.BLEND_ADD:
			for y := int32(0); y < frameHeight; y++ {
				cy := y + frameStartY
				for x := int32(0); x < frameWidth; x++ {
					cx := x + frameStartX
					canvas[c][cy][cx] = ref[cy][cx] + frameBuffer[y][x]
				}
			}
			break
		case frame.BLEND_MULT:
			for y := int32(0); y < frameHeight; y++ {
				cy := y + frameStartY
				if ref != nil {
					for x := int32(0); x < frameWidth; x++ {
						cx := x + frameStartX
						newSample := frameBuffer[y][x]
						if info.Clamp {
							newSample = util.Clamp3Float32(newSample, 0.0, 1.0)
						}
						canvas[c][cy][cx] = newSample * ref[cy][cx]
					}
				} else {
					util.FillFloat32(canvas[c][cy], uint32(frameStartX), uint32(frameWidth), 0.0)
				}
			}
			break
		case frame.BLEND_BLEND:
			for cy := frameStartY; cy < frameHeight+frameStartY; cy++ {
				for cx := frameStartX; cx < frameWidth+frameStartX; cx++ {
					var oldAlpha float32

					if jxl.imageHeader.HasAlpha() {
						oldAlpha = 1.0
					} else {
						if ref != nil {
							oldAlpha = refBuffer[imageColours+int(info.AlphaChannel)][cy][cx]
						} else {
							oldAlpha = 0.0
						}
					}
					var newAlpha float32
					if jxl.imageHeader.HasAlpha() {
						newAlpha = 1.0
					} else {
						newAlpha = imgFrame.GetImageSample(int32(uint32(frameColours)+info.AlphaChannel), cx, cy)
					}

					if info.Clamp {
						newAlpha = util.Clamp3Float32(newAlpha, 0.0, 1.0)
					}
					var alpha = float32(1.0)
					var oldSample float32
					if ref != nil {
						oldSample = ref[cy][cx]
					} else {
						oldSample = 0.0
					}
					newSample := imgFrame.GetImageSample(frameC, cx, cy)
					if isAlpha || hasAlpha && !premult {
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
		default:
			return errors.New("Illegal blend Mode")
		}
	}
	return nil
}

func (jxl *JXLCodestreamDecoder) copyToCanvas(canvas [][]float32, start util.Point, off util.Point, size util.Dimension, frameBuffer [][]float32) {
	for y := uint32(0); y < size.Height; y++ {
		copy(canvas[y+uint32(start.X)][off.X:], frameBuffer[y+uint32(off.Y)][off.X:uint32(off.X)+size.Width])
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
