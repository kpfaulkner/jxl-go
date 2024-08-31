package entropy

import (
	"errors"
	bbits "math/bits"

	"github.com/kpfaulkner/jxl-go/jxlio"
	"github.com/kpfaulkner/jxl-go/util"
)

type VLCTable struct {
	table [][]int
	bits  int
}

func NewVLCTable(bits int, table [][]int) (rcvr *VLCTable) {
	rcvr = &VLCTable{}
	rcvr.bits = bits
	rcvr.table = table
	return
}
func NewVLCTableWithSymbols(bits int, lengths []int, symbols []int) (*VLCTable, error) {
	rcvr := &VLCTable{}
	rcvr.bits = bits
	table := util.MakeMatrix2D[int](1<<bits, 2)
	codes := make([]int, len(lengths))
	nLengths := make([]int, len(lengths))
	nSymbols := make([]int, len(lengths))
	count := 0
	code := 0
	for i := 0; i < len(lengths); i++ {
		currentLen := lengths[i]
		if currentLen > 0 {
			nLengths[count] = currentLen
			if len(symbols) > 0 {
				nSymbols[count] = symbols[i]
			} else {
				nSymbols[count] = i
			}
			codes[count] = int(code)
			count++
		} else if currentLen < 0 {
			currentLen = -currentLen
		} else {
			continue
		}
		code += 1 << (32 - currentLen)
		if code > 1<<32 {
			return nil, errors.New("Too many VLC codes")
		}
	}
	if code != 1<<32 {
		return nil, errors.New("Not enough VLC codes")
	}
	for i := 0; i < count; i++ {
		if nLengths[i] <= bits {
			index := bbits.Reverse(uint(codes[i]))
			number := 1 << (bits - nLengths[i])
			offset := 1 << nLengths[i]
			for j := 0; j < number; j++ {
				oldSymbol := table[index][0]
				oldLen := table[index][1]
				if (oldLen > 0 || oldSymbol > 0) && (oldLen != nLengths[i] || oldSymbol != nSymbols[i]) {
					return nil, errors.New("Illegal VLC codes")
				}
				table[index][0] = nSymbols[i]
				table[index][1] = nLengths[i]
				index += uint(offset)
			}
		} else {
			return nil, errors.New("Table size too small")
		}
	}
	for i := 0; i < len(table); i++ {
		if table[i][1] == 0 {
			table[i][0] = -1
		}
	}
	rcvr.table = table
	return rcvr, nil
}
func (rcvr *VLCTable) GetVLC(reader *jxlio.Bitreader) (int, error) {

	index := reader.MustShowBits(rcvr.bits)
	symbol := rcvr.table[index][0]
	length := rcvr.table[index][1]
	err := reader.SkipBits(uint32(length))
	if err != nil {
		return 0, err
	}

	return symbol, nil
}
