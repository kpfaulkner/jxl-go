package image

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewImageBuffer(t *testing.T) {
	buf, err := NewImageBuffer(TYPE_INT, 5, 5)
	assert.Nil(t, err)

	assert.NotNil(t, buf)
}

func TestNewImageBufferFromInts(t *testing.T) {
	origBuf := [][]int32{{1, 2, 3}, {4, 5, 6}, {7, 8, 9}}
	buf := NewImageBufferFromInts(origBuf)
	assert.NotNil(t, buf)
}

func TestNewImageBufferFromFloats(t *testing.T) {
	origBuf := [][]float32{{1, 2, 3}, {4, 5, 6}, {7, 8, 9}}
	buf := NewImageBufferFromFloats(origBuf)
	assert.NotNil(t, buf)
}

func TestNewImageBufferFromImageBuffer(t *testing.T) {
	origBuf := [][]int32{{1, 2, 3}, {4, 5, 6}, {7, 8, 9}}
	buf := NewImageBufferFromInts(origBuf)

	buf2 := NewImageBufferFromImageBuffer(buf)
	assert.NotNil(t, buf2)
}

func TestEqualsInt(t *testing.T) {
	origBuf := [][]int32{{1, 2, 3}, {4, 5, 6}, {7, 8, 9}}
	buf := NewImageBufferFromInts(origBuf)
	buf2 := NewImageBufferFromInts(origBuf)

	if !buf.Equals(*buf2) {
		t.Failed()
	}
}

func TestEqualsFloat(t *testing.T) {
	origBuf := [][]float32{{1, 2, 3}, {4, 5, 6}, {7, 8, 9}}
	buf := NewImageBufferFromFloats(origBuf)
	buf2 := NewImageBufferFromFloats(origBuf)

	if !buf.Equals(*buf2) {
		t.Failed()
	}
}

func TestCastToFloat(t *testing.T) {
	origBuf := [][]int32{{1, 2, 3}, {4, 5, 6}, {7, 8, 9}}
	buf := NewImageBufferFromInts(origBuf)

	// confirm floats are nil to begin with
	assert.Nil(t, buf.FloatBuffer)
	err := buf.castToFloatBuffer(10)
	assert.Nil(t, err)
	assert.NotNil(t, buf.FloatBuffer)
}

func TestCastToInt(t *testing.T) {
	origBuf := [][]float32{{1, 2, 3}, {4, 5, 6}, {7, 8, 9}}
	buf := NewImageBufferFromFloats(origBuf)

	// confirm int are nil to begin with
	assert.Nil(t, buf.IntBuffer)
	err := buf.castToIntBuffer(10)
	assert.Nil(t, err)
	assert.NotNil(t, buf.IntBuffer)
}

func TestCastToIntIfFloat(t *testing.T) {
	origBuf := [][]float32{{1, 2, 3}, {4, 5, 6}, {7, 8, 9}}
	buf := NewImageBufferFromFloats(origBuf)

	err := buf.CastToIntIfFloat(10)
	assert.Nil(t, err)
}

func TestCastToFloatIfInt(t *testing.T) {
	origBuf := [][]int32{{1, 2, 3}, {4, 5, 6}, {7, 8, 9}}
	buf := NewImageBufferFromInts(origBuf)

	err := buf.CastToFloatIfInt(10)
	assert.Nil(t, err)
}

func TestClamp(t *testing.T) {
	origBuf := [][]int32{{1, 2, 3}, {4, 5, 6}, {7, 8, 9}}
	buf := NewImageBufferFromInts(origBuf)

	err := buf.Clamp(10)
	assert.Nil(t, err)
}

func TestImageBufferSliceEquals(t *testing.T) {
	origBuf := [][]int32{{1, 2, 3}, {4, 5, 6}, {7, 8, 9}}
	buf := NewImageBufferFromInts(origBuf)
	buf2 := NewImageBufferFromInts(origBuf)

	if !ImageBufferSliceEquals([]ImageBuffer{*buf}, []ImageBuffer{*buf2}) {
		t.Failed()
	}
}

//func TestNewTransformInfo(t *testing.T) {
//	for _, tc := range []struct {
//		name           string
//		expectedResult int32
//		bitsData       []uint64
//		u32Data        []uint32
//		expectErr      bool
//	}{
//		{
//			name:           "RCT test",
//			bitsData:       []uint64{0},
//			u32Data:        []uint32{0, 0},
//			expectedResult: 1,
//			expectErr:      false,
//		},
//		{
//			name:           "Palette test",
//			bitsData:       []uint64{1, 5},
//			u32Data:        []uint32{0, 0, 0, 0},
//			expectedResult: 1,
//			expectErr:      false,
//		},
//		{
//			name:           "Squeeze test",
//			bitsData:       []uint64{2, 5},
//			u32Data:        []uint32{0, 0, 0, 0},
//			expectedResult: 1,
//			expectErr:      false,
//		},
//	} {
//		t.Run(tc.name, func(t *testing.T) {
//
//			bitReader := testcommon.NewFakeBitReader()
//			bitReader.ReadBitsData = tc.bitsData
//			bitReader.ReadU32Data = tc.u32Data
//			res, err := NewTransformInfo(bitReader)
//			if err != nil {
//				t.Errorf("got error when none was expected : %v", err)
//			}
//
//			if err != nil && !tc.expectErr {
//				t.Errorf("got error when none was expected : %v", err)
//			}
//			if err == nil && tc.expectErr {
//				t.Errorf("expected error but got none")
//			}
//			if err != nil && tc.expectErr {
//				return
//			}
//
//			// just make sure not nil for now
//			assert.NotNil(t, res)
//		})
//	}
//
//}
//
//func TestNewModularStreamWithChannels(t *testing.T) {
//	for _, tc := range []struct {
//		name         string
//		frame        Framer
//		streamIndex  int
//		channelCount int
//		ecStart      int
//		channelArray []ModularChannel
//		bitsData     []uint64
//		boolData     []bool
//		u32Data      []uint32
//		expectErr    bool
//	}{
//		{
//			name:      "success no channels",
//			bitsData:  []uint64{0},
//			u32Data:   []uint32{0, 0},
//			expectErr: false,
//		},
//		{
//			name:         "success 1 channel, PALETTE",
//			channelCount: 1,
//			frame:        NewFakeFramer(),
//			boolData: []bool{
//				true,
//				true, // wpparams default
//			},
//			bitsData: []uint64{
//				1, // pallette
//				0, // dPred
//
//			},
//			u32Data: []uint32{
//				1, // nbTransforms
//				0, //
//				0, // numC
//				0, // nbColours
//				0, // nbDeltas
//			},
//			expectErr: false,
//		},
//		{
//			name:         "success 1 channel, SQUEEZE",
//			channelCount: 1,
//			frame:        NewFakeFramer(),
//			boolData: []bool{
//				true,
//				true, // wpparams default
//			},
//			bitsData: []uint64{
//				2, // pallette
//				0, // dPred
//
//			},
//			u32Data: []uint32{
//				1, // nbTransforms
//				0, //
//				0, // numC
//				0, // nbColours
//				0, // nbDeltas
//			},
//			expectErr: false,
//		},
//	} {
//		t.Run(tc.name, func(t *testing.T) {
//
//			bitReader := testcommon.NewFakeBitReader()
//			bitReader.ReadBitsData = tc.bitsData
//			bitReader.ReadU32Data = tc.u32Data
//			bitReader.ReadBoolData = tc.boolData
//			res, err := NewModularStreamWithChannels(bitReader, tc.frame, tc.streamIndex, tc.channelCount, tc.ecStart, tc.channelArray)
//			if err != nil {
//				t.Errorf("got error when none was expected : %v", err)
//			}
//
//			if err != nil && !tc.expectErr {
//				t.Errorf("got error when none was expected : %v", err)
//			}
//			if err == nil && tc.expectErr {
//				t.Errorf("expected error but got none")
//			}
//			if err != nil && tc.expectErr {
//				return
//			}
//
//			// just make sure not nil for now
//			assert.NotNil(t, res)
//		})
//	}
//
//}
