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
		readBitsData []uint64
		readU32Data  []uint32
	}{
		{
			name:         "success",
			expectErr:    false,
			readBoolData: []bool{true, false, false, false, true, true, true, true, true, true},
			readBitsData: []uint64{0x0, 0x0, 0x0, 0x2, 0x0, 0x1, 0x1, 0x0, 0x1, 0x2, 0x0, 0x1, 0x2, 0x0, 0x0, 0x0, 0x3, 0x1, 0x0, 0x0, 0x0, 0x0, 0x9, 0x6, 0x9, 0x6, 0x3, 0x0, 0x0, 0x6, 0x5, 0x2, 0x0, 0x7, 0x6, 0x4, 0x5, 0x1, 0x2, 0x0, 0x7, 0x8},
			readU32Data:  []uint32{0x200, 0x3},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {

			bitReader := testcommon.NewFakeBitReader()
			bitReader.ReadBoolData = tc.readBoolData
			bitReader.ReadBitsData = tc.readBitsData
			bitReader.ReadU32Data = tc.readU32Data
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
