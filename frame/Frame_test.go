package frame

import (
	"bytes"
	"testing"

	"github.com/kpfaulkner/jxl-go/bundle"
	"github.com/kpfaulkner/jxl-go/colour"
	"github.com/kpfaulkner/jxl-go/image"
	"github.com/kpfaulkner/jxl-go/jxlio"
	"github.com/kpfaulkner/jxl-go/options"
	"github.com/kpfaulkner/jxl-go/testcommon"
	"github.com/kpfaulkner/jxl-go/util"
)

func TestCtxFunc(t *testing.T) {
	if got := ctxFunc(0); got != 0 {
		t.Errorf("ctxFunc(0) = %d; want 0", got)
	}
	if got := ctxFunc(1); got != 1 {
		t.Errorf("ctxFunc(1) = %d; want 1", got)
	}
	if got := ctxFunc(1000); got < 0 {
		t.Errorf("ctxFunc(1000) = %d; want non-negative", got)
	}
}

func TestAbsFloat32(t *testing.T) {
	if got := absFloat32(-3.5); got != 3.5 {
		t.Errorf("absFloat32(-3.5) = %f; want 3.5", got)
	}
	if got := absFloat32(2.25); got != 2.25 {
		t.Errorf("absFloat32(2.25) = %f; want 2.25", got)
	}
}

func TestMirrorCoord(t *testing.T) {
	// inside range
	if got := mirrorCoord(2, 5); got != 2 {
		t.Errorf("mirrorCoord(2,5) = %d; want 2", got)
	}
	// negative -> mirrored
	if got := mirrorCoord(-1, 4); got < 0 || got >= 4 {
		t.Errorf("mirrorCoord(-1,4) = %d; want in [0,3]", got)
	}
	// beyond -> mirrored
	if got := mirrorCoord(10, 6); got < 0 || got >= 6 {
		t.Errorf("mirrorCoord(10,6) = %d; want in [0,5]", got)
	}
}

func TestLocationFunctions(t *testing.T) {
	f := &Frame{}
	f.lfGroupRowStride = 4
	f.groupRowStride = 8

	lfLoc := f.getLFGroupLocation(5)
	if lfLoc == nil {
		t.Fatal("getLFGroupLocation returned nil")
	}
	gLoc := f.getGroupLocation(10)
	if gLoc == nil {
		t.Fatal("getGroupLocation returned nil")
	}

	// groupPosInLFGroup should compute difference
	f.lfGroups = make([]*LFGroup, 0)
	p := f.groupPosInLFGroup(int32(lfLoc.Y), uint32(10))
	_ = p // just ensure it runs
}

func TestGetColourChannelCountAndIsVisible(t *testing.T) {
	f := &Frame{}
	f.GlobalMetadata = &bundle.ImageHeader{}
	f.Header = &FrameHeader{}

	// initialize ColourEncoding to avoid nil deref
	f.GlobalMetadata.ColourEncoding = &colour.ColourEncodingBundle{ColourEncoding: colour.CE_RGB}

	// default: not XYB and not VARDCT
	f.GlobalMetadata.XybEncoded = false
	f.Header.Encoding = MODULAR
	if got := f.GetColourChannelCount(); got != f.GlobalMetadata.GetColourChannelCount() {
		t.Errorf("GetColourChannelCount mismatch: got %d", got)
	}

	f.GlobalMetadata.XybEncoded = true
	if got := f.GetColourChannelCount(); got != 3 {
		t.Errorf("GetColourChannelCount XYB = %d; want 3", got)
	}

	// IsVisible tests
	f.Header.FrameType = REGULAR_FRAME
	f.Header.Duration = 0
	f.Header.IsLast = false
	if !f.IsVisible() {
		t.Errorf("IsVisible expected true for REGULAR_FRAME")
	}
	f.Header.FrameType = SKIP_PROGRESSIVE
	f.Header.Duration = 1
	if !f.IsVisible() {
		t.Errorf("IsVisible expected true for SKIP_PROGRESSIVE with duration")
	}
}

func TestGetPaddedFrameSize(t *testing.T) {
	f := &Frame{}
	f.Header = &FrameHeader{}
	f.GlobalMetadata = &bundle.ImageHeader{}

	// Test for MODULAR (non-VARDCT)
	f.Header.Encoding = MODULAR
	f.Header.jpegUpsamplingX = []int32{0, 0, 0}
	f.Header.jpegUpsamplingY = []int32{0, 0, 0}
	f.Header.Bounds = &util.Rectangle{Origin: util.Point{}, Size: util.Dimension{Width: 100, Height: 50}}
	f.Header.Upsampling = 1
	padded, err := f.GetPaddedFrameSize()
	if err != nil {
		t.Fatalf("GetPaddedFrameSize returned error: %v", err)
	}
	if padded.Width == 0 || padded.Height == 0 {
		t.Fatalf("unexpected padded size: %+v", padded)
	}

	// Test for VARDCT - should round up to multiple of 8
	f.Header.Encoding = VARDCT
	f.Header.Bounds = &util.Rectangle{Origin: util.Point{}, Size: util.Dimension{Width: 17, Height: 9}}
	padded, err = f.GetPaddedFrameSize()
	if err != nil {
		t.Fatalf("GetPaddedFrameSize returned error for VARDCT: %v", err)
	}
	if padded.Width%8 != 0 || padded.Height%8 != 0 {
		t.Fatalf("VARDCT padded not multiple of 8: %+v", padded)
	}
}

func TestCopyFloatBuffersAndGenerateSignatures(t *testing.T) {
	// build simple image buffers
	ib, err := image.NewImageBuffer(image.TYPE_FLOAT, 2, 3)
	if err != nil {
		t.Fatalf("NewImageBuffer failed: %v", err)
	}
	ib.FloatBuffer[0] = []float32{1, 2, 3}
	ib.FloatBuffer[1] = []float32{4, 5, 6}
	f := &Frame{}
	f.Buffer = []image.ImageBuffer{*ib}

	copied := copyFloatBuffers(f.Buffer, 1)
	if len(copied) != 1 {
		t.Fatalf("copyFloatBuffers returned wrong length")
	}
	// modify original and check copy unaffected
	ib.FloatBuffer[0][0] = 99
	if copied[0][0][0] == 99 {
		t.Fatalf("copyFloatBuffers did not produce independent copy")
	}

	// Test generateSignaturesForBuffer
	f.Buffer = []image.ImageBuffer{*ib}
	sigs := f.generateSignaturesForBuffer(0)
	if len(sigs) != int(f.Buffer[0].Height) {
		t.Fatalf("generateSignaturesForBuffer length mismatch: got %d want %d", len(sigs), f.Buffer[0].Height)
	}
}

func TestGetGroupAndLFGroupSize(t *testing.T) {
	f := &Frame{}
	f.Header = &FrameHeader{groupDim: 8, lfGroupDim: 64}
	f.Header.Bounds = &util.Rectangle{Origin: util.Point{}, Size: util.Dimension{Width: 100, Height: 50}}
	f.groupRowStride = util.CeilDiv(f.Header.Bounds.Size.Width, f.Header.groupDim)
	f.lfGroupRowStride = util.CeilDiv(f.Header.Bounds.Size.Width, f.Header.groupDim<<3)
	f.numLFGroups = f.lfGroupRowStride * util.CeilDiv(f.Header.Bounds.Size.Height, f.Header.groupDim<<3)

	sz, err := f.getGroupSize(0)
	if err != nil {
		t.Fatalf("getGroupSize error: %v", err)
	}
	if sz.Width == 0 || sz.Height == 0 {
		t.Fatalf("getGroupSize returned zero size: %+v", sz)
	}

	lfsz, err := f.getLFGroupSize(0)
	if err != nil {
		t.Fatalf("getLFGroupSize error: %v", err)
	}
	if lfsz.Width == 0 || lfsz.Height == 0 {
		t.Fatalf("getLFGroupSize returned zero size: %+v", lfsz)
	}
}

func TestGetNumLFGroups(t *testing.T) {
	f := &Frame{numLFGroups: 7}
	if got := f.getNumLFGroups(); got != 7 {
		t.Fatalf("getNumLFGroups = %d; want 7", got)
	}
}

func TestUpsamplePerformUpsampling(t *testing.T) {
	// Basic invocation to ensure no panic; extensive correctness is covered elsewhere
	f := &Frame{}
	f.Header = &FrameHeader{Upsampling: 2}
	// initialize BitDepth and upsample weights to avoid nil deref
	f.GlobalMetadata = &bundle.ImageHeader{BitDepth: &bundle.BitDepthHeader{BitsPerSample: 8}}
	// set default up weights so GetUpWeights succeeds
	f.GlobalMetadata.Up2Weights = bundle.DEFAULT_UP2
	f.GlobalMetadata.Up4Weights = bundle.DEFAULT_UP4
	f.GlobalMetadata.Up8Weights = bundle.DEFAULT_UP8
	f.Buffer = make([]image.ImageBuffer, 1)
	ib, err := image.NewImageBuffer(image.TYPE_FLOAT, 2, 2)
	if err != nil {
		t.Fatalf("NewImageBuffer failed: %v", err)
	}
	ib.FloatBuffer[0] = []float32{1, 2}
	ib.FloatBuffer[1] = []float32{3, 4}
	f.Buffer[0] = *ib
	// call performUpsampling directly
	if _, err := f.performUpsampling(f.Buffer[0], 0); err != nil {
		t.Fatalf("performUpsampling error: %v", err)
	}
}

// TestGettersSetters tests the getter/setter methods
func TestGettersSetters(t *testing.T) {
	f := &Frame{}

	// Test getGlobalTree and setGlobalTree
	tree := &MATreeNode{}
	f.setGlobalTree(tree)
	if got := f.getGlobalTree(); got != tree {
		t.Errorf("getGlobalTree returned wrong tree")
	}

	// Test getGlobalMetadata
	meta := &bundle.ImageHeader{}
	f.GlobalMetadata = meta
	if got := f.getGlobalMetadata(); got != meta {
		t.Errorf("getGlobalMetadata returned wrong metadata")
	}

	// Test getPasses
	passes := []Pass{{}, {}}
	f.passes = passes
	if got := f.getPasses(); len(got) != 2 {
		t.Errorf("getPasses returned wrong length: got %d want 2", len(got))
	}

	// Test getFrameHeader
	header := &FrameHeader{}
	f.Header = header
	if got := f.getFrameHeader(); got != header {
		t.Errorf("getFrameHeader returned wrong header")
	}

	// Test getHFGlobal
	hfg := &HFGlobal{}
	f.hfGlobal = hfg
	if got := f.getHFGlobal(); got != hfg {
		t.Errorf("getHFGlobal returned wrong hfGlobal")
	}

	// Test getLFGlobal
	lfg := &LFGlobal{}
	f.LfGlobal = lfg
	if got := f.getLFGlobal(); got != lfg {
		t.Errorf("getLFGlobal returned wrong lfGlobal")
	}
}

// TestNewFrameWithReader tests the Frame constructor
func TestNewFrameWithReader(t *testing.T) {
	data := []byte{0x00, 0x01, 0x02, 0x03}
	reader := jxlio.NewBitStreamReader(bytes.NewReader(data))
	imageHeader := &bundle.ImageHeader{}
	opts := options.NewJXLOptions(nil)

	frame := NewFrameWithReader(reader, imageHeader, opts)

	if frame == nil {
		t.Fatal("NewFrameWithReader returned nil")
	}
	if frame.GlobalMetadata != imageHeader {
		t.Error("GlobalMetadata not set correctly")
	}
	if frame.options != opts {
		t.Error("options not set correctly")
	}
	if frame.reader == nil {
		t.Error("reader not set")
	}
}

// TestSkipFrameData tests skipping frame data
func TestSkipFrameData(t *testing.T) {
	// Create data with enough bytes
	data := make([]byte, 100)
	for i := range data {
		data[i] = byte(i)
	}
	reader := jxlio.NewBitStreamReader(bytes.NewReader(data))

	f := &Frame{
		reader:     reader,
		tocLengths: []uint32{10, 20, 30},
	}

	err := f.SkipFrameData()
	if err != nil {
		t.Fatalf("SkipFrameData returned error: %v", err)
	}
}

// TestGetBitreader tests the getBitreader method
func TestGetBitreader(t *testing.T) {
	data := []byte{0x00, 0x01, 0x02, 0x03}
	reader := jxlio.NewBitStreamReader(bytes.NewReader(data))

	// Test with single TOC entry
	f := &Frame{
		tocLengths: []uint32{10},
		bitreaders: []jxlio.BitReader{reader},
	}

	br, err := f.getBitreader(0)
	if err != nil {
		t.Fatalf("getBitreader(0) returned error: %v", err)
	}
	if br != reader {
		t.Error("getBitreader returned wrong reader for single entry")
	}

	// Test with multiple TOC entries
	reader2 := jxlio.NewBitStreamReader(bytes.NewReader(data))
	f2 := &Frame{
		tocLengths: []uint32{10, 20},
		bitreaders: []jxlio.BitReader{reader, reader2},
	}

	br1, err := f2.getBitreader(0)
	if err != nil {
		t.Fatalf("getBitreader(0) returned error: %v", err)
	}
	if br1 != reader {
		t.Error("getBitreader(0) returned wrong reader")
	}

	br2, err := f2.getBitreader(1)
	if err != nil {
		t.Fatalf("getBitreader(1) returned error: %v", err)
	}
	if br2 != reader2 {
		t.Error("getBitreader(1) returned wrong reader")
	}
}

// TestNotImplementedFunctions tests functions that return "not implemented" errors
func TestNotImplementedFunctions(t *testing.T) {
	f := &Frame{
		LfGlobal: &LFGlobal{},
	}

	// Test InitializeNoise with empty parameters (should return nil)
	err := f.InitializeNoise(0)
	if err != nil {
		t.Errorf("InitializeNoise with empty params should return nil, got: %v", err)
	}

	// Test InitializeNoise with noise parameters (should return error)
	f.LfGlobal.noiseParameters = []NoiseParameters{{}}
	err = f.InitializeNoise(0)
	if err == nil {
		t.Error("InitializeNoise with noise params should return error")
	}

	// Test RenderSplines without splines (should return nil)
	f.LfGlobal.splines = nil
	err = f.RenderSplines()
	if err != nil {
		t.Errorf("RenderSplines with nil splines should return nil, got: %v", err)
	}

	// Test RenderSplines with splines (should return error)
	f.LfGlobal.splines = []SplinesBundle{{}}
	err = f.RenderSplines()
	if err == nil {
		t.Error("RenderSplines with splines should return error")
	}

	// Test SynthesizeNoise without noise params (should return nil)
	f.LfGlobal.noiseParameters = nil
	err = f.SynthesizeNoise()
	if err != nil {
		t.Errorf("SynthesizeNoise with nil params should return nil, got: %v", err)
	}

	// Test SynthesizeNoise with noise params (should return error)
	f.LfGlobal.noiseParameters = []NoiseParameters{{}}
	err = f.SynthesizeNoise()
	if err == nil {
		t.Error("SynthesizeNoise with params should return error")
	}
}

// TestGetColourChannelCountInt32 tests the int32 version of getColourChannelCount
func TestGetColourChannelCountInt32(t *testing.T) {
	f := &Frame{
		GlobalMetadata: &bundle.ImageHeader{
			ColourEncoding: &colour.ColourEncodingBundle{ColourEncoding: colour.CE_RGB},
		},
		Header: &FrameHeader{
			Encoding: MODULAR,
		},
	}

	// Not XYB and not VARDCT
	f.GlobalMetadata.XybEncoded = false
	got := f.getColourChannelCount()
	if got != int32(f.GlobalMetadata.GetColourChannelCount()) {
		t.Errorf("getColourChannelCount() = %d; want %d", got, f.GlobalMetadata.GetColourChannelCount())
	}

	// XYB encoded
	f.GlobalMetadata.XybEncoded = true
	got = f.getColourChannelCount()
	if got != 3 {
		t.Errorf("getColourChannelCount() XYB = %d; want 3", got)
	}

	// VARDCT
	f.GlobalMetadata.XybEncoded = false
	f.Header.Encoding = VARDCT
	got = f.getColourChannelCount()
	if got != 3 {
		t.Errorf("getColourChannelCount() VARDCT = %d; want 3", got)
	}
}

// TestGetLFGroupForGroup tests the getLFGroupForGroup method
func TestGetLFGroupForGroup(t *testing.T) {
	lfg := &LFGroup{}

	f := &Frame{
		groupRowStride:   8,
		lfGroupRowStride: 1,
		lfGroups:         []*LFGroup{lfg},
	}

	got := f.getLFGroupForGroup(0)
	if got != lfg {
		t.Error("getLFGroupForGroup returned wrong LFGroup")
	}
}

// TestInvertSubsampling tests the invertSubsampling method
func TestInvertSubsampling(t *testing.T) {
	f := &Frame{
		GlobalMetadata: &bundle.ImageHeader{
			BitDepth: &bundle.BitDepthHeader{BitsPerSample: 8},
		},
		Header: &FrameHeader{
			jpegUpsamplingX: []int32{0, 0, 0},
			jpegUpsamplingY: []int32{0, 0, 0},
		},
	}

	// Create 3 float buffers
	f.Buffer = make([]image.ImageBuffer, 3)
	for c := 0; c < 3; c++ {
		ib, err := image.NewImageBuffer(image.TYPE_FLOAT, 4, 4)
		if err != nil {
			t.Fatalf("NewImageBuffer failed: %v", err)
		}
		for y := 0; y < 4; y++ {
			for x := 0; x < 4; x++ {
				ib.FloatBuffer[y][x] = float32(y*4 + x)
			}
		}
		f.Buffer[c] = *ib
	}

	// Test with no upsampling
	err := f.invertSubsampling()
	if err != nil {
		t.Fatalf("invertSubsampling returned error: %v", err)
	}

	// Test with X upsampling
	f.Header.jpegUpsamplingX = []int32{1, 0, 0}
	f.Header.jpegUpsamplingY = []int32{0, 0, 0}
	// Re-create buffer for channel 0 with half width
	ib, err := image.NewImageBuffer(image.TYPE_FLOAT, 4, 2)
	if err != nil {
		t.Fatalf("NewImageBuffer failed: %v", err)
	}
	for y := 0; y < 4; y++ {
		for x := 0; x < 2; x++ {
			ib.FloatBuffer[y][x] = float32(y*2 + x)
		}
	}
	f.Buffer[0] = *ib

	err = f.invertSubsampling()
	if err != nil {
		t.Fatalf("invertSubsampling with X upsampling returned error: %v", err)
	}

	// Check that channel 0 was upsampled
	if f.Buffer[0].Width != 4 {
		t.Errorf("After X upsampling, width = %d; want 4", f.Buffer[0].Width)
	}
}

// TestInvertSubsamplingY tests Y-axis subsampling inversion
func TestInvertSubsamplingY(t *testing.T) {
	f := &Frame{
		GlobalMetadata: &bundle.ImageHeader{
			BitDepth: &bundle.BitDepthHeader{BitsPerSample: 8},
		},
		Header: &FrameHeader{
			jpegUpsamplingX: []int32{0, 0, 0},
			jpegUpsamplingY: []int32{1, 0, 0},
		},
	}

	f.Buffer = make([]image.ImageBuffer, 3)
	// Channel 0 has half height
	ib0, _ := image.NewImageBuffer(image.TYPE_FLOAT, 2, 4)
	for y := 0; y < 2; y++ {
		for x := 0; x < 4; x++ {
			ib0.FloatBuffer[y][x] = float32(y*4 + x)
		}
	}
	f.Buffer[0] = *ib0

	// Other channels normal
	for c := 1; c < 3; c++ {
		ib, _ := image.NewImageBuffer(image.TYPE_FLOAT, 4, 4)
		for y := 0; y < 4; y++ {
			for x := 0; x < 4; x++ {
				ib.FloatBuffer[y][x] = float32(y*4 + x)
			}
		}
		f.Buffer[c] = *ib
	}

	err := f.invertSubsampling()
	if err != nil {
		t.Fatalf("invertSubsampling with Y upsampling returned error: %v", err)
	}

	if f.Buffer[0].Height != 4 {
		t.Errorf("After Y upsampling, height = %d; want 4", f.Buffer[0].Height)
	}
}

// TestPerformGabConvolution tests the Gab convolution filter
func TestPerformGabConvolution(t *testing.T) {
	f := &Frame{
		GlobalMetadata: &bundle.ImageHeader{
			BitDepth:       &bundle.BitDepthHeader{BitsPerSample: 8},
			XybEncoded:     false,
			ColourEncoding: &colour.ColourEncodingBundle{ColourEncoding: colour.CE_RGB},
		},
		Header: &FrameHeader{
			Encoding:          MODULAR,
			restorationFilter: NewRestorationFilter(),
		},
		options: &options.JXLOptions{MaxGoroutines: 1},
	}

	// Create 3 float buffers
	f.Buffer = make([]image.ImageBuffer, 3)
	for c := 0; c < 3; c++ {
		ib, err := image.NewImageBuffer(image.TYPE_FLOAT, 8, 8)
		if err != nil {
			t.Fatalf("NewImageBuffer failed: %v", err)
		}
		for y := 0; y < 8; y++ {
			for x := 0; x < 8; x++ {
				ib.FloatBuffer[y][x] = float32(y*8+x) / 64.0
			}
		}
		f.Buffer[c] = *ib
	}

	err := f.performGabConvolution()
	if err != nil {
		t.Fatalf("performGabConvolution returned error: %v", err)
	}

	// Verify buffer sizes are preserved
	for c := 0; c < 3; c++ {
		if f.Buffer[c].Width != 8 || f.Buffer[c].Height != 8 {
			t.Errorf("Channel %d size changed: %dx%d", c, f.Buffer[c].Width, f.Buffer[c].Height)
		}
	}
}

// TestDisplayBuffers tests the display buffer functions (for coverage)
func TestDisplayBuffers(t *testing.T) {
	frameBuffer := [][][]float32{
		{{1.0, 2.0}, {3.0, 4.0}},
		{{5.0, 6.0}, {7.0, 8.0}},
	}

	// Just ensure it doesn't panic
	displayBuffers("test", frameBuffer)

	singleBuffer := [][]float32{{1.0, 2.0}, {3.0, 4.0}}
	displayBuffer("test", singleBuffer)
}

// TestUpsampleFull tests the full Upsample method
func TestUpsampleFull(t *testing.T) {
	f := &Frame{
		GlobalMetadata: &bundle.ImageHeader{
			BitDepth:         &bundle.BitDepthHeader{BitsPerSample: 8},
			ExtraChannelInfo: []bundle.ExtraChannelInfo{},
			ColourEncoding:   &colour.ColourEncodingBundle{ColourEncoding: colour.CE_RGB},
			Up2Weights:       bundle.DEFAULT_UP2,
			Up4Weights:       bundle.DEFAULT_UP4,
			Up8Weights:       bundle.DEFAULT_UP8,
		},
		Header: &FrameHeader{
			Upsampling:   2,
			EcUpsampling: []uint32{},
			Bounds:       &util.Rectangle{Origin: util.Point{X: 0, Y: 0}, Size: util.Dimension{Width: 4, Height: 4}},
			groupDim:     8,
		},
	}

	f.Buffer = make([]image.ImageBuffer, 1)
	ib, err := image.NewImageBuffer(image.TYPE_FLOAT, 4, 4)
	if err != nil {
		t.Fatalf("NewImageBuffer failed: %v", err)
	}
	for y := 0; y < 4; y++ {
		for x := 0; x < 4; x++ {
			ib.FloatBuffer[y][x] = float32(y*4+x) / 16.0
		}
	}
	f.Buffer[0] = *ib

	err = f.Upsample()
	if err != nil {
		t.Fatalf("Upsample returned error: %v", err)
	}

	// Check dimensions were doubled
	if f.Header.Bounds.Size.Width != 8 || f.Header.Bounds.Size.Height != 8 {
		t.Errorf("After Upsample, size = %dx%d; want 8x8", f.Header.Bounds.Size.Width, f.Header.Bounds.Size.Height)
	}
}

// TestReadPermutation tests the readPermutation function
func TestReadPermutation(t *testing.T) {
	// Create a fake bit reader with minimal data
	data := make([]byte, 100)
	reader := jxlio.NewBitStreamReader(bytes.NewReader(data))

	// Create a fake entropy streamer that returns controlled values
	// The first symbol determines the end value (how many lehmer codes to read)
	// Then lehmer codes follow
	stream := &FakeEntropyStreamer{
		FakeSymbols: []int32{
			2, // end = 2 (read 2 lehmer codes)
			0, // lehmer[0] = 0
			1, // lehmer[1] = 1
		},
	}

	perm, err := readPermutation(reader, stream, 4, 0)
	if err != nil {
		t.Fatalf("readPermutation returned error: %v", err)
	}

	if len(perm) != 4 {
		t.Errorf("Permutation length = %d; want 4", len(perm))
	}
}

// TestReadPermutationWithSkip tests readPermutation with skip parameter
func TestReadPermutationWithSkip(t *testing.T) {
	data := make([]byte, 100)
	reader := jxlio.NewBitStreamReader(bytes.NewReader(data))

	stream := &FakeEntropyStreamer{
		FakeSymbols: []int32{
			1, // end = 1
			0, // lehmer[1] = 0 (skip=1, so we start at index 1)
		},
	}

	perm, err := readPermutation(reader, stream, 4, 1)
	if err != nil {
		t.Fatalf("readPermutation with skip returned error: %v", err)
	}

	if len(perm) != 4 {
		t.Errorf("Permutation length = %d; want 4", len(perm))
	}
}

// TestReadPermutationIllegalEnd tests readPermutation with illegal end value
func TestReadPermutationIllegalEnd(t *testing.T) {
	data := make([]byte, 100)
	reader := jxlio.NewBitStreamReader(bytes.NewReader(data))

	stream := &FakeEntropyStreamer{
		FakeSymbols: []int32{
			10, // end = 10 (illegal for size=4, skip=0)
		},
	}

	_, err := readPermutation(reader, stream, 4, 0)
	if err == nil {
		t.Error("readPermutation should return error for illegal end value")
	}
}

// TestCtxFuncEdgeCases tests ctxFunc edge cases
func TestCtxFuncEdgeCases(t *testing.T) {
	// Test with larger values to hit the min(7, ...) case
	if got := ctxFunc(100000); got != 7 {
		t.Errorf("ctxFunc(100000) = %d; want 7 (capped)", got)
	}

	// Test with various values
	testCases := []struct {
		input    int64
		expected int
	}{
		{0, 0},
		{1, 1},
		{2, 2},
		{3, 2},
		{7, 3},
		{15, 4},
		{127, 7},
		{255, 7},
	}

	for _, tc := range testCases {
		got := ctxFunc(tc.input)
		if got != tc.expected {
			t.Errorf("ctxFunc(%d) = %d; want %d", tc.input, got, tc.expected)
		}
	}
}

// TestAbsFloat32EdgeCases tests absFloat32 edge cases
func TestAbsFloat32EdgeCases(t *testing.T) {
	// Test zero
	if got := absFloat32(0.0); got != 0.0 {
		t.Errorf("absFloat32(0.0) = %f; want 0.0", got)
	}

	// Test negative zero
	if got := absFloat32(-0.0); got != 0.0 {
		t.Errorf("absFloat32(-0.0) = %f; want 0.0", got)
	}

	// Test very small numbers
	if got := absFloat32(-1e-38); got != 1e-38 {
		t.Errorf("absFloat32(-1e-38) = %f; want 1e-38", got)
	}

	// Test large numbers
	if got := absFloat32(-1e38); got != 1e38 {
		t.Errorf("absFloat32(-1e38) = %f; want 1e38", got)
	}
}

// TestMirrorCoordEdgeCases tests mirrorCoord edge cases
func TestMirrorCoordEdgeCases(t *testing.T) {
	// Test exactly at boundary
	if got := mirrorCoord(0, 10); got != 0 {
		t.Errorf("mirrorCoord(0, 10) = %d; want 0", got)
	}

	// Test at size-1
	if got := mirrorCoord(9, 10); got != 9 {
		t.Errorf("mirrorCoord(9, 10) = %d; want 9", got)
	}

	// Test at size (should mirror)
	if got := mirrorCoord(10, 10); got < 0 || got >= 10 {
		t.Errorf("mirrorCoord(10, 10) = %d; want in [0, 9]", got)
	}

	// Test large negative
	if got := mirrorCoord(-5, 10); got < 0 || got >= 10 {
		t.Errorf("mirrorCoord(-5, 10) = %d; want in [0, 9]", got)
	}
}

// TestPerformUpsamplingNoUpsampling tests when k=1 (no upsampling)
func TestPerformUpsamplingNoUpsampling(t *testing.T) {
	f := &Frame{
		GlobalMetadata: &bundle.ImageHeader{
			BitDepth:         &bundle.BitDepthHeader{BitsPerSample: 8},
			ColourEncoding:   &colour.ColourEncodingBundle{ColourEncoding: colour.CE_RGB},
			ExtraChannelInfo: []bundle.ExtraChannelInfo{},
			Up2Weights:       bundle.DEFAULT_UP2,
		},
		Header: &FrameHeader{
			Upsampling:   1, // No upsampling
			EcUpsampling: []uint32{},
		},
	}

	ib, err := image.NewImageBuffer(image.TYPE_FLOAT, 4, 4)
	if err != nil {
		t.Fatalf("NewImageBuffer failed: %v", err)
	}

	result, err := f.performUpsampling(*ib, 0)
	if err != nil {
		t.Fatalf("performUpsampling error: %v", err)
	}

	// With k=1, the function returns a pointer to the input buffer
	// Just verify result is valid and has same dimensions
	if result == nil {
		t.Error("performUpsampling result should not be nil")
	}
	if result.Width != ib.Width || result.Height != ib.Height {
		t.Errorf("performUpsampling with k=1 should preserve size: got %dx%d, want %dx%d",
			result.Width, result.Height, ib.Width, ib.Height)
	}
}

// TestGetPaddedFrameSizeVARDCT tests GetPaddedFrameSize with VARDCT encoding
func TestGetPaddedFrameSizeVARDCT(t *testing.T) {
	f := &Frame{
		Header: &FrameHeader{
			Encoding:        VARDCT,
			jpegUpsamplingX: []int32{0, 0, 0},
			jpegUpsamplingY: []int32{0, 0, 0},
			Bounds: &util.Rectangle{
				Origin: util.Point{},
				Size:   util.Dimension{Width: 17, Height: 9},
			},
		},
	}

	padded, err := f.GetPaddedFrameSize()
	if err != nil {
		t.Fatalf("GetPaddedFrameSize returned error: %v", err)
	}

	// Should be rounded up to multiple of 8
	if padded.Width%8 != 0 {
		t.Errorf("VARDCT Width not multiple of 8: %d", padded.Width)
	}
	if padded.Height%8 != 0 {
		t.Errorf("VARDCT Height not multiple of 8: %d", padded.Height)
	}
}

// TestGetPaddedFrameSizeWithUpsampling tests with JPEG upsampling factors
func TestGetPaddedFrameSizeWithUpsampling(t *testing.T) {
	f := &Frame{
		Header: &FrameHeader{
			Encoding:        MODULAR,
			jpegUpsamplingX: []int32{1, 0, 0}, // 2x horizontal for channel 0
			jpegUpsamplingY: []int32{1, 0, 0}, // 2x vertical for channel 0
			Bounds: &util.Rectangle{
				Origin: util.Point{},
				Size:   util.Dimension{Width: 16, Height: 16},
			},
		},
	}

	padded, err := f.GetPaddedFrameSize()
	if err != nil {
		t.Fatalf("GetPaddedFrameSize returned error: %v", err)
	}

	// With upsampling factor of 2, size should be adjusted
	if padded.Width == 0 || padded.Height == 0 {
		t.Errorf("Padded size should not be zero: %+v", padded)
	}
}

// TestGroupPosInLFGroup tests the groupPosInLFGroup calculation
func TestGroupPosInLFGroup(t *testing.T) {
	f := &Frame{
		groupRowStride:   8,
		lfGroupRowStride: 1,
	}

	// Group 0 in LFGroup 0
	pos := f.groupPosInLFGroup(0, 0)
	if pos.X != 0 || pos.Y != 0 {
		t.Errorf("groupPosInLFGroup(0, 0) = (%d, %d); want (0, 0)", pos.X, pos.Y)
	}

	// Group 1 in LFGroup 0
	pos = f.groupPosInLFGroup(0, 1)
	if pos.X != 1 || pos.Y != 0 {
		t.Errorf("groupPosInLFGroup(0, 1) = (%d, %d); want (1, 0)", pos.X, pos.Y)
	}
}

// TestIsVisibleEdgeCases tests IsVisible edge cases
func TestIsVisibleEdgeCases(t *testing.T) {
	f := &Frame{
		Header: &FrameHeader{},
	}

	// LF_FRAME should not be visible
	f.Header.FrameType = LF_FRAME
	f.Header.Duration = 0
	f.Header.IsLast = false
	if f.IsVisible() {
		t.Error("LF_FRAME should not be visible")
	}

	// REFERENCE_ONLY should not be visible
	f.Header.FrameType = REFERENCE_ONLY
	if f.IsVisible() {
		t.Error("REFERENCE_ONLY should not be visible")
	}

	// SKIP_PROGRESSIVE with IsLast=true should be visible
	f.Header.FrameType = SKIP_PROGRESSIVE
	f.Header.Duration = 0
	f.Header.IsLast = true
	if !f.IsVisible() {
		t.Error("SKIP_PROGRESSIVE with IsLast should be visible")
	}
}

// TestGenerateSignaturesForBuffer tests the signature generation function
func TestGenerateSignaturesForBuffer(t *testing.T) {
	f := &Frame{}
	ib, err := image.NewImageBuffer(image.TYPE_FLOAT, 4, 4)
	if err != nil {
		t.Fatalf("NewImageBuffer failed: %v", err)
	}

	// Fill with simple values
	for y := 0; y < 4; y++ {
		for x := 0; x < 4; x++ {
			ib.FloatBuffer[y][x] = float32(y*4 + x)
		}
	}
	f.Buffer = []image.ImageBuffer{*ib}

	sigs := f.generateSignaturesForBuffer(0)
	if len(sigs) != 4 {
		t.Errorf("Expected 4 signatures, got %d", len(sigs))
	}

	// Each signature should be non-empty string
	for i, sig := range sigs {
		if sig == "" {
			t.Errorf("Signature %d is empty", i)
		}
	}
}

// TestReadPermutationIllegalLehmer tests readPermutation with illegal lehmer values
func TestReadPermutationIllegalLehmer(t *testing.T) {
	data := make([]byte, 100)
	reader := jxlio.NewBitStreamReader(bytes.NewReader(data))

	// Lehmer value that exceeds size-i
	stream := &FakeEntropyStreamer{
		FakeSymbols: []int32{
			2, // end = 2
			0, // lehmer[0] = 0 (valid)
			5, // lehmer[1] = 5 (invalid for size=4, i=1: 5 >= 4-1=3)
		},
	}

	_, err := readPermutation(reader, stream, 4, 0)
	if err == nil {
		t.Error("readPermutation should return error for illegal lehmer value")
	}
}

// TestReadPermutationZeroEnd tests readPermutation with end=0
func TestReadPermutationZeroEnd(t *testing.T) {
	data := make([]byte, 100)
	reader := jxlio.NewBitStreamReader(bytes.NewReader(data))

	stream := &FakeEntropyStreamer{
		FakeSymbols: []int32{
			0, // end = 0 (no lehmer codes to read)
		},
	}

	perm, err := readPermutation(reader, stream, 4, 0)
	if err != nil {
		t.Fatalf("readPermutation with end=0 returned error: %v", err)
	}

	if len(perm) != 4 {
		t.Errorf("Permutation length = %d; want 4", len(perm))
	}

	// With end=0, should return identity permutation
	for i, v := range perm {
		if v != uint32(i) {
			t.Errorf("perm[%d] = %d; want %d (identity)", i, v, i)
		}
	}
}

// TestGetGroupSizeEdge tests getGroupSize with edge cases
func TestGetGroupSizeEdge(t *testing.T) {
	f := &Frame{
		Header: &FrameHeader{
			groupDim: 128,
			Bounds: &util.Rectangle{
				Origin: util.Point{},
				Size:   util.Dimension{Width: 256, Height: 256},
			},
			jpegUpsamplingX: []int32{0, 0, 0},
			jpegUpsamplingY: []int32{0, 0, 0},
			Encoding:        MODULAR,
		},
	}
	f.groupRowStride = 2

	// Test group at edge
	sz, err := f.getGroupSize(1) // Second group in first row
	if err != nil {
		t.Fatalf("getGroupSize error: %v", err)
	}
	if sz.Width != 128 || sz.Height != 128 {
		t.Errorf("getGroupSize(1) = %+v; want 128x128", sz)
	}
}

// TestGetLFGroupSizeEdge tests getLFGroupSize with edge cases
func TestGetLFGroupSizeEdge(t *testing.T) {
	f := &Frame{
		Header: &FrameHeader{
			lfGroupDim: 256,
			Bounds: &util.Rectangle{
				Origin: util.Point{},
				Size:   util.Dimension{Width: 300, Height: 300},
			},
			jpegUpsamplingX: []int32{0, 0, 0},
			jpegUpsamplingY: []int32{0, 0, 0},
			Encoding:        MODULAR,
		},
	}
	f.lfGroupRowStride = 2

	// Test LF group that spans partial area
	sz, err := f.getLFGroupSize(1) // Second LF group
	if err != nil {
		t.Fatalf("getLFGroupSize error: %v", err)
	}
	// Should be clamped to remaining size (300-256=44)
	if sz.Width != 44 {
		t.Errorf("getLFGroupSize(1).Width = %d; want 44", sz.Width)
	}
}

// TestInvertSubsamplingBothAxes tests invertSubsampling with both X and Y
func TestInvertSubsamplingBothAxes(t *testing.T) {
	f := &Frame{
		GlobalMetadata: &bundle.ImageHeader{
			BitDepth: &bundle.BitDepthHeader{BitsPerSample: 8},
		},
		Header: &FrameHeader{
			jpegUpsamplingX: []int32{1, 0, 0}, // 2x horizontal for channel 0
			jpegUpsamplingY: []int32{1, 0, 0}, // 2x vertical for channel 0
		},
	}

	f.Buffer = make([]image.ImageBuffer, 3)
	// Channel 0: half width and half height
	ib0, _ := image.NewImageBuffer(image.TYPE_FLOAT, 2, 2)
	for y := 0; y < 2; y++ {
		for x := 0; x < 2; x++ {
			ib0.FloatBuffer[y][x] = float32(y*2 + x)
		}
	}
	f.Buffer[0] = *ib0

	// Other channels normal size
	for c := 1; c < 3; c++ {
		ib, _ := image.NewImageBuffer(image.TYPE_FLOAT, 4, 4)
		for y := 0; y < 4; y++ {
			for x := 0; x < 4; x++ {
				ib.FloatBuffer[y][x] = float32(y*4 + x)
			}
		}
		f.Buffer[c] = *ib
	}

	err := f.invertSubsampling()
	if err != nil {
		t.Fatalf("invertSubsampling returned error: %v", err)
	}

	// Channel 0 should now be 4x4
	if f.Buffer[0].Width != 4 || f.Buffer[0].Height != 4 {
		t.Errorf("After both axes upsampling, size = %dx%d; want 4x4",
			f.Buffer[0].Width, f.Buffer[0].Height)
	}
}

// TestPerformGabConvolutionMultipleWorkers tests Gab convolution with multiple goroutines
func TestPerformGabConvolutionMultipleWorkers(t *testing.T) {
	f := &Frame{
		GlobalMetadata: &bundle.ImageHeader{
			BitDepth:       &bundle.BitDepthHeader{BitsPerSample: 8},
			XybEncoded:     false,
			ColourEncoding: &colour.ColourEncodingBundle{ColourEncoding: colour.CE_RGB},
		},
		Header: &FrameHeader{
			Encoding:          MODULAR,
			restorationFilter: NewRestorationFilter(),
		},
		options: &options.JXLOptions{MaxGoroutines: 4}, // Multiple workers
	}

	// Create 3 float buffers
	f.Buffer = make([]image.ImageBuffer, 3)
	for c := 0; c < 3; c++ {
		ib, err := image.NewImageBuffer(image.TYPE_FLOAT, 16, 16)
		if err != nil {
			t.Fatalf("NewImageBuffer failed: %v", err)
		}
		for y := 0; y < 16; y++ {
			for x := 0; x < 16; x++ {
				ib.FloatBuffer[y][x] = float32(y*16+x) / 256.0
			}
		}
		f.Buffer[c] = *ib
	}

	err := f.performGabConvolution()
	if err != nil {
		t.Fatalf("performGabConvolution returned error: %v", err)
	}

	// Verify buffer sizes are preserved
	for c := 0; c < 3; c++ {
		if f.Buffer[c].Width != 16 || f.Buffer[c].Height != 16 {
			t.Errorf("Channel %d size changed: %dx%d", c, f.Buffer[c].Width, f.Buffer[c].Height)
		}
	}
}

// TestGenerateSignaturesForBufferRow288 tests the special row 288 case
func TestGenerateSignaturesForBufferRow288(t *testing.T) {
	f := &Frame{}
	// Create buffer with at least 289 rows to hit row 288 special case
	ib, err := image.NewImageBuffer(image.TYPE_FLOAT, 290, 4)
	if err != nil {
		t.Fatalf("NewImageBuffer failed: %v", err)
	}

	// Fill with simple values (row 288 will be checked specially)
	for y := 0; y < 290; y++ {
		for x := 0; x < 4; x++ {
			ib.FloatBuffer[y][x] = float32(y*4 + x)
		}
	}
	f.Buffer = []image.ImageBuffer{*ib}

	sigs := f.generateSignaturesForBuffer(0)
	if len(sigs) != 290 {
		t.Errorf("Expected 290 signatures, got %d", len(sigs))
	}
}

// TestGetGroupSizeError tests getGroupSize error handling
func TestGetGroupSizeError(t *testing.T) {
	f := &Frame{
		Header: &FrameHeader{
			groupDim: 128,
			Bounds: &util.Rectangle{
				Origin: util.Point{},
				Size:   util.Dimension{Width: 256, Height: 256},
			},
			jpegUpsamplingX: []int32{0, 0, 0},
			jpegUpsamplingY: []int32{0, 0, 0},
			Encoding:        MODULAR,
		},
	}
	f.groupRowStride = 2

	// Test getting size for valid group
	sz, err := f.getGroupSize(0)
	if err != nil {
		t.Fatalf("getGroupSize(0) error: %v", err)
	}
	if sz.Width == 0 || sz.Height == 0 {
		t.Error("getGroupSize returned zero dimension")
	}
}

// TestGetLFGroupSizeError tests getLFGroupSize error handling
func TestGetLFGroupSizeError(t *testing.T) {
	f := &Frame{
		Header: &FrameHeader{
			lfGroupDim: 256,
			Bounds: &util.Rectangle{
				Origin: util.Point{},
				Size:   util.Dimension{Width: 256, Height: 256},
			},
			jpegUpsamplingX: []int32{0, 0, 0},
			jpegUpsamplingY: []int32{0, 0, 0},
			Encoding:        MODULAR,
		},
	}
	f.lfGroupRowStride = 1

	// Test getting size for valid LF group
	sz, err := f.getLFGroupSize(0)
	if err != nil {
		t.Fatalf("getLFGroupSize(0) error: %v", err)
	}
	if sz.Width == 0 || sz.Height == 0 {
		t.Error("getLFGroupSize returned zero dimension")
	}
}

// TestUpsampleWithExtraChannels tests Upsample with extra channels
func TestUpsampleWithExtraChannels(t *testing.T) {
	f := &Frame{
		GlobalMetadata: &bundle.ImageHeader{
			BitDepth: &bundle.BitDepthHeader{BitsPerSample: 8},
			ExtraChannelInfo: []bundle.ExtraChannelInfo{
				{BitDepth: bundle.BitDepthHeader{BitsPerSample: 8}},
			},
			ColourEncoding: &colour.ColourEncodingBundle{ColourEncoding: colour.CE_GRAY},
			Up2Weights:     bundle.DEFAULT_UP2,
			Up4Weights:     bundle.DEFAULT_UP4,
			Up8Weights:     bundle.DEFAULT_UP8,
		},
		Header: &FrameHeader{
			Upsampling:   2,
			EcUpsampling: []uint32{1}, // Extra channel has no upsampling
			Bounds:       &util.Rectangle{Origin: util.Point{X: 0, Y: 0}, Size: util.Dimension{Width: 4, Height: 4}},
			groupDim:     8,
			Encoding:     MODULAR, // Not VARDCT, not XYB
		},
	}

	// Create 2 buffers: 1 grey color channel + 1 extra channel
	// GetColourChannelCount() returns 1 for CE_GREY
	f.Buffer = make([]image.ImageBuffer, 2)
	for c := 0; c < 2; c++ {
		ib, err := image.NewImageBuffer(image.TYPE_FLOAT, 4, 4)
		if err != nil {
			t.Fatalf("NewImageBuffer failed: %v", err)
		}
		for y := 0; y < 4; y++ {
			for x := 0; x < 4; x++ {
				ib.FloatBuffer[y][x] = float32(y*4+x) / 16.0
			}
		}
		f.Buffer[c] = *ib
	}

	err := f.Upsample()
	if err != nil {
		t.Fatalf("Upsample returned error: %v", err)
	}

	// Grey color channel should be 8x8, extra channel should remain 4x4
	if f.Buffer[0].Width != 8 || f.Buffer[0].Height != 8 {
		t.Errorf("Grey color channel after Upsample = %dx%d; want 8x8",
			f.Buffer[0].Width, f.Buffer[0].Height)
	}
	if f.Buffer[1].Width != 4 || f.Buffer[1].Height != 4 {
		t.Errorf("Extra channel after Upsample = %dx%d; want 4x4",
			f.Buffer[1].Width, f.Buffer[1].Height)
	}
}

func TestDecodeFrame(t *testing.T) {

	frame := &Frame{
		tocPermutation: nil,
		//tocLengths:       []uint32{0x513d, 0x0, 0x0, 0x0, 0x0, 0x0, 0xdf34, 0xe433, 0xd43a, 0xda74, 0xec3f, 0xe0d9, 0xe9d8, 0xe7d7, 0xe80e, 0xe5a3, 0xe00d, 0xec7f, 0xadc5, 0xc9d8, 0xe879, 0xd567, 0xe11a, 0xf591, 0xd533, 0xe39c, 0xf26a, 0xeba7, 0xe10e, 0xe863, 0xe6cd, 0xad6e, 0xc76a, 0xcbef, 0xe202, 0xd202, 0xdd76, 0xefd8, 0x105bf, 0x109bd, 0xf359, 0xdf25, 0xdfc5, 0xe8b2, 0xb115, 0xc7db, 0xc67e, 0xd835, 0xde84, 0xfe5b, 0xe894, 0xd494, 0xf95c, 0xf79d, 0xdb79, 0xe675, 0xe7d1, 0xacaa, 0xe109, 0xcce3, 0xdcbb, 0xe996, 0xe4df, 0xeb92, 0x107b0, 0xfa0b, 0xfa4b, 0xe0d0, 0xd954, 0xdf3d, 0xaaf6, 0xc74b, 0xd90a, 0xe26d, 0xe92a, 0x12702, 0xebaa, 0x109da, 0x10857, 0xf1fd, 0xe995, 0xeb3e, 0xe6bb, 0xac69, 0xd207, 0xd59d, 0xd50f, 0xc90f, 0xfb02, 0x115af, 0x1100a, 0x1152d, 0xfa05, 0xe408, 0xe849, 0xddb5, 0x9d28, 0xc9f2, 0xd731, 0xdd8f, 0xdbea, 0xcf0c, 0xee08, 0xfdd9, 0xde1d, 0xd5b9, 0xd596, 0xdca4, 0xde90, 0xa57e, 0xcfcb, 0xe193, 0xdb39, 0xda07, 0xd7c5, 0xc722, 0xe195, 0xd6c5, 0xcf2c, 0xafb1, 0x85c0, 0xdb91, 0xa78a, 0x810a, 0x7d9d, 0x7719, 0x7d6b, 0x83f2, 0x7e83, 0x88ab, 0x8485, 0x77ee, 0x58f1, 0x48ad, 0x7bf5, 0x5cdd},
		tocLengths: []uint32{1},
		lfGroups:   nil,
		Buffer:     nil,
		passes:     nil,
		bitreaders: nil,
		GlobalMetadata: &bundle.ImageHeader{
			BitDepth: &bundle.BitDepthHeader{
				BitsPerSample:    0,
				ExpBits:          0,
				UsesFloatSamples: false,
			},
			ColourEncoding: &colour.ColourEncodingBundle{
				ColourEncoding: colour.CE_RGB,
			},
		},
		options: &options.JXLOptions{ParseOnly: false, RenderVarblocks: false, MaxGoroutines: 24},
		reader:  testcommon.NewFakeBitReader(),
		Header: &FrameHeader{
			jpegUpsamplingX: []int32{0, 0, 0},
			jpegUpsamplingY: []int32{0, 0, 0},
			EcUpsampling:    nil,
			EcBlendingInfo:  nil,
			name:            "",
			Bounds: &util.Rectangle{
				Origin: util.Point{},
				Size: util.Dimension{
					Width:  5,
					Height: 5,
				},
			},
			restorationFilter: &RestorationFilter{},
			extensions:        nil,
			passes: &PassesInfo{
				numPasses: 0,
			},
			BlendingInfo:    nil,
			Flags:           0,
			FrameType:       0,
			Width:           0,
			Height:          0,
			Upsampling:      0,
			LfLevel:         0,
			groupDim:        0,
			Encoding:        MODULAR,
			groupSizeShift:  0,
			lfGroupDim:      0,
			logGroupDim:     0,
			logLFGroupDIM:   0,
			xqmScale:        0,
			bqmScale:        0,
			Duration:        0,
			timecode:        0,
			SaveAsReference: 0,
			SaveBeforeCT:    false,
			DoYCbCr:         false,
			haveCrop:        false,
			IsLast:          false,
		},
		globalTree: nil,
		hfGlobal:   nil,
		LfGlobal: &LFGlobal{
			frame:           nil,
			Patches:         nil,
			splines:         nil,
			noiseParameters: nil,
			lfDequant:       nil,
			hfBlockCtx:      nil,
			lfChanCorr:      nil,
			globalScale:     0,
			quantLF:         0,
			scaledDequant:   nil,
			globalModular: &ModularStream{
				channels: []*ModularChannel{},
			},
		},
		groupRowStride:   0xd,
		lfGroupRowStride: 0x2,
		numGroups:        0x82,
		numLFGroups:      0x4,
		permutatedTOC:    false,
		decoded:          false,
	}

	reader := testcommon.NewFakeBitReader()
	reader.ReadBoolData = []bool{true, false}
	frame.reader = reader

	err := frame.DecodeFrame(nil, NewFakeLFGlobalWithReaderFunc)

	if err != nil {
		t.Errorf("Error decoding frame: %v", err)
	}
}

//func TestDecodeFrameWithRealFile(t *testing.T) {
//	f, err := os.ReadFile(`../testdata/unittest.jxl`)
//	if err != nil {
//		log.Errorf("Error opening file: %v\n", err)
//		return
//	}
//
//	r := bytes.NewReader(f)
//	jxl := core.NewJXLDecoder(r, nil)
//
//	jxl.Decode()
//
//	var jxlImage *core.JXLImage
//	if jxlImage, err = jxl.Decode(); err != nil {
//		fmt.Printf("Error decoding: %v\n", err)
//		t.Errorf("Error decoding: %v\n", err)
//	}
//
//	fmt.Printf("XXXXX %+v\n", jxlImage)
//
//}

func NewFakeLFGlobalWithReaderFunc(reader jxlio.BitReader, parent Framer, hfBlockContextFunc NewHFBlockContextFunc) (*LFGlobal, error) {
	return &LFGlobal{
		frame:           nil,
		Patches:         nil,
		splines:         nil,
		noiseParameters: nil,
		lfDequant:       nil,
		hfBlockCtx:      nil,
		lfChanCorr:      nil,
		globalScale:     0,
		quantLF:         0,
		scaledDequant:   nil,
		globalModular: &ModularStream{
			channels: []*ModularChannel{},
		},
	}, nil
}
