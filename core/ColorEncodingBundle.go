package core

import (
	"github.com/kpfaulkner/jxl-go/jxlio"
)

const (
	CE_RGB   = 0
	WP_D65   = 1
	PRI_SRGB = 2
	TF_SRGB  = 3

	RT_RELATIVE = 4
)

type CIEXY struct {
}

type CIEPrimaries struct {
}

type ColorEncodingBundle struct {
	useIccProfile   bool
	colorEncoding   uint32
	whitePoint      uint32
	white           *CIEXY
	primaries       uint32
	prim            *CIEPrimaries
	tf              uint32
	renderingIntent uint32
}

func NewColorEncodingBundle(reader *jxlio.Bitreader) (*ColorEncodingBundle, error) {
	ceb := &ColorEncodingBundle{}
	ceb.useIccProfile = false
	ceb.colorEncoding = CE_RGB
	ceb.whitePoint = WP_D65
	ceb.white = getWhitePoint(ceb.whitePoint)
	ceb.primaries = PRI_SRGB
	ceb.prim = getPrimaries(ceb.primaries)
	ceb.tf = TF_SRGB
	ceb.renderingIntent = RT_RELATIVE
	return ceb, nil
}

func getPrimaries(primaries uint32) *CIEPrimaries {

}

func getWhitePoint(point uint32) *CIEXY {

}
