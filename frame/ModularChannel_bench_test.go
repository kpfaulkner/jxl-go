package frame

import (
	"github.com/kpfaulkner/jxl-go/util"
	"testing"
)

func BenchmarkModularChannelDecode(b *testing.B) {
	// Setup a representative ModularChannel
	width, height := 256, 256
	mc := NewModularChannelWithAllParams(int32(height), int32(width), 0, 0, true)

	// We need a tree and other dependencies, but for a simple allocation benchmark
	// we can just call allocate() which uses MakeMatrix2D
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mc.buffer = nil
		mc.allocate()
	}
}

func BenchmarkMakeMatrix2D(b *testing.B) {
	width, height := 256, 256
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = util.MakeMatrix2D[int32](height, width)
	}
}

func BenchmarkMakeMatrix2DPooled(b *testing.B) {
	width, height := 256, 256
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m := util.MakeMatrix2DPooled[int32](height, width)
		util.ReturnMatrix2DToPool(m)
	}
}
