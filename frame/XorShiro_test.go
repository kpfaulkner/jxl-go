package frame

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestNewXorShiroWith4Seeds(t *testing.T) {
	xs := NewXorShiroWith4Seeds(1, 2, 3, 4)
	assert.NotNil(t, xs)
	assert.Equal(t, 8, len(xs.state0))
	assert.Equal(t, 8, len(xs.state1))
	assert.Equal(t, 8, len(xs.batch))
}

func TestXorShiroNextLong(t *testing.T) {
	xs := NewXorShiroWith2Seeds(12345, 67890)
	l1 := xs.nextLong()
	l2 := xs.nextLong()
	assert.NotEqual(t, l1, l2)
}

func TestXorShiroFill(t *testing.T) {
	xs := NewXorShiroWith2Seeds(1, 2)
	bits := make([]int64, 10)
	xs.fill(bits)
	for _, b := range bits {
		assert.NotEqual(t, int64(0), b)
	}
}
