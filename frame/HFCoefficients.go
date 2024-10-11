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
	xxxxxxxxxxx
	return hf, nil
}
