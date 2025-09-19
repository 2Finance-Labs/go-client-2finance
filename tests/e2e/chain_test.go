package e2e_test


import (
	"testing"
	"time"
)


func TestChainBasics(t *testing.T) {
	c := setupClient(t)
	w, _ := createWallet(t, c)

	if got, err := c.GetWallet(w.PublicKey); err != nil {
	t.Fatalf("GetWallet: %v", err)
	} else {
	var w2 struct{ PublicKey string `json:"public_key"` }
	unmarshalState(t, got.States[0].Object, &w2)
	if w2.PublicKey != w.PublicKey { t.Fatalf("wallet mismatch: %s != %s", w2.PublicKey, w.PublicKey) }
	}


	nonce, err := c.GetNonce(w.PublicKey)
	if err != nil { t.Fatalf("GetNonce: %v", err) }
	if nonce < 0 { t.Fatalf("invalid nonce: %d", nonce) }


	// Transactions / Logs (best effort – may be zero depending on backend retention)
	if txs, err := c.ListTransactions(w.PublicKey, "", "", nil, 0, 1, 10, true); err == nil && len(txs) > 0 {
	_, _ = c.ListLogs([]string{"wallet_created"}, 0, txs[0].Hash, nil, "", 1, 10, true)
	}


	// Blocks (best effort)
	_, _ = c.ListBlocks(0, time.Time{}, "", "", "", 1, 5, true)
}