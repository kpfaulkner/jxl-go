package main

import (
	"fmt"
	"unsafe"

	"github.com/kpfaulkner/jxl-go/bundle"
	"github.com/kpfaulkner/jxl-go/color"
	"github.com/kpfaulkner/jxl-go/core"
	"github.com/kpfaulkner/jxl-go/image"
)

// displays sizes of main structs to determine any padding wasteage
func main() {

	bitDepthHeader := bundle.BitDepthHeader{}
	fmt.Printf("Size of BitDepthHeader: %d bytes\n", unsafe.Sizeof(bitDepthHeader))

	extensions := bundle.Extensions{}
	fmt.Printf("Size of Extensions: %d bytes\n", unsafe.Sizeof(extensions))

	extraChannelInfo := bundle.ExtraChannelInfo{}
	fmt.Printf("Size of ExtraChannelInfo: %d bytes\n", unsafe.Sizeof(extraChannelInfo))

	fmt.Printf("Alignment of extraChannelInfo.EcType: %d bytes\n", unsafe.Alignof(extraChannelInfo.EcType))
	fmt.Printf("Alignment of extraChannelInfo.DimShift: %d bytes\n", unsafe.Alignof(extraChannelInfo.DimShift))
	fmt.Printf("Alignment of extraChannelInfo.EcType: %d bytes\n", unsafe.Alignof(extraChannelInfo.EcType))
	fmt.Printf("Alignment of extraChannelInfo.Red: %d bytes\n", unsafe.Alignof(extraChannelInfo.Red))
	fmt.Printf("Alignment of extraChannelInfo.Green: %d bytes\n", unsafe.Alignof(extraChannelInfo.Green))
	fmt.Printf("Alignment of extraChannelInfo.Blue: %d bytes\n", unsafe.Alignof(extraChannelInfo.Blue))
	fmt.Printf("Alignment of extraChannelInfo.Solidity: %d bytes\n", unsafe.Alignof(extraChannelInfo.Solidity))
	fmt.Printf("Alignment of extraChannelInfo.BitDepth: %d bytes\n", unsafe.Alignof(extraChannelInfo.BitDepth))
	fmt.Printf("Alignment of extraChannelInfo.Name: %d bytes\n", unsafe.Alignof(extraChannelInfo.Name))
	fmt.Printf("Alignment of extraChannelInfo.AlphaAssociated: %d bytes\n", unsafe.Alignof(extraChannelInfo.AlphaAssociated))

	imageHeader := bundle.ImageHeader{}
	fmt.Printf("Size of ImageHeader: %d bytes\n", unsafe.Sizeof(imageHeader))
	fmt.Printf("Alignment of imageHeader.Level: %d bytes\n", unsafe.Alignof(imageHeader.Level))
	fmt.Printf("Alignment of imageHeader.Orientation: %d bytes\n", unsafe.Alignof(imageHeader.Orientation))
	fmt.Printf("Alignment of imageHeader.OrientedWidth: %d bytes\n", unsafe.Alignof(imageHeader.OrientedWidth))
	fmt.Printf("Alignment of imageHeader.OrientedHeight: %d bytes\n", unsafe.Alignof(imageHeader.OrientedHeight))
	fmt.Printf("Alignment of imageHeader.Size: %d bytes\n", unsafe.Alignof(imageHeader.Size))
	fmt.Printf("Alignment of imageHeader.PreviewSize: %d bytes\n", unsafe.Alignof(imageHeader.PreviewSize))
	fmt.Printf("Alignment of imageHeader.AnimationHeader: %d bytes\n", unsafe.Alignof(imageHeader.AnimationHeader))
	fmt.Printf("Alignment of imageHeader.AlphaIndices: %d bytes\n", unsafe.Alignof(imageHeader.AlphaIndices))
	fmt.Printf("Alignment of imageHeader.UpWeights: %d bytes\n", unsafe.Alignof(imageHeader.UpWeights))
	fmt.Printf("Alignment of imageHeader.XybEncoded: %d bytes\n", unsafe.Alignof(imageHeader.XybEncoded))

	colourEncodingBundle := color.ColorEncodingBundle{}
	fmt.Printf("Size of ColorEncodingBundle: %d bytes\n", unsafe.Sizeof(colourEncodingBundle))

	opsInverseMatrix := color.OpsinInverseMatrix{}
	fmt.Printf("Size of OpsinInverseMatrix: %d bytes\n", unsafe.Sizeof(opsInverseMatrix))
	fmt.Printf("Alignment of OpsinInverseMatrix.WhitePoint: %d bytes\n", unsafe.Alignof(opsInverseMatrix.WhitePoint))
	fmt.Printf("Size of OpsinInverseMatrix.WhitePoint: %d bytes\n", unsafe.Sizeof(opsInverseMatrix.WhitePoint))
	fmt.Printf("Alignment of OpsinInverseMatrix.CbrtOpsinBias: %d bytes\n", unsafe.Alignof(opsInverseMatrix.CbrtOpsinBias))
	fmt.Printf("Size of OpsinInverseMatrix.CbrtOpsinBias: %d bytes\n", unsafe.Sizeof(opsInverseMatrix.CbrtOpsinBias))
	fmt.Printf("Alignment of OpsinInverseMatrix.Matrix: %d bytes\n", unsafe.Alignof(opsInverseMatrix.Matrix))
	fmt.Printf("Size of OpsinInverseMatrix.Matrix: %d bytes\n", unsafe.Sizeof(opsInverseMatrix.Matrix))
	fmt.Printf("Alignment of OpsinInverseMatrix.Primaries: %d bytes\n", unsafe.Alignof(opsInverseMatrix.Primaries))
	fmt.Printf("Size of OpsinInverseMatrix.Primaries: %d bytes\n", unsafe.Sizeof(opsInverseMatrix.Primaries))

	toneMapping := color.ToneMapping{}
	fmt.Printf("Size of ToneMapping: %d bytes\n", unsafe.Sizeof(toneMapping))

	imageBuffer := image.ImageBuffer{}
	fmt.Printf("Size of ImageBuffer: %d bytes\n", unsafe.Sizeof(imageBuffer))

	jxlImage := core.JXLImage{}
	fmt.Printf("Size of JXLImage: %d bytes\n", unsafe.Sizeof(jxlImage))

}
