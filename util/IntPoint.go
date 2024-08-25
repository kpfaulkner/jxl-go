package util

import "math"

var (
	ONE  = IntPoint{1, 1}
	ZERO = IntPoint{0, 0}
)

// TODO(kpfaulkner) confirm if X/Y should be signed or unsigned.
type IntPoint struct {
	X uint32
	Y uint32
}

func coordinates(index uint32, rowStride uint32) IntPoint {
	return IntPoint{
		X: index % rowStride,
		Y: index / rowStride,
	}
}

func NewIntPoint(dim int) IntPoint {
	return IntPoint{uint32(dim), uint32(dim)}
}

func NewIntPointWithXY(x uint32, y uint32) IntPoint {
	return IntPoint{x, y}
}

func (ip IntPoint) times(factor uint32) IntPoint {
	return IntPoint{ip.X * factor, ip.Y * factor}
}

func (ip IntPoint) timesWithIntPoint(p IntPoint) IntPoint {
	return IntPoint{ip.X * p.X, ip.Y * p.Y}
}

func (ip IntPoint) ceilDiv(factor uint32) IntPoint {
	return IntPoint{CeilDiv(ip.X, factor), CeilDiv(ip.Y, factor)}
}

func (ip IntPoint) ceilDivWithIntPoint(p IntPoint) IntPoint {
	return IntPoint{CeilDiv(ip.X, p.X), CeilDiv(ip.Y, p.Y)}
}

func (ip IntPoint) transpose() IntPoint {
	return IntPoint{X: ip.Y, Y: ip.X}
}

func (ip IntPoint) unwrapCoord(rowStride uint32) uint32 {
	return ip.Y*rowStride + ip.X
}

func (ip IntPoint) shiftRight(hShift int, vShift int) IntPoint {
	x := ip.X >> uint32(hShift)
	y := ip.Y >> uint32(vShift)
	return IntPoint{X: x, Y: y}
}

func (ip IntPoint) shiftRightWithShift(shift int) IntPoint {
	return ip.shiftRight(shift, shift)
}

func (ip IntPoint) shiftRightWithIntPoint(p IntPoint) IntPoint {
	return ip.shiftRight(int(p.X), int(p.Y))
}

func (ip IntPoint) shiftLeftWithShift(shift int) IntPoint {
	return ip.shiftLeft(shift, shift)
}

func (ip IntPoint) shiftLeftWithIntPoint(shift IntPoint) IntPoint {
	return ip.shiftLeft(int(shift.X), int(shift.Y))
}

func (ip IntPoint) shiftLeft(hShift int, vShift int) IntPoint {
	x := ip.X << uint32(hShift)
	y := ip.Y << uint32(vShift)
	return IntPoint{X: x, Y: y}
}

func (ip IntPoint) minus(p IntPoint) IntPoint {
	return IntPoint{X: ip.X - p.X, Y: ip.Y - p.Y}
}

func (ip IntPoint) min(p IntPoint) IntPoint {
	return IntPoint{X: uint32(math.Min(float64(ip.X), float64(p.X))), Y: uint32(math.Min(float64(ip.Y), float64(p.Y)))}
}
