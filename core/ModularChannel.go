package core

import (
	"errors"
	"fmt"
	"math"

	"github.com/kpfaulkner/jxl-go/entropy"
	"github.com/kpfaulkner/jxl-go/jxlio"
	"github.com/kpfaulkner/jxl-go/util"
)

var (
	oneL24OverKP1 = make([]int64, 64)
)

type ModularChannel struct {
	ModularChannelInfo
	buffer  [][]int32
	decoded bool
	err     [][][]int32
	pred    [][]int32
	subpred []int32
	weight  []int32
}

func init() {
	for i := int64(0); i < int64(len(oneL24OverKP1)); i++ {
		oneL24OverKP1[i] = (1 << 24) / (i + 1)
	}
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

func (mc *ModularChannel) prediction(x int32, y int32, k int32) (int32, error) {
	var n, v, nw, w int32
	switch k {
	case 0:
		return 0, nil
	case 1:
		return mc.west(x, y), nil
	case 2:
		return mc.north(x, y), nil
	case 3:
		return (mc.west(x, y) + mc.north(x, y)) / 2, nil
	case 4:
		w = mc.west(x, y)
		n = mc.north(x, y)
		nw = mc.northWest(x, y)
		if util.Abs(n-nw) < util.Abs(w-nw) {
			return w, nil
		}
		return n, nil
	case 5:
		w = mc.west(x, y)
		n = mc.north(x, y)
		v = w + n - mc.northWest(x, y)
		return util.Clamp3(v, n, w), nil
	case 6:
		return (mc.pred[y][x] + 3) >> 3, nil
	case 7:
		return mc.northEast(x, y), nil
	case 8:
		return mc.northWest(x, y), nil
	case 9:
		return mc.westWest(x, y), nil
	case 10:
		return (mc.west(x, y) + mc.northWest(x, y)) / 2, nil
	case 11:
		return (mc.north(x, y) + mc.northWest(x, y)) / 2, nil
	case 12:
		return (mc.north(x, y) + mc.northEast(x, y)) / 2, nil
	case 13:
		return (6*mc.north(x, y) - 2*mc.northNorth(x, y) + 7*mc.west(x, y) +
			mc.westWest(x, y) + mc.northEastEast(x, y) + 3*mc.northEast(x, y) + 8) / 16, nil
	default:
		return 0, errors.New("Illegal predictor state")
	}
}

func (mc *ModularChannel) prePredictWP(wpParams *WPParams, x int32, y int32) (int32, error) {

	n3 := mc.north(x, y) << 3
	nw3 := mc.northWest(x, y) << 3
	ne3 := mc.northEastEast(x, y) << 3
	w3 := mc.west(x, y) << 3
	nn3 := mc.northNorth(x, y) << 3
	tN := mc.errorNorth(x, y, 4)
	tW := mc.errorWest(x, y, 4)
	tNE := mc.errorNorthEast(x, y, 4)
	tNW := mc.errorNorthWest(x, y, 4)
	mc.subpred[0] = w3 + ne3 - n3
	mc.subpred[1] = n3 - ((tW+tN+tNE)*(int32(wpParams.param1)))>>5
	mc.subpred[2] = w3 - ((tW+tN+tNW)*(int32(wpParams.param2)))>>5
	mc.subpred[3] = n3 - ((tNW*wpParams.param3a +
		tN*wpParams.param3b +
		tNE*wpParams.param3c +
		(nn3-n3)*wpParams.param3d +
		(nw3-w3)*wpParams.param3e) >> 5)

	wSum := int32(0)
	for e := int32(0); e < 4; e++ {
		//eSum := mc.errorNorth(x, y, e) + mc.errorWest(x, y, e) + mc.errorNorthWest(x, y, e) +
		//	mc.errorWestWest(x, y, e) + mc.errorNorthEast(x, y, e)
		en := mc.errorNorth(x, y, e)
		ew := mc.errorWest(x, y, e)
		enw := mc.errorNorthWest(x, y, e)
		eww := mc.errorWestWest(x, y, e)
		ene := mc.errorNorthEast(x, y, e)
		eSum := en + ew + enw + eww + ene
		if x+1 == int32(mc.width) {
			eSum += mc.errorWest(x, y, e)
		}
		shift := util.FloorLog1p(int64(eSum)) - 5
		if shift < 0 {
			shift = 0
		}
		mc.weight[e] = int32(4 + ((wpParams.weight[e] * oneL24OverKP1[eSum>>shift]) >> shift))
		wSum += mc.weight[e]
	}

	logWeight := util.FloorLog1p(int64(wSum)-1) - 4
	wSum = 0
	for e := 0; e < 4; e++ {
		mc.weight[e] = mc.weight[e] >> logWeight
		wSum += mc.weight[e]
	}
	s := int64((wSum >> 1) - 1)
	for e := 0; e < 4; e++ {
		s += int64(mc.subpred[e]) * int64(mc.weight[e])
	}
	mc.pred[y][x] = int32((s * oneL24OverKP1[wSum-1]) >> 24)
	if (tN^tW)|(tN^tNW) <= 0 {
		mc.pred[y][x] = util.Clamp(mc.pred[y][x], int32(w3), int32(n3), int32(ne3))
	}

	maxError := tW
	if util.Abs(tN) > util.Abs(maxError) {
		maxError = tN
	}
	if util.Abs(tNW) > util.Abs(maxError) {
		maxError = tNW
	}
	if util.Abs(tNE) > util.Abs(maxError) {
		maxError = tNE
	}

	return maxError, nil
}

// Could try and use IfThenElse but that gets messy quickly. Prefer some simple 'if' statements.
func (mc *ModularChannel) west(x int32, y int32) int32 {
	if x > 0 {
		return mc.buffer[y][x-1]
	}
	if y > 0 {
		return mc.buffer[y-1][x]
	}
	return 0
}

func (mc *ModularChannel) north(x int32, y int32) int32 {
	if y > 0 {
		return mc.buffer[y-1][x]
	}
	if x > 0 {
		return mc.buffer[y][x-1]
	}
	return 0
}

func (mc *ModularChannel) northWest(x int32, y int32) int32 {
	if x > 0 && y > 0 {
		return mc.buffer[y-1][x-1]
	}
	return mc.west(x, y)
}

func (mc *ModularChannel) northEast(x int32, y int32) int32 {
	if x+1 < int32(mc.width) && y > 0 {
		return mc.buffer[y-1][x+1]
	}
	return mc.north(x, y)
}

func (mc *ModularChannel) northNorth(x int32, y int32) int32 {
	if y > 1 {
		return mc.buffer[y-2][x]
	}
	return mc.north(x, y)
}

func (mc *ModularChannel) northEastEast(x int32, y int32) int32 {
	if x+2 < int32(mc.width) && y > 0 {
		return mc.buffer[y-1][x+2]
	}
	return mc.northEast(x, y)
}

func (mc *ModularChannel) westWest(x int32, y int32) int32 {
	if x > 1 {
		return mc.buffer[y][x-2]
	}
	return mc.west(x, y)
}

func (mc *ModularChannel) errorNorth(x int32, y int32, e int32) int32 {
	if y > 0 {
		return mc.err[e][y-1][x]
	}
	return 0
}
func (mc *ModularChannel) errorWest(x int32, y int32, e int32) int32 {
	if x > 0 {
		return mc.err[e][y][x-1]
	}
	return 0
}
func (mc *ModularChannel) errorWestWest(x int32, y int32, e int32) int32 {
	if x > 1 {
		return mc.err[e][y][x-2]
	}
	return 0
}

func (mc *ModularChannel) errorNorthWest(x int32, y int32, e int32) int32 {
	if x > 0 && y > 0 {
		return mc.err[e][y-1][x-1]
	}
	return mc.errorNorth(x, y, e)
}

func (mc *ModularChannel) errorNorthEast(x int32, y int32, e int32) int32 {
	if x+1 < int32(mc.width) && y > 0 {
		return mc.err[e][y-1][x+1]
	}

	return mc.errorNorth(x, y, e)
}

func (mc *ModularChannel) walkerFunc(k int32) int32 {
	return 0
}

func (mc *ModularChannel) decode(reader *jxlio.Bitreader, stream *entropy.EntropyStream,
	wpParams *WPParams, tree *MATree, parent *ModularStream, channelIndex int32, streamIndex int32, distMultiplier int) error {

	fmt.Printf("decode start : bits read %d\n", reader.BitsRead())
	if mc.decoded {
		return nil
	}

	tree = tree.compactify(channelIndex, streamIndex)
	useWP := mc.forceWP || tree.useWeightedPredictor()
	if useWP {
		mc.err = util.MakeMatrix3D[int32](5, mc.height, mc.width)
		mc.pred = util.MakeMatrix2D[int32](mc.height, mc.width)
		mc.subpred = make([]int32, 4)
		mc.weight = make([]int32, 4)
	}

	////////////////////////////////////////////////////////////////
	// WARNING:
	// below when y0=0, the values being read are incorrect. This is causing
	// Java goes from bits read from 15x476 entries to 4x493
	// where as this cgoes does 5x476 then 5x492... and gets worse from there.
	// FIXME(kpfaulkner)
	///////////////////////////////////////////////////////////
	var err error
	for y0 := 0; y0 < mc.height; y0++ {
		fmt.Printf("decode start : y : %d bits read %d\n", y0, reader.BitsRead())
		if y0 == 0 {
			fmt.Printf("snoop\n")
		}
		y := int32(y0)
		refinedTree := tree.compactifyWithY(channelIndex, streamIndex, int32(y))
		for x0 := 0; x0 < mc.width; x0++ {
			if y0 == 0 && x0 == 1 {

				fmt.Printf("snoop\n")
			}
			if y0 == 0 && x0 == 2 {
				// weight goes wrong.
				fmt.Printf("snoop\n")
			}
			if y0 == 0 && x0 == 20 {
				fmt.Printf("snoop\n")
			}
			if y0 == 0 && x0 == 103 && reader.BitsRead() == 476 {
				fmt.Printf("snoop\n")
			}
			x := int32(x0)
			var maxError int32
			if useWP {
				maxError, err = mc.prePredictWP(wpParams, int32(x), int32(y))
				if err != nil {
					return err
				}
			} else {
				maxError = 0
			}

			leafNode, err := refinedTree.walk(func(k int32) (int32, error) {
				switch k {
				case 0:
					return channelIndex, nil
				case 1:
					return streamIndex, nil
				case 2:
					return y, nil
				case 3:
					return x, nil
				case 4:
					return util.Abs(mc.north(x, y)), nil
				case 5:
					return util.Abs(mc.west(x, y)), nil
				case 6:
					return mc.north(x, y), nil
				case 7:
					return mc.west(x, y), nil
				case 8:
					if x > 0 {
						return mc.west(x, y) - (mc.west(x-1, y) + mc.north(x-1, y) - mc.northWest(x-1, y)), nil
					}
					return mc.west(x, y), nil
				case 9:
					return mc.west(x, y) + mc.north(x, y) - mc.northWest(x, y), nil
				case 10:
					return mc.west(x, y) - mc.northWest(x, y), nil
				case 11:
					return mc.northWest(x, y) - mc.north(x, y), nil
				case 12:
					return mc.north(x, y) - mc.northEast(x, y), nil
				case 13:
					return mc.north(x, y) - mc.northNorth(x, y), nil
				case 14:
					return mc.west(x, y) - mc.westWest(x, y), nil
				case 15:
					return maxError, nil
				default:
					if k-16 >= 4*channelIndex {
						return 0, nil
					}
					k2 := int32(16)
					for j := channelIndex - 1; j >= 0; j-- {
						channel, ok := parent.channels[j].(*ModularChannel)
						if !ok {
							return 0, errors.New("channel not a ModularChannel")
						}
						if channel.width != mc.width || channel.height != mc.height ||
							channel.hshift != mc.hshift || channel.vshift != mc.vshift {
							continue
						}
						if k2+4 <= k {
							k2 += 4
							continue
						}
						rC := channel.buffer[y][x]
						if k2 == k {

							k2++
							return util.Abs(rC), nil
						}
						k2++
						if k2 == k {
							k2++
							return rC, nil
						}
						k2++
						var rW int32
						var rN int32
						var rNW int32
						var rG int32
						if x > 0 {
							rW = channel.buffer[y][x-1]
						} else {
							rW = 0
						}

						if y > 0 {
							rN = channel.buffer[y-1][x]
						} else {
							rN = rW
						}
						if x > 0 && y > 0 {
							rNW = channel.buffer[y-1][x-1]
						} else {
							rNW = rW
						}
						rG = rC - util.Clamp3(rW+rN-rNW, rN, rW)
						if k2 == k {

							k2++
							return util.Abs(rG), nil
						}
						k2++
						if k2 == k {
							k2++
							return rG, nil
						}
						k2++

					}
					return 0, nil
				}
			})
			if err != nil {
				return err
			}
			diff, err := stream.ReadSymbolWithMultiplier(reader, int(leafNode.context), distMultiplier)
			if err != nil {
				return err
			}
			fmt.Printf("decode bits read %d\n", reader.BitsRead())
			if reader.BitsRead() == 832 {
				fmt.Printf("snoop\n")
			}
			diff = jxlio.UnpackSigned(uint32(diff))*leafNode.multiplier + leafNode.offset
			p, err := mc.prediction(x, y, leafNode.predictor)
			if err != nil {
				return err
			}
			trueValue := diff + p
			mc.buffer[y][x] = trueValue
			if useWP {
				for e := 0; e < 4; e++ {
					if e == 1 && y == 0 && x == 1 {
						fmt.Printf("snoop\n")
					}
					mc.err[e][y][x] = int32(math.Abs(float64(mc.subpred[e]-(trueValue<<3)))+3) >> 3
				}
				mc.err[4][y][x] = mc.pred[y][x] - (trueValue << 3)
			}
		}
	}
	return nil
}
