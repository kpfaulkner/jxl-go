package core

import (
	"bytes"
	"os"
	"reflect"
	"testing"

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
		data               []uint8
		expectErr          bool
		expectedBoxHeaders []ContainerBoxHeader
	}{
		{
			name:               "success",
			data:               []uint8{0x01},
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
