package frame

import (
	"testing"

	"github.com/kpfaulkner/jxl-go/bundle"
	"github.com/kpfaulkner/jxl-go/testcommon"
	"github.com/kpfaulkner/jxl-go/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func makeParent(width, height uint32, xybEncoded bool, numExtraChannels int) *bundle.ImageHeader {
	ih := &bundle.ImageHeader{
		Size:       util.Dimension{Width: width, Height: height},
		XybEncoded: xybEncoded,
	}
	ih.ExtraChannelInfo = make([]bundle.ExtraChannelInfo, numExtraChannels)
	return ih
}

func TestNewFrameHeaderWithReader_AllDefault_XybEncoded(t *testing.T) {
	parent := makeParent(100, 200, true, 0)
	reader := &testcommon.FakeBitReader{
		ReadBoolData: []bool{true}, // allDefault = true
	}

	fh, err := NewFrameHeaderWithReader(reader, parent)
	require.NoError(t, err)

	assert.Equal(t, uint32(REGULAR_FRAME), fh.FrameType)
	assert.Equal(t, uint32(VARDCT), fh.Encoding)
	assert.Equal(t, uint64(0), fh.Flags)
	assert.False(t, fh.DoYCbCr)
	assert.Equal(t, uint32(1), fh.Upsampling)
	assert.Empty(t, fh.EcUpsampling)
	assert.Equal(t, uint32(1), fh.groupSizeShift)
	assert.Equal(t, uint32(256), fh.groupDim)
	assert.Equal(t, uint32(2048), fh.lfGroupDim)
	assert.Equal(t, uint32(8), fh.logGroupDim)
	assert.Equal(t, uint32(11), fh.logLFGroupDIM)
	assert.Equal(t, uint32(3), fh.xqmScale)
	assert.Equal(t, uint32(2), fh.bqmScale)
	assert.Equal(t, uint32(0), fh.LfLevel)
	assert.False(t, fh.haveCrop)

	require.NotNil(t, fh.Bounds)
	assert.Equal(t, int32(0), fh.Bounds.Origin.X)
	assert.Equal(t, int32(0), fh.Bounds.Origin.Y)
	assert.Equal(t, uint32(100), fh.Bounds.Size.Width)
	assert.Equal(t, uint32(200), fh.Bounds.Size.Height)

	// normalFrame = false (allDefault), so IsLast = (FrameType == REGULAR_FRAME) = true
	assert.True(t, fh.IsLast)
	assert.Equal(t, uint32(0), fh.SaveAsReference)
	assert.False(t, fh.SaveBeforeCT)
	assert.Equal(t, "", fh.name)
	assert.Equal(t, uint32(0), fh.Duration)

	require.NotNil(t, fh.BlendingInfo)
	assert.Equal(t, uint32(BLEND_REPLACE), fh.BlendingInfo.Mode)

	assert.NotNil(t, fh.restorationFilter)
	assert.NotNil(t, fh.extensions)
	assert.NotNil(t, fh.passes)

	assert.Equal(t, []int32{0, 0, 0}, fh.jpegUpsamplingX)
	assert.Equal(t, []int32{0, 0, 0}, fh.jpegUpsamplingY)
}

func TestNewFrameHeaderWithReader_AllDefault_NotXybEncoded(t *testing.T) {
	parent := makeParent(100, 200, false, 0)
	reader := &testcommon.FakeBitReader{
		ReadBoolData: []bool{true}, // allDefault = true
	}

	fh, err := NewFrameHeaderWithReader(reader, parent)
	require.NoError(t, err)

	assert.Equal(t, uint32(VARDCT), fh.Encoding)
	assert.False(t, fh.DoYCbCr)
	// Not XybEncoded, so xqmScale/bqmScale get the else branch values
	assert.Equal(t, uint32(2), fh.xqmScale)
	assert.Equal(t, uint32(2), fh.bqmScale)
}

func TestNewFrameHeaderWithReader_ErrorOnFirstBool(t *testing.T) {
	parent := makeParent(100, 200, true, 0)
	reader := &testcommon.FakeBitReader{
		ReadBoolData: []bool{}, // empty -> error
	}

	fh, err := NewFrameHeaderWithReader(reader, parent)
	assert.Error(t, err)
	assert.Nil(t, fh)
}

func TestNewFrameHeaderWithReader_RegularFrame_VarDCT_XybEncoded(t *testing.T) {
	parent := makeParent(100, 200, true, 0)

	reader := &testcommon.FakeBitReader{
		ReadBoolData: []bool{
			false, // allDefault = false
			false, // haveCrop = false
			true,  // IsLast = true
			true,  // RestorationFilter allDefault = true
		},
		ReadBitsData: []uint64{
			0, // FrameType = REGULAR_FRAME
			0, // Encoding = VARDCT
			0, // upsampling (1 << 0 = 1)
			3, // xqmScale
			2, // bqmScale
		},
		ReadU32Data: []uint32{
			1, // numPasses = 1
			0, // BlendingInfo mode = BLEND_REPLACE
			0, // nameLen = 0
		},
	}

	fh, err := NewFrameHeaderWithReader(reader, parent)
	require.NoError(t, err)

	assert.Equal(t, uint32(REGULAR_FRAME), fh.FrameType)
	assert.Equal(t, uint32(VARDCT), fh.Encoding)
	assert.Equal(t, uint64(0), fh.Flags)
	assert.False(t, fh.DoYCbCr)
	assert.Equal(t, uint32(1), fh.Upsampling)
	assert.Equal(t, uint32(3), fh.xqmScale)
	assert.Equal(t, uint32(2), fh.bqmScale)
	assert.Equal(t, uint32(0), fh.LfLevel)
	assert.False(t, fh.haveCrop)
	assert.True(t, fh.IsLast)
	assert.Equal(t, uint32(0), fh.SaveAsReference)
	assert.False(t, fh.SaveBeforeCT)
	assert.Equal(t, "", fh.name)

	require.NotNil(t, fh.Bounds)
	assert.Equal(t, uint32(100), fh.Bounds.Size.Width)
	assert.Equal(t, uint32(200), fh.Bounds.Size.Height)
}

func TestNewFrameHeaderWithReader_LFFrame(t *testing.T) {
	parent := makeParent(100, 200, true, 0)

	reader := &testcommon.FakeBitReader{
		ReadBoolData: []bool{
			false, // allDefault = false
			true,  // RestorationFilter allDefault = true
		},
		ReadBitsData: []uint64{
			1, // FrameType = LF_FRAME
			0, // Encoding = VARDCT
			0, // upsampling (1 << 0 = 1)
			3, // xqmScale
			2, // bqmScale
			0, // LfLevel (0 + 1 = 1)
		},
		ReadU32Data: []uint32{
			1, // numPasses = 1
			0, // nameLen = 0
		},
	}

	fh, err := NewFrameHeaderWithReader(reader, parent)
	require.NoError(t, err)

	assert.Equal(t, uint32(LF_FRAME), fh.FrameType)
	assert.Equal(t, uint32(1), fh.LfLevel)
	// normalFrame = false (LF_FRAME), IsLast = (FrameType == REGULAR_FRAME) = false
	assert.False(t, fh.IsLast)
	// !allDefault && FrameType != LF_FRAME is false, so haveCrop = false
	assert.False(t, fh.haveCrop)
	// !allDefault && FrameType != LF_FRAME && !IsLast is false, so SaveAsReference = 0
	assert.Equal(t, uint32(0), fh.SaveAsReference)

	// Bounds adjusted for LfLevel=1: CeilDiv(size, 1<<(3*1)) = CeilDiv(size, 8)
	// CeilDiv(200, 8) = 25, CeilDiv(100, 8) = 13
	require.NotNil(t, fh.Bounds)
	assert.Equal(t, uint32(13), fh.Bounds.Size.Width)
	assert.Equal(t, uint32(25), fh.Bounds.Size.Height)
}

func TestNewFrameHeaderWithReader_ModularEncoding(t *testing.T) {
	parent := makeParent(100, 200, false, 0)

	reader := &testcommon.FakeBitReader{
		ReadBoolData: []bool{
			false, // allDefault = false
			false, // DoYCbCr = false
			false, // haveCrop = false
			true,  // IsLast = true
			true,  // RestorationFilter allDefault = true
		},
		ReadBitsData: []uint64{
			0, // FrameType = REGULAR_FRAME
			1, // Encoding = MODULAR
			0, // upsampling (1 << 0 = 1)
			2, // groupSizeShift = 2
		},
		ReadU32Data: []uint32{
			1, // numPasses = 1
			0, // BlendingInfo mode = BLEND_REPLACE
			0, // nameLen = 0
		},
	}

	fh, err := NewFrameHeaderWithReader(reader, parent)
	require.NoError(t, err)

	assert.Equal(t, uint32(MODULAR), fh.Encoding)
	assert.Equal(t, uint32(2), fh.groupSizeShift)
	assert.Equal(t, uint32(512), fh.groupDim)       // 128 << 2
	assert.Equal(t, uint32(4096), fh.lfGroupDim)     // 512 << 3
	assert.Equal(t, uint32(9), fh.logGroupDim)       // CeilLog2(512)
	assert.Equal(t, uint32(12), fh.logLFGroupDIM)    // CeilLog2(4096)
	assert.Equal(t, uint32(2), fh.xqmScale)          // not xyb, so default 2
	assert.Equal(t, uint32(2), fh.bqmScale)
}

func TestNewFrameHeaderWithReader_DoYCbCr(t *testing.T) {
	parent := makeParent(100, 200, false, 0)

	reader := &testcommon.FakeBitReader{
		ReadBoolData: []bool{
			false, // allDefault = false
			true,  // DoYCbCr = true
			false, // haveCrop = false
			true,  // IsLast = true
			true,  // RestorationFilter allDefault = true
		},
		ReadBitsData: []uint64{
			0, // FrameType = REGULAR_FRAME
			0, // Encoding = VARDCT
			1, // jpeg mode[0] = 1 -> Y=1, X=1
			2, // jpeg mode[1] = 2 -> Y=0, X=1
			3, // jpeg mode[2] = 3 -> Y=1, X=0
			0, // upsampling (1 << 0 = 1)
		},
		ReadU32Data: []uint32{
			1, // numPasses = 1
			0, // BlendingInfo mode = BLEND_REPLACE
			0, // nameLen = 0
		},
	}

	fh, err := NewFrameHeaderWithReader(reader, parent)
	require.NoError(t, err)

	assert.True(t, fh.DoYCbCr)

	// After inversion: maxJPY=1, maxJPX=1
	// jpegUpsamplingY = [1-1, 1-0, 1-1] = [0, 1, 0]
	// jpegUpsamplingX = [1-1, 1-1, 1-0] = [0, 0, 1]
	assert.Equal(t, []int32{0, 1, 0}, fh.jpegUpsamplingY)
	assert.Equal(t, []int32{0, 0, 1}, fh.jpegUpsamplingX)
}

func TestNewFrameHeaderWithReader_HaveCrop_SkipProgressive(t *testing.T) {
	parent := makeParent(100, 200, false, 0)

	reader := &testcommon.FakeBitReader{
		ReadBoolData: []bool{
			false, // allDefault = false
			false, // DoYCbCr = false
			true,  // haveCrop = true
			true,  // IsLast = true
			true,  // RestorationFilter allDefault = true
		},
		ReadBitsData: []uint64{
			3, // FrameType = SKIP_PROGRESSIVE
			0, // Encoding = VARDCT
			0, // upsampling (1 << 0 = 1)
			0, // BlendingInfo source (not fullFrame)
		},
		ReadU32Data: []uint32{
			1,  // numPasses = 1
			10, // x0 -> UnpackSigned(10) = 5
			20, // y0 -> UnpackSigned(20) = 10
			50, // width
			80, // height
			0,  // BlendingInfo mode = BLEND_REPLACE
			0,  // nameLen = 0
		},
	}

	fh, err := NewFrameHeaderWithReader(reader, parent)
	require.NoError(t, err)

	assert.Equal(t, uint32(SKIP_PROGRESSIVE), fh.FrameType)
	assert.True(t, fh.haveCrop)

	require.NotNil(t, fh.Bounds)
	assert.Equal(t, int32(5), fh.Bounds.Origin.X)
	assert.Equal(t, int32(10), fh.Bounds.Origin.Y)
	assert.Equal(t, uint32(50), fh.Bounds.Size.Width)
	assert.Equal(t, uint32(80), fh.Bounds.Size.Height)
	assert.True(t, fh.IsLast)
}

func TestNewFrameHeaderWithReader_ReferenceOnly(t *testing.T) {
	parent := makeParent(100, 200, false, 0)

	reader := &testcommon.FakeBitReader{
		ReadBoolData: []bool{
			false, // allDefault = false
			false, // DoYCbCr = false
			false, // haveCrop = false
			false, // SaveBeforeCT = false
			true,  // RestorationFilter allDefault = true
		},
		ReadBitsData: []uint64{
			2, // FrameType = REFERENCE_ONLY
			0, // Encoding = VARDCT
			0, // upsampling (1 << 0 = 1)
			1, // saveAsReference = 1
		},
		ReadU32Data: []uint32{
			0, // nameLen = 0
		},
	}

	fh, err := NewFrameHeaderWithReader(reader, parent)
	require.NoError(t, err)

	assert.Equal(t, uint32(REFERENCE_ONLY), fh.FrameType)
	// normalFrame = false, so IsLast = (FrameType == REGULAR_FRAME) = false
	assert.False(t, fh.IsLast)
	assert.Equal(t, uint32(1), fh.SaveAsReference)
	assert.False(t, fh.SaveBeforeCT)

	// passes should be default (not read from reader for REFERENCE_ONLY)
	assert.NotNil(t, fh.passes)
	assert.Equal(t, uint32(1), fh.passes.numPasses)
}

func TestNewFrameHeaderWithReader_SaveBeforeCT(t *testing.T) {
	parent := makeParent(100, 200, true, 0)

	reader := &testcommon.FakeBitReader{
		ReadBoolData: []bool{
			false, // allDefault = false
			false, // haveCrop = false
			false, // IsLast = false
			true,  // SaveBeforeCT = true
			true,  // RestorationFilter allDefault = true
		},
		ReadBitsData: []uint64{
			0, // FrameType = REGULAR_FRAME
			0, // Encoding = VARDCT
			0, // upsampling (1 << 0 = 1)
			3, // xqmScale
			2, // bqmScale
			0, // saveAsReference = 0
		},
		ReadU32Data: []uint32{
			1, // numPasses = 1
			0, // BlendingInfo mode = BLEND_REPLACE
			0, // nameLen = 0
		},
	}

	fh, err := NewFrameHeaderWithReader(reader, parent)
	require.NoError(t, err)

	assert.False(t, fh.IsLast)
	assert.Equal(t, uint32(0), fh.SaveAsReference)
	assert.True(t, fh.SaveBeforeCT)
}

func TestNewFrameHeaderWithReader_ExtraChannels(t *testing.T) {
	parent := makeParent(100, 200, true, 2)

	reader := &testcommon.FakeBitReader{
		ReadBoolData: []bool{
			false, // allDefault = false
			false, // haveCrop = false
			true,  // IsLast = true
			true,  // RestorationFilter allDefault = true
		},
		ReadBitsData: []uint64{
			0, // FrameType = REGULAR_FRAME
			0, // Encoding = VARDCT
			0, // upsampling (1 << 0 = 1)
			1, // ecUpsampling[0] (1 << 1 = 2)
			2, // ecUpsampling[1] (1 << 2 = 4)
			3, // xqmScale
			2, // bqmScale
		},
		ReadU32Data: []uint32{
			1, // numPasses = 1
			0, // BlendingInfo mode = BLEND_REPLACE (main)
			0, // EcBlendingInfo[0] mode = BLEND_REPLACE
			0, // EcBlendingInfo[1] mode = BLEND_REPLACE
			0, // nameLen = 0
		},
	}

	fh, err := NewFrameHeaderWithReader(reader, parent)
	require.NoError(t, err)

	assert.Equal(t, []uint32{2, 4}, fh.EcUpsampling)
	assert.Len(t, fh.EcBlendingInfo, 2)
	assert.Equal(t, uint32(BLEND_REPLACE), fh.EcBlendingInfo[0].Mode)
	assert.Equal(t, uint32(BLEND_REPLACE), fh.EcBlendingInfo[1].Mode)
}

func TestNewFrameHeaderWithReader_AllDefault_ExtraChannels(t *testing.T) {
	parent := makeParent(100, 200, true, 3)

	reader := &testcommon.FakeBitReader{
		ReadBoolData: []bool{true}, // allDefault = true
	}

	fh, err := NewFrameHeaderWithReader(reader, parent)
	require.NoError(t, err)

	// allDefault sets all EcUpsampling to 1
	assert.Equal(t, []uint32{1, 1, 1}, fh.EcUpsampling)
	// normalFrame=false, so EcBlendingInfo uses default BlendingInfo
	assert.Len(t, fh.EcBlendingInfo, 3)
	for i := 0; i < 3; i++ {
		assert.Equal(t, uint32(BLEND_REPLACE), fh.EcBlendingInfo[i].Mode)
	}
}

func TestNewFrameHeaderWithReader_LFFrame_HigherLfLevel(t *testing.T) {
	parent := makeParent(100, 200, true, 0)

	reader := &testcommon.FakeBitReader{
		ReadBoolData: []bool{
			false, // allDefault = false
			true,  // RestorationFilter allDefault = true
		},
		ReadBitsData: []uint64{
			1, // FrameType = LF_FRAME
			0, // Encoding = VARDCT
			0, // upsampling (1 << 0 = 1)
			3, // xqmScale
			2, // bqmScale
			2, // LfLevel (2 + 1 = 3)
		},
		ReadU32Data: []uint32{
			1, // numPasses = 1
			0, // nameLen = 0
		},
	}

	fh, err := NewFrameHeaderWithReader(reader, parent)
	require.NoError(t, err)

	assert.Equal(t, uint32(3), fh.LfLevel)
	// CeilDiv(200, 1<<(3*3)) = CeilDiv(200, 512) = 1
	// CeilDiv(100, 1<<(3*3)) = CeilDiv(100, 512) = 1
	assert.Equal(t, uint32(1), fh.Bounds.Size.Width)
	assert.Equal(t, uint32(1), fh.Bounds.Size.Height)
}

func TestNewFrameHeaderWithReader_GroupDimDefaults(t *testing.T) {
	tests := []struct {
		name           string
		groupSizeShift uint32
		expectedDim    uint32
		expectedLFDim  uint32
		expectedLogGD  uint32
		expectedLogLF  uint32
	}{
		{"shift 0", 0, 128, 1024, 7, 10},
		{"shift 1", 1, 256, 2048, 8, 11},
		{"shift 2", 2, 512, 4096, 9, 12},
		{"shift 3", 3, 1024, 8192, 10, 13},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			parent := makeParent(100, 200, false, 0)
			reader := &testcommon.FakeBitReader{
				ReadBoolData: []bool{
					false, // allDefault = false
					false, // DoYCbCr = false
					false, // haveCrop = false
					true,  // IsLast = true
					true,  // RestorationFilter allDefault = true
				},
				ReadBitsData: []uint64{
					0,                      // FrameType = REGULAR_FRAME
					1,                      // Encoding = MODULAR
					0,                      // upsampling
					uint64(tc.groupSizeShift), // groupSizeShift
				},
				ReadU32Data: []uint32{
					1, // numPasses = 1
					0, // BlendingInfo mode
					0, // nameLen
				},
			}

			fh, err := NewFrameHeaderWithReader(reader, parent)
			require.NoError(t, err)

			assert.Equal(t, tc.expectedDim, fh.groupDim)
			assert.Equal(t, tc.expectedLFDim, fh.lfGroupDim)
			assert.Equal(t, tc.expectedLogGD, fh.logGroupDim)
			assert.Equal(t, tc.expectedLogLF, fh.logLFGroupDIM)
		})
	}
}

func TestNewFrameHeaderWithReader_UpsamplingValues(t *testing.T) {
	tests := []struct {
		name             string
		upsamplingBits   uint64
		expectedUpsampling uint32
	}{
		{"1x", 0, 1},
		{"2x", 1, 2},
		{"4x", 2, 4},
		{"8x", 3, 8},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			parent := makeParent(800, 600, true, 0)
			reader := &testcommon.FakeBitReader{
				ReadBoolData: []bool{
					false, // allDefault = false
					false, // haveCrop = false
					true,  // IsLast = true
					true,  // RestorationFilter allDefault = true
				},
				ReadBitsData: []uint64{
					0,                    // FrameType = REGULAR_FRAME
					0,                    // Encoding = VARDCT
					tc.upsamplingBits,    // upsampling
					3,                    // xqmScale
					2,                    // bqmScale
				},
				ReadU32Data: []uint32{
					1, // numPasses = 1
					0, // BlendingInfo mode
					0, // nameLen
				},
			}

			fh, err := NewFrameHeaderWithReader(reader, parent)
			require.NoError(t, err)

			assert.Equal(t, tc.expectedUpsampling, fh.Upsampling)
		})
	}
}

func TestNewFrameHeaderWithReader_JpegUpsamplingModes(t *testing.T) {
	tests := []struct {
		name       string
		modes      []uint64
		expectedY  []int32
		expectedX  []int32
	}{
		{
			"all mode 0 (no subsampling)",
			[]uint64{0, 0, 0},
			[]int32{0, 0, 0},
			[]int32{0, 0, 0},
		},
		{
			"all mode 1 (4:2:0 all channels)",
			[]uint64{1, 1, 1},
			[]int32{0, 0, 0},
			[]int32{0, 0, 0},
		},
		{
			"mixed modes 1,2,3",
			[]uint64{1, 2, 3},
			[]int32{0, 1, 0},
			[]int32{0, 0, 1},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			parent := makeParent(100, 200, false, 0)
			reader := &testcommon.FakeBitReader{
				ReadBoolData: []bool{
					false, // allDefault = false
					true,  // DoYCbCr = true
					false, // haveCrop = false
					true,  // IsLast = true
					true,  // RestorationFilter allDefault = true
				},
				ReadBitsData: append([]uint64{
					0, // FrameType = REGULAR_FRAME
					0, // Encoding = VARDCT
				}, append(tc.modes, 0)...), // jpeg modes + upsampling
				ReadU32Data: []uint32{
					1, // numPasses = 1
					0, // BlendingInfo mode
					0, // nameLen
				},
			}

			fh, err := NewFrameHeaderWithReader(reader, parent)
			require.NoError(t, err)

			assert.True(t, fh.DoYCbCr)
			assert.Equal(t, tc.expectedY, fh.jpegUpsamplingY)
			assert.Equal(t, tc.expectedX, fh.jpegUpsamplingX)
		})
	}
}

func TestDisplayDebug_NilHeader(t *testing.T) {
	var fh *FrameHeader
	// Should not panic
	fh.DisplayDebug()
}

func TestDisplayDebug_ValidHeader(t *testing.T) {
	fh := &FrameHeader{
		FrameType:       REGULAR_FRAME,
		Encoding:        VARDCT,
		Flags:           NOISE | USE_LF_FRAME,
		DoYCbCr:         false,
		IsLast:          true,
		Width:           100,
		Height:          200,
		Upsampling:      1,
		LfLevel:         0,
		groupDim:        256,
		lfGroupDim:      2048,
		logGroupDim:     8,
		logLFGroupDIM:   11,
		jpegUpsamplingX: []int32{0, 0, 0},
		jpegUpsamplingY: []int32{0, 0, 0},
		EcUpsampling:    []uint32{},
		Bounds: &util.Rectangle{
			Origin: util.Point{X: 0, Y: 0},
			Size:   util.Dimension{Width: 100, Height: 200},
		},
		BlendingInfo: &BlendingInfo{Mode: BLEND_REPLACE},
		passes:       NewPassesInfo(),
		restorationFilter: NewRestorationFilter(),
		extensions:   bundle.NewExtensions(),
	}

	// Should not panic
	fh.DisplayDebug()
}

func TestDisplayDebug_NilFields(t *testing.T) {
	fh := &FrameHeader{
		jpegUpsamplingX: []int32{0, 0, 0},
		jpegUpsamplingY: []int32{0, 0, 0},
		EcUpsampling:    []uint32{},
	}

	// Should not panic even with nil Bounds, BlendingInfo, passes, etc.
	fh.DisplayDebug()
}
