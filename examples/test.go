package main

import (
	"bytes"
	"fmt"
	"image"
	"image/png"
	"os"
	"path"
	"time"
	//"github.com/pkg/profile"
	"github.com/kpfaulkner/jxl-go/core"
	"github.com/pkg/profile"
	log "github.com/sirupsen/logrus"
)

func main() {

	// church image exercises different code pathways compared to lenna.jxl
	file := `c:/temp/ken-0-4.jxl`
	//file := `c:/temp/dont_ask_why_i_have_this.jxl`
	//file := `../testdata/patches.jxl`
	//file := `../testdata/alpha-triangles.jxl`
	//file := `../testdata/unittest.jxl`

	defer profile.Start(profile.CPUProfile, profile.ProfilePath(".")).Stop()
	//defer profile.Start(profile.MemProfile, profile.ProfilePath(".")).Stop()
	//defer profile.Start(profile.MemProfileAllocs, profile.ProfilePath(".")).Stop()

	f, err := os.ReadFile(file)
	if err != nil {
		log.Errorf("Error opening file: %v\n", err)
		return
	}

	start := time.Now()
	var img image.Image
	for count := 0; count < 100; count++ {
		r := bytes.NewReader(f)
		jxl := core.NewJXLDecoder(r, nil)
		start := time.Now()
		var jxlImage *core.JXLImage
		if jxlImage, err = jxl.Decode(); err != nil {
			fmt.Printf("Error decoding: %v\n", err)
			return
		}
		fmt.Printf("decoding took %d ms\n", time.Since(start).Milliseconds())
		fmt.Printf("Has alpha %v\n", jxlImage.HasAlpha())
		fmt.Printf("Num extra channels (inc alpha) %d\n", jxlImage.NumExtraChannels())

		if ct, err := jxlImage.GetExtraChannelType(0); err == nil {
			fmt.Printf("channel 3 type %d\n", ct)
		}

		// convert to regular Go image.Image
		img, err = jxlImage.ToImage()
		if err != nil {
			fmt.Printf("error when making image %v\n", err)
		}

	}

	end := time.Now()
	fmt.Printf("decoding total time %d ms\n", end.Sub(start).Milliseconds())
	buf := new(bytes.Buffer)
	if err := png.Encode(buf, img); err != nil {
		log.Fatalf("boomage %v", err)
	}
	ext := path.Ext(file)
	pngFileName := file[:len(file)-len(ext)] + ".png"
	err = os.WriteFile(pngFileName, buf.Bytes(), 0666)
	if err != nil {
		log.Fatalf("boomage %v", err)
	}
}
