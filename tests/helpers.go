package tests

import (
	"encoding/json"
	"testing"
)

// Igual ao padrão do Mint: Marshal(raw) -> Unmarshal(bytes, &dest)
func UnmarshalState[T any](t *testing.T, raw any, dest *T) {
	t.Helper()

	b, err := json.Marshal(raw)
	if err != nil {
		t.Fatalf("Error marshaling state object: %v", err)
	}

	if err := json.Unmarshal(b, dest); err != nil {
		t.Fatalf("Error unmarshalling into %T: %v", *dest, err)
	}
}

// Variante mais “enxuta” (retorna o valor pronto)
func MustState[T any](t *testing.T, raw any) T {
	t.Helper()

	var v T
	UnmarshalState(t, raw, &v)
	return v
}
