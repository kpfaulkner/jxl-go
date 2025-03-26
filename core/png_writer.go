package core

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"hash/crc32"
	"io"
)

// WritePNG instead of using standard golang image/png package since we need
// to write out ICC Profile which doesn't seem to be supported by the standard package.
func WritePNG(jxlImage *JXLImage, output io.Writer) error {

	// PNG header
	header := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
	output.Write(header)

	if err := writeIHDR(jxlImage, output); err != nil {
		return err
	}

	if err := writeICCP(jxlImage, output); err != nil {
		return err
	}

	if err := writeIDAT(jxlImage, output); err != nil {
		return err
	}
	return nil
}

func writeICCP(image *JXLImage, output io.Writer) error {

	var buf bytes.Buffer
	//output.Write([]byte{0x69, 0x43, 0x43, 0x50})
	buf.Write([]byte{0x69, 0x43, 0x43, 0x50})
	buf.Write([]byte("jxlatte")) // using jxlatte just to compare files
	buf.WriteByte(0x00)
	buf.WriteByte(0x00)

	// decompress icc?
	//w := zlib.NewWriter(&buf)

	//iccReader := bytes.NewReader(image.iccProfile)

	var compressedICC bytes.Buffer
	w, err := zlib.NewWriterLevel(&compressedICC, 0)
	if err != nil {
		return err
	}
	w.Write(image.iccProfile)
	w.Flush()
	w.Close()

	//buf.Write(image.iccProfile)
	b := compressedICC.Bytes()
	buf.Write(b)
	//w.Close()

	rawBytes := buf.Bytes()
	buf2 := make([]byte, 4)
	binary.BigEndian.PutUint32(buf2, uint32(len(rawBytes))-4)
	output.Write(buf2)
	// ICCP

	output.Write(rawBytes)
	checksum := crc32.ChecksumIEEE(rawBytes)
	binary.BigEndian.PutUint32(buf2, checksum)
	output.Write(buf2)
	return nil
}

func writeIHDR(jxlImage *JXLImage, output io.Writer) error {

	// colourmode is 6 if include alpha...  otherwise 2
	colourMode := byte(2)
	if jxlImage.HasAlpha() {
		colourMode = 6
	}

	ihdr := make([]byte, 17)
	copy(ihdr[:4], []byte{'I', 'H', 'D', 'R'})
	binary.BigEndian.PutUint32(ihdr[4:], jxlImage.Width)
	binary.BigEndian.PutUint32(ihdr[8:], jxlImage.Height)
	ihdr[12] = byte(jxlImage.imageHeader.BitDepth.BitsPerSample)
	ihdr[13] = colourMode
	ihdr[14] = 0
	ihdr[15] = 0
	ihdr[16] = 0

	sizeBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(sizeBytes, uint32(len(ihdr)-4))
	if _, err := output.Write(sizeBytes); err != nil {
		return err
	}

	if _, err := output.Write(ihdr); err != nil {
		return err
	}

	checksum := crc32.ChecksumIEEE(ihdr)
	checkSumBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(checkSumBytes, checksum)
	if _, err := output.Write(checkSumBytes); err != nil {
		return err
	}

	return nil
}

func writeIDAT(jxlImage *JXLImage, output io.Writer) error {

	//if _, err := buf.Write([]byte("IDAT")); err != nil {
	//	return err
	//}

	var buf bytes.Buffer
	buf.Write([]byte("IDAT"))

	var compressedBytes bytes.Buffer
	w, err := zlib.NewWriterLevel(&compressedBytes, 0)
	if err != nil {
		return err
	}

	for y := uint32(0); y < jxlImage.Height; y++ {
		w.Write([]byte{0})
		for x := uint32(0); x < jxlImage.Width; x++ {

			// FIXME(kpfaulkner) remove 3 assumption
			for c := 0; c < 3; c++ {
				w.Write([]byte{byte(jxlImage.Buffer[c].FloatBuffer[y][x] * 255)})
			}
			if jxlImage.HasAlpha() {
				w.Write([]byte{byte(jxlImage.Buffer[3].FloatBuffer[y][x] * 255)})
			}
		}
	}
	w.Close()

	buf.Write(compressedBytes.Bytes())
	bb := buf.Bytes()
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, uint32(len(bb))-4)
	output.Write(b)
	output.Write(bb)
	checksum := crc32.ChecksumIEEE(bb)
	binary.BigEndian.PutUint32(b, checksum)
	output.Write(b)

	return nil
}
