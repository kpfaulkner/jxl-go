package core

import (
	"github.com/kpfaulkner/jxl-go/entropy"
	"github.com/kpfaulkner/jxl-go/jxlio"
	"github.com/kpfaulkner/jxl-go/util"
)

type ModularChannel struct {
	ModularChannelInfo
	buffer  [][]int32
	decoded bool
}

func NewModularChannelFromInfo(info ModularChannelInfo) *ModularChannel {

	return NewModularChannelWithAllParams(info.width, info.height, info.hshift, info.vshift, util.IntPoint{0, 0}, false)
}

func NewModularChannelWithAllParams(width int, height int, hshift int32, vshift int32, origin util.IntPoint, forceWP bool) *ModularChannel {
	mc := &ModularChannel{
		ModularChannelInfo: ModularChannelInfo{
			width:   width,
			height:  height,
			hshift:  hshift,
			vshift:  vshift,
			origin:  origin,
			forceWP: forceWP,
		},
	}

	if width == 0 || height == 0 {
		mc.buffer = make([][]int32, 0)
	} else {
		mc.buffer = util.MakeMatrix2D[int32](height, width)
	}
	return mc
}

func (mc *ModularChannel) decode(reader *jxlio.Bitreader, stream *entropy.EntropyStream,
	wpParams *WPParams, tree *MATree, parent *ModularStream, channelIndex int, streamIndex int, distMultiplier int) error {

	panic("decode not implemented")
	//if mc.decoded {
	//	return nil
	//}
	//
	//tree = tree.compactify(channelIndex, streamIndex)
	//useWP := mc.forceWP || tree.useWeightedPredictor()
	//if useWP {
	//	mc.err = make3DSlice(5, mc.height, mc.width)
	//	mc.pred = make2DSlice(mc.height, mc.width)
	//	mc.subpred = make([]int, 4)
	//	mc.weight = make([]int, 4)
	//}
	//
	//for y0 := 0; y0 < mc.height; y0++ {
	//	y := y0
	//	refinedTree := tree.compactifyWithY(channelIndex, streamIndex, y)
	//	for x0 := 0; x0 < mc.width; x0++ {
	//		x := x0
	//		var maxError int
	//		if useWP {
	//			maxError = mc.prePredictWP(wpParams, x, y)
	//		} else {
	//			maxError = 0
	//		}
	//
	//		leafNode := refinedTree.walk(mc.getRefinedTreeWalker(channelIndex, streamIndex, x, y, parent, maxError))
	//		diff, err := stream.readSymbolWithMultiplier(reader, leafNode.context, distMultiplier)
	//		if err != nil {
	//			return err
	//		}
	//		diff = bits.UnpackSigned(diff)*leafNode.multiplier + leafNode.offset
	//		trueValue := diff + mc.prediction(x, y, leafNode.predictor)
	//		mc.set(x, y, trueValue)
	//		if useWP {
	//			for e := 0; e < 4; e++ {
	//				mc.err[e][y][x] = int(math.Abs(float64(mc.subpred[e]-(trueValue<<3)))+3) >> 3
	//			}
	//			mc.err[4][y][x] = mc.pred[y][x] - (trueValue << 3)
	//		}
	//	}
	//}
	//mc.decoded = true
	return nil
}
