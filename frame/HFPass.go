package frame

import (
	"errors"
	"fmt"
	"slices"

	"github.com/kpfaulkner/jxl-go/entropy"
	"github.com/kpfaulkner/jxl-go/jxlio"
	"github.com/kpfaulkner/jxl-go/util"
)

type HFPass struct {
	order         [][][]util.Point
	naturalOrder  [][]util.Point
	contextStream *entropy.EntropyStream
}

func NewHFPassWithReader(reader *jxlio.Bitreader, frame *Frame, passIndex uint32) (*HFPass, error) {
	hfp := &HFPass{}
	hfp.naturalOrder = util.MakeMatrix2D[util.Point](13, 0)
	hfp.order = util.MakeMatrix3D[util.Point](13, 3, 0)
	usedOrders, err := reader.ReadU32(05, 0, 19, 0, 0, 0, 0, 13)
	if err != nil {
		return nil, err
	}
	var stream *entropy.EntropyStream
	if usedOrders != 0 {
		if stream, err = entropy.NewEntropyStreamWithReaderAndNumDists(reader, 8); err != nil {
			return nil, err
		}
	} else {
		stream = nil
	}

	for b := int32(0); b < 13; b++ {
		naturalOrder, err := hfp.getNaturalOrder(b)
		if err != nil {
			return nil, err
		}
		l := len(naturalOrder)
		if b == 3 {
			fmt.Printf("snoop\n")
		}

		for c := 0; c < 3; c++ {
			if usedOrders&(1<<uint32(b)) != 0 {
				hfp.order[b][c] = make([]util.Point, l)
				perm, err := readPermutation(reader, stream, uint32(l), uint32(l/64))
				if err != nil {
					return nil, err
				}
				for i := 0; i < len(hfp.order[b][c]); i++ {
					hfp.order[b][c][i] = naturalOrder[perm[i]]
				}
			} else {
				hfp.order[b][c] = naturalOrder
			}
		}
	}
	if stream != nil && !stream.ValidateFinalState() {
		return nil, errors.New("ANS state decoding error")
	}
	numContexts := 495 * frame.hfGlobal.numHFPresets * frame.LfGlobal.hfBlockCtx.numClusters
	contextStream, err := entropy.NewEntropyStreamWithReaderAndNumDists(reader, int(numContexts))
	if err != nil {
		return nil, err
	}

	hfp.contextStream = contextStream
	return hfp, nil
}

func (hfp *HFPass) getNaturalOrder(i int32) ([]util.Point, error) {
	if len(hfp.naturalOrder[i]) != 0 {
		return hfp.naturalOrder[i], nil
	}

	var tt *TransformType
	var err error
	if tt, err = getByOrderID(i); err != nil {
		return nil, err
	}

	l := tt.blockWidth * tt.blockHeight
	hfp.naturalOrder[i] = make([]util.Point, l)
	for y := int32(0); y < tt.blockHeight; y++ {
		for x := int32(0); x < tt.blockWidth; x++ {
			hfp.naturalOrder[i][y*tt.blockWidth+x] = util.Point{X: x, Y: y}
		}
	}

	sorterFunc, err := getNaturalOrderFunc(i)
	if err != nil {
		return nil, err
	}
	slices.SortFunc(hfp.naturalOrder[i], sorterFunc)
	return hfp.naturalOrder[i], nil

}

func getNaturalOrderFunc(i int32) (func(a util.Point, b util.Point) int, error) {

	tt, err := getByOrderID(i)
	if err != nil {
		return nil, err
	}
	return func(a util.Point, b util.Point) int {
		maxDim := util.Max(tt.dctSelectHeight, tt.dctSelectWidth)
		aLLF := a.Y < tt.dctSelectHeight && a.X < tt.dctSelectWidth
		bLLF := b.Y < tt.dctSelectHeight && b.X < tt.dctSelectWidth
		if aLLF && !bLLF {
			return -1
		}
		if !aLLF && bLLF {
			return 1
		}
		if aLLF && bLLF {
			if b.Y != a.Y {
				return int(a.Y - b.Y)
			}
			return int(a.X - b.X)
		}

		aSY := a.Y * maxDim / tt.dctSelectHeight
		aSX := a.X * maxDim / tt.dctSelectWidth
		bSY := b.Y * maxDim / tt.dctSelectHeight
		bSX := b.Y * maxDim / tt.dctSelectWidth
		aKey1 := aSY + aSX
		bKey1 := bSY + bSX
		if aKey1 != bKey1 {
			return int(aKey1 - bKey1)
		}

		aKey2 := aSX - aSY
		bKey2 := bSX - bSY
		if (aKey1 & 1) == 1 {
			aKey2 = -aKey2
		}
		if (bKey1 & 1) == 1 {
			bKey2 = -bKey2
		}
		return int(aKey2 - bKey2)
	}, nil
}
