package frame

import (
	"errors"

	"github.com/kpfaulkner/jxl-go/bundle"
	"github.com/kpfaulkner/jxl-go/entropy"
	"github.com/kpfaulkner/jxl-go/image"
	"github.com/kpfaulkner/jxl-go/jxlio"
	"github.com/kpfaulkner/jxl-go/util"
)

type FakeFramer struct {
	lfGroup                *LFGroup
	hfGlobal               *HFGlobal
	lfGlobal               *LFGlobal
	header                 *FrameHeader
	passes                 []Pass
	groupSize              *util.Dimension
	groupPosInLFGroupPoint *util.Point
	imageHeader            *bundle.ImageHeader
}

func (f *FakeFramer) getLFGroupSize(lfGroupID int32) (util.Dimension, error) {
	//TODO implement me
	return util.Dimension{
		Width:  5,
		Height: 5,
	}, nil
}

func (f *FakeFramer) getNumLFGroups() uint32 {
	return 0
}

func (f *FakeFramer) getLFGroupLocation(lfGroupID int32) *util.Point {
	//TODO implement me
	panic("implement me")
}

func (f *FakeFramer) getGlobalTree() *MATreeNode {
	t := &MATreeNode{stream: &entropy.EntropyStream{}}
	return t
}
func (f *FakeFramer) setGlobalTree(tree *MATreeNode) {}

func (f *FakeFramer) getLFGroupForGroup(groupID int32) *LFGroup {
	return f.lfGroup
}

func (f *FakeFramer) getHFGlobal() *HFGlobal {
	return f.hfGlobal
}

func (f *FakeFramer) getLFGlobal() *LFGlobal {
	return f.lfGlobal
}

func (f *FakeFramer) getFrameHeader() *FrameHeader {
	return f.header
}

func (f *FakeFramer) getPasses() []Pass {
	return f.passes
}

func (f *FakeFramer) getGroupSize(groupID int32) (util.Dimension, error) {
	return *f.groupSize, nil
}

func (f *FakeFramer) groupPosInLFGroup(lfGroupID int32, groupID uint32) util.Point {
	return *f.groupPosInLFGroupPoint
}

func (f *FakeFramer) getGlobalMetadata() *bundle.ImageHeader {
	return f.imageHeader
}

func NewFakeFramer(encoding uint32) Framer {
	ff := &FakeFramer{
		header: &FrameHeader{
			jpegUpsamplingX: []int32{0, 0, 0},
			jpegUpsamplingY: []int32{0, 0, 0},
			Bounds: &util.Rectangle{
				Origin: util.Point{},
				Size:   util.Dimension{5, 5},
			},
			passes:   NewPassesInfo(),
			Encoding: encoding,
		},
		lfGlobal: NewLFGlobal(),
		hfGlobal: &HFGlobal{
			numHFPresets: 1,
		},
		imageHeader: &bundle.ImageHeader{
			ExtraChannelInfo: []bundle.ExtraChannelInfo{{
				DimShift: 0,
			}},
		},
	}
	ff.lfGlobal.scaledDequant = []float32{1, 1, 1}
	return ff
}

func NewFakeHFBlockContextFunc(reader jxlio.BitReader, readClusterMap func(reader jxlio.BitReader, clusterMap []int, maxClusters int) (int, error)) (*HFBlockContext, error) {
	return nil, nil
}

func NewFakeHFMetadataFunc(reader jxlio.BitReader, parent *LFGroup, frame Framer) (*HFMetadata, error) {

	return &HFMetadata{}, nil
}

func NewFakeLFCoeffientsFunc(reader jxlio.BitReader, parent *LFGroup, frame Framer, lfBuffer []image.ImageBuffer, modularStreamFunc NewModularStreamFunc) (*LFCoefficients, error) {
	return &LFCoefficients{}, nil
}

type FakeEntropyStreamer struct {
	FakeSymbols []int32
}

func (f FakeEntropyStreamer) GetDists() []entropy.SymbolDistribution {
	return nil
}

func (f FakeEntropyStreamer) GetState() *entropy.ANSState {
	//TODO implement me
	panic("implement me")
}

func (f *FakeEntropyStreamer) ReadSymbol(reader jxlio.BitReader, context int) (int32, error) {

	if len(f.FakeSymbols) == 0 {
		return 0, errors.New("no more symbols")
	}

	symbol := f.FakeSymbols[0]
	f.FakeSymbols = f.FakeSymbols[1:]
	return symbol, nil
}

func (f FakeEntropyStreamer) TryReadSymbol(reader jxlio.BitReader, context int) int32 {
	return 0
}

func (f FakeEntropyStreamer) ReadSymbolWithMultiplier(reader jxlio.BitReader, context int, distanceMultiplier int32) (int32, error) {
	return 0, nil
}

func (f FakeEntropyStreamer) ReadHybridInteger(reader jxlio.BitReader, config *entropy.HybridIntegerConfig, token int32) (int32, error) {
	//TODO implement me
	panic("implement me")
}

func (f FakeEntropyStreamer) ValidateFinalState() bool {
	return true
}

func NewFakeEntropyStreamer() entropy.EntropyStreamer {
	return &FakeEntropyStreamer{}
}

func NewFakeEntropyStreamerFunc(reader jxlio.BitReader, numDists int, readClusterMapFunc entropy.ReadClusterMapFunc) (entropy.EntropyStreamer, error) {
	return &FakeEntropyStreamer{}, nil
}

func NewFakeEntropyWithReaderFunc(reader jxlio.BitReader, numDists int, disallowLZ77 bool, readClusterMapFunc func(reader jxlio.BitReader, clusterMap []int, maxClusters int) (int, error)) (entropy.EntropyStreamer, error) {
	return &FakeEntropyStreamer{}, nil
}
