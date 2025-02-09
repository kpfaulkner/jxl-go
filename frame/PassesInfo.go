package frame

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
	if numPasses, err := reader.ReadU32(1, 0, 2, 0, 3, 0, 4, 3); err != nil {
		return nil, err
	} else {
		pi.numPasses = numPasses
	}
	if pi.numPasses != 1 {
		if numDS, err := reader.ReadU32(0, 0, 1, 0, 2, 0, 3, 1); err != nil {
			return nil, err
		} else {
			pi.numDS = numDS
		}
	} else {
		pi.numDS = 0
	}

	pi.shift = make([]uint32, pi.numPasses)
	for i := uint32(0); i < pi.numPasses-1; i++ {
		pi.shift[i] = uint32(reader.MustReadBits(2))
	}
	pi.shift[pi.numPasses-1] = 0
	pi.downSample = make([]uint32, pi.numDS+1)
	for i := 0; i < int(pi.numDS); i++ {
		pi.downSample[i] = 1 << reader.MustReadBits(2)
	}
	pi.lastPass = make([]uint32, pi.numDS+1)
	for i := 0; i < int(pi.numDS); i++ {
		if lastPass, err := reader.ReadU32(0, 0, 1, 0, 2, 0, 0, 3); err != nil {
			return nil, err
		} else {
			pi.lastPass[i] = lastPass
		}
	}
	pi.downSample[pi.numDS] = 1
	pi.lastPass[pi.numDS] = pi.numPasses - 1

	return pi, nil
}
