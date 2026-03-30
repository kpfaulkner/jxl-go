package testcommon

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBitWriter_WriteBit(t *testing.T) {
	bw := NewBitWriter()
	bw.WriteBit(1)
	bw.WriteBit(0)
	bw.WriteBit(1)
	bw.WriteBit(1)

	bytes := bw.Bytes()
	// 1 + 0*2 + 1*4 + 1*8 = 13 (0xD)
	assert.Equal(t, uint8(13), bytes[0])
}

func TestBitWriter_WriteBits(t *testing.T) {
	bw := NewBitWriter()
	bw.WriteBits(10, 4) // 1010 in bits (binary 10)

	bytes := bw.Bytes()
	assert.Equal(t, uint8(10), bytes[0])
}

func TestBitWriter_WriteU8(t *testing.T) {
	bw := NewBitWriter()
	bw.WriteU8(4)
	// WriteU8(4):
	// val != 0 -> WriteBit(1)
	// (1<<(n+1)) <= 4 -> n=2 (1<<3 = 8 > 4)
	// WriteBits(2, 3) -> 010 (bits for 2)
	// WriteBits(4-(1<<2), 2) -> WriteBits(0, 2) -> 00
	// Result: 1 + 010 + 00 = 1 0 1 0 0 0 = 1 + 4 = 5?
	// Wait, let's re-calculate:
	// bit 0: 1
	// bits 1-3: 0, 1, 0 (2 in 3 bits)
	// bits 4-5: 0, 0 (0 in 2 bits)
	// Combined bits: 101000 (least significant first) -> 1 + 0*2 + 1*4 + 0*8 + 0*16 + 0*32 = 5

	bytes := bw.Bytes()
	assert.Equal(t, uint8(5), bytes[0])
}

func TestBitWriter_FullByte(t *testing.T) {
	bw := NewBitWriter()
	bw.WriteBits(0xFF, 8)
	bw.WriteBit(1)

	bytes := bw.Bytes()
	assert.Equal(t, 2, len(bytes))
	assert.Equal(t, uint8(0xFF), bytes[0])
	assert.Equal(t, uint8(1), bytes[1])
}
