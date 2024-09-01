package core

import "github.com/kpfaulkner/jxl-go/jxlio"

type WPParams struct {
	param1  int
	param2  int
	param3a int
	param3b int
	param3c int
	param3d int
	param3e int
	weight  [4]int64
}

func NewWPParams(reader *jxlio.Bitreader) *WPParams {
	wp := WPParams{}
	if reader.MustReadBool() {
		wp.param1 = 16
		wp.param2 = 10
		wp.param3a = 7
		wp.param3b = 7
		wp.param3c = 7
		wp.param3d = 0
		wp.param3e = 0
		wp.weight[0] = 13
		wp.weight[1] = 12
		wp.weight[2] = 12
		wp.weight[3] = 12
	} else {
		wp.param1 = int(reader.MustReadBits(5))
		wp.param2 = int(reader.MustReadBits(5))
		wp.param3a = int(reader.MustReadBits(5))
		wp.param3b = int(reader.MustReadBits(5))
		wp.param3c = int(reader.MustReadBits(5))
		wp.param3d = int(reader.MustReadBits(5))
		wp.param3e = int(reader.MustReadBits(5))
		wp.weight[0] = int(reader.MustReadBits(4))
		wp.weight[1] = int(reader.MustReadBits(4))
		wp.weight[2] = int(reader.MustReadBits(4))
		wp.weight[3] = int(reader.MustReadBits(4))
	}

	return &wp
}
