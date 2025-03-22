package colour

import (
	"reflect"
	"testing"

	"github.com/kpfaulkner/jxl-go/util"
)

func TestGetConversionMatrix(t *testing.T) {

	for _, tc := range []struct {
		name            string
		targetPrim      CIEPrimaries
		targetWP        CIEXY
		currentPrim     CIEPrimaries
		currentWP       CIEXY
		expectedResults [][]float32
		expectErr       bool
	}{
		{
			name: "target matches current",
			targetPrim: CIEPrimaries{
				Red:   &CIEXY{X: 0.6399987, Y: 0.33001015},
				Green: &CIEXY{X: 0.3000038, Y: 0.60000336},
				Blue:  &CIEXY{X: 0.15000205, Y: 0.059997205},
			},
			targetWP: CIEXY{
				X: 0.3127,
				Y: 0.329,
			},
			currentPrim: CIEPrimaries{
				Red:   &CIEXY{X: 0.6399987, Y: 0.33001015},
				Green: &CIEXY{X: 0.3000038, Y: 0.60000336},
				Blue:  &CIEXY{X: 0.15000205, Y: 0.059997205},
			},
			currentWP: CIEXY{
				X: 0.3127,
				Y: 0.329,
			},
			expectedResults: util.MatrixIdentity(3),
			expectErr:       false,
		},
		{
			name: "success",
			targetPrim: CIEPrimaries{
				Red:   &CIEXY{X: 0.6499987, Y: 0.33001015},
				Green: &CIEXY{X: 0.3000038, Y: 0.60000336},
				Blue:  &CIEXY{X: 0.15000205, Y: 0.059997205},
			},
			targetWP: CIEXY{
				X: 0.3127,
				Y: 0.329,
			},
			currentPrim: CIEPrimaries{
				Red:   &CIEXY{X: 0.6399987, Y: 0.33001015},
				Green: &CIEXY{X: 0.3000038, Y: 0.60000336},
				Blue:  &CIEXY{X: 0.15000205, Y: 0.059997205},
			},
			currentWP: CIEXY{
				X: 0.3127,
				Y: 0.329,
			},
			expectedResults: [][]float32{
				{0.9999998, -0.00000017508864, 0.00000000000000},
				{0.006320032, 0.99368, 0.000000007450581},
				{0.0062607024, -0.000000014901161, 0.9937391},
			},
			expectErr: false,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {

			result, err := GetConversionMatrix(tc.targetPrim, tc.targetWP, tc.currentPrim, tc.currentWP)
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

			eq := reflect.DeepEqual(result, tc.expectedResults)
			if !eq {
				t.Errorf("expected %v but got %v", tc.expectedResults, result)
			}
		})
	}
}
