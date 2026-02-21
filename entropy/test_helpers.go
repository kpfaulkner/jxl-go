package entropy

import "github.com/kpfaulkner/jxl-go/jxlio"

// FakeSymbolDistribution implements SymbolDistribution for testing.
// It returns predefined symbols in sequence.
type FakeSymbolDistribution struct {
	Symbols             []int32
	Cfg                 *HybridIntegerConfig
	idx                 int
	ActivateStateOnRead bool
}

func (f *FakeSymbolDistribution) ReadSymbol(reader jxlio.BitReader, state *ANSState) (int32, error) {
	if f.ActivateStateOnRead {
		state.HasState = true
		state.State = 0
	}
	if f.idx >= len(f.Symbols) {
		return 0, nil
	}
	sym := f.Symbols[f.idx]
	f.idx++
	return sym, nil
}

func (f *FakeSymbolDistribution) SetConfig(config *HybridIntegerConfig) {
	f.Cfg = config
}

func (f *FakeSymbolDistribution) GetConfig() *HybridIntegerConfig {
	return f.Cfg
}

// NewEntropyStreamForTest creates a controllable EntropyStream for testing.
// All contexts map to the single provided distribution.
func NewEntropyStreamForTest(numContexts int, dist SymbolDistribution) *EntropyStream {
	es := &EntropyStream{}
	es.clusterMap = make([]int, numContexts)
	es.dists = []SymbolDistribution{dist}
	es.ansState = &ANSState{State: -1, HasState: false}
	return es
}
