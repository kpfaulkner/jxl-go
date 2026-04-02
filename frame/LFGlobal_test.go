package frame

import (
	"testing"

	"github.com/kpfaulkner/jxl-go/testcommon"
	"github.com/stretchr/testify/assert"
)

func TestNewLFGlobal(t *testing.T) {
	g := NewLFGlobal()
	assert.NotNil(t, g)
}

func Disabled_TestNewLFGlobalWithReader(t *testing.T) {

	for _, tc := range []struct {
		name               string
		boolData           []bool
		enumData           []int32
		u32Data            []uint32
		f16Data            []float32
		bitsData           []uint64
		parent             Framer
		hfBlockContextFunc NewHFBlockContextFunc
		expectedResult     LFCoefficients
		expectErr          bool
	}{
		{
			name:               "success",
			parent:             NewFakeFramer(VARDCT),
			hfBlockContextFunc: NewFakeHFBlockContextFunc,
			boolData:           []bool{true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true},
			u32Data:            []uint32{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			bitsData:           []uint64{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			expectErr:          false,
		},
		{
			name:               "success modular",
			parent:             NewFakeFramer(MODULAR),
			hfBlockContextFunc: NewFakeHFBlockContextFunc,
			boolData:           []bool{true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true},
			u32Data:            []uint32{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			bitsData:           []uint64{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			expectErr:          false,
		},
		{
			name:               "patches error",
			parent:             NewFakeFramerWithFlags(VARDCT, PATCHES),
			hfBlockContextFunc: NewFakeHFBlockContextFunc,
			expectErr:          true,
		},
		{
			name:               "splines error",
			parent:             NewFakeFramerWithFlags(VARDCT, SPLINES),
			hfBlockContextFunc: NewFakeHFBlockContextFunc,
			expectErr:          true,
		},
		{
			name:               "noise error",
			parent:             NewFakeFramerWithFlags(VARDCT, NOISE),
			hfBlockContextFunc: NewFakeHFBlockContextFunc,
			expectErr:          true,
		},
		{
			name:               "readDequant false",
			parent:             NewFakeFramer(MODULAR),
			hfBlockContextFunc: NewFakeHFBlockContextFunc,
			boolData:           []bool{false, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true},
			f16Data:            []float32{1.0, 1.0, 1.0},
			u32Data:            []uint32{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			bitsData:           []uint64{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
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
			lfg, err := NewLFGlobalWithReader(bitReader, tc.parent, tc.hfBlockContextFunc)
			if err != nil && !tc.expectErr {
				t.Errorf("got error when none was expected : %v", err)
			}
			if err == nil && tc.expectErr {
				t.Errorf("expected error but got none")
			}
			if err != nil && tc.expectErr {
				return
			}

			if lfg == nil {
				t.Errorf("nil LFGlobal")
			}
		})
	}
}
