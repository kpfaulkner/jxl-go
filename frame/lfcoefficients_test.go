package frame

import (
	"errors"
	"math"
	"testing"

	"github.com/kpfaulkner/jxl-go/jxlio"
	"github.com/kpfaulkner/jxl-go/testcommon"
	"github.com/kpfaulkner/jxl-go/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// FakeModularStreamer implements ModularStreamer for testing.
type FakeModularStreamer struct {
	decodedBuffer  [][][]int32
	decodeErr      error
	decodeCalled   bool
}

func (f *FakeModularStreamer) decodeChannels(reader jxlio.BitReader, partial bool) error {
	f.decodeCalled = true
	return f.decodeErr
}

func (f *FakeModularStreamer) getDecodedBuffer() [][][]int32 {
	return f.decodedBuffer
}

func (f *FakeModularStreamer) applyTransforms() error {
	return nil
}

func (f *FakeModularStreamer) getChannels() []*ModularChannel {
	return nil
}

// helper to build a Framer with specific settings for LFCoefficients tests
func makeLFCoeffFrame(flags uint64, jpegUpsamplingY, jpegUpsamplingX []int32, scaledDequant []float32, lfThresholds [][]int32) *FakeFramer {
	ff := &FakeFramer{
		header: &FrameHeader{
			jpegUpsamplingX: jpegUpsamplingX,
			jpegUpsamplingY: jpegUpsamplingY,
			Flags:           flags,
			passes:          NewPassesInfo(),
			Encoding:        VARDCT,
		},
		lfGlobal: NewLFGlobal(),
	}
	ff.lfGlobal.scaledDequant = scaledDequant
	// Set valid lfChanCorr defaults to avoid division by zero (colorFactor=0)
	// With xFactorLF=128, bFactorLF=128: kX = 0.0, kB = 0.0 (no correlation)
	ff.lfGlobal.lfChanCorr = &LFChannelCorrelation{
		colorFactor:      84,
		baseCorrelationX: 0.0,
		baseCorrelationB: 0.0,
		xFactorLF:        128,
		bFactorLF:        128,
	}
	if lfThresholds != nil {
		ff.lfGlobal.hfBlockCtx.lfThresholds = lfThresholds
	}
	return ff
}

func TestGetLFIndex_NoThresholds(t *testing.T) {
	thresholds := [][]int32{{}, {}, {}}
	frame := makeLFCoeffFrame(0, []int32{0, 0, 0}, []int32{0, 0, 0}, []float32{1, 1, 1}, thresholds)

	lf := &LFCoefficients{frame: frame}
	// cMap = {1, 0, 2}, so lfQuant channels are accessed as [1], [0], [2]
	lfQuant := util.MakeMatrix3D[int32](3, 2, 2)

	result := lf.getLFIndex(lfQuant, frame.lfGlobal.hfBlockCtx, 0, 0)
	// With no thresholds: all index[i] = 0
	// lfIndex = ((0 * 1 + 0) * 1) + 0 = 0
	assert.Equal(t, int32(0), result)
}

func TestGetLFIndex_SingleThresholdPerChannel(t *testing.T) {
	// Each channel has one threshold
	thresholds := [][]int32{{5}, {10}, {15}}
	frame := makeLFCoeffFrame(0, []int32{0, 0, 0}, []int32{0, 0, 0}, []float32{1, 1, 1}, thresholds)

	lf := &LFCoefficients{frame: frame}

	// cMap = {1, 0, 2}: i=0 -> cMap[0]=1, i=1 -> cMap[1]=0, i=2 -> cMap[2]=2
	// So: index[0] checks lfQuant[1], index[1] checks lfQuant[0], index[2] checks lfQuant[2]
	lfQuant := util.MakeMatrix3D[int32](3, 2, 2)

	tests := []struct {
		name     string
		vals     [3]int32 // values at [channel][0][0] for channels 0, 1, 2
		expected int32
	}{
		{
			"all below thresholds",
			[3]int32{3, 3, 3}, // ch0=3 < thresh[1]=10, ch1=3 < thresh[0]=5, ch2=3 < thresh[2]=15
			// index = [0, 0, 0]
			// lfIndex = ((0 * 2) + 0) * 2 + 0 = 0
			0,
		},
		{
			"channel 0 exceeds (via cMap[0]=1, lfQuant[1])",
			[3]int32{3, 6, 3}, // lfQuant[1][0][0]=6 > thresh[0]=5 -> index[0]=1
			// index = [1, 0, 0]
			// lfIndex = ((1 * 2) + 0) * 2 + 0 = 4
			4,
		},
		{
			"all exceed thresholds",
			[3]int32{11, 6, 16}, // ch0=11>10(index[1]=1), ch1=6>5(index[0]=1), ch2=16>15(index[2]=1)
			// index = [1, 1, 1]
			// lfIndex = ((1 * 2) + 1) * 2 + 1 = 7
			7,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			lfQuant[0][0][0] = tc.vals[0]
			lfQuant[1][0][0] = tc.vals[1]
			lfQuant[2][0][0] = tc.vals[2]

			result := lf.getLFIndex(lfQuant, frame.lfGlobal.hfBlockCtx, 0, 0)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestGetLFIndex_MultipleThresholds(t *testing.T) {
	// Channel 0 has 2 thresholds, channel 1 has 1, channel 2 has 3
	thresholds := [][]int32{{5, 10}, {20}, {1, 2, 3}}
	frame := makeLFCoeffFrame(0, []int32{0, 0, 0}, []int32{0, 0, 0}, []float32{1, 1, 1}, thresholds)

	lf := &LFCoefficients{frame: frame}
	lfQuant := util.MakeMatrix3D[int32](3, 1, 1)

	// cMap = {1, 0, 2}
	// i=0: lfQuant[1] vs thresholds[0] = {5, 10}
	// i=1: lfQuant[0] vs thresholds[1] = {20}
	// i=2: lfQuant[2] vs thresholds[2] = {1, 2, 3}

	// Set values: ch1=12 > 5 and > 10 -> index[0]=2
	//             ch0=25 > 20 -> index[1]=1
	//             ch2=2 > 1 but not > 2 -> index[2]=1
	lfQuant[0][0][0] = 25
	lfQuant[1][0][0] = 12
	lfQuant[2][0][0] = 2

	result := lf.getLFIndex(lfQuant, frame.lfGlobal.hfBlockCtx, 0, 0)

	// lfIndex = index[0] = 2
	// lfIndex *= len(thresholds[2])+1 = 4 -> lfIndex = 8
	// lfIndex += index[2] = 1 -> lfIndex = 9
	// lfIndex *= len(thresholds[1])+1 = 2 -> lfIndex = 18
	// lfIndex += index[1] = 1 -> lfIndex = 19
	assert.Equal(t, int32(19), result)
}

func TestGetLFIndex_WithJpegUpsampling(t *testing.T) {
	thresholds := [][]int32{{5}, {5}, {5}}
	frame := makeLFCoeffFrame(0, []int32{1, 0, 0}, []int32{0, 0, 0}, []float32{1, 1, 1}, thresholds)

	lf := &LFCoefficients{frame: frame}
	// With jpegUpsamplingY[0]=1, sy = y >> 1 for channel i=0
	// lfQuant needs to be large enough: at least 2 rows for channel cMap[0]=1
	lfQuant := util.MakeMatrix3D[int32](3, 2, 2)
	lfQuant[1][0][0] = 10 // channel cMap[0]=1, row 0
	lfQuant[1][1][0] = 2  // channel cMap[0]=1, row 1

	// y=0: sy = 0>>1 = 0 -> lfQuant[1][0][0] = 10 > 5 -> index[0]=1
	result0 := lf.getLFIndex(lfQuant, frame.lfGlobal.hfBlockCtx, 0, 0)
	// y=1: sy = 1>>1 = 0 -> still lfQuant[1][0][0] = 10 > 5 -> index[0]=1
	result1 := lf.getLFIndex(lfQuant, frame.lfGlobal.hfBlockCtx, 1, 0)

	// Both should give the same result since y>>1 maps both to row 0
	assert.Equal(t, result0, result1)
}

func TestPopulatedLFIndex(t *testing.T) {
	thresholds := [][]int32{{5}, {10}, {15}}
	frame := makeLFCoeffFrame(0, []int32{0, 0, 0}, []int32{0, 0, 0}, []float32{1, 1, 1}, thresholds)

	parent := &LFGroup{
		size: util.Dimension{Width: 2, Height: 2},
	}

	lf := &LFCoefficients{
		frame:   frame,
		lfIndex: util.MakeMatrix2D[int32](2, 2),
	}

	lfQuant := util.MakeMatrix3D[int32](3, 2, 2)
	// Set different values at different positions
	lfQuant[0][0][0] = 0  // below all
	lfQuant[1][0][0] = 0
	lfQuant[2][0][0] = 0

	lfQuant[0][0][1] = 11 // ch0 > thresh[1]=10 -> index[1]=1
	lfQuant[1][0][1] = 6  // ch1 > thresh[0]=5 -> index[0]=1
	lfQuant[2][0][1] = 16 // ch2 > thresh[2]=15 -> index[2]=1

	err := lf.populatedLFIndex(parent, lfQuant)
	require.NoError(t, err)

	assert.Equal(t, int32(0), lf.lfIndex[0][0])
	// (0,1): index=[1,1,1], lfIndex = ((1*2)+1)*2+1 = 7
	assert.Equal(t, int32(7), lf.lfIndex[0][1])
}

func TestAdaptiveSmooth_SinglePixel(t *testing.T) {
	// 1x1 has no interior, edges get copied directly
	coeff := util.MakeMatrix3D[float32](3, 1, 1)
	coeff[0][0][0] = 5.0
	coeff[1][0][0] = 10.0
	coeff[2][0][0] = 15.0

	scaledDequant := []float32{1.0, 1.0, 1.0}
	result := adaptiveSmooth(coeff, scaledDequant)

	assert.Equal(t, float32(5.0), result[0][0][0])
	assert.Equal(t, float32(10.0), result[1][0][0])
	assert.Equal(t, float32(15.0), result[2][0][0])
}

func TestAdaptiveSmooth_2x2(t *testing.T) {
	// 2x2: first and last row are edges, no interior processing
	coeff := util.MakeMatrix3D[float32](3, 2, 2)
	for c := 0; c < 3; c++ {
		coeff[c][0][0] = float32(c + 1)
		coeff[c][0][1] = float32(c + 2)
		coeff[c][1][0] = float32(c + 3)
		coeff[c][1][1] = float32(c + 4)
	}

	scaledDequant := []float32{1.0, 1.0, 1.0}
	result := adaptiveSmooth(coeff, scaledDequant)

	// All pixels are on edge rows (y=0 or y=height-1), so they get copied
	for c := 0; c < 3; c++ {
		assert.Equal(t, coeff[c][0][0], result[c][0][0])
		assert.Equal(t, coeff[c][0][1], result[c][0][1])
		assert.Equal(t, coeff[c][1][0], result[c][1][0])
		assert.Equal(t, coeff[c][1][1], result[c][1][1])
	}
}

func TestAdaptiveSmooth_3x3_EdgesCopied(t *testing.T) {
	// 3x3: row 0 and row 2 are edges, col 0 and col 2 are edges for row 1
	coeff := util.MakeMatrix3D[float32](3, 3, 3)
	for c := 0; c < 3; c++ {
		for y := 0; y < 3; y++ {
			for x := 0; x < 3; x++ {
				coeff[c][y][x] = float32(c*9 + y*3 + x + 1)
			}
		}
	}

	scaledDequant := []float32{1.0, 1.0, 1.0}
	result := adaptiveSmooth(coeff, scaledDequant)

	for c := 0; c < 3; c++ {
		// Top row copied
		assert.Equal(t, coeff[c][0][0], result[c][0][0])
		assert.Equal(t, coeff[c][0][1], result[c][0][1])
		assert.Equal(t, coeff[c][0][2], result[c][0][2])
		// Bottom row copied
		assert.Equal(t, coeff[c][2][0], result[c][2][0])
		assert.Equal(t, coeff[c][2][1], result[c][2][1])
		assert.Equal(t, coeff[c][2][2], result[c][2][2])
		// Middle row edges copied
		assert.Equal(t, coeff[c][1][0], result[c][1][0])
		assert.Equal(t, coeff[c][1][2], result[c][1][2])
	}
}

func TestAdaptiveSmooth_3x3_InteriorComputation(t *testing.T) {
	// Use uniform values so weighted == sample and gap is controlled
	coeff := util.MakeMatrix3D[float32](3, 3, 3)
	for c := 0; c < 3; c++ {
		for y := 0; y < 3; y++ {
			for x := 0; x < 3; x++ {
				coeff[c][y][x] = 1.0
			}
		}
	}

	scaledDequant := []float32{1.0, 1.0, 1.0}
	result := adaptiveSmooth(coeff, scaledDequant)

	// With all 1.0 coefficients:
	// weighted = 0.05226...*1 + 0.20345...*4 + 0.03348...*4 = 1.0 (approximately)
	// gap_val = |1.0 - ~1.0| * 1.0 ≈ 0
	// gap starts at 0.5, g = max(0.5, ~0) = 0.5
	// After transform: gap = max(0, 3.0 - 4.0*0.5) = max(0, 1.0) = 1.0
	// result = (1.0 - weighted)*gap + weighted ≈ 1.0

	// The interior pixel should be approximately 1.0
	assert.InDelta(t, 1.0, float64(result[0][1][1]), 0.01)
	assert.InDelta(t, 1.0, float64(result[1][1][1]), 0.01)
	assert.InDelta(t, 1.0, float64(result[2][1][1]), 0.01)
}

func TestAdaptiveSmooth_3x3_StrongEdge(t *testing.T) {
	// Strong edge: center pixel very different from neighbors
	coeff := util.MakeMatrix3D[float32](3, 3, 3)
	for c := 0; c < 3; c++ {
		for y := 0; y < 3; y++ {
			for x := 0; x < 3; x++ {
				coeff[c][y][x] = 0.0
			}
		}
		coeff[c][1][1] = 100.0 // strong spike
	}

	scaledDequant := []float32{1.0, 1.0, 1.0}
	result := adaptiveSmooth(coeff, scaledDequant)

	// With a strong edge, |sample - weighted| * sd should be large
	// gap should saturate and become max(0, 3.0 - 4.0*large) = 0
	// result = (100 - weighted)*0 + weighted = weighted (strong smoothing)
	// weighted = 0.05226...*100 + 0.20345...*0 + 0.03348...*0 = 5.226...
	expected := float32(0.05226273532324128) * 100.0
	assert.InDelta(t, expected, result[0][1][1], 0.01)
}

func TestAdaptiveSmooth_LargeScaledDequant(t *testing.T) {
	// Use a spike pattern where the center differs significantly from neighbors
	// so the weighted average differs from the sample, giving a non-zero gap
	makeCoeff := func() [][][]float32 {
		coeff := util.MakeMatrix3D[float32](3, 3, 3)
		for c := 0; c < 3; c++ {
			for y := 0; y < 3; y++ {
				for x := 0; x < 3; x++ {
					coeff[c][y][x] = 0.0
				}
			}
			coeff[c][1][1] = 10.0 // spike at center
		}
		return coeff
	}

	smallSD := []float32{0.001, 0.001, 0.001}
	largeSD := []float32{1000.0, 1000.0, 1000.0}

	resultSmall := adaptiveSmooth(makeCoeff(), smallSD)
	resultLarge := adaptiveSmooth(makeCoeff(), largeSD)

	originalCenter := float32(10.0)

	// With small scaledDequant, gap is small -> gap transform yields ~3.0 -> less smoothing
	// With large scaledDequant, gap is amplified -> gap transform yields ~0 -> more smoothing
	smallDiff := math.Abs(float64(resultSmall[0][1][1] - originalCenter))
	largeDiff := math.Abs(float64(resultLarge[0][1][1] - originalCenter))

	// Large SD should cause more smoothing (larger deviation from original)
	assert.Greater(t, largeDiff, smallDiff)
}

func TestNewLFCoefficientsWithReader_AdaptiveSmoothingWithSubsampling(t *testing.T) {
	// Adaptive smoothing + subsampled should return error
	frame := makeLFCoeffFrame(0, []int32{1, 0, 0}, []int32{0, 0, 0}, []float32{1, 1, 1}, nil)

	parent := &LFGroup{
		size: util.Dimension{Width: 4, Height: 4},
	}

	reader := &testcommon.FakeBitReader{
		ReadBitsData: []uint64{0}, // extraPrecision
	}

	lf, err := NewLFCoefficientsWithReader(reader, parent, frame, nil, nil)
	assert.Error(t, err)
	assert.Nil(t, lf)
	assert.Contains(t, err.Error(), "Adaptive smoothing is incompatible with subsampling")
}

func TestNewLFCoefficientsWithReader_ReadBitsError(t *testing.T) {
	// Skip adaptive smoothing by setting SKIP_ADAPTIVE_LF_SMOOTHING flag
	frame := makeLFCoeffFrame(SKIP_ADAPTIVE_LF_SMOOTHING, []int32{0, 0, 0}, []int32{0, 0, 0}, []float32{1, 1, 1}, nil)

	parent := &LFGroup{
		size: util.Dimension{Width: 2, Height: 2},
	}

	reader := &testcommon.FakeBitReader{
		ReadBitsData: []uint64{}, // empty -> error on ReadBits
	}

	lf, err := NewLFCoefficientsWithReader(reader, parent, frame, nil, nil)
	assert.Error(t, err)
	assert.Nil(t, lf)
}

func TestNewLFCoefficientsWithReader_ModularStreamFuncError(t *testing.T) {
	frame := makeLFCoeffFrame(SKIP_ADAPTIVE_LF_SMOOTHING, []int32{0, 0, 0}, []int32{0, 0, 0}, []float32{1, 1, 1}, nil)

	parent := &LFGroup{
		size:      util.Dimension{Width: 2, Height: 2},
		lfGroupID: 0,
	}

	reader := &testcommon.FakeBitReader{
		ReadBitsData: []uint64{0}, // extraPrecision = 0
	}

	modularFunc := func(reader jxlio.BitReader, frame Framer, streamIndex int, channelCount int, ecStart int, channelArray []ModularChannel) (ModularStreamer, error) {
		return nil, errors.New("modular stream error")
	}

	lf, err := NewLFCoefficientsWithReader(reader, parent, frame, nil, modularFunc)
	assert.Error(t, err)
	assert.Nil(t, lf)
	assert.Contains(t, err.Error(), "modular stream error")
}

func TestNewLFCoefficientsWithReader_DecodeChannelsError(t *testing.T) {
	frame := makeLFCoeffFrame(SKIP_ADAPTIVE_LF_SMOOTHING, []int32{0, 0, 0}, []int32{0, 0, 0}, []float32{1, 1, 1}, nil)

	parent := &LFGroup{
		size:      util.Dimension{Width: 2, Height: 2},
		lfGroupID: 0,
	}

	reader := &testcommon.FakeBitReader{
		ReadBitsData: []uint64{0},
	}

	fakeStream := &FakeModularStreamer{
		decodeErr: errors.New("decode error"),
	}

	modularFunc := func(reader jxlio.BitReader, frame Framer, streamIndex int, channelCount int, ecStart int, channelArray []ModularChannel) (ModularStreamer, error) {
		return fakeStream, nil
	}

	lf, err := NewLFCoefficientsWithReader(reader, parent, frame, nil, modularFunc)
	assert.Error(t, err)
	assert.Nil(t, lf)
	assert.Contains(t, err.Error(), "decode error")
}

func TestNewLFCoefficientsWithReader_Success_NoSmoothing(t *testing.T) {
	thresholds := [][]int32{{5}, {10}, {15}}
	frame := makeLFCoeffFrame(SKIP_ADAPTIVE_LF_SMOOTHING, []int32{0, 0, 0}, []int32{0, 0, 0}, []float32{2.0, 3.0, 4.0}, thresholds)

	parent := &LFGroup{
		size:      util.Dimension{Width: 2, Height: 2},
		lfGroupID: 0,
	}

	reader := &testcommon.FakeBitReader{
		ReadBitsData: []uint64{0}, // extraPrecision = 0
	}

	// cMap = {1, 0, 2}: decoded buffer channel mapping
	// i=0 -> scaledDequant[0]=2.0, writes to dequantLFCoeff[0], reads from lfQuant[cMap[0]]=lfQuant[1]
	// i=1 -> scaledDequant[1]=3.0, writes to dequantLFCoeff[1], reads from lfQuant[cMap[1]]=lfQuant[0]
	// i=2 -> scaledDequant[2]=4.0, writes to dequantLFCoeff[2], reads from lfQuant[cMap[2]]=lfQuant[2]
	decodedBuf := util.MakeMatrix3D[int32](3, 2, 2)
	decodedBuf[0][0][0] = 10 // -> dequantLFCoeff[1][0][0] = 10 * 3.0 = 30.0
	decodedBuf[1][0][0] = 5  // -> dequantLFCoeff[0][0][0] = 5 * 2.0 = 10.0
	decodedBuf[2][0][0] = 3  // -> dequantLFCoeff[2][0][0] = 3 * 4.0 = 12.0

	fakeStream := &FakeModularStreamer{
		decodedBuffer: decodedBuf,
	}

	modularFunc := func(reader jxlio.BitReader, frame Framer, streamIndex int, channelCount int, ecStart int, channelArray []ModularChannel) (ModularStreamer, error) {
		return fakeStream, nil
	}

	lf, err := NewLFCoefficientsWithReader(reader, parent, frame, nil, modularFunc)
	require.NoError(t, err)
	require.NotNil(t, lf)

	// With SKIP_ADAPTIVE_LF_SMOOTHING and no subsampling, dequantLFCoeff is assigned directly
	// extraPrecision=0: xx=1<<0=1, sd = scaledDequant[i] / 1
	assert.InDelta(t, 10.0, float64(lf.dequantLFCoeff[0][0][0]), 0.001)
	assert.InDelta(t, 30.0, float64(lf.dequantLFCoeff[1][0][0]), 0.001)
	assert.InDelta(t, 12.0, float64(lf.dequantLFCoeff[2][0][0]), 0.001)
}

func TestNewLFCoefficientsWithReader_ExtraPrecision(t *testing.T) {
	frame := makeLFCoeffFrame(SKIP_ADAPTIVE_LF_SMOOTHING, []int32{0, 0, 0}, []int32{0, 0, 0}, []float32{8.0, 8.0, 8.0}, nil)

	parent := &LFGroup{
		size:      util.Dimension{Width: 1, Height: 1},
		lfGroupID: 0,
	}

	reader := &testcommon.FakeBitReader{
		ReadBitsData: []uint64{2}, // extraPrecision = 2, so xx = 1<<2 = 4
	}

	decodedBuf := util.MakeMatrix3D[int32](3, 1, 1)
	decodedBuf[0][0][0] = 4  // -> dequant[1] = 4 * (8.0/4) = 8.0
	decodedBuf[1][0][0] = 12 // -> dequant[0] = 12 * (8.0/4) = 24.0
	decodedBuf[2][0][0] = 8  // -> dequant[2] = 8 * (8.0/4) = 16.0

	fakeStream := &FakeModularStreamer{decodedBuffer: decodedBuf}

	modularFunc := func(reader jxlio.BitReader, frame Framer, streamIndex int, channelCount int, ecStart int, channelArray []ModularChannel) (ModularStreamer, error) {
		return fakeStream, nil
	}

	lf, err := NewLFCoefficientsWithReader(reader, parent, frame, nil, modularFunc)
	require.NoError(t, err)

	// sd = scaledDequant[i] / (1 << extraPrecision) = 8.0 / 4.0 = 2.0
	assert.InDelta(t, 24.0, float64(lf.dequantLFCoeff[0][0][0]), 0.001)
	assert.InDelta(t, 8.0, float64(lf.dequantLFCoeff[1][0][0]), 0.001)
	assert.InDelta(t, 16.0, float64(lf.dequantLFCoeff[2][0][0]), 0.001)
}

func TestNewLFCoefficientsWithReader_ChannelCorrelation(t *testing.T) {
	// Test the chroma-from-luma correlation step (subSampled=false, adaptiveSmoothing off)
	frame := makeLFCoeffFrame(SKIP_ADAPTIVE_LF_SMOOTHING, []int32{0, 0, 0}, []int32{0, 0, 0}, []float32{1.0, 1.0, 1.0}, nil)
	// Set up channel correlation
	frame.lfGlobal.lfChanCorr = &LFChannelCorrelation{
		colorFactor:      84,
		baseCorrelationX: 0.0,
		baseCorrelationB: 1.0,
		xFactorLF:        128,
		bFactorLF:        128,
	}

	parent := &LFGroup{
		size:      util.Dimension{Width: 1, Height: 1},
		lfGroupID: 0,
	}

	reader := &testcommon.FakeBitReader{
		ReadBitsData: []uint64{0},
	}

	decodedBuf := util.MakeMatrix3D[int32](3, 1, 1)
	// cMap = {1, 0, 2}
	decodedBuf[0][0][0] = 0  // -> dequant[1] (Y channel) = 0 * 1.0 = 0
	decodedBuf[1][0][0] = 10 // -> dequant[0] (X channel) = 10 * 1.0 = 10
	decodedBuf[2][0][0] = 20 // -> dequant[2] (B channel) = 20 * 1.0 = 20

	fakeStream := &FakeModularStreamer{decodedBuffer: decodedBuf}
	modularFunc := func(reader jxlio.BitReader, frame Framer, streamIndex int, channelCount int, ecStart int, channelArray []ModularChannel) (ModularStreamer, error) {
		return fakeStream, nil
	}

	lf, err := NewLFCoefficientsWithReader(reader, parent, frame, nil, modularFunc)
	require.NoError(t, err)

	// kX = baseCorrelationX + (xFactorLF - 128) / colorFactor = 0.0 + 0/84 = 0.0
	// kB = baseCorrelationB + (bFactorLF - 128) / colorFactor = 1.0 + 0/84 = 1.0
	// dequantY = dequantLFCoeff[1] = 0.0
	// dequantX = dequantLFCoeff[0] += kX * dequantY = 10.0 + 0.0 * 0.0 = 10.0
	// dequantB = dequantLFCoeff[2] += kB * dequantY = 20.0 + 1.0 * 0.0 = 20.0
	assert.InDelta(t, 10.0, float64(lf.dequantLFCoeff[0][0][0]), 0.001)
	assert.InDelta(t, 0.0, float64(lf.dequantLFCoeff[1][0][0]), 0.001)
	assert.InDelta(t, 20.0, float64(lf.dequantLFCoeff[2][0][0]), 0.001)
}

func TestNewLFCoefficientsWithReader_ChannelCorrelation_NonZeroY(t *testing.T) {
	// Note: xFactorLF and bFactorLF are uint32; values < 128 cause uint32 underflow
	// in (xFactorLF - 128). Use values >= 128 to test the correlation correctly.
	frame := makeLFCoeffFrame(SKIP_ADAPTIVE_LF_SMOOTHING, []int32{0, 0, 0}, []int32{0, 0, 0}, []float32{1.0, 1.0, 1.0}, nil)
	frame.lfGlobal.lfChanCorr = &LFChannelCorrelation{
		colorFactor:      84,
		baseCorrelationX: 0.5,
		baseCorrelationB: 0.0,
		xFactorLF:        212, // (212-128)/84 = 84/84 = 1.0 -> kX = 0.5 + 1.0 = 1.5
		bFactorLF:        170, // (170-128)/84 = 42/84 = 0.5 -> kB = 0.0 + 0.5 = 0.5
	}

	parent := &LFGroup{
		size:      util.Dimension{Width: 1, Height: 1},
		lfGroupID: 0,
	}

	decodedBuf := util.MakeMatrix3D[int32](3, 1, 1)
	decodedBuf[0][0][0] = 6 // -> dequant[1] (Y) = 6
	decodedBuf[1][0][0] = 4 // -> dequant[0] (X) = 4
	decodedBuf[2][0][0] = 8 // -> dequant[2] (B) = 8

	reader := &testcommon.FakeBitReader{ReadBitsData: []uint64{0}}
	fakeStream := &FakeModularStreamer{decodedBuffer: decodedBuf}
	modularFunc := func(reader jxlio.BitReader, frame Framer, streamIndex int, channelCount int, ecStart int, channelArray []ModularChannel) (ModularStreamer, error) {
		return fakeStream, nil
	}

	lf, err := NewLFCoefficientsWithReader(reader, parent, frame, nil, modularFunc)
	require.NoError(t, err)

	// kX = 0.5 + (212-128)/84 = 0.5 + 1.0 = 1.5
	// kB = 0.0 + (170-128)/84 = 0.0 + 0.5 = 0.5
	// Y = 6.0
	// X = 4.0 + 1.5 * 6.0 = 4.0 + 9.0 = 13.0
	// B = 8.0 + 0.5 * 6.0 = 8.0 + 3.0 = 11.0
	assert.InDelta(t, 13.0, float64(lf.dequantLFCoeff[0][0][0]), 0.01)
	assert.InDelta(t, 6.0, float64(lf.dequantLFCoeff[1][0][0]), 0.01)
	assert.InDelta(t, 11.0, float64(lf.dequantLFCoeff[2][0][0]), 0.01)
}

func TestNewLFCoefficientsWithReader_SubsampledSkipsCorrelation(t *testing.T) {
	// Subsampled + SKIP_ADAPTIVE_LF_SMOOTHING should skip channel correlation.
	// jpegUpsamplingY[0]=1: channel i=0 (cMap[0]=1) has sizeY = 2>>1 = 1.
	// The decoded buffer must match these reduced dimensions per channel.
	frame := makeLFCoeffFrame(SKIP_ADAPTIVE_LF_SMOOTHING, []int32{1, 0, 0}, []int32{0, 0, 0}, []float32{1.0, 1.0, 1.0}, nil)
	frame.lfGlobal.lfChanCorr = &LFChannelCorrelation{
		colorFactor:      84,
		baseCorrelationX: 0.5,
		baseCorrelationB: 1.0,
		xFactorLF:        212,
		bFactorLF:        128,
	}

	parent := &LFGroup{
		size:      util.Dimension{Width: 2, Height: 2},
		lfGroupID: 0,
	}
	reader := &testcommon.FakeBitReader{ReadBitsData: []uint64{0}}

	// The dequant loop iterates over len(lfQuant[c]) rows, which must <= dequantLFCoeff[i] rows.
	// dequantLFCoeff channel sizes:
	//   i=0 (cMap[0]=1): sizeY = 2>>1 = 1, so lfQuant[1] must have 1 row
	//   i=1 (cMap[1]=0): sizeY = 2>>0 = 2, so lfQuant[0] must have 2 rows
	//   i=2 (cMap[2]=2): sizeY = 2>>0 = 2, so lfQuant[2] must have 2 rows
	decodedBuf := make([][][]int32, 3)
	decodedBuf[0] = util.MakeMatrix2D[int32](2, 2) // ch0 -> dequant[1], 2 rows
	decodedBuf[1] = util.MakeMatrix2D[int32](1, 2) // ch1 -> dequant[0], 1 row
	decodedBuf[2] = util.MakeMatrix2D[int32](2, 2) // ch2 -> dequant[2], 2 rows

	decodedBuf[0][0][0] = 5
	decodedBuf[1][0][0] = 10
	decodedBuf[2][0][0] = 15

	fakeStream := &FakeModularStreamer{decodedBuffer: decodedBuf}
	modularFunc := func(reader jxlio.BitReader, frame Framer, streamIndex int, channelCount int, ecStart int, channelArray []ModularChannel) (ModularStreamer, error) {
		return fakeStream, nil
	}

	lf, err := NewLFCoefficientsWithReader(reader, parent, frame, nil, modularFunc)
	require.NoError(t, err)

	// Since subSampled=true, channel correlation is skipped entirely
	// dequant[0] = lfQuant[cMap[0]=1][0][0] * scaledDequant[0] = 10 * 1.0 = 10.0
	assert.InDelta(t, 10.0, float64(lf.dequantLFCoeff[0][0][0]), 0.001)
}

func TestNewLFCoefficientsWithReader_LFIndexPopulated(t *testing.T) {
	thresholds := [][]int32{{5}, {10}, {15}}
	frame := makeLFCoeffFrame(SKIP_ADAPTIVE_LF_SMOOTHING, []int32{0, 0, 0}, []int32{0, 0, 0}, []float32{1.0, 1.0, 1.0}, thresholds)

	parent := &LFGroup{
		size:      util.Dimension{Width: 2, Height: 1},
		lfGroupID: 0,
	}
	reader := &testcommon.FakeBitReader{ReadBitsData: []uint64{0}}

	decodedBuf := util.MakeMatrix3D[int32](3, 1, 2)
	// Position (0,0): all zero -> all below thresholds -> lfIndex = 0
	decodedBuf[0][0][0] = 0
	decodedBuf[1][0][0] = 0
	decodedBuf[2][0][0] = 0
	// Position (0,1): values above thresholds
	decodedBuf[0][0][1] = 11 // ch0 (via cMap[1]=0) > thresh[1]=10 -> index[1]=1
	decodedBuf[1][0][1] = 6  // ch1 (via cMap[0]=1) > thresh[0]=5 -> index[0]=1
	decodedBuf[2][0][1] = 16 // ch2 (via cMap[2]=2) > thresh[2]=15 -> index[2]=1

	fakeStream := &FakeModularStreamer{decodedBuffer: decodedBuf}
	modularFunc := func(reader jxlio.BitReader, frame Framer, streamIndex int, channelCount int, ecStart int, channelArray []ModularChannel) (ModularStreamer, error) {
		return fakeStream, nil
	}

	lf, err := NewLFCoefficientsWithReader(reader, parent, frame, nil, modularFunc)
	require.NoError(t, err)

	assert.Equal(t, int32(0), lf.lfIndex[0][0])
	// index=[1,1,1]: ((1*2)+1)*2+1 = 7
	assert.Equal(t, int32(7), lf.lfIndex[0][1])
}

func TestNewLFCoefficientsWithReader_ModularStreamParams(t *testing.T) {
	frame := makeLFCoeffFrame(SKIP_ADAPTIVE_LF_SMOOTHING, []int32{0, 0, 0}, []int32{0, 0, 0}, []float32{1.0, 1.0, 1.0}, nil)

	parent := &LFGroup{
		size:      util.Dimension{Width: 4, Height: 3},
		lfGroupID: 5,
	}
	reader := &testcommon.FakeBitReader{ReadBitsData: []uint64{0}}

	var capturedStreamIndex int
	var capturedChannelCount int
	var capturedChannelArray []ModularChannel

	decodedBuf := util.MakeMatrix3D[int32](3, 3, 4)
	fakeStream := &FakeModularStreamer{decodedBuffer: decodedBuf}
	modularFunc := func(reader jxlio.BitReader, frame Framer, streamIndex int, channelCount int, ecStart int, channelArray []ModularChannel) (ModularStreamer, error) {
		capturedStreamIndex = streamIndex
		capturedChannelCount = channelCount
		capturedChannelArray = channelArray
		return fakeStream, nil
	}

	_, err := NewLFCoefficientsWithReader(reader, parent, frame, nil, modularFunc)
	require.NoError(t, err)

	// streamIndex = 1 + parent.lfGroupID = 1 + 5 = 6
	assert.Equal(t, 6, capturedStreamIndex)
	// channelCount = len(info) = 3
	assert.Equal(t, 3, capturedChannelCount)
	// Verify channel dimensions match parent size (no upsampling)
	require.Len(t, capturedChannelArray, 3)
	// cMap = {1, 0, 2}: info[cMap[0]]=info[1], info[cMap[1]]=info[0], info[cMap[2]]=info[2]
	assert.Equal(t, uint32(3), capturedChannelArray[1].size.Height)
	assert.Equal(t, uint32(4), capturedChannelArray[1].size.Width)
}

// ---- Additional getLFIndex tests ----

func TestGetLFIndex_MultiplePositions(t *testing.T) {
	// Test that getLFIndex works correctly at different (y,x) positions within a grid
	thresholds := [][]int32{{10}, {20}, {30}}
	frame := makeLFCoeffFrame(0, []int32{0, 0, 0}, []int32{0, 0, 0}, []float32{1, 1, 1}, thresholds)

	lf := &LFCoefficients{frame: frame}
	// 3x3 grid with varied values
	lfQuant := util.MakeMatrix3D[int32](3, 3, 3)

	// Position (0,0): all below thresholds
	// Position (1,1): channel 1 above threshold
	lfQuant[1][1][1] = 15 // cMap[0]=1, check vs thresholds[0]={10}: 15 > 10 -> index[0]=1

	// Position (2,2): all above thresholds
	lfQuant[0][2][2] = 25 // cMap[1]=0, check vs thresholds[1]={20}: 25 > 20 -> index[1]=1
	lfQuant[1][2][2] = 15 // cMap[0]=1, check vs thresholds[0]={10}: 15 > 10 -> index[0]=1
	lfQuant[2][2][2] = 35 // cMap[2]=2, check vs thresholds[2]={30}: 35 > 30 -> index[2]=1

	hfctx := frame.lfGlobal.hfBlockCtx

	// (0,0): index=[0,0,0], lfIndex = ((0*2)+0)*2+0 = 0
	assert.Equal(t, int32(0), lf.getLFIndex(lfQuant, hfctx, 0, 0))
	// (1,1): index=[1,0,0], lfIndex = ((1*2)+0)*2+0 = 4
	assert.Equal(t, int32(4), lf.getLFIndex(lfQuant, hfctx, 1, 1))
	// (2,2): index=[1,1,1], lfIndex = ((1*2)+1)*2+1 = 7
	assert.Equal(t, int32(7), lf.getLFIndex(lfQuant, hfctx, 2, 2))
}

func TestGetLFIndex_ThresholdExactBoundary(t *testing.T) {
	// Values at exact threshold should NOT count (condition is strictly greater than)
	thresholds := [][]int32{{10}, {20}, {30}}
	frame := makeLFCoeffFrame(0, []int32{0, 0, 0}, []int32{0, 0, 0}, []float32{1, 1, 1}, thresholds)

	lf := &LFCoefficients{frame: frame}
	lfQuant := util.MakeMatrix3D[int32](3, 1, 1)

	// Set all channels to exactly their threshold values
	// cMap = {1, 0, 2}: i=0 checks lfQuant[1] vs thresholds[0]={10}
	lfQuant[1][0][0] = 10 // exactly at threshold[0]=10, NOT > 10
	lfQuant[0][0][0] = 20 // exactly at threshold[1]=20, NOT > 20
	lfQuant[2][0][0] = 30 // exactly at threshold[2]=30, NOT > 30

	result := lf.getLFIndex(lfQuant, frame.lfGlobal.hfBlockCtx, 0, 0)
	// All at exact boundary -> none exceed -> all index[i]=0
	assert.Equal(t, int32(0), result)

	// Now set values one above threshold
	lfQuant[1][0][0] = 11
	lfQuant[0][0][0] = 21
	lfQuant[2][0][0] = 31

	result = lf.getLFIndex(lfQuant, frame.lfGlobal.hfBlockCtx, 0, 0)
	// All exceed -> index=[1,1,1], lfIndex = ((1*2)+1)*2+1 = 7
	assert.Equal(t, int32(7), result)
}

func TestGetLFIndex_BothXAndYUpsampling(t *testing.T) {
	// Test with both X and Y jpeg upsampling simultaneously
	thresholds := [][]int32{{5}, {5}, {5}}
	// Channel 0: both X and Y upsampling by 1 (halving both dimensions)
	frame := makeLFCoeffFrame(0, []int32{1, 0, 0}, []int32{1, 0, 0}, []float32{1, 1, 1}, thresholds)

	lf := &LFCoefficients{frame: frame}
	// Need enough rows/cols for full-size access
	lfQuant := util.MakeMatrix3D[int32](3, 4, 4)

	// For channel i=0 (cMap[0]=1): sy = y>>1, sx = x>>1
	// So positions (0,0), (0,1), (1,0), (1,1) all map to lfQuant[1][0][0]
	lfQuant[1][0][0] = 10 // > 5

	hfctx := frame.lfGlobal.hfBlockCtx

	// All four positions should yield the same lfIndex since they map to the same cell
	r00 := lf.getLFIndex(lfQuant, hfctx, 0, 0)
	r01 := lf.getLFIndex(lfQuant, hfctx, 0, 1)
	r10 := lf.getLFIndex(lfQuant, hfctx, 1, 0)
	r11 := lf.getLFIndex(lfQuant, hfctx, 1, 1)

	assert.Equal(t, r00, r01)
	assert.Equal(t, r00, r10)
	assert.Equal(t, r00, r11)

	// Position (2,2) maps to lfQuant[1][1][1] which is 0 (< 5)
	r22 := lf.getLFIndex(lfQuant, hfctx, 2, 2)
	assert.NotEqual(t, r00, r22)
}

// ---- Additional populatedLFIndex tests ----

func TestPopulatedLFIndex_LargerGrid(t *testing.T) {
	thresholds := [][]int32{{5, 15}, {10}, {20}}
	frame := makeLFCoeffFrame(0, []int32{0, 0, 0}, []int32{0, 0, 0}, []float32{1, 1, 1}, thresholds)

	parent := &LFGroup{
		size: util.Dimension{Width: 3, Height: 3},
	}

	lf := &LFCoefficients{
		frame:   frame,
		lfIndex: util.MakeMatrix2D[int32](3, 3),
	}

	lfQuant := util.MakeMatrix3D[int32](3, 3, 3)

	// Set varied values across the grid
	// (0,0): all zero -> lfIndex = 0
	// (1,0): lfQuant[1][1][0] = 7 -> exceeds thresholds[0]={5,15}: 7>5 but not >15 -> index[0]=1
	lfQuant[1][1][0] = 7
	// (2,2): lfQuant[1][2][2] = 20 -> exceeds both 5 and 15 -> index[0]=2
	//        lfQuant[0][2][2] = 12 -> exceeds 10 -> index[1]=1
	//        lfQuant[2][2][2] = 25 -> exceeds 20 -> index[2]=1
	lfQuant[1][2][2] = 20
	lfQuant[0][2][2] = 12
	lfQuant[2][2][2] = 25

	err := lf.populatedLFIndex(parent, lfQuant)
	require.NoError(t, err)

	// (0,0): index=[0,0,0]
	// lfIndex = ((0 * 2) + 0) * 2 + 0 = 0
	assert.Equal(t, int32(0), lf.lfIndex[0][0])

	// (1,0): index=[1,0,0]
	// lfIndex = 1 * (1+1) = 2, 2 + 0 = 2, 2 * (1+1) = 4, 4 + 0 = 4
	assert.Equal(t, int32(4), lf.lfIndex[1][0])

	// (2,2): index=[2,1,1]
	// lfIndex = 2 * 2 = 4, 4 + 1 = 5, 5 * 2 = 10, 10 + 1 = 11
	assert.Equal(t, int32(11), lf.lfIndex[2][2])
}

// ---- Additional adaptiveSmooth tests ----

func TestAdaptiveSmooth_5x5_MultipleInteriorPixels(t *testing.T) {
	// 5x5 matrix: rows 1-3 are interior, cols 1-3 are interior within those rows
	// This tests that multiple interior pixels are each computed independently
	coeff := util.MakeMatrix3D[float32](3, 5, 5)
	// Set a pattern: channel 0 has a horizontal gradient
	for y := 0; y < 5; y++ {
		for x := 0; x < 5; x++ {
			coeff[0][y][x] = float32(x) * 2.0
		}
	}
	// Channel 1: uniform
	for y := 0; y < 5; y++ {
		for x := 0; x < 5; x++ {
			coeff[1][y][x] = 5.0
		}
	}
	// Channel 2: spike at (2,2)
	coeff[2][2][2] = 50.0

	scaledDequant := []float32{1.0, 1.0, 1.0}
	result := adaptiveSmooth(coeff, scaledDequant)

	// Edge rows (0 and 4) should be copied exactly
	for x := 0; x < 5; x++ {
		assert.Equal(t, coeff[0][0][x], result[0][0][x])
		assert.Equal(t, coeff[0][4][x], result[0][4][x])
		assert.Equal(t, coeff[2][0][x], result[2][0][x])
		assert.Equal(t, coeff[2][4][x], result[2][4][x])
	}
	// Edge columns on interior rows should be copied exactly
	for y := 1; y <= 3; y++ {
		assert.Equal(t, coeff[0][y][0], result[0][y][0])
		assert.Equal(t, coeff[0][y][4], result[0][y][4])
	}
	// Uniform channel 1: interior should remain approximately 5.0
	for y := 1; y <= 3; y++ {
		for x := 1; x <= 3; x++ {
			assert.InDelta(t, 5.0, float64(result[1][y][x]), 0.01,
				"channel 1 interior (%d,%d)", y, x)
		}
	}
	// Channel 2 spike: (2,2) should be smoothed closer to neighbors (0)
	assert.Less(t, result[2][2][2], float32(50.0), "spike should be smoothed down")
}

func TestAdaptiveSmooth_CrossChannelGapInteraction(t *testing.T) {
	// The gap is shared across all 3 channels (max of gap across channels).
	// If channel 0 has a strong edge but channels 1,2 are smooth,
	// the gap from channel 0 should affect channels 1,2.
	coeff := util.MakeMatrix3D[float32](3, 3, 3)

	// Channel 0: strong spike at center
	coeff[0][1][1] = 100.0

	// Channel 1: mild difference at center
	for y := 0; y < 3; y++ {
		for x := 0; x < 3; x++ {
			coeff[1][y][x] = 5.0
		}
	}
	coeff[1][1][1] = 6.0 // small deviation

	// Channel 2: uniform
	for y := 0; y < 3; y++ {
		for x := 0; x < 3; x++ {
			coeff[2][y][x] = 10.0
		}
	}

	scaledDequant := []float32{1.0, 1.0, 1.0}
	result := adaptiveSmooth(coeff, scaledDequant)

	// The strong edge in channel 0 should drive the gap high (> 0.5),
	// making gap transform = max(0, 3 - 4*large) = 0
	// This means ALL channels get fully smoothed at (1,1):
	// result = (sample - weighted) * 0 + weighted = weighted

	// Channel 2 was uniform, so weighted ≈ 10.0, result ≈ 10.0
	assert.InDelta(t, 10.0, float64(result[2][1][1]), 0.01)

	// Channel 0: heavily smoothed from 100 toward weighted average
	assert.Less(t, result[0][1][1], float32(10.0), "ch0 spike should be strongly smoothed")
}

func TestAdaptiveSmooth_PreciseWeightedAverage(t *testing.T) {
	// Verify the exact weighted average formula:
	// weighted = 0.05226273532324128*sample + 0.20345139757231578*adjacent + 0.0334829185968739*diag
	// When gap transform = 0, result = weighted (full smoothing)
	coeff := util.MakeMatrix3D[float32](3, 3, 3)

	// Set specific values for channel 0 only (others zero -> won't affect gap much)
	coeff[0][0][0] = 1.0
	coeff[0][0][1] = 2.0
	coeff[0][0][2] = 3.0
	coeff[0][1][0] = 4.0
	coeff[0][1][1] = 5.0 // center sample
	coeff[0][1][2] = 6.0
	coeff[0][2][0] = 7.0
	coeff[0][2][1] = 8.0
	coeff[0][2][2] = 9.0

	// Use large scaledDequant to force gap to saturate -> gap transform = 0
	// Then result = (sample - weighted)*0 + weighted = weighted
	scaledDequant := []float32{1000.0, 1000.0, 1000.0}
	result := adaptiveSmooth(coeff, scaledDequant)

	// Compute expected weighted average at (1,1) for channel 0
	sample := float32(5.0)
	adjacent := float32(4.0 + 6.0 + 2.0 + 8.0) // left, right, above, below
	diag := float32(1.0 + 3.0 + 7.0 + 9.0)      // corners
	expectedWeighted := float32(0.05226273532324128*float64(sample) +
		0.20345139757231578*float64(adjacent) +
		0.0334829185968739*float64(diag))

	assert.InDelta(t, float64(expectedWeighted), float64(result[0][1][1]), 0.001)
}

func TestAdaptiveSmooth_GapClampedToZeroMinimum(t *testing.T) {
	// When gap is very large (strong edge), gap transform = max(0, 3-4*gap) = 0
	// This means result = (sample-weighted)*0 + weighted = weighted
	coeff := util.MakeMatrix3D[float32](3, 3, 3)
	// Huge spike at center
	coeff[0][1][1] = 1000.0

	scaledDequant := []float32{100.0, 100.0, 100.0}
	result := adaptiveSmooth(coeff, scaledDequant)

	// Weighted = 0.05226...*1000 + 0 + 0 = 52.26...
	expectedWeighted := float32(0.05226273532324128 * 1000.0)
	assert.InDelta(t, float64(expectedWeighted), float64(result[0][1][1]), 0.01)
}

func TestAdaptiveSmooth_DifferentScaledDequantPerChannel(t *testing.T) {
	// Each channel has a different scaledDequant, affecting gap differently
	// Gap is the max across all channels
	coeff := util.MakeMatrix3D[float32](3, 3, 3)
	// All channels have a spike at center
	for c := 0; c < 3; c++ {
		coeff[c][1][1] = 10.0
	}

	// Channel 0 has large sd (amplifies gap), channels 1,2 have tiny sd
	scaledDequant := []float32{1000.0, 0.001, 0.001}
	result := adaptiveSmooth(coeff, scaledDequant)

	// Channel 0's large sd drives the shared gap to saturate (gap transform = 0)
	// So all channels should be fully smoothed to their weighted averages
	expectedWeighted := float32(0.05226273532324128 * 10.0)
	for c := 0; c < 3; c++ {
		assert.InDelta(t, float64(expectedWeighted), float64(result[c][1][1]), 0.01,
			"channel %d should be fully smoothed due to ch0's large sd", c)
	}
}

func TestAdaptiveSmooth_4x4(t *testing.T) {
	// 4x4: rows 1-2 are interior, cols 1-2 are interior within those rows
	// Tests that 2 interior rows and 2 interior cols work correctly
	coeff := util.MakeMatrix3D[float32](3, 4, 4)
	for c := 0; c < 3; c++ {
		for y := 0; y < 4; y++ {
			for x := 0; x < 4; x++ {
				coeff[c][y][x] = 1.0
			}
		}
	}
	// Spikes at both interior pixels of row 1
	coeff[0][1][1] = 50.0
	coeff[0][1][2] = 50.0

	scaledDequant := []float32{1.0, 1.0, 1.0}
	result := adaptiveSmooth(coeff, scaledDequant)

	// Edge rows (0 and 3) should be copied exactly
	for x := 0; x < 4; x++ {
		assert.Equal(t, float32(1.0), result[0][0][x])
		assert.Equal(t, float32(1.0), result[0][3][x])
	}
	// Edge columns of interior rows should be copied
	assert.Equal(t, float32(1.0), result[0][1][0])
	assert.Equal(t, float32(1.0), result[0][1][3])
	assert.Equal(t, float32(1.0), result[0][2][0])
	assert.Equal(t, float32(1.0), result[0][2][3])

	// Interior spikes should be smoothed
	assert.Less(t, result[0][1][1], float32(50.0))
	assert.Less(t, result[0][1][2], float32(50.0))
	// Uniform channels should remain approximately 1.0
	assert.InDelta(t, 1.0, float64(result[1][1][1]), 0.01)
	assert.InDelta(t, 1.0, float64(result[1][2][2]), 0.01)
}

// ---- Additional NewLFCoefficientsWithReader tests ----

func TestNewLFCoefficientsWithReader_WithAdaptiveSmoothing(t *testing.T) {
	// No SKIP_ADAPTIVE_LF_SMOOTHING, no USE_LF_FRAME, no subsampling -> adaptive smoothing runs
	thresholds := [][]int32{{}, {}, {}}
	frame := makeLFCoeffFrame(0, []int32{0, 0, 0}, []int32{0, 0, 0}, []float32{1.0, 1.0, 1.0}, thresholds)

	parent := &LFGroup{
		size:      util.Dimension{Width: 3, Height: 3},
		lfGroupID: 0,
	}

	reader := &testcommon.FakeBitReader{ReadBitsData: []uint64{0}}

	// Create a decoded buffer with a spike to verify smoothing occurs
	decodedBuf := util.MakeMatrix3D[int32](3, 3, 3)
	// cMap = {1,0,2}: channel 1 in decoded buf maps to dequantLFCoeff[0]
	// Put a spike that will be visible after dequant
	decodedBuf[0][1][1] = 100 // -> dequant[1] (Y channel)
	decodedBuf[1][1][1] = 0   // -> dequant[0] (X channel)
	decodedBuf[2][1][1] = 0   // -> dequant[2] (B channel)

	fakeStream := &FakeModularStreamer{decodedBuffer: decodedBuf}
	modularFunc := func(reader jxlio.BitReader, frame Framer, streamIndex int, channelCount int, ecStart int, channelArray []ModularChannel) (ModularStreamer, error) {
		return fakeStream, nil
	}

	lf, err := NewLFCoefficientsWithReader(reader, parent, frame, nil, modularFunc)
	require.NoError(t, err)
	require.NotNil(t, lf)

	// After channel correlation (kX=0, kB=0) and adaptive smoothing:
	// The Y channel spike at (1,1) should be smoothed
	// Original dequant[1][1][1] = 100.0, but after smoothing it should be reduced
	assert.Less(t, lf.dequantLFCoeff[1][1][1], float32(100.0),
		"adaptive smoothing should reduce the spike")
	// Edge pixels should be copied from the dequantized values
	assert.Equal(t, float32(0.0), lf.dequantLFCoeff[1][0][0])
}

func TestNewLFCoefficientsWithReader_ChannelCorrelation_MultiPixel(t *testing.T) {
	// Verify channel correlation is applied to all pixels in a multi-pixel grid
	frame := makeLFCoeffFrame(SKIP_ADAPTIVE_LF_SMOOTHING, []int32{0, 0, 0}, []int32{0, 0, 0}, []float32{1.0, 1.0, 1.0}, nil)
	frame.lfGlobal.lfChanCorr = &LFChannelCorrelation{
		colorFactor:      84,
		baseCorrelationX: 0.0,
		baseCorrelationB: 0.0,
		xFactorLF:        212, // kX = (212-128)/84 = 1.0
		bFactorLF:        212, // kB = (212-128)/84 = 1.0
	}

	parent := &LFGroup{
		size:      util.Dimension{Width: 2, Height: 2},
		lfGroupID: 0,
	}
	reader := &testcommon.FakeBitReader{ReadBitsData: []uint64{0}}

	decodedBuf := util.MakeMatrix3D[int32](3, 2, 2)
	// cMap = {1, 0, 2}
	// Y channel: decodedBuf[0] -> dequantLFCoeff[1]
	decodedBuf[0][0][0] = 10
	decodedBuf[0][0][1] = 20
	decodedBuf[0][1][0] = 30
	decodedBuf[0][1][1] = 40
	// X channel: decodedBuf[1] -> dequantLFCoeff[0]
	decodedBuf[1][0][0] = 1
	decodedBuf[1][0][1] = 2
	decodedBuf[1][1][0] = 3
	decodedBuf[1][1][1] = 4
	// B channel: decodedBuf[2] -> dequantLFCoeff[2]
	decodedBuf[2][0][0] = 5
	decodedBuf[2][0][1] = 6
	decodedBuf[2][1][0] = 7
	decodedBuf[2][1][1] = 8

	fakeStream := &FakeModularStreamer{decodedBuffer: decodedBuf}
	modularFunc := func(reader jxlio.BitReader, frame Framer, streamIndex int, channelCount int, ecStart int, channelArray []ModularChannel) (ModularStreamer, error) {
		return fakeStream, nil
	}

	lf, err := NewLFCoefficientsWithReader(reader, parent, frame, nil, modularFunc)
	require.NoError(t, err)

	// kX = 1.0, kB = 1.0
	// For each (y,x): X[y][x] += kX * Y[y][x], B[y][x] += kB * Y[y][x]
	// (0,0): X = 1 + 1.0*10 = 11, B = 5 + 1.0*10 = 15
	assert.InDelta(t, 11.0, float64(lf.dequantLFCoeff[0][0][0]), 0.01)
	assert.InDelta(t, 15.0, float64(lf.dequantLFCoeff[2][0][0]), 0.01)
	// (0,1): X = 2 + 1.0*20 = 22, B = 6 + 1.0*20 = 26
	assert.InDelta(t, 22.0, float64(lf.dequantLFCoeff[0][0][1]), 0.01)
	assert.InDelta(t, 26.0, float64(lf.dequantLFCoeff[2][0][1]), 0.01)
	// (1,0): X = 3 + 1.0*30 = 33, B = 7 + 1.0*30 = 37
	assert.InDelta(t, 33.0, float64(lf.dequantLFCoeff[0][1][0]), 0.01)
	assert.InDelta(t, 37.0, float64(lf.dequantLFCoeff[2][1][0]), 0.01)
	// (1,1): X = 4 + 1.0*40 = 44, B = 8 + 1.0*40 = 48
	assert.InDelta(t, 44.0, float64(lf.dequantLFCoeff[0][1][1]), 0.01)
	assert.InDelta(t, 48.0, float64(lf.dequantLFCoeff[2][1][1]), 0.01)
	// Y channel unchanged
	assert.InDelta(t, 10.0, float64(lf.dequantLFCoeff[1][0][0]), 0.01)
	assert.InDelta(t, 40.0, float64(lf.dequantLFCoeff[1][1][1]), 0.01)
}

func TestNewLFCoefficientsWithReader_DequantMultiplePixels(t *testing.T) {
	// Verify dequantization at multiple positions with varying scaledDequant values
	frame := makeLFCoeffFrame(SKIP_ADAPTIVE_LF_SMOOTHING, []int32{0, 0, 0}, []int32{0, 0, 0}, []float32{2.0, 0.5, 3.0}, nil)

	parent := &LFGroup{
		size:      util.Dimension{Width: 3, Height: 2},
		lfGroupID: 0,
	}
	reader := &testcommon.FakeBitReader{ReadBitsData: []uint64{0}} // extraPrecision=0

	decodedBuf := util.MakeMatrix3D[int32](3, 2, 3)
	// Fill with known values
	// cMap = {1,0,2}
	// i=0: reads from lfQuant[1], writes to dequantLFCoeff[0] with sd=2.0
	// i=1: reads from lfQuant[0], writes to dequantLFCoeff[1] with sd=0.5
	// i=2: reads from lfQuant[2], writes to dequantLFCoeff[2] with sd=3.0
	decodedBuf[0][0][0] = 10
	decodedBuf[0][0][1] = 20
	decodedBuf[0][0][2] = 30
	decodedBuf[0][1][0] = 40
	decodedBuf[1][0][0] = 5
	decodedBuf[1][0][1] = 15
	decodedBuf[1][1][2] = 25
	decodedBuf[2][0][0] = 7
	decodedBuf[2][1][1] = 14

	fakeStream := &FakeModularStreamer{decodedBuffer: decodedBuf}
	modularFunc := func(reader jxlio.BitReader, frame Framer, streamIndex int, channelCount int, ecStart int, channelArray []ModularChannel) (ModularStreamer, error) {
		return fakeStream, nil
	}

	lf, err := NewLFCoefficientsWithReader(reader, parent, frame, nil, modularFunc)
	require.NoError(t, err)

	// dequantLFCoeff[0] = lfQuant[1] * 2.0
	assert.InDelta(t, 10.0, float64(lf.dequantLFCoeff[0][0][0]), 0.001) // 5*2.0
	assert.InDelta(t, 30.0, float64(lf.dequantLFCoeff[0][0][1]), 0.001) // 15*2.0
	assert.InDelta(t, 50.0, float64(lf.dequantLFCoeff[0][1][2]), 0.001) // 25*2.0
	// dequantLFCoeff[1] = lfQuant[0] * 0.5
	assert.InDelta(t, 5.0, float64(lf.dequantLFCoeff[1][0][0]), 0.001)  // 10*0.5
	assert.InDelta(t, 10.0, float64(lf.dequantLFCoeff[1][0][1]), 0.001) // 20*0.5
	assert.InDelta(t, 15.0, float64(lf.dequantLFCoeff[1][0][2]), 0.001) // 30*0.5
	assert.InDelta(t, 20.0, float64(lf.dequantLFCoeff[1][1][0]), 0.001) // 40*0.5
	// dequantLFCoeff[2] = lfQuant[2] * 3.0
	assert.InDelta(t, 21.0, float64(lf.dequantLFCoeff[2][0][0]), 0.001) // 7*3.0
	assert.InDelta(t, 42.0, float64(lf.dequantLFCoeff[2][1][1]), 0.001) // 14*3.0
}

func TestNewLFCoefficientsWithReader_SubsampledXOnly(t *testing.T) {
	// Test with X-only upsampling (jpegUpsamplingX[1]=1, others zero)
	frame := makeLFCoeffFrame(SKIP_ADAPTIVE_LF_SMOOTHING, []int32{0, 0, 0}, []int32{0, 1, 0}, []float32{1.0, 1.0, 1.0}, nil)

	parent := &LFGroup{
		size:      util.Dimension{Width: 4, Height: 2},
		lfGroupID: 0,
	}
	reader := &testcommon.FakeBitReader{ReadBitsData: []uint64{0}}

	// Channel dimensions:
	// i=0 (cMap[0]=1): sizeY=2>>0=2, sizeX=4>>0=4 -> lfQuant[1] needs 2x4
	// i=1 (cMap[1]=0): sizeY=2>>0=2, sizeX=4>>1=2 -> lfQuant[0] needs 2x2
	// i=2 (cMap[2]=2): sizeY=2>>0=2, sizeX=4>>0=4 -> lfQuant[2] needs 2x4
	decodedBuf := make([][][]int32, 3)
	decodedBuf[0] = util.MakeMatrix2D[int32](2, 2) // ch0
	decodedBuf[1] = util.MakeMatrix2D[int32](2, 4) // ch1
	decodedBuf[2] = util.MakeMatrix2D[int32](2, 4) // ch2

	decodedBuf[0][0][0] = 10
	decodedBuf[1][0][0] = 20
	decodedBuf[2][0][0] = 30

	fakeStream := &FakeModularStreamer{decodedBuffer: decodedBuf}
	modularFunc := func(reader jxlio.BitReader, frame Framer, streamIndex int, channelCount int, ecStart int, channelArray []ModularChannel) (ModularStreamer, error) {
		return fakeStream, nil
	}

	lf, err := NewLFCoefficientsWithReader(reader, parent, frame, nil, modularFunc)
	require.NoError(t, err)

	// subSampled=true (jpegUpsamplingX[1]=1), so channel correlation is skipped
	// dequant[0] = lfQuant[1] * 1.0 = 20
	assert.InDelta(t, 20.0, float64(lf.dequantLFCoeff[0][0][0]), 0.001)
	// dequant[1] = lfQuant[0] * 1.0 = 10
	assert.InDelta(t, 10.0, float64(lf.dequantLFCoeff[1][0][0]), 0.001)
	// dequant[2] = lfQuant[2] * 1.0 = 30
	assert.InDelta(t, 30.0, float64(lf.dequantLFCoeff[2][0][0]), 0.001)

	// Verify dequantLFCoeff dimensions: channel i=1 has reduced width
	assert.Len(t, lf.dequantLFCoeff[1][0], 2, "channel 1 should have halved width")
	assert.Len(t, lf.dequantLFCoeff[0][0], 4, "channel 0 should have full width")
}

func TestNewLFCoefficientsWithReader_ModularChannelDimensionsWithUpsampling(t *testing.T) {
	// Verify that the ModularChannel dimensions passed to the modular stream function
	// correctly account for jpeg upsampling
	frame := makeLFCoeffFrame(SKIP_ADAPTIVE_LF_SMOOTHING, []int32{1, 0, 0}, []int32{0, 1, 0}, []float32{1, 1, 1}, nil)

	parent := &LFGroup{
		size:      util.Dimension{Width: 8, Height: 4},
		lfGroupID: 0,
	}
	reader := &testcommon.FakeBitReader{ReadBitsData: []uint64{0}}

	var capturedChannels []ModularChannel
	// Build decoded buffer matching expected dimensions
	// i=0 (cMap[0]=1): sizeY=4>>1=2, sizeX=8>>0=8
	// i=1 (cMap[1]=0): sizeY=4>>0=4, sizeX=8>>1=4
	// i=2 (cMap[2]=2): sizeY=4>>0=4, sizeX=8>>0=8
	decodedBuf := make([][][]int32, 3)
	decodedBuf[0] = util.MakeMatrix2D[int32](4, 4) // ch0 -> dequant[1]
	decodedBuf[1] = util.MakeMatrix2D[int32](2, 8) // ch1 -> dequant[0]
	decodedBuf[2] = util.MakeMatrix2D[int32](4, 8) // ch2 -> dequant[2]

	fakeStream := &FakeModularStreamer{decodedBuffer: decodedBuf}
	modularFunc := func(reader jxlio.BitReader, frame Framer, streamIndex int, channelCount int, ecStart int, channelArray []ModularChannel) (ModularStreamer, error) {
		capturedChannels = channelArray
		return fakeStream, nil
	}

	_, err := NewLFCoefficientsWithReader(reader, parent, frame, nil, modularFunc)
	require.NoError(t, err)

	require.Len(t, capturedChannels, 3)
	// cMap = {1, 0, 2}
	// i=0: info[cMap[0]]=info[1] has height=4>>1=2, width=8>>0=8
	assert.Equal(t, uint32(2), capturedChannels[1].size.Height)
	assert.Equal(t, uint32(8), capturedChannels[1].size.Width)
	assert.Equal(t, int32(1), capturedChannels[1].vshift)
	assert.Equal(t, int32(0), capturedChannels[1].hshift)
	// i=1: info[cMap[1]]=info[0] has height=4>>0=4, width=8>>1=4
	assert.Equal(t, uint32(4), capturedChannels[0].size.Height)
	assert.Equal(t, uint32(4), capturedChannels[0].size.Width)
	assert.Equal(t, int32(0), capturedChannels[0].vshift)
	assert.Equal(t, int32(1), capturedChannels[0].hshift)
	// i=2: info[cMap[2]]=info[2] has height=4>>0=4, width=8>>0=8
	assert.Equal(t, uint32(4), capturedChannels[2].size.Height)
	assert.Equal(t, uint32(8), capturedChannels[2].size.Width)
}

func TestNewLFCoefficientsWithReader_SkipSmoothingFlagOnly(t *testing.T) {
	// With only SKIP_ADAPTIVE_LF_SMOOTHING set (no USE_LF_FRAME),
	// smoothing is disabled and dequantLFCoeff is assigned directly without adaptiveSmooth.
	// A spike pattern that would normally be smoothed should remain unchanged.
	frame := makeLFCoeffFrame(SKIP_ADAPTIVE_LF_SMOOTHING, []int32{0, 0, 0}, []int32{0, 0, 0}, []float32{1.0, 1.0, 1.0}, nil)

	parent := &LFGroup{
		size:      util.Dimension{Width: 3, Height: 3},
		lfGroupID: 0,
	}
	reader := &testcommon.FakeBitReader{ReadBitsData: []uint64{0}}

	// Put a spike at center that adaptive smoothing would reduce
	decodedBuf := util.MakeMatrix3D[int32](3, 3, 3)
	decodedBuf[0][1][1] = 100 // Y channel spike

	fakeStream := &FakeModularStreamer{decodedBuffer: decodedBuf}
	modularFunc := func(reader jxlio.BitReader, frame Framer, streamIndex int, channelCount int, ecStart int, channelArray []ModularChannel) (ModularStreamer, error) {
		return fakeStream, nil
	}

	lf, err := NewLFCoefficientsWithReader(reader, parent, frame, nil, modularFunc)
	require.NoError(t, err)

	// With smoothing skipped, the spike should remain at its dequantized value
	// cMap = {1,0,2}: decodedBuf[0] -> dequantLFCoeff[1]
	assert.InDelta(t, 100.0, float64(lf.dequantLFCoeff[1][1][1]), 0.001,
		"spike should NOT be smoothed when SKIP_ADAPTIVE_LF_SMOOTHING is set")
}

func TestNewLFCoefficientsWithReader_ExtraPrecisionMaxValue(t *testing.T) {
	// extraPrecision is read as 2 bits, max value is 3 -> xx = 1<<3 = 8
	frame := makeLFCoeffFrame(SKIP_ADAPTIVE_LF_SMOOTHING, []int32{0, 0, 0}, []int32{0, 0, 0}, []float32{16.0, 16.0, 16.0}, nil)

	parent := &LFGroup{
		size:      util.Dimension{Width: 1, Height: 1},
		lfGroupID: 0,
	}
	reader := &testcommon.FakeBitReader{ReadBitsData: []uint64{3}} // max extraPrecision = 3

	decodedBuf := util.MakeMatrix3D[int32](3, 1, 1)
	decodedBuf[0][0][0] = 8
	decodedBuf[1][0][0] = 16
	decodedBuf[2][0][0] = 24

	fakeStream := &FakeModularStreamer{decodedBuffer: decodedBuf}
	modularFunc := func(reader jxlio.BitReader, frame Framer, streamIndex int, channelCount int, ecStart int, channelArray []ModularChannel) (ModularStreamer, error) {
		return fakeStream, nil
	}

	lf, err := NewLFCoefficientsWithReader(reader, parent, frame, nil, modularFunc)
	require.NoError(t, err)

	// sd = 16.0 / (1<<3) = 16.0 / 8 = 2.0
	// cMap = {1,0,2}
	// dequant[0] = lfQuant[1] * 2.0 = 16 * 2.0 = 32.0
	assert.InDelta(t, 32.0, float64(lf.dequantLFCoeff[0][0][0]), 0.001)
	// dequant[1] = lfQuant[0] * 2.0 = 8 * 2.0 = 16.0
	assert.InDelta(t, 16.0, float64(lf.dequantLFCoeff[1][0][0]), 0.001)
	// dequant[2] = lfQuant[2] * 2.0 = 24 * 2.0 = 48.0
	assert.InDelta(t, 48.0, float64(lf.dequantLFCoeff[2][0][0]), 0.001)
}
