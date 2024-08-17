package color

import (
	"errors"

	"github.com/kpfaulkner/jxl-go/jxlio"
)

type ColorEncodingBundle struct {
	useIccProfile   bool
	colorEncoding   int32
	whitePoint      int32
	white           *CIEXY
	primaries       int32
	prim            *CIEPrimaries
	tf              uint32
	renderingIntent uint32
}

func NewColorEncodingBundle() (*ColorEncodingBundle, error) {
	ceb := &ColorEncodingBundle{}
	ceb.useIccProfile = false
	ceb.colorEncoding = CE_RGB
	ceb.whitePoint = WP_D65
	ceb.white = getWhitePoint(ceb.whitePoint)
	ceb.primaries = PRI_SRGB
	ceb.prim = getPrimaries(ceb.primaries)
	ceb.tf = TF_SRGB
	ceb.renderingIntent = RI_RELATIVE
	return ceb, nil
}

func NewColorEncodingBundleWithReader(reader *jxlio.Bitreader) (*ColorEncodingBundle, error) {
	ceb := &ColorEncodingBundle{}
	allDefault := reader.MustReadBool()

	var err error
	if !allDefault {
		ceb.useIccProfile = reader.MustReadBool()
	}

	if !allDefault {
		ceb.colorEncoding = reader.MustReadEnum()
	} else {
		ceb.colorEncoding = CE_RGB
	}

	if ValidateColorEncoding(ceb.colorEncoding) {
		return nil, errors.New("Invalid ColorSpace enum")
	}

	if !allDefault && !ceb.useIccProfile && ceb.colorEncoding != CE_XYB {
		ceb.whitePoint = reader.MustReadEnum()
	} else {
		ceb.whitePoint = WP_D65
	}

	if ValidateWhitePoint(ceb.whitePoint) {
		return nil, errors.New("Invalid WhitePoint enum")
	}

	if ceb.whitePoint == WP_CUSTOM {
		white, err := NewCustomXY(reader)
		if err != nil {
			return nil, err
		}
		ceb.white = &white.CIEXY
	} else {
		ceb.white = getWhitePoint(ceb.whitePoint)
	}

	if !allDefault && !ceb.useIccProfile && ceb.colorEncoding != CE_XYB && ceb.colorEncoding != CE_GRAY {
		ceb.primaries = reader.MustReadEnum()
	} else {
		ceb.primaries = PRI_SRGB
	}

	return ceb, nil
}

func getPrimaries(primaries uint32) *CIEPrimaries {

}

func getWhitePoint(point int32) *CIEXY {

}
