package color

const (
	PRI_SRGB   int32 = 1
	PRI_CUSTOM int32 = 2
	PRI_BT2100 int32 = 9
	PRI_P3     int32 = 11

	WP_D50    int32 = -1
	WP_D65    int32 = 1
	WP_CUSTOM int32 = 2
	WP_E      int32 = 10
	WP_DCI    int32 = 11

	CE_RGB     int32 = 0
	CE_GRAY    int32 = 1
	CE_XYB     int32 = 2
	CE_UNKNOWN int32 = 3

	RI_PERCEPTUAL uint32 = 0
	RI_RELATIVE   uint32 = 1
	RI_SATURATION uint32 = 2
	RI_ABSOLUTE   uint32 = 3

	TF_BT709   uint32 = 1 + (1 << 24)
	TF_UNKNOWN uint32 = 2 + (1 << 24)
	TF_LINEAR  uint32 = 8 + (1 << 24)
	TF_SRGB    uint32 = 13 + (1 << 24)
	TF_PQ      uint32 = 16 + (1 << 24)
	TF_DCI     uint32 = 17 + (1 << 24)
	TF_HLG     uint32 = 18 + (1 << 24)
)

func ValidateColorEncoding(colorEncoding int32) bool {
	return colorEncoding >= 0 && colorEncoding <= 3
}

func ValidateWhitePoint(whitePoint int32) bool {
	return whitePoint == WP_D65 || whitePoint == WP_CUSTOM || whitePoint == WP_E || whitePoint == WP_DCI
}
