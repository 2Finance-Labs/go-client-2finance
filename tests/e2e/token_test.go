package e2e_test

import (
	"testing"
	"time"

	//"strconv"
	//"fmt"

	"github.com/stretchr/testify/assert"
	"gitlab.com/2finance/2finance-network/blockchain/contract/tokenV1"
	"gitlab.com/2finance/2finance-network/blockchain/contract/tokenV1/domain"
	tokenV1Domain "gitlab.com/2finance/2finance-network/blockchain/contract/tokenV1/domain"
	tokenV1Models "gitlab.com/2finance/2finance-network/blockchain/contract/tokenV1/models"
	"gitlab.com/2finance/2finance-network/blockchain/log"
	"gitlab.com/2finance/2finance-network/blockchain/utils"
)

func TestTokenFlowFungible(t *testing.T) {
	c := setupClient(t)
	owner, ownerPriv := createWallet(t, c)
	c.SetPrivateKey(ownerPriv)

	deployedContract, err := c.DeployContract1(tokenV1.TOKEN_CONTRACT_V1)
	if err != nil {
		t.Fatalf("DeployContract: %v", err)
	}
	contractLog, err := utils.UnmarshalLog[log.Log](deployedContract.Logs[0])
	if err != nil {
		t.Fatalf("UnmarshalLog (AddWallet.Logs[0]): %v", err)
	}

	// ------------------
	//    CREATE TOKEN
	// ------------------
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
	feeAddress := owner.PublicKey
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
		owner.PublicKey,
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
	assert.Equal(t, tok.AccessMode, accessMode, "token access policy mode mismatch")
	assert.Equal(t, tok.AccessUsers[ownerPub], accessUsers[ownerPub], "token access policy users mismatch")
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
	assert.Equal(t, mint.MintTo, owner.PublicKey, "mint to address mismatch")
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
	assert.Equal(t, balance.OwnerAddress, owner.PublicKey, "balance wallet address mismatch")
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

	// ------------------
	//        MINT
	// ------------------
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
	assert.Equal(t, mint2.MintTo, owner.PublicKey, "mint to address mismatch")
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
	assert.Equal(t, balance2.OwnerAddress, owner.PublicKey, "balance wallet address mismatch")
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

	// ------------------
	//        BURN
	// ------------------
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

	// ------------------
	//    ALLOW USERS
	// ------------------

	// Add allow users
	_, receiverPub, _ := createWallet(t, c)
	c.SetPrivateKey(ownerPriv)

	allowUsers, err := c.AddAllowUsers(tok.Address, domain.ALLOW_ACCESS_MODE, map[string]bool{
		receiverPub: true,
	})
	if err != nil {
		t.Fatalf("AllowUsers: %v", err)
	}

	unmarshalLogAllow, err := utils.UnmarshalLog[log.Log](allowUsers.Logs[0])
	if err != nil {
		t.Fatalf("UnmarshalLog (AllowUsers.Logs[0]): %v", err)
	}
	assert.Equal(t, unmarshalLogAllow.LogType, tokenV1Domain.TOKEN_USERS_ADDED_LOG, "allow users log type mismatch")

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

	// // Remove allow users
	// removeAllowUsers, err := c.RemoveAllowUsers(tok.Address, domain.ALLOW_ACCESS_MODE, map[string]bool{
	// 	receiverPub: true,
	// })
	// if err != nil {
	// 	t.Fatalf("RemoveAllowUsers: %v", err)
	// }

	// unmarshalLogRemoveAllow, err := utils.UnmarshalLog[log.Log](removeAllowUsers.Logs[0])
	// if err != nil {
	// 	t.Fatalf("UnmarshalLog (RemoveAllowUsers.Logs[0]): %v", err)
	// }
	// assert.Equal(t, unmarshalLogRemoveAllow.LogType, tokenV1Domain.TOKEN_USERS_REMOVED_LOG, "remove allow users log type mismatch")

	// removeAllow, err := utils.UnmarshalEvent[tokenV1Domain.AccessPolicy](unmarshalLogRemoveAllow.Event)
	// if err != nil {
	// 	t.Fatalf("UnmarshalEvent (RemoveAllowUsers.Logs[0]): %v", err)
	// }
	// assert.Equal(t, removeAllow.Address, tok.Address, "access policy token address mismatch")
	// assert.Equal(t, removeAllow.AccessMode, tokenV1Domain.ALLOW_ACCESS_MODE, "access policy mode mismatch after remove allow")
	// assert.Equal(t, removeAllow.AccessUsers[receiverPub], false, "access policy users mismatch after remove allow")

	// getTokenOut5, err := c.GetToken(tok.Address, "", "")
	// if err != nil {
	// 	t.Fatalf("GetToken: %v", err)
	// }

	// var tokenState5 tokenV1Models.TokenStateModel
	// err = utils.UnmarshalState[tokenV1Models.TokenStateModel](getTokenOut5.States[0].Object, &tokenState5)
	// if err != nil {
	// 	t.Fatalf("UnmarshalState (GetToken.States[0]): %v", err)
	// }
	// assert.Equal(t, tokenState5.AccessMode, tokenV1Domain.ALLOW_ACCESS_MODE, "token access mode mismatch after remove allow")
	// assert.Equal(t, tokenState5.AccessUsers[receiverPub], false, "token access users mismatch after remove allow")

	// ------------------
	//    DENY USERS
	// ------------------

	// // Add deny users
	// denyUsers, err := c.AddDenyUsers(tok.Address, tokenV1Domain.DENY_ACCESS_MODE, map[string]bool{
	// 	receiverPub: true,
	// })
	// if err != nil {
	// 	t.Fatalf("AddDenyUsers: %v", err)
	// }

	// unmarshalLogDeny, err := utils.UnmarshalLog[log.Log](denyUsers.Logs[0])
	// if err != nil {
	// 	t.Fatalf("UnmarshalLog (AddDenyUsers.Logs[0]): %v", err)
	// }
	// assert.Equal(t, unmarshalLogDeny.LogType, tokenV1Domain.TOKEN_USERS_ADDED_LOG, "add deny users log type mismatch")

	// deny, err := utils.UnmarshalEvent[tokenV1Domain.AccessPolicy](unmarshalLogDeny.Event)
	// if err != nil {
	// 	t.Fatalf("UnmarshalEvent (AddDenyUsers.Logs[0]): %v", err)
	// }
	// assert.Equal(t, deny.Address, tok.Address, "access policy token address mismatch")
	// assert.Equal(t, deny.AccessMode, tokenV1Domain.DENY_ACCESS_MODE, "access policy mode mismatch after add deny")
	// assert.Equal(t, deny.AccessUsers[receiverPub], true, "access policy users mismatch after add deny")

	// getTokenOut6, err := c.GetToken(tok.Address, "", "")
	// if err != nil {
	// 	t.Fatalf("GetToken: %v", err)
	// }

	// var tokenState6 tokenV1Models.TokenStateModel
	// err = utils.UnmarshalState[tokenV1Models.TokenStateModel](getTokenOut6.States[0].Object, &tokenState6)
	// if err != nil {
	// 	t.Fatalf("UnmarshalState (GetToken.States[0]): %v", err)
	// }
	// assert.Equal(t, tokenState6.AccessMode, tokenV1Domain.DENY_ACCESS_MODE, "token access mode mismatch after add deny")
	// assert.Equal(t, tokenState6.AccessUsers[receiverPub], true, "token access users mismatch after add deny")

	// // Remove deny users
	// removeDenyUsers, err := c.RemoveDenyUsers(tok.Address, tokenV1Domain.DENY_ACCESS_MODE, map[string]bool{
	// 	receiverPub: true,
	// })
	// if err != nil {
	// 	t.Fatalf("RemoveDenyUsers: %v", err)
	// }

	// unmarshalLogRemoveDeny, err := utils.UnmarshalLog[log.Log](removeDenyUsers.Logs[0])
	// if err != nil {
	// 	t.Fatalf("UnmarshalLog (RemoveDenyUsers.Logs[0]): %v", err)
	// }
	// assert.Equal(t, unmarshalLogRemoveDeny.LogType, tokenV1Domain.TOKEN_USERS_REMOVED_LOG, "remove deny users log type mismatch")

	// removeDeny, err := utils.UnmarshalEvent[tokenV1Domain.AccessPolicy](unmarshalLogRemoveDeny.Event)
	// if err != nil {
	// 	t.Fatalf("UnmarshalEvent (RemoveDenyUsers.Logs[0]): %v", err)
	// }
	// assert.Equal(t, removeDeny.Address, tok.Address, "access policy token address mismatch")
	// assert.Equal(t, removeDeny.AccessMode, tokenV1Domain.DENY_ACCESS_MODE, "access policy mode mismatch after remove deny")
	// assert.Equal(t, removeDeny.AccessUsers[receiverPub], false, "access policy users mismatch after remove deny")

	// getTokenOut7, err := c.GetToken(tok.Address, "", "")
	// if err != nil {
	// 	t.Fatalf("GetToken: %v", err)
	// }

	// var tokenState7 tokenV1Models.TokenStateModel
	// err = utils.UnmarshalState[tokenV1Models.TokenStateModel](getTokenOut7.States[0].Object, &tokenState7)
	// if err != nil {
	// 	t.Fatalf("UnmarshalState (GetToken.States[0]): %v", err)
	// }
	// assert.Equal(t, tokenState7.AccessMode, tokenV1Domain.DENY_ACCESS_MODE, "token access mode mismatch after remove deny")
	// assert.Equal(t, tokenState7.AccessUsers[receiverPub], false, "token access users mismatch after remove deny")

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

	// // Pause Token
	// pauseToken, err := c.PauseToken(tok.Address, true)
	// if err != nil {
	// 	t.Fatalf("PauseToken: %v", err)
	// }

	// unmarshalLogPause, err := utils.UnmarshalLog[log.Log](pauseToken.Logs[0])
	// if err != nil {
	// 	t.Fatalf("UnmarshalLog (PauseToken.Logs[0]): %v", err)
	// }
	// assert.Equal(t, unmarshalLogPause.LogType, tokenV1Domain.TOKEN_PAUSED_LOG, "pause token log type mismatch")
	// pause, err := utils.UnmarshalEvent[tokenV1Domain.Token](unmarshalLogPause.Event)
	// if err != nil {
	// 	t.Fatalf("UnmarshalEvent (PauseToken.Logs[0]): %v", err)
	// }
	// assert.Equal(t, pause.Address, tok.Address, "pause token address mismatch")
	// assert.Equal(t, pause.Paused, true, "token paused state mismatch after pause")

	// getTokenOut6, err := c.GetToken(tok.Address, "", "")
	// if err != nil {
	// 	t.Fatalf("GetToken: %v", err)
	// }

	// var tokenState6 tokenV1Models.TokenStateModel
	// err = utils.UnmarshalState[tokenV1Models.TokenStateModel](getTokenOut6.States[0].Object, &tokenState6)
	// if err != nil {
	// 	t.Fatalf("UnmarshalState (GetToken.States[0]): %v", err)
	// }

	// assert.Equal(t, tokenState6.Paused, true, "token paused state mismatch after pause")

	// // Unpause Token
	// unpauseToken, err := c.UnpauseToken(tok.Address, false)
	// if err != nil {
	// 	t.Fatalf("UnpauseToken: %v", err)
	// }
	// unmarshalLogUnpause, err := utils.UnmarshalLog[log.Log](unpauseToken.Logs[0])
	// if err != nil {
	// 	t.Fatalf("UnmarshalLog (UnpauseToken.Logs[0]): %v", err)
	// }
	// assert.Equal(t, unmarshalLogUnpause.LogType, tokenV1Domain.TOKEN_UNPAUSED_LOG, "unpause token log type mismatch")
	// unpause, err := utils.UnmarshalEvent[tokenV1Domain.Token](unmarshalLogUnpause.Event)
	// if err != nil {
	// 	t.Fatalf("UnmarshalEvent (UnpauseToken.Logs[0]): %v", err)
	// }
	// assert.Equal(t, unpause.Address, tok.Address, "unpause token address mismatch")
	// assert.Equal(t, unpause.Paused, false, "token paused state mismatch after unpause")

	// getTokenOut7, err := c.GetToken(tok.Address, "", "")
	// if err != nil {
	// 	t.Fatalf("GetToken: %v", err)
	// }

	// var tokenState7 tokenV1Models.TokenStateModel
	// err = utils.UnmarshalState[tokenV1Models.TokenStateModel](getTokenOut7.States[0].Object, &tokenState7)
	// if err != nil {
	// 	t.Fatalf("UnmarshalState (GetToken.States[0]): %v", err)
	// }

	// assert.Equal(t, tokenState7.Paused, false, "token paused state mismatch after unpause")

	// // Fee tiers
	// updateFeeTiers, er := c.UpdateFeeTiers(tok.Address, []map[string]interface{}{
	// 	{
	// 		"min_amount": "0",
	// 		"max_amount": amt(10_000, decimals),
	// 		"min_volume": "0",
	// 		"max_volume": amt(100_000, decimals),
	// 		"fee_bps":    50,
	// 	},
	// })
	// if er != nil {
	// 	t.Fatalf("UpdateFeeTiers: %v", er)
	// }

	// unmarshalLogFeeTiers, err := utils.UnmarshalLog[log.Log](updateFeeTiers.Logs[0])
	// if err != nil {
	// 	t.Fatalf("UnmarshalLog (UpdateFeeTiers.Logs[0]): %v", err)
	// }
	// assert.Equal(t, unmarshalLogFeeTiers.LogType, tokenV1Domain.TOKEN_FEE_UPDATED_LOG, "update fee tiers log type mismatch")

	// feeTiersEvent, err := utils.UnmarshalEvent[tokenV1Domain.FeeTiers](unmarshalLogFeeTiers.Event)
	// if err != nil {
	// 	t.Fatalf("UnmarshalEvent (UpdateFeeTiers.Logs[0]): %v", err)
	// }
	// assert.Equal(t, feeTiersEvent.FeeTiersList[0].MinAmount, "0", "fee tiers min amount mismatch")
	// assert.Equal(t, feeTiersEvent.FeeTiersList[0].MaxAmount, amt(10_000, decimals), "fee tiers max amount mismatch")
	// assert.Equal(t, feeTiersEvent.FeeTiersList[0].MinVolume, "0", "fee tiers min volume mismatch")
	// assert.Equal(t, feeTiersEvent.FeeTiersList[0].MaxVolume, amt(100_000, decimals), "fee tiers max volume mismatch")
	// assert.Equal(t, feeTiersEvent.FeeTiersList[0].FeeBps, 50, "fee tiers bps mismatch")

	// getTokenOut8, err := c.GetToken(tok.Address, "", "")
	// if err != nil {
	// 	t.Fatalf("GetToken: %v", err)
	// }

	// var tokenState8 tokenV1Models.TokenStateModel
	// err = utils.UnmarshalState[tokenV1Models.TokenStateModel](getTokenOut8.States[0].Object, &tokenState8)
	// if err != nil {
	// 	t.Fatalf("UnmarshalState (GetToken.States[0]): %v", err)
	// }

	// assert.Equal(t, tokenState8.FeeTiersList[0].MinAmount, "0", "token state fee tiers min amount mismatch")
	// assert.Equal(t, tokenState8.FeeTiersList[0].MaxAmount, amt(10_000, decimals), "token state fee tiers max amount mismatch")
	// assert.Equal(t, tokenState8.FeeTiersList[0].MinVolume, "0", "token state fee tiers min volume mismatch")
	// assert.Equal(t, tokenState8.FeeTiersList[0].MaxVolume, amt(100_000, decimals), "token state fee tiers max volume mismatch")
	// assert.Equal(t, tokenState8.FeeTiersList[0].FeeBps, 50, "token state fee tiers bps mismatch")

	// // Fee address
	// updateFeeAddress, er := c.UpdateFeeAddress(tok.Address, feeAddress)
	// if er != nil {
	// 	t.Fatalf("UpdateFeeAddress: %v", er)
	// }

	// unmarshalLogFeeAddress, err := utils.UnmarshalLog[log.Log](updateFeeAddress.Logs[0])
	// if err != nil {
	// 	t.Fatalf("UnmarshalLog (UpdateFeeAddress.Logs[0]): %v", err)
	// }
	// assert.Equal(t, unmarshalLogFeeAddress.LogType, tokenV1Domain.TOKEN_FEE_ADDRESS_UPDATED_LOG, "update fee address log type mismatch")

	// feeAddressEvent, err := utils.UnmarshalEvent[tokenV1Domain.Fee](unmarshalLogFeeAddress.Event)
	// if err != nil {
	// 	t.Fatalf("UnmarshalEvent (UpdateFeeAddress.Logs[0]): %v", err)
	// }
	// assert.Equal(t, feeAddressEvent.TokenAddress, tok.Address, "fee address event token address mismatch")
	// assert.Equal(t, feeAddressEvent.FeeAddress, feeAddress, "fee address event fee address mismatch")

	// getTokenOut9, err := c.GetToken(tok.Address, "", "")
	// if err != nil {
	// 	t.Fatalf("GetToken: %v", err)
	// }

	// var tokenState9 tokenV1Models.TokenStateModel
	// err = utils.UnmarshalState[tokenV1Models.TokenStateModel](getTokenOut9.States[0].Object, &tokenState9)
	// if err != nil {
	// 	t.Fatalf("UnmarshalState (GetToken.States[0]): %v", err)
	// }

	// assert.Equal(t, tokenState9.FeeAddress, feeAddress, "token state fee address mismatch")

	// // Metadata
	// updateMetadata, err := c.UpdateMetadata(
	// 	tok.Address,
	// 	"2F-NEW"+randSuffix(4),
	// 	"2Finance New",
	// 	decimals,
	// 	"Updated by tests",
	// 	"https://example.com/img.png",
	// 	"https://example.com",
	// 	map[string]string{"twitter": "https://x.com/2f"},
	// 	map[string]string{"category": "DeFi"},
	// 	map[string]string{"tag1": "e2e"},
	// 	creator,
	// 	"https://creator",
	// 	time.Now().Add(30*24*time.Hour),
	// )
	// if err != nil {
	// 	t.Fatalf("UpdateMetadata: %v", err)
	// }

	// unmarshalLogMetadata, err := utils.UnmarshalLog[log.Log](updateMetadata.Logs[0])
	// if err != nil {
	// 	t.Fatalf("UnmarshalLog (UpdateMetadata.Logs[0]): %v", err)
	// }
	// assert.Equal(t, unmarshalLogMetadata.LogType, tokenV1Domain.TOKEN_METADATA_UPDATED_LOG, "update metadata log type mismatch")

	// metadataEvent, err := utils.UnmarshalEvent[tokenV1Domain.Token](unmarshalLogMetadata.Event)
	// if err != nil {
	// 	t.Fatalf("UnmarshalEvent (UpdateMetadata.Logs[0]): %v", err)
	// }
	// assert.Equal(t, metadataEvent.Address, tok.Address, "update metadata event token address mismatch")
	// assert.Equal(t, metadataEvent.Symbol, "2F-NEW"+randSuffix(4), "update metadata event token symbol mismatch")
	// assert.Equal(t, metadataEvent.Name, "2Finance New", "update metadata event token name mismatch")
	// assert.Equal(t, metadataEvent.Decimals, decimals, "update metadata event token decimals mismatch")
	// assert.Equal(t, metadataEvent.Description, "Updated by tests", "update metadata event token description mismatch")
	// assert.Equal(t, metadataEvent.Image, "https://example.com/img.png", "update metadata event token image mismatch")
	// assert.Equal(t, metadataEvent.Website, "https://example.com", "update metadata event token website mismatch")
	// assert.Equal(t, metadataEvent.TagsSocialMedia["twitter"], "https://x.com/2f", "update metadata event token tags social mismatch")
	// assert.Equal(t, metadataEvent.TagsCategory["category"], "DeFi", "update metadata event token tags category mismatch")
	// assert.Equal(t, metadataEvent.Tags["tag1"], "e2e", "update metadata event token tags mismatch")
	// assert.Equal(t, metadataEvent.Creator, creator, "update metadata event token creator mismatch")
	// assert.Equal(t, metadataEvent.CreatorWebsite, "https://creator", "update metadata event token creator website mismatch")

	// getTokenOut10, err := c.GetToken(tok.Address, "", "")
	// if err != nil {
	// 	t.Fatalf("GetToken: %v", err)
	// }

	// var tokenState10 tokenV1Models.TokenStateModel
	// err = utils.UnmarshalState[tokenV1Models.TokenStateModel](getTokenOut10.States[0].Object, &tokenState10)
	// if err != nil {
	// 	t.Fatalf("UnmarshalState (GetToken.States[0]): %v", err)
	// }
	// assert.Equal(t, tokenState10.Symbol, "2F-NEW"+randSuffix(4), "token symbol mismatch after metadata update")
	// assert.Equal(t, tokenState10.Name, "2Finance New", "token name mismatch after metadata update")
	// assert.Equal(t, tokenState10.Description, "Updated by tests", "token description mismatch after metadata update")
	// assert.Equal(t, tokenState10.Image, "https://example.com/img.png", "token image mismatch after metadata update")
	// assert.Equal(t, tokenState10.Website, "https://example.com", "token website mismatch after metadata update")
	// assert.Equal(t, tokenState10.TagsSocialMedia["twitter"], "https://x.com/2f", "token tags social mismatch after metadata update")
	// assert.Equal(t, tokenState10.TagsCategory["category"], "DeFi", "token tags category mismatch after metadata update")
	// assert.Equal(t, tokenState10.Tags["tag1"], "e2e", "token tags mismatch after metadata update")
	// assert.Equal(t, tokenState10.Creator, creator, "token creator mismatch after metadata update")
	// assert.Equal(t, tokenState10.CreatorWebsite, "https://creator", "token creator website mismatch after metadata update")

	// // Revoke authorities
	// revokeMint, err := c.RevokeMintAuthority(tok.Address, true)
	// if err != nil {
	// 	t.Fatalf("RevokeMintAuthority: %v", err)
	// }

	// unmarshalLogRevokeMint, err := utils.UnmarshalLog[log.Log](revokeMint.Logs[0])
	// if err != nil {
	// 	t.Fatalf("UnmarshalLog (RevokeMintAuthority.Logs[0]): %v", err)
	// }
	// assert.Equal(t, unmarshalLogRevokeMint.LogType, tokenV1Domain.TOKEN_MINT_AUTHORITY_REVOKED_LOG, "revoke mint authority log type mismatch")

	// revokeMintEvent, err := utils.UnmarshalEvent[tokenV1Domain.Token](unmarshalLogRevokeMint.Event)
	// if err != nil {
	// 	t.Fatalf("UnmarshalEvent (RevokeMintAuthority.Logs[0]): %v", err)
	// }
	// assert.Equal(t, revokeMintEvent.Address, tok.Address, "revoke mint authority event token address mismatch")
	// assert.Equal(t, revokeMintEvent.MintAuthorityRevoked, true, "revoke mint authority event mint authority revoked mismatch")

	// getTokenOut11, err := c.GetToken(tok.Address, "", "")
	// if err != nil {
	// 	t.Fatalf("GetToken: %v", err)
	// }

	// var tokenState11 tokenV1Models.TokenStateModel
	// err = utils.UnmarshalState[tokenV1Models.TokenStateModel](getTokenOut11.States[0].Object, &tokenState11)
	// if err != nil {
	// 	t.Fatalf("UnmarshalState (GetToken.States[0]): %v", err)
	// }
	// assert.Equal(t, tokenState11.MintAuthorityRevoked, true, "token mint authority revoked mismatch after revoke")

	// revokeUpdate, err := c.RevokeUpdateAuthority(tok.Address, true)
	// if err != nil {
	// 	t.Fatalf("RevokeUpdateAuthority: %v", err)
	// }

	// unmarshalLogRevokeUpdate, err := utils.UnmarshalLog[log.Log](revokeUpdate.Logs[0])
	// if err != nil {
	// 	t.Fatalf("UnmarshalLog (RevokeUpdateAuthority.Logs[0]): %v", err)
	// }
	// assert.Equal(t, unmarshalLogRevokeUpdate.LogType, tokenV1Domain.TOKEN_UPDATE_AUTHORITY_REVOKED_LOG, "revoke update authority log type mismatch")

	// revokeUpdateEvent, err := utils.UnmarshalEvent[tokenV1Domain.Token](unmarshalLogRevokeUpdate.Event)
	// if err != nil {
	// 	t.Fatalf("UnmarshalEvent (RevokeUpdateAuthority.Logs[0]): %v", err)
	// }
	// assert.Equal(t, revokeUpdateEvent.Address, tok.Address, "revoke update authority event token address mismatch")
	// assert.Equal(t, revokeUpdateEvent.UpdateAuthorityRevoked, true, "revoke update authority event update authority revoked mismatch")

	// getTokenOut12, err := c.GetToken(tok.Address, "", "")
	// if err != nil {
	// 	t.Fatalf("GetToken: %v", err)
	// }

	// var tokenState12 tokenV1Models.TokenStateModel
	// err = utils.UnmarshalState[tokenV1Models.TokenStateModel](getTokenOut12.States[0].Object, &tokenState12)
	// if err != nil {
	// 	t.Fatalf("UnmarshalState (GetToken.States[0]): %v", err)
	// }
	// assert.Equal(t, tokenState12.UpdateAuthorityRevoked, true, "token update authority revoked mismatch after revoke")

	// // Freeze
	// freezeWallet, err := c.FreezeWallet(tok.Address, ownerPub)
	// if err != nil {
	// 	t.Fatalf("FreezeWallet: %v", err)
	// }

	// unmarshalLogFreeze, err := utils.UnmarshalLog[log.Log](freezeWallet.Logs[0])
	// if err != nil {
	// 	t.Fatalf("UnmarshalLog (FreezeWallet.Logs[0]): %v", err)
	// }
	// assert.Equal(t, unmarshalLogFreeze.LogType, tokenV1Domain.TOKEN_FREEZE_ACCOUNT_LOG, "freeze wallet log type mismatch")
	// freezeEvent, err := utils.UnmarshalEvent[tokenV1Domain.Freeze](unmarshalLogFreeze.Event)
	// if err != nil {
	// 	t.Fatalf("UnmarshalEvent (FreezeWallet.Logs[0]): %v", err)
	// }
	// assert.Equal(t, freezeEvent.TokenAddress, tok.Address, "freeze wallet event token address mismatch")
	// assert.Equal(t, freezeEvent.FrozenAccounts[ownerPub], ownerPub, "freeze wallet event wallet address mismatch")

	// getTokenOut13, err := c.GetToken(tok.Address, "", "")
	// if err != nil {
	// 	t.Fatalf("GetToken: %v", err)
	// }

	// var tokenState13 tokenV1Models.TokenStateModel
	// err = utils.UnmarshalState[tokenV1Models.TokenStateModel](getTokenOut13.States[0].Object, &tokenState13)
	// if err != nil {
	// 	t.Fatalf("UnmarshalState (GetToken.States[0]): %v", err)
	// }
	// assert.Equal(t, tokenState13.FrozenAccounts[ownerPub], ownerPub, "token frozen accounts mismatch after freeze")

	// // Unfreeze
	// unfreezeWallet, err := c.UnfreezeWallet(tok.Address, ownerPub)
	// if err != nil {
	// 	t.Fatalf("UnfreezeWallet: %v", err)
	// }

	// unmarshalLogUnfreeze, err := utils.UnmarshalLog[log.Log](unfreezeWallet.Logs[0])
	// if err != nil {
	// 	t.Fatalf("UnmarshalLog (UnfreezeWallet.Logs[0]): %v", err)
	// }
	// assert.Equal(t, unmarshalLogUnfreeze.LogType, tokenV1Domain.TOKEN_UNFREEZE_ACCOUNT_LOG, "unfreeze wallet log type mismatch")
	// unfreezeEvent, err := utils.UnmarshalEvent[tokenV1Domain.Freeze](unmarshalLogUnfreeze.Event)
	// if err != nil {
	// 	t.Fatalf("UnmarshalEvent (UnfreezeWallet.Logs[0]): %v", err)
	// }
	// assert.Equal(t, unfreezeEvent.TokenAddress, tok.Address, "unfreeze wallet event token address mismatch")
	// assert.Equal(t, unfreezeEvent.FrozenAccounts[ownerPub], "", "unfreeze wallet event wallet address mismatch")

	// getTokenOut14, err := c.GetToken(tok.Address, "", "")
	// if err != nil {
	// 	t.Fatalf("GetToken: %v", err)
	// }

	// var tokenState14 tokenV1Models.TokenStateModel
	// err = utils.UnmarshalState[tokenV1Models.TokenStateModel](getTokenOut14.States[0].Object, &tokenState14)
	// if err != nil {
	// 	t.Fatalf("UnmarshalState (GetToken.States[0]): %v", err)
	// }
	// assert.Equal(t, tokenState14.FrozenAccounts[ownerPub], "", "token frozen accounts mismatch after unfreeze")

	// // Balances / Listings
	// if _, err := c.GetTokenBalance(tok.Address, owner.PublicKey); err != nil {
	// 	t.Fatalf("GetTokenBalance(owner): %v", err)
	// }
	// if _, err := c.ListTokenBalances(tok.Address, "", 1, 10, true); err != nil {
	// 	t.Fatalf("ListTokenBalances: %v", err)
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
