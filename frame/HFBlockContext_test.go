package frame

import (
	"testing"

	"github.com/kpfaulkner/jxl-go/jxlio"
	"github.com/kpfaulkner/jxl-go/testcommon"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewHFBlockContextWithReader(t *testing.T) {

	for _, tc := range []struct {
		name           string
		boolData       []bool
		enumData       []int32
		u32Data        []uint32
		bitsData       []uint64
		expectedResult HFBlockContext
		expectErr      bool
	}{
		{
			name:      "no data",
			expectErr: true,
		},
		{
			name: "success use default",
			boolData: []bool{
				true,
			},
			expectedResult: HFBlockContext{
				lfThresholds:  [][]int32{{}, {}, {}},
				clusterMap:    []int{0, 1, 2, 2, 3, 3, 4, 5, 6, 6, 6, 6, 6, 7, 8, 9, 9, 10, 11, 12, 13, 14, 14, 14, 14, 14, 7, 8, 9, 9, 10, 11, 12, 13, 14, 14, 14, 14, 14},
				numClusters:   15,
				qfThresholds:  []int32{},
				numLFContexts: 1,
			},

			expectErr: false,
		},
		{
			name: "success non-default",
			u32Data: []uint32{
				1, 2, 2,
				1, // nbQfThread U32 read
			},
			bitsData: []uint64{
				0, 1, 2,
				1, // nbQfThread
			},
			boolData: []bool{
				false,
				true, // simple clustering
			},
			expectedResult: HFBlockContext{
				lfThresholds:  [][]int32{{}, {-1}, {1, 1}},
				clusterMap:    make([]int, 468),
				numClusters:   1,
				qfThresholds:  []int32{2},
				numLFContexts: 6,
			},

			expectErr: false,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {

			bitReader := &testcommon.FakeBitReader{
				ReadU32Data:  tc.u32Data,
				ReadBitsData: tc.bitsData,
				ReadBoolData: tc.boolData,
			}
			hf, err := NewHFBlockContextWithReader(bitReader, fakeReadClusterMap)
			if err != nil && !tc.expectErr {
				t.Fatalf("got error when none was expected : %v", err)
			}
			if err == nil && tc.expectErr {
				t.Fatalf("expected error but got none")
			}
			if err != nil && tc.expectErr {
				return
			}
			assert.Equal(t, tc.expectedResult, *hf)

		})
	}
}

func fakeReadClusterMap(reader jxlio.BitReader, clusterMap []int, maxClusters int) (int, error) {
	return 1, nil
}

func TestNewHFBlockContextWithReader_ErrorPaths(t *testing.T) {
	t.Run("Error on useDefault", func(t *testing.T) {
		reader := &testcommon.FakeBitReader{
			ReadBoolData: []bool{}, // Error
		}
		hf, err := NewHFBlockContextWithReader(reader, fakeReadClusterMap)
		assert.Error(t, err)
		assert.Nil(t, hf)
	})

	t.Run("Error on nbLFThresh", func(t *testing.T) {
		reader := &testcommon.FakeBitReader{
			ReadBoolData: []bool{false},
			ReadBitsData: []uint64{}, // Error on first nbLFThresh
		}
		hf, err := NewHFBlockContextWithReader(reader, fakeReadClusterMap)
		assert.Error(t, err)
		assert.Nil(t, hf)
	})

	t.Run("Error on lfThresholds value", func(t *testing.T) {
		reader := &testcommon.FakeBitReader{
			ReadBoolData: []bool{false},
			ReadBitsData: []uint64{1, 0, 0}, // nbLFThresh = [1, 0, 0]
			ReadU32Data:  []uint32{},        // Error on ReadU32
		}
		hf, err := NewHFBlockContextWithReader(reader, fakeReadClusterMap)
		assert.Error(t, err)
		assert.Nil(t, hf)
	})

	t.Run("Error on nbQfThread", func(t *testing.T) {
		reader := &testcommon.FakeBitReader{
			ReadBoolData: []bool{false},
			ReadBitsData: []uint64{0, 0, 0}, // nbLFThresh = [0, 0, 0]
			// nbQfThread read next
		}
		hf, err := NewHFBlockContextWithReader(reader, fakeReadClusterMap)
		assert.Error(t, err)
		assert.Nil(t, hf)
	})

	t.Run("Error on qfThresholds value", func(t *testing.T) {
		reader := &testcommon.FakeBitReader{
			ReadBoolData: []bool{false},
			ReadBitsData: []uint64{0, 0, 0, 1}, // nbLFThresh=[0,0,0], nbQfThread=1
			ReadU32Data:  []uint32{},           // Error on qfThresholds ReadU32
		}
		hf, err := NewHFBlockContextWithReader(reader, fakeReadClusterMap)
		assert.Error(t, err)
		assert.Nil(t, hf)
	})

	t.Run("HF block size too large", func(t *testing.T) {
		// bSize = 39 * (nbQfThread + 1) * (nbLFThresh[0] + 1) * (nbLFThresh[1] + 1) * (nbLFThresh[2] + 1)
		// Max allowed 39 * 64
		// If nbQfThread = 15, nbLFThresh = [1, 1, 1]
		// bSize = 39 * 16 * 2 * 2 * 2 = 39 * 128 > 39 * 64
		reader := &testcommon.FakeBitReader{
			ReadBoolData: []bool{false},
			ReadBitsData: []uint64{1, 1, 1, 15}, // nbLFThresh=[1,1,1], nbQfThread=15
			ReadU32Data:  []uint32{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		}
		hf, err := NewHFBlockContextWithReader(reader, fakeReadClusterMap)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "HF block size too large")
		assert.Nil(t, hf)
	})

	t.Run("Error on readClusterMap", func(t *testing.T) {
		reader := &testcommon.FakeBitReader{
			ReadBoolData: []bool{false},
			ReadBitsData: []uint64{0, 0, 0, 0}, // nbLFThresh=[0,0,0], nbQfThread=0
		}
		errReadClusterMap := func(reader jxlio.BitReader, clusterMap []int, maxClusters int) (int, error) {
			return 0, assert.AnError
		}
		hf, err := NewHFBlockContextWithReader(reader, errReadClusterMap)
		assert.Error(t, err)
		assert.Nil(t, hf)
	})
}

func TestNewHFBlockContextWithReader_UseDefault(t *testing.T) {
	// THIS TEST SHOULD FAIL IF THE BUG IS PRESENT (unless the test provides data it shouldn't read)
	reader := &testcommon.FakeBitReader{
		ReadBoolData: []bool{true},
	}
	hf, err := NewHFBlockContextWithReader(reader, fakeReadClusterMap)
	require.NoError(t, err)
	require.NotNil(t, hf)

	assert.Equal(t, int32(15), hf.numClusters)
	assert.Equal(t, int32(1), hf.numLFContexts)
	assert.Len(t, hf.clusterMap, 39)
}
