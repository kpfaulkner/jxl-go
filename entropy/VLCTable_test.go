package entropy

import (
	"bytes"
	"testing"

	"github.com/kpfaulkner/jxl-go/jxlio"
	"github.com/stretchr/testify/assert"
)

func TestVLCTable(t *testing.T) {
	// Test NewVLCTable
	tableData := [][]int32{{5, 1}, {10, 1}}
	vlc := NewVLCTable(1, tableData)
	assert.Equal(t, int32(1), vlc.bits)
	assert.Equal(t, tableData, vlc.table)

	// Test GetVLC
	data := []byte{0x00} // bit 0 -> index 0 -> symbol 5
	br := jxlio.NewBitStreamReader(bytes.NewReader(data))
	sym, err := vlc.GetVLC(br)
	assert.NoError(t, err)
	assert.Equal(t, int32(5), sym)

	data = []byte{0x01} // bit 1 -> index 1 -> symbol 10
	br = jxlio.NewBitStreamReader(bytes.NewReader(data))
	sym, err = vlc.GetVLC(br)
	assert.NoError(t, err)
	assert.Equal(t, int32(10), sym)
}

func TestVLCTableWithSymbols(t *testing.T) {
	// 2 symbols, length 1 each (covers all 2^1 codes)
	lengths := []int32{1, 1}
	symbols := []int32{100, 200}
	vlc, err := NewVLCTableWithSymbols(1, lengths, symbols)
	assert.NoError(t, err)
	assert.NotNil(t, vlc)

	// Test GetVLC
	data := []byte{0x00} // symbol 100
	br := jxlio.NewBitStreamReader(bytes.NewReader(data))
	sym, err := vlc.GetVLC(br)
	assert.NoError(t, err)
	assert.Equal(t, int32(100), sym)

	data = []byte{0x01} // symbol 200
	br = jxlio.NewBitStreamReader(bytes.NewReader(data))
	sym, err = vlc.GetVLC(br)
	assert.NoError(t, err)
	assert.Equal(t, int32(200), sym)
}

func TestVLCTableWithSymbolsError(t *testing.T) {
	// Not enough codes
	lengths := []int32{1} // only half of the space for bits=1
	symbols := []int32{100}
	_, err := NewVLCTableWithSymbols(1, lengths, symbols)
	assert.Error(t, err)
	assert.Equal(t, "Not enough VLC codes", err.Error())

	// Too many codes
	lengths = []int32{1, 1, 1}
	symbols = []int32{100, 200, 300}
	_, err = NewVLCTableWithSymbols(1, lengths, symbols)
	assert.Error(t, err)
	assert.Equal(t, "Too many VLC codes", err.Error())
}
