package frame

import (
	"math"
	"testing"

	"github.com/kpfaulkner/jxl-go/bundle"
	"github.com/kpfaulkner/jxl-go/colour"
	"github.com/kpfaulkner/jxl-go/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- layBlock (package-level) tests ---

func TestLayBlock_CopiesBlockToBuffer(t *testing.T) {
	src := util.MakeMatrix2D[float32](4, 4)
	dst := util.MakeMatrix2D[float32](8, 8)
	for y := 0; y < 4; y++ {
		for x := 0; x < 4; x++ {
			src[y][x] = float32(y*10 + x)
		}
	}

	layBlock(src, dst, util.Point{X: 0, Y: 0}, util.Point{X: 2, Y: 3}, util.Dimension{Height: 4, Width: 4})

	for y := 0; y < 4; y++ {
		for x := 0; x < 4; x++ {
			assert.Equal(t, float32(y*10+x), dst[y+3][x+2], "dst[%d][%d]", y+3, x+2)
		}
	}
	// Check surrounding cells are untouched
	assert.Equal(t, float32(0), dst[0][0])
	assert.Equal(t, float32(0), dst[3][0])
}

func TestLayBlock_WithSourceOffset(t *testing.T) {
	src := util.MakeMatrix2D[float32](8, 8)
	dst := util.MakeMatrix2D[float32](4, 4)
	src[2][3] = 42.0
	src[3][4] = 99.0

	layBlock(src, dst, util.Point{X: 3, Y: 2}, util.Point{X: 0, Y: 0}, util.Dimension{Height: 2, Width: 2})

	assert.Equal(t, float32(42.0), dst[0][0])
	assert.Equal(t, float32(0), dst[0][1])
	assert.Equal(t, float32(99.0), dst[1][1])
}

func TestLayBlock_SingleCell(t *testing.T) {
	src := [][]float32{{7.5}}
	dst := util.MakeMatrix2D[float32](3, 3)

	layBlock(src, dst, util.Point{X: 0, Y: 0}, util.Point{X: 1, Y: 1}, util.Dimension{Height: 1, Width: 1})

	assert.Equal(t, float32(7.5), dst[1][1])
	assert.Equal(t, float32(0), dst[0][0])
}

func TestLayBlock_FullRowCopy(t *testing.T) {
	src := [][]float32{{1, 2, 3, 4}}
	dst := util.MakeMatrix2D[float32](1, 4)

	layBlock(src, dst, util.Point{X: 0, Y: 0}, util.Point{X: 0, Y: 0}, util.Dimension{Height: 1, Width: 4})

	assert.Equal(t, []float32{1, 2, 3, 4}, dst[0])
}

// --- (g *PassGroup) layBlock (method) tests ---

func TestPassGroupLayBlock_SameAsPackageLevel(t *testing.T) {
	pg := &PassGroup{}
	src := util.MakeMatrix2D[float32](2, 2)
	src[0][0] = 1.0
	src[0][1] = 2.0
	src[1][0] = 3.0
	src[1][1] = 4.0
	dst := util.MakeMatrix2D[float32](4, 4)

	pg.layBlock(src, dst, util.Point{X: 0, Y: 0}, util.Point{X: 1, Y: 2}, util.Dimension{Height: 2, Width: 2})

	assert.Equal(t, float32(1.0), dst[2][1])
	assert.Equal(t, float32(2.0), dst[2][2])
	assert.Equal(t, float32(3.0), dst[3][1])
	assert.Equal(t, float32(4.0), dst[3][2])
}

// --- auxDCT2 tests ---

func TestAuxDCT2_Size2_BasicButterfly(t *testing.T) {
	pg := &PassGroup{}
	// 8x8 coefficients, but we only care about the 1x1 region (s/2=1) for the butterfly
	coeffs := util.MakeMatrix2D[float32](8, 8)
	result := util.MakeMatrix2D[float32](8, 8)
	coeffs[0][0] = 10.0
	coeffs[0][1] = 2.0
	coeffs[1][0] = 3.0
	coeffs[1][1] = 1.0

	pg.auxDCT2(coeffs, result, util.Point{X: 0, Y: 0}, util.Point{X: 0, Y: 0}, 2)

	// Butterfly: c00=10, c01=2, c10=3, c11=1
	// r00 = 10+2+3+1 = 16
	// r01 = 10+2-3-1 = 8
	// r10 = 10-2+3-1 = 10
	// r11 = 10-2-3+1 = 6
	assert.Equal(t, float32(16.0), result[0][0])
	assert.Equal(t, float32(8.0), result[0][1])
	assert.Equal(t, float32(10.0), result[1][0])
	assert.Equal(t, float32(6.0), result[1][1])
}

func TestAuxDCT2_Size4_FourButterflies(t *testing.T) {
	pg := &PassGroup{}
	coeffs := util.MakeMatrix2D[float32](8, 8)
	result := util.MakeMatrix2D[float32](8, 8)
	// Set all to 1.0 for a simple test
	for y := 0; y < 4; y++ {
		for x := 0; x < 4; x++ {
			coeffs[y][x] = 1.0
		}
	}

	pg.auxDCT2(coeffs, result, util.Point{X: 0, Y: 0}, util.Point{X: 0, Y: 0}, 4)

	// num = 4/2 = 2, so 2x2 butterflies
	// For each (iy, ix): c00=c01=c10=c11=1
	// r00=4, r01=0, r10=0, r11=0
	for iy := 0; iy < 2; iy++ {
		for ix := 0; ix < 2; ix++ {
			assert.Equal(t, float32(4.0), result[iy*2][ix*2], "result[%d][%d]", iy*2, ix*2)
			assert.Equal(t, float32(0.0), result[iy*2][ix*2+1], "result[%d][%d]", iy*2, ix*2+1)
			assert.Equal(t, float32(0.0), result[iy*2+1][ix*2], "result[%d][%d]", iy*2+1, ix*2)
			assert.Equal(t, float32(0.0), result[iy*2+1][ix*2+1], "result[%d][%d]", iy*2+1, ix*2+1)
		}
	}
}

func TestAuxDCT2_Size2_WithOffset(t *testing.T) {
	pg := &PassGroup{}
	coeffs := util.MakeMatrix2D[float32](16, 16)
	result := util.MakeMatrix2D[float32](16, 16)
	// Place data at offset (4,4)
	coeffs[4][4] = 5.0
	coeffs[4][5] = 3.0
	coeffs[5][4] = 2.0
	coeffs[5][5] = 1.0

	pg.auxDCT2(coeffs, result, util.Point{X: 4, Y: 4}, util.Point{X: 0, Y: 0}, 2)

	// r00 = 5+3+2+1 = 11
	assert.Equal(t, float32(11.0), result[0][0])
	// r01 = 5+3-2-1 = 5
	assert.Equal(t, float32(5.0), result[0][1])
	// r10 = 5-3+2-1 = 3
	assert.Equal(t, float32(3.0), result[1][0])
	// r11 = 5-3-2+1 = 1
	assert.Equal(t, float32(1.0), result[1][1])
}

func TestAuxDCT2_CopiesFullBlock_BeforeButterfly(t *testing.T) {
	pg := &PassGroup{}
	coeffs := util.MakeMatrix2D[float32](8, 8)
	result := util.MakeMatrix2D[float32](8, 8)
	// Set values outside the butterfly region
	coeffs[5][5] = 77.0

	pg.auxDCT2(coeffs, result, util.Point{X: 0, Y: 0}, util.Point{X: 0, Y: 0}, 2)

	// The layBlock call copies the full 8x8 block first, then butterfly overwrites the 2x2 region
	assert.Equal(t, float32(77.0), result[5][5])
}

func TestAuxDCT2_Size8_LargestButterfly(t *testing.T) {
	pg := &PassGroup{}
	coeffs := util.MakeMatrix2D[float32](8, 8)
	result := util.MakeMatrix2D[float32](8, 8)
	// One non-zero value at (0,0)
	coeffs[0][0] = 4.0

	pg.auxDCT2(coeffs, result, util.Point{X: 0, Y: 0}, util.Point{X: 0, Y: 0}, 8)

	// num = 8/2 = 4, butterfly is 4x4
	// Only (iy=0, ix=0) has non-zero input: c00=4, c01=c10=c11=0
	// r00 = 4, r01 = 4, r10 = 4, r11 = 4
	assert.Equal(t, float32(4.0), result[0][0])
	assert.Equal(t, float32(4.0), result[0][1])
	assert.Equal(t, float32(4.0), result[1][0])
	assert.Equal(t, float32(4.0), result[1][1])
}

// --- invertAFV tests ---

func TestInvertAFV_ZeroInput_AFV0(t *testing.T) {
	pg := &PassGroup{}
	coeffs := util.MakeMatrix2D[float32](16, 16)
	buffer := util.MakeMatrix2D[float32](16, 16)
	scratch := util.MakeMatrix3D[float32](5, 256, 256)

	err := pg.invertAFV(coeffs, buffer, AFV0, util.Point{X: 0, Y: 0}, util.Point{X: 0, Y: 0}, scratch)
	require.NoError(t, err)

	// With all-zero input, output should be all zero
	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			assert.Equal(t, float32(0), buffer[y][x], "buffer[%d][%d]", y, x)
		}
	}
}

func TestInvertAFV_AllVariants_ZeroInput(t *testing.T) {
	for _, tc := range []struct {
		name string
		tt   *TransformType
	}{
		{"AFV0", AFV0},
		{"AFV1", AFV1},
		{"AFV2", AFV2},
		{"AFV3", AFV3},
	} {
		t.Run(tc.name, func(t *testing.T) {
			pg := &PassGroup{}
			coeffs := util.MakeMatrix2D[float32](16, 16)
			buffer := util.MakeMatrix2D[float32](16, 16)
			scratch := util.MakeMatrix3D[float32](5, 256, 256)

			err := pg.invertAFV(coeffs, buffer, tc.tt, util.Point{X: 0, Y: 0}, util.Point{X: 0, Y: 0}, scratch)
			require.NoError(t, err)

			// Zero in → zero out
			for y := 0; y < 8; y++ {
				for x := 0; x < 8; x++ {
					assert.Equal(t, float32(0), buffer[y][x], "buffer[%d][%d]", y, x)
				}
			}
		})
	}
}

func TestInvertAFV_SingleDCCoeff_ProducesNonZero(t *testing.T) {
	pg := &PassGroup{}
	coeffs := util.MakeMatrix2D[float32](16, 16)
	buffer := util.MakeMatrix2D[float32](16, 16)
	scratch := util.MakeMatrix3D[float32](5, 256, 256)

	// Set DC coefficient - this should produce non-zero output
	coeffs[0][0] = 1.0

	err := pg.invertAFV(coeffs, buffer, AFV0, util.Point{X: 0, Y: 0}, util.Point{X: 0, Y: 0}, scratch)
	require.NoError(t, err)

	// At least some output cells should be non-zero
	hasNonZero := false
	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			if buffer[y][x] != 0 {
				hasNonZero = true
				break
			}
		}
	}
	assert.True(t, hasNonZero, "expected non-zero output from DC coefficient")
}

func TestInvertAFV_WithOffset(t *testing.T) {
	pg := &PassGroup{}
	coeffs := util.MakeMatrix2D[float32](16, 16)
	buffer := util.MakeMatrix2D[float32](16, 16)
	scratch := util.MakeMatrix3D[float32](5, 256, 256)

	coeffs[2][4] = 1.0

	err := pg.invertAFV(coeffs, buffer, AFV0, util.Point{X: 4, Y: 2}, util.Point{X: 4, Y: 2}, scratch)
	require.NoError(t, err)

	hasNonZero := false
	for y := 2; y < 10; y++ {
		for x := 4; x < 12; x++ {
			if buffer[y][x] != 0 {
				hasNonZero = true
				break
			}
		}
	}
	assert.True(t, hasNonZero, "expected non-zero output in offset region")
}

func TestInvertAFV_FlipSymmetry(t *testing.T) {
	// AFV0 has flipX=0, flipY=0
	// AFV1 has flipX=1, flipY=0
	// AFV2 has flipX=0, flipY=1
	// AFV3 has flipX=1, flipY=1
	// The 4x4 AFV basis subblock should be placed in different quadrants
	pg := &PassGroup{}
	coeffs := util.MakeMatrix2D[float32](16, 16)
	coeffs[0][0] = 1.0
	coeffs[1][0] = 0.5
	// Set AC coefficients so the AFV basis produces a non-uniform 4x4 block.
	// Without this, all basis outputs are identical and flipping has no visible effect.
	coeffs[0][2] = 0.8 // maps to scratchBlock[0][0][1] in the AFV basis
	coeffs[2][0] = 0.3 // maps to scratchBlock[0][1][0] in the AFV basis

	results := make(map[string][][]float32)
	for _, tc := range []struct {
		name string
		tt   *TransformType
	}{
		{"AFV0", AFV0},
		{"AFV1", AFV1},
		{"AFV2", AFV2},
		{"AFV3", AFV3},
	} {
		buffer := util.MakeMatrix2D[float32](16, 16)
		scratch := util.MakeMatrix3D[float32](5, 256, 256)
		err := pg.invertAFV(coeffs, buffer, tc.tt, util.Point{X: 0, Y: 0}, util.Point{X: 0, Y: 0}, scratch)
		require.NoError(t, err)
		results[tc.name] = buffer
	}

	// AFV0 and AFV1 differ by flipX: the 4x4 AFV basis subblock is placed
	// at different x quadrants. Check that the full 8x8 output differs.
	differs01 := false
	for y := 0; y < 8 && !differs01; y++ {
		for x := 0; x < 8; x++ {
			if results["AFV0"][y][x] != results["AFV1"][y][x] {
				differs01 = true
				break
			}
		}
	}
	assert.True(t, differs01, "AFV0 and AFV1 outputs should differ (different flipX)")

	// AFV0 and AFV2 differ by flipY
	differs02 := false
	for y := 0; y < 8 && !differs02; y++ {
		for x := 0; x < 8; x++ {
			if results["AFV0"][y][x] != results["AFV2"][y][x] {
				differs02 = true
				break
			}
		}
	}
	assert.True(t, differs02, "AFV0 and AFV2 outputs should differ (different flipY)")

	// AFV0 and AFV3 differ by both flipX and flipY
	differs03 := false
	for y := 0; y < 8 && !differs03; y++ {
		for x := 0; x < 8; x++ {
			if results["AFV0"][y][x] != results["AFV3"][y][x] {
				differs03 = true
				break
			}
		}
	}
	assert.True(t, differs03, "AFV0 and AFV3 outputs should differ (different flipX and flipY)")
}

// --- Release tests ---

func TestRelease_NilHFCoefficients(t *testing.T) {
	pg := &PassGroup{hfCoefficients: nil}
	// Should not panic
	pg.Release()
	assert.Nil(t, pg.hfCoefficients)
}

func TestRelease_ClearsHFCoefficients(t *testing.T) {
	hf := &HFCoefficients{
		quantizedCoeffs: util.MakeMatrix3D[int32](3, 8, 8),
		dequantHFCoeff:  util.MakeMatrix3D[float32](3, 8, 8),
	}
	pg := &PassGroup{hfCoefficients: hf}

	pg.Release()

	assert.Nil(t, pg.hfCoefficients)
}

func TestRelease_CalledTwice(t *testing.T) {
	hf := &HFCoefficients{
		quantizedCoeffs: util.MakeMatrix3D[int32](3, 4, 4),
		dequantHFCoeff:  util.MakeMatrix3D[float32](3, 4, 4),
	}
	pg := &PassGroup{hfCoefficients: hf}

	pg.Release()
	// Second call should be safe
	pg.Release()
	assert.Nil(t, pg.hfCoefficients)
}

// --- invertVarDCT tests ---

// makeHFCoeffForInvertVarDCT creates an HFCoefficients with all dependencies
// properly configured for bakeDequantizedCoeffs() to succeed.
// DC values should be set via the returned LFGroup's lfCoeff.dequantLFCoeff[c][blockY][blockX].
func makeHFCoeffForInvertVarDCT(tt *TransformType, blocks []*util.Point) (*HFCoefficients, *LFGroup) {
	pH := 16 // big enough for any 8x8 block type
	pW := 16

	dctSelect := util.MakeMatrix2D[*TransformType](8, 8)
	hfMultiplier := util.MakeMatrix2D[int32](8, 8)
	for _, b := range blocks {
		if b != nil {
			dctSelect[b.Y][b.X] = tt
			hfMultiplier[b.Y][b.X] = 1
		}
	}

	hfStreamBuffer := make([][][]int32, 2)
	hfStreamBuffer[0] = util.MakeMatrix2D[int32](1, 1) // xFactorHF
	hfStreamBuffer[1] = util.MakeMatrix2D[int32](1, 1) // bFactorHF

	hfMetadata := &HFMetadata{
		dctSelect:      dctSelect,
		hfMultiplier:   hfMultiplier,
		hfStreamBuffer: hfStreamBuffer,
	}

	lfg := &LFGroup{
		hfMetadata: hfMetadata,
		lfCoeff: &LFCoefficients{
			dequantLFCoeff: util.MakeMatrix3D[float32](3, 8, 8),
			lfIndex:        util.MakeMatrix2D[int32](8, 8),
		},
	}

	// Build weights for all parameter indices up to tt.parameterIndex
	numParams := int(tt.parameterIndex + 1)
	weights := make([][][][]float32, numParams)
	for p := 0; p < numParams; p++ {
		weights[p] = make([][][]float32, 3)
		for c := 0; c < 3; c++ {
			weights[p][c] = util.MakeMatrix2D[float32](pH, pW)
			for y := 0; y < pH; y++ {
				for x := 0; x < pW; x++ {
					weights[p][c][y][x] = 1.0
				}
			}
		}
	}

	ff := &FakeFramer{
		header: &FrameHeader{
			jpegUpsamplingX: []int32{0, 0, 0},
			jpegUpsamplingY: []int32{0, 0, 0},
			xqmScale:        2, // 0.8^(2-2) = 1.0
			bqmScale:        2, // 0.8^(2-2) = 1.0
			passes:          NewPassesInfo(),
			Encoding:        VARDCT,
		},
		lfGlobal: NewLFGlobal(),
		hfGlobal: &HFGlobal{
			numHFPresets: 1,
			weights:      weights,
		},
		imageHeader: &bundle.ImageHeader{
			OpsinInverseMatrix: colour.NewOpsinInverseMatrix(),
		},
	}
	ff.lfGlobal.globalScale = 65536 // 65536/65536 = 1.0
	ff.lfGlobal.lfChanCorr = &LFChannelCorrelation{
		colorFactor:      84,
		baseCorrelationX: 0.0,
		baseCorrelationB: 0.0,
		xFactorLF:        128,
		bFactorLF:        128,
	}

	hfCoeff := &HFCoefficients{
		frame:           ff,
		lfg:             lfg,
		quantizedCoeffs: util.MakeMatrix3D[int32](3, pH, pW),
		dequantHFCoeff:  util.MakeMatrix3D[float32](3, pH, pW),
		groupPos:        util.Point{X: 0, Y: 0},
		blocks:          blocks,
	}

	return hfCoeff, lfg
}

func makeMinimalFrameForInvertVarDCT() *Frame {
	return &Frame{
		Header: &FrameHeader{
			jpegUpsamplingX: []int32{0, 0, 0},
			jpegUpsamplingY: []int32{0, 0, 0},
			Bounds: &util.Rectangle{
				Origin: util.Point{},
				Size:   util.Dimension{Width: 256, Height: 256},
			},
		},
		groupRowStride: 1,
	}
}

func TestInvertVarDCT_EmptyBlocks(t *testing.T) {
	frame := makeMinimalFrameForInvertVarDCT()
	hfCoeff, lfg := makeHFCoeffForInvertVarDCT(DCT8, []*util.Point{})

	pg := &PassGroup{
		frame:          frame,
		hfCoefficients: hfCoeff,
		lfg:            lfg,
		groupID:        0,
	}

	frameBuffer := util.MakeMatrix3D[float32](3, 256, 256)
	err := pg.invertVarDCT(frameBuffer, nil)
	require.NoError(t, err)

	// No blocks → frame buffer should remain zero
	for c := 0; c < 3; c++ {
		for y := 0; y < 8; y++ {
			for x := 0; x < 8; x++ {
				assert.Equal(t, float32(0), frameBuffer[c][y][x])
			}
		}
	}
}

func TestInvertVarDCT_NilBlock_Skipped(t *testing.T) {
	frame := makeMinimalFrameForInvertVarDCT()
	hfCoeff, lfg := makeHFCoeffForInvertVarDCT(DCT8, []*util.Point{nil, nil})

	pg := &PassGroup{
		frame:          frame,
		hfCoefficients: hfCoeff,
		lfg:            lfg,
		groupID:        0,
	}

	frameBuffer := util.MakeMatrix3D[float32](3, 256, 256)
	err := pg.invertVarDCT(frameBuffer, nil)
	require.NoError(t, err)
}

func TestInvertVarDCT_MethodDCT2_SingleBlock(t *testing.T) {
	frame := makeMinimalFrameForInvertVarDCT()
	hfCoeff, lfg := makeHFCoeffForInvertVarDCT(DCT2, []*util.Point{{X: 0, Y: 0}})

	// Set DC value via dequantLFCoeff (finalizeLLF propagates it to dequantHFCoeff)
	lfg.lfCoeff.dequantLFCoeff[0][0][0] = 4.0

	pg := &PassGroup{
		frame:          frame,
		hfCoefficients: hfCoeff,
		lfg:            lfg,
		groupID:        0,
	}

	frameBuffer := util.MakeMatrix3D[float32](3, 256, 256)
	err := pg.invertVarDCT(frameBuffer, nil)
	require.NoError(t, err)

	// DCT2 with DC should propagate to all 8x8 pixels
	hasNonZero := false
	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			if frameBuffer[0][y][x] != 0 {
				hasNonZero = true
			}
		}
	}
	assert.True(t, hasNonZero, "DCT2 should produce non-zero output from DC coeff")
}

func TestInvertVarDCT_MethodDCT2_AllPixelsEqual(t *testing.T) {
	// DC-only input through DCT2 should produce uniform output
	frame := makeMinimalFrameForInvertVarDCT()
	hfCoeff, lfg := makeHFCoeffForInvertVarDCT(DCT2, []*util.Point{{X: 0, Y: 0}})

	lfg.lfCoeff.dequantLFCoeff[0][0][0] = 1.0

	pg := &PassGroup{
		frame:          frame,
		hfCoefficients: hfCoeff,
		lfg:            lfg,
	}

	frameBuffer := util.MakeMatrix3D[float32](3, 256, 256)
	err := pg.invertVarDCT(frameBuffer, nil)
	require.NoError(t, err)

	// All 64 pixels in the 8x8 block should be the same value (DC propagation)
	firstVal := frameBuffer[0][0][0]
	assert.NotEqual(t, float32(0), firstVal)
	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			assert.InDelta(t, float64(firstVal), float64(frameBuffer[0][y][x]), 1e-5,
				"all pixels should be equal for DC-only input, [%d][%d]", y, x)
		}
	}
}

func TestInvertVarDCT_MethodDCT_SingleBlock(t *testing.T) {
	frame := makeMinimalFrameForInvertVarDCT()
	hfCoeff, lfg := makeHFCoeffForInvertVarDCT(DCT8, []*util.Point{{X: 0, Y: 0}})

	lfg.lfCoeff.dequantLFCoeff[0][0][0] = 2.0

	pg := &PassGroup{
		frame:          frame,
		hfCoefficients: hfCoeff,
		lfg:            lfg,
	}

	frameBuffer := util.MakeMatrix3D[float32](3, 256, 256)
	err := pg.invertVarDCT(frameBuffer, nil)
	require.NoError(t, err)

	// DC-only DCT should produce uniform output across 8x8
	firstVal := frameBuffer[0][0][0]
	assert.NotEqual(t, float32(0), firstVal)
	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			assert.InDelta(t, float64(firstVal), float64(frameBuffer[0][y][x]), 1e-5,
				"DC-only DCT [%d][%d]", y, x)
		}
	}
}

func TestInvertVarDCT_MethodHornuss_SingleBlock(t *testing.T) {
	frame := makeMinimalFrameForInvertVarDCT()
	hfCoeff, lfg := makeHFCoeffForInvertVarDCT(HORNUSS, []*util.Point{{X: 0, Y: 0}})

	lfg.lfCoeff.dequantLFCoeff[0][0][0] = 3.0

	pg := &PassGroup{
		frame:          frame,
		hfCoefficients: hfCoeff,
		lfg:            lfg,
	}

	frameBuffer := util.MakeMatrix3D[float32](3, 256, 256)
	err := pg.invertVarDCT(frameBuffer, nil)
	require.NoError(t, err)

	hasNonZero := false
	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			if frameBuffer[0][y][x] != 0 {
				hasNonZero = true
			}
		}
	}
	assert.True(t, hasNonZero, "HORNUSS should produce non-zero from DC coeff")
}

func TestInvertVarDCT_MethodAFV_SingleBlock(t *testing.T) {
	frame := makeMinimalFrameForInvertVarDCT()
	hfCoeff, lfg := makeHFCoeffForInvertVarDCT(AFV0, []*util.Point{{X: 0, Y: 0}})

	lfg.lfCoeff.dequantLFCoeff[0][0][0] = 1.0

	pg := &PassGroup{
		frame:          frame,
		hfCoefficients: hfCoeff,
		lfg:            lfg,
	}

	frameBuffer := util.MakeMatrix3D[float32](3, 256, 256)
	err := pg.invertVarDCT(frameBuffer, nil)
	require.NoError(t, err)

	hasNonZero := false
	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			if frameBuffer[0][y][x] != 0 {
				hasNonZero = true
			}
		}
	}
	assert.True(t, hasNonZero, "AFV should produce non-zero from DC coeff")
}

func TestInvertVarDCT_MethodDCT8_4_SingleBlock(t *testing.T) {
	frame := makeMinimalFrameForInvertVarDCT()
	hfCoeff, lfg := makeHFCoeffForInvertVarDCT(DCT8_4, []*util.Point{{X: 0, Y: 0}})

	lfg.lfCoeff.dequantLFCoeff[0][0][0] = 2.0

	pg := &PassGroup{
		frame:          frame,
		hfCoefficients: hfCoeff,
		lfg:            lfg,
	}

	frameBuffer := util.MakeMatrix3D[float32](3, 256, 256)
	err := pg.invertVarDCT(frameBuffer, nil)
	require.NoError(t, err)

	hasNonZero := false
	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			if frameBuffer[0][y][x] != 0 {
				hasNonZero = true
			}
		}
	}
	assert.True(t, hasNonZero, "DCT8_4 should produce non-zero from DC coeff")
}

func TestInvertVarDCT_MethodDCT4_8_SingleBlock(t *testing.T) {
	frame := makeMinimalFrameForInvertVarDCT()
	hfCoeff, lfg := makeHFCoeffForInvertVarDCT(DCT4_8, []*util.Point{{X: 0, Y: 0}})

	lfg.lfCoeff.dequantLFCoeff[0][0][0] = 2.0

	pg := &PassGroup{
		frame:          frame,
		hfCoefficients: hfCoeff,
		lfg:            lfg,
	}

	frameBuffer := util.MakeMatrix3D[float32](3, 256, 256)
	err := pg.invertVarDCT(frameBuffer, nil)
	require.NoError(t, err)

	hasNonZero := false
	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			if frameBuffer[0][y][x] != 0 {
				hasNonZero = true
			}
		}
	}
	assert.True(t, hasNonZero, "DCT4_8 should produce non-zero from DC coeff")
}

func TestInvertVarDCT_UpsampledChannel_Skipped(t *testing.T) {
	frame := makeMinimalFrameForInvertVarDCT()
	// jpegUpsamplingY[1] = 1 means channel 1 is upsampled by 2x
	frame.Header.jpegUpsamplingY = []int32{0, 1, 0}

	hfCoeff, lfg := makeHFCoeffForInvertVarDCT(DCT8, []*util.Point{{X: 0, Y: 0}})
	// Match frame header upsampling on the FakeFramer used by hfCoeff
	hfCoeff.frame.getFrameHeader().jpegUpsamplingY = []int32{0, 1, 0}

	lfg.lfCoeff.dequantLFCoeff[1][0][0] = 10.0

	pg := &PassGroup{
		frame:          frame,
		hfCoefficients: hfCoeff,
		lfg:            lfg,
	}

	frameBuffer := util.MakeMatrix3D[float32](3, 256, 256)
	err := pg.invertVarDCT(frameBuffer, nil)
	require.NoError(t, err)

	// The key test is it doesn't panic with upsampled channels.
}

func TestInvertVarDCT_MultipleBlocks(t *testing.T) {
	frame := makeMinimalFrameForInvertVarDCT()
	hfCoeff, lfg := makeHFCoeffForInvertVarDCT(DCT8, []*util.Point{{X: 0, Y: 0}, {X: 1, Y: 0}})

	// Set DC values for both blocks via dequantLFCoeff
	lfg.lfCoeff.dequantLFCoeff[0][0][0] = 1.0
	lfg.lfCoeff.dequantLFCoeff[0][0][1] = 2.0

	pg := &PassGroup{
		frame:          frame,
		hfCoefficients: hfCoeff,
		lfg:            lfg,
	}

	frameBuffer := util.MakeMatrix3D[float32](3, 256, 256)
	err := pg.invertVarDCT(frameBuffer, nil)
	require.NoError(t, err)

	// Both blocks should produce non-zero output
	block0NonZero := false
	block1NonZero := false
	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			if frameBuffer[0][y][x] != 0 {
				block0NonZero = true
			}
			if frameBuffer[0][y][x+8] != 0 {
				block1NonZero = true
			}
		}
	}
	assert.True(t, block0NonZero, "block 0 should have non-zero output")
	assert.True(t, block1NonZero, "block 1 should have non-zero output")
}

func TestInvertVarDCT_DCT_InverseDCT_Energy(t *testing.T) {
	frame := makeMinimalFrameForInvertVarDCT()
	hfCoeff, lfg := makeHFCoeffForInvertVarDCT(DCT8, []*util.Point{{X: 0, Y: 0}})

	lfg.lfCoeff.dequantLFCoeff[0][0][0] = 8.0

	pg := &PassGroup{
		frame:          frame,
		hfCoefficients: hfCoeff,
		lfg:            lfg,
	}

	frameBuffer := util.MakeMatrix3D[float32](3, 256, 256)
	err := pg.invertVarDCT(frameBuffer, nil)
	require.NoError(t, err)

	// For DC-only, all pixels should be equal
	first := frameBuffer[0][0][0]
	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			assert.InDelta(t, float64(first), float64(frameBuffer[0][y][x]), 1e-4,
				"DC-only should give uniform output at [%d][%d]", y, x)
		}
	}
	assert.True(t, math.Abs(float64(first)) > 0.001, "DC output should be non-zero")
}

func TestInvertVarDCT_AllChannelsProcessed(t *testing.T) {
	frame := makeMinimalFrameForInvertVarDCT()
	hfCoeff, lfg := makeHFCoeffForInvertVarDCT(DCT8, []*util.Point{{X: 0, Y: 0}})

	lfg.lfCoeff.dequantLFCoeff[0][0][0] = 1.0
	lfg.lfCoeff.dequantLFCoeff[1][0][0] = 2.0
	lfg.lfCoeff.dequantLFCoeff[2][0][0] = 3.0

	pg := &PassGroup{
		frame:          frame,
		hfCoefficients: hfCoeff,
		lfg:            lfg,
	}

	frameBuffer := util.MakeMatrix3D[float32](3, 256, 256)
	err := pg.invertVarDCT(frameBuffer, nil)
	require.NoError(t, err)

	// All 3 channels should have non-zero output
	for c := 0; c < 3; c++ {
		hasNonZero := false
		for y := 0; y < 8 && !hasNonZero; y++ {
			for x := 0; x < 8; x++ {
				if frameBuffer[c][y][x] != 0 {
					hasNonZero = true
					break
				}
			}
		}
		assert.True(t, hasNonZero, "channel %d should have non-zero output", c)
	}
}
