package main

import (
	"bytes"
	"fmt"
	"image/png"
	//"image/png"
	"os"
	"path"
	"strings"
	//"path"
	"time"

	//"github.com/pkg/profile"
	"github.com/kpfaulkner/jxl-go/core"
	log "github.com/sirupsen/logrus"
)

func main() {

	filePaths := []string{

		// not looking like ref.png but same as jxlatte output
		//`C:\Users\kenfa\projects\conformance\testcases\spot\input.jxl|spot.png`,
		//
		//`C:\Users\kenfa\projects\conformance\testcases\upsampling\input.jxl|upsampling.png`,
		//`C:\Users\kenfa\projects\conformance\testcases\sunset_logo\input.jxl|sunset_logo.png`,
		//`C:\Users\kenfa\projects\conformance\testcases\cafe\input.jxl|cafe.png`,
		`C:\Users\kenfa\projects\conformance\testcases\delta_palette\input.jxl|delta_palette.png`,
		//`c:\temp\ken-0-4.jxl|ken-0-4.png`,
		//`..\testdata\unittest.jxl|unittest.png`,
		//`..\testdata\bench.jxl|bench.png`,
		//`..\testdata\alpha-triangles.jxl|alpha-triangles.png`,
		//`..\testdata\bbb.jxl|bbb.png`,
		//`..\testdata\ants-lossless.jxl|ants-lossless.png`,
		//`..\testdata\lenna.jxl|lenna.png`,
		//`..\testdata\quilt.jxl|quilt.png`,
		//`..\testdata\wb-rainbow.jxl|wb-rainbow.png`,
		//`..\testdata\ants.jxl|ants.png`,
		//`..\testdata\blendmodes_5.jxl|blendmodes_5.png`,
		//`..\testdata\lossless.jxl|lossless.png`,
		//`..\testdata\sollevante-hdr.jxl|sollevante-hdr.png`,
		//`..\testdata\white.jxl|white.png`,
		//`..\testdata\art.jxl|art.png`,
		//`..\testdata\church.jxl|church.png`,
		//`..\testdata\patches-lossless.jxl|patches-lossless.png`,
		//`..\testdata\tiny2.jxl|tiny2.png`,
	}

	destinationDir := `c:\temp\jxlresults\`
	for _, file := range filePaths {
		fileDetails := strings.Split(file, "|")
		orig := fileDetails[0]
		newFile := fileDetails[1]
		fmt.Printf("file %s\n", orig)
		f, err := os.ReadFile(orig)
		if err != nil {
			log.Errorf("Error opening file: %v\n", err)
			return
		}

		start := time.Now()
		//var img image.Image
		for count := 0; count < 1; count++ {
			r := bytes.NewReader(f)
			jxl := core.NewJXLDecoder(r, nil)
			start := time.Now()
			//p := profile.Start(profile.CPUProfile, profile.ProfilePath("."))
			var jxlImage *core.JXLImage
			if jxlImage, err = jxl.Decode(); err != nil {
				fmt.Printf("Error decoding: %v\n", err)
				return
			}
			//p.Stop()
			fmt.Printf("decoding took %d ms\n", time.Since(start).Milliseconds())
			fmt.Printf("Has alpha %v\n", jxlImage.HasAlpha())
			fmt.Printf("Num extra channels (inc alpha) %d\n", jxlImage.NumExtraChannels())

			if ct, err := jxlImage.GetExtraChannelType(0); err == nil {
				fmt.Printf("channel 3 type %d\n", ct)
			}

			pngFileName := path.Join(destinationDir, newFile)

			// if ICC profile then use custom PNG writer... otherwise use default Go encoder.
			if jxlImage.HasICCProfile() || true {
				f, err := os.Create(pngFileName)
				if err != nil {
					log.Fatalf("boomage %v", err)
				}
				defer f.Close()
				core.WritePNG(jxlImage, f)
			} else {

				// convert to regular Go image.Image
				img, err := jxlImage.ToImage()
				if err != nil {
					fmt.Printf("error when making image %v\n", err)
				}

				buf := new(bytes.Buffer)
				if err := png.Encode(buf, img); err != nil {
					log.Fatalf("boomage %v", err)
				}

				err = os.WriteFile(pngFileName, buf.Bytes(), 0666)
				if err != nil {
					log.Fatalf("boomage %v", err)
				}
			}

		}

		end := time.Now()
		fmt.Printf("decoding total time %d ms\n", end.Sub(start).Milliseconds())
	}
}
