package jxlio

import "io"

const (
	tempBufSize = 10000
)

type IOHelper struct {
}

func NewIOHelper() (rcvr *IOHelper) {
	rcvr = &IOHelper{}
	return
}

func ReadFully(in io.ReadSeeker, buffer []byte, offset int, len int) (int, error) {
	remaining := len

	// randomly selected tempBuf
	tempBuf := make([]byte, tempBufSize)
	for remaining > 0 {
		//count := in.Read(buffer, offset+len-remaining, remaining)
		count, err := in.Read(tempBuf)
		if err != nil {
			return 0, err
		}

		if count <= 0 {
			break
		}

		// copy tempBuf to buffer.
		buffer = append(buffer, tempBuf[:count]...)
		remaining -= count
	}
	return remaining, nil
}

func ReadFully2(in io.ReadSeeker, buffer []byte) (int, error) {
	return ReadFully(in, buffer, 0, len(buffer))
}

// FIXME(kpfaulkner) really unsure what this is supposed to do. Skip some content... then read more?
func SkipFully(in io.ReadSeeker, n int64) (int, error) {
	remaining := n
	var sz int64
	if n < tempBufSize {
		sz = n
	} else {
		sz = tempBufSize
	}

	tempBuf := make([]byte, sz)
	for remaining > 0 {
		skipped, err := in.Read(tempBuf)
		if err != nil {
			return 0, err
		}

		remaining -= int64(skipped)
		if skipped == 0 {
			break
		}
	}
	if remaining == 0 {
		return 0, nil
	}
	buffer := make([]byte, 4096)
	for remaining > int64(len(buffer)) {
		k, err := ReadFully2(in, buffer)
		if err != nil {
			return 0, err
		}
		remaining = remaining - int64(len(buffer)) + int64(k)
		if k != 0 {
			return int(remaining), nil
		}
	}
	return ReadFully(in, buffer, 0, int(remaining))
}
