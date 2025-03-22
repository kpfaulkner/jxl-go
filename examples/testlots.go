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
	log "github.com/sirupsen/logrus"
)

func main() {

	filePaths := []string{
		`..\testdata\alpha-triangles.jxl`,
		`..\testdata\bbb.jxl`,
		//`..\testdata\george-tiled.jxl`,
		//`..\testdata\patches.jxl`,
		`..\testdata\unittest.jxl`,
		`..\testdata\ants-lossless.jxl`,
		`..\testdata\bench.jxl`,
		`..\testdata\lenna.jxl`,
		`..\testdata\quilt.jxl`,
		`..\testdata\wb-rainbow.jxl`,
		`..\testdata\ants.jxl`,
		`..\testdata\blendmodes_5.jxl`,
		`..\testdata\lossless.jxl`,
		`..\testdata\sollevante-hdr.jxl`,
		`..\testdata\white.jxl`,
		`..\testdata\art.jxl`,
		`..\testdata\church.jxl`,
		`..\testdata\patches-lossless.jxl`,
		`..\testdata\tiny2.jxl`,
	}

	for _, file := range filePaths {
		fmt.Printf("file %s\n", file)
		f, err := os.ReadFile(file)
		if err != nil {
			log.Errorf("Error opening file: %v\n", err)
			return
		}

		start := time.Now()
		var img image.Image
		for count := 0; count < 1; count++ {
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
}
