package colour

import (
	"bytes"
	"testing"

	"github.com/kpfaulkner/jxl-go/jxlio"
)

func TestNewCustomXY(t *testing.T) {

	for _, tc := range []struct {
		name             string
		data             []uint8
		expectedCustomXY CustomXY
		expectErr        bool
	}{
		{
			name: "success",
			data: []uint8{0x40, 0x01, 0x02, 0x03, 0x04, 0x40, 0x01, 0x02, 0x03, 0x04},
			expectedCustomXY: CustomXY{
				CIEXY: CIEXY{X: 0.016424, Y: 0.001027},
			},
			expectErr: false,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {

			bitReader := jxlio.NewBitStreamReader(bytes.NewReader(tc.data))
			customXY, err := NewCustomXY(bitReader)
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

			if customXY.X != tc.expectedCustomXY.X {
				t.Errorf("expected X %f got %f", tc.expectedCustomXY.X, customXY.X)
			}

			if customXY.Y != tc.expectedCustomXY.Y {
				t.Errorf("expected Y %f got %f", tc.expectedCustomXY.Y, customXY.Y)
			}

		})
	}
}
