package frame

import (
	"github.com/kpfaulkner/jxl-go/entropy"
	"github.com/kpfaulkner/jxl-go/jxlio"
	"github.com/kpfaulkner/jxl-go/util"
)

type HFCoefficients struct {
	hfPreset        int32
	groupID         uint32
	frame           *Frame
	hfctx           *HFBlockContext
	lfg             *LFGroup
	stream          *entropy.EntropyStream
	quantizedCoeffs [][][]int32
	dequantHFCoeff  [][][]float32
	groupPos        util.Point
	blocks          []util.Point
}

func NewHFCoefficientsWithReader(reader *jxlio.Bitreader, frame *Frame, pass uint32, group uint32) (*HFCoefficients, error) {
	hf := &HFCoefficients{}

	hfPreset, err := reader.ReadBits(uint32(util.CeilLog1p(frame.hfGlobal.numHFPresets - 1)))
	if err != nil {
		return nil, err
	}
	hf.hfPreset = int32(hfPreset)
	hf.groupID = group
	hf.frame = frame
	hf.hfctx = frame.LfGlobal.hfBlockCtx
	hf.lfg = frame.getLFGroupForGroup(int32(group))
	offset := 495 * hf.hfctx.numClusters * hf.hfPreset
	header := frame.Header
	shift := header.passes.shift[pass]
	hfPass := hf.frame.passes[pass].hfPass
	size, err := hf.frame.getLFGroupSize(int32(hf.groupID))
	if err != nil {
		return nil, err
	}
	nonZeros := util.MakeMatrix3D[int32](3, 32, 32)
	hf.stream = entropy.NewEntropyStreamWithStream(hfPass.contextStream)
	hf.quantizedCoeffs = util.MakeMatrix3D[int32](3, 0, 0)
	hf.dequantHFCoeff = util.MakeMatrix3D[float32](3, 0, 0)

	for c := 0; c < 3; c++ {
		sY := size.Height >> header.jpegUpsamplingY[c]
		sX := size.Width >> header.jpegUpsamplingX[c]
		hf.quantizedCoeffs[c] = util.MakeMatrix2D[int32](sY, sX)
		hf.dequantHFCoeff[c] = util.MakeMatrix2D[float32](sY, sX)
	}

	hf.groupPos = hf.frame.groupPosInLFGroup(hf.lfg.lfGroupID, hf.groupID)
	hf.groupPos.Y <<= 5
	hf.groupPos.X <<= 5
	hf.blocks = make([]util.Point, len(hf.lfg.hfMetadata.blockList))
	for i := 0; i < len(hf.lfg.hfMetadata.blockList); i++ {
		posInLfg := hf.lfg.hfMetadata.blockList[i]
		groupY := posInLfg.Y - hf.groupPos.Y
		groupX := posInLfg.X - hf.groupPos.X
		if groupY < 0 || groupX < 0 || groupY >= 32 || groupX >= 32 {
			continue
		}
		hf.blocks[i] = posInLfg
		tt := hf.lfg.hfMetadata.dctSelect[posInLfg.Y][posInLfg.X]
		flip := tt.flip()
		hfMult := hf.lfg.hfMetadata.hfMultiplier[posInLfg.Y][posInLfg.X]
		lfIndex := hf.lfg.lfCoeff.lfIndex[posInLfg.Y][posInLfg.X]
		numBlocks := tt.dctSelectHeight * tt.dctSelectWidth
		for _, c := range cMap {
			sGroupY := groupY >> header.jpegUpsamplingY[c]
			sGroupX := groupX >> header.jpegUpsamplingX[c]
			if groupY != sGroupY<<header.jpegUpsamplingY[c] || groupX != sGroupX<<header.jpegUpsamplingX[c] {
				continue
			}

			pixelGroupY := sGroupY << 3
			pixelGroupX := sGroupX << 3
			predicted := getPredictedNonZeros(nonZeros, c, sGroupY, sGroupX)
			blockCtx := getBlockContext(c, tt.orderID, hfMult, lfIndex)
		}

	}
	return nil, nil
}

func (hf *HFCoefficients) getBlockContext(c int, orderID int32, hfMult int32, lfIndex int32) int32 {

	var idx int
	if c < 2 {
		idx = 1 - c
	} else {
		idx = c
	}
	idx = idx*13 + int(orderID)
	idx *= len(hf.hfctx.qfThresholds) + 1
	for _, t := range hf.hfctx.qfThresholds {
		if hfMult > t {
			idx++
		}
	}
	idx *= int(hf.hfctx.numLFContexts)
	return int32(hf.hfctx.clusterMap[int32(idx)+lfIndex])
}

func getPredictedNonZeros(nonZeros [][][]int32, c int, y int32, x int32) int32 {
	if x == 0 && y == 0 {
		return 32
	}
	if x == 0 {
		return nonZeros[c][y-1][x]
	}

	if y == 0 {
		return nonZeros[c][0][x-1]
	}

	return (nonZeros[c][y-1][x] + nonZeros[c][y][x-1] + 1) >> 1
}
