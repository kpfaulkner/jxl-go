package core

import (
	"bytes"
	"os"
	"reflect"
	"testing"

	"github.com/kpfaulkner/jxl-go/bundle"
	"github.com/kpfaulkner/jxl-go/colour"
	frame2 "github.com/kpfaulkner/jxl-go/frame"
	"github.com/kpfaulkner/jxl-go/image"
	"github.com/kpfaulkner/jxl-go/jxlio"
	"github.com/kpfaulkner/jxl-go/options"
	"github.com/kpfaulkner/jxl-go/util"
)

func GenerateTestBitReader(t *testing.T, filename string) jxlio.BitReader {
	data, err := os.ReadFile(filename)
	if err != nil {
		t.Errorf("error reading test jxl file : %v", err)
		return nil
	}
	br := jxlio.NewBitStreamReader(bytes.NewReader(data))

	return br
}

// TestReadSignatureAndBoxes tests the ReadSignatureAndBoxes function.
// For testing, instead of providing a mock BitStreamReader and having to fake all the data
// I'll provide a real BitStreamReader and test the function with real data. May eventually
// swap it out for a mock BitStreamReader.
func TestReadSignatureAndBoxes(t *testing.T) {

	for _, tc := range []struct {
		name               string
		expectErr          bool
		expectedBoxHeaders []ContainerBoxHeader
	}{
		{
			name:               "success",
			expectErr:          false,
			expectedBoxHeaders: []ContainerBoxHeader{{BoxType: JXLC, BoxSize: 7085536, IsLast: false, Offset: 40}},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {

			br := GenerateTestBitReader(t, "../testdata/unittest.jxl")
			decoder := NewJXLCodestreamDecoder(br, nil)
			err := decoder.readSignatureAndBoxes()
			if err != nil && tc.expectErr {
				// got what we wanted..
				return
			}

			if err == nil && tc.expectErr {
				t.Errorf("expected error but got none")
			}

			if err != nil && !tc.expectErr {
				t.Errorf("expected no error but got %v", err)
			}

			if len(decoder.boxHeaders) != len(tc.expectedBoxHeaders) {
				t.Errorf("expected %d box headers but got %d", len(tc.expectedBoxHeaders), len(decoder.boxHeaders))
			}

			if !reflect.DeepEqual(decoder.boxHeaders, tc.expectedBoxHeaders) {
				t.Errorf("actual box headers does not match expected box headers")
			}
		})
	}
}

func TestGetImageHeader(t *testing.T) {

	for _, tc := range []struct {
		name           string
		data           []uint8
		expectErr      bool
		expectedHeader bundle.ImageHeader
	}{
		{
			name:      "success",
			expectErr: false,
			expectedHeader: bundle.ImageHeader{
				Level: 5,
				Size: util.Dimension{
					Width:  3264,
					Height: 2448,
				},
				Orientation: 1,
				BitDepth: &bundle.BitDepthHeader{
					UsesFloatSamples: false,
					BitsPerSample:    8,
					ExpBits:          0,
				},
				OrientedWidth:       3264,
				OrientedHeight:      2448,
				Modular16BitBuffers: true,
				ExtraChannelInfo:    nil,
				XybEncoded:          false,
				ColourEncoding: &colour.ColourEncodingBundle{
					UseIccProfile:  false,
					ColourEncoding: 0,
					WhitePoint:     1,
					White: &colour.CIEXY{
						X: 0.3127,
						Y: 0.329,
					},
					Primaries: 1,
					Prim: &colour.CIEPrimaries{
						Red: &colour.CIEXY{
							X: 0.6399987,
							Y: 0.33001015,
						},
						Green: &colour.CIEXY{
							X: 0.3000038,
							Y: 0.60000336,
						},
						Blue: &colour.CIEXY{
							X: 0.15000205,
							Y: 0.059997205,
						},
					},
					Tf:              16777229,
					RenderingIntent: 0,
				},
				AlphaIndices: nil,
				ToneMapping: &colour.ToneMapping{
					IntensityTarget:      255,
					MinNits:              0,
					RelativeToMaxDisplay: false,
					LinearBelow:          0,
				},
				Extensions:         nil,
				OpsinInverseMatrix: nil,
				Up2Weights:         nil,
				Up4Weights:         nil,
				Up8Weights:         nil,
				EncodedICC:         nil,
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {

			br := GenerateTestBitReader(t, "../testdata/unittest.jxl")
			opts := options.NewJXLOptions(nil)
			decoder := NewJXLCodestreamDecoder(br, opts)

			header, err := decoder.GetImageHeader()
			if err != nil && tc.expectErr {
				// got what we wanted..
				return
			}

			if err == nil && tc.expectErr {
				t.Errorf("expected error but got none")
			}

			if err != nil && !tc.expectErr {
				t.Errorf("expected no error but got %v", err)
			}

			// not going to deepequals the whole struct. Will check a few key fields and will extend this later on.
			if header.Level != tc.expectedHeader.Level {
				t.Errorf("expected level %d but got %d", tc.expectedHeader.Level, header.Level)
			}

			if header.Size.Width != tc.expectedHeader.Size.Width {
				t.Errorf("expected width %d but got %d", tc.expectedHeader.Size.Width, header.Size.Width)
			}
			if header.Size.Height != tc.expectedHeader.Size.Height {
				t.Errorf("expected height %d but got %d", tc.expectedHeader.Size.Height, header.Size.Height)
			}
			if header.BitDepth.BitsPerSample != tc.expectedHeader.BitDepth.BitsPerSample {
				t.Errorf("expected bits per sample %d but got %d", tc.expectedHeader.BitDepth.BitsPerSample, header.BitDepth.BitsPerSample)
			}
			if header.BitDepth.UsesFloatSamples != tc.expectedHeader.BitDepth.UsesFloatSamples {
				t.Errorf("expected uses float samples %t but got %t", tc.expectedHeader.BitDepth.UsesFloatSamples, header.BitDepth.UsesFloatSamples)
			}
			if header.BitDepth.ExpBits != tc.expectedHeader.BitDepth.ExpBits {
				t.Errorf("expected exp bits %d but got %d", tc.expectedHeader.BitDepth.ExpBits, header.BitDepth.ExpBits)
			}
			if header.ColourEncoding.UseIccProfile != tc.expectedHeader.ColourEncoding.UseIccProfile {
				t.Errorf("expected use icc profile %t but got %t", tc.expectedHeader.ColourEncoding.UseIccProfile, header.ColourEncoding.UseIccProfile)
			}

			if !header.ColourEncoding.Prim.Matches(tc.expectedHeader.ColourEncoding.Prim) {
				t.Errorf("expected primaries to match but they did not")
			}
		})
	}
}

func TestDecode(t *testing.T) {

	for _, tc := range []struct {
		name      string
		data      []uint8
		filename  string
		expectErr bool
	}{
		{
			name:      "success",
			expectErr: false,
			filename:  "../testdata/unittest.jxl",
		},
		{
			name:      "success 2",
			expectErr: false,
			filename:  "../testdata/tiny2.jxl",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {

			br := GenerateTestBitReader(t, tc.filename)
			opts := options.NewJXLOptions(nil)
			decoder := NewJXLCodestreamDecoder(br, opts)

			_, err := decoder.decode()
			if err != nil && tc.expectErr {
				// got what we wanted..
				return
			}

			if err == nil && tc.expectErr {
				t.Errorf("expected error but got none")
			}

			if err != nil && !tc.expectErr {
				t.Errorf("expected no error but got %v", err)
			}
		})
	}
}

func TestBlendFrame(t *testing.T) {

	for _, tc := range []struct {
		name        string
		dataGenFunc func(*testing.T) (*JXLCodestreamDecoder, *image.ImageBuffer, error, frame2.Frame)
	}{
		{
			name:        "success replace",
			dataGenFunc: generateBlendReplaceTestData,
		},
		{
			name:        "success add",
			dataGenFunc: generateBlendAddTestData,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {

			jxl, canvas, err, frame := tc.dataGenFunc(t)

			err = jxl.blendFrame([]image.ImageBuffer{*canvas}, &frame)
			if err != nil {
				t.Errorf("error blending frame : %v", err)
			}
		})
	}
}

func generateBlendReplaceTestData(t *testing.T) (*JXLCodestreamDecoder, *image.ImageBuffer, error, frame2.Frame) {

	// Need to simplify the code so mocking out structs with fake data
	// is easier.
	jxl := NewJXLCodestreamDecoder(nil, nil)

	jxl.imageHeader = &bundle.ImageHeader{
		Size: util.Dimension{Width: 10, Height: 10},
		ColourEncoding: &colour.ColourEncodingBundle{
			ColourEncoding: colour.CE_RGB,
		},
	}
	bufferRef, _ := image.NewImageBuffer(image.TYPE_INT, 10, 10)
	jxl.reference = make([][]image.ImageBuffer, 1)
	jxl.reference[0] = []image.ImageBuffer{*bufferRef}

	canvas, err := image.NewImageBuffer(image.TYPE_INT, 10, 10)
	if err != nil {
		t.Errorf("error creating image buffer : %v", err)
	}

	buffer, _ := image.NewImageBuffer(image.TYPE_INT, 10, 10)
	frame := frame2.Frame{
		Buffer: []image.ImageBuffer{*buffer},
		Header: &frame2.FrameHeader{
			BlendingInfo: &frame2.BlendingInfo{
				Source: 0,
			},
			Bounds: &util.Rectangle{
				Origin: util.Point{
					X: 0,
					Y: 0,
				},
				Size: util.Dimension{
					Width:  10,
					Height: 10,
				},
			},
		},
		GlobalMetadata: &bundle.ImageHeader{
			ColourEncoding: &colour.ColourEncodingBundle{
				ColourEncoding: colour.CE_RGB,
			},
		},
	}
	return jxl, canvas, err, frame
}

func generateBlendAddTestData(t *testing.T) (*JXLCodestreamDecoder, *image.ImageBuffer, error, frame2.Frame) {

	// Need to simplify the code so mocking out structs with fake data
	// is easier.
	jxl := NewJXLCodestreamDecoder(nil, nil)

	jxl.imageHeader = &bundle.ImageHeader{
		Size: util.Dimension{Width: 10, Height: 10},
		ColourEncoding: &colour.ColourEncodingBundle{
			ColourEncoding: colour.CE_RGB,
		},
	}
	bufferRef, _ := image.NewImageBuffer(image.TYPE_INT, 10, 10)
	jxl.reference = make([][]image.ImageBuffer, 1)
	jxl.reference[0] = []image.ImageBuffer{*bufferRef}

	canvas, err := image.NewImageBuffer(image.TYPE_INT, 10, 10)
	if err != nil {
		t.Errorf("error creating image buffer : %v", err)
	}

	buffer, _ := image.NewImageBuffer(image.TYPE_INT, 10, 10)
	frame := frame2.Frame{
		Buffer: []image.ImageBuffer{*buffer},
		Header: &frame2.FrameHeader{
			BlendingInfo: &frame2.BlendingInfo{
				Source: 0,
				Mode:   frame2.BLEND_ADD,
			},
			Bounds: &util.Rectangle{
				Origin: util.Point{
					X: 0,
					Y: 0,
				},
				Size: util.Dimension{
					Width:  10,
					Height: 10,
				},
			},
		},
		GlobalMetadata: &bundle.ImageHeader{
			ColourEncoding: &colour.ColourEncodingBundle{
				ColourEncoding: colour.CE_RGB,
			},
		},
	}
	return jxl, canvas, err, frame
}

func TestPerformBlending(t *testing.T) {

	for _, tc := range []struct {
		name                string
		canvas              []image.ImageBuffer
		info                *frame2.BlendingInfo
		frameBuffer         image.ImageBuffer
		canvasIdx           int32
		ref                 image.ImageBuffer
		frameHeight         int32
		frameStartY         int32
		frameOffsetY        int32
		frameWidth          int32
		frameStartX         int32
		frameOffsetX        int32
		hasAlpha            bool
		refBuffer           []image.ImageBuffer
		imageColours        int
		frameBuffers        []image.ImageBuffer
		frameColours        int
		isAlpha             bool
		premult             bool
		expectedImageBuffer image.ImageBuffer
		expectErr           bool
	}{
		{
			name: "blend add is int, success",
			canvas: []image.ImageBuffer{{
				Width:       10,
				Height:      10,
				BufferType:  image.TYPE_INT,
				FloatBuffer: nil,
				IntBuffer:   makeFullMatrix[int32](10, 10, 2),
			}},
			info: &frame2.BlendingInfo{
				Mode: frame2.BLEND_ADD,
			},

			frameBuffer: image.ImageBuffer{
				Width:       10,
				Height:      10,
				BufferType:  image.TYPE_INT,
				FloatBuffer: nil,
				IntBuffer:   makeFullMatrix[int32](10, 10, 3),
			},
			canvasIdx: 0,
			ref: image.ImageBuffer{
				Width:       10,
				Height:      10,
				BufferType:  image.TYPE_INT,
				FloatBuffer: nil,
				IntBuffer:   makeFullMatrix[int32](10, 10, 4),
			},
			frameHeight:  10,
			frameStartY:  0,
			frameOffsetY: 0,
			frameWidth:   10,
			frameStartX:  0,
			frameOffsetX: 0,
			hasAlpha:     false,
			refBuffer:    nil,
			imageColours: 0,
			frameBuffers: nil,
			frameColours: 0,
			isAlpha:      false,
			premult:      false,
			expectedImageBuffer: image.ImageBuffer{
				Width:      10,
				Height:     10,
				BufferType: image.TYPE_INT,
				IntBuffer:  makeFullMatrix[int32](10, 10, 7),
			},
			expectErr: false,
		},
		{
			name: "blend add is float, success",
			canvas: []image.ImageBuffer{{
				Width:       10,
				Height:      10,
				BufferType:  image.TYPE_FLOAT,
				FloatBuffer: makeFullMatrix[float32](10, 10, 2),
			}},
			info: &frame2.BlendingInfo{
				Mode: frame2.BLEND_ADD,
			},

			frameBuffer: image.ImageBuffer{
				Width:       10,
				Height:      10,
				BufferType:  image.TYPE_FLOAT,
				FloatBuffer: makeFullMatrix[float32](10, 10, 3),
			},
			canvasIdx: 0,
			ref: image.ImageBuffer{
				Width:       10,
				Height:      10,
				BufferType:  image.TYPE_FLOAT,
				FloatBuffer: makeFullMatrix[float32](10, 10, 4),
			},
			frameHeight:  10,
			frameStartY:  0,
			frameOffsetY: 0,
			frameWidth:   10,
			frameStartX:  0,
			frameOffsetX: 0,
			hasAlpha:     false,
			refBuffer:    nil,
			imageColours: 0,
			frameBuffers: nil,
			frameColours: 0,
			isAlpha:      false,
			premult:      false,
			expectedImageBuffer: image.ImageBuffer{
				Width:       10,
				Height:      10,
				BufferType:  image.TYPE_FLOAT,
				FloatBuffer: makeFullMatrix[float32](10, 10, 7),
			},
			expectErr: false,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {

			jxl := NewJXLCodestreamDecoder(nil, nil)
			err := jxl.performBlending(tc.canvas, tc.info, tc.frameBuffer, tc.canvasIdx, tc.ref, tc.frameHeight, tc.frameStartY, tc.frameOffsetY, tc.frameWidth, tc.frameStartX, tc.frameOffsetX, tc.hasAlpha, tc.refBuffer, tc.imageColours, tc.frameBuffers, tc.frameColours, tc.isAlpha, tc.premult)

			if tc.expectErr && err == nil {
				t.Errorf("expected error but got none")
			}
			if !tc.expectErr && err != nil {
				t.Errorf("expected no error but got %v", err)
			}

			if !tc.expectErr && !reflect.DeepEqual(tc.canvas[0].IntBuffer, tc.expectedImageBuffer.IntBuffer) {
				t.Errorf("expected %v but got %v", tc.expectedImageBuffer.IntBuffer, tc.canvas[0].IntBuffer)
			}
		})
	}
}

func makeFullMatrix[T any](height int, width int, val T) [][]T {
	mat := util.MakeMatrix2D[T](height, width)
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			mat[y][x] = val
		}
	}
	return mat
}
