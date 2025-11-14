package frame

import (
	"testing"

	"github.com/kpfaulkner/jxl-go/testcommon"
	"github.com/stretchr/testify/assert"
)

func TestNewPassWithReader(t *testing.T) {

	bitReader := testcommon.NewFakeBitReader()
	fakeFrame := NewFakeFramer(MODULAR)
	fakeFrame.getLFGlobal().globalModular = &ModularStream{channels: []*ModularChannel{{}}}
	res, err := NewPassWithReader(bitReader, fakeFrame, 0, 0)
	assert.Nil(t, err)
	assert.NotNil(t, res)

	bitReader = testcommon.NewFakeBitReader()
	bitReader.ReadU32Data = []uint32{0, 0, 0}
	fakeFrame = NewFakeFramer(VARDCT)
	fakeFrame.getLFGlobal().globalModular = &ModularStream{channels: []*ModularChannel{{}}}
	res, err = NewPassWithReader(bitReader, fakeFrame, 0, 0)
	assert.Nil(t, err)
	assert.NotNil(t, res)

	// cherry pick a few values to make sure lengths are correct.
	// TODO(kpfaulkner) write some proper tests!!!
	assert.Equal(t, uint32(3), res.minShift)
	assert.Equal(t, uint32(3), res.maxShift)
	assert.Equal(t, 13, len(res.hfPass.order))

}
