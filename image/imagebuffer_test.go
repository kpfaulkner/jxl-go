package image

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewImageBuffer(t *testing.T) {
	buf, err := NewImageBuffer(TYPE_INT, 5, 5)
	assert.Nil(t, err)

	assert.NotNil(t, buf)
}

func TestNewImageBufferFromInts(t *testing.T) {
	origBuf := [][]int32{{1, 2, 3}, {4, 5, 6}, {7, 8, 9}}
	buf := NewImageBufferFromInts(origBuf)
	assert.NotNil(t, buf)
}

func TestNewImageBufferFromFloats(t *testing.T) {
	origBuf := [][]float32{{1, 2, 3}, {4, 5, 6}, {7, 8, 9}}
	buf := NewImageBufferFromFloats(origBuf)
	assert.NotNil(t, buf)
}

func TestNewImageBufferFromImageBuffer(t *testing.T) {
	origBuf := [][]int32{{1, 2, 3}, {4, 5, 6}, {7, 8, 9}}
	buf := NewImageBufferFromInts(origBuf)

	buf2 := NewImageBufferFromImageBuffer(buf)
	assert.NotNil(t, buf2)
}

func TestEqualsInt(t *testing.T) {
	origBuf := [][]int32{{1, 2, 3}, {4, 5, 6}, {7, 8, 9}}
	buf := NewImageBufferFromInts(origBuf)
	buf2 := NewImageBufferFromInts(origBuf)

	if !buf.Equals(*buf2) {
		t.Failed()
	}
}

func TestEqualsFloat(t *testing.T) {
	origBuf := [][]float32{{1, 2, 3}, {4, 5, 6}, {7, 8, 9}}
	buf := NewImageBufferFromFloats(origBuf)
	buf2 := NewImageBufferFromFloats(origBuf)

	if !buf.Equals(*buf2) {
		t.Failed()
	}
}

func TestCastToFloat(t *testing.T) {
	origBuf := [][]int32{{1, 2, 3}, {4, 5, 6}, {7, 8, 9}}
	buf := NewImageBufferFromInts(origBuf)

	// confirm floats are nil to begin with
	assert.Nil(t, buf.FloatBuffer)
	err := buf.castToFloatBuffer(10)
	assert.Nil(t, err)
	assert.NotNil(t, buf.FloatBuffer)
}

func TestCastToInt(t *testing.T) {
	origBuf := [][]float32{{1, 2, 3}, {4, 5, 6}, {7, 8, 9}}
	buf := NewImageBufferFromFloats(origBuf)

	// confirm int are nil to begin with
	assert.Nil(t, buf.IntBuffer)
	err := buf.castToIntBuffer(10)
	assert.Nil(t, err)
	assert.NotNil(t, buf.IntBuffer)
}

func TestCastToIntIfFloat(t *testing.T) {
	origBuf := [][]float32{{1, 2, 3}, {4, 5, 6}, {7, 8, 9}}
	buf := NewImageBufferFromFloats(origBuf)

	err := buf.CastToIntIfMax(10)
	assert.Nil(t, err)
}

func TestCastToFloatIfInt(t *testing.T) {
	origBuf := [][]int32{{1, 2, 3}, {4, 5, 6}, {7, 8, 9}}
	buf := NewImageBufferFromInts(origBuf)

	err := buf.CastToFloatIfMax(10)
	assert.Nil(t, err)
}

func TestClamp(t *testing.T) {
	origBuf := [][]int32{{1, 2, 3}, {4, 5, 6}, {7, 8, 9}}
	buf := NewImageBufferFromInts(origBuf)

	err := buf.Clamp(10)
	assert.Nil(t, err)
}

func TestImageBufferSliceEquals(t *testing.T) {
	origBuf := [][]int32{{1, 2, 3}, {4, 5, 6}, {7, 8, 9}}
	buf := NewImageBufferFromInts(origBuf)
	buf2 := NewImageBufferFromInts(origBuf)

	if !ImageBufferSliceEquals([]ImageBuffer{*buf}, []ImageBuffer{*buf2}) {
		t.Failed()
	}
}
