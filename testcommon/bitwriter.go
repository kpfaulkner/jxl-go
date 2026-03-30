package testcommon

// BitWriter is a helper to write bits into a byte slice.
// Used for crafting mock bitstreams in tests.
// It allows for fine-grained control over bit-level data construction,
// which is essential for testing low-level decoders and readers.
type BitWriter struct {
	data []byte
	curr byte
	bits int
}

// NewBitWriter initializes and returns a pointer to a new BitWriter instance.
func NewBitWriter() *BitWriter {
	return &BitWriter{}
}

// WriteBit appends a single bit to the current bitstream.
// Bits are written to the current byte starting from the least significant bit (LSB).
// When a full byte (8 bits) is reached, it is appended to the internal data slice.
func (bw *BitWriter) WriteBit(bit uint8) {
	if bit != 0 {
		bw.curr |= (1 << bw.bits)
	}
	bw.bits++
	if bw.bits == 8 {
		bw.data = append(bw.data, bw.curr)
		bw.curr = 0
		bw.bits = 0
	}
}

// WriteBits appends multiple bits to the bitstream.
// The bits are extracted from the provided uint64 value,
// starting from the LSB and moving towards the most significant bit.
func (bw *BitWriter) WriteBits(val uint64, numBits int) {
	for i := 0; i < numBits; i++ {
		bw.WriteBit(uint8((val >> i) & 1))
	}
}

// WriteU8 implements a specific JXL-style variable-length encoding for small integers.
// If the value is 0, it writes a single 0 bit.
// If the value is non-zero, it writes a 1 bit, followed by a 3-bit exponent (n),
// and finally the remaining bits to represent the value (val - 2^n).
func (bw *BitWriter) WriteU8(val int) {
	if val == 0 {
		bw.WriteBit(0)
		return
	}
	bw.WriteBit(1)
	n := 0
	for (1 << (n + 1)) <= val {
		n++
	}
	bw.WriteBits(uint64(n), 3)
	bw.WriteBits(uint64(val-(1<<n)), n)
}

// Bytes returns the accumulated byte slice representing the bitstream.
// If there are remaining bits that haven't formed a full byte,
// the current partially-filled byte is appended before returning the slice.
func (bw *BitWriter) Bytes() []byte {
	if bw.bits > 0 {
		return append(bw.data, bw.curr)
	}
	return bw.data
}
