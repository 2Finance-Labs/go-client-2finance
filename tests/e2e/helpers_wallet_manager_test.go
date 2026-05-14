package e2e_test

func cloneBytes(data []byte) []byte {
	if data == nil {
		return nil
	}

	cloned := make([]byte, len(data))
	copy(cloned, data)

	return cloned
}