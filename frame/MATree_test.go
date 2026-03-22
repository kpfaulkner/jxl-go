package frame

import (
	"testing"

	"github.com/kpfaulkner/jxl-go/entropy"
	"github.com/kpfaulkner/jxl-go/jxlio"
	"github.com/kpfaulkner/jxl-go/testcommon"
	"github.com/stretchr/testify/assert"
)

func TestNewMATreeWithReader(t *testing.T) {

	for _, tc := range []struct {
		name                    string
		boolData                []bool
		enumData                []int32
		u32Data                 []uint32
		f16Data                 []float32
		bitsData                []uint64
		parent                  Framer
		entropyStream           entropy.EntropyStreamWithReaderAndNumDistsFunc
		entropyStreamWithReader entropy.EntropyStreamWithReaderFunc
		displayTree             bool
		expectedResult          LFCoefficients
		entropySymbols          []int32
		expectErr               bool
	}{
		//{
		//	name:                    "success, property < 0",
		//	parent:                  NewFakeFramer(),
		//	boolData:                []bool{},
		//	enumData:                nil,
		//	u32Data:                 []uint32{},
		//	bitsData:                []uint64{},
		//	entropyStream:           NewFakeEntropyStreamerFunc,
		//	entropyStreamWithReader: NewFakeEntropyWithReaderFunc,
		//	entropySymbols:          []int32{0},
		//	expectErr:               false,
		//},
		{
			name:     "success, simple tree",
			parent:   NewFakeFramer(VARDCT),
			boolData: []bool{},
			entropyStream: func(reader jxlio.BitReader, numDists int, readClusterMapFunc entropy.ReadClusterMapFunc) (entropy.EntropyStreamer, error) {
				fakeStreamer := &FakeEntropyStreamer{}
				// Root: property=0 (ReadSymbol returns 1), value=5
				// Left: property=-1 (ReadSymbol returns 0), predictor=0, offset=0, mulLog=0, mulBits=0
				// Right: property=-1 (ReadSymbol returns 0), predictor=1, offset=1, mulLog=1, mulBits=1
				fakeStreamer.FakeSymbols = []int32{
					1, // property+1 = 1 => property=0
					0, // property+1 = 0 => property=-1 (Leaf)
					0, // predictor=0
					0, // mulLog=0
					0, // mulBits=0
					0, // property+1 = 0 => property=-1 (Leaf)
					1, // predictor=1
					1, // mulLog=1
					1, // mulBits=1
				}
				fakeStreamer.FakeTrySymbols = []int32{
					5, // TryReadSymbol(reader, 0) => value=5 (unpacked signed)
					0, // TryReadSymbol(reader, 3) => offset=0 (unpacked signed)
					2, // TryReadSymbol(reader, 3) => offset=1 (unpacked signed 2 => 1)
				}
				return fakeStreamer, nil
			},
			entropyStreamWithReader: NewFakeEntropyWithReaderFunc,
			displayTree:             true,
			expectErr:               false,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {

			bitReader := &testcommon.FakeBitReader{
				ReadU32Data:  tc.u32Data,
				ReadBitsData: tc.bitsData,
				ReadBoolData: tc.boolData,
				ReadF16Data:  tc.f16Data,
			}
			tree, err := NewMATreeWithReader(bitReader, tc.entropyStream, tc.entropyStreamWithReader)
			if err != nil && !tc.expectErr {
				t.Errorf("got error when none was expected : %v", err)
			}
			if err == nil && tc.expectErr {
				t.Errorf("expected error but got none")
			}
			if err != nil && tc.expectErr {
				return
			}

			if tree == nil {
				t.Errorf("nil MATreeNode")
			}

			if tc.displayTree {
				DisplayTree(tree, 0)
			}
		})
	}
}

func TestMATreeMethods(t *testing.T) {
	// Simple tree:
	// Root: property=0, value=5
	//   Left (Leaf): predictor=0, context=0
	//   Right (Leaf): predictor=1, context=1
	left := &MATreeNode{property: -1, predictor: 0, context: 0}
	right := &MATreeNode{property: -1, predictor: 1, context: 1}
	root := &MATreeNode{
		property:       0,
		value:          5,
		leftChildNode:  left,
		rightChildNode: right,
	}
	left.parent = root
	right.parent = root

	// Test getSize
	assert.Equal(t, 3, root.getSize())

	// Test compactify (property 0 is channelIndex)
	// channelIndex 10 > 5 => Left
	assert.Equal(t, left, root.compactify(10, 0))
	// channelIndex 2 <= 5 => Right
	assert.Equal(t, right, root.compactify(2, 0))

	// Test compactifyWithY (property 0 is channelIndex, property 1 is streamIndex, property 2 is y)
	// channelIndex 10 > 5 => Left
	assert.Equal(t, left, root.compactifyWithY(10, 0, 0))

	root2 := &MATreeNode{
		property:       2, // y
		value:          10,
		leftChildNode:  left,
		rightChildNode: right,
	}
	// y 15 > 10 => Left
	assert.Equal(t, left, root2.compactifyWithY(0, 0, 15))

	// Test useWeightedPredictor
	left.predictor = 6 // weighted predictor
	assert.True(t, root.useWeightedPredictor())
	left.predictor = 0
	assert.False(t, root.useWeightedPredictor())

	// Test walk
	res, err := root.walk(func(inp int32) (int32, error) {
		if inp == 0 {
			return 10, nil // > 5 => left
		}
		return 0, nil
	})
	assert.NoError(t, err)
	assert.Equal(t, left, res)
}

func TestMATreeDisplay(t *testing.T) {

	tree := &MATreeNode{}
	tree.leftChildNode = &MATreeNode{
		parent:   tree,
		property: -1,
	}
	tree.rightChildNode = &MATreeNode{
		parent:   tree,
		property: -1,
	}

	DisplayTree(tree, 0)
}
