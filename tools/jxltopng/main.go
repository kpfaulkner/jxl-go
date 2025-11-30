package main

import (
	"bytes"
	"flag"
	"fmt"
	"image/png"
	"os"
	"time"

	"github.com/kpfaulkner/jxl-go/core"
	log "github.com/sirupsen/logrus"
)

func main() {
	infile := flag.String("i", "", "input jxl file")
	outfile := flag.String("o", "", "output png file")
	flag.Parse()

	if *infile == "" || *outfile == "" {
		fmt.Printf("both input and output files must be specified\n")
		os.Exit(1)
	}

	f, err := os.ReadFile(*infile)
	if err != nil {
		log.Errorf("Error opening file: %v\n", err)
		return
	}

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

	startEncoding := time.Now()
	// if ICC profile then use custom PNG writer... otherwise use default Go encoder.
	if jxlImage.HasICCProfile() {
		f, err := os.Create(*outfile)
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
		startEncodingToImg := time.Now()
		// convert to regular Go image.Image
		img, err := jxlImage.ToImage()
		if err != nil {
			fmt.Printf("error when making image %v\n", err)
		}

		fmt.Printf("encoding to img took %d ms\n", time.Since(startEncodingToImg).Milliseconds())
		buf := new(bytes.Buffer)
		if err := png.Encode(buf, img); err != nil {
			log.Fatalf("boomage %v", err)
		}

		err = os.WriteFile(*outfile, buf.Bytes(), 0666)
		if err != nil {
			log.Fatalf("boomage %v", err)
		}
	}

	fmt.Printf("encoding took %d ms\n", time.Since(startEncoding).Milliseconds())
}
