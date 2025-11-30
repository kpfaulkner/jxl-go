package core

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"hash/crc32"
	"io"

	"github.com/kpfaulkner/jxl-go/colour"
)

type PNGWriter struct {
	bitDepth     int32
	colourMode   byte
	width        uint32
	height       uint32
	alphaIndex   int32
	hdr          bool
	writeSRGBICC bool
}

// WritePNG instead of using standard golang image/png package since we need
// to write out ICC Profile which doesn't seem to be supported by the standard package.
func (w *PNGWriter) WritePNG(jxlImage *JXLImage, output io.Writer) error {

	w.hdr = jxlImage.isHDR()
	var bitDepth int32
	if w.hdr {
		bitDepth = 16
	} else {
		if jxlImage.imageHeader.BitDepth.BitsPerSample > 8 {
			bitDepth = 16
		} else {
			bitDepth = 8
		}
	}

	w.bitDepth = bitDepth
	gray := jxlImage.ColorEncoding == colour.CE_GRAY

	primaries := colour.CM_PRI_SRGB
	tf := colour.TF_SRGB
	if jxlImage.isHDR() {
		primaries = colour.CM_PRI_BT2100
		tf = colour.TF_PQ
	}
	whitePoint := colour.CM_WP_D65

	if jxlImage.iccProfile == nil {
		// transforms in place
		img, err := jxlImage.transform(primaries, whitePoint, tf, PEAK_DETECT_AUTO)
		if err != nil {
			return err
		}
		jxlImage = img
	}
	maxValue := int32(^(^0 << bitDepth))
	w.width = jxlImage.Width
	w.height = jxlImage.Height
	w.alphaIndex = jxlImage.alphaIndex

	var colourMode byte
	if gray {
		if jxlImage.alphaIndex >= 0 {
			colourMode = 4
		} else {
			colourMode = 0
		}
	} else {
		if jxlImage.alphaIndex >= 0 {
			colourMode = 6
		} else {
			colourMode = 2
		}
	}
	w.colourMode = colourMode

	coerce := jxlImage.alphaIsPremultiplied
	buffer, err := jxlImage.getBuffer(false)
	if err != nil {
		return err
	}
	if !coerce {
		for c := 0; c < len(buffer); c++ {
			if buffer[c].IsInt() && jxlImage.bitDepths[c] != uint32(bitDepth) {
				coerce = true
				break
			}
		}
	}
	if coerce {
		for c := 0; c < len(buffer); c++ {
			if err := buffer[c].CastToFloatIfMax(^(^0 << jxlImage.bitDepths[c])); err != nil {
				return err
			}
		}
	}

	if jxlImage.alphaIsPremultiplied {
		panic("not implemented")
	}
	for c := 0; c < len(buffer); c++ {
		if buffer[c].IsInt() && jxlImage.bitDepths[c] == uint32(bitDepth) {
			if err := buffer[c].Clamp(maxValue); err != nil {
				return err
			}
		} else {
			if err := buffer[c].CastToIntIfMax(maxValue); err != nil {
				return err
			}
		}
	}

	//newImg := jxlImage.create24BitImage(buffer)
	//buf := new(bytes.Buffer)
	//if err := png.Encode(buf, newImg); err != nil {
	//	panic(err)
	//}
	//
	//pngFileName := `c:\temp\test-image.png`
	//err = os.WriteFile(pngFileName, buf.Bytes(), 0666)
	//if err != nil {
	//	panic(err)
	//}
	//return nil

	// PNG header
	header := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
	_, err = output.Write(header)
	if err != nil {
		return err
	}

	if err := w.writeIHDR(jxlImage, output); err != nil {
		return err
	}

	if w.hdr || len(jxlImage.iccProfile) != 0 || w.writeSRGBICC {
		if err := w.writeICCP(jxlImage, output); err != nil {
			return err
		}
	} else {
		// if we have an ICC profile then write it out
		if err := w.writeSRGB(jxlImage, output); err != nil {
			return err
		}
	}

	if err := w.writeIDAT(jxlImage, output); err != nil {
		return err
	}
	_, err = output.Write([]byte{0, 0, 0, 0})
	if err != nil {
		return err
	}
	_, err = output.Write([]byte{0x49, 0x45, 0x4E, 0x44})
	if err != nil {
		return err
	}
	_, err = output.Write([]byte{0xAE, 0x42, 0x60, 0x82})
	if err != nil {
		return err
	}
	return nil
}

func (w *PNGWriter) writeICCP(image *JXLImage, output io.Writer) error {

	var buf bytes.Buffer
	buf.Write([]byte{0x69, 0x43, 0x43, 0x50})
	buf.Write([]byte("jxlatte"))
	buf.WriteByte(0x00)
	buf.WriteByte(0x00)
	var compressedICC bytes.Buffer
	wr, err := zlib.NewWriterLevel(&compressedICC, 1)
	if err != nil {
		return err
	}
	if _, err = wr.Write(image.iccProfile); err != nil {
		return err
	}

	if err = wr.Flush(); err != nil {
		return err
	}
	if err = wr.Close(); err != nil {
		return err
	}

	b := compressedICC.Bytes()
	buf.Write(b)

	rawBytes := buf.Bytes()
	buf2 := make([]byte, 4)
	binary.BigEndian.PutUint32(buf2, uint32(len(rawBytes))-4)
	if _, err = output.Write(buf2); err != nil {
		return err
	}
	if _, err = output.Write(rawBytes); err != nil {
		return err
	}

	checksum := crc32.ChecksumIEEE(rawBytes)
	binary.BigEndian.PutUint32(buf2, checksum)
	if _, err = output.Write(buf2); err != nil {
		return err
	}
	return nil
}

func (w *PNGWriter) writeSRGB(image *JXLImage, output io.Writer) error {

	if w.hdr {
		return nil
	}
	var buf bytes.Buffer
	//output.Write([]byte{0x69, 0x43, 0x43, 0x50})
	if _, err := buf.Write([]byte{0x00, 0x00, 0x00, 0x01}); err != nil {
		return err
	}

	// using jxlatte just to compare files
	if _, err := buf.Write([]byte{0x73, 0x52, 0x47, 0x42}); err != nil {
		return err
	}
	if err := buf.WriteByte(0x01); err != nil {
		return err
	}
	if _, err := buf.Write([]byte{0xD9, 0xC9, 0x2C, 0x7F}); err != nil {
		return err
	}

	rawBytes := buf.Bytes()
	//buf2 := make([]byte, 4)
	//binary.BigEndian.PutUint32(buf2, uint32(len(rawBytes))-4)
	//output.Write(buf2)
	if _, err := output.Write(rawBytes); err != nil {
		return err
	}

	return nil
}

func (w *PNGWriter) writeIHDR(jxlImage *JXLImage, output io.Writer) error {

	ihdr := make([]byte, 17)
	copy(ihdr[:4], []byte{'I', 'H', 'D', 'R'})
	binary.BigEndian.PutUint32(ihdr[4:], jxlImage.Width)
	binary.BigEndian.PutUint32(ihdr[8:], jxlImage.Height)
	ihdr[12] = byte(w.bitDepth)
	ihdr[13] = w.colourMode
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

func (w *PNGWriter) writeIDAT(jxlImage *JXLImage, output io.Writer) error {

	var buf bytes.Buffer
	buf.Write([]byte("IDAT"))

	var compressedBytes bytes.Buffer
	wr, err := zlib.NewWriterLevel(&compressedBytes, zlib.NoCompression)
	if err != nil {
		return err
	}

	var bitDepth int32
	if jxlImage.imageHeader.BitDepth.BitsPerSample > 8 {
		bitDepth = 16
	} else {
		bitDepth = 8
	}

	maxValue := int32(^(^0 << bitDepth))

	if err = jxlImage.Buffer[0].CastToIntIfMax(maxValue); err != nil {
		return err
	}

	if len(jxlImage.Buffer) > 1 {
		if err = jxlImage.Buffer[1].CastToIntIfMax(maxValue); err != nil {
			return err
		}
	}

	if len(jxlImage.Buffer) > 2 {
		if err = jxlImage.Buffer[2].CastToIntIfMax(maxValue); err != nil {
			return err
		}
	}

	if jxlImage.HasAlpha() {
		if err = jxlImage.Buffer[3].CastToIntIfMax(maxValue); err != nil {
			return err
		}
	}

	for c := 0; c < len(jxlImage.Buffer); c++ {
		if jxlImage.Buffer[c].IsInt() && jxlImage.bitDepths[c] == uint32(bitDepth) {
			if err := jxlImage.Buffer[c].Clamp(maxValue); err != nil {
				return err
			}
		} else {
			if err := jxlImage.Buffer[c].CastToIntIfMax(maxValue); err != nil {
				return err
			}
		}
	}
	for y := uint32(0); y < jxlImage.Height; y++ {
		if _, err := wr.Write([]byte{0}); err != nil {
			return err
		}
		for x := uint32(0); x < jxlImage.Width; x++ {

			// FIXME(kpfaulkner) remove 3 assumption
			for c := 0; c < jxlImage.imageHeader.GetColourChannelCount(); c++ {
				dat := jxlImage.Buffer[c].IntBuffer[y][x]
				if jxlImage.bitDepths[c] == 8 {
					if _, err := wr.Write([]byte{byte(dat)}); err != nil {
						return err
					}
				} else {
					byte1 := dat & 0xFF
					byte2 := dat & 0xFF00
					byte2 >>= 8
					if _, err := wr.Write([]byte{byte(byte2), byte(byte1)}); err != nil {
						return err
					}
				}

			}
			if jxlImage.HasAlpha() {
				dat := jxlImage.Buffer[3].IntBuffer[y][x]
				if jxlImage.bitDepths[3] == 8 {
					if _, err := wr.Write([]byte{byte(dat)}); err != nil {
						return err
					}
				} else {
					byte1 := dat & 0xFF
					byte2 := dat & 0xFF00
					byte2 >>= 8

					if _, err := wr.Write([]byte{byte(byte2), byte(byte1)}); err != nil {
						return err
					}
				}
			}
		}
	}
	wr.Close()

	if _, err := buf.Write(compressedBytes.Bytes()); err != nil {
		return err
	}
	bb := buf.Bytes()
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, uint32(len(bb))-4)
	if _, err := output.Write(b); err != nil {
		return err
	}
	if _, err := output.Write(bb); err != nil {
		return err
	}
	checksum := crc32.ChecksumIEEE(bb)
	binary.BigEndian.PutUint32(b, checksum)
	if _, err := output.Write(b); err != nil {
		return err
	}

	return nil
}
