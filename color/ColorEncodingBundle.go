package color

import (
	"errors"

	"github.com/kpfaulkner/jxl-go/jxlio"
)

type ColorEncodingBundle struct {
	UseIccProfile   bool
	ColorEncoding   int32
	WhitePoint      int32
	White           *CIEXY
	Primaries       int32
	Prim            *CIEPrimaries
	Tf              int32
	RenderingIntent int32
}

func NewColorEncodingBundle() (*ColorEncodingBundle, error) {
	ceb := &ColorEncodingBundle{}
	ceb.UseIccProfile = false
	ceb.ColorEncoding = CE_RGB
	ceb.WhitePoint = WP_D65
	ceb.White = getWhitePoint(ceb.WhitePoint)
	ceb.Primaries = PRI_SRGB
	ceb.Prim = getPrimaries(ceb.Primaries)
	ceb.Tf = TF_SRGB
	ceb.RenderingIntent = RI_RELATIVE
	return ceb, nil
}

func NewColorEncodingBundleWithReader(reader *jxlio.Bitreader) (*ColorEncodingBundle, error) {
	ceb := &ColorEncodingBundle{}
	allDefault := reader.MustReadBool()

	if !allDefault {
		ceb.UseIccProfile = reader.MustReadBool()
	}

	if !allDefault {
		ceb.ColorEncoding = reader.MustReadEnum()
	} else {
		ceb.ColorEncoding = CE_RGB
	}

	if ValidateColorEncoding(ceb.ColorEncoding) {
		return nil, errors.New("Invalid ColorSpace enum")
	}

	if !allDefault && !ceb.UseIccProfile && ceb.ColorEncoding != CE_XYB {
		ceb.WhitePoint = reader.MustReadEnum()
	} else {
		ceb.WhitePoint = WP_D65
	}

	if ValidateWhitePoint(ceb.WhitePoint) {
		return nil, errors.New("Invalid WhitePoint enum")
	}

	if ceb.WhitePoint == WP_CUSTOM {
		white, err := NewCustomXY(reader)
		if err != nil {
			return nil, err
		}
		ceb.White = &white.CIEXY
	} else {
		ceb.White = getWhitePoint(ceb.WhitePoint)
	}

	if !allDefault && !ceb.UseIccProfile && ceb.ColorEncoding != CE_XYB && ceb.ColorEncoding != CE_GRAY {
		ceb.Primaries = reader.MustReadEnum()
	} else {
		ceb.Primaries = PRI_SRGB
	}

	if !ValidatePrimaries(ceb.Primaries) {
		return nil, errors.New("Invalid Primaries enum")
	}

	if ceb.Primaries == PRI_CUSTOM {
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
		ceb.Prim = NewCIEPrimaries(&pRed.CIEXY, &pGreen.CIEXY, &pBlue.CIEXY)
	} else {
		ceb.Prim = GetPrimaries(ceb.Primaries)
	}

	if !allDefault && !ceb.UseIccProfile {
		useGamma := reader.MustReadBool()
		if useGamma {
			ceb.Tf = int32(reader.MustReadBits(24))
		} else {
			ceb.Tf = (1 << 24) + reader.MustReadEnum()
		}
		if ValidateTransfer(ceb.Tf) {
			return nil, errors.New("Illegal transfer function")
		}
		ceb.RenderingIntent = reader.MustReadEnum()
		if ValidateRenderingIntent(ceb.RenderingIntent) {
			return nil, errors.New("Invalid RenderingIntent enum")
		}
	} else {
		ceb.Tf = TF_SRGB
		ceb.RenderingIntent = RI_RELATIVE
	}

	return ceb, nil
}

func getWhitePoint(point int32) *CIEXY {

}
