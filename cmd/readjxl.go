package main

import (
	"bytes"
	"fmt"
	"image"
	"image/png"
	"os"
	"time"

	"github.com/kpfaulkner/jxl-go/core"
	log "github.com/sirupsen/logrus"
)

func main() {
	fmt.Printf("So it begins...\n")

	//defer profile.Start(profile.TraceProfile, profile.ProfilePath(`.`)).Stop()
	//defer profile.Start(profile.CPUProfile, profile.ProfilePath(`.`)).Stop()
	//defer profile.Start(profile.BlockProfile, profile.ProfilePath(`.`)).Stop()
	//defer profile.Start(profile.MemProfileHeap, profile.MemProfileRate(1), profile.ProfilePath(`.`)).Stop()
	//defer profile.Start(profile.MemProfileAllocs, profile.MemProfileRate(1), profile.ProfilePath(`.`)).Stop()

	file := `../testdata/lossless.jxl`
	//file := `../testdata/lenna.jxl`
	//file := `c:\temp\work.jxl`
	//file := `c:\temp\tiny2.jxl`
	//file := `c:\temp\input.jxl`
	f, err := os.ReadFile(file)
	if err != nil {
		log.Errorf("Error opening file: %v\n", err)
		return
	}

	r := bytes.NewReader(f)
	jxl := core.NewJXLDecoder(r)

	var img image.Image
	start := time.Now()
	if img, err = jxl.Decode(); err != nil {
		fmt.Printf("Error decoding: %v\n", err)
		return
	}
	fmt.Printf("decoding took %d ms\n", time.Since(start).Milliseconds())
	fmt.Printf("img %+v\n", img.Bounds())

	//return

	buf := new(bytes.Buffer)
	if err := png.Encode(buf, img); err != nil {
		log.Fatalf("boomage %v", err)
	}
	err = os.WriteFile(`c:\temp\test.png`, buf.Bytes(), 0666)
	if err != nil {
		log.Fatalf("boomage %v", err)
	}
}
