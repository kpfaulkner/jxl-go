package frame

import (
	"errors"

	"github.com/kpfaulkner/jxl-go/color"
	"github.com/kpfaulkner/jxl-go/entropy"
	"github.com/kpfaulkner/jxl-go/jxlio"
)

type LFGlobal struct {
	frame           *Frame
	Patches         []Patch
	splines         *SplinesBundle
	noiseParameters []NoiseParameters
	lfDequant       []float32
	hfBlockCtx      *HFBlockContext
	lfChanCorr      *LFChannelCorrelation
	//gModular        *GlobalModular
	globalScale   int32
	quantLF       int32
	scaledDequant []float32
	globalModular *ModularStream
}

func NewLFGlobal() *LFGlobal {
	lf := &LFGlobal{}
	lf.lfDequant = []float32{1.0 / 4096.0, 1.0 / 512.0, 1.0 / 256.0}
	lf.scaledDequant = make([]float32, 3)
	return lf
}

func NewLFGlobalWithReader(reader *jxlio.Bitreader, parent *Frame) (*LFGlobal, error) {

	lf := NewLFGlobal()
	lf.frame = parent
	extra := len(lf.frame.globalMetadata.ExtraChannelInfo)
	if lf.frame.Header.Flags&PATCHES != 0 {

		// TODO(kpfaulkner) not used yet with the lossless images I'm trying.
		panic("Patches not implemented yet")
		stream, err := entropy.NewEntropyStreamWithReaderAndNumDists(reader, 10)
		if err != nil {
			return nil, err
		}
		numPatches, err := stream.ReadSymbol(reader, 0)
		if err != nil {
			return nil, err
		}
		lf.Patches = make([]Patch, numPatches)
		for i := 0; i < int(numPatches); i++ {
			lf.Patches[i], err = NewPatchWithStreamAndReader(stream, reader, len(parent.globalMetadata.ExtraChannelInfo), len(parent.globalMetadata.AlphaIndices))
			if err != nil {
				return nil, err
			}
		}

	} else {
		lf.Patches = []Patch{}
	}

	if lf.frame.Header.Flags&SPLINES != 0 {
		if lf.frame.globalMetadata.GetColourChannelCount() < 3 {
			return nil, errors.New("Cannot do splines in grayscale")
		}
		var err error
		lf.splines, err = NewSplinesBundleWithReader(reader)
		if err != nil {
			return nil, err
		}
	} else {
		lf.splines = nil
	}

	if lf.frame.Header.Flags&NOISE != 0 {
		panic("noise not implemented yet")
	} else {
		lf.noiseParameters = nil
	}

	if !reader.MustReadBool() {
		for i := 0; i < 3; i++ {
			lf.lfDequant[i] = reader.MustReadF16() * (1.0 / 128.0)
		}
	}

	var err error
	if lf.frame.Header.Encoding == VARDCT {
		//lf.quantizer, err = NewQuantizerWithReader(reader, lf.lfDequant)
		//if err != nil {
		//	return nil, err
		//}
		globalScale, err := reader.ReadU32(1, 11, 2049, 11, 4097, 12, 8193, 16)
		if err != nil {
			return nil, err
		}
		lf.globalScale = int32(globalScale)
		quantLF, err := reader.ReadU32(16, 0, 1, 5, 1, 8, 1, 16)
		if err != nil {
			return nil, err
		}
		lf.quantLF = int32(quantLF)
		for i := 0; i < 3; i++ {
			lf.scaledDequant[i] = (1 << 16) * lf.lfDequant[i] / float32(lf.globalScale*lf.quantLF)
		}
		lf.hfBlockCtx, err = NewHFBlockContextWithReader(reader)
		if err != nil {
			return nil, err
		}
		lf.lfChanCorr, err = NewLFChannelCorrelationWithReader(reader)
		if err != nil {
			return nil, err
		}
	} else {
		lf.globalScale = 0
		lf.quantLF = 0
		lf.hfBlockCtx = nil
		lf.lfChanCorr, err = NewLFChannelCorrelation()
		if err != nil {
			return nil, err
		}
	}

	hasGlobalTree, err := reader.ReadBool()
	if err != nil {
		return nil, err
	}
	var globalTree *MATree
	if hasGlobalTree {
		globalTree, err = NewMATreeWithReader(reader)
		if err != nil {
			return nil, err
		}
	} else {
		globalTree = nil
	}
	lf.frame.globalTree = globalTree
	subModularChannelCount := extra
	ecStart := 0
	if lf.frame.Header.Encoding == MODULAR {
		if lf.frame.Header.DoYCbCr && !lf.frame.globalMetadata.XybEncoded &&
			lf.frame.globalMetadata.ColorEncoding.ColorEncoding == color.CE_GRAY {
			ecStart = 1
		} else {
			ecStart = 3
		}
	}
	subModularChannelCount += ecStart

	globalModular, err := NewModularStreamWithReader(reader, parent, 0, subModularChannelCount, ecStart)
	if err != nil {
		return nil, err
	}
	lf.globalModular = globalModular
	if err = lf.globalModular.decodeChannels(reader, true); err != nil {
		return nil, err
	}

	return lf, nil
}
