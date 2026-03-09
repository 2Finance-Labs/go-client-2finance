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
	accessMode := tokenV1Domain.DENY_ACCESS_MODE
	accessUsers := map[string]bool{}
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
	assert.Equal(t, tok.AccessUsers[owner.PublicKey], accessUsers[owner.PublicKey], "token access policy users mismatch")
	assert.Equal(t, tok.FrozenAccounts[owner.PublicKey], frozenAccounts[owner.PublicKey], "token frozen accounts mismatch")
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
	assert.Equal(t, tokenState.AccessUsers[owner.PublicKey], accessUsers[owner.PublicKey], "token access policy users mismatch")
	assert.Equal(t, tokenState.FrozenAccounts[owner.PublicKey], frozenAccounts[owner.PublicKey], "token frozen accounts mismatch")
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
	mintToken, err := c.MintToken(tok.Address, owner.PublicKey, mintAmount, decimals, tok.TokenType)
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
	//   TRANSFER TOKEN
	// ------------------
	// receiver, receiverPriv := createWallet(t, c)
	// c.SetPrivateKey(ownerPriv)

	// transferAmount := "500000"
	// transferToken, err := c.TransferToken(tok.Address, receiver.PublicKey, transferAmount, decimals, tok.TokenType, "")
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
	// assert.Equal(t, transfer.FromAddress, owner.PublicKey, "transfer from address mismatch")
	// assert.Equal(t, transfer.ToAddress, receiver.PublicKey, "transfer to address mismatch")
	// assert.Equal(t, transfer.Amount, transferAmount, "transfer amount mismatch")
	// assert.Equal(t, transfer.TokenType, tokenType, "transfer token type mismatch")
	// assert.Equal(t, transfer.UUID, "", "transfer UUID mismatch") // Should be nil for fungible tokens

	// unmarshalLogBalance4, err := utils.UnmarshalLog[log.Log](transferToken.Logs[1])
	// if err != nil {
	// 	t.Fatalf("UnmarshalLog (TransferToken.Logs[1]): %v", err)
	// }
	// assert.Equal(t, unmarshalLogBalance4.LogType, tokenV1Domain.TOKEN_BALANCE_DECREASED_LOG, "balance decreased log type mismatch")

	// balance4, err := utils.UnmarshalEvent[tokenV1Domain.Balance](unmarshalLogBalance4.Event)
	// if err != nil {
	// 	t.Fatalf("UnmarshalEvent (TransferToken.Logs[1]): %v", err)
	// }

	// assert.Equal(t, balance4.TokenAddress, tok.Address, "balance token address mismatch")
	// assert.Equal(t, balance4.OwnerAddress, owner.PublicKey, "balance wallet address mismatch")
	// assert.Equal(t, balance4.Amount, mintAmount, "balance amount mismatch after transfer")
	// assert.Equal(t, balance4.TokenType, tokenType, "balance token type mismatch")
	// assert.Equal(t, balance4.TokenUUIDList, []string(nil), "balance token UUID list mismatch") // Should be nil for fungible tokens

	// getTokenOut10, err := c.GetToken(tok.Address, "", "")
	// if err != nil {
	// 	t.Fatalf("GetToken: %v", err)
	// }

	// var tokenState10 tokenV1Models.TokenStateModel
	// err = utils.UnmarshalState[tokenV1Models.TokenStateModel](getTokenOut10.States[0].Object, &tokenState10)
	// if err != nil {
	// 	t.Fatalf("UnmarshalState (GetToken.States[0]): %v", err)
	// }

	// subTotalSupply, err = utils.SubBigIntStrings(subTotalSupply, transferAmount)
	// if err != nil {
	// 	t.Fatalf("SubBigIntStrings: %v", err)
	// }

	// assert.Equal(t, tokenState10.TotalSupply, subTotalSupply, "token total supply mismatch after transfer")

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
	assert.Equal(t, burn.BurnFrom, owner.PublicKey, "burn from address mismatch")
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
	assert.Equal(t, balance3.OwnerAddress, owner.PublicKey, "balance wallet address mismatch")
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
	// CHANGE ACCESS MODE
	// ------------------
	changeAccessMode, err := c.ChangeAccessMode(tok.Address, tokenV1Domain.ALLOW_ACCESS_MODE)
	if err != nil {
		t.Fatalf("ChangeAccessMode: %v", err)
	}

	unmarshalLogChangeAccess, err := utils.UnmarshalLog[log.Log](changeAccessMode.Logs[0])
	if err != nil {
		t.Fatalf("UnmarshalLog (ChangeAccessMode.Logs[0]): %v", err)
	}
	assert.Equal(t, unmarshalLogChangeAccess.LogType, tokenV1Domain.TOKEN_ACCESS_MODE_CHANGED_LOG, "change access mode log type mismatch")
	changeAccess, err := utils.UnmarshalEvent[tokenV1Domain.AccessPolicy](unmarshalLogChangeAccess.Event)
	if err != nil {
		t.Fatalf("UnmarshalEvent (ChangeAccessMode.Logs[0]): %v", err)
	}
	assert.Equal(t, changeAccess.Address, tok.Address, "access policy token address mismatch")
	assert.Equal(t, changeAccess.AccessMode, tokenV1Domain.ALLOW_ACCESS_MODE, "access policy mode mismatch after change access mode")

	getTokenOut4, err := c.GetToken(tok.Address, "", "")
	if err != nil {
		t.Fatalf("GetToken: %v", err)
	}

	var tokenState4 tokenV1Models.TokenStateModel
	err = utils.UnmarshalState[tokenV1Models.TokenStateModel](getTokenOut4.States[0].Object, &tokenState4)
	if err != nil {
		t.Fatalf("UnmarshalState (GetToken.States[0]): %v", err)
	}
	assert.Equal(t, tokenV1Domain.ALLOW_ACCESS_MODE, tokenState4.AccessMode, "token access mode mismatch after change access mode")

	// ------------------
	//    ALLOW USERS
	// ------------------

	// Add allow users
	receiver, _ := createWallet(t, c)
	c.SetPrivateKey(ownerPriv)

	allowUsers, err := c.AddAllowUsers(tok.Address, domain.ALLOW_ACCESS_MODE, map[string]bool{
		receiver.PublicKey: true,
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
	assert.Equal(t, allow.AccessUsers[receiver.PublicKey], true, "access policy users mismatch after allow")

	getTokenOut5, err := c.GetToken(tok.Address, "", "")
	if err != nil {
		t.Fatalf("GetToken: %v", err)
	}

	var tokenState5 tokenV1Models.TokenStateModel
	err = utils.UnmarshalState[tokenV1Models.TokenStateModel](getTokenOut5.States[0].Object, &tokenState5)
	if err != nil {
		t.Fatalf("UnmarshalState (GetToken.States[0]): %v", err)
	}

	assert.Equal(t, tokenState5.AccessMode, tokenV1Domain.ALLOW_ACCESS_MODE, "token access mode mismatch after allow")
	assert.Equal(t, tokenState5.AccessUsers[receiver.PublicKey], true, "token access users mismatch after allow")

	// Remove allow users
	removeAllowUsers, err := c.RemoveAllowUsers(tok.Address, domain.ALLOW_ACCESS_MODE, map[string]bool{
		receiver.PublicKey: true,
	})
	if err != nil {
		t.Fatalf("RemoveAllowUsers: %v", err)
	}

	unmarshalLogRemoveAllow, err := utils.UnmarshalLog[log.Log](removeAllowUsers.Logs[0])
	if err != nil {
		t.Fatalf("UnmarshalLog (RemoveAllowUsers.Logs[0]): %v", err)
	}
	assert.Equal(t, tokenV1Domain.TOKEN_USERS_REMOVED_LOG, unmarshalLogRemoveAllow.LogType, "remove allow users log type mismatch")

	removeAllow, err := utils.UnmarshalEvent[tokenV1Domain.AccessPolicy](unmarshalLogRemoveAllow.Event)
	if err != nil {
		t.Fatalf("UnmarshalEvent (RemoveAllowUsers.Logs[0]): %v", err)
	}
	assert.Equal(t, tok.Address, removeAllow.Address, "access policy token address mismatch")
	assert.Equal(t, tokenV1Domain.ALLOW_ACCESS_MODE, removeAllow.AccessMode, "access policy mode mismatch after remove allow")
	assert.Contains(t, removeAllow.AccessUsers, receiver.PublicKey, "removed user should be present in remove event payload")
	assert.True(t, removeAllow.AccessUsers[receiver.PublicKey], "removed user payload should be true in remove event")

	getTokenOut6, err := c.GetToken(tok.Address, "", "")
	if err != nil {
		t.Fatalf("GetToken: %v", err)
	}

	var tokenState6 tokenV1Models.TokenStateModel
	err = utils.UnmarshalState[tokenV1Models.TokenStateModel](getTokenOut6.States[0].Object, &tokenState6)
	if err != nil {
		t.Fatalf("UnmarshalState (GetToken.States[0]): %v", err)
	}
	assert.Equal(t, tokenV1Domain.ALLOW_ACCESS_MODE, tokenState6.AccessMode, "token access mode mismatch after remove allow")

	_, exists := tokenState6.AccessUsers[receiver.PublicKey]
	assert.False(t, exists, "user should have been removed from token access users")

	// ------------------
	// CHANGE ACCESS MODE
	// ------------------
	changeAccessMode2, err := c.ChangeAccessMode(tok.Address, tokenV1Domain.DENY_ACCESS_MODE)
	if err != nil {
		t.Fatalf("ChangeAccessMode: %v", err)
	}

	unmarshalLogChangeAccess2, err := utils.UnmarshalLog[log.Log](changeAccessMode2.Logs[0])
	if err != nil {
		t.Fatalf("UnmarshalLog (ChangeAccessMode.Logs[0]): %v", err)
	}
	assert.Equal(t, unmarshalLogChangeAccess2.LogType, tokenV1Domain.TOKEN_ACCESS_MODE_CHANGED_LOG, "change access mode log type mismatch")
	changeAccess2, err := utils.UnmarshalEvent[tokenV1Domain.AccessPolicy](unmarshalLogChangeAccess2.Event)
	if err != nil {
		t.Fatalf("UnmarshalEvent (ChangeAccessMode.Logs[0]): %v", err)
	}
	assert.Equal(t, changeAccess2.Address, tok.Address, "access policy token address mismatch")
	assert.Equal(t, changeAccess2.AccessMode, tokenV1Domain.DENY_ACCESS_MODE, "access policy mode mismatch after change access mode")

	getTokenOut7, err := c.GetToken(tok.Address, "", "")
	if err != nil {
		t.Fatalf("GetToken: %v", err)
	}

	var tokenState7 tokenV1Models.TokenStateModel
	err = utils.UnmarshalState[tokenV1Models.TokenStateModel](getTokenOut7.States[0].Object, &tokenState7)
	if err != nil {
		t.Fatalf("UnmarshalState (GetToken.States[0]): %v", err)
	}
	assert.Equal(t, tokenState7.AccessMode, tokenV1Domain.DENY_ACCESS_MODE, "token access mode mismatch after change access mode")

	// ------------------
	//    DENY USERS
	// ------------------

	// Add deny users
	denyUsers, err := c.AddDenyUsers(tok.Address, tokenV1Domain.DENY_ACCESS_MODE, map[string]bool{
		receiver.PublicKey: true,
	})
	if err != nil {
		t.Fatalf("AddDenyUsers: %v", err)
	}

	unmarshalLogDeny, err := utils.UnmarshalLog[log.Log](denyUsers.Logs[0])
	if err != nil {
		t.Fatalf("UnmarshalLog (AddDenyUsers.Logs[0]): %v", err)
	}
	assert.Equal(t, unmarshalLogDeny.LogType, tokenV1Domain.TOKEN_USERS_ADDED_LOG, "add deny users log type mismatch")

	deny, err := utils.UnmarshalEvent[tokenV1Domain.AccessPolicy](unmarshalLogDeny.Event)
	if err != nil {
		t.Fatalf("UnmarshalEvent (AddDenyUsers.Logs[0]): %v", err)
	}
	assert.Equal(t, deny.Address, tok.Address, "access policy token address mismatch")
	assert.Equal(t, deny.AccessMode, tokenV1Domain.DENY_ACCESS_MODE, "access policy mode mismatch after add deny")
	assert.Equal(t, deny.AccessUsers[receiver.PublicKey], true, "access policy users mismatch after add deny")

	getTokenOut8, err := c.GetToken(tok.Address, "", "")
	if err != nil {
		t.Fatalf("GetToken: %v", err)
	}

	var tokenState8 tokenV1Models.TokenStateModel
	err = utils.UnmarshalState[tokenV1Models.TokenStateModel](getTokenOut8.States[0].Object, &tokenState8)
	if err != nil {
		t.Fatalf("UnmarshalState (GetToken.States[0]): %v", err)
	}
	assert.Equal(t, tokenState8.AccessMode, tokenV1Domain.DENY_ACCESS_MODE, "token access mode mismatch after add deny")
	assert.Equal(t, tokenState8.AccessUsers[receiver.PublicKey], true, "token access users mismatch after add deny")

	// Remove deny users
	removeDenyUsers, err := c.RemoveDenyUsers(tok.Address, tokenV1Domain.DENY_ACCESS_MODE, map[string]bool{
		receiver.PublicKey: true,
	})
	if err != nil {
		t.Fatalf("RemoveDenyUsers: %v", err)
	}

	unmarshalLogRemoveDeny, err := utils.UnmarshalLog[log.Log](removeDenyUsers.Logs[0])
	if err != nil {
		t.Fatalf("UnmarshalLog (RemoveDenyUsers.Logs[0]): %v", err)
	}
	assert.Equal(t, unmarshalLogRemoveDeny.LogType, tokenV1Domain.TOKEN_USERS_REMOVED_LOG, "remove deny users log type mismatch")

	removeDeny, err := utils.UnmarshalEvent[tokenV1Domain.AccessPolicy](unmarshalLogRemoveDeny.Event)
	if err != nil {
		t.Fatalf("UnmarshalEvent (RemoveDenyUsers.Logs[0]): %v", err)
	}
	assert.Equal(t, removeDeny.Address, tok.Address, "access policy token address mismatch")
	assert.Equal(t, removeDeny.AccessMode, tokenV1Domain.DENY_ACCESS_MODE, "access policy mode mismatch after remove deny")
	assert.Contains(t, removeDeny.AccessUsers, receiver.PublicKey, "removed user should be present in remove event payload")
	assert.True(t, removeDeny.AccessUsers[receiver.PublicKey], "removed user payload should be true in remove event")

	getTokenOut9, err := c.GetToken(tok.Address, "", "")
	if err != nil {
		t.Fatalf("GetToken: %v", err)
	}

	var tokenState9 tokenV1Models.TokenStateModel
	err = utils.UnmarshalState[tokenV1Models.TokenStateModel](getTokenOut9.States[0].Object, &tokenState9)
	if err != nil {
		t.Fatalf("UnmarshalState (GetToken.States[0]): %v", err)
	}
	assert.Equal(t, tokenState9.AccessMode, tokenV1Domain.DENY_ACCESS_MODE, "token access mode mismatch after remove deny")
	_, exists2 := tokenState9.AccessUsers[receiver.PublicKey]
	assert.False(t, exists2, "user should have been removed from token access users")

	// ------------------
	//       PAUSE
	// ------------------
	pauseToken, err := c.PauseToken(tok.Address, true)
	if err != nil {
		t.Fatalf("PauseToken: %v", err)
	}

	unmarshalLogPause, err := utils.UnmarshalLog[log.Log](pauseToken.Logs[0])
	if err != nil {
		t.Fatalf("UnmarshalLog (PauseToken.Logs[0]): %v", err)
	}
	assert.Equal(t, tokenV1Domain.TOKEN_PAUSED_LOG, unmarshalLogPause.LogType, "pause token log type mismatch")

	pause, err := utils.UnmarshalEvent[tokenV1Domain.PausePolicy](unmarshalLogPause.Event)
	if err != nil {
		t.Fatalf("UnmarshalEvent (PauseToken.Logs[0]): %v", err)
	}
	assert.Equal(t, tok.Address, pause.TokenAddress, "pause token address mismatch")
	assert.Equal(t, true, pause.Enabled, "token paused state mismatch after pause")

	getTokenOut11, err := c.GetToken(tok.Address, "", "")
	if err != nil {
		t.Fatalf("GetToken: %v", err)
	}

	var tokenState11 tokenV1Models.TokenStateModel
	err = utils.UnmarshalState[tokenV1Models.TokenStateModel](getTokenOut11.States[0].Object, &tokenState11)
	if err != nil {
		t.Fatalf("UnmarshalState (GetToken.States[0]): %v", err)
	}

	assert.Equal(t, true, tokenState11.Paused, "token paused state mismatch after pause")

	// ------------------
	//      UNPAUSE
	// ------------------
	unpauseToken, err := c.UnpauseToken(tok.Address, false)
	if err != nil {
		t.Fatalf("UnpauseToken: %v", err)
	}

	unmarshalLogUnpause, err := utils.UnmarshalLog[log.Log](unpauseToken.Logs[0])
	if err != nil {
		t.Fatalf("UnmarshalLog (UnpauseToken.Logs[0]): %v", err)
	}

	assert.Equal(t, tokenV1Domain.TOKEN_UNPAUSED_LOG, unmarshalLogUnpause.LogType, "unpause token log type mismatch")

	unpause, err := utils.UnmarshalEvent[tokenV1Domain.PausePolicy](unmarshalLogUnpause.Event)
	if err != nil {
		t.Fatalf("UnmarshalEvent (UnpauseToken.Logs[0]): %v", err)
	}

	assert.Equal(t, tok.Address, unpause.TokenAddress, "unpause token address mismatch")
	assert.Equal(t, false, unpause.Enabled, "token paused state mismatch after unpause")

	getTokenOut12, err := c.GetToken(tok.Address, "", "")
	if err != nil {
		t.Fatalf("GetToken: %v", err)
	}

	var tokenState12 tokenV1Models.TokenStateModel
	err = utils.UnmarshalState[tokenV1Models.TokenStateModel](getTokenOut12.States[0].Object, &tokenState12)
	if err != nil {
		t.Fatalf("UnmarshalState (GetToken.States[0]): %v", err)
	}

	assert.Equal(t, false, tokenState12.Paused, "token paused state mismatch after unpause")

	// ------------------
	//     FEE TIERS
	// ------------------
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
	assert.Equal(t, unmarshalLogFeeTiers.LogType, tokenV1Domain.TOKEN_FEE_UPDATED_LOG, "update fee tiers log type mismatch")

	feeTiersEvent, err := utils.UnmarshalEvent[tokenV1Domain.FeeTiers](unmarshalLogFeeTiers.Event)
	if err != nil {
		t.Fatalf("UnmarshalEvent (UpdateFeeTiers.Logs[0]): %v", err)
	}
	assert.Equal(t, feeTiersEvent.FeeTiersList[0].MinAmount, "0", "fee tiers min amount mismatch")
	assert.Equal(t, feeTiersEvent.FeeTiersList[0].MaxAmount, amt(10_000, decimals), "fee tiers max amount mismatch")
	assert.Equal(t, feeTiersEvent.FeeTiersList[0].MinVolume, "0", "fee tiers min volume mismatch")
	assert.Equal(t, feeTiersEvent.FeeTiersList[0].MaxVolume, amt(100_000, decimals), "fee tiers max volume mismatch")
	assert.Equal(t, feeTiersEvent.FeeTiersList[0].FeeBps, 50, "fee tiers bps mismatch")

	getTokenOut13, err := c.GetToken(tok.Address, "", "")
	if err != nil {
		t.Fatalf("GetToken: %v", err)
	}

	var tokenState13 tokenV1Models.TokenStateModel
	err = utils.UnmarshalState[tokenV1Models.TokenStateModel](getTokenOut13.States[0].Object, &tokenState13)
	if err != nil {
		t.Fatalf("UnmarshalState (GetToken.States[0]): %v", err)
	}

	assert.Equal(t, tokenState13.FeeTiersList[0].MinAmount, "0", "token state fee tiers min amount mismatch")
	assert.Equal(t, tokenState13.FeeTiersList[0].MaxAmount, amt(10_000, decimals), "token state fee tiers max amount mismatch")
	assert.Equal(t, tokenState13.FeeTiersList[0].MinVolume, "0", "token state fee tiers min volume mismatch")
	assert.Equal(t, tokenState13.FeeTiersList[0].MaxVolume, amt(100_000, decimals), "token state fee tiers max volume mismatch")
	assert.Equal(t, tokenState13.FeeTiersList[0].FeeBps, 50, "token state fee tiers bps mismatch")

	// ------------------
	//    FEE ADDRESS
	// ------------------
	updateFeeAddress, er := c.UpdateFeeAddress(tok.Address, feeAddress)
	if er != nil {
		t.Fatalf("UpdateFeeAddress: %v", er)
	}

	unmarshalLogFeeAddress, err := utils.UnmarshalLog[log.Log](updateFeeAddress.Logs[0])
	if err != nil {
		t.Fatalf("UnmarshalLog (UpdateFeeAddress.Logs[0]): %v", err)
	}
	assert.Equal(t, unmarshalLogFeeAddress.LogType, tokenV1Domain.TOKEN_FEE_ADDRESS_UPDATED_LOG, "update fee address log type mismatch")

	feeAddressEvent, err := utils.UnmarshalEvent[tokenV1Domain.Fee](unmarshalLogFeeAddress.Event)
	if err != nil {
		t.Fatalf("UnmarshalEvent (UpdateFeeAddress.Logs[0]): %v", err)
	}
	assert.Equal(t, feeAddressEvent.TokenAddress, tok.Address, "fee address event token address mismatch")
	assert.Equal(t, feeAddressEvent.FeeAddress, feeAddress, "fee address event fee address mismatch")

	getTokenOut14, err := c.GetToken(tok.Address, "", "")
	if err != nil {
		t.Fatalf("GetToken: %v", err)
	}

	var tokenState14 tokenV1Models.TokenStateModel
	err = utils.UnmarshalState[tokenV1Models.TokenStateModel](getTokenOut14.States[0].Object, &tokenState14)
	if err != nil {
		t.Fatalf("UnmarshalState (GetToken.States[0]): %v", err)
	}

	assert.Equal(t, tokenState14.FeeAddress, feeAddress, "token state fee address mismatch")

	// ------------------
	//  METADATA UPDATE
	// ------------------
	newSymbol := "2F-NEW" + randSuffix(4)

	updateMetadata, err := c.UpdateMetadata(
		tok.Address,
		newSymbol,
		"2Finance New",
		decimals,
		"Updated by tests",
		"https://example.com/img.png",
		"https://example.com",
		map[string]string{"twitter": "https://x.com/2f"},
		map[string]string{"category": "DeFi"},
		map[string]string{"tag1": "e2e"},
		creator,
		"https://creator",
		time.Now().Add(30*24*time.Hour),
	)
	if err != nil {
		t.Fatalf("UpdateMetadata: %v", err)
	}

	unmarshalLogMetadata, err := utils.UnmarshalLog[log.Log](updateMetadata.Logs[0])
	if err != nil {
		t.Fatalf("UnmarshalLog (UpdateMetadata.Logs[0]): %v", err)
	}
	assert.Equal(t, tokenV1Domain.TOKEN_METADATA_UPDATED_LOG, unmarshalLogMetadata.LogType, "update metadata log type mismatch")

	metadataEvent, err := utils.UnmarshalEvent[tokenV1Domain.Token](unmarshalLogMetadata.Event)
	if err != nil {
		t.Fatalf("UnmarshalEvent (UpdateMetadata.Logs[0]): %v", err)
	}
	assert.Equal(t, tok.Address, metadataEvent.Address, "update metadata event token address mismatch")
	assert.Equal(t, newSymbol, metadataEvent.Symbol, "update metadata event token symbol mismatch")
	assert.Equal(t, "2Finance New", metadataEvent.Name, "update metadata event token name mismatch")
	assert.Equal(t, decimals, metadataEvent.Decimals, "update metadata event token decimals mismatch")
	assert.Equal(t, "Updated by tests", metadataEvent.Description, "update metadata event token description mismatch")
	assert.Equal(t, "https://example.com/img.png", metadataEvent.Image, "update metadata event token image mismatch")
	assert.Equal(t, "https://example.com", metadataEvent.Website, "update metadata event token website mismatch")
	assert.Equal(t, "https://x.com/2f", metadataEvent.TagsSocialMedia["twitter"], "update metadata event token tags social mismatch")
	assert.Equal(t, "DeFi", metadataEvent.TagsCategory["category"], "update metadata event token tags category mismatch")
	assert.Equal(t, "e2e", metadataEvent.Tags["tag1"], "update metadata event token tags mismatch")
	assert.Equal(t, creator, metadataEvent.Creator, "update metadata event token creator mismatch")
	assert.Equal(t, "https://creator", metadataEvent.CreatorWebsite, "update metadata event token creator website mismatch")

	getTokenOut15, err := c.GetToken(tok.Address, "", "")
	if err != nil {
		t.Fatalf("GetToken: %v", err)
	}

	var tokenState15 tokenV1Models.TokenStateModel
	err = utils.UnmarshalState[tokenV1Models.TokenStateModel](getTokenOut15.States[0].Object, &tokenState15)
	if err != nil {
		t.Fatalf("UnmarshalState (GetToken.States[0]): %v", err)
	}
	assert.Equal(t, newSymbol, tokenState15.Symbol, "token symbol mismatch after metadata update")
	assert.Equal(t, "2Finance New", tokenState15.Name, "token name mismatch after metadata update")
	assert.Equal(t, "Updated by tests", tokenState15.Description, "token description mismatch after metadata update")
	assert.Equal(t, "https://example.com/img.png", tokenState15.Image, "token image mismatch after metadata update")
	assert.Equal(t, "https://example.com", tokenState15.Website, "token website mismatch after metadata update")
	assert.Equal(t, "https://x.com/2f", tokenState15.TagsSocialMedia["twitter"], "token tags social mismatch after metadata update")
	assert.Equal(t, "DeFi", tokenState15.TagsCategory["category"], "token tags category mismatch after metadata update")
	assert.Equal(t, "e2e", tokenState15.Tags["tag1"], "token tags mismatch after metadata update")
	assert.Equal(t, creator, tokenState15.Creator, "token creator mismatch after metadata update")
	assert.Equal(t, "https://creator", tokenState15.CreatorWebsite, "token creator website mismatch after metadata update")

	// ------------------
	//      FREEZE
	// ------------------
	freezeWallet, err := c.FreezeWallet(tok.Address, owner.PublicKey)
	if err != nil {
		t.Fatalf("FreezeWallet: %v", err)
	}

	unmarshalLogFreeze, err := utils.UnmarshalLog[log.Log](freezeWallet.Logs[0])
	if err != nil {
		t.Fatalf("UnmarshalLog (FreezeWallet.Logs[0]): %v", err)
	}

	assert.Equal(t, tokenV1Domain.TOKEN_FREEZE_ACCOUNT_LOG, unmarshalLogFreeze.LogType, "freeze wallet log type mismatch")

	freezeEvent, err := utils.UnmarshalEvent[tokenV1Domain.Freeze](unmarshalLogFreeze.Event)
	if err != nil {
		t.Fatalf("UnmarshalEvent (FreezeWallet.Logs[0]): %v", err)
	}

	assert.Equal(t, tok.Address, freezeEvent.TokenAddress, "freeze wallet event token address mismatch")
	assert.Equal(t, true, freezeEvent.FrozenAccounts[owner.PublicKey], "freeze wallet event wallet address mismatch")

	getTokenOut16, err := c.GetToken(tok.Address, "", "")
	if err != nil {
		t.Fatalf("GetToken: %v", err)
	}

	var tokenState16 tokenV1Models.TokenStateModel
	err = utils.UnmarshalState[tokenV1Models.TokenStateModel](getTokenOut16.States[0].Object, &tokenState16)
	if err != nil {
		t.Fatalf("UnmarshalState (GetToken.States[0]): %v", err)
	}

	assert.Equal(t, true, tokenState16.FrozenAccounts[owner.PublicKey], "token frozen accounts mismatch after freeze")

	// ------------------
	//      UNFREEZE
	// ------------------
	unfreezeWallet, err := c.UnfreezeWallet(tok.Address, owner.PublicKey)
	if err != nil {
		t.Fatalf("UnfreezeWallet: %v", err)
	}

	unmarshalLogUnfreeze, err := utils.UnmarshalLog[log.Log](unfreezeWallet.Logs[0])
	if err != nil {
		t.Fatalf("UnmarshalLog (UnfreezeWallet.Logs[0]): %v", err)
	}
	assert.Equal(t, tokenV1Domain.TOKEN_UNFREEZE_ACCOUNT_LOG, unmarshalLogUnfreeze.LogType, "unfreeze wallet log type mismatch")

	unfreezeEvent, err := utils.UnmarshalEvent[tokenV1Domain.Freeze](unmarshalLogUnfreeze.Event)
	if err != nil {
		t.Fatalf("UnmarshalEvent (UnfreezeWallet.Logs[0]): %v", err)
	}
	assert.Equal(t, tok.Address, unfreezeEvent.TokenAddress, "unfreeze wallet event token address mismatch")
	assert.Equal(t, false, unfreezeEvent.FrozenAccounts[owner.PublicKey], "unfreeze wallet event wallet address mismatch")

	getTokenOut17, err := c.GetToken(tok.Address, "", "")
	if err != nil {
		t.Fatalf("GetToken: %v", err)
	}

	var tokenState17 tokenV1Models.TokenStateModel
	err = utils.UnmarshalState[tokenV1Models.TokenStateModel](getTokenOut17.States[0].Object, &tokenState17)
	if err != nil {
		t.Fatalf("UnmarshalState (GetToken.States[0]): %v", err)
	}

	_, exists3 := tokenState17.FrozenAccounts[owner.PublicKey]
	assert.False(t, exists3, "token frozen accounts mismatch after unfreeze")

	// ------------------
	//     UPDATE GLB
	// ------------------
	newGLB := "https://example.com/assets/token.glb"
	updateGlbFile, err := c.UpdateGlbFile(tok.Address, newGLB)
	if err != nil {
		t.Fatalf("UpdateGlbFile: %v", err)
	}

	unmarshalLogUpdateGlbFile, err := utils.UnmarshalLog[log.Log](updateGlbFile.Logs[0])
	if err != nil {
		t.Fatalf("UnmarshalLog (UpdateGlbFile.Logs[0]): %v", err)
	}
	assert.Equal(t, tokenV1Domain.TOKEN_UPDATE_GLB_FILE_LOG, unmarshalLogUpdateGlbFile.LogType, "update glb file log type mismatch")

	updateGlbFileEvent, err := utils.UnmarshalEvent[tokenV1Domain.Token](unmarshalLogUpdateGlbFile.Event)
	if err != nil {
		t.Fatalf("UnmarshalEvent (UpdateGlbFile.Logs[0]): %v", err)
	}
	assert.Equal(t, tok.Address, updateGlbFileEvent.Address, "update glb file token address mismatch")
	assert.Equal(t, newGLB, updateGlbFileEvent.AssetGLBUri, "update glb file asset uri mismatch")

	getTokenOut18, err := c.GetToken(tok.Address, "", "")
	if err != nil {
		t.Fatalf("GetToken: %v", err)
	}

	var tokenState18 tokenV1Models.TokenStateModel
	err = utils.UnmarshalState[tokenV1Models.TokenStateModel](getTokenOut18.States[0].Object, &tokenState18)
	if err != nil {
		t.Fatalf("UnmarshalState (GetToken.States[0]): %v", err)
	}
	assert.Equal(t, newGLB, tokenState18.AssetGLBUri, "token asset glb uri mismatch after update")

	// ------------------
	//    TRANSFERABLE
	// ------------------
	// transferableToken, err := c.TransferableToken(tok.Address, true)
	// if err != nil {
	// 	t.Fatalf("TransferableToken: %v", err)
	// }

	// unmarshalLogTransferableToken, err := utils.UnmarshalLog[log.Log](transferableToken.Logs[0])
	// if err != nil {
	// 	t.Fatalf("UnmarshalLog (TransferableToken.Logs[0]): %v", err)
	// }
	// assert.Equal(t, tokenV1Domain.TOKEN_TRANSFERABLE_LOG, unmarshalLogTransferableToken.LogType, "transferable token log type mismatch")

	// transferableTokenEvent, err := utils.UnmarshalEvent[tokenV1Domain.Token](unmarshalLogTransferableToken.Event)
	// if err != nil {
	// 	t.Fatalf("UnmarshalEvent (TransferableToken.Logs[0]): %v", err)
	// }
	// assert.Equal(t, tok.Address, transferableTokenEvent.Address, "transferable token address mismatch")
	// assert.Equal(t, true, transferableTokenEvent.Transferable, "token transferable state mismatch after transferable")

	// getTokenOut19, err := c.GetToken(tok.Address, "", "")
	// if err != nil {
	// 	t.Fatalf("GetToken: %v", err)
	// }

	// var tokenState19 tokenV1Models.TokenStateModel
	// err = utils.UnmarshalState[tokenV1Models.TokenStateModel](getTokenOut19.States[0].Object, &tokenState19)
	// if err != nil {
	// 	t.Fatalf("UnmarshalState (GetToken.States[0]): %v", err)
	// }
	// assert.Equal(t, true, tokenState19.Transferable, "token transferable mismatch after transferable update")

	// ------------------
	// REVOKE AUTHORITIES
	// ------------------
	// Revoke freeze authority
	revokeFreezeAuthority, err := c.RevokeFreezeAuthority(tok.Address, true)
	if err != nil {
		t.Fatalf("RevokeFreezeAuthority: %v", err)
	}

	unmarshalLogRevokeFreezeAuthority, err := utils.UnmarshalLog[log.Log](revokeFreezeAuthority.Logs[0])
	if err != nil {
		t.Fatalf("UnmarshalLog (RevokeFreezeAuthority.Logs[0]): %v", err)
	}
	assert.Equal(t, tokenV1Domain.TOKEN_FREEZE_AUTHORITY_REVOKED_LOG, unmarshalLogRevokeFreezeAuthority.LogType, "revoke freeze authority log type mismatch")

	revokeFreezeAuthorityEvent, err := utils.UnmarshalEvent[tokenV1Domain.Token](unmarshalLogRevokeFreezeAuthority.Event)
	if err != nil {
		t.Fatalf("UnmarshalEvent (RevokeFreezeAuthority.Logs[0]): %v", err)
	}
	assert.Equal(t, tok.Address, revokeFreezeAuthorityEvent.Address, "revoke freeze authority token address mismatch")
	assert.Equal(t, true, revokeFreezeAuthorityEvent.FreezeAuthorityRevoked, "revoke freeze authority state mismatch")

	getTokenOut20, err := c.GetToken(tok.Address, "", "")
	if err != nil {
		t.Fatalf("GetToken: %v", err)
	}

	var tokenState20 tokenV1Models.TokenStateModel
	err = utils.UnmarshalState[tokenV1Models.TokenStateModel](getTokenOut20.States[0].Object, &tokenState20)
	if err != nil {
		t.Fatalf("UnmarshalState (GetToken.States[0]): %v", err)
	}
	assert.Equal(t, true, tokenState20.FreezeAuthorityRevoked, "token revoked freeze authority mismatch after revoke")

	// Revoke mint authority
	revokeMint, err := c.RevokeMintAuthority(tok.Address, true)
	if err != nil {
		t.Fatalf("RevokeMintAuthority: %v", err)
	}

	unmarshalLogRevokeMint, err := utils.UnmarshalLog[log.Log](revokeMint.Logs[0])
	if err != nil {
		t.Fatalf("UnmarshalLog (RevokeMintAuthority.Logs[0]): %v", err)
	}
	assert.Equal(t, unmarshalLogRevokeMint.LogType, tokenV1Domain.TOKEN_MINT_AUTHORITY_REVOKED_LOG, "revoke mint authority log type mismatch")

	revokeMintEvent, err := utils.UnmarshalEvent[tokenV1Domain.Token](unmarshalLogRevokeMint.Event)
	if err != nil {
		t.Fatalf("UnmarshalEvent (RevokeMintAuthority.Logs[0]): %v", err)
	}
	assert.Equal(t, revokeMintEvent.Address, tok.Address, "revoke mint authority event token address mismatch")
	assert.Equal(t, revokeMintEvent.MintAuthorityRevoked, true, "revoke mint authority event mint authority revoked mismatch")

	getTokenOut21, err := c.GetToken(tok.Address, "", "")
	if err != nil {
		t.Fatalf("GetToken: %v", err)
	}

	var tokenState21 tokenV1Models.TokenStateModel
	err = utils.UnmarshalState[tokenV1Models.TokenStateModel](getTokenOut21.States[0].Object, &tokenState21)
	if err != nil {
		t.Fatalf("UnmarshalState (GetToken.States[0]): %v", err)
	}
	assert.Equal(t, tokenState21.MintAuthorityRevoked, true, "token mint authority revoked mismatch after revoke")

	// Revoke update authority
	revokeUpdate, err := c.RevokeUpdateAuthority(tok.Address, true)
	if err != nil {
		t.Fatalf("RevokeUpdateAuthority: %v", err)
	}

	unmarshalLogRevokeUpdate, err := utils.UnmarshalLog[log.Log](revokeUpdate.Logs[0])
	if err != nil {
		t.Fatalf("UnmarshalLog (RevokeUpdateAuthority.Logs[0]): %v", err)
	}
	assert.Equal(t, unmarshalLogRevokeUpdate.LogType, tokenV1Domain.TOKEN_UPDATE_AUTHORITY_REVOKED_LOG, "revoke update authority log type mismatch")

	revokeUpdateEvent, err := utils.UnmarshalEvent[tokenV1Domain.Token](unmarshalLogRevokeUpdate.Event)
	if err != nil {
		t.Fatalf("UnmarshalEvent (RevokeUpdateAuthority.Logs[0]): %v", err)
	}
	assert.Equal(t, revokeUpdateEvent.Address, tok.Address, "revoke update authority event token address mismatch")
	assert.Equal(t, revokeUpdateEvent.UpdateAuthorityRevoked, true, "revoke update authority event update authority revoked mismatch")

	getTokenOut22, err := c.GetToken(tok.Address, "", "")
	if err != nil {
		t.Fatalf("GetToken: %v", err)
	}

	var tokenState22 tokenV1Models.TokenStateModel
	err = utils.UnmarshalState[tokenV1Models.TokenStateModel](getTokenOut22.States[0].Object, &tokenState22)
	if err != nil {
		t.Fatalf("UnmarshalState (GetToken.States[0]): %v", err)
	}
	assert.Equal(t, tokenState22.UpdateAuthorityRevoked, true, "token update authority revoked mismatch after revoke")

	// ------------------
	// GETTERS | LISTINGS
	// ------------------
	if _, err := c.GetTokenBalance(tok.Address, owner.PublicKey); err != nil {
		t.Fatalf("GetTokenBalance(owner): %v", err)
	}
	if _, err := c.ListTokenBalances(tok.Address, "", 1, 10, true); err != nil {
		t.Fatalf("ListTokenBalances: %v", err)
	}
	if _, err := c.ListTokens("", "", "", 1, 10, true); err != nil {
		t.Fatalf("ListTokens: %v", err)
	}
}

func TestTokenFlowNonFungible(t *testing.T) {
	c := setupClient(t)
	owner, ownerPriv := createWallet(t, c)
	c.SetPrivateKey(ownerPriv)

	deployedContract, err := c.DeployContract1(tokenV1.TOKEN_CONTRACT_V1)
	if err != nil {
		t.Fatalf("DeployContract: %v", err)
	}
	contractLog, err := utils.UnmarshalLog[log.Log](deployedContract.Logs[0])
	if err != nil {
		t.Fatalf("UnmarshalLog (DeployContract.Logs[0]): %v", err)
	}

	// ------------------
	//    CREATE TOKEN
	// ------------------
	address := contractLog.ContractAddress
	decimals := 0
	tokenType := tokenV1Domain.NON_FUNGIBLE
	stablecoin := false
	symbol := "2NFT" + randSuffix(4)
	name := "2Finance NFT"
	totalSupply := "1"
	description := "e2e non fungible token created by tests"
	image := "https://example.com/image.png"
	website := "https://example.com"
	tagsSocial := map[string]string{"twitter": "https://twitter.com/2finance"}
	tagsCat := map[string]string{"category": "Collectibles"}
	tags := map[string]string{"tag1": "NFT", "tag2": "Blockchain"}
	creator := "2Finance Test"
	creatorWebsite := "https://creator.example"
	accessMode := tokenV1Domain.DENY_ACCESS_MODE
	accessUsers := map[string]bool{}
	frozenAccounts := map[string]bool{}
	feeTiers := []map[string]interface{}{}
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
	assert.Equal(t, tokenV1Domain.TOKEN_CREATED_LOG, unmarshalLogToken.LogType, "add-token log type mismatch")

	tok, err := utils.UnmarshalEvent[tokenV1Domain.Token](unmarshalLogToken.Event)
	if err != nil {
		t.Fatalf("UnmarshalEvent (AddToken.Logs[0]): %v", err)
	}

	assert.Equal(t, address, tok.Address, "token address mismatch")
	assert.Equal(t, symbol, tok.Symbol, "token symbol mismatch")
	assert.Equal(t, name, tok.Name, "token name mismatch")
	assert.Equal(t, decimals, tok.Decimals, "token decimals mismatch")
	assert.Equal(t, totalSupply, tok.TotalSupply, "token total supply mismatch")
	assert.Equal(t, description, tok.Description, "token description mismatch")
	assert.Equal(t, image, tok.Image, "token image mismatch")
	assert.Equal(t, website, tok.Website, "token website mismatch")
	assert.Equal(t, tagsSocial["twitter"], tok.TagsSocialMedia["twitter"], "token tags social mismatch")
	assert.Equal(t, tagsCat["category"], tok.TagsCategory["category"], "token tags category mismatch")
	assert.Equal(t, tags["tag1"], tok.Tags["tag1"], "token tags mismatch")
	assert.Equal(t, creator, tok.Creator, "token creator mismatch")
	assert.Equal(t, creatorWebsite, tok.CreatorWebsite, "token creator website mismatch")
	assert.Equal(t, accessMode, tok.AccessMode, "token access mode mismatch")
	assert.Equal(t, feeAddress, tok.FeeAddress, "token fee address mismatch")
	assert.Equal(t, freezeAuthorityRevoked, tok.FreezeAuthorityRevoked, "token freeze authority revoked mismatch")
	assert.Equal(t, mintAuthorityRevoked, tok.MintAuthorityRevoked, "token mint authority revoked mismatch")
	assert.Equal(t, updateAuthorityRevoked, tok.UpdateAuthorityRevoked, "token update authority revoked mismatch")
	assert.Equal(t, paused, tok.Paused, "token paused mismatch")
	assert.Equal(t, assetGLBUri, tok.AssetGLBUri, "token asset GLB URI mismatch")
	assert.Equal(t, tokenType, tok.TokenType, "token type mismatch")
	assert.Equal(t, transferable, tok.Transferable, "token transferable mismatch")
	assert.Equal(t, stablecoin, tok.Stablecoin, "token stablecoin mismatch")

	getTokenOut, err := c.GetToken(tok.Address, "", "")
	if err != nil {
		t.Fatalf("GetToken: %v", err)
	}

	var tokenState tokenV1Models.TokenStateModel
	err = utils.UnmarshalState[tokenV1Models.TokenStateModel](getTokenOut.States[0].Object, &tokenState)
	if err != nil {
		t.Fatalf("UnmarshalState (GetToken.States[0]): %v", err)
	}

	assert.Equal(t, address, tokenState.Address, "token address mismatch")
	assert.Equal(t, symbol, tokenState.Symbol, "token symbol mismatch")
	assert.Equal(t, name, tokenState.Name, "token name mismatch")
	assert.Equal(t, decimals, tokenState.Decimals, "token decimals mismatch")
	assert.Equal(t, totalSupply, tokenState.TotalSupply, "token total supply mismatch")
	assert.Equal(t, description, tokenState.Description, "token description mismatch")
	assert.Equal(t, image, tokenState.Image, "token image mismatch")
	assert.Equal(t, website, tokenState.Website, "token website mismatch")
	assert.Equal(t, tagsSocial["twitter"], tokenState.TagsSocialMedia["twitter"], "token tags social mismatch")
	assert.Equal(t, tagsCat["category"], tokenState.TagsCategory["category"], "token tags category mismatch")
	assert.Equal(t, tags["tag1"], tokenState.Tags["tag1"], "token tags mismatch")
	assert.Equal(t, creator, tokenState.Creator, "token creator mismatch")
	assert.Equal(t, creatorWebsite, tokenState.CreatorWebsite, "token creator website mismatch")
	assert.Equal(t, accessMode, tokenState.AccessMode, "token access mode mismatch")
	assert.Equal(t, feeAddress, tokenState.FeeAddress, "token fee address mismatch")
	assert.Equal(t, freezeAuthorityRevoked, tokenState.FreezeAuthorityRevoked, "token freeze authority revoked mismatch")
	assert.Equal(t, mintAuthorityRevoked, tokenState.MintAuthorityRevoked, "token mint authority revoked mismatch")
	assert.Equal(t, updateAuthorityRevoked, tokenState.UpdateAuthorityRevoked, "token update authority revoked mismatch")
	assert.Equal(t, paused, tokenState.Paused, "token paused mismatch")
	assert.Equal(t, assetGLBUri, tokenState.AssetGLBUri, "token asset GLB URI mismatch")
	assert.Equal(t, tokenType, tokenState.TokenType, "token type mismatch")
	assert.Equal(t, transferable, tokenState.Transferable, "token transferable mismatch")
	assert.Equal(t, stablecoin, tokenState.Stablecoin, "token stablecoin mismatch")

	// ------------------
	//        MINT
	// ------------------
	// mintAmount := "1"
	// mintToken, err := c.MintToken(tok.Address, owner.PublicKey, mintAmount, decimals, tok.TokenType)
	// if err != nil {
	// 	t.Fatalf("MintToken: %v", err)
	// }

	// unmarshalLogMint, err := utils.UnmarshalLog[log.Log](mintToken.Logs[0])
	// if err != nil {
	// 	t.Fatalf("UnmarshalLog (MintToken.Logs[0]): %v", err)
	// }
	// assert.Equal(t, tokenV1Domain.TOKEN_MINTED_LOG, unmarshalLogMint.LogType, "mint log type mismatch")

	// mint, err := utils.UnmarshalEvent[tokenV1Domain.Mint](unmarshalLogMint.Event)
	// if err != nil {
	// 	t.Fatalf("UnmarshalEvent (MintToken.Logs[0]): %v", err)
	// }

	// assert.Equal(t, tok.Address, mint.TokenAddress, "mint token address mismatch")
	// assert.Equal(t, owner.PublicKey, mint.MintTo, "mint to address mismatch")
	// assert.Equal(t, mintAmount, mint.Amount, "mint amount mismatch")
	// assert.Equal(t, tokenType, mint.TokenType, "mint token type mismatch")
	// require.Len(t, mint.TokenUUIDList, 1, "mint token uuid list mismatch")

	// mintedUUID := mint.TokenUUIDList[0]
	// assert.NotEmpty(t, mintedUUID, "minted UUID should not be empty")

	// unmarshalLogSupply, err := utils.UnmarshalLog[log.Log](mintToken.Logs[1])
	// if err != nil {
	// 	t.Fatalf("UnmarshalLog (MintToken.Logs[1]): %v", err)
	// }
	// assert.Equal(t, tokenV1Domain.TOKEN_TOTAL_SUPPLY_INCREASED_LOG, unmarshalLogSupply.LogType, "supply log type mismatch")

	// supply, err := utils.UnmarshalEvent[tokenV1Domain.Supply](unmarshalLogSupply.Event)
	// if err != nil {
	// 	t.Fatalf("UnmarshalEvent (MintToken.Logs[1]): %v", err)
	// }
	// assert.Equal(t, tok.Address, supply.TokenAddress, "supply token address mismatch")
	// assert.Equal(t, mintAmount, supply.Amount, "total supply mismatch after mint")

	// unmarshalLogBalance, err := utils.UnmarshalLog[log.Log](mintToken.Logs[2])
	// if err != nil {
	// 	t.Fatalf("UnmarshalLog (MintToken.Logs[2]): %v", err)
	// }
	// assert.Equal(t, tokenV1Domain.TOKEN_BALANCE_INCREASED_LOG, unmarshalLogBalance.LogType, "balance log type mismatch")

	// balance, err := utils.UnmarshalEvent[tokenV1Domain.Balance](unmarshalLogBalance.Event)
	// if err != nil {
	// 	t.Fatalf("UnmarshalEvent (MintToken.Logs[2]): %v", err)
	// }

	// assert.Equal(t, tok.Address, balance.TokenAddress, "balance token address mismatch")
	// assert.Equal(t, owner.PublicKey, balance.OwnerAddress, "balance wallet address mismatch")
	// assert.Equal(t, mintAmount, balance.Amount, "balance amount mismatch after mint")
	// assert.Equal(t, tokenType, balance.TokenType, "balance token type mismatch")
	// require.Len(t, balance.TokenUUIDList, 1, "balance token UUID list mismatch after mint")
	// assert.Equal(t, mintedUUID, balance.TokenUUIDList[0], "balance token UUID mismatch after mint")

	// getTokenOut2, err := c.GetToken(tok.Address, "", "")
	// if err != nil {
	// 	t.Fatalf("GetToken: %v", err)
	// }

	// var tokenState2 tokenV1Models.TokenStateModel
	// err = utils.UnmarshalState[tokenV1Models.TokenStateModel](getTokenOut2.States[0].Object, &tokenState2)
	// if err != nil {
	// 	t.Fatalf("UnmarshalState (GetToken.States[0]): %v", err)
	// }
	// assert.Equal(t, "1", tokenState2.TotalSupply, "token total supply mismatch after mint")

	// ------------------
	//   TRANSFER TOKEN
	// ------------------
	receiver, _ := createWallet(t, c)
	c.SetPrivateKey(ownerPriv)

	// transferAmount := "1"
	// transferToken, err := c.TransferToken(tok.Address, receiver.PublicKey, transferAmount, decimals, tok.TokenType, mintedUUID)
	// if err != nil {
	// 	t.Fatalf("TransferToken: %v", err)
	// }

	// unmarshalLogTransfer, err := utils.UnmarshalLog[log.Log](transferToken.Logs[0])
	// if err != nil {
	// 	t.Fatalf("UnmarshalLog (TransferToken.Logs[0]): %v", err)
	// }
	// assert.Equal(t, tokenV1Domain.TOKEN_TRANSFERRED_LOG, unmarshalLogTransfer.LogType, "transfer log type mismatch")

	// transfer, err := utils.UnmarshalEvent[tokenV1Domain.Transfer](unmarshalLogTransfer.Event)
	// if err != nil {
	// 	t.Fatalf("UnmarshalEvent (TransferToken.Logs[0]): %v", err)
	// }

	// assert.Equal(t, tok.Address, transfer.TokenAddress, "transfer token address mismatch")
	// assert.Equal(t, owner.PublicKey, transfer.FromAddress, "transfer from address mismatch")
	// assert.Equal(t, receiver.PublicKey, transfer.ToAddress, "transfer to address mismatch")
	// assert.Equal(t, transferAmount, transfer.Amount, "transfer amount mismatch")
	// assert.Equal(t, tokenType, transfer.TokenType, "transfer token type mismatch")
	// assert.Equal(t, mintedUUID, transfer.UUID, "transfer UUID mismatch")

	// unmarshalLogBalanceFrom, err := utils.UnmarshalLog[log.Log](transferToken.Logs[1])
	// if err != nil {
	// 	t.Fatalf("UnmarshalLog (TransferToken.Logs[1]): %v", err)
	// }
	// assert.Equal(t, tokenV1Domain.TOKEN_BALANCE_DECREASED_LOG, unmarshalLogBalanceFrom.LogType, "sender balance decreased log type mismatch")

	// balanceFrom, err := utils.UnmarshalEvent[tokenV1Domain.Balance](unmarshalLogBalanceFrom.Event)
	// if err != nil {
	// 	t.Fatalf("UnmarshalEvent (TransferToken.Logs[1]): %v", err)
	// }
	// assert.Equal(t, tok.Address, balanceFrom.TokenAddress, "sender balance token address mismatch")
	// assert.Equal(t, owner.PublicKey, balanceFrom.OwnerAddress, "sender balance owner mismatch")
	// assert.Equal(t, "0", balanceFrom.Amount, "sender balance amount mismatch after transfer")
	// assert.Equal(t, tokenType, balanceFrom.TokenType, "sender balance token type mismatch")

	// unmarshalLogBalanceTo, err := utils.UnmarshalLog[log.Log](transferToken.Logs[2])
	// if err != nil {
	// 	t.Fatalf("UnmarshalLog (TransferToken.Logs[2]): %v", err)
	// }
	// assert.Equal(t, tokenV1Domain.TOKEN_BALANCE_INCREASED_LOG, unmarshalLogBalanceTo.LogType, "receiver balance increased log type mismatch")

	// balanceTo, err := utils.UnmarshalEvent[tokenV1Domain.Balance](unmarshalLogBalanceTo.Event)
	// if err != nil {
	// 	t.Fatalf("UnmarshalEvent (TransferToken.Logs[2]): %v", err)
	// }
	// assert.Equal(t, tok.Address, balanceTo.TokenAddress, "receiver balance token address mismatch")
	// assert.Equal(t, receiver.PublicKey, balanceTo.OwnerAddress, "receiver balance owner mismatch")
	// assert.Equal(t, "1", balanceTo.Amount, "receiver balance amount mismatch after transfer")
	// assert.Equal(t, tokenType, balanceTo.TokenType, "receiver balance token type mismatch")
	// require.Len(t, balanceTo.TokenUUIDList, 1, "receiver balance token uuid list mismatch")
	// assert.Equal(t, mintedUUID, balanceTo.TokenUUIDList[0], "receiver balance token uuid mismatch")

	// ------------------
	//        BURN
	// ------------------
	// c.SetPrivateKey(receiverPriv)

	// burnToken, err := c.BurnToken(tok.Address, "1", decimals, tok.TokenType, mintedUUID)
	// if err != nil {
	// 	t.Fatalf("BurnToken: %v", err)
	// }

	// unmarshalLogBurn, err := utils.UnmarshalLog[log.Log](burnToken.Logs[0])
	// if err != nil {
	// 	t.Fatalf("UnmarshalLog (BurnToken.Logs[0]): %v", err)
	// }
	// assert.Equal(t, tokenV1Domain.TOKEN_BURNED_LOG, unmarshalLogBurn.LogType, "burn log type mismatch")

	// burn, err := utils.UnmarshalEvent[tokenV1Domain.Burn](unmarshalLogBurn.Event)
	// if err != nil {
	// 	t.Fatalf("UnmarshalEvent (BurnToken.Logs[0]): %v", err)
	// }

	// assert.Equal(t, tok.Address, burn.TokenAddress, "burn token address mismatch")
	// assert.Equal(t, receiver.PublicKey, burn.BurnFrom, "burn from address mismatch")
	// assert.Equal(t, "1", burn.Amount, "burn amount mismatch")
	// assert.Equal(t, tokenType, burn.TokenType, "burn token type mismatch")
	// assert.Equal(t, mintedUUID, burn.UUID, "burn token UUID mismatch")

	// unmarshalLogSupply2, err := utils.UnmarshalLog[log.Log](burnToken.Logs[1])
	// if err != nil {
	// 	t.Fatalf("UnmarshalLog (BurnToken.Logs[1]): %v", err)
	// }
	// assert.Equal(t, tokenV1Domain.TOKEN_TOTAL_SUPPLY_DECREASED_LOG, unmarshalLogSupply2.LogType, "supply log type mismatch")

	// supply2, err := utils.UnmarshalEvent[tokenV1Domain.Supply](unmarshalLogSupply2.Event)
	// if err != nil {
	// 	t.Fatalf("UnmarshalEvent (BurnToken.Logs[1]): %v", err)
	// }
	// assert.Equal(t, tok.Address, supply2.TokenAddress, "supply token address mismatch")
	// assert.Equal(t, "1", supply2.Amount, "total supply mismatch after burn")

	// unmarshalLogBalance2, err := utils.UnmarshalLog[log.Log](burnToken.Logs[2])
	// if err != nil {
	// 	t.Fatalf("UnmarshalLog (BurnToken.Logs[2]): %v", err)
	// }
	// assert.Equal(t, tokenV1Domain.TOKEN_BALANCE_DECREASED_LOG, unmarshalLogBalance2.LogType, "balance log type mismatch")

	// balance2, err := utils.UnmarshalEvent[tokenV1Domain.Balance](unmarshalLogBalance2.Event)
	// if err != nil {
	// 	t.Fatalf("UnmarshalEvent (BurnToken.Logs[2]): %v", err)
	// }

	// assert.Equal(t, tok.Address, balance2.TokenAddress, "balance token address mismatch")
	// assert.Equal(t, receiver.PublicKey, balance2.OwnerAddress, "balance wallet address mismatch")
	// assert.Equal(t, "0", balance2.Amount, "balance amount mismatch after burn")
	// assert.Equal(t, tokenType, balance2.TokenType, "balance token type mismatch")

	// getTokenOut3, err := c.GetToken(tok.Address, "", "")
	// if err != nil {
	// 	t.Fatalf("GetToken: %v", err)
	// }

	// var tokenState3 tokenV1Models.TokenStateModel
	// err = utils.UnmarshalState[tokenV1Models.TokenStateModel](getTokenOut3.States[0].Object, &tokenState3)
	// if err != nil {
	// 	t.Fatalf("UnmarshalState (GetToken.States[0]): %v", err)
	// }

	// sumTotalSupply, err := utils.AddBigIntStrings(totalSupply, mintAmount)
	// if err != nil {
	// 	t.Fatalf("AddBigIntStrings: %v", err)
	// }

	// assert.Equal(t, sumTotalSupply, tokenState2.TotalSupply, "token total supply mismatch after mint")

	// c.SetPrivateKey(ownerPriv)

	// ------------------
	// CHANGE ACCESS MODE
	// ------------------
	changeAccessMode, err := c.ChangeAccessMode(tok.Address, tokenV1Domain.ALLOW_ACCESS_MODE)
	if err != nil {
		t.Fatalf("ChangeAccessMode: %v", err)
	}

	unmarshalLogChangeAccess, err := utils.UnmarshalLog[log.Log](changeAccessMode.Logs[0])
	if err != nil {
		t.Fatalf("UnmarshalLog (ChangeAccessMode.Logs[0]): %v", err)
	}
	assert.Equal(t, tokenV1Domain.TOKEN_ACCESS_MODE_CHANGED_LOG, unmarshalLogChangeAccess.LogType, "change access mode log type mismatch")

	changeAccess, err := utils.UnmarshalEvent[tokenV1Domain.AccessPolicy](unmarshalLogChangeAccess.Event)
	if err != nil {
		t.Fatalf("UnmarshalEvent (ChangeAccessMode.Logs[0]): %v", err)
	}
	assert.Equal(t, tok.Address, changeAccess.Address, "access policy token address mismatch")
	assert.Equal(t, tokenV1Domain.ALLOW_ACCESS_MODE, changeAccess.AccessMode, "access policy mode mismatch after change access mode")

	getTokenOut4, err := c.GetToken(tok.Address, "", "")
	if err != nil {
		t.Fatalf("GetToken: %v", err)
	}

	var tokenState4 tokenV1Models.TokenStateModel
	err = utils.UnmarshalState[tokenV1Models.TokenStateModel](getTokenOut4.States[0].Object, &tokenState4)
	if err != nil {
		t.Fatalf("UnmarshalState (GetToken.States[0]): %v", err)
	}
	assert.Equal(t, tokenV1Domain.ALLOW_ACCESS_MODE, tokenState4.AccessMode, "token access mode mismatch after change access mode")

	// ------------------
	//    ALLOW USERS
	// ------------------
	allowUsers, err := c.AddAllowUsers(tok.Address, tokenV1Domain.ALLOW_ACCESS_MODE, map[string]bool{
		receiver.PublicKey: true,
	})
	if err != nil {
		t.Fatalf("AddAllowUsers: %v", err)
	}

	unmarshalLogAllow, err := utils.UnmarshalLog[log.Log](allowUsers.Logs[0])
	if err != nil {
		t.Fatalf("UnmarshalLog (AddAllowUsers.Logs[0]): %v", err)
	}
	assert.Equal(t, tokenV1Domain.TOKEN_USERS_ADDED_LOG, unmarshalLogAllow.LogType, "allow users log type mismatch")

	allow, err := utils.UnmarshalEvent[tokenV1Domain.AccessPolicy](unmarshalLogAllow.Event)
	if err != nil {
		t.Fatalf("UnmarshalEvent (AddAllowUsers.Logs[0]): %v", err)
	}
	assert.Equal(t, tok.Address, allow.Address, "access policy token address mismatch")
	assert.Equal(t, tokenV1Domain.ALLOW_ACCESS_MODE, allow.AccessMode, "access policy mode mismatch after allow")
	assert.Equal(t, true, allow.AccessUsers[receiver.PublicKey], "access policy users mismatch after allow")

	getTokenOut5, err := c.GetToken(tok.Address, "", "")
	if err != nil {
		t.Fatalf("GetToken: %v", err)
	}

	var tokenState5 tokenV1Models.TokenStateModel
	err = utils.UnmarshalState[tokenV1Models.TokenStateModel](getTokenOut5.States[0].Object, &tokenState5)
	if err != nil {
		t.Fatalf("UnmarshalState (GetToken.States[0]): %v", err)
	}
	assert.Equal(t, tokenV1Domain.ALLOW_ACCESS_MODE, tokenState5.AccessMode, "token access mode mismatch after allow")
	assert.Equal(t, true, tokenState5.AccessUsers[receiver.PublicKey], "token access users mismatch after allow")

	removeAllowUsers, err := c.RemoveAllowUsers(tok.Address, tokenV1Domain.ALLOW_ACCESS_MODE, map[string]bool{
		receiver.PublicKey: true,
	})
	if err != nil {
		t.Fatalf("RemoveAllowUsers: %v", err)
	}

	unmarshalLogRemoveAllow, err := utils.UnmarshalLog[log.Log](removeAllowUsers.Logs[0])
	if err != nil {
		t.Fatalf("UnmarshalLog (RemoveAllowUsers.Logs[0]): %v", err)
	}
	assert.Equal(t, tokenV1Domain.TOKEN_USERS_REMOVED_LOG, unmarshalLogRemoveAllow.LogType, "remove allow users log type mismatch")

	removeAllow, err := utils.UnmarshalEvent[tokenV1Domain.AccessPolicy](unmarshalLogRemoveAllow.Event)
	if err != nil {
		t.Fatalf("UnmarshalEvent (RemoveAllowUsers.Logs[0]): %v", err)
	}
	assert.Equal(t, tok.Address, removeAllow.Address, "access policy token address mismatch")
	assert.Equal(t, tokenV1Domain.ALLOW_ACCESS_MODE, removeAllow.AccessMode, "access policy mode mismatch after remove allow")
	assert.Contains(t, removeAllow.AccessUsers, receiver.PublicKey, "removed user should be present in remove event payload")
	assert.True(t, removeAllow.AccessUsers[receiver.PublicKey], "removed user payload should be true in remove event")

	getTokenOut6, err := c.GetToken(tok.Address, "", "")
	if err != nil {
		t.Fatalf("GetToken: %v", err)
	}

	var tokenState6 tokenV1Models.TokenStateModel
	err = utils.UnmarshalState[tokenV1Models.TokenStateModel](getTokenOut6.States[0].Object, &tokenState6)
	if err != nil {
		t.Fatalf("UnmarshalState (GetToken.States[0]): %v", err)
	}
	_, exists := tokenState6.AccessUsers[receiver.PublicKey]
	assert.False(t, exists, "user should have been removed from token access users")

	// ------------------
	// CHANGE ACCESS MODE
	// ------------------
	changeAccessMode2, err := c.ChangeAccessMode(tok.Address, tokenV1Domain.DENY_ACCESS_MODE)
	if err != nil {
		t.Fatalf("ChangeAccessMode: %v", err)
	}

	unmarshalLogChangeAccess2, err := utils.UnmarshalLog[log.Log](changeAccessMode2.Logs[0])
	if err != nil {
		t.Fatalf("UnmarshalLog (ChangeAccessMode.Logs[0]): %v", err)
	}
	assert.Equal(t, tokenV1Domain.TOKEN_ACCESS_MODE_CHANGED_LOG, unmarshalLogChangeAccess2.LogType, "change access mode log type mismatch")

	changeAccess2, err := utils.UnmarshalEvent[tokenV1Domain.AccessPolicy](unmarshalLogChangeAccess2.Event)
	if err != nil {
		t.Fatalf("UnmarshalEvent (ChangeAccessMode.Logs[0]): %v", err)
	}
	assert.Equal(t, tok.Address, changeAccess2.Address, "access policy token address mismatch")
	assert.Equal(t, tokenV1Domain.DENY_ACCESS_MODE, changeAccess2.AccessMode, "access policy mode mismatch after change access mode")

	getTokenOut7, err := c.GetToken(tok.Address, "", "")
	if err != nil {
		t.Fatalf("GetToken: %v", err)
	}

	var tokenState7 tokenV1Models.TokenStateModel
	err = utils.UnmarshalState[tokenV1Models.TokenStateModel](getTokenOut7.States[0].Object, &tokenState7)
	if err != nil {
		t.Fatalf("UnmarshalState (GetToken.States[0]): %v", err)
	}
	assert.Equal(t, tokenV1Domain.DENY_ACCESS_MODE, tokenState7.AccessMode, "token access mode mismatch after change access mode")

	// ------------------
	//    DENY USERS
	// ------------------
	denyUsers, err := c.AddDenyUsers(tok.Address, tokenV1Domain.DENY_ACCESS_MODE, map[string]bool{
		receiver.PublicKey: true,
	})
	if err != nil {
		t.Fatalf("AddDenyUsers: %v", err)
	}

	unmarshalLogDeny, err := utils.UnmarshalLog[log.Log](denyUsers.Logs[0])
	if err != nil {
		t.Fatalf("UnmarshalLog (AddDenyUsers.Logs[0]): %v", err)
	}
	assert.Equal(t, tokenV1Domain.TOKEN_USERS_ADDED_LOG, unmarshalLogDeny.LogType, "add deny users log type mismatch")

	deny, err := utils.UnmarshalEvent[tokenV1Domain.AccessPolicy](unmarshalLogDeny.Event)
	if err != nil {
		t.Fatalf("UnmarshalEvent (AddDenyUsers.Logs[0]): %v", err)
	}
	assert.Equal(t, tok.Address, deny.Address, "access policy token address mismatch")
	assert.Equal(t, tokenV1Domain.DENY_ACCESS_MODE, deny.AccessMode, "access policy mode mismatch after add deny")
	assert.Equal(t, true, deny.AccessUsers[receiver.PublicKey], "access policy users mismatch after add deny")

	getTokenOut8, err := c.GetToken(tok.Address, "", "")
	if err != nil {
		t.Fatalf("GetToken: %v", err)
	}

	var tokenState8 tokenV1Models.TokenStateModel
	err = utils.UnmarshalState[tokenV1Models.TokenStateModel](getTokenOut8.States[0].Object, &tokenState8)
	if err != nil {
		t.Fatalf("UnmarshalState (GetToken.States[0]): %v", err)
	}
	assert.Equal(t, tokenV1Domain.DENY_ACCESS_MODE, tokenState8.AccessMode, "token access mode mismatch after add deny")
	assert.Equal(t, true, tokenState8.AccessUsers[receiver.PublicKey], "token access users mismatch after add deny")

	removeDenyUsers, err := c.RemoveDenyUsers(tok.Address, tokenV1Domain.DENY_ACCESS_MODE, map[string]bool{
		receiver.PublicKey: true,
	})
	if err != nil {
		t.Fatalf("RemoveDenyUsers: %v", err)
	}

	unmarshalLogRemoveDeny, err := utils.UnmarshalLog[log.Log](removeDenyUsers.Logs[0])
	if err != nil {
		t.Fatalf("UnmarshalLog (RemoveDenyUsers.Logs[0]): %v", err)
	}
	assert.Equal(t, tokenV1Domain.TOKEN_USERS_REMOVED_LOG, unmarshalLogRemoveDeny.LogType, "remove deny users log type mismatch")

	removeDeny, err := utils.UnmarshalEvent[tokenV1Domain.AccessPolicy](unmarshalLogRemoveDeny.Event)
	if err != nil {
		t.Fatalf("UnmarshalEvent (RemoveDenyUsers.Logs[0]): %v", err)
	}
	assert.Equal(t, tok.Address, removeDeny.Address, "access policy token address mismatch")
	assert.Equal(t, tokenV1Domain.DENY_ACCESS_MODE, removeDeny.AccessMode, "access policy mode mismatch after remove deny")
	assert.Contains(t, removeDeny.AccessUsers, receiver.PublicKey, "removed user should be present in remove event payload")
	assert.True(t, removeDeny.AccessUsers[receiver.PublicKey], "removed user payload should be true in remove event")

	getTokenOut9, err := c.GetToken(tok.Address, "", "")
	if err != nil {
		t.Fatalf("GetToken: %v", err)
	}

	var tokenState9 tokenV1Models.TokenStateModel
	err = utils.UnmarshalState[tokenV1Models.TokenStateModel](getTokenOut9.States[0].Object, &tokenState9)
	if err != nil {
		t.Fatalf("UnmarshalState (GetToken.States[0]): %v", err)
	}
	_, exists2 := tokenState9.AccessUsers[receiver.PublicKey]
	assert.False(t, exists2, "user should have been removed from token access users")

	// ------------------
	//       PAUSE
	// ------------------
	pauseToken, err := c.PauseToken(tok.Address, true)
	if err != nil {
		t.Fatalf("PauseToken: %v", err)
	}

	unmarshalLogPause, err := utils.UnmarshalLog[log.Log](pauseToken.Logs[0])
	if err != nil {
		t.Fatalf("UnmarshalLog (PauseToken.Logs[0]): %v", err)
	}
	assert.Equal(t, tokenV1Domain.TOKEN_PAUSED_LOG, unmarshalLogPause.LogType, "pause token log type mismatch")

	pause, err := utils.UnmarshalEvent[tokenV1Domain.PausePolicy](unmarshalLogPause.Event)
	if err != nil {
		t.Fatalf("UnmarshalEvent (PauseToken.Logs[0]): %v", err)
	}
	assert.Equal(t, tok.Address, pause.TokenAddress, "pause token address mismatch")
	assert.Equal(t, true, pause.Enabled, "token paused state mismatch after pause")

	getTokenOut10, err := c.GetToken(tok.Address, "", "")
	if err != nil {
		t.Fatalf("GetToken: %v", err)
	}

	var tokenState10 tokenV1Models.TokenStateModel
	err = utils.UnmarshalState[tokenV1Models.TokenStateModel](getTokenOut10.States[0].Object, &tokenState10)
	if err != nil {
		t.Fatalf("UnmarshalState (GetToken.States[0]): %v", err)
	}
	assert.Equal(t, true, tokenState10.Paused, "token paused state mismatch after pause")

	// ------------------
	//      UNPAUSE
	// ------------------
	unpauseToken, err := c.UnpauseToken(tok.Address, false)
	if err != nil {
		t.Fatalf("UnpauseToken: %v", err)
	}

	unmarshalLogUnpause, err := utils.UnmarshalLog[log.Log](unpauseToken.Logs[0])
	if err != nil {
		t.Fatalf("UnmarshalLog (UnpauseToken.Logs[0]): %v", err)
	}
	assert.Equal(t, tokenV1Domain.TOKEN_UNPAUSED_LOG, unmarshalLogUnpause.LogType, "unpause token log type mismatch")

	unpause, err := utils.UnmarshalEvent[tokenV1Domain.PausePolicy](unmarshalLogUnpause.Event)
	if err != nil {
		t.Fatalf("UnmarshalEvent (UnpauseToken.Logs[0]): %v", err)
	}
	assert.Equal(t, tok.Address, unpause.TokenAddress, "unpause token address mismatch")
	assert.Equal(t, false, unpause.Enabled, "token paused state mismatch after unpause")

	getTokenOut11, err := c.GetToken(tok.Address, "", "")
	if err != nil {
		t.Fatalf("GetToken: %v", err)
	}

	var tokenState11 tokenV1Models.TokenStateModel
	err = utils.UnmarshalState[tokenV1Models.TokenStateModel](getTokenOut11.States[0].Object, &tokenState11)
	if err != nil {
		t.Fatalf("UnmarshalState (GetToken.States[0]): %v", err)
	}
	assert.Equal(t, false, tokenState11.Paused, "token paused state mismatch after unpause")

	// ------------------
	//     FEE TIERS
	// ------------------
	updateFeeTiers, er := c.UpdateFeeTiers(tok.Address, []map[string]interface{}{
		{
			"min_amount": "0",
			"max_amount": "1",
			"min_volume": "0",
			"max_volume": "100",
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
	assert.Equal(t, tokenV1Domain.TOKEN_FEE_UPDATED_LOG, unmarshalLogFeeTiers.LogType, "update fee tiers log type mismatch")

	feeTiersEvent, err := utils.UnmarshalEvent[tokenV1Domain.FeeTiers](unmarshalLogFeeTiers.Event)
	if err != nil {
		t.Fatalf("UnmarshalEvent (UpdateFeeTiers.Logs[0]): %v", err)
	}
	assert.Equal(t, "0", feeTiersEvent.FeeTiersList[0].MinAmount, "fee tiers min amount mismatch")
	assert.Equal(t, "1", feeTiersEvent.FeeTiersList[0].MaxAmount, "fee tiers max amount mismatch")
	assert.Equal(t, "0", feeTiersEvent.FeeTiersList[0].MinVolume, "fee tiers min volume mismatch")
	assert.Equal(t, "100", feeTiersEvent.FeeTiersList[0].MaxVolume, "fee tiers max volume mismatch")
	assert.Equal(t, 50, feeTiersEvent.FeeTiersList[0].FeeBps, "fee tiers bps mismatch")

	getTokenOut12, err := c.GetToken(tok.Address, "", "")
	if err != nil {
		t.Fatalf("GetToken: %v", err)
	}

	var tokenState12 tokenV1Models.TokenStateModel
	err = utils.UnmarshalState[tokenV1Models.TokenStateModel](getTokenOut12.States[0].Object, &tokenState12)
	if err != nil {
		t.Fatalf("UnmarshalState (GetToken.States[0]): %v", err)
	}
	assert.Equal(t, "0", tokenState12.FeeTiersList[0].MinAmount, "token state fee tiers min amount mismatch")
	assert.Equal(t, "1", tokenState12.FeeTiersList[0].MaxAmount, "token state fee tiers max amount mismatch")
	assert.Equal(t, "0", tokenState12.FeeTiersList[0].MinVolume, "token state fee tiers min volume mismatch")
	assert.Equal(t, "100", tokenState12.FeeTiersList[0].MaxVolume, "token state fee tiers max volume mismatch")
	assert.Equal(t, 50, tokenState12.FeeTiersList[0].FeeBps, "token state fee tiers bps mismatch")

	// ------------------
	//    FEE ADDRESS
	// ------------------
	updateFeeAddress, er := c.UpdateFeeAddress(tok.Address, feeAddress)
	if er != nil {
		t.Fatalf("UpdateFeeAddress: %v", er)
	}

	unmarshalLogFeeAddress, err := utils.UnmarshalLog[log.Log](updateFeeAddress.Logs[0])
	if err != nil {
		t.Fatalf("UnmarshalLog (UpdateFeeAddress.Logs[0]): %v", err)
	}
	assert.Equal(t, tokenV1Domain.TOKEN_FEE_ADDRESS_UPDATED_LOG, unmarshalLogFeeAddress.LogType, "update fee address log type mismatch")

	feeAddressEvent, err := utils.UnmarshalEvent[tokenV1Domain.Fee](unmarshalLogFeeAddress.Event)
	if err != nil {
		t.Fatalf("UnmarshalEvent (UpdateFeeAddress.Logs[0]): %v", err)
	}
	assert.Equal(t, tok.Address, feeAddressEvent.TokenAddress, "fee address event token address mismatch")
	assert.Equal(t, feeAddress, feeAddressEvent.FeeAddress, "fee address event fee address mismatch")

	getTokenOut13, err := c.GetToken(tok.Address, "", "")
	if err != nil {
		t.Fatalf("GetToken: %v", err)
	}

	var tokenState13 tokenV1Models.TokenStateModel
	err = utils.UnmarshalState[tokenV1Models.TokenStateModel](getTokenOut13.States[0].Object, &tokenState13)
	if err != nil {
		t.Fatalf("UnmarshalState (GetToken.States[0]): %v", err)
	}
	assert.Equal(t, feeAddress, tokenState13.FeeAddress, "token state fee address mismatch")

	// ------------------
	//  METADATA UPDATE
	// ------------------
	newSymbol := "2NFT-NEW" + randSuffix(4)

	updateMetadata, err := c.UpdateMetadata(
		tok.Address,
		newSymbol,
		"2Finance NFT New",
		decimals,
		"Updated NFT by tests",
		"https://example.com/img.png",
		"https://example.com",
		map[string]string{"twitter": "https://x.com/2f"},
		map[string]string{"category": "Collectibles"},
		map[string]string{"tag1": "e2e-nft"},
		creator,
		"https://creator",
		time.Now().Add(30*24*time.Hour),
	)
	if err != nil {
		t.Fatalf("UpdateMetadata: %v", err)
	}

	unmarshalLogMetadata, err := utils.UnmarshalLog[log.Log](updateMetadata.Logs[0])
	if err != nil {
		t.Fatalf("UnmarshalLog (UpdateMetadata.Logs[0]): %v", err)
	}
	assert.Equal(t, tokenV1Domain.TOKEN_METADATA_UPDATED_LOG, unmarshalLogMetadata.LogType, "update metadata log type mismatch")

	metadataEvent, err := utils.UnmarshalEvent[tokenV1Domain.Token](unmarshalLogMetadata.Event)
	if err != nil {
		t.Fatalf("UnmarshalEvent (UpdateMetadata.Logs[0]): %v", err)
	}
	assert.Equal(t, tok.Address, metadataEvent.Address, "update metadata event token address mismatch")
	assert.Equal(t, newSymbol, metadataEvent.Symbol, "update metadata event token symbol mismatch")
	assert.Equal(t, "2Finance NFT New", metadataEvent.Name, "update metadata event token name mismatch")
	assert.Equal(t, decimals, metadataEvent.Decimals, "update metadata event token decimals mismatch")
	assert.Equal(t, "Updated NFT by tests", metadataEvent.Description, "update metadata event token description mismatch")
	assert.Equal(t, "https://example.com/img.png", metadataEvent.Image, "update metadata event token image mismatch")
	assert.Equal(t, "https://example.com", metadataEvent.Website, "update metadata event token website mismatch")
	assert.Equal(t, "https://x.com/2f", metadataEvent.TagsSocialMedia["twitter"], "update metadata event token tags social mismatch")
	assert.Equal(t, "Collectibles", metadataEvent.TagsCategory["category"], "update metadata event token tags category mismatch")
	assert.Equal(t, "e2e-nft", metadataEvent.Tags["tag1"], "update metadata event token tags mismatch")
	assert.Equal(t, creator, metadataEvent.Creator, "update metadata event token creator mismatch")
	assert.Equal(t, "https://creator", metadataEvent.CreatorWebsite, "update metadata event token creator website mismatch")

	getTokenOut14, err := c.GetToken(tok.Address, "", "")
	if err != nil {
		t.Fatalf("GetToken: %v", err)
	}

	var tokenState14 tokenV1Models.TokenStateModel
	err = utils.UnmarshalState[tokenV1Models.TokenStateModel](getTokenOut14.States[0].Object, &tokenState14)
	if err != nil {
		t.Fatalf("UnmarshalState (GetToken.States[0]): %v", err)
	}
	assert.Equal(t, newSymbol, tokenState14.Symbol, "token symbol mismatch after metadata update")
	assert.Equal(t, "2Finance NFT New", tokenState14.Name, "token name mismatch after metadata update")
	assert.Equal(t, "Updated NFT by tests", tokenState14.Description, "token description mismatch after metadata update")
	assert.Equal(t, "https://example.com/img.png", tokenState14.Image, "token image mismatch after metadata update")
	assert.Equal(t, "https://example.com", tokenState14.Website, "token website mismatch after metadata update")
	assert.Equal(t, "https://x.com/2f", tokenState14.TagsSocialMedia["twitter"], "token tags social mismatch after metadata update")
	assert.Equal(t, "Collectibles", tokenState14.TagsCategory["category"], "token tags category mismatch after metadata update")
	assert.Equal(t, "e2e-nft", tokenState14.Tags["tag1"], "token tags mismatch after metadata update")
	assert.Equal(t, creator, tokenState14.Creator, "token creator mismatch after metadata update")
	assert.Equal(t, "https://creator", tokenState14.CreatorWebsite, "token creator website mismatch after metadata update")

	// ------------------
	//      FREEZE
	// ------------------
	freezeWallet, err := c.FreezeWallet(tok.Address, owner.PublicKey)
	if err != nil {
		t.Fatalf("FreezeWallet: %v", err)
	}

	unmarshalLogFreeze, err := utils.UnmarshalLog[log.Log](freezeWallet.Logs[0])
	if err != nil {
		t.Fatalf("UnmarshalLog (FreezeWallet.Logs[0]): %v", err)
	}
	assert.Equal(t, tokenV1Domain.TOKEN_FREEZE_ACCOUNT_LOG, unmarshalLogFreeze.LogType, "freeze wallet log type mismatch")

	freezeEvent, err := utils.UnmarshalEvent[tokenV1Domain.Freeze](unmarshalLogFreeze.Event)
	if err != nil {
		t.Fatalf("UnmarshalEvent (FreezeWallet.Logs[0]): %v", err)
	}
	assert.Equal(t, tok.Address, freezeEvent.TokenAddress, "freeze wallet event token address mismatch")
	assert.Equal(t, true, freezeEvent.FrozenAccounts[owner.PublicKey], "freeze wallet event wallet address mismatch")

	getTokenOut15, err := c.GetToken(tok.Address, "", "")
	if err != nil {
		t.Fatalf("GetToken: %v", err)
	}

	var tokenState15 tokenV1Models.TokenStateModel
	err = utils.UnmarshalState[tokenV1Models.TokenStateModel](getTokenOut15.States[0].Object, &tokenState15)
	if err != nil {
		t.Fatalf("UnmarshalState (GetToken.States[0]): %v", err)
	}
	assert.Equal(t, true, tokenState15.FrozenAccounts[owner.PublicKey], "token frozen accounts mismatch after freeze")

	// ------------------
	//      UNFREEZE
	// ------------------
	unfreezeWallet, err := c.UnfreezeWallet(tok.Address, owner.PublicKey)
	if err != nil {
		t.Fatalf("UnfreezeWallet: %v", err)
	}

	unmarshalLogUnfreeze, err := utils.UnmarshalLog[log.Log](unfreezeWallet.Logs[0])
	if err != nil {
		t.Fatalf("UnmarshalLog (UnfreezeWallet.Logs[0]): %v", err)
	}
	assert.Equal(t, tokenV1Domain.TOKEN_UNFREEZE_ACCOUNT_LOG, unmarshalLogUnfreeze.LogType, "unfreeze wallet log type mismatch")

	unfreezeEvent, err := utils.UnmarshalEvent[tokenV1Domain.Freeze](unmarshalLogUnfreeze.Event)
	if err != nil {
		t.Fatalf("UnmarshalEvent (UnfreezeWallet.Logs[0]): %v", err)
	}
	assert.Equal(t, tok.Address, unfreezeEvent.TokenAddress, "unfreeze wallet event token address mismatch")
	assert.Equal(t, false, unfreezeEvent.FrozenAccounts[owner.PublicKey], "unfreeze wallet event wallet address mismatch")

	getTokenOut16, err := c.GetToken(tok.Address, "", "")
	if err != nil {
		t.Fatalf("GetToken: %v", err)
	}

	var tokenState16 tokenV1Models.TokenStateModel
	err = utils.UnmarshalState[tokenV1Models.TokenStateModel](getTokenOut16.States[0].Object, &tokenState16)
	if err != nil {
		t.Fatalf("UnmarshalState (GetToken.States[0]): %v", err)
	}
	_, exists3 := tokenState16.FrozenAccounts[owner.PublicKey]
	assert.False(t, exists3, "token frozen accounts mismatch after unfreeze")

	// ------------------
	//     UPDATE GLB
	// ------------------
	newGLB := "https://example.com/assets/token.glb"
	updateGlbFile, err := c.UpdateGlbFile(tok.Address, newGLB)
	if err != nil {
		t.Fatalf("UpdateGlbFile: %v", err)
	}

	unmarshalLogUpdateGlbFile, err := utils.UnmarshalLog[log.Log](updateGlbFile.Logs[0])
	if err != nil {
		t.Fatalf("UnmarshalLog (UpdateGlbFile.Logs[0]): %v", err)
	}
	assert.Equal(t, tokenV1Domain.TOKEN_UPDATE_GLB_FILE_LOG, unmarshalLogUpdateGlbFile.LogType, "update glb file log type mismatch")

	updateGlbFileEvent, err := utils.UnmarshalEvent[tokenV1Domain.Token](unmarshalLogUpdateGlbFile.Event)
	if err != nil {
		t.Fatalf("UnmarshalEvent (UpdateGlbFile.Logs[0]): %v", err)
	}
	assert.Equal(t, tok.Address, updateGlbFileEvent.Address, "update glb file token address mismatch")
	assert.Equal(t, newGLB, updateGlbFileEvent.AssetGLBUri, "update glb file asset uri mismatch")

	getTokenOut17, err := c.GetToken(tok.Address, "", "")
	if err != nil {
		t.Fatalf("GetToken: %v", err)
	}

	var tokenState17 tokenV1Models.TokenStateModel
	err = utils.UnmarshalState[tokenV1Models.TokenStateModel](getTokenOut17.States[0].Object, &tokenState17)
	if err != nil {
		t.Fatalf("UnmarshalState (GetToken.States[0]): %v", err)
	}
	assert.Equal(t, newGLB, tokenState17.AssetGLBUri, "token asset glb uri mismatch after update")

	// ------------------
	// REVOKE AUTHORITIES
	// ------------------
	revokeFreezeAuthority, err := c.RevokeFreezeAuthority(tok.Address, true)
	if err != nil {
		t.Fatalf("RevokeFreezeAuthority: %v", err)
	}

	unmarshalLogRevokeFreezeAuthority, err := utils.UnmarshalLog[log.Log](revokeFreezeAuthority.Logs[0])
	if err != nil {
		t.Fatalf("UnmarshalLog (RevokeFreezeAuthority.Logs[0]): %v", err)
	}
	assert.Equal(t, tokenV1Domain.TOKEN_FREEZE_AUTHORITY_REVOKED_LOG, unmarshalLogRevokeFreezeAuthority.LogType, "revoke freeze authority log type mismatch")

	revokeFreezeAuthorityEvent, err := utils.UnmarshalEvent[tokenV1Domain.Token](unmarshalLogRevokeFreezeAuthority.Event)
	if err != nil {
		t.Fatalf("UnmarshalEvent (RevokeFreezeAuthority.Logs[0]): %v", err)
	}
	assert.Equal(t, tok.Address, revokeFreezeAuthorityEvent.Address, "revoke freeze authority token address mismatch")
	assert.Equal(t, true, revokeFreezeAuthorityEvent.FreezeAuthorityRevoked, "revoke freeze authority state mismatch")

	getTokenOut18, err := c.GetToken(tok.Address, "", "")
	if err != nil {
		t.Fatalf("GetToken: %v", err)
	}

	var tokenState18 tokenV1Models.TokenStateModel
	err = utils.UnmarshalState[tokenV1Models.TokenStateModel](getTokenOut18.States[0].Object, &tokenState18)
	if err != nil {
		t.Fatalf("UnmarshalState (GetToken.States[0]): %v", err)
	}
	assert.Equal(t, true, tokenState18.FreezeAuthorityRevoked, "token revoked freeze authority mismatch after revoke")

	revokeMint, err := c.RevokeMintAuthority(tok.Address, true)
	if err != nil {
		t.Fatalf("RevokeMintAuthority: %v", err)
	}

	unmarshalLogRevokeMint, err := utils.UnmarshalLog[log.Log](revokeMint.Logs[0])
	if err != nil {
		t.Fatalf("UnmarshalLog (RevokeMintAuthority.Logs[0]): %v", err)
	}
	assert.Equal(t, tokenV1Domain.TOKEN_MINT_AUTHORITY_REVOKED_LOG, unmarshalLogRevokeMint.LogType, "revoke mint authority log type mismatch")

	revokeMintEvent, err := utils.UnmarshalEvent[tokenV1Domain.Token](unmarshalLogRevokeMint.Event)
	if err != nil {
		t.Fatalf("UnmarshalEvent (RevokeMintAuthority.Logs[0]): %v", err)
	}
	assert.Equal(t, tok.Address, revokeMintEvent.Address, "revoke mint authority event token address mismatch")
	assert.Equal(t, true, revokeMintEvent.MintAuthorityRevoked, "revoke mint authority event mint authority revoked mismatch")

	getTokenOut19, err := c.GetToken(tok.Address, "", "")
	if err != nil {
		t.Fatalf("GetToken: %v", err)
	}

	var tokenState19 tokenV1Models.TokenStateModel
	err = utils.UnmarshalState[tokenV1Models.TokenStateModel](getTokenOut19.States[0].Object, &tokenState19)
	if err != nil {
		t.Fatalf("UnmarshalState (GetToken.States[0]): %v", err)
	}
	assert.Equal(t, true, tokenState19.MintAuthorityRevoked, "token mint authority revoked mismatch after revoke")

	revokeUpdate, err := c.RevokeUpdateAuthority(tok.Address, true)
	if err != nil {
		t.Fatalf("RevokeUpdateAuthority: %v", err)
	}

	unmarshalLogRevokeUpdate, err := utils.UnmarshalLog[log.Log](revokeUpdate.Logs[0])
	if err != nil {
		t.Fatalf("UnmarshalLog (RevokeUpdateAuthority.Logs[0]): %v", err)
	}
	assert.Equal(t, tokenV1Domain.TOKEN_UPDATE_AUTHORITY_REVOKED_LOG, unmarshalLogRevokeUpdate.LogType, "revoke update authority log type mismatch")

	revokeUpdateEvent, err := utils.UnmarshalEvent[tokenV1Domain.Token](unmarshalLogRevokeUpdate.Event)
	if err != nil {
		t.Fatalf("UnmarshalEvent (RevokeUpdateAuthority.Logs[0]): %v", err)
	}
	assert.Equal(t, tok.Address, revokeUpdateEvent.Address, "revoke update authority event token address mismatch")
	assert.Equal(t, true, revokeUpdateEvent.UpdateAuthorityRevoked, "revoke update authority event update authority revoked mismatch")

	getTokenOut20, err := c.GetToken(tok.Address, "", "")
	if err != nil {
		t.Fatalf("GetToken: %v", err)
	}

	var tokenState20 tokenV1Models.TokenStateModel
	err = utils.UnmarshalState[tokenV1Models.TokenStateModel](getTokenOut20.States[0].Object, &tokenState20)
	if err != nil {
		t.Fatalf("UnmarshalState (GetToken.States[0]): %v", err)
	}
	assert.Equal(t, true, tokenState20.UpdateAuthorityRevoked, "token update authority revoked mismatch after revoke")

	// ------------------
	// GETTERS | LISTINGS
	// ------------------
	// if _, err := c.GetTokenBalance(tok.Address, owner.PublicKey); err != nil {
	// 	t.Fatalf("GetTokenBalance(owner): %v", err)
	// }
	// if _, err := c.GetTokenBalance(tok.Address, receiver.PublicKey); err != nil {
	// 	t.Fatalf("GetTokenBalance(receiver): %v", err)
	// }
	// if _, err := c.ListTokenBalances(tok.Address, "", 1, 10, true); err != nil {
	// 	t.Fatalf("ListTokenBalances: %v", err)
	// }
	// if _, err := c.ListTokens("", "", "", 1, 10, true); err != nil {
	// 	t.Fatalf("ListTokens: %v", err)
	// }
}
