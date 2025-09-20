package colour

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/kpfaulkner/jxl-go/testcommon"
)

func TestColourEncodingBundle(t *testing.T) {
	ceb, _ := NewColourEncodingBundle()

	expectedCEB := ColourEncodingBundle{ColourEncoding: 0, WhitePoint: 1, Primaries: 1, Tf: 16777229, RenderingIntent: 1, UseIccProfile: false,
		Prim: &CIEPrimaries{
			Red:   &CIEXY{X: 0.6399987, Y: 0.33001015},
			Green: &CIEXY{X: 0.3000038, Y: 0.60000336},
			Blue:  &CIEXY{X: 0.15000205, Y: 0.059997205},
		},
		White: &CIEXY{X: 0.3127, Y: 0.329}}

	if !reflect.DeepEqual(*ceb, expectedCEB) {
		t.Errorf("expected ColourEncodingBundle %+v, got %+v", expectedCEB, *ceb)
	}

}

func TestColourEncodingBundleWithReader(t *testing.T) {

	for _, tc := range []struct {
		name        string
		readData    bool
		expectedCEB ColourEncodingBundle
		boolData    []bool
		enumData    []int32
		expectErr   bool
		skipBytes   uint32
		jxlFilePath string
	}{
		//{
		//	name:      "no data",
		//	data:      []uint8{},
		//	readData:  false,
		//	expectErr: true,
		//},
		{
			name:     "success, allDefault",
			boolData: []bool{true},
			expectedCEB: ColourEncodingBundle{ColourEncoding: 0, WhitePoint: 1, Primaries: 1, Tf: 16777229, RenderingIntent: 1, UseIccProfile: false,
				Prim: &CIEPrimaries{
					Red:   &CIEXY{X: 0.6399987, Y: 0.33001015},
					Green: &CIEXY{X: 0.3000038, Y: 0.60000336},
					Blue:  &CIEXY{X: 0.15000205, Y: 0.059997205},
				},
				White: &CIEXY{X: 0.3127, Y: 0.329}},

			expectErr: false,
		},
		{
			name:      "NOT allDefault, invalid whitepoint",
			boolData:  []bool{false},
			expectErr: true,
		},
		{
			name: "success, NOT  allDefault",
			boolData: []bool{
				false, // not all default
				false, // not ICC
				false, // gamma
			},
			enumData: []int32{
				1, // colour encoding
				1, // whitepoint
				1, // not gamma enum
				1, // rendering intent
			},

			expectedCEB: ColourEncodingBundle{ColourEncoding: 1, WhitePoint: 1, Primaries: 1, Tf: 16777217, RenderingIntent: 1, UseIccProfile: false,
				Prim: &CIEPrimaries{
					Red:   &CIEXY{X: 0.6399987, Y: 0.33001015},
					Green: &CIEXY{X: 0.3000038, Y: 0.60000336},
					Blue:  &CIEXY{X: 0.15000205, Y: 0.059997205},
				},
				White: &CIEXY{X: 0.3127, Y: 0.329}},

			expectErr: false,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {

			bitReader := &testcommon.FakeBitReader{
				ReadBoolData: tc.boolData,
				ReadEnumData: tc.enumData,
			}
			ceb, err := NewColourEncodingBundleWithReader(bitReader)
			if err != nil && !tc.expectErr {
				t.Errorf("got error when none was expected : %v", err)
			}
			if err == nil && tc.expectErr {
				t.Errorf("expected error but got none")
			}
			if err != nil && tc.expectErr {
				return
			}

			fmt.Printf("ceb %#v\n", *ceb)

			fmt.Printf("ColourEncodingBundle: %+v\n", *ceb)

			if !reflect.DeepEqual(*ceb, tc.expectedCEB) {
				t.Errorf("expected ColourEncodingBundle %+v, got %+v", tc.expectedCEB, *ceb)
			}

		})
	}
}
