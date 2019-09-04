package worker

import (
	"bytes"
	"fmt"
	"io"
	"os"
)

func ReadFile(path string, maxSize int64) ([]byte, error) {

	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// SAST: exception 'utils.ReadFile prone to resource exhaustion'
	size, err := f.Seek(0, io.SeekEnd)
	if err != nil {
		return nil, err
	} else if size > maxSize {
		return nil, fmt.Errorf("invalid file size: %d", size)
	}

	if _, err := f.Seek(0, io.SeekStart); err != nil {
		return nil, err
	}

	buf := bytes.NewBuffer(make([]byte, 0, size))
	if _, err := io.Copy(buf, f); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
