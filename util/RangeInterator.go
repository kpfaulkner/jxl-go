package util

import (
	"io"
)

// COULD try out the new Go 1.23 iter package, but to keep backwards compatibility will
// just use something basic and simple.
func RangeIterator(startX uint32, startY uint32, endX uint32, endY uint32) func() (*IntPoint, error) {
	x := startX
	y := startY
	return func() (*IntPoint, error) {
		if x > endX {
			x = startX
			y++
		}
		if y > endY {
			return nil, io.EOF
		}
		x++
		return &IntPoint{x, y}, nil
	}
}

func RangeIteratorWithIntPoint(ip IntPoint) func() (*IntPoint, error) {
	return RangeIterator(0, 0, ip.X, ip.Y)
}
