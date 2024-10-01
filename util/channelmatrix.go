package util

type ChannelData[T any] struct {
	data     []T
	dim2Size int
	dim3Size int
}

// ChannelMatrix will have first dimension as used to identify channel
// but then dimensions 2 and 3 are flattened into a single one.
// This is a compromise between entirely flattening out a 3D matrix
// and being able to process individual channels separately.
type ChannelMatrix[T any] struct {
	data       []ChannelData[T]
	channelDim int
	dim2Size   int
	dim3Size   int
}

func NewChannelData[T any](dim2Size int, dim3Size int) *ChannelData[T] {
	d := &ChannelData[T]{
		data:     make([]T, dim2Size*dim3Size),
		dim2Size: dim2Size,
		dim3Size: dim3Size,
	}
	return d
}

func (m *ChannelData[T]) Get(dim2Index int, dim3Index int) T {
	return m.data[dim2Index*m.dim2Size+dim3Index]
}

func (m *ChannelData[T]) Set(dim2Index int, dim3Index int, val T) error {
	m.data[dim2Index*m.dim2Size+dim3Index] = val
	return nil
}

func (m *ChannelData[T]) GetDim2(dim2Index int) []T {
	return m.data[dim2Index*m.dim2Size : dim2Index*m.dim2Size+m.dim2Size]
}

func (m *ChannelData[T]) SetDim2(dim2Index int, val []T) error {
	temp := append(m.data[:dim2Index*m.dim2Size], val...)
	m.data = append(temp, m.data[:dim2Index*m.dim2Size+m.dim2Size]...)
	return nil
}

func NewChannelMatrix3D[T any](channelSize int, dim2Size int, dim3Size int) *ChannelMatrix[T] {
	m := &ChannelMatrix[T]{
		data:     make([]ChannelData[T], channelSize),
		dim2Size: dim2Size,
		dim3Size: dim3Size,
	}

	for i, _ := range m.data {
		m.data[i] = *NewChannelData[T](dim2Size, dim3Size)
	}
	return m
}

func (m *ChannelMatrix[T]) Get(channelIndex int, dim2Index int, dim3Index int) T {
	ch := m.data[channelIndex]
	return ch.Get(dim2Index, dim3Index)
}

func (m *ChannelMatrix[T]) Set(channelIndex int, dim2Index int, dim3Index int, val T) error {
	ch := m.data[channelIndex]
	ch.Set(dim2Index, dim3Index, val)
	return nil
}

func (m *ChannelMatrix[T]) SetDataForChannel(channelIndex int, val ChannelData[T]) error {
	m.data[channelIndex] = val
	return nil
}

func (m *ChannelMatrix[T]) GetDataForChannel(ch int) ChannelData[T] {
	return m.data[ch]
}

func (m *ChannelMatrix[T]) NumberOfChannels() int {
	return m.channelDim
}

func (m *ChannelMatrix[T]) Dim2() int {
	return m.dim2Size
}

func (m *ChannelMatrix[T]) Dim3() int {
	return m.dim3Size
}

func (m *ChannelMatrix[T]) CreateChannelOfSize(chSize int) []T {
	return make([]T, chSize)
}
