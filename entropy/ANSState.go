package entropy

import "errors"

type ANSState struct {
	state    int32
	hasState bool
}

func NewANSState() (rcvr *ANSState) {
	rcvr = &ANSState{}
	rcvr.hasState = false
	return
}
func (rcvr *ANSState) GetState() (int32, error) {
	if rcvr.hasState {
		return rcvr.state, nil
	}

	return 0, errors.New("ANS state has not been initialized")
}

func (rcvr *ANSState) HasState() bool {
	return rcvr.hasState
}
func (rcvr *ANSState) Reset() {
	rcvr.hasState = false
}
func (rcvr *ANSState) SetState(state int32) {

	rcvr.state = state
	rcvr.hasState = true
}
