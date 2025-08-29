package testcommon

import "fmt"

// FakeBitReader is a mock implementation of a bit reader for testing purposes.
// Will populate the functions as required
type FakeBitReader struct {
	ReadF16Data  []float32
	ReadBoolData []bool
}

func (fbr *FakeBitReader) ReadBytesToBuffer(buffer []uint8, numBytes uint32) error {
	//TODO implement me
	panic("implement me")
}

func (fbr *FakeBitReader) ReadBits(bits uint32) (uint64, error) {
	//TODO implement me
	panic("implement me")
}

func (fbr *FakeBitReader) ReadByteArrayWithOffsetAndLength(buffer []byte, offset int64, length uint32) error {
	//TODO implement me
	panic("implement me")
}

func (fbr *FakeBitReader) ReadByte() (uint8, error) {
	//TODO implement me
	panic("implement me")
}

func (fbr *FakeBitReader) ReadEnum() (int32, error) {
	//TODO implement me
	panic("implement me")
}

func (fbr *FakeBitReader) ReadICCVarint() (int32, error) {
	//TODO implement me
	panic("implement me")
}

func (fbr *FakeBitReader) ReadU32(c0 int, u0 int, c1 int, u1 int, c2 int, u2 int, c3 int, u3 int) (uint32, error) {
	//TODO implement me
	panic("implement me")
}

func (fbr *FakeBitReader) ReadU8() (int, error) {
	//TODO implement me
	panic("implement me")
}

func (fbr *FakeBitReader) GetBitsCount() uint64 {
	//TODO implement me
	panic("implement me")
}

func (fbr *FakeBitReader) ZeroPadToByte() error {
	//TODO implement me
	panic("implement me")
}

func NewFakeBitReader() *FakeBitReader {
	return &FakeBitReader{}
}

func (fbr *FakeBitReader) ReadBool() (bool, error) {
	if len(fbr.ReadBoolData) > 0 {
		val := fbr.ReadBoolData[0]
		fbr.ReadBoolData = fbr.ReadBoolData[1:]
		return val, nil
	}
	return false, fmt.Errorf("No more data")
}

func (fbr *FakeBitReader) ReadU64() (uint64, error) {
	return 0, nil
}

func (fbr *FakeBitReader) ReadF16() (float32, error) {
	if len(fbr.ReadF16Data) > 0 {
		val := fbr.ReadF16Data[0]
		fbr.ReadF16Data = fbr.ReadF16Data[1:]
		return val, nil
	}
	return 0, fmt.Errorf("No more data")
}

func (fbr *FakeBitReader) ReadU16() (uint16, error) {
	return 0, nil
}

func (fbr *FakeBitReader) AlignToByte() error {
	return nil
}

func (fbr *FakeBitReader) BitsRead() uint64 {
	return 0
}

func (fbr *FakeBitReader) BytesRead() uint64 {
	return 0
}

func (fbr *FakeBitReader) SetBytePosition(pos uint64) error {
	return nil
}

func (fbr *FakeBitReader) BytePosition() uint64 {
	return 0
}

func (fbr *FakeBitReader) Close() error {
	return nil
}

func (fbr *FakeBitReader) ReadBytes(numBytes uint32) ([]byte, error) {
	return make([]byte, numBytes), nil
}

func (fbr *FakeBitReader) ReadU32Array(length uint32, a, b, c, d, e, f, g, h uint32) ([]uint32, error) {
	return make([]uint32, length), nil
}

func (fbr *FakeBitReader) ReadU8Array(length uint32) ([]uint8, error) {
	return make([]uint8, length), nil
}
func (fbr *FakeBitReader) ReadF16Array(length uint32) ([]float32, error) {
	return make([]float32, length), nil
}

func (fbr *FakeBitReader) Seek(offset int64, whence int) (int64, error) {
	return 0, nil
}

func (fbr *FakeBitReader) ReadBitsNoAdvance(bits uint32) (uint32, error) {
	return 0, nil
}

func (fbr *FakeBitReader) SkipBits(bits uint32) error {
	return nil
}

func (fbr *FakeBitReader) AtEnd() bool {
	return false
}

func (fbr *FakeBitReader) ShowBits(bits int) (uint64, error) {
	return 0, nil
}

func (fbr *FakeBitReader) Skip(bytes uint32) (int64, error) {
	return int64(bytes), nil
}

func (fbr *FakeBitReader) ReadBytesUint64(noBytes int) (uint64, error) {
	return 0, nil
}

func (fbr *FakeBitReader) ReadBytesInt(noBytes int) (int64, error) {
	return 0, nil
}

func (fbr *FakeBitReader) ReadBytesFloat32(noBytes int) (float32, error) {
	return 0, nil
}

func (fbr *FakeBitReader) ReadBytesFloat64(noBytes int) (float64, error) {
	return 0, nil
}

func (fbr *FakeBitReader) ReadBytesIntArray(noBytes int, length uint32) ([]int64, error) {
	return make([]int64, length), nil
}

func (fbr *FakeBitReader) Reset() error {
	return nil
}
