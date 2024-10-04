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

// ReadByteArrayWithOffsetAndLength tests reading set of bytes with offset and length.
func TestReadByteArrayWithOffsetAndLength(t *testing.T) {

	for _, tc := range []struct {
		name       string
		data       []uint8
		offset     int64
		length     uint32
		bufferSize int
		expected   []uint8
		expectErr  bool
	}{
		{
			name:       "Read offset 0, length 1",
			data:       []uint8{0x01, 0x02, 0x03, 0x04, 0x05},
			offset:     0,
			length:     1,
			bufferSize: 1,
			expected:   []uint8{0x01},
			expectErr:  false,
		},
		{
			name:       "Read offset 0, length 2",
			data:       []uint8{0x01, 0x02, 0x03, 0x04, 0x05},
			offset:     0,
			length:     2,
			bufferSize: 2,
			expected:   []uint8{0x01, 0x02},
			expectErr:  false,
		},
		{
			name:       "Read offset 1, length 2",
			data:       []uint8{0x01, 0x02, 0x03, 0x04, 0x05},
			offset:     1,
			length:     2,
			bufferSize: 2,
			expected:   []uint8{0x02, 0x03},
			expectErr:  false,
		},
		{
			name:       "Read offset 0, length too large, 0's at end",
			data:       []uint8{0x01, 0x02},
			offset:     0,
			length:     3,
			bufferSize: 3,
			expected:   []uint8{0x01, 0x02, 0x00},
			expectErr:  true,
		},
		{
			name:       "Read offset 0, length 1, no data",
			data:       []uint8{},
			offset:     0,
			length:     1,
			bufferSize: 1,
			expected:   []uint8{0x00},
			expectErr:  true,
		},
		{
			name:       "Read offset too large",
			data:       []uint8{0x01, 0x02},
			offset:     3,
			length:     1,
			bufferSize: 1,
			expected:   []uint8{0x00},
			expectErr:  true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {

			data := bytes.NewReader(tc.data)
			br := NewBitreader(data)

			buffer := make([]uint8, tc.bufferSize)
			err := br.ReadByteArrayWithOffsetAndLength(buffer, tc.offset, tc.length)
			if err != nil && !tc.expectErr {
				t.Errorf("got error when none was expected : %v", err)
			}

			if err == nil && tc.expectErr {
				t.Errorf("expected error but got none")
			}

			if !bytes.Equal(tc.expected, buffer) {
				t.Errorf("expected %v but got %v", tc.expected, buffer)
			}
		})
	}
}
