package core

import (
	"errors"

	"github.com/kpfaulkner/jxl-go/jxlio"
	log "github.com/sirupsen/logrus"
)

type PreviewHeader struct {
	height uint32
	width  uint32
}

func NewPreviewHeader(reader *jxlio.Bitreader) (*PreviewHeader, error) {
	ph := &PreviewHeader{}

	var err error

	div8 := reader.MustReadBool()
	if div8 {
		ph.height = reader.MustReadU32(16, 0, 32, 0, 1, 5, 33, 9)
	} else {
		ph.height = reader.MustReadU32(1, 6, 65, 8, 321, 10, 1345, 12)
	}
	ratio := reader.MustReadBits(3)
	if ratio != 0 {
		ph.width, err = getWidthFromRatio(ratio, ph.height)
		if err != nil {
			log.Errorf("Error getting width from ratio: %v\n", err)
			return nil, err
		}
	} else {
		if div8 {
			ph.width = reader.MustReadU32(16, 0, 32, 0, 1, 5, 33, 9)
		} else {
			ph.width = reader.MustReadU32(1, 6, 65, 8, 321, 10, 1345, 12)
		}
	}

	if ph.width > 4096 || ph.height > 4096 {
		log.Errorf("preview width or preview height too large: %d, %d", ph.width, ph.height)
		return nil, errors.New("preview width or preview height too large")
	}

	return ph, nil
}
