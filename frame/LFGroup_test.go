package frame

import (
	"testing"

	"github.com/kpfaulkner/jxl-go/image"
	"github.com/kpfaulkner/jxl-go/testcommon"
)

func TestNewLFGroupWithReader(t *testing.T) {

	for _, tc := range []struct {
		name               string
		replaced           []ModularChannel
		lfBuffer           []image.ImageBuffer
		boolData           []bool
		enumData           []int32
		u32Data            []uint32
		f16Data            []float32
		bitsData           []uint64
		parent             Framer
		lfCoefficientsFunc NewLFCoefficientsWithReaderFunc
		hfMetadataFunc     NewHFMetadataWithReaderFunc
		expectedResult     LFCoefficients
		expectErr          bool
	}{
		//{
		//	name:      "no data",
		//	parent:    NewFakeFramer(),
		//	expectErr: true,
		//},
		{
			name:               "success",
			parent:             NewFakeFramer(),
			lfCoefficientsFunc: NewFakeLFCoeffientsFunc,
			hfMetadataFunc:     NewFakeHFMetadataFunc,
			boolData:           []bool{true, true, false},
			enumData:           nil,
			u32Data:            []uint32{0, 0},
			bitsData:           []uint64{},
			expectErr:          false,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {

			bitReader := &testcommon.FakeBitReader{
				ReadU32Data:  tc.u32Data,
				ReadBitsData: tc.bitsData,
				ReadBoolData: tc.boolData,
				ReadF16Data:  tc.f16Data,
			}
			lfg, err := NewLFGroupWithReader(bitReader, tc.parent, 0, tc.replaced, tc.lfBuffer, tc.lfCoefficientsFunc, tc.hfMetadataFunc)
			if err != nil && !tc.expectErr {
				t.Errorf("got error when none was expected : %v", err)
			}
			if err == nil && tc.expectErr {
				t.Errorf("expected error but got none")
			}
			if err != nil && tc.expectErr {
				return
			}

			// Not checking contents of lfg yet, just want to make sure it's not nil for now
			if lfg == nil {
				t.Errorf("nil LFGlobal")
			}
		})
	}
}
