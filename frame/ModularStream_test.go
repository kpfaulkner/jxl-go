package frame

import (
	"testing"

	"github.com/kpfaulkner/jxl-go/testcommon"
	"github.com/stretchr/testify/assert"
)

func TestNewSqueezeParam(t *testing.T) {
	bitReader := testcommon.NewFakeBitReader()
	bitReader.ReadU32Data = []uint32{0, 0}
	bitReader.ReadBoolData = []bool{false, false}
	sq, err := NewSqueezeParam(bitReader)
	if err != nil {
		t.Errorf("got error when none was expected : %v", err)
	}
	assert.NotNil(t, sq)
}

func TestNewTransformInfo(t *testing.T) {
	for _, tc := range []struct {
		name           string
		expectedResult int32
		bitsData       []uint64
		u32Data        []uint32
		expectErr      bool
	}{
		{
			name:           "RCT test",
			bitsData:       []uint64{0},
			u32Data:        []uint32{0, 0},
			expectedResult: 1,
			expectErr:      false,
		},
		{
			name:           "Palette test",
			bitsData:       []uint64{1, 5},
			u32Data:        []uint32{0, 0, 0, 0},
			expectedResult: 1,
			expectErr:      false,
		},
		{
			name:           "Squeeze test",
			bitsData:       []uint64{2, 5},
			u32Data:        []uint32{0, 0, 0, 0},
			expectedResult: 1,
			expectErr:      false,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {

			bitReader := testcommon.NewFakeBitReader()
			bitReader.ReadBitsData = tc.bitsData
			bitReader.ReadU32Data = tc.u32Data
			res, err := NewTransformInfo(bitReader)
			if err != nil {
				t.Errorf("got error when none was expected : %v", err)
			}

			if err != nil && !tc.expectErr {
				t.Errorf("got error when none was expected : %v", err)
			}
			if err == nil && tc.expectErr {
				t.Errorf("expected error but got none")
			}
			if err != nil && tc.expectErr {
				return
			}

			// just make sure not nil for now
			assert.NotNil(t, res)
		})
	}

}

func TestNewModularStreamWithChannels(t *testing.T) {
	for _, tc := range []struct {
		name         string
		frame        Framer
		streamIndex  int
		channelCount int
		ecStart      int
		channelArray []ModularChannel
		bitsData     []uint64
		boolData     []bool
		u32Data      []uint32
		expectErr    bool
	}{
		{
			name:      "success no channels",
			bitsData:  []uint64{0},
			u32Data:   []uint32{0, 0},
			expectErr: false,
		},
		{
			name:         "success 1 channel, PALETTE",
			channelCount: 1,
			frame:        NewFakeFramer(VARDCT),
			boolData: []bool{
				true,
				true, // wpparams default
			},
			bitsData: []uint64{
				1, // pallette
				0, // dPred

			},
			u32Data: []uint32{
				1, // nbTransforms
				0, //
				0, // numC
				0, // nbColours
				0, // nbDeltas
			},
			expectErr: false,
		},
		{
			name:         "success 1 channel, SQUEEZE",
			channelCount: 1,
			frame:        NewFakeFramer(VARDCT),
			boolData: []bool{
				true,
				true, // wpparams default
			},
			bitsData: []uint64{
				2, // pallette
				0, // dPred

			},
			u32Data: []uint32{
				1, // nbTransforms
				0, //
				0, // numC
				0, // nbColours
				0, // nbDeltas
			},
			expectErr: false,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {

			bitReader := testcommon.NewFakeBitReader()
			bitReader.ReadBitsData = tc.bitsData
			bitReader.ReadU32Data = tc.u32Data
			bitReader.ReadBoolData = tc.boolData
			res, err := NewModularStreamWithChannels(bitReader, tc.frame, tc.streamIndex, tc.channelCount, tc.ecStart, tc.channelArray)
			if err != nil {
				t.Errorf("got error when none was expected : %v", err)
			}

			if err != nil && !tc.expectErr {
				t.Errorf("got error when none was expected : %v", err)
			}
			if err == nil && tc.expectErr {
				t.Errorf("expected error but got none")
			}
			if err != nil && tc.expectErr {
				return
			}

			// just make sure not nil for now
			assert.NotNil(t, res)
		})
	}

}
