package util

import (
	"cmp"
	"math"
	"math/bits"
)

func SignedPow(base float32, exponent float32) float32 {
	if base < 0 {
		return float32(math.Pow(float64(-base), float64(exponent)))
	}
	return float32(math.Pow(float64(base), float64(exponent)))
}

func CeilLog1p(x int64) int {
	xx := bits.LeadingZeros64(uint64(x))
	return 64 - xx
}

func Max[T cmp.Ordered](args ...T) T {
	if len(args) == 0 {
		return *new(T)
	}

	if isNan(args[0]) {
		return args[0]
	}

	max := args[0]
	for _, arg := range args[1:] {

		if isNan(arg) {
			return arg
		}

		if arg > max {
			max = arg
		}
	}
	return max
}

func Min[T cmp.Ordered](args ...T) T {
	if len(args) == 0 {
		return *new(T)
	}

	if isNan(args[0]) {
		return args[0]
	}

	min := args[0]
	for _, arg := range args[1:] {

		if isNan(arg) {
			return arg
		}

		if arg < min {
			min = arg
		}
	}
	return min
}

func isNan[T comparable](arg T) bool {
	return arg != arg
}
