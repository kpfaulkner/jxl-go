package main

import (
	"fmt"
	"os"
	"time"

	"github.com/kpfaulkner/jxl-go/core"
	"github.com/kpfaulkner/jxl-go/imageformats"
)

func main() {
	fmt.Printf("So it begins...\n")

	jxl := core.NewJXLDecoder(core.WithInputFilename(`../testdata/lossless.jxl`))

	var img *core.JXLImage
	var err error
	start := time.Now()
	if img, err = jxl.Decode(); err != nil {
		fmt.Printf("Error decoding: %v\n", err)
		return
	}
	fmt.Printf("decoding took %d ms\n", time.Since(start).Milliseconds())

	// now convert to PNG for moment.
	pfmFile, err := os.Create(`c:\temp\lossless-jxl-go.pfm`)
	if err != nil {
		fmt.Printf("Error opening output file: %v\n", err)
		return
	}
	defer pfmFile.Close()
	imageformats.WritePFM(img, pfmFile)
}
