package core

import (
	"bytes"
	"testing"

	"github.com/kpfaulkner/jxl-go/colour"
	"github.com/stretchr/testify/assert"
)

func TestWriteSRGB(t *testing.T) {
	testCases := []struct {
		name           string
		intent         int32
		expectedIntent byte
	}{
		{
			name:           "Relative",
			intent:         colour.RI_RELATIVE,
			expectedIntent: 0x01,
		},
		{
			name:           "Perceptual",
			intent:         colour.RI_PERCEPTUAL,
			expectedIntent: 0x00,
		},
		{
			name:           "Saturation",
			intent:         colour.RI_SATURATION,
			expectedIntent: 0x02,
		},
		{
			name:           "Absolute",
			intent:         colour.RI_ABSOLUTE,
			expectedIntent: 0x03,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			writer := &PNGWriter{}
			img := &JXLImage{}
			img.imageHeader.ColourEncoding = &colour.ColourEncodingBundle{
				RenderingIntent: tc.intent,
			}

			var buf bytes.Buffer
			err := writer.writeSRGB(img, &buf)
			assert.Nil(t, err)

			output := buf.Bytes()
			// PNG sRGB chunk: 4 bytes length, 4 bytes "sRGB", 1 byte intent, 4 bytes CRC
			assert.Equal(t, 13, len(output))
			
			// Length should be 1
			assert.Equal(t, byte(0), output[0])
			assert.Equal(t, byte(0), output[1])
			assert.Equal(t, byte(0), output[2])
			assert.Equal(t, byte(1), output[3])

			// Chunk type
			assert.Equal(t, []byte("sRGB"), output[4:8])

			// Intent
			assert.Equal(t, tc.expectedIntent, output[8], "Intent mismatch for %s", tc.name)
		})
	}
}
