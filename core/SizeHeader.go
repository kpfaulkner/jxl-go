package core

import (
	"fmt"

	"github.com/kpfaulkner/jxl-go/jxlio"
	"github.com/kpfaulkner/jxl-go/util"
	log "github.com/sirupsen/logrus"
)

func readSizeHeader(reader *jxlio.Bitreader, level int32) (*Dimension, error) {
	dim := &Dimension{}
	var err error

	div8 := reader.MustReadBool()
	if div8 {
		dim.height = 1 + uint32(reader.MustReadBits(5))<<3
	} else {
		dim.height = reader.MustReadU32(1, 9, 1, 13, 1, 18, 1, 30)
	}
	ratio := reader.MustReadBits(3)
	if ratio != 0 {
		dim.width, err = getWidthFromRatio(uint32(ratio), dim.height)
		if err != nil {

			log.Errorf("Error getting width from ratio: %v\n", err)
			return nil, err
		}
	} else {
		if div8 {
			dim.width = 1 + uint32(reader.MustReadBits(5))<<3
		} else {
			dim.width = reader.MustReadU32(1, 9, 1, 13, 1, 18, 1, 30)
		}
	}

	maxDim := util.IfThenElse[uint64](level <= 5, 1<<18, 1<<28)
	maxTimes := util.IfThenElse[uint64](level <= 5, 1<<30, 1<<40)
	if dim.width > uint32(maxDim) || dim.height > uint32(maxDim) {
		log.Errorf("Invalid size header: %d x %d", dim.width, dim.height)
		return nil, fmt.Errorf("Invalid size header: %d x %d", dim.width, dim.height)
	}
	if uint64(dim.width*dim.height) > maxTimes {
		log.Errorf("Width times height too large: %d %d", dim.width, dim.height)
		return nil, fmt.Errorf("Width times height too large: %d %d", dim.width, dim.height)
	}

	return dim, nil
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
