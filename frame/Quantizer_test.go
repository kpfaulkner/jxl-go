package frame

import (
	"bytes"
	"testing"

	"github.com/kpfaulkner/jxl-go/jxlio"
	"github.com/kpfaulkner/jxl-go/testcommon"
	"github.com/stretchr/testify/assert"
)

func TestNewQuantizerWithReader(t *testing.T) {
	bw := testcommon.NewBitWriter()

	// globalScale = 1
	bw.WriteBits(0, 2)  // Selector 00
	bw.WriteBits(0, 11) // Offset 0

	// quantLF = 16
	bw.WriteBits(0, 2) // Selector 00
	// Offset 0 (0 bits)

	br := jxlio.NewBitStreamReader(bytes.NewReader(bw.Bytes()))
	lfDequant := []float32{1.0, 1.0, 1.0}

	q, err := NewQuantizerWithReader(br, lfDequant)
	assert.NoError(t, err)
	assert.NotNil(t, q)
	assert.Equal(t, uint32(1), q.globalScale)
	assert.Equal(t, uint32(16), q.quantLF)

	// Check scaledDequant
	// (1<<16) * 1.0 / (1 * 16) = 65536 / 16 = 4096
	assert.Equal(t, float32(4096.0), q.scaledDequant[0])
}
