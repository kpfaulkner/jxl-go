package jxlio

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
)

type Bitreader struct {
	// stream/reader we're using most of the time
	stream io.ReadSeeker

	index       uint8
	currentByte uint8
	tempIndex   int
	bitsRead    uint64
}

func NewBitreaderWithIndex(in io.ReadSeeker, index int) *Bitreader {

	br := NewBitreader(in)
	br.tempIndex = index
	return br
}

func NewBitreader(in io.ReadSeeker) *Bitreader {

	br := &Bitreader{}
	br.stream = in
	return br
}

// utter hack to seek about the place. TODO(kpfaulkner) confirm this really works.
func (br *Bitreader) Seek(offset int64, whence int) (int64, error) {
	n, err := br.stream.Seek(offset, whence)
	if err != nil {
		return 0, err
	}
	return n, err
}

func (br *Bitreader) Reset() error {

	_, err := br.stream.Seek(0, io.SeekStart)
	if err != nil {
		return err
	}

	// reset tracking
	br.index = 0
	br.currentByte = 0
	return nil
}

func (br *Bitreader) AtEnd() bool {

	_, err := br.ShowBits(1)
	if err != nil {
		return true
	}
	return false
}

func (br *Bitreader) Close() {

	// FIXME(kpfaulkner)
	//br.in.Close()
	panic("Bitreader.Close() not implemented")
}

// loop one byte at a time and read.... not efficient but will rework later FIXME(kpfaulkner)
// Most of the time we probably just want to fill the buffer... but have seen that in some cases
// we might just want to partially populate the buffer. Hence the numBytes parameter.
func (br *Bitreader) ReadBytesToBufferOrig(buffer []uint8, numBytes uint32) error {
	for i := uint32(0); i < numBytes; i++ {
		b, err := br.ReadBits(8)
		if err != nil {
			return err
		}
		buffer[i] = uint8(b)
	}
	return nil
}

// ReadBytesToBuffer
// If part way through a byte then fail. Need to be aligned for this to work.
func (br *Bitreader) ReadBytesToBuffer(buffer []uint8, numBytes uint32) error {

	if br.index != 0 {
		return errors.New("Bitreader cache not aligned")
	}

	n, err := br.stream.Read(buffer[:numBytes])
	if err != nil {
		return err
	}

	if n != int(numBytes) {
		panic("boom")
		return errors.New("unable to read all bytes")
	}
	return nil
}

// read single bit and will cache the current byte we're working on.
// Need to look more at how JXLatte does it..
func (br *Bitreader) readBit() (uint8, error) {
	if br.index == 0 {
		buffer := make([]byte, 1)
		_, err := br.stream.Read(buffer)
		if err != nil {
			return 0, err
		}
		br.currentByte = buffer[0]
	}

	v := (br.currentByte & (1 << br.index)) != 0
	br.index = (br.index + 1) % 8

	br.bitsRead++
	if v {
		return 1, nil
	} else {
		return 0, nil
	}
}

func (br *Bitreader) ReadBits(bits uint32) (uint64, error) {

	if bits == 0 {
		return 0, nil
	}

	if bits < 1 || bits > 64 {

		return 0, errors.New("num bits must be between 1 and 64")
	}
	var v uint64
	for i := uint32(0); i < bits; i++ {
		bit, err := br.readBit()
		if err != nil {
			return 0, err
		}
		v |= uint64(bit) << i

	}
	return v, nil
}

func (br *Bitreader) ReadByteArrayWithOffsetAndLength(buffer []byte, offset int64, length uint32) error {
	if length == 0 {
		return nil
	}

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

func (br *Bitreader) MustReadEnum() int32 {
	v, err := br.ReadEnum()
	if err != nil {
		panic("MustReadEnum panic")
	}
	return v
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
	biased_exp := uint32(bits16) >> 10 & 0x1F
	if biased_exp == 31 {
		return 0, errors.New("illegal infinite/NaN float16")
	}

	biased_exp += 127 - 15
	sign := bits16 & 0x8000 << 16
	total := uint32(sign) | biased_exp<<23 | uint32(mantissa)
	return math.Float32frombits(total), nil
}

func (br *Bitreader) MustReadF16() float32 {
	v, err := br.ReadF16()
	if err != nil {
		panic("unable to ReadF16")
	}
	return v
}

func (br *Bitreader) ReadICCVarint() (int, error) {
	value := 0
	for shift := 0; shift < 63; shift += 7 {
		b, err := br.ReadBits(8)
		if err != nil {
			return 0, err
		}
		value |= int(b) & 127 << shift
		if b <= 127 {
			break
		}
	}
	if value > math.MaxInt32 {
		return 0, errors.New("ICC varint overflow")

	}
	return value, nil
}

func (br *Bitreader) MustReadU32(c0 int, u0 int, c1 int, u1 int, c2 int, u2 int, c3 int, u3 int) uint32 {

	v, err := br.ReadU32(c0, u0, c1, u1, c2, u2, c3, u3)
	if err != nil {
		panic("unable to read U32")
	}
	return v
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

func (br *Bitreader) MustReadU64() uint64 {
	v, err := br.ReadU64()
	if err != nil {
		panic("unable to read U64")
	}
	return v
}

func (br *Bitreader) ReadBool() (bool, error) {
	v, err := br.readBit()
	if err != nil {
		return false, err
	}
	return v == 1, nil
}

func (br *Bitreader) MustReadBool() bool {
	ok, err := br.ReadBool()
	if err != nil {
		// really need to remove these panics and force error inspection
		panic(err)
	}
	return ok
}

func (br *Bitreader) MustReadBits(bits uint32) uint64 {
	v, err := br.ReadBits(bits)
	if err != nil {

		// really need to remove these panics and force error inspection
		panic(err)
	}
	return v
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
	for br.MustReadBool() {
		if shift == 60 {
			value |= uint64(br.MustReadBits(4)) << shift
			break
		}
		value |= uint64(br.MustReadBits(8)) << shift
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

func (br *Bitreader) ShowBits(bits int) (uint64, error) {

	curPos, err := br.Seek(0, io.SeekCurrent)
	if err != nil {
		return 0, err
	}
	oldCur := br.currentByte
	oldIndex := br.index
	oldBitsRead := br.bitsRead

	b, err := br.ReadBits(uint32(bits))
	if err != nil {
		return 0, err
	}

	_, err = br.Seek(curPos, io.SeekStart)
	if err != nil {
		return 0, err
	}
	br.currentByte = oldCur
	br.index = oldIndex
	br.bitsRead = oldBitsRead

	return b, nil
}

func (br *Bitreader) SkipBits(bits uint32) error {
	numBytes := bits / 8
	if numBytes > 0 {
		buffer := make([]byte, numBytes)
		_, err := br.stream.Read(buffer)
		if err != nil {
			return err
		}
		br.currentByte = buffer[numBytes-1]
	}

	// read bits so we can keep track of where we are.
	for i := numBytes * 8; i < bits; i++ {
		_, err := br.readBit()
		if err != nil {
			return err
		}
	}
	return nil
}

func (br *Bitreader) Skip(bytes uint32) (int64, error) {
	err := br.SkipBits(bytes << 3)
	if err != nil {
		return 0, err
	}
	return int64(bytes), nil
}

func (br *Bitreader) GetBytePos() int64 {
	pos, _ := br.Seek(0, io.SeekCurrent)
	return pos
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
	//remaining := br.index % 8
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

func UnpackSignedU32(x uint32) int32 {

	panic("boom")
	base := int32(x >> 1)
	if x&1 == 0 {
		return base
	} else {
		return -base - 1
	}
}
