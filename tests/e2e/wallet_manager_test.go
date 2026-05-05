package e2e_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/2Finance-Labs/go-client-2finance/wallet_manager"
	"github.com/stretchr/testify/require"
)

func TestWalletManagerE2E_LockUnlockRealFlow(t *testing.T) {
	// -------------------------
	// ARRANGE
	// -------------------------
	owner := "owner-address-test"
	password := "StrongPassword123!"
	wrongPassword := "WrongPassword123!"

	originalPrivateKey := []byte("test-private-key-value")

	walletDir := t.TempDir()
	walletPath := filepath.Join(walletDir, "owner-address-test.wallet")

	manager := wallet_manager.NewWalletManager(owner, walletPath)

	// -------------------------
	// ACT: CREATE LOCAL WALLET
	// -------------------------
	privateKeyToCreateWallet := cloneBytes(originalPrivateKey)

	err := manager.CreateLocalWallet(privateKeyToCreateWallet, password)

	// -------------------------
	// ASSERT: CREATE LOCAL WALLET
	// -------------------------
	require.NoError(t, err)

	_, err = os.Stat(walletPath)
	require.NoError(t, err, "wallet file should be created locally")

	require.False(t, manager.IsUnlocked(), "wallet should be locked after CreateLocalWallet()")

	require.NotEqual(
		t,
		originalPrivateKey,
		privateKeyToCreateWallet,
		"CreateLocalWallet() should clear the input private key slice from memory",
	)

	for _, b := range privateKeyToCreateWallet {
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

	// -------------------------
	// ASSERT: GET PRIVATE KEY WITHOUT PASSWORD
	// SignTransaction is not in passwordRequiredMethods,
	// so it can use the 2-minute unlocked session.
	// -------------------------
	unlockedPrivateKey, err := manager.GetPrivateKey("SignTransaction", "")

	require.NoError(t, err)
	require.Equal(t, originalPrivateKey, unlockedPrivateKey)

	// -------------------------
	// ASSERT: returned private key is a clone
	// -------------------------
	unlockedPrivateKey[0] = 'X'

	unlockedPrivateKeyAgain, err := manager.GetPrivateKey("SignTransaction", "")

	require.NoError(t, err)
	require.Equal(t, originalPrivateKey, unlockedPrivateKeyAgain)

	// -------------------------
	// ASSERT: sensitive method requires password
	// ExportPrivateKey is in passwordRequiredMethods.
	// -------------------------
	_, err = manager.GetPrivateKey("ExportPrivateKey", "")
	require.EqualError(t, err, "password is required")

	exportedPrivateKey, err := manager.GetPrivateKey("ExportPrivateKey", password)

	require.NoError(t, err)
	require.Equal(t, originalPrivateKey, exportedPrivateKey)

	// -------------------------
	// ACT: LOCK
	// -------------------------
	err = manager.Lock()

	// -------------------------
	// ASSERT: LOCK
	// -------------------------
	require.NoError(t, err)
	require.False(t, manager.IsUnlocked())

	_, err = manager.GetPrivateKey("SignTransaction", "")
	require.EqualError(t, err, "wallet is locked")

	// -------------------------
	// ASSERT: after locked, password can unlock again
	// -------------------------
	privateKeyAfterRelock, err := manager.GetPrivateKey("SignTransaction", password)

	require.NoError(t, err)
	require.Equal(t, originalPrivateKey, privateKeyAfterRelock)
}

func TestWalletManagerE2E_UnlockAfterNewManagerInstance(t *testing.T) {
	// Esse teste simula o app fechando e abrindo de novo.
	// O primeiro manager cria o arquivo.
	// O segundo manager lê o arquivo e desbloqueia a wallet.

	// -------------------------
	// ARRANGE
	// -------------------------
	owner := "owner-address-test"
	password := "StrongPassword123!"
	originalPrivateKey := []byte("test-private-key-value")

	walletDir := t.TempDir()
	walletPath := filepath.Join(walletDir, "owner-address-test.wallet")

	firstManager := wallet_manager.NewWalletManager(owner, walletPath)

	// -------------------------
	// ACT: FIRST INSTANCE CREATES LOCAL WALLET
	// -------------------------
	privateKeyToCreateWallet := cloneBytes(originalPrivateKey)

	err := firstManager.CreateLocalWallet(privateKeyToCreateWallet, password)
	require.NoError(t, err)

	require.False(t, firstManager.IsUnlocked())

	// -------------------------
	// ACT: SECOND INSTANCE UNLOCKS WALLET
	// -------------------------
	secondManager := wallet_manager.NewWalletManager(owner, walletPath)

	err = secondManager.Unlock(password)

	// -------------------------
	// ASSERT
	// -------------------------
	require.NoError(t, err)
	require.True(t, secondManager.IsUnlocked())

	privateKeyFromSecondManager, err := secondManager.GetPrivateKey("SignTransaction", "")

	require.NoError(t, err)
	require.Equal(t, originalPrivateKey, privateKeyFromSecondManager)
}

func TestWalletManagerE2E_OwnerMismatch(t *testing.T) {
	// Esse teste garante que uma wallet criada para um owner
	// não seja aberta por outro owner.

	// -------------------------
	// ARRANGE
	// -------------------------
	owner := "owner-address-test"
	anotherOwner := "another-owner-address-test"
	password := "StrongPassword123!"
	privateKey := []byte("test-private-key-value")

	walletDir := t.TempDir()
	walletPath := filepath.Join(walletDir, "owner-address-test.wallet")

	manager := wallet_manager.NewWalletManager(owner, walletPath)

	err := manager.CreateLocalWallet(privateKey, password)
	require.NoError(t, err)

	anotherManager := wallet_manager.NewWalletManager(anotherOwner, walletPath)

	// -------------------------
	// ACT
	// -------------------------
	err = anotherManager.Unlock(password)

	// -------------------------
	// ASSERT
	// -------------------------
	require.Error(t, err)
	require.Contains(t, err.Error(), "wallet owner mismatch")
	require.False(t, anotherManager.IsUnlocked())
}

func TestWalletManagerE2E_InvalidInputs(t *testing.T) {
	walletDir := t.TempDir()
	walletPath := filepath.Join(walletDir, "owner-address-test.wallet")

	manager := wallet_manager.NewWalletManager("owner-address-test", walletPath)

	err := manager.CreateLocalWallet(nil, "StrongPassword123!")
	require.Error(t, err)
	require.Contains(t, err.Error(), "private key is required")

	err = manager.CreateLocalWallet([]byte("private-key"), "")
	require.EqualError(t, err, "password is required")

	err = manager.Unlock("")
	require.EqualError(t, err, "password is required")

	_, err = manager.GetPrivateKey("SignTransaction", "")
	require.ErrorContains(t, err, "wallet is locked")
}