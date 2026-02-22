package frame

import (
	"math"
	"testing"

	"github.com/kpfaulkner/jxl-go/testcommon"
	"github.com/kpfaulkner/jxl-go/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- quantMult tests ---

func TestQuantMult(t *testing.T) {
	tests := []struct {
		name     string
		v        float32
		expected float32
	}{
		{"zero returns 1", 0.0, 1.0},
		{"positive 1 returns 2", 1.0, 2.0},
		{"positive 2 returns 3", 2.0, 3.0},
		{"positive 0.5 returns 1.5", 0.5, 1.5},
		{"negative -1 returns 0.5", -1.0, 0.5},           // 1/(1+1) = 0.5
		{"negative -0.5 returns 2/3", -0.5, 1.0 / 1.5},   // 1/(1+0.5)
		{"negative -3 returns 0.25", -3.0, 0.25},          // 1/(1+3) = 0.25
		{"small positive", 0.001, 1.001},
		{"boundary at zero positive side", 0.0, 1.0},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := quantMult(tc.v)
			assert.InDelta(t, tc.expected, float64(result), 1e-5)
		})
	}
}

// --- interpolate tests ---

func TestInterpolate_SingleBand(t *testing.T) {
	// With single band, always returns that band value
	result := interpolate(0, []float32{42.0})
	assert.Equal(t, float32(42.0), result)

	result = interpolate(999, []float32{42.0})
	assert.Equal(t, float32(42.0), result)
}

func TestInterpolate_IntegerPosition(t *testing.T) {
	bands := []float32{10.0, 20.0, 30.0}
	// At position 0: a=10, b=20, pow(20/10, 0)=1 → 10
	assert.InDelta(t, 10.0, float64(interpolate(0.0, bands)), 1e-5)
	// At position 1: a=20, b=30, pow(30/20, 0)=1 → 20
	assert.InDelta(t, 20.0, float64(interpolate(1.0, bands)), 1e-5)
}

func TestInterpolate_BeyondRange(t *testing.T) {
	bands := []float32{10.0, 20.0}
	// scaledIndex=5, 5+1=6 > 1 → return bands[1]=20
	result := interpolate(5.0, bands)
	assert.Equal(t, float32(20.0), result)
}

func TestInterpolate_FractionalPosition(t *testing.T) {
	bands := []float32{100.0, 200.0}
	// a=100, b=200, fracIndex=0.5
	// 100 * pow(200/100, 0.5) = 100 * sqrt(2) ≈ 141.42
	result := interpolate(0.5, bands)
	expected := 100.0 * math.Sqrt(2)
	assert.InDelta(t, expected, float64(result), 0.1)
}

func TestInterpolate_Midpoint(t *testing.T) {
	bands := []float32{100.0, 100.0, 100.0}
	// All bands equal → any position returns 100
	assert.InDelta(t, 100.0, float64(interpolate(0.5, bands)), 1e-5)
	assert.InDelta(t, 100.0, float64(interpolate(1.5, bands)), 1e-5)
}

// --- getDCTQuantWeights tests ---

func TestGetDCTQuantWeights_OutputSize(t *testing.T) {
	hfg := &HFGlobal{}
	weights := hfg.getDCTQuantWeights(8, 16, []float64{1000, 0})
	assert.Equal(t, 8, len(weights))
	assert.Equal(t, 16, len(weights[0]))
}

func TestGetDCTQuantWeights_OriginValue(t *testing.T) {
	hfg := &HFGlobal{}
	// At (0,0): dist=0, interpolate(0, bands) = bands[0] = params[0]
	weights := hfg.getDCTQuantWeights(8, 8, []float64{3150.0, 0.0, -0.4})
	assert.InDelta(t, 3150.0, float64(weights[0][0]), 0.01)
}

func TestGetDCTQuantWeights_ConstantBands(t *testing.T) {
	hfg := &HFGlobal{}
	// params[1]=0 → quantMult(0)=1 → bands all equal params[0]
	weights := hfg.getDCTQuantWeights(4, 4, []float64{500.0, 0.0})
	// All weights should be 500 since bands = [500, 500]
	for y := 0; y < 4; y++ {
		for x := 0; x < 4; x++ {
			assert.InDelta(t, 500.0, float64(weights[y][x]), 0.1,
				"weight at (%d,%d)", y, x)
		}
	}
}

func TestGetDCTQuantWeights_Symmetry(t *testing.T) {
	hfg := &HFGlobal{}
	// For square matrices, weights[y][x] == weights[x][y]
	weights := hfg.getDCTQuantWeights(8, 8, []float64{1000.0, -0.5, -0.3})
	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			assert.InDelta(t, float64(weights[y][x]), float64(weights[x][y]), 1e-4,
				"weight[%d][%d] should equal weight[%d][%d]", y, x, x, y)
		}
	}
}

func TestGetDCTQuantWeights_DecreasingBands(t *testing.T) {
	hfg := &HFGlobal{}
	// Negative params → quantMult returns <1 → bands decrease
	weights := hfg.getDCTQuantWeights(8, 8, []float64{1000.0, -1.0, -1.0})
	// Origin should be highest
	assert.Greater(t, weights[0][0], weights[7][7])
}

// --- generateWeights tests ---

func makeHFGlobalWithDefaults() *HFGlobal {
	hfg := &HFGlobal{
		params:  defaultParams,
		weights: util.MakeMatrix4D[float32](17, 3, 0, 0),
	}
	return hfg
}

func TestGenerateWeights_ModeDCT_Size(t *testing.T) {
	hfg := makeHFGlobalWithDefaults()
	err := hfg.generateWeights(0) // index=0: DCT8, matrixHeight=8, matrixWidth=8
	require.NoError(t, err)
	for c := 0; c < 3; c++ {
		assert.Equal(t, 8, len(hfg.weights[0][c]), "channel %d height", c)
		assert.Equal(t, 8, len(hfg.weights[0][c][0]), "channel %d width", c)
	}
}

func TestGenerateWeights_ModeDCT_LargerSize(t *testing.T) {
	hfg := makeHFGlobalWithDefaults()
	err := hfg.generateWeights(4) // index=4: DCT16, matrixHeight=16, matrixWidth=16
	require.NoError(t, err)
	for c := 0; c < 3; c++ {
		assert.Equal(t, 16, len(hfg.weights[4][c]), "channel %d height", c)
		assert.Equal(t, 16, len(hfg.weights[4][c][0]), "channel %d width", c)
	}
}

func TestGenerateWeights_ModeHornuss(t *testing.T) {
	hfg := makeHFGlobalWithDefaults()
	err := hfg.generateWeights(1) // index=1: HORNUSS
	require.NoError(t, err)

	// After inversion, w[0][0] = 1/1.0 = 1.0
	assert.InDelta(t, 1.0, float64(hfg.weights[1][0][0][0]), 1e-6)
	// w[0][1] was param[0][1]=3160 → inverted: 1/3160
	assert.InDelta(t, 1.0/3160.0, float64(hfg.weights[1][0][0][1]), 1e-8)
	// w[1][0] was param[0][1]=3160 → inverted
	assert.InDelta(t, 1.0/3160.0, float64(hfg.weights[1][0][1][0]), 1e-8)
	// w[1][1] was param[0][2]=3160 → inverted
	assert.InDelta(t, 1.0/3160.0, float64(hfg.weights[1][0][1][1]), 1e-8)
	// All other cells were param[0][0]=280 → inverted: 1/280
	assert.InDelta(t, 1.0/280.0, float64(hfg.weights[1][0][3][3]), 1e-8)
	assert.InDelta(t, 1.0/280.0, float64(hfg.weights[1][0][7][7]), 1e-8)
}

func TestGenerateWeights_ModeDCT2(t *testing.T) {
	hfg := makeHFGlobalWithDefaults()
	err := hfg.generateWeights(2) // index=2: DCT2
	require.NoError(t, err)

	// c=0: param[0] = {3840, 2560, 1280, 640, 480, 300}
	// After inversion: w[0][0] = 1/1 = 1.0
	assert.InDelta(t, 1.0, float64(hfg.weights[2][0][0][0]), 1e-6)
	// w[0][1] = 1/3840
	assert.InDelta(t, 1.0/3840.0, float64(hfg.weights[2][0][0][1]), 1e-8)
	// w[1][1] = 1/2560
	assert.InDelta(t, 1.0/2560.0, float64(hfg.weights[2][0][1][1]), 1e-8)
	// w[0][2] (from param[2]=1280) → 1/1280
	assert.InDelta(t, 1.0/1280.0, float64(hfg.weights[2][0][0][2]), 1e-8)
	// w[4][4] (from param[5]=300) → 1/300
	assert.InDelta(t, 1.0/300.0, float64(hfg.weights[2][0][4][4]), 1e-8)
}

func TestGenerateWeights_ModeDCT4(t *testing.T) {
	hfg := makeHFGlobalWithDefaults()
	err := hfg.generateWeights(3) // index=3: DCT4
	require.NoError(t, err)

	// DCT4 produces 8x8 weights
	for c := 0; c < 3; c++ {
		assert.Equal(t, 8, len(hfg.weights[3][c]), "channel %d height", c)
		assert.Equal(t, 8, len(hfg.weights[3][c][0]), "channel %d width", c)
	}

	// All weights should be positive after inversion
	for c := 0; c < 3; c++ {
		for y := 0; y < 8; y++ {
			for x := 0; x < 8; x++ {
				assert.Greater(t, hfg.weights[3][c][y][x], float32(0),
					"c=%d y=%d x=%d should be positive", c, y, x)
			}
		}
	}
}

func TestGenerateWeights_ModeDCT4_8(t *testing.T) {
	hfg := makeHFGlobalWithDefaults()
	err := hfg.generateWeights(9) // index=9: DCT4_8
	require.NoError(t, err)

	// DCT4_8 produces 8x8 weights
	for c := 0; c < 3; c++ {
		assert.Equal(t, 8, len(hfg.weights[9][c]))
		assert.Equal(t, 8, len(hfg.weights[9][c][0]))
	}
}

func TestGenerateWeights_ModeAFV(t *testing.T) {
	hfg := makeHFGlobalWithDefaults()
	err := hfg.generateWeights(10) // index=10: AFV
	require.NoError(t, err)

	// AFV produces 8x8 weights
	for c := 0; c < 3; c++ {
		assert.Equal(t, 8, len(hfg.weights[10][c]))
		assert.Equal(t, 8, len(hfg.weights[10][c][0]))
	}

	// After inversion, w[0][0] = 1/1 = 1.0
	assert.InDelta(t, 1.0, float64(hfg.weights[10][0][0][0]), 1e-6)
}

func TestGenerateWeights_AllDefaultIndices(t *testing.T) {
	hfg := makeHFGlobalWithDefaults()
	for i := 0; i < 17; i++ {
		err := hfg.generateWeights(i)
		require.NoError(t, err, "generateWeights failed for index %d", i)
	}
}

func TestGenerateWeights_WeightsArePositive(t *testing.T) {
	hfg := makeHFGlobalWithDefaults()
	for i := 0; i < 17; i++ {
		err := hfg.generateWeights(i)
		require.NoError(t, err)
	}
	for i := 0; i < 17; i++ {
		for c := 0; c < 3; c++ {
			for y := 0; y < len(hfg.weights[i][c]); y++ {
				for x := 0; x < len(hfg.weights[i][c][y]); x++ {
					assert.Greater(t, hfg.weights[i][c][y][x], float32(0),
						"index=%d c=%d y=%d x=%d", i, c, y, x)
				}
			}
		}
	}
}

func TestGenerateWeights_NegativeWeightError(t *testing.T) {
	hfg := &HFGlobal{
		params:  make([]DCTParam, 17),
		weights: util.MakeMatrix4D[float32](17, 3, 0, 0),
	}
	// Set up HORNUSS with negative param → produces negative weight → error on inversion check
	hfg.params[1] = DCTParam{
		param: [][]float32{
			{-1.0, 1.0, 1.0},
			{1.0, 1.0, 1.0},
			{1.0, 1.0, 1.0},
		},
		mode: MODE_HORNUSS,
	}
	err := hfg.generateWeights(1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Invalid weight")
}

// --- getAFVTransformWeights tests ---

func TestGetAFVTransformWeights_OutputSize(t *testing.T) {
	hfg := makeHFGlobalWithDefaults()
	weights, err := hfg.getAFVTransformWeights(10, 0)
	require.NoError(t, err)
	assert.Equal(t, 8, len(weights))
	assert.Equal(t, 8, len(weights[0]))
}

func TestGetAFVTransformWeights_CornerValues(t *testing.T) {
	hfg := makeHFGlobalWithDefaults()
	// c=0, params[10].param[0] = [3072, 3072, 256, 256, 256, 414, 0, 0, 0]
	weights, err := hfg.getAFVTransformWeights(10, 0)
	require.NoError(t, err)

	// weight[0][0] = 1 (explicit)
	assert.InDelta(t, 1.0, float64(weights[0][0]), 1e-6)
	// weight[1][0] = param[c][0] = 3072
	assert.InDelta(t, 3072.0, float64(weights[1][0]), 1e-4)
	// weight[0][1] = param[c][1] = 3072
	assert.InDelta(t, 3072.0, float64(weights[0][1]), 1e-4)
	// weight[2][0] = param[c][2] = 256
	assert.InDelta(t, 256.0, float64(weights[2][0]), 1e-4)
	// weight[0][2] = param[c][3] = 256
	assert.InDelta(t, 256.0, float64(weights[0][2]), 1e-4)
	// weight[2][2] = param[c][4] = 256
	assert.InDelta(t, 256.0, float64(weights[2][2]), 1e-4)
}

func TestGetAFVTransformWeights_NegativeBandError(t *testing.T) {
	hfg := &HFGlobal{
		params:  make([]DCTParam, 17),
		weights: util.MakeMatrix4D[float32](17, 3, 0, 0),
	}
	hfg.params[10] = defaultParams[10]
	// Make param[c][5] negative → bands[0] < 0 → error
	hfg.params[10].param = [][]float32{
		{3072, 3072, 256, 256, 256, -1, 0, 0, 0},
		{1024, 1024, 50, 50, 50, 58, 0, 0, 0},
		{384, 384, 12, 12, 12, 22, -0.25, -0.25, -0.25},
	}
	_, err := hfg.getAFVTransformWeights(10, 0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Invalid band")
}

// --- readDCTParams tests ---

func TestReadDCTParams_SingleParam(t *testing.T) {
	hfg := &HFGlobal{}
	reader := &testcommon.FakeBitReader{
		ReadBitsData: []uint64{0}, // numParams=0+1=1
		ReadF16Data:  []float32{10.0, 20.0, 30.0},
	}

	vals, err := hfg.readDCTParams(reader)
	require.NoError(t, err)
	assert.Equal(t, 3, len(vals))
	assert.Equal(t, 1, len(vals[0]))
	// First (only) value of each channel ×64
	assert.InDelta(t, 640.0, vals[0][0], 1e-3)
	assert.InDelta(t, 1280.0, vals[1][0], 1e-3)
	assert.InDelta(t, 1920.0, vals[2][0], 1e-3)
}

func TestReadDCTParams_MultipleParams(t *testing.T) {
	hfg := &HFGlobal{}
	reader := &testcommon.FakeBitReader{
		ReadBitsData: []uint64{2}, // numParams=2+1=3
		ReadF16Data:  []float32{1, 2, 3, 4, 5, 6, 7, 8, 9},
	}

	vals, err := hfg.readDCTParams(reader)
	require.NoError(t, err)
	assert.Equal(t, 3, len(vals))
	assert.Equal(t, 3, len(vals[0]))

	// c=0: [1*64, 2, 3]
	assert.InDelta(t, 64.0, vals[0][0], 1e-3)
	assert.InDelta(t, 2.0, vals[0][1], 1e-3)
	assert.InDelta(t, 3.0, vals[0][2], 1e-3)

	// c=1: [4*64, 5, 6]
	assert.InDelta(t, 256.0, vals[1][0], 1e-3)
	assert.InDelta(t, 5.0, vals[1][1], 1e-3)

	// c=2: [7*64, 8, 9]
	assert.InDelta(t, 448.0, vals[2][0], 1e-3)
	assert.InDelta(t, 9.0, vals[2][2], 1e-3)
}

func TestReadDCTParams_ReadBitsError(t *testing.T) {
	hfg := &HFGlobal{}
	reader := &testcommon.FakeBitReader{} // empty → error on ReadBits(4)

	_, err := hfg.readDCTParams(reader)
	assert.Error(t, err)
}

func TestReadDCTParams_ReadF16Error(t *testing.T) {
	hfg := &HFGlobal{}
	reader := &testcommon.FakeBitReader{
		ReadBitsData: []uint64{0}, // numParams=1
		// No F16 data → error on first ReadF16
	}

	_, err := hfg.readDCTParams(reader)
	assert.Error(t, err)
}

// --- setupDCTParam tests ---

func TestSetupDCTParam_ModeLibrary(t *testing.T) {
	hfg := &HFGlobal{
		params: make([]DCTParam, 17),
	}
	reader := &testcommon.FakeBitReader{
		ReadBitsData: []uint64{MODE_LIBRARY}, // mode=0
	}

	err := hfg.setupDCTParam(reader, nil, 5)
	require.NoError(t, err)

	// Should copy defaultParams[5]
	assert.True(t, hfg.params[5].Equals(defaultParams[5]))
}

func TestSetupDCTParam_ModeHornuss(t *testing.T) {
	hfg := &HFGlobal{
		params: make([]DCTParam, 17),
	}
	reader := &testcommon.FakeBitReader{
		ReadBitsData: []uint64{MODE_HORNUSS},
		ReadF16Data:  []float32{1, 2, 3, 4, 5, 6, 7, 8, 9},
	}

	err := hfg.setupDCTParam(reader, nil, 1) // index=1 valid for HORNUSS
	require.NoError(t, err)

	assert.Equal(t, int32(MODE_HORNUSS), hfg.params[1].mode)
	assert.Equal(t, float32(1), hfg.params[1].denominator)
	// param[0] = [1*64, 2*64, 3*64]
	assert.InDelta(t, 64.0, float64(hfg.params[1].param[0][0]), 1e-3)
	assert.InDelta(t, 128.0, float64(hfg.params[1].param[0][1]), 1e-3)
	assert.InDelta(t, 192.0, float64(hfg.params[1].param[0][2]), 1e-3)
	// param[2] = [7*64, 8*64, 9*64]
	assert.InDelta(t, 448.0, float64(hfg.params[1].param[2][0]), 1e-3)
}

func TestSetupDCTParam_ModeDCT(t *testing.T) {
	hfg := &HFGlobal{
		params: make([]DCTParam, 17),
	}
	reader := &testcommon.FakeBitReader{
		ReadBitsData: []uint64{MODE_DCT, 0}, // mode=6, then numParams=0+1=1
		ReadF16Data:  []float32{10, 20, 30},
	}

	err := hfg.setupDCTParam(reader, nil, 0)
	require.NoError(t, err)

	assert.Equal(t, int32(MODE_DCT), hfg.params[0].mode)
	assert.Nil(t, hfg.params[0].param)
	// dctParam[0][0] = 10*64 = 640
	assert.InDelta(t, 640.0, hfg.params[0].dctParam[0][0], 1e-3)
	assert.InDelta(t, 1280.0, hfg.params[0].dctParam[1][0], 1e-3)
	assert.InDelta(t, 1920.0, hfg.params[0].dctParam[2][0], 1e-3)
}

func TestSetupDCTParam_ModeDCT2(t *testing.T) {
	hfg := &HFGlobal{
		params: make([]DCTParam, 17),
	}
	// MODE_DCT2 needs 3×6=18 F16 values
	f16Data := make([]float32, 18)
	for i := range f16Data {
		f16Data[i] = float32(i + 1)
	}
	reader := &testcommon.FakeBitReader{
		ReadBitsData: []uint64{MODE_DCT2},
		ReadF16Data:  f16Data,
	}

	err := hfg.setupDCTParam(reader, nil, 2) // index=2 valid for DCT2
	require.NoError(t, err)

	assert.Equal(t, int32(MODE_DCT2), hfg.params[2].mode)
	assert.Equal(t, 3, len(hfg.params[2].param))
	assert.Equal(t, 6, len(hfg.params[2].param[0]))
	// Values are *64
	assert.InDelta(t, 64.0, float64(hfg.params[2].param[0][0]), 1e-3)
}

func TestSetupDCTParam_InvalidIndexForMode(t *testing.T) {
	hfg := &HFGlobal{
		params: make([]DCTParam, 17),
	}
	reader := &testcommon.FakeBitReader{
		ReadBitsData: []uint64{MODE_HORNUSS}, // HORNUSS only valid for index 0-3, 9, 10
	}

	err := hfg.setupDCTParam(reader, nil, 5) // index=5 → invalid for HORNUSS
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Invalid index")
}

func TestSetupDCTParam_ReadBitsError(t *testing.T) {
	hfg := &HFGlobal{
		params: make([]DCTParam, 17),
	}
	reader := &testcommon.FakeBitReader{} // empty → error

	err := hfg.setupDCTParam(reader, nil, 0)
	assert.Error(t, err)
}

// --- NewHFGlobalWithReader tests ---

func TestNewHFGlobalWithReader_DefaultParams(t *testing.T) {
	reader := &testcommon.FakeBitReader{
		ReadBoolData: []bool{true},   // quantAllDefault=true
		ReadBitsData: []uint64{0},    // numPresets (CeilLog1p(0)=0 bits → ReadBits(0))
	}
	frame := &Frame{numGroups: 1}

	hfg, err := NewHFGlobalWithReader(reader, frame)
	require.NoError(t, err)
	require.NotNil(t, hfg)

	// Params should be defaultParams
	for i := 0; i < 17; i++ {
		assert.True(t, hfg.params[i].Equals(defaultParams[i]),
			"params[%d] should match default", i)
	}
	assert.Equal(t, int32(1), hfg.numHFPresets) // 1 + 0 = 1
}

func TestNewHFGlobalWithReader_NumPresets(t *testing.T) {
	reader := &testcommon.FakeBitReader{
		ReadBoolData: []bool{true},
		ReadBitsData: []uint64{5}, // numPresets value
	}
	frame := &Frame{numGroups: 8} // CeilLog1p(7)=3 → ReadBits(3)

	hfg, err := NewHFGlobalWithReader(reader, frame)
	require.NoError(t, err)

	assert.Equal(t, int32(6), hfg.numHFPresets) // 1 + 5 = 6
}

func TestNewHFGlobalWithReader_AllModeLibrary(t *testing.T) {
	// quantAllDefault=false, all 17 modes are MODE_LIBRARY(0)
	readBitsData := make([]uint64, 18) // 17 for modes + 1 for numPresets
	// All zeros: MODE_LIBRARY for each index, then 0 for numPresets

	reader := &testcommon.FakeBitReader{
		ReadBoolData: []bool{false},
		ReadBitsData: readBitsData,
	}
	frame := &Frame{numGroups: 1}

	hfg, err := NewHFGlobalWithReader(reader, frame)
	require.NoError(t, err)
	require.NotNil(t, hfg)

	// Each param should match defaultParams (since MODE_LIBRARY copies defaults)
	for i := 0; i < 17; i++ {
		assert.True(t, hfg.params[i].Equals(defaultParams[i]),
			"params[%d] should match default", i)
	}
}

func TestNewHFGlobalWithReader_ReadBoolError(t *testing.T) {
	reader := &testcommon.FakeBitReader{} // empty → ReadBool error

	hfg, err := NewHFGlobalWithReader(reader, &Frame{numGroups: 1})
	assert.Error(t, err)
	assert.Nil(t, hfg)
}

func TestNewHFGlobalWithReader_WeightsGenerated(t *testing.T) {
	reader := &testcommon.FakeBitReader{
		ReadBoolData: []bool{true},
		ReadBitsData: []uint64{0},
	}
	frame := &Frame{numGroups: 1}

	hfg, err := NewHFGlobalWithReader(reader, frame)
	require.NoError(t, err)

	// Verify weights were generated for all 17 indices
	assert.Equal(t, 17, len(hfg.weights))
	for i := 0; i < 17; i++ {
		assert.Equal(t, 3, len(hfg.weights[i]), "index %d should have 3 channels", i)
		for c := 0; c < 3; c++ {
			assert.Greater(t, len(hfg.weights[i][c]), 0,
				"index=%d c=%d should have rows", i, c)
		}
	}
}

func TestNewHFGlobalWithReader_VerifySpecificWeights(t *testing.T) {
	reader := &testcommon.FakeBitReader{
		ReadBoolData: []bool{true},
		ReadBitsData: []uint64{0},
	}
	frame := &Frame{numGroups: 1}

	hfg, err := NewHFGlobalWithReader(reader, frame)
	require.NoError(t, err)

	// HORNUSS (index=1): w[0][0] should be 1/1 = 1.0
	assert.InDelta(t, 1.0, float64(hfg.weights[1][0][0][0]), 1e-6)

	// DCT2 (index=2): w[0][0] should be 1/1 = 1.0
	assert.InDelta(t, 1.0, float64(hfg.weights[2][0][0][0]), 1e-6)

	// AFV (index=10): w[0][0] should be 1/1 = 1.0
	assert.InDelta(t, 1.0, float64(hfg.weights[10][0][0][0]), 1e-6)
}

func TestNewHFGlobalWithReader_NumPresetsReadError(t *testing.T) {
	reader := &testcommon.FakeBitReader{
		ReadBoolData: []bool{true},
		// No ReadBitsData → error when reading numPresets
	}
	frame := &Frame{numGroups: 1}

	hfg, err := NewHFGlobalWithReader(reader, frame)
	assert.Error(t, err)
	assert.Nil(t, hfg)
}

// --- getHFPresets test ---

func TestGetHFPresets(t *testing.T) {
	hfg := &HFGlobal{numHFPresets: 42}
	assert.Equal(t, int32(42), hfg.getHFPresets())
}

// --- setupDefaultParams tests ---

func TestSetupDefaultParams_Count(t *testing.T) {
	// defaultParams is set up via init()
	assert.Equal(t, 17, len(defaultParams))
}

func TestSetupDefaultParams_Modes(t *testing.T) {
	expectedModes := []int32{
		MODE_DCT, MODE_HORNUSS, MODE_DCT2, MODE_DCT4, // 0-3
		MODE_DCT, MODE_DCT, MODE_DCT, MODE_DCT, MODE_DCT, // 4-8
		MODE_DCT4_8, MODE_AFV, // 9-10
		MODE_DCT, MODE_DCT, MODE_DCT, MODE_DCT, MODE_DCT, MODE_DCT, // 11-16
	}
	for i, expected := range expectedModes {
		assert.Equal(t, expected, defaultParams[i].mode,
			"defaultParams[%d] mode", i)
	}
}

func TestSetupDefaultParams_Denominators(t *testing.T) {
	// All default params have denominator=1
	for i := 0; i < 17; i++ {
		assert.Equal(t, float32(1), defaultParams[i].denominator,
			"defaultParams[%d] denominator", i)
	}
}

func TestSetupDefaultParams_DCT8HasDctParams(t *testing.T) {
	// index=0 (DCT8): dctParam should have 3 channels, 6 values each
	assert.Equal(t, 3, len(defaultParams[0].dctParam))
	assert.Equal(t, 6, len(defaultParams[0].dctParam[0]))
	assert.InDelta(t, 3150.0, defaultParams[0].dctParam[0][0], 1e-3)
}

func TestSetupDefaultParams_HornussHasParam(t *testing.T) {
	// index=1 (HORNUSS): param should have 3 channels, 3 values each
	assert.Equal(t, 3, len(defaultParams[1].param))
	assert.Equal(t, 3, len(defaultParams[1].param[0]))
	assert.InDelta(t, 280.0, float64(defaultParams[1].param[0][0]), 1e-3)
}
