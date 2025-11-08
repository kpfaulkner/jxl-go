package frame

import (
	"testing"

	"github.com/kpfaulkner/jxl-go/entropy"
	"github.com/kpfaulkner/jxl-go/jxlio"
	"github.com/kpfaulkner/jxl-go/testcommon"
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
			name:     "success, property == 0",
			parent:   NewFakeFramer(),
			boolData: []bool{},
			enumData: nil,
			u32Data:  []uint32{},
			bitsData: []uint64{},
			entropyStream: func(reader jxlio.BitReader, numDists int, readClusterMapFunc entropy.ReadClusterMapFunc) (entropy.EntropyStreamer, error) {
				fakeStreamer := &FakeEntropyStreamer{}

				fakeStreamer.FakeSymbols = []int32{1, 0, 1, 0, 0, 0, 0, 0, 0}
				return fakeStreamer, nil
			},
			entropyStreamWithReader: NewFakeEntropyWithReaderFunc,
			displayTree:             true,
			entropySymbols:          []int32{1},
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
