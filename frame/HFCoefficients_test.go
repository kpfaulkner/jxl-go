package frame

import (
	"errors"
	"math"
	"testing"

	"github.com/kpfaulkner/jxl-go/bundle"
	"github.com/kpfaulkner/jxl-go/colour"
	"github.com/kpfaulkner/jxl-go/entropy"
	"github.com/kpfaulkner/jxl-go/testcommon"
	"github.com/kpfaulkner/jxl-go/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetPredictedNonZeros(t *testing.T) {
	nonZeros := util.MakeMatrix3D[int32](3, 4, 4)

	// Set up values for channel 0
	nonZeros[0][0][0] = 10
	nonZeros[0][0][1] = 20
	nonZeros[0][0][2] = 30
	nonZeros[0][1][0] = 15
	nonZeros[0][1][1] = 25
	nonZeros[0][2][0] = 40

	tests := []struct {
		name     string
		c        int
		y, x     int32
		expected int32
	}{
		{"origin returns 32", 0, 0, 0, 32},
		{"x=0 returns above", 0, 1, 0, 10},
		{"x=0 row2 returns above", 0, 2, 0, 15},
		{"y=0 returns left", 0, 0, 1, 10},
		{"y=0 col2 returns left", 0, 0, 2, 20},
		// interior: (nonZeros[c][y-1][x] + nonZeros[c][y][x-1] + 1) >> 1
		{"interior (1,1)", 0, 1, 1, (20 + 15 + 1) >> 1},
		{"interior (1,2)", 0, 1, 2, (30 + 25 + 1) >> 1},
		{"channel 1 origin", 1, 0, 0, 32},
		{"channel 1 all zeros interior", 1, 1, 1, (0 + 0 + 1) >> 1},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := getPredictedNonZeros(nonZeros, tc.c, tc.y, tc.x)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestGetNonZeroContext(t *testing.T) {
	hf := &HFCoefficients{
		hfctx: &HFBlockContext{
			numClusters: 5,
		},
	}

	tests := []struct {
		name      string
		predicted int32
		ctx       int32
		expected  int32
	}{
		{"predicted 0", 0, 2, 2 + 5*0},
		{"predicted 1", 1, 2, 2 + 5*1},
		{"predicted 7 (below threshold)", 7, 2, 2 + 5*7},
		{"predicted 8 (at threshold)", 8, 3, 3 + 5*(4+8/2)},
		{"predicted 10", 10, 0, 0 + 5*(4+10/2)},
		{"predicted 64 (at max)", 64, 1, 1 + 5*(4+64/2)},
		{"predicted 65 (clamped to 64)", 65, 1, 1 + 5*(4+64/2)},
		{"predicted 100 (clamped to 64)", 100, 0, 0 + 5*(4+64/2)},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := hf.getNonZeroContext(tc.predicted, tc.ctx)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestGetNonZeroContext_BoundaryAt8(t *testing.T) {
	hf := &HFCoefficients{
		hfctx: &HFBlockContext{
			numClusters: 10,
		},
	}

	// Below 8: ctx + numClusters * predicted
	below := hf.getNonZeroContext(7, 0)
	assert.Equal(t, int32(10*7), below)

	// At 8: ctx + numClusters * (4 + predicted/2)
	at := hf.getNonZeroContext(8, 0)
	assert.Equal(t, int32(10*(4+4)), at)
}

func TestGetCoefficientContext(t *testing.T) {
	hf := &HFCoefficients{}

	tests := []struct {
		name      string
		k         int32
		nonZeros  int32
		numBlocks int32
		prev      int32
		expected  int32
	}{
		{
			"basic case",
			1, 1, 1, 0,
			// nonZeros = (1+1-1)/1 = 1, k = 1/1 = 1
			(coeffNumNonzeroCtx[1] + coeffFreqCtx[1]) * 2,
		},
		{
			"prev=1 adds 1",
			1, 1, 1, 1,
			(coeffNumNonzeroCtx[1]+coeffFreqCtx[1])*2 + 1,
		},
		{
			"multi-block division",
			2, 4, 2, 0,
			// nonZeros = (4+2-1)/2 = 2, k = 2/2 = 1
			(coeffNumNonzeroCtx[2] + coeffFreqCtx[1]) * 2,
		},
		{
			"larger values",
			4, 10, 1, 1,
			// nonZeros = 10, k = 4
			(coeffNumNonzeroCtx[10]+coeffFreqCtx[4])*2 + 1,
		},
		{
			"max table indices",
			32, 32, 1, 0,
			// nonZeros = 32, k = 32
			(coeffNumNonzeroCtx[32] + coeffFreqCtx[32]) * 2,
		},
		{
			"numBlocks > 1 division rounding",
			3, 5, 2, 0,
			// nonZeros = (5+2-1)/2 = 3, k = 3/2 = 1
			(coeffNumNonzeroCtx[3] + coeffFreqCtx[1]) * 2,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := hf.getCoefficientContext(tc.k, tc.nonZeros, tc.numBlocks, tc.prev)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestGetBlockContext(t *testing.T) {
	hf := &HFCoefficients{
		hfctx: &HFBlockContext{
			qfThresholds:  []int32{5, 10, 20},
			numLFContexts: 4,
			clusterMap:    make([]int, 500),
			numClusters:   3,
		},
	}

	tests := []struct {
		name         string
		c            int
		orderID      int32
		hfMult       int32
		lfIndex      int32
		clusterValue int
	}{
		{"c=0 orderID=0 low hfMult", 0, 0, 3, 0, 42},
		{"c=1 orderID=0 low hfMult", 1, 0, 3, 0, 17},
		{"c=2 orderID=0 low hfMult", 2, 0, 3, 0, 99},
		{"c=0 orderID=0 high hfMult", 0, 0, 25, 0, 55},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Compute expected index to place the cluster value
			var idx int
			if tc.c < 2 {
				idx = 1 - tc.c
			} else {
				idx = tc.c
			}
			idx = idx*13 + int(tc.orderID)
			idx *= len(hf.hfctx.qfThresholds) + 1
			for _, threshold := range hf.hfctx.qfThresholds {
				if tc.hfMult > threshold {
					idx++
				}
			}
			idx *= int(hf.hfctx.numLFContexts)
			finalIdx := int32(idx) + tc.lfIndex

			require.Less(t, int(finalIdx), len(hf.hfctx.clusterMap))
			hf.hfctx.clusterMap[finalIdx] = tc.clusterValue

			result := hf.getBlockContext(tc.c, tc.orderID, tc.hfMult, tc.lfIndex)
			assert.Equal(t, int32(tc.clusterValue), result)
		})
	}
}

func TestGetBlockContext_QfThresholdCounting(t *testing.T) {
	hf := &HFCoefficients{
		hfctx: &HFBlockContext{
			qfThresholds:  []int32{5, 10, 20},
			numLFContexts: 1,
			clusterMap:    make([]int, 500),
			numClusters:   1,
		},
	}

	// c=1 (idx_c=0), orderID=0 -> base=0
	// base *= (3+1)=4
	// Then threshold count is added, then *= numLFContexts=1

	// hfMult=3: exceeds 0 thresholds -> offset 0
	hf.hfctx.clusterMap[0] = 10
	assert.Equal(t, int32(10), hf.getBlockContext(1, 0, 3, 0))

	// hfMult=6: exceeds 1 threshold (5) -> offset 1
	hf.hfctx.clusterMap[1] = 20
	assert.Equal(t, int32(20), hf.getBlockContext(1, 0, 6, 0))

	// hfMult=15: exceeds 2 thresholds (5,10) -> offset 2
	hf.hfctx.clusterMap[2] = 30
	assert.Equal(t, int32(30), hf.getBlockContext(1, 0, 15, 0))

	// hfMult=25: exceeds 3 thresholds (5,10,20) -> offset 3
	hf.hfctx.clusterMap[3] = 40
	assert.Equal(t, int32(40), hf.getBlockContext(1, 0, 25, 0))
}

func TestGetBlockContext_ChannelMapping(t *testing.T) {
	hf := &HFCoefficients{
		hfctx: &HFBlockContext{
			qfThresholds:  []int32{},
			numLFContexts: 1,
			clusterMap:    make([]int, 200),
			numClusters:   1,
		},
	}

	// With no qf thresholds: idx = idx_c * 13 * 1 * 1 = idx_c * 13
	// c=0: idx_c = 1-0 = 1 -> idx = 13
	hf.hfctx.clusterMap[13] = 1
	assert.Equal(t, int32(1), hf.getBlockContext(0, 0, 0, 0))

	// c=1: idx_c = 1-1 = 0 -> idx = 0
	hf.hfctx.clusterMap[0] = 2
	assert.Equal(t, int32(2), hf.getBlockContext(1, 0, 0, 0))

	// c=2: idx_c = 2 -> idx = 26
	hf.hfctx.clusterMap[26] = 3
	assert.Equal(t, int32(3), hf.getBlockContext(2, 0, 0, 0))
}

func TestNewHFCoefficientsWithReader_NilFrame(t *testing.T) {
	reader := &testcommon.FakeBitReader{}
	hf, err := NewHFCoefficientsWithReader(reader, nil, 0, 0)
	assert.Error(t, err)
	assert.Nil(t, hf)
	assert.Contains(t, err.Error(), "nil")
}

func TestNewHFCoefficientsWithReader_NilReader(t *testing.T) {
	frame := NewFakeFramer(VARDCT)
	hf, err := NewHFCoefficientsWithReader(nil, frame, 0, 0)
	assert.Error(t, err)
	assert.Nil(t, hf)
	assert.Contains(t, err.Error(), "nil")
}

func TestNewHFCoefficientsWithReader_BothNil(t *testing.T) {
	hf, err := NewHFCoefficientsWithReader(nil, nil, 0, 0)
	assert.Error(t, err)
	assert.Nil(t, hf)
}

func TestNewHFCoefficientsWithReader_ReadBitsError(t *testing.T) {
	frame := NewFakeFramer(VARDCT)
	reader := &testcommon.FakeBitReader{
		ReadBitsData: []uint64{}, // empty -> error on first ReadBits
	}
	hf, err := NewHFCoefficientsWithReader(reader, frame, 0, 0)
	assert.Error(t, err)
	assert.Nil(t, hf)
}

func TestRelease(t *testing.T) {
	hf := &HFCoefficients{
		quantizedCoeffs: util.MakeMatrix3D[int32](3, 4, 4),
		dequantHFCoeff:  util.MakeMatrix3D[float32](3, 4, 4),
	}

	require.NotNil(t, hf.quantizedCoeffs)
	require.NotNil(t, hf.dequantHFCoeff)

	hf.Release()

	assert.Nil(t, hf.quantizedCoeffs)
	assert.Nil(t, hf.dequantHFCoeff)
}

func TestRelease_AlreadyNil(t *testing.T) {
	hf := &HFCoefficients{}
	// Should not panic when fields are already nil
	hf.Release()
	assert.Nil(t, hf.quantizedCoeffs)
	assert.Nil(t, hf.dequantHFCoeff)
}

func TestRelease_Idempotent(t *testing.T) {
	hf := &HFCoefficients{
		quantizedCoeffs: util.MakeMatrix3D[int32](3, 2, 2),
		dequantHFCoeff:  util.MakeMatrix3D[float32](3, 2, 2),
	}

	hf.Release()
	hf.Release() // second call should not panic

	assert.Nil(t, hf.quantizedCoeffs)
	assert.Nil(t, hf.dequantHFCoeff)
}

func TestDisplayHFCoefficients(t *testing.T) {
	hf := &HFCoefficients{
		quantizedCoeffs: util.MakeMatrix3D[int32](3, 4, 4),
		dequantHFCoeff:  util.MakeMatrix3D[float32](3, 4, 4),
	}

	hf.quantizedCoeffs[0][0][0] = 5
	hf.quantizedCoeffs[1][1][1] = -3
	hf.dequantHFCoeff[0][0][0] = 1.5
	hf.dequantHFCoeff[2][2][2] = -0.5

	// Should not panic
	hf.DisplayHFCoefficients()
}

// ---- Additional getPredictedNonZeros tests ----

func TestGetPredictedNonZeros_LargeValuesRounding(t *testing.T) {
	nonZeros := util.MakeMatrix3D[int32](3, 4, 4)

	tests := []struct {
		name     string
		above    int32
		left     int32
		expected int32
	}{
		// (above + left + 1) >> 1
		{"even sum rounds down", 10, 10, (10 + 10 + 1) >> 1},  // 10
		{"odd sum rounds up", 10, 11, (10 + 11 + 1) >> 1},     // 11
		{"large even", 100, 200, (100 + 200 + 1) >> 1},        // 150
		{"large odd", 99, 200, (99 + 200 + 1) >> 1},           // 150
		{"zero and nonzero", 0, 100, (0 + 100 + 1) >> 1},      // 50
		{"both zero", 0, 0, (0 + 0 + 1) >> 1},                 // 0
		{"max int32 safe", 1000, 1000, (1000 + 1000 + 1) >> 1}, // 1000
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			nonZeros[0][0][1] = tc.above // nonZeros[c][y-1][x] at position (1,1)
			nonZeros[0][1][0] = tc.left  // nonZeros[c][y][x-1] at position (1,1)
			result := getPredictedNonZeros(nonZeros, 0, 1, 1)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestGetPredictedNonZeros_AllChannels(t *testing.T) {
	nonZeros := util.MakeMatrix3D[int32](3, 2, 2)
	nonZeros[0][0][1] = 10
	nonZeros[0][1][0] = 20
	nonZeros[1][0][1] = 30
	nonZeros[1][1][0] = 40
	nonZeros[2][0][1] = 50
	nonZeros[2][1][0] = 60

	// Each channel should compute independently
	assert.Equal(t, int32((10+20+1)>>1), getPredictedNonZeros(nonZeros, 0, 1, 1))
	assert.Equal(t, int32((30+40+1)>>1), getPredictedNonZeros(nonZeros, 1, 1, 1))
	assert.Equal(t, int32((50+60+1)>>1), getPredictedNonZeros(nonZeros, 2, 1, 1))
}

// ---- Additional getNonZeroContext tests ----

func TestGetNonZeroContext_NegativePredicted(t *testing.T) {
	hf := &HFCoefficients{
		hfctx: &HFBlockContext{
			numClusters: 5,
		},
	}

	// Negative predicted is < 8, so takes the direct path: ctx + numClusters * predicted
	// This would result in a negative value which is unusual but follows the code path
	result := hf.getNonZeroContext(-1, 0)
	assert.Equal(t, int32(5*(-1)), result)
}

func TestGetNonZeroContext_ZeroContext(t *testing.T) {
	hf := &HFCoefficients{
		hfctx: &HFBlockContext{
			numClusters: 1,
		},
	}

	// With numClusters=1 and ctx=0:
	// predicted < 8: 0 + 1 * predicted = predicted
	assert.Equal(t, int32(0), hf.getNonZeroContext(0, 0))
	assert.Equal(t, int32(7), hf.getNonZeroContext(7, 0))
	// predicted >= 8: 0 + 1 * (4 + predicted/2)
	assert.Equal(t, int32(4+8/2), hf.getNonZeroContext(8, 0))
}

// ---- Additional getCoefficientContext tests ----

func TestGetCoefficientContext_TableLookups(t *testing.T) {
	hf := &HFCoefficients{}

	tests := []struct {
		name      string
		k         int32
		nonZeros  int32
		numBlocks int32
		prev      int32
		expected  int32
	}{
		{
			"k/numBlocks=0 uses coeffFreqCtx[0]=-1",
			0, 1, 1, 0,
			// nonZeros=(1+1-1)/1=1, k=0/1=0
			(coeffNumNonzeroCtx[1] + coeffFreqCtx[0]) * 2,
		},
		{
			"high k index uses coeffFreqCtx[63]=30",
			63, 1, 1, 0,
			(coeffNumNonzeroCtx[1] + coeffFreqCtx[63]) * 2,
		},
		{
			"high nonzeros uses coeffNumNonzeroCtx[63]=206",
			1, 63, 1, 0,
			(coeffNumNonzeroCtx[63] + coeffFreqCtx[1]) * 2,
		},
		{
			"both at zero index",
			0, 0, 1, 0,
			// nonZeros=(0+1-1)/1=0, k=0
			(coeffNumNonzeroCtx[0] + coeffFreqCtx[0]) * 2,
		},
		{
			"large numBlocks division",
			10, 20, 4, 1,
			// nonZeros = (20+4-1)/4 = 5, k = 10/4 = 2
			(coeffNumNonzeroCtx[5]+coeffFreqCtx[2])*2 + 1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := hf.getCoefficientContext(tc.k, tc.nonZeros, tc.numBlocks, tc.prev)
			assert.Equal(t, tc.expected, result)
		})
	}
}

// ---- Additional getBlockContext tests ----

func TestGetBlockContext_NonZeroOrderID(t *testing.T) {
	hf := &HFCoefficients{
		hfctx: &HFBlockContext{
			qfThresholds:  []int32{},
			numLFContexts: 1,
			clusterMap:    make([]int, 200),
			numClusters:   1,
		},
	}

	// With no qf thresholds and numLFContexts=1:
	// idx = idx_c * 13 + orderID
	// idx *= 1 (no thresholds)
	// idx *= 1 (numLFContexts)
	// So idx = idx_c * 13 + orderID

	// c=1: idx_c=0, orderID=5 -> idx = 5
	hf.hfctx.clusterMap[5] = 77
	assert.Equal(t, int32(77), hf.getBlockContext(1, 5, 0, 0))

	// c=0: idx_c=1, orderID=12 -> idx = 13 + 12 = 25
	hf.hfctx.clusterMap[25] = 88
	assert.Equal(t, int32(88), hf.getBlockContext(0, 12, 0, 0))

	// c=2: idx_c=2, orderID=3 -> idx = 26 + 3 = 29
	hf.hfctx.clusterMap[29] = 99
	assert.Equal(t, int32(99), hf.getBlockContext(2, 3, 0, 0))
}

func TestGetBlockContext_NonZeroLfIndex(t *testing.T) {
	hf := &HFCoefficients{
		hfctx: &HFBlockContext{
			qfThresholds:  []int32{},
			numLFContexts: 4,
			clusterMap:    make([]int, 500),
			numClusters:   1,
		},
	}

	// c=1: idx_c=0, orderID=0, no thresholds
	// idx = 0 * 13 + 0 = 0
	// idx *= 1 (no thresholds) -> 0
	// idx *= 4 (numLFContexts) -> 0
	// final = 0 + lfIndex

	hf.hfctx.clusterMap[0] = 10
	hf.hfctx.clusterMap[1] = 20
	hf.hfctx.clusterMap[2] = 30
	hf.hfctx.clusterMap[3] = 40

	assert.Equal(t, int32(10), hf.getBlockContext(1, 0, 0, 0))
	assert.Equal(t, int32(20), hf.getBlockContext(1, 0, 0, 1))
	assert.Equal(t, int32(30), hf.getBlockContext(1, 0, 0, 2))
	assert.Equal(t, int32(40), hf.getBlockContext(1, 0, 0, 3))
}

func TestGetBlockContext_QfThresholdExactBoundary(t *testing.T) {
	hf := &HFCoefficients{
		hfctx: &HFBlockContext{
			qfThresholds:  []int32{10},
			numLFContexts: 1,
			clusterMap:    make([]int, 200),
			numClusters:   1,
		},
	}

	// c=1 (idx_c=0), orderID=0
	// idx = 0, idx *= 2 (1 threshold + 1)
	// hfMult=10 is NOT > 10 (strictly greater than) -> offset 0 -> idx = 0
	hf.hfctx.clusterMap[0] = 50
	assert.Equal(t, int32(50), hf.getBlockContext(1, 0, 10, 0))

	// hfMult=11 IS > 10 -> offset 1 -> idx = 1
	hf.hfctx.clusterMap[1] = 60
	assert.Equal(t, int32(60), hf.getBlockContext(1, 0, 11, 0))
}

func TestGetBlockContext_CombinedChannelOrderThreshold(t *testing.T) {
	hf := &HFCoefficients{
		hfctx: &HFBlockContext{
			qfThresholds:  []int32{5},
			numLFContexts: 2,
			clusterMap:    make([]int, 200),
			numClusters:   1,
		},
	}

	// c=0, idx_c=1, orderID=2, hfMult=10 (>5, offset 1), lfIndex=1
	// idx = 1*13 + 2 = 15
	// idx *= 2 (1 threshold + 1) -> 30
	// idx += 1 (hfMult>5) -> 31
	// idx *= 2 (numLFContexts) -> 62
	// final = 62 + 1 = 63
	hf.hfctx.clusterMap[63] = 42
	assert.Equal(t, int32(42), hf.getBlockContext(0, 2, 10, 1))
}

// ---- Display method tests ----

func TestDisplayHFCoefficients_EmptyMatrices(t *testing.T) {
	hf := &HFCoefficients{
		quantizedCoeffs: util.MakeMatrix3D[int32](3, 0, 0),
		dequantHFCoeff:  util.MakeMatrix3D[float32](3, 0, 0),
	}
	// Should not panic with zero-sized matrices
	hf.DisplayHFCoefficients()
}

func TestDisplayHFCoefficients_SingleElement(t *testing.T) {
	hf := &HFCoefficients{
		quantizedCoeffs: util.MakeMatrix3D[int32](3, 1, 1),
		dequantHFCoeff:  util.MakeMatrix3D[float32](3, 1, 1),
	}
	hf.quantizedCoeffs[0][0][0] = 100
	hf.quantizedCoeffs[1][0][0] = -100
	hf.dequantHFCoeff[0][0][0] = 3.14
	hf.dequantHFCoeff[2][0][0] = -2.71

	// Should not panic
	hf.DisplayHFCoefficients()
}

// ---- dequantizeHFCoefficients tests ----

// makeHFCoeffForDequant creates an HFCoefficients struct suitable for testing dequantizeHFCoefficients.
// It sets up a single DCT8 block at position (0,0) with all weights=1.0 and globalScale=1.0.
func makeHFCoeffForDequant(hfMult int32) *HFCoefficients {
	tt := DCT8 // parameterIndex=0, orderID=0, pixel 8x8, dctSelect 1x1

	dctSelect := make([][]*TransformType, 1)
	dctSelect[0] = make([]*TransformType, 1)
	dctSelect[0][0] = tt

	hfMultiplier := util.MakeMatrix2D[int32](1, 1)
	hfMultiplier[0][0] = hfMult

	hfMetadata := &HFMetadata{
		dctSelect:    dctSelect,
		hfMultiplier: hfMultiplier,
		blockList:    []util.Point{{X: 0, Y: 0}},
	}

	lfg := &LFGroup{
		hfMetadata: hfMetadata,
		lfCoeff: &LFCoefficients{
			dequantLFCoeff: util.MakeMatrix3D[float32](3, 1, 1),
			lfIndex:        util.MakeMatrix2D[int32](1, 1),
		},
	}

	// weights[parameterIndex=0][channel][y][x] - all 1.0
	weights := make([][][][]float32, 1)
	weights[0] = make([][][]float32, 3)
	for c := 0; c < 3; c++ {
		weights[0][c] = util.MakeMatrix2D[float32](8, 8)
		for y := 0; y < 8; y++ {
			for x := 0; x < 8; x++ {
				weights[0][c][y][x] = 1.0
			}
		}
	}

	ff := &FakeFramer{
		header: &FrameHeader{
			jpegUpsamplingX: []int32{0, 0, 0},
			jpegUpsamplingY: []int32{0, 0, 0},
			xqmScale:        2, // 0.8^(2-2) = 0.8^0 = 1.0
			bqmScale:        2, // 0.8^(2-2) = 0.8^0 = 1.0
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
	ff.lfGlobal.globalScale = 65536 // -> 65536.0 / 65536 = 1.0

	quantizedCoeffs := util.MakeMatrix3D[int32](3, 8, 8)
	dequantHFCoeff := util.MakeMatrix3D[float32](3, 8, 8)

	return &HFCoefficients{
		frame:           ff,
		lfg:             lfg,
		quantizedCoeffs: quantizedCoeffs,
		dequantHFCoeff:  dequantHFCoeff,
		groupPos:        util.Point{X: 0, Y: 0},
		blocks:          []*util.Point{{X: 0, Y: 0}},
	}
}

func TestDequantizeHFCoefficients_SmallCoeffs(t *testing.T) {
	// Test the qbclut path for small coefficients (-1, 0, 1)
	hf := makeHFCoeffForDequant(1)
	oim := colour.NewOpsinInverseMatrix()

	// DCT8: dctSelectHeight=1, dctSelectWidth=1
	// Inner loop skips y<1 && x<1, i.e., only skips (0,0)
	// Position (0,1) is processed: y=0, x=1

	// Channel 0: coeff=0 -> quant = qbclut[0][0+1] = 0.0
	hf.quantizedCoeffs[0][0][1] = 0
	// Channel 0: coeff=1 -> quant = qbclut[0][1+1] = QuantBias[0]
	hf.quantizedCoeffs[0][0][2] = 1
	// Channel 0: coeff=-1 -> quant = qbclut[0][-1+1] = -QuantBias[0]
	hf.quantizedCoeffs[0][0][3] = -1

	err := hf.dequantizeHFCoefficients()
	require.NoError(t, err)

	// With globalScale=1.0, scaleFactor[0]=1.0, hfMultiplier=1, weights=1.0:
	// dequantHFCoeff = quant * 1.0 * 1.0
	assert.Equal(t, float32(0.0), hf.dequantHFCoeff[0][0][1])
	assert.InDelta(t, float64(oim.QuantBias[0]), float64(hf.dequantHFCoeff[0][0][2]), 0.0001)
	assert.InDelta(t, float64(-oim.QuantBias[0]), float64(hf.dequantHFCoeff[0][0][3]), 0.0001)
}

func TestDequantizeHFCoefficients_LargeCoeffs(t *testing.T) {
	// Test the large coefficient path: quant = float32(coeff) - QuantBiasNumerator/float32(coeff)
	hf := makeHFCoeffForDequant(1)
	oim := colour.NewOpsinInverseMatrix()

	// coeff=5 at position (0,1)
	hf.quantizedCoeffs[0][0][1] = 5

	// coeff=-3 at position (0,2)
	hf.quantizedCoeffs[0][0][2] = -3

	// coeff=2 (boundary: not in [-1,1]) at position (0,3)
	hf.quantizedCoeffs[0][0][3] = 2

	// coeff=-2 at position (0,4)
	hf.quantizedCoeffs[0][0][4] = -2

	err := hf.dequantizeHFCoefficients()
	require.NoError(t, err)

	// quant = coeff - QuantBiasNumerator / coeff
	expected5 := float32(5) - oim.QuantBiasNumerator/float32(5)
	expectedN3 := float32(-3) - oim.QuantBiasNumerator/float32(-3)
	expected2 := float32(2) - oim.QuantBiasNumerator/float32(2)
	expectedN2 := float32(-2) - oim.QuantBiasNumerator/float32(-2)

	assert.InDelta(t, float64(expected5), float64(hf.dequantHFCoeff[0][0][1]), 0.0001)
	assert.InDelta(t, float64(expectedN3), float64(hf.dequantHFCoeff[0][0][2]), 0.0001)
	assert.InDelta(t, float64(expected2), float64(hf.dequantHFCoeff[0][0][3]), 0.0001)
	assert.InDelta(t, float64(expectedN2), float64(hf.dequantHFCoeff[0][0][4]), 0.0001)
}

func TestDequantizeHFCoefficients_ScaleFactors(t *testing.T) {
	// Test that xqmScale and bqmScale affect channels 0 and 2 respectively
	hf := makeHFCoeffForDequant(1)
	header := hf.frame.getFrameHeader()
	header.xqmScale = 3 // scaleFactor[0] = globalScale * 0.8^(3-2) = 1.0 * 0.8
	header.bqmScale = 4 // scaleFactor[2] = globalScale * 0.8^(4-2) = 1.0 * 0.64

	// Set coeff=5 at (0,1) for all 3 channels
	for c := 0; c < 3; c++ {
		hf.quantizedCoeffs[c][0][1] = 5
	}

	err := hf.dequantizeHFCoefficients()
	require.NoError(t, err)

	oim := colour.NewOpsinInverseMatrix()
	quant5 := float32(5) - oim.QuantBiasNumerator/float32(5)

	// Channel 0: scaleFactor = 1.0 * 0.8^1 = 0.8
	expected0 := quant5 * float32(math.Pow(0.8, 1.0))
	// Channel 1: scaleFactor = 1.0 (no xqm/bqm scale)
	expected1 := quant5 * 1.0
	// Channel 2: scaleFactor = 1.0 * 0.8^2 = 0.64
	expected2 := quant5 * float32(math.Pow(0.8, 2.0))

	assert.InDelta(t, float64(expected0), float64(hf.dequantHFCoeff[0][0][1]), 0.001)
	assert.InDelta(t, float64(expected1), float64(hf.dequantHFCoeff[1][0][1]), 0.001)
	assert.InDelta(t, float64(expected2), float64(hf.dequantHFCoeff[2][0][1]), 0.001)
}

func TestDequantizeHFCoefficients_HfMultiplierDivision(t *testing.T) {
	// Test that hfMultiplier divides the scale factor
	hf := makeHFCoeffForDequant(4) // hfMultiplier = 4

	hf.quantizedCoeffs[0][0][1] = 5

	err := hf.dequantizeHFCoefficients()
	require.NoError(t, err)

	oim := colour.NewOpsinInverseMatrix()
	quant5 := float32(5) - oim.QuantBiasNumerator/float32(5)
	// sfc = scaleFactor[0] / hfMultiplier = 1.0 / 4.0 = 0.25
	expected := quant5 * 0.25

	assert.InDelta(t, float64(expected), float64(hf.dequantHFCoeff[0][0][1]), 0.001)
}

func TestDequantizeHFCoefficients_WeightMultiplication(t *testing.T) {
	// Test that per-position weights are applied
	hf := makeHFCoeffForDequant(1)

	// Set a non-uniform weight at a specific position
	weights := hf.frame.getHFGlobal().weights
	// DCT8 flip()=true: wy=x, wx = x^y^wy = y
	// For position (y=0, x=1): wy=1, wx=0 -> w3[1][0]
	weights[0][0][1][0] = 2.5

	hf.quantizedCoeffs[0][0][1] = 5

	err := hf.dequantizeHFCoefficients()
	require.NoError(t, err)

	oim := colour.NewOpsinInverseMatrix()
	quant5 := float32(5) - oim.QuantBiasNumerator/float32(5)
	expected := quant5 * 1.0 * 2.5 // sfc=1.0, weight=2.5

	assert.InDelta(t, float64(expected), float64(hf.dequantHFCoeff[0][0][1]), 0.001)
}

func TestDequantizeHFCoefficients_NilBlockSkipped(t *testing.T) {
	hf := makeHFCoeffForDequant(1)
	// Replace blocks with a nil entry
	hf.blocks = []*util.Point{nil}

	hf.quantizedCoeffs[0][0][1] = 100

	err := hf.dequantizeHFCoefficients()
	require.NoError(t, err)

	// Nothing should be dequantized since the block is nil
	assert.Equal(t, float32(0.0), hf.dequantHFCoeff[0][0][1])
}

func TestDequantizeHFCoefficients_SkipsDctSelectRegion(t *testing.T) {
	// Verify that pixels within the dctSelect region (y < dctSelectHeight && x < dctSelectWidth)
	// are skipped in dequantization
	hf := makeHFCoeffForDequant(1)

	// DCT8: dctSelectHeight=1, dctSelectWidth=1
	// (0,0) should be skipped
	hf.quantizedCoeffs[0][0][0] = 999

	err := hf.dequantizeHFCoefficients()
	require.NoError(t, err)

	// (0,0) should remain at 0.0 (not dequantized)
	assert.Equal(t, float32(0.0), hf.dequantHFCoeff[0][0][0])
}

func TestDequantizeHFCoefficients_MultipleChannels(t *testing.T) {
	// Verify all 3 channels are dequantized independently with their own QuantBias
	hf := makeHFCoeffForDequant(1)
	oim := colour.NewOpsinInverseMatrix()

	// Set coeff=1 at (0,1) for all channels to test per-channel QuantBias
	for c := 0; c < 3; c++ {
		hf.quantizedCoeffs[c][0][1] = 1
	}

	err := hf.dequantizeHFCoefficients()
	require.NoError(t, err)

	// Each channel should use its own QuantBias[c]
	for c := 0; c < 3; c++ {
		assert.InDelta(t, float64(oim.QuantBias[c]), float64(hf.dequantHFCoeff[c][0][1]), 0.0001,
			"channel %d should use QuantBias[%d]", c, c)
	}
	// Verify they're actually different (default QuantBias values differ per channel)
	assert.NotEqual(t, hf.dequantHFCoeff[0][0][1], hf.dequantHFCoeff[1][0][1])
}

func TestDequantizeHFCoefficients_GlobalScaleEffect(t *testing.T) {
	// Test with a different globalScale value
	hf := makeHFCoeffForDequant(1)
	hf.frame.getLFGlobal().globalScale = 32768 // globalScale = 65536/32768 = 2.0

	hf.quantizedCoeffs[0][0][1] = 5

	err := hf.dequantizeHFCoefficients()
	require.NoError(t, err)

	oim := colour.NewOpsinInverseMatrix()
	quant5 := float32(5) - oim.QuantBiasNumerator/float32(5)
	// scaleFactor[0] = 2.0 * 0.8^0 = 2.0, sfc = 2.0/1 = 2.0
	expected := quant5 * 2.0

	assert.InDelta(t, float64(expected), float64(hf.dequantHFCoeff[0][0][1]), 0.001)
}

// ---- chromaFromLuma tests ----

func makeHFCoeffForChroma() *HFCoefficients {
	tt := DCT8 // pixelHeight=8, pixelWidth=8

	dctSelect := make([][]*TransformType, 1)
	dctSelect[0] = make([]*TransformType, 1)
	dctSelect[0][0] = tt

	// hfStreamBuffer: 2 channels (x and b factors), at least 1 row, 1 col
	hfStreamBuffer := make([][][]int32, 2)
	hfStreamBuffer[0] = util.MakeMatrix2D[int32](1, 1) // xFactorHF
	hfStreamBuffer[1] = util.MakeMatrix2D[int32](1, 1) // bFactorHF

	hfMetadata := &HFMetadata{
		dctSelect:      dctSelect,
		hfMultiplier:   util.MakeMatrix2D[int32](1, 1),
		blockList:      []util.Point{{X: 0, Y: 0}},
		hfStreamBuffer: hfStreamBuffer,
	}

	lfg := &LFGroup{
		hfMetadata: hfMetadata,
	}

	ff := &FakeFramer{
		header: &FrameHeader{
			jpegUpsamplingX: []int32{0, 0, 0},
			jpegUpsamplingY: []int32{0, 0, 0},
			passes:          NewPassesInfo(),
			Encoding:        VARDCT,
		},
		lfGlobal: NewLFGlobal(),
	}
	// Set valid lfChanCorr defaults
	ff.lfGlobal.lfChanCorr = &LFChannelCorrelation{
		colorFactor:      84,
		baseCorrelationX: 0.0,
		baseCorrelationB: 1.0,
		xFactorLF:        128,
		bFactorLF:        128,
	}

	dequantHFCoeff := util.MakeMatrix3D[float32](3, 8, 8)

	return &HFCoefficients{
		frame:          ff,
		lfg:            lfg,
		dequantHFCoeff: dequantHFCoeff,
		groupPos:       util.Point{X: 0, Y: 0},
		blocks:         []*util.Point{{X: 0, Y: 0}},
	}
}

func TestChromaFromLuma_ReturnsEarlyWithUpsampling(t *testing.T) {
	hf := makeHFCoeffForChroma()
	// Set upsampling on one channel to trigger early return
	hf.frame.getFrameHeader().jpegUpsamplingX[1] = 1

	// Pre-fill Y channel with values
	hf.dequantHFCoeff[1][0][0] = 10.0
	// Set X channel to something we can verify doesn't change
	hf.dequantHFCoeff[0][0][0] = 5.0
	hf.dequantHFCoeff[2][0][0] = 7.0

	err := hf.chromaFromLuma()
	require.NoError(t, err)

	// X and B channels should be unchanged since function returned early
	assert.Equal(t, float32(5.0), hf.dequantHFCoeff[0][0][0])
	assert.Equal(t, float32(7.0), hf.dequantHFCoeff[2][0][0])
}

func TestChromaFromLuma_AppliesCorrelation(t *testing.T) {
	hf := makeHFCoeffForChroma()

	// Set hfStreamBuffer factors
	hf.lfg.hfMetadata.hfStreamBuffer[0][0][0] = 42  // xFactor
	hf.lfg.hfMetadata.hfStreamBuffer[1][0][0] = -84 // bFactor

	lfc := hf.frame.getLFGlobal().lfChanCorr
	// kX = baseCorrelationX + xFactor * (1/colorFactor) = 0.0 + 42/84 = 0.5
	// kB = baseCorrelationB + bFactor * (1/colorFactor) = 1.0 + (-84)/84 = 0.0
	expectedKX := lfc.baseCorrelationX + float32(42)/float32(lfc.colorFactor)
	expectedKB := lfc.baseCorrelationB + float32(-84)/float32(lfc.colorFactor)

	// Set Y channel values at a few positions
	hf.dequantHFCoeff[1][0][0] = 10.0
	hf.dequantHFCoeff[1][0][1] = 20.0
	hf.dequantHFCoeff[1][3][5] = 5.0

	// Set initial X and B values
	hf.dequantHFCoeff[0][0][0] = 1.0
	hf.dequantHFCoeff[0][0][1] = 2.0
	hf.dequantHFCoeff[2][0][0] = 3.0
	hf.dequantHFCoeff[2][0][1] = 4.0

	err := hf.chromaFromLuma()
	require.NoError(t, err)

	// X[y][x] += kX * Y[y][x]
	assert.InDelta(t, 1.0+float64(expectedKX)*10.0, float64(hf.dequantHFCoeff[0][0][0]), 0.001)
	assert.InDelta(t, 2.0+float64(expectedKX)*20.0, float64(hf.dequantHFCoeff[0][0][1]), 0.001)
	// B[y][x] += kB * Y[y][x]
	assert.InDelta(t, 3.0+float64(expectedKB)*10.0, float64(hf.dequantHFCoeff[2][0][0]), 0.001)
	assert.InDelta(t, 4.0+float64(expectedKB)*20.0, float64(hf.dequantHFCoeff[2][0][1]), 0.001)
	// Y channel should be unchanged
	assert.Equal(t, float32(10.0), hf.dequantHFCoeff[1][0][0])
}

func TestChromaFromLuma_ZeroYChannel(t *testing.T) {
	hf := makeHFCoeffForChroma()
	hf.lfg.hfMetadata.hfStreamBuffer[0][0][0] = 42
	hf.lfg.hfMetadata.hfStreamBuffer[1][0][0] = 42

	// Y channel is all zeros -> correlation adds nothing
	hf.dequantHFCoeff[0][0][1] = 5.0
	hf.dequantHFCoeff[2][0][1] = 7.0

	err := hf.chromaFromLuma()
	require.NoError(t, err)

	// X and B should be unchanged when Y is zero
	assert.Equal(t, float32(5.0), hf.dequantHFCoeff[0][0][1])
	assert.Equal(t, float32(7.0), hf.dequantHFCoeff[2][0][1])
}

func TestChromaFromLuma_NilBlockSkipped(t *testing.T) {
	hf := makeHFCoeffForChroma()
	hf.blocks = []*util.Point{nil}
	hf.lfg.hfMetadata.hfStreamBuffer[0][0][0] = 42
	hf.lfg.hfMetadata.hfStreamBuffer[1][0][0] = 42

	hf.dequantHFCoeff[1][0][0] = 100.0
	hf.dequantHFCoeff[0][0][0] = 1.0

	err := hf.chromaFromLuma()
	require.NoError(t, err)

	// No correlation applied since block is nil
	assert.Equal(t, float32(1.0), hf.dequantHFCoeff[0][0][0])
}

func TestChromaFromLuma_BaseCorrelationValues(t *testing.T) {
	hf := makeHFCoeffForChroma()

	// Set hf factors to 0 -> kX/kB depend only on base correlations
	hf.lfg.hfMetadata.hfStreamBuffer[0][0][0] = 0
	hf.lfg.hfMetadata.hfStreamBuffer[1][0][0] = 0

	lfc := hf.frame.getLFGlobal().lfChanCorr
	lfc.baseCorrelationX = 2.0
	lfc.baseCorrelationB = 3.0

	// kX = 2.0 + 0/84 = 2.0
	// kB = 3.0 + 0/84 = 3.0

	hf.dequantHFCoeff[1][0][1] = 4.0 // Y value
	hf.dequantHFCoeff[0][0][1] = 0.0 // X initial
	hf.dequantHFCoeff[2][0][1] = 0.0 // B initial

	err := hf.chromaFromLuma()
	require.NoError(t, err)

	// X += 2.0 * 4.0 = 8.0
	assert.InDelta(t, 8.0, float64(hf.dequantHFCoeff[0][0][1]), 0.001)
	// B += 3.0 * 4.0 = 12.0
	assert.InDelta(t, 12.0, float64(hf.dequantHFCoeff[2][0][1]), 0.001)
}

// ---- bakeDequantizedCoeffs orchestration test ----

func TestBakeDequantizedCoeffs_CallsAllThreeSteps(t *testing.T) {
	// bakeDequantizedCoeffs calls dequantize, chromaFromLuma, then finalizeLLF in order.
	// We test that it runs without error for a minimal valid setup.
	hf := makeHFCoeffForDequant(1)

	// Add hfStreamBuffer for chromaFromLuma
	hf.lfg.hfMetadata.hfStreamBuffer = make([][][]int32, 2)
	hf.lfg.hfMetadata.hfStreamBuffer[0] = util.MakeMatrix2D[int32](1, 1)
	hf.lfg.hfMetadata.hfStreamBuffer[1] = util.MakeMatrix2D[int32](1, 1)

	// Set valid lfChanCorr
	hf.frame.getLFGlobal().lfChanCorr = &LFChannelCorrelation{
		colorFactor:      84,
		baseCorrelationX: 0.0,
		baseCorrelationB: 0.0,
		xFactorLF:        128,
		bFactorLF:        128,
	}

	// Ensure lfCoeff.dequantLFCoeff is properly sized for finalizeLLF
	// DCT8: dctSelectHeight=1, dctSelectWidth=1
	// finalizeLLF reads dqlf[c][sLfgY][sLfgX] where sLfgY=pos.Y, sLfgX=pos.X
	hf.lfg.lfCoeff.dequantLFCoeff = util.MakeMatrix3D[float32](3, 1, 1)

	// Set a quantized coefficient to verify the full pipeline runs
	hf.quantizedCoeffs[0][0][1] = 3

	err := hf.bakeDequantizedCoeffs()
	require.NoError(t, err)

	// The coefficient should have been dequantized (non-zero after pipeline)
	// Exact value depends on dequant + chromaFromLuma + finalizeLLF
	// Just verify the pipeline completed and produced some output
	assert.NotEqual(t, float32(0.0), hf.dequantHFCoeff[0][0][1],
		"dequantized coefficient should be non-zero after full pipeline")
}

// ---- makeHFCoeffForDequantWithTT helper ----

// makeHFCoeffForDequantWithTT creates an HFCoefficients suitable for testing
// dequantizeHFCoefficients with a specific TransformType.
func makeHFCoeffForDequantWithTT(tt *TransformType, hfMult int32) *HFCoefficients {
	dctSelect := make([][]*TransformType, 1)
	dctSelect[0] = make([]*TransformType, 1)
	dctSelect[0][0] = tt

	hfMultiplier := util.MakeMatrix2D[int32](1, 1)
	hfMultiplier[0][0] = hfMult

	hfMetadata := &HFMetadata{
		dctSelect:    dctSelect,
		hfMultiplier: hfMultiplier,
		blockList:    []util.Point{{X: 0, Y: 0}},
	}

	lfg := &LFGroup{
		hfMetadata: hfMetadata,
		lfCoeff: &LFCoefficients{
			dequantLFCoeff: util.MakeMatrix3D[float32](3, 1, 1),
			lfIndex:        util.MakeMatrix2D[int32](1, 1),
		},
	}

	numParams := int(tt.parameterIndex + 1)
	weights := make([][][][]float32, numParams)
	for p := 0; p < numParams; p++ {
		weights[p] = make([][][]float32, 3)
		for c := 0; c < 3; c++ {
			weights[p][c] = util.MakeMatrix2D[float32](int(tt.pixelHeight), int(tt.pixelWidth))
			for y := 0; y < int(tt.pixelHeight); y++ {
				for x := 0; x < int(tt.pixelWidth); x++ {
					weights[p][c][y][x] = 1.0
				}
			}
		}
	}

	ff := &FakeFramer{
		header: &FrameHeader{
			jpegUpsamplingX: []int32{0, 0, 0},
			jpegUpsamplingY: []int32{0, 0, 0},
			xqmScale:        2,
			bqmScale:        2,
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
	ff.lfGlobal.globalScale = 65536

	pH := int(tt.pixelHeight)
	pW := int(tt.pixelWidth)
	quantizedCoeffs := util.MakeMatrix3D[int32](3, pH, pW)
	dequantHFCoeff := util.MakeMatrix3D[float32](3, pH, pW)

	return &HFCoefficients{
		frame:           ff,
		lfg:             lfg,
		quantizedCoeffs: quantizedCoeffs,
		dequantHFCoeff:  dequantHFCoeff,
		groupPos:        util.Point{X: 0, Y: 0},
		blocks:          []*util.Point{{X: 0, Y: 0}},
	}
}

// ---- dequantizeHFCoefficients: flip behavior tests ----

func TestDequantizeHFCoefficients_NoFlipHornuss(t *testing.T) {
	// HORNUSS has flip()=false (transformMethod=METHOD_HORNUSS, not METHOD_DCT)
	// With flip=false: wy=y, wx=x^y^y=x → weight access is w3[y][x] (normal)
	hf := makeHFCoeffForDequantWithTT(HORNUSS, 1)
	oim := colour.NewOpsinInverseMatrix()

	weights := hf.frame.getHFGlobal().weights
	weights[HORNUSS.parameterIndex][0][0][1] = 2.5

	hf.quantizedCoeffs[0][0][1] = 5

	err := hf.dequantizeHFCoefficients()
	require.NoError(t, err)

	quant5 := float32(5) - oim.QuantBiasNumerator/float32(5)
	expected := quant5 * 1.0 * 2.5 // sfc=1.0, weight=2.5
	assert.InDelta(t, float64(expected), float64(hf.dequantHFCoeff[0][0][1]), 0.001)
}

func TestDequantizeHFCoefficients_FlipVsNoFlipWeightAccess(t *testing.T) {
	// At pixel (y=0, x=1):
	//   DCT8 (flip=true):    wy=x=1, wx=x^y^wy=0 → w3[1][0]
	//   HORNUSS (flip=false): wy=y=0, wx=x^y^wy=1 → w3[0][1]
	oim := colour.NewOpsinInverseMatrix()
	quant5 := float32(5) - oim.QuantBiasNumerator/float32(5)

	// DCT8 (flip=true): at (y=0,x=1) uses w3[1][0]
	hfFlip := makeHFCoeffForDequant(1)
	wFlip := hfFlip.frame.getHFGlobal().weights
	wFlip[0][0][1][0] = 3.0 // w3[1][0] used by DCT8
	wFlip[0][0][0][1] = 7.0 // w3[0][1] NOT used by DCT8 at this position
	hfFlip.quantizedCoeffs[0][0][1] = 5

	err := hfFlip.dequantizeHFCoefficients()
	require.NoError(t, err)
	assert.InDelta(t, float64(quant5*3.0), float64(hfFlip.dequantHFCoeff[0][0][1]), 0.001,
		"DCT8 (flip=true) should use weight at w3[1][0]")

	// HORNUSS (flip=false): at (y=0,x=1) uses w3[0][1]
	hfNoFlip := makeHFCoeffForDequantWithTT(HORNUSS, 1)
	wNoFlip := hfNoFlip.frame.getHFGlobal().weights
	wNoFlip[HORNUSS.parameterIndex][0][1][0] = 3.0 // NOT used by HORNUSS at this position
	wNoFlip[HORNUSS.parameterIndex][0][0][1] = 7.0 // w3[0][1] used by HORNUSS
	hfNoFlip.quantizedCoeffs[0][0][1] = 5

	err = hfNoFlip.dequantizeHFCoefficients()
	require.NoError(t, err)
	assert.InDelta(t, float64(quant5*7.0), float64(hfNoFlip.dequantHFCoeff[0][0][1]), 0.001,
		"HORNUSS (flip=false) should use weight at w3[0][1]")
}

// ---- dequantizeHFCoefficients: upsampling skip test ----

func TestDequantizeHFCoefficients_JpegUpsamplingSkipsChannel(t *testing.T) {
	// Block at pos (1,0): groupY=1. With jpegUpsamplingY[2]=1:
	// Channel 2: sGroupY=1>>1=0, sGroupY<<1=0 != 1 → SKIP
	// Channels 0,1: sGroupY=1>>0=1, sGroupY<<0=1 == 1 → processed
	tt := DCT8

	dctSelect := make([][]*TransformType, 2)
	dctSelect[0] = make([]*TransformType, 1)
	dctSelect[1] = make([]*TransformType, 1)
	dctSelect[1][0] = tt

	hfMultiplier := util.MakeMatrix2D[int32](2, 1)
	hfMultiplier[1][0] = 1

	hfMetadata := &HFMetadata{
		dctSelect:    dctSelect,
		hfMultiplier: hfMultiplier,
		blockList:    []util.Point{{X: 0, Y: 1}},
	}

	lfg := &LFGroup{
		hfMetadata: hfMetadata,
		lfCoeff: &LFCoefficients{
			dequantLFCoeff: util.MakeMatrix3D[float32](3, 2, 1),
			lfIndex:        util.MakeMatrix2D[int32](2, 1),
		},
	}

	weights := make([][][][]float32, 1)
	weights[0] = make([][][]float32, 3)
	for c := 0; c < 3; c++ {
		weights[0][c] = util.MakeMatrix2D[float32](8, 8)
		for y := 0; y < 8; y++ {
			for x := 0; x < 8; x++ {
				weights[0][c][y][x] = 1.0
			}
		}
	}

	ff := &FakeFramer{
		header: &FrameHeader{
			jpegUpsamplingX: []int32{0, 0, 0},
			jpegUpsamplingY: []int32{0, 0, 1}, // channel 2 has Y upsampling
			xqmScale:        2,
			bqmScale:        2,
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
	ff.lfGlobal.globalScale = 65536

	// Channels 0,1: pixelGroupY=8, need rows 8..15; Channel 2: skipped
	quantizedCoeffs := util.MakeMatrix3D[int32](3, 16, 8)
	dequantHFCoeff := util.MakeMatrix3D[float32](3, 16, 8)

	hf := &HFCoefficients{
		frame:           ff,
		lfg:             lfg,
		quantizedCoeffs: quantizedCoeffs,
		dequantHFCoeff:  dequantHFCoeff,
		groupPos:        util.Point{X: 0, Y: 0},
		blocks:          []*util.Point{{X: 0, Y: 1}},
	}

	// Channel 0,1: processed at pixelGroupY=8, position (8,1) is first non-dctSelect
	hf.quantizedCoeffs[0][8][1] = 5
	hf.quantizedCoeffs[1][8][1] = 5
	// Channel 2: skipped, set at a position that would be processed if not skipped
	hf.quantizedCoeffs[2][0][1] = 5

	err := hf.dequantizeHFCoefficients()
	require.NoError(t, err)

	// Channels 0,1 should be dequantized
	assert.NotEqual(t, float32(0.0), hf.dequantHFCoeff[0][8][1])
	assert.NotEqual(t, float32(0.0), hf.dequantHFCoeff[1][8][1])
	// Channel 2 should NOT be dequantized (skipped due to upsampling mismatch)
	assert.Equal(t, float32(0.0), hf.dequantHFCoeff[2][0][1])
}

// ---- dequantizeHFCoefficients: multiple blocks test ----

func TestDequantizeHFCoefficients_MultipleBlocks(t *testing.T) {
	tt := DCT8

	dctSelect := make([][]*TransformType, 1)
	dctSelect[0] = make([]*TransformType, 2)
	dctSelect[0][0] = tt
	dctSelect[0][1] = tt

	hfMultiplier := util.MakeMatrix2D[int32](1, 2)
	hfMultiplier[0][0] = 1
	hfMultiplier[0][1] = 1

	hfMetadata := &HFMetadata{
		dctSelect:    dctSelect,
		hfMultiplier: hfMultiplier,
		blockList:    []util.Point{{X: 0, Y: 0}, {X: 1, Y: 0}},
	}

	lfg := &LFGroup{
		hfMetadata: hfMetadata,
		lfCoeff: &LFCoefficients{
			dequantLFCoeff: util.MakeMatrix3D[float32](3, 1, 2),
			lfIndex:        util.MakeMatrix2D[int32](1, 2),
		},
	}

	weights := make([][][][]float32, 1)
	weights[0] = make([][][]float32, 3)
	for c := 0; c < 3; c++ {
		weights[0][c] = util.MakeMatrix2D[float32](8, 8)
		for y := 0; y < 8; y++ {
			for x := 0; x < 8; x++ {
				weights[0][c][y][x] = 1.0
			}
		}
	}

	ff := &FakeFramer{
		header: &FrameHeader{
			jpegUpsamplingX: []int32{0, 0, 0},
			jpegUpsamplingY: []int32{0, 0, 0},
			xqmScale:        2,
			bqmScale:        2,
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
	ff.lfGlobal.globalScale = 65536

	// Block 0: columns 0-7, Block 1: columns 8-15
	quantizedCoeffs := util.MakeMatrix3D[int32](3, 8, 16)
	dequantHFCoeff := util.MakeMatrix3D[float32](3, 8, 16)

	hf := &HFCoefficients{
		frame:           ff,
		lfg:             lfg,
		quantizedCoeffs: quantizedCoeffs,
		dequantHFCoeff:  dequantHFCoeff,
		groupPos:        util.Point{X: 0, Y: 0},
		blocks:          []*util.Point{{X: 0, Y: 0}, {X: 1, Y: 0}},
	}

	hf.quantizedCoeffs[0][0][1] = 5  // block 0: position (0,1)
	hf.quantizedCoeffs[0][0][9] = 10 // block 1: pixelGroupX=8 + x=1

	err := hf.dequantizeHFCoefficients()
	require.NoError(t, err)

	oim := colour.NewOpsinInverseMatrix()
	quant5 := float32(5) - oim.QuantBiasNumerator/float32(5)
	quant10 := float32(10) - oim.QuantBiasNumerator/float32(10)

	assert.InDelta(t, float64(quant5), float64(hf.dequantHFCoeff[0][0][1]), 0.001,
		"block 0 coefficient should be dequantized")
	assert.InDelta(t, float64(quant10), float64(hf.dequantHFCoeff[0][0][9]), 0.001,
		"block 1 coefficient should be dequantized")
}

// ---- dequantizeHFCoefficients: multiple positions test ----

func TestDequantizeHFCoefficients_MultiplePositions(t *testing.T) {
	// Verify several positions across a full DCT8 block
	hf := makeHFCoeffForDequant(1)
	oim := colour.NewOpsinInverseMatrix()

	type posCoeff struct {
		y, x  int
		coeff int32
	}
	positions := []posCoeff{
		{0, 1, 3},    // large positive
		{0, 7, -5},   // large negative
		{1, 0, 2},    // boundary (just outside [-1,1])
		{3, 4, 10},   // interior large
		{7, 7, -1},   // small negative
		{4, 2, 0},    // zero
		{6, 6, 1},    // small positive
	}

	for _, p := range positions {
		hf.quantizedCoeffs[0][p.y][p.x] = p.coeff
	}

	err := hf.dequantizeHFCoefficients()
	require.NoError(t, err)

	for _, p := range positions {
		var expectedQuant float32
		if p.coeff > -2 && p.coeff < 2 {
			switch p.coeff {
			case -1:
				expectedQuant = -oim.QuantBias[0]
			case 0:
				expectedQuant = 0
			case 1:
				expectedQuant = oim.QuantBias[0]
			}
		} else {
			expectedQuant = float32(p.coeff) - oim.QuantBiasNumerator/float32(p.coeff)
		}
		assert.InDelta(t, float64(expectedQuant), float64(hf.dequantHFCoeff[0][p.y][p.x]), 0.0001,
			"position (%d,%d) coeff=%d", p.y, p.x, p.coeff)
	}

	// Verify (0,0) is still 0 (skipped dctSelect region)
	assert.Equal(t, float32(0.0), hf.dequantHFCoeff[0][0][0])
}

// ---- chromaFromLuma: Y upsampling early return ----

func TestChromaFromLuma_ReturnsEarlyWithYUpsampling(t *testing.T) {
	hf := makeHFCoeffForChroma()
	hf.frame.getFrameHeader().jpegUpsamplingY[2] = 1

	hf.dequantHFCoeff[1][0][0] = 10.0
	hf.dequantHFCoeff[0][0][0] = 5.0
	hf.dequantHFCoeff[2][0][0] = 7.0

	err := hf.chromaFromLuma()
	require.NoError(t, err)

	// Should return early, no correlation applied
	assert.Equal(t, float32(5.0), hf.dequantHFCoeff[0][0][0])
	assert.Equal(t, float32(7.0), hf.dequantHFCoeff[2][0][0])
}

// ---- chromaFromLuma: multiple blocks ----

func TestChromaFromLuma_MultipleBlocks(t *testing.T) {
	tt := DCT8

	dctSelect := make([][]*TransformType, 1)
	dctSelect[0] = make([]*TransformType, 2)
	dctSelect[0][0] = tt
	dctSelect[0][1] = tt

	hfStreamBuffer := make([][][]int32, 2)
	hfStreamBuffer[0] = util.MakeMatrix2D[int32](1, 1)
	hfStreamBuffer[1] = util.MakeMatrix2D[int32](1, 1)
	hfStreamBuffer[0][0][0] = 42  // xFactor
	hfStreamBuffer[1][0][0] = -42 // bFactor

	hfMetadata := &HFMetadata{
		dctSelect:      dctSelect,
		hfMultiplier:   util.MakeMatrix2D[int32](1, 2),
		blockList:      []util.Point{{X: 0, Y: 0}, {X: 1, Y: 0}},
		hfStreamBuffer: hfStreamBuffer,
	}

	lfg := &LFGroup{
		hfMetadata: hfMetadata,
	}

	ff := &FakeFramer{
		header: &FrameHeader{
			jpegUpsamplingX: []int32{0, 0, 0},
			jpegUpsamplingY: []int32{0, 0, 0},
			passes:          NewPassesInfo(),
			Encoding:        VARDCT,
		},
		lfGlobal: NewLFGlobal(),
	}
	ff.lfGlobal.lfChanCorr = &LFChannelCorrelation{
		colorFactor:      84,
		baseCorrelationX: 0.0,
		baseCorrelationB: 1.0,
		xFactorLF:        128,
		bFactorLF:        128,
	}

	// 2 blocks: 8 rows x 16 columns
	dequantHFCoeff := util.MakeMatrix3D[float32](3, 8, 16)

	hf := &HFCoefficients{
		frame:          ff,
		lfg:            lfg,
		dequantHFCoeff: dequantHFCoeff,
		groupPos:       util.Point{X: 0, Y: 0},
		blocks:         []*util.Point{{X: 0, Y: 0}, {X: 1, Y: 0}},
	}

	// Set Y values in both blocks
	hf.dequantHFCoeff[1][0][0] = 10.0 // block 0
	hf.dequantHFCoeff[1][0][8] = 20.0 // block 1

	lfc := ff.lfGlobal.lfChanCorr
	expectedKX := lfc.baseCorrelationX + float32(42)/float32(lfc.colorFactor)
	expectedKB := lfc.baseCorrelationB + float32(-42)/float32(lfc.colorFactor)

	err := hf.chromaFromLuma()
	require.NoError(t, err)

	// Block 0: X[0][0] += kX * Y[0][0]
	assert.InDelta(t, float64(expectedKX)*10.0, float64(hf.dequantHFCoeff[0][0][0]), 0.001)
	// Block 1: X[0][8] += kX * Y[0][8]
	assert.InDelta(t, float64(expectedKX)*20.0, float64(hf.dequantHFCoeff[0][0][8]), 0.001)
	// Block 0: B[0][0] += kB * Y[0][0]
	assert.InDelta(t, float64(expectedKB)*10.0, float64(hf.dequantHFCoeff[2][0][0]), 0.001)
	// Block 1: B[0][8] += kB * Y[0][8]
	assert.InDelta(t, float64(expectedKB)*20.0, float64(hf.dequantHFCoeff[2][0][8]), 0.001)
}

// ---- finalizeLLF tests ----

func TestFinalizeLLF_NilBlockSkipped(t *testing.T) {
	hf := makeHFCoeffForDequant(1)
	hf.blocks = []*util.Point{nil}

	hf.lfg.lfCoeff.dequantLFCoeff[0][0][0] = 5.0
	// Pre-set dequant value to verify it's not overwritten
	hf.dequantHFCoeff[0][0][0] = 99.0

	err := hf.finalizeLLF()
	require.NoError(t, err)

	// With nil block, nothing should change
	assert.Equal(t, float32(99.0), hf.dequantHFCoeff[0][0][0])
}

func TestFinalizeLLF_CopiesLFCoeff(t *testing.T) {
	hf := makeHFCoeffForDequant(1)

	// For DCT8: dctSelectSize=(1,1), ForwardDCT2D copies single value
	// llfScale[0][0] = 1.0 for DCT8
	hf.lfg.lfCoeff.dequantLFCoeff[0][0][0] = 5.0
	hf.lfg.lfCoeff.dequantLFCoeff[1][0][0] = 3.0
	hf.lfg.lfCoeff.dequantLFCoeff[2][0][0] = -2.0

	err := hf.finalizeLLF()
	require.NoError(t, err)

	// ForwardDCT2D(1x1) copies value, then *= llfScale[0][0] = 1.0
	assert.InDelta(t, 5.0, float64(hf.dequantHFCoeff[0][0][0]), 0.001)
	assert.InDelta(t, 3.0, float64(hf.dequantHFCoeff[1][0][0]), 0.001)
	assert.InDelta(t, -2.0, float64(hf.dequantHFCoeff[2][0][0]), 0.001)
}

func TestFinalizeLLF_LlfScaleMultiplication(t *testing.T) {
	// Use a custom TransformType with non-trivial llfScale to verify multiplication
	ttCopy := *DCT8
	ttCopy.llfScale = [][]float32{{2.5}}

	hf := makeHFCoeffForDequantWithTT(&ttCopy, 1)

	hf.lfg.lfCoeff.dequantLFCoeff[0][0][0] = 4.0
	hf.lfg.lfCoeff.dequantLFCoeff[1][0][0] = -3.0
	hf.lfg.lfCoeff.dequantLFCoeff[2][0][0] = 10.0

	err := hf.finalizeLLF()
	require.NoError(t, err)

	// ForwardDCT2D(1x1) copies value, then *= llfScale[0][0] = 2.5
	assert.InDelta(t, 10.0, float64(hf.dequantHFCoeff[0][0][0]), 0.001)
	assert.InDelta(t, -7.5, float64(hf.dequantHFCoeff[1][0][0]), 0.001)
	assert.InDelta(t, 25.0, float64(hf.dequantHFCoeff[2][0][0]), 0.001)
}

func TestFinalizeLLF_UpsamplingSkipsChannel(t *testing.T) {
	// Block at pos (1,0): groupY=1. With jpegUpsamplingY[2]=1:
	// Channel 2: skipped (sGroupY=0, 0<<1=0 != 1)
	// Channels 0,1: processed (sGroupY=1, pixelGroupY=8)
	tt := DCT8

	dctSelect := make([][]*TransformType, 2)
	dctSelect[0] = make([]*TransformType, 1)
	dctSelect[1] = make([]*TransformType, 1)
	dctSelect[1][0] = tt

	hfMetadata := &HFMetadata{
		dctSelect:    dctSelect,
		hfMultiplier: util.MakeMatrix2D[int32](2, 1),
		blockList:    []util.Point{{X: 0, Y: 1}},
	}

	lfg := &LFGroup{
		hfMetadata: hfMetadata,
		lfCoeff: &LFCoefficients{
			dequantLFCoeff: util.MakeMatrix3D[float32](3, 2, 1),
			lfIndex:        util.MakeMatrix2D[int32](2, 1),
		},
	}

	ff := &FakeFramer{
		header: &FrameHeader{
			jpegUpsamplingX: []int32{0, 0, 0},
			jpegUpsamplingY: []int32{0, 0, 1}, // channel 2 has Y upsampling
			passes:          NewPassesInfo(),
			Encoding:        VARDCT,
		},
		lfGlobal: NewLFGlobal(),
	}

	dequantHFCoeff := util.MakeMatrix3D[float32](3, 16, 8)

	hf := &HFCoefficients{
		frame:          ff,
		lfg:            lfg,
		dequantHFCoeff: dequantHFCoeff,
		groupPos:       util.Point{X: 0, Y: 0},
		blocks:         []*util.Point{{X: 0, Y: 1}},
	}

	// Set LF coefficients for channels 0,1 at sLfgY=1
	lfg.lfCoeff.dequantLFCoeff[0][1][0] = 5.0
	lfg.lfCoeff.dequantLFCoeff[1][1][0] = 3.0
	// Channel 2 marker (should remain unchanged)
	hf.dequantHFCoeff[2][0][0] = 99.0

	err := hf.finalizeLLF()
	require.NoError(t, err)

	// Channels 0,1: ForwardDCT2D copies LF coeff to pixelGroupY=8
	assert.InDelta(t, 5.0, float64(hf.dequantHFCoeff[0][8][0]), 0.001)
	assert.InDelta(t, 3.0, float64(hf.dequantHFCoeff[1][8][0]), 0.001)
	// Channel 2: skipped, marker value unchanged
	assert.Equal(t, float32(99.0), hf.dequantHFCoeff[2][0][0])
}

// ---- NewHFCoefficientsWithReader tests ----

// makeFramerForNewHFCoeff creates a FakeFramer suitable for testing NewHFCoefficientsWithReader.
// The contextStream uses a zero-value entropy.EntropyStream, so any ReadSymbol call will fail.
// This makes it suitable for tests where no blocks are within the group range.
func makeFramerForNewHFCoeff(blockList []util.Point) *FakeFramer {
	contextStream := &entropy.EntropyStream{}
	hfPass := &HFPass{
		contextStream: contextStream,
		order:         make([][][]util.Point, 13),
	}
	for i := 0; i < 13; i++ {
		hfPass.order[i] = make([][]util.Point, 3)
	}

	// Size arrays to cover all block positions in the LFGroup
	rows, cols := 1, 1
	for _, p := range blockList {
		if int(p.Y)+1 > rows {
			rows = int(p.Y) + 1
		}
		if int(p.X)+1 > cols {
			cols = int(p.X) + 1
		}
	}

	dctSelect := make([][]*TransformType, rows)
	for i := 0; i < rows; i++ {
		dctSelect[i] = make([]*TransformType, cols)
		for j := 0; j < cols; j++ {
			dctSelect[i][j] = DCT8
		}
	}

	hfMultiplier := util.MakeMatrix2D[int32](rows, cols)
	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			hfMultiplier[i][j] = 1
		}
	}

	hfMetadata := &HFMetadata{
		blockList:    blockList,
		dctSelect:    dctSelect,
		hfMultiplier: hfMultiplier,
	}

	lfCoeff := &LFCoefficients{
		lfIndex: util.MakeMatrix2D[int32](rows, cols),
	}

	lfGroup := &LFGroup{
		hfMetadata: hfMetadata,
		lfCoeff:    lfCoeff,
	}

	hfBlockCtx := &HFBlockContext{
		numClusters:   1,
		qfThresholds:  []int32{},
		numLFContexts: 1,
		clusterMap:    []int{0},
	}

	ff := &FakeFramer{
		header: &FrameHeader{
			jpegUpsamplingX: []int32{0, 0, 0},
			jpegUpsamplingY: []int32{0, 0, 0},
			passes: &PassesInfo{
				shift:     []uint32{0},
				numPasses: 1,
			},
			Encoding: VARDCT,
		},
		lfGlobal: &LFGlobal{
			hfBlockCtx: hfBlockCtx,
		},
		hfGlobal: &HFGlobal{
			numHFPresets: 1,
		},
		lfGroup:                lfGroup,
		passes:                 []Pass{{hfPass: hfPass}},
		groupSize:              &util.Dimension{Width: 256, Height: 256},
		groupPosInLFGroupPoint: &util.Point{X: 0, Y: 0},
	}

	return ff
}

func TestNewHFCoefficientsWithReader_GetGroupSizeError(t *testing.T) {
	ff := makeFramerForNewHFCoeff([]util.Point{})
	ff.groupSizeError = errors.New("group size error")

	reader := &testcommon.FakeBitReader{
		ReadBitsData: []uint64{0},
	}

	hf, err := NewHFCoefficientsWithReader(reader, ff, 0, 0)
	assert.Error(t, err)
	assert.Nil(t, hf)
	assert.Contains(t, err.Error(), "group size error")
}

func TestNewHFCoefficientsWithReader_EmptyBlockList(t *testing.T) {
	ff := makeFramerForNewHFCoeff([]util.Point{})

	reader := &testcommon.FakeBitReader{
		ReadBitsData: []uint64{0},
	}

	hf, err := NewHFCoefficientsWithReader(reader, ff, 0, 0)
	require.NoError(t, err)
	require.NotNil(t, hf)

	assert.Equal(t, int32(0), hf.hfPreset)
	assert.Equal(t, uint32(0), hf.groupID)
	assert.Equal(t, 0, len(hf.blocks))
	assert.NotNil(t, hf.stream)
	assert.NotNil(t, hf.quantizedCoeffs)
	assert.NotNil(t, hf.dequantHFCoeff)
}

func TestNewHFCoefficientsWithReader_VerifyFieldsSet(t *testing.T) {
	// Block at (50,50) is far outside 32x32 group range
	blockList := []util.Point{{X: 50, Y: 50}}
	ff := makeFramerForNewHFCoeff(blockList)

	reader := &testcommon.FakeBitReader{
		ReadBitsData: []uint64{0},
	}

	hf, err := NewHFCoefficientsWithReader(reader, ff, 0, 5)
	require.NoError(t, err)
	require.NotNil(t, hf)

	assert.Equal(t, int32(0), hf.hfPreset)
	assert.Equal(t, uint32(5), hf.groupID)
	assert.NotNil(t, hf.hfctx)
	assert.NotNil(t, hf.lfg)
	assert.NotNil(t, hf.stream)
	assert.NotNil(t, hf.quantizedCoeffs)
	assert.NotNil(t, hf.dequantHFCoeff)
	assert.Equal(t, 1, len(hf.blocks))
	// Block at (50,50) is outside group range → nil
	assert.Nil(t, hf.blocks[0])
}

func TestNewHFCoefficientsWithReader_AllBlocksOutsideGroup(t *testing.T) {
	// groupPos = (0,0). Blocks at Y>=32 or X>=32 are outside
	blockList := []util.Point{
		{X: 0, Y: 32},
		{X: 32, Y: 0},
		{X: 32, Y: 32},
	}
	ff := makeFramerForNewHFCoeff(blockList)

	reader := &testcommon.FakeBitReader{
		ReadBitsData: []uint64{0},
	}

	hf, err := NewHFCoefficientsWithReader(reader, ff, 0, 0)
	require.NoError(t, err)
	require.NotNil(t, hf)

	assert.Equal(t, 3, len(hf.blocks))
	for i, b := range hf.blocks {
		assert.Nil(t, b, "block %d should be nil (outside range)", i)
	}
}

func TestNewHFCoefficientsWithReader_BufferSizingNoUpsampling(t *testing.T) {
	ff := makeFramerForNewHFCoeff([]util.Point{})
	ff.groupSize = &util.Dimension{Width: 128, Height: 64}

	reader := &testcommon.FakeBitReader{
		ReadBitsData: []uint64{0},
	}

	hf, err := NewHFCoefficientsWithReader(reader, ff, 0, 0)
	require.NoError(t, err)

	// No upsampling: all channels have same dimensions as group size
	for c := 0; c < 3; c++ {
		assert.Equal(t, 64, len(hf.quantizedCoeffs[c]), "channel %d quantized height", c)
		assert.Equal(t, 128, len(hf.quantizedCoeffs[c][0]), "channel %d quantized width", c)
		assert.Equal(t, 64, len(hf.dequantHFCoeff[c]), "channel %d dequant height", c)
		assert.Equal(t, 128, len(hf.dequantHFCoeff[c][0]), "channel %d dequant width", c)
	}
}

func TestNewHFCoefficientsWithReader_BufferSizingWithUpsampling(t *testing.T) {
	ff := makeFramerForNewHFCoeff([]util.Point{})
	ff.groupSize = &util.Dimension{Width: 128, Height: 64}
	ff.header.jpegUpsamplingY[2] = 1 // channel 2 height halved
	ff.header.jpegUpsamplingX[0] = 1 // channel 0 width halved

	reader := &testcommon.FakeBitReader{
		ReadBitsData: []uint64{0},
	}

	hf, err := NewHFCoefficientsWithReader(reader, ff, 0, 0)
	require.NoError(t, err)

	// Channel 0: height=64, width=128>>1=64
	assert.Equal(t, 64, len(hf.quantizedCoeffs[0]))
	assert.Equal(t, 64, len(hf.quantizedCoeffs[0][0]))

	// Channel 1: height=64, width=128 (no upsampling)
	assert.Equal(t, 64, len(hf.quantizedCoeffs[1]))
	assert.Equal(t, 128, len(hf.quantizedCoeffs[1][0]))

	// Channel 2: height=64>>1=32, width=128
	assert.Equal(t, 32, len(hf.quantizedCoeffs[2]))
	assert.Equal(t, 128, len(hf.quantizedCoeffs[2][0]))
}

func TestNewHFCoefficientsWithReader_GroupPosShiftedBy5(t *testing.T) {
	ff := makeFramerForNewHFCoeff([]util.Point{})
	ff.groupPosInLFGroupPoint = &util.Point{X: 3, Y: 2}

	reader := &testcommon.FakeBitReader{
		ReadBitsData: []uint64{0},
	}

	hf, err := NewHFCoefficientsWithReader(reader, ff, 0, 0)
	require.NoError(t, err)

	// groupPos = groupPosInLFGroup << 5
	assert.Equal(t, int32(2<<5), hf.groupPos.Y) // 64
	assert.Equal(t, int32(3<<5), hf.groupPos.X) // 96
}

func TestNewHFCoefficientsWithReader_HfPresetReadValue(t *testing.T) {
	ff := makeFramerForNewHFCoeff([]util.Point{})
	ff.hfGlobal.numHFPresets = 8 // CeilLog1p(7) = 3 → ReadBits(3)

	reader := &testcommon.FakeBitReader{
		ReadBitsData: []uint64{5}, // hfPreset = 5
	}

	hf, err := NewHFCoefficientsWithReader(reader, ff, 0, 0)
	require.NoError(t, err)

	assert.Equal(t, int32(5), hf.hfPreset)
}

func TestNewHFCoefficientsWithReader_GroupIDPropagated(t *testing.T) {
	ff := makeFramerForNewHFCoeff([]util.Point{})

	reader := &testcommon.FakeBitReader{
		ReadBitsData: []uint64{0},
	}

	hf, err := NewHFCoefficientsWithReader(reader, ff, 0, 7)
	require.NoError(t, err)

	assert.Equal(t, uint32(7), hf.groupID)
}

func TestNewHFCoefficientsWithReader_BlockBoundaryOutside(t *testing.T) {
	// groupPos = (0, 0). All blocks at boundary or beyond are filtered out
	blockList := []util.Point{
		{X: 32, Y: 0},  // groupX=32 >= 32: OUTSIDE
		{X: 0, Y: 32},  // groupY=32 >= 32: OUTSIDE
		{X: 32, Y: 32}, // both >= 32: OUTSIDE
	}
	ff := makeFramerForNewHFCoeff(blockList)

	reader := &testcommon.FakeBitReader{
		ReadBitsData: []uint64{0},
	}

	hf, err := NewHFCoefficientsWithReader(reader, ff, 0, 0)
	require.NoError(t, err)
	require.NotNil(t, hf)

	assert.Equal(t, 3, len(hf.blocks))
	for i, b := range hf.blocks {
		assert.Nil(t, b, "block %d should be nil (outside range)", i)
	}
}

func TestNewHFCoefficientsWithReader_NegativeGroupPosFiltersBlocks(t *testing.T) {
	// groupPosInLFGroup=(2,2) → groupPos=(64,64)
	// Block at (0,0): groupY=0-64=-64 < 0 → outside
	blockList := []util.Point{{X: 0, Y: 0}}
	ff := makeFramerForNewHFCoeff(blockList)
	ff.groupPosInLFGroupPoint = &util.Point{X: 2, Y: 2}

	reader := &testcommon.FakeBitReader{
		ReadBitsData: []uint64{0},
	}

	hf, err := NewHFCoefficientsWithReader(reader, ff, 0, 0)
	require.NoError(t, err)
	require.NotNil(t, hf)

	assert.Equal(t, 1, len(hf.blocks))
	assert.Nil(t, hf.blocks[0])
	assert.Equal(t, int32(64), hf.groupPos.Y)
	assert.Equal(t, int32(64), hf.groupPos.X)
}

func TestNewHFCoefficientsWithReader_BlockInsideGroupTriggersRead(t *testing.T) {
	// Block at (0,0) is inside the 32x32 group range → ReadSymbol called
	// Zero-value entropy stream causes "Context cannot be bigger" error
	blockList := []util.Point{{X: 0, Y: 0}}
	ff := makeFramerForNewHFCoeff(blockList)

	reader := &testcommon.FakeBitReader{
		ReadBitsData: []uint64{0},
	}

	hf, err := NewHFCoefficientsWithReader(reader, ff, 0, 0)
	assert.Error(t, err)
	assert.Nil(t, hf)
	assert.Contains(t, err.Error(), "Context")
}

func TestNewHFCoefficientsWithReader_BlockAt31IsInside(t *testing.T) {
	// Block at (31,31) is at the max valid position (31 < 32) → inside → triggers read
	blockList := []util.Point{{X: 31, Y: 31}}
	ff := makeFramerForNewHFCoeff(blockList)

	reader := &testcommon.FakeBitReader{
		ReadBitsData: []uint64{0},
	}

	hf, err := NewHFCoefficientsWithReader(reader, ff, 0, 0)
	assert.Error(t, err)
	assert.Nil(t, hf)
	// Confirms block at (31,31) was NOT filtered out
	assert.Contains(t, err.Error(), "Context")
}

func TestNewHFCoefficientsWithReader_HfctxAndLfgSet(t *testing.T) {
	ff := makeFramerForNewHFCoeff([]util.Point{})

	reader := &testcommon.FakeBitReader{
		ReadBitsData: []uint64{0},
	}

	hf, err := NewHFCoefficientsWithReader(reader, ff, 0, 0)
	require.NoError(t, err)

	// Verify hfctx comes from lfGlobal.hfBlockCtx
	assert.Same(t, ff.lfGlobal.hfBlockCtx, hf.hfctx)
	// Verify lfg comes from getLFGroupForGroup
	assert.Same(t, ff.lfGroup, hf.lfg)
}

// makeFramerForNewHFCoeffWithStream creates a FakeFramer with a controllable entropy stream
// for testing the full coefficient-reading path in NewHFCoefficientsWithReader.
// orderEntries controls how many entries per channel in the DCT8 order (1 LLF + N-1 HF).
func makeFramerForNewHFCoeffWithStream(blockList []util.Point, dist *entropy.FakeSymbolDistribution, orderEntries int) *FakeFramer {
	contextStream := entropy.NewEntropyStreamForTest(495, dist)

	hfPass := &HFPass{
		contextStream: contextStream,
		order:         make([][][]util.Point, 13),
	}
	for i := 0; i < 13; i++ {
		hfPass.order[i] = make([][]util.Point, 3)
	}
	// Set up order for DCT8 (orderID=0)
	for c := 0; c < 3; c++ {
		order := make([]util.Point, orderEntries)
		order[0] = util.Point{X: 0, Y: 0} // LLF entry
		for j := 1; j < orderEntries; j++ {
			order[j] = util.Point{X: int32(j), Y: 0} // HF entries
		}
		hfPass.order[0][c] = order
	}

	rows, cols := 1, 1
	for _, p := range blockList {
		if int(p.Y)+1 > rows {
			rows = int(p.Y) + 1
		}
		if int(p.X)+1 > cols {
			cols = int(p.X) + 1
		}
	}

	dctSelect := make([][]*TransformType, rows)
	for i := 0; i < rows; i++ {
		dctSelect[i] = make([]*TransformType, cols)
		for j := 0; j < cols; j++ {
			dctSelect[i][j] = DCT8
		}
	}

	hfMultiplier := util.MakeMatrix2D[int32](rows, cols)
	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			hfMultiplier[i][j] = 1
		}
	}

	hfMetadata := &HFMetadata{
		blockList:    blockList,
		dctSelect:    dctSelect,
		hfMultiplier: hfMultiplier,
	}

	lfCoeff := &LFCoefficients{
		lfIndex: util.MakeMatrix2D[int32](rows, cols),
	}

	lfGroup := &LFGroup{
		hfMetadata: hfMetadata,
		lfCoeff:    lfCoeff,
	}

	// clusterMap needs at least 27 entries (c=2: idx=2*13+0=26)
	hfBlockCtxClusterMap := make([]int, 40)
	hfBlockCtx := &HFBlockContext{
		numClusters:   1,
		qfThresholds:  []int32{},
		numLFContexts: 1,
		clusterMap:    hfBlockCtxClusterMap,
	}

	ff := &FakeFramer{
		header: &FrameHeader{
			jpegUpsamplingX: []int32{0, 0, 0},
			jpegUpsamplingY: []int32{0, 0, 0},
			passes: &PassesInfo{
				shift:     []uint32{0},
				numPasses: 1,
			},
			Encoding: VARDCT,
		},
		lfGlobal: &LFGlobal{
			hfBlockCtx: hfBlockCtx,
		},
		hfGlobal: &HFGlobal{
			numHFPresets: 1,
		},
		lfGroup:                lfGroup,
		passes:                 []Pass{{hfPass: hfPass}},
		groupSize:              &util.Dimension{Width: 256, Height: 256},
		groupPosInLFGroupPoint: &util.Point{X: 0, Y: 0},
	}

	return ff
}

func TestNewHFCoefficientsWithReader_NonZeroZeroAllChannels(t *testing.T) {
	// Block at (0,0) is inside group. All channels get nonZero=0 → no coefficient reads.
	// cMap order: c=1, c=0, c=2
	dist := &entropy.FakeSymbolDistribution{
		Symbols: []int32{0, 0, 0}, // nonZero=0 for each channel
		Cfg:     entropy.NewHybridIntegerConfig(15, 0, 0),
	}
	ff := makeFramerForNewHFCoeffWithStream([]util.Point{{X: 0, Y: 0}}, dist, 2)

	reader := &testcommon.FakeBitReader{
		ReadBitsData: []uint64{0}, // hfPreset
	}

	hf, err := NewHFCoefficientsWithReader(reader, ff, 0, 0)
	require.NoError(t, err)
	require.NotNil(t, hf)

	// Block was included
	assert.NotNil(t, hf.blocks[0])

	// All quantizedCoeffs should remain zero
	for c := 0; c < 3; c++ {
		for y := 0; y < len(hf.quantizedCoeffs[c]); y++ {
			for x := 0; x < len(hf.quantizedCoeffs[c][y]); x++ {
				assert.Equal(t, int32(0), hf.quantizedCoeffs[c][y][x],
					"channel %d pos (%d,%d) should be zero", c, y, x)
			}
		}
	}
}

func TestNewHFCoefficientsWithReader_SingleCoeffPerChannel(t *testing.T) {
	// Block at (0,0), nonZero=1 per channel, one non-zero coeff each.
	// cMap order: c=1, c=0, c=2
	// Order entry [1] = {X:1, Y:0}. With flip=true (DCT8): posY=1, posX=0.
	// UnpackSigned(2)=1, UnpackSigned(4)=2, UnpackSigned(6)=3
	dist := &entropy.FakeSymbolDistribution{
		Symbols: []int32{
			1, 2, // c=1: nonZero=1, coeff uc=2
			1, 4, // c=0: nonZero=1, coeff uc=4
			1, 6, // c=2: nonZero=1, coeff uc=6
		},
		Cfg: entropy.NewHybridIntegerConfig(15, 0, 0),
	}
	ff := makeFramerForNewHFCoeffWithStream([]util.Point{{X: 0, Y: 0}}, dist, 2)

	reader := &testcommon.FakeBitReader{
		ReadBitsData: []uint64{0},
	}

	hf, err := NewHFCoefficientsWithReader(reader, ff, 0, 0)
	require.NoError(t, err)
	require.NotNil(t, hf)

	// With flip=true, order[1]={X:1,Y:0} → posY=1, posX=0
	assert.Equal(t, int32(1), hf.quantizedCoeffs[1][1][0]) // UnpackSigned(2)=1
	assert.Equal(t, int32(2), hf.quantizedCoeffs[0][1][0]) // UnpackSigned(4)=2
	assert.Equal(t, int32(3), hf.quantizedCoeffs[2][1][0]) // UnpackSigned(6)=3
}

func TestNewHFCoefficientsWithReader_NonZeroMismatchError(t *testing.T) {
	// nonZero=2 for c=1, but ucoeffLen=1 (2-entry order) and coeff=0 → nonZero stays 2
	dist := &entropy.FakeSymbolDistribution{
		Symbols: []int32{2, 0}, // c=1: nonZero=2, coeff uc=0 (no decrement)
		Cfg:     entropy.NewHybridIntegerConfig(15, 0, 0),
	}
	ff := makeFramerForNewHFCoeffWithStream([]util.Point{{X: 0, Y: 0}}, dist, 2)

	reader := &testcommon.FakeBitReader{
		ReadBitsData: []uint64{0},
	}

	hf, err := NewHFCoefficientsWithReader(reader, ff, 0, 0)
	assert.Error(t, err)
	assert.Nil(t, hf)
	assert.Contains(t, err.Error(), "nonZero != 0")
}

func TestNewHFCoefficientsWithReader_ShiftAppliedToCoeff(t *testing.T) {
	// shift=2: UnpackSigned(2)=1, then 1<<2=4
	dist := &entropy.FakeSymbolDistribution{
		Symbols: []int32{1, 2, 0, 0}, // c=1: nonZero=1, coeff uc=2; c=0,c=2: nonZero=0
		Cfg:     entropy.NewHybridIntegerConfig(15, 0, 0),
	}
	ff := makeFramerForNewHFCoeffWithStream([]util.Point{{X: 0, Y: 0}}, dist, 2)
	ff.header.passes.shift[0] = 2

	reader := &testcommon.FakeBitReader{
		ReadBitsData: []uint64{0},
	}

	hf, err := NewHFCoefficientsWithReader(reader, ff, 0, 0)
	require.NoError(t, err)

	assert.Equal(t, int32(4), hf.quantizedCoeffs[1][1][0]) // 1 << 2 = 4
}

func TestNewHFCoefficientsWithReader_ValidateFinalStateError(t *testing.T) {
	// ActivateStateOnRead makes ReadSymbol set HasState=true, State=0
	// After all reads, ValidateFinalState returns false (State != 0x130000)
	dist := &entropy.FakeSymbolDistribution{
		Symbols:             []int32{0, 0, 0},
		Cfg:                 entropy.NewHybridIntegerConfig(15, 0, 0),
		ActivateStateOnRead: true,
	}
	ff := makeFramerForNewHFCoeffWithStream([]util.Point{{X: 0, Y: 0}}, dist, 2)

	reader := &testcommon.FakeBitReader{
		ReadBitsData: []uint64{0},
	}

	hf, err := NewHFCoefficientsWithReader(reader, ff, 0, 0)
	assert.Error(t, err)
	assert.Nil(t, hf)
	assert.Contains(t, err.Error(), "Illegal final state")
}

func TestNewHFCoefficientsWithReader_MultipleCoeffsRead(t *testing.T) {
	// 3-entry order: ucoeffLen=2. nonZero=2, two non-zero coefficients for c=1.
	// order[1]={X:1,Y:0} → flip → posY=1,posX=0
	// order[2]={X:2,Y:0} → flip → posY=2,posX=0
	dist := &entropy.FakeSymbolDistribution{
		Symbols: []int32{
			2, 4, 6, // c=1: nonZero=2, coeff1 uc=4, coeff2 uc=6
			0, // c=0: nonZero=0
			0, // c=2: nonZero=0
		},
		Cfg: entropy.NewHybridIntegerConfig(15, 0, 0),
	}
	ff := makeFramerForNewHFCoeffWithStream([]util.Point{{X: 0, Y: 0}}, dist, 3)

	reader := &testcommon.FakeBitReader{
		ReadBitsData: []uint64{0},
	}

	hf, err := NewHFCoefficientsWithReader(reader, ff, 0, 0)
	require.NoError(t, err)

	// UnpackSigned(4)=2, UnpackSigned(6)=3
	assert.Equal(t, int32(2), hf.quantizedCoeffs[1][1][0])
	assert.Equal(t, int32(3), hf.quantizedCoeffs[1][2][0])
}

func TestNewHFCoefficientsWithReader_OddUcNegativeCoeff(t *testing.T) {
	// Odd uc values produce negative coefficients via UnpackSigned
	// UnpackSigned(3) = -(3+1)>>1 = -2
	dist := &entropy.FakeSymbolDistribution{
		Symbols: []int32{1, 3, 0, 0}, // c=1: nonZero=1, coeff uc=3; c=0,c=2: 0
		Cfg:     entropy.NewHybridIntegerConfig(15, 0, 0),
	}
	ff := makeFramerForNewHFCoeffWithStream([]util.Point{{X: 0, Y: 0}}, dist, 2)

	reader := &testcommon.FakeBitReader{
		ReadBitsData: []uint64{0},
	}

	hf, err := NewHFCoefficientsWithReader(reader, ff, 0, 0)
	require.NoError(t, err)

	assert.Equal(t, int32(-2), hf.quantizedCoeffs[1][1][0])
}
