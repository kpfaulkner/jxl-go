package frame

import (
	"reflect"
	"testing"

	"github.com/kpfaulkner/jxl-go/entropy"
	"github.com/kpfaulkner/jxl-go/jxlio"
	"github.com/kpfaulkner/jxl-go/testcommon"
)

func TestNewHFPassWithReader(t *testing.T) {

	for _, tc := range []struct {
		name               string
		frame              Framer
		passIndex          uint32
		boolData           []bool
		enumData           []int32
		u32Data            []uint32
		bitsData           []uint64
		readClusterMapFunc entropy.ReadClusterMapFunc
		entropyStreamFunc  entropy.EntropyStreamWithReaderAndNumDistsFunc
		expectedResult     HFPass
		expectErr          bool
	}{
		{
			name:      "no data",
			expectErr: true,
		},
		{
			name: "success",
			frame: &FakeFramer{
				hfGlobal: &HFGlobal{numHFPresets: 1},
				lfGlobal: &LFGlobal{
					hfBlockCtx: &HFBlockContext{},
				},
			},
			readClusterMapFunc: fakeReadClusterMap,
			entropyStreamFunc: func(reader jxlio.BitReader, numDists int, readClusterMapFunc ReadClusterMapFunc) (*entropy.EntropyStream, error) {
				return nil, nil
			},
			boolData: []bool{
				false, // usesLZ77 for entity stream reader
				true,  // simple clustering
				true,  // prefix codes
			},
			enumData: nil,
			u32Data:  []uint32{1},
			bitsData: []uint64{
				1,
				1,
				0,
			},
			expectedResult: HFPass{},
			expectErr:      false,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {

			bitReader := &testcommon.FakeBitReader{
				ReadU32Data:  tc.u32Data,
				ReadBitsData: tc.bitsData,
				ReadBoolData: tc.boolData,
			}
			hpass, err := NewHFPassWithReader(bitReader, tc.frame, tc.passIndex, tc.readClusterMapFunc, tc.entropyStreamFunc)
			if err != nil && !tc.expectErr {
				t.Errorf("got error when none was expected : %v", err)
			}
			if err == nil && tc.expectErr {
				t.Errorf("expected error but got none")
			}
			if err != nil && tc.expectErr {
				return
			}
			if !reflect.DeepEqual(*hpass, tc.expectedResult) {
				t.Errorf("expected HFPass %+v, got %+v", tc.expectedResult, *hpass)
			}

		})
	}
}
