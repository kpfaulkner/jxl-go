package frame

import (
	"reflect"
	"testing"

	"github.com/kpfaulkner/jxl-go/entropy"
	"github.com/kpfaulkner/jxl-go/testcommon"
	"github.com/kpfaulkner/jxl-go/util"
)

//func TestNewHFCoefficientsWithReader(t *testing.T) {
//
//	for _, tc := range []struct {
//		name      string
//		data      []uint8
//		filename  string
//		expectErr bool
//	}{
//		{
//			name:      "success",
//			expectErr: false,
//			filename:  "../testdata/unittest.jxl",
//		},
//		{
//			name:      "success 2",
//			expectErr: false,
//			filename:  "../testdata/tiny2.jxl",
//		},
//	} {
//		t.Run(tc.name, func(t *testing.T) {
//
//			file, err := os.Open(tc.filename)
//			if err != nil {
//				t.Errorf("error opening test jxl file : %v", err)
//				return
//			}
//			defer file.Close() // Ensure the file is closed
//
//			// The *os.File directly implements io.ReadSeeker
//			var readSeeker io.ReadSeeker = file
//
//			opts := options.NewJXLOptions(nil)
//			decoder := core.NewJXLDecoder(readSeeker, opts)
//
//			_, err = decoder.Decode()
//			if err != nil && tc.expectErr {
//				// got what we wanted..
//				return
//			}
//
//			if err == nil && tc.expectErr {
//				t.Errorf("expected error but got none")
//			}
//
//			if err != nil && !tc.expectErr {
//				t.Errorf("expected no error but got %v", err)
//			}
//		})
//	}
//}

func TestNewHFCoefficientsWithReaderOrig(t *testing.T) {

	for _, tc := range []struct {
		name           string
		frame          Framer
		pass           uint32
		group          uint32
		boolData       []bool
		enumData       []int32
		u32Data        []uint32
		bitsData       []uint64
		expectedResult HFCoefficients
		expectErr      bool
	}{
		{
			name:      "no data",
			expectErr: true,
		},
		{
			name: "success",
			frame: &FakeFramer{
				lfGroup: &LFGroup{
					hfMetadata: &HFMetadata{

						blockList: []util.Point{{
							X: 0,
							Y: 0,
						}},
					},
					lfGroupID: 0,
				},
				hfGlobal: &HFGlobal{numHFPresets: 1},
				lfGlobal: &LFGlobal{
					hfBlockCtx: &HFBlockContext{},
				},
				header: &FrameHeader{
					jpegUpsamplingX: []int32{0, 0, 0},
					jpegUpsamplingY: []int32{0, 0, 0},
					passes: &PassesInfo{
						shift: []uint32{0},
					},
				},
				passes: []Pass{{
					hfPass: &HFPass{
						contextStream: &entropy.EntropyStream{},
					},
				}},
				groupSize:              &util.Dimension{Width: 1, Height: 1},
				groupPosInLFGroupPoint: &util.Point{X: 0, Y: 0},
				imageHeader:            nil,
			},
			pass:     0,
			group:    0,
			boolData: []bool{},
			enumData: nil,
			u32Data:  []uint32{},
			bitsData: []uint64{
				0, // hfPreset
			},
			expectedResult: HFCoefficients{},
			expectErr:      false,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {

			bitReader := &testcommon.FakeBitReader{
				ReadU32Data:  tc.u32Data,
				ReadBitsData: tc.bitsData,
				ReadBoolData: tc.boolData,
			}
			hfc, err := NewHFCoefficientsWithReader(bitReader, tc.frame, tc.pass, tc.group)
			if err != nil && !tc.expectErr {
				t.Errorf("got error when none was expected : %v", err)
			}
			if err == nil && tc.expectErr {
				t.Errorf("expected error but got none")
			}
			if err != nil && tc.expectErr {
				return
			}
			if !reflect.DeepEqual(*hfc, tc.expectedResult) {
				t.Errorf("expected HFBlockContext %+v, got %+v", tc.expectedResult, *hfc)
			}

		})
	}
}
