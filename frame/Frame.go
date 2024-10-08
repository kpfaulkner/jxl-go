package frame

import (
	"bytes"
	"errors"
	"sync"

	"golang.org/x/sync/errgroup"

	"github.com/kpfaulkner/jxl-go/bundle"
	"github.com/kpfaulkner/jxl-go/entropy"
	"github.com/kpfaulkner/jxl-go/jxlio"
	"github.com/kpfaulkner/jxl-go/options"
	"github.com/kpfaulkner/jxl-go/util"
)

var (
	cMap = []int{1, 0, 2}
)

type Frame struct {
	globalMetadata   *bundle.ImageHeader
	options          *options.JXLOptions
	reader           *jxlio.Bitreader
	Header           *FrameHeader
	width            uint32
	height           uint32
	bounds           util.Rectangle
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
	LfGlobal   *LFGlobal
	Buffer     [][][]float32
	globalTree *MATree
	hfGlobal   *HFGlobal
	passes     []Pass
}

func (f *Frame) ReadFrameHeader() (FrameHeader, error) {

	f.reader.ZeroPadToByte()
	var err error
	f.Header, err = NewFrameHeaderWithReader(f.reader, f.globalMetadata)
	if err != nil {
		return FrameHeader{}, err
	}

	f.bounds = util.Rectangle{
		Origin: f.Header.Bounds.Origin,
		Size:   f.Header.Bounds.Size,
	}

	f.groupRowStride = util.CeilDiv(f.bounds.Size.Width, f.Header.groupDim)
	f.lfGroupRowStride = util.CeilDiv(f.bounds.Size.Width, f.Header.groupDim<<3)
	f.numGroups = f.groupRowStride * util.CeilDiv(f.bounds.Size.Height, f.Header.groupDim)
	f.numLFGroups = f.lfGroupRowStride * util.CeilDiv(f.bounds.Size.Height, f.Header.groupDim<<3)

	return *f.Header, nil
}

func (f *Frame) ReadTOC() error {
	var tocEntries uint32

	if f.numGroups == 1 && f.Header.passes.numPasses == 1 {
		tocEntries = 1
	} else {
		tocEntries = 1 + f.numLFGroups + 1 + f.numGroups*f.Header.passes.numPasses
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
	if tocEntries != 1 && !f.options.ParseOnly {
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

func NewFrameWithReader(reader *jxlio.Bitreader, imageHeader *bundle.ImageHeader, options *options.JXLOptions) *Frame {

	frame := &Frame{
		globalMetadata: imageHeader,
		options:        options,
		reader:         reader,
	}

	return frame
}

func (f *Frame) SkipFrameData() error {
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
	return jxlio.NewBitreaderWithIndex(bytes.NewReader(f.buffers[permutedIndex]), index), nil
}

func (f *Frame) DecodeFrame(lfBuffer [][][]float32) error {

	if f.decoded {
		return nil
	}
	f.decoded = true

	lfGlobalBitReader, err := f.getBitreader(0)
	if err != nil {
		return err
	}
	f.LfGlobal, err = NewLFGlobalWithReader(lfGlobalBitReader, f)
	if err != nil {
		return err
	}

	paddedSize, err := f.GetPaddedFrameSize()
	if err != nil {
		return err
	}

	f.Buffer = make([][][]float32, f.GetColorChannelCount()+len(f.globalMetadata.ExtraChannelInfo))
	for c := 0; c < len(f.Buffer); c++ {
		if c < 3 && c < f.GetColorChannelCount() {
			//shiftedSize := paddedSize.ShiftRightWithIntPoint(f.Header.jpegUpsampling[c])
			shiftedHeight := paddedSize.Height >> f.Header.jpegUpsamplingY[c]
			shiftedWidth := paddedSize.Width >> f.Header.jpegUpsamplingX[c]
			f.Buffer[c] = util.MakeMatrix2D[float32](int(shiftedHeight), int(shiftedWidth))
		} else {
			f.Buffer[c] = util.MakeMatrix2D[float32](int(paddedSize.Height), int(paddedSize.Width))
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

	if f.Header.Encoding == VARDCT {
		panic("VARDCT not implemented")
	} else {
		f.hfGlobal = nil
	}

	err = f.decodePasses(hfGlobalReader)
	if err != nil {
		return err
	}

	err = f.decodePassGroupsConcurrent()
	if err != nil {
		return err
	}

	err = f.LfGlobal.gModular.Stream.applyTransforms()
	if err != nil {
		return err
	}

	modularBuffer := f.LfGlobal.gModular.Stream.getDecodedBuffer()
	for c := 0; c < len(modularBuffer); c++ {
		cIn := c
		var scaleFactor float32
		isModularColour := f.Header.Encoding == MODULAR && c < f.GetColorChannelCount()
		isModularXYB := f.globalMetadata.XybEncoded && isModularColour
		var cOut int
		if isModularXYB {
			cOut = cMap[c]
		} else {
			cOut = c
		}
		cOut += len(f.Buffer) - len(modularBuffer)
		ecIndex := c
		if f.Header.Encoding == MODULAR {
			ecIndex -= f.globalMetadata.GetColourChannelCount()
		}
		if isModularXYB {
			scaleFactor = f.LfGlobal.lfDequant[cOut]
		} else if isModularColour && f.globalMetadata.BitDepth.ExpBits != 0 {
			scaleFactor = 1.0
		} else if isModularColour {
			step1 := f.globalMetadata.BitDepth.BitsPerSample
			step2 := ^int32(0) << step1
			step3 := ^step2
			scaleFactor = 1.0 / float32(step3)
		} else {
			scaleFactor = float32(1.0 / ^(^uint32(0) << f.globalMetadata.ExtraChannelInfo[ecIndex].BitDepth.BitsPerSample))
		}

		cOutSection := f.Buffer[cOut]

		// Have tried getting local references to commonly refered arrays (eg modularBufferCin := modularBuffer[cIn])
		// but had negative performance... unsure why.
		if isModularXYB && cIn == 2 {
			for y := uint32(0); y < f.height; y++ {

				// get reference to sub slices to do not have to repeat lookups.
				row := cOutSection[y]
				for x := uint32(0); x < f.width; x++ {
					row[x] = scaleFactor * float32(modularBuffer[0][y][x]+modularBuffer[2][y][x])
				}
			}
		} else {
			// FIXME(kpfaulkner) change Matrices to be 1D slice with helper functions
			// NOTE: have tried using goroutine pool to process each row concurrently, didn't really make a
			// noticeable difference, so have reverted back to simple embedded loops.
			height := f.bounds.Size.Height
			for y := uint32(0); y < height; y++ {
				modularBufferY := modularBuffer[cIn][y]
				width := f.bounds.Size.Width
				// get reference to sub slices to do not have to repeat lookups.
				row := cOutSection[y]
				for x := uint32(0); x < width; x++ {
					row[x] = scaleFactor * float32(modularBufferY[x])
				}
			}
		}
	}

	f.invertSubsampling()

	if f.Header.restorationFilter.gab {
		f.performGabConvolution()
	}

	if f.Header.restorationFilter.epfIterations > 0 {
		f.performEdgePreservingFilter()
	}
	return nil
}

func (f *Frame) IsVisible() bool {
	return f.Header.FrameType == REGULAR_FRAME || f.Header.FrameType == SKIP_PROGRESSIVE && (f.Header.Duration != 0 || f.Header.IsLast)
}

func (f *Frame) GetColorChannelCount() int {
	if f.globalMetadata.XybEncoded || f.Header.Encoding == VARDCT {
		return 3
	}
	return f.globalMetadata.GetColourChannelCount()
}

func (f *Frame) GetPaddedFrameSize() (util.Dimension, error) {

	factorY := 1 << util.Max(f.Header.jpegUpsamplingY...)
	factorX := 1 << util.Max(f.Header.jpegUpsamplingX...)
	var width uint32
	var height uint32
	if f.Header.Encoding == VARDCT {
		height = (f.bounds.Size.Height + 7) >> 3
		width = (f.bounds.Size.Width + 7) >> 3
	} else {
		width = f.bounds.Size.Width
		height = f.bounds.Size.Height
	}

	height = util.CeilDiv(height, uint32(factorY))
	width = util.CeilDiv(width, uint32(factorX))
	if f.Header.Encoding == VARDCT {
		return util.Dimension{
			Width:  (width * uint32(factorX)) << 3,
			Height: (height * uint32(factorY)) << 3,
		}, nil
	} else {
		return util.Dimension{
			Width:  width * uint32(factorX),
			Height: height * uint32(factorY),
		}, nil
	}
}

func (f *Frame) decodeLFGroups(lfBuffer [][][]float32) error {

	lfReplacementChannels := []*ModularChannel{}
	lfReplacementChannelIndicies := []int{}

	for i := 0; i < len(f.LfGlobal.gModular.Stream.channels); i++ {
		ch := f.LfGlobal.gModular.Stream.channels[i]
		if !ch.decoded {
			if ch.hshift >= 3 && ch.vshift >= 3 {
				lfReplacementChannelIndicies = append(lfReplacementChannelIndicies, i)
				height := f.Header.lfGroupDim >> ch.vshift
				width := f.Header.lfGroupDim >> ch.hshift
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
		frameSize, err := f.GetPaddedFrameSize()
		if err != nil {
			return err
		}
		for i, info := range replaced {
			lfHeight := frameSize.Height >> info.vshift
			lfWidth := frameSize.Width >> info.hshift
			info.origin.Y = uint32(lfGroupPos.Y) * info.size.Height
			info.origin.X = uint32(lfGroupPos.X) * info.size.Width
			info.size.Height = util.Min(info.size.Height, lfHeight-info.origin.Y)
			info.size.Width = util.Min(info.size.Width, lfWidth-info.origin.X)
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
			channel := f.LfGlobal.gModular.Stream.channels[index]
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
	f.passes = make([]Pass, f.Header.passes.numPasses)
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

func (f *Frame) decodePassGroupsSerial() error {
	numPasses := len(f.passes)
	numGroups := int(f.numGroups)
	passGroups := util.MakeMatrix2D[PassGroup](numPasses, numGroups)

	var eg errgroup.Group
	for pass0 := 0; pass0 < numPasses; pass0++ {
		pass := pass0

		for group0 := 0; group0 < numGroups; group0++ {

			iPass := pass0
			iGroup := group0
			eg.Go(func() error {
				group := iGroup
				br, err := f.getBitreader(2 + int(f.numLFGroups) + iPass*int(f.numGroups) + group)
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
					passGroupSize := util.NewIntPoint(int(f.Header.groupDim)).ShiftRightWithIntPoint(shift)
					rowStride := util.CeilDiv(uint32(info.size.Width), passGroupSize.X)
					pos := util.Coordinates(uint32(group), rowStride).TimesWithIntPoint(passGroupSize)
					chanSize := util.NewIntPointWithXY(uint32(info.size.Width), uint32(info.size.Height))
					info.origin = pos
					size := passGroupSize.Min(chanSize.Minus(info.origin))
					info.size.Width = size.X
					info.size.Height = size.Y
					replaced[i] = info
				}

				pg, err := NewPassGroupWithReader(br, f, uint32(iPass), uint32(group), replaced)
				if err != nil {
					return err
				}
				passGroups[pass][group] = *pg
				return nil
			})
		}
		if err := eg.Wait(); err != nil {
			return err
		}
	}

	for pass := 0; pass < numPasses; pass++ {
		j := 0
		for i := 0; i < len(f.passes[pass].replacedChannels); i++ {
			ii := i
			jj := j
			eg.Go(func() error {
				channel := f.LfGlobal.gModular.Stream.channels[ii]
				channel.allocate()
				for group := 0; group < int(f.numGroups); group++ {
					newChannelInfo := passGroups[pass][group].modularStream.channels[jj]
					buff := newChannelInfo.buffer
					for y := 0; y < len(buff); y++ {
						idx := y + int(newChannelInfo.origin.Y)
						copy(channel.buffer[idx][newChannelInfo.origin.X:], buff[y][:len(buff[y])])
					}
				}
				return nil
			})
			j++
		}
		if err := eg.Wait(); err != nil {
			return err
		}
	}
	if f.Header.Encoding == VARDCT {
		panic("VARDCT not implemented")
	}

	return nil
}

func (f *Frame) doProcessing(iPass int, iGroup int, passGroups [][]PassGroup) error {

	br, err := f.getBitreader(2 + int(f.numLFGroups) + iPass*int(f.numGroups) + iGroup)
	if err != nil {
		return err
	}

	replaced := []ModularChannel{}
	for _, r := range f.passes[iPass].replacedChannels {
		mc := NewModularChannelFromChannel(r)
		replaced = append(replaced, *mc)
	}
	for i := 0; i < len(replaced); i++ {
		info := replaced[i]
		shift := util.NewIntPointWithXY(uint32(info.hshift), uint32(info.vshift))
		passGroupSize := util.NewIntPoint(int(f.Header.groupDim)).ShiftRightWithIntPoint(shift)
		rowStride := util.CeilDiv(uint32(info.size.Width), passGroupSize.X)
		pos := util.Coordinates(uint32(iGroup), rowStride).TimesWithIntPoint(passGroupSize)
		chanSize := util.NewIntPointWithXY(uint32(info.size.Width), uint32(info.size.Height))
		info.origin = pos
		size := passGroupSize.Min(chanSize.Minus(info.origin))
		info.size.Width = size.X
		info.size.Height = size.Y
		replaced[i] = info
	}

	pg, err := NewPassGroupWithReader(br, f, uint32(iPass), uint32(iGroup), replaced)
	if err != nil {
		return err
	}
	passGroups[iPass][iGroup] = *pg
	return nil

}

func (f *Frame) decodePassGroupsConcurrent() error {
	numPasses := len(f.passes)
	numGroups := int(f.numGroups)
	passGroups := util.MakeMatrix2D[PassGroup](numPasses, numGroups)

	type Inp struct {
		iPass  int
		iGroup int
	}
	inputChan := make(chan Inp, numPasses*numGroups)
	var wg sync.WaitGroup

	numberOfWorkers := 40
	for i := 0; i < numberOfWorkers; i++ {
		wg.Add(1)
		go func() {
			for inp := range inputChan {
				f.doProcessing(inp.iPass, inp.iGroup, passGroups)
			}
			defer wg.Done()
		}()
	}
	for pass0 := 0; pass0 < numPasses; pass0++ {
		pass := pass0

		for group0 := 0; group0 < numGroups; group0++ {
			inputChan <- Inp{
				iPass:  pass,
				iGroup: group0,
			}
		}
	}
	close(inputChan)
	wg.Wait()

	var eg errgroup.Group

	for pass := 0; pass < numPasses; pass++ {
		j := 0
		for i := 0; i < len(f.passes[pass].replacedChannels); i++ {
			ii := i
			jj := j
			eg.Go(func() error {
				channel := f.LfGlobal.gModular.Stream.channels[ii]
				channel.allocate()
				for group := 0; group < int(f.numGroups); group++ {
					newChannelInfo := passGroups[pass][group].modularStream.channels[jj]
					buff := newChannelInfo.buffer
					for y := 0; y < len(buff); y++ {
						idx := y + int(newChannelInfo.origin.Y)
						copy(channel.buffer[idx][newChannelInfo.origin.X:], buff[y][:len(buff[y])])
					}
				}
				return nil
			})
			j++
		}
		if err := eg.Wait(); err != nil {
			return err
		}
	}
	if f.Header.Encoding == VARDCT {
		panic("VARDCT not implemented")
	}

	return nil
}

func (f *Frame) invertSubsampling() {
	for c := 0; c < 3; c++ {
		xShift := f.Header.jpegUpsamplingX[c]
		yShift := f.Header.jpegUpsamplingY[c]
		for xShift > 0 {
			xShift--
			oldChannel := f.Buffer[c]
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
			f.Buffer[c] = newChannel
		}
		for yShift > 0 {
			yShift--
			oldChannel := f.Buffer[c]
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
			f.Buffer[c] = newChannel
		}
	}
}

func (f *Frame) performGabConvolution() error {
	panic("not implemented")
}

func (f *Frame) performEdgePreservingFilter() error {
	panic("not implemented")
}

func (f *Frame) InitializeNoise(seed0 int64) error {
	if f.LfGlobal.noiseParameters == nil || len(f.LfGlobal.noiseParameters) == 0 {
		return nil
	}
	// FIXME(kpfaulkner) yet to do.
	panic("not implemented")

	//rowStride := util.CeilDiv(f.Header.Width, f.Header.groupDim)
	//localNoiseBuffer := util.MakeMatrix3D[float32](3, int(f.Header.Height), int(f.Header.Width))
	//numGroups := rowStride * util.CeilDiv(f.Header.Height, f.Header.groupDim)
	//for group := uint32(0); group < numGroups; group++ {
	//	groupXYUp := util.Coordinates(group, rowStride).Times(f.Header.Upsampling)
	//	for iy := uint32(0); iy < f.Header.Upsampling; iy++ {
	//		for ix := uint32(0); ix < f.Header.Upsampling; ix++ {
	//			x0 := (groupXYUp.X + ix) * f.Header.groupDim
	//			y0 := (groupXYUp.Y + iy) * f.Header.groupDim
	//
	//		}
	//	}
	//}
}

func (f *Frame) GetImageSample(c int32, x int32, y int32) float32 {

	frameY := y - f.bounds.Origin.Y
	frameX := x - f.bounds.Origin.X

	if frameY < 0 || frameX < 0 || frameY >= int32(f.bounds.Size.Height) || frameX >= int32(f.bounds.Size.Width) {
		return 0
	}
	return f.Buffer[c][frameY][frameX]
}

func (f *Frame) Upsample() error {
	var err error
	for c := 0; c < len(f.Buffer); c++ {
		f.Buffer[c], err = f.performUpsampling(f.Buffer[c], c)
		if err != nil {
			return err
		}
	}
	f.bounds.Size.Height *= f.Header.Upsampling
	f.bounds.Size.Width *= f.Header.Upsampling

	f.bounds.Origin.Y *= int32(f.Header.Upsampling)
	f.bounds.Origin.X *= int32(f.Header.Upsampling)
	return nil
}

func (f *Frame) performUpsampling(buffer [][]float32, c int) ([][]float32, error) {

	colour := f.GetColorChannelCount()
	var k uint32
	if c < colour {
		k = f.Header.Upsampling
	} else {
		k = f.Header.EcUpsampling[c-colour]
	}
	if k == 1 {
		return buffer, nil
	}

	// FIXME(kpfaulkner) not implemented
	panic("not implemented")
}

func (f *Frame) RenderSplines() error {
	if f.LfGlobal.splines == nil {
		return nil
	}

	panic("RenderSplines not implemented")
}

func (f *Frame) SynthesizeNoise() error {
	if f.LfGlobal.noiseParameters == nil {
		return nil
	}

	panic("SynthesizeNoise not implemented")
}

func (f *Frame) getLFGroupSize(lfGroupID int32) (util.Dimension, error) {
	pos := f.getLFGroupLocation(lfGroupID)
	paddedSize, err := f.GetPaddedFrameSize()
	if err != nil {
		return util.Dimension{}, err
	}

	height := util.Min(f.Header.lfGroupDim, paddedSize.Height-uint32(pos.Y)*f.Header.lfGroupDim)
	width := util.Min(f.Header.lfGroupDim, paddedSize.Width-uint32(pos.X)*f.Header.lfGroupDim)
	return util.Dimension{
		Height: height,
		Width:  width,
	}, nil
}

func (f *Frame) getLFGroupLocation(lfGroupID int32) *util.Point {
	return util.NewPoint(lfGroupID/int32(f.lfGroupRowStride), lfGroupID%int32(f.lfGroupRowStride))
}

func (f *Frame) getGroupLocation(groupID int32) *util.Point {
	return util.NewPoint(groupID/int32(f.groupRowStride), groupID%int32(f.groupRowStride))
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
