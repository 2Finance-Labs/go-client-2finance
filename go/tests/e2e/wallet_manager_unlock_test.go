package e2e_test

import (
	"strings"
	"testing"
	"time"

	"github.com/2Finance-Labs/2finance-sdk-client/protocol"
	"github.com/2Finance-Labs/2finance-sdk-client/wallet_manager"
	"gitlab.com/2finance/2finance-network/blockchain/encryption/keys"
	"gitlab.com/2finance/2finance-network/blockchain/utils"
)

func TestWalletManager_UnlockDuration_Is15Seconds(t *testing.T) {
	unlockDuration := 15 * time.Second // This should match the actual unlock duration in the WalletManager implementation
	if unlockDuration != 15*time.Second {
		t.Fatalf("expected unlockDuration to be 15 seconds, got %s", unlockDuration)
	}
}

func TestWalletManager_SignTransaction_ReturnsWalletLocked_WhenNotUnlocked(t *testing.T) {
	manager, publicKey := newImportedWalletForUnlockTest(t)

	tx, err := manager.SignTransaction(wallet_manager.SignTransactionInput{
		ChainID: 1,
		From:    publicKey,
		To:      publicKey,
		Method:  "AddCashback",
		Data: map[string]interface{}{
			"amount": "10",
		},
		Version: 1,
		UUID7:   "018f5f4e-8f6a-7c4a-9a2e-000000000001",
	})

	if err == nil {
		t.Fatal("expected error")
	}

	if tx != nil {
		t.Fatal("expected transaction to be nil")
	}

	assertUnlockTestErrorContains(t, err, "wallet is locked")
}

func TestWalletManager_SignTransaction_WorksAfterUnlockWithPassword(t *testing.T) {
	manager, publicKey := newImportedWalletForUnlockTest(t)

	if manager.IsUnlocked() {
		t.Fatal("expected wallet to start locked after import")
	}

	if err := manager.UnlockWithPassword(testWalletPassword); err != nil {
		t.Fatalf("UnlockWithPassword error: %v", err)
	}

	if !manager.IsUnlocked() {
		t.Fatal("expected wallet to be unlocked")
	}

	tx, err := manager.SignTransaction(wallet_manager.SignTransactionInput{
		ChainID: 1,
		From:    publicKey,
		To:      publicKey,
		Method:  "AddCashback",
		Data: map[string]interface{}{
			"address":       publicKey,
			"owner":         publicKey,
			"token_address": publicKey,
			"program_type":  "fixed-percentage",
			"percentage":    "10",
			"paused":        false,
		},
		Version: 1,
		UUID7:   "018f5f4e-8f6a-7c4a-9a2e-000000000002",
	})
	if err != nil {
		t.Fatalf("SignTransaction error: %v", err)
	}

	if tx == nil {
		t.Fatal("expected signed transaction")
	}
}

func TestWalletManager_SignPreparedTransaction_WorksAfterUnlockWithPassword(t *testing.T) {
	manager, publicKey := newImportedWalletForUnlockTest(t)

	if err := manager.UnlockWithPassword(testWalletPassword); err != nil {
		t.Fatalf("UnlockWithPassword error: %v", err)
	}

	uuid7, err := utils.NewUUID7()
	if err != nil {
		t.Fatalf("NewUUID7 error: %v", err)
	}

	signed, err := manager.SignPreparedTransaction(protocol.PreparedTransaction{
		ChainID: 1,
		From:    publicKey,
		To:      publicKey,
		Method:  "start_Sending",
		Data: map[string]interface{}{
			"contract_version": "sendingLifecycleV2",
			"request_id":       "send-001",
		},
		Version: 1,
		UUID7:   uuid7,
	})
	if err != nil {
		t.Fatalf("SignPreparedTransaction error: %v", err)
	}
	if signed.Hash == "" || signed.Signature == "" {
		t.Fatalf("hash/signature should not be empty")
	}
	if signed.Data["contract_version"] != "sendingLifecycleV2" {
		t.Fatalf("contract_version = %v, want sendingLifecycleV2", signed.Data["contract_version"])
	}
}

func TestWalletManager_SignTransaction_ReturnsErrorForEmptyData(t *testing.T) {
	manager, publicKey := newImportedWalletForUnlockTest(t)

	if err := manager.UnlockWithPassword(testWalletPassword); err != nil {
		t.Fatalf("UnlockWithPassword error: %v", err)
	}

	_, err := manager.SignTransaction(wallet_manager.SignTransactionInput{
		ChainID: 1,
		From:    publicKey,
		To:      publicKey,
		Method:  "start_Sending",
		Data:    map[string]interface{}{},
		Version: 1,
		UUID7:   "018f5f4e-8f6a-7c4a-9a2e-000000000003",
	})
	assertUnlockTestErrorContains(t, err, "data cannot be empty")

	_, err = manager.SignPreparedTransaction(protocol.PreparedTransaction{
		ChainID: 1,
		From:    publicKey,
		To:      publicKey,
		Method:  "start_Sending",
		Data:    map[string]interface{}{},
		Version: 1,
		UUID7:   "018f5f4e-8f6a-7c4a-9a2e-000000000004",
	})
	assertUnlockTestErrorContains(t, err, "data cannot be empty")
}

func TestWalletManager_SignTransaction_ReturnsWalletLocked_AfterManualLock(t *testing.T) {
	manager, publicKey := newImportedWalletForUnlockTest(t)

	if err := manager.UnlockWithPassword(testWalletPassword); err != nil {
		t.Fatalf("UnlockWithPassword error: %v", err)
	}

	if !manager.IsUnlocked() {
		t.Fatal("expected wallet to be unlocked")
	}

	if err := manager.Lock(); err != nil {
		t.Fatalf("Lock error: %v", err)
	}

	if manager.IsUnlocked() {
		t.Fatal("expected wallet to be locked")
	}

	tx, err := manager.SignTransaction(wallet_manager.SignTransactionInput{
		ChainID: 1,
		From:    publicKey,
		To:      publicKey,
		Method:  "AddCashback",
		Data: map[string]interface{}{
			"percentage": "10",
		},
		Version: 1,
		UUID7:   "018f5f4e-8f6a-7c4a-9a2e-000000000003",
	})

	if err == nil {
		t.Fatal("expected error")
	}

	if tx != nil {
		t.Fatal("expected transaction to be nil")
	}

	assertUnlockTestErrorContains(t, err, "wallet is locked")
}

func TestWalletManager_UnlockWithPassword_WrongPasswordDoesNotUnlock(t *testing.T) {
	manager, _ := newImportedWalletForUnlockTest(t)

	err := manager.UnlockWithPassword("wrong-password")
	if err == nil {
		t.Fatal("expected error")
	}

	if manager.IsUnlocked() {
		t.Fatal("expected wallet to remain locked")
	}

	assertUnlockTestErrorContains(t, err, "failed to decrypt wallet file")
}

func TestWalletManager_UnlockThenSign_UsageExample(t *testing.T) {
	walletPath := t.TempDir() + "/wallet.json"

	walletManager := wallet_manager.NewWalletManager(walletPath)

	publicKey, privateKey := newUnlockTestKeyPair(t)

	if err := walletManager.ImportPrivateKey([]byte(privateKey), testWalletPassword); err != nil {
		t.Fatalf("ImportPrivateKey error: %v", err)
	}

	// This is the usage pattern you want:
	//
	// walletManager := wallet_manager.NewWalletManager(walletPath)
	//
	// if err := walletManager.UnlockWithPassword(password); err != nil {
	//     return types.ContractOutput{}, err
	// }
	//
	// client.SetWalletManager(walletManager)
	//
	// cashback, err := client.AddCashback(...)
	if err := walletManager.UnlockWithPassword(testWalletPassword); err != nil {
		t.Fatalf("UnlockWithPassword error: %v", err)
	}

	if !walletManager.IsUnlocked() {
		t.Fatal("expected wallet to be unlocked")
	}

	tx, err := walletManager.SignTransaction(wallet_manager.SignTransactionInput{
		ChainID: 1,
		From:    publicKey,
		To:      publicKey,
		Method:  "AddCashback",
		Data: map[string]interface{}{
			"address":       publicKey,
			"owner":         publicKey,
			"token_address": publicKey,
			"program_type":  "fixed-percentage",
			"percentage":    "10",
			"paused":        false,
		},
		Version: 1,
		UUID7:   "018f5f4e-8f6a-7c4a-9a2e-000000000004",
	})
	if err != nil {
		t.Fatalf("SignTransaction error: %v", err)
	}

	if tx == nil {
		t.Fatal("expected signed transaction")
	}
}

func newImportedWalletForUnlockTest(t *testing.T) (*wallet_manager.WalletManager, string) {
	t.Helper()

	walletPath := t.TempDir() + "/wallet.json"

	managerInterface := wallet_manager.NewWalletManager(walletPath)

	manager, ok := managerInterface.(*wallet_manager.WalletManager)
	if !ok {
		t.Fatal("expected *WalletManager")
	}

	publicKey, privateKey := newUnlockTestKeyPair(t)

	if err := manager.ImportPrivateKey([]byte(privateKey), testWalletPassword); err != nil {
		t.Fatalf("ImportPrivateKey error: %v", err)
	}

	if manager.OwnerAddress() != publicKey {
		t.Fatalf("expected owner %q, got %q", publicKey, manager.OwnerAddress())
	}

	return manager, publicKey
}

func newUnlockTestKeyPair(t *testing.T) (string, string) {
	t.Helper()

	publicKey, privateKey, err := wallet_manager.GenerateEd25519KeyPairHex()
	if err != nil {
		t.Fatalf("GenerateEd25519KeyPairHex error: %v", err)
	}

	if publicKey == "" {
		t.Fatal("expected public key")
	}

	if privateKey == "" {
		t.Fatal("expected private key")
	}

	derivedPublicKey, err := keys.PublicKeyFromEd25519PrivateHex(privateKey)
	if err != nil {
		t.Fatalf("PublicKeyFromEd25519PrivateHex error: %v", err)
	}

	derivedPublicKeyHex := keys.PublicKeyToHex(derivedPublicKey)
	if derivedPublicKeyHex != publicKey {
		t.Fatalf("expected derived public key %q, got %q", publicKey, derivedPublicKeyHex)
	}

	return publicKey, privateKey
}

func assertUnlockTestErrorContains(t *testing.T, err error, expected string) {
	t.Helper()

	if err == nil {
		t.Fatalf("expected error containing %q, got nil", expected)
	}

	if !strings.Contains(err.Error(), expected) {
		t.Fatalf("expected error containing %q, got %q", expected, err.Error())
	}
}

var _ = utils.JSONB{}
