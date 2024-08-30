package core

import (
	"github.com/kpfaulkner/jxl-go/bundle"
	"github.com/kpfaulkner/jxl-go/jxlio"
)

type RestorationFilter struct {
	gab                bool
	customGab          bool
	gab1Weights        []float32
	gab2Weights        []float32
	epfIterations      uint32
	epfSharpCustom     bool
	epfSharpLut        []float32
	epfChannelScale    []float32
	epfSigmaCustom     bool
	epfQuantMul        float32
	epfPass0SigmaScale float32
	epfPass2SigmaScale float32
	epfBorderSadMul    float32
	epfSigmaForModular float32
	extensions         *bundle.Extensions
	epfWeightCustom    bool
}

func NewRestorationFilter() *RestorationFilter {
	rf := &RestorationFilter{}
	rf.epfSharpLut = []float32{0, 1 / 7, 2 / 7, 3 / 7, 4 / 7, 5 / 7, 6 / 7, 1}
	rf.epfChannelScale = []float32{40.0, 5.0, 3.5}
	rf.gab1Weights = []float32{0.115169525, 0.115169525, 0.115169525}
	rf.gab2Weights = []float32{0.061248592, 0.061248592, 0.061248592}
	rf.gab = true
	rf.customGab = false
	rf.epfIterations = 2
	rf.epfSharpCustom = false
	rf.epfWeightCustom = false
	rf.epfSigmaCustom = false
	rf.epfQuantMul = 0.46
	rf.epfPass0SigmaScale = 0.9
	rf.epfPass2SigmaScale = 6.5
	rf.epfBorderSadMul = 2.0 / 3.0
	rf.epfSigmaForModular = 1.0
	rf.extensions = bundle.NewExtensions()
	for i := 0; i < 8; i++ {
		rf.epfSharpLut[i] *= rf.epfQuantMul
	}

	return rf
}

func NewRestorationFilterWithReader(reader *jxlio.Bitreader, encoding uint32) (*RestorationFilter, error) {
	rf := &RestorationFilter{}
	rf.epfSharpLut = []float32{0, 1 / 7, 2 / 7, 3 / 7, 4 / 7, 5 / 7, 6 / 7, 1}
	rf.epfChannelScale = []float32{40.0, 5.0, 3.5}
	allDefault := reader.MustReadBool()
	if allDefault {
		rf.gab = true
	} else {
		rf.gab = reader.MustReadBool()
	}

	if !allDefault && rf.gab {
		rf.customGab = reader.MustReadBool()
	} else {
		rf.customGab = false
	}

	if rf.customGab {
		for i := 0; i < 3; i++ {
			rf.gab1Weights[i] = reader.MustReadF16()
			rf.gab2Weights[i] = reader.MustReadF16()
		}
	}

	if allDefault {
		rf.epfIterations = 2
	} else {
		rf.epfIterations = uint32(reader.MustReadBits(2))
	}

	if !allDefault && rf.epfIterations > 0 && encoding == VARDCT {
		rf.epfSharpCustom = reader.MustReadBool()
	} else {
		rf.epfSharpCustom = false
	}
	if rf.epfSharpCustom {
		for i := 0; i < len(rf.epfSharpLut); i++ {
			rf.epfSharpLut[i] = reader.MustReadF16()
		}

	}

	if !allDefault && rf.epfIterations > 9 {
		rf.epfWeightCustom = reader.MustReadBool()
	} else {
		rf.epfWeightCustom = false
	}
	if rf.epfWeightCustom {
		for i := 0; i < len(rf.epfChannelScale); i++ {
			rf.epfChannelScale[i] = reader.MustReadF16()
		}
		reader.MustReadBits(32) // ??? what do we do with this data?
	}

	if !allDefault && rf.epfIterations > 0 {
		rf.epfSigmaCustom = reader.MustReadBool()
	} else {
		rf.epfSigmaCustom = false
	}

	var err error
	if rf.epfSigmaCustom && encoding == VARDCT {
		rf.epfQuantMul, err = reader.ReadF16()
		if err != nil {
			return nil, err
		}
	} else {
		rf.epfQuantMul = 0.46
	}
	if rf.epfSigmaCustom {
		rf.epfPass0SigmaScale = reader.MustReadF16()
		rf.epfPass2SigmaScale = reader.MustReadF16()
		rf.epfBorderSadMul = reader.MustReadF16()
	} else {
		rf.epfPass0SigmaScale = 0.9
		rf.epfPass2SigmaScale = 6.5
		rf.epfBorderSadMul = 2.0 / 3.0
	}

	if !allDefault && rf.epfIterations > 0 && encoding == MODULAR {
		rf.epfSigmaForModular = reader.MustReadF16()
	} else {
		rf.epfSigmaForModular = 1.0
	}

	if allDefault {
		rf.extensions = bundle.NewExtensions()
	} else {
		rf.extensions, err = bundle.NewExtensionsWithReader(reader)
		if err != nil {
			return nil, err
		}
	}

	for i := 0; i < 8; i++ {
		rf.epfSharpLut[i] *= rf.epfQuantMul
	}

	return rf, nil
}
