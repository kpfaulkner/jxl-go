package core

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"

	"github.com/kpfaulkner/jxl-go/jxlio"
)

var (
	JPEGXLHEADERALT = [12]byte{0x00, 0x00, 0x00, 0x0C, 0x4A, 0x58, 0x4C, 0x20, 0x0D, 0x0A, 0x87, 0x0A}

	JXLL = makeTag([]byte{'j', 'x', 'l', 'l'}, 0, 4)
	JXLP = makeTag([]byte{'j', 'x', 'l', 'p'}, 0, 4)
	JXLC = makeTag([]byte{'j', 'x', 'l', 'c'}, 0, 4)
)

type ContainerBoxHeader struct {
	BoxType   uint64
	BoxSize   uint64
	IsLast    bool
	Offset    int64 // offset compared to very beginning of file.
	Processed bool  // indicated if finished with.
}

type BoxReader struct {
	reader *jxlio.Bitreader
	level  int
}

func NewBoxReader(reader *jxlio.Bitreader) *BoxReader {
	return &BoxReader{
		reader: reader,
		level:  5,
	}
}

func (br *BoxReader) ReadBoxHeader() ([]ContainerBoxHeader, error) {
	buffer := make([]byte, 12)
	err := br.reader.ReadByteArrayWithOffsetAndLength(buffer, 0, 12)
	if err != nil {
		return nil, err
	}

	if !bytes.Equal(buffer, JPEGXLHEADERALT[:]) {
		return nil, fmt.Errorf("invalid magic number: %+v", buffer)
	}

	var containerBoxHeaders []ContainerBoxHeader
	if containerBoxHeaders, err = br.readAllBoxes(); err != nil {
		return nil, err
	}

	return containerBoxHeaders, nil
}

// readAllBoxes read all boxes and gets offset/size details so we can read them later.
func (br *BoxReader) readAllBoxesOrig() ([]ContainerBoxHeader, error) {

	//var boxHeaders []ContainerBoxHeader
	//var bSize uint32
	//var bType uint64
	//var offset uint64
	//for {
	//	boxSize, tag, skip, err := br.readSizeAndType()
	//	if err != nil {
	//		// end of file... return
	//		if err == io.EOF {
	//			return boxHeaders, nil
	//		}
	//		return nil, err
	//	}
	//
	//	// seek ahead to the next box.
	//	newPos, err := br.reader.Seek(int64(bSize), io.SeekCurrent)
	//	if err != nil {
	//		return nil, err
	//	}
	//
	//	fmt.Printf("new pos %d\n", newPos)
	//
	//	boxHeaders = append(boxHeaders,
	//		ContainerBoxHeader{
	//			Size:    bSize,
	//			BoxType: bType,
	//			IsLast:  false, // not sure we need this yet.
	//			Offset:  offset,
	//		})
	//}
	return nil, nil
}

func (br *BoxReader) readAllBoxes() ([]ContainerBoxHeader, error) {

	var boxHeaders []ContainerBoxHeader
	//boxSizeArray := make([]byte, 4)
	boxSizeArray := make([]byte, 8)
	boxTag := make([]byte, 4)
	for {

		err := br.reader.ReadBytesToBuffer(boxSizeArray, 4)
		if err != nil {
			if err == io.EOF {
				// simple end of file... return with boxHeaders
				return boxHeaders, nil
			}
			return nil, err
		}

		boxSize := makeTag(boxSizeArray, 0, 4)
		if boxSize == 1 {
			err = br.reader.ReadBytesToBuffer(boxSizeArray, 8)
			if err != nil {
				return nil, err
			}
			boxSize = makeTag(boxSizeArray, 0, 8)
			if boxSize > 0 {
				boxSize -= 8
			}
		}
		if boxSize > 0 {
			boxSize -= 8
		}
		if boxSize < 0 {
			return nil, errors.New("invalid box size")
		}

		err = br.reader.ReadBytesToBuffer(boxTag, 4)
		if err != nil {
			return nil, err
		}
		tag := makeTag(boxTag, 0, 4)

		// check boxType...  if we dont know the box type, just skip over the bytes and keep reading.
		switch tag {
		case JXLP:
			// reads next 4 bytes as additional tag?
			err = br.reader.ReadBytesToBuffer(boxTag, 4)
			if err != nil {
				return nil, err
			}
			boxSize -= 4

			// fileoffset...  directly from ReadSeeker?
			pos, err := br.reader.Seek(0, io.SeekCurrent)
			if err != nil {
				return nil, err
			}
			bh := ContainerBoxHeader{
				BoxType:   tag,
				BoxSize:   boxSize,
				IsLast:    false,
				Offset:    pos,
				Processed: false,
			}
			boxHeaders = append(boxHeaders, bh)

		case JXLL:
			if boxSize != 1 {
				return nil, errors.New("JXLL box size should be 1")
			}
			l, err := br.reader.ReadByte()
			if err != nil {
				return nil, err
			}
			if l != 5 && l != 10 {
				return nil, errors.New("invalid level")
			}
			br.level = int(l)

		case JXLC:
			// fileoffset...  directly from ReadSeeker?
			pos, err := br.reader.Seek(0, io.SeekCurrent)
			if err != nil {
				return nil, err
			}
			bh := ContainerBoxHeader{
				BoxType:   tag,
				BoxSize:   boxSize,
				IsLast:    false,
				Offset:    pos,
				Processed: false,
			}
			boxHeaders = append(boxHeaders, bh)
			// skip past this box.
			s, err := br.SkipFully(int64(boxSize))
			if err != nil {
				return nil, err
			}
			fmt.Printf("S is %d\n", s)

		default:
			// skip over the bytes
			if boxSize > 0 {
				s, err := br.SkipFully(int64(boxSize))
				if err != nil {
					return nil, err
				}
				if s != 0 {
					return nil, errors.New("truncated extra box")
				}
			} else {
				panic("java read supplyExceptionally... unsure why?")
			}
		}
	}

	return boxHeaders, nil
}

func (br *BoxReader) readBox() error {

	return nil
}

func (br *BoxReader) readSizeAndType() (offset uint64, boxSize uint64, tag uint64, skip bool, err error) {
	//
	//boxSizeArray, err := br.reader.ReadByteArray(4)
	//if err != nil {
	//	return 0, 0, 0, false, err
	//}
	//
	//boxSize = makeTag(boxSizeArray, 0, 4)
	//if boxSize == 1 {
	//	boxSizeArray, err = br.reader.ReadByteArray(8)
	//	if err != nil {
	//		return 0, 0, 0, false, err
	//	}
	//	boxSize = makeTag(boxSizeArray, 0, 8)
	//	if boxSize > 0 {
	//		boxSize -= 8
	//	}
	//}
	//if boxSize > 0 {
	//	boxSize -= 8
	//}
	//if boxSize < 0 {
	//	return 0, 0, 0, false, fmt.Errorf("invalid box size: %d", boxSize)
	//}
	//
	//boxTag, err := br.reader.ReadByteArray(4)
	//if err != nil {
	//	return 0, 0, 0, false, err
	//}
	//tag = makeTag(boxTag, 0, 4)
	//if tag == JXLL {
	//	if boxSize != 1 {
	//		return 0, 0, 0, false, fmt.Errorf("JXLL box size should be 1")
	//	}
	//	l, err := br.reader.ReadByte()
	//	if err != nil {
	//		return 0, 0, 0, false, err
	//	}
	//	if l != 5 && l != 10 {
	//		return 0, 0, 0, false, fmt.Errorf("invalid level %d", l)
	//	}
	//	br.level = int(l)
	//	return boxSize, tag, true, nil
	//}
	//if tag == JXLP {
	//	boxTag, err = br.reader.ReadByteArray(4)
	//	if err != nil {
	//		return 0, 0, 0, false, errors.New("truncated sequence number")
	//	}
	//	boxSize -= 4
	//}
	//
	//// if JXLP or JXLC then read contents immediately
	//if tag == JXLP || tag == JXLC {
	//	return boxSize, tag, false, nil
	//} else {
	//
	//	// otherwise skip the box?
	//	if boxSize > 0 {
	//		s, err := br.SkipFully(int64(boxSize))
	//		if err != nil {
	//			return 0, 0, 0, false, err
	//		}
	//		if s != 0 {
	//			return 0, 0, 0, false, errors.New("truncated extra box")
	//		}
	//	} else {
	//		panic("java read supplyExceptionally... unsure why?")
	//	}
	//}
	//
	//return boxSize, tag, false, nil
	return 0, 0, 0, false, nil
}

// returns number of bytes that were NOT skipped.
func (br *BoxReader) SkipFully(i int64) (int64, error) {
	n, err := br.reader.Skip(uint32(i))
	return i - n, err
}

func makeTag(bytes []uint8, offset int, length int) uint64 {
	tag := uint64(0)
	for i := offset; i < offset+length; i++ {
		tag = (tag << 8) | uint64(bytes[i])&0xFF
	}
	return tag
}

func fromBeToLe(le uint32) uint32 {
	return uint32(binary.BigEndian.Uint32([]byte{byte(le), byte(le >> 8), byte(le >> 16), byte(le >> 24)}))
}
