package entropy

import (
	"github.com/kpfaulkner/jxl-go/jxlio"
	"github.com/kpfaulkner/jxl-go/util"
)

var level0Table = NewVLCTable(4, [][]int{{0, 2}, {4, 2}, {3, 2}, {2, 3}, {0, 2}, {4, 2}, {3, 2}, {1, 4}, {0, 2}, {4, 2}, {3, 2}, {2, 3}, {0, 2}, {4, 2}, {3, 2}, {5, 4}})
var codelenMap []int

type PrefixSymbolDistribution struct {
	*SymbolDistributionBase
	table         *VLCTable
	defaultSymbol int
}

func NewPrefixSymbolDistributionWithReader(reader *jxlio.Bitreader, alphabetSize int) (rcvr *PrefixSymbolDistribution) {
	rcvr = &PrefixSymbolDistribution{}
	rcvr.alphabetSize = alphabetSize
	rcvr.logAlphabetSize = util.CeilLog1p(int64(alphabetSize - 1))
	if rcvr.alphabetSize == 1 {
		rcvr.table = nil
		rcvr.defaultSymbol = 0
		return rcvr
	}

	hskip := reader.MustReadBits(2)
	if hskip == 1 {
		rcvr.populateSimplePrefix(reader)
	} else {
		rcvr.populateComplexPrefix(reader, int(hskip))
	}
	return
}
func (rcvr *PrefixSymbolDistribution) populateComplexPrefix(reader *jxlio.Bitreader, hskip int) error {

	panic("not implemented")
	//level1Lengths := make([]int, 18)
	//level1Codecounts := make([]int, 19)
	//level1Codecounts[0] = hskip
	//totalCode := 0
	//numCodes := 0
	//for i := hskip; i < 18; i++ {
	//	code := level0Table.GetVLC(reader)
	//	level1Lengths[codelenMap[i]] = code
	//
	//	level1Codecounts[code]++
	//	if code != 0 {
	//		totalCode += 32 >> code
	//		numCodes++
	//	}
	//	if totalCode >= 32 {
	//		level1Codecounts[0] += 17 - i
	//		break
	//	}
	//}
	//if totalCode != 32 && numCodes >= 2 || numCodes < 1 {
	//	return errors.New("Invalid Level 1 Prefix codes")
	//}
	//for i := 1; i < 19; i++ {
	//	level1Codecounts[i] += level1Codecounts[i-1]
	//}
	//level1LengthsScrambled := make([]int, 18)
	//level1Symbols := make([]int, 18)
	//for i := 17; i >= 0; i-- {
	//	level1Codecounts[level1Lengths[i]]--
	//	index := level1Codecounts[level1Lengths[i]]
	//	level1LengthsScrambled[index] = level1Lengths[i]
	//	level1Symbols[index] = i
	//}
	//
	//leve11Table := NewVLCTable(5, [][]int{level1LengthsScrambled, level1Symbols})
	//totalCode = 0
	//var prevRepeatCount int32
	//var prevZeroCount int32
	//level2Lengths := make([]int, rcvr.alphabetSize)
	//level2Symbols := make([]int, rcvr.alphabetSize)
	//level2Counts := make([]int, rcvr.alphabetSize+1)
	//prev := 8
	//for i := int32(0); i < int32(rcvr.alphabetSize); i++ {
	//	code := leve11Table.GetVLC(reader)
	//	if code == 16 {
	//		extra := 3 + reader.MustReadBits(2)
	//		if prevRepeatCount > 0 {
	//			extra = 4*(prevRepeatCount-2) - prevRepeatCount + extra
	//		}
	//		for j := int32(0); j < extra; j++ {
	//			level2Lengths[i+j] = prev
	//		}
	//		totalCode += int(uint32(32768)>>prev) * int(extra)
	//		i += extra - 1
	//		prevRepeatCount += extra
	//		prevZeroCount = 0
	//		level2Counts[prev] += int(extra)
	//	} else if code == 17 {
	//		extra := 3 + reader.MustReadBits(3)
	//		if prevZeroCount > 0 {
	//			extra = 8*(prevZeroCount-2) - prevZeroCount + extra
	//		}
	//		i += extra - 1
	//		prevRepeatCount = 0
	//		prevZeroCount += extra
	//		level2Counts[0] += int(extra)
	//	} else {
	//		level2Lengths[i] = code
	//		prevRepeatCount = 0
	//		prevZeroCount = 0
	//		if code != 0 {
	//			// uint32 casting due to in Java its using unsigned shift, right? Zero fill? (>>>).
	//			// Go recommendation is cast to uint32 first.
	//			totalCode += int(uint32(32768) >> code)
	//			prev = code
	//		}
	//		level2Counts[code]++
	//	}
	//	if totalCode >= 32768 {
	//		level2Counts[0] += rcvr.alphabetSize - int(i) - 1
	//		break
	//	}
	//}
	//if totalCode != 32768 && level2Counts[0] < rcvr.alphabetSize-1 {
	//	return errors.New("Invalid Level 2 Prefix Codes")
	//}
	//for i := 1; i <= rcvr.alphabetSize; i++ {
	//	level2Counts[i] += level2Counts[i-1]
	//}
	//level2LengthsScrambled := make([]int, rcvr.alphabetSize)
	//for i := rcvr.alphabetSize - 1; i >= 0; i-- {
	//	level2Counts[level2Lengths[i]]--
	//	index := level2Counts[level2Lengths[i]]
	//	level2LengthsScrambled[index] = level2Lengths[i]
	//	level2Symbols[index] = i
	//}
	//var err error
	//rcvr.table, err = NewVLCTableWithSymbols(15, level2LengthsScrambled, level2Symbols)
	//if err != nil {
	//	return err
	//}
	return nil
}

func (rcvr *PrefixSymbolDistribution) populateSimplePrefix(reader *jxlio.Bitreader) error {
	panic("not implemented")
	//symbols := make([]int, 4)
	//var lens []int = nil
	//nsym := 1 + reader.MustReadBits(2)
	//treeSelect := false
	//bits := 0
	//for i := 0; i < int(nsym); i++ {
	//	symbols[i] = int(reader.MustReadBits(rcvr.logAlphabetSize))
	//}
	//if nsym == 4 {
	//	treeSelect = reader.MustReadBool()
	//}
	//switch nsym {
	//case 1:
	//	rcvr.table = nil
	//	rcvr.defaultSymbol = symbols[0]
	//	return nil
	//case 2:
	//	bits = 1
	//	lens = []int{1, 1, 0, 0}
	//	if symbols[0] > symbols[1] {
	//		temp := symbols[1]
	//		symbols[1] = symbols[0]
	//		symbols[0] = temp
	//	}
	//case 3:
	//	bits = 2
	//	lens = []int{1, 2, 2, 0}
	//	if symbols[1] > symbols[2] {
	//		temp := symbols[2]
	//		symbols[2] = symbols[1]
	//		symbols[1] = temp
	//	}
	//case 4:
	//	if treeSelect {
	//		bits = 3
	//		lens = []int{1, 2, 3, 3}
	//		if symbols[2] > symbols[3] {
	//			temp := symbols[3]
	//			symbols[3] = symbols[2]
	//			symbols[2] = temp
	//		}
	//	} else {
	//		bits = 2
	//		lens = []int{2, 2, 2, 2}
	//		slices.Sort(symbols)
	//	}
	//}
	//var err error
	//rcvr.table, err = NewVLCTableWithSymbols(bits, lens, symbols)
	//if err != nil {
	//	return err
	//}
	return nil
}

func (rcvr *PrefixSymbolDistribution) ReadSymbol(reader *jxlio.Bitreader, state *ANSState) (int, error) {
	if rcvr.table == nil {
		return rcvr.defaultSymbol, nil
	}
	return rcvr.table.GetVLC(reader)
}
