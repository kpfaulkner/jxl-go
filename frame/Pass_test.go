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
}
