package core

import "github.com/kpfaulkner/jxl-go/jxlio"

type RestorationFilter struct {
	gab         bool
	customGab   bool
	gab1Weights []float32
	gab2Weights []float32
}

func NewRestorationFilter() *RestorationFilter {
	rf := &RestorationFilter{}
	rf.gab1Weights = []float32{0.115169525, 0.115169525, 0.115169525}
	rf.gab2Weights = []float32{0.061248592, 0.061248592, 0.061248592}
	return rf
}

func NewRestorationFilterWithReader(reader *jxlio.Bitreader, encoding uint32) (*RestorationFilter, error) {
	rf := &RestorationFilter{}

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
	
	// TODO(kpfaulkner)

	return rf, nil
}
