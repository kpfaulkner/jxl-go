package image

const (
	TYPE_INT   = 0
	TYPE_FLOAT = 1
)

type ImageBuffer struct {
	Width      int32
	Height     int32
	bufferType int
	BufferType int

	// image data can be either float or int based. Keep separate buffers and just
	// reference each one as required. If conversion will be required then that might get
	// expensive, but will optimise/revisit later.
	FloatBuffer [][]float32
	IntBuffer   [][]int32
}

func NewImageBuffer(t int, height int32, width int32) *ImageBuffer {
	panic("not implemented")
}

func NewImageBufferFromInts(buffer [][]int32) *ImageBuffer {
	panic("not implemented")
}
func NewImageBufferFromFloats(buffer [][]float32) *ImageBuffer {
	panic("not implemented")
}

// Equals compares two ImageBuffers and returns true if they are equal.
func (ib *ImageBuffer) Equals(other ImageBuffer) bool {

	panic("not implemented")
	return true
}

func (ib *ImageBuffer) IsFloat() bool {
	return ib.bufferType == TYPE_FLOAT
}

func (ib *ImageBuffer) IsInt() bool {
	return ib.bufferType == TYPE_INT
}

func (ib *ImageBuffer) CastToFloatIfInt(maxValue int32) error {
	panic("not implemented")
}

func ImageBufferEquals(a []ImageBuffer, b []ImageBuffer) bool {
	panic("not implemented")
	return true
}
