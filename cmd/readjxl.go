package main

import (
	"bytes"
	"fmt"
	"image/png"
	"os"
	"time"

	"github.com/kpfaulkner/jxl-go/core"
	"github.com/pkg/profile"
	log "github.com/sirupsen/logrus"
)

func main() {
	fmt.Printf("So it begins...\n")

	defer profile.Start(profile.TraceProfile, profile.ProfilePath(`.`)).Stop()
	//defer profile.Start(profile.CPUProfile, profile.ProfilePath(`.`)).Stop()
	//defer profile.Start(profile.BlockProfile, profile.ProfilePath(`.`)).Stop()
	//defer profile.Start(profile.MemProfileHeap, profile.MemProfileRate(1), profile.ProfilePath(`.`)).Stop()
	//defer profile.Start(profile.MemProfileAllocs, profile.MemProfileRate(1), profile.ProfilePath(`.`)).Stop()

	file := `../testdata/lossless.jxl`
	//file := `../testdata/lenna.jxl`
	//file := `c:\temp\work.jxl`
	//file := `c:\temp\from-nwf.jxl`
	//file := `c:\temp\ken-0-0.jxl`
	//file := `c:\temp\tiny2.jxl`
	//file := `c:\temp\tiny4.jxl`
	//file := `c:\temp\tiny5.jxl`
	//file := `c:\temp\input.jxl`
	f, err := os.ReadFile(file)
	if err != nil {
		log.Errorf("Error opening file: %v\n", err)
		return
	}

	r := bytes.NewReader(f)
	start := time.Now()
	jxl := core.NewJXLDecoder(r)

	var jxlImage *core.JXLImage

	if jxlImage, err = jxl.Decode(); err != nil {
		fmt.Printf("Error decoding: %v\n", err)
		return
	}
	fmt.Printf("decoding took %d ms\n", time.Since(start).Milliseconds())

	return

	// convert to regular Go image.Image
	img, err := jxlImage.ToImage()
	if err != nil {
		fmt.Printf("error when making image %v\n", err)
	}

	buf := new(bytes.Buffer)
	if err := png.Encode(buf, img); err != nil {
		log.Fatalf("boomage %v", err)
	}
	err = os.WriteFile(`c:\temp\test.png`, buf.Bytes(), 0666)
	if err != nil {
		log.Fatalf("boomage %v", err)
	}
}
