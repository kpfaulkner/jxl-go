package frame

import (
	"testing"

	"github.com/kpfaulkner/jxl-go/testcommon"
	"github.com/stretchr/testify/assert"
)

func TestNewPassesInfo(t *testing.T) {
	pi := NewPassesInfo()
	assert.NotNil(t, pi)
}

func TestNewPassesInfoWithReader(t *testing.T) {

	for _, tc := range []struct {
		name           string
		boolData       []bool
		enumData       []int32
		u32Data        []uint32
		f16Data        []float32
		bitsData       []uint64
		expectedResult PassesInfo
		expectErr      bool
	}{
		{
			name:      "no data",
			expectErr: true,
		},
		{
			name:     "success",
			u32Data:  []uint32{2, 1, 1, 1},
			bitsData: []uint64{1, 1},
			expectedResult: PassesInfo{
				shift:      []uint32{1, 0},
				downSample: []uint32{2, 1},
				lastPass:   []uint32{1, 1},
				numPasses:  2,
				numDS:      1,
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
			pi, err := NewPassesInfoWithReader(bitReader)
			if err != nil && !tc.expectErr {
				t.Errorf("got error when none was expected : %v", err)
			}
			if err == nil && tc.expectErr {
				t.Errorf("expected error but got none")
			}
			if err != nil && tc.expectErr {
				return
			}
			assert.NotNil(t, pi)
			assert.Equal(t, tc.expectedResult, *pi)
		})
	}
}
