package frame

import (
	"github.com/kpfaulkner/jxl-go/bundle"
	"github.com/kpfaulkner/jxl-go/util"
)

type FakeFramer struct {
	lfGroup  *LFGroup
	hfGlobal *HFGlobal
	lfGlobal *LFGlobal
	header   *FrameHeader
	passes   []Pass
}

func (f FakeFramer) getLFGroupForGroup(groupID int32) *LFGroup {
	//TODO implement me
	panic("implement me")
}

func (f FakeFramer) getHFGlobal() *HFGlobal {
	//TODO implement me
	panic("implement me")
}

func (f FakeFramer) getLFGlobal() *LFGlobal {
	//TODO implement me
	panic("implement me")
}

func (f FakeFramer) getFrameHeader() *FrameHeader {
	//TODO implement me
	panic("implement me")
}

func (f FakeFramer) getPasses() []Pass {
	//TODO implement me
	panic("implement me")
}

func (f FakeFramer) getGroupSize(groupID int32) (util.Dimension, error) {
	//TODO implement me
	panic("implement me")
}

func (f FakeFramer) groupPosInLFGroup(lfGroupID int32, groupID uint32) util.Point {
	//TODO implement me
	panic("implement me")
}

func (f FakeFramer) getGlobalMetadata() *bundle.ImageHeader {
	//TODO implement me
	panic("implement me")
}

func NewFakeFramer() Framer {
	return &FakeFramer{}
}
