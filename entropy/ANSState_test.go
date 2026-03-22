package entropy

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestANSState(t *testing.T) {
	state := &ANSState{State: 123, HasState: true}
	assert.Equal(t, int32(123), state.State)
	assert.True(t, state.HasState)

	state.SetState(456)
	assert.Equal(t, int32(456), state.State)
	assert.True(t, state.HasState)
}
