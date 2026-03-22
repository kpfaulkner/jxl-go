package frame

import (
	"bytes"
	"testing"

	"github.com/kpfaulkner/jxl-go/jxlio"
	"github.com/stretchr/testify/assert"
)

type BitWriter struct {
	data []byte
	byte byte
	bits int
}

func (bw *BitWriter) WriteBit(bit uint8) {
	if bit != 0 {
		bw.byte |= (1 << bw.bits)
	}
	bw.bits++
	if bw.bits == 8 {
		bw.data = append(bw.data, bw.byte)
		bw.byte = 0
		bw.bits = 0
	}
}

func (bw *BitWriter) WriteBits(val uint64, numBits int) {
	for i := 0; i < numBits; i++ {
		bw.WriteBit(uint8((val >> i) & 1))
	}
}

func (bw *BitWriter) WriteU32(val uint32, dist []int) {
	// Simple implementation for fixed distributions if needed, 
	// but Quantizer uses ReadU32 with specific distributions.
	// We need to match the distributions used in Quantizer.go:
	// globalScale: 1, 11, 2049, 11, 4097, 12, 8193, 16
	// quantLF: 16, 0, 1, 5, 1, 8, 1, 16
	
	// Since implementing a full U32 writer matching the distribution is complex,
	// we can try to craft bits that hit simple cases if possible, or use the 
	// existing bit writer logic if we understand the U32 encoding well enough.
	// For now, let's use the simplest path if possible.
	
	// U32 distribution logic (from bitreader.go likely):
	// Read bits for selector (2 bits usually?)
	// Then read bits for offset.
	// Then read extra bits.
	
	// Let's use 1 as globalScale. 
	// selector 0 (val 1) -> 0 bits offset, 0 bits extra?
	// globalScale dist: val=1, bits=11 -> ???
	// Wait, ReadU32(val0, bits0, val1, bits1, val2, bits2, val3, bits3)
	// selector is 2 bits.
	// 00 -> val0 + ReadBits(bits0)
	// 01 -> val1 + ReadBits(bits1)
	// 10 -> val2 + ReadBits(bits2)
	// 11 -> val3 + ReadBits(bits3)
	
	// To get globalScale = 1:
	// Selector 00. val0=1, bits0=11.
	// We need ReadBits(11) to be 0.
	
	// To get quantLF = 16:
	// Selector 00. val0=16, bits0=0.
	// We need ReadBits(0) -> 0.
}

func (bw *BitWriter) Bytes() []byte {
	if bw.bits > 0 {
		return append(bw.data, bw.byte)
	}
	return bw.data
}

func TestNewQuantizerWithReader(t *testing.T) {
	bw := &BitWriter{}
	
	// globalScale = 1
	bw.WriteBits(0, 2) // Selector 00
	bw.WriteBits(0, 11) // Offset 0
	
	// quantLF = 16
	bw.WriteBits(0, 2) // Selector 00
	// Offset 0 (0 bits)
	
	br := jxlio.NewBitStreamReader(bytes.NewReader(bw.Bytes()))
	lfDequant := []float32{1.0, 1.0, 1.0}
	
	q, err := NewQuantizerWithReader(br, lfDequant)
	assert.NoError(t, err)
	assert.NotNil(t, q)
	assert.Equal(t, uint32(1), q.globalScale)
	assert.Equal(t, uint32(16), q.quantLF)
	
	// Check scaledDequant
	// (1<<16) * 1.0 / (1 * 16) = 65536 / 16 = 4096
	assert.Equal(t, float32(4096.0), q.scaledDequant[0])
}
