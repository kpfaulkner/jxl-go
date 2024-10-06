package frame

import "github.com/kpfaulkner/jxl-go/jxlio"

type Quantizer struct {
	globalScale   uint32
	quantLF       uint32
	scaledDequant []float32
}

func NewQuantizerWithReader(reader *jxlio.Bitreader, lfDequant []float32) (*Quantizer, error) {
	q := &Quantizer{}
	var err error
	q.globalScale, err = reader.ReadU32(1, 11, 2049, 11, 4097, 12, 8193, 16)
	if err != nil {
		return nil, err
	}
	q.quantLF, err = reader.ReadU32(16, 0, 1, 5, 1, 8, 1, 16)
	if err != nil {
		return nil, err
	}

	for i := 0; i < 3; i++ {
		q.scaledDequant[i] = (1 << 16) * lfDequant[i] / float32(q.globalScale*q.quantLF)
	}
	return q, nil
}
