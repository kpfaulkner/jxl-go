package frame

import (
	"reflect"
	"testing"

	"github.com/kpfaulkner/jxl-go/jxlio"
	"github.com/kpfaulkner/jxl-go/testcommon"
)

func TestNewHFCoefficientsWithReader(t *testing.T) {

	for _, tc := range []struct {
		name           string
		frame          Framer
		pass           uint32
		group          uint32
		boolData       []bool
		enumData       []int32
		u32Data        []uint32
		bitsData       []uint64
		expectedResult HFCoefficients
		expectErr      bool
	}{
		{
			name:      "no data",
			expectErr: true,
		},
		{
			name:    "success",
			pass:    0,
			group:   0,
			frame:   NewFakeFramer(),
			u32Data: []uint32{},
			bitsData: []uint64{
				0, // hfPreset
			},
			boolData:       []bool{},
			expectedResult: HFCoefficients{},

			expectErr: false,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {

			bitReader := &testcommon.FakeBitReader{
				ReadU32Data:  tc.u32Data,
				ReadBitsData: tc.bitsData,
				ReadBoolData: tc.boolData,
			}
			hfc, err := NewHFCoefficientsWithReader(bitReader, tc.frame, tc.pass, tc.group)
			if err != nil && !tc.expectErr {
				t.Errorf("got error when none was expected : %v", err)
			}
			if err == nil && tc.expectErr {
				t.Errorf("expected error but got none")
			}
			if err != nil && tc.expectErr {
				return
			}
			if !reflect.DeepEqual(*hfc, tc.expectedResult) {
				t.Errorf("expected HFBlockContext %+v, got %+v", tc.expectedResult, *hf)
			}

		})
	}
}
