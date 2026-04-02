package frame

import (
	"testing"

	"github.com/kpfaulkner/jxl-go/testcommon"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRestorationFilterWithReader_AllDefault(t *testing.T) {
	reader := &testcommon.FakeBitReader{
		ReadBoolData: []bool{true},
	}
	rf, err := NewRestorationFilterWithReader(reader, VARDCT)
	require.NoError(t, err)
	assert.True(t, rf.gab)
	assert.Equal(t, uint32(2), rf.epfIterations)
}

func TestNewRestorationFilterWithReader_CustomGab(t *testing.T) {
	reader := &testcommon.FakeBitReader{
		ReadBoolData: []bool{
			false, // allDefault
			true,  // gab
			true,  // customGab
			false, // epfSharpCustom (iterations 1 > 0, VARDCT)
			false, // epfSigmaCustom (iterations 1 > 0)
		},
		ReadF16Data: []float32{
			0.1, 0.2, // gab1Weights[0], gab2Weights[0]
			0.3, 0.4, // gab1Weights[1], gab2Weights[1]
			0.5, 0.6, // gab1Weights[2], gab2Weights[2]
		},
		ReadBitsData: []uint64{
			1, // epfIterations
		},
		ReadU64Data: []uint64{0}, // ExtensionsKey
	}
	rf, err := NewRestorationFilterWithReader(reader, VARDCT)
	require.NoError(t, err)
	assert.True(t, rf.gab)
	assert.True(t, rf.customGab)
	assert.Equal(t, float32(0.1), rf.gab1Weights[0])
	assert.Equal(t, float32(0.6), rf.gab2Weights[2])
	assert.Equal(t, uint32(1), rf.epfIterations)
}

func TestNewRestorationFilterWithReader_CustomEpfSharp(t *testing.T) {
	reader := &testcommon.FakeBitReader{
		ReadBoolData: []bool{
			false, // allDefault
			false, // gab
			true,  // epfSharpCustom (iterations 1 > 0, VARDCT)
			false, // epfSigmaCustom (iterations 1 > 0)
		},
		ReadBitsData: []uint64{
			1, // epfIterations
		},
		ReadF16Data: []float32{
			0.1, 0.2, 0.3, 0.4, 0.5, 0.6, 0.7, 0.8, // epfSharpLut
		},
		ReadU64Data: []uint64{0}, // ExtensionsKey
	}
	rf, err := NewRestorationFilterWithReader(reader, VARDCT)
	require.NoError(t, err)
	assert.True(t, rf.epfSharpCustom)
}

func TestNewRestorationFilterWithReader_CustomEpfWeight(t *testing.T) {
	reader := &testcommon.FakeBitReader{
		ReadBoolData: []bool{
			false, // allDefault
			false, // gab
			false, // epfSharpCustom (iterations 10 > 0, VARDCT)
			true,  // epfWeightCustom (iterations 10 > 9)
			false, // epfSigmaCustom (iterations 10 > 0)
		},
		ReadBitsData: []uint64{
			10, // epfIterations > 9
			0,  // readBits(32)
		},
		ReadF16Data: []float32{
			1.1, 2.2, 3.3, // epfChannelScale
		},
		ReadU64Data: []uint64{0}, // ExtensionsKey
	}
	rf, err := NewRestorationFilterWithReader(reader, VARDCT)
	require.NoError(t, err)
	assert.True(t, rf.epfWeightCustom)
	assert.Equal(t, float32(1.1), rf.epfChannelScale[0])
}

func TestNewRestorationFilterWithReader_CustomEpfSigma_Vardct(t *testing.T) {
	reader := &testcommon.FakeBitReader{
		ReadBoolData: []bool{
			false, // allDefault
			false, // gab
			false, // epfSharpCustom (iterations 1 > 0, VARDCT)
			true,  // epfSigmaCustom (iterations 1 > 0)
		},
		ReadBitsData: []uint64{
			1, // epfIterations
		},
		ReadF16Data: []float32{
			0.5, // epfQuantMul
			1.1, // epfPass0SigmaScale
			2.2, // epfPass2SigmaScale
			3.3, // epfBorderSadMul
		},
		ReadU64Data: []uint64{0}, // ExtensionsKey
	}
	rf, err := NewRestorationFilterWithReader(reader, VARDCT)
	require.NoError(t, err)
	assert.True(t, rf.epfSigmaCustom)
}

func TestNewRestorationFilterWithReader_ModularSigma(t *testing.T) {
	reader := &testcommon.FakeBitReader{
		ReadBoolData: []bool{
			false, // allDefault
			false, // gab
			false, // epfSigmaCustom (iterations 1 > 0)
			// no epfSharpCustom because encoding is MODULAR
		},
		ReadBitsData: []uint64{
			1, // epfIterations
		},
		ReadF16Data: []float32{
			2.5, // epfSigmaForModular
		},
		ReadU64Data: []uint64{0}, // ExtensionsKey
	}
	rf, err := NewRestorationFilterWithReader(reader, MODULAR)
	require.NoError(t, err)
	assert.Equal(t, float32(2.5), rf.epfSigmaForModular)
}

func TestNewRestorationFilterWithReader_NoIterations(t *testing.T) {
	reader := &testcommon.FakeBitReader{
		ReadBoolData: []bool{
			false, // allDefault
			false, // gab
			// no epfSharpCustom because iterations == 0
			// no epfSigmaCustom because iterations == 0
			// no epfSigmaForModular because iterations == 0
		},
		ReadBitsData: []uint64{
			0, // epfIterations
		},
		ReadU64Data: []uint64{0}, // ExtensionsKey
	}
	rf, err := NewRestorationFilterWithReader(reader, VARDCT)
	require.NoError(t, err)
	assert.Equal(t, uint32(0), rf.epfIterations)
}

func TestNewRestorationFilterWithReader_SigmaCustom_Modular(t *testing.T) {
	reader := &testcommon.FakeBitReader{
		ReadBoolData: []bool{
			false, // allDefault
			false, // gab
			true,  // epfSigmaCustom (iterations 1 > 0)
		},
		ReadBitsData: []uint64{
			1, // epfIterations
		},
		ReadF16Data: []float32{
			1.1, // epfPass0SigmaScale
			2.2, // epfPass2SigmaScale
			3.3, // epfBorderSadMul
			// no epfQuantMul because encoding is MODULAR
			4.4, // epfSigmaForModular (iterations 1 > 0, MODULAR)
		},
		ReadU64Data: []uint64{0}, // ExtensionsKey
	}
	rf, err := NewRestorationFilterWithReader(reader, MODULAR)
	require.NoError(t, err)
	assert.True(t, rf.epfSigmaCustom)
	assert.Equal(t, float32(0.46), rf.epfQuantMul) // Default
	assert.Equal(t, float32(4.4), rf.epfSigmaForModular)
}

func TestNewRestorationFilterWithReader_Error(t *testing.T) {
	// EOF on first ReadBool (allDefault)
	reader := &testcommon.FakeBitReader{
		ReadBoolData: []bool{},
	}
	_, err := NewRestorationFilterWithReader(reader, VARDCT)
	assert.Error(t, err)

	// EOF on second ReadBool (rf.gab)
	reader = &testcommon.FakeBitReader{
		ReadBoolData: []bool{false},
	}
	_, err = NewRestorationFilterWithReader(reader, VARDCT)
	assert.Error(t, err)

	// EOF on rf.customGab
	reader = &testcommon.FakeBitReader{
		ReadBoolData: []bool{false, true},
	}
	_, err = NewRestorationFilterWithReader(reader, VARDCT)
	assert.Error(t, err)

	// EOF on rf.gab1Weights
	reader = &testcommon.FakeBitReader{
		ReadBoolData: []bool{false, true, true},
		ReadF16Data:  []float32{0.1, 0.2, 0.3}, // Only some of the 6 weights
	}
	_, err = NewRestorationFilterWithReader(reader, VARDCT)
	assert.Error(t, err)

	// EOF on rf.epfIterations
	reader = &testcommon.FakeBitReader{
		ReadBoolData: []bool{false, false}, // not all default, gab=false
		ReadBitsData: []uint64{},           // No bits for iterations
	}
	_, err = NewRestorationFilterWithReader(reader, VARDCT)
	assert.Error(t, err)

	// EOF on rf.epfSharpCustom
	reader = &testcommon.FakeBitReader{
		ReadBoolData: []bool{false, false},
		ReadBitsData: []uint64{1}, // iterations > 0, VARDCT
		// Missing rf.epfSharpCustom bool
	}
	_, err = NewRestorationFilterWithReader(reader, VARDCT)
	assert.Error(t, err)

	// EOF on rf.epfSharpLut
	reader = &testcommon.FakeBitReader{
		ReadBoolData: []bool{false, false, true},
		ReadBitsData: []uint64{1},
		ReadF16Data:  []float32{0.1}, // Only one value for LUT (needs 8)
	}
	_, err = NewRestorationFilterWithReader(reader, VARDCT)
	assert.Error(t, err)

	// EOF on rf.epfWeightCustom
	reader = &testcommon.FakeBitReader{
		ReadBoolData: []bool{false, false, false},
		ReadBitsData: []uint64{10}, // iterations > 9
		// Missing rf.epfWeightCustom bool
	}
	_, err = NewRestorationFilterWithReader(reader, VARDCT)
	assert.Error(t, err)

	// EOF on rf.epfChannelScale
	reader = &testcommon.FakeBitReader{
		ReadBoolData: []bool{false, false, false, true},
		ReadBitsData: []uint64{10},
		ReadF16Data:  []float32{0.1}, // Only one value (needs 3)
	}
	_, err = NewRestorationFilterWithReader(reader, VARDCT)
	assert.Error(t, err)

	// EOF on ReadBits(32) after epfChannelScale
	reader = &testcommon.FakeBitReader{
		ReadBoolData: []bool{false, false, false, true},
		ReadBitsData: []uint64{10}, // iterations > 9
		ReadF16Data:  []float32{1.1, 2.2, 3.3},
		// Missing 32 bits
	}
	_, err = NewRestorationFilterWithReader(reader, VARDCT)
	assert.Error(t, err)

	// EOF on rf.epfSigmaCustom
	reader = &testcommon.FakeBitReader{
		ReadBoolData: []bool{false, false, false},
		ReadBitsData: []uint64{1},
		// Missing rf.epfSigmaCustom bool
	}
	_, err = NewRestorationFilterWithReader(reader, VARDCT)
	assert.Error(t, err)

	// EOF on rf.epfQuantMul
	reader = &testcommon.FakeBitReader{
		ReadBoolData: []bool{false, false, false, true},
		ReadBitsData: []uint64{1},
		ReadF16Data:  []float32{}, // Needs 1 for QuantMul
	}
	_, err = NewRestorationFilterWithReader(reader, VARDCT)
	assert.Error(t, err)

	// EOF on rf.epfPass0SigmaScale etc
	reader = &testcommon.FakeBitReader{
		ReadBoolData: []bool{false, false, false, true},
		ReadBitsData: []uint64{1},
		ReadF16Data:  []float32{0.5, 0.6}, // Needs more for scales
	}
	_, err = NewRestorationFilterWithReader(reader, VARDCT)
	assert.Error(t, err)

	// EOF on rf.epfSigmaForModular
	reader = &testcommon.FakeBitReader{
		ReadBoolData: []bool{false, false, false},
		ReadBitsData: []uint64{1},
		// Missing rf.epfSigmaForModular f16
	}
	_, err = NewRestorationFilterWithReader(reader, MODULAR)
	assert.Error(t, err)

	// EOF on bundle.NewExtensionsWithReader
	reader = &testcommon.FakeBitReader{
		ReadBoolData: []bool{false, false, false, false},
		ReadBitsData: []uint64{0},
		ReadU64Data:  []uint64{}, // EOF on extension key
	}
	_, err = NewRestorationFilterWithReader(reader, VARDCT)
	assert.Error(t, err)
}
