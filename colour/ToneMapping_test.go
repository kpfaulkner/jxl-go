package colour

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/kpfaulkner/jxl-go/jxlio"
)

func TestNewToneMappingWithReader(t *testing.T) {

	for _, tc := range []struct {
		name                string
		data                []uint8
		expectedToneMapping ToneMapping
		expectErr           bool
	}{
		{
			name: "success",
			data: []uint8{0b01000000, 0b01000000, 0b01000000, 0b01000000, 0b01000000, 0b01000000, 0b01000000},
			expectedToneMapping: ToneMapping{
				IntensityTarget:      0.008056641,
				MinNits:              0.008056641,
				LinearBelow:          0.00049591064,
				RelativeToMaxDisplay: false,
			},
			expectErr: false,
		},

		{
			name: "success default",
			data: []uint8{0b01000001, 0b01000000, 0b01000000, 0b01000000, 0b01000000, 0b01000000, 0b01000000},
			expectedToneMapping: ToneMapping{
				IntensityTarget:      255,
				MinNits:              0,
				LinearBelow:          0,
				RelativeToMaxDisplay: false,
			},
			expectErr: false,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {

			bitReader := jxlio.NewBitStreamReader(bytes.NewReader(tc.data))
			toneMapping, err := NewToneMappingWithReader(bitReader)
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

			if !tc.expectErr && !reflect.DeepEqual(*toneMapping, tc.expectedToneMapping) {
				t.Errorf("expected ToneMapping %+v, got %+v", tc.expectedToneMapping, toneMapping)
			}

		})
	}
}
