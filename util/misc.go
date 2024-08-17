package util

func IfThenElse[T any](condition bool, a T, b T) T {
	if condition {
		return a
	}
	return b
}
