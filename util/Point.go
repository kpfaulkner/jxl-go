package util

type Point struct {
	X int32
	Y int32
}

func NewPoint(y int32, x int32) *Point {
	return &Point{
		X: x,
		Y: y,
	}
}
