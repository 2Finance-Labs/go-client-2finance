package e2e_test


import (
	"testing"
	"time"
	client2f "github.com/2Finance-Labs/go-client-2finance/client_2finance"

	tokenV1Domain "gitlab.com/2finance/2finance-network/blockchain/contract/tokenV1/domain"
)


func TestTokenFlow(t *testing.T) {
	c := setupClient(t)
	owner, ownerPriv := createWallet(t, c)
	c.SetPrivateKey(ownerPriv)


	dec := 6
	tok := createBasicToken(t, c, owner.PublicKey, dec)


	// Mint & Burn
	if _, err := c.MintToken(tok.Address, owner.PublicKey, amt(35, dec), dec); err != nil { t.Fatalf("MintToken: %v", err) }
	if _, err := c.BurnToken(tok.Address, amt(12, dec), dec); err != nil { t.Fatalf("BurnToken: %v", err) }


	// Transfer to a new allowed wallet
	receiver, _ := createWallet(t, c)
	c.SetPrivateKey(ownerPriv)
	if _, err := c.AllowUsers(tok.Address, map[string]bool{receiver.PublicKey: true}); err != nil { t.Fatalf("AllowUsers: %v", err) }
	trOut, err := c.TransferToken(tok.Address, receiver.PublicKey, amt(1, dec), dec)
	if err != nil { t.Fatalf("TransferToken: %v", err) }
	var tr tokenV1Domain.Transfer
	unmarshalState(t, trOut.States[0].Object, &tr)
	if tr.ToAddress != receiver.PublicKey { t.Fatalf("transfer to mismatch: %s != %s", tr.ToAddress, receiver.PublicKey) }
	

	c.SetPrivateKey(ownerPriv)
	// Fee tiers & address
	if _, err := c.UpdateFeeTiers(tok.Address, []map[string]interface{}{{"min_amount": "0", "max_amount": amt(10_000, dec), "min_volume": "0", "max_volume": amt(100_000, dec), "fee_bps": 25}}); err != nil { t.Fatalf("UpdateFeeTiers: %v", err) }
	if _, err := c.UpdateFeeAddress(tok.Address, owner.PublicKey); err != nil { t.Fatalf("UpdateFeeAddress: %v", err) }


	// Metadata / Authorities / Pause
	if _, err := c.UpdateMetadata(tok.Address, "2F-NEW"+randSuffix(4), "2Finance New", dec, "Updated by tests", "https://example.com/img.png", "https://example.com", map[string]string{"twitter": "https://x.com/2f"}, map[string]string{"category": "DeFi"}, map[string]string{"tag":"e2e"}, "creator", "https://creator", time.Now().Add(30*24*time.Hour)); err != nil { t.Fatalf("UpdateMetadata: %v", err) }
	if _, err := c.RevokeMintAuthority(tok.Address, true); err != nil { t.Fatalf("RevokeMintAuthority: %v", err) }
	if _, err := c.RevokeUpdateAuthority(tok.Address, true); err != nil { t.Fatalf("RevokeUpdateAuthority: %v", err) }
	if _, err := c.PauseToken(tok.Address, true); err != nil { t.Fatalf("PauseToken: %v", err) }
	if _, err := c.UnpauseToken(tok.Address, false); err != nil { t.Fatalf("UnpauseToken: %v", err) }


	// Balances / Listings
	if _, err := c.GetTokenBalance(tok.Address, owner.PublicKey); err != nil { t.Fatalf("GetTokenBalance(owner): %v", err) }
	if _, err := c.ListTokenBalances(tok.Address, "", 1, 10, true); err != nil { t.Fatalf("ListTokenBalances: %v", err) }
	if _, err := c.GetToken(tok.Address, "", ""); err != nil { t.Fatalf("GetToken: %v", err) }
	if _, err := c.ListTokens("", "", "", 1, 10, true); err != nil { t.Fatalf("ListTokens: %v", err) }
}


// createBasicToken creates a minimal token owned by ownerPub.
func createBasicToken(t *testing.T, c client2f.Client2FinanceNetwork, ownerPub string, decimals int) tokenV1Domain.Token {
	t.Helper()
	symbol := "2F" + randSuffix(4)
	name := "2Finance"
	totalSupply := amt(1_000_000, decimals)
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
		{"min_amount": "0", "max_amount": amt(10_000, decimals), "min_volume": "0", "max_volume": amt(100_000, decimals), "fee_bps": 50},
	}
	feeAddress := ownerPub
	freezeAuthorityRevoked := false
	mintAuthorityRevoked := false
	updateAuthorityRevoked := false
	paused := false
	expiredAt := time.Time{}

	out, err := c.AddToken(symbol, name, decimals, totalSupply, description, ownerPub, image, website, tagsSocial, tagsCat, tags, creator, creatorWebsite, allowUsers, blockUsers, feeTiers, feeAddress, freezeAuthorityRevoked, mintAuthorityRevoked, updateAuthorityRevoked, paused, expiredAt)
	if err != nil { t.Fatalf("AddToken: %v", err) }
	var tok tokenV1Domain.Token
	unmarshalState(t, out.States[0].Object, &tok)
	if tok.Address == "" { t.Fatalf("token address empty") }
	return tok
}


func createMint(t *testing.T, c client2f.Client2FinanceNetwork, token tokenV1Domain.Token, to string, amount string, decimals int) tokenV1Domain.Mint {
	t.Helper()
	out, err := c.MintToken(token.Address, to, amount, decimals)
	if err != nil { t.Fatalf("MintToken: %v", err) }
	var m tokenV1Domain.Mint
	unmarshalState(t, out.States[0].Object, &m)
	if m.TokenAddress != token.Address { t.Fatalf("mint token mismatch: %s != %s", m.TokenAddress, token.Address) }
	return m
}


func createBurn(t *testing.T, c client2f.Client2FinanceNetwork, token tokenV1Domain.Token, amount string, decimals int) tokenV1Domain.Burn {
	t.Helper()
	out, err := c.BurnToken(token.Address, amount, decimals)
	if err != nil { t.Fatalf("BurnToken: %v", err) }
	var b tokenV1Domain.Burn
	unmarshalState(t, out.States[0].Object, &b)
	if b.TokenAddress != token.Address { t.Fatalf("burn token mismatch: %s != %s", b.TokenAddress, token.Address) }
	return b
}


func createTransfer(t *testing.T, c client2f.Client2FinanceNetwork, token tokenV1Domain.Token, to string, amount string, decimals int) tokenV1Domain.Transfer {
	t.Helper()
	out, err := c.TransferToken(token.Address, to, amount, decimals)
	if err != nil { t.Fatalf("TransferToken: %v", err) }
	var tr tokenV1Domain.Transfer
	unmarshalState(t, out.States[0].Object, &tr)
	if tr.ToAddress != to { t.Fatalf("transfer to mismatch: %s != %s", tr.ToAddress, to) }
	return tr
}


func approveSpender(t *testing.T, c client2f.Client2FinanceNetwork, token tokenV1Domain.Token, owner, spender, amount string) {
	t.Helper()
	if _, err := c.ApproveSpender(token.Address, owner, spender, amount, time.Now().Add(30*time.Minute)); err != nil {
		t.Fatalf("ApproveSpender: %v", err)
	}
}
