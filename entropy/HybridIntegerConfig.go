package entropy

import (
	"errors"

	"github.com/kpfaulkner/jxl-go/jxlio"
	"github.com/kpfaulkner/jxl-go/util"
)

type HybridIntegerConfig struct {
	SplitExponent int32
	MsbInToken    int32
	LsbInToken    int32
}

func NewHybridIntegerConfig(splitExponent int32, msbInToken int32, lsbInToken int32) *HybridIntegerConfig {
	hic := &HybridIntegerConfig{}
	hic.SplitExponent = splitExponent
	hic.MsbInToken = msbInToken
	hic.LsbInToken = lsbInToken
	return hic
}

func NewHybridIntegerConfigWithReader(reader *jxlio.Bitreader, logAlphabetSize int32) (*HybridIntegerConfig, error) {
	hic := &HybridIntegerConfig{}
	hic.SplitExponent = int32(reader.MustReadBits(uint32(util.CeilLog1p(int64(logAlphabetSize)))))
	if hic.SplitExponent == logAlphabetSize {
		hic.MsbInToken = 0
		hic.LsbInToken = 0
		return hic, nil
	}
	hic.MsbInToken = int32(reader.MustReadBits(uint32(util.CeilLog1p(int64(hic.SplitExponent)))))
	if hic.MsbInToken > hic.SplitExponent {
		return nil, errors.New("msbInToken is too large")
	}
	hic.LsbInToken = int32(reader.MustReadBits(uint32(util.CeilLog1p(int64(hic.SplitExponent - hic.MsbInToken)))))
	if hic.MsbInToken+hic.LsbInToken > hic.SplitExponent {
		return nil, errors.New("msbInToken + lsbInToken is too large")
	}
	return hic, nil
}
