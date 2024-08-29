package core

import (
	"errors"

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
	buffers [][]uint8
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

	f.permutatedTOC = f.reader.TryReadBool()
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
		for i := 0; i < int(tocEntries); i++ {
			f.tocPermutation[i] = uint32(i)
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
	buffer, err := f.reader.ReadBytesToSlice(uint64(length) + 4)
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
		_, err := f.reader.ReadBytesToSlice(uint64(f.tocLengths[i]))
		if err != nil {
			return err
		}
	}
	return nil
}

func (f *Frame) decodeFrame(lfBuffer [][][]float32) error {

	// TODO(kpfaulkner)
	return nil
}
