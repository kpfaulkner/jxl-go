package main

import (
	"bytes"
	"fmt"
	"image/png"
	"os"
	"path"
	"strings"
	"time"

	"github.com/kpfaulkner/jxl-go/core"
	"github.com/pkg/profile"
	log "github.com/sirupsen/logrus"
)

func main() {

	filePaths := []string{

		`c:/temp/ants.jxl|ants-lossless.png`,
	}

	//p := profile.Start(profile.MemProfileHeap, profile.ProfilePath("."))
	//p := profile.Start(profile.MemProfileHeap, profile.ProfilePath("."))
	//p := profile.Start(profile.MemProfileRate(1), profile.ProfilePath("."))
	p := profile.Start(profile.CPUProfile, profile.ProfilePath("."))
	defer p.Stop()

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
			if true || jxlImage.HasICCProfile() {
				f, err := os.Create(pngFileName)
				if err != nil {
					log.Fatalf("boomage %v", err)
				}
				defer f.Close()
				pngWriter := core.PNGWriter{}
				err = pngWriter.WritePNG(jxlImage, f)
				if err != nil {
					log.Fatalf("boomage %v", err)
				}
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
