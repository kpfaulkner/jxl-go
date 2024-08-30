package core

import "github.com/kpfaulkner/jxl-go/jxlio"

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

func NewPassesInfoWithReader(reader *jxlio.Bitreader) (*PassesInfo, error) {
	pi := &PassesInfo{}
	pi.numPasses = reader.MustReadU32(1, 0, 2, 0, 3, 0, 4, 3)
	if pi.numPasses != 1 {
		pi.numDS = reader.MustReadU32(0, 0, 1, 0, 2, 0, 3, 1)
	} else {
		pi.numDS = 0
	}

	pi.shift = make([]uint32, pi.numPasses-1)
	for i := 0; i < len(pi.shift); i++ {
		pi.shift[i] = uint32(reader.MustReadBits(2))
	}
	pi.downSample = make([]uint32, pi.numDS+1)
	for i := 0; i < int(pi.numDS); i++ {
		pi.downSample[i] = 1 << reader.MustReadBits(2)
	}
	pi.lastPass = make([]uint32, pi.numDS+1)
	for i := 0; i < int(pi.numDS); i++ {
		pi.lastPass[i] = reader.MustReadU32(0, 0, 1, 0, 2, 0, 0, 3)
	}
	pi.downSample[pi.numDS] = 1
	pi.lastPass[pi.numDS] = pi.numPasses - 1

	return pi, nil
}
