package entropy

import (
	"errors"
	"fmt"

	"github.com/kpfaulkner/jxl-go/jxlio"
)

type EntropyStream struct {
	usesLZ77       bool
	lz77MinSymbol  int32
	lz77MinLength  int32
	lzLengthConfig *HybridIntegerConfig
	window         []int32
	clusterMap     []int
	state          *ANSState

	dists           []SymbolDistribution
	logAlphabetSize int
	numToCopy77     int
	copyPos77       int
	numDecoded77    int
}

func NewEntropyStreamWithReaderAndNumDists(reader *jxlio.Bitreader, numDists int) (*EntropyStream, error) {
	return NewEntropyStreamWithReader(reader, numDists, false)
}

func NewEntropyStreamWithStream(stream *EntropyStream) *EntropyStream {
	es := &EntropyStream{}
	es.usesLZ77 = stream.usesLZ77
	es.lz77MinLength = stream.lz77MinLength
	es.lz77MinSymbol = stream.lz77MinSymbol
	es.lzLengthConfig = stream.lzLengthConfig
	es.clusterMap = stream.clusterMap
	es.dists = stream.dists
	es.logAlphabetSize = stream.logAlphabetSize
	if es.usesLZ77 {
		es.window = make([]int32, 1<<20)
	}
	es.state = NewANSState()
	return es
}

func NewEntropyStreamWithReader(reader *jxlio.Bitreader, numDists int, disallowLZ77 bool) (*EntropyStream, error) {

	var err error
	if numDists <= 0 {
		return nil, errors.New("Num Dists must be positive")
	}

	x := reader.MustShowBits(32)
	fmt.Printf("XXXXX BEGIN %d\n", int32(x))

	es := &EntropyStream{}
	es.state = NewANSState()
	es.usesLZ77 = reader.MustReadBool()
	if es.usesLZ77 {
		if disallowLZ77 {
			return nil, errors.New("Nested distributions cannot use LZ77")
		}
		es.lz77MinSymbol = int32(reader.MustReadU32(224, 0, 512, 0, 4096, 0, 8, 15))
		es.lz77MinLength = int32(reader.MustReadU32(3, 0, 4, 0, 5, 2, 9, 8))
		numDists++
		es.lzLengthConfig, err = NewHybridIntegerConfigWithReader(reader, 8)
		if err != nil {
			return nil, err
		}
		es.window = make([]int32, 1<<20)
	}

	es.clusterMap = make([]int, numDists)
	numClusters, err := ReadClusterMap(reader, es.clusterMap, numDists)
	if err != nil {
		return nil, err
	}

	es.dists = make([]SymbolDistribution, numClusters)
	prefixCodes := reader.MustReadBool()
	if prefixCodes {
		es.logAlphabetSize = 15
	} else {
		es.logAlphabetSize = 5 + int(reader.MustReadBits(2))
	}

	configs := make([]*HybridIntegerConfig, len(es.dists))
	for i := 0; i < len(configs); i++ {
		configs[i], err = NewHybridIntegerConfigWithReader(reader, es.logAlphabetSize)
		if err != nil {
			return nil, err
		}
	}

	if prefixCodes {
		alphabetSizes := make([]int, len(es.dists))
		for i := 0; i < len(es.dists); i++ {
			if reader.MustReadBool() {
				n := reader.MustReadBits(4)
				alphabetSizes[i] = 1 + int(1<<n+reader.MustReadBits(uint32(n)))
			} else {
				alphabetSizes[i] = 1
			}
		}
		for i := 0; i < len(es.dists); i++ {
			es.dists[i] = NewPrefixSymbolDistributionWithReader(reader, alphabetSizes[i])
		}
	} else {
		if es.logAlphabetSize == 35 {
			fmt.Printf("snoop\n")
		}
		x := reader.MustShowBits(32)
		fmt.Printf("XXXXX PRIOR %d\n", int32(x))
		for i := 0; i < len(es.dists); i++ {
			d, err := NewANSSymbolDistribution(reader, es.logAlphabetSize)
			if err != nil {
				return nil, err
			}
			es.dists[i] = d
		}
		x = reader.MustShowBits(32)
		//fmt.Printf("XXXXX %d\n", int32(x))
		if x == 172542 {
			fmt.Printf("snoop\n")
		}
	}

	for i := 0; i < len(es.dists); i++ {
		es.dists[i].SetConfig(configs[i])
	}

	return es, nil

}

func ReadClusterMap(reader *jxlio.Bitreader, clusterMap []int, maxClusters int) (int, error) {
	numDists := len(clusterMap)
	if numDists == 1 {
		clusterMap[0] = 0
	} else if reader.MustReadBool() {
		nbits := reader.MustReadBits(2)
		for i := 0; i < numDists; i++ {
			clusterMap[i] = int(reader.MustReadBits(uint32(nbits)))
		}
	} else {
		useMtf := reader.MustReadBool()
		nested, err := NewEntropyStreamWithReader(reader, 1, numDists <= 2)
		if err != nil {
			return 0, err
		}

		for i := 0; i < numDists; i++ {
			c, err := nested.ReadSymbol(reader, 0)
			clusterMap[i] = int(c)
			if err != nil {
				return 0, err
			}
		}

		if !nested.ValidateFinalState() {
			return 0, errors.New("nested distribution")
		}

		if useMtf {
			mtf := make([]int, 256)
			for i := 0; i < 256; i++ {
				mtf[i] = i
			}
			for i := 0; i < numDists; i++ {
				index := clusterMap[i]
				clusterMap[i] = mtf[index]
				if index != 0 {
					value := mtf[index]
					for j := index; j > 0; j-- {
						mtf[j] = mtf[j-1]
					}
					mtf[0] = value
				}
			}
		}
	}
	numClusters := 0
	for i := 0; i < numDists; i++ {
		if clusterMap[i] >= numClusters {
			numClusters = clusterMap[i] + 1
		}
	}
	if numClusters > maxClusters {
		return 0, errors.New("Too many clusters")
	}
	return numClusters, nil
}

func (es *EntropyStream) ReadSymbol(reader *jxlio.Bitreader, context int) (int32, error) {
	return es.ReadSymbolWithMultiplier(reader, context, 0)
}

func (es *EntropyStream) TryReadSymbol(reader *jxlio.Bitreader, context int) int32 {
	v, err := es.ReadSymbol(reader, context)
	if err != nil {
		panic(err)
	}
	return v
}

func (es *EntropyStream) ReadSymbolWithMultiplier(reader *jxlio.Bitreader, context int, distanceMultiplier int) (int32, error) {
	fmt.Printf("ReadSymbolWithMultiplier current pos %d\n", reader.BitsRead())
	if es.numToCopy77 > 0 {
		es.copyPos77++
		hybridInt := es.window[es.copyPos77&0xFFFFF]
		es.numToCopy77--
		es.numDecoded77++
		es.window[es.numDecoded77&0xFFFFF] = hybridInt
		return hybridInt, nil
	}
	if context >= len(es.clusterMap) {
		return 0, errors.New("Context cannot be bigger than bundle length")
	}
	if es.clusterMap[context] >= len(es.dists) {
		return 0, errors.New("Cluster Map points to nonexisted distribution")
	}

	dist := es.dists[es.clusterMap[context]]
	t, err := dist.ReadSymbol(reader, es.state)
	token := int32(t)
	if err != nil {
		return 0, err
	}

	if es.usesLZ77 && token >= es.lz77MinSymbol {
		panic("not implemented")
		//lz77dist := es.dists[es.clusterMap[len(es.clusterMap)-1]]
		//es.numToCopy77 = es.lz77MinLength + es.mustReadHybridInteger(reader, es.lzLengthConfig, token-es.lz77MinSymbol)
		//token, err = lz77dist.ReadSymbol(reader, es.state)
		//if err != nil {
		//	return 0, err
		//}
		//distance, err := es.readHybridInteger(reader, lz77dist.GetConfig(), token)
		//if err != nil {
		//	return 0, err
		//}
		//
		//if distanceMultiplier == 0 {
		//	distance++
		//} else if distance < 120 {
		//	distance = SPECIAL_DISTANCES[distance][0] + distanceMultiplier*SPECIAL_DISTANCES[distance][1]
		//} else {
		//	distance -= 119
		//}
		//if distance > 1<<20 {
		//	distance = 1 << 20
		//}
		//if distance > es.numDecoded77 {
		//	distance = es.numDecoded77
		//}
		//es.copyPos77 = es.numDecoded77 - distance
		//return es.ReadSymbol(reader, context)
	}
	hybridInt, err := es.readHybridInteger(reader, dist.GetConfig(), token)
	if err != nil {
		return 0, err
	}
	if es.usesLZ77 {
		es.numDecoded77++
		es.window[es.numDecoded77&0xFFFFF] = hybridInt
	}
	return hybridInt, nil
}

func (es *EntropyStream) readHybridInteger(reader *jxlio.Bitreader, config *HybridIntegerConfig, token int32) (int32, error) {
	split := 1 << config.SplitExponent
	if token < int32(split) {
		return token, nil
	}
	n := config.SplitExponent - config.LsbInToken - config.MsbInToken + int(uint32(token-int32(split))>>(config.MsbInToken+config.LsbInToken))
	if n > 32 {
		return 0, errors.New("n is too large")
	}
	low := token & ((1 << config.LsbInToken) - 1)
	token = int32(uint32(token) >> config.LsbInToken)
	token &= (1 << config.MsbInToken) - 1
	token |= 1 << config.MsbInToken
	return ((int32(token<<n)|int32(reader.MustReadBits(uint32(n))))<<int32(config.LsbInToken) | int32(low)), nil
}

func (es *EntropyStream) ValidateFinalState() bool {
	if !es.state.HasState() {
		return true
	}
	s, err := es.state.GetState()
	if err != nil || s != 0x130000 {
		return false
	}

	return true
}
