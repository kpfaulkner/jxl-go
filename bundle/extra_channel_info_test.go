package bundle

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/kpfaulkner/jxl-go/jxlio"
	"github.com/kpfaulkner/jxl-go/testcommon"
)

func TestExtraChannelInfo(t *testing.T) {

	for _, tc := range []struct {
		name           string
		data           []uint8
		readData       bool
		expectErr      bool
		expectedResult ExtraChannelInfo
	}{
		{
			name:      "no data",
			data:      []uint8{},
			readData:  false,
			expectErr: true,
		},
		{
			name:      "success",
			data:      []uint8{},
			readData:  true,
			expectErr: false,
			expectedResult: ExtraChannelInfo{
				EcType: 0,
				BitDepth: BitDepthHeader{
					UsesFloatSamples: false,
					BitsPerSample:    9,
					ExpBits:          0,
				},
				DimShift:        0,
				name:            "",
				AlphaAssociated: false,
				red:             0,
				green:           0,
				blue:            0,
				solidity:        0,
				cfaIndex:        1,
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {

			var bitReader *jxlio.Bitreader
			if tc.readData {
				bitReader = testcommon.GenerateTestBitReader(t, `../testdata/alpha-triangles.jxl`)
				// skip first 40 bytes due to box headers.
				bitReader.SkipBits(49)
			} else {
				bitReader = jxlio.NewBitreader(bytes.NewReader(tc.data))
			}

			eci, err := NewExtraChannelInfoWithReader(bitReader)
			if err != nil && tc.expectErr {
				return
			}
			if err == nil && tc.expectErr {
				t.Errorf("expected error but got none")
				return
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

			if !tc.expectErr && !reflect.DeepEqual(eci.BitDepth, tc.expectedResult.BitDepth) {
				t.Errorf("expected bitdepth %+v, got %+v", tc.expectedResult.BitDepth, eci.BitDepth)
			}

			if !tc.expectErr && eci.EcType != tc.expectedResult.EcType {
				t.Errorf("expected EcType %+v, got %+v", tc.expectedResult.EcType, eci.EcType)
			}

			if !tc.expectErr && eci.cfaIndex != tc.expectedResult.cfaIndex {
				t.Errorf("expected cfaIndex %+v, got %+v", tc.expectedResult.cfaIndex, eci.cfaIndex)
			}

		})
	}
}
