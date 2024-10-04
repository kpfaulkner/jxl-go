package core

import (
	"errors"
	"math"

	"github.com/kpfaulkner/jxl-go/bundle"
	"github.com/kpfaulkner/jxl-go/color"
	"github.com/kpfaulkner/jxl-go/entropy"
	"github.com/kpfaulkner/jxl-go/jxlio"
)

const (
	CODESTREAM_HEADER uint32 = 0x0AFF
)

var (
	DEFAULT_UP2 = []float32{
		-0.01716200, -0.03452303, -0.04022174, -0.02921014, -0.00624645,
		0.14111091, 0.28896755, 0.00278718, -0.01610267, 0.56661550,
		0.03777607, -0.01986694, -0.03144731, -0.01185068, -0.00213539}

	DEFAULT_UP4 = []float32{
		-0.02419067, -0.03491987, -0.03693351, -0.03094285, -0.00529785,
		-0.01663432, -0.03556863, -0.03888905, -0.03516850, -0.00989469,
		0.23651958, 0.33392945, -0.01073543, -0.01313181, -0.03556694,
		0.13048175, 0.40103025, 0.03951150, -0.02077584, 0.46914198,
		-0.00209270, -0.01484589, -0.04064806, 0.18942530, 0.56279892,
		0.06674400, -0.02335494, -0.03551682, -0.00754830, -0.02267919,
		-0.02363578, 0.00315804, -0.03399098, -0.01359519, -0.00091653,
		-0.00335467, -0.01163294, -0.01610294, -0.00974088, -0.00191622,
		-0.01095446, -0.03198464, -0.04455121, -0.02799790, -0.00645912,
		0.06390599, 0.22963888, 0.00630981, -0.01897349, 0.67537268,
		0.08483369, -0.02534994, -0.02205197, -0.01667999, -0.00384443}
	DEFAULT_UP8 = []float32{
		-0.02928613, -0.03706353, -0.03783812, -0.03324558, -0.00447632,
		-0.02519406, -0.03752601, -0.03901508, -0.03663285, -0.00646649,
		-0.02066407, -0.03838633, -0.04002101, -0.03900035, -0.00901973,
		-0.01626393, -0.03954148, -0.04046620, -0.03979621, -0.01224485,
		0.29895328, 0.35757708, -0.02447552, -0.01081748, -0.04314594,
		0.23903219, 0.41119301, -0.00573046, -0.01450239, -0.04246845,
		0.17567618, 0.45220643, 0.02287757, -0.01936783, -0.03583255,
		0.11572472, 0.47416733, 0.06284440, -0.02685066, 0.42720050,
		-0.02248939, -0.01155273, -0.04562755, 0.28689496, 0.49093869,
		-0.00007891, -0.01545926, -0.04562659, 0.21238920, 0.53980934,
		0.03369474, -0.02070211, -0.03866988, 0.14229550, 0.56593398,
		0.08045181, -0.02888298, -0.03680918, -0.00542229, -0.02920477,
		-0.02788574, -0.02118180, -0.03942402, -0.00775547, -0.02433614,
		-0.03193943, -0.02030828, -0.04044014, -0.01074016, -0.01930822,
		-0.03620399, -0.01974125, -0.03919545, -0.01456093, -0.00045072,
		-0.00360110, -0.01020207, -0.01231907, -0.00638988, -0.00071592,
		-0.00279122, -0.00957115, -0.01288327, -0.00730937, -0.00107783,
		-0.00210156, -0.00890705, -0.01317668, -0.00813895, -0.00153491,
		-0.02128481, -0.04173044, -0.04831487, -0.03293190, -0.00525260,
		-0.01720322, -0.04052736, -0.05045706, -0.03607317, -0.00738030,
		-0.01341764, -0.03965629, -0.05151616, -0.03814886, -0.01005819,
		0.18968273, 0.33063684, -0.01300105, -0.01372950, -0.04017465,
		0.13727832, 0.36402234, 0.01027890, -0.01832107, -0.03365072,
		0.08734506, 0.38194295, 0.04338228, -0.02525993, 0.56408126,
		0.00458352, -0.01648227, -0.04887868, 0.24585519, 0.62026135,
		0.04314807, -0.02213737, -0.04158014, 0.16637289, 0.65027023,
		0.09621636, -0.03101388, -0.04082742, -0.00904519, -0.02790922,
		-0.02117818, 0.00798662, -0.03995711, -0.01243427, -0.02231705,
		-0.02946266, 0.00992055, -0.03600283, -0.01684920, -0.00111684,
		-0.00411204, -0.01297130, -0.01723725, -0.01022545, -0.00165306,
		-0.00313110, -0.01218016, -0.01763266, -0.01125620, -0.00231663,
		-0.01374149, -0.03797620, -0.05142937, -0.03117307, -0.00581914,
		-0.01064003, -0.03608089, -0.05272168, -0.03375670, -0.00795586,
		0.09628104, 0.27129991, -0.00353779, -0.01734151, -0.03153981,
		0.05686230, 0.28500998, 0.02230594, -0.02374955, 0.68214326,
		0.05018048, -0.02320852, -0.04383616, 0.18459474, 0.71517975,
		0.10805613, -0.03263677, -0.03637639, -0.01394373, -0.02511203,
		-0.01728636, 0.05407331, -0.02867568, -0.01893131, -0.00240854,
		-0.00446511, -0.01636187, -0.02377053, -0.01522848, -0.00333334,
		-0.00819975, -0.02964169, -0.04499287, -0.02745350, -0.00612408,
		0.02727416, 0.19446600, 0.00159832, -0.02232473, 0.74982506,
		0.11452620, -0.03348048, -0.01605681, -0.02070339, -0.00458223}
)

type ImageHeader struct {
	level           int32
	size            *Dimension
	orientation     uint32
	intrinsicSize   *Dimension
	previewHeader   *PreviewHeader
	animationHeader *AnimationHeader
	bitDepth        *BitDepthHeader

	orientedWidth       uint32
	orientedHeight      uint32
	modular16BitBuffers bool

	extraChannelInfo []ExtraChannelInfo
	xybEncoded       bool
	colorEncoding    *color.ColorEncodingBundle
	alphaIndices     []int32

	toneMapping        *color.ToneMapping
	extensions         *bundle.Extensions
	opsinInverseMatrix *color.OpsinInverseMatrix

	up2Weights []float32
	up4Weights []float32
	up8Weights []float32

	encodedICC []byte
}

func NewImageHeader() *ImageHeader {
	ih := &ImageHeader{}
	return ih
}

func ParseImageHeader(reader *jxlio.Bitreader, level int32) (*ImageHeader, error) {
	header := NewImageHeader()

	headerBits := reader.MustReadBits(16)
	if uint32(headerBits) != CODESTREAM_HEADER {
		return nil, errors.New("Not a JXL codestream: 0xFF0A magic mismatch")
	}

	err := header.setLevel(level)
	if err != nil {
		return nil, err
	}

	header.size, err = readSizeHeader(reader, level)
	if err != nil {
		return nil, err
	}

	allDefault := reader.MustReadBool()
	extraFields := false
	if !allDefault {
		extraFields = reader.MustReadBool()
	}

	if extraFields {
		header.orientation = 1 + uint32(reader.MustReadBits(3))
		if reader.MustReadBool() {
			header.intrinsicSize, err = readSizeHeader(reader, level)
			if err != nil {
				return nil, err
			}
		}
		if reader.MustReadBool() {
			header.previewHeader, err = NewPreviewHeader(reader)
			if err != nil {
				return nil, err
			}
		}
		if reader.MustReadBool() {
			header.animationHeader, err = NewAnimationHeader(reader)
			if err != nil {
				return nil, err
			}
		}
	} else {
		header.orientation = 1
	}

	if header.orientation > 4 {
		header.orientedWidth = header.size.height
		header.orientedHeight = header.size.width
	} else {
		header.orientedWidth = header.size.width
		header.orientedHeight = header.size.height
	}

	if allDefault {
		header.bitDepth = NewBitDepthHeader()
		header.modular16BitBuffers = true
		header.extraChannelInfo = []ExtraChannelInfo{}
		header.xybEncoded = true
		header.colorEncoding, err = color.NewColorEncodingBundle()
		if err != nil {
			return nil, err
		}
	} else {
		header.bitDepth = NewBitDepthHeaderWithReader(reader)
		header.modular16BitBuffers = reader.MustReadBool()
		extraChannelCount := reader.MustReadU32(0, 0, 1, 0, 2, 4, 1, 12)
		header.extraChannelInfo = make([]ExtraChannelInfo, extraChannelCount)
		alphaIndicies := make([]int32, extraChannelCount)
		numAlphaChannels := 0

		for i := 0; i < int(extraChannelCount); i++ {
			eci, err := NewExtraChannelInfoWithReader(reader)
			if err != nil {
				return nil, err
			}
			header.extraChannelInfo[i] = *eci

			if header.extraChannelInfo[i].ecType == bundle.ALPHA {
				alphaIndicies[numAlphaChannels] = int32(i)
				numAlphaChannels++
			}
		}
		header.alphaIndices = make([]int32, numAlphaChannels)
		copy(header.alphaIndices, alphaIndicies[:numAlphaChannels])
		header.xybEncoded = reader.MustReadBool()
		header.colorEncoding, err = color.NewColorEncodingBundleWithReader(reader)
		if err != nil {
			return nil, err
		}
	}

	if extraFields {
		header.toneMapping, err = color.NewToneMappingWithReader(reader)
		if err != nil {
			return nil, err
		}
	} else {
		header.toneMapping = color.NewToneMapping()
	}

	if allDefault {
		header.extensions = bundle.NewExtensions()
	} else {
		header.extensions, err = bundle.NewExtensionsWithReader(reader)
		if err != nil {
			return nil, err
		}
	}

	defaultMatrix := reader.MustReadBool()
	if !defaultMatrix && header.xybEncoded {
		header.opsinInverseMatrix = color.NewOpsinInverseMatrixWithReader(reader)
	} else {
		header.opsinInverseMatrix = color.NewOpsinInverseMatrix()
	}

	var cwMask int32
	if defaultMatrix {
		cwMask = 0
	} else {
		cwMask = int32(reader.MustReadBits(3))
	}

	if cwMask&1 != 0 {
		header.up2Weights = make([]float32, 15)
		for i := 0; i < len(header.up2Weights); i++ {
			header.up2Weights[i] = reader.MustReadF16()
		}
	} else {
		header.up2Weights = DEFAULT_UP2
	}

	if cwMask&2 != 0 {
		header.up4Weights = make([]float32, 55)
		for i := 0; i < len(header.up4Weights); i++ {
			header.up4Weights[i] = reader.MustReadF16()
		}
	} else {
		header.up4Weights = DEFAULT_UP4
	}

	if cwMask&4 != 0 {
		header.up8Weights = make([]float32, 210)
		for i := 0; i < len(header.up8Weights); i++ {
			header.up8Weights[i] = reader.MustReadF16()
		}
	} else {
		header.up8Weights = DEFAULT_UP8
	}

	if header.colorEncoding.UseIccProfile {
		encodedSize := reader.MustReadU64()

		// check MaxUint32 or MaxInt32
		if encodedSize > math.MaxUint32 {
			return nil, errors.New("Invalid encoded size")
		}
		header.encodedICC = make([]byte, encodedSize)
		iccDistribution, err := entropy.NewEntropyStreamWithReaderAndNumDists(reader, 41)
		if err != nil {
			return nil, err
		}
		for i := 0; i < int(encodedSize); i++ {
			cc, err := iccDistribution.ReadSymbol(reader, GetICCContext(header.encodedICC, i))
			if err != nil {
				return nil, err
			}
			header.encodedICC[i] = byte(cc)
		}
		if !iccDistribution.ValidateFinalState() {
			return nil, errors.New("ICC Stream")
		}
	}
	reader.ZeroPadToByte()

	return header, nil
}
func (h *ImageHeader) getColourChannelCount() int {
	if h.colorEncoding.ColorEncoding == color.CE_GRAY {
		return 1
	}

	return 3
}

func (h *ImageHeader) GetSize() (uint32, uint32) {
	return h.size.width, h.size.height
}

func (h *ImageHeader) GetColourModel() int32 {
	return h.colorEncoding.ColorEncoding
}

func (h *ImageHeader) setLevel(level int32) error {
	if level != 5 && level != 10 {
		return errors.New("invalid bitstream")
	}
	h.level = level
	return nil
}

func (h *ImageHeader) hasAlpha() bool {
	return len(h.alphaIndices) > 0
}

func (h *ImageHeader) getTotalChannelCount() int {
	return len(h.extraChannelInfo) + h.getColourChannelCount()
}

func (h *ImageHeader) getDecodedICC() []byte {
	return nil

	// TODO(kpfaulkner) NOT IMPLEMENTED YET... but test images do not have ICC component... so skipping for now.
}

func GetICCContext(buffer []byte, index int) int {
	if index <= 128 {
		return 0
	}
	b1 := int(buffer[index-1]) & 0xFF
	b2 := int(buffer[index-2]) & 0xFF
	var p1 int
	var p2 int
	if b1 >= 'a' && b1 <= 'z' || b1 >= 'A' && b1 <= 'Z' {
		p1 = 0
	} else if b1 >= '0' && b1 <= '9' || b1 == '.' || b1 == ',' {
		p1 = 1
	} else if b1 <= 1 {
		p1 = 2 + b1
	} else if b1 > 1 && b1 < 16 {
		p1 = 4
	} else if b1 > 240 && b1 < 255 {
		p1 = 5
	} else if b1 == 255 {
		p1 = 6
	} else {
		p1 = 7
	}
	if b2 >= 'a' && b2 <= 'z' || b2 >= 'A' && b2 <= 'Z' {
		p2 = 0
	} else if b2 >= '0' && b2 <= '9' || b2 == '.' || b2 == ',' {
		p2 = 1
	} else if b2 < 16 {
		p2 = 2
	} else if b2 > 240 {
		p2 = 3
	} else {
		p2 = 4
	}
	return 1 + p1 + 8*p2
}
