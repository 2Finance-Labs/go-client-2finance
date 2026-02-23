package e2e_test

import (
	"log"
	"testing"
	"time"

	client2f "github.com/2Finance-Labs/go-client-2finance/client_2finance"
	"github.com/2Finance-Labs/go-client-2finance/tests"

	"gitlab.com/2finance/2finance-network/blockchain/contract/contractV1/models"
	"gitlab.com/2finance/2finance-network/blockchain/contract/tokenV1"
	"gitlab.com/2finance/2finance-network/blockchain/contract/tokenV1/domain"
	tokenV1Domain "gitlab.com/2finance/2finance-network/blockchain/contract/tokenV1/domain"
)

func TestTokenFlowFungible(t *testing.T) {
	c := setupClient(t)
	owner, ownerPriv := createWallet(t, c)
	c.SetPrivateKey(ownerPriv)

	dec := 6
	tokenType := tokenV1Domain.FUNGIBLE
	stablecoin := false

	tok := createBasicToken(t, c, owner.PublicKey, dec, true, tokenType, stablecoin)

	log.Printf("Token Address: %s", tok.Address)
	log.Printf("Token Symbol: %s", tok.Symbol)
	log.Printf("Token Name: %s", tok.Name)
	log.Printf("Token Decimals: %d", tok.Decimals)
	log.Printf("Token Total Supply: %s", tok.TotalSupply)
	log.Printf("Token Description: %s", tok.Description)
	log.Printf("Token Image: %s", tok.Image)
	log.Printf("Token Website: %s", tok.Website)
	log.Printf("Token Tags Social: %+v", tok.TagsSocialMedia)
	log.Printf("Token Tags Category: %+v", tok.TagsCategory)
	log.Printf("Token Tags: %+v", tok.Tags)
	log.Printf("Token Creator: %s", tok.Creator)
	log.Printf("Token Creator Website: %s", tok.CreatorWebsite)
	log.Printf("Token Access Policy Mode: %s", tok.AccessPolicy.Mode)
	log.Printf("Token Access Policy Users: %+v", tok.AccessPolicy.Users)
	log.Printf("Token Frozen Accounts: %+v", tok.FrozenAccounts)
	log.Printf("Token Fee Tiers: %+v", tok.FeeTiersList)
	log.Printf("Token Fee Address: %s", tok.FeeAddress)
	log.Printf("Token Freeze Authority Revoked: %v", tok.FreezeAuthorityRevoked)
	log.Printf("Token Mint Authority Revoked: %v", tok.MintAuthorityRevoked)
	log.Printf("Token Update Authority Revoked: %v", tok.UpdateAuthorityRevoked)
	log.Printf("Token Paused: %v", tok.Paused)
	log.Printf("Token Expired At: %s", tok.ExpiredAt.String())
	log.Printf("Token Asset GLB URI: %s", tok.AssetGLBUri)
	log.Printf("Token Type: %s", tok.TokenType)
	log.Printf("Token Transferable: %v", tok.Transferable)
	log.Printf("Token Stablecoin: %v", tok.Stablecoin)

	// Mint
	mintOut, err := c.MintToken(tok.Address, owner.PublicKey, amt(35, dec), dec, tok.TokenType)
	if err != nil {
		t.Fatalf("MintToken: %v", err)
	}
	var mint tokenV1Domain.Mint
	tests.UnmarshalState(t, mintOut.States[0].Object, &mint)

	log.Printf("Mint TokenAddress: %s\n", mint.TokenAddress)
	log.Printf("Mint ToAddress: %s\n", mint.MintTo)
	log.Printf("Mint Amount: %s\n", mint.Amount)
	log.Printf("Mint TokenType: %s\n", mint.TokenType)
	log.Printf("Mint TokenUUIDList: %+v\n", mint.TokenUUIDList)

	// Burn
	burnOut, err := c.BurnToken(tok.Address, amt(12, dec), dec, tok.TokenType, "")
	if err != nil {
		t.Fatalf("BurnToken: %v", err)
	}
	var burn tokenV1Domain.Burn
	tests.UnmarshalState(t, burnOut.States[0].Object, &burn)

	log.Printf("Burn TokenAddress: %s\n", burn.TokenAddress)
	log.Printf("Burn FromAddress: %s\n", burn.BurnFrom)
	log.Printf("Burn Amount: %s\n", burn.Amount)
	log.Printf("Burn TokenType: %s\n", burn.TokenType)
	log.Printf("Burn TokenUUIDList: %+v\n", burn.UUID)

	// Transfer to allowed wallet
	receiver, _ := createWallet(t, c)
	c.SetPrivateKey(ownerPriv)

	allowOut, err := c.AllowUsers(tok.Address, map[string]bool{
		receiver.PublicKey: true,
	})
	if err != nil {
		t.Fatalf("AllowUsers: %v", err)
	}
	var accessPolicy tokenV1Domain.AccessPolicy
	tests.UnmarshalState(t, allowOut.States[0].Object, &accessPolicy)

	log.Printf("AllowUsers Mode: %s\n", accessPolicy.Mode)
	log.Printf("AllowUsers Users: %+v\n", accessPolicy.Users)

	trOut, err := c.TransferToken(
		tok.Address,
		receiver.PublicKey,
		amt(1, dec),
		dec,
		tok.TokenType,
		"",
	)
	if err != nil {
		t.Fatalf("TransferToken: %v", err)
	}
	var tr tokenV1Domain.Transfer
	tests.UnmarshalState(t, trOut.States[0].Object, &tr)

	log.Printf("Transfer TokenAddress: %s\n", tr.TokenAddress)
	log.Printf("Transfer FromAddress: %s\n", tr.FromAddress)
	log.Printf("Transfer ToAddress: %s\n", tr.ToAddress)
	log.Printf("Transfer Amount: %s\n", tr.Amount)
	log.Printf("Transfer TokenType: %s\n", tr.TokenType)
	log.Printf("Transfer UUID: %s\n", tr.UUID)

	if tr.ToAddress != receiver.PublicKey {
		t.Fatalf("transfer mismatch: %s != %s", tr.ToAddress, receiver.PublicKey)
	}

	// Fee tiers
	feeTiersOut, err := c.UpdateFeeTiers(tok.Address, []map[string]interface{}{
		{
			"min_amount": "0",
			"max_amount": amt(10_000, dec),
			"min_volume": "0",
			"max_volume": amt(100_000, dec),
			"fee_bps":    25,
		},
	})
	if err != nil {
		t.Fatalf("UpdateFeeTiers: %v", err)
	}
	var feeTiers tokenV1Domain.FeeTiers
	tests.UnmarshalState(t, feeTiersOut.States[0].Object, &feeTiers)

	log.Printf("UpdateFeeTiers FeeTiersList: %+v\n", feeTiers.FeeTiersList)

	// Fee address
	feeOut, err := c.UpdateFeeAddress(tok.Address, owner.PublicKey)
	if err != nil {
		t.Fatalf("UpdateFeeAddress: %v", err)
	}
	var fee tokenV1Domain.Fee
	tests.UnmarshalState(t, feeOut.States[0].Object, &fee)

	log.Printf("UpdateFeeAddress TokenAddress: %s\n", fee.TokenAddress)
	log.Printf("UpdateFeeAddress FeeAddress: %s\n", fee.FeeAddress)

	// Metadata
	metaOut, err := c.UpdateMetadata(
		tok.Address,
		"2F-NEW"+randSuffix(4),
		"2Finance New",
		dec,
		"Updated by tests",
		"https://example.com/img.png",
		"https://example.com",
		map[string]string{"twitter": "https://x.com/2f"},
		map[string]string{"category": "DeFi"},
		map[string]string{"tag": "e2e"},
		"creator",
		"https://creator",
		time.Now().Add(30*24*time.Hour),
	)
	if err != nil {
		t.Fatalf("UpdateMetadata: %v", err)
	}
	var meta tokenV1Domain.Token
	tests.UnmarshalState(t, metaOut.States[0].Object, &meta)

	log.Printf("UpdateMetadata TokenAddress: %s\n", meta.Address)
	log.Printf("UpdateMetadata Symbol: %s\n", meta.Symbol)
	log.Printf("UpdateMetadata Name: %s\n", meta.Name)
	log.Printf("UpdateMetadata Decimals: %d\n", meta.Decimals)
	log.Printf("UpdateMetadata Description: %s\n", meta.Description)
	log.Printf("UpdateMetadata Image: %s\n", meta.Image)
	log.Printf("UpdateMetadata Website: %s\n", meta.Website)
	log.Printf("UpdateMetadata TagsSocialMedia: %+v\n", meta.TagsSocialMedia)
	log.Printf("UpdateMetadata TagsCategory: %+v\n", meta.TagsCategory)
	log.Printf("UpdateMetadata Tags: %+v\n", meta.Tags)
	log.Printf("UpdateMetadata Creator: %s\n", meta.Creator)
	log.Printf("UpdateMetadata CreatorWebsite: %s\n", meta.CreatorWebsite)
	log.Printf("UpdateMetadata ExpiredAt: %s\n", meta.ExpiredAt)

	// Revoke Mint Authority
	revMintOut, err := c.RevokeMintAuthority(tok.Address, true)
	if err != nil {
		t.Fatalf("RevokeMintAuthority: %v", err)
	}
	var revMint tokenV1Domain.Token
	tests.UnmarshalState(t, revMintOut.States[0].Object, &revMint)

	log.Printf("RevokeMintAuthority TokenAddress: %s\n", revMint.Address)
	log.Printf("RevokeMintAuthority MintAuthorityRevoked: %v\n", revMint.MintAuthorityRevoked)

	// Revoke Update Authority
	revUpdOut, err := c.RevokeUpdateAuthority(tok.Address, true)
	if err != nil {
		t.Fatalf("RevokeUpdateAuthority: %v", err)
	}
	var revUpd tokenV1Domain.Token
	tests.UnmarshalState(t, revUpdOut.States[0].Object, &revUpd)

	log.Printf("RevokeUpdateAuthority TokenAddress: %s\n", revUpd.Address)
	log.Printf("RevokeUpdateAuthority UpdateAuthorityRevoked: %v\n", revUpd.UpdateAuthorityRevoked)

	// Pause
	pauseOut, err := c.PauseToken(tok.Address, true)
	if err != nil {
		t.Fatalf("PauseToken: %v", err)
	}
	var pause tokenV1Domain.Token
	tests.UnmarshalState(t, pauseOut.States[0].Object, &pause)

	log.Printf("PauseToken TokenAddress: %s\n", pause.Address)
	log.Printf("PauseToken Paused: %v\n", pause.Paused)

	// Unpause
	unpauseOut, err := c.UnpauseToken(tok.Address, false)
	if err != nil {
		t.Fatalf("UnpauseToken: %v", err)
	}
	var unpause tokenV1Domain.Token
	tests.UnmarshalState(t, unpauseOut.States[0].Object, &unpause)

	log.Printf("UnpauseToken TokenAddress: %s\n", unpause.Address)
	log.Printf("UnpauseToken Paused: %v\n", unpause.Paused)

	// Freeze wallet
	freezeOut, err := c.FreezeWallet(tok.Address, owner.PublicKey)
	if err != nil {
		t.Fatalf("FreezeWallet: %v", err)
	}
	var freeze tokenV1Domain.Token
	tests.UnmarshalState(t, freezeOut.States[0].Object, &freeze)

	log.Printf("FreezeWallet TokenAddress: %s\n", freeze.Address)
	log.Printf("FreezeWallet Wallet: %s\n", freeze.Owner)
	log.Printf("FreezeWallet Frozen Accounts: %v\n", freeze.FrozenAccounts)

	// Unfreeze wallet
	unfreezeOut, err := c.UnfreezeWallet(tok.Address, owner.PublicKey)
	if err != nil {
		t.Fatalf("UnfreezeWallet: %v", err)
	}
	var unfreeze tokenV1Domain.Token
	tests.UnmarshalState(t, unfreezeOut.States[0].Object, &unfreeze)

	log.Printf("UnfreezeWallet TokenAddress: %s\n", unfreeze.Address)
	log.Printf("UnfreezeWallet Wallet: %s\n", unfreeze.Owner)
	log.Printf("UnfreezeWallet Frozen Accounts: %v\n", unfreeze.FrozenAccounts)

	// Balances / Listings
	getBalOut, err := c.GetTokenBalance(tok.Address, owner.PublicKey)
	if err != nil {
		t.Fatalf("GetTokenBalance(owner): %v", err)
	}
	var bal tokenV1Domain.Balance
	tests.UnmarshalState(t, getBalOut.States[0].Object, &bal)

	log.Printf("GetTokenBalance TokenAddress: %s\n", bal.TokenAddress)
	log.Printf("GetTokenBalance OwnerAddress: %s\n", bal.OwnerAddress)
	log.Printf("GetTokenBalance Amount: %s\n", bal.Amount)

	listBalOut, err := c.ListTokenBalances(tok.Address, "", 1, 10, true)
	if err != nil {
		t.Fatalf("ListTokenBalances: %v", err)
	}
	log.Printf("Balance Listed Successfully:\n%+v\n", listBalOut)

	getTokOut, err := c.GetToken(tok.Address, "", "")
	if err != nil {
		t.Fatalf("GetToken: %v", err)
	}
	var got tokenV1Domain.Token
	tests.UnmarshalState(t, getTokOut.States[0].Object, &got)

	log.Printf("GetToken Address: %s\n", got.Address)
	log.Printf("GetToken Symbol: %s\n", got.Symbol)
	log.Printf("GetToken Name: %s\n", got.Name)
	log.Printf("GetToken TotalSupply: %s\n", got.TotalSupply)
	log.Printf("GetToken FeeAddress: %s\n", got.FeeAddress)
	log.Printf("GetToken Paused: %v\n", got.Paused)

	listTokOut, err := c.ListTokens("", "", "", 1, 10, true)
	if err != nil {
		t.Fatalf("ListTokens: %v", err)
	}

	log.Printf("Token Listed Successfully:\n%+v\n", listTokOut)
}

func TestTokenFlowNonFungible(t *testing.T) {
	c := setupClient(t)
	owner, ownerPriv := createWallet(t, c)
	c.SetPrivateKey(ownerPriv)

	dec := 0
	tokenType := tokenV1Domain.NON_FUNGIBLE
	stablecoin := false

	// keep same "createBasicToken" call pattern used in fungible flow
	tok := createBasicToken(t, c, owner.PublicKey, dec, false, tokenType, stablecoin)

	log.Printf("Token Address: %s", tok.Address)
	log.Printf("Token Symbol: %s", tok.Symbol)
	log.Printf("Token Name: %s", tok.Name)
	log.Printf("Token Decimals: %d", tok.Decimals)
	log.Printf("Token Total Supply: %s", tok.TotalSupply)
	log.Printf("Token Description: %s", tok.Description)
	log.Printf("Token Image: %s", tok.Image)
	log.Printf("Token Website: %s", tok.Website)
	log.Printf("Token Tags Social: %+v", tok.TagsSocialMedia)
	log.Printf("Token Tags Category: %+v", tok.TagsCategory)
	log.Printf("Token Tags: %+v", tok.Tags)
	log.Printf("Token Creator: %s", tok.Creator)
	log.Printf("Token Creator Website: %s", tok.CreatorWebsite)
	log.Printf("Token Access Policy Mode: %s", tok.AccessPolicy.Mode)
	log.Printf("Token Access Policy Users: %+v", tok.AccessPolicy.Users)
	log.Printf("Token Frozen Accounts: %+v", tok.FrozenAccounts)
	log.Printf("Token Fee Tiers: %+v", tok.FeeTiersList)
	log.Printf("Token Fee Address: %s", tok.FeeAddress)
	log.Printf("Token Freeze Authority Revoked: %v", tok.FreezeAuthorityRevoked)
	log.Printf("Token Mint Authority Revoked: %v", tok.MintAuthorityRevoked)
	log.Printf("Token Update Authority Revoked: %v", tok.UpdateAuthorityRevoked)
	log.Printf("Token Paused: %v", tok.Paused)
	log.Printf("Token Expired At: %s", tok.ExpiredAt.String())
	log.Printf("Token Asset GLB URI: %s", tok.AssetGLBUri)
	log.Printf("Token Type: %s", tok.TokenType)
	log.Printf("Token Transferable: %v", tok.Transferable)
	log.Printf("Token Stablecoin: %v", tok.Stablecoin)

	// Mint NFT
	mintOut, err := c.MintToken(tok.Address, owner.PublicKey, amt(35, dec), dec, tok.TokenType)
	if err != nil {
		t.Fatalf("MintToken NFT: %v", err)
	}
	var mint tokenV1Domain.Mint
	tests.UnmarshalState(t, mintOut.States[0].Object, &mint)

	log.Printf("Mint TokenAddress: %s\n", mint.TokenAddress)
	log.Printf("Mint ToAddress: %s\n", mint.MintTo)
	log.Printf("Mint Amount: %s\n", mint.Amount)
	log.Printf("Mint TokenType: %s\n", mint.TokenType)
	log.Printf("Mint TokenUUIDList: %+v\n", mint.TokenUUIDList)

	if len(mint.TokenUUIDList) != 35 {
		t.Fatalf("expected %d uuid, got %d", 35, len(mint.TokenUUIDList))
	}

	// Burn 1 NFT (with UUID)
	burnOut, err := c.BurnToken(
		tok.Address,
		amt(1, dec),
		dec,
		tok.TokenType,
		mint.TokenUUIDList[0],
	)
	if err != nil {
		t.Fatalf("BurnToken: %v", err)
	}
	var burn tokenV1Domain.Burn
	tests.UnmarshalState(t, burnOut.States[0].Object, &burn)

	log.Printf("Burn TokenAddress: %s\n", burn.TokenAddress)
	log.Printf("Burn FromAddress: %s\n", burn.BurnFrom)
	log.Printf("Burn Amount: %s\n", burn.Amount)
	log.Printf("Burn TokenType: %s\n", burn.TokenType)
	log.Printf("Burn TokenUUIDList: %+v\n", burn.UUID)

	// Transfer NFT to allowed wallet
	receiver, _ := createWallet(t, c)
	c.SetPrivateKey(ownerPriv)

	allowOut, err := c.AllowUsers(tok.Address, map[string]bool{
		receiver.PublicKey: true,
	})
	if err != nil {
		t.Fatalf("AllowUsers: %v", err)
	}
	var accessPolicy tokenV1Domain.AccessPolicy
	tests.UnmarshalState(t, allowOut.States[0].Object, &accessPolicy)

	log.Printf("AllowUsers Mode: %s\n", accessPolicy.Mode)
	log.Printf("AllowUsers Users: %+v\n", accessPolicy.Users)

	trOut, err := c.TransferToken(
		tok.Address,
		receiver.PublicKey,
		amt(1, dec),
		dec,
		tok.TokenType,
		mint.TokenUUIDList[1], // use a UUID that was not burned
	)
	if err != nil {
		t.Fatalf("Transfer NFT: %v", err)
	}
	var tr tokenV1Domain.Transfer
	tests.UnmarshalState(t, trOut.States[0].Object, &tr)

	log.Printf("Transfer TokenAddress: %s\n", tr.TokenAddress)
	log.Printf("Transfer FromAddress: %s\n", tr.FromAddress)
	log.Printf("Transfer ToAddress: %s\n", tr.ToAddress)
	log.Printf("Transfer Amount: %s\n", tr.Amount)
	log.Printf("Transfer TokenType: %s\n", tr.TokenType)
	log.Printf("Transfer UUID: %s\n", tr.UUID)

	if tr.ToAddress != receiver.PublicKey {
		t.Fatalf("transfer mismatch: %s != %s", tr.ToAddress, receiver.PublicKey)
	}

	// Fee tiers (same pattern as fungible flow)
	feeTiersOut, err := c.UpdateFeeTiers(tok.Address, []map[string]interface{}{
		{
			"min_amount": "0",
			"max_amount": amt(10_000, dec),
			"min_volume": "0",
			"max_volume": amt(100_000, dec),
			"fee_bps":    25,
		},
	})
	if err != nil {
		t.Fatalf("UpdateFeeTiers: %v", err)
	}
	var feeTiers tokenV1Domain.FeeTiers
	tests.UnmarshalState(t, feeTiersOut.States[0].Object, &feeTiers)

	log.Printf("UpdateFeeTiers FeeTiersList: %+v\n", feeTiers.FeeTiersList)

	// Fee address
	feeOut, err := c.UpdateFeeAddress(tok.Address, owner.PublicKey)
	if err != nil {
		t.Fatalf("UpdateFeeAddress: %v", err)
	}
	var fee tokenV1Domain.Fee
	tests.UnmarshalState(t, feeOut.States[0].Object, &fee)

	log.Printf("UpdateFeeAddress TokenAddress: %s\n", fee.TokenAddress)
	log.Printf("UpdateFeeAddress FeeAddress: %s\n", fee.FeeAddress)

	// Metadata
	metaOut, err := c.UpdateMetadata(
		tok.Address,
		"2F-NEW"+randSuffix(4),
		"2Finance New",
		dec,
		"Updated by tests",
		"https://example.com/img.png",
		"https://example.com",
		map[string]string{"twitter": "https://x.com/2f"},
		map[string]string{"category": "DeFi"},
		map[string]string{"tag": "e2e"},
		"creator",
		"https://creator",
		time.Now().Add(30*24*time.Hour),
	)
	if err != nil {
		t.Fatalf("UpdateMetadata: %v", err)
	}
	var meta tokenV1Domain.Token
	tests.UnmarshalState(t, metaOut.States[0].Object, &meta)

	log.Printf("UpdateMetadata TokenAddress: %s\n", meta.Address)
	log.Printf("UpdateMetadata Symbol: %s\n", meta.Symbol)
	log.Printf("UpdateMetadata Name: %s\n", meta.Name)
	log.Printf("UpdateMetadata Decimals: %d\n", meta.Decimals)
	log.Printf("UpdateMetadata Description: %s\n", meta.Description)
	log.Printf("UpdateMetadata Image: %s\n", meta.Image)
	log.Printf("UpdateMetadata Website: %s\n", meta.Website)
	log.Printf("UpdateMetadata TagsSocialMedia: %+v\n", meta.TagsSocialMedia)
	log.Printf("UpdateMetadata TagsCategory: %+v\n", meta.TagsCategory)
	log.Printf("UpdateMetadata Tags: %+v\n", meta.Tags)
	log.Printf("UpdateMetadata Creator: %s\n", meta.Creator)
	log.Printf("UpdateMetadata CreatorWebsite: %s\n", meta.CreatorWebsite)
	log.Printf("UpdateMetadata ExpiredAt: %s\n", meta.ExpiredAt)

	// Revoke Mint Authority
	revMintOut, err := c.RevokeMintAuthority(tok.Address, true)
	if err != nil {
		t.Fatalf("RevokeMintAuthority: %v", err)
	}
	var revMint tokenV1Domain.Token
	tests.UnmarshalState(t, revMintOut.States[0].Object, &revMint)

	log.Printf("RevokeMintAuthority TokenAddress: %s\n", revMint.Address)
	log.Printf("RevokeMintAuthority MintAuthorityRevoked: %v\n", revMint.MintAuthorityRevoked)

	// Revoke Update Authority
	revUpdOut, err := c.RevokeUpdateAuthority(tok.Address, true)
	if err != nil {
		t.Fatalf("RevokeUpdateAuthority: %v", err)
	}
	var revUpd tokenV1Domain.Token
	tests.UnmarshalState(t, revUpdOut.States[0].Object, &revUpd)

	log.Printf("RevokeUpdateAuthority TokenAddress: %s\n", revUpd.Address)
	log.Printf("RevokeUpdateAuthority UpdateAuthorityRevoked: %v\n", revUpd.UpdateAuthorityRevoked)

	// Pause
	pauseOut, err := c.PauseToken(tok.Address, true)
	if err != nil {
		t.Fatalf("PauseToken: %v", err)
	}
	var pause tokenV1Domain.Token
	tests.UnmarshalState(t, pauseOut.States[0].Object, &pause)

	log.Printf("PauseToken TokenAddress: %s\n", pause.Address)
	log.Printf("PauseToken Paused: %v\n", pause.Paused)

	// Unpause
	unpauseOut, err := c.UnpauseToken(tok.Address, false)
	if err != nil {
		t.Fatalf("UnpauseToken: %v", err)
	}
	var unpause tokenV1Domain.Token
	tests.UnmarshalState(t, unpauseOut.States[0].Object, &unpause)

	log.Printf("UnpauseToken TokenAddress: %s\n", unpause.Address)
	log.Printf("UnpauseToken Paused: %v\n", unpause.Paused)

	// Freeze wallet
	freezeOut, err := c.FreezeWallet(tok.Address, owner.PublicKey)
	if err != nil {
		t.Fatalf("FreezeWallet: %v", err)
	}
	var freeze tokenV1Domain.Token
	tests.UnmarshalState(t, freezeOut.States[0].Object, &freeze)

	log.Printf("FreezeWallet TokenAddress: %s\n", freeze.Address)
	log.Printf("FreezeWallet Wallet: %s\n", freeze.Owner)
	log.Printf("FreezeWallet Frozen Accounts: %v\n", freeze.FrozenAccounts)

	// Unfreeze wallet
	unfreezeOut, err := c.UnfreezeWallet(tok.Address, owner.PublicKey)
	if err != nil {
		t.Fatalf("UnfreezeWallet: %v", err)
	}
	var unfreeze tokenV1Domain.Token
	tests.UnmarshalState(t, unfreezeOut.States[0].Object, &unfreeze)

	log.Printf("UnfreezeWallet TokenAddress: %s\n", unfreeze.Address)
	log.Printf("UnfreezeWallet Wallet: %s\n", unfreeze.Owner)
	log.Printf("UnfreezeWallet Frozen Accounts: %v\n", unfreeze.FrozenAccounts)

	// Balances / Listings (same pattern as fungible flow)
	getBalOut, err := c.GetTokenBalance(tok.Address, owner.PublicKey)
	if err != nil {
		t.Fatalf("GetTokenBalance(owner): %v", err)
	}
	var bal tokenV1Domain.Balance
	tests.UnmarshalState(t, getBalOut.States[0].Object, &bal)

	log.Printf("GetTokenBalance TokenAddress: %s\n", bal.TokenAddress)
	log.Printf("GetTokenBalance OwnerAddress: %s\n", bal.OwnerAddress)
	log.Printf("GetTokenBalance Amount: %s\n", bal.Amount)

	listBalOut, err := c.ListTokenBalances(tok.Address, "", 1, 10, true)
	if err != nil {
		t.Fatalf("ListTokenBalances: %v", err)
	}
	log.Printf("Balance Listed Successfully:\n%+v\n", listBalOut)

	getTokOut, err := c.GetToken(tok.Address, "", "")
	if err != nil {
		t.Fatalf("GetToken: %v", err)
	}
	var got tokenV1Domain.Token
	tests.UnmarshalState(t, getTokOut.States[0].Object, &got)

	log.Printf("GetToken Address: %s\n", got.Address)
	log.Printf("GetToken Symbol: %s\n", got.Symbol)
	log.Printf("GetToken Name: %s\n", got.Name)
	log.Printf("GetToken TotalSupply: %s\n", got.TotalSupply)
	log.Printf("GetToken FeeAddress: %s\n", got.FeeAddress)
	log.Printf("GetToken Paused: %v\n", got.Paused)

	listTokOut, err := c.ListTokens("", "", "", 1, 10, true)
	if err != nil {
		t.Fatalf("ListTokens: %v", err)
	}
	log.Printf("Token Listed Successfully:\n%+v\n", listTokOut)
}

// createBasicToken creates a minimal token owned by ownerPub.
func createBasicToken(
	t *testing.T,
	c client2f.Client2FinanceNetwork,
	ownerPub string,
	decimals int,
	requireFee bool,
	tokenType string,
	stablecoin bool,
) tokenV1Domain.Token {
	t.Helper()

	symbol := "2F" + randSuffix(4)
	name := "2Finance"
	var totalSupply string
	if tokenType == tokenV1Domain.NON_FUNGIBLE {
		totalSupply = "1"
	} else {
		totalSupply = amt(1_000_000, decimals)
	}
	description := "e2e token created by tests"
	image := "https://example.com/image.png"
	website := "https://example.com"
	tagsSocial := map[string]string{"twitter": "https://twitter.com/2finance"}
	tagsCat := map[string]string{"category": "DeFi"}
	tags := map[string]string{"tag1": "DeFi", "tag2": "Blockchain"}
	creator := "2Finance Test"
	creatorWebsite := "https://creator.example"
	accessPolicy := domain.AccessPolicy{
		Mode: domain.ALLOW,
		Users: map[string]bool{
			ownerPub: true,
		},
	}
	frozenAccounts := map[string]bool{}
	feeTiers := []map[string]interface{}{}

	if requireFee {
		feeTiers = []map[string]interface{}{
			{
				"min_amount": "0",
				"max_amount": amt(10_000, decimals),
				"min_volume": "0",
				"max_volume": amt(100_000, decimals),
				"fee_bps":    50,
			},
		}
	}

	feeAddress := ownerPub
	freezeAuthorityRevoked := false
	mintAuthorityRevoked := false
	updateAuthorityRevoked := false
	paused := false
	expiredAt := time.Time{}

	contractState := models.ContractStateModel{}
	deployedContract, err := c.DeployContract1(tokenV1.TOKEN_CONTRACT_V1)
	if err != nil {
		t.Fatalf("DeployContract: %v", err)
	}
	unmarshalState(t, deployedContract.States[0].Object, &contractState)
	address := contractState.Address

	assetGLBUri := "https://example.com/asset.glb"
	transferable := true

	out, err := c.AddToken(
		address,
		symbol,
		name,
		decimals,
		totalSupply,
		description,
		ownerPub,
		image,
		website,
		tagsSocial,
		tagsCat,
		tags,
		creator,
		creatorWebsite,
		accessPolicy,
		frozenAccounts,
		feeTiers,
		feeAddress,
		freezeAuthorityRevoked,
		mintAuthorityRevoked,
		updateAuthorityRevoked,
		paused,
		expiredAt,
		assetGLBUri,
		tokenType,
		transferable,
		stablecoin,
	)
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

func createMint(t *testing.T, c client2f.Client2FinanceNetwork, token tokenV1Domain.Token, to string, amount string, decimals int, tokenType string) tokenV1Domain.Mint {
	t.Helper()
	out, err := c.MintToken(token.Address, to, amount, decimals, tokenType)
	if err != nil {
		t.Fatalf("MintToken: %v", err)
	}
	var m tokenV1Domain.Mint
	unmarshalState(t, out.States[0].Object, &m)
	if m.TokenAddress != token.Address {
		t.Fatalf("mint token mismatch: %s != %s", m.TokenAddress, token.Address)
	}
	return m
}

func createBurn(t *testing.T, c client2f.Client2FinanceNetwork, token tokenV1Domain.Token, amount string, decimals int, tokenType, uuid string) tokenV1Domain.Burn {
	t.Helper()
	out, err := c.BurnToken(token.Address, amount, decimals, tokenType, uuid)
	if err != nil {
		t.Fatalf("BurnToken: %v", err)
	}
	var b tokenV1Domain.Burn
	unmarshalState(t, out.States[0].Object, &b)
	if b.TokenAddress != token.Address {
		t.Fatalf("burn token mismatch: %s != %s", b.TokenAddress, token.Address)
	}
	return b
}

func createTransfer(t *testing.T, c client2f.Client2FinanceNetwork, token tokenV1Domain.Token, to string, amount string, decimals int, tokenType, uuid string) tokenV1Domain.Transfer {
	t.Helper()
	out, err := c.TransferToken(token.Address, to, amount, decimals, tokenType, uuid)
	if err != nil {
		t.Fatalf("TransferToken: %v", err)
	}
	var tr tokenV1Domain.Transfer
	unmarshalState(t, out.States[0].Object, &tr)
	if tr.ToAddress != to {
		t.Fatalf("transfer to mismatch: %s != %s", tr.ToAddress, to)
	}
	return tr
}
