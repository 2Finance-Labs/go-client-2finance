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
	"net/http"
	"os"
	"time"

	client2f "github.com/2Finance-Labs/2finance-sdk-client/client_2finance"
	"github.com/2Finance-Labs/2finance-sdk-client/wallet_manager"
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

	config.Load_config(env, "./../../../.env")

	baseURL := strings.TrimSpace(os.Getenv("TWO_FINANCE_NETWORK_URL"))
	if baseURL == "" {
		baseURL = "http://127.0.0.1:19295"
	}
	c := client2f.NewV2(baseURL, http.DefaultClient, wallet)
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
	pub, priv, err := wallet_manager.GenerateEd25519KeyPairHex()
	if err != nil {
		t.Fatalf("GenerateEd25519KeyPairHex: %v", err)
	}

	return pub, priv
}

func useWallet(t *testing.T, c client2f.Client2FinanceNetwork, wm wallet_manager.IWalletManager) {
	t.Helper()

	c.SetWalletManager(wm)
}

func importAndUnlockWallet(t *testing.T, wm wallet_manager.IWalletManager, expectedPublicKey string, privateKey string) {
	t.Helper()

	privateKeyBytes := []byte(privateKey)

	if err := wm.ImportPrivateKey(privateKeyBytes, E2E_WALLET_PASSWORD); err != nil {
		t.Fatalf("ImportPrivateKey: %v", err)
	}

	gotPublicKey := wm.OwnerAddress()
	if gotPublicKey != expectedPublicKey {
		t.Fatalf("wallet public key mismatch: want %s, got %s", expectedPublicKey, gotPublicKey)
	}

	if err := wm.UnlockWithPassword(E2E_WALLET_PASSWORD); err != nil {
		t.Fatalf("UnlockWithPassword: %v", err)
	}
}
