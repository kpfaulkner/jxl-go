package core

import "github.com/kpfaulkner/jxl-go/jxlio"

type WPParams struct {
	param1  int
	param2  int
	param3a int32
	param3b int32
	param3c int32
	param3d int32
	param3e int32
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
		wp.param3a = int32(reader.MustReadBits(5))
		wp.param3b = int32(reader.MustReadBits(5))
		wp.param3c = int32(reader.MustReadBits(5))
		wp.param3d = int32(reader.MustReadBits(5))
		wp.param3e = int32(reader.MustReadBits(5))
		wp.weight[0] = int64(reader.MustReadBits(4))
		wp.weight[1] = int64(reader.MustReadBits(4))
		wp.weight[2] = int64(reader.MustReadBits(4))
		wp.weight[3] = int64(reader.MustReadBits(4))
	}

	return &wp
}
