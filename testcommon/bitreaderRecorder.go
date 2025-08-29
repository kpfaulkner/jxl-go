package testcommon

import (
	"github.com/kpfaulkner/jxl-go/jxlio"
)

type BitReaderRecorder struct {
	ReadF16Data                          []float32
	ReadBoolData                         []bool
	ReadBytesData                        [][]byte
	ReadBitsData                         []uint64
	ReadByteArrayWithOffsetAndLengthData [][]byte
	ReadByteData                         []byte
	ReadEnumData                         []int32
	ReadICCVarintData                    []int32
	ReadU64Data                          []uint64
	ReadU32Data                          []uint32
	ReadU16Data                          []uint16
	ReadU8Data                           []int
	GetBitsCountData                     []uint64
	ZeroPadToByteData                    []error
	ReadBoolCallCount                    int
	ShowBitsData                         []uint64
	SkipData                             []int64
	ReadBytesUint64Data                  []uint64
	ReadBytesToBufferData                [][]uint8
	SeekData                             []int64
	AtEndData                            []bool

	realBitReader jxlio.BitReader
}

func NewBitReaderRecorder(realBitReader jxlio.BitReader) *BitReaderRecorder {
	br := &BitReaderRecorder{
		realBitReader: realBitReader,
	}

	return br
}

func (fbr *BitReaderRecorder) ReadBytesToBuffer(buffer []uint8, numBytes uint32) error {
	err := fbr.realBitReader.ReadBytesToBuffer(buffer, numBytes)
	if err != nil {
		return err
	}

	fbr.ReadBytesToBufferData = append(fbr.ReadBytesToBufferData, buffer)
	return nil
}

func (fbr *BitReaderRecorder) ReadBits(bits uint32) (uint64, error) {

	res, err := fbr.realBitReader.ReadBits(bits)
	if err != nil {
		return 0, err
	}
	fbr.ReadBitsData = append(fbr.ReadBitsData, res)
	return res, nil
}

func (fbr *BitReaderRecorder) ReadByteArrayWithOffsetAndLength(buffer []byte, offset int64, length uint32) error {
	err := fbr.realBitReader.ReadByteArrayWithOffsetAndLength(buffer, offset, length)
	if err != nil {
		return err
	}

	fbr.ReadByteArrayWithOffsetAndLengthData = append(fbr.ReadByteArrayWithOffsetAndLengthData, buffer)
	return nil
}

func (fbr *BitReaderRecorder) ReadByte() (uint8, error) {
	res, err := fbr.realBitReader.ReadByte()
	if err != nil {
		return 0, err
	}
	fbr.ReadByteData = append(fbr.ReadByteData, res)
	return res, nil
}

func (fbr *BitReaderRecorder) ReadEnum() (int32, error) {
	res, err := fbr.realBitReader.ReadEnum()
	if err != nil {
		return 0, err
	}
	fbr.ReadEnumData = append(fbr.ReadEnumData, res)
	return res, nil
}

func (fbr *BitReaderRecorder) ReadICCVarint() (int32, error) {
	res, err := fbr.realBitReader.ReadICCVarint()
	if err != nil {
		return 0, err
	}
	fbr.ReadICCVarintData = append(fbr.ReadICCVarintData, res)
	return res, nil
}

func (fbr *BitReaderRecorder) ReadU32(c0 int, u0 int, c1 int, u1 int, c2 int, u2 int, c3 int, u3 int) (uint32, error) {
	res, err := fbr.realBitReader.ReadU32(c0, u0, c1, u1, c2, u2, c3, u3)
	if err != nil {
		return 0, err
	}
	fbr.ReadU32Data = append(fbr.ReadU32Data, res)
	return res, nil
}

func (fbr *BitReaderRecorder) ReadU8() (int, error) {
	res, err := fbr.realBitReader.ReadU8()
	if err != nil {
		return 0, err
	}
	fbr.ReadU8Data = append(fbr.ReadU8Data, res)
	return res, nil
}

func (fbr *BitReaderRecorder) GetBitsCount() uint64 {
	//TODO implement me
	panic("implement me")
}

func (fbr *BitReaderRecorder) ZeroPadToByte() error {
	err := fbr.realBitReader.ZeroPadToByte()
	return err
}

func (fbr *BitReaderRecorder) ReadBool() (bool, error) {
	res, err := fbr.realBitReader.ReadBool()
	if err != nil {
		return false, err
	}
	fbr.ReadBoolData = append(fbr.ReadBoolData, res)
	return res, nil
}

func (fbr *BitReaderRecorder) ReadU64() (uint64, error) {
	res, err := fbr.realBitReader.ReadU64()
	if err != nil {
		return 0, err
	}
	fbr.ReadU64Data = append(fbr.ReadU64Data, res)
	return res, nil
}

func (fbr *BitReaderRecorder) ReadF16() (float32, error) {
	res, err := fbr.realBitReader.ReadF16()
	if err != nil {
		return 0, err
	}
	fbr.ReadF16Data = append(fbr.ReadF16Data, res)
	return res, nil
}

func (fbr *BitReaderRecorder) BitsRead() uint64 {
	res := fbr.realBitReader.BitsRead()
	fbr.ReadBitsData = append(fbr.ReadBitsData, res)
	return res
}

func (fbr *BitReaderRecorder) Seek(offset int64, whence int) (int64, error) {
	res, err := fbr.realBitReader.Seek(offset, whence)
	if err != nil {
		return 0, err
	}
	fbr.SeekData = append(fbr.SeekData, res)
	return res, nil
}

func (fbr *BitReaderRecorder) SkipBits(bits uint32) error {
	return fbr.realBitReader.SkipBits(bits)
}

func (fbr *BitReaderRecorder) AtEnd() bool {
	res := fbr.realBitReader.AtEnd()

	fbr.AtEndData = append(fbr.AtEndData, res)
	return res
}

func (fbr *BitReaderRecorder) ShowBits(bits int) (uint64, error) {
	res, err := fbr.realBitReader.ShowBits(bits)
	if err != nil {
		return 0, err
	}
	fbr.ShowBitsData = append(fbr.ShowBitsData, res)
	return res, nil
}

func (fbr *BitReaderRecorder) Skip(bytes uint32) (int64, error) {
	res, err := fbr.realBitReader.Skip(bytes)
	if err != nil {
		return 0, err
	}
	fbr.SkipData = append(fbr.SkipData, res)
	return res, nil
}

func (fbr *BitReaderRecorder) ReadBytesUint64(noBytes int) (uint64, error) {
	res, err := fbr.realBitReader.ReadBytesUint64(noBytes)
	if err != nil {
		return 0, err
	}
	fbr.ReadBytesUint64Data = append(fbr.ReadBytesUint64Data, res)
	return res, nil
}

func (fbr *BitReaderRecorder) Reset() error {
	return nil
}
