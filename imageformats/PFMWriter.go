package imageformats

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/kpfaulkner/jxl-go/color"
	"github.com/kpfaulkner/jxl-go/core"
)

func WritePFM(jxlImage *core.JXLImage, output io.Writer) error {

	gray := jxlImage.ColorEncoding == color.CE_GRAY
	width := jxlImage.Width
	height := jxlImage.Height

	pf := "Pf"
	if !gray {
		pf = "PF"
	}
	header := fmt.Sprintf("%s\n%d %d\n1.0\n", pf, width, height)
	output.Write([]byte(header))
	buffer := jxlImage.Buffer
	cCount := 1
	if !gray {
		cCount = 3
	}
	var buf bytes.Buffer
	for y := height - 1; y >= 0; y-- {
		for x := 0; x < width; x++ {
			p := y*width + x
			for c := 0; c < cCount; c++ {
				buf.Reset()
				err := binary.Write(&buf, binary.BigEndian, buffer[c][p])
				if err != nil {
					fmt.Println("binary.Write failed:", err)
					return err
				}
				_, err = output.Write(buf.Bytes())
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}
