package network

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/url"
	"strings"
	"time"

	"github.com/2Finance-Labs/2finance-sdk-client/internal/service"
	"github.com/2Finance-Labs/2finance-sdk-client/protocol"
)

const (
	transactionsV2Path           = "/v2/2finance-network/transactions"
	protocolV2SubmissionAttempts = 6
)

type ProtocolV2Commit struct {
	Sequence           uint64     `json:"sequence"`
	AdmissionHeight    uint64     `json:"admission_height"`
	Status             string     `json:"status"`
	BlockHeight        *uint64    `json:"block_height,omitempty"`
	BlockHash          string     `json:"block_hash,omitempty"`
	MembershipPosition *uint64    `json:"membership_position,omitempty"`
	CommittedAt        time.Time  `json:"committed_at"`
	IncludedAt         *time.Time `json:"included_at,omitempty"`
}

type ProtocolV2Artifact struct {
	Runtime           string `json:"runtime"`
	RuntimeVersion    uint32 `json:"runtime_version"`
	ExecutionSpecHash string `json:"execution_spec_hash"`
	ReceiptHash       string `json:"receipt_hash"`
	ArtifactHash      string `json:"artifact_hash"`
	LogRoot           string `json:"log_root"`
	WriteSetRoot      string `json:"write_set_root"`
	StateRootBefore   string `json:"state_root_before"`
	StateRootAfter    string `json:"state_root_after"`
	TimestampHash     string `json:"timestamp_hash"`
}

type ProtocolV2Membership struct {
	Position        uint64 `json:"position"`
	TransactionHash string `json:"transaction_hash"`
	CommitSequence  uint64 `json:"commit_sequence"`
}

type ProtocolV2Finality struct {
	BlockEnvelope     json.RawMessage        `json:"block_envelope"`
	CommitCertificate json.RawMessage        `json:"commit_certificate"`
	Membership        []ProtocolV2Membership `json:"membership"`
}

type ProtocolV2TransactionStatus struct {
	ProtocolVersion   uint32              `json:"protocol_version"`
	ChainID           string              `json:"chain_id"`
	TransactionHash   string              `json:"transaction_hash"`
	SignedTransaction json.RawMessage     `json:"signed_transaction"`
	Commit            ProtocolV2Commit    `json:"commit"`
	Artifact          ProtocolV2Artifact  `json:"artifact"`
	Finality          *ProtocolV2Finality `json:"finality,omitempty"`
}

func (status ProtocolV2TransactionStatus) DecodeSignedTransaction() (protocol.SignedTransaction, error) {
	var signed protocol.SignedTransaction
	if err := json.Unmarshal(status.SignedTransaction, &signed); err != nil {
		return protocol.SignedTransaction{}, err
	}
	return signed, nil
}

type ProtocolV2SubmitResult struct {
	ExecutionOutput json.RawMessage             `json:"execution_output"`
	Transaction     ProtocolV2TransactionStatus `json:"transaction"`
}

// ProtocolV2ExecutionLog is the runtime-neutral log shape returned for both
// Native Go events and EVM LOG opcodes.
type ProtocolV2ExecutionLog struct {
	ChainID         string          `json:"chain_id"`
	TransactionHash string          `json:"transaction_hash"`
	LogIndex        uint64          `json:"log_index"`
	SourceLogIndex  uint64          `json:"source_log_index"`
	Runtime         string          `json:"runtime"`
	ContractAddress string          `json:"contract_address"`
	ContractVersion string          `json:"contract_version"`
	EventSignature  string          `json:"event_signature"`
	Topics          []string        `json:"topics"`
	Data            json.RawMessage `json:"data"`
	RawData         string          `json:"raw_data"`
	LeafHash        string          `json:"leaf_hash"`
	CreatedAt       time.Time       `json:"created_at"`
}

type ProtocolV2ExecutionLogsPage struct {
	Logs  []ProtocolV2ExecutionLog `json:"logs"`
	Page  int                      `json:"page"`
	Limit int                      `json:"limit"`
}

type protocolV2Response[T any] struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data T      `json:"data"`
}

// SubmitSignedTransactionV2 executes and commits the signed business
// transaction before block formation. The returned status may already contain
// finality when an idempotent submission is queried after inclusion.
func (c *Client) SubmitSignedTransactionV2(
	ctx context.Context,
	tx protocol.SignedTransaction,
) (ProtocolV2SubmitResult, error) {
	if strings.TrimSpace(tx.Hash) == "" || strings.TrimSpace(tx.Signature) == "" {
		return ProtocolV2SubmitResult{}, errors.New("network v2: signed transaction hash and signature are required")
	}
	for attempt := 1; attempt <= protocolV2SubmissionAttempts; attempt++ {
		var response protocolV2Response[ProtocolV2SubmitResult]
		err := c.Post(ctx, transactionsV2Path, tx, &response)
		if err == nil {
			return response.Data, nil
		}
		if !isTransientProtocolV2SubmissionError(err) || attempt == protocolV2SubmissionAttempts {
			return ProtocolV2SubmitResult{}, err
		}
		if err := waitProtocolV2SubmissionRetry(ctx, attempt); err != nil {
			return ProtocolV2SubmitResult{}, err
		}
	}
	return ProtocolV2SubmitResult{}, errors.New("network v2: transaction submission attempts exhausted")
}

// TransactionFinalityV2 returns the exact signed body, commit-first metadata,
// deterministic execution artifact and, after inclusion, the block envelope,
// commit certificate and ordered membership proof.
func (c *Client) TransactionFinalityV2(
	ctx context.Context,
	transactionHash string,
) (ProtocolV2TransactionStatus, error) {
	hash, err := normalizedTransactionHashV2(transactionHash)
	if err != nil {
		return ProtocolV2TransactionStatus{}, err
	}
	var response protocolV2Response[ProtocolV2TransactionStatus]
	path := transactionsV2Path + "/" + url.PathEscape(hash) + "/finality"
	if err := c.Get(ctx, path, &response); err != nil {
		return ProtocolV2TransactionStatus{}, err
	}
	return response.Data, nil
}

func (c *Client) ExecutionLogsV2(
	ctx context.Context,
	query url.Values,
) (ProtocolV2ExecutionLogsPage, error) {
	path := "/v2/2finance-network/logs"
	if encoded := query.Encode(); encoded != "" {
		path += "?" + encoded
	}
	var response protocolV2Response[ProtocolV2ExecutionLogsPage]
	if err := c.Get(ctx, path, &response); err != nil {
		return ProtocolV2ExecutionLogsPage{}, err
	}
	return response.Data, nil
}

func normalizedTransactionHashV2(value string) (string, error) {
	value = strings.ToLower(strings.TrimSpace(value))
	decoded, err := hex.DecodeString(value)
	if err != nil || len(decoded) != 32 {
		return "", errors.New("network v2: transaction hash must be 32-byte hexadecimal")
	}
	return value, nil
}

func isTransientProtocolV2SubmissionError(err error) bool {
	var httpErr *service.HTTPError
	if !errors.As(err, &httpErr) || httpErr.StatusCode != 422 {
		return false
	}
	message := strings.ToLower(string(httpErr.Body))
	return strings.Contains(message, "can't serialize access") ||
		strings.Contains(message, "try restarting transaction") ||
		strings.Contains(message, "transaction is killed") ||
		strings.Contains(message, "lock wait timeout") ||
		strings.Contains(message, "context deadline exceeded")
}

func waitProtocolV2SubmissionRetry(ctx context.Context, attempt int) error {
	delay := 50 * time.Millisecond * time.Duration(1<<(attempt-1))
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
