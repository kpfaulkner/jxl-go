package color

type CIEPrimaries struct {
	red   *CIEXY
	green *CIEXY
	blue  *CIEXY
}

func NewCIEPrimaries(red *CIEXY, green *CIEXY, blue *CIEXY) *CIEPrimaries {
	cp := &CIEPrimaries{}
	cp.red = red
	cp.green = green
	cp.blue = blue
	return cp
}
