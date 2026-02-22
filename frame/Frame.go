package frame

import (
	"bytes"
	"errors"
	"fmt"
	"math"
	"sync"

	log "github.com/sirupsen/logrus"

	"github.com/kpfaulkner/jxl-go/bundle"
	"github.com/kpfaulkner/jxl-go/entropy"
	"github.com/kpfaulkner/jxl-go/image"
	"github.com/kpfaulkner/jxl-go/jxlio"
	"github.com/kpfaulkner/jxl-go/options"
	"github.com/kpfaulkner/jxl-go/util"
)

var (
	cMap = []int{1, 0, 2}

	SQRT_H = math.Sqrt(0.5)

	epfCross = []util.Point{
		{Y: 0, X: 0},
		{Y: 0, X: -1}, {Y: 0, X: 1},
		{Y: -1, X: 0}, {Y: 1, X: 0}}

	epfDoubleCross = []util.Point{{X: 0, Y: 0},
		{X: 0, Y: -1}, {X: 0, Y: 1}, {X: -1, Y: 0}, {X: 1, Y: 0},
		{X: -1, Y: 1}, {X: 1, Y: 1}, {X: 1, Y: -1}, {X: -1, Y: -1},
		{X: 0, Y: -2}, {X: 0, Y: 2}, {X: 2, Y: 0}, {X: -2, Y: 0}}
)

type NewLFGlobalWithReaderFunc func(reader jxlio.BitReader, parent Framer, hfBlockContextFunc NewHFBlockContextFunc) (*LFGlobal, error)

type ReadPermutationFunc func(reader jxlio.BitReader, stream entropy.EntropyStreamer, size uint32, skip uint32) ([]uint32, error)
type Inp struct {
	iPass  int
	iGroup int
}

type Framer interface {
	getLFGroupForGroup(groupID int32) *LFGroup
	getHFGlobal() *HFGlobal
	getLFGlobal() *LFGlobal
	getFrameHeader() *FrameHeader
	getPasses() []Pass
	getGroupSize(groupID int32) (util.Dimension, error)
	groupPosInLFGroup(lfGroupID int32, groupID uint32) util.Point
	getGlobalMetadata() *bundle.ImageHeader
	getLFGroupLocation(lfGroupID int32) *util.Point
	getGlobalTree() *MATreeNode
	setGlobalTree(tree *MATreeNode)
	getLFGroupSize(lfGroupID int32) (util.Dimension, error)
	getNumLFGroups() uint32
}

type Frame struct {
	tocPermutation []uint32
	tocLengths     []uint32
	lfGroups       []*LFGroup
	Buffer         []image.ImageBuffer
	passes         []Pass
	bitreaders     []jxlio.BitReader
	GlobalMetadata *bundle.ImageHeader
	options        *options.JXLOptions
	reader         jxlio.BitReader
	Header         *FrameHeader
	globalTree     *MATreeNode
	hfGlobal       *HFGlobal
	LfGlobal       *LFGlobal

	groupRowStride   uint32
	lfGroupRowStride uint32
	numGroups        uint32
	numLFGroups      uint32
	permutatedTOC    bool

	decoded bool
}

func (f *Frame) getGlobalTree() *MATreeNode {
	return f.globalTree
}
func (f *Frame) setGlobalTree(tree *MATreeNode) {
	f.globalTree = tree
}

func (f *Frame) getGlobalMetadata() *bundle.ImageHeader {
	return f.GlobalMetadata
}

func (f *Frame) getPasses() []Pass {
	return f.passes
}

func (f *Frame) getFrameHeader() *FrameHeader {
	return f.Header
}

func (f *Frame) ReadFrameHeader() (FrameHeader, error) {

	err := f.reader.ZeroPadToByte()
	if err != nil {
		return FrameHeader{}, fmt.Errorf("error zero padding to byte: %v", err)
	}
	f.Header, err = NewFrameHeaderWithReader(f.reader, f.GlobalMetadata)
	if err != nil {
		return FrameHeader{}, err
	}

	f.Header.Bounds = &util.Rectangle{
		Origin: f.Header.Bounds.Origin,
		Size:   f.Header.Bounds.Size,
	}

	f.groupRowStride = util.CeilDiv(f.Header.Bounds.Size.Width, f.Header.groupDim)
	f.lfGroupRowStride = util.CeilDiv(f.Header.Bounds.Size.Width, f.Header.groupDim<<3)
	f.numGroups = f.groupRowStride * util.CeilDiv(f.Header.Bounds.Size.Height, f.Header.groupDim)
	f.numLFGroups = f.lfGroupRowStride * util.CeilDiv(f.Header.Bounds.Size.Height, f.Header.groupDim<<3)

	return *f.Header, nil
}

func (f *Frame) ReadTOC() error {
	var tocEntries uint32

	if f.numGroups == 1 && f.Header.passes.numPasses == 1 {
		tocEntries = 1
	} else {
		tocEntries = 1 + f.numLFGroups + 1 + f.numGroups*f.Header.passes.numPasses
	}

	var err error
	if f.permutatedTOC, err = f.reader.ReadBool(); err != nil {
		return err
	}
	if f.permutatedTOC {
		tocStream, err := entropy.NewEntropyStreamWithReaderAndNumDists(f.reader, 8, entropy.ReadClusterMap)
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
		f.tocPermutation = nil
		//f.tocPermutation = make([]uint32, tocEntries)
		//for i := uint32(0); i < tocEntries; i++ {
		//	a := i
		//	f.tocPermutation[i] = a
		//}
	}
	if err = f.reader.ZeroPadToByte(); err != nil {
		return err
	}
	f.tocLengths = make([]uint32, tocEntries)

	for i := 0; i < int(tocEntries); i++ {
		if tocLengths, err := f.reader.ReadU32(0, 10, 1024, 14, 17408, 22, 4211712, 30); err != nil {
			return err
		} else {
			f.tocLengths[i] = tocLengths
		}
	}

	if err = f.reader.ZeroPadToByte(); err != nil {
		return err
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

// absFloat32 returns the absolute value of a float32 using bit manipulation.
// This is ~3x faster than float32(math.Abs(float64(x))) as it avoids:
// 1. float32 -> float64 conversion
// 2. float64 -> float32 conversion
// 3. The overhead of math.Abs function call
func absFloat32(x float32) float32 {
	return math.Float32frombits(math.Float32bits(x) & 0x7FFFFFFF)
}

func readPermutation(reader jxlio.BitReader, stream entropy.EntropyStreamer, size uint32, skip uint32) ([]uint32, error) {
	end, err := stream.ReadSymbol(reader, ctxFunc(int64(size)))
	if err != nil {
		return nil, err
	}

	if uint32(end) > size-skip {
		return nil, errors.New("illegal end value in lehmer sequence")
	}

	lehmer := make([]uint32, size)
	for i := skip; i < uint32(end)+skip; i++ {
		ctxVal := uint32(0)
		if i > skip {
			ctxVal = lehmer[i-1]
		}
		ii, err := stream.ReadSymbol(reader, ctxFunc(int64(ctxVal)))
		if err != nil {
			return nil, err
		}
		lehmer[i] = uint32(ii)
		if lehmer[i] >= size-i {
			return nil, errors.New("illegal end value in lehmer sequence")
		}
	}

	// Convert Lehmer code to permutation using Fenwick tree for O(n log n)
	// instead of O(n²) slice splicing
	permutation := make([]uint32, size)

	// Fenwick tree: tree[i] stores count of available elements in its range
	// Initialize with all 1s (all elements available)
	tree := make([]int, size+1)
	for i := 1; i <= int(size); i++ {
		tree[i] = i & (-i) // lowbit(i) = count of 1s in range for all-1s initialization
	}

	// Find k-th available element (1-indexed) using binary search on Fenwick tree
	findKth := func(k int) int {
		pos := 0
		mask := 1
		for mask <= int(size) {
			mask <<= 1
		}
		for mask >>= 1; mask > 0; mask >>= 1 {
			if pos+mask <= int(size) && tree[pos+mask] < k {
				k -= tree[pos+mask]
				pos += mask
			}
		}
		return pos + 1
	}

	// Mark element at pos as used (subtract 1 from all ranges containing pos)
	markUsed := func(pos int) {
		for pos <= int(size) {
			tree[pos]--
			pos += pos & (-pos)
		}
	}

	for i := uint32(0); i < size; i++ {
		// Find the (lehmer[i]+1)-th available element
		val := findKth(int(lehmer[i]) + 1)
		permutation[i] = uint32(val - 1) // Convert to 0-indexed
		markUsed(val)
	}

	return permutation, nil
}

func NewFrameWithReader(reader jxlio.BitReader, imageHeader *bundle.ImageHeader, options *options.JXLOptions) *Frame {

	frame := &Frame{
		GlobalMetadata: imageHeader,
		options:        options,
		reader:         reader,
	}

	return frame
}

func (f *Frame) SkipFrameData() error {
	for i := 0; i < len(f.tocLengths); i++ {
		_, err := f.reader.Skip(f.tocLengths[i])
		if err != nil {
			return err
		}
	}
	return nil
}

// gets a bit reader for each TOC entry???
func (f *Frame) getBitreader(index int) (jxlio.BitReader, error) {
	var i uint32
	if len(f.tocLengths) <= 1 {
		i = 0
	} else {
		i = uint32(index)
	}

	return f.bitreaders[i], nil
}

func (f *Frame) getHFGlobal() *HFGlobal {
	return f.hfGlobal
}

func (f *Frame) getLFGlobal() *LFGlobal {
	return f.LfGlobal
}

func (f *Frame) DecodeFrame(lfBuffer []image.ImageBuffer, newLFGlobalWithReader NewLFGlobalWithReaderFunc) error {

	if f.decoded {
		return nil
	}
	f.decoded = true

	err := f.setupBitReaders()
	if err != nil {
		return err
	}

	lfGlobalBitReader, err := f.getBitreader(0)
	if err != nil {
		return err
	}
	f.LfGlobal, err = newLFGlobalWithReader(lfGlobalBitReader, f, NewHFBlockContextWithReader)
	if err != nil {
		return err
	}

	paddedSize, err := f.GetPaddedFrameSize()
	if err != nil {
		return err
	}

	numColours := f.GetColourChannelCount()
	f.Buffer = make([]image.ImageBuffer, numColours+len(f.GlobalMetadata.ExtraChannelInfo))

	for c := 0; c < len(f.Buffer); c++ {
		channelSize := util.Dimension{
			Width:  paddedSize.Width,
			Height: paddedSize.Height,
		}
		if c < 3 && c < f.GetColourChannelCount() {
			channelSize.Height >>= f.Header.jpegUpsamplingY[c]
			channelSize.Width >>= f.Header.jpegUpsamplingX[c]
		}
		var isFloat bool
		if c < f.GetColourChannelCount() {
			isFloat = f.GlobalMetadata.XybEncoded || f.Header.Encoding == VARDCT ||
				f.GlobalMetadata.BitDepth.ExpBits != 0
		} else {
			isFloat = f.GlobalMetadata.ExtraChannelInfo[c-numColours].BitDepth.ExpBits != 0
		}
		typeToUse := image.TYPE_INT
		if isFloat {
			typeToUse = image.TYPE_FLOAT
		}
		buf, err := image.NewImageBuffer(typeToUse, int32(channelSize.Height), int32(channelSize.Width))
		if err != nil {
			return err
		}
		f.Buffer[c] = *buf
	}

	err = f.decodeLFGroups(lfBuffer)
	if err != nil {
		log.Errorf("Error decoding LFGroups %v", err)
		return err
	}

	hfGlobalReader, err := f.getBitreader(1 + int(f.numLFGroups))
	if err != nil {
		return err
	}

	if f.Header.Encoding == VARDCT {
		f.hfGlobal, err = NewHFGlobalWithReader(hfGlobalReader, f)
		if err != nil {
			return err
		}
	} else {
		f.hfGlobal = nil
	}

	err = f.decodePasses(hfGlobalReader)
	if err != nil {
		return err
	}

	// bench.jxl, after this Buffer[x].FloatBuffer is NOT zeroed out.. but jxlatte is
	err = f.decodePassGroupsConcurrent()
	if err != nil {
		return err
	}

	err = f.LfGlobal.globalModular.applyTransforms()
	if err != nil {
		return err
	}

	modularBuffer := f.LfGlobal.globalModular.getDecodedBuffer()
	for c := 0; c < len(modularBuffer); c++ {
		cIn := c
		isModularColour := f.Header.Encoding == MODULAR && c < f.GetColourChannelCount()
		isModularXYB := f.GlobalMetadata.XybEncoded && isModularColour
		var cOut int
		if isModularXYB {
			cOut = cMap[c]
		} else {
			cOut = c
		}
		cOut += len(f.Buffer) - len(modularBuffer)

		var scaleFactor float32
		if isModularXYB {
			scaleFactor = f.LfGlobal.lfDequant[cOut]
		} else {
			scaleFactor = 1.0
		}

		if isModularXYB && cIn == 2 {
			outBuffer := f.Buffer[cOut].FloatBuffer
			for y := uint32(0); y < f.Header.Bounds.Size.Height; y++ {
				for x := uint32(0); x < f.Header.Bounds.Size.Width; x++ {
					outBuffer[y][x] = scaleFactor * float32(modularBuffer[0][y][x]+modularBuffer[2][y][x])
				}
			}
		} else if f.Buffer[cOut].IsFloat() {

			outBuffer := f.Buffer[cOut].FloatBuffer
			for y := uint32(0); y < f.Header.Bounds.Size.Height; y++ {
				for x := uint32(0); x < f.Header.Bounds.Size.Width; x++ {
					outBuffer[y][x] = scaleFactor * float32(modularBuffer[cIn][y][x])
				}
			}
		} else {
			outBuffer := f.Buffer[cOut].IntBuffer
			for y := uint32(0); y < f.Header.Bounds.Size.Height; y++ {
				//copy(outBuffer[y], modularBuffer[cIn][y])
				for x := uint32(0); x < f.Header.Bounds.Size.Width; x++ {
					outBuffer[y][x] = int32(scaleFactor * float32(modularBuffer[cIn][y][x]))
				}
			}
		}
	}

	if err := f.invertSubsampling(); err != nil {
		return nil
	}

	if f.Header.restorationFilter.gab {
		if err := f.performGabConvolution(); err != nil {
			return err
		}
	}

	if f.Header.restorationFilter.epfIterations > 0 {
		if err = f.performEdgePreservingFilter(); err != nil {
			return err
		}
	}
	return nil
}

func (f *Frame) setupBitReaders() error {
	f.bitreaders = make([]jxlio.BitReader, len(f.tocLengths))
	if len(f.tocLengths) != 1 {
		for i := 0; i < len(f.tocLengths); i++ {
			buffer, err := f.readBuffer(i)
			if err != nil {
				return err
			}
			f.bitreaders[i] = jxlio.NewBitStreamReader(bytes.NewReader(buffer))
		}
	} else {
		f.bitreaders[0] = f.reader
	}
	return nil
}

func (f *Frame) IsVisible() bool {
	return f.Header.FrameType == REGULAR_FRAME || f.Header.FrameType == SKIP_PROGRESSIVE && (f.Header.Duration != 0 || f.Header.IsLast)
}

func (f *Frame) GetColourChannelCount() int {
	if f.GlobalMetadata.XybEncoded || f.Header.Encoding == VARDCT {
		return 3
	}
	return f.GlobalMetadata.GetColourChannelCount()
}

func (f *Frame) GetPaddedFrameSize() (util.Dimension, error) {

	factorY := 1 << util.Max(f.Header.jpegUpsamplingY...)
	factorX := 1 << util.Max(f.Header.jpegUpsamplingX...)
	var width uint32
	var height uint32
	if f.Header.Encoding == VARDCT {
		height = (f.Header.Bounds.Size.Height + 7) >> 3
		width = (f.Header.Bounds.Size.Width + 7) >> 3
	} else {
		width = f.Header.Bounds.Size.Width
		height = f.Header.Bounds.Size.Height
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

func (f *Frame) decodeLFGroups(lfBuffer []image.ImageBuffer) error {

	lfReplacementChannels := []*ModularChannel{}
	lfReplacementChannelIndicies := []int{}

	for i := 0; i < len(f.LfGlobal.globalModular.getChannels()); i++ {
		ch := f.LfGlobal.globalModular.getChannels()[i]
		if !ch.decoded {
			if ch.hshift >= 3 && ch.vshift >= 3 {
				lfReplacementChannelIndicies = append(lfReplacementChannelIndicies, i)
				height := f.Header.lfGroupDim >> ch.vshift
				width := f.Header.lfGroupDim >> ch.hshift
				lfReplacementChannels = append(lfReplacementChannels, NewModularChannelWithAllParams(int32(height), int32(width), ch.vshift, ch.hshift, false))
			}
		}
	}

	f.lfGroups = make([]*LFGroup, f.numLFGroups)

	// Hoist invariant computation out of the loop
	frameSize, err := f.GetPaddedFrameSize()
	if err != nil {
		return err
	}

	// Pre-compute lfHeight/lfWidth for each replacement channel (invariant per channel)
	numReplacements := len(lfReplacementChannels)
	lfHeights := make([]uint32, numReplacements)
	lfWidths := make([]uint32, numReplacements)
	templateHeights := make([]uint32, numReplacements)
	templateWidths := make([]uint32, numReplacements)
	for i, r := range lfReplacementChannels {
		lfHeights[i] = frameSize.Height >> r.vshift
		lfWidths[i] = frameSize.Width >> r.hshift
		templateHeights[i] = r.size.Height
		templateWidths[i] = r.size.Width
	}

	for lfGroupID := uint32(0); lfGroupID < f.numLFGroups; lfGroupID++ {
		reader, err := f.getBitreader(1 + int(lfGroupID))
		if err != nil {
			return err
		}

		lfGroupPos := f.getLFGroupLocation(int32(lfGroupID))

		// Pre-allocate with correct capacity, use index assignment instead of append
		replaced := make([]ModularChannel, numReplacements)
		for i, r := range lfReplacementChannels {
			// Copy the template channel directly instead of calling NewModularChannelFromChannel
			replaced[i] = ModularChannel{
				size:    r.size,
				origin:  r.origin,
				hshift:  r.hshift,
				vshift:  r.vshift,
				decoded: r.decoded,
				// buffer is nil - will be allocated by NewLFGroupWithReader if needed
			}

			// Update origin and size for this specific LF group
			replaced[i].origin.Y = lfGroupPos.Y * int32(templateHeights[i])
			replaced[i].origin.X = lfGroupPos.X * int32(templateWidths[i])
			replaced[i].size.Height = util.Min(templateHeights[i], lfHeights[i]-uint32(replaced[i].origin.Y))
			replaced[i].size.Width = util.Min(templateWidths[i], lfWidths[i]-uint32(replaced[i].origin.X))
		}

		f.lfGroups[lfGroupID], err = NewLFGroupWithReader(reader, f, int32(lfGroupID), replaced, lfBuffer, NewLFCoefficientsWithReader, NewHFMetadataWithReader)
		if err != nil {
			return err
		}
	}

	// Allocate all replacement channels once BEFORE copying data from LF groups
	// (previously was allocating redundantly for each LF group)
	channels := f.LfGlobal.globalModular.getChannels()
	for j := 0; j < len(lfReplacementChannelIndicies); j++ {
		channels[lfReplacementChannelIndicies[j]].allocate()
	}

	// Copy data from each LF group into the pre-allocated channels
	for lfGroupID := uint32(0); lfGroupID < f.numLFGroups; lfGroupID++ {
		for j := 0; j < len(lfReplacementChannelIndicies); j++ {
			index := lfReplacementChannelIndicies[j]
			channel := channels[index]
			newChannelInfo := f.lfGroups[lfGroupID].modularLFGroup.getChannels()[index]
			newChannel := newChannelInfo.buffer
			for y := 0; y < len(newChannel); y++ {
				copy(channel.buffer[int32(y)+newChannelInfo.origin.Y], newChannel[y])
			}
		}
	}
	return nil
}

func (f *Frame) decodePasses(reader jxlio.BitReader) error {

	var err error
	f.passes = make([]Pass, f.Header.passes.numPasses)
	for pass := 0; pass < len(f.passes); pass++ {
		prevMinShift := uint32(0)
		if pass > 0 {
			prevMinShift = f.passes[pass-1].minShift
		}

		f.passes[pass], err = NewPassWithReader(reader, f, uint32(pass), prevMinShift, NewHFPassWithReader)
		if err != nil {
			return err
		}
	}
	return nil
}

func (f *Frame) startWorker(inputChan chan Inp, passGroups [][]PassGroup) {
	for inp := range inputChan {
		if err := f.doProcessing(inp.iPass, inp.iGroup, passGroups); err != nil {
			log.Errorf("Error processing %v %v %v", inp.iPass, inp.iGroup, err)
		}

	}
}

func (f *Frame) doProcessing(iPass int, iGroup int, passGroups [][]PassGroup) error {

	br, err := f.getBitreader(2 + int(f.numLFGroups) + iPass*int(f.numGroups) + iGroup)
	if err != nil {
		return err
	}

	replaced := []ModularChannel{}
	for _, r := range f.passes[iPass].replacedChannels {
		// remove any replacedChannels that are nil/empty
		if r != nil {
			mc := NewModularChannelFromChannel(*r)
			replaced = append(replaced, *mc)
		}
	}

	for i := 0; i < len(replaced); i++ {
		info := replaced[i]
		groupHeight := f.Header.groupDim >> info.vshift
		groupWidth := f.Header.groupDim >> info.hshift
		rowStride := util.CeilDiv(info.size.Width, groupWidth)
		info.origin.Y = int32((uint32(iGroup) / rowStride) * groupHeight)
		info.origin.X = int32((uint32(iGroup) % rowStride) * groupWidth)
		info.size.Height = util.Min[uint32](info.size.Height-uint32(info.origin.Y), uint32(groupHeight))
		info.size.Width = util.Min[uint32](info.size.Width-uint32(info.origin.X), uint32(groupWidth))
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

	inputChan := make(chan Inp, numPasses*numGroups)

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

	wg := sync.WaitGroup{}
	for i := 0; i < f.options.MaxGoroutines; i++ {
		wg.Add(1)
		go func() {
			f.startWorker(inputChan, passGroups)
			wg.Done()
		}()
	}

	wg.Wait()

	for pass := 0; pass < numPasses; pass++ {
		j := 0
		for i := 0; i < len(f.passes[pass].replacedChannels); i++ {
			if f.passes[pass].replacedChannels[i] == nil {
				continue
			}
			ii := i
			jj := j
			channel := f.LfGlobal.globalModular.getChannels()[ii]
			channel.allocate()
			for group := 0; group < int(f.numGroups); group++ {
				newChannelInfo := passGroups[pass][group].modularStream.getChannels()[jj]
				buff := newChannelInfo.buffer
				for y := 0; y < len(buff); y++ {
					idx := y + int(newChannelInfo.origin.Y)
					copy(channel.buffer[idx][newChannelInfo.origin.X:], buff[y][:len(buff[y])])
				}
			}
			j++
		}
	}

	if f.Header.Encoding == VARDCT {

		// get floating point version of frame buffer
		buffers := util.MakeMatrix3DPooled[float32](3, 0, 0)
		defer util.ReturnMatrix3DToPool(buffers)
		for c := 0; c < 3; c++ {
			if err := f.Buffer[c].CastToFloatIfMax(^(^0 << f.GlobalMetadata.BitDepth.BitsPerSample)); err != nil {
				return err
			}
			buffers[c] = f.Buffer[c].FloatBuffer
		}

		for pass := 0; pass < numPasses; pass++ {
			for group := 0; group < numGroups; group++ {
				passGroup := &passGroups[pass][group]
				var prev *PassGroup
				if pass > 0 {
					prev = &passGroups[pass-1][group]
				} else {
					prev = nil
				}
				if err := passGroup.invertVarDCT(buffers, prev); err != nil {
					return err
				}
			}
		}

		// Release HFCoefficients buffers back to pool
		for pass := 0; pass < numPasses; pass++ {
			for group := 0; group < numGroups; group++ {
				passGroups[pass][group].Release()
			}
		}
	}

	return nil
}

// nolint
func displayBuffers(text string, frameBuffer [][][]float32) {
	total := 0.0
	for c := 0; c < len(frameBuffer); c++ {
		for y := 0; y < len(frameBuffer[c]); y++ {
			for x := 0; x < len(frameBuffer[c][y]); x++ {
				total += float64(frameBuffer[c][y][x])
			}
		}
	}
}

// nolint
func displayBuffer(text string, frameBuffer [][]float32) {
	total := 0.0

	for y := 0; y < len(frameBuffer); y++ {
		//fmt.Printf("Row %d: %v\n", y, frameBuffer[c][y])
		//fmt.Printf("Row %d: ", y)
		for x := 0; x < len(frameBuffer[y]); x++ {
			//fmt.Printf("%f ", frameBuffer[c][y][x])
			total += float64(frameBuffer[y][x])

		}
	}
}

func (f *Frame) invertSubsampling() error {
	for c := 0; c < 3; c++ {
		xShift := f.Header.jpegUpsamplingX[c]
		yShift := f.Header.jpegUpsamplingY[c]

		xShift--
		for xShift+1 > 0 {
			oldBuffer := f.Buffer[c]
			if err := oldBuffer.CastToFloatIfMax(^(^0 << f.GlobalMetadata.BitDepth.BitsPerSample)); err != nil {
				log.Errorf("Error casting buffer to float %v", err)
				return err
			}
			oldChannel := oldBuffer.FloatBuffer
			newBuffer, err := image.NewImageBuffer(image.TYPE_FLOAT, oldBuffer.Height, oldBuffer.Width*2)
			if err != nil {
				log.Errorf("Error creating new buffer %v", err)
				return err
			}
			newChannel := newBuffer.FloatBuffer
			for y := 0; y < len(oldChannel); y++ {
				oldRow := oldChannel[y]
				//newRow := make([]float32, len(oldRow)*2)
				newRow := newChannel[y]
				for x := 0; x < len(oldRow); x++ {
					b75 := 0.75 * oldRow[x]
					xx := 0
					if x != 0 {
						xx = x - 1
					}
					newRow[2*x] = b75 + 0.25*oldRow[xx]
					xx = x + 1
					if x+1 == len(oldRow) {
						xx = len(oldRow) - 1
					}
					newRow[2*x+1] = b75 + 0.25*oldRow[xx]
				}
				newChannel[y] = newRow
			}
			f.Buffer[c] = *newBuffer
			xShift--
		}
		yShift--
		for yShift+1 > 0 {
			oldBuffer := f.Buffer[c]
			if err := oldBuffer.CastToFloatIfMax(^(^0 << f.GlobalMetadata.BitDepth.BitsPerSample)); err != nil {
				log.Errorf("Error casting buffer to float %v", err)
				return err
			}
			oldChannel := oldBuffer.FloatBuffer
			newBuffer, err := image.NewImageBuffer(image.TYPE_FLOAT, oldBuffer.Height*2, oldBuffer.Width)
			if err != nil {
				log.Errorf("Error creating new buffer %v", err)
				return err
			}
			newChannel := newBuffer.FloatBuffer
			for y := 0; y < len(oldChannel); y++ {
				oldRow := oldChannel[y]
				yy := y - 1
				if y == 0 {
					yy = 0
				}
				oldRowPrev := oldChannel[yy]
				yy = y + 1
				if y+1 == len(oldChannel) {
					yy = len(oldChannel) - 1
				}
				oldRowNext := oldChannel[yy]
				firstNewRow := newChannel[2*y]
				secondNewRow := newChannel[2*y+1]
				for x := 0; x < len(oldRow); x++ {
					b75 := 0.75 * oldRow[x]
					firstNewRow[x] = b75 + 0.25*oldRowPrev[x]
					secondNewRow[x] = b75 + 0.25*oldRowNext[x]
				}
			}
			f.Buffer[c] = *newBuffer
			yShift--
		}
	}
	return nil
}

func (f *Frame) performGabConvolution() error {

	// f.Buffer[c] is already wrong at this stage. We have -0.003936676 where as should be -0.000675
	colours := f.getColourChannelCount()
	normGabBase := make([]float32, colours)
	normGabAdj := make([]float32, colours)
	normGabDiag := make([]float32, colours)
	for c := int32(0); c < colours; c++ {
		gabW1 := f.Header.restorationFilter.gab1Weights[c]
		gabW2 := f.Header.restorationFilter.gab2Weights[c]
		mult := 1.0 / (1.0 + 4.0*(gabW1+gabW2))
		normGabBase[c] = mult
		normGabAdj[c] = gabW1 * mult
		normGabDiag[c] = gabW2 * mult
	}

	for c := int32(0); c < colours; c++ {
		if err := f.Buffer[c].CastToFloatIfMax(^(^0 << f.GlobalMetadata.BitDepth.BitsPerSample)); err != nil {
			return err
		}
		height := f.Buffer[c].Height
		width := f.Buffer[c].Width
		buffC := f.Buffer[c].FloatBuffer
		newBuffer, err := image.NewImageBuffer(image.TYPE_FLOAT, height, width)
		if err != nil {
			return err
		}
		newBufferF := newBuffer.FloatBuffer

		// Capture values for closures
		gabBase := normGabBase[c]
		gabAdj := normGabAdj[c]
		gabDiag := normGabDiag[c]

		// Worker function to process a range of rows
		processRows := func(startY, endY int32) {
			for y := startY; y < endY; y++ {
				var north int32
				if y == 0 {
					north = 0
				} else {
					north = y - 1
				}
				var south int32
				if y+1 == height {
					south = height - 1
				} else {
					south = y + 1
				}

				buffR := buffC[y]
				buffN := buffC[north]
				buffS := buffC[south]
				newBuffR := newBufferF[y]

				for x := int32(0); x < width; x++ {
					var west int32
					if x == 0 {
						west = 0
					} else {
						west = x - 1
					}
					var east int32
					if x+1 == width {
						east = width - 1
					} else {
						east = x + 1
					}
					adj := buffR[west] + buffR[east] + buffN[x] + buffS[x]
					diag := buffN[west] + buffN[east] + buffS[west] + buffS[east]
					newBuffR[x] = gabBase*buffR[x] + gabAdj*adj + gabDiag*diag
				}
			}
		}

		// Divide work among goroutines
		numWorkers := f.options.MaxGoroutines
		if numWorkers < 1 {
			numWorkers = 1
		}
		rowsPerWorker := (height + int32(numWorkers) - 1) / int32(numWorkers)

		var wg sync.WaitGroup
		for w := 0; w < numWorkers; w++ {
			startY := int32(w) * rowsPerWorker
			endY := startY + rowsPerWorker
			if endY > height {
				endY = height
			}
			if startY >= height {
				break
			}
			wg.Add(1)
			go func(sy, ey int32) {
				defer wg.Done()
				processRows(sy, ey)
			}(startY, endY)
		}
		wg.Wait()

		f.Buffer[c] = *newBuffer
	}
	return nil
}

func (f *Frame) performEdgePreservingFilter() error {

	stepMultiplier := float32(1.65) * 4 * float32(1-SQRT_H)
	paddedSize, err := f.GetPaddedFrameSize()
	if err != nil {
		return err
	}
	blockHeight := (paddedSize.Height + 7) >> 3
	blockWidth := (paddedSize.Width + 7) >> 3
	inverseSigma := util.MakeMatrix2DPooled[float32](int(blockHeight), int(blockWidth))
	defer util.ReturnMatrix2DToPool(inverseSigma)
	colours := f.getColourChannelCount()
	if f.Header.Encoding == MODULAR {
		inv := 1.0 / f.Header.restorationFilter.epfSigmaForModular
		for y := 0; y < int(blockHeight); y++ {
			for i := 0; i < len(inverseSigma[y]); i++ {
				inverseSigma[y][i] = inv
			}
		}
	} else {
		globalScale := float32(65536.0) / float32(f.LfGlobal.globalScale)
		for y := int32(0); y < int32(blockHeight); y++ {
			lfY := y >> 8
			bY := y - (lfY << 8)
			lfR := lfY * int32(f.lfGroupRowStride)
			for x := int32(0); x < int32(blockWidth); x++ {
				lfX := x >> 8
				bX := x - (lfX << 8)
				lfg := f.lfGroups[lfR+lfX]
				hf := lfg.hfMetadata.hfMultiplier[bY][bX]
				sharpness := lfg.hfMetadata.hfStreamBuffer[3][bY][bX]
				if sharpness < 0 || sharpness > 7 {
					return errors.New("sharpness value out of range")
				}
				sigma := globalScale * float32(f.Header.restorationFilter.epfSharpLut[sharpness]) / float32(hf)
				inverseSigma[y][x] = 1.0 / sigma
			}
		}
	}

	outputBuffer := make([]image.ImageBuffer, colours)
	for c := int32(0); c < colours; c++ {
		if err = f.Buffer[c].CastToFloatIfMax(^(^0 << f.GlobalMetadata.BitDepth.BitsPerSample)); err != nil {
			return err
		}
		outBuf, err := image.NewImageBuffer(image.TYPE_FLOAT, int32(paddedSize.Height), int32(paddedSize.Width))
		if err != nil {
			return err
		}
		outputBuffer[c] = *outBuf
	}

	// Cache channel scales to avoid repeated field lookups
	channelScales := f.Header.restorationFilter.epfChannelScale
	borderSadMul := f.Header.restorationFilter.epfBorderSadMul

	// Threshold for skipping EPF (1.0 / 0.3)
	const skipThreshold = float32(1.0 / 0.3)

	height := int32(paddedSize.Height)
	width := int32(paddedSize.Width)

	for i := 0; i < 3; i++ {
		if i == 0 && f.Header.restorationFilter.epfIterations < 3 {
			continue
		}
		if i == 2 && f.Header.restorationFilter.epfIterations < 2 {
			break
		}

		// copy first 3 (well number of colours we have) buffers
		inputBuffers := copyFloatBuffers(f.Buffer, colours)
		outputBuffers := copyFloatBuffers(outputBuffer, colours)
		defer util.ReturnMatrix3DToPool(inputBuffers)
		// Note: Don't return outputBuffers to pool - they get swapped into f.Buffer

		var sigmaScale float32
		if i == 0 {
			sigmaScale = stepMultiplier * f.Header.restorationFilter.epfPass0SigmaScale
		} else if i == 2 {
			sigmaScale = stepMultiplier * f.Header.restorationFilter.epfPass2SigmaScale
		} else {
			sigmaScale = stepMultiplier
		}

		useDoubleCross := i == 0
		useDistance2 := i == 2

		// Worker function to process a range of rows
		processRows := func(startY, endY int32) {
			// Pre-allocate per-worker buffers
			var sumChannels [3]float32 // Max 3 colour channels

			for y := startY; y < endY; y++ {
				// Check if we're on a block boundary row for border SAD multiplier
				modY := y & 0b111
				isBorderY := modY == 0 || modY == 7
				invSigmaRow := inverseSigma[y>>3]

				// Get row pointers for all channels to avoid repeated indexing
				var inRows, outRows [3][]float32
				for c := int32(0); c < colours; c++ {
					inRows[c] = inputBuffers[c][y]
					outRows[c] = outputBuffers[c][y]
				}

				for x := int32(0); x < width; x++ {
					s := invSigmaRow[x>>3]
					if s > skipThreshold {
						for c := int32(0); c < colours; c++ {
							outRows[c][x] = inRows[c][x]
						}
						continue
					}

					// Check if we're on a block boundary for border SAD multiplier
					modX := x & 0b111
					isBorder := isBorderY || modX == 0 || modX == 7
					sigmaScaleS := sigmaScale * s

					sumWeights := float32(0)
					for c := int32(0); c < colours; c++ {
						sumChannels[c] = 0
					}

					if useDoubleCross {
						// Process epfDoubleCross (13 points) with inlined distance1
						for ci := 0; ci < 13; ci++ {
							crossY := epfDoubleCross[ci].Y
							crossX := epfDoubleCross[ci].X

							// Inline epfDistance1: sum over epfCross (5 points)
							dist := float32(0)
							for c := int32(0); c < colours; c++ {
								buffC := inputBuffers[c]
								scale := channelScales[c]
								// Unroll epfCross loop (5 iterations with known offsets)
								// Point 0: (0, 0)
								pY0 := y
								pX0 := x
								dY0 := mirrorCoord(y+crossY, height)
								dX0 := mirrorCoord(x+crossX, width)
								dist += absFloat32(buffC[pY0][pX0]-buffC[dY0][dX0]) * scale

								// Point 1: (0, -1)
								pY1 := y
								pX1 := mirrorCoord(x-1, width)
								dY1 := mirrorCoord(y+crossY, height)
								dX1 := mirrorCoord(x+crossX-1, width)
								dist += absFloat32(buffC[pY1][pX1]-buffC[dY1][dX1]) * scale

								// Point 2: (0, 1)
								pY2 := y
								pX2 := mirrorCoord(x+1, width)
								dY2 := mirrorCoord(y+crossY, height)
								dX2 := mirrorCoord(x+crossX+1, width)
								dist += absFloat32(buffC[pY2][pX2]-buffC[dY2][dX2]) * scale

								// Point 3: (-1, 0)
								pY3 := mirrorCoord(y-1, height)
								pX3 := x
								dY3 := mirrorCoord(y+crossY-1, height)
								dX3 := mirrorCoord(x+crossX, width)
								dist += absFloat32(buffC[pY3][pX3]-buffC[dY3][dX3]) * scale

								// Point 4: (1, 0)
								pY4 := mirrorCoord(y+1, height)
								pX4 := x
								dY4 := mirrorCoord(y+crossY+1, height)
								dX4 := mirrorCoord(x+crossX, width)
								dist += absFloat32(buffC[pY4][pX4]-buffC[dY4][dX4]) * scale
							}

							// Inline epfWeight
							if isBorder {
								dist *= borderSadMul
							}
							weight := 1.0 - dist*sigmaScaleS
							if weight < 0 {
								weight = 0
							}
							sumWeights += weight

							mY := mirrorCoord(y+crossY, height)
							mX := mirrorCoord(x+crossX, width)
							for c := int32(0); c < colours; c++ {
								sumChannels[c] += inputBuffers[c][mY][mX] * weight
							}
						}
					} else if useDistance2 {
						// Process epfCross (5 points) with inlined distance2
						for ci := 0; ci < 5; ci++ {
							crossY := epfCross[ci].Y
							crossX := epfCross[ci].X

							// Inline epfDistance2: simple single-point distance
							dist := float32(0)
							dY := mirrorCoord(y+crossY, height)
							dX := mirrorCoord(x+crossX, width)
							for c := int32(0); c < colours; c++ {
								dist += absFloat32(inputBuffers[c][y][x]-inputBuffers[c][dY][dX]) * channelScales[c]
							}

							// Inline epfWeight
							if isBorder {
								dist *= borderSadMul
							}
							weight := 1.0 - dist*sigmaScaleS
							if weight < 0 {
								weight = 0
							}
							sumWeights += weight

							for c := int32(0); c < colours; c++ {
								sumChannels[c] += inputBuffers[c][dY][dX] * weight
							}
						}
					} else {
						// Process epfCross (5 points) with inlined distance1
						for ci := 0; ci < 5; ci++ {
							crossY := epfCross[ci].Y
							crossX := epfCross[ci].X

							// Inline epfDistance1
							dist := float32(0)
							for c := int32(0); c < colours; c++ {
								buffC := inputBuffers[c]
								scale := channelScales[c]
								// Unrolled epfCross
								pY0 := y
								pX0 := x
								dY0 := mirrorCoord(y+crossY, height)
								dX0 := mirrorCoord(x+crossX, width)
								dist += absFloat32(buffC[pY0][pX0]-buffC[dY0][dX0]) * scale

								pY1 := y
								pX1 := mirrorCoord(x-1, width)
								dY1 := mirrorCoord(y+crossY, height)
								dX1 := mirrorCoord(x+crossX-1, width)
								dist += absFloat32(buffC[pY1][pX1]-buffC[dY1][dX1]) * scale

								pY2 := y
								pX2 := mirrorCoord(x+1, width)
								dY2 := mirrorCoord(y+crossY, height)
								dX2 := mirrorCoord(x+crossX+1, width)
								dist += absFloat32(buffC[pY2][pX2]-buffC[dY2][dX2]) * scale

								pY3 := mirrorCoord(y-1, height)
								pX3 := x
								dY3 := mirrorCoord(y+crossY-1, height)
								dX3 := mirrorCoord(x+crossX, width)
								dist += absFloat32(buffC[pY3][pX3]-buffC[dY3][dX3]) * scale

								pY4 := mirrorCoord(y+1, height)
								pX4 := x
								dY4 := mirrorCoord(y+crossY+1, height)
								dX4 := mirrorCoord(x+crossX, width)
								dist += absFloat32(buffC[pY4][pX4]-buffC[dY4][dX4]) * scale
							}

							// Inline epfWeight
							if isBorder {
								dist *= borderSadMul
							}
							weight := 1.0 - dist*sigmaScaleS
							if weight < 0 {
								weight = 0
							}
							sumWeights += weight

							mY := mirrorCoord(y+crossY, height)
							mX := mirrorCoord(x+crossX, width)
							for c := int32(0); c < colours; c++ {
								sumChannels[c] += inputBuffers[c][mY][mX] * weight
							}
						}
					}

					invSumWeights := 1.0 / sumWeights
					for c := int32(0); c < colours; c++ {
						outRows[c][x] = sumChannels[c] * invSumWeights
					}
				}
			}
		}

		// Divide work among goroutines
		numWorkers := f.options.MaxGoroutines
		if numWorkers < 1 {
			numWorkers = 1
		}
		rowsPerWorker := (height + int32(numWorkers) - 1) / int32(numWorkers)

		var wg sync.WaitGroup
		for w := 0; w < numWorkers; w++ {
			startY := int32(w) * rowsPerWorker
			endY := startY + rowsPerWorker
			if endY > height {
				endY = height
			}
			if startY >= height {
				break
			}
			wg.Add(1)
			go func(sy, ey int32) {
				defer wg.Done()
				processRows(sy, ey)
			}(startY, endY)
		}
		wg.Wait()

		for c := 0; c < int(colours); c++ {
			tmp := f.Buffer[c]
			f.Buffer[c].FloatBuffer = outputBuffers[c]
			outputBuffer[c] = tmp
		}
	}
	return nil
}

// mirrorCoord is an optimized version of MirrorCoordinate for EPF hot path.
// For most pixels (not on edges), this is just a bounds check.
func mirrorCoord(coord int32, size int32) int32 {
	if coord >= 0 && coord < size {
		return coord
	}
	return util.MirrorCoordinate(coord, size)
}

func copyFloatBuffers(buffer []image.ImageBuffer, colours int32) [][][]float32 {
	data := util.MakeMatrix3DPooled[float32](int(colours), int(buffer[0].Height), int(buffer[0].Width))
	for c := int32(0); c < colours; c++ {
		for y := int32(0); y < buffer[c].Height; y++ {
			copy(data[c][y], buffer[c].FloatBuffer[y])
		}
	}
	return data
}

func (f *Frame) getNumLFGroups() uint32 {
	return f.numLFGroups
}

func (f *Frame) InitializeNoise(seed0 int64) error {
	if len(f.LfGlobal.noiseParameters) == 0 {
		return nil
	}

	return errors.New("noise not implemented")

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

func (f *Frame) Upsample() error {
	for c := 0; c < len(f.Buffer); c++ {
		if buf, err := f.performUpsampling(f.Buffer[c], c); err != nil {
			return err
		} else {
			f.Buffer[c] = *buf
		}
	}
	f.Header.Bounds.Size.Height *= f.Header.Upsampling
	f.Header.Bounds.Size.Width *= f.Header.Upsampling

	f.Header.Bounds.Origin.Y *= int32(f.Header.Upsampling)
	f.Header.Bounds.Origin.X *= int32(f.Header.Upsampling)
	f.groupRowStride = util.CeilDiv(f.Header.Bounds.Size.Width, f.Header.groupDim)
	f.lfGroupRowStride = util.CeilDiv(f.Header.Bounds.Size.Width, f.Header.groupDim<<3)
	f.numGroups = f.groupRowStride * util.CeilDiv(f.Header.Bounds.Size.Height, f.Header.groupDim)
	f.numLFGroups = f.lfGroupRowStride * util.CeilDiv(f.Header.Bounds.Size.Height, f.Header.groupDim<<3)
	return nil
}

func (f *Frame) performUpsampling(ib image.ImageBuffer, c int) (*image.ImageBuffer, error) {

	colour := f.GetColourChannelCount()
	var k uint32
	if c < colour {
		k = f.Header.Upsampling
	} else {
		k = f.Header.EcUpsampling[c-colour]
	}
	if k == 1 {
		return &ib, nil
	}

	var depth uint32
	if c < colour {
		depth = f.GlobalMetadata.BitDepth.BitsPerSample
	} else {
		depth = f.GlobalMetadata.ExtraChannelInfo[c-colour].BitDepth.BitsPerSample
	}

	if err := ib.CastToFloatIfMax(^(^0 << depth)); err != nil {
		return nil, err
	}

	buffer := ib.FloatBuffer
	l := util.CeilLog1p(k-1) - 1
	up, err := f.GlobalMetadata.GetUpWeights()
	if err != nil {
		return nil, err
	}
	upWeights := up[l]
	newBuffer := util.MakeMatrix2D[float32](len(buffer)*int(k), 0)
	for y := 0; y < len(buffer); y++ {
		for ky := 0; ky < int(k); ky++ {
			newBuffer[y*int(k)+ky] = make([]float32, len(buffer[y])*int(k))
			for x := 0; x < len(buffer[y]); x++ {
				for kx := 0; kx < int(k); kx++ {
					weights := upWeights[ky][kx]
					total := float32(0.0)
					min := float32(math.MaxFloat32)
					max := float32(math.SmallestNonzeroFloat32)
					for iy := 0; iy < 5; iy++ {
						for ix := 0; ix < 5; ix++ {
							newY := util.MirrorCoordinate(int32(y)+int32(iy)-2, int32(len(buffer)))
							newX := util.MirrorCoordinate(int32(x)+int32(ix)-2, int32(len(buffer[newY])))
							sample := buffer[newY][newX]
							if sample < min {
								min = sample
							}
							if sample > max {
								max = sample
							}
							total += weights[iy][ix] * sample
						}
					}
					var val float32
					if total < min {
						val = min
					} else if total > max {
						val = max
					} else {
						val = total
					}
					newBuffer[y*int(k)+ky][x*int(k)+kx] = val
				}
			}
		}
	}

	return image.NewImageBufferFromFloats(newBuffer), nil

}

func (f *Frame) RenderSplines() error {
	if f.LfGlobal.splines == nil {
		return nil
	}
	return errors.New("RenderSplines not implemented")
}

func (f *Frame) SynthesizeNoise() error {
	if f.LfGlobal.noiseParameters == nil {
		return nil
	}

	return errors.New("SynthesizeNoise not implemented")
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
	return f.lfGroups[idx]
}

func (f *Frame) groupPosInLFGroup(lfGroupID int32, groupID uint32) util.Point {

	gr := f.getGroupLocation(int32(groupID))
	lf := f.getLFGroupLocation(lfGroupID)
	gr2 := *gr
	gr2.Y = gr.Y - lf.Y<<3
	gr2.X = gr.X - lf.X<<3
	return gr2
}

func (f *Frame) getGroupSize(groupID int32) (util.Dimension, error) {

	pos := f.getGroupLocation(groupID)
	paddedSize, err := f.GetPaddedFrameSize()
	if err != nil {
		return util.Dimension{}, err
	}
	height := util.Min(f.Header.groupDim, paddedSize.Height-uint32(pos.Y)*f.Header.groupDim)
	width := util.Min(f.Header.groupDim, paddedSize.Width-uint32(pos.X)*f.Header.groupDim)
	return util.Dimension{
		Height: height,
		Width:  width,
	}, nil
}

func (f *Frame) getColourChannelCount() int32 {
	if f.GlobalMetadata.XybEncoded || f.Header.Encoding == VARDCT {
		return 3
	}
	return int32(f.GlobalMetadata.GetColourChannelCount())
}

// generate a total (signature?) for each row of each channel in the buffer.
// This is just to see if we can compare Go and Java
// Assume float buffer
// nolint   (skip linting for now... will be used later)
func (f *Frame) generateSignaturesForBuffer(idx int) []string {
	sigs := []string{}
	var c = f.Buffer[idx]
	for y := int32(0); y < int32(len(c.FloatBuffer)); y++ {
		sig := float64(0)
		xx := c.FloatBuffer[y]
		if y == 288 {
			var cc float32
			for x := int32(0); x < int32(len(xx)); x++ {
				cc += c.FloatBuffer[y][x]
				nanCheck := fmt.Sprintf("%.4f", cc)
				if nanCheck == "NaN" {
					fmt.Print("NAN!\n")
					checkVal := c.FloatBuffer[y][x]
					fmt.Printf("Nan check value %f\n", checkVal)
					//fmt.Printf("range %+v\n", c.FloatBuffer[y][x-10:x+10])
				}
			}
			fmt.Printf("xx %f\n", cc)
		}

		for x := int32(0); x < int32(len(xx)); x++ {
			sig += float64(c.FloatBuffer[y][x])
		}
		sigs = append(sigs, fmt.Sprintf("%.4f", sig))
	}

	return sigs
}
