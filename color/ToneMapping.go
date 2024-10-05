package color

import (
	"errors"

	"github.com/kpfaulkner/jxl-go/jxlio"
)

type ToneMapping struct {
	IntensityTarget      float32
	MinNits              float32
	RelativeToMaxDisplay bool
	LinearBelow          float32
}

func NewToneMapping() *ToneMapping {
	tm := &ToneMapping{}

	tm.IntensityTarget = 255.0
	tm.MinNits = 0.0
	tm.RelativeToMaxDisplay = false
	tm.LinearBelow = 0

	return tm
}

func NewToneMappingWithReader(reader *jxlio.Bitreader) (*ToneMapping, error) {
	tm := &ToneMapping{}
	if reader.MustReadBool() {
		tm.IntensityTarget = 255.0
		tm.MinNits = 0.0
		tm.RelativeToMaxDisplay = false
		tm.LinearBelow = 0
	} else {
		tm.IntensityTarget = reader.MustReadF16()
		if tm.IntensityTarget <= 0 {
			return nil, errors.New("Intensity Target must be positive")
		}
		tm.MinNits = reader.MustReadF16()
		if tm.MinNits < 0 {
			return nil, errors.New("Min Nits must be positive")
		}
		if tm.MinNits > tm.IntensityTarget {
			return nil, errors.New("Min Nits must be at most the Intensity Target")
		}
		tm.RelativeToMaxDisplay = reader.MustReadBool()
		tm.LinearBelow = reader.MustReadF16()
		if tm.RelativeToMaxDisplay && (tm.LinearBelow < 0 || tm.LinearBelow > 1) {
			return nil, errors.New("Linear Below out of relative range")
		}
		if !tm.RelativeToMaxDisplay && tm.LinearBelow < 0 {
			return nil, errors.New("Linear Below must be nonnegative")
		}
	}
	return tm, nil
}

func (tm *ToneMapping) GetIntensityTarget() float32 {
	return tm.IntensityTarget
}
