package entropy

import (
	"errors"
	"fmt"

	"github.com/kpfaulkner/jxl-go/jxlio"
	"github.com/kpfaulkner/jxl-go/util"
)

var distPrefixTable = NewVLCTable(7, [][]int{{10, 3}, {12, 7}, {7, 3}, {3, 4}, {6, 3}, {8, 3}, {9, 3}, {5, 4}, {10, 3}, {4, 4}, {7, 3}, {1, 4}, {6, 3}, {8, 3}, {9, 3}, {2, 4}, {10, 3}, {0, 5}, {7, 3}, {3, 4}, {6, 3}, {8, 3}, {9, 3}, {5, 4}, {10, 3}, {4, 4}, {7, 3}, {1, 4}, {6, 3}, {8, 3}, {9, 3}, {2, 4}, {10, 3}, {11, 6}, {7, 3}, {3, 4}, {6, 3}, {8, 3}, {9, 3}, {5, 4}, {10, 3}, {4, 4}, {7, 3}, {1, 4}, {6, 3}, {8, 3}, {9, 3}, {2, 4}, {10, 3}, {0, 5}, {7, 3}, {3, 4}, {6, 3}, {8, 3}, {9, 3}, {5, 4}, {10, 3}, {4, 4}, {7, 3}, {1, 4}, {6, 3}, {8, 3}, {9, 3}, {2, 4}, {10, 3}, {13, 7}, {7, 3}, {3, 4}, {6, 3}, {8, 3}, {9, 3}, {5, 4}, {10, 3}, {4, 4}, {7, 3}, {1, 4}, {6, 3}, {8, 3}, {9, 3}, {2, 4}, {10, 3}, {0, 5}, {7, 3}, {3, 4}, {6, 3}, {8, 3}, {9, 3}, {5, 4}, {10, 3}, {4, 4}, {7, 3}, {1, 4}, {6, 3}, {8, 3}, {9, 3}, {2, 4}, {10, 3}, {11, 6}, {7, 3}, {3, 4}, {6, 3}, {8, 3}, {9, 3}, {5, 4}, {10, 3}, {4, 4}, {7, 3}, {1, 4}, {6, 3}, {8, 3}, {9, 3}, {2, 4}, {10, 3}, {0, 5}, {7, 3}, {3, 4}, {6, 3}, {8, 3}, {9, 3}, {5, 4}, {10, 3}, {4, 4}, {7, 3}, {1, 4}, {6, 3}, {8, 3}, {9, 3}, {2, 4}})

type ANSSymbolDistribution struct {
	*SymbolDistributionBase
	frequencies []int
	cutoffs     []int
	symbols     []int
	offsets     []int
}

func NewANSSymbolDistribution(reader *jxlio.Bitreader, logAlphabetSize int) (*ANSSymbolDistribution, error) {
	asd := &ANSSymbolDistribution{}
	asd.logAlphabetSize = logAlphabetSize
	uniqPos := -1
	if reader.TryReadBool() {
		//asd.asd = 1 << logAlphabetSize
		//asd.frequencies = make([]int, alphabetSize)
		if reader.TryReadBool() {
			v1 := reader.ReadU8()
			v2 := reader.ReadU8()
			if v1 == v2 {
				return nil, errors.New("Overlapping dual peak distribution")
			}
			asd.alphabetSize = 1 + util.Max(v1, v2)
			if asd.alphabetSize > (1 << asd.logAlphabetSize) {
				return nil, errors.New(fmt.Sprintf("Illegal Alphabet size : %d", asd.alphabetSize))
			}
			asd.frequencies = make([]int, asd.alphabetSize)
			asd.frequencies[v1] = int(reader.TryReadBits(12))
			asd.frequencies[v2] = 1<<12 - asd.frequencies[v1]
			if asd.frequencies[v1] == 0 {
				uniqPos = v2
			}
		} else {
			x := reader.ReadU8()
			if x >= len(asd.frequencies) {
				return nil, errors.New("Invalid frequency position")
			}
			asd.frequencies[x] = 1 << 12
			uniqPos = x
		}
	} else if reader.TryReadBool() {
		asd.alphabetSize = 1 + reader.ReadU8()
		if asd.alphabetSize == 1 {
			uniqPos = 0
		}
		asd.frequencies = make([]int, asd.alphabetSize)
		for i := 0; i < asd.alphabetSize; i++ {
			asd.frequencies[i] = (1 << 12) / asd.alphabetSize
		}
		for i := 0; i < (1<<12)%asd.alphabetSize; i++ {
			asd.frequencies[i]++
		}
	} else {
		var l int
		for l = 0; l < 3; l++ {
			if !reader.TryReadBool() {
				break
			}
		}
		shift := reader.TryReadBits(uint64(l)) | 1<<l - 1
		if shift > 13 {
			return nil, errors.New("Shift > 13")
		}
		asd.alphabetSize = 3 + reader.ReadU8()
		if asd.alphabetSize > (1 << asd.logAlphabetSize) {
			return nil, errors.New(fmt.Sprintf("Illegal Alphabet size : %d", asd.alphabetSize))
		}

		asd.frequencies = make([]int, asd.alphabetSize)
		logCounts := make([]int, asd.alphabetSize)
		same := make([]int, asd.alphabetSize)
		omitLog := -1
		omitPos := -1
		var err error
		for i := 0; i < asd.alphabetSize; i++ {
			logCounts[i], err = distPrefixTable.GetVLC(reader)
			if err != nil {
				return nil, err
			}
			if logCounts[i] == 13 {
				rle := reader.ReadU8()
				same[i] = rle + 5
				i += rle + 3
				continue
			}
			if logCounts[i] > omitLog {
				omitLog = logCounts[i]
				omitPos = i
			}
		}
		if omitPos < 0 || omitPos+1 < asd.alphabetSize && logCounts[omitPos+1] == 13 {
			return nil, errors.New("Invalid OmitPos")
		}
		totalCount := 0
		numSame := 0
		prev := 0
		for i := 0; i < asd.alphabetSize; i++ {
			if same[i] != 0 {
				numSame = same[i] - 1
				if i > 0 {
					prev = asd.frequencies[i-1]
				} else {
					prev = 0
				}
			}
			if numSame != 0 {
				asd.frequencies[i] = prev
				numSame--
			} else {
				if i == omitPos || logCounts[i] == 0 {
					continue
				}
				if logCounts[i] == 1 {
					asd.frequencies[i] = 1
				} else {
					bitcount := uint32(shift) - uint32(12-logCounts[i]+1)>>1
					if bitcount < 0 {
						bitcount = 0
					}
					if bitcount > uint32(logCounts[i])-1 {
						bitcount = uint32(logCounts[i])
					}
					asd.frequencies[i] = int(1<<(logCounts[i]-1) + reader.TryReadBits(uint64(bitcount))<<(logCounts[i]-1-int(bitcount)))
				}
			}
			totalCount += asd.frequencies[i]
		}
		asd.frequencies[omitPos] = 1<<12 - totalCount
	}
	asd.generateAliasMapping(uniqPos)
	return asd, nil
}

func (asd *ANSSymbolDistribution) generateAliasMapping(uniqPos int) {
	asd.logBucketSize = 12 - asd.logAlphabetSize
	bucketSize := 1 << asd.logBucketSize
	tableSize := 1 << asd.logAlphabetSize
	overfull := util.NewDeque[int]()
	underfull := util.NewDeque[int]()

	asd.symbols = make([]int, tableSize)
	asd.cutoffs = make([]int, tableSize)
	asd.offsets = make([]int, tableSize)
	if uniqPos >= 0 {
		for i := 0; i < tableSize; i++ {
			asd.symbols[i] = uniqPos
			asd.offsets[i] = i * bucketSize
			asd.cutoffs[i] = 0
		}
		return
	}

	for i := 0; i < asd.alphabetSize; i++ {
		asd.cutoffs[i] = asd.frequencies[i]
		if asd.cutoffs[i] > bucketSize {
			overfull.AddFirst(i)
		} else if asd.cutoffs[i] < bucketSize {
			overfull.AddFirst(i)
		}
	}
	for i := asd.alphabetSize; i < tableSize; i++ {
		overfull.AddFirst(i)
	}
	for !overfull.IsEmpty() {
		u := underfull.RemoveFirst()
		o := underfull.RemoveFirst()
		by := bucketSize - asd.cutoffs[*u]
		asd.cutoffs[*o] -= by
		asd.symbols[*u] = *o
		asd.offsets[*u] = asd.cutoffs[*o]
		if asd.cutoffs[*o] < bucketSize {
			overfull.AddFirst(*o)
		} else if asd.cutoffs[*o] > bucketSize {
			overfull.AddFirst(*o)
		}
	}
	for i := 0; i < tableSize; i++ {
		if asd.cutoffs[i] == bucketSize {
			asd.symbols[i] = i
			asd.offsets[i] = 0
			asd.cutoffs[i] = 0
		} else {
			asd.offsets[i] -= asd.cutoffs[i]
		}
	}
}

func (asd *ANSSymbolDistribution) ReadSymbol(reader *jxlio.Bitreader, stateObj *ANSState) (int, error) {
	var state int
	var err error
	if stateObj.HasState() {
		state, err = stateObj.GetState()
		if err != nil {
			return 0, err
		}
	} else {
		state = int(reader.TryReadBits(32))
	}

	index := state & 0xFFF
	i := uint32(index) >> asd.logBucketSize
	pos := index & (1<<asd.logBucketSize - 1)
	greater := pos >= asd.cutoffs[i]
	var symbol int
	var offset int
	if greater {
		symbol = asd.symbols[i]
		offset = asd.offsets[i] + pos
	} else {
		symbol = int(i)
		offset = pos
	}

	state = asd.frequencies[symbol]*int(uint32(state)>>12) + offset
	if state&0xFFFF0000 == 0 {
		state = state<<16 | int(reader.TryReadBits(16))
	}
	stateObj.SetState(state)
	return symbol, nil
}
