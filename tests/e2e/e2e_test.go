package e2e_test

import (
	"path/filepath"
	"strings"
	"testing"

	"context"
	"crypto/rand"
	"encoding/hex"

	//"encoding/json"
	"fmt"
	"math/big"
	"os"
	"time"

	client2f "github.com/2Finance-Labs/go-client-2finance/client_2finance"
	"github.com/2Finance-Labs/go-client-2finance/wallet_manager"
	"gitlab.com/2finance/2finance-network/config"
	// "gitlab.com/2finance/2finance-network/blockchain/log"
)

const E2E_WALLET_PASSWORD = "E2E-Wallet-Password-123!"

// ----------------------------------------------------------------------------
// Helpers
// ----------------------------------------------------------------------------
type e2eSigner struct {
	PublicKey  string
	PrivateKey string
	Wallet     wallet_manager.IWalletManager
}

func setupSignerWallet(t *testing.T) e2eSigner {
	t.Helper()

	wm := setupWalletManager(t)

	pub, priv := genKey(t, wm)

	importAndUnlockWallet(t, wm, pub, priv)

	return e2eSigner{
		PublicKey:  pub,
		PrivateKey: priv,
		Wallet:     wm,
	}
}

func setupClient(t *testing.T, wallet wallet_manager.IWalletManager) client2f.Client2FinanceNetwork {
	t.Helper()

	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "prod"
	}

	config.Load_config(env, "./../../.env")

	emqxHost := fmt.Sprintf(
		"%s://%s:%s",
		config.EMQX_SCHEME,
		config.EMQX_HOST,
		config.EMQX_PORT,
	)

	base := config.EMQX_CLIENT_ID
	if base == "" {
		base = "e2e"
	}

	id := fmt.Sprintf("%s-%s", base, randSuffix(8))

	c := client2f.New(emqxHost, id, false, wallet)
	c.SetChainID(config.CHAIN_ID)

	return c
}

func sanitizeFileName(name string) string {
	replacer := strings.NewReplacer(
		"/", "_",
		"\\", "_",
		" ", "_",
		":", "_",
	)

	return replacer.Replace(name)
}

const E2E_WALLET_DIR = "wallets"

func setupWalletManager(t *testing.T) wallet_manager.IWalletManager {
	t.Helper()

	walletDir := os.Getenv("E2E_WALLET_DIR")
	if walletDir == "" {
		walletDir = E2E_WALLET_DIR
	}

	if err := os.MkdirAll(walletDir, 0700); err != nil {
		t.Fatalf("MkdirAll wallet dir: %v", err)
	}

	walletFileName := fmt.Sprintf(
		"%s-%d-%s.wallet",
		sanitizeFileName(t.Name()),
		time.Now().UnixNano(),
		randSuffix(8),
	)

	walletPath := filepath.Join(walletDir, walletFileName)

	return wallet_manager.NewWalletManager(walletPath)
}

func randSuffix(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)[:n]
}

// amt builds integer string respecting decimals (unscaled * 10^decimals)
func amt(unscaled int64, decimals int) string {
	p := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimals)), nil)
	v := new(big.Int).Mul(big.NewInt(unscaled), p)
	return v.String()
}

func waitUntil(t *testing.T, d time.Duration, pred func() bool) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), d)
	defer cancel()
	tick := time.NewTicker(20 * time.Millisecond)
	defer tick.Stop()
	for {
		select {
		case <-ctx.Done():
			t.Fatalf("timeout waiting for condition")
		case <-tick.C:
			if pred() {
				return
			}
		}
	}
}

func genKey(t *testing.T, w wallet_manager.IWalletManager) (pub, priv string) {
	pub, priv, err := w.GenerateEd25519KeyPairHex()
	if err != nil {
		t.Fatalf("GenerateEd25519KeyPairHex: %v", err)
	}

	return pub, priv
}

func importAndUnlockWallet(t *testing.T, wm wallet_manager.IWalletManager, expectedPublicKey string, privateKey string) {
	t.Helper()

	privateKeyBytes := []byte(privateKey)

	if err := wm.ImportWallet(privateKeyBytes, E2E_WALLET_PASSWORD); err != nil {
		t.Fatalf("ImportWallet: %v", err)
	}

	gotPublicKey := wm.GetPublicKey()
	if gotPublicKey != expectedPublicKey {
		t.Fatalf("wallet public key mismatch: want %s, got %s", expectedPublicKey, gotPublicKey)
	}

	if err := wm.Unlock(E2E_WALLET_PASSWORD); err != nil {
		t.Fatalf("Unlock: %v", err)
	}
}

func useWallet(t *testing.T, c client2f.Client2FinanceNetwork, wm wallet_manager.IWalletManager) {
	t.Helper()

	c.SetWalletManager(wm)
}
