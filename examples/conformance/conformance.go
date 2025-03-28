package main

import (
	"bytes"
	"flag"
	"fmt"
	"image/png"
	//"image/png"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/kpfaulkner/jxl-go/core"
	log "github.com/sirupsen/logrus"
)

var (
	failedFiles    []string
	succeededFiles []string
)

var hasError bool

func doProcessing(filename string) error {
	defer func() {
		if r := recover(); r != nil {
			hasError = true
		}
	}()
	fmt.Printf("file %s\n", filename)
	f, err := os.ReadFile(filename)
	if err != nil {
		log.Errorf("Error opening file: %v\n", err)
		return err
	}

	start := time.Now()
	//var img image.Image
	r := bytes.NewReader(f)
	jxl := core.NewJXLDecoder(r, nil)
	//p := profile.Start(profile.CPUProfile, profile.ProfilePath("."))
	var jxlImage *core.JXLImage
	if jxlImage, err = jxl.Decode(); err != nil {
		fmt.Printf("Error decoding: %v\n", err)
		return err
	}
	//p.Stop()
	fmt.Printf("decoding took %d ms\n", time.Since(start).Milliseconds())
	fmt.Printf("Has alpha %v\n", jxlImage.HasAlpha())
	fmt.Printf("Num extra channels (inc alpha) %d\n", jxlImage.NumExtraChannels())

	if ct, err := jxlImage.GetExtraChannelType(0); err == nil {
		fmt.Printf("channel 3 type %d\n", ct)
	}

	ext := path.Ext(filename)
	pngFileName := filename[:len(filename)-len(ext)] + ".png"

	// if ICC profile then use custom PNG writer... otherwise use default Go encoder.
	if jxlImage.HasICCProfile() {
		f, err := os.Create(pngFileName)
		if err != nil {
			return err
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
			return err
		}

		err = os.WriteFile(pngFileName, buf.Bytes(), 0666)
		if err != nil {
			return err
		}
	}

	end := time.Now()
	fmt.Printf("decoding total time %d ms\n", end.Sub(start).Milliseconds())
	return nil
}

func main() {

	conformanceDir := flag.String("c", `C:\Users\kenfa\projects\conformance\testcases`, "Conformance directory")
	flag.Parse()

	if *conformanceDir == "" {
		fmt.Printf("Conformance directory not specified")
		return
	}

	var allFiles []string
	err := filepath.Walk(*conformanceDir,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if filepath.Ext(path) == ".jxl" {
				allFiles = append(allFiles, path)
			}
			return nil
		})
	if err != nil {
		log.Println(err)
	}

	for _, file := range allFiles {
		hasError = false
		err = doProcessing(file)
		if err != nil {
			hasError = true
		}

		if hasError {
			failedFiles = append(failedFiles, file)
		} else {
			succeededFiles = append(succeededFiles, file)
		}

	}

	for _, f := range succeededFiles {
		fmt.Printf("%s succeeded\n", f)
	}

	for _, f := range failedFiles {
		fmt.Printf("%s failed\n", f)
	}
}
