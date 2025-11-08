package frame

import (
	"fmt"
	"testing"

	"github.com/kpfaulkner/jxl-go/testcommon"
	"github.com/kpfaulkner/jxl-go/util"
	"github.com/stretchr/testify/assert"
)

func TestNewModularChannelWithAllParams(t *testing.T) {
	mc := NewModularChannelWithAllParams(5, 5, 1, 1, true)
	assert.NotNil(t, mc)
}

func TestAllocate(t *testing.T) {
	mc := NewModularChannelWithAllParams(5, 5, 1, 1, true)
	assert.NotNil(t, mc)

	mc.allocate()
	assert.NotNil(t, mc.buffer)
}

func TestPrediction(t *testing.T) {

	for _, tc := range []struct {
		name           string
		mc             *ModularChannel
		x              int32
		y              int32
		k              int32
		expectedResult int32
		expectErr      bool
	}{
		{
			name:           "success k=0",
			mc:             NewModularChannelWithAllParams(5, 5, 1, 1, true),
			x:              0,
			y:              0,
			k:              0,
			expectedResult: 0,
			expectErr:      false,
		},
		{
			name:           "success k=1",
			mc:             makeFakeModularChannel(),
			x:              1,
			y:              1,
			k:              1,
			expectedResult: 1,
			expectErr:      false,
		},
		{
			name:           "success k=2",
			mc:             makeFakeModularChannel(),
			x:              1,
			y:              1,
			k:              2,
			expectedResult: 5,
			expectErr:      false,
		},
		{
			name:           "success k=3",
			mc:             makeFakeModularChannel(),
			x:              1,
			y:              1,
			k:              3,
			expectedResult: 3,
			expectErr:      false,
		},
		{
			name:           "success k=4",
			mc:             makeFakeModularChannel(),
			x:              1,
			y:              1,
			k:              4,
			expectedResult: 5,
			expectErr:      false,
		},
		{
			name:           "success k=5",
			mc:             makeFakeModularChannel(),
			x:              1,
			y:              1,
			k:              5,
			expectedResult: 5,
			expectErr:      false,
		},
		{
			name:           "success k=6",
			mc:             makeFakeModularChannel(),
			x:              1,
			y:              1,
			k:              6,
			expectedResult: 4,
			expectErr:      false,
		},
		{
			name:           "success k=7",
			mc:             makeFakeModularChannel(),
			x:              1,
			y:              1,
			k:              7,
			expectedResult: 10,
			expectErr:      false,
		},
		{
			name:           "success k=8",
			mc:             makeFakeModularChannel(),
			x:              1,
			y:              1,
			k:              8,
			expectedResult: 0,
			expectErr:      false,
		},
		{
			name:           "success k=9",
			mc:             makeFakeModularChannel(),
			x:              1,
			y:              1,
			k:              9,
			expectedResult: 1,
			expectErr:      false,
		},
		{
			name:           "success k=10",
			mc:             makeFakeModularChannel(),
			x:              1,
			y:              1,
			k:              10,
			expectedResult: 0,
			expectErr:      false,
		},
		{
			name:           "success k=11",
			mc:             makeFakeModularChannel(),
			x:              1,
			y:              1,
			k:              11,
			expectedResult: 2,
			expectErr:      false,
		},
		{
			name:           "success k=12",
			mc:             makeFakeModularChannel(),
			x:              1,
			y:              1,
			k:              12,
			expectedResult: 7,
			expectErr:      false,
		},
		{
			name:           "success k=13",
			mc:             makeFakeModularChannel(),
			x:              1,
			y:              1,
			k:              13,
			expectedResult: 5,
			expectErr:      false,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {

			res, err := tc.mc.prediction(tc.x, tc.y, tc.k)
			if err != nil && !tc.expectErr {
				t.Errorf("got error when none was expected : %v", err)
			}
			if err == nil && tc.expectErr {
				t.Errorf("expected error but got none")
			}
			if err != nil && tc.expectErr {
				return
			}

			assert.Equal(t, tc.expectedResult, res, tc.name)
		})
	}
}

func TestGetWalkerFunc(t *testing.T) {

	for _, tc := range []struct {
		name           string
		mc             *ModularChannel
		parent         *ModularStream
		x              int32
		y              int32
		k              int32
		channelIndex   int32
		streamIndex    int32
		expectedResult int32
		expectErr      bool
	}{
		{
			name:           "success k=0",
			mc:             makeFakeModularChannel(),
			x:              1,
			y:              1,
			channelIndex:   1,
			streamIndex:    1,
			k:              0,
			expectedResult: 1,
			expectErr:      false,
		},
		{
			name:           "success k=1",
			mc:             makeFakeModularChannel(),
			x:              1,
			y:              1,
			channelIndex:   1,
			streamIndex:    1,
			k:              1,
			expectedResult: 1,
			expectErr:      false,
		},
		{
			name:           "success k=2",
			mc:             makeFakeModularChannel(),
			x:              1,
			y:              1,
			channelIndex:   1,
			streamIndex:    1,
			k:              2,
			expectedResult: 1,
			expectErr:      false,
		},
		{
			name:           "success k=3",
			mc:             makeFakeModularChannel(),
			x:              1,
			y:              1,
			channelIndex:   1,
			streamIndex:    1,
			k:              3,
			expectedResult: 1,
			expectErr:      false,
		},
		{
			name:           "success k=4",
			mc:             makeFakeModularChannel(),
			x:              1,
			y:              1,
			channelIndex:   1,
			streamIndex:    1,
			k:              4,
			expectedResult: 5,
			expectErr:      false,
		},
		{
			name:           "success k=5",
			mc:             makeFakeModularChannel(),
			x:              1,
			y:              1,
			channelIndex:   1,
			streamIndex:    1,
			k:              5,
			expectedResult: 1,
			expectErr:      false,
		},
		{
			name:           "success k=6",
			mc:             makeFakeModularChannel(),
			x:              1,
			y:              1,
			channelIndex:   1,
			streamIndex:    1,
			k:              6,
			expectedResult: 5,
			expectErr:      false,
		},
		{
			name:           "success k=7",
			mc:             makeFakeModularChannel(),
			x:              1,
			y:              1,
			channelIndex:   1,
			streamIndex:    1,
			k:              7,
			expectedResult: 1,
			expectErr:      false,
		},
		{
			name:           "success k=8",
			mc:             makeFakeModularChannel(),
			x:              1,
			y:              1,
			channelIndex:   1,
			streamIndex:    1,
			k:              8,
			expectedResult: 1,
			expectErr:      false,
		},
		{
			name:           "success k=9",
			mc:             makeFakeModularChannel(),
			x:              1,
			y:              1,
			channelIndex:   1,
			streamIndex:    1,
			k:              9,
			expectedResult: 6,
			expectErr:      false,
		},
		{
			name:           "success k=10",
			mc:             makeFakeModularChannel(),
			x:              1,
			y:              1,
			channelIndex:   1,
			streamIndex:    1,
			k:              10,
			expectedResult: 1,
			expectErr:      false,
		},
		{
			name:           "success k=11",
			mc:             makeFakeModularChannel(),
			x:              1,
			y:              1,
			channelIndex:   1,
			streamIndex:    1,
			k:              11,
			expectedResult: -5,
			expectErr:      false,
		},
		{
			name:           "success k=12",
			mc:             makeFakeModularChannel(),
			x:              1,
			y:              1,
			channelIndex:   1,
			streamIndex:    1,
			k:              12,
			expectedResult: -5,
			expectErr:      false,
		},
		{
			name:           "success k=13",
			mc:             makeFakeModularChannel(),
			x:              1,
			y:              1,
			channelIndex:   1,
			streamIndex:    1,
			k:              13,
			expectedResult: 0,
			expectErr:      false,
		},
		{
			name:           "success k=14",
			mc:             makeFakeModularChannel(),
			x:              1,
			y:              1,
			channelIndex:   1,
			streamIndex:    1,
			k:              14,
			expectedResult: 0,
			expectErr:      false,
		},
		{
			name:           "success k=15",
			mc:             makeFakeModularChannel(),
			x:              1,
			y:              1,
			channelIndex:   1,
			streamIndex:    1,
			k:              15,
			expectedResult: 1,
			expectErr:      false,
		},
		{
			name:           "success k=16 (triggers default)",
			mc:             makeFakeModularChannel(),
			parent:         makeFakeModularStream(),
			x:              1,
			y:              1,
			channelIndex:   1,
			streamIndex:    1,
			k:              16,
			expectedResult: 0,
			expectErr:      false,
		},
		{
			name:           "success k=17 (triggers default)",
			mc:             makeFakeModularChannel(),
			parent:         makeFakeModularStream(),
			x:              1,
			y:              1,
			channelIndex:   1,
			streamIndex:    1,
			k:              17,
			expectedResult: 0,
			expectErr:      false,
		},
		{
			name:           "success k=18 (triggers default)",
			mc:             makeFakeModularChannel(),
			parent:         makeFakeModularStream(),
			x:              1,
			y:              1,
			channelIndex:   1,
			streamIndex:    1,
			k:              18,
			expectedResult: 0,
			expectErr:      false,
		},
		{
			name:           "success k=19 (triggers default)",
			mc:             makeFakeModularChannel(),
			parent:         makeFakeModularStream(),
			x:              1,
			y:              1,
			channelIndex:   1,
			streamIndex:    1,
			k:              19,
			expectedResult: 0,
			expectErr:      false,
		},
		{
			name:           "success k=20 (triggers default)",
			mc:             makeFakeModularChannel(),
			parent:         makeFakeModularStream(),
			x:              1,
			y:              1,
			channelIndex:   1,
			streamIndex:    1,
			k:              20,
			expectedResult: 0,
			expectErr:      false,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {

			walkerFunc := tc.mc.getWalkerFunc(tc.channelIndex, tc.streamIndex, tc.x, tc.y, 1, tc.parent)
			res, err := walkerFunc(tc.k)
			if err != nil && !tc.expectErr {
				t.Errorf("got error when none was expected : %v", err)
			}
			if err == nil && tc.expectErr {
				t.Errorf("expected error but got none")
			}
			if err != nil && tc.expectErr {
				return
			}

			assert.Equal(t, tc.expectedResult, res, tc.name)
		})
	}
}

func TestDecode(t *testing.T) {
	mc := makeFakeModularChannel()

	bitReader := testcommon.NewFakeBitReader()
	entropyStreamer := NewFakeEntropyStreamer()
	parent := makeFakeModularStream()

	tree := NewFakeTree()
	wpParams := &WPParams{}
	err := mc.decode(bitReader, entropyStreamer, wpParams, tree, parent, 1, 1, 1)
	if err != nil {
		t.Errorf("got error when none was expected : %v", err)
	}
}

func TestPrePredictWP(t *testing.T) {
	mc := makeFakeModularChannel()

	wpParams := &WPParams{}
	res, err := mc.prePredictWP(wpParams, 0, 0)
	if err != nil {
		t.Errorf("got error when none was expected : %v", err)
	}
	fmt.Printf("XXX res %d\n", res)
}

func makeFakeModularStream() *ModularStream {

	ms := &ModularStream{
		channels: []*ModularChannel{{
			size:   util.Dimension{5, 5},
			hshift: 1,
			vshift: 1,
			buffer: util.MakeMatrix2D[int32](5, 5),
		}},
	}

	return ms
}

func makeFakeModularChannel() *ModularChannel {
	mc := NewModularChannelWithAllParams(5, 5, 1, 1, true)
	mc.allocate()

	i := int32(0)
	// populate with pattern.
	for y := 0; y < 5; y++ {
		for x := 0; x < 5; x++ {
			mc.buffer[x][y] = i
			i++
		}
	}

	// not the right type of data... but just want to confirm consistency with unit tests
	mc.pred = util.MakeMatrix2D[int32](5, 5)
	for y := 0; y < 5; y++ {
		for x := 0; x < 5; x++ {
			mc.pred[x][y] = i
			i++
		}
	}

	mc.subpred = []int32{0, 0, 0, 0}
	mc.weight = []int32{0, 0, 0, 0}
	return mc
}

// Makes a real tree with fake data.
// Due to the way the tree is structured (each node being of same type), making an interface
// and mocking this is rather awkward. So will just make a REAL tree with controlled data.
func NewFakeTree() *MATreeNode {

	var tree = &MATreeNode{
		leftChildNode:  &MATreeNode{property: -1},
		rightChildNode: &MATreeNode{property: -1},
	}
	return tree
}
