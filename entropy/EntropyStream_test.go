package entropy

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"slices"
	"testing"

	"github.com/kpfaulkner/jxl-go/testcommon"
	"github.com/stretchr/testify/assert"
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

func TestReadHybridInteger(t *testing.T) {
	es := &EntropyStream{}

	for _, tc := range []struct {
		name     string
		config   *HybridIntegerConfig
		token    int32
		bitsData []uint64
		expected int32
	}{
		{
			name:     "token below split",
			config:   NewHybridIntegerConfig(4, 0, 0),
			token:    10,
			expected: 10,
		},
		{
			name:     "token above split, simple",
			config:   NewHybridIntegerConfig(2, 0, 0), // split = 4
			token:    5,                               // n = 2 - 0 - 0 + (5-4)>>0 = 3
			bitsData: []uint64{7},                     // 3 bits = 111 (binary)
			// low = 5 & 0 = 0
			// token = 5 >> 0 = 5
			// token &= 0
			// token |= 1 = 1
			// result = (1<<3 | 7) << 0 | 0 = 15
			expected: 15,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			bitReader := testcommon.NewFakeBitReader()
			bitReader.ReadBitsData = tc.bitsData

			res, err := es.ReadHybridInteger(bitReader, tc.config, tc.token)
			assert.NoError(t, err)
			assert.Equal(t, tc.expected, res)
		})
	}
}

func TestReadClusterMapSimple(t *testing.T) {
	// simpleClustering = true
	// nbits = 2
	// clusterMap entries
	readBoolData := []bool{true}
	readBitsData := []uint64{2, 0, 1, 2, 3} // nbits=2, then 4 clusters

	bitReader := testcommon.NewFakeBitReader()
	bitReader.ReadBoolData = readBoolData
	bitReader.ReadBitsData = readBitsData

	clusterMap := make([]int, 4)
	numClusters, err := ReadClusterMap(bitReader, clusterMap, 10)
	assert.NoError(t, err)
	assert.Equal(t, 4, numClusters)
	assert.Equal(t, []int{0, 1, 2, 3}, clusterMap)
}
