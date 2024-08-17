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
	tf              int32
	renderingIntent int32
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

	if !ValidatePrimaries(ceb.primaries) {
		return nil, errors.New("Invalid Primaries enum")
	}

	if ceb.primaries == PRI_CUSTOM {
		pRed, err := NewCustomXY(reader)
		if err != nil {
			return nil, err
		}
		pGreen, err := NewCustomXY(reader)
		if err != nil {
			return nil, err
		}
		pBlue, err := NewCustomXY(reader)
		if err != nil {
			return nil, err
		}
		ceb.prim = NewCIEPrimaries(&pRed.CIEXY, &pGreen.CIEXY, &pBlue.CIEXY)
	} else {
		ceb.prim = GetPrimaries(ceb.primaries)
	}

	if !allDefault && !ceb.useIccProfile {
		useGamma := reader.MustReadBool()
		if useGamma {
			ceb.tf = int32(reader.MustReadBits(24))
		} else {
			ceb.tf = (1 << 24) + reader.MustReadEnum()
		}
		if ValidateTransfer(ceb.tf) {
			return nil, errors.New("Illegal transfer function")
		}
		ceb.renderingIntent = reader.MustReadEnum()
		if ValidateRenderingIntent(ceb.renderingIntent) {
			return nil, errors.New("Invalid RenderingIntent enum")
		}
	} else {
		ceb.tf = TF_SRGB
		ceb.renderingIntent = RI_RELATIVE
	}

	return ceb, nil
}

func getWhitePoint(point int32) *CIEXY {

}
