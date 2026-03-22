package frame

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewNoiseGroupWithHeader(t *testing.T) {
	// Setup FrameHeader
	header := &FrameHeader{
		groupDim: 128,
		Width:    256,
		Height:   256,
	}
	
	// Setup noiseBuffer
	// 3 channels, 256x256
	noiseBuffer := make([][][]float32, 3)
	for i := 0; i < 3; i++ {
		noiseBuffer[i] = make([][]float32, 256)
		for j := 0; j < 256; j++ {
			noiseBuffer[i][j] = make([]float32, 256)
		}
	}
	
	seed0 := int64(12345)
	x0 := int32(0)
	y0 := int32(0)
	
	ng := NewNoiseGroupWithHeader(header, seed0, noiseBuffer, x0, y0)
	assert.NotNil(t, ng)
	assert.NotNil(t, ng.rng)
	
	// Check if noise buffer is populated with non-zero values
	// First few values should be generated.
	// XorShiro generates pseudo-random numbers, converted to float.
	// We just check if they are not all 0.
	
	hasNonZero := false
	for i := 0; i < 3; i++ {
		for j := 0; j < 128; j++ { // within group dim
			for k := 0; k < 128; k++ {
				if noiseBuffer[i][j][k] != 0 {
					hasNonZero = true
					break
				}
			}
		}
	}
	assert.True(t, hasNonZero, "Noise buffer should contain non-zero values")
}
