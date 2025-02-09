package frame

import (
	"errors"
	"image"

	"github.com/kpfaulkner/jxl-go/entropy"
	"github.com/kpfaulkner/jxl-go/jxlio"
	"github.com/kpfaulkner/jxl-go/util"
)

type SplinesBundle struct {
	numSplines    int32
	quantAdjust   int32
	controlCounts []int32
	controlPoints [][]image.Point
	coeffX        [][]int32
	coeffY        [][]int32
	coeffB        [][]int32
	coeffSigma    [][]int32
}

func NewSplinesBundleWithReader(reader *jxlio.Bitreader) (*SplinesBundle, error) {
	sb := &SplinesBundle{}

	var stream *entropy.EntropyStream
	var err error
	if stream, err = entropy.NewEntropyStreamWithReaderAndNumDists(reader, 6); err != nil {
		return nil, err
	}
	if sb.numSplines, err = stream.ReadSymbol(reader, 2); err != nil {
		return nil, err
	} else {
		sb.numSplines++
	}

	splinePos := make([]image.Point, sb.numSplines)
	for i := 0; i < int(sb.numSplines); i++ {
		x, err := stream.ReadSymbol(reader, 1)
		if err != nil {
			return nil, err
		}
		y, err := stream.ReadSymbol(reader, 1)
		if err != nil {
			return nil, err
		}
		if i != 0 {
			x = jxlio.UnpackSigned(uint32(x)) + int32(splinePos[i-1].X)
			y = jxlio.UnpackSigned(uint32(y)) + int32(splinePos[i-1].Y)
		}
		splinePos[i] = image.Point{X: int(x), Y: int(y)}
	}

	var sym int32
	if sym, err = stream.ReadSymbol(reader, 0); err != nil {
		return nil, err
	}
	sb.quantAdjust = jxlio.UnpackSigned(uint32(sym))
	sb.controlCounts = make([]int32, sb.numSplines)
	sb.controlPoints = util.MakeMatrix2D[image.Point](int(sb.numSplines), 0)
	sb.coeffX = util.MakeMatrix2D[int32](int(sb.numSplines), 32)
	sb.coeffY = util.MakeMatrix2D[int32](int(sb.numSplines), 32)
	sb.coeffB = util.MakeMatrix2D[int32](int(sb.numSplines), 32)
	sb.coeffSigma = util.MakeMatrix2D[int32](int(sb.numSplines), 32)
	for i := int32(0); i < sb.numSplines; i++ {
		if sym, err = stream.ReadSymbol(reader, 3); err != nil {
			return nil, err
		}
		sb.controlCounts[i] = 1 + sym
		sb.controlPoints[i] = make([]image.Point, int(sb.controlCounts[i]))
		sb.controlPoints[i][0] = image.Point{splinePos[i].X, splinePos[i].Y}
		deltaY := make([]int32, sb.controlCounts[i]-1)
		deltaX := make([]int32, len(deltaY))
		var data int32

		for j := 0; j < len(deltaY); j++ {
			if data, err = stream.ReadSymbol(reader, 4); err != nil {
				return nil, err
			}
			deltaX[j] = jxlio.UnpackSigned(uint32(data))
			if data, err = stream.ReadSymbol(reader, 4); err != nil {
				return nil, err
			}
			deltaY[j] = jxlio.UnpackSigned(uint32(data))
		}
		cY := sb.controlPoints[i][0].Y
		cX := sb.controlPoints[i][0].X
		dY := 0
		dX := 0
		for j := int32(1); j < sb.controlCounts[i]; j++ {
			dY += int(deltaY[j-1])
			dX += int(deltaX[j-1])
			cY += dY
			cX += dX
			sb.controlPoints[i][j] = image.Point{cX, cY}
		}

		for j := 0; j < 32; j++ {
			if data, err = stream.ReadSymbol(reader, 5); err != nil {
				return nil, err
			}
			sb.coeffX[i][j] = data
		}
		for j := 0; j < 32; j++ {
			if data, err = stream.ReadSymbol(reader, 5); err != nil {
				return nil, err
			}
			sb.coeffY[i][j] = data
		}
		for j := 0; j < 32; j++ {
			if data, err = stream.ReadSymbol(reader, 5); err != nil {
				return nil, err
			}
			sb.coeffB[i][j] = data
		}
		for j := 0; j < 32; j++ {
			if data, err = stream.ReadSymbol(reader, 5); err != nil {
				return nil, err
			}
			sb.coeffSigma[i][j] = data
		}
	}
	if !stream.ValidateFinalState() {
		return nil, errors.New("Illegal final ANS state")
	}
	return sb, nil
}
