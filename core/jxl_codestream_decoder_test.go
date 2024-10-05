package core

import (
	"bytes"
	"os"
	"reflect"
	"testing"

	"github.com/kpfaulkner/jxl-go/bundle"
	"github.com/kpfaulkner/jxl-go/color"
	"github.com/kpfaulkner/jxl-go/jxlio"
)

func generateTestBitReader(t *testing.T) *jxlio.Bitreader {
	data, err := os.ReadFile(`../testdata/unittest.jxl`)
	if err != nil {
		t.Errorf("error reading test jxl file : %v", err)
		return nil
	}
	br := jxlio.NewBitreader(bytes.NewReader(data))

	return br
}

// TestReadSignatureAndBoxes tests the ReadSignatureAndBoxes function.
// For testing, instead of providing a mock Bitreader and having to fake all the data
// I'll provide a real Bitreader and test the function with real data. May eventually
// swap it out for a mock Bitreader.
func TestReadSignatureAndBoxes(t *testing.T) {

	for _, tc := range []struct {
		name               string
		expectErr          bool
		expectedBoxHeaders []ContainerBoxHeader
	}{
		{
			name:               "success",
			expectErr:          false,
			expectedBoxHeaders: []ContainerBoxHeader{{BoxType: JXLC, BoxSize: 7085536, IsLast: false, Offset: 40}},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {

			br := generateTestBitReader(t)
			decoder := NewJXLCodestreamDecoder(br, nil)
			err := decoder.readSignatureAndBoxes()
			if err != nil && tc.expectErr {
				// got what we wanted..
				return
			}

			if err == nil && tc.expectErr {
				t.Errorf("expected error but got none")
			}

			if err != nil && !tc.expectErr {
				t.Errorf("expected no error but got %v", err)
			}

			if len(decoder.boxHeaders) != len(tc.expectedBoxHeaders) {
				t.Errorf("expected %d box headers but got %d", len(tc.expectedBoxHeaders), len(decoder.boxHeaders))
			}

			if !reflect.DeepEqual(decoder.boxHeaders, tc.expectedBoxHeaders) {
				t.Errorf("actual box headers does not match expected box headers")
			}
		})
	}
}

func TestGetImageHeader(t *testing.T) {

	for _, tc := range []struct {
		name           string
		data           []uint8
		expectErr      bool
		expectedHeader bundle.ImageHeader
	}{
		{
			name:      "success",
			expectErr: false,
			expectedHeader: bundle.ImageHeader{
				level: 5,
				size: &Dimension{
					width:  3264,
					height: 2448,
				},
				orientation: 1,
				bitDepth: &bundle.BitDepthHeader{
					UsesFloatSamples: false,
					BitsPerSample:    8,
					ExpBits:          0,
				},
				orientedWidth:       3264,
				orientedHeight:      2448,
				modular16BitBuffers: true,
				extraChannelInfo:    nil,
				xybEncoded:          false,
				colorEncoding: &color.ColorEncodingBundle{
					UseIccProfile: false,
					ColorEncoding: 0,
					WhitePoint:    1,
					White: &color.CIEXY{
						X: 0.3127,
						Y: 0.329,
					},
					Primaries: 1,
					Prim: &color.CIEPrimaries{
						Red: &color.CIEXY{
							X: 0.6399987,
							Y: 0.33001015,
						},
						Green: &color.CIEXY{
							X: 0.3000038,
							Y: 0.60000336,
						},
						Blue: &color.CIEXY{
							X: 0.15000205,
							Y: 0.059997205,
						},
					},
					Tf:              16777229,
					RenderingIntent: 0,
				},
				alphaIndices: nil,
				toneMapping: &color.ToneMapping{
					IntensityTarget:      255,
					MinNits:              0,
					RelativeToMaxDisplay: false,
					LinearBelow:          0,
				},
				extensions:         nil,
				opsinInverseMatrix: nil,
				up2Weights:         nil,
				up4Weights:         nil,
				up8Weights:         nil,
				encodedICC:         nil,
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {

			br := generateTestBitReader(t)
			decoder := NewJXLCodestreamDecoder(br, nil)

			header, err := decoder.GetImageHeader()
			if err != nil && tc.expectErr {
				// got what we wanted..
				return
			}

			if err == nil && tc.expectErr {
				t.Errorf("expected error but got none")
			}

			if err != nil && !tc.expectErr {
				t.Errorf("expected no error but got %v", err)
			}

			// not going to deepequals the whole struct. Will check a few key fields and will extend this later on.
			if header.level != tc.expectedHeader.level {
				t.Errorf("expected level %d but got %d", tc.expectedHeader.level, header.level)
			}

			if header.size.width != tc.expectedHeader.size.width {
				t.Errorf("expected width %d but got %d", tc.expectedHeader.size.width, header.size.width)
			}
			if header.size.height != tc.expectedHeader.size.height {
				t.Errorf("expected height %d but got %d", tc.expectedHeader.size.height, header.size.height)
			}
			if header.bitDepth.bitsPerSample != tc.expectedHeader.bitDepth.bitsPerSample {
				t.Errorf("expected bits per sample %d but got %d", tc.expectedHeader.bitDepth.bitsPerSample, header.bitDepth.bitsPerSample)
			}
			if header.bitDepth.usesFloatSamples != tc.expectedHeader.bitDepth.usesFloatSamples {
				t.Errorf("expected uses float samples %t but got %t", tc.expectedHeader.bitDepth.usesFloatSamples, header.bitDepth.usesFloatSamples)
			}
			if header.bitDepth.expBits != tc.expectedHeader.bitDepth.expBits {
				t.Errorf("expected exp bits %d but got %d", tc.expectedHeader.bitDepth.expBits, header.bitDepth.expBits)
			}
			if header.colorEncoding.UseIccProfile != tc.expectedHeader.colorEncoding.UseIccProfile {
				t.Errorf("expected use icc profile %t but got %t", tc.expectedHeader.colorEncoding.UseIccProfile, header.colorEncoding.UseIccProfile)
			}

			if !header.colorEncoding.Prim.Matches(tc.expectedHeader.colorEncoding.Prim) {
				t.Errorf("expected primaries to match but they did not")
			}
		})
	}
}
