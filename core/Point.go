package core

type Point struct {
	X int32
	Y int32
}

func NewPoint(x int32, y int32) *Point {
	return &Point{
		X: x,
		Y: y,
	}
}
