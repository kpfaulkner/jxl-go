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

func (cp *CIEPrimaries) Matches(b *CIEPrimaries) bool {

	if b == nil {
		return false
	}
	return cp.red.Matches(b.red) && cp.green.Matches(b.green) && cp.blue.Matches(b.blue)
}
