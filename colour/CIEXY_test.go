package colour

import (
	"reflect"
	"testing"
)

func TestAdaptWhitePoint(t *testing.T) {

	for _, tc := range []struct {
		name            string
		targetWP        *CIEXY
		currentWP       *CIEXY
		expectedResults [][]float32
		expectErr       bool
	}{
		{
			name:            "no targetWP not currentWP",
			targetWP:        nil,
			currentWP:       nil,
			expectedResults: [][]float32{[]float32{1.065479, 0.033983506, -0.042857543}, []float32{0.04747088, 0.9720429, -0.01575847}, []float32{-0.0062496495, 0.009066325, 0.8170424}},
			expectErr:       false,
		},

		{
			name: "success",
			targetWP: &CIEXY{
				X: 0.34577,
				Y: 0.34577,
			},
			currentWP: &CIEXY{
				X: 0.3137,
				Y: 0.326,
			},
			expectedResults: [][]float32{[]float32{1.0583518, 0.029857645, -0.04368171}, []float32{0.0410195, 0.97784656, -0.015669692}, []float32{-0.0069154147, 0.010499056, 0.80369115}},
			expectErr:       false,
		},

		{
			name: "invalid target",
			targetWP: &CIEXY{
				X: -0.34577,
				Y: 0.34577,
			},
			currentWP: &CIEXY{
				X: 0.3137,
				Y: 0.326,
			},
			expectErr: true,
		},
		{
			name: "invalid current",
			targetWP: &CIEXY{
				X: 0.34577,
				Y: 0.34577,
			},
			currentWP: &CIEXY{
				X: -0.3137,
				Y: 0.326,
			},
			expectErr: true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {

			res, err := AdaptWhitePoint(tc.targetWP, tc.currentWP)
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

			if !tc.expectErr && !reflect.DeepEqual(res, tc.expectedResults) {
				t.Errorf("expected CIEXY %+v, got %+v", tc.expectedResults, res)
			}

		})
	}
}
