package frame

import (
	"testing"

	"github.com/kpfaulkner/jxl-go/image"
	"github.com/kpfaulkner/jxl-go/jxlio"
	"github.com/kpfaulkner/jxl-go/testcommon"
	"github.com/kpfaulkner/jxl-go/util"
)

func TestNNewLFCoefficientsWithReader(t *testing.T) {

	for _, tc := range []struct {
		name           string
		boolData       []bool
		enumData       []int32
		u32Data        []uint32
		f16Data        []float32
		bitsData       []uint64
		parent         *LFGroup
		frame          Framer
		lfBuffer       []image.ImageBuffer
		expectedResult LFCoefficients
		expectErr      bool
	}{
		//{
		//	name: "no data",
		//	parent: &LFGroup{
		//		size: util.Dimension{
		//			Width:  5,
		//			Height: 5,
		//		},
		//	},
		//	frame:     NewFakeFramer(),
		//	expectErr: true,
		//},
		{
			name:  "success",
			frame: NewFakeFramer(),
			parent: &LFGroup{
				size: util.Dimension{
					Width:  5,
					Height: 5,
				},
			},
			boolData: []bool{true},
			enumData: nil,
			u32Data:  []uint32{},
			bitsData: []uint64{
				0,
			},
			expectedResult: LFCoefficients{},
			expectErr:      false,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {

			bitReader := &testcommon.FakeBitReader{
				ReadU32Data:  tc.u32Data,
				ReadBitsData: tc.bitsData,
				ReadBoolData: tc.boolData,
				ReadF16Data:  tc.f16Data,
			}
			lfcc, err := NewLFCoefficientsWithReader(bitReader, tc.parent, tc.frame, tc.lfBuffer, fakeNewModularStream)
			if err != nil && !tc.expectErr {
				t.Errorf("got error when none was expected : %v", err)
			}
			if err == nil && tc.expectErr {
				t.Errorf("expected error but got none")
			}
			if err != nil && tc.expectErr {
				return
			}

			// Not checking contents of lfcc yet, just want to make sure it's not nil for now
			if lfcc == nil {
				t.Errorf("nil LF Coefficients")
			}
		})
	}
}

type fakeModularStream struct{}

func (fms *fakeModularStream) applyTransforms() error {
	return nil
}

func (fms *fakeModularStream) getChannels() []*ModularChannel {
	//TODO implement me
	panic("implement me")
}

func (fms *fakeModularStream) decodeChannels(reader jxlio.BitReader, partial bool) error {

	// TODO(kpfaulkner) 20251025
	return nil
}

func (fms *fakeModularStream) getDecodedBuffer() [][][]int32 {
	return util.MakeMatrix3D[int32](5, 5, 5)
}

func fakeNewModularStream(reader jxlio.BitReader, frame Framer, index int, count int, start int, array []ModularChannel) (ModularStreamer, error) {
	return &fakeModularStream{}, nil
}
