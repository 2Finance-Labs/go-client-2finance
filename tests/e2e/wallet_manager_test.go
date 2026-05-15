package e2e_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/2Finance-Labs/go-client-2finance/wallet_manager"
	"github.com/stretchr/testify/require"
	"gitlab.com/2finance/2finance-network/blockchain/contract/cashbackV1"
	"gitlab.com/2finance/2finance-network/blockchain/contract/walletV1"
	"gitlab.com/2finance/2finance-network/blockchain/utils"
)

func TestWalletManagerE2E_LockUnlockRealFlow(t *testing.T) {
	// -------------------------
	// ARRANGE
	// -------------------------
	password := "StrongPassword123!"
	wrongPassword := "WrongPassword123!"

	walletDir := t.TempDir()
	walletPath := filepath.Join(walletDir, "owner-address-test.wallet")

	manager := wallet_manager.NewWalletManager(walletPath)

	originalPublicKey, originalPrivateKey, err := manager.GenerateEd25519KeyPairHex()
	require.NoError(t, err)

	originalPrivateKeyBytes := []byte(originalPrivateKey)

	// -------------------------
	// ACT: IMPORT WALLET
	// -------------------------
	privateKeyToImport := cloneBytes(originalPrivateKeyBytes)

	err = manager.ImportWallet(privateKeyToImport, password)

	// -------------------------
	// ASSERT: IMPORT WALLET
	// -------------------------
	require.NoError(t, err)

	_, err = os.Stat(walletPath)
	require.NoError(t, err, "wallet file should be created locally")

	require.False(t, manager.IsUnlocked(), "wallet should be locked after ImportWallet()")
	require.Equal(t, originalPublicKey, manager.GetPublicKey(), "wallet public key should be derived from imported private key")

	require.NotEqual(
		t,
		originalPrivateKeyBytes,
		privateKeyToImport,
		"ImportWallet() should clear the input private key slice from memory",
	)

	for _, b := range privateKeyToImport {
		require.Equal(t, byte(0), b, "private key input slice should be zeroed")
	}

	// -------------------------
	// ASSERT: WRONG PASSWORD
	// -------------------------
	err = manager.Unlock(wrongPassword)
	require.Error(t, err, "unlocking with wrong password should fail")
	require.False(t, manager.IsUnlocked(), "wallet should remain locked after wrong password")

	// -------------------------
	// ACT: UNLOCK
	// -------------------------
	err = manager.Unlock(password)

	// -------------------------
	// ASSERT: UNLOCK
	// -------------------------
	require.NoError(t, err)
	require.True(t, manager.IsUnlocked(), "wallet should be unlocked after correct password")
	require.Equal(t, originalPublicKey, manager.GetPublicKey(), "wallet public key should remain loaded after unlock")

	// -------------------------
	// ASSERT: SIGN WITHOUT PASSWORD WHILE UNLOCKED
	// -------------------------
	nonSensitiveMethod := walletV1.METHOD_GET_WALLET_BY_ADDRESS

	data, err := utils.MapToJSONB(map[string]interface{}{
		"address": originalPublicKey,
	})
	require.NoError(t, err)

	uuid7, err := utils.NewUUID7()
	require.NoError(t, err)

	signedTx, err := manager.SignTransaction(
		1,
		originalPublicKey,
		originalPublicKey,
		nonSensitiveMethod,
		data,
		1,
		uuid7,
	)
	require.NoError(t, err)
	require.NotNil(t, signedTx)
	require.NotEmpty(t, signedTx.Signature)
	require.NotEmpty(t, signedTx.Hash)

	// -------------------------
	// ASSERT: SENSITIVE METHOD CONFIGURATION
	// -------------------------
	sensitiveMethod := cashbackV1.METHOD_WITHDRAW_CASHBACK

	require.False(t, manager.PasswordIsRequired(sensitiveMethod))

	err = manager.AddRequiredPasswordMethods(sensitiveMethod)
	require.NoError(t, err)

	require.True(t, manager.PasswordIsRequired(sensitiveMethod))

	// -------------------------
	// ASSERT: SIGN WITH PASSWORD REQUIRES PASSWORD
	// -------------------------
	uuid7, err = utils.NewUUID7()
	require.NoError(t, err)

	_, err = manager.SignTransactionWithPassword(
		1,
		originalPublicKey,
		originalPublicKey,
		sensitiveMethod,
		data,
		1,
		uuid7,
		"",
	)
	require.EqualError(t, err, "password is required")

	uuid7, err = utils.NewUUID7()
	require.NoError(t, err)

	signedTx, err = manager.SignTransactionWithPassword(
		1,
		originalPublicKey,
		originalPublicKey,
		sensitiveMethod,
		data,
		1,
		uuid7,
		password,
	)
	require.NoError(t, err)
	require.NotNil(t, signedTx)
	require.NotEmpty(t, signedTx.Signature)
	require.NotEmpty(t, signedTx.Hash)

	// -------------------------
	// ACT: LOCK
	// -------------------------
	err = manager.Lock()

	// -------------------------
	// ASSERT: LOCK
	// -------------------------
	require.NoError(t, err)
	require.False(t, manager.IsUnlocked())

	uuid7, err = utils.NewUUID7()
	require.NoError(t, err)

	_, err = manager.SignTransaction(
		1,
		originalPublicKey,
		originalPublicKey,
		nonSensitiveMethod,
		data,
		1,
		uuid7,
	)
	require.Error(t, err)

	// -------------------------
	// ASSERT: UNLOCK AGAIN
	// -------------------------
	err = manager.Unlock(password)
	require.NoError(t, err)
	require.True(t, manager.IsUnlocked())

	uuid7, err = utils.NewUUID7()
	require.NoError(t, err)

	signedTx, err = manager.SignTransaction(
		1,
		originalPublicKey,
		originalPublicKey,
		nonSensitiveMethod,
		data,
		1,
		uuid7,
	)
	require.NoError(t, err)
	require.NotNil(t, signedTx)
	require.NotEmpty(t, signedTx.Signature)
	require.NotEmpty(t, signedTx.Hash)
}

func TestWalletManagerE2E_UnlockAfterNewManagerInstance(t *testing.T) {
	// Esse teste simula o app fechando e abrindo de novo.
	// O primeiro manager cria o arquivo.
	// O segundo manager lê o arquivo e desbloqueia a wallet.

	// -------------------------
	// ARRANGE
	// -------------------------
	password := "StrongPassword123!"

	walletDir := t.TempDir()
	walletPath := filepath.Join(walletDir, "owner-address-test.wallet")

	firstManager := wallet_manager.NewWalletManager(walletPath)

	originalPublicKey, originalPrivateKey, err := firstManager.GenerateEd25519KeyPairHex()
	require.NoError(t, err)

	originalPrivateKeyBytes := []byte(originalPrivateKey)

	// -------------------------
	// ACT: FIRST INSTANCE IMPORTS WALLET
	// -------------------------
	privateKeyToImport := cloneBytes(originalPrivateKeyBytes)

	err = firstManager.ImportWallet(privateKeyToImport, password)
	require.NoError(t, err)

	require.False(t, firstManager.IsUnlocked())
	require.Equal(t, originalPublicKey, firstManager.GetPublicKey())

	// -------------------------
	// ACT: SECOND INSTANCE UNLOCKS WALLET
	// -------------------------
	secondManager := wallet_manager.NewWalletManager(walletPath)

	err = secondManager.Unlock(password)

	// -------------------------
	// ASSERT
	// -------------------------
	require.NoError(t, err)
	require.True(t, secondManager.IsUnlocked())
	require.Equal(t, originalPublicKey, secondManager.GetPublicKey())

	data, err := utils.MapToJSONB(map[string]interface{}{
		"address": originalPublicKey,
	})
	require.NoError(t, err)

	uuid7, err := utils.NewUUID7()
	require.NoError(t, err)

	signedTx, err := secondManager.SignTransaction(
		1,
		originalPublicKey,
		originalPublicKey,
		walletV1.METHOD_GET_WALLET_BY_ADDRESS,
		data,
		1,
		uuid7,
	)

	require.NoError(t, err)
	require.NotNil(t, signedTx)
	require.NotEmpty(t, signedTx.Signature)
	require.NotEmpty(t, signedTx.Hash)
}

func TestWalletManagerE2E_WrongWalletFileDoesNotMatchExpectedPublicKey(t *testing.T) {
	// Esse teste substitui o antigo OwnerMismatch.
	// No fluxo novo, o owner/publicKey é derivado da private key importada.
	// Então não faz mais sentido passar "anotherOwner" no construtor.
	// A validação correta é garantir que uma instância nova carregue o publicKey real do arquivo.

	// -------------------------
	// ARRANGE
	// -------------------------
	password := "StrongPassword123!"

	walletDir := t.TempDir()
	walletPath := filepath.Join(walletDir, "owner-address-test.wallet")

	manager := wallet_manager.NewWalletManager(walletPath)

	originalPublicKey, originalPrivateKey, err := manager.GenerateEd25519KeyPairHex()
	require.NoError(t, err)

	privateKeyToImport := []byte(originalPrivateKey)

	err = manager.ImportWallet(privateKeyToImport, password)
	require.NoError(t, err)

	anotherManager := wallet_manager.NewWalletManager(walletPath)

	// -------------------------
	// ACT
	// -------------------------
	err = anotherManager.Unlock(password)

	// -------------------------
	// ASSERT
	// -------------------------
	require.NoError(t, err)
	require.True(t, anotherManager.IsUnlocked())
	require.Equal(t, originalPublicKey, anotherManager.GetPublicKey())
}

func TestWalletManagerE2E_InvalidInputs(t *testing.T) {
	walletDir := t.TempDir()
	walletPath := filepath.Join(walletDir, "owner-address-test.wallet")

	manager := wallet_manager.NewWalletManager(walletPath)

	err := manager.ImportWallet(nil, "StrongPassword123!")
	require.Error(t, err)
	require.Contains(t, err.Error(), "private key is required")

	err = manager.ImportWallet([]byte("private-key"), "")
	require.EqualError(t, err, "password is required")

	err = manager.Unlock("")
	require.EqualError(t, err, "password is required")

	publicKey, _, err := manager.GenerateEd25519KeyPairHex()
	require.NoError(t, err)

	data, err := utils.MapToJSONB(map[string]interface{}{
		"address": publicKey,
	})
	require.NoError(t, err)

	uuid7, err := utils.NewUUID7()
	require.NoError(t, err)

	_, err = manager.SignTransaction(
		1,
		publicKey,
		publicKey,
		walletV1.METHOD_GET_WALLET_BY_ADDRESS,
		data,
		1,
		uuid7,
	)
	require.Error(t, err)

	err = manager.AddRequiredPasswordMethods("")
	require.EqualError(t, err, "method name is required")
}

func TestWalletManagerE2E_RotatePassword(t *testing.T) {
	// -------------------------
	// ARRANGE
	// -------------------------
	currentPassword := "StrongPassword123!"
	newPassword := "NewStrongPassword123!"

	walletDir := t.TempDir()
	walletPath := filepath.Join(walletDir, "owner-address-test.wallet")

	manager := wallet_manager.NewWalletManager(walletPath)

	originalPublicKey, originalPrivateKey, err := manager.GenerateEd25519KeyPairHex()
	require.NoError(t, err)

	originalPrivateKeyBytes := []byte(originalPrivateKey)

	privateKeyToImport := cloneBytes(originalPrivateKeyBytes)

	err = manager.ImportWallet(privateKeyToImport, currentPassword)
	require.NoError(t, err)

	_, err = os.Stat(walletPath)
	require.NoError(t, err, "wallet file should be created locally")

	require.Equal(t, originalPublicKey, manager.GetPublicKey())

	// -------------------------
	// ASSERT: CURRENT PASSWORD WORKS BEFORE ROTATION
	// -------------------------
	err = manager.Unlock(currentPassword)
	require.NoError(t, err)
	require.True(t, manager.IsUnlocked())

	data, err := utils.MapToJSONB(map[string]interface{}{
		"address": originalPublicKey,
	})
	require.NoError(t, err)

	uuid7, err := utils.NewUUID7()
	require.NoError(t, err)

	signedTx, err := manager.SignTransaction(
		1,
		originalPublicKey,
		originalPublicKey,
		walletV1.METHOD_GET_WALLET_BY_ADDRESS,
		data,
		1,
		uuid7,
	)
	require.NoError(t, err)
	require.NotNil(t, signedTx)

	err = manager.Lock()
	require.NoError(t, err)
	require.False(t, manager.IsUnlocked())

	// -------------------------
	// ACT: ROTATE PASSWORD
	// -------------------------
	err = manager.RotatePassword(currentPassword, newPassword)
	require.NoError(t, err)

	require.False(t, manager.IsUnlocked(), "wallet should be locked after password rotation")

	// -------------------------
	// ASSERT: OLD PASSWORD DOES NOT WORK
	// -------------------------
	err = manager.Unlock(currentPassword)
	require.Error(t, err, "old password should not unlock wallet after rotation")
	require.False(t, manager.IsUnlocked())

	// -------------------------
	// ASSERT: NEW PASSWORD WORKS
	// -------------------------
	err = manager.Unlock(newPassword)
	require.NoError(t, err)
	require.True(t, manager.IsUnlocked())
	require.Equal(t, originalPublicKey, manager.GetPublicKey())

	uuid7, err = utils.NewUUID7()
	require.NoError(t, err)

	signedTx, err = manager.SignTransaction(
		1,
		originalPublicKey,
		originalPublicKey,
		walletV1.METHOD_GET_WALLET_BY_ADDRESS,
		data,
		1,
		uuid7,
	)
	require.NoError(t, err)
	require.NotNil(t, signedTx)
	require.NotEmpty(t, signedTx.Signature)
	require.NotEmpty(t, signedTx.Hash)
}

func TestWalletManagerE2E_RotatePasswordInvalidInputs(t *testing.T) {
	currentPassword := "StrongPassword123!"
	newPassword := "NewStrongPassword123!"

	walletDir := t.TempDir()
	walletPath := filepath.Join(walletDir, "owner-address-test.wallet")

	manager := wallet_manager.NewWalletManager(walletPath)

	_, originalPrivateKey, err := manager.GenerateEd25519KeyPairHex()
	require.NoError(t, err)

	privateKeyToImport := []byte(originalPrivateKey)

	err = manager.ImportWallet(privateKeyToImport, currentPassword)
	require.NoError(t, err)

	err = manager.RotatePassword("", newPassword)
	require.EqualError(t, err, "current password is required")

	err = manager.RotatePassword(currentPassword, "")
	require.EqualError(t, err, "new password is required")

	err = manager.RotatePassword(currentPassword, currentPassword)
	require.EqualError(t, err, "new password must be different from current password")

	err = manager.RotatePassword("WrongPassword123!", newPassword)
	require.Error(t, err)
	require.ErrorContains(t, err, "failed to decrypt wallet file with current password")
}

func TestWalletManagerE2E_AddRequiredPasswordMethods(t *testing.T) {
	walletDir := t.TempDir()
	walletPath := filepath.Join(walletDir, "owner-address-test.wallet")

	manager := wallet_manager.NewWalletManager(walletPath)

	sensitiveMethod := cashbackV1.METHOD_WITHDRAW_CASHBACK

	require.False(t, manager.PasswordIsRequired(sensitiveMethod))

	err := manager.AddRequiredPasswordMethods(sensitiveMethod)
	require.NoError(t, err)

	require.True(t, manager.PasswordIsRequired(sensitiveMethod))

	err = manager.AddRequiredPasswordMethods("")
	require.EqualError(t, err, "method name is required")
}

func TestWalletManagerE2E_SignTransactionWithPasswordRequiresPassword(t *testing.T) {
	password := "StrongPassword123!"

	walletDir := t.TempDir()
	walletPath := filepath.Join(walletDir, "owner-address-test.wallet")

	manager := wallet_manager.NewWalletManager(walletPath)

	publicKey, privateKey, err := manager.GenerateEd25519KeyPairHex()
	require.NoError(t, err)

	err = manager.ImportWallet([]byte(privateKey), password)
	require.NoError(t, err)

	err = manager.Unlock(password)
	require.NoError(t, err)
	require.True(t, manager.IsUnlocked())

	data, err := utils.MapToJSONB(map[string]interface{}{
		"address":       publicKey,
		"amount":        "100",
		"token_address": publicKey,
		"token_type":    "fungible",
		"uuid":          "",
	})
	require.NoError(t, err)

	uuid7, err := utils.NewUUID7()
	require.NoError(t, err)

	_, err = manager.SignTransactionWithPassword(
		1,
		publicKey,
		publicKey,
		cashbackV1.METHOD_WITHDRAW_CASHBACK,
		data,
		1,
		uuid7,
		"",
	)
	require.EqualError(t, err, "password is required")

	uuid7, err = utils.NewUUID7()
	require.NoError(t, err)

	signedTx, err := manager.SignTransactionWithPassword(
		1,
		publicKey,
		publicKey,
		cashbackV1.METHOD_WITHDRAW_CASHBACK,
		data,
		1,
		uuid7,
		password,
	)

	require.NoError(t, err)
	require.NotNil(t, signedTx)
	require.NotEmpty(t, signedTx.Signature)
	require.NotEmpty(t, signedTx.Hash)
}

func TestWalletManagerE2E_NonSensitiveMethodUsesUnlockedSession(t *testing.T) {
	password := "StrongPassword123!"

	walletDir := t.TempDir()
	walletPath := filepath.Join(walletDir, "owner-address-test.wallet")

	manager := wallet_manager.NewWalletManager(walletPath)

	publicKey, privateKey, err := manager.GenerateEd25519KeyPairHex()
	require.NoError(t, err)

	err = manager.ImportWallet([]byte(privateKey), password)
	require.NoError(t, err)

	err = manager.Unlock(password)
	require.NoError(t, err)
	require.True(t, manager.IsUnlocked())

	nonSensitiveMethod := walletV1.METHOD_GET_WALLET_BY_ADDRESS

	require.False(t, manager.PasswordIsRequired(nonSensitiveMethod))

	data, err := utils.MapToJSONB(map[string]interface{}{
		"address": publicKey,
	})
	require.NoError(t, err)

	uuid7, err := utils.NewUUID7()
	require.NoError(t, err)

	signedTx, err := manager.SignTransaction(
		1,
		publicKey,
		publicKey,
		nonSensitiveMethod,
		data,
		1,
		uuid7,
	)

	require.NoError(t, err)
	require.NotNil(t, signedTx)
	require.NotEmpty(t, signedTx.Signature)
	require.NotEmpty(t, signedTx.Hash)
}