package entropy

import (
	"fmt"
	"testing"

	"github.com/kpfaulkner/jxl-go/testcommon"
)

func TestNewEntropyStreamWithReaderAndNumDists(t *testing.T) {

	for _, tc := range []struct {
		name         string
		expectErr    bool
		readBoolData []bool
	}{
		{
			name:         "success",
			expectErr:    false,
			readBoolData: []bool{false},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {

			bitReader := testcommon.NewFakeBitReader()
			bitReader.ReadBoolData = tc.readBoolData
			entropyStream, err := NewEntropyStreamWithReaderAndNumDists(bitReader, 8)

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

			fmt.Printf("entropyStream: %#v\n", *entropyStream)

		})
	}
}
