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

	// -------------------------
	// Token (validate + log)
	// -------------------------
	if tok.Address == "" {
		t.Fatalf("token address empty")
	}
	if tok.Symbol == "" {
		t.Fatalf("token symbol empty")
	}
	if tok.Name == "" {
		t.Fatalf("token name empty")
	}
	if tok.Decimals != dec {
		t.Fatalf("token decimals mismatch: got %d want %d", tok.Decimals, dec)
	}
	if tok.TokenType != tokenType {
		t.Fatalf("token type mismatch: got %s want %s", tok.TokenType, tokenType)
	}
	if tok.Stablecoin != stablecoin {
		t.Fatalf("token stablecoin mismatch: got %v want %v", tok.Stablecoin, stablecoin)
	}
	if tok.Creator == "" {
		t.Fatalf("token creator empty")
	}
	if tok.AccessPolicy.Mode == "" {
		t.Fatalf("token access policy mode empty")
	}
	if tok.AccessPolicy.Users == nil {
		t.Fatalf("token access policy users nil")
	}
	if !tok.AccessPolicy.Users[owner.PublicKey] {
		t.Fatalf("token access policy must include owner: %s", owner.PublicKey)
	}

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

	// -------------------------
	// Mint (envelope + unmarshal + validate + log)
	// -------------------------
	mintOut, err := c.MintToken(tok.Address, owner.PublicKey, amt(35, dec), dec, tok.TokenType)
	if err != nil {
		t.Fatalf("MintToken: %v", err)
	}
	if len(mintOut.States) == 0 {
		t.Fatalf("MintToken returned empty States")
	}
	if mintOut.States[0].Object == nil {
		t.Fatalf("MintToken returned nil state object")
	}

	var mint tokenV1Domain.Mint
	tests.UnmarshalState(t, mintOut.States[0].Object, &mint)

	if mint.TokenAddress != tok.Address {
		t.Fatalf("Mint TokenAddress mismatch: got %s want %s", mint.TokenAddress, tok.Address)
	}
	if mint.MintTo != owner.PublicKey {
		t.Fatalf("Mint ToAddress mismatch: got %s want %s", mint.MintTo, owner.PublicKey)
	}
	expectedMintAmount := amt(35, dec)
	if mint.Amount != expectedMintAmount {
		t.Fatalf("Mint Amount mismatch: got %s want %s", mint.Amount, expectedMintAmount)
	}
	if mint.TokenType != tok.TokenType {
		t.Fatalf("Mint TokenType mismatch: got %s want %s", mint.TokenType, tok.TokenType)
	}
	if tok.TokenType == tokenV1Domain.FUNGIBLE && len(mint.TokenUUIDList) != 0 {
		t.Fatalf("Fungible token should not generate UUIDs")
	}
	if tok.TokenType == tokenV1Domain.NON_FUNGIBLE && len(mint.TokenUUIDList) == 0 {
		t.Fatalf("NFT mint must generate UUIDs")
	}

	log.Printf("Mint Output States: %+v", mintOut.States)
	log.Printf("Mint Output Logs: %+v", mintOut.Logs)
	log.Printf("Mint Output Delegated Call: %+v", mintOut.DelegatedCall)

	log.Printf("Mint TokenAddress: %s", mint.TokenAddress)
	log.Printf("Mint ToAddress: %s", mint.MintTo)
	log.Printf("Mint Amount: %s", mint.Amount)
	log.Printf("Mint TokenType: %s", mint.TokenType)
	log.Printf("Mint TokenUUIDList: %+v", mint.TokenUUIDList)

	// -------------------------
	// Burn (envelope + unmarshal + validate + log)
	// -------------------------
	burnOut, err := c.BurnToken(tok.Address, amt(12, dec), dec, tok.TokenType, "")
	if err != nil {
		t.Fatalf("BurnToken: %v", err)
	}
	if len(burnOut.States) == 0 {
		t.Fatalf("BurnToken returned empty States")
	}
	if burnOut.States[0].Object == nil {
		t.Fatalf("BurnToken returned nil state object")
	}

	var burn tokenV1Domain.Burn
	tests.UnmarshalState(t, burnOut.States[0].Object, &burn)

	if burn.TokenAddress != tok.Address {
		t.Fatalf("Burn TokenAddress mismatch: got %s want %s", burn.TokenAddress, tok.Address)
	}
	if burn.BurnFrom != owner.PublicKey {
		t.Fatalf("Burn FromAddress mismatch: got %s want %s", burn.BurnFrom, owner.PublicKey)
	}
	expectedBurnAmount := amt(12, dec)
	if burn.Amount != expectedBurnAmount {
		t.Fatalf("Burn Amount mismatch: got %s want %s", burn.Amount, expectedBurnAmount)
	}
	if burn.TokenType != tok.TokenType {
		t.Fatalf("Burn TokenType mismatch: got %s want %s", burn.TokenType, tok.TokenType)
	}
	if tok.TokenType == tokenV1Domain.FUNGIBLE && burn.UUID != "" {
		t.Fatalf("Fungible burn should not have UUID, got %q", burn.UUID)
	}

	log.Printf("Burn Output States: %+v", burnOut.States)
	log.Printf("Burn Output Logs: %+v", burnOut.Logs)
	log.Printf("Burn Output Delegated Call: %+v", burnOut.DelegatedCall)

	log.Printf("Burn TokenAddress: %s", burn.TokenAddress)
	log.Printf("Burn FromAddress: %s", burn.BurnFrom)
	log.Printf("Burn Amount: %s", burn.Amount)
	log.Printf("Burn TokenType: %s", burn.TokenType)
	log.Printf("Burn UUID: %s", burn.UUID)

	// -------------------------
	// AllowUsers (envelope + unmarshal + validate + log)
	// -------------------------
	receiver, _ := createWallet(t, c)
	c.SetPrivateKey(ownerPriv)

	allowOut, err := c.AllowUsers(tok.Address, map[string]bool{
		receiver.PublicKey: true,
	})
	if err != nil {
		t.Fatalf("AllowUsers: %v", err)
	}
	if len(allowOut.States) == 0 {
		t.Fatalf("AllowUsers returned empty States")
	}
	if allowOut.States[0].Object == nil {
		t.Fatalf("AllowUsers returned nil state object")
	}

	var accessPolicy tokenV1Domain.AccessPolicy
	tests.UnmarshalState(t, allowOut.States[0].Object, &accessPolicy)

	if accessPolicy.Mode == "" {
		t.Fatalf("AllowUsers Mode empty")
	}
	if accessPolicy.Users == nil {
		t.Fatalf("AllowUsers Users nil")
	}
	if !accessPolicy.Users[receiver.PublicKey] {
		t.Fatalf("AllowUsers missing receiver in allowlist: %s", receiver.PublicKey)
	}

	log.Printf("AllowUsers Output States: %+v", allowOut.States)
	log.Printf("AllowUsers Output Logs: %+v", allowOut.Logs)
	log.Printf("AllowUsers Output Delegated Call: %+v", allowOut.DelegatedCall)

	log.Printf("AllowUsers Mode: %s", accessPolicy.Mode)
	log.Printf("AllowUsers Users: %+v", accessPolicy.Users)

	// -------------------------
	// Transfer (envelope + unmarshal + validate + log)
	// -------------------------
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
	if len(trOut.States) == 0 {
		t.Fatalf("TransferToken returned empty States")
	}
	if trOut.States[0].Object == nil {
		t.Fatalf("TransferToken returned nil state object")
	}

	var tr tokenV1Domain.Transfer
	tests.UnmarshalState(t, trOut.States[0].Object, &tr)

	if tr.TokenAddress != tok.Address {
		t.Fatalf("Transfer TokenAddress mismatch: got %s want %s", tr.TokenAddress, tok.Address)
	}
	if tr.FromAddress != owner.PublicKey {
		t.Fatalf("Transfer FromAddress mismatch: got %s want %s", tr.FromAddress, owner.PublicKey)
	}
	if tr.ToAddress != receiver.PublicKey {
		t.Fatalf("Transfer ToAddress mismatch: got %s want %s", tr.ToAddress, receiver.PublicKey)
	}
	expectedTransferAmount := amt(1, dec)
	if tr.Amount != expectedTransferAmount {
		t.Fatalf("Transfer Amount mismatch: got %s want %s", tr.Amount, expectedTransferAmount)
	}
	if tr.TokenType != tok.TokenType {
		t.Fatalf("Transfer TokenType mismatch: got %s want %s", tr.TokenType, tok.TokenType)
	}
	if tok.TokenType == tokenV1Domain.FUNGIBLE && tr.UUID != "" {
		t.Fatalf("Fungible transfer should not have UUID, got %q", tr.UUID)
	}

	log.Printf("Transfer Output States: %+v", trOut.States)
	log.Printf("Transfer Output Logs: %+v", trOut.Logs)
	log.Printf("Transfer Output Delegated Call: %+v", trOut.DelegatedCall)

	log.Printf("Transfer TokenAddress: %s", tr.TokenAddress)
	log.Printf("Transfer FromAddress: %s", tr.FromAddress)
	log.Printf("Transfer ToAddress: %s", tr.ToAddress)
	log.Printf("Transfer Amount: %s", tr.Amount)
	log.Printf("Transfer TokenType: %s", tr.TokenType)
	log.Printf("Transfer UUID: %s", tr.UUID)

	// -------------------------
	// Fee tiers (envelope + unmarshal + validate + log)
	// -------------------------
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
	if len(feeTiersOut.States) == 0 {
		t.Fatalf("UpdateFeeTiers returned empty States")
	}
	if feeTiersOut.States[0].Object == nil {
		t.Fatalf("UpdateFeeTiers returned nil state object")
	}

	var feeTiers tokenV1Domain.FeeTiers
	tests.UnmarshalState(t, feeTiersOut.States[0].Object, &feeTiers)

	if feeTiers.FeeTiersList == nil || len(feeTiers.FeeTiersList) == 0 {
		t.Fatalf("UpdateFeeTiers returned empty FeeTiersList")
	}

	log.Printf("UpdateFeeTiers Output States: %+v", feeTiersOut.States)
	log.Printf("UpdateFeeTiers Output Logs: %+v", feeTiersOut.Logs)
	log.Printf("UpdateFeeTiers Output Delegated Call: %+v", feeTiersOut.DelegatedCall)

	log.Printf("UpdateFeeTiers FeeTiersList: %+v", feeTiers.FeeTiersList)

	// -------------------------
	// Fee address (envelope + unmarshal + validate + log)
	// -------------------------
	feeOut, err := c.UpdateFeeAddress(tok.Address, owner.PublicKey)
	if err != nil {
		t.Fatalf("UpdateFeeAddress: %v", err)
	}
	if len(feeOut.States) == 0 {
		t.Fatalf("UpdateFeeAddress returned empty States")
	}
	if feeOut.States[0].Object == nil {
		t.Fatalf("UpdateFeeAddress returned nil state object")
	}

	var fee tokenV1Domain.Fee
	tests.UnmarshalState(t, feeOut.States[0].Object, &fee)

	if fee.TokenAddress != tok.Address {
		t.Fatalf("UpdateFeeAddress TokenAddress mismatch: got %s want %s", fee.TokenAddress, tok.Address)
	}
	if fee.FeeAddress != owner.PublicKey {
		t.Fatalf("UpdateFeeAddress FeeAddress mismatch: got %s want %s", fee.FeeAddress, owner.PublicKey)
	}

	log.Printf("UpdateFeeAddress Output States: %+v", feeOut.States)
	log.Printf("UpdateFeeAddress Output Logs: %+v", feeOut.Logs)
	log.Printf("UpdateFeeAddress Output Delegated Call: %+v", feeOut.DelegatedCall)

	log.Printf("UpdateFeeAddress TokenAddress: %s", fee.TokenAddress)
	log.Printf("UpdateFeeAddress FeeAddress: %s", fee.FeeAddress)

	// -------------------------
	// Metadata (envelope + unmarshal + validate + log)
	// -------------------------
	newSymbol := "2F-NEW" + randSuffix(4)
	newName := "2Finance New"

	metaOut, err := c.UpdateMetadata(
		tok.Address,
		newSymbol,
		newName,
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
	if len(metaOut.States) == 0 {
		t.Fatalf("UpdateMetadata returned empty States")
	}
	if metaOut.States[0].Object == nil {
		t.Fatalf("UpdateMetadata returned nil state object")
	}

	var meta tokenV1Domain.Token
	tests.UnmarshalState(t, metaOut.States[0].Object, &meta)

	if meta.Address != tok.Address {
		t.Fatalf("UpdateMetadata Address mismatch: got %s want %s", meta.Address, tok.Address)
	}
	if meta.Symbol != newSymbol {
		t.Fatalf("UpdateMetadata Symbol mismatch: got %s want %s", meta.Symbol, newSymbol)
	}
	if meta.Name != newName {
		t.Fatalf("UpdateMetadata Name mismatch: got %s want %s", meta.Name, newName)
	}
	if meta.Decimals != dec {
		t.Fatalf("UpdateMetadata Decimals mismatch: got %d want %d", meta.Decimals, dec)
	}

	log.Printf("UpdateMetadata Output States: %+v", metaOut.States)
	log.Printf("UpdateMetadata Output Logs: %+v", metaOut.Logs)
	log.Printf("UpdateMetadata Output Delegated Call: %+v", metaOut.DelegatedCall)

	log.Printf("UpdateMetadata TokenAddress: %s", meta.Address)
	log.Printf("UpdateMetadata Symbol: %s", meta.Symbol)
	log.Printf("UpdateMetadata Name: %s", meta.Name)
	log.Printf("UpdateMetadata Decimals: %d", meta.Decimals)
	log.Printf("UpdateMetadata Description: %s", meta.Description)
	log.Printf("UpdateMetadata Image: %s", meta.Image)
	log.Printf("UpdateMetadata Website: %s", meta.Website)
	log.Printf("UpdateMetadata TagsSocialMedia: %+v", meta.TagsSocialMedia)
	log.Printf("UpdateMetadata TagsCategory: %+v", meta.TagsCategory)
	log.Printf("UpdateMetadata Tags: %+v", meta.Tags)
	log.Printf("UpdateMetadata Creator: %s", meta.Creator)
	log.Printf("UpdateMetadata CreatorWebsite: %s", meta.CreatorWebsite)
	log.Printf("UpdateMetadata ExpiredAt: %s", meta.ExpiredAt)

	// -------------------------
	// Revoke Mint Authority (envelope + unmarshal + validate + log)
	// -------------------------
	revMintOut, err := c.RevokeMintAuthority(tok.Address, true)
	if err != nil {
		t.Fatalf("RevokeMintAuthority: %v", err)
	}
	if len(revMintOut.States) == 0 {
		t.Fatalf("RevokeMintAuthority returned empty States")
	}
	if revMintOut.States[0].Object == nil {
		t.Fatalf("RevokeMintAuthority returned nil state object")
	}

	var revMint tokenV1Domain.Token
	tests.UnmarshalState(t, revMintOut.States[0].Object, &revMint)

	if revMint.Address != tok.Address {
		t.Fatalf("RevokeMintAuthority Address mismatch: got %s want %s", revMint.Address, tok.Address)
	}
	if !revMint.MintAuthorityRevoked {
		t.Fatalf("RevokeMintAuthority expected MintAuthorityRevoked=true")
	}

	log.Printf("RevokeMintAuthority Output States: %+v", revMintOut.States)
	log.Printf("RevokeMintAuthority Output Logs: %+v", revMintOut.Logs)
	log.Printf("RevokeMintAuthority Output Delegated Call: %+v", revMintOut.DelegatedCall)

	log.Printf("RevokeMintAuthority TokenAddress: %s", revMint.Address)
	log.Printf("RevokeMintAuthority MintAuthorityRevoked: %v", revMint.MintAuthorityRevoked)

	// -------------------------
	// Revoke Update Authority (envelope + unmarshal + validate + log)
	// -------------------------
	revUpdOut, err := c.RevokeUpdateAuthority(tok.Address, true)
	if err != nil {
		t.Fatalf("RevokeUpdateAuthority: %v", err)
	}
	if len(revUpdOut.States) == 0 {
		t.Fatalf("RevokeUpdateAuthority returned empty States")
	}
	if revUpdOut.States[0].Object == nil {
		t.Fatalf("RevokeUpdateAuthority returned nil state object")
	}

	var revUpd tokenV1Domain.Token
	tests.UnmarshalState(t, revUpdOut.States[0].Object, &revUpd)

	if revUpd.Address != tok.Address {
		t.Fatalf("RevokeUpdateAuthority Address mismatch: got %s want %s", revUpd.Address, tok.Address)
	}
	if !revUpd.UpdateAuthorityRevoked {
		t.Fatalf("RevokeUpdateAuthority expected UpdateAuthorityRevoked=true")
	}

	log.Printf("RevokeUpdateAuthority Output States: %+v", revUpdOut.States)
	log.Printf("RevokeUpdateAuthority Output Logs: %+v", revUpdOut.Logs)
	log.Printf("RevokeUpdateAuthority Output Delegated Call: %+v", revUpdOut.DelegatedCall)

	log.Printf("RevokeUpdateAuthority TokenAddress: %s", revUpd.Address)
	log.Printf("RevokeUpdateAuthority UpdateAuthorityRevoked: %v", revUpd.UpdateAuthorityRevoked)

	// -------------------------
	// Pause (envelope + unmarshal + validate + log)
	// -------------------------
	pauseOut, err := c.PauseToken(tok.Address, true)
	if err != nil {
		t.Fatalf("PauseToken: %v", err)
	}
	if len(pauseOut.States) == 0 {
		t.Fatalf("PauseToken returned empty States")
	}
	if pauseOut.States[0].Object == nil {
		t.Fatalf("PauseToken returned nil state object")
	}

	var pause tokenV1Domain.Token
	tests.UnmarshalState(t, pauseOut.States[0].Object, &pause)

	if pause.Address != tok.Address {
		t.Fatalf("PauseToken Address mismatch: got %s want %s", pause.Address, tok.Address)
	}
	if !pause.Paused {
		t.Fatalf("PauseToken expected Paused=true")
	}

	log.Printf("PauseToken Output States: %+v", pauseOut.States)
	log.Printf("PauseToken Output Logs: %+v", pauseOut.Logs)
	log.Printf("PauseToken Output Delegated Call: %+v", pauseOut.DelegatedCall)

	log.Printf("PauseToken TokenAddress: %s", pause.Address)
	log.Printf("PauseToken Paused: %v", pause.Paused)

	// -------------------------
	// Unpause (envelope + unmarshal + validate + log)
	// -------------------------
	unpauseOut, err := c.UnpauseToken(tok.Address, false)
	if err != nil {
		t.Fatalf("UnpauseToken: %v", err)
	}
	if len(unpauseOut.States) == 0 {
		t.Fatalf("UnpauseToken returned empty States")
	}
	if unpauseOut.States[0].Object == nil {
		t.Fatalf("UnpauseToken returned nil state object")
	}

	var unpause tokenV1Domain.Token
	tests.UnmarshalState(t, unpauseOut.States[0].Object, &unpause)

	if unpause.Address != tok.Address {
		t.Fatalf("UnpauseToken Address mismatch: got %s want %s", unpause.Address, tok.Address)
	}
	if unpause.Paused {
		t.Fatalf("UnpauseToken expected Paused=false")
	}

	log.Printf("UnpauseToken Output States: %+v", unpauseOut.States)
	log.Printf("UnpauseToken Output Logs: %+v", unpauseOut.Logs)
	log.Printf("UnpauseToken Output Delegated Call: %+v", unpauseOut.DelegatedCall)

	log.Printf("UnpauseToken TokenAddress: %s", unpause.Address)
	log.Printf("UnpauseToken Paused: %v", unpause.Paused)

	// -------------------------
	// Freeze wallet (envelope + unmarshal + validate + log)
	// -------------------------
	freezeOut, err := c.FreezeWallet(tok.Address, owner.PublicKey)
	if err != nil {
		t.Fatalf("FreezeWallet: %v", err)
	}
	if len(freezeOut.States) == 0 {
		t.Fatalf("FreezeWallet returned empty States")
	}
	if freezeOut.States[0].Object == nil {
		t.Fatalf("FreezeWallet returned nil state object")
	}

	var freeze tokenV1Domain.Token
	tests.UnmarshalState(t, freezeOut.States[0].Object, &freeze)

	if freeze.Address != tok.Address {
		t.Fatalf("FreezeWallet Address mismatch: got %s want %s", freeze.Address, tok.Address)
	}
	if freeze.Owner == "" {
		t.Fatalf("FreezeWallet Owner empty")
	}
	if freeze.FrozenAccounts == nil {
		t.Fatalf("FreezeWallet FrozenAccounts nil")
	}
	if !freeze.FrozenAccounts[owner.PublicKey] {
		t.Fatalf("FreezeWallet expected owner to be frozen: %s", owner.PublicKey)
	}

	log.Printf("FreezeWallet Output States: %+v", freezeOut.States)
	log.Printf("FreezeWallet Output Logs: %+v", freezeOut.Logs)
	log.Printf("FreezeWallet Output Delegated Call: %+v", freezeOut.DelegatedCall)

	log.Printf("FreezeWallet TokenAddress: %s", freeze.Address)
	log.Printf("FreezeWallet Wallet: %s", freeze.Owner)
	log.Printf("FreezeWallet Frozen Accounts: %v", freeze.FrozenAccounts)

	// -------------------------
	// Unfreeze wallet (envelope + unmarshal + validate + log)
	// -------------------------
	unfreezeOut, err := c.UnfreezeWallet(tok.Address, owner.PublicKey)
	if err != nil {
		t.Fatalf("UnfreezeWallet: %v", err)
	}
	if len(unfreezeOut.States) == 0 {
		t.Fatalf("UnfreezeWallet returned empty States")
	}
	if unfreezeOut.States[0].Object == nil {
		t.Fatalf("UnfreezeWallet returned nil state object")
	}

	var unfreeze tokenV1Domain.Token
	tests.UnmarshalState(t, unfreezeOut.States[0].Object, &unfreeze)

	if unfreeze.Address != tok.Address {
		t.Fatalf("UnfreezeWallet Address mismatch: got %s want %s", unfreeze.Address, tok.Address)
	}
	if unfreeze.FrozenAccounts == nil {
		t.Fatalf("UnfreezeWallet FrozenAccounts nil")
	}
	if unfreeze.FrozenAccounts[owner.PublicKey] {
		t.Fatalf("UnfreezeWallet expected owner to be unfrozen: %s", owner.PublicKey)
	}

	log.Printf("UnfreezeWallet Output States: %+v", unfreezeOut.States)
	log.Printf("UnfreezeWallet Output Logs: %+v", unfreezeOut.Logs)
	log.Printf("UnfreezeWallet Output Delegated Call: %+v", unfreezeOut.DelegatedCall)

	log.Printf("UnfreezeWallet TokenAddress: %s", unfreeze.Address)
	log.Printf("UnfreezeWallet Wallet: %s", unfreeze.Owner)
	log.Printf("UnfreezeWallet Frozen Accounts: %v", unfreeze.FrozenAccounts)

	// -------------------------
	// GetTokenBalance (envelope + unmarshal + validate + log)
	// -------------------------
	getBalOut, err := c.GetTokenBalance(tok.Address, owner.PublicKey)
	if err != nil {
		t.Fatalf("GetTokenBalance(owner): %v", err)
	}
	if len(getBalOut.States) == 0 {
		t.Fatalf("GetTokenBalance returned empty States")
	}
	if getBalOut.States[0].Object == nil {
		t.Fatalf("GetTokenBalance returned nil state object")
	}

	var bal tokenV1Domain.Balance
	tests.UnmarshalState(t, getBalOut.States[0].Object, &bal)

	if bal.TokenAddress != tok.Address {
		t.Fatalf("GetTokenBalance TokenAddress mismatch: got %s want %s", bal.TokenAddress, tok.Address)
	}
	if bal.OwnerAddress != owner.PublicKey {
		t.Fatalf("GetTokenBalance OwnerAddress mismatch: got %s want %s", bal.OwnerAddress, owner.PublicKey)
	}
	if bal.Amount == "" {
		t.Fatalf("GetTokenBalance Amount empty")
	}

	log.Printf("GetTokenBalance Output States: %+v", getBalOut.States)
	log.Printf("GetTokenBalance Output Logs: %+v", getBalOut.Logs)
	log.Printf("GetTokenBalance Output Delegated Call: %+v", getBalOut.DelegatedCall)

	log.Printf("GetTokenBalance TokenAddress: %s", bal.TokenAddress)
	log.Printf("GetTokenBalance OwnerAddress: %s", bal.OwnerAddress)
	log.Printf("GetTokenBalance Amount: %s", bal.Amount)

	// -------------------------
	// ListTokenBalances (envelope + unmarshal + validate + log)
	// -------------------------
	listBalOut, err := c.ListTokenBalances(tok.Address, "", 1, 10, true)
	if err != nil {
		t.Fatalf("ListTokenBalances: %v", err)
	}
	if len(listBalOut.States) == 0 {
		t.Fatalf("ListTokenBalances returned empty States")
	}
	if listBalOut.States[0].Object == nil {
		t.Fatalf("ListTokenBalances returned nil state object")
	}

	var balList []tokenV1Domain.Balance
	tests.UnmarshalState(t, listBalOut.States[0].Object, &balList)

	if len(balList) == 0 {
		t.Fatalf("ListTokenBalances returned empty list")
	}

	log.Printf("ListTokenBalances Output States: %+v", listBalOut.States)
	log.Printf("ListTokenBalances Output Logs: %+v", listBalOut.Logs)
	log.Printf("ListTokenBalances Output Delegated Call: %+v", listBalOut.DelegatedCall)

	log.Printf("Balance Listed Successfully: %+v", balList)

	// -------------------------
	// GetToken (envelope + unmarshal + validate + log)
	// -------------------------
	getTokOut, err := c.GetToken(tok.Address, "", "")
	if err != nil {
		t.Fatalf("GetToken: %v", err)
	}
	if len(getTokOut.States) == 0 {
		t.Fatalf("GetToken returned empty States")
	}
	if getTokOut.States[0].Object == nil {
		t.Fatalf("GetToken returned nil state object")
	}

	var got tokenV1Domain.Token
	tests.UnmarshalState(t, getTokOut.States[0].Object, &got)

	if got.Address != tok.Address {
		t.Fatalf("GetToken Address mismatch: got %s want %s", got.Address, tok.Address)
	}
	if got.Symbol == "" || got.Name == "" {
		t.Fatalf("GetToken Symbol/Name empty: symbol=%q name=%q", got.Symbol, got.Name)
	}
	if got.TokenType != tok.TokenType {
		t.Fatalf("GetToken TokenType mismatch: got %s want %s", got.TokenType, tok.TokenType)
	}

	log.Printf("GetToken Output States: %+v", getTokOut.States)
	log.Printf("GetToken Output Logs: %+v", getTokOut.Logs)
	log.Printf("GetToken Output Delegated Call: %+v", getTokOut.DelegatedCall)

	log.Printf("GetToken Address: %s", got.Address)
	log.Printf("GetToken Symbol: %s", got.Symbol)
	log.Printf("GetToken Name: %s", got.Name)
	log.Printf("GetToken TotalSupply: %s", got.TotalSupply)
	log.Printf("GetToken FeeAddress: %s", got.FeeAddress)
	log.Printf("GetToken Paused: %v", got.Paused)

	// -------------------------
	// ListTokens (envelope + unmarshal + validate + log)
	// -------------------------
	listTokOut, err := c.ListTokens("", "", "", 1, 10, true)
	if err != nil {
		t.Fatalf("ListTokens: %v", err)
	}
	if len(listTokOut.States) == 0 {
		t.Fatalf("ListTokens returned empty States")
	}
	if listTokOut.States[0].Object == nil {
		t.Fatalf("ListTokens returned nil state object")
	}

	var tokList []tokenV1Domain.Token
	tests.UnmarshalState(t, listTokOut.States[0].Object, &tokList)

	if len(tokList) == 0 {
		t.Fatalf("ListTokens returned empty list")
	}

	log.Printf("ListTokens Output States: %+v", listTokOut.States)
	log.Printf("ListTokens Output Logs: %+v", listTokOut.Logs)
	log.Printf("ListTokens Output Delegated Call: %+v", listTokOut.DelegatedCall)

	log.Printf("Token Listed Successfully: %+v", tokList)
}

func TestTokenFlowNonFungible(t *testing.T) {
	c := setupClient(t)
	owner, ownerPriv := createWallet(t, c)
	c.SetPrivateKey(ownerPriv)

	dec := 0
	tokenType := tokenV1Domain.NON_FUNGIBLE
	stablecoin := false

	tok := createBasicToken(t, c, owner.PublicKey, dec, false, tokenType, stablecoin)

	// -------------------------
	// Token (validate + log)
	// -------------------------
	if tok.Address == "" {
		t.Fatalf("token address empty")
	}
	if tok.Symbol == "" {
		t.Fatalf("token symbol empty")
	}
	if tok.Name == "" {
		t.Fatalf("token name empty")
	}
	if tok.Decimals != dec {
		t.Fatalf("token decimals mismatch: got %d want %d", tok.Decimals, dec)
	}
	if tok.TokenType != tokenType {
		t.Fatalf("token type mismatch: got %s want %s", tok.TokenType, tokenType)
	}
	if tok.Stablecoin != stablecoin {
		t.Fatalf("token stablecoin mismatch: got %v want %v", tok.Stablecoin, stablecoin)
	}
	if tok.Creator == "" {
		t.Fatalf("token creator empty")
	}
	if tok.AccessPolicy.Mode == "" {
		t.Fatalf("token access policy mode empty")
	}
	if tok.AccessPolicy.Users == nil {
		t.Fatalf("token access policy users nil")
	}
	if !tok.AccessPolicy.Users[owner.PublicKey] {
		t.Fatalf("token access policy must include owner: %s", owner.PublicKey)
	}

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

	// -------------------------
	// Mint NFT (envelope + unmarshal + validate + log)
	// -------------------------
	mintOut, err := c.MintToken(tok.Address, owner.PublicKey, amt(35, dec), dec, tok.TokenType)
	if err != nil {
		t.Fatalf("MintToken NFT: %v", err)
	}
	if len(mintOut.States) == 0 {
		t.Fatalf("MintToken returned empty States")
	}
	if mintOut.States[0].Object == nil {
		t.Fatalf("MintToken returned nil state object")
	}

	var mint tokenV1Domain.Mint
	tests.UnmarshalState(t, mintOut.States[0].Object, &mint)

	if mint.TokenAddress != tok.Address {
		t.Fatalf("Mint TokenAddress mismatch: got %s want %s", mint.TokenAddress, tok.Address)
	}
	if mint.MintTo != owner.PublicKey {
		t.Fatalf("Mint ToAddress mismatch: got %s want %s", mint.MintTo, owner.PublicKey)
	}
	expectedMintAmount := amt(35, dec) // with dec=0, should be "35"
	if mint.Amount != expectedMintAmount {
		t.Fatalf("Mint Amount mismatch: got %s want %s", mint.Amount, expectedMintAmount)
	}
	if mint.TokenType != tok.TokenType {
		t.Fatalf("Mint TokenType mismatch: got %s want %s", mint.TokenType, tok.TokenType)
	}
	if len(mint.TokenUUIDList) != 35 {
		t.Fatalf("expected %d uuid, got %d", 35, len(mint.TokenUUIDList))
	}
	// sanity: UUIDs not empty
	for i, u := range mint.TokenUUIDList {
		if u == "" {
			t.Fatalf("mint uuid[%d] empty", i)
		}
	}

	log.Printf("Mint Output States: %+v", mintOut.States)
	log.Printf("Mint Output Logs: %+v", mintOut.Logs)
	log.Printf("Mint Output Delegated Call: %+v", mintOut.DelegatedCall)

	log.Printf("Mint TokenAddress: %s", mint.TokenAddress)
	log.Printf("Mint ToAddress: %s", mint.MintTo)
	log.Printf("Mint Amount: %s", mint.Amount)
	log.Printf("Mint TokenType: %s", mint.TokenType)
	log.Printf("Mint TokenUUIDList: %+v", mint.TokenUUIDList)

	// -------------------------
	// Burn 1 NFT (envelope + unmarshal + validate + log)
	// -------------------------
	burnUUID := mint.TokenUUIDList[0]
	burnOut, err := c.BurnToken(
		tok.Address,
		amt(1, dec),
		dec,
		tok.TokenType,
		burnUUID,
	)
	if err != nil {
		t.Fatalf("BurnToken: %v", err)
	}
	if len(burnOut.States) == 0 {
		t.Fatalf("BurnToken returned empty States")
	}
	if burnOut.States[0].Object == nil {
		t.Fatalf("BurnToken returned nil state object")
	}

	var burn tokenV1Domain.Burn
	tests.UnmarshalState(t, burnOut.States[0].Object, &burn)

	if burn.TokenAddress != tok.Address {
		t.Fatalf("Burn TokenAddress mismatch: got %s want %s", burn.TokenAddress, tok.Address)
	}
	if burn.BurnFrom != owner.PublicKey {
		t.Fatalf("Burn FromAddress mismatch: got %s want %s", burn.BurnFrom, owner.PublicKey)
	}
	expectedBurnAmount := amt(1, dec) // "1"
	if burn.Amount != expectedBurnAmount {
		t.Fatalf("Burn Amount mismatch: got %s want %s", burn.Amount, expectedBurnAmount)
	}
	if burn.TokenType != tok.TokenType {
		t.Fatalf("Burn TokenType mismatch: got %s want %s", burn.TokenType, tok.TokenType)
	}
	if burn.UUID != burnUUID {
		t.Fatalf("Burn UUID mismatch: got %q want %q", burn.UUID, burnUUID)
	}

	log.Printf("Burn Output States: %+v", burnOut.States)
	log.Printf("Burn Output Logs: %+v", burnOut.Logs)
	log.Printf("Burn Output Delegated Call: %+v", burnOut.DelegatedCall)

	log.Printf("Burn TokenAddress: %s", burn.TokenAddress)
	log.Printf("Burn FromAddress: %s", burn.BurnFrom)
	log.Printf("Burn Amount: %s", burn.Amount)
	log.Printf("Burn TokenType: %s", burn.TokenType)
	log.Printf("Burn UUID: %s", burn.UUID)

	// -------------------------
	// AllowUsers (envelope + unmarshal + validate + log)
	// -------------------------
	receiver, _ := createWallet(t, c)
	c.SetPrivateKey(ownerPriv)

	allowOut, err := c.AllowUsers(tok.Address, map[string]bool{
		receiver.PublicKey: true,
	})
	if err != nil {
		t.Fatalf("AllowUsers: %v", err)
	}
	if len(allowOut.States) == 0 {
		t.Fatalf("AllowUsers returned empty States")
	}
	if allowOut.States[0].Object == nil {
		t.Fatalf("AllowUsers returned nil state object")
	}

	var accessPolicy tokenV1Domain.AccessPolicy
	tests.UnmarshalState(t, allowOut.States[0].Object, &accessPolicy)

	if accessPolicy.Mode == "" {
		t.Fatalf("AllowUsers Mode empty")
	}
	if accessPolicy.Users == nil {
		t.Fatalf("AllowUsers Users nil")
	}
	if !accessPolicy.Users[receiver.PublicKey] {
		t.Fatalf("AllowUsers missing receiver in allowlist: %s", receiver.PublicKey)
	}

	log.Printf("AllowUsers Output States: %+v", allowOut.States)
	log.Printf("AllowUsers Output Logs: %+v", allowOut.Logs)
	log.Printf("AllowUsers Output Delegated Call: %+v", allowOut.DelegatedCall)

	log.Printf("AllowUsers Mode: %s", accessPolicy.Mode)
	log.Printf("AllowUsers Users: %+v", accessPolicy.Users)

	// -------------------------
	// Transfer NFT (envelope + unmarshal + validate + log)
	// -------------------------
	transferUUID := mint.TokenUUIDList[1] // uuid that was not burned
	if transferUUID == burnUUID {
		t.Fatalf("transferUUID equals burned UUID, test data invalid")
	}
	if transferUUID == "" {
		t.Fatalf("transferUUID empty")
	}

	trOut, err := c.TransferToken(
		tok.Address,
		receiver.PublicKey,
		amt(1, dec),
		dec,
		tok.TokenType,
		transferUUID,
	)
	if err != nil {
		t.Fatalf("Transfer NFT: %v", err)
	}
	if len(trOut.States) == 0 {
		t.Fatalf("TransferToken returned empty States")
	}
	if trOut.States[0].Object == nil {
		t.Fatalf("TransferToken returned nil state object")
	}

	var tr tokenV1Domain.Transfer
	tests.UnmarshalState(t, trOut.States[0].Object, &tr)

	if tr.TokenAddress != tok.Address {
		t.Fatalf("Transfer TokenAddress mismatch: got %s want %s", tr.TokenAddress, tok.Address)
	}
	if tr.FromAddress != owner.PublicKey {
		t.Fatalf("Transfer FromAddress mismatch: got %s want %s", tr.FromAddress, owner.PublicKey)
	}
	if tr.ToAddress != receiver.PublicKey {
		t.Fatalf("Transfer ToAddress mismatch: got %s want %s", tr.ToAddress, receiver.PublicKey)
	}
	expectedTransferAmount := amt(1, dec) // "1"
	if tr.Amount != expectedTransferAmount {
		t.Fatalf("Transfer Amount mismatch: got %s want %s", tr.Amount, expectedTransferAmount)
	}
	if tr.TokenType != tok.TokenType {
		t.Fatalf("Transfer TokenType mismatch: got %s want %s", tr.TokenType, tok.TokenType)
	}
	if tr.UUID != transferUUID {
		t.Fatalf("Transfer UUID mismatch: got %q want %q", tr.UUID, transferUUID)
	}

	log.Printf("Transfer Output States: %+v", trOut.States)
	log.Printf("Transfer Output Logs: %+v", trOut.Logs)
	log.Printf("Transfer Output Delegated Call: %+v", trOut.DelegatedCall)

	log.Printf("Transfer TokenAddress: %s", tr.TokenAddress)
	log.Printf("Transfer FromAddress: %s", tr.FromAddress)
	log.Printf("Transfer ToAddress: %s", tr.ToAddress)
	log.Printf("Transfer Amount: %s", tr.Amount)
	log.Printf("Transfer TokenType: %s", tr.TokenType)
	log.Printf("Transfer UUID: %s", tr.UUID)

	// -------------------------
	// Fee tiers (envelope + unmarshal + validate + log)
	// -------------------------
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
	if len(feeTiersOut.States) == 0 {
		t.Fatalf("UpdateFeeTiers returned empty States")
	}
	if feeTiersOut.States[0].Object == nil {
		t.Fatalf("UpdateFeeTiers returned nil state object")
	}

	var feeTiers tokenV1Domain.FeeTiers
	tests.UnmarshalState(t, feeTiersOut.States[0].Object, &feeTiers)

	if feeTiers.FeeTiersList == nil || len(feeTiers.FeeTiersList) == 0 {
		t.Fatalf("UpdateFeeTiers returned empty FeeTiersList")
	}

	log.Printf("UpdateFeeTiers Output States: %+v", feeTiersOut.States)
	log.Printf("UpdateFeeTiers Output Logs: %+v", feeTiersOut.Logs)
	log.Printf("UpdateFeeTiers Output Delegated Call: %+v", feeTiersOut.DelegatedCall)

	log.Printf("UpdateFeeTiers FeeTiersList: %+v", feeTiers.FeeTiersList)

	// -------------------------
	// Fee address (envelope + unmarshal + validate + log)
	// -------------------------
	feeOut, err := c.UpdateFeeAddress(tok.Address, owner.PublicKey)
	if err != nil {
		t.Fatalf("UpdateFeeAddress: %v", err)
	}
	if len(feeOut.States) == 0 {
		t.Fatalf("UpdateFeeAddress returned empty States")
	}
	if feeOut.States[0].Object == nil {
		t.Fatalf("UpdateFeeAddress returned nil state object")
	}

	var fee tokenV1Domain.Fee
	tests.UnmarshalState(t, feeOut.States[0].Object, &fee)

	if fee.TokenAddress != tok.Address {
		t.Fatalf("UpdateFeeAddress TokenAddress mismatch: got %s want %s", fee.TokenAddress, tok.Address)
	}
	if fee.FeeAddress != owner.PublicKey {
		t.Fatalf("UpdateFeeAddress FeeAddress mismatch: got %s want %s", fee.FeeAddress, owner.PublicKey)
	}

	log.Printf("UpdateFeeAddress Output States: %+v", feeOut.States)
	log.Printf("UpdateFeeAddress Output Logs: %+v", feeOut.Logs)
	log.Printf("UpdateFeeAddress Output Delegated Call: %+v", feeOut.DelegatedCall)

	log.Printf("UpdateFeeAddress TokenAddress: %s", fee.TokenAddress)
	log.Printf("UpdateFeeAddress FeeAddress: %s", fee.FeeAddress)

	// -------------------------
	// Metadata (envelope + unmarshal + validate + log)
	// -------------------------
	newSymbol := "2F-NEW" + randSuffix(4)
	newName := "2Finance New"

	metaOut, err := c.UpdateMetadata(
		tok.Address,
		newSymbol,
		newName,
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
	if len(metaOut.States) == 0 {
		t.Fatalf("UpdateMetadata returned empty States")
	}
	if metaOut.States[0].Object == nil {
		t.Fatalf("UpdateMetadata returned nil state object")
	}

	var meta tokenV1Domain.Token
	tests.UnmarshalState(t, metaOut.States[0].Object, &meta)

	if meta.Address != tok.Address {
		t.Fatalf("UpdateMetadata Address mismatch: got %s want %s", meta.Address, tok.Address)
	}
	if meta.Symbol != newSymbol {
		t.Fatalf("UpdateMetadata Symbol mismatch: got %s want %s", meta.Symbol, newSymbol)
	}
	if meta.Name != newName {
		t.Fatalf("UpdateMetadata Name mismatch: got %s want %s", meta.Name, newName)
	}
	if meta.Decimals != dec {
		t.Fatalf("UpdateMetadata Decimals mismatch: got %d want %d", meta.Decimals, dec)
	}

	log.Printf("UpdateMetadata Output States: %+v", metaOut.States)
	log.Printf("UpdateMetadata Output Logs: %+v", metaOut.Logs)
	log.Printf("UpdateMetadata Output Delegated Call: %+v", metaOut.DelegatedCall)

	log.Printf("UpdateMetadata TokenAddress: %s", meta.Address)
	log.Printf("UpdateMetadata Symbol: %s", meta.Symbol)
	log.Printf("UpdateMetadata Name: %s", meta.Name)
	log.Printf("UpdateMetadata Decimals: %d", meta.Decimals)
	log.Printf("UpdateMetadata Description: %s", meta.Description)
	log.Printf("UpdateMetadata Image: %s", meta.Image)
	log.Printf("UpdateMetadata Website: %s", meta.Website)
	log.Printf("UpdateMetadata TagsSocialMedia: %+v", meta.TagsSocialMedia)
	log.Printf("UpdateMetadata TagsCategory: %+v", meta.TagsCategory)
	log.Printf("UpdateMetadata Tags: %+v", meta.Tags)
	log.Printf("UpdateMetadata Creator: %s", meta.Creator)
	log.Printf("UpdateMetadata CreatorWebsite: %s", meta.CreatorWebsite)
	log.Printf("UpdateMetadata ExpiredAt: %s", meta.ExpiredAt)

	// -------------------------
	// Revoke Mint Authority (envelope + unmarshal + validate + log)
	// -------------------------
	revMintOut, err := c.RevokeMintAuthority(tok.Address, true)
	if err != nil {
		t.Fatalf("RevokeMintAuthority: %v", err)
	}
	if len(revMintOut.States) == 0 {
		t.Fatalf("RevokeMintAuthority returned empty States")
	}
	if revMintOut.States[0].Object == nil {
		t.Fatalf("RevokeMintAuthority returned nil state object")
	}

	var revMint tokenV1Domain.Token
	tests.UnmarshalState(t, revMintOut.States[0].Object, &revMint)

	if revMint.Address != tok.Address {
		t.Fatalf("RevokeMintAuthority Address mismatch: got %s want %s", revMint.Address, tok.Address)
	}
	if !revMint.MintAuthorityRevoked {
		t.Fatalf("RevokeMintAuthority expected MintAuthorityRevoked=true")
	}

	log.Printf("RevokeMintAuthority Output States: %+v", revMintOut.States)
	log.Printf("RevokeMintAuthority Output Logs: %+v", revMintOut.Logs)
	log.Printf("RevokeMintAuthority Output Delegated Call: %+v", revMintOut.DelegatedCall)

	log.Printf("RevokeMintAuthority TokenAddress: %s", revMint.Address)
	log.Printf("RevokeMintAuthority MintAuthorityRevoked: %v", revMint.MintAuthorityRevoked)

	// -------------------------
	// Revoke Update Authority (envelope + unmarshal + validate + log)
	// -------------------------
	revUpdOut, err := c.RevokeUpdateAuthority(tok.Address, true)
	if err != nil {
		t.Fatalf("RevokeUpdateAuthority: %v", err)
	}
	if len(revUpdOut.States) == 0 {
		t.Fatalf("RevokeUpdateAuthority returned empty States")
	}
	if revUpdOut.States[0].Object == nil {
		t.Fatalf("RevokeUpdateAuthority returned nil state object")
	}

	var revUpd tokenV1Domain.Token
	tests.UnmarshalState(t, revUpdOut.States[0].Object, &revUpd)

	if revUpd.Address != tok.Address {
		t.Fatalf("RevokeUpdateAuthority Address mismatch: got %s want %s", revUpd.Address, tok.Address)
	}
	if !revUpd.UpdateAuthorityRevoked {
		t.Fatalf("RevokeUpdateAuthority expected UpdateAuthorityRevoked=true")
	}

	log.Printf("RevokeUpdateAuthority Output States: %+v", revUpdOut.States)
	log.Printf("RevokeUpdateAuthority Output Logs: %+v", revUpdOut.Logs)
	log.Printf("RevokeUpdateAuthority Output Delegated Call: %+v", revUpdOut.DelegatedCall)

	log.Printf("RevokeUpdateAuthority TokenAddress: %s", revUpd.Address)
	log.Printf("RevokeUpdateAuthority UpdateAuthorityRevoked: %v", revUpd.UpdateAuthorityRevoked)

	// -------------------------
	// Pause (envelope + unmarshal + validate + log)
	// -------------------------
	pauseOut, err := c.PauseToken(tok.Address, true)
	if err != nil {
		t.Fatalf("PauseToken: %v", err)
	}
	if len(pauseOut.States) == 0 {
		t.Fatalf("PauseToken returned empty States")
	}
	if pauseOut.States[0].Object == nil {
		t.Fatalf("PauseToken returned nil state object")
	}

	var pause tokenV1Domain.Token
	tests.UnmarshalState(t, pauseOut.States[0].Object, &pause)

	if pause.Address != tok.Address {
		t.Fatalf("PauseToken Address mismatch: got %s want %s", pause.Address, tok.Address)
	}
	if !pause.Paused {
		t.Fatalf("PauseToken expected Paused=true")
	}

	log.Printf("PauseToken Output States: %+v", pauseOut.States)
	log.Printf("PauseToken Output Logs: %+v", pauseOut.Logs)
	log.Printf("PauseToken Output Delegated Call: %+v", pauseOut.DelegatedCall)

	log.Printf("PauseToken TokenAddress: %s", pause.Address)
	log.Printf("PauseToken Paused: %v", pause.Paused)

	// -------------------------
	// Unpause (envelope + unmarshal + validate + log)
	// -------------------------
	unpauseOut, err := c.UnpauseToken(tok.Address, false)
	if err != nil {
		t.Fatalf("UnpauseToken: %v", err)
	}
	if len(unpauseOut.States) == 0 {
		t.Fatalf("UnpauseToken returned empty States")
	}
	if unpauseOut.States[0].Object == nil {
		t.Fatalf("UnpauseToken returned nil state object")
	}

	var unpause tokenV1Domain.Token
	tests.UnmarshalState(t, unpauseOut.States[0].Object, &unpause)

	if unpause.Address != tok.Address {
		t.Fatalf("UnpauseToken Address mismatch: got %s want %s", unpause.Address, tok.Address)
	}
	if unpause.Paused {
		t.Fatalf("UnpauseToken expected Paused=false")
	}

	log.Printf("UnpauseToken Output States: %+v", unpauseOut.States)
	log.Printf("UnpauseToken Output Logs: %+v", unpauseOut.Logs)
	log.Printf("UnpauseToken Output Delegated Call: %+v", unpauseOut.DelegatedCall)

	log.Printf("UnpauseToken TokenAddress: %s", unpause.Address)
	log.Printf("UnpauseToken Paused: %v", unpause.Paused)

	// -------------------------
	// Freeze wallet (envelope + unmarshal + validate + log)
	// -------------------------
	freezeOut, err := c.FreezeWallet(tok.Address, owner.PublicKey)
	if err != nil {
		t.Fatalf("FreezeWallet: %v", err)
	}
	if len(freezeOut.States) == 0 {
		t.Fatalf("FreezeWallet returned empty States")
	}
	if freezeOut.States[0].Object == nil {
		t.Fatalf("FreezeWallet returned nil state object")
	}

	var freeze tokenV1Domain.Token
	tests.UnmarshalState(t, freezeOut.States[0].Object, &freeze)

	if freeze.Address != tok.Address {
		t.Fatalf("FreezeWallet Address mismatch: got %s want %s", freeze.Address, tok.Address)
	}
	if freeze.Owner == "" {
		t.Fatalf("FreezeWallet Owner empty")
	}
	if freeze.FrozenAccounts == nil {
		t.Fatalf("FreezeWallet FrozenAccounts nil")
	}
	if !freeze.FrozenAccounts[owner.PublicKey] {
		t.Fatalf("FreezeWallet expected owner to be frozen: %s", owner.PublicKey)
	}

	log.Printf("FreezeWallet Output States: %+v", freezeOut.States)
	log.Printf("FreezeWallet Output Logs: %+v", freezeOut.Logs)
	log.Printf("FreezeWallet Output Delegated Call: %+v", freezeOut.DelegatedCall)

	log.Printf("FreezeWallet TokenAddress: %s", freeze.Address)
	log.Printf("FreezeWallet Wallet: %s", freeze.Owner)
	log.Printf("FreezeWallet Frozen Accounts: %v", freeze.FrozenAccounts)

	// -------------------------
	// Unfreeze wallet (envelope + unmarshal + validate + log)
	// -------------------------
	unfreezeOut, err := c.UnfreezeWallet(tok.Address, owner.PublicKey)
	if err != nil {
		t.Fatalf("UnfreezeWallet: %v", err)
	}
	if len(unfreezeOut.States) == 0 {
		t.Fatalf("UnfreezeWallet returned empty States")
	}
	if unfreezeOut.States[0].Object == nil {
		t.Fatalf("UnfreezeWallet returned nil state object")
	}

	var unfreeze tokenV1Domain.Token
	tests.UnmarshalState(t, unfreezeOut.States[0].Object, &unfreeze)

	if unfreeze.Address != tok.Address {
		t.Fatalf("UnfreezeWallet Address mismatch: got %s want %s", unfreeze.Address, tok.Address)
	}
	if unfreeze.FrozenAccounts == nil {
		t.Fatalf("UnfreezeWallet FrozenAccounts nil")
	}
	if unfreeze.FrozenAccounts[owner.PublicKey] {
		t.Fatalf("UnfreezeWallet expected owner to be unfrozen: %s", owner.PublicKey)
	}

	log.Printf("UnfreezeWallet Output States: %+v", unfreezeOut.States)
	log.Printf("UnfreezeWallet Output Logs: %+v", unfreezeOut.Logs)
	log.Printf("UnfreezeWallet Output Delegated Call: %+v", unfreezeOut.DelegatedCall)

	log.Printf("UnfreezeWallet TokenAddress: %s", unfreeze.Address)
	log.Printf("UnfreezeWallet Wallet: %s", unfreeze.Owner)
	log.Printf("UnfreezeWallet Frozen Accounts: %v", unfreeze.FrozenAccounts)

	// -------------------------
	// GetTokenBalance (envelope + unmarshal + validate + log)
	// -------------------------
	getBalOut, err := c.GetTokenBalance(tok.Address, owner.PublicKey)
	if err != nil {
		t.Fatalf("GetTokenBalance(owner): %v", err)
	}
	if len(getBalOut.States) == 0 {
		t.Fatalf("GetTokenBalance returned empty States")
	}
	if getBalOut.States[0].Object == nil {
		t.Fatalf("GetTokenBalance returned nil state object")
	}

	var bal tokenV1Domain.Balance
	tests.UnmarshalState(t, getBalOut.States[0].Object, &bal)

	if bal.TokenAddress != tok.Address {
		t.Fatalf("GetTokenBalance TokenAddress mismatch: got %s want %s", bal.TokenAddress, tok.Address)
	}
	if bal.OwnerAddress != owner.PublicKey {
		t.Fatalf("GetTokenBalance OwnerAddress mismatch: got %s want %s", bal.OwnerAddress, owner.PublicKey)
	}
	if bal.Amount == "" {
		t.Fatalf("GetTokenBalance Amount empty")
	}

	log.Printf("GetTokenBalance Output States: %+v", getBalOut.States)
	log.Printf("GetTokenBalance Output Logs: %+v", getBalOut.Logs)
	log.Printf("GetTokenBalance Output Delegated Call: %+v", getBalOut.DelegatedCall)

	log.Printf("GetTokenBalance TokenAddress: %s", bal.TokenAddress)
	log.Printf("GetTokenBalance OwnerAddress: %s", bal.OwnerAddress)
	log.Printf("GetTokenBalance Amount: %s", bal.Amount)

	// -------------------------
	// ListTokenBalances (envelope + unmarshal + validate + log)
	// -------------------------
	listBalOut, err := c.ListTokenBalances(tok.Address, "", 1, 10, true)
	if err != nil {
		t.Fatalf("ListTokenBalances: %v", err)
	}
	if len(listBalOut.States) == 0 {
		t.Fatalf("ListTokenBalances returned empty States")
	}
	if listBalOut.States[0].Object == nil {
		t.Fatalf("ListTokenBalances returned nil state object")
	}

	var balList []tokenV1Domain.Balance
	tests.UnmarshalState(t, listBalOut.States[0].Object, &balList)

	if len(balList) == 0 {
		t.Fatalf("ListTokenBalances returned empty list")
	}

	log.Printf("ListTokenBalances Output States: %+v", listBalOut.States)
	log.Printf("ListTokenBalances Output Logs: %+v", listBalOut.Logs)
	log.Printf("ListTokenBalances Output Delegated Call: %+v", listBalOut.DelegatedCall)

	log.Printf("Balance Listed Successfully: %+v", balList)

	// -------------------------
	// GetToken (envelope + unmarshal + validate + log)
	// -------------------------
	getTokOut, err := c.GetToken(tok.Address, "", "")
	if err != nil {
		t.Fatalf("GetToken: %v", err)
	}
	if len(getTokOut.States) == 0 {
		t.Fatalf("GetToken returned empty States")
	}
	if getTokOut.States[0].Object == nil {
		t.Fatalf("GetToken returned nil state object")
	}

	var got tokenV1Domain.Token
	tests.UnmarshalState(t, getTokOut.States[0].Object, &got)

	if got.Address != tok.Address {
		t.Fatalf("GetToken Address mismatch: got %s want %s", got.Address, tok.Address)
	}
	if got.Symbol == "" || got.Name == "" {
		t.Fatalf("GetToken Symbol/Name empty: symbol=%q name=%q", got.Symbol, got.Name)
	}
	if got.TokenType != tok.TokenType {
		t.Fatalf("GetToken TokenType mismatch: got %s want %s", got.TokenType, tok.TokenType)
	}

	log.Printf("GetToken Output States: %+v", getTokOut.States)
	log.Printf("GetToken Output Logs: %+v", getTokOut.Logs)
	log.Printf("GetToken Output Delegated Call: %+v", getTokOut.DelegatedCall)

	log.Printf("GetToken Address: %s", got.Address)
	log.Printf("GetToken Symbol: %s", got.Symbol)
	log.Printf("GetToken Name: %s", got.Name)
	log.Printf("GetToken TotalSupply: %s", got.TotalSupply)
	log.Printf("GetToken FeeAddress: %s", got.FeeAddress)
	log.Printf("GetToken Paused: %v", got.Paused)

	// -------------------------
	// ListTokens (envelope + unmarshal + validate + log)
	// -------------------------
	listTokOut, err := c.ListTokens("", "", "", 1, 10, true)
	if err != nil {
		t.Fatalf("ListTokens: %v", err)
	}
	if len(listTokOut.States) == 0 {
		t.Fatalf("ListTokens returned empty States")
	}
	if listTokOut.States[0].Object == nil {
		t.Fatalf("ListTokens returned nil state object")
	}

	var tokList []tokenV1Domain.Token
	tests.UnmarshalState(t, listTokOut.States[0].Object, &tokList)

	if len(tokList) == 0 {
		t.Fatalf("ListTokens returned empty list")
	}

	log.Printf("ListTokens Output States: %+v", listTokOut.States)
	log.Printf("ListTokens Output Logs: %+v", listTokOut.Logs)
	log.Printf("ListTokens Output Delegated Call: %+v", listTokOut.DelegatedCall)

	log.Printf("Token Listed Successfully: %+v", tokList)
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
