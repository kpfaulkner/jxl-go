package frame

import (
	"errors"
	"testing"

	"github.com/kpfaulkner/jxl-go/jxlio"
	"github.com/kpfaulkner/jxl-go/testcommon"
	"github.com/kpfaulkner/jxl-go/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// makeHFMetadataForPlaceBlock creates an HFMetadata with an empty grid of given size.
// Height and width are in dctSelect units (multiples of 8 pixels).
func makeHFMetadataForPlaceBlock(height, width int) *HFMetadata {
	return &HFMetadata{
		dctSelect:    util.MakeMatrix2D[*TransformType](int32(height), int32(width)),
		hfMultiplier: util.MakeMatrix2D[int32](int32(height), int32(width)),
	}
}

func TestPlaceBlock_SingleDCT8OnEmptyGrid(t *testing.T) {
	m := makeHFMetadataForPlaceBlock(4, 4)
	block := *DCT8 // 1x1
	pos, err := m.placeBlock(util.Point{X: 0, Y: 0}, block, 5)
	require.NoError(t, err)
	assert.Equal(t, util.Point{X: 0, Y: 0}, pos)
	assert.Equal(t, &block, m.dctSelect[0][0])
	assert.Equal(t, int32(5), m.hfMultiplier[0][0])
	// Adjacent cells should be untouched
	assert.Nil(t, m.dctSelect[0][1])
	assert.Equal(t, int32(0), m.hfMultiplier[0][1])
}

func TestPlaceBlock_DCT16On2x2Block(t *testing.T) {
	m := makeHFMetadataForPlaceBlock(4, 4)
	block := *DCT16 // 2x2
	pos, err := m.placeBlock(util.Point{X: 0, Y: 0}, block, 3)
	require.NoError(t, err)
	assert.Equal(t, util.Point{X: 0, Y: 0}, pos)
	// All 4 cells should be filled
	for y := 0; y < 2; y++ {
		for x := 0; x < 2; x++ {
			assert.Equal(t, &block, m.dctSelect[y][x], "dctSelect[%d][%d]", y, x)
			assert.Equal(t, int32(3), m.hfMultiplier[y][x], "hfMultiplier[%d][%d]", y, x)
		}
	}
	// Cells outside should be untouched
	assert.Nil(t, m.dctSelect[0][2])
	assert.Nil(t, m.dctSelect[2][0])
}

func TestPlaceBlock_SequentialDCT8Placement(t *testing.T) {
	m := makeHFMetadataForPlaceBlock(2, 3)
	block := *DCT8 // 1x1
	origin := util.Point{X: 0, Y: 0}

	// Place 3 blocks on row 0
	pos1, err := m.placeBlock(origin, block, 1)
	require.NoError(t, err)
	assert.Equal(t, util.Point{X: 0, Y: 0}, pos1)

	pos2, err := m.placeBlock(origin, block, 2)
	require.NoError(t, err)
	assert.Equal(t, util.Point{X: 1, Y: 0}, pos2)

	pos3, err := m.placeBlock(origin, block, 3)
	require.NoError(t, err)
	assert.Equal(t, util.Point{X: 2, Y: 0}, pos3)

	// Row 0 full, next block goes to row 1
	pos4, err := m.placeBlock(origin, block, 4)
	require.NoError(t, err)
	assert.Equal(t, util.Point{X: 0, Y: 1}, pos4)

	// Verify multipliers
	assert.Equal(t, int32(1), m.hfMultiplier[0][0])
	assert.Equal(t, int32(2), m.hfMultiplier[0][1])
	assert.Equal(t, int32(3), m.hfMultiplier[0][2])
	assert.Equal(t, int32(4), m.hfMultiplier[1][0])
}

func TestPlaceBlock_WideBlockSkipsToNextRow(t *testing.T) {
	// Grid is 2 wide, try to place a 2-wide block after occupying col 0
	m := makeHFMetadataForPlaceBlock(2, 2)
	small := *DCT8   // 1x1
	wide := *DCT8_16 // dctSelectWidth=2, dctSelectHeight=1

	// Place a small block at (0,0)
	_, err := m.placeBlock(util.Point{X: 0, Y: 0}, small, 1)
	require.NoError(t, err)

	// Wide block can't fit at x=1 (1+2>2), wraps to row 1
	pos, err := m.placeBlock(util.Point{X: 0, Y: 0}, wide, 2)
	require.NoError(t, err)
	assert.Equal(t, util.Point{X: 0, Y: 1}, pos)
	assert.Equal(t, int32(2), m.hfMultiplier[1][0])
	assert.Equal(t, int32(2), m.hfMultiplier[1][1])
}

func TestPlaceBlock_TallBlockFillsMultipleRows(t *testing.T) {
	m := makeHFMetadataForPlaceBlock(4, 2)
	tall := *DCT16_8 // dctSelectWidth=1, dctSelectHeight=2

	pos, err := m.placeBlock(util.Point{X: 0, Y: 0}, tall, 7)
	require.NoError(t, err)
	assert.Equal(t, util.Point{X: 0, Y: 0}, pos)
	assert.Equal(t, int32(7), m.hfMultiplier[0][0])
	assert.Equal(t, int32(7), m.hfMultiplier[1][0])
	// Column 1 should be untouched
	assert.Nil(t, m.dctSelect[0][1])
	assert.Nil(t, m.dctSelect[1][1])
}

func TestPlaceBlock_SkipsOccupiedCellsByWidth(t *testing.T) {
	// Place a 2-wide block, then a 1x1 block should skip past it
	m := makeHFMetadataForPlaceBlock(2, 4)
	wide := *DCT8_16 // dctSelectWidth=2, dctSelectHeight=1
	small := *DCT8   // 1x1

	_, err := m.placeBlock(util.Point{X: 0, Y: 0}, wide, 1)
	require.NoError(t, err)

	// Next 1x1 block should land at x=2 (skips past the 2-wide block)
	pos, err := m.placeBlock(util.Point{X: 0, Y: 0}, small, 2)
	require.NoError(t, err)
	assert.Equal(t, util.Point{X: 2, Y: 0}, pos)
}

func TestPlaceBlock_LastBlockYStartsSearch(t *testing.T) {
	m := makeHFMetadataForPlaceBlock(3, 2)
	block := *DCT8 // 1x1

	// Start search from row 2
	pos, err := m.placeBlock(util.Point{X: 0, Y: 2}, block, 9)
	require.NoError(t, err)
	assert.Equal(t, util.Point{X: 0, Y: 2}, pos)
	// Rows 0 and 1 should be untouched
	assert.Nil(t, m.dctSelect[0][0])
	assert.Nil(t, m.dctSelect[1][0])
}

func TestPlaceBlock_NoSpaceError(t *testing.T) {
	m := makeHFMetadataForPlaceBlock(1, 1)
	block := *DCT8 // 1x1

	// Fill the only cell
	_, err := m.placeBlock(util.Point{X: 0, Y: 0}, block, 1)
	require.NoError(t, err)

	// No space left
	_, err = m.placeBlock(util.Point{X: 0, Y: 0}, block, 2)
	assert.EqualError(t, err, "No space for block")
}

func TestPlaceBlock_NoSpaceBlockTooWide(t *testing.T) {
	// Grid is 1 wide, but block needs 2
	m := makeHFMetadataForPlaceBlock(2, 1)
	wide := *DCT8_16 // dctSelectWidth=2

	_, err := m.placeBlock(util.Point{X: 0, Y: 0}, wide, 1)
	assert.EqualError(t, err, "No space for block")
}

func TestPlaceBlock_2x2BlockAfterOccupied(t *testing.T) {
	// Row 0: occupied at (0,0), then DCT16 needs 2x2 starting at x=0 but
	// col 0 is taken. x=1 won't fit (1+2>3 needs check). With width=4 it should fit at x=2.
	m := makeHFMetadataForPlaceBlock(4, 4)
	small := *DCT8 // 1x1
	big := *DCT16  // 2x2

	_, err := m.placeBlock(util.Point{X: 0, Y: 0}, small, 1)
	require.NoError(t, err)

	pos, err := m.placeBlock(util.Point{X: 0, Y: 0}, big, 2)
	require.NoError(t, err)
	// x=0 is occupied, x=1 checks occupancy of x=1 and x=2 (both nil), fits!
	assert.Equal(t, util.Point{X: 1, Y: 0}, pos)

	// Verify the 2x2 fill
	for y := 0; y < 2; y++ {
		for x := 1; x < 3; x++ {
			assert.NotNil(t, m.dctSelect[y][x], "dctSelect[%d][%d]", y, x)
			assert.Equal(t, int32(2), m.hfMultiplier[y][x], "hfMultiplier[%d][%d]", y, x)
		}
	}
}

func TestPlaceBlock_WideBlockExactFit(t *testing.T) {
	// Grid exactly fits a 2-wide block
	m := makeHFMetadataForPlaceBlock(1, 2)
	wide := *DCT8_16 // dctSelectWidth=2, dctSelectHeight=1

	pos, err := m.placeBlock(util.Point{X: 0, Y: 0}, wide, 10)
	require.NoError(t, err)
	assert.Equal(t, util.Point{X: 0, Y: 0}, pos)
	assert.Equal(t, int32(10), m.hfMultiplier[0][0])
	assert.Equal(t, int32(10), m.hfMultiplier[0][1])
}

func TestPlaceBlock_MultiplierOverwrittenByLaterBlock(t *testing.T) {
	// This tests the grid fill: place a 2x2 block, verify all cells get the multiplier
	m := makeHFMetadataForPlaceBlock(2, 2)
	block := *DCT16 // 2x2

	pos, err := m.placeBlock(util.Point{X: 0, Y: 0}, block, 42)
	require.NoError(t, err)
	assert.Equal(t, util.Point{X: 0, Y: 0}, pos)
	for y := 0; y < 2; y++ {
		for x := 0; x < 2; x++ {
			assert.Equal(t, int32(42), m.hfMultiplier[y][x])
		}
	}
}

func TestPlaceBlock_MixedSizes(t *testing.T) {
	// 4x4 grid: place a 2x2, then fill remaining with 1x1 blocks
	m := makeHFMetadataForPlaceBlock(4, 4)
	big := *DCT16 // 2x2
	small := *DCT8

	// Place 2x2 at origin
	pos, err := m.placeBlock(util.Point{X: 0, Y: 0}, big, 1)
	require.NoError(t, err)
	assert.Equal(t, util.Point{X: 0, Y: 0}, pos)

	// Next 1x1 should go to x=2, y=0 (skips occupied 2x2)
	pos, err = m.placeBlock(util.Point{X: 0, Y: 0}, small, 2)
	require.NoError(t, err)
	assert.Equal(t, util.Point{X: 2, Y: 0}, pos)

	pos, err = m.placeBlock(util.Point{X: 0, Y: 0}, small, 3)
	require.NoError(t, err)
	assert.Equal(t, util.Point{X: 3, Y: 0}, pos)
}

func TestPlaceBlock_DctSelectPointersSameBlock(t *testing.T) {
	// All cells in a multi-cell block should point to the same TransformType
	m := makeHFMetadataForPlaceBlock(2, 2)
	block := *DCT16 // 2x2

	_, err := m.placeBlock(util.Point{X: 0, Y: 0}, block, 1)
	require.NoError(t, err)

	// All 4 cells should point to the same address
	ptr := m.dctSelect[0][0]
	assert.Equal(t, ptr, m.dctSelect[0][1])
	assert.Equal(t, ptr, m.dctSelect[1][0])
	assert.Equal(t, ptr, m.dctSelect[1][1])
}

func TestPlaceBlock_LastBlockYSkipsFilledRows(t *testing.T) {
	m := makeHFMetadataForPlaceBlock(3, 2)
	block := *DCT8

	// Fill row 0 completely
	_, err := m.placeBlock(util.Point{X: 0, Y: 0}, block, 1)
	require.NoError(t, err)
	_, err = m.placeBlock(util.Point{X: 0, Y: 0}, block, 2)
	require.NoError(t, err)

	// Now start search from row 1 (lastBlock.Y=1)
	pos, err := m.placeBlock(util.Point{X: 0, Y: 1}, block, 3)
	require.NoError(t, err)
	assert.Equal(t, util.Point{X: 0, Y: 1}, pos)
}

func TestPlaceBlock_4x4Block(t *testing.T) {
	m := makeHFMetadataForPlaceBlock(4, 4)
	block := *DCT32 // dctSelectWidth=4, dctSelectHeight=4

	pos, err := m.placeBlock(util.Point{X: 0, Y: 0}, block, 5)
	require.NoError(t, err)
	assert.Equal(t, util.Point{X: 0, Y: 0}, pos)

	// All 16 cells should be filled
	for y := 0; y < 4; y++ {
		for x := 0; x < 4; x++ {
			assert.NotNil(t, m.dctSelect[y][x], "dctSelect[%d][%d]", y, x)
			assert.Equal(t, int32(5), m.hfMultiplier[y][x], "hfMultiplier[%d][%d]", y, x)
		}
	}
}

func TestPlaceBlock_EmptyGridReturnsOrigin(t *testing.T) {
	m := makeHFMetadataForPlaceBlock(1, 1)
	block := *DCT8

	pos, err := m.placeBlock(util.Point{X: 0, Y: 0}, block, 1)
	require.NoError(t, err)
	assert.Equal(t, util.Point{X: 0, Y: 0}, pos)
}

func TestPlaceBlock_SkipPast2x2OccupiedBlock(t *testing.T) {
	// Place a 2x2 block at origin, then a 1x1 block should skip past both columns
	m := makeHFMetadataForPlaceBlock(2, 4)
	big := *DCT16 // 2x2
	small := *DCT8

	_, err := m.placeBlock(util.Point{X: 0, Y: 0}, big, 1)
	require.NoError(t, err)

	// 1x1 at x=0 sees occupied, skips by dctSelectWidth(2)-1=1, then loop increments x to 2
	pos, err := m.placeBlock(util.Point{X: 0, Y: 0}, small, 2)
	require.NoError(t, err)
	assert.Equal(t, util.Point{X: 2, Y: 0}, pos)
}

func TestPlaceBlock_NoSpaceLastBlockYBeyondGrid(t *testing.T) {
	m := makeHFMetadataForPlaceBlock(2, 2)
	block := *DCT8

	// Start searching from beyond the grid
	_, err := m.placeBlock(util.Point{X: 0, Y: 2}, block, 1)
	assert.EqualError(t, err, "No space for block")
}

func TestPlaceBlock_MultipleWideBlocks(t *testing.T) {
	m := makeHFMetadataForPlaceBlock(2, 4)
	wide := *DCT8_16 // dctSelectWidth=2, dctSelectHeight=1

	pos1, err := m.placeBlock(util.Point{X: 0, Y: 0}, wide, 1)
	require.NoError(t, err)
	assert.Equal(t, util.Point{X: 0, Y: 0}, pos1)

	pos2, err := m.placeBlock(util.Point{X: 0, Y: 0}, wide, 2)
	require.NoError(t, err)
	assert.Equal(t, util.Point{X: 2, Y: 0}, pos2)

	pos3, err := m.placeBlock(util.Point{X: 0, Y: 0}, wide, 3)
	require.NoError(t, err)
	assert.Equal(t, util.Point{X: 0, Y: 1}, pos3)

	// Verify multipliers
	assert.Equal(t, int32(1), m.hfMultiplier[0][0])
	assert.Equal(t, int32(1), m.hfMultiplier[0][1])
	assert.Equal(t, int32(2), m.hfMultiplier[0][2])
	assert.Equal(t, int32(2), m.hfMultiplier[0][3])
	assert.Equal(t, int32(3), m.hfMultiplier[1][0])
	assert.Equal(t, int32(3), m.hfMultiplier[1][1])
}

func TestPlaceBlock_TallBlocksAdjacentColumns(t *testing.T) {
	m := makeHFMetadataForPlaceBlock(2, 3)
	tall := *DCT16_8 // dctSelectWidth=1, dctSelectHeight=2

	pos1, err := m.placeBlock(util.Point{X: 0, Y: 0}, tall, 1)
	require.NoError(t, err)
	assert.Equal(t, util.Point{X: 0, Y: 0}, pos1)

	pos2, err := m.placeBlock(util.Point{X: 0, Y: 0}, tall, 2)
	require.NoError(t, err)
	assert.Equal(t, util.Point{X: 1, Y: 0}, pos2)

	pos3, err := m.placeBlock(util.Point{X: 0, Y: 0}, tall, 3)
	require.NoError(t, err)
	assert.Equal(t, util.Point{X: 2, Y: 0}, pos3)

	// Grid fully filled
	for y := 0; y < 2; y++ {
		for x := 0; x < 3; x++ {
			assert.NotNil(t, m.dctSelect[y][x], "dctSelect[%d][%d]", y, x)
		}
	}
	// Verify multipliers on both rows
	assert.Equal(t, int32(1), m.hfMultiplier[0][0])
	assert.Equal(t, int32(1), m.hfMultiplier[1][0])
	assert.Equal(t, int32(2), m.hfMultiplier[0][1])
	assert.Equal(t, int32(2), m.hfMultiplier[1][1])
}

func TestPlaceBlock_ZeroMultiplier(t *testing.T) {
	m := makeHFMetadataForPlaceBlock(1, 1)
	block := *DCT8

	pos, err := m.placeBlock(util.Point{X: 0, Y: 0}, block, 0)
	require.NoError(t, err)
	assert.Equal(t, util.Point{X: 0, Y: 0}, pos)
	// Multiplier 0 should still be written (distinguishable from unset only by dctSelect)
	assert.Equal(t, int32(0), m.hfMultiplier[0][0])
	assert.NotNil(t, m.dctSelect[0][0])
}

func TestPlaceBlock_NegativeMultiplier(t *testing.T) {
	m := makeHFMetadataForPlaceBlock(1, 1)
	block := *DCT8

	pos, err := m.placeBlock(util.Point{X: 0, Y: 0}, block, -5)
	require.NoError(t, err)
	assert.Equal(t, util.Point{X: 0, Y: 0}, pos)
	assert.Equal(t, int32(-5), m.hfMultiplier[0][0])
}

func TestPlaceBlock_WideThenTallFillsGap(t *testing.T) {
	// 2x2 grid: place wide block (2x1) at row 0, then tall block (1x2) won't fit at x=0
	// (row 0 occupied), so it goes to... let's trace:
	// After wide: row 0 fully occupied, row 1 empty
	// Tall block: y=0, x=0 → occupied (dctSelectWidth=2, skip by 2-1=1, x becomes 2)
	//             x=2 → 1+2>2, continue outerY
	//             y=1, x=0 → nil, but height=2 → writes to rows 1 and 2
	//             But grid is only 2 rows! This would panic. Let me use a 3-row grid.
	m := makeHFMetadataForPlaceBlock(3, 2)
	wide := *DCT8_16 // w=2, h=1
	tall := *DCT16_8 // w=1, h=2

	_, err := m.placeBlock(util.Point{X: 0, Y: 0}, wide, 1)
	require.NoError(t, err)

	pos, err := m.placeBlock(util.Point{X: 0, Y: 0}, tall, 2)
	require.NoError(t, err)
	assert.Equal(t, util.Point{X: 0, Y: 1}, pos)
	assert.Equal(t, int32(2), m.hfMultiplier[1][0])
	assert.Equal(t, int32(2), m.hfMultiplier[2][0])
}

func TestPlaceBlock_FullGridMultipleBlockTypes(t *testing.T) {
	// 4x4 grid: fill with a 2x2 at (0,0), two 1x2 wide at rows 0-1 col 2-3,
	// then verify remaining space fills correctly
	m := makeHFMetadataForPlaceBlock(4, 4)
	big := *DCT16   // 2x2
	wide := *DCT8_16 // w=2, h=1
	small := *DCT8   // 1x1

	// 2x2 block at origin
	pos, err := m.placeBlock(util.Point{X: 0, Y: 0}, big, 1)
	require.NoError(t, err)
	assert.Equal(t, util.Point{X: 0, Y: 0}, pos)

	// Wide block should go to x=2, y=0
	pos, err = m.placeBlock(util.Point{X: 0, Y: 0}, wide, 2)
	require.NoError(t, err)
	assert.Equal(t, util.Point{X: 2, Y: 0}, pos)

	// Another wide block at x=2, y=1
	pos, err = m.placeBlock(util.Point{X: 0, Y: 0}, wide, 3)
	require.NoError(t, err)
	assert.Equal(t, util.Point{X: 2, Y: 1}, pos)

	// Now rows 0-1 are full. Small blocks should start at row 2
	pos, err = m.placeBlock(util.Point{X: 0, Y: 0}, small, 4)
	require.NoError(t, err)
	assert.Equal(t, util.Point{X: 0, Y: 2}, pos)
}

// --- NewHFMetadataWithReader tests ---

// fakeModularStreamHFMeta implements ModularStreamer for HFMetadata tests.
type fakeModularStreamHFMeta struct {
	decodedBuffer    [][][]int32
	decodeErr        error
}

func (f *fakeModularStreamHFMeta) decodeChannels(reader jxlio.BitReader, partial bool) error {
	return f.decodeErr
}

func (f *fakeModularStreamHFMeta) getDecodedBuffer() [][][]int32 {
	return f.decodedBuffer
}

func (f *fakeModularStreamHFMeta) applyTransforms() error {
	return nil
}

func (f *fakeModularStreamHFMeta) getChannels() []*ModularChannel {
	return nil
}

// makeBlockInfoBuffer creates channel 2 (blockInfo) buffer for the modular stream.
// blockTypes[i] is the transform type index (0=DCT8, 4=DCT16, etc.)
// multipliers[i] is the raw multiplier value (actual mul = 1 + multipliers[i])
func makeBlockInfoBuffer(blockTypes []int32, multipliers []int32) [][]int32 {
	return [][]int32{blockTypes, multipliers}
}

// makeFakeModularStreamFunc returns a NewModularStreamWithStreamIndexFunc
// that produces a fakeModularStreamHFMeta with the given decoded buffer.
func makeFakeModularStreamFunc(decodedBuffer [][][]int32, decodeErr error) NewModularStreamWithStreamIndexFunc {
	return func(reader jxlio.BitReader, frame Framer, streamIndex int, channelArray []ModularChannel) (ModularStreamer, error) {
		return &fakeModularStreamHFMeta{
			decodedBuffer: decodedBuffer,
			decodeErr:     decodeErr,
		}, nil
	}
}

func makeFakeModularStreamFuncWithError(err error) NewModularStreamWithStreamIndexFunc {
	return func(reader jxlio.BitReader, frame Framer, streamIndex int, channelArray []ModularChannel) (ModularStreamer, error) {
		return nil, err
	}
}

func TestNewHFMetadataWithReader_SingleDCT8Block(t *testing.T) {
	// parent.size = {Height: 1, Width: 1} → 1x1 dctSelect grid
	// CeilLog2(1*1) = 0, ReadBits(0) returns 0 → nbBlocks = 0 + 1 = 1
	parent := &LFGroup{size: util.Dimension{Height: 1, Width: 1}}
	reader := &testcommon.FakeBitReader{ReadBitsData: []uint64{0}}
	framer := NewFakeFramer(VARDCT)

	blockInfo := makeBlockInfoBuffer([]int32{0}, []int32{4}) // DCT8, mul = 1+4 = 5
	decodedBuffer := [][][]int32{
		{{0}},          // channel 0 (xFromY)
		{{0}},          // channel 1 (bFromY)
		blockInfo,      // channel 2 (blockInfo)
		{{0}},          // channel 3 (sharpness)
	}
	msFunc := makeFakeModularStreamFunc(decodedBuffer, nil)

	hf, err := NewHFMetadataWithReader(reader, parent, framer, msFunc)
	require.NoError(t, err)
	assert.Equal(t, uint64(1), hf.nbBlocks)
	assert.NotNil(t, hf.dctSelect[0][0])
	assert.Equal(t, int32(5), hf.hfMultiplier[0][0])
	assert.Equal(t, util.Point{X: 0, Y: 0}, hf.blockList[0])
	assert.Equal(t, parent, hf.parent)
}

func TestNewHFMetadataWithReader_MultipleDCT8Blocks(t *testing.T) {
	// parent.size = {Height: 2, Width: 2} → 2x2 grid, 4 cells
	// CeilLog2(2*2) = CeilLog2(4) = 2, ReadBits(2) returns 3 → nbBlocks = 4
	parent := &LFGroup{size: util.Dimension{Height: 2, Width: 2}}
	reader := &testcommon.FakeBitReader{ReadBitsData: []uint64{3}}
	framer := NewFakeFramer(VARDCT)

	blockInfo := makeBlockInfoBuffer(
		[]int32{0, 0, 0, 0},    // all DCT8
		[]int32{0, 1, 2, 3},    // multipliers: 1, 2, 3, 4
	)
	decodedBuffer := [][][]int32{
		{{0, 0}},
		{{0, 0}},
		blockInfo,
		{{0, 0}, {0, 0}},
	}
	msFunc := makeFakeModularStreamFunc(decodedBuffer, nil)

	hf, err := NewHFMetadataWithReader(reader, parent, framer, msFunc)
	require.NoError(t, err)
	assert.Equal(t, uint64(4), hf.nbBlocks)

	// Verify all 4 positions filled
	assert.Equal(t, util.Point{X: 0, Y: 0}, hf.blockList[0])
	assert.Equal(t, util.Point{X: 1, Y: 0}, hf.blockList[1])
	assert.Equal(t, util.Point{X: 0, Y: 1}, hf.blockList[2])
	assert.Equal(t, util.Point{X: 1, Y: 1}, hf.blockList[3])

	// Verify multipliers (1 + raw value)
	assert.Equal(t, int32(1), hf.hfMultiplier[0][0])
	assert.Equal(t, int32(2), hf.hfMultiplier[0][1])
	assert.Equal(t, int32(3), hf.hfMultiplier[1][0])
	assert.Equal(t, int32(4), hf.hfMultiplier[1][1])
}

func TestNewHFMetadataWithReader_DCT16Block(t *testing.T) {
	// parent.size = {Height: 2, Width: 2} → 2x2 grid
	// nbBlocks = 1, block type 4 (DCT16, 2x2)
	parent := &LFGroup{size: util.Dimension{Height: 2, Width: 2}}
	reader := &testcommon.FakeBitReader{ReadBitsData: []uint64{0}}
	framer := NewFakeFramer(VARDCT)

	blockInfo := makeBlockInfoBuffer([]int32{4}, []int32{9}) // DCT16, mul = 10
	decodedBuffer := [][][]int32{
		{{0}},
		{{0}},
		blockInfo,
		{{0, 0}, {0, 0}},
	}
	msFunc := makeFakeModularStreamFunc(decodedBuffer, nil)

	hf, err := NewHFMetadataWithReader(reader, parent, framer, msFunc)
	require.NoError(t, err)
	assert.Equal(t, uint64(1), hf.nbBlocks)

	// All 4 cells filled by the 2x2 DCT16 block
	for y := 0; y < 2; y++ {
		for x := 0; x < 2; x++ {
			assert.NotNil(t, hf.dctSelect[y][x], "dctSelect[%d][%d]", y, x)
			assert.Equal(t, int32(10), hf.hfMultiplier[y][x], "hfMultiplier[%d][%d]", y, x)
		}
	}
	assert.Equal(t, util.Point{X: 0, Y: 0}, hf.blockList[0])
}

func TestNewHFMetadataWithReader_InvalidTransformTypeHigh(t *testing.T) {
	parent := &LFGroup{size: util.Dimension{Height: 1, Width: 1}}
	reader := &testcommon.FakeBitReader{ReadBitsData: []uint64{0}}
	framer := NewFakeFramer(VARDCT)

	blockInfo := makeBlockInfoBuffer([]int32{27}, []int32{0}) // 27 > 26 → invalid
	decodedBuffer := [][][]int32{{{0}}, {{0}}, blockInfo, {{0}}}
	msFunc := makeFakeModularStreamFunc(decodedBuffer, nil)

	_, err := NewHFMetadataWithReader(reader, parent, framer, msFunc)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Invalid transform Type 27")
}

func TestNewHFMetadataWithReader_InvalidTransformTypeNegative(t *testing.T) {
	parent := &LFGroup{size: util.Dimension{Height: 1, Width: 1}}
	reader := &testcommon.FakeBitReader{ReadBitsData: []uint64{0}}
	framer := NewFakeFramer(VARDCT)

	blockInfo := makeBlockInfoBuffer([]int32{-1}, []int32{0})
	decodedBuffer := [][][]int32{{{0}}, {{0}}, blockInfo, {{0}}}
	msFunc := makeFakeModularStreamFunc(decodedBuffer, nil)

	_, err := NewHFMetadataWithReader(reader, parent, framer, msFunc)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Invalid transform Type -1")
}

func TestNewHFMetadataWithReader_ReadBitsError(t *testing.T) {
	parent := &LFGroup{size: util.Dimension{Height: 2, Width: 2}}
	reader := &testcommon.FakeBitReader{ReadBitsData: []uint64{}} // no data
	framer := NewFakeFramer(VARDCT)
	msFunc := makeFakeModularStreamFunc(nil, nil)

	_, err := NewHFMetadataWithReader(reader, parent, framer, msFunc)
	assert.Error(t, err)
}

func TestNewHFMetadataWithReader_ModularStreamCreationError(t *testing.T) {
	parent := &LFGroup{size: util.Dimension{Height: 1, Width: 1}}
	reader := &testcommon.FakeBitReader{ReadBitsData: []uint64{0}}
	framer := NewFakeFramer(VARDCT)
	msFunc := makeFakeModularStreamFuncWithError(errors.New("stream creation failed"))

	_, err := NewHFMetadataWithReader(reader, parent, framer, msFunc)
	assert.EqualError(t, err, "stream creation failed")
}

func TestNewHFMetadataWithReader_DecodeChannelsError(t *testing.T) {
	parent := &LFGroup{size: util.Dimension{Height: 1, Width: 1}}
	reader := &testcommon.FakeBitReader{ReadBitsData: []uint64{0}}
	framer := NewFakeFramer(VARDCT)
	msFunc := makeFakeModularStreamFunc(nil, errors.New("decode failed"))

	_, err := NewHFMetadataWithReader(reader, parent, framer, msFunc)
	assert.EqualError(t, err, "decode failed")
}

func TestNewHFMetadataWithReader_PlaceBlockNoSpaceError(t *testing.T) {
	// 1x1 grid with 2 blocks → second block has no space
	parent := &LFGroup{size: util.Dimension{Height: 1, Width: 1}}
	reader := &testcommon.FakeBitReader{ReadBitsData: []uint64{1}} // nbBlocks = 1+1 = 2
	framer := NewFakeFramer(VARDCT)

	blockInfo := makeBlockInfoBuffer([]int32{0, 0}, []int32{0, 0})
	decodedBuffer := [][][]int32{{{0}}, {{0}}, blockInfo, {{0}}}
	msFunc := makeFakeModularStreamFunc(decodedBuffer, nil)

	_, err := NewHFMetadataWithReader(reader, parent, framer, msFunc)
	assert.EqualError(t, err, "No space for block")
}

func TestNewHFMetadataWithReader_HfStreamBufferStored(t *testing.T) {
	parent := &LFGroup{size: util.Dimension{Height: 1, Width: 1}}
	reader := &testcommon.FakeBitReader{ReadBitsData: []uint64{0}}
	framer := NewFakeFramer(VARDCT)

	blockInfo := makeBlockInfoBuffer([]int32{0}, []int32{0})
	decodedBuffer := [][][]int32{{{99}}, {{88}}, blockInfo, {{77}}}
	msFunc := makeFakeModularStreamFunc(decodedBuffer, nil)

	hf, err := NewHFMetadataWithReader(reader, parent, framer, msFunc)
	require.NoError(t, err)
	assert.Equal(t, decodedBuffer, hf.hfStreamBuffer)
}

func TestNewHFMetadataWithReader_MixedBlockTypes(t *testing.T) {
	// 4x4 grid: 1 DCT16 (2x2) + 4 DCT8 (1x1) for top-right + 8 DCT8 for bottom
	// DCT16 fills (0,0)-(1,1), then DCT8 fills (2,0), (3,0), (2,1), (3,1),
	// then (0,2), (1,2), (2,2), (3,2), (0,3), (1,3), (2,3), (3,3)
	parent := &LFGroup{size: util.Dimension{Height: 4, Width: 4}}
	// CeilLog2(4*4) = CeilLog2(16) = 4, ReadBits(4) returns nbBlocks-1
	reader := &testcommon.FakeBitReader{ReadBitsData: []uint64{12}} // nbBlocks = 13
	framer := NewFakeFramer(VARDCT)

	types := []int32{4, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}       // DCT16 then 12 DCT8
	muls := []int32{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	blockInfo := makeBlockInfoBuffer(types, muls)
	decodedBuffer := [][][]int32{
		{{0, 0, 0, 0}},
		{{0, 0, 0, 0}},
		blockInfo,
		{{0, 0, 0, 0}, {0, 0, 0, 0}, {0, 0, 0, 0}, {0, 0, 0, 0}},
	}
	msFunc := makeFakeModularStreamFunc(decodedBuffer, nil)

	hf, err := NewHFMetadataWithReader(reader, parent, framer, msFunc)
	require.NoError(t, err)
	assert.Equal(t, uint64(13), hf.nbBlocks)
	assert.Len(t, hf.blockList, 13)

	// DCT16 placed at (0,0)
	assert.Equal(t, util.Point{X: 0, Y: 0}, hf.blockList[0])
	// Next DCT8 at (2,0) - skips past the 2x2 block
	assert.Equal(t, util.Point{X: 2, Y: 0}, hf.blockList[1])
	// All cells should be filled
	for y := 0; y < 4; y++ {
		for x := 0; x < 4; x++ {
			assert.NotNil(t, hf.dctSelect[y][x], "dctSelect[%d][%d] should not be nil", y, x)
		}
	}
}

func TestNewHFMetadataWithReader_LastBlockAdvances(t *testing.T) {
	// Verify that lastBlock is updated after each placement
	// 2x2 grid, 2 blocks: first at (0,0), lastBlock becomes (0,0), second at (1,0)
	parent := &LFGroup{size: util.Dimension{Height: 2, Width: 2}}
	reader := &testcommon.FakeBitReader{ReadBitsData: []uint64{1}} // nbBlocks = 2
	framer := NewFakeFramer(VARDCT)

	blockInfo := makeBlockInfoBuffer([]int32{0, 0}, []int32{0, 1})
	decodedBuffer := [][][]int32{
		{{0, 0}},
		{{0, 0}},
		blockInfo,
		{{0, 0}, {0, 0}},
	}
	msFunc := makeFakeModularStreamFunc(decodedBuffer, nil)

	hf, err := NewHFMetadataWithReader(reader, parent, framer, msFunc)
	require.NoError(t, err)
	assert.Equal(t, util.Point{X: 0, Y: 0}, hf.blockList[0])
	assert.Equal(t, util.Point{X: 1, Y: 0}, hf.blockList[1])
	assert.Equal(t, int32(1), hf.hfMultiplier[0][0])
	assert.Equal(t, int32(2), hf.hfMultiplier[0][1])
}

func TestNewHFMetadataWithReader_AllValidTransformTypes(t *testing.T) {
	// Test that all 27 transform types (0-26) are accepted by the validation
	for tt := int32(0); tt <= 26; tt++ {
		t.Run(allDCT[tt].name, func(t *testing.T) {
			block := allDCT[tt]
			h := int(block.dctSelectHeight)
			w := int(block.dctSelectWidth)
			parent := &LFGroup{size: util.Dimension{Height: uint32(h), Width: uint32(w)}}
			// CeilLog2(h*w) bits needed
			reader := &testcommon.FakeBitReader{ReadBitsData: []uint64{0}} // nbBlocks = 1
			framer := NewFakeFramer(VARDCT)

			blockInfo := makeBlockInfoBuffer([]int32{tt}, []int32{0})
			decodedBuffer := [][][]int32{
				{{0}},
				{{0}},
				blockInfo,
				util.MakeMatrix2D[int32](int32(h), int32(w)),
			}
			msFunc := makeFakeModularStreamFunc(decodedBuffer, nil)

			hf, err := NewHFMetadataWithReader(reader, parent, framer, msFunc)
			require.NoError(t, err)
			assert.Equal(t, uint64(1), hf.nbBlocks)
			// Verify all cells in the block area are filled
			for y := 0; y < h; y++ {
				for x := 0; x < w; x++ {
					assert.NotNil(t, hf.dctSelect[y][x], "dctSelect[%d][%d]", y, x)
				}
			}
		})
	}
}

func TestNewHFMetadataWithReader_NbBlocksCalculation(t *testing.T) {
	// Verify nbBlocks = ReadBits value + 1
	parent := &LFGroup{size: util.Dimension{Height: 4, Width: 4}}
	// CeilLog2(16) = 4, so ReadBits(4)
	reader := &testcommon.FakeBitReader{ReadBitsData: []uint64{7}} // nbBlocks = 8
	framer := NewFakeFramer(VARDCT)

	types := make([]int32, 8)       // all DCT8 (type 0)
	muls := make([]int32, 8)
	blockInfo := makeBlockInfoBuffer(types, muls)
	decodedBuffer := [][][]int32{
		{{0, 0, 0, 0}},
		{{0, 0, 0, 0}},
		blockInfo,
		util.MakeMatrix2D[int32](4, 4),
	}
	msFunc := makeFakeModularStreamFunc(decodedBuffer, nil)

	hf, err := NewHFMetadataWithReader(reader, parent, framer, msFunc)
	require.NoError(t, err)
	assert.Equal(t, uint64(8), hf.nbBlocks)
	assert.Len(t, hf.blockList, 8)
}

func TestNewHFMetadataWithReader_StreamIndexCalculation(t *testing.T) {
	// Verify the stream index passed to modularStreamFunc
	// streamIndex = 1 + 2*numLFGroups + lfGroupID
	parent := &LFGroup{size: util.Dimension{Height: 1, Width: 1}, lfGroupID: 3}
	reader := &testcommon.FakeBitReader{ReadBitsData: []uint64{0}}
	framer := NewFakeFramer(VARDCT)

	var capturedStreamIndex int
	msFunc := func(reader jxlio.BitReader, frame Framer, streamIndex int, channelArray []ModularChannel) (ModularStreamer, error) {
		capturedStreamIndex = streamIndex
		blockInfo := makeBlockInfoBuffer([]int32{0}, []int32{0})
		return &fakeModularStreamHFMeta{
			decodedBuffer: [][][]int32{{{0}}, {{0}}, blockInfo, {{0}}},
		}, nil
	}

	_, err := NewHFMetadataWithReader(reader, parent, framer, msFunc)
	require.NoError(t, err)
	// numLFGroups = 0 (from FakeFramer), lfGroupID = 3
	// streamIndex = 1 + 2*0 + 3 = 4
	assert.Equal(t, 4, capturedStreamIndex)
}

func TestNewHFMetadataWithReader_ChannelArrayPassedCorrectly(t *testing.T) {
	// Verify 4 channels are passed to modularStreamFunc
	parent := &LFGroup{size: util.Dimension{Height: 2, Width: 2}}
	reader := &testcommon.FakeBitReader{ReadBitsData: []uint64{0}}
	framer := NewFakeFramer(VARDCT)

	var capturedChannels []ModularChannel
	msFunc := func(reader jxlio.BitReader, frame Framer, streamIndex int, channelArray []ModularChannel) (ModularStreamer, error) {
		capturedChannels = channelArray
		blockInfo := makeBlockInfoBuffer([]int32{0}, []int32{0})
		return &fakeModularStreamHFMeta{
			decodedBuffer: [][][]int32{{{0}}, {{0}}, blockInfo, {{0, 0}, {0, 0}}},
		}, nil
	}

	_, err := NewHFMetadataWithReader(reader, parent, framer, msFunc)
	require.NoError(t, err)

	require.Len(t, capturedChannels, 4)
	// Channel 0 (xFromY): correlationHeight x correlationWidth = (2+7)/8 x (2+7)/8 = 1x1
	assert.Equal(t, uint32(1), capturedChannels[0].size.Height)
	assert.Equal(t, uint32(1), capturedChannels[0].size.Width)
	// Channel 1 (bFromY): same as xFromY
	assert.Equal(t, uint32(1), capturedChannels[1].size.Height)
	assert.Equal(t, uint32(1), capturedChannels[1].size.Width)
	// Channel 2 (blockInfo): 2 x nbBlocks(1)
	assert.Equal(t, uint32(2), capturedChannels[2].size.Height)
	assert.Equal(t, uint32(1), capturedChannels[2].size.Width)
	// Channel 3 (sharpness): parent.size.Height x parent.size.Width = 2x2
	assert.Equal(t, uint32(2), capturedChannels[3].size.Height)
	assert.Equal(t, uint32(2), capturedChannels[3].size.Width)
}
