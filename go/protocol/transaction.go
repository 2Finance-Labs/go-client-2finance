package protocol

import (
	"encoding/json"
	"fmt"

	"gitlab.com/2finance/2finance-network/blockchain/transaction"
)

// PreparedTransaction is the canonical unsigned transaction envelope shared
// with MCP tools. Hash and signature are expected to be empty before signing.
type PreparedTransaction struct {
	ChainID       uint8                              `json:"chain_id"`
	From          string                             `json:"from"`
	To            string                             `json:"to"`
	Method        string                             `json:"method"`
	Data          map[string]interface{}             `json:"data"`
	Version       uint8                              `json:"version"`
	UUID7         string                             `json:"uuid7"`
	Hash          string                             `json:"hash"`
	Signature     string                             `json:"signature"`
	Authorization *transaction.AuthorizationEnvelope `json:"authorization,omitempty"`
}

// SignedTransaction keeps the same JSON shape after the wallet fills hash and
// signature. Keeping a separate type makes call sites explicit about state.
type SignedTransaction struct {
	ChainID       uint8                              `json:"chain_id"`
	From          string                             `json:"from"`
	To            string                             `json:"to"`
	Method        string                             `json:"method"`
	Data          map[string]interface{}             `json:"data"`
	Version       uint8                              `json:"version"`
	UUID7         string                             `json:"uuid7"`
	Hash          string                             `json:"hash"`
	Signature     string                             `json:"signature"`
	Authorization *transaction.AuthorizationEnvelope `json:"authorization,omitempty"`
}

type PreparedTransactionResult struct {
	Workflow            string              `json:"workflow"`
	UnsignedTransaction PreparedTransaction `json:"unsigned_transaction"`
	NextStep            string              `json:"next_step"`
}

func SignedTransactionFromNetwork(tx *transaction.Transaction) (SignedTransaction, error) {
	if tx == nil {
		return SignedTransaction{}, fmt.Errorf("transaction is nil")
	}

	data := map[string]interface{}{}
	if len(tx.Data) > 0 {
		if err := json.Unmarshal(tx.Data, &data); err != nil {
			return SignedTransaction{}, fmt.Errorf("failed to unmarshal transaction data: %w", err)
		}
	}

	return SignedTransaction{
		ChainID:       tx.ChainID,
		From:          tx.From,
		To:            tx.To,
		Method:        tx.Method,
		Data:          data,
		Version:       tx.Version,
		UUID7:         tx.UUID7,
		Hash:          tx.Hash,
		Signature:     tx.Signature,
		Authorization: tx.Authorization,
	}, nil
}

func (tx SignedTransaction) ToNetworkTransaction() (*transaction.Transaction, error) {
	data, err := json.Marshal(tx.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal transaction data: %w", err)
	}

	return &transaction.Transaction{
		ChainID:       tx.ChainID,
		From:          tx.From,
		To:            tx.To,
		Method:        tx.Method,
		Data:          data,
		Version:       tx.Version,
		UUID7:         tx.UUID7,
		Hash:          tx.Hash,
		Signature:     tx.Signature,
		Authorization: tx.Authorization,
	}, nil
}
