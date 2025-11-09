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

func TestNewLFGlobalWithReader(t *testing.T) {

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
		//{
		//	name:      "no data",
		//	parent:    NewFakeFramer(),
		//	expectErr: true,
		//},
		{
			name:               "success",
			parent:             NewFakeFramer(),
			hfBlockContextFunc: NewFakeHFBlockContextFunc,
			boolData:           []bool{true, true, false, true, true},
			enumData:           nil,
			u32Data:            []uint32{0, 0, 0},
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

			// Not checking contents of lfg yet, just want to make sure it's not nil for now
			if lfg == nil {
				t.Errorf("nil LFGlobal")
			}
		})
	}
}
