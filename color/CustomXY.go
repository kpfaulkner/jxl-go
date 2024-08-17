package color

import (
	"github.com/kpfaulkner/jxl-go/jxlio"
)

type CustomXY struct {
	CIEXY
}

func NewCustomXY(reader *jxlio.Bitreader) (*CustomXY, error) {
	cxy := &CustomXY{}

	ciexy, err := cxy.readCustom(reader)
	if err != nil {
		return nil, err
	}
	cxy.CIEXY = *ciexy
	return cxy, nil
}

func (cxy *CustomXY) readCustom(reader *jxlio.Bitreader) (*CIEXY, error) {
	ux := reader.MustReadU32(0, 19, 524288, 19, 1048576, 20, 2097152, 21)
	uy := reader.MustReadU32(0, 19, 524288, 19, 1048576, 20, 2097152, 21)
	x := float32(jxlio.UnpackSigned(int32(ux))) * 1e-6
	y := float32(jxlio.UnpackSigned(int32(uy))) * 1e-6

	return NewCIEXY(x, y), nil
}
