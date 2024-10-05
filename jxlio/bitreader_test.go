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

// TestReadByte tests reading a single byte...
func TestReadByte(t *testing.T) {

	for _, tc := range []struct {
		name      string
		data      []uint8
		expected  uint8
		expectErr bool
	}{
		{
			name:      "Read byte",
			data:      []uint8{0x01, 0x02, 0x03, 0x04, 0x05},
			expected:  0x01,
			expectErr: false,
		},
		{
			name:      "Read byte no data",
			data:      []uint8{},
			expected:  0x00,
			expectErr: true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {

			data := bytes.NewReader(tc.data)
			br := NewBitreader(data)

			resp, err := br.ReadByte()
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

// TestReadU32 tests reading of U32 type
// First byte of the test data will be used for choice response.
// Choice response should only be 2 bits long, so will need to read the first 6 bits and discard.
func TestReadU32(t *testing.T) {

	for _, tc := range []struct {
		name      string
		data      []uint8
		expected  uint32
		expectErr bool
	}{
		{
			// First byte will be used for choiceResponse. Given choice response is 2 bits, we need to skip the first 6 bits
			// This is why its set to 0x40. First 6 bits are 0, but then the 7th is 1... which is wanted for this test.
			name:      "ReadU32 success",
			data:      []uint8{0x40, 0x01, 0x02, 0x03, 0x04},
			expected:  514,
			expectErr: false,
		},
		{
			name:      "ReadU32 not enough data",
			data:      []uint8{0x40, 0x01},
			expected:  0,
			expectErr: true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {

			data := bytes.NewReader(tc.data)
			br := NewBitreader(data)
			br.SkipBits(6)

			resp, err := br.ReadU32(1, 9, 1, 13, 1, 18, 1, 30)
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

// TestReadU64 tests reading of U64 type
// NOT COMPLETE... (unused in library yet, except for response of 0).
func TestReadU64(t *testing.T) {

	for _, tc := range []struct {
		name      string
		data      []uint8
		expected  uint64
		expectErr bool
	}{
		{
			// First byte will be used for choiceResponse. Given choice response is 2 bits, we need to skip the first 6 bits
			// This is why its set to 0x40. First 6 bits are 0, but then the 7th is 1... which is wanted for this test.
			name:      "ReadU64 read 0",
			data:      []uint8{0x0, 0x01, 0x02, 0x03, 0x04},
			expected:  0,
			expectErr: false,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {

			data := bytes.NewReader(tc.data)
			br := NewBitreader(data)
			err := br.SkipBits(6)
			if err != nil {
				t.Errorf("error skipping bits : %v", err)
			}

			resp, err := br.ReadU64()
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

func TestReadU8(t *testing.T) {

	for _, tc := range []struct {
		name      string
		data      []uint8
		expected  int
		expectErr bool
	}{
		{
			name:      "ReadU8 initial bit is 0, result 0",
			data:      []uint8{0x00},
			expected:  0,
			expectErr: false,
		},
		{
			name:      "ReadU8 initial bit is 1, next 3 are 7 then following 7 bits are 1010101",
			data:      []uint8{0xF0, 0x55},
			expected:  213,
			expectErr: false,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {

			data := bytes.NewReader(tc.data)
			br := NewBitreader(data)
			err := br.SkipBits(4)

			resp, err := br.ReadU8()
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

func TestSkipBits(t *testing.T) {

	for _, tc := range []struct {
		name            string
		data            []uint8
		numBitsToSkip   uint32
		expectedNextBit uint8
		expectErr       bool
	}{
		{
			name:          "Skip but no data",
			data:          []uint8{},
			numBitsToSkip: 1,
			expectErr:     true,
		},
		{
			name:            "Skip 1, next read should be 0",
			data:            []uint8{0x01},
			numBitsToSkip:   1,
			expectedNextBit: 0,
			expectErr:       false,
		},
		{
			name:            "Skip 1, next read should be 1",
			data:            []uint8{0x03},
			numBitsToSkip:   1,
			expectedNextBit: 1,
			expectErr:       false,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {

			data := bytes.NewReader(tc.data)
			br := NewBitreader(data)

			err := br.SkipBits(tc.numBitsToSkip)
			if err != nil && !tc.expectErr {
				t.Errorf("got error when none was expected : %v", err)
				return
			}

			if err != nil && tc.expectErr {
				// all good return
				return
			}

			if err == nil && tc.expectErr {
				t.Errorf("expected error but got none")
			}

			resp, err := br.readBit()
			if err != nil {
				t.Errorf("error reading bit : %v", err)
			}

			if resp != tc.expectedNextBit {
				t.Errorf("expected %v but got %v", tc.expectedNextBit, resp)
			}

		})
	}
}
