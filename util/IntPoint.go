package util

import "math"

var (
	ONE  = IntPoint{1, 1}
	ZERO = IntPoint{0, 0}
)

type IntPoint struct {
	x uint32
	y uint32
}

func coordinates(index uint32, rowStride uint32) IntPoint {
	return IntPoint{
		x: index % rowStride,
		y: index / rowStride,
	}
}

func NewIntPoint(dim int) IntPoint {
	return IntPoint{uint32(dim), uint32(dim)}
}

func (ip IntPoint) times(factor uint32) IntPoint {
	return IntPoint{ip.x * factor, ip.y * factor}
}

func (ip IntPoint) timesWithIntPoint(p IntPoint) IntPoint {
	return IntPoint{ip.x * p.x, ip.y * p.y}
}

func (ip IntPoint) ceilDiv(factor uint32) IntPoint {
	return IntPoint{CeilDiv(ip.x, factor), CeilDiv(ip.y, factor)}
}

func (ip IntPoint) ceilDivWithIntPoint(p IntPoint) IntPoint {
	return IntPoint{CeilDiv(ip.x, p.x), CeilDiv(ip.y, p.y)}
}

func (ip IntPoint) transpose() IntPoint {
	return IntPoint{x: ip.y, y: ip.x}
}

func (ip IntPoint) unwrapCoord(rowStride uint32) uint32 {
	return ip.y*rowStride + ip.x
}

func (ip IntPoint) shiftRight(hShift int, vShift int) IntPoint {
	x := ip.x >> uint32(hShift)
	y := ip.y >> uint32(vShift)
	return IntPoint{x: x, y: y}
}

func (ip IntPoint) shiftRightWithShift(shift int) IntPoint {
	return ip.shiftRight(shift, shift)
}

func (ip IntPoint) shiftRightWithIntPoint(p IntPoint) IntPoint {
	return ip.shiftRight(int(p.x), int(p.y))
}

func (ip IntPoint) shiftLeftWithShift(shift int) IntPoint {
	return ip.shiftLeft(shift, shift)
}

func (ip IntPoint) shiftLeftWithIntPoint(shift IntPoint) IntPoint {
	return ip.shiftLeft(int(shift.x), int(shift.y))
}

func (ip IntPoint) shiftLeft(hShift int, vShift int) IntPoint {
	x := ip.x << uint32(hShift)
	y := ip.y << uint32(vShift)
	return IntPoint{x: x, y: y}
}

func (ip IntPoint) minus(p IntPoint) IntPoint {
	return IntPoint{x: ip.x - p.x, y: ip.y - p.y}
}

func (ip IntPoint) min(p IntPoint) IntPoint {
	return IntPoint{x: uint32(math.Min(float64(ip.x), float64(p.x))), y: uint32(math.Min(float64(ip.y), float64(p.y)))}
}
