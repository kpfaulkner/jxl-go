package main

import (
	"bytes"
	"fmt"
	"image"
	"image/png"
	"os"
	"time"

	"github.com/kpfaulkner/jxl-go/core"
	"github.com/sirupsen/logrus"
)

func main() {
	fmt.Printf("So it begins...\n")

	//defer profile.Start(profile.TraceProfile, profile.ProfilePath(`.`)).Stop()
	//defer profile.Start(profile.CPUProfile, profile.ProfilePath(`.`)).Stop()
	//defer profile.Start(profile.BlockProfile, profile.ProfilePath(`.`)).Stop()
	//defer profile.Start(profile.MemProfileHeap, profile.MemProfileRate(1), profile.ProfilePath(`.`)).Stop()

	jxl := core.NewJXLDecoder(core.WithInputFilename(`../testdata/lossless.jxl`), core.ReadFileIntoMemory())

	var img image.Image
	var err error
	start := time.Now()
	if img, err = jxl.Decode(); err != nil {
		fmt.Printf("Error decoding: %v\n", err)
		return
	}
	fmt.Printf("decoding took %d ms\n", time.Since(start).Milliseconds())
	//return
	fmt.Printf("img %+v\n", img.Bounds())

	buf := new(bytes.Buffer)
	if err := png.Encode(buf, img); err != nil {
		logrus.Fatalf("boomage %v", err)
	}
	err = os.WriteFile(`c:\temp\test.png`, buf.Bytes(), 0666)
	if err != nil {
		logrus.Fatalf("boomage %v", err)
	}

	// now convert to PNG for moment.
	//pfmFile, err := os.Create(`c:\temp\lossless-jxl-go.pfm`)
	//if err != nil {
	//	fmt.Printf("Error opening output file: %v\n", err)
	//	return
	//}
	//defer pfmFile.Close()
	//imageformats.WritePFM(img, pfmFile)
}
