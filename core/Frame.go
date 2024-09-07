package core

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/kpfaulkner/jxl-go/entropy"
	"github.com/kpfaulkner/jxl-go/jxlio"
	"github.com/kpfaulkner/jxl-go/util"
)

type Frame struct {
	globalMetadata   *ImageHeader
	options          *JXLOptions
	reader           *jxlio.Bitreader
	header           *FrameHeader
	width            uint32
	height           uint32
	groupRowStride   uint32
	lfGroupRowStride uint32
	numGroups        uint32
	numLFGroups      uint32
	permutatedTOC    bool
	tocPermutation   []uint32
	tocLengths       []uint32

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
	f.width = f.header.width
	f.height = f.header.height
	f.width = util.CeilDiv(f.width, f.header.upsampling)
	f.height = util.CeilDiv(f.height, f.header.upsampling)
	f.width = util.CeilDiv(f.width, 1<<(f.header.lfLevel*3))
	f.height = util.CeilDiv(f.height, 1<<(f.header.lfLevel*3))
	f.groupRowStride = util.CeilDiv(f.width, f.header.groupDim)
	f.lfGroupRowStride = util.CeilDiv(f.width, f.header.groupDim<<3)
	f.numGroups = f.groupRowStride * util.CeilDiv(f.height, f.header.groupDim)
	f.numLFGroups = f.lfGroupRowStride * util.CeilDiv(f.height, f.header.groupDim<<3)
	f.readTOC()
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
	fmt.Printf("XXXX getBitreader index %d : perm %d\n", index, permutedIndex)
	return jxlio.NewBitreader(bytes.NewReader(f.buffers[permutedIndex]), true), nil
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
			shiftedSize := paddedSize.ShiftRightWithIntPoint(f.header.jpegUpsampling[c])
			f.buffer[c] = util.MakeMatrix2D[float32](int(shiftedSize.Y), int(shiftedSize.X))
		} else {
			f.buffer[c] = util.MakeMatrix2D[float32](int(paddedSize.Y), int(paddedSize.X))
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
			cOut = f.cMap[c]
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
			scaleFactor = 1.0 / ^(^0 << f.globalMetadata.bitDepth.bitsPerSample)
		}
		if isModularXYB && cIn == 2 {
			for y := uint32(0); y < f.height; y++ {
				for x := uint32(0); x < f.width; x++ {
					f.buffer[cOut][y][x] = scaleFactor * float32(modularBuffer[0][y][x]+modularBuffer[2][y][x])
				}
			}
		} else {
			for y := uint32(0); y < f.height; y++ {
				for x := uint32(0); x < f.width; x++ {
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

	panic("boom")
	// TODO(kpfaulkner)
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

func (f *Frame) getPaddedFrameSize() (util.IntPoint, error) {

	if f.header.encoding == VARDCT {
		return util.NewIntPointWithXY(f.width, f.height).CeilDiv(8).Times(8), nil
	} else {
		return util.NewIntPointWithXY(f.width, f.height), nil
	}
}

func (f *Frame) decodeLFGroups(lfBuffer [][][]float32) error {

	//lfReplacementChannels := []ModularChannelInfo{}
	//lfReplacementCHannelIndicies := []int{}

	//for i:=0;i<f.lfGlobal.gModular.
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

func (f *Frame) decodePassGroups() error {
	numPasses := len(f.passes)
	passGroups := make([][]PassGroup, numPasses)

	for pass := 0; pass < numPasses; pass++ {
		for group := 4; group < int(f.numGroups); group++ {
			br, err := f.getBitreader(2 + int(f.numLFGroups) + pass*int(f.numGroups) + group)
			if err != nil {
				return err
			}
			replaced := f.passes[pass].replacedChannels
			for i := 0; i < len(replaced); i++ {
				info := replaced[i]
				shift := util.NewIntPointWithXY(uint32(info.hshift), uint32(info.vshift))
				passGroupSize := util.NewIntPoint(int(f.header.groupDim)).ShiftRightWithIntPoint(shift)
				rowStride := util.CeilDiv(uint32(info.width), passGroupSize.X)
				pos := util.Coordinates(uint32(group), rowStride).TimesWithIntPoint(passGroupSize)
				chanSize := util.NewIntPointWithXY(uint32(info.width), uint32(info.height))
				info.origin = pos
				size := passGroupSize.Min(chanSize.Minus(info.origin))
				info.width = int(size.X)
				info.height = int(size.Y)
				replaced[i] = info
			}
			pg, err := NewPassGroupWithReader(br, f, uint32(pass), uint32(group), replaced)
			if err != nil {
				return err
			}
			//f.passes[pass].replacedChannels = replaced
			passGroups[pass][group] = *pg
			fmt.Printf("%v\n", br)
		}
	}

	for pass := 0; pass < numPasses; pass++ {
		j := 0
		fmt.Printf("%v\n", j)
		for i := 0; i < len(f.passes[pass].replacedChannels); i++ {
			//if f.passes[pass].replacedChannels[i] == nil {
			//	continue
			//}
			channel, ok := f.lfGlobal.gModular.stream.channels[i].(*ModularChannel)
			if !ok {
				return errors.New("trying to get ModularChannel when one didn't exist")
			}

			//public static native void arraycopy(Object src,  int  srcPos,
			//	Object dest, int destPos,
			//	int length);
			//
			for group := 0; group < int(f.numGroups); group++ {
				newChannelInfo := passGroups[pass][group].modularPassGroupInfo[j]
				buff := passGroups[pass][group].modularPassGroupBuffer[j]
				for y := 0; y < newChannelInfo.height; y++ {
					channel.buffer[y+int(newChannelInfo.origin.Y)] = buff[y]
				}
			}
			f.lfGlobal.gModular.stream.channels[i] = channel
			j++
		}
	}
	if f.header.encoding == VARDCT {
		panic("VARDCT not implemented")
	}

	return nil
}

func (f *Frame) decodePassGroupsORIG() error {
	numPasses := len(f.passes)
	passGroups := make([][]PassGroup, numPasses)

	for pass := 0; pass < numPasses; pass++ {
		for group := 0; group < int(f.numGroups); group++ {
			br, err := f.getBitreader(2 + int(f.numLFGroups) + pass*int(f.numGroups) + group)
			if err != nil {
				return err
			}
			replaced := f.passes[pass].replacedChannels
			for i := 0; i < len(replaced); i++ {
				info := replaced[i]
				shift := util.NewIntPointWithXY(uint32(info.hshift), uint32(info.vshift))
				passGroupSize := util.NewIntPoint(int(f.header.groupDim)).ShiftRightWithIntPoint(shift)
				rowStride := util.CeilDiv(uint32(info.width), passGroupSize.X)
				pos := util.Coordinates(uint32(group), rowStride).TimesWithIntPoint(passGroupSize)
				chanSize := util.NewIntPointWithXY(uint32(info.width), uint32(info.height))
				info.origin = pos
				size := passGroupSize.Min(chanSize.Minus(info.origin))
				info.width = int(size.X)
				info.height = int(size.Y)
				replaced[i] = info
			}
			pg, err := NewPassGroupWithReader(br, f, uint32(pass), uint32(group), replaced)
			if err != nil {
				return err
			}
			//f.passes[pass].replacedChannels = replaced
			passGroups[pass][group] = *pg
			fmt.Printf("%v\n", br)
		}
	}

	for pass := 0; pass < numPasses; pass++ {
		j := 0
		fmt.Printf("%v\n", j)
		for i := 0; i < len(f.passes[pass].replacedChannels); i++ {
			//if f.passes[pass].replacedChannels[i] == nil {
			//	continue
			//}
			channel, ok := f.lfGlobal.gModular.stream.channels[i].(*ModularChannel)
			if !ok {
				return errors.New("trying to get ModularChannel when one didn't exist")
			}

			//public static native void arraycopy(Object src,  int  srcPos,
			//	Object dest, int destPos,
			//	int length);
			//
			for group := 0; group < int(f.numGroups); group++ {
				newChannelInfo := passGroups[pass][group].modularPassGroupInfo[j]
				buff := passGroups[pass][group].modularPassGroupBuffer[j]
				for y := 0; y < newChannelInfo.height; y++ {
					channel.buffer[y+int(newChannelInfo.origin.Y)] = buff[y]
				}
			}
			f.lfGlobal.gModular.stream.channels[i] = channel
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
		xShift := f.header.jpegUpsampling[c].X
		yShift := f.header.jpegUpsampling[c].Y
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
					xx := len(oldRow) - 1
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

	//rowStride := util.CeilDiv(f.header.width, f.header.groupDim)
	//localNoiseBuffer := util.MakeMatrix3D[float32](3, int(f.header.height), int(f.header.width))
	//numGroups := rowStride * util.CeilDiv(f.header.height, f.header.groupDim)
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
