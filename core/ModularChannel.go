package core

import (
	"math"

	"github.com/kpfaulkner/jxl-go/entropy"
	"github.com/kpfaulkner/jxl-go/jxlio"
	"github.com/kpfaulkner/jxl-go/util"
)

type ModularChannel struct {
	ModularChannelInfo
	buffer  [][]uint32
	decoded bool
	err     [][][]uint32
	pred    [][]int32
	subpred []int32
	weight  []int32
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
		mc.buffer = make([][]uint32, 0)
	} else {
		mc.buffer = util.MakeMatrix2D[uint32](height, width)
	}
	return mc
}

func (mc *ModularChannel) prePredictWP(wpParams *WPParams, x int32, y int32) (int32, error) {

	return 0, nil
}

// Could try and use IfThenElse but that gets messy quickly. Prefer some simple 'if' statements.
func (mc *ModularChannel) west(x int32, y int32) uint32 {
	if x > 0 {
		return mc.buffer[y][x-1]
	}
	if y > 0 {
		return mc.buffer[y-1][x]
	}
	return 0
}

func (mc *ModularChannel) north(x int32, y int32) uint32 {
	if y > 0 {
		return mc.buffer[y-1][x]
	}
	if x > 0 {
		return mc.buffer[y][x-1]
	}
	return 0
}

func (mc *ModularChannel) northWest(x int32, y int32) uint32 {
	if x > 0 && y > 0 {
		return mc.buffer[y-1][x-1]
	}
	return mc.west(x, y)
}

func (mc *ModularChannel) northEast(x int32, y int32) uint32 {
	if x+1 < int32(mc.width) && y > 0 {
		return mc.buffer[y-1][x+1]
	}
	return mc.north(x, y)
}

func (mc *ModularChannel) northNorth(x int32, y int32) uint32 {
	if y > 1 {
		return mc.buffer[y-2][x]
	}
	return mc.north(x, y)
}

func (mc *ModularChannel) northEastEast(x int32, y int32) uint32 {
	if x+2 < int32(mc.width) && y > 0 {
		return mc.buffer[y-1][x+2]
	}
	return mc.northEast(x, y)
}

func (mc *ModularChannel) westWest(x int32, y int32) uint32 {
	if x > 1 {
		return mc.buffer[y][x-2]
	}
	return mc.west(x, y)
}

func (mc *ModularChannel) errorNorth(x int32, y int32, e int32) uint32 {
	if y > 0 {
		return mc.err[e][y-1][x]
	}
	return 0
}

func (mc *ModularChannel) errorWestWest(x int32, y int32, e int32) uint32 {
	if x > 1 {
		return mc.err[e][y][x-2]
	}
	return 0
}

func (mc *ModularChannel) errorNorthWest(x int32, y int32, e int32) uint32 {
	if x > 0 && y > 0 {
		return mc.err[e][y-1][x-1]
	}
	return mc.errorNorth(x, y, e)
}

func (mc *ModularChannel) errorNorthEast(x int32, y int32, e int32) uint32 {
	if x+1 < int32(mc.width) && y > 0 {
		return mc.err[e][y-1][x+1]
	}

	return mc.errorNorth(x, y, e)
}

func (mc *ModularChannel) decode(reader *jxlio.Bitreader, stream *entropy.EntropyStream,
	wpParams *WPParams, tree *MATree, parent *ModularStream, channelIndex int32, streamIndex int32, distMultiplier int) error {

	if mc.decoded {
		return nil
	}

	tree = tree.compactify(channelIndex, streamIndex)
	useWP := mc.forceWP || tree.useWeightedPredictor()
	if useWP {
		mc.err = util.MakeMatrix3D[uint32](5, mc.height, mc.width)
		mc.pred = util.MakeMatrix2D[int32](mc.height, mc.width)
		mc.subpred = make([]int32, 4)
		mc.weight = make([]int32, 4)
	}

	for y0 := 0; y0 < mc.height; y0++ {
		y := y0
		refinedTree := tree.compactifyWithY(channelIndex, streamIndex, int32(y))
		for x0 := 0; x0 < mc.width; x0++ {
			x := x0
			var maxError int
			if useWP {
				maxError = mc.prePredictWP(wpParams, x, y)
			} else {
				maxError = 0
			}

			leafNode := refinedTree.walk(mc.getRefinedTreeWalker(channelIndex, streamIndex, x, y, parent, maxError))
			diff, err := stream.readSymbolWithMultiplier(reader, leafNode.context, distMultiplier)
			if err != nil {
				return err
			}
			diff = bits.UnpackSigned(diff)*leafNode.multiplier + leafNode.offset
			trueValue := diff + mc.prediction(x, y, leafNode.predictor)
			mc.set(x, y, trueValue)
			if useWP {
				for e := 0; e < 4; e++ {
					mc.err[e][y][x] = int(math.Abs(float64(mc.subpred[e]-(trueValue<<3)))+3) >> 3
				}
				mc.err[4][y][x] = mc.pred[y][x] - (trueValue << 3)
			}
		}
	}
	mc.decoded = true
	panic("decode not implemented")

	return nil
}
