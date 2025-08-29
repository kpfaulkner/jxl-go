package testcommon

import (
	"bytes"
	"os"
	"testing"

	"github.com/kpfaulkner/jxl-go/jxlio"
)

func GenerateTestBitReader(t *testing.T, filepath string) jxlio.BitReader {
	data, err := os.ReadFile(filepath)
	if err != nil {
		t.Errorf("error reading test jxl file : %v", err)
		return nil
	}
	br := jxlio.NewBitStreamReader(bytes.NewReader(data))

	return br
}
