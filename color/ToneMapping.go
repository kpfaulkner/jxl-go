package color

import (
	"errors"

	"github.com/kpfaulkner/jxl-go/jxlio"
)

type ToneMapping struct {
	intensityTarget      float32
	minNits              float32
	relativeToMaxDisplay bool
	linearBelow          float32
}

func NewToneMapping() *ToneMapping {
	tm := &ToneMapping{}

	tm.intensityTarget = 255.0
	tm.minNits = 0.0
	tm.relativeToMaxDisplay = false
	tm.linearBelow = 0

	return tm
}

func NewToneMappingWithReader(reader *jxlio.Bitreader) (*ToneMapping, error) {
	tm := &ToneMapping{}
	if reader.MustReadBool() {
		tm.intensityTarget = 255.0
		tm.minNits = 0.0
		tm.relativeToMaxDisplay = false
		tm.linearBelow = 0
	} else {
		tm.intensityTarget = reader.MustReadF16()
		if tm.intensityTarget <= 0 {
			return nil, errors.New("Intensity Target must be positive")
		}
		tm.minNits = reader.MustReadF16()
		if tm.minNits < 0 {
			return nil, errors.New("Min Nits must be positive")
		}
		if tm.minNits > tm.intensityTarget {
			return nil, errors.New("Min Nits must be at most the Intensity Target")
		}
		tm.relativeToMaxDisplay = reader.MustReadBool()
		tm.linearBelow = reader.MustReadF16()
		if tm.relativeToMaxDisplay && (tm.linearBelow < 0 || tm.linearBelow > 1) {
			return nil, errors.New("Linear Below out of relative range")
		}
		if !tm.relativeToMaxDisplay && tm.linearBelow < 0 {
			return nil, errors.New("Linear Below must be nonnegative")
		}
	}
	return tm, nil
}

func (tm *ToneMapping) GetIntensityTarget() float32 {
	return tm.intensityTarget
}
