package color

type CIEXY struct {
	x float32
	y float32
}

func NewCIEXY(x float32, y float32) *CIEXY {
	cxy := &CIEXY{}
	cxy.x = x
	cxy.y = y
	return cxy
}
