package main

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"strings"
	"time"

	"github.com/kpfaulkner/jxl-go/core"
	log "github.com/sirupsen/logrus"
)

func main() {

	filePaths := []string{
		`..\testdata\test-benchmark.jxl|test-benchmark.png`,
		`..\testdata\sollevante-hdr.jxl|sollevante-hdr.png`,
		`..\testdata\ants-lossless.jxl|ants-lossless.png`,
		`..\testdata\bbb.jxl|bbb.png`,
		`..\testdata\bench.jxl|bench.png`,
		`..\testdata\spot.jxl|spot.png`,
		`..\testdata\upsampling.jxl|upsampling.png`,
		`..\testdata\sunset_logo.jxl|sunset_logo.png`,
		`..\testdata\cafe.jxl|cafe.png`,
		`..\testdata\delta_palette.jxl|delta_palette.png`,
		`..\testdata\grayscale.jxl|grayscale.png`,
		`..\testdata\unittest.jxl|unittest.png`,
		`..\testdata\bench.jxl|bench.png`,
		`..\testdata\alpha-triangles.jxl|alpha-triangles.png`,
		`..\testdata\lenna.jxl|lenna.png`,
		`..\testdata\quilt.jxl|quilt.png`,
		`..\testdata\wb-rainbow.jxl|wb-rainbow.png`,
		`..\testdata\ants.jxl|ants.png`,
		`..\testdata\blendmodes_5.jxl|blendmodes_5.png`,
		`..\testdata\lossless.jxl|lossless.png`,
		`..\testdata\white.jxl|white.png`,
		`..\testdata\art.jxl|art.png`,
		`..\testdata\church.jxl|church.png`,
		`..\testdata\tiny2.jxl|tiny2.png`,
	}

	// hardcoded output dir for now
	destinationDir := `c:\temp\jxlresults\`

	totalTime := time.Now()
	for _, file := range filePaths {
		fileDetails := strings.Split(file, "|")
		orig := fileDetails[0]
		newFile := fileDetails[1]
		//fmt.Printf("file %s\n", orig)
		f, err := os.ReadFile(orig)
		if err != nil {
			log.Errorf("Error opening file: %v\n", err)
			return
		}

		start := time.Now()
		r := bytes.NewReader(f)
		jxl := core.NewJXLDecoder(r, nil)
		//p := profile.Start(profile.CPUProfile, profile.ProfilePath("."))
		var jxlImage *core.JXLImage
		if jxlImage, err = jxl.Decode(); err != nil {
			fmt.Printf("Error decoding: %v\n", err)
			continue
		}
		//p.Stop()
		decodingDuration := time.Since(start)

		//fmt.Printf("Has alpha %v\n", jxlImage.HasAlpha())
		//fmt.Printf("Num extra channels (inc alpha) %d\n", jxlImage.NumExtraChannels())

		pngFileName := path.Join(destinationDir, newFile)

		ff, err := os.Create(pngFileName)
		if err != nil {
			log.Fatalf("boomage %v", err)
		}
		defer ff.Close()
		pngWriter := core.PNGWriter{}
		pngWriter.WritePNG(jxlImage, ff)

		// produces some recognisable output for alpha-triangles but is NOT correct. (stripes)
		//// if ICC profile then use custom PNG writer... otherwise use default Go encoder.
		//if jxlImage.HasICCProfile() {
		//	//f, err := os.Create(pngFileName)
		//	//if err != nil {
		//	//	log.Fatalf("boomage %v", err)
		//	//}
		//	//defer f.Close()
		//	//core.WritePNG(jxlImage, f)
		//} else {
		//
		//	// convert to regular Go image.Image
		//	img, err := jxlImage.ToImage()
		//	if err != nil {
		//		fmt.Printf("error when making image %v\n", err)
		//	}
		//
		//	buf := new(bytes.Buffer)
		//	if err := png.Encode(buf, img); err != nil {
		//		log.Fatalf("boomage %v", err)
		//	}
		//
		//	pngFileName = `c:\temp\testold.png`
		//	err = os.WriteFile(pngFileName, buf.Bytes(), 0666)
		//	if err != nil {
		//		log.Fatalf("boomage %v", err)
		//	}
		//}

		end := time.Now()
		fmt.Printf("decoding %s took %d ms but including png total time %d ms\n", orig, decodingDuration.Milliseconds(), end.Sub(start).Milliseconds())
	}
	fmt.Printf("TOTAL time for all decoding %d ms\n", time.Since(totalTime).Milliseconds())
}
