package main

import (
	"fmt"

	"github.com/kpfaulkner/jxl-go/core"
)

func main() {
	fmt.Printf("So it begins...\n")

	jxl := core.NewJXLDecoder(core.WithInputFilename(`c:\temp\lossless.jxl`))

	if err := jxl.Decode(); err != nil {
		fmt.Printf("Error decoding: %v\n", err)
	} else {
		fmt.Printf("Decoded successfully\n")
	}

}
