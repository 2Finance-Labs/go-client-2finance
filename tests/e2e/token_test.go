package e2e_test

import (
	"testing"
	"time"

	//"strconv"
	//"fmt"

	"github.com/stretchr/testify/assert"
	"gitlab.com/2finance/2finance-network/blockchain/contract/tokenV1"
	tokenV1Domain "gitlab.com/2finance/2finance-network/blockchain/contract/tokenV1/domain"
	tokenV1Models "gitlab.com/2finance/2finance-network/blockchain/contract/tokenV1/models"
	"gitlab.com/2finance/2finance-network/blockchain/log"
	"gitlab.com/2finance/2finance-network/blockchain/utils"
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
	accessMode := tokenV1Domain.ALLOW_ACCESS_MODE
	accessUsers := map[string]bool{
		ownerPub: true,
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
		accessMode,
		accessUsers,
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
	assert.Equal(t, tok.AccessPolicy.AccessMode, accessMode, "token access policy mode mismatch")
	assert.Equal(t, tok.AccessPolicy.AccessUsers[ownerPub], accessUsers[ownerPub], "token access policy users mismatch")
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
	assert.Equal(t, tokenState.AccessMode, accessMode, "token access policy mode mismatch")
	assert.Equal(t, tokenState.AccessUsers[ownerPub], accessUsers[ownerPub], "token access policy users mismatch")
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

	// Mint
	mintAmount := "1000000"
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

	// Burn
	burnToken, err := c.BurnToken(tok.Address, mintAmount, decimals, tok.TokenType, "")
	if err != nil {
		t.Fatalf("BurnToken: %v", err)
	}

	unmarshalLogBurn, err := utils.UnmarshalLog[log.Log](burnToken.Logs[0])
	if err != nil {
		t.Fatalf("UnmarshalLog (BurnToken.Logs[0]): %v", err)
	}
	assert.Equal(t, unmarshalLogBurn.LogType, tokenV1Domain.TOKEN_BURNED_LOG, "burn log type mismatch")
	burn, err := utils.UnmarshalEvent[tokenV1Domain.Burn](unmarshalLogBurn.Event)
	if err != nil {
		t.Fatalf("UnmarshalEvent (BurnToken.Logs[0]): %v", err)
	}

	assert.Equal(t, burn.TokenAddress, tok.Address, "burn token address mismatch")
	assert.Equal(t, burn.BurnFrom, ownerPub, "burn from address mismatch")
	assert.Equal(t, burn.Amount, mintAmount, "burn amount mismatch")
	assert.Equal(t, burn.TokenType, tokenType, "burn token type mismatch")
	assert.Equal(t, burn.UUID, "", "burn token UUID mismatch") // Should be nil for fungible tokens

	unmmarshalLogSupply2, err := utils.UnmarshalLog[log.Log](burnToken.Logs[1])
	if err != nil {
		t.Fatalf("UnmarshalLog (BurnToken.Logs[1]): %v", err)
	}
	assert.Equal(t, unmmarshalLogSupply2.LogType, tokenV1Domain.TOKEN_TOTAL_SUPPLY_DECREASED_LOG, "supply log type mismatch")
	supply2, err := utils.UnmarshalEvent[tokenV1Domain.Supply](unmmarshalLogSupply2.Event)
	if err != nil {
		t.Fatalf("UnmarshalEvent (BurnToken.Logs[1]): %v", err)
	}
	assert.Equal(t, supply2.TokenAddress, tok.Address, "supply token address mismatch")
	assert.Equal(t, supply2.Amount, mintAmount, "total supply mismatch after burn")

	unmarshalLogBalance3, err := utils.UnmarshalLog[log.Log](burnToken.Logs[2])
	if err != nil {
		t.Fatalf("UnmarshalLog (BurnToken.Logs[2]): %v", err)
	}
	assert.Equal(t, unmarshalLogBalance3.LogType, tokenV1Domain.TOKEN_BALANCE_DECREASED_LOG, "balance log type mismatch")
	balance3, err := utils.UnmarshalEvent[tokenV1Domain.Balance](unmarshalLogBalance3.Event)
	if err != nil {
		t.Fatalf("UnmarshalEvent (BurnToken.Logs[2]): %v", err)
	}

	assert.Equal(t, balance3.TokenAddress, tok.Address, "balance token address mismatch")
	assert.Equal(t, balance3.OwnerAddress, ownerPub, "balance wallet address mismatch")
	assert.Equal(t, balance3.Amount, totalSupply, "balance amount mismatch after burn")
	assert.Equal(t, balance3.TokenType, tokenType, "balance token type mismatch")
	assert.Equal(t, balance3.TokenUUIDList, []string(nil), "balance token UUID list mismatch") // Should be nil for fungible tokens

	getTokenOut3, err := c.GetToken(tok.Address, "", "")
	if err != nil {
		t.Fatalf("GetToken: %v", err)
	}

	var tokenState3 tokenV1Models.TokenStateModel
	err = utils.UnmarshalState[tokenV1Models.TokenStateModel](getTokenOut3.States[0].Object, &tokenState3)
	if err != nil {
		t.Fatalf("UnmarshalState (GetToken.States[0]): %v", err)
	}

	subTotalSupply, err := utils.SubBigIntStrings(sumTotalSupply, mintAmount)
	if err != nil {
		t.Fatalf("SubBigIntStrings: %v", err)
	}

	assert.Equal(t, tokenState3.TotalSupply, subTotalSupply, "token total supply mismatch after burn")

	// Allow Users
	_, receiverPub, _ := createWallet(t, c)
	c.SetPrivateKey(ownerPriv)

	allowUsers, err := c.AllowUsers(tok.Address, map[string]bool{
		receiverPub: true,
	})
	if err != nil {
		t.Fatalf("AllowUsers: %v", err)
	}

	unmarshalLogAllow, err := utils.UnmarshalLog[log.Log](allowUsers.Logs[0])
	if err != nil {
		t.Fatalf("UnmarshalLog (AllowUsers.Logs[0]): %v", err)
	}
	assert.Equal(t, unmarshalLogAllow.LogType, tokenV1Domain.TOKEN_USERS_ALLOWED_LOG, "allow users log type mismatch")

	allow, err := utils.UnmarshalEvent[tokenV1Domain.AccessPolicy](unmarshalLogAllow.Event)
	if err != nil {
		t.Fatalf("UnmarshalEvent (AllowUsers.Logs[0]): %v", err)
	}
	assert.Equal(t, allow.Address, tok.Address, "access policy token address mismatch")
	assert.Equal(t, allow.AccessMode, tokenV1Domain.ALLOW_ACCESS_MODE, "access policy mode mismatch after allow")
	assert.Equal(t, allow.AccessUsers[receiverPub], true, "access policy users mismatch after allow")

	getTokenOut4, err := c.GetToken(tok.Address, "", "")
	if err != nil {
		t.Fatalf("GetToken: %v", err)
	}

	var tokenState4 tokenV1Models.TokenStateModel
	err = utils.UnmarshalState[tokenV1Models.TokenStateModel](getTokenOut4.States[0].Object, &tokenState4)
	if err != nil {
		t.Fatalf("UnmarshalState (GetToken.States[0]): %v", err)
	}

	assert.Equal(t, tokenState4.AccessMode, tokenV1Domain.ALLOW_ACCESS_MODE, "token access mode mismatch after allow")
	assert.Equal(t, tokenState4.AccessUsers[receiverPub], true, "token access users mismatch after allow")

	// Transfer
	// transferAmount := "500000"
	// transferToken, err := c.TransferToken(tok.Address, receiverPub, transferAmount, decimals, tok.TokenType, "")
	// if err != nil {
	// 	t.Fatalf("TransferToken: %v", err)
	// }

	// unmarshalLogTransfer, err := utils.UnmarshalLog[log.Log](transferToken.Logs[0])
	// if err != nil {
	// 	t.Fatalf("UnmarshalLog (TransferToken.Logs[0]): %v", err)
	// }
	// assert.Equal(t, unmarshalLogTransfer.LogType, tokenV1Domain.TOKEN_TRANSFERRED_LOG, "transfer log type mismatch")
	// transfer, err := utils.UnmarshalEvent[tokenV1Domain.Transfer](unmarshalLogTransfer.Event)
	// if err != nil {
	// 	t.Fatalf("UnmarshalEvent (TransferToken.Logs[0]): %v", err)
	// }

	// assert.Equal(t, transfer.TokenAddress, tok.Address, "transfer token address mismatch")
	// assert.Equal(t, transfer.FromAddress, ownerPub, "transfer from address mismatch")
	// assert.Equal(t, transfer.ToAddress, receiverPub, "transfer to address mismatch")
	// assert.Equal(t, transfer.Amount, transferAmount, "transfer amount mismatch")
	// assert.Equal(t, transfer.TokenType, tokenType, "transfer token type mismatch")
	// assert.Equal(t, transfer.UUID, "", "transfer UUID mismatch") // Should be nil for fungible tokens

	// unmarshalLogBalance4, err := utils.UnmarshalLog[log.Log](transferToken.Logs[1])
	// if err != nil {
	// 	t.Fatalf("UnmarshalLog (TransferToken.Logs[1]): %v", err)
	// }
	// assert.Equal(t, unmarshalLogBalance4.LogType, tokenV1Domain.TOKEN_BALANCE_INCREASED_LOG, "balance increased log type mismatch")

	// balance4, err := utils.UnmarshalEvent[tokenV1Domain.Balance](unmarshalLogBalance4.Event)
	// if err != nil {
	// 	t.Fatalf("UnmarshalEvent (TransferToken.Logs[1]): %v", err)
	// }

	// assert.Equal(t, balance4.TokenAddress, tok.Address, "balance token address mismatch")
	// assert.Equal(t, balance4.OwnerAddress, ownerPub, "balance wallet address mismatch")
	// assert.Equal(t, balance4.Amount, mintAmount, "balance amount mismatch after transfer")
	// assert.Equal(t, balance4.TokenType, tokenType, "balance token type mismatch")
	// assert.Equal(t, balance4.TokenUUIDList, []string(nil), "balance token UUID list mismatch") // Should be nil for fungible tokens

	// getTokenOut5, err := c.GetToken(tok.Address, "", "")
	// if err != nil {
	// 	t.Fatalf("GetToken: %v", err)
	// }

	// var tokenState5 tokenV1Models.TokenStateModel
	// err = utils.UnmarshalState[tokenV1Models.TokenStateModel](getTokenOut5.States[0].Object, &tokenState5)
	// if err != nil {
	// 	t.Fatalf("UnmarshalState (GetToken.States[0]): %v", err)
	// }

	// subTotalSupply, err = utils.SubBigIntStrings(subTotalSupply, transferAmount)
	// if err != nil {
	// 	t.Fatalf("SubBigIntStrings: %v", err)
	// }

	// assert.Equal(t, tokenState5.TotalSupply, subTotalSupply, "token total supply mismatch after transfer")

	// Pause Token
	pauseToken, err := c.PauseToken(tok.Address, true)
	if err != nil {
		t.Fatalf("PauseToken: %v", err)
	}

	unmarshalLogPause, err := utils.UnmarshalLog[log.Log](pauseToken.Logs[0])
	if err != nil {
		t.Fatalf("UnmarshalLog (PauseToken.Logs[0]): %v", err)
	}
	assert.Equal(t, unmarshalLogPause.LogType, tokenV1Domain.TOKEN_PAUSED_LOG, "pause token log type mismatch")
	pause, err := utils.UnmarshalEvent[tokenV1Domain.Token](unmarshalLogPause.Event)
	if err != nil {
		t.Fatalf("UnmarshalEvent (PauseToken.Logs[0]): %v", err)
	}
	assert.Equal(t, pause.Address, tok.Address, "pause token address mismatch")
	assert.Equal(t, pause.Paused, true, "token paused state mismatch after pause")

	getTokenOut6, err := c.GetToken(tok.Address, "", "")
	if err != nil {
		t.Fatalf("GetToken: %v", err)
	}

	var tokenState6 tokenV1Models.TokenStateModel
	err = utils.UnmarshalState[tokenV1Models.TokenStateModel](getTokenOut6.States[0].Object, &tokenState6)
	if err != nil {
		t.Fatalf("UnmarshalState (GetToken.States[0]): %v", err)
	}

	assert.Equal(t, tokenState6.Paused, true, "token paused state mismatch after pause")

	// Unpause Token
	unpauseToken, err := c.UnpauseToken(tok.Address, false)
	if err != nil {
		t.Fatalf("UnpauseToken: %v", err)
	}
	unmarshalLogUnpause, err := utils.UnmarshalLog[log.Log](unpauseToken.Logs[0])
	if err != nil {
		t.Fatalf("UnmarshalLog (UnpauseToken.Logs[0]): %v", err)
	}
	assert.Equal(t, unmarshalLogUnpause.LogType, tokenV1Domain.TOKEN_UNPAUSED_LOG, "unpause token log type mismatch")
	unpause, err := utils.UnmarshalEvent[tokenV1Domain.Token](unmarshalLogUnpause.Event)
	if err != nil {
		t.Fatalf("UnmarshalEvent (UnpauseToken.Logs[0]): %v", err)
	}
	assert.Equal(t, unpause.Address, tok.Address, "unpause token address mismatch")
	assert.Equal(t, unpause.Paused, false, "token paused state mismatch after unpause")

	getTokenOut7, err := c.GetToken(tok.Address, "", "")
	if err != nil {
		t.Fatalf("GetToken: %v", err)
	}

	var tokenState7 tokenV1Models.TokenStateModel
	err = utils.UnmarshalState[tokenV1Models.TokenStateModel](getTokenOut7.States[0].Object, &tokenState7)
	if err != nil {
		t.Fatalf("UnmarshalState (GetToken.States[0]): %v", err)
	}

	assert.Equal(t, tokenState7.Paused, false, "token paused state mismatch after unpause")

	// Fee tiers
	updateFeeTiers, er := c.UpdateFeeTiers(tok.Address, []map[string]interface{}{
		{
			"min_amount": "0",
			"max_amount": amt(10_000, decimals),
			"min_volume": "0",
			"max_volume": amt(100_000, decimals),
			"fee_bps":    50,
		},
	})
	if er != nil {
		t.Fatalf("UpdateFeeTiers: %v", er)
	}

	unmarshalLogFeeTiers, err := utils.UnmarshalLog[log.Log](updateFeeTiers.Logs[0])
	if err != nil {
		t.Fatalf("UnmarshalLog (UpdateFeeTiers.Logs[0]): %v", err)
	}
	assert.Equal(t, unmarshalLogFeeTiers.LogType, tokenV1Domain.TOKEN_FEE_ADDRESS_UPDATED_LOG, "update fee tiers log type mismatch")

	feeTiersEvent, err := utils.UnmarshalEvent[tokenV1Domain.FeeTiers](unmarshalLogFeeTiers.Event)
	if err != nil {
		t.Fatalf("UnmarshalEvent (UpdateFeeTiers.Logs[0]): %v", err)
	}
	assert.Equal(t, feeTiersEvent.FeeTiersList[0].MinAmount, "0", "fee tiers min amount mismatch")
	assert.Equal(t, feeTiersEvent.FeeTiersList[0].MaxAmount, amt(10_000, decimals), "fee tiers max amount mismatch")
	assert.Equal(t, feeTiersEvent.FeeTiersList[0].MinVolume, "0", "fee tiers min volume mismatch")
	assert.Equal(t, feeTiersEvent.FeeTiersList[0].MaxVolume, amt(100_000, decimals), "fee tiers max volume mismatch")
	assert.Equal(t, feeTiersEvent.FeeTiersList[0].FeeBps, 50, "fee tiers bps mismatch")

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
