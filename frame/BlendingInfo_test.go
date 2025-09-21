package frame

import (
	"reflect"
	"testing"

	"github.com/kpfaulkner/jxl-go/testcommon"
)

func TestNewBlendingInfo(t *testing.T) {
	bi := NewBlendingInfo()
	expectedBI := BlendingInfo{Mode: BLEND_REPLACE, AlphaChannel: 0, Clamp: false, Source: 0}

	if !reflect.DeepEqual(*bi, expectedBI) {
		t.Errorf("expected BlendinfInfo %+v, got %+v", expectedBI, *bi)
	}

}

func TestNewBlendingInfoWithReader(t *testing.T) {

	for _, tc := range []struct {
		name       string
		extra      bool
		fullFrame  bool
		expectedBI BlendingInfo
		boolData   []bool
		enumData   []int32
		u32Data    []uint32
		bitsData   []uint64
		expectErr  bool
	}{
		{
			name:      "no data",
			expectErr: true,
		},
		{
			name:      "no extra, mode blend",
			extra:     false,
			fullFrame: false,
			u32Data: []uint32{
				BLEND_BLEND, // mode
			},
			bitsData: []uint64{
				0b11, // source
			},
			expectedBI: BlendingInfo{
				Mode:         2,
				AlphaChannel: 0,
				Clamp:        false,
				Source:       3,
			},
			expectErr: false,
		},
		{
			name:      "extra, mode blend",
			extra:     true,
			fullFrame: false,
			u32Data: []uint32{
				BLEND_BLEND, // mode
				0x3,         // alpha channel
			},
			bitsData: []uint64{
				0b11, // source
			},
			boolData: []bool{
				true, // clamp
			},
			expectedBI: BlendingInfo{
				Mode:         2,
				AlphaChannel: 3,
				Clamp:        true,
				Source:       3,
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
			bi, err := NewBlendingInfoWithReader(bitReader, tc.extra, tc.fullFrame)
			if err != nil && !tc.expectErr {
				t.Errorf("got error when none was expected : %v", err)
			}
			if err == nil && tc.expectErr {
				t.Errorf("expected error but got none")
			}
			if err != nil && tc.expectErr {
				return
			}

			if !reflect.DeepEqual(*bi, tc.expectedBI) {
				t.Errorf("expected BundleInfo %+v, got %+v", tc.expectedBI, *bi)
			}

		})
	}
}
