package frame

import (
	"github.com/kpfaulkner/jxl-go/entropy"
	"github.com/kpfaulkner/jxl-go/jxlio"
)

type LFGlobal struct {
	frame           *Frame
	Patches         []Patch
	splines         []SplinesBundle
	noiseParameters []NoiseParameters
	lfDequant       []float32
	quantizer       *Quantizer
	hfBlockCtx      *HFBlockContext
	lfChanCorr      *LFChannelCorrelation
	gModular        *GlobalModular
}

func NewLFGlobal() *LFGlobal {
	lf := &LFGlobal{}
	lf.lfDequant = []float32{1.0 / 4096.0, 1.0 / 512.0, 1.0 / 256.0}
	return lf
}

func NewLFGlobalWithReader(reader *jxlio.Bitreader, parent *Frame) (*LFGlobal, error) {

	lf := NewLFGlobal()
	lf.frame = parent

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
		panic("Splines not implemented yet")
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
		lf.quantizer, err = NewQuantizerWithReader(reader, lf.lfDequant)
		if err != nil {
			return nil, err
		}
		lf.hfBlockCtx, err = NewHF
		panic("VARDCT not implemented")
	} else {
		lf.quantizer = nil
		lf.hfBlockCtx = nil
		lf.lfChanCorr, err = NewLFChannelCorrelation()
		if err != nil {
			return nil, err
		}
	}

	lf.gModular, err = NewGlobalModularWithReader(reader, lf.frame)
	if err != nil {
		return nil, err
	}

	return lf, nil
}
