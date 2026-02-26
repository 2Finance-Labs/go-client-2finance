package tests

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gitlab.com/2finance/2finance-network/blockchain/log"
	"gitlab.com/2finance/2finance-network/blockchain/types"
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

func RequireStates(t *testing.T, states []any, want int) {
	t.Helper()
	require.Len(t, states, want, "unexpected number of states")
	for i := 0; i < want; i++ {
		require.NotNil(t, states[i], "state[%d] is nil", i)
	}
}

func RequireStateObjectsNotNil(t *testing.T, outStates []types.State, want int) {
	t.Helper()
	require.Len(t, outStates, want, "unexpected number of states")
	for i := 0; i < want; i++ {
		require.NotNil(t, outStates[i].Object, "states[%d].Object is nil", i)
	}
}

func RequireLogsBase(t *testing.T, logs []log.Log, want int) {
	t.Helper()
	require.Len(t, logs, want, "unexpected number of logs")

	txHash := logs[0].TransactionHash
	contractAddr := logs[0].ContractAddress
	contractVer := logs[0].ContractVersion

	for i := range logs {
		AssertLogBase(t, logs[i])
		require.Equal(t, txHash, logs[i].TransactionHash, "logs[%d] tx hash mismatch", i)
		require.Equal(t, contractAddr, logs[i].ContractAddress, "logs[%d] contract address mismatch", i)
		require.Equal(t, contractVer, logs[i].ContractVersion, "logs[%d] contract version mismatch", i)
	}
}

func RequireLogTypesInOrder(t *testing.T, logs []log.Log, want []string) {
	t.Helper()
	require.Len(t, logs, len(want), "log count mismatch for log type validation")
	for i := range want {
		require.Equal(t, want[i], logs[i].LogType, "logs[%d].LogType mismatch", i)
	}
}

func FindLogByType(t *testing.T, logs []log.Log, logType string) log.Log {
	t.Helper()
	for _, l := range logs {
		if l.LogType == logType {
			return l
		}
	}
	t.Fatalf("log of type %q not found", logType)
	return log.Log{}
}

func RequireMapFieldString(t *testing.T, m map[string]any, key string) string {
	t.Helper()
	v, ok := m[key]
	require.True(t, ok, "event missing field %q. event=%v", key, m)
	s, ok := v.(string)
	require.True(t, ok, "event field %q not a string (got %T=%v)", key, v, v)
	return s
}
