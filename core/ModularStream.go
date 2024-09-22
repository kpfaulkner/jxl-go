package core

import (
	"errors"
	"fmt"

	"github.com/kpfaulkner/jxl-go/entropy"
	"github.com/kpfaulkner/jxl-go/jxlio"
)

const (
	RCT     = 0
	PALETTE = 1
	SQUEEZE = 2
)

var (
	permutationLUT = [][]int{
		{0, 1, 2}, {1, 2, 0}, {2, 0, 1},
		{0, 2, 1}, {1, 0, 2}, {2, 1, 0}}
)

type SqueezeParam struct {
	horizontal bool
	inPlace    bool
	beginC     int
	numC       int
}

func NewSqueezeParam(reader *jxlio.Bitreader) SqueezeParam {
	sp := SqueezeParam{}
	sp.horizontal = reader.MustReadBool()
	sp.inPlace = reader.MustReadBool()
	sp.beginC = int(reader.MustReadU32(3, 6, 10, 13, 0, 8, 72, 1096))
	sp.numC = int(reader.MustReadU32(0, 0, 0, 4, 1, 2, 3, 4))
	return sp
}

type TransformInfo struct {
	tr        int
	beginC    int
	rctType   int
	numC      int
	nbColours int
	nbDeltas  int
	dPred     int
	sp        []SqueezeParam
}

func NewTransformInfo(reader *jxlio.Bitreader) TransformInfo {

	ti := TransformInfo{}

	tr := reader.MustReadBits(2)
	if tr != SQUEEZE {
		ti.beginC = int(reader.MustReadU32(0, 3, 8, 6, 72, 10, 1096, 13))
	} else {
		ti.beginC = 0
	}

	if tr == RCT {
		ti.rctType = int(reader.MustReadU32(6, 0, 0, 2, 2, 4, 10, 6))
	} else {
		ti.rctType = 0
	}

	if tr == PALETTE {
		ti.numC = int(reader.MustReadU32(1, 0, 3, 0, 4, 0, 1, 13))
		ti.nbColours = int(reader.MustReadU32(0, 8, 256, 10, 1280, 12, 5376, 16))
		ti.nbDeltas = int(reader.MustReadU32(0, 0, 1, 8, 257, 10, 1281, 16))
		ti.dPred = int(reader.MustReadBits(4))
	} else {
		ti.numC = 0
		ti.nbColours = 0
		ti.nbDeltas = 0
		ti.dPred = 0
	}

	if tr == SQUEEZE {
		numSq := reader.MustReadU32(0, 0, 1, 4, 9, 6, 41, 8)
		ti.sp = make([]SqueezeParam, numSq)
		for i := 0; i < int(numSq); i++ {
			ti.sp[i] = NewSqueezeParam(reader)
		}
	} else {
		ti.sp = nil
	}

	ti.tr = int(tr)
	return ti
}

type ModularStream struct {
	frame        *Frame
	streamIndex  int
	channelCount int
	ecStart      int

	// HACK HACK HACK... utterly hate this. Using it to convert between ModularChannelInfo and ModularChannel
	// FIXME(kpfaulkner) refactor this to be cleaner.
	channels []*ModularChannel

	// This feels utterly dirty. But ModularChannelInfo is just a few primatives
	// and doesn't really need an interface. I will probably change my mind on this...
	//channels       []any
	tree           *MATree
	wpParams       *WPParams
	transforms     []TransformInfo
	distMultiplier int
	nbMetaChannels int
	stream         *entropy.EntropyStream
	transformed    bool
	squeezeMap     map[int][]SqueezeParam
}

func NewModularStreamWithStreamIndex(reader *jxlio.Bitreader, frame *Frame, streamIndex int, channelArray []ModularChannel) (*ModularStream, error) {
	return NewModularStreamWithChannels(reader, frame, streamIndex, len(channelArray), 0, channelArray)
}

func NewModularStreamWithReader(reader *jxlio.Bitreader, frame *Frame, streamIndex int, channelCount int, ecStart int) (*ModularStream, error) {
	return NewModularStreamWithChannels(reader, frame, streamIndex, channelCount, ecStart, nil)
}

// ModularStream.java line 63... TODO(kpfaulkner) continue
func NewModularStreamWithChannels(reader *jxlio.Bitreader, frame *Frame, streamIndex int, channelCount int, ecStart int, channelArray []ModularChannel) (*ModularStream, error) {
	ms := &ModularStream{}
	ms.streamIndex = streamIndex
	ms.frame = frame
	ms.squeezeMap = make(map[int][]SqueezeParam)

	if channelCount == 0 {
		ms.tree = nil
		ms.wpParams = nil
		ms.transforms = []TransformInfo{}
		ms.distMultiplier = 1
		return ms, nil
	}

	useGlobalTree := reader.MustReadBool()
	ms.wpParams = NewWPParams(reader)
	nbTransforms, err := reader.ReadU32(0, 0, 1, 0, 2, 4, 18, 8)
	if err != nil {
		return nil, err
	}

	ms.transforms = make([]TransformInfo, nbTransforms)
	for i := 0; i < int(nbTransforms); i++ {
		ms.transforms[i] = NewTransformInfo(reader)
	}

	w := int32(frame.header.width)
	h := int32(frame.header.height)

	if channelArray == nil || len(channelArray) == 0 {
		for i := 0; i < channelCount; i++ {
			var dimShift int32
			if i < ecStart {
				dimShift = 0
			} else {
				dimShift = frame.globalMetadata.extraChannelInfo[i-ecStart].dimShift
			}
			ms.channels = append(ms.channels, NewModularChannelWithAllParams(w, h, dimShift, dimShift, false))
		}
	} else {
		//ms.channels = append(ms.channels, channelArray...)
		for _, c := range channelArray {
			ms.channels = append(ms.channels, &c)
		}
	}

	for i := 0; i < int(nbTransforms); i++ {

		if ms.transforms[i].tr == PALETTE {

			if ms.transforms[i].beginC < ms.nbMetaChannels {
				ms.nbMetaChannels += 2 - ms.transforms[i].numC
			} else {
				ms.nbMetaChannels++
			}
			start := ms.transforms[i].beginC + 1
			for j := start; j < ms.transforms[i].beginC+ms.transforms[i].numC; j++ {
				ms.channels = append(ms.channels[:start], ms.channels[start+1:]...)
			}
			if ms.transforms[i].nbDeltas > 0 && ms.transforms[i].dPred == 6 {
				mc := ms.channels[ms.transforms[i].beginC]
				mc.forceWP = true
				ms.channels[ms.transforms[i].beginC] = mc
			}
			mc := NewModularChannelWithAllParams(int32(ms.transforms[i].nbColours), int32(ms.transforms[i].numC), -1, -1, false)
			ms.channels = append([]*ModularChannel{mc}, ms.channels...)

		} else if ms.transforms[i].tr == SQUEEZE {
			panic("TODO implement squeeze transform")
			//squeezeList := []SqueezeParam{}
			//if len(ms.transforms[i].sp) == 0 {
			//	first := ms.nbMetaChannels
			//	count := len(ms.channels) - first
			//	w = ms.channels[first].getWidth()
			//	h = ms.channels[first].getHeight()
			//	if count > 2 && ms.channels[first+1].getWidth() == w && ms.channels[first+1].getHeight() == h {
			//		squeezeList = append(squeezeList, SqueezeParam{horizontal: true, inPlace: false, beginC: first + 1, numC: 2})
			//		squeezeList = append(squeezeList, SqueezeParam{horizontal: false, inPlace: false, beginC: first + 1, numC: 2})
			//	}
			//	if h >= w && h > 8 {
			//		squeezeList = append(squeezeList, SqueezeParam{horizontal: false, inPlace: true, beginC: first, numC: count})
			//		h = (h + 1) / 2
			//	}
			//
			//	for w > 8 || h > 8 {
			//		if w > 8 {
			//			squeezeList = append(squeezeList, SqueezeParam{horizontal: true, inPlace: true, beginC: first, numC: count})
			//			w = (w + 1) / 2
			//		}
			//		if h > 8 {
			//			squeezeList = append(squeezeList, SqueezeParam{horizontal: false, inPlace: true, beginC: first, numC: count})
			//			h = (h + 1) / 2
			//		}
			//	}

		} else if ms.transforms[i].tr == RCT {
			//squeezeList = append(squeezeList, ms.transforms[i].sp...)
			continue
		} else {
			return nil, fmt.Errorf("illegal transform type %d", ms.transforms[i].tr)
		}
	}
	if !useGlobalTree {
		tree, err := NewMATreeWithReader(reader)
		if err != nil {
			return nil, err
		}
		ms.tree = tree
	} else {
		ms.tree = frame.globalTree
	}

	ms.stream = entropy.NewEntropyStreamWithStream(ms.tree.stream)

	// get max width from all channels.
	maxWidth := int32(0)
	for _, c := range ms.channels {
		if c.width > maxWidth {
			maxWidth = c.width
		}
	}
	ms.distMultiplier = int(maxWidth)
	return ms, nil
}

func (ms *ModularStream) decodeChannels(reader *jxlio.Bitreader, partial bool) error {

	groupDim := int32(ms.frame.header.groupDim)
	for i := 0; i < len(ms.channels); i++ {
		fmt.Printf("decodeChannels bitread %d\n", reader.BitsRead())
		channel := ms.channels[i]
		if partial && i >= ms.nbMetaChannels &&
			(channel.width > groupDim || channel.height > groupDim) {
			break
		}
		err := channel.decode(reader, ms.stream, ms.wpParams, ms.tree, ms, int32(i), int32(ms.streamIndex), ms.distMultiplier)
		if err != nil {
			return err
		}
	}

	if ms.stream != nil && !ms.stream.ValidateFinalState() {
		return errors.New("illegal final modular state")
	}
	if !partial {
		err := ms.applyTransforms()
		if err != nil {
			return err
		}
	}

	return nil
}

func (ms *ModularStream) applyTransforms() error {

	// HERE  20240921
	if ms.transformed {
		return nil
	}
	ms.transformed = true
	var err error
	for i := len(ms.transforms) - 1; i >= 0; i-- {
		if ms.transforms[i].tr == SQUEEZE {
			spa := ms.squeezeMap[i]
			for j := len(spa) - 1; j >= 0; j-- {
				sp := spa[j]
				begin := sp.beginC
				end := begin + sp.numC - 1
				var offset int
				if sp.inPlace {
					offset = end + 1
				} else {
					offset = len(ms.channels) + begin - end - 1
				}
				for c := begin; c <= end; c++ {
					r := offset + c - begin
					ch := ms.channels[c]
					residu := ms.channels[r]
					var output *ModularChannel
					if sp.horizontal {
						outputInfo := NewModularChannelWithAllParams(ch.width+residu.width, ch.height, ch.hshift-1, ch.vshift, false)
						output, err = inverseHorizontalSqueeze(outputInfo, ch, residu)
						if err != nil {
							return err
						}
					} else {

						outputInfo := NewModularChannelWithAllParams(ch.width, ch.height+residu.height, ch.hshift, ch.vshift-1, false)
						output, err = inverseHorizontalSqueeze(outputInfo, ch, residu)
						if err != nil {
							return err
						}
					}
					ms.channels[c] = output
				}
				for c := 0; c < end-begin+1; c++ {
					ms.channels = append(ms.channels[:offset], ms.channels[offset+1:]...)
				}
			}
		} else if ms.transforms[i].tr == RCT {

			// HERE... need to implement
			permutation := ms.transforms[i].rctType / 7
			transType := ms.transforms[i].rctType % 7
			v := [3]*ModularChannel{}
			start := ms.transforms[i].beginC
			var err error
			for j := 0; j < 3; j++ {
				v[j] = ms.channels[start+j]
			}
			var rct func(int32, int32) error
			switch transType {
			case 0:
				rct = func(x int32, y int32) error {
					return nil
				}
				break

			case 1:
				rct = func(x int32, y int32) error {
					v[2].buffer[y][x] += v[0].buffer[y][x]
					return nil
				}
			case 2:
				rct = func(x int32, y int32) error {
					v[1].buffer[y][x] += v[0].buffer[y][x]
					return nil
				}
				break
			case 3:
				rct = func(x int32, y int32) error {
					a := v[0].buffer[y][x]
					v[2].buffer[y][x] += a
					v[1].buffer[y][x] += a
					return nil
				}
				break
			case 4:
				rct = func(x int32, y int32) error {
					v[1].buffer[y][x] += (v[0].buffer[y][x] + v[2].buffer[y][x]) >> 1
					return nil
				}
				break
			case 5:
				rct = func(x int32, y int32) error {
					a := v[0].buffer[y][x]
					ac := a + v[2].buffer[y][x]
					v[1].buffer[y][x] += (a + ac) >> 1
					v[2].buffer[y][x] = ac
					return nil
				}
				break
			case 6:
				rct = func(x int32, y int32) error {
					b := v[1].buffer[y][x]
					c := v[2].buffer[y][x]
					tmp := v[0].buffer[y][x] - (c >> 1)
					f := tmp - (b >> 1)
					v[0].buffer[y][x] = f + b
					v[1].buffer[y][x] = c + tmp
					v[2].buffer[y][x] = f
					return nil
				}
				break
			default:
				return errors.New("illegal RCT type")
			}

			for y := int32(0); y < v[0].height; y++ {
				for x := int32(0); x < v[0].width; x++ {
					err = rct(int32(x), int32(y))
					if err != nil {
						return err
					}
				}
			}

			for j := 0; j < 3; j++ {
				ms.channels[start+permutationLUT[permutation][j]] = v[j]
			}
		} else if ms.transforms[i].tr == PALETTE {
			panic("ModularStream::applyTransforms PALETTE not implemented")
		}
	}
	return nil
}

func inverseHorizontalSqueeze(channel *ModularChannel, orig *ModularChannel, res *ModularChannel) (*ModularChannel, error) {

	if channel.height != orig.height+res.height ||
		(orig.height != res.height && orig.height != 1+res.height) ||
		channel.width != orig.width || res.width != orig.width {
		return nil, errors.New("Corrupted squeeze transform")
	}

	channel.allocate()

	for y := uint32(0); y < channel.size.height; y++ {
		for x := uint32(0); x < channel.size.width; x++ {
			avg := orig.buffer[y][x]
			residu := res.buffer[y][x]
			var nextAvg int32
			if x+1 < uint32(orig.width) {
				nextAvg = orig.buffer[y][x+1]
			} else {
				nextAvg = avg
			}
			var left int32
			if x > 0 {
				left = orig.buffer[y][2*x-1]
			} else {
				nextAvg = avg
			}
			diff := residu + tendancy(left, avg, nextAvg)
			first := avg + diff/2
			channel.buffer[y][2*x] = first
			channel.buffer[y][2*x+1] = first - diff
		}
	}
	if orig.size.width > res.size.width {
		xs := 2 * res.size.width
		for y := uint32(0); y < channel.size.height; y++ {
			channel.buffer[y][xs] = orig.buffer[y][res.size.width]
		}
	}

	return channel, nil
}

func tendancy(a int32, b int32, c int32) int32 {
	if a >= b && b >= c {
		x := (4*a - 3*c - b + 6) / 12
		d := 2 * (a - b)
		e := 2 * (b - c)
		if (x - (x & 1)) > d {
			x = d + 1
		}
		if (x + (x & 1)) > e {
			x = e
		}
		return x
	}

	if a <= b && b <= c {
		x := (4*a - 3*c - b - 6) / 12
		d := 2 * (a - b)
		e := 2 * (b - c)
		if (x + (x & 1)) < d {
			x = d - 1
		}
		if (x - (x & 1)) < e {
			x = e
		}
		return x
	}

	return 0
}

func (ms *ModularStream) getDecodedBuffer() [][][]int32 {
	bands := make([][][]int32, len(ms.channels))
	for i := 0; i < len(bands); i++ {
		bands[i] = ms.channels[i].buffer
	}
	return bands
}
