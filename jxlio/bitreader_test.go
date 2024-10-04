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
			name:      "result 1",
			data:      []uint8{0x01},
			expected:  1,
			expectErr: false,
		},
		{
			name:      "result 0",
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

// TestReadbits tests the reading multiple bits.
func TestReadbits(t *testing.T) {

	for _, tc := range []struct {
		name      string
		data      []uint8
		numBits   uint32
		expected  uint64
		expectErr bool
	}{
		{
			name:      "4 bits",
			data:      []uint8{0x0F},
			numBits:   4,
			expected:  15,
			expectErr: false,
		},
		{
			name:      "7 bits",
			data:      []uint8{0xFF},
			numBits:   7,
			expected:  127,
			expectErr: false,
		},
		{
			name:      "8 bits",
			data:      []uint8{0xFF},
			numBits:   8,
			expected:  255,
			expectErr: false,
		},
		{
			name:      "10 bits, expecting b1011111111",
			data:      []uint8{0xFF, 0x02},
			numBits:   10,
			expected:  0x02FF,
			expectErr: false,
		},
		{
			name:      "32 bits",
			data:      []uint8{0xFF, 0x02, 0x03, 0xD4},
			numBits:   32,
			expected:  0xD40302FF,
			expectErr: false,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {

			data := bytes.NewReader(tc.data)
			br := NewBitreader(data)

			resp, err := br.ReadBits(tc.numBits)
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
