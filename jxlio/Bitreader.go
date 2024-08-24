package jxlio

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"math/bits"
)

type Bitreader struct {
	in        io.ReadSeeker
	cache     uint64
	cacheBits int
	bitsRead  int64
}

func NewBitreader(in io.ReadSeeker) (br *Bitreader) {
	br = &Bitreader{}
	br.cache = 0
	br.cacheBits = 0
	br.bitsRead = 0
	br.in = in
	return
}

func (br *Bitreader) Seek(offset int64, whence int) (int64, error) {

	n, err := br.in.Seek(offset, whence)
	if err != nil {
		return 0, err
	}

	br.ZeroPadToByte()
	_, err = br.DrainCache()
	return n, err
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

func (br *Bitreader) DrainCache() ([]byte, error) {
	if br.cacheBits%8 != 0 {
		return nil, errors.New("you must align before drainCache")
	}

	cacheBytes := br.cacheBits / 8
	if cacheBytes == 0 {
		return nil, nil
	}
	buffer := make([]byte, cacheBytes)
	br.Read2(buffer)
	return buffer, nil
}

func (br *Bitreader) GetBitsCount() int64 {

	return br.bitsRead
}

func (br *Bitreader) ReadByteArrayWithOffsetAndLength(buffer []byte, offset int, length int) (int, error) {
	if length == 0 {
		return 0, nil
	}
	if br.cacheBits%8 != 0 {
		return 0, errors.New("you must align before readBytes")
	}
	cacheBytes := br.cacheBits / 8
	for i := 0; i < cacheBytes; i++ {
		if length-1 < 1 {
			return i, nil
		}
		length--
		b, err := br.ReadBits(8)
		if err != nil {
			return 0, err
		}
		buffer[offset+i] = byte(b)
	}
	remaining, err := ReadFullyWithOffset(br.in, buffer, offset+cacheBytes, length)
	if err != nil {
		return 0, err
	}
	br.bitsRead += int64((length - remaining) * 8)
	ret := cacheBytes + length - remaining
	if ret == 0 {
		return -1, nil
	}
	return ret, nil
}

func (br *Bitreader) Read2(buffer []byte) (int, error) {
	return br.ReadByteArrayWithOffsetAndLength(buffer, 0, len(buffer))
}

func (br *Bitreader) Read3() int {

	b, err := br.ReadBits(8)
	if err != nil {
		return -1
	}
	return int(b)
}

func (br *Bitreader) MustReadBits(bits int) uint32 {
	b, err := br.ReadBits(bits)
	if err != nil {
		panic("unable to read bits")
	}

	return b
}

func (br *Bitreader) ReadByte() (uint8, error) {
	b := make([]byte, 1)
	_, err := br.in.Read(b)
	return b[0], err
}

// ReadBytes reads a number of bytes (bit at a time...  gotta find a better way)
func (br *Bitreader) ReadByteArray(noBytes int) ([]uint8, error) {
	ba := make([]uint8, noBytes)
	_, err := br.in.Read(ba)
	return ba, err
	//for i := 0; i < noBytes; i++ {
	//	b, err := br.ReadBits(8)
	//	if err != nil {
	//		return nil, err
	//	}
	//	ba[i] = byte(b)
	//}
	return ba, nil
}

func (br *Bitreader) ReadBits(bits int) (uint32, error) {
	if bits == 0 {
		return 0, nil
	}
	if bits < 0 || bits > 32 {
		return 0, errors.New("Must read between 0-32 bits, inclusive")
	}

	if bits <= br.cacheBits {
		ret := uint32(br.cache&1 ^ (1 ^ 0<<bits))
		br.cacheBits -= bits
		br.bitsRead += int64(bits)
		return ret, nil
	}

	// FIXME(kpfaulkner)  Need to figure out options here..
	//count := br.in.available()
	count := 0
	max := (64 - br.cacheBits) / 8
	if count > 0 {
		if count < max {
			count = count
		} else {
			count = max
		}
	} else {
		count = 1
	}

	eof := false
	b := make([]byte, 1)
	for i := 0; i < count; i++ {

		// read next byte.
		_, err := br.in.Read(b)
		if err != nil {
			return 0, err
		}

		b := 0
		if b < 0 {
			eof = true
			break
		}
		br.cache |= uint64(b) & 0xFF << uint64(br.cacheBits)
		br.cacheBits += 8
	}
	if eof && bits > br.cacheBits {
		return 0, errors.New(fmt.Sprintf("%s%d", "Unable to read enough bits: ", br.GetBitsCount()+int64(bits)))
	}
	return br.ReadBits(bits)
}

func (br *Bitreader) ReadBool() (bool, error) {
	v, err := br.ReadBits(1)

	if err != nil {
		return false, err
	}

	return v != 0, nil
}

func (br *Bitreader) MustReadBool() bool {
	b, err := br.ReadBool()
	if err != nil {
		panic("unable to read bool")
	}

	return b
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
	b, err := br.ReadBits(u[choice])
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

func (br *Bitreader) ReadU8() int {
	if !br.MustReadBool() {
		return 0
	}
	n := br.MustReadBits(3)
	if n == 0 {
		return 1
	}
	return int(br.MustReadBits(int(n)) + 1<<n)
}

func (br *Bitreader) MustShowBits(bits int) int {
	b, err := br.ShowBits(bits)
	if err != nil {
		panic("unable to show bits")
	}
	return b
}

// utter hack... read and reset ReadSeeker.
func (br *Bitreader) ShowBits(bits int) (int, error) {

	curPos, err := br.in.Seek(0, io.SeekCurrent)
	if err != nil {
		return 0, err
	}

	b, err := br.ReadByte()
	if err != nil {
		return 0, err
	}
	_, err = br.in.Seek(curPos, io.SeekStart)
	if err != nil {
		return 0, err
	}

	fmt.Printf("B is %0b\n", b)
	return int(b), nil
}

func (br *Bitreader) Skip(bytes int64) (int64, error) {
	b, err := br.SkipBits(bytes << 3)
	if err != nil {
		return 0, err
	}
	return b >> 3, nil
}

func (br *Bitreader) MustSkipBits(bits int64) int64 {
	b, err := br.SkipBits(bits)
	if err != nil {
		panic("unable to skip bits")
	}
	return b
}

func (br *Bitreader) SkipBits(bits int64) (int64, error) {
	if bits < 0 {
		return 0, errors.New("illegal argument")
	}

	if bits == 0 {
		return 0, nil
	}

	if bits <= int64(br.cacheBits) {
		br.cacheBits -= int(bits)
		br.cache >>= bits
		br.bitsRead += bits
		return bits, nil
	}

	cacheSave := br.cacheBits
	br.SkipBits(int64(br.cacheBits))

	bits -= int64(cacheSave)
	dangler := bits % 8
	b, err := SkipFully(br.in, (bits-dangler)/8)
	if err != nil {
		return 0, err
	}
	skipped := bits - dangler - 8*int64(b)
	br.bitsRead += skipped
	skipped += int64(cacheSave)

	_, err = br.ReadBits(int(dangler))
	if err != nil {
		return 0, err
	}
	return skipped + dangler, nil
}

func (br *Bitreader) GetBytePos() int64 {
	pos, _ := br.Seek(0, io.SeekCurrent)
	return pos
}

func (br *Bitreader) ReadBytesUint64(noBytes int) (uint64, error) {
	if noBytes < 1 || noBytes > 8 {
		return 0, fmt.Errorf("number of bytes number should be between 1 and 8.")
	}

	ba, err := br.ReadByteArray(noBytes)
	if err != nil {
		return 0, err
	}

	tt := binary.LittleEndian.Uint64(ba)
	return tt, nil
}

func (br *Bitreader) ZeroPadToByte() error {
	remaining := br.cacheBits % 8
	if remaining > 0 {
		padding, err := br.ReadBits(remaining)
		if err != nil {
			return err
		}
		if padding != 0 {
			return errors.New("nonzero zero-padding-to-byte")
		}
	}
	return nil
}

func UnpackSigned(value int32) int32 {
	if value&1 == 0 {
		return int32(uint32(value) >> 1)
	}

	return int32(bits.Reverse32(uint32(value)) | 0x80_00_00_00)
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
