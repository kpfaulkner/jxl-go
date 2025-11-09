package main

import (
	"fmt"
	"reflect"

	"github.com/kpfaulkner/jxl-go/frame"
)

// displays sizes of main structs to determine any padding wasteage
func memStats(input any) {

	rType := reflect.TypeOf(input)
	//fmt.Printf("Size of %s : %d bytes\n", rType.Name(), unsafe.Sizeof(input))
	fmt.Printf("Size of %s : %d bytes\n", rType.Name(), rType.Size())

	currentMaxAlignment := 1000000

	totalBytes := 0
	idealCurrentOffset := 0
	//rValue := reflect.ValueOf(bitDepthHeader)
	if rType.Kind() == reflect.Struct {
		for i := 0; i < rType.NumField(); i++ {

			fmt.Printf("  Name %s\n", rType.FieldByIndex([]int{i}).Name)
			fmt.Printf("    Current ttl  : %d\n", totalBytes)
			fmt.Printf("    Offset of    : %d bytes\n", rType.FieldByIndex([]int{i}).Offset)
			fmt.Printf("    Size of      : %d bytes\n", rType.FieldByIndex([]int{i}).Type.Size())
			fmt.Printf("    Alignment of : %d bytes", rType.FieldByIndex([]int{i}).Type.Align())

			if idealCurrentOffset != int(rType.FieldByIndex([]int{i}).Offset) {
				fmt.Printf(" XXXX not ideal offset, %d vs %d\n", idealCurrentOffset, int(rType.FieldByIndex([]int{i}).Offset))
			}

			if rType.FieldByIndex([]int{i}).Type.Align() > currentMaxAlignment {
				fmt.Printf("  ***check\n")
			} else {
				fmt.Printf("\n")
			}

			idealCurrentOffset += int(rType.FieldByIndex([]int{i}).Type.Size())
			currentMaxAlignment = rType.FieldByIndex([]int{i}).Type.Align()
			totalBytes += int(rType.FieldByIndex([]int{i}).Type.Size())

			fmt.Println()
		}
	}

	fmt.Printf("Total size calculated from field %d\n", totalBytes)
}

func main() {
	//memStats(core.JXLImage{})

	//memStats(entropy.EntropyStream{})
	//memStats(color.OpsinInverseMatrix{})

	//memStats(entropy.ANSSymbolDistribution{})

	//memStats(frame.Pass{})
	//memStats(frame.PassesInfo{})
	//memStats(frame.FrameHeader{})
	//memStats(frame.MATreeNode{})
	//memStats(jxlio.Bitreader{})
	//memStats(frame.Frame{})
	//memStats(frame.PassGroup{})
	memStats(frame.Quantizer{})
}
