package core

import "github.com/kpfaulkner/jxl-go/jxlio"

type GlobalModular struct {
	frame *Frame
}

func NewGlobalModularWithReader(reader *jxlio.Bitreader, parent *Frame) (*GlobalModular, error) {
	gm := &GlobalModular{}
	gm.frame = parent

	hasGlobalTree := reader.MustReadBool()
	if hasGlobalTree {
		mg.globalTree = NewMATreeWithReader(reader)
	} else {

	}
	return gm, nil
}
