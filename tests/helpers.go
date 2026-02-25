package tests

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gitlab.com/2finance/2finance-network/blockchain/log"
	"gitlab.com/2finance/2finance-network/blockchain/utils"
)

func UnmarshalJSONB[T any](t *testing.T, jb utils.JSONB) T {
	t.Helper()

	data, err := json.Marshal(jb)
	require.NoError(t, err, "json.Marshal(JSONB) failed")

	var out T
	err = json.Unmarshal(data, &out)
	require.NoError(t, err, "json.Unmarshal(Event) failed. Raw=%s", string(data))

	return out
}

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

func AssertLogBase(t *testing.T, l log.Log) {
	t.Helper()

	require.NotEmpty(t, l.LogType, "log_type empty")
	require.NotZero(t, l.LogIndex, "log_index is zero")
	require.NotEmpty(t, l.TransactionHash, "transaction_hash empty")
	require.NotEmpty(t, l.ContractVersion, "contract_version empty")
	require.NotEmpty(t, l.ContractAddress, "contract_address empty")
	require.False(t, l.CreatedAt.IsZero(), "created_at is zero")

	require.LessOrEqual(t, l.CreatedAt.Unix(), time.Now().Add(2*time.Minute).Unix(), "created_at seems in the future")
}

func RequireMapFieldString(t *testing.T, m map[string]any, key string) string {
	t.Helper()
	v, ok := m[key]
	require.True(t, ok, "event missing field %q. event=%v", key, m)

	s, ok := v.(string)
	require.True(t, ok, "event field %q not a string (got %T=%v)", key, v, v)
	return s
}

func RequireMapFieldEqual(t *testing.T, m map[string]any, key string, want any) {
	t.Helper()
	got, ok := m[key]
	require.True(t, ok, "event missing field %q. event=%v", key, m)
	require.Equal(t, want, got, "event field %q mismatch", key)
}

func RequireMapFieldNumberAsUint(t *testing.T, m map[string]any, key string) uint {
	t.Helper()
	v, ok := m[key]
	require.True(t, ok, "event missing field %q. event=%v", key, m)

	// json.Unmarshal em map usa float64 para números
	f, ok := v.(float64)
	require.True(t, ok, "event field %q not numeric (got %T=%v)", key, v, v)
	if f < 0 {
		t.Fatalf("event field %q is negative: %v", key, f)
	}
	return uint(f)
}

func FailEventUnknownShape(t *testing.T, m map[string]any) {
	t.Helper()
	t.Fatalf("unknown event shape: %v", fmt.Sprintf("%#v", m))
}
