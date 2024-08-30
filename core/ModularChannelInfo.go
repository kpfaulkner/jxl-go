package core

import (
	"github.com/kpfaulkner/jxl-go/util"
)

type ModularChannelInfo struct {
	width   int
	height  int
	hshift  int32
	vshift  int32
	origin  util.IntPoint
	forceWP bool
}

func NewModularChannelInfo(width int, height int, hshift int32, vshift int32) *ModularChannelInfo {
	return NewModularChannelInfoWithAllParams(width, height, hshift, vshift, util.IntPoint{0, 0}, false)
}

func NewModularChannelInfoFromInfo(info ModularChannelInfo) *ModularChannelInfo {
	return NewModularChannelInfoWithAllParams(info.width, info.height, info.hshift, info.vshift, info.origin, info.forceWP)
}

func NewModularChannelInfoWithAllParams(width int, height int, hshift int32, vshift int32, origin util.IntPoint, forceWP bool) *ModularChannelInfo {
	mc := ModularChannelInfo{
		width:   width,
		height:  height,
		hshift:  hshift,
		vshift:  vshift,
		origin:  origin,
		forceWP: forceWP,
	}
	return &mc
}
