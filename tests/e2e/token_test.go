package e2e_test

import (
	"testing"
	"time"
	//"strconv"
	//"fmt"

	"github.com/stretchr/testify/assert"
	"gitlab.com/2finance/2finance-network/blockchain/contract/tokenV1"
	"gitlab.com/2finance/2finance-network/blockchain/utils"
	"gitlab.com/2finance/2finance-network/blockchain/log"
	tokenV1Domain "gitlab.com/2finance/2finance-network/blockchain/contract/tokenV1/domain"
	tokenV1Models "gitlab.com/2finance/2finance-network/blockchain/contract/tokenV1/models"
)

func TestTokenFlowFungible(t *testing.T) {
	c := setupClient(t)
	_, ownerPub, ownerPriv := createWallet(t, c)
	c.SetPrivateKey(ownerPriv)

	deployedContract, err := c.DeployContract1(tokenV1.TOKEN_CONTRACT_V1)
	if err != nil {
		t.Fatalf("DeployContract: %v", err)
	}
	contractLog, err := utils.UnmarshalLog[log.Log](deployedContract.Logs[0])
	if err != nil {
		t.Fatalf("UnmarshalLog (AddWallet.Logs[0]): %v", err)
	}
	address := contractLog.ContractAddress

	decimals := 6
	tokenType := tokenV1Domain.FUNGIBLE
	stablecoin := false
	symbol := "2F" + randSuffix(4)
	name := "2Finance"
	var totalSupply string
	totalSupply = "100000000"
	description := "e2e token created by tests"
	image := "https://example.com/image.png"
	website := "https://example.com"
	tagsSocial := map[string]string{"twitter": "https://twitter.com/2finance"}
	tagsCat := map[string]string{"category": "DeFi"}
	tags := map[string]string{"tag1": "DeFi", "tag2": "Blockchain"}
	creator := "2Finance Test"
	creatorWebsite := "https://creator.example"
	accessPolicy := tokenV1Domain.AccessPolicy{
		Mode: tokenV1Domain.ALLOW,
		Users: map[string]bool{
			ownerPub: true,
		},
	}
	frozenAccounts := map[string]bool{}
	feeTiers := []map[string]interface{}{}
	requireFee := false
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

	unmarshalLogToken, err := utils.UnmarshalLog[log.Log](out.Logs[0])
	if err != nil {
		t.Fatalf("UnmarshalLog (AddToken.Logs[0]): %v", err)
	}
	assert.Equal(t, unmarshalLogToken.LogType, tokenV1Domain.TOKEN_CREATED_LOG, "add-token log type mismatch")
	tok, err := utils.UnmarshalEvent[tokenV1Domain.Token](unmarshalLogToken.Event)
	if err != nil {
		t.Fatalf("UnmarshalEvent (AddToken.Logs[0]): %v", err)
	}

	assert.Equal(t, tok.Address, address, "token address mismatch")
	assert.Equal(t, tok.Symbol, symbol, "token symbol mismatch")
	assert.Equal(t, tok.Name, name, "token name mismatch")
	assert.Equal(t, tok.Decimals, decimals, "token decimals mismatch")
	assert.Equal(t, tok.TotalSupply, totalSupply, "token total supply mismatch")
	assert.Equal(t, tok.Description, description, "token description mismatch")
	assert.Equal(t, tok.Image, image, "token image mismatch")
	assert.Equal(t, tok.Website, website, "token website mismatch")
	assert.Equal(t, tok.TagsSocialMedia["twitter"], tagsSocial["twitter"], "token tags social mismatch")
	assert.Equal(t, tok.TagsCategory["category"], tagsCat["category"], "token tags category mismatch")
	assert.Equal(t, tok.Tags["tag1"], tags["tag1"], "token tags mismatch")
	assert.Equal(t, tok.Creator, creator, "token creator mismatch")
	assert.Equal(t, tok.CreatorWebsite, creatorWebsite, "token creator website mismatch")
	assert.Equal(t, tok.AccessPolicy.Mode, accessPolicy.Mode, "token access policy mode mismatch")
	assert.Equal(t, tok.AccessPolicy.Users[ownerPub], accessPolicy.Users[ownerPub], "token access policy users mismatch")
	assert.Equal(t, tok.FrozenAccounts[ownerPub], frozenAccounts[ownerPub], "token frozen accounts mismatch")
	// Skipping fee tiers deep equality for simplicity
	assert.Equal(t, tok.FeeAddress, feeAddress, "token fee address mismatch")
	assert.Equal(t, tok.FreezeAuthorityRevoked, freezeAuthorityRevoked, "token freeze authority revoked mismatch")
	assert.Equal(t, tok.MintAuthorityRevoked, mintAuthorityRevoked, "token mint authority revoked mismatch")
	assert.Equal(t, tok.UpdateAuthorityRevoked, updateAuthorityRevoked, "token update authority revoked mismatch")
	assert.Equal(t, tok.Paused, paused, "token paused mismatch")
	// Skipping expiredAt deep equality for simplicity
	assert.Equal(t, tok.AssetGLBUri, assetGLBUri, "token asset GLB URI mismatch")
	assert.Equal(t, tok.TokenType, tokenType, "token type mismatch")
	assert.Equal(t, tok.Transferable, transferable, "token transferable mismatch")
	assert.Equal(t, tok.Stablecoin, stablecoin, "token stablecoin mismatch")

	unmarshalLogMint, err := utils.UnmarshalLog[log.Log](out.Logs[1])
	if err != nil {
		t.Fatalf("UnmarshalLog (AddToken.Logs[1]): %v", err)
	}
	assert.Equal(t, unmarshalLogMint.LogType, tokenV1Domain.TOKEN_MINTED_LOG, "mint log type mismatch")
	mint, err := utils.UnmarshalEvent[tokenV1Domain.Mint](unmarshalLogMint.Event)
	if err != nil {
		t.Fatalf("UnmarshalEvent (AddToken.Logs[1]): %v", err)
	}

	assert.Equal(t, mint.TokenAddress, tok.Address, "mint to address mismatch")
	assert.Equal(t, mint.MintTo, ownerPub, "mint to address mismatch")
	assert.Equal(t, mint.Amount, totalSupply, "mint amount mismatch")
	assert.Equal(t, mint.TokenType, tokenType, "mint token type mismatch")
	assert.Equal(t, mint.TokenUUIDList, []string(nil), "mint token UUID list mismatch") // Should be nil for fungible tokens

	unmarshalLogBalance, err := utils.UnmarshalLog[log.Log](out.Logs[2])
	if err != nil {
		t.Fatalf("UnmarshalLog (AddToken.Logs[2]): %v", err)
	}
	assert.Equal(t, unmarshalLogBalance.LogType, tokenV1Domain.TOKEN_BALANCE_INCREASED_LOG, "balance log type mismatch")

	balance, err := utils.UnmarshalEvent[tokenV1Domain.Balance](unmarshalLogBalance.Event)
	if err != nil {
		t.Fatalf("UnmarshalEvent (AddToken.Logs[2]): %v", err)
	}

	assert.Equal(t, balance.TokenAddress, tok.Address, "balance token address mismatch")
	assert.Equal(t, balance.OwnerAddress, ownerPub, "balance wallet address mismatch")
	assert.Equal(t, balance.Amount, totalSupply, "balance amount mismatch")
	assert.Equal(t, balance.TokenType, tokenType, "balance token type mismatch")
	assert.Equal(t, balance.TokenUUIDList, []string(nil), "balance token UUID list mismatch") // Should be nil for fungible tokens

	getTokenOut, err := c.GetToken(tok.Address, "", "")
	if err != nil {
		t.Fatalf("GetToken: %v", err)
	}

	var tokenState tokenV1Models.TokenStateModel
	err = utils.UnmarshalState[tokenV1Models.TokenStateModel](getTokenOut.States[0].Object, &tokenState)
	if err != nil {
		t.Fatalf("UnmarshalState (GetToken.States[0]): %v", err)
	}

	assert.Equal(t, tokenState.Address, address, "token address mismatch")
	assert.Equal(t, tokenState.Symbol, symbol, "token symbol mismatch")
	assert.Equal(t, tokenState.Name, name, "token name mismatch")
	assert.Equal(t, tokenState.Decimals, decimals, "token decimals mismatch")
	assert.Equal(t, tokenState.TotalSupply, totalSupply, "token total supply mismatch")
	assert.Equal(t, tokenState.Description, description, "token description mismatch")
	assert.Equal(t, tokenState.Image, image, "token image mismatch")
	assert.Equal(t, tokenState.Website, website, "token website mismatch")
	assert.Equal(t, tokenState.TagsSocialMedia["twitter"], tagsSocial["twitter"], "token tags social mismatch")
	assert.Equal(t, tokenState.TagsCategory["category"], tagsCat["category"], "token tags category mismatch")
	assert.Equal(t, tokenState.Tags["tag1"], tags["tag1"], "token tags mismatch")
	assert.Equal(t, tokenState.Creator, creator, "token creator mismatch")
	assert.Equal(t, tokenState.CreatorWebsite, creatorWebsite, "token creator website mismatch")
	assert.Equal(t, tokenState.AccessPolicy.Mode, accessPolicy.Mode, "token access policy mode mismatch")
	assert.Equal(t, tokenState.AccessPolicy.Users[ownerPub], accessPolicy.Users[ownerPub], "token access policy users mismatch")
	assert.Equal(t, tokenState.FrozenAccounts[ownerPub], frozenAccounts[ownerPub], "token frozen accounts mismatch")
	// Skipping fee tiers deep equality for simplicity
	assert.Equal(t, tokenState.FeeAddress, feeAddress, "token fee address mismatch")
	assert.Equal(t, tokenState.FreezeAuthorityRevoked, freezeAuthorityRevoked, "token freeze authority revoked mismatch")
	assert.Equal(t, tokenState.MintAuthorityRevoked, mintAuthorityRevoked, "token mint authority revoked mismatch")
	assert.Equal(t, tokenState.UpdateAuthorityRevoked, updateAuthorityRevoked, "token update authority revoked mismatch")
	assert.Equal(t, tokenState.Paused, paused, "token paused mismatch")
	// Skipping expiredAt deep equality for simplicity
	assert.Equal(t, tokenState.AssetGLBUri, assetGLBUri, "token asset GLB URI mismatch")
	assert.Equal(t, tokenState.TokenType, tokenType, "token type mismatch")
	assert.Equal(t, tokenState.Transferable, transferable, "token transferable mismatch")
	assert.Equal(t, tokenState.Stablecoin, stablecoin, "token stablecoin mismatch")

	mintAmount := "1000000"
	// Mint & Burn
	mintToken, err := c.MintToken(tok.Address, ownerPub, mintAmount, decimals, tok.TokenType)
	if err != nil {
		t.Fatalf("MintToken: %v", err)
	}

	unmarshalLogMint2, err := utils.UnmarshalLog[log.Log](mintToken.Logs[0])
	if err != nil {
		t.Fatalf("UnmarshalLog (MintToken.Logs[0]): %v", err)
	}
	assert.Equal(t, unmarshalLogMint2.LogType, tokenV1Domain.TOKEN_MINTED_LOG, "mint log type mismatch")
	mint2, err := utils.UnmarshalEvent[tokenV1Domain.Mint](unmarshalLogMint2.Event)
	if err != nil {
		t.Fatalf("UnmarshalEvent (MintToken.Logs[0]): %v", err)
	}

	assert.Equal(t, mint2.TokenAddress, tok.Address, "mint to address mismatch")
	assert.Equal(t, mint2.MintTo, ownerPub, "mint to address mismatch")
	assert.Equal(t, mint2.Amount, mintAmount, "mint amount mismatch")
	assert.Equal(t, mint2.TokenType, tokenType, "mint token type mismatch")
	assert.Equal(t, mint2.TokenUUIDList, []string(nil), "mint token UUID list mismatch") // Should be nil for fungible tokens

	unmmarshalLogSupply, err := utils.UnmarshalLog[log.Log](mintToken.Logs[1])
	if err != nil {
		t.Fatalf("UnmarshalLog (MintToken.Logs[1]): %v", err)
	}
	assert.Equal(t, unmmarshalLogSupply.LogType, tokenV1Domain.TOKEN_TOTAL_SUPPLY_INCREASED_LOG, "supply log type mismatch")
	supply, err := utils.UnmarshalEvent[tokenV1Domain.Supply](unmmarshalLogSupply.Event)
	if err != nil {
		t.Fatalf("UnmarshalEvent (MintToken.Logs[1]): %v", err)
	}
	assert.Equal(t, supply.TokenAddress, tok.Address, "supply token address mismatch")
	assert.Equal(t, supply.Amount, mintAmount, "total supply mismatch after mint")

	unmarshalLogBalance2, err := utils.UnmarshalLog[log.Log](mintToken.Logs[2])
	if err != nil {
		t.Fatalf("UnmarshalLog (MintToken.Logs[2]): %v", err)
	}
	assert.Equal(t, unmarshalLogBalance2.LogType, tokenV1Domain.TOKEN_BALANCE_INCREASED_LOG, "balance log type mismatch")
	balance2, err := utils.UnmarshalEvent[tokenV1Domain.Balance](unmarshalLogBalance2.Event)
	if err != nil {
		t.Fatalf("UnmarshalEvent (MintToken.Logs[2]): %v", err)
	}

	assert.Equal(t, balance2.TokenAddress, tok.Address, "balance token address mismatch")
	assert.Equal(t, balance2.OwnerAddress, ownerPub, "balance wallet address mismatch")
	assert.Equal(t, balance2.Amount, mintAmount, "balance amount mismatch after mint")
	assert.Equal(t, balance2.TokenType, tokenType, "balance token type mismatch")
	assert.Equal(t, balance2.TokenUUIDList, []string(nil), "balance token UUID list mismatch") // Should be nil for fungible tokens

	
	getTokenOut2, err := c.GetToken(tok.Address, "", "")
	if err != nil {
		t.Fatalf("GetToken: %v", err)
	}
	
	var tokenState2 tokenV1Models.TokenStateModel
	err = utils.UnmarshalState[tokenV1Models.TokenStateModel](getTokenOut2.States[0].Object, &tokenState2)
	if err != nil {
		t.Fatalf("UnmarshalState (GetToken.States[0]): %v", err)
	}

	sumTotalSupply, err := utils.AddBigIntStrings(totalSupply, mintAmount)
	if err != nil {
		t.Fatalf("AddBigIntStrings: %v", err)
	}

	assert.Equal(t, tokenState2.TotalSupply, sumTotalSupply, "token total supply mismatch after mint")
	
	
	// if _, err := c.BurnToken(tok.Address, amt(12, decimals), decimals, tok.TokenType, ""); err != nil {
	// 	t.Fatalf("BurnToken: %v", err)
	// }

	// // // Transfer to allowed wallet
	// receiver, _ := createWallet(t, c)
	// c.SetPrivateKey(ownerPriv)

	// if _, err := c.AllowUsers(tok.Address, map[string]bool{
	// 	receiver.PublicKey: true,
	// }); err != nil {
	// 	t.Fatalf("AllowUsers: %v", err)
	// }

	// trOut, err := c.TransferToken(
	// 	tok.Address,
	// 	receiver.PublicKey,
	// 	amt(1, dec),
	// 	dec,
	// 	tok.TokenType,
	// 	"",
	// )
	// if err != nil {
	// 	t.Fatalf("TransferToken: %v", err)
	// }

	// var tr tokenV1Domain.Transfer
	// unmarshalState(t, trOut.States[0].Object, &tr)
	// if tr.ToAddress != receiver.PublicKey {
	// 	t.Fatalf("transfer mismatch: %s != %s", tr.ToAddress, receiver.PublicKey)
	// }

	// // Fee tiers & address
	// if _, err := c.UpdateFeeTiers(tok.Address, []map[string]interface{}{
	// 	{
	// 		"min_amount": "0",
	// 		"max_amount": amt(10_000, dec),
	// 		"min_volume": "0",
	// 		"max_volume": amt(100_000, dec),
	// 		"fee_bps":    25,
	// 	},
	// }); err != nil {
	// 	t.Fatalf("UpdateFeeTiers: %v", err)
	// }

	// if _, err := c.UpdateFeeAddress(tok.Address, owner.PublicKey); err != nil {
	// 	t.Fatalf("UpdateFeeAddress: %v", err)
	// }

	// // Metadata / Authorities / Pause
	// if _, err := c.UpdateMetadata(
	// 	tok.Address,
	// 	"2F-NEW"+randSuffix(4),
	// 	"2Finance New",
	// 	dec,
	// 	"Updated by tests",
	// 	"https://example.com/img.png",
	// 	"https://example.com",
	// 	map[string]string{"twitter": "https://x.com/2f"},
	// 	map[string]string{"category": "DeFi"},
	// 	map[string]string{"tag": "e2e"},
	// 	"creator",
	// 	"https://creator",
	// 	time.Now().Add(30*24*time.Hour),
	// ); err != nil {
	// 	t.Fatalf("UpdateMetadata: %v", err)
	// }

	// if _, err := c.RevokeMintAuthority(tok.Address, true); err != nil {
	// 	t.Fatalf("RevokeMintAuthority: %v", err)
	// }
	// if _, err := c.RevokeUpdateAuthority(tok.Address, true); err != nil {
	// 	t.Fatalf("RevokeUpdateAuthority: %v", err)
	// }
	// if _, err := c.PauseToken(tok.Address, true); err != nil {
	// 	t.Fatalf("PauseToken: %v", err)
	// }
	// if _, err := c.UnpauseToken(tok.Address, false); err != nil {
	// 	t.Fatalf("UnpauseToken: %v", err)
	// }
	// if _, err := c.FreezeWallet(tok.Address, owner.PublicKey); err != nil {
	// 	t.Fatalf("FreezeWallet: %v", err)
	// }
	// if _, err := c.UnfreezeWallet(tok.Address, owner.PublicKey); err != nil {
	// 	t.Fatalf("UnfreezeWallet: %v", err)
	// }


	// // Balances / Listings
	// if _, err := c.GetTokenBalance(tok.Address, owner.PublicKey); err != nil {
	// 	t.Fatalf("GetTokenBalance(owner): %v", err)
	// }
	// if _, err := c.ListTokenBalances(tok.Address, "", 1, 10, true); err != nil {
	// 	t.Fatalf("ListTokenBalances: %v", err)
	// }
	// if _, err := c.GetToken(tok.Address, "", ""); err != nil {
	// 	t.Fatalf("GetToken: %v", err)
	// }
	// if _, err := c.ListTokens("", "", "", 1, 10, true); err != nil {
	// 	t.Fatalf("ListTokens: %v", err)
	// }
}

func TestTokenFlowNonFungible(t *testing.T) {
	// c := setupClient(t)
	// owner, ownerPriv := createWallet(t, c)
	// c.SetPrivateKey(ownerPriv)

	// dec := 0
	// tokenType := tokenV1Domain.NON_FUNGIBLE

	// stablecoin := false

	// tok := createBasicToken(t, c, owner.PublicKey, dec, false, tokenType, stablecoin)

	// amount := "35"

	// // Mint NFT
	// mintOut, err := c.MintToken(tok.Address, owner.PublicKey, amount, dec, tok.TokenType)
	// if err != nil {
	// 	t.Fatalf("MintToken NFT: %v", err)
	// }

	// var mint tokenV1Domain.Mint
	// unmarshalState(t, mintOut.States[0].Object, &mint)

	// amountInt, _ := strconv.Atoi(amount)
	// if len(mint.TokenUUIDList) != amountInt {
	// 	t.Fatalf("expected %d uuid, got %d", amountInt, len(mint.TokenUUIDList))
	// }

	// if _, err := c.BurnToken(
	// 	tok.Address,
	// 	amt(1, dec),
	// 	dec,
	// 	tok.TokenType,
	// 	mint.TokenUUIDList[0],
	// ); err != nil {
	// 	t.Fatalf("BurnToken: %v", err)
	// }

	// // Transfer NFT
	// receiver, _ := createWallet(t, c)
	// c.SetPrivateKey(ownerPriv)

	// // ✅ NECESSÁRIO em ALLOW
	// if _, err := c.AllowUsers(tok.Address, map[string]bool{
	// 	receiver.PublicKey: true,
	// }); err != nil {
	// 	t.Fatalf("AllowUsers: %v", err)
	// }

	// trOut, err := c.TransferToken(
	// 	tok.Address,
	// 	receiver.PublicKey,
	// 	amt(1, dec),
	// 	dec,
	// 	tok.TokenType,
	// 	mint.TokenUUIDList[0],
	// )
	// if err != nil {
	// 	t.Fatalf("Transfer NFT: %v", err)
	// }

	// if _, err := c.FreezeWallet(tok.Address, owner.PublicKey); err != nil {
	// 	t.Fatalf("FreezeWallet: %v", err)
	// }
	// if _, err := c.UnfreezeWallet(tok.Address, owner.PublicKey); err != nil {
	// 	t.Fatalf("UnfreezeWallet: %v", err)
	// }

	// var tr tokenV1Domain.Transfer
	// unmarshalState(t, trOut.States[0].Object, &tr)
	// if tr.ToAddress != receiver.PublicKey {
	// 	t.Fatalf("transfer mismatch: %s != %s", tr.ToAddress, receiver.PublicKey)
	// }
}

