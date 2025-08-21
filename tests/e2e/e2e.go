package e2e_test

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	client2f "github.com/2Finance-Labs/go-client-2finance/client_2finance"
	tokenV1Domain "gitlab.com/2finance/2finance-network/blockchain/contract/tokenV1/domain"
	walletDomain "gitlab.com/2finance/2finance-network/blockchain/contract/walletV1/domain"
	"gitlab.com/2finance/2finance-network/config"
)

// --- Helpers -----------------------------------------------------------------

func setupClient(t *testing.T) client2f.Client2FinanceNetwork {
	t.Helper()
	// Allow overriding ENV (same pattern as your main)
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "prod"
	}
	config.Load_config(env, "./.env")

	emqxHost := fmt.Sprintf("%s://%s:%s", config.EMQX_SCHEME, config.EMQX_HOST, config.EMQX_PORT)
	c := client2f.New(emqxHost, config.EMQX_CLIENT_ID, false)
	return c
}

func randSuffix(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)[:n]
}

// unmarshalState decodes an arbitrary state object into a typed struct.
func unmarshalState[T any](t *testing.T, obj any, out *T) {
	t.Helper()
	by, err := json.Marshal(obj)
	if err != nil {
		t.Fatalf("marshal state: %v", err)
	}
	if err := json.Unmarshal(by, out); err != nil {
		t.Fatalf("unmarshal state: %v", err)
	}
}

// createWallet generates a keypair, registers the wallet and returns the parsed state.
func createWallet(t *testing.T, c client2f.Client2FinanceNetwork) walletDomain.Wallet {
	t.Helper()
	pub, priv, err := c.GenerateKeyEd25519()
	if err != nil {
		t.Fatalf("GenerateKeyEd25519: %v", err)
	}
	c.SetPrivateKey(priv)

	wOut, err := c.AddWallet(pub)
	if err != nil {
		t.Fatalf("AddWallet: %v", err)
	}
	var w walletDomain.Wallet
	unmarshalState(t, wOut.States[0].Object, &w)
	if w.PublicKey == "" {
		t.Fatalf("wallet public key empty")
	}
	return w
}

// createBasicToken creates a minimal token owned by ownerPub.
func createBasicToken(t *testing.T, c client2f.Client2FinanceNetwork, ownerPub string) tokenV1Domain.Token {
	t.Helper()
	symbol := "2F" + randSuffix(4)
	name := "2Finance"
	decimals := 3
	totalSupply := "10"
	description := "e2e token created by tests"
	image := "https://example.com/image.png"
	website := "https://example.com"
	tagsSocial := map[string]string{"twitter": "https://twitter.com/2finance"}
	tagsCat := map[string]string{"category": "DeFi"}
	tags := map[string]string{"tag1": "DeFi", "tag2": "Blockchain"}
	creator := "2Finance Test"
	creatorWebsite := "https://creator.example"
	allowUsers := map[string]bool{}
	blockUsers := map[string]bool{}
	feeTiers := []map[string]interface{}{
		{"min_amount": "0", "max_amount": "1000000000000000000", "min_volume": "0", "max_volume": "10000000000000000000", "fee_bps": 50},
	}
	feeAddress := "fe1b01a9861bb265b141c00517d7697c8a0d8286492a14d776ca33ffdded43c1"
	freezeAuthorityRevoked := false
	mintAuthorityRevoked := false
	updateAuthorityRevoked := false
	paused := false
	expiredAt := time.Time{}

	out, err := c.AddToken(symbol, name, decimals, totalSupply, description, ownerPub, image, website, tagsSocial, tagsCat, tags, creator, creatorWebsite, allowUsers, blockUsers, feeTiers, feeAddress, freezeAuthorityRevoked, mintAuthorityRevoked, updateAuthorityRevoked, paused, expiredAt)
	if err != nil {
		t.Fatalf("AddToken: %v", err)
	}
	var tok tokenV1Domain.Token
	unmarshalState(t, out.States[0].Object, &tok)
	if tok.Address == "" {
		t.Fatalf("token address empty")
	}
	return tok
}

// --- Tests -------------------------------------------------------------------

func TestWallet(t *testing.T) { // testWallet
	c := setupClient(t)
	w := createWallet(t, c)

	// Fetch it back to ensure GetWallet works
	got, err := c.GetWallet(w.PublicKey)
	if err != nil {
		t.Fatalf("GetWallet: %v", err)
	}
	var w2 walletDomain.Wallet
	unmarshalState(t, got.States[0].Object, &w2)
	if w2.PublicKey != w.PublicKey {
		t.Fatalf("wallet mismatch: have %s want %s", w2.PublicKey, w.PublicKey)
	}

	nonce, err := c.GetNonce(w.PublicKey)
	if err != nil {
		t.Fatalf("GetNonce: %v", err)
	}
	if nonce < 0 {
		t.Fatalf("invalid nonce: %d", nonce)
	}
}

func TestToken(t *testing.T) { // testToken
	c := setupClient(t)
	owner := createWallet(t, c)

	// create token
	tok := createBasicToken(t, c, owner.PublicKey)

	// Mint to owner
	mintOut, err := c.MintToken(tok.Address, owner.PublicKey, "35", tok.Decimals)
	if err != nil {
		t.Fatalf("MintToken: %v", err)
	}
	var mint tokenV1Domain.Mint
	unmarshalState(t, mintOut.States[0].Object, &mint)
	if mint.TokenAddress != tok.Address {
		t.Fatalf("mint token mismatch: %s != %s", mint.TokenAddress, tok.Address)
	}

	// Burn from owner
	burnOut, err := c.BurnToken(tok.Address, "12", tok.Decimals)
	if err != nil {
		t.Fatalf("BurnToken: %v", err)
	}
	var burn tokenV1Domain.Burn
	unmarshalState(t, burnOut.States[0].Object, &burn)
	if burn.TokenAddress != tok.Address {
		t.Fatalf("burn token mismatch: %s != %s", burn.TokenAddress, tok.Address)
	}

	// Transfer to a new wallet allowed by AllowUsers
	to := createWallet(t, c)
	_, _, err = c.GenerateKeyEd25519() // just exercising API; not strictly needed
	if err != nil {
		t.Fatalf("GenerateKeyEd25519: %v", err)
	}
	if _, err := c.AllowUsers(tok.Address, map[string]bool{to.PublicKey: true}); err != nil {
		t.Fatalf("AllowUsers: %v", err)
	}

	trOut, err := c.TransferToken(tok.Address, to.PublicKey, "1", tok.Decimals)
	if err != nil {
		t.Fatalf("TransferToken: %v", err)
	}
	var tr tokenV1Domain.Transfer
	unmarshalState(t, trOut.States[0].Object, &tr)
	if tr.ToAddress != to.PublicKey {
		t.Fatalf("transfer to mismatch: %s != %s", tr.ToAddress, to.PublicKey)
	}

	// Check balances
	if _, err := c.GetTokenBalance(tok.Address, owner.PublicKey); err != nil {
		t.Fatalf("GetTokenBalance(owner): %v", err)
	}
	if _, err := c.GetTokenBalance(tok.Address, to.PublicKey); err != nil {
		t.Fatalf("GetTokenBalance(to): %v", err)
	}

	// Pause / Unpause
	if _, err := c.PauseToken(tok.Address, true); err != nil {
		t.Fatalf("PauseToken: %v", err)
	}
	if _, err := c.UnpauseToken(tok.Address, false); err != nil {
		t.Fatalf("UnpauseToken: %v", err)
	}

	// Update metadata
	newSymbol := "2F-NEW" + randSuffix(4)
	_, err = c.UpdateMetadata(
		tok.Address,
		newSymbol,
		"2Finance New",
		tok.Decimals+1,
		"Updated by tests",
		"https://example.com/image-new.png",
		"https://example-new.com",
		map[string]string{"twitter": "https://twitter.com/2finance-new"},
		map[string]string{"category": "DeFi New"},
		map[string]string{"tag1": "DeFi New"},
		"2Finance Creator New",
		"https://creator-new.com",
		time.Now().Add(30*24*time.Hour),
	)
	if err != nil {
		t.Fatalf("UpdateMetadata: %v", err)
	}
}

func TestFaucet(t *testing.T) { // testFaucet
	// If the Faucet endpoints aren't available in your current client build,
	// keep this as a scaffold and remove the Skip once implemented.
	t.Skip("Faucet E2E not implemented in this client yet — enable once available.")

	// Example skeleton (uncomment when methods exist):
	// c := setupClient(t)
	// owner := createWallet(t, c)
	// tok := createBasicToken(t, c, owner.PublicKey)
	// startAt := time.Now().Add(1 * time.Minute)
	// expireAt := time.Now().Add(10 * time.Minute)
	// requestLimit := 5
	// out, err := c.AddFaucet(owner.PublicKey, tok.Address, startAt, expireAt, false, requestLimit)
	// if err != nil { t.Fatalf("AddFaucet: %v", err) }
	// // parse faucet state here and continue with DepositFunds / WithdrawFunds / Pause / Unpause / ClaimFunds
}

func TestCoupons(t *testing.T) { // testCoupons
	// Same note as Faucet — enable once the coupon contract calls are wired in the client.
	t.Skip("Coupons E2E not implemented in this client yet — enable once available.")

	// Skeleton for when available (align names with your client):
	// c := setupClient(t)
	// owner := createWallet(t, c)
	// tok := createBasicToken(t, c, owner.PublicKey)
	// start := time.Now().Add(5 * time.Second)
	// end := time.Now().Add(30 * time.Minute)
	// out, err := c.AddCoupon(owner.PublicKey, tok.Address, "percentage", "250", "", "", start, end, false, true, 100, 5, "<sha256_passcode>")
	// if err != nil { t.Fatalf("AddCoupon: %v", err) }
	// // UpdateCoupon / Pause / Unpause / GetCoupon / ListCoupons / Redeem…
}
