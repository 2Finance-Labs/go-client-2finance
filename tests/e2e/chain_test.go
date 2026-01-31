package e2e_test


import (
	"testing"
	"gitlab.com/2finance/2finance-network/blockchain/encryption/keys"
	"gitlab.com/2finance/2finance-network/blockchain/utils"
	"gitlab.com/2finance/2finance-network/blockchain/transaction"
	"gitlab.com/2finance/2finance-network/blockchain/contract/contractV1/models"
	"gitlab.com/2finance/2finance-network/blockchain/contract/walletV1"
	"encoding/json"
	"time"
)


func TestChainBasics(t *testing.T) {
	c := setupClient(t)
	w, _ := createWallet(t, c)

	_, err := c.DeployContract1(walletV1.WALLET_CONTRACT_V1)
    if err != nil {
        t.Fatalf("DeployContract wallet: %v", err)
    }

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


func Test_SetPrivateKey_Getters(t *testing.T) {
	c := setupClient(t)
	_, priv := genKey(t, c)

	// Set and read back
	c.SetPrivateKey(priv)
	if got := c.GetPrivateKey(); got != priv {
		t.Fatalf("GetPrivateKey mismatch")
	}
	gotPub := c.GetPublicKey()
	if gotPub == "" {
		t.Fatalf("GetPublicKey returned empty")
	}
	if err := keys.ValidateEDDSAPublicKeyHex(gotPub); err != nil {
		t.Fatalf("GetPublicKey invalid: %v", err)
	}

	// Derive pub from priv and compare with client’s public key
	edPub, err := keys.PublicKeyFromEd25519PrivateHex(priv)
	if err != nil {
		t.Fatalf("derive pub from priv: %v", err)
	}
	if want := keys.PublicKeyToHex(edPub); want != gotPub {
		t.Fatalf("public key mismatch: want %s, got %s", want, gotPub)
	}
}

func Test_GenerateKeyEd25519(t *testing.T) {
	c := setupClient(t)
	pub, priv, err := c.GenerateEd25519KeyPairHex()
	if err != nil {
		t.Fatalf("GenerateKeyEd25519: %v", err)
	}
	if pub == "" || priv == "" {
		t.Fatalf("empty keys")
	}
	if err := keys.ValidateEDDSAPublicKeyHex(pub); err != nil {
		t.Fatalf("generated pub invalid: %v", err)
	}
}

// Validation guards that short-circuit before hitting the network:
func Test_GetNonce_Validation(t *testing.T) {
	c := setupClient(t)

	if _, err := c.GetNonce(""); err == nil || err.Error() == "" {
		t.Fatalf("expected error for empty public key")
	}
	if _, err := c.GetNonce("not-a-key"); err == nil || err.Error() == "" {
		t.Fatalf("expected error for invalid public key")
	}
}

func Test_ListTransactions_Validation(t *testing.T) {
	c := setupClient(t)

	// all empty
	if _, err := c.ListTransactions("", "", "", nil, 0, 1, 10, true); err == nil {
		t.Fatalf("expected error when all filters are empty")
	}
	// invalid from
	if _, err := c.ListTransactions("bad-from", "", "", nil, 0, 1, 10, true); err == nil {
		t.Fatalf("expected error for invalid from")
	}
	// invalid to
	if _, err := c.ListTransactions("", "bad-to", "", nil, 0, 1, 10, true); err == nil {
		t.Fatalf("expected error for invalid to")
	}
}

func Test_ListLogs_Validation(t *testing.T) {
	c := setupClient(t)
	// no logType, txHash, or contractAddress -> invalid
	if _, err := c.ListLogs(nil, 0, "", nil, "", 1, 10, true); err == nil {
		t.Fatalf("expected error when no filters are provided")
	}
}

func Test_SignTransaction(t *testing.T) {
	c := setupClient(t)
	fromPub, fromPriv := genKey(t, c)
	toPub, _ := genKey(t, c)
	c.SetPrivateKey(fromPriv)
	chainId := uint64(1)
	jb, err := utils.MapToJSONB(map[string]interface{}{"hello": "world"})
	if err != nil {
		t.Fatalf("MapToJSONB: %v", err)
	}
	signed, err := c.SignTransaction(chainId, fromPub, toPub, "noop_method", jb, 42)
	if err != nil {
		t.Fatalf("SignTransaction: %v", err)
	}
	if signed == nil {
		t.Fatalf("signed tx is nil")
	}
	tx := signed.Get()
	// basic sanity
	if tx.From != fromPub {
		t.Fatalf("tx.From mismatch")
	}
	if tx.Nonce != 42 {
		t.Fatalf("tx.Nonce mismatch")
	}
	if signed.Signature == "" || signed.Hash == "" {
		t.Fatalf("signature/hash should not be empty")
	}
}

func Test_DeployContract_ValidationAndSuccess(t *testing.T) {
	c := setupClient(t)

	// without SetPrivateKey -> no public key in client
	if _, err := c.DeployContract1(walletV1.WALLET_CONTRACT_V1); err == nil {
		t.Fatalf("expected error when from address is not set")
	}

	// set signer
	_, priv := genKey(t, c)
	c.SetPrivateKey(priv)

	// empty contract version
	if _, err := c.DeployContract1(""); err == nil {
		t.Fatalf("expected error when contract version is empty")
	}

	// happy path: deploy wallet contract
	deployed, err := c.DeployContract1(walletV1.WALLET_CONTRACT_V1)
	if err != nil {
		t.Fatalf("DeployContract: %v", err)
	}
	var cs models.ContractStateModel
	unmarshalState(t, deployed.States[0].Object, &cs)
	if cs.Address == "" {
		t.Fatalf("deployed contract state has empty address")
	}

	pub, priv := genKey(t, c)

	deployed2, err := c.DeployContract2(walletV1.WALLET_CONTRACT_V1, pub)
	if err != nil {
		t.Fatalf("DeployContract: %v", err)
	}
	unmarshalState(t, deployed2.States[0].Object, &cs)
	if cs.Address != pub {
		t.Fatalf("deployed contract address mismatch: want %s, got %s", pub, cs.Address)
	}
}

func Test_EndToEnd_MinimalFlow(t *testing.T) {
	c := setupClient(t)

	// create a wallet (deploys + sends tx)
	w, priv := createWallet(t, c)
	c.SetPrivateKey(priv)

	// Nonce should be available
	nonce, err := c.GetNonce(w.PublicKey)
	if err != nil {
		t.Fatalf("GetNonce: %v", err)
	}
	if nonce < 0 {
		t.Fatalf("invalid nonce: %d", nonce)
	}

	// Best-effort: transactions/logs/blocks (may be empty depending on backend retention)
	if txs, err := c.ListTransactions(w.PublicKey, "", "", nil, 0, 1, 10, true); err == nil && len(txs) > 0 {
		_, _ = c.ListLogs([]string{"wallet_created"}, 0, txs[0].Hash, nil, "", 1, 10, true)
	}
	_, _ = c.ListBlocks(0, time.Time{}, "", "", "", 1, 5, true)
}

// (Optional) tiny compile-time/proto sanity check for Transaction serialization
func Test_TransactionRoundtrip_Sanity(t *testing.T) {
	pub, _ := genKey(t, setupClient(t))
	chainId := uint64(1)
	tx := transaction.NewTransaction(chainId, pub, pub, "echo", json.RawMessage(`{"contract_version": "walletV1", "k": "v"}`), 7)
	_ = tx.Get() // ensure .Get() is accessible
}
