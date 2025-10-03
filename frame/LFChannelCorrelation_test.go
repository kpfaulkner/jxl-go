package frame

import (
	"reflect"
	"testing"

	"github.com/kpfaulkner/jxl-go/testcommon"
)

func TestNewLFChannelCorrelationWithReaderAndDefault(t *testing.T) {

	for _, tc := range []struct {
		name           string
		boolData       []bool
		enumData       []int32
		u32Data        []uint32
		f16Data        []float32
		bitsData       []uint64
		expectedResult LFChannelCorrelation
		expectErr      bool
	}{
		{
			name:      "no data",
			expectErr: true,
		},
		{
			name:     "success with default",
			boolData: []bool{true},
			enumData: nil,
			u32Data:  []uint32{},
			bitsData: []uint64{},
			expectedResult: LFChannelCorrelation{
				colorFactor:      84,
				baseCorrelationX: 0.0,
				baseCorrelationB: 1.0,
				xFactorLF:        128,
				bFactorLF:        128,
			},
			expectErr: false,
		},
		{
			name:     "success without default",
			boolData: []bool{false},
			enumData: nil,
			u32Data:  []uint32{1},
			bitsData: []uint64{1, 1},
			f16Data:  []float32{1, 1},
			expectedResult: LFChannelCorrelation{
				colorFactor:      1,
				baseCorrelationX: 1,
				baseCorrelationB: 1,
				xFactorLF:        1,
				bFactorLF:        1,
			},
			expectErr: false,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {

			bitReader := &testcommon.FakeBitReader{
				ReadU32Data:  tc.u32Data,
				ReadBitsData: tc.bitsData,
				ReadBoolData: tc.boolData,
				ReadF16Data:  tc.f16Data,
			}
			lfcc, err := NewLFChannelCorrelationWithReader(bitReader)
			if err != nil && !tc.expectErr {
				t.Errorf("got error when none was expected : %v", err)
			}
			if err == nil && tc.expectErr {
				t.Errorf("expected error but got none")
			}
			if err != nil && tc.expectErr {
				return
			}

			if !reflect.DeepEqual(*lfcc, tc.expectedResult) {
				t.Errorf("expected LFChannelCorrelation %+v, got %+v", tc.expectedResult, *lfcc)
			}
		})
	}
}
