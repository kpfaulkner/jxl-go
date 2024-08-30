package core

import "github.com/kpfaulkner/jxl-go/jxlio"

type GlobalModular struct {
	frame      *Frame
	globalTree *MATree
	//stream     *ModularStream
}

func NewGlobalModularWithReader(reader *jxlio.Bitreader, parent *Frame) (*GlobalModular, error) {
	gm := &GlobalModular{}
	gm.frame = parent

	var err error
	hasGlobalTree := reader.MustReadBool()
	if hasGlobalTree {
		gm.globalTree, err = NewMATreeWithReader(reader)
		if err != nil {
			return nil, err
		}
	} else {
		gm.globalTree = nil
	}
	return gm, nil
}
