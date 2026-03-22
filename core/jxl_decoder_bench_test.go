package core

import (
	"bytes"
	"os"
	"testing"

	"github.com/kpfaulkner/jxl-go/jxlio"
	"github.com/kpfaulkner/jxl-go/options"
)

func benchmarkDecode(b *testing.B, filename string) {
	data, err := os.ReadFile(filename)
	if err != nil {
		b.Fatalf("failed to read test file: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r := bytes.NewReader(data)
		br := jxlio.NewBitStreamReader(r)
		opts := options.NewJXLOptions(nil)
		decoder := NewJXLCodestreamDecoder(br, opts)
		_, err := decoder.decode()
		if err != nil {
			b.Fatalf("failed to decode: %v", err)
		}
	}
}

func BenchmarkDecodeUnittest(b *testing.B) {
	benchmarkDecode(b, "../testdata/unittest.jxl")
}

func BenchmarkDecodeTiny2(b *testing.B) {
	benchmarkDecode(b, "../testdata/tiny2.jxl")
}

func BenchmarkDecodeLossless(b *testing.B) {
	benchmarkDecode(b, "../testdata/lossless.jxl")
}

func BenchmarkDecodeGrayscale(b *testing.B) {
	benchmarkDecode(b, "../testdata/grayscale.jxl")
}
