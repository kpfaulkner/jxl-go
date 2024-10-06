package frame

import (
	"errors"

	"github.com/kpfaulkner/jxl-go/jxlio"
	"github.com/kpfaulkner/jxl-go/util"
)

type LFCoefficients struct {
	dequantLFCoeff [][][]float32
	lfIndex        [][]int32
	frame          *Frame
}

func NewLFCoefficientsWithReader(reader *jxlio.Bitreader, parent LFGroup, frame *Frame, lfBuffer [][][]float32) (*LFCoefficients, error) {
	lf := &LFCoefficients{}

	lf.frame = frame
	lf.lfIndex = util.MakeMatrix2D[int32](parent.size.Height, parent.size.Width)
	header := frame.Header
	adapativeSmoothing := (header.Flags & (SKIP_ADAPTIVE_LF_SMOOTHING | USE_LF_FRAME)) == 0
	info := make([]*ModularChannel, 3)
	dequantLFCoeff := util.MakeMatrix3D[float32](3, 0, 0)
	subSampled := header.jpegUpsamplingY[0] != 0 || header.jpegUpsamplingY[1] != 0 || header.jpegUpsamplingY[2] != 0 ||
		header.jpegUpsamplingX[0] != 0 || header.jpegUpsamplingX[1] != 0 || header.jpegUpsamplingX[2] != 0
	if adapativeSmoothing && subSampled {
		return nil, errors.New("Adaptive smoothing is incompatible with subsampling")
	}
	for i := 0; i < 3; i++ {
		sizeY := parent.size.Height >> header.jpegUpsamplingY[i]
		sizeX := parent.size.Width >> header.jpegUpsamplingX[i]
		info[cMap[i]] = NewModularChannelWithAllParams(sizeY, sizeX, header.jpegUpsamplingY[i], header.jpegUpsamplingX[i], false)
		dequantLFCoeff[i] = util.MakeMatrix2D[float32, uint32](sizeY, sizeX)
	}

	if (header.Flags & USE_LF_FRAME) != 0 {
		pos := frame.getLFGroupLocation(parent.lfGroupID)
		pY := pos.Y << 8
		pX := pos.X << 8
		lf.dequantLFCoeff = dequantLFCoeff
		for c := 0; c < 3; c++ {
			for y := int32(0); y < int32(len(dequantLFCoeff[c])); y++ {
				for x, d := pX, 0; x < pX+int32(len(dequantLFCoeff[c][y])); x, d = x+1, d+1 {
					dequantLFCoeff[c][y][d] = lfBuffer[c][pY+y][x]
				}
			}
		}
	}

	panic("not implemented yet")

	return lf, nil
}
