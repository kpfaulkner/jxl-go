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
			name:      "success alpha",
			data:      []byte{0xf0, 0x3, 0x8, 0x7e, 0xc0, 0x48, 0xc6, 0xc8, 0x86, 0x16, 0x7, 0x80, 0x87, 0xe0, 0x7, 0x6e, 0x76, 0x0, 0x0, 0xe, 0x77, 0x0, 0x78, 0x8, 0x4f, 0x18, 0x3, 0x50, 0x68, 0x48, 0x52, 0x2, 0xc, 0xd, 0x0, 0x3e, 0x80, 0x30, 0x1, 0xd3, 0xc0, 0x34, 0x70, 0x0, 0xa9, 0x39, 0xbc, 0x90, 0xc4, 0x36, 0x9d, 0xd7, 0x9, 0xaa, 0x17, 0x54, 0xe8, 0x62, 0x8d, 0x5f, 0x6e, 0xb4, 0xb4, 0xf0, 0x25, 0x9f, 0x70, 0x18, 0x56, 0x2f, 0x39, 0xd, 0x31, 0x94, 0x8b, 0x91, 0x40, 0x32, 0x28, 0x17, 0x1, 0x5f, 0x23, 0x18, 0x7b, 0x8, 0xa1, 0x89, 0x34, 0xc0, 0xc6, 0x6e, 0xa6, 0x81, 0x75, 0x89, 0x56, 0x82, 0x21, 0xf4},
			readData:  false,
			expectErr: false,
			expectedResult: ExtraChannelInfo{
				EcType: 0,
				BitDepth: BitDepthHeader{
					UsesFloatSamples: false,
					BitsPerSample:    16,
					ExpBits:          0,
				},
				DimShift:        0,
				Name:            "",
				AlphaAssociated: false,
				Red:             0,
				Green:           0,
				Blue:            0,
				Solidity:        0,
				CfaIndex:        1,
			},
		},

		{
			name:      "success spot",
			data:      []byte{0x4, 0x3f, 0x60, 0x24, 0x63, 0x64, 0x43, 0x8b, 0x3, 0xc0, 0x43, 0xf0, 0x3, 0x37, 0x3b, 0x0, 0x0, 0x87, 0x3b, 0x0, 0x3c, 0x84, 0x27, 0x8c, 0x1, 0x28, 0x34, 0x24, 0x29, 0x1, 0x86, 0x6, 0x0, 0x1f, 0x40, 0x98, 0x80, 0x69, 0x60, 0x1a, 0x38, 0x80, 0xd4, 0x1c, 0x5e, 0x48, 0x62, 0x9b, 0xce, 0xeb, 0x4, 0xd5, 0xb, 0x2a, 0x74, 0xb1, 0xc6, 0x2f, 0x37, 0x5a, 0x5a, 0xf8, 0x92, 0x4f, 0x38, 0xc, 0xab, 0x97, 0x9c, 0x86, 0x18, 0xca, 0xc5, 0x48, 0x20, 0x19, 0x94, 0x8b, 0x80, 0xaf, 0x11, 0x8c, 0x3d, 0x84, 0xd0, 0x44, 0x1a, 0x60, 0x63, 0x37, 0xd3, 0xc0, 0xba, 0x44, 0x2b, 0xc1, 0x10, 0x7a, 0x21, 0x56},
			readData:  false,
			expectErr: false,
			expectedResult: ExtraChannelInfo{
				EcType: 2,
				BitDepth: BitDepthHeader{
					UsesFloatSamples: false,
					BitsPerSample:    16,
					ExpBits:          0,
				},
				DimShift:        0,
				Name:            "",
				AlphaAssociated: false,
				Red:             0.19604492,
				Green:           0.39208984,
				Blue:            0.5878906,
				Solidity:        1,
				CfaIndex:        1,
			},
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
					BitsPerSample:    8,
					ExpBits:          0,
				},
				DimShift:        0,
				Name:            "",
				AlphaAssociated: false,
				Red:             0,
				Green:           0,
				Blue:            0,
				Solidity:        0,
				CfaIndex:        1,
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {

			var bitReader jxlio.BitReader
			if tc.readData {
				bitReader = testcommon.GenerateTestBitReader(t, `../testdata/spot.jxl`)
				// skip first 40 bytes due to box headers.
				bitReader.Skip(224)
				bitReader.SkipBits(57)
			} else {
				bitReader = jxlio.NewBitStreamReader(bytes.NewReader(tc.data))
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

			if !tc.expectErr && !reflect.DeepEqual(eci.Red, tc.expectedResult.Red) {
				t.Errorf("expected red %+v, got %+v", tc.expectedResult.Red, eci.Red)
			}

			if !tc.expectErr && !reflect.DeepEqual(eci.Green, tc.expectedResult.Green) {
				t.Errorf("expected green %+v, got %+v", tc.expectedResult.Green, eci.Green)
			}

			if !tc.expectErr && !reflect.DeepEqual(eci.Blue, tc.expectedResult.Blue) {
				t.Errorf("expected blue %+v, got %+v", tc.expectedResult.Blue, eci.Blue)
			}

			if !tc.expectErr && !reflect.DeepEqual(eci.BitDepth, tc.expectedResult.BitDepth) {
				t.Errorf("expected bitdepth %+v, got %+v", tc.expectedResult.BitDepth, eci.BitDepth)
			}

			if !tc.expectErr && eci.EcType != tc.expectedResult.EcType {
				t.Errorf("expected EcType %+v, got %+v", tc.expectedResult.EcType, eci.EcType)
			}

			if !tc.expectErr && eci.CfaIndex != tc.expectedResult.CfaIndex {
				t.Errorf("expected CfaIndex %+v, got %+v", tc.expectedResult.CfaIndex, eci.CfaIndex)
			}

		})
	}
}
