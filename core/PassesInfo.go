package core

type PassesInfo struct {
	numPasses  uint32
	numDS      uint32
	shift      []uint32
	downSample []uint32
	lastPass   []uint32
}

func NewPassesInfo() *PassesInfo {
	pi := &PassesInfo{}
	pi.numPasses = 1
	pi.numDS = 0
	pi.shift = []uint32{}
	pi.downSample = []uint32{}
	pi.lastPass = []uint32{}
	return pi
}
