package entropy

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"slices"
	"testing"

	"github.com/kpfaulkner/jxl-go/testcommon"
)

func TestNewEntropyStreamWithReaderAndNumDists(t *testing.T) {

	for _, tc := range []struct {
		name                  string
		expectErr             bool
		readBoolData          []bool
		readBitsData          []uint64
		readU8Data            []int
		readU32Data           []uint32
		showBitsData          []uint64
		expectedEntropyStream EntropyStream
		expectedDistSignature []byte
	}{
		{
			name:      "success",
			expectErr: false,

			readBoolData: []bool{false, true, true},
			readBitsData: []uint64{0x2, 0x0, 0x1, 0x2, 0x3, 0x0},
			showBitsData: []uint64{0xa, 0x2, 0xc, 0x3, 0xc, 0x7, 0x3, 0x10, 0x18, 0x1c, 0x1e, 0xf, 0x5, 0x19},

			// just text signature for now
			expectedDistSignature: []byte{193, 119, 6, 136, 1, 214, 83, 177, 115, 115, 47, 99, 24, 210, 6, 197, 212, 115, 25, 49, 161, 170, 45, 202, 3, 67, 76, 62, 179, 254, 245, 56},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {

			bitReader := testcommon.NewFakeBitReader()
			bitReader.ReadBoolData = tc.readBoolData
			bitReader.ReadBitsData = tc.readBitsData
			bitReader.ReadU8Data = tc.readU8Data
			bitReader.ReadU32Data = tc.readU32Data
			bitReader.ShowBitsData = tc.showBitsData

			entropyStream, err := NewEntropyStreamWithReaderAndNumDists(bitReader, 1, ReadClusterMap)

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

			// generate signature of dist table for comparison
			sig := generateVLCTableSignature(entropyStream)

			if slices.Compare(tc.expectedDistSignature, sig) != 0 {
				t.Errorf("entropy stream dist signature mismatch, got %v, want %v", sig, tc.expectedDistSignature)
			}

		})
	}
}

func generateVLCTableSignature(es EntropyStreamer) []byte {
	buf := new(bytes.Buffer)

	if pre, ok := es.GetDists()[0].(*PrefixSymbolDistribution); ok {
		for _, row := range pre.table.table {
			for _, val := range row {
				err := binary.Write(buf, binary.LittleEndian, val)
				if err != nil {
					fmt.Println("Error writing int32:", err)
					return nil
				}
			}
		}
	}
	sig := sha256.Sum256(buf.Bytes())
	return sig[:]
}
