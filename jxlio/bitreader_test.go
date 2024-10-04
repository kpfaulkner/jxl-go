package jxlio

import (
	"bytes"
	"testing"
)

// TestReadbit tests the reading a single bit.
func TestReadbit(t *testing.T) {

	for _, tc := range []struct {
		name      string
		data      []uint8
		expected  uint8
		expectErr bool
	}{
		{
			name:      "little endian 1",
			data:      []uint8{0x01},
			expected:  1,
			expectErr: false,
		},
		{
			name:      "little endian 0",
			data:      []uint8{0xFE},
			expected:  0,
			expectErr: false,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {

			data := bytes.NewReader(tc.data)
			br := NewBitreader(data)

			resp, err := br.readBit()
			if err != nil && !tc.expectErr {
				t.Errorf("got error when none was expected : %v", err)
			}

			if err == nil && tc.expectErr {
				t.Errorf("expected error but got none")
			}

			if resp != tc.expected {
				t.Errorf("expected %v but got %v", tc.expected, resp)
			}
		})
	}
}

// TestReadbits tests the reading of multiple bits
//func TestReadbits(t *testing.T) {
//
//	for _, tc := range []struct {
//		name         string
//		data         []uint8
//		littleEndian bool
//		expected     uint8
//		expectErr    bool
//	}{
//		{
//			name:         "little endian 0",
//			data:         []uint8{0x01, 0xFF},
//			littleEndian: true,
//			expected:     0,
//			expectErr:    false,
//		},
//		{
//			name:         "little endian 1",
//			data:         []uint8{0xFF, 0x01},
//			littleEndian: true,
//			expected:     1,
//			expectErr:    false,
//		},
//		{
//			name:         "big endian 0",
//			data:         []uint8{0x01, 0xFF},
//			littleEndian: true,
//			expected:     0,
//			expectErr:    false,
//		},
//		{
//			name:         "big endian 1",
//			data:         []uint8{0xFF, 0x01},
//			littleEndian: true,
//			expected:     1,
//			expectErr:    false,
//		},
//	} {
//		t.Run(tc.name, func(t *testing.T) {
//
//			data := bytes.NewReader(tc.data)
//			br := NewBitreader(data, false)
//
//			resp, err := br.readBit()
//			if err != nil && !tc.expectErr {
//				t.Errorf("got error when none was expected : %v", err)
//			}
//
//			if err == nil && tc.expectErr {
//				t.Errorf("expected error but got none")
//			}
//
//			if resp != tc.expected {
//				t.Errorf("expected %v but got %v", tc.expected, resp)
//			}
//		})
//	}
//
//}
