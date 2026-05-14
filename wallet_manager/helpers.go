package wallet_manager

import (
	"crypto/rand"
	"io"
)

func randomBytes(size int) ([]byte, error) {
	data := make([]byte, size)

	if _, err := io.ReadFull(rand.Reader, data); err != nil {
		return nil, err
	}

	return data, nil
}

func clearBytes(data []byte) {
	for i := range data {
		data[i] = 0
	}
}

func cloneBytes(data []byte) []byte {
	if data == nil {
		return nil
	}

	cloned := make([]byte, len(data))
	copy(cloned, data)

	return cloned
}