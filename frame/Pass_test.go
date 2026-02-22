package frame

import (
	"errors"
	"testing"

	"github.com/kpfaulkner/jxl-go/entropy"
	"github.com/kpfaulkner/jxl-go/jxlio"
	"github.com/kpfaulkner/jxl-go/testcommon"
	"github.com/kpfaulkner/jxl-go/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeHFPassFunc returns a NewHFPassWithReaderFunc that returns a fixed HFPass or error.
func fakeHFPassFunc(hfPass *HFPass, err error) NewHFPassWithReaderFunc {
	return func(reader jxlio.BitReader, frame Framer, passIndex uint32,
		readClusterMapFunc entropy.ReadClusterMapFunc,
		newEntropyStreamWithReader entropy.EntropyStreamWithReaderAndNumDistsFunc,
		readPermutation ReadPermutationFunc) (*HFPass, error) {
		return hfPass, err
	}
}

// makeFramerForPass creates a FakeFramer configured for Pass tests.
// encoding: VARDCT or MODULAR
// channels: the ModularChannels that globalModular.getChannels() will return
// passes: optional custom PassesInfo (nil uses default)
func makeFramerForPass(encoding uint32, channels []*ModularChannel, passes *PassesInfo) Framer {
	ff := NewFakeFramer(encoding).(*FakeFramer)
	if passes != nil {
		ff.header.passes = passes
	}
	ff.lfGlobal.globalModular = &ModularStream{channels: channels}
	return ff
}

func TestNewPassWithReader_ModularPassIndex0_DefaultPasses(t *testing.T) {
	// passIndex=0, default passes (lastPass=[], downSample=[])
	// maxShift=3, no match in lastPass → minShift=maxShift=3
	channels := []*ModularChannel{
		{hshift: 0, vshift: 0, decoded: false},
	}
	framer := makeFramerForPass(MODULAR, channels, nil)
	reader := testcommon.NewFakeBitReader()

	p, err := NewPassWithReader(reader, framer, 0, 0, fakeHFPassFunc(nil, nil))
	require.NoError(t, err)
	assert.Equal(t, uint32(3), p.maxShift)
	assert.Equal(t, uint32(3), p.minShift)
	assert.Nil(t, p.hfPass)
}

func TestNewPassWithReader_PassIndex0_MaxShiftIs3(t *testing.T) {
	channels := []*ModularChannel{{decoded: true}}
	framer := makeFramerForPass(MODULAR, channels, nil)
	reader := testcommon.NewFakeBitReader()

	p, err := NewPassWithReader(reader, framer, 0, 99, fakeHFPassFunc(nil, nil))
	require.NoError(t, err)
	// passIndex==0 → maxShift=3 regardless of prevMinShift
	assert.Equal(t, uint32(3), p.maxShift)
}

func TestNewPassWithReader_PassIndexGt0_MaxShiftIsPrevMinShift(t *testing.T) {
	channels := []*ModularChannel{{decoded: true}}
	framer := makeFramerForPass(MODULAR, channels, nil)
	reader := testcommon.NewFakeBitReader()

	p, err := NewPassWithReader(reader, framer, 1, 7, fakeHFPassFunc(nil, nil))
	require.NoError(t, err)
	assert.Equal(t, uint32(7), p.maxShift)
}

func TestNewPassWithReader_MinShiftFromDownSample(t *testing.T) {
	// lastPass=[0], downSample=[4] → n=0 matched for passIndex=0
	// minShift = CeilLog1p(4-1) = CeilLog1p(3) = 2
	pi := &PassesInfo{
		numPasses:  1,
		numDS:      1,
		lastPass:   []uint32{0},
		downSample: []uint32{4},
		shift:      []uint32{0},
	}
	channels := []*ModularChannel{{decoded: true}}
	framer := makeFramerForPass(MODULAR, channels, pi)
	reader := testcommon.NewFakeBitReader()

	p, err := NewPassWithReader(reader, framer, 0, 0, fakeHFPassFunc(nil, nil))
	require.NoError(t, err)
	assert.Equal(t, uint32(3), p.maxShift)
	assert.Equal(t, uint32(2), p.minShift)
}

func TestNewPassWithReader_MinShiftFromDownSample_Value8(t *testing.T) {
	// downSample[0]=8 → CeilLog1p(8-1)=CeilLog1p(7)=3
	pi := &PassesInfo{
		numPasses:  1,
		lastPass:   []uint32{0},
		downSample: []uint32{8},
		shift:      []uint32{0},
	}
	channels := []*ModularChannel{{decoded: true}}
	framer := makeFramerForPass(MODULAR, channels, pi)
	reader := testcommon.NewFakeBitReader()

	p, err := NewPassWithReader(reader, framer, 0, 0, fakeHFPassFunc(nil, nil))
	require.NoError(t, err)
	assert.Equal(t, uint32(3), p.minShift)
}

func TestNewPassWithReader_MinShiftFromDownSample_Value2(t *testing.T) {
	// downSample[0]=2 → CeilLog1p(2-1)=CeilLog1p(1)=1
	pi := &PassesInfo{
		numPasses:  1,
		lastPass:   []uint32{0},
		downSample: []uint32{2},
		shift:      []uint32{0},
	}
	channels := []*ModularChannel{{decoded: true}}
	framer := makeFramerForPass(MODULAR, channels, pi)
	reader := testcommon.NewFakeBitReader()

	p, err := NewPassWithReader(reader, framer, 0, 0, fakeHFPassFunc(nil, nil))
	require.NoError(t, err)
	assert.Equal(t, uint32(1), p.minShift)
}

func TestNewPassWithReader_MinShiftFromDownSample_Value1(t *testing.T) {
	// downSample[0]=1 → CeilLog1p(1-1)=CeilLog1p(0)=0
	pi := &PassesInfo{
		numPasses:  1,
		lastPass:   []uint32{0},
		downSample: []uint32{1},
		shift:      []uint32{0},
	}
	channels := []*ModularChannel{{decoded: true}}
	framer := makeFramerForPass(MODULAR, channels, pi)
	reader := testcommon.NewFakeBitReader()

	p, err := NewPassWithReader(reader, framer, 0, 0, fakeHFPassFunc(nil, nil))
	require.NoError(t, err)
	assert.Equal(t, uint32(0), p.minShift)
}

func TestNewPassWithReader_NoMatchInLastPass_MinShiftEqualsMaxShift(t *testing.T) {
	// lastPass=[1], but passIndex=0 → no match → minShift=maxShift
	pi := &PassesInfo{
		numPasses:  2,
		lastPass:   []uint32{1},
		downSample: []uint32{4},
		shift:      []uint32{0, 0},
	}
	channels := []*ModularChannel{{decoded: true}}
	framer := makeFramerForPass(MODULAR, channels, pi)
	reader := testcommon.NewFakeBitReader()

	p, err := NewPassWithReader(reader, framer, 0, 0, fakeHFPassFunc(nil, nil))
	require.NoError(t, err)
	assert.Equal(t, p.maxShift, p.minShift)
}

func TestNewPassWithReader_MultipleLastPassEntries_FirstMatch(t *testing.T) {
	// lastPass=[0, 1], downSample=[4, 8] → passIndex=0 matches at index 0
	// minShift = CeilLog1p(4-1) = 2
	pi := &PassesInfo{
		numPasses:  2,
		lastPass:   []uint32{0, 1},
		downSample: []uint32{4, 8},
		shift:      []uint32{0, 0},
	}
	channels := []*ModularChannel{{decoded: true}}
	framer := makeFramerForPass(MODULAR, channels, pi)
	reader := testcommon.NewFakeBitReader()

	p, err := NewPassWithReader(reader, framer, 0, 0, fakeHFPassFunc(nil, nil))
	require.NoError(t, err)
	assert.Equal(t, uint32(2), p.minShift)
}

func TestNewPassWithReader_MultipleLastPassEntries_SecondMatch(t *testing.T) {
	// lastPass=[0, 1], downSample=[4, 8] → passIndex=1 matches at index 1
	// minShift = CeilLog1p(8-1) = 3
	pi := &PassesInfo{
		numPasses:  2,
		lastPass:   []uint32{0, 1},
		downSample: []uint32{4, 8},
		shift:      []uint32{0, 0},
	}
	channels := []*ModularChannel{{decoded: true}}
	framer := makeFramerForPass(MODULAR, channels, pi)
	reader := testcommon.NewFakeBitReader()

	p, err := NewPassWithReader(reader, framer, 1, 5, fakeHFPassFunc(nil, nil))
	require.NoError(t, err)
	assert.Equal(t, uint32(3), p.minShift)
	assert.Equal(t, uint32(5), p.maxShift) // prevMinShift
}

func TestNewPassWithReader_ReplacedChannels_AllDecoded(t *testing.T) {
	// All channels already decoded → all replacedChannels are nil
	channels := []*ModularChannel{
		{hshift: 0, vshift: 0, decoded: true},
		{hshift: 1, vshift: 1, decoded: true},
	}
	framer := makeFramerForPass(MODULAR, channels, nil)
	reader := testcommon.NewFakeBitReader()

	p, err := NewPassWithReader(reader, framer, 0, 0, fakeHFPassFunc(nil, nil))
	require.NoError(t, err)
	assert.Len(t, p.replacedChannels, 2)
	assert.Nil(t, p.replacedChannels[0])
	assert.Nil(t, p.replacedChannels[1])
}

func TestNewPassWithReader_ReplacedChannels_UndecidedInRange(t *testing.T) {
	// minShift=maxShift=3 (default passes, passIndex=0)
	// Channel with min(hshift,vshift)=3 → 3 <= 3 && 3 < 3 → FALSE (3 < 3 is false)
	// So no channel will match when minShift==maxShift
	channels := []*ModularChannel{
		{hshift: 3, vshift: 3, decoded: false, size: util.Dimension{Width: 2, Height: 2}},
	}
	framer := makeFramerForPass(MODULAR, channels, nil)
	reader := testcommon.NewFakeBitReader()

	p, err := NewPassWithReader(reader, framer, 0, 0, fakeHFPassFunc(nil, nil))
	require.NoError(t, err)
	assert.Nil(t, p.replacedChannels[0])
}

func TestNewPassWithReader_ReplacedChannels_InShiftRange(t *testing.T) {
	// Use passes with downSample to get minShift < maxShift
	// lastPass=[0], downSample=[4] → minShift=2, maxShift=3
	// Channel with min(hshift,vshift)=2 → 2 <= 2 && 2 < 3 → TRUE → replaced
	pi := &PassesInfo{
		numPasses:  1,
		lastPass:   []uint32{0},
		downSample: []uint32{4},
		shift:      []uint32{0},
	}
	channels := []*ModularChannel{
		{hshift: 2, vshift: 5, decoded: false, size: util.Dimension{Width: 4, Height: 4}},
	}
	framer := makeFramerForPass(MODULAR, channels, pi)
	reader := testcommon.NewFakeBitReader()

	p, err := NewPassWithReader(reader, framer, 0, 0, fakeHFPassFunc(nil, nil))
	require.NoError(t, err)
	require.NotNil(t, p.replacedChannels[0])
	assert.Equal(t, uint32(4), p.replacedChannels[0].size.Width)
	assert.Equal(t, uint32(4), p.replacedChannels[0].size.Height)
}

func TestNewPassWithReader_ReplacedChannels_ShiftBelowRange(t *testing.T) {
	// minShift=2, maxShift=3, channel shift=1 → 2 <= 1 is false → not replaced
	pi := &PassesInfo{
		numPasses:  1,
		lastPass:   []uint32{0},
		downSample: []uint32{4},
		shift:      []uint32{0},
	}
	channels := []*ModularChannel{
		{hshift: 1, vshift: 1, decoded: false},
	}
	framer := makeFramerForPass(MODULAR, channels, pi)
	reader := testcommon.NewFakeBitReader()

	p, err := NewPassWithReader(reader, framer, 0, 0, fakeHFPassFunc(nil, nil))
	require.NoError(t, err)
	assert.Nil(t, p.replacedChannels[0])
}

func TestNewPassWithReader_ReplacedChannels_ShiftAboveRange(t *testing.T) {
	// minShift=2, maxShift=3, channel shift=3 → 3 < 3 is false → not replaced
	pi := &PassesInfo{
		numPasses:  1,
		lastPass:   []uint32{0},
		downSample: []uint32{4},
		shift:      []uint32{0},
	}
	channels := []*ModularChannel{
		{hshift: 3, vshift: 3, decoded: false},
	}
	framer := makeFramerForPass(MODULAR, channels, pi)
	reader := testcommon.NewFakeBitReader()

	p, err := NewPassWithReader(reader, framer, 0, 0, fakeHFPassFunc(nil, nil))
	require.NoError(t, err)
	assert.Nil(t, p.replacedChannels[0])
}

func TestNewPassWithReader_ReplacedChannels_MinOfShiftsUsed(t *testing.T) {
	// min(hshift=5, vshift=2) = 2, which is in range [2,3)
	pi := &PassesInfo{
		numPasses:  1,
		lastPass:   []uint32{0},
		downSample: []uint32{4},
		shift:      []uint32{0},
	}
	channels := []*ModularChannel{
		{hshift: 5, vshift: 2, decoded: false, size: util.Dimension{Width: 3, Height: 3}},
	}
	framer := makeFramerForPass(MODULAR, channels, pi)
	reader := testcommon.NewFakeBitReader()

	p, err := NewPassWithReader(reader, framer, 0, 0, fakeHFPassFunc(nil, nil))
	require.NoError(t, err)
	require.NotNil(t, p.replacedChannels[0])
}

func TestNewPassWithReader_ReplacedChannels_MixOfDecodedAndUndecoded(t *testing.T) {
	// minShift=2, maxShift=3
	// Channel 0: decoded=true → nil
	// Channel 1: decoded=false, shift=2 → replaced
	// Channel 2: decoded=false, shift=1 → nil (below range)
	// Channel 3: decoded=false, shift=2 → replaced
	pi := &PassesInfo{
		numPasses:  1,
		lastPass:   []uint32{0},
		downSample: []uint32{4},
		shift:      []uint32{0},
	}
	channels := []*ModularChannel{
		{hshift: 2, vshift: 2, decoded: true, size: util.Dimension{Width: 1, Height: 1}},
		{hshift: 2, vshift: 2, decoded: false, size: util.Dimension{Width: 2, Height: 2}},
		{hshift: 1, vshift: 1, decoded: false, size: util.Dimension{Width: 3, Height: 3}},
		{hshift: 2, vshift: 3, decoded: false, size: util.Dimension{Width: 4, Height: 4}},
	}
	framer := makeFramerForPass(MODULAR, channels, pi)
	reader := testcommon.NewFakeBitReader()

	p, err := NewPassWithReader(reader, framer, 0, 0, fakeHFPassFunc(nil, nil))
	require.NoError(t, err)
	assert.Len(t, p.replacedChannels, 4)
	assert.Nil(t, p.replacedChannels[0])
	assert.NotNil(t, p.replacedChannels[1])
	assert.Nil(t, p.replacedChannels[2])
	assert.NotNil(t, p.replacedChannels[3])
}

func TestNewPassWithReader_ReplacedChannels_NoChannels(t *testing.T) {
	framer := makeFramerForPass(MODULAR, []*ModularChannel{}, nil)
	reader := testcommon.NewFakeBitReader()

	p, err := NewPassWithReader(reader, framer, 0, 0, fakeHFPassFunc(nil, nil))
	require.NoError(t, err)
	assert.Len(t, p.replacedChannels, 0)
}

func TestNewPassWithReader_VARDCT_CreatesHFPass(t *testing.T) {
	channels := []*ModularChannel{{decoded: true}}
	framer := makeFramerForPass(VARDCT, channels, nil)
	reader := testcommon.NewFakeBitReader()
	expectedHFPass := &HFPass{usedOrders: 42}

	p, err := NewPassWithReader(reader, framer, 0, 0, fakeHFPassFunc(expectedHFPass, nil))
	require.NoError(t, err)
	assert.Equal(t, expectedHFPass, p.hfPass)
}

func TestNewPassWithReader_VARDCT_HFPassError(t *testing.T) {
	channels := []*ModularChannel{{decoded: true}}
	framer := makeFramerForPass(VARDCT, channels, nil)
	reader := testcommon.NewFakeBitReader()

	_, err := NewPassWithReader(reader, framer, 0, 0, fakeHFPassFunc(nil, errors.New("hfpass failed")))
	assert.EqualError(t, err, "hfpass failed")
}

func TestNewPassWithReader_MODULAR_NoHFPass(t *testing.T) {
	channels := []*ModularChannel{{decoded: true}}
	framer := makeFramerForPass(MODULAR, channels, nil)
	reader := testcommon.NewFakeBitReader()

	p, err := NewPassWithReader(reader, framer, 0, 0, fakeHFPassFunc(nil, nil))
	require.NoError(t, err)
	assert.Nil(t, p.hfPass)
}

func TestNewPassWithReader_VARDCT_PassesAlsoProcessesChannels(t *testing.T) {
	// Even with VARDCT, channel replacement logic still runs
	pi := &PassesInfo{
		numPasses:  1,
		lastPass:   []uint32{0},
		downSample: []uint32{4},
		shift:      []uint32{0},
	}
	channels := []*ModularChannel{
		{hshift: 2, vshift: 2, decoded: false, size: util.Dimension{Width: 10, Height: 10}},
	}
	framer := makeFramerForPass(VARDCT, channels, pi)
	reader := testcommon.NewFakeBitReader()
	expectedHFPass := &HFPass{}

	p, err := NewPassWithReader(reader, framer, 0, 0, fakeHFPassFunc(expectedHFPass, nil))
	require.NoError(t, err)
	assert.NotNil(t, p.hfPass)
	require.NotNil(t, p.replacedChannels[0])
	assert.Equal(t, uint32(10), p.replacedChannels[0].size.Width)
}

func TestNewPassWithReader_ReplacedChannelIsCopy(t *testing.T) {
	// Verify the replaced channel is a copy, not a reference to the original
	pi := &PassesInfo{
		numPasses:  1,
		lastPass:   []uint32{0},
		downSample: []uint32{4},
		shift:      []uint32{0},
	}
	original := &ModularChannel{
		hshift: 2, vshift: 2, decoded: false,
		size: util.Dimension{Width: 3, Height: 3},
	}
	framer := makeFramerForPass(MODULAR, []*ModularChannel{original}, pi)
	reader := testcommon.NewFakeBitReader()

	p, err := NewPassWithReader(reader, framer, 0, 0, fakeHFPassFunc(nil, nil))
	require.NoError(t, err)
	require.NotNil(t, p.replacedChannels[0])
	// Should be a different pointer
	assert.NotSame(t, original, p.replacedChannels[0])
	// But same dimensions
	assert.Equal(t, original.size, p.replacedChannels[0].size)
	assert.Equal(t, original.hshift, p.replacedChannels[0].hshift)
	assert.Equal(t, original.vshift, p.replacedChannels[0].vshift)
}

func TestNewPassWithReader_MultiPass_Scenario(t *testing.T) {
	// Simulate 2-pass scenario:
	// Pass 0: maxShift=3, lastPass=[0], downSample=[4] → minShift=2
	// Pass 1: maxShift=prevMinShift=2, no lastPass match → minShift=2
	pi := &PassesInfo{
		numPasses:  2,
		numDS:      1,
		lastPass:   []uint32{0, 1},
		downSample: []uint32{4, 1},
		shift:      []uint32{0, 0},
	}
	channels := []*ModularChannel{
		{hshift: 2, vshift: 2, decoded: false, size: util.Dimension{Width: 5, Height: 5}},
		{hshift: 1, vshift: 1, decoded: false, size: util.Dimension{Width: 5, Height: 5}},
	}

	// Pass 0
	framer := makeFramerForPass(MODULAR, channels, pi)
	reader := testcommon.NewFakeBitReader()
	p0, err := NewPassWithReader(reader, framer, 0, 0, fakeHFPassFunc(nil, nil))
	require.NoError(t, err)
	assert.Equal(t, uint32(3), p0.maxShift)
	assert.Equal(t, uint32(2), p0.minShift)
	// Channel 0 (shift=2): 2 <= 2 && 2 < 3 → replaced
	assert.NotNil(t, p0.replacedChannels[0])
	// Channel 1 (shift=1): 2 <= 1 → false → nil
	assert.Nil(t, p0.replacedChannels[1])

	// Pass 1: prevMinShift = p0.minShift = 2
	p1, err := NewPassWithReader(reader, framer, 1, p0.minShift, fakeHFPassFunc(nil, nil))
	require.NoError(t, err)
	assert.Equal(t, uint32(2), p1.maxShift)
	// lastPass=[0,1], passIndex=1 matches at index 1, downSample[1]=1
	// minShift = CeilLog1p(1-1) = CeilLog1p(0) = 0
	assert.Equal(t, uint32(0), p1.minShift)
	// Channel 0 (shift=2): 0 <= 2 && 2 < 2 → false (2 < 2)
	assert.Nil(t, p1.replacedChannels[0])
	// Channel 1 (shift=1): 0 <= 1 && 1 < 2 → true
	assert.NotNil(t, p1.replacedChannels[1])
}

func TestNewPassWithReader_HFPassFuncReceivesCorrectPassIndex(t *testing.T) {
	channels := []*ModularChannel{{decoded: true}}
	framer := makeFramerForPass(VARDCT, channels, nil)
	reader := testcommon.NewFakeBitReader()

	var capturedPassIndex uint32
	customHFPassFunc := func(reader jxlio.BitReader, frame Framer, passIndex uint32,
		readClusterMapFunc entropy.ReadClusterMapFunc,
		newEntropyStreamWithReader entropy.EntropyStreamWithReaderAndNumDistsFunc,
		readPermutation ReadPermutationFunc) (*HFPass, error) {
		capturedPassIndex = passIndex
		return &HFPass{}, nil
	}

	_, err := NewPassWithReader(reader, framer, 5, 0, customHFPassFunc)
	require.NoError(t, err)
	assert.Equal(t, uint32(5), capturedPassIndex)
}
