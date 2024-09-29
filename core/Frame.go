package core

import (
	"bytes"
	"errors"

	"github.com/kpfaulkner/jxl-go/entropy"
	"github.com/kpfaulkner/jxl-go/jxlio"
	"github.com/kpfaulkner/jxl-go/util"
)

var (
	cMap = []int{0, 2, 1}
)

type Frame struct {
	globalMetadata   *ImageHeader
	options          *JXLOptions
	reader           *jxlio.Bitreader
	header           *FrameHeader
	width            uint32
	height           uint32
	bounds           Rectangle
	groupRowStride   uint32
	lfGroupRowStride uint32
	numGroups        uint32
	numLFGroups      uint32
	permutatedTOC    bool
	tocPermutation   []uint32
	tocLengths       []uint32

	lfGroups []*LFGroup
	// unsure about this
	buffers    [][]uint8
	decoded    bool
	lfGlobal   *LFGlobal
	buffer     [][][]float32
	globalTree *MATree
	hfGlobal   *HFGlobal
	passes     []Pass
}

func (f *Frame) readFrameHeader() (FrameHeader, error) {

	f.reader.ZeroPadToByte()
	var err error
	f.header, err = NewFrameHeaderWithReader(f.reader, f.globalMetadata)
	if err != nil {
		return FrameHeader{}, err
	}

	f.bounds = Rectangle{
		origin: f.header.bounds.origin,
		size:   f.header.bounds.size,
	}

	f.groupRowStride = util.CeilDiv(f.bounds.size.width, f.header.groupDim)
	f.lfGroupRowStride = util.CeilDiv(f.bounds.size.width, f.header.groupDim<<3)
	f.numGroups = f.groupRowStride * util.CeilDiv(f.bounds.size.height, f.header.groupDim)
	f.numLFGroups = f.lfGroupRowStride * util.CeilDiv(f.bounds.size.height, f.header.groupDim<<3)

	return *f.header, nil
}

func (f *Frame) readTOC() error {
	var tocEntries uint32

	if f.numGroups == 1 && f.header.passes.numPasses == 1 {
		tocEntries = 1
	} else {
		tocEntries = 1 + f.numLFGroups + 1 + f.numGroups*f.header.passes.numPasses
	}

	f.permutatedTOC = f.reader.MustReadBool()
	if f.permutatedTOC {
		tocStream, err := entropy.NewEntropyStreamWithReaderAndNumDists(f.reader, 8)
		if err != nil {
			return err
		}
		f.tocPermutation, err = readPermutation(f.reader, tocStream, tocEntries, 0)
		if err != nil {
			return err
		}
		if !tocStream.ValidateFinalState() {
			return errors.New("invalid final ANS state decoding TOC")
		}
	} else {
		f.tocPermutation = make([]uint32, tocEntries)
		for i := uint32(0); i < tocEntries; i++ {
			a := i
			f.tocPermutation[i] = a
		}
	}
	f.reader.ZeroPadToByte()
	f.tocLengths = make([]uint32, tocEntries)

	for i := 0; i < int(tocEntries); i++ {
		f.tocLengths[i] = f.reader.MustReadU32(0, 10, 1024, 14, 17408, 22, 4211712, 30)
	}

	f.reader.ZeroPadToByte()

	f.buffers = make([][]uint8, tocEntries)

	// TODO(kpfaulkner) potentially make this more concurrent?
	if tocEntries != 1 && !f.options.parseOnly {
		for i := 0; i < int(tocEntries); i++ {
			b, err := f.readBuffer(i)
			if err != nil {
				return err
			}
			f.buffers[i] = b
		}
	}
	return nil
}

// TODO(kpfaulkner) really need to check this.
func (f *Frame) readBuffer(index int) ([]uint8, error) {
	length := f.tocLengths[index]
	buffer := make([]uint8, length+4)
	err := f.reader.ReadBytesToBuffer(buffer, length)
	if err != nil {
		return nil, err
	}
	if len(buffer) < int(length) {
		return nil, errors.New("unable to read full TOC entry")
	}

	return buffer, nil
}

func ctxFunc(x int64) int {
	return min(7, util.CeilLog1p(x))
}

func readPermutation(reader *jxlio.Bitreader, stream *entropy.EntropyStream, size uint32, skip uint32) ([]uint32, error) {
	end, err := stream.ReadSymbol(reader, ctxFunc(int64(size)))
	if err != nil {
		return nil, err
	}

	if uint32(end) > size-skip {
		return nil, errors.New("illegal end value in lehmer sequence")
	}

	lehmer := make([]uint32, size)
	for i := skip; i < uint32(end)+skip; i++ {
		ii, err := stream.ReadSymbol(reader, ctxFunc(int64(util.IfThenElse(i > skip, lehmer[i-1], 0))))
		if err != nil {
			return nil, err
		}
		lehmer[i] = uint32(ii)
		if lehmer[i] >= size-i {
			return nil, errors.New("illegal end value in lehmer sequence")
		}
	}

	var temp []uint32
	permutation := make([]uint32, size)
	for i := 0; i < int(size); i++ {
		temp = append(temp, uint32(i))
	}

	for i, index := range lehmer {
		permutation[i] = temp[index]
	}

	return permutation, nil
}

func NewFrameWithReader(reader *jxlio.Bitreader, imageHeader *ImageHeader, options *JXLOptions) *Frame {

	frame := &Frame{
		globalMetadata: imageHeader,
		options:        options,
		reader:         reader,
	}

	return frame
}

func (f *Frame) skipFrameData() error {
	for i := 0; i < len(f.tocLengths); i++ {
		buffer := make([]byte, f.tocLengths[i])
		err := f.reader.ReadBytesToBuffer(buffer, f.tocLengths[i])
		if err != nil {
			return err
		}
	}
	return nil
}

// gets a bit reader for each TOC entry???
func (f *Frame) getBitreader(index int) (*jxlio.Bitreader, error) {
	if len(f.tocLengths) == 1 {
		panic("getBitreader panic... unsure what to do")
	}
	permutedIndex := f.tocPermutation[index]
	return jxlio.NewBitreaderWithIndex(bytes.NewReader(f.buffers[permutedIndex]), true, index), nil
}

func (f *Frame) decodeFrame(lfBuffer [][][]float32) error {

	if f.decoded {
		return nil
	}
	f.decoded = true

	lfGlobalBitReader, err := f.getBitreader(0)
	if err != nil {
		return err
	}
	f.lfGlobal, err = NewLFGlobalWithReader(lfGlobalBitReader, f)
	if err != nil {
		return err
	}

	paddedSize, err := f.getPaddedFrameSize()
	if err != nil {
		return err
	}

	f.buffer = make([][][]float32, f.getColorChannelCount()+len(f.globalMetadata.extraChannelInfo))
	for c := 0; c < len(f.buffer); c++ {
		if c < 3 && c < f.getColorChannelCount() {
			//shiftedSize := paddedSize.ShiftRightWithIntPoint(f.header.jpegUpsampling[c])
			shiftedHeight := paddedSize.height >> f.header.jpegUpsamplingY[c]
			shiftedWidth := paddedSize.width >> f.header.jpegUpsamplingX[c]
			f.buffer[c] = util.MakeMatrix2D[float32](int(shiftedHeight), int(shiftedWidth))
		} else {
			f.buffer[c] = util.MakeMatrix2D[float32](int(paddedSize.height), int(paddedSize.width))
		}
	}

	err = f.decodeLFGroups(lfBuffer)
	if err != nil {
		return err
	}

	hfGlobalReader, err := f.getBitreader(1 + int(f.numLFGroups))
	if err != nil {
		return err
	}

	if f.header.encoding == VARDCT {
		panic("VARDCT not implemented")
	} else {
		f.hfGlobal = nil
	}

	err = f.decodePasses(hfGlobalReader)
	if err != nil {
		return err
	}

	err = f.decodePassGroups()
	if err != nil {
		return err
	}

	err = f.lfGlobal.gModular.stream.applyTransforms()
	if err != nil {
		return err
	}

	modularBuffer := f.lfGlobal.gModular.stream.getDecodedBuffer()
	for c := 0; c < len(modularBuffer); c++ {
		cIn := c
		var scaleFactor float32
		isModularColour := f.header.encoding == MODULAR && c < f.getColorChannelCount()
		isModularXYB := f.globalMetadata.xybEncoded && isModularColour
		var cOut int
		if isModularXYB {
			cOut = cMap[c]
		} else {
			cOut = c
		}
		cOut += len(f.buffer) - len(modularBuffer)
		ecIndex := c
		if f.header.encoding == MODULAR {
			ecIndex -= f.globalMetadata.getColourChannelCount()
		}
		if isModularXYB {
			scaleFactor = f.lfGlobal.lfDequant[cOut]
		} else if isModularColour && f.globalMetadata.bitDepth.expBits != 0 {
			scaleFactor = 1.0
		} else if isModularColour {
			// FIXME(kpfaulkner) need to check this.
			//scaleFactor = float32(1.0 / ^(^uint32(0) << f.globalMetadata.bitDepth.bitsPerSample))
			step1 := f.globalMetadata.bitDepth.bitsPerSample
			step2 := ^int32(0) << step1
			step3 := ^step2
			scaleFactor = 1.0 / float32(step3)
		} else {
			scaleFactor = float32(1.0 / ^(^uint32(0) << f.globalMetadata.extraChannelInfo[ecIndex].bitDepth.bitsPerSample))
		}
		if isModularXYB && cIn == 2 {
			for y := uint32(0); y < f.height; y++ {
				for x := uint32(0); x < f.width; x++ {
					f.buffer[cOut][y][x] = scaleFactor * float32(modularBuffer[0][y][x]+modularBuffer[2][y][x])
				}
			}
		} else {
			for y := uint32(0); y < f.bounds.size.height; y++ {
				for x := uint32(0); x < f.bounds.size.width; x++ {
					f.buffer[cOut][y][x] = scaleFactor * float32(modularBuffer[cIn][y][x])
				}
			}
		}

	}
	f.invertSubsampling()

	if f.header.restorationFilter.gab {
		f.performGabConvolution()
	}

	if f.header.restorationFilter.epfIterations > 0 {
		f.performEdgePreservingFilter()
	}
	return nil
}

func (f *Frame) isVisible() bool {
	return f.header.frameType == REGULAR_FRAME || f.header.frameType == SKIP_PROGRESSIVE && (f.header.duration != 0 || f.header.isLast)
}

func (f *Frame) getColorChannelCount() int {
	if f.globalMetadata.xybEncoded || f.header.encoding == VARDCT {
		return 3
	}
	return f.globalMetadata.getColourChannelCount()
}

func (f *Frame) getPaddedFrameSize() (Dimension, error) {

	factorY := 1 << util.Max(f.header.jpegUpsamplingY...)
	factorX := 1 << util.Max(f.header.jpegUpsamplingX...)
	var width uint32
	var height uint32
	if f.header.encoding == VARDCT {
		panic("VARDCT not implemented")
	} else {
		width = f.bounds.size.width
		height = f.bounds.size.height
	}

	height = util.CeilDiv(height, uint32(factorY))
	width = util.CeilDiv(width, uint32(factorX))
	if f.header.encoding == VARDCT {
		panic("VARDCT not implemented")
	} else {
		return Dimension{
			width:  width * uint32(factorX),
			height: height * uint32(factorY),
		}, nil
	}
}

func (f *Frame) decodeLFGroups(lfBuffer [][][]float32) error {

	lfReplacementChannels := []*ModularChannel{}
	lfReplacementChannelIndicies := []int{}

	for i := 0; i < len(f.lfGlobal.gModular.stream.channels); i++ {
		ch := f.lfGlobal.gModular.stream.channels[i]
		if !ch.decoded {
			if ch.hshift >= 3 && ch.vshift >= 3 {
				lfReplacementChannelIndicies = append(lfReplacementChannelIndicies, i)
				height := f.header.lfGroupDim >> ch.vshift
				width := f.header.lfGroupDim >> ch.hshift
				lfReplacementChannels = append(lfReplacementChannels, NewModularChannelWithAllParams(int32(height), int32(width), ch.hshift, ch.vshift, false))
			}
		}
	}

	f.lfGroups = make([]*LFGroup, f.numLFGroups)

	for lfGroupID := uint32(0); lfGroupID < f.numLFGroups; lfGroupID++ {
		reader, err := f.getBitreader(1 + int(lfGroupID))
		if err != nil {
			return err
		}

		lfGroupPos := f.getLFGroupLocation(int32(lfGroupID))
		replaced := make([]ModularChannel, len(lfReplacementChannels))
		for _, r := range lfReplacementChannels {
			replaced = append(replaced, *NewModularChannelFromChannel(*r))
		}
		frameSize, err := f.getPaddedFrameSize()
		if err != nil {
			return err
		}
		for i, info := range replaced {
			lfHeight := frameSize.height >> info.vshift
			lfWidth := frameSize.width >> info.hshift
			info.origin.Y = uint32(lfGroupPos.Y) * info.size.height
			info.origin.X = uint32(lfGroupPos.X) * info.size.width
			info.size.height = util.Min(info.size.height, lfHeight-info.origin.Y)
			info.size.width = util.Min(info.size.width, lfWidth-info.origin.X)
			replaced[i] = info
		}
		f.lfGroups[lfGroupID], err = NewLFGroup(reader, f, int32(lfGroupID), replaced, lfBuffer)
		if err != nil {
			return err
		}
	}

	for lfGroupID := uint32(0); lfGroupID < f.numLFGroups; lfGroupID++ {
		for j := 0; j < len(lfReplacementChannelIndicies); j++ {
			index := lfReplacementChannelIndicies[j]
			channel := f.lfGlobal.gModular.stream.channels[index]
			newChannelInfo := f.lfGroups[lfGroupID].modularLFGroup.channels[index]
			newChannel := newChannelInfo.buffer
			for y := 0; y < len(newChannel); y++ {
				copy(channel.buffer[uint32(y)+newChannelInfo.origin.Y], newChannel[y])
			}
		}
	}
	return nil
}

func (f *Frame) decodePasses(reader *jxlio.Bitreader) error {

	var err error
	f.passes = make([]Pass, f.header.passes.numPasses)
	for pass := 0; pass < len(f.passes); pass++ {
		prevMinShift := uint32(0)
		if pass > 0 {
			prevMinShift = f.passes[pass-1].minShift
		}

		f.passes[pass], err = NewPassWithReader(reader, f, uint32(pass), prevMinShift)
		if err != nil {
			return err
		}
	}
	return nil
}

// JXLatte seems to break the processing (somehow?  thread race condition?) into group 0->22
// then 23 onwards. So going to do the same just to make it easier to compare outputs.
func (f *Frame) decodePassGroups() error {
	numPasses := len(f.passes)
	numGroups := int(f.numGroups)
	passGroups := util.MakeMatrix2D[PassGroup](numPasses, numGroups)

	for pass0 := 0; pass0 < numPasses; pass0++ {
		pass := pass0
		for group0 := 0; group0 < numGroups; group0++ {
			group := group0
			br, err := f.getBitreader(2 + int(f.numLFGroups) + pass*int(f.numGroups) + group)
			if err != nil {
				return err
			}

			replaced := []ModularChannel{}
			for _, r := range f.passes[pass].replacedChannels {
				mc := NewModularChannelFromChannel(r)
				replaced = append(replaced, *mc)
			}
			for i := 0; i < len(replaced); i++ {
				info := replaced[i]
				shift := util.NewIntPointWithXY(uint32(info.hshift), uint32(info.vshift))
				passGroupSize := util.NewIntPoint(int(f.header.groupDim)).ShiftRightWithIntPoint(shift)
				rowStride := util.CeilDiv(uint32(info.size.width), passGroupSize.X)
				pos := util.Coordinates(uint32(group), rowStride).TimesWithIntPoint(passGroupSize)
				chanSize := util.NewIntPointWithXY(uint32(info.size.width), uint32(info.size.height))
				info.origin = pos
				size := passGroupSize.Min(chanSize.Minus(info.origin))
				info.size.width = size.X
				info.size.height = size.Y
				replaced[i] = info
			}

			pg, err := NewPassGroupWithReader(br, f, uint32(pass), uint32(group), replaced)
			if err != nil {
				return err
			}
			//f.passes[pass].replacedChannels = replaced
			passGroups[pass][group] = *pg
		}
	}

	for pass := 0; pass < numPasses; pass++ {
		j := 0
		for i := 0; i < len(f.passes[pass].replacedChannels); i++ {
			channel := f.lfGlobal.gModular.stream.channels[i]
			channel.allocate()
			for group := 0; group < int(f.numGroups); group++ {
				newChannelInfo := passGroups[pass][group].modularStream.channels[j]
				buff := newChannelInfo.buffer
				for y := 0; y < len(buff); y++ {
					//channel.Buffer[y+int(newChannelInfo.origin.Y)] = buff[y]
					idx := y + int(newChannelInfo.origin.Y)
					//copy(channel.Buffer[idx], buff[y])
					copy(channel.buffer[idx][newChannelInfo.origin.X:], buff[y][:len(buff[y])])
				}
			}
			//f.lfGlobal.gModular.stream.channels[i] = channel
			j++
		}
	}
	if f.header.encoding == VARDCT {
		panic("VARDCT not implemented")
	}

	return nil
}

func (f *Frame) invertSubsampling() {
	for c := 0; c < 3; c++ {
		xShift := f.header.jpegUpsamplingX[c]
		yShift := f.header.jpegUpsamplingY[c]
		for xShift > 0 {
			xShift--
			oldChannel := f.buffer[c]
			newChannel := util.MakeMatrix2D[float32](len(oldChannel), 0)
			for y := 0; y < len(oldChannel); y++ {
				oldRow := oldChannel[y]
				newRow := make([]float32, len(oldRow)*2)
				for x := 0; x < len(oldRow); x++ {
					b75 := 0.075 * oldRow[x]
					xx := 0
					if x != 0 {
						xx = x - 1
					}
					newRow[2*x] = b75 + 0.25*oldRow[xx]
					xx = len(oldRow) - 1
					if x+1 == len(oldRow) {
						xx = x + 1
					}
					newRow[2*x+1] = b75 + 0.25*oldRow[xx]
				}
				newChannel[y] = newRow
			}
			f.buffer[c] = newChannel
		}
		for yShift > 0 {
			yShift--
			oldChannel := f.buffer[c]
			newChannel := util.MakeMatrix2D[float32](len(oldChannel)*2, 0)
			for y := 0; y < len(oldChannel); y++ {
				oldRow := oldChannel[y]
				xx := 0
				if y == 0 {
					xx = y - 1
				}
				oldRowPrev := oldChannel[xx]
				xx = len(oldChannel) - 1
				if y+1 == len(oldChannel) {
					xx = y + 1
				}
				oldRowNext := oldChannel[xx]
				firstNewRow := make([]float32, len(oldRow))
				secondNewRow := make([]float32, len(oldRow))
				for x := 0; x < len(oldRow); x++ {
					b75 := 0.075 * oldRow[x]
					firstNewRow[x] = b75 + 0.25*oldRowPrev[x]
					secondNewRow[x] = b75 + 0.25*oldRowNext[x]
				}
				newChannel[2*y] = firstNewRow
				newChannel[2*y+1] = secondNewRow
			}
			f.buffer[c] = newChannel
		}
	}
}

func (f *Frame) performGabConvolution() error {
	panic("not implemented")
}

func (f *Frame) performEdgePreservingFilter() error {
	panic("not implemented")
}

func (f *Frame) initializeNoise(seed0 int64) error {
	if f.lfGlobal.noiseParameters == nil || len(f.lfGlobal.noiseParameters) == 0 {
		return nil
	}
	// FIXME(kpfaulkner) yet to do.
	panic("not implemented")

	//rowStride := util.CeilDiv(f.header.Width, f.header.groupDim)
	//localNoiseBuffer := util.MakeMatrix3D[float32](3, int(f.header.Height), int(f.header.Width))
	//numGroups := rowStride * util.CeilDiv(f.header.Height, f.header.groupDim)
	//for group := uint32(0); group < numGroups; group++ {
	//	groupXYUp := util.Coordinates(group, rowStride).Times(f.header.upsampling)
	//	for iy := uint32(0); iy < f.header.upsampling; iy++ {
	//		for ix := uint32(0); ix < f.header.upsampling; ix++ {
	//			x0 := (groupXYUp.X + ix) * f.header.groupDim
	//			y0 := (groupXYUp.Y + iy) * f.header.groupDim
	//
	//		}
	//	}
	//}
}

func (f *Frame) getImageSample(c int32, x int32, y int32) float32 {

	frameY := y - f.bounds.origin.Y
	frameX := x - f.bounds.origin.X

	if frameY < 0 || frameX < 0 || frameY >= int32(f.bounds.size.height) || frameX >= int32(f.bounds.size.width) {
		return 0
	}
	return f.buffer[c][frameY][frameX]
}

func (f *Frame) upsample() error {
	var err error
	for c := 0; c < len(f.buffer); c++ {
		f.buffer[c], err = f.performUpsampling(f.buffer[c], c)
		if err != nil {
			return err
		}
	}
	f.bounds.size.height *= f.header.upsampling
	f.bounds.size.width *= f.header.upsampling

	f.bounds.origin.Y *= int32(f.header.upsampling)
	f.bounds.origin.X *= int32(f.header.upsampling)
	return nil
}

func (f *Frame) performUpsampling(buffer [][]float32, c int) ([][]float32, error) {

	colour := f.getColorChannelCount()
	var k uint32
	if c < colour {
		k = f.header.upsampling
	} else {
		k = f.header.ecUpsampling[c-colour]
	}
	if k == 1 {
		return buffer, nil
	}

	// FIXME(kpfaulkner) not implemented
	panic("not implemented")
}

func (f *Frame) renderSplines() error {
	if f.lfGlobal.splines == nil {
		return nil
	}

	panic("renderSplines not implemented")
}

func (f *Frame) synthesizeNoise() error {
	if f.lfGlobal.noiseParameters == nil {
		return nil
	}

	panic("synthesizeNoise not implemented")
}

func (f *Frame) getLFGroupSize(lfGroupID int32) (Dimension, error) {
	pos := f.getLFGroupLocation(lfGroupID)
	paddedSize, err := f.getPaddedFrameSize()
	if err != nil {
		return Dimension{}, err
	}

	height := util.Min(f.header.lfGroupDim, paddedSize.height-uint32(pos.Y)*f.header.lfGroupDim)
	width := util.Min(f.header.lfGroupDim, paddedSize.width-uint32(pos.X)*f.header.lfGroupDim)
	return Dimension{
		height: height,
		width:  width,
	}, nil
}

func (f *Frame) getLFGroupLocation(lfGroupID int32) *Point {
	return NewPoint(lfGroupID/int32(f.lfGroupRowStride), lfGroupID%int32(f.lfGroupRowStride))
}

func (f *Frame) getGroupLocation(groupID int32) *Point {
	return NewPoint(groupID/int32(f.groupRowStride), groupID%int32(f.groupRowStride))
}

func (f *Frame) getLFGroupForGroup(groupID int32) *LFGroup {
	pos := f.getGroupLocation(groupID)
	idx := (pos.Y>>3)*int32(f.lfGroupRowStride) + (pos.X >> 3)
	//idx1 := uint32(pos.Y) >> 3
	//idx2 := uint32(f.lfGroupRowStride)
	//idx3 := uint32(pos.X >> 3)
	//idx := idx1*idx2 + idx3
	return f.lfGroups[idx]
}
