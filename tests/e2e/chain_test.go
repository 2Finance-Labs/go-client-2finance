package e2e_test


import (
	"testing"
	"gitlab.com/2finance/2finance-network/blockchain/encryption/keys"
	"gitlab.com/2finance/2finance-network/blockchain/utils"
	"gitlab.com/2finance/2finance-network/blockchain/transaction"
	// "gitlab.com/2finance/2finance-network/blockchain/contract/contractV1/models"
	"gitlab.com/2finance/2finance-network/blockchain/contract/contractV1/domain"
	"gitlab.com/2finance/2finance-network/blockchain/contract/walletV1"
	"gitlab.com/2finance/2finance-network/blockchain/log"
	"encoding/json"
	"time"
	"github.com/stretchr/testify/assert"
)


func TestContractDeployment1(t *testing.T) {
	wm := setupWalletManager(t)
	c := setupClient(t, wm)

	contractOutput, err := c.DeployContract1(walletV1.WALLET_CONTRACT_V1)
    if err != nil {
        t.Fatalf("DeployContract wallet: %v", err)
    }

	unmarshaledLog, err := utils.UnmarshalLog[log.Log](contractOutput.Logs[0])
	if err != nil {
		t.Fatalf("UnmarshalLog: %v", err)
	}

	unmarshaledEvent, err := utils.UnmarshalEvent[domain.Contract](unmarshaledLog.Event)
	if err != nil {
		t.Fatalf("UnmarshalEvent: %v", err)
	}

	assert.NotEmpty(t, unmarshaledEvent.Address, "deployed contract address should not be empty")
	assert.Equal(t, unmarshaledEvent.ContractVersion, walletV1.WALLET_CONTRACT_V1, "deployed contract version mismatch")

	if txs, err := c.ListTransactions(unmarshaledEvent.Address, "", "", nil, 0, 1, 10, true); err == nil && len(txs) > 0 {
		_, _ = c.ListLogs([]string{"deploy contract"}, 0, txs[0].Hash, nil, "", 1, 10, true)
	}
}

func TestContractDeployment2(t *testing.T) {
	wm := setupWalletManager(t)
	c := setupClient(t, wm)

	genKeyPub, genKeyPriv := genKey(t, wm)
	wm.SetPrivateKey(genKeyPriv)

	contractOutput, err := c.DeployContract2(walletV1.WALLET_CONTRACT_V1, genKeyPub)
	if err != nil {
		t.Fatalf("DeployContract2 wallet: %v", err)
	}

	unmarshaledLog, err := utils.UnmarshalLog[log.Log](contractOutput.Logs[0])
	if err != nil {
		t.Fatalf("UnmarshalLog: %v", err)
	}

	unmarshaledEvent, err := utils.UnmarshalEvent[domain.Contract](unmarshaledLog.Event)
	if err != nil {
		t.Fatalf("UnmarshalEvent: %v", err)
	}

	assert.NotEmpty(t, unmarshaledEvent.Address, "deployed contract address should not be empty")
	assert.Equal(t, unmarshaledEvent.ContractVersion, walletV1.WALLET_CONTRACT_V1, "deployed contract version mismatch")

	if txs, err := c.ListTransactions(unmarshaledEvent.Address, "", "", nil, 0, 1, 10, true); err == nil && len(txs) > 0 {
		_, _ = c.ListLogs([]string{"deploy contract"}, 0, txs[0].Hash, nil, "", 1, 10, true)
	}
}

func Test_SetPrivateKey_Getters(t *testing.T) {
	wm := setupWalletManager(t)
	_, priv := genKey(t, wm)
	wm.SetPrivateKey(priv)

	// Set and read back
	// if got := c.GetPrivateKey(); got != priv {
	// 	t.Fatalf("GetPrivateKey mismatch")
	// }
	gotPub := wm.GetPublicKey()
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
	wm := setupWalletManager(t)
	pub, priv, err := wm.GenerateEd25519KeyPairHex()
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

func Test_ListTransactions_Validation(t *testing.T) {
	wm := setupWalletManager(t)
	c := setupClient(t, wm)

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
	wm := setupWalletManager(t)
	c := setupClient(t, wm)
	// no logType, txHash, or contractAddress -> invalid
	if _, err := c.ListLogs(nil, 0, "", nil, "", 1, 10, true); err == nil {
		t.Fatalf("expected error when no filters are provided")
	}
}

func Test_SignTransaction(t *testing.T) {
	wm := setupWalletManager(t)
	fromPub, fromPriv := genKey(t, wm)
	toPub, _ := genKey(t, wm)
	wm.SetPrivateKey(fromPriv)
	chainId := uint8(1)
	version := uint8(1)
	uuid7, err := utils.NewUUID7()
	if err != nil {
		t.Fatalf("NewUUID7: %v", err)
	}
	jb, err := utils.MapToJSONB(map[string]interface{}{"hello": "world"})
	if err != nil {
		t.Fatalf("MapToJSONB: %v", err)
	}
	signed, err := wm.SignTransaction(chainId, fromPub, toPub, "noop_method", jb, version, uuid7)
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
	if tx.Version != 1 {
		t.Fatalf("tx.Version mismatch")
	}
	if signed.Signature == "" || signed.Hash == "" {
		t.Fatalf("signature/hash should not be empty")
	}
}

func Test_DeployContract_ValidationAndSuccess(t *testing.T) {
	wm := setupWalletManager(t)
	c := setupClient(t, wm)

	// without SetPrivateKey -> no public key in client
	if _, err := c.DeployContract1(walletV1.WALLET_CONTRACT_V1); err == nil {
		t.Fatalf("expected error when from address is not set")
	}

	// set signer
	_, priv := genKey(t, wm)
	wm.SetPrivateKey(priv)

	// empty contract version
	if _, err := c.DeployContract1(""); err == nil {
		t.Fatalf("expected error when contract version is empty")
	}

	// happy path: deploy wallet contract
	deployedContract, err := c.DeployContract1(walletV1.WALLET_CONTRACT_V1)
	if err != nil {
		t.Fatalf("DeployContract: %v", err)
	}

	contractLog, err := utils.UnmarshalLog[log.Log](deployedContract.Logs[0])
	if err != nil {
		t.Fatalf("UnmarshalLog (AddWallet.Logs[0]): %v", err)
	}

	contractDomain, err := utils.UnmarshalEvent[domain.Contract](contractLog.Event)
	if err != nil {
		t.Fatalf("UnmarshalEvent (AddWallet.Logs[0]): %v", err)
	}

	assert.NotEmpty(t, contractDomain.Address, "deployed contract address should not be empty")
	assert.Equal(t, contractDomain.ContractVersion, walletV1.WALLET_CONTRACT_V1, "deployed contract version mismatch")

	pub, priv := genKey(t, wm)
	wm.SetPrivateKey(priv)

	deployed2, err := c.DeployContract2(walletV1.WALLET_CONTRACT_V1, pub)
	if err != nil {
		t.Fatalf("DeployContract: %v", err)
	}
	contractLog2, err := utils.UnmarshalLog[log.Log](deployed2.Logs[0])
	if err != nil {
		t.Fatalf("UnmarshalLog (DeployContract2.Logs[0]): %v", err)
	}
	contractDomain2, err := utils.UnmarshalEvent[domain.Contract](contractLog2.Event)
	if err != nil {
		t.Fatalf("UnmarshalEvent (DeployContract2.Logs[0]): %v", err)
	}
	assert.NotEmpty(t, contractDomain2.Address, "deployed contract address should not be empty")
	assert.Equal(t, contractDomain2.ContractVersion, walletV1.WALLET_CONTRACT_V1, "deployed contract version mismatch")
}

func Test_EndToEnd_MinimalFlow(t *testing.T) {
	wm := setupWalletManager(t)
	c := setupClient(t, wm)

	// create a wallet (deploys + sends tx)
	w, priv := createWallet(t, c, wm)
	wm.SetPrivateKey(priv)

	// Best-effort: transactions/logs/blocks (may be empty depending on backend retention)
	if txs, err := c.ListTransactions(w.PublicKey, "", "", nil, 0, 1, 10, true); err == nil && len(txs) > 0 {
		_, _ = c.ListLogs([]string{"wallet_created"}, 0, txs[0].Hash, nil, "", 1, 10, true)
	}
	_, _ = c.ListBlocks(0, time.Time{}, "", "", "", 1, 5, true)
}

// (Optional) tiny compile-time/proto sanity check for Transaction serialization
func Test_TransactionRoundtrip_Sanity(t *testing.T) {
	wm := setupWalletManager(t)
	pub, priv := genKey(t, wm)
	wm.SetPrivateKey(priv)

	chainId := uint8(1)
	version := uint8(1)
	uuid7, err := utils.NewUUID7()
	if err != nil {
		t.Fatalf("NewUUID7: %v", err)
	}
	tx := transaction.NewTransaction(chainId, pub, pub, "echo", json.RawMessage(`{"contract_version": "walletV1", "k": "v"}`), version, uuid7)
	_ = tx.Get() // ensure .Get() is accessible
}
