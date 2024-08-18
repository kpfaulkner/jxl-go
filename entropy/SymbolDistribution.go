package entropy

import "github.com/kpfaulkner/jxl-go/jxlio"

type SymbolDistribution interface {
	ReadSymbol(reader *jxlio.Bitreader, state *ANSState) (int, error)
	SetConfig(config *HybridIntegerConfig)
	GetConfig() *HybridIntegerConfig
}

type SymbolDistributionBase struct {
	config          *HybridIntegerConfig
	logBucketSize   int
	alphabetSize    int
	logAlphabetSize int
}

func NewSymbolDistributionBase() *SymbolDistributionBase {
	rcvr := &SymbolDistributionBase{}
	return rcvr
}

func (rcvr *SymbolDistributionBase) ReadSymbol(reader *jxlio.Bitreader, state *ANSState) (int, error) {

	return 0, nil
}

func (rcvr *SymbolDistributionBase) SetConfig(config *HybridIntegerConfig) {
	rcvr.config = config
}

func (rcvr *SymbolDistributionBase) GetConfig() *HybridIntegerConfig {
	return rcvr.config
}
