package main

import (
	"fmt"
	"reflect"

	"github.com/kpfaulkner/jxl-go/core"
)

// displays sizes of main structs to determine any padding wasteage
func memStats(input any) {

	rType := reflect.TypeOf(input)
	//fmt.Printf("Size of %s : %d bytes\n", rType.Name(), unsafe.Sizeof(input))
	fmt.Printf("Size of %s : %d bytes\n", rType.Name(), rType.Size())

	//rValue := reflect.ValueOf(bitDepthHeader)
	if rType.Kind() == reflect.Struct {
		for i := 0; i < rType.NumField(); i++ {

			fmt.Printf("  Name %s\n", rType.FieldByIndex([]int{i}).Name)
			fmt.Printf("    Offset of    : %d bytes\n", rType.FieldByIndex([]int{i}).Offset)
			fmt.Printf("    Size of      : %d bytes\n", rType.FieldByIndex([]int{i}).Type.Size())
			fmt.Printf("    Alignment of : %d bytes\n", rType.FieldByIndex([]int{i}).Type.Align())
			fmt.Println()
		}
	}
}

func main() {
	memStats(core.JXLImage{})
}
