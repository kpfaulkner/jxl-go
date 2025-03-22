package bundle

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/kpfaulkner/jxl-go/color"
	"github.com/kpfaulkner/jxl-go/jxlio"
	"github.com/kpfaulkner/jxl-go/testcommon"
	"github.com/kpfaulkner/jxl-go/util"
)

// TestParseImageHeader tests the reading and parsing of image header.
func TestParseImageHeader(t *testing.T) {

	for _, tc := range []struct {
		name           string
		data           []uint8
		readData       bool
		expectErr      bool
		expectedHeader ImageHeader
	}{
		{
			name:      "no data",
			data:      []uint8{},
			readData:  false,
			expectErr: true,
		},
		{
			name:      "success, no extra channels",
			data:      []uint8{},
			readData:  true,
			expectErr: false,
			expectedHeader: ImageHeader{
				Level:           5,
				Size:            util.Dimension{Width: 3264, Height: 2448},
				Orientation:     1,
				intrinsicSize:   util.Dimension{},
				PreviewSize:     nil,
				AnimationHeader: nil,
				BitDepth: &BitDepthHeader{
					UsesFloatSamples: false,
					BitsPerSample:    8,
					ExpBits:          0,
				},
				OrientedWidth:       3264,
				OrientedHeight:      2448,
				Modular16BitBuffers: true,
				ExtraChannelInfo:    []ExtraChannelInfo{},
				XybEncoded:          false,
				ColourEncoding: &color.ColourEncodingBundle{
					UseIccProfile:   false,
					ColourEncoding:  0,
					WhitePoint:      1,
					White:           &color.CIEXY{X: 0.3127, Y: 0.329},
					Primaries:       1,
					Prim:            &color.CIEPrimaries{Red: &color.CIEXY{X: 0.6399987, Y: 0.33001015}, Green: &color.CIEXY{X: 0.3000038, Y: 0.60000336}, Blue: &color.CIEXY{X: 0.15000205, Y: 0.059997205}},
					Tf:              16777229,
					RenderingIntent: 0,
				},
				AlphaIndices: nil,
				ToneMapping: &color.ToneMapping{
					IntensityTarget:      255,
					MinNits:              0,
					RelativeToMaxDisplay: false,
					LinearBelow:          0,
				},
				Extensions: &Extensions{
					ExtensionsKey: 0,
					Payloads:      [64][]byte{},
				},
				OpsinInverseMatrix: &color.OpsinInverseMatrix{
					Matrix:             [][]float32{[]float32{11.031567, -9.866944, -0.16462299}, []float32{-3.2541473, 4.4187703, -0.16462299}, []float32{-3.6588514, 2.712923, 1.9459282}},
					OpsinBias:          []float32{-0.0037930734, -0.0037930734, -0.0037930734},
					QuantBias:          []float32{0.94534993, 0.9299455, 0.9500649},
					QuantBiasNumerator: 0.145,
					Primaries: color.CIEPrimaries{
						Red:   &color.CIEXY{X: 0.6399987, Y: 0.33001015},
						Green: &color.CIEXY{X: 0.3000038, Y: 0.60000336},
						Blue:  &color.CIEXY{X: 0.15000205, Y: 0.059997205},
					},
					WhitePoint:    color.CIEXY{X: 0.3127, Y: 0.329},
					CbrtOpsinBias: []float32{-0.1559542, -0.1559542, -0.1559542},
				},
				Up2Weights: []float32{-0.017162, -0.03452303, -0.04022174, -0.02921014, -0.00624645, 0.14111091, 0.28896755, 0.00278718, -0.01610267, 0.5666155, 0.03777607, -0.01986694, -0.03144731, -0.01185068, -0.00213539},
				Up4Weights: []float32{-0.02419067, -0.03491987, -0.03693351, -0.03094285, -0.00529785, -0.01663432, -0.03556863, -0.03888905, -0.0351685, -0.00989469, 0.23651958, 0.33392945, -0.01073543, -0.01313181, -0.03556694, 0.13048175, 0.40103024, 0.0395115, -0.02077584, 0.469142, -0.0020927, -0.01484589, -0.04064806, 0.1894253, 0.5627989, 0.066744, -0.02335494, -0.03551682, -0.0075483, -0.02267919, -0.02363578, 0.00315804, -0.03399098, -0.01359519, -0.00091653, -0.00335467, -0.01163294, -0.01610294, -0.00974088, -0.00191622, -0.01095446, -0.03198464, -0.04455121, -0.0279979, -0.00645912, 0.06390599, 0.22963887, 0.00630981, -0.01897349, 0.67537266, 0.08483369, -0.02534994, -0.02205197, -0.01667999, -0.00384443},
				Up8Weights: []float32{-0.02928613, -0.03706353, -0.03783812, -0.03324558, -0.00447632, -0.02519406, -0.03752601, -0.03901508, -0.03663285, -0.00646649, -0.02066407, -0.03838633, -0.04002101, -0.03900035, -0.00901973, -0.01626393, -0.03954148, -0.0404662, -0.03979621, -0.01224485, 0.2989533, 0.3575771, -0.02447552, -0.01081748, -0.04314594, 0.2390322, 0.411193, -0.00573046, -0.01450239, -0.04246845, 0.17567618, 0.45220643, 0.02287757, -0.01936783, -0.03583255, 0.11572472, 0.47416732, 0.0628444, -0.02685066, 0.4272005, -0.02248939, -0.01155273, -0.04562755, 0.28689495, 0.4909387, -7.891e-05, -0.01545926, -0.04562659, 0.2123892, 0.53980935, 0.03369474, -0.02070211, -0.03866988, 0.1422955, 0.565934, 0.08045181, -0.02888298, -0.03680918, -0.00542229, -0.02920477, -0.02788574, -0.0211818, -0.03942402, -0.00775547, -0.02433614, -0.03193943, -0.02030828, -0.04044014, -0.01074016, -0.01930822, -0.03620399, -0.01974125, -0.03919545, -0.01456093, -0.00045072, -0.0036011, -0.01020207, -0.01231907, -0.00638988, -0.00071592, -0.00279122, -0.00957115, -0.01288327, -0.00730937, -0.00107783, -0.00210156, -0.00890705, -0.01317668, -0.00813895, -0.00153491, -0.02128481, -0.04173044, -0.04831487, -0.0329319, -0.0052526, -0.01720322, -0.04052736, -0.05045706, -0.03607317, -0.0073803, -0.01341764, -0.03965629, -0.05151616, -0.03814886, -0.01005819, 0.18968274, 0.33063683, -0.01300105, -0.0137295, -0.04017465, 0.13727832, 0.36402234, 0.0102789, -0.01832107, -0.03365072, 0.08734506, 0.38194296, 0.04338228, -0.02525993, 0.56408125, 0.00458352, -0.01648227, -0.04887868, 0.2458552, 0.6202614, 0.04314807, -0.02213737, -0.04158014, 0.1663729, 0.6502702, 0.09621636, -0.03101388, -0.04082742, -0.00904519, -0.02790922, -0.02117818, 0.00798662, -0.03995711, -0.01243427, -0.02231705, -0.02946266, 0.00992055, -0.03600283, -0.0168492, -0.00111684, -0.00411204, -0.0129713, -0.01723725, -0.01022545, -0.00165306, -0.0031311, -0.01218016, -0.01763266, -0.0112562, -0.00231663, -0.01374149, -0.0379762, -0.05142937, -0.03117307, -0.00581914, -0.01064003, -0.03608089, -0.05272168, -0.0337567, -0.00795586, 0.09628104, 0.2712999, -0.00353779, -0.01734151, -0.03153981, 0.0568623, 0.28500998, 0.02230594, -0.02374955, 0.6821433, 0.05018048, -0.02320852, -0.04383616, 0.18459474, 0.71517974, 0.10805613, -0.03263677, -0.03637639, -0.01394373, -0.02511203, -0.01728636, 0.05407331, -0.02867568, -0.01893131, -0.00240854, -0.00446511, -0.01636187, -0.02377053, -0.01522848, -0.00333334, -0.00819975, -0.02964169, -0.04499287, -0.0274535, -0.00612408, 0.02727416, 0.194466, 0.00159832, -0.02232473, 0.74982506, 0.1145262, -0.03348048, -0.01605681, -0.02070339, -0.00458223},
				UpWeights:  nil,
				EncodedICC: nil,
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {

			var bitReader *jxlio.Bitreader
			if tc.readData {
				bitReader = testcommon.GenerateTestBitReader(t, `../testdata/unittest.jxl`)
				// skip first 40 bytes due to box headers.
				bitReader.Skip(40)
			} else {
				bitReader = jxlio.NewBitreader(bytes.NewReader(tc.data))
			}

			header, err := ParseImageHeader(bitReader, 5)
			if err != nil && !tc.expectErr {
				t.Errorf("got error when none was expected : %v", err)
			}
			if err == nil && tc.expectErr {
				t.Errorf("expected error but got none")
			}
			if err != nil && tc.expectErr {
				return
			}

			if !tc.expectErr && !reflect.DeepEqual(header.BitDepth, tc.expectedHeader.BitDepth) {
				t.Errorf("expected bitdepth %+v, got %+v", tc.expectedHeader, header)
			}

			if !tc.expectErr && !reflect.DeepEqual(header.Size, tc.expectedHeader.Size) {
				t.Errorf("expected size %+v, got %+v", tc.expectedHeader, header)
			}

			if !tc.expectErr && !reflect.DeepEqual(header.OpsinInverseMatrix, tc.expectedHeader.OpsinInverseMatrix) {
				t.Errorf("expected OpsinInverseMatrix %+v, got %+v", tc.expectedHeader, header)
			}

			if !tc.expectErr && !reflect.DeepEqual(header.ExtraChannelInfo, tc.expectedHeader.ExtraChannelInfo) {
				t.Errorf("expected ExtraChannelInfo %+v, got %+v", tc.expectedHeader, header)
			}

			if !tc.expectErr && !reflect.DeepEqual(header.UpWeights, tc.expectedHeader.UpWeights) {
				t.Errorf("expected UpWeights %+v, got %+v", tc.expectedHeader, header)
			}

			if !tc.expectErr && !reflect.DeepEqual(header.Up2Weights, tc.expectedHeader.Up2Weights) {
				t.Errorf("expected Up2Weights %+v, got %+v", tc.expectedHeader, header)
			}

			if !tc.expectErr && !reflect.DeepEqual(header.Up4Weights, tc.expectedHeader.Up4Weights) {
				t.Errorf("expected Up4Weights %+v, got %+v", tc.expectedHeader, header)
			}

			if !tc.expectErr && !reflect.DeepEqual(header.Up8Weights, tc.expectedHeader.Up8Weights) {
				t.Errorf("expected Up8Weights %+v, got %+v", tc.expectedHeader, header)
			}

			if !tc.expectErr && !reflect.DeepEqual(header.ToneMapping, tc.expectedHeader.ToneMapping) {
				t.Errorf("expected ToneMapping %+v, got %+v", tc.expectedHeader, header)
			}

			if !tc.expectErr && !reflect.DeepEqual(header.ColourEncoding, tc.expectedHeader.ColourEncoding) {
				t.Errorf("expected ColourEncoding %+v, got %+v", tc.expectedHeader, header)
			}
		})
	}
}

func TestGetUpWeights(t *testing.T) {

	for _, tc := range []struct {
		name           string
		data           []uint8
		readData       bool
		expectErr      bool
		expectedHeader ImageHeader
	}{
		{
			name:      "no data",
			data:      []uint8{},
			readData:  false,
			expectErr: true,
		},
		{
			name:      "success, no extra channels",
			data:      []uint8{},
			readData:  true,
			expectErr: false,
			expectedHeader: ImageHeader{
				Level:           5,
				Size:            util.Dimension{Width: 3264, Height: 2448},
				Orientation:     1,
				intrinsicSize:   util.Dimension{},
				PreviewSize:     nil,
				AnimationHeader: nil,
				BitDepth: &BitDepthHeader{
					UsesFloatSamples: false,
					BitsPerSample:    8,
					ExpBits:          0,
				},
				OrientedWidth:       3264,
				OrientedHeight:      2448,
				Modular16BitBuffers: true,
				ExtraChannelInfo:    []ExtraChannelInfo{},
				XybEncoded:          false,
				ColourEncoding: &color.ColourEncodingBundle{
					UseIccProfile:   false,
					ColourEncoding:  0,
					WhitePoint:      1,
					White:           &color.CIEXY{X: 0.3127, Y: 0.329},
					Primaries:       1,
					Prim:            &color.CIEPrimaries{Red: &color.CIEXY{X: 0.6399987, Y: 0.33001015}, Green: &color.CIEXY{X: 0.3000038, Y: 0.60000336}, Blue: &color.CIEXY{X: 0.15000205, Y: 0.059997205}},
					Tf:              16777229,
					RenderingIntent: 0,
				},
				AlphaIndices: nil,
				ToneMapping: &color.ToneMapping{
					IntensityTarget:      255,
					MinNits:              0,
					RelativeToMaxDisplay: false,
					LinearBelow:          0,
				},
				Extensions: &Extensions{
					ExtensionsKey: 0,
					Payloads:      [64][]byte{},
				},
				OpsinInverseMatrix: &color.OpsinInverseMatrix{
					Matrix:             [][]float32{[]float32{11.031567, -9.866944, -0.16462299}, []float32{-3.2541473, 4.4187703, -0.16462299}, []float32{-3.6588514, 2.712923, 1.9459282}},
					OpsinBias:          []float32{-0.0037930734, -0.0037930734, -0.0037930734},
					QuantBias:          []float32{0.94534993, 0.9299455, 0.9500649},
					QuantBiasNumerator: 0.145,
					Primaries: color.CIEPrimaries{
						Red:   &color.CIEXY{X: 0.6399987, Y: 0.33001015},
						Green: &color.CIEXY{X: 0.3000038, Y: 0.60000336},
						Blue:  &color.CIEXY{X: 0.15000205, Y: 0.059997205},
					},
					WhitePoint:    color.CIEXY{X: 0.3127, Y: 0.329},
					CbrtOpsinBias: []float32{-0.1559542, -0.1559542, -0.1559542},
				},
				Up2Weights: []float32{-0.017162, -0.03452303, -0.04022174, -0.02921014, -0.00624645, 0.14111091, 0.28896755, 0.00278718, -0.01610267, 0.5666155, 0.03777607, -0.01986694, -0.03144731, -0.01185068, -0.00213539},
				Up4Weights: []float32{-0.02419067, -0.03491987, -0.03693351, -0.03094285, -0.00529785, -0.01663432, -0.03556863, -0.03888905, -0.0351685, -0.00989469, 0.23651958, 0.33392945, -0.01073543, -0.01313181, -0.03556694, 0.13048175, 0.40103024, 0.0395115, -0.02077584, 0.469142, -0.0020927, -0.01484589, -0.04064806, 0.1894253, 0.5627989, 0.066744, -0.02335494, -0.03551682, -0.0075483, -0.02267919, -0.02363578, 0.00315804, -0.03399098, -0.01359519, -0.00091653, -0.00335467, -0.01163294, -0.01610294, -0.00974088, -0.00191622, -0.01095446, -0.03198464, -0.04455121, -0.0279979, -0.00645912, 0.06390599, 0.22963887, 0.00630981, -0.01897349, 0.67537266, 0.08483369, -0.02534994, -0.02205197, -0.01667999, -0.00384443},
				Up8Weights: []float32{-0.02928613, -0.03706353, -0.03783812, -0.03324558, -0.00447632, -0.02519406, -0.03752601, -0.03901508, -0.03663285, -0.00646649, -0.02066407, -0.03838633, -0.04002101, -0.03900035, -0.00901973, -0.01626393, -0.03954148, -0.0404662, -0.03979621, -0.01224485, 0.2989533, 0.3575771, -0.02447552, -0.01081748, -0.04314594, 0.2390322, 0.411193, -0.00573046, -0.01450239, -0.04246845, 0.17567618, 0.45220643, 0.02287757, -0.01936783, -0.03583255, 0.11572472, 0.47416732, 0.0628444, -0.02685066, 0.4272005, -0.02248939, -0.01155273, -0.04562755, 0.28689495, 0.4909387, -7.891e-05, -0.01545926, -0.04562659, 0.2123892, 0.53980935, 0.03369474, -0.02070211, -0.03866988, 0.1422955, 0.565934, 0.08045181, -0.02888298, -0.03680918, -0.00542229, -0.02920477, -0.02788574, -0.0211818, -0.03942402, -0.00775547, -0.02433614, -0.03193943, -0.02030828, -0.04044014, -0.01074016, -0.01930822, -0.03620399, -0.01974125, -0.03919545, -0.01456093, -0.00045072, -0.0036011, -0.01020207, -0.01231907, -0.00638988, -0.00071592, -0.00279122, -0.00957115, -0.01288327, -0.00730937, -0.00107783, -0.00210156, -0.00890705, -0.01317668, -0.00813895, -0.00153491, -0.02128481, -0.04173044, -0.04831487, -0.0329319, -0.0052526, -0.01720322, -0.04052736, -0.05045706, -0.03607317, -0.0073803, -0.01341764, -0.03965629, -0.05151616, -0.03814886, -0.01005819, 0.18968274, 0.33063683, -0.01300105, -0.0137295, -0.04017465, 0.13727832, 0.36402234, 0.0102789, -0.01832107, -0.03365072, 0.08734506, 0.38194296, 0.04338228, -0.02525993, 0.56408125, 0.00458352, -0.01648227, -0.04887868, 0.2458552, 0.6202614, 0.04314807, -0.02213737, -0.04158014, 0.1663729, 0.6502702, 0.09621636, -0.03101388, -0.04082742, -0.00904519, -0.02790922, -0.02117818, 0.00798662, -0.03995711, -0.01243427, -0.02231705, -0.02946266, 0.00992055, -0.03600283, -0.0168492, -0.00111684, -0.00411204, -0.0129713, -0.01723725, -0.01022545, -0.00165306, -0.0031311, -0.01218016, -0.01763266, -0.0112562, -0.00231663, -0.01374149, -0.0379762, -0.05142937, -0.03117307, -0.00581914, -0.01064003, -0.03608089, -0.05272168, -0.0337567, -0.00795586, 0.09628104, 0.2712999, -0.00353779, -0.01734151, -0.03153981, 0.0568623, 0.28500998, 0.02230594, -0.02374955, 0.6821433, 0.05018048, -0.02320852, -0.04383616, 0.18459474, 0.71517974, 0.10805613, -0.03263677, -0.03637639, -0.01394373, -0.02511203, -0.01728636, 0.05407331, -0.02867568, -0.01893131, -0.00240854, -0.00446511, -0.01636187, -0.02377053, -0.01522848, -0.00333334, -0.00819975, -0.02964169, -0.04499287, -0.0274535, -0.00612408, 0.02727416, 0.194466, 0.00159832, -0.02232473, 0.74982506, 0.1145262, -0.03348048, -0.01605681, -0.02070339, -0.00458223},
				UpWeights:  nil,
				EncodedICC: nil,
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {

			var bitReader *jxlio.Bitreader
			if tc.readData {
				bitReader = testcommon.GenerateTestBitReader(t, `../testdata/unittest.jxl`)
				// skip first 40 bytes due to box headers.
				bitReader.Skip(40)
			} else {
				bitReader = jxlio.NewBitreader(bytes.NewReader(tc.data))
			}

			header, err := ParseImageHeader(bitReader, 5)
			if err != nil && !tc.expectErr {
				t.Errorf("got error when none was expected : %v", err)
			}
			if err == nil && tc.expectErr {
				t.Errorf("expected error but got none")
			}
			if err != nil && tc.expectErr {
				return
			}

			if !tc.expectErr && !reflect.DeepEqual(header.BitDepth, tc.expectedHeader.BitDepth) {
				t.Errorf("expected bitdepth %+v, got %+v", tc.expectedHeader, header)
			}

			if !tc.expectErr && !reflect.DeepEqual(header.Size, tc.expectedHeader.Size) {
				t.Errorf("expected size %+v, got %+v", tc.expectedHeader, header)
			}

			if !tc.expectErr && !reflect.DeepEqual(header.OpsinInverseMatrix, tc.expectedHeader.OpsinInverseMatrix) {
				t.Errorf("expected OpsinInverseMatrix %+v, got %+v", tc.expectedHeader, header)
			}

			if !tc.expectErr && !reflect.DeepEqual(header.ExtraChannelInfo, tc.expectedHeader.ExtraChannelInfo) {
				t.Errorf("expected ExtraChannelInfo %+v, got %+v", tc.expectedHeader, header)
			}

			if !tc.expectErr && !reflect.DeepEqual(header.UpWeights, tc.expectedHeader.UpWeights) {
				t.Errorf("expected UpWeights %+v, got %+v", tc.expectedHeader, header)
			}

			if !tc.expectErr && !reflect.DeepEqual(header.Up2Weights, tc.expectedHeader.Up2Weights) {
				t.Errorf("expected Up2Weights %+v, got %+v", tc.expectedHeader, header)
			}

			if !tc.expectErr && !reflect.DeepEqual(header.Up4Weights, tc.expectedHeader.Up4Weights) {
				t.Errorf("expected Up4Weights %+v, got %+v", tc.expectedHeader, header)
			}

			if !tc.expectErr && !reflect.DeepEqual(header.Up8Weights, tc.expectedHeader.Up8Weights) {
				t.Errorf("expected Up8Weights %+v, got %+v", tc.expectedHeader, header)
			}

			if !tc.expectErr && !reflect.DeepEqual(header.ToneMapping, tc.expectedHeader.ToneMapping) {
				t.Errorf("expected ToneMapping %+v, got %+v", tc.expectedHeader, header)
			}

			if !tc.expectErr && !reflect.DeepEqual(header.ColourEncoding, tc.expectedHeader.ColourEncoding) {
				t.Errorf("expected ColourEncoding %+v, got %+v", tc.expectedHeader, header)
			}
		})
	}
}
