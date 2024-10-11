package frame

import "github.com/kpfaulkner/jxl-go/jxlio"

type HFCoefficients struct {
	hfPreset uint32
}

func NewHFCoefficientsWithReader(reader *jxlio.Bitreader, frame *Frame, pass uint32, group uint32) (*HFCoefficients, error {
	hf := &HFCoefficients{}


	return hf, nil
}
