package core

import (
	"fmt"

	"github.com/kpfaulkner/jxl-go/jxlio"
	"github.com/kpfaulkner/jxl-go/util"
	log "github.com/sirupsen/logrus"
)

type SizeHeader struct {
	height uint32
	width  uint32
}

func NewSizeHeader(reader *jxlio.Bitreader, level int32) (*SizeHeader, error) {
	sh := &SizeHeader{}
	var err error

	div8 := reader.TryReadBool()
	if div8 {
		sh.height = 1 + uint32(reader.TryReadBits(5))<<3
	} else {
		sh.height = reader.MustReadU32(1, 9, 1, 13, 1, 18, 1, 30)
	}
	ratio := reader.TryReadBits(3)
	if ratio != 0 {
		sh.width, err = getWidthFromRatio(uint32(ratio), sh.height)
		if err != nil {

			log.Errorf("Error getting width from ratio: %v\n", err)
			return nil, err
		}
	} else {
		if div8 {
			sh.width = 1 + uint32(reader.TryReadBits(5))<<3
		} else {
			sh.width = reader.MustReadU32(1, 9, 1, 13, 1, 18, 1, 30)
		}
	}

	maxDim := util.IfThenElse[uint64](level <= 5, 1<<18, 1<<28)
	maxTimes := util.IfThenElse[uint64](level <= 5, 1<<30, 1<<40)
	if sh.width > uint32(maxDim) || sh.height > uint32(maxDim) {
		log.Errorf("Invalid size header: %d x %d", sh.width, sh.height)
		return nil, fmt.Errorf("Invalid size header: %d x %d", sh.width, sh.height)
	}
	if uint64(sh.width*sh.height) > maxTimes {
		log.Errorf("Width times height too large: %d %d", sh.width, sh.height)
		return nil, fmt.Errorf("Width times height too large: %d %d", sh.width, sh.height)
	}

	return sh, nil
}

func getWidthFromRatio(ratio uint32, height uint32) (uint32, error) {
	switch ratio {
	case 1:
		return height, nil
	case 2:
		return height * 6 / 5, nil
	case 3:
		return height * 4 / 3, nil
	case 4:
		return height * 3 / 2, nil
	case 5:
		return height * 16 / 9, nil
	case 6:
		return height * 5 / 4, nil
	case 7:
		return height * 2, nil
	default:
		return 0, fmt.Errorf("invalid ratio: %d", ratio)
	}
}
