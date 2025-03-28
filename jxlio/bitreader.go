package jxlio

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"

	"github.com/pektezol/bitreader"
)

type Bitreader struct {
	//stream      io.ReadSeeker

	readerAtOrigin *bitreader.Reader
	reader         *bitreader.Reader
	bitsRead       uint64
	tempIndex      int
	index          uint8
	currentByte    uint8
}

func NewBitreaderWithIndex(in io.ReadSeeker, index int) *Bitreader {

	br := NewBitreader(in)
	br.tempIndex = index
	return br
}

func NewBitreader(in io.ReadSeeker) *Bitreader {

	br := &Bitreader{}
	r := bitreader.NewReader(in, true)
	br.reader = r
	br.readerAtOrigin, _ = r.Fork()
	return br
}

// utter hack to seek about the place. TODO(kpfaulkner) confirm this really works.
func (br *Bitreader) Seek(offset int64, whence int) (int64, error) {

	if whence != io.SeekStart {
		panic("seek boomage")
	}

	newReaderAtOrigin, err := br.readerAtOrigin.Fork()
	if err != nil {
		return 0, err
	}
	newReaderAtOrigin.SkipBytes(uint64(offset))

	br.reader = newReaderAtOrigin
	return 0, nil
}

func (br *Bitreader) AtEnd() bool {

	if _, err := br.reader.ReadRemainingBits(); err != nil {
		return true
	}

	return false
}

// ReadBytesToBuffer
// If part way through a byte then fail. Need to be aligned for this to work.
func (br *Bitreader) ReadBytesToBuffer(buffer []uint8, numBytes uint32) error {

	buffer, err := br.reader.ReadBytesToSlice(uint64(numBytes))
	if err != nil {
		return err
	}
	return nil

}

// read single bit and will cache the current byte we're working on.
func (br *Bitreader) readBit() (uint8, error) {

	b, err := br.reader.ReadBits(1)
	if err != nil {
		return 0, err
	}
	return uint8(b), nil
}

func (br *Bitreader) ReadBits(bits uint32) (uint64, error) {
	if bits > 64 {
		fmt.Printf("snoop\n")
	}
	if bits <= 0 {
		return 0, nil
	}
	return br.reader.ReadBits(uint64(bits))
}

func (br *Bitreader) ReadByteArrayWithOffsetAndLength(buffer []byte, offset int64, length uint32) error {
	if length == 0 {
		return nil
	}

	// remove seek!
	_, err := br.Seek(offset, io.SeekStart)
	if err != nil {
		return err
	}

	err = br.ReadBytesToBuffer(buffer, length)
	if err != nil {
		return err
	}
	return nil
}

func (br *Bitreader) ReadByte() (uint8, error) {
	v, err := br.ReadBits(8)
	if err != nil {
		return 0, err
	}
	return uint8(v), nil
}

func (br *Bitreader) ReadEnum() (int32, error) {
	constant, err := br.ReadU32(0, 0, 1, 0, 2, 4, 18, 6)
	if err != nil {
		return 0, err
	}
	if constant > 63 {
		return 0, errors.New("enum constant > 63")
	}
	return int32(constant), nil
}

func (br *Bitreader) ReadF16() (float32, error) {
	bits16, err := br.ReadBits(16)
	if err != nil {
		return 0, err
	}

	mantissa := bits16 & 0x3FF
	biased_exp := (uint32(bits16) >> 10) & 0x1F
	sign := (bits16 >> 15) & 1
	if biased_exp == 31 {
		return 0, errors.New("illegal infinite/NaN float16")
	}

	if biased_exp == 0 {
		return (1.0 - 2.0*float32(sign)) * float32(mantissa) / 16777216.0, nil
	}

	biased_exp += 127 - 15
	mantissa = mantissa << 13
	sign = sign << 31

	total := uint32(sign) | biased_exp<<23 | uint32(mantissa)
	return math.Float32frombits(total), nil
}

func (br *Bitreader) ReadICCVarint() (int32, error) {
	value := int32(0)
	for shift := 0; shift < 63; shift += 7 {
		b, err := br.ReadBits(8)
		if err != nil {
			return 0, err
		}
		value |= int32(b) & 127 << shift
		if b <= 127 {
			break
		}
	}
	if value > math.MaxInt32 {
		return 0, errors.New("ICC varint overflow")

	}
	return value, nil
}

func (br *Bitreader) ReadU32(c0 int, u0 int, c1 int, u1 int, c2 int, u2 int, c3 int, u3 int) (uint32, error) {
	choice, err := br.ReadBits(2)
	if err != nil {
		return 0, err
	}

	c := []int{c0, c1, c2, c3}
	u := []int{u0, u1, u2, u3}
	b, err := br.ReadBits(uint32(u[choice]))
	if err != nil {
		return 0, err
	}
	return uint32(c[choice]) + uint32(b), nil
}

func (br *Bitreader) ReadBool() (bool, error) {
	v, err := br.readBit()
	if err != nil {
		return false, err
	}
	return v == 1, nil
}

func (br *Bitreader) ReadU64() (uint64, error) {
	index, err := br.ReadBits(2)
	if err != nil {
		return 0, err
	}

	if index == 0 {
		return 0, nil
	}

	if index == 1 {
		b, err := br.ReadBits(4)
		if err != nil {
			return 0, err
		}
		return 1 + uint64(b), nil
	}

	if index == 2 {
		b, err := br.ReadBits(8)
		if err != nil {
			return 0, err
		}
		return 17 + uint64(b), nil
	}

	value2, err := br.ReadBits(12)
	if err != nil {
		return 0, err
	}
	value := uint64(value2)

	shift := 12
	var boolCheck bool
	for {
		if boolCheck, err = br.ReadBool(); err != nil {
			return 0, err
		}
		if !boolCheck {
			break
		}
		if shift == 60 {

			if data, err := br.ReadBits(4); err != nil {
				return 0, err
			} else {
				value |= data << shift
			}
			break
		}
		if data, err := br.ReadBits(8); err != nil {
			return 0, err
		} else {
			value |= data << shift
		}
		shift += 8
	}
	return value, nil
}

func (br *Bitreader) ReadU8() (int, error) {

	b, err := br.ReadBool()
	if err != nil {
		return 0, err
	}

	if !b {
		return 0, nil
	}
	n, err := br.ReadBits(3)
	if err != nil {
		return 0, err
	}
	if n == 0 {
		return 1, nil
	}

	nn, err := br.ReadBits(uint32(n))
	if err != nil {
		return 0, err
	}
	return int(nn + 1<<n), nil
}

func (br *Bitreader) GetBitsCount() uint64 {
	return br.bitsRead
}

func (br *Bitreader) ShowBits(bits int) (uint64, error) {

	tempReader, err := br.reader.Fork()
	if err != nil {
		return 0, err
	}

	if bits > 64 {
		fmt.Printf("snoop\n")
	}

	b, err := tempReader.ReadBits(uint64(bits))
	if err != nil {
		return 0, err
	}
	return b, nil
}

func (br *Bitreader) SkipBits(bits uint32) error {
	return br.reader.SkipBits(uint64(bits))
}

func (br *Bitreader) Skip(bytes uint32) error {
	return br.reader.SkipBytes(uint64(bytes))
}

func (br *Bitreader) ReadBytesUint64(noBytes int) (uint64, error) {
	if noBytes < 1 || noBytes > 8 {
		return 0, fmt.Errorf("number of bytes number should be between 1 and 8.")
	}

	ba := make([]byte, 8)
	err := br.ReadBytesToBuffer(ba, uint32(noBytes))
	if err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint64(ba), nil
}

func (br *Bitreader) ZeroPadToByte() error {

	if br.index == 0 {
		return nil
	}
	remaining := 8 - br.index
	if remaining > 0 {
		_, err := br.ReadBits(uint32(remaining))
		if err != nil {
			return err
		}
	}
	return nil
}

func (br *Bitreader) BitsRead() uint64 {
	return br.bitsRead
}

// JPEGXL spec states unpackedsigned is
// equivalent to u / 2 if u is even, and -(u + 1) / 2 if u is odd
func UnpackSigned(value uint32) int32 {
	if value&1 == 0 {
		return int32(value >> 1)
	}

	return -(int32(value) + 1) >> 1
}

func UnpackSigned64(value uint64) int64 {
	if value&1 == 0 {
		return int64(value >> 1)
	}

	return -(int64(value) + 1) >> 1
}
