package core

// JXLImage contains the core information about the JXL image.
type JXLImage struct {
	width  int
	height int
}

func NewJXLImageWithBuffer(buffer [][][]float32, header ImageHeader) (*JXLImage, error) {
	jxl := &JXLImage{}
	jxl.width = len(buffer[0][0])
	jxl.height = len(buffer[0])
	
	return jxl, nil
}
