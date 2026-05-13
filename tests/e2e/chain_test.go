package e2e_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gitlab.com/2finance/2finance-network/blockchain/contract/contractV1/domain"
	"gitlab.com/2finance/2finance-network/blockchain/contract/walletV1"
	"gitlab.com/2finance/2finance-network/blockchain/encryption/keys"
	"gitlab.com/2finance/2finance-network/blockchain/log"
	"gitlab.com/2finance/2finance-network/blockchain/transaction"
	"gitlab.com/2finance/2finance-network/blockchain/utils"
)

func TestContractDeployment1(t *testing.T) {
	signer := setupSignerWallet(t)
	c := setupClient(t, signer.Wallet)

	useWallet(t, c, signer.Wallet)

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
	assert.Equal(t, walletV1.WALLET_CONTRACT_V1, unmarshaledEvent.ContractVersion, "deployed contract version mismatch")

	if txs, err := c.ListTransactions(unmarshaledEvent.Address, "", "", nil, 0, 1, 10, true); err == nil && len(txs) > 0 {
		_, _ = c.ListLogs([]string{"deploy contract"}, 0, txs[0].Hash, nil, "", 1, 10, true)
	}
}

func TestContractDeployment2(t *testing.T) {
	signer := setupSignerWallet(t)
	c := setupClient(t, signer.Wallet)

	tmpWM := setupWalletManager(t)
	contractAddress, _ := genKey(t, tmpWM)

	useWallet(t, c, signer.Wallet)

	contractOutput, err := c.DeployContract2(walletV1.WALLET_CONTRACT_V1, contractAddress)
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
	assert.Equal(t, walletV1.WALLET_CONTRACT_V1, unmarshaledEvent.ContractVersion, "deployed contract version mismatch")
	assert.Equal(t, contractAddress, unmarshaledEvent.Address, "deployed contract address mismatch")

	if txs, err := c.ListTransactions(unmarshaledEvent.Address, "", "", nil, 0, 1, 10, true); err == nil && len(txs) > 0 {
		_, _ = c.ListLogs([]string{"deploy contract"}, 0, txs[0].Hash, nil, "", 1, 10, true)
	}
}

func Test_ImportWallet_Getters(t *testing.T) {
	signer := setupSignerWallet(t)

	gotPub := signer.Wallet.GetPublicKey()
	if gotPub == "" {
		t.Fatalf("GetPublicKey returned empty")
	}

	if err := keys.ValidateEDDSAPublicKeyHex(gotPub); err != nil {
		t.Fatalf("GetPublicKey invalid: %v", err)
	}

	edPub, err := keys.PublicKeyFromEd25519PrivateHex(signer.PrivateKey)
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

	if _, err := c.ListTransactions("", "", "", nil, 0, 1, 10, true); err == nil {
		t.Fatalf("expected error when all filters are empty")
	}

	if _, err := c.ListTransactions("bad-from", "", "", nil, 0, 1, 10, true); err == nil {
		t.Fatalf("expected error for invalid from")
	}

	if _, err := c.ListTransactions("", "bad-to", "", nil, 0, 1, 10, true); err == nil {
		t.Fatalf("expected error for invalid to")
	}
}

func Test_ListLogs_Validation(t *testing.T) {
	wm := setupWalletManager(t)
	c := setupClient(t, wm)

	if _, err := c.ListLogs(nil, 0, "", nil, "", 1, 10, true); err == nil {
		t.Fatalf("expected error when no filters are provided")
	}
}

func Test_SignTransaction(t *testing.T) {
	signer := setupSignerWallet(t)

	tmpWM := setupWalletManager(t)
	toPub, _ := genKey(t, tmpWM)

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

	signed, err := signer.Wallet.SignTransaction(
		chainId,
		signer.PublicKey,
		toPub,
		"noop_method",
		jb,
		version,
		uuid7,
	)
	if err != nil {
		t.Fatalf("SignTransaction: %v", err)
	}

	if signed == nil {
		t.Fatalf("signed tx is nil")
	}

	tx := signed.Get()

	if tx.From != signer.PublicKey {
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

	if _, err := c.DeployContract1(walletV1.WALLET_CONTRACT_V1); err == nil {
		t.Fatalf("expected error when from address is not set")
	}

	signer := setupSignerWallet(t)
	useWallet(t, c, signer.Wallet)

	if _, err := c.DeployContract1(""); err == nil {
		t.Fatalf("expected error when contract version is empty")
	}

	deployedContract, err := c.DeployContract1(walletV1.WALLET_CONTRACT_V1)
	if err != nil {
		t.Fatalf("DeployContract1: %v", err)
	}

	contractLog, err := utils.UnmarshalLog[log.Log](deployedContract.Logs[0])
	if err != nil {
		t.Fatalf("UnmarshalLog (DeployContract1.Logs[0]): %v", err)
	}

	contractDomain, err := utils.UnmarshalEvent[domain.Contract](contractLog.Event)
	if err != nil {
		t.Fatalf("UnmarshalEvent (DeployContract1.Logs[0]): %v", err)
	}

	assert.NotEmpty(t, contractDomain.Address, "deployed contract address should not be empty")
	assert.Equal(t, walletV1.WALLET_CONTRACT_V1, contractDomain.ContractVersion, "deployed contract version mismatch")

	tmpWM := setupWalletManager(t)
	contractAddress, _ := genKey(t, tmpWM)

	deployed2, err := c.DeployContract2(walletV1.WALLET_CONTRACT_V1, contractAddress)
	if err != nil {
		t.Fatalf("DeployContract2: %v", err)
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
	assert.Equal(t, walletV1.WALLET_CONTRACT_V1, contractDomain2.ContractVersion, "deployed contract version mismatch")
	assert.Equal(t, contractAddress, contractDomain2.Address, "deployed contract address mismatch")
}

func Test_EndToEnd_MinimalFlow(t *testing.T) {
	signer := setupSignerWallet(t)
	c := setupClient(t, signer.Wallet)

	useWallet(t, c, signer.Wallet)

	w := createWallet(t, c, signer.PublicKey)

	if txs, err := c.ListTransactions(w.PublicKey, "", "", nil, 0, 1, 10, true); err == nil && len(txs) > 0 {
		_, _ = c.ListLogs([]string{"wallet_created"}, 0, txs[0].Hash, nil, "", 1, 10, true)
	}

	_, _ = c.ListBlocks(0, time.Time{}, "", "", "", 1, 5, true)
}

func Test_TransactionRoundtrip_Sanity(t *testing.T) {
	signer := setupSignerWallet(t)

	chainId := uint8(1)
	version := uint8(1)

	uuid7, err := utils.NewUUID7()
	if err != nil {
		t.Fatalf("NewUUID7: %v", err)
	}

	tx := transaction.NewTransaction(
		chainId,
		signer.PublicKey,
		signer.PublicKey,
		"echo",
		json.RawMessage(`{"contract_version": "walletV1", "k": "v"}`),
		version,
		uuid7,
	)

	_ = tx.Get()
}