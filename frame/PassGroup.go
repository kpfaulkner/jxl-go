package frame

import (
	"errors"

	"github.com/kpfaulkner/jxl-go/jxlio"
	"github.com/kpfaulkner/jxl-go/util"
)

type PassGroup struct {
	modularPassGroupBuffer [][][]int32
	modularStream          *ModularStream
	frame                  *Frame
	groupID                uint32
	passID                 uint32
	hfCoefficients         *HFCoefficients
	lfg                    *LFGroup
}

func NewPassGroupWithReader(reader *jxlio.Bitreader, frame *Frame, pass uint32, group uint32, replacedChannels []ModularChannel) (*PassGroup, error) {

	pg := &PassGroup{}
	pg.frame = frame
	pg.groupID = group
	pg.passID = pass
	if frame.Header.Encoding == VARDCT {
		coeff, err := NewHFCoefficientsWithReader(reader, frame, pass, group)
		if err != nil {
			return nil, err
		}
		pg.hfCoefficients = coeff
	} else {
		pg.hfCoefficients = nil
	}

	stream, err := NewModularStreamWithStreamIndex(reader, frame, int(18+3*frame.numLFGroups+frame.numGroups*pass+group), replacedChannels)
	if err != nil {
		return nil, err
	}

	pg.modularStream = stream
	err = stream.decodeChannels(reader, false)
	if err != nil {
		return nil, err
	}

	pg.lfg = frame.getLFGroupForGroup(int32(group))

	return pg, nil
}

func (g PassGroup) invertVarDCT(frameBuffer [][][]float32, prev *PassGroup) error {
	header := g.frame.Header
	//zero := util.Point{}
	if prev != nil {
		panic("not implemented")
	}

	if err := g.hfCoefficients.bakeDequantizedCoeffs(); err != nil {
		return err
	}

	groupLocation := g.frame.getGroupLocation(int32(g.groupID))
	groupLocation.Y <<= 8
	groupLocation.X <<= 8

	coeffs := g.hfCoefficients.dequantHFCoeff
	scratchBlock := util.MakeMatrix3D[float32](5, 256, 256)
	for i := 0; i < len(g.hfCoefficients.blocks); i++ {
		posInLFG := g.hfCoefficients.blocks[i]
		// Zero value then continue? TODO(kpfaulkner) check this!
		if posInLFG.X == 0 && posInLFG.Y == 0 {
			continue
		}
		tt := g.lfg.hfMetadata.dctSelect[posInLFG.Y][posInLFG.X]
		groupY := posInLFG.Y - g.hfCoefficients.groupPos.Y
		groupX := posInLFG.X - g.hfCoefficients.groupPos.X
		for c := 0; c < 3; c++ {
			sGroupY := groupY >> header.jpegUpsamplingY[c]
			sGroupX := groupX >> header.jpegUpsamplingX[c]
			if sGroupY<<header.jpegUpsamplingY[c] != groupY ||
				sGroupX<<header.jpegUpsamplingX[c] != groupX {
				continue
			}
			ppg := util.Point{
				X: sGroupX << 3,
				Y: sGroupY << 3,
			}
			ppf := util.Point{
				X: ppg.X + (groupLocation.X >> header.jpegUpsamplingX[c]),
				Y: ppg.Y + (groupLocation.Y >> header.jpegUpsamplingY[c]),
			}
			//var foeff0 float32
			//var foeff1 float32
			//lfs := make([]float32, 2)
			switch tt.transformMethod {
			case METHOD_DCT:
				util.InverseDCT2D(coeffs[c], frameBuffer[c], ppg, ppf, tt.getPixelSize(), scratchBlock[0], scratchBlock[1], false)
				break
			case METHOD_DCT8_4:
				panic("not implemented")
			case METHOD_DCT4_8:
				panic("not implemented")
			case METHOD_AFV:
				panic("not implemented")
			case METHOD_DCT2:
				panic("not implemented")
			case METHOD_HORNUSS:
				panic("not implemented")
			case METHOD_DCT4:
				panic("not implemented")
			default:
				return errors.New("Transform not implemented")
			}
		}

	}
	return nil
}
