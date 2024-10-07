package entropy

type ANSState struct {
	State int32
}

func NewANSState() *ANSState {
	s := &ANSState{}
	return s
}
