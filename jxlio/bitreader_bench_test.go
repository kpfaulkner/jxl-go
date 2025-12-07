package jxlio

import (
	"bytes"
	"testing"
)

// BenchmarkReadBitsSmall benchmarks reading small bit amounts (1-8 bits) which is common in entropy coding
func BenchmarkReadBitsSmall(b *testing.B) {
	// Create a large data set
	data := make([]byte, 1024*1024) // 1MB
	for i := range data {
		data[i] = byte(i & 0xFF)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		reader := bytes.NewReader(data)
		br := NewBitStreamReader(reader)

		// Read many small bit amounts (typical for entropy decoding)
		for j := 0; j < 10000; j++ {
			br.ReadBits(1)
			br.ReadBits(2)
			br.ReadBits(3)
			br.ReadBits(4)
			br.ReadBits(8)
		}
	}
}

// BenchmarkReadBitsMedium benchmarks reading medium bit amounts (8-16 bits)
func BenchmarkReadBitsMedium(b *testing.B) {
	data := make([]byte, 1024*1024)
	for i := range data {
		data[i] = byte(i & 0xFF)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		reader := bytes.NewReader(data)
		br := NewBitStreamReader(reader)

		for j := 0; j < 10000; j++ {
			br.ReadBits(8)
			br.ReadBits(12)
			br.ReadBits(16)
		}
	}
}

// BenchmarkReadBitsLarge benchmarks reading larger bit amounts (32 bits)
func BenchmarkReadBitsLarge(b *testing.B) {
	data := make([]byte, 1024*1024)
	for i := range data {
		data[i] = byte(i & 0xFF)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		reader := bytes.NewReader(data)
		br := NewBitStreamReader(reader)

		for j := 0; j < 5000; j++ {
			br.ReadBits(32)
		}
	}
}

// BenchmarkReadBitsMixed benchmarks a realistic mix of bit read sizes
func BenchmarkReadBitsMixed(b *testing.B) {
	data := make([]byte, 1024*1024)
	for i := range data {
		data[i] = byte(i & 0xFF)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		reader := bytes.NewReader(data)
		br := NewBitStreamReader(reader)

		// Simulate realistic entropy decoding pattern
		for j := 0; j < 5000; j++ {
			br.ReadBits(1)  // Symbol flag
			br.ReadBits(4)  // Small integer
			br.ReadBits(8)  // Byte value
			br.ReadBits(2)  // Choice
			br.ReadBits(12) // Medium integer
			br.ReadBits(1)  // Boolean
		}
	}
}

// BenchmarkReadByte benchmarks reading full bytes
func BenchmarkReadByte(b *testing.B) {
	data := make([]byte, 1024*1024)
	for i := range data {
		data[i] = byte(i & 0xFF)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		reader := bytes.NewReader(data)
		br := NewBitStreamReader(reader)

		for j := 0; j < 10000; j++ {
			br.ReadByte()
		}
	}
}

// BenchmarkReadU32 benchmarks ReadU32 which is heavily used
func BenchmarkReadU32(b *testing.B) {
	data := make([]byte, 1024*1024)
	for i := range data {
		data[i] = byte(i & 0xFF)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		reader := bytes.NewReader(data)
		br := NewBitStreamReader(reader)

		for j := 0; j < 5000; j++ {
			br.ReadU32(0, 0, 1, 0, 2, 4, 18, 6)
		}
	}
}

// BenchmarkReadU64 benchmarks ReadU64
func BenchmarkReadU64(b *testing.B) {
	data := make([]byte, 1024*1024)
	for i := range data {
		data[i] = byte(i & 0xFF)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		reader := bytes.NewReader(data)
		br := NewBitStreamReader(reader)

		for j := 0; j < 5000; j++ {
			br.ReadU64()
		}
	}
}