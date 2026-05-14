package e2e_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/2finance/2finance-network/blockchain/contract/tokenV1"
	tokenV1Domain "gitlab.com/2finance/2finance-network/blockchain/contract/tokenV1/domain"
	tokenV1Models "gitlab.com/2finance/2finance-network/blockchain/contract/tokenV1/models"
	"gitlab.com/2finance/2finance-network/blockchain/log"
	"gitlab.com/2finance/2finance-network/blockchain/utils"
)

func TestTokenFlowNonFungible(t *testing.T) {
	// ------------------
	//   LOCAL WALLETS
	// ------------------
	ownerSigner := setupSignerWallet(t)
	receiverSigner := setupSignerWallet(t)

	c := setupClient(t, ownerSigner.Wallet)

	// ------------------
	//   ON-CHAIN WALLETS
	// ------------------
	useWallet(t, c, ownerSigner.Wallet)
	owner := createWallet(t, c, ownerSigner.PublicKey)

	useWallet(t, c, receiverSigner.Wallet)
	receiver := createWallet(t, c, receiverSigner.PublicKey)

	tmpWM := setupWalletManager(t)

	// ------------------
	//   DEPLOY TOKEN
	// ------------------
	useWallet(t, c, ownerSigner.Wallet)

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
	useWallet(t, c, ownerSigner.Wallet)

	address := contractLog.ContractAddress
	decimals := 0
	tokenType := tokenV1Domain.NON_FUNGIBLE
	assetType := tokenV1Domain.TOKEN_ASSET_TYPE
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
	allowUsers := map[string]bool{}
	blockedUsers := map[string]bool{}
	frozenAccounts := map[string]bool{}
	feeTiers := []map[string]interface{}{}

	feeAddress, _ := genKey(t, tmpWM)
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
		allowUsers,
		blockedUsers,
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
		assetType,
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
	assert.Empty(t, tok.AllowedUsers, "token allowed users mismatch")
	assert.Empty(t, tok.BlockedUsers, "token blocked users mismatch")
	assert.Empty(t, tok.FrozenAccounts, "token frozen accounts mismatch")
	assert.Equal(t, feeAddress, tok.FeeAddress, "token fee address mismatch")
	assert.Equal(t, freezeAuthorityRevoked, tok.FreezeAuthorityRevoked, "token freeze authority revoked mismatch")
	assert.Equal(t, mintAuthorityRevoked, tok.MintAuthorityRevoked, "token mint authority revoked mismatch")
	assert.Equal(t, updateAuthorityRevoked, tok.UpdateAuthorityRevoked, "token update authority revoked mismatch")
	assert.Equal(t, paused, tok.Paused, "token paused mismatch")
	assert.Equal(t, assetGLBUri, tok.AssetGLBUri, "token asset GLB URI mismatch")
	assert.Equal(t, tokenType, tok.TokenType, "token type mismatch")
	assert.Equal(t, transferable, tok.Transferable, "token transferable mismatch")
	assert.Equal(t, assetType, tok.AssetType, "token asset type mismatch")

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
	assert.Empty(t, tokenState.AllowedUsers, "token allowed users mismatch")
	assert.Empty(t, tokenState.BlockedUsers, "token blocked users mismatch")
	assert.Empty(t, tokenState.FrozenAccounts, "token frozen accounts mismatch")
	assert.Equal(t, feeAddress, tokenState.FeeAddress, "token fee address mismatch")
	assert.Equal(t, freezeAuthorityRevoked, tokenState.FreezeAuthorityRevoked, "token freeze authority revoked mismatch")
	assert.Equal(t, mintAuthorityRevoked, tokenState.MintAuthorityRevoked, "token mint authority revoked mismatch")
	assert.Equal(t, updateAuthorityRevoked, tokenState.UpdateAuthorityRevoked, "token update authority revoked mismatch")
	assert.Equal(t, paused, tokenState.Paused, "token paused mismatch")
	assert.Equal(t, assetGLBUri, tokenState.AssetGLBUri, "token asset GLB URI mismatch")
	assert.Equal(t, tokenType, tokenState.TokenType, "token type mismatch")
	assert.Equal(t, transferable, tokenState.Transferable, "token transferable mismatch")
	assert.Equal(t, assetType, tokenState.AssetType, "token asset type mismatch")

	// ------------------
	//        MINT
	// ------------------
	useWallet(t, c, ownerSigner.Wallet)

	mintAmount := "10"
	mintToken, err := c.MintToken(tok.Address, owner.PublicKey, mintAmount)
	if err != nil {
		t.Fatalf("MintToken: %v", err)
	}

	unmarshalLogMint, err := utils.UnmarshalLog[log.Log](mintToken.Logs[0])
	if err != nil {
		t.Fatalf("UnmarshalLog (MintToken.Logs[0]): %v", err)
	}
	assert.Equal(t, tokenV1Domain.TOKEN_MINTED_NFT_LOG, unmarshalLogMint.LogType, "mint log type mismatch")

	mint, err := utils.UnmarshalEvent[tokenV1Domain.MintNFT](unmarshalLogMint.Event)
	if err != nil {
		t.Fatalf("UnmarshalEvent (MintToken.Logs[0]): %v", err)
	}

	assert.Equal(t, tok.Address, mint.TokenAddress, "mint token address mismatch")
	assert.Equal(t, owner.PublicKey, mint.MintTo, "mint to address mismatch")
	assert.Equal(t, mintAmount, mint.Amount, "mint amount mismatch")
	assert.Equal(t, tokenType, mint.TokenType, "mint token type mismatch")
	assert.Len(t, mint.TokenUUIDList, 10, "mint token UUID list length mismatch")

	mintedUUID := mint.TokenUUIDList[0]
	assert.NotEmpty(t, mintedUUID, "minted UUID should not be empty")

	unmarshalLogSupply, err := utils.UnmarshalLog[log.Log](mintToken.Logs[1])
	if err != nil {
		t.Fatalf("UnmarshalLog (MintToken.Logs[1]): %v", err)
	}
	assert.Equal(t, tokenV1Domain.TOKEN_TOTAL_SUPPLY_INCREASED_LOG, unmarshalLogSupply.LogType, "supply log type mismatch")

	supply, err := utils.UnmarshalEvent[tokenV1Domain.Supply](unmarshalLogSupply.Event)
	if err != nil {
		t.Fatalf("UnmarshalEvent (MintToken.Logs[1]): %v", err)
	}
	assert.Equal(t, tok.Address, supply.TokenAddress, "supply token address mismatch")
	assert.Equal(t, mintAmount, supply.Amount, "total supply mismatch after mint")

	unmarshalLogBalance, err := utils.UnmarshalLog[log.Log](mintToken.Logs[2])
	if err != nil {
		t.Fatalf("UnmarshalLog (MintToken.Logs[2]): %v", err)
	}
	assert.Equal(t, tokenV1Domain.TOKEN_BALANCE_INCREASED_NFT_LOG, unmarshalLogBalance.LogType, "balance log type mismatch")

	balance, err := utils.UnmarshalEvent[tokenV1Domain.BalanceNFT](unmarshalLogBalance.Event)
	if err != nil {
		t.Fatalf("UnmarshalEvent (MintToken.Logs[2]): %v", err)
	}

	assert.Equal(t, tok.Address, balance.TokenAddress, "balance token address mismatch")
	assert.Equal(t, owner.PublicKey, balance.OwnerAddress, "balance wallet address mismatch")
	assert.Equal(t, tokenType, balance.TokenType, "balance token type mismatch")
	require.Len(t, balance.TokenUUIDList, 10, "balance token UUID list mismatch after mint")
	assert.Equal(t, mintedUUID, balance.TokenUUIDList[0], "balance token UUID mismatch after mint")

	getTokenOut2, err := c.GetToken(tok.Address, "", "")
	if err != nil {
		t.Fatalf("GetToken: %v", err)
	}

	var tokenState2 tokenV1Models.TokenStateModel
	err = utils.UnmarshalState[tokenV1Models.TokenStateModel](getTokenOut2.States[0].Object, &tokenState2)
	if err != nil {
		t.Fatalf("UnmarshalState (GetToken.States[0]): %v", err)
	}
	assert.Equal(t, "11", tokenState2.TotalSupply, "token total supply mismatch after mint")

	// ------------------
	//   TRANSFER TOKEN
	// ------------------
	useWallet(t, c, ownerSigner.Wallet)

	_, err = c.AddAllowedUsers(tok.Address, map[string]bool{
		receiver.PublicKey: true,
	})
	assert.NoError(t, err, "AddAllowedUsers failed")

	mintedUUID = mint.TokenUUIDList[0]

	transferToken, err := c.TransferToken(tok.Address, receiver.PublicKey, "", []string{mintedUUID})
	if err != nil {
		t.Fatalf("TransferToken: %v", err)
	}

	unmarshalLogTransfer, err := utils.UnmarshalLog[log.Log](transferToken.Logs[0])
	if err != nil {
		t.Fatalf("UnmarshalLog (TransferToken.Logs[0]): %v", err)
	}
	assert.Equal(t, tokenV1Domain.TOKEN_TRANSFERRED_NFT_LOG, unmarshalLogTransfer.LogType, "transfer log type mismatch")

	transferNFT, err := utils.UnmarshalEvent[tokenV1Domain.TransferNFT](unmarshalLogTransfer.Event)
	if err != nil {
		t.Fatalf("UnmarshalEvent (TransferToken.Logs[0]): %v", err)
	}

	assert.Equal(t, tok.Address, transferNFT.TokenAddress, "transfer token address mismatch")
	assert.Equal(t, owner.PublicKey, transferNFT.FromAddress, "transfer from address mismatch")
	assert.Equal(t, receiver.PublicKey, transferNFT.ToAddress, "transfer to address mismatch")
	assert.Equal(t, tokenType, transferNFT.TokenType, "transfer token type mismatch")
	assert.Equal(t, mintedUUID, transferNFT.TokenUUIDList[0], "transfer UUID mismatch")

	unmarshalLogBalanceFrom, err := utils.UnmarshalLog[log.Log](transferToken.Logs[1])
	if err != nil {
		t.Fatalf("UnmarshalLog (TransferToken.Logs[1]): %v", err)
	}
	assert.Equal(t, tokenV1Domain.TOKEN_BALANCE_DECREASED_NFT_LOG, unmarshalLogBalanceFrom.LogType, "sender balance decreased log type mismatch")

	balanceFrom, err := utils.UnmarshalEvent[tokenV1Domain.BalanceNFT](unmarshalLogBalanceFrom.Event)
	if err != nil {
		t.Fatalf("UnmarshalEvent (TransferToken.Logs[1]): %v", err)
	}
	assert.Equal(t, tok.Address, balanceFrom.TokenAddress, "sender balance token address mismatch")
	assert.Equal(t, owner.PublicKey, balanceFrom.OwnerAddress, "sender balance owner mismatch")
	assert.Equal(t, tokenType, balanceFrom.TokenType, "sender balance token type mismatch")
	assert.Len(t, balanceFrom.TokenUUIDList, 1, "sender balance token uuid list length mismatch after transfer")
	assert.Equal(t, mintedUUID, balanceFrom.TokenUUIDList[0], "sender balance token uuid mismatch after transfer")

	unmarshalLogBalanceTo, err := utils.UnmarshalLog[log.Log](transferToken.Logs[2])
	if err != nil {
		t.Fatalf("UnmarshalLog (TransferToken.Logs[2]): %v", err)
	}
	assert.Equal(t, tokenV1Domain.TOKEN_BALANCE_INCREASED_NFT_LOG, unmarshalLogBalanceTo.LogType, "receiver balance increased log type mismatch")

	balanceTo, err := utils.UnmarshalEvent[tokenV1Domain.BalanceNFT](unmarshalLogBalanceTo.Event)
	if err != nil {
		t.Fatalf("UnmarshalEvent (TransferToken.Logs[2]): %v", err)
	}
	assert.Equal(t, tok.Address, balanceTo.TokenAddress, "receiver balance token address mismatch")
	assert.Equal(t, receiver.PublicKey, balanceTo.OwnerAddress, "receiver balance owner mismatch")
	assert.Equal(t, tokenType, balanceTo.TokenType, "receiver balance token type mismatch")
	require.Len(t, balanceTo.TokenUUIDList, 1, "receiver balance token uuid list mismatch")
	assert.Equal(t, mintedUUID, balanceTo.TokenUUIDList[0], "receiver balance token uuid mismatch")

	getBalanceReceiver, err := c.GetTokenBalanceNFT(tok.Address, receiver.PublicKey, mintedUUID)
	if err != nil {
		t.Fatalf("GetTokenBalance (receiver): %v", err)
	}

	var balanceReceiver tokenV1Models.BalanceStateModel
	err = utils.UnmarshalState[tokenV1Models.BalanceStateModel](getBalanceReceiver.States[0].Object, &balanceReceiver)
	if err != nil {
		t.Fatalf("UnmarshalState (GetTokenBalance receiver): %v", err)
	}

	assert.Equal(t, tok.Address, balanceReceiver.TokenAddress, "receiver balance token address mismatch")
	assert.Equal(t, receiver.PublicKey, balanceReceiver.OwnerAddress, "receiver balance owner mismatch")
	assert.Equal(t, mintedUUID, balanceReceiver.TokenUUID, "receiver balance token UUID mismatch after transfer")
	assert.Equal(t, "1", balanceReceiver.Amount, "receiver balance amount mismatch after transfer")
	assert.NotNil(t, balanceReceiver.CreatedAt, "receiver balance created at should not be nil")
	assert.NotNil(t, balanceReceiver.UpdatedAt, "receiver balance updated at should not be nil")
	assert.Equal(t, tokenV1Domain.NON_FUNGIBLE, balanceReceiver.TokenType, "receiver balance token type mismatch after transfer")

	// ------------------
	//        BURN
	// ------------------
	useWallet(t, c, receiverSigner.Wallet)

	listOfUUIDs := []string{mintedUUID}
	burnToken, err := c.BurnToken(tok.Address, "", listOfUUIDs)
	if err != nil {
		t.Fatalf("BurnToken: %v", err)
	}

	unmarshalLogBurn, err := utils.UnmarshalLog[log.Log](burnToken.Logs[0])
	if err != nil {
		t.Fatalf("UnmarshalLog (BurnToken.Logs[0]): %v", err)
	}
	assert.Equal(t, tokenV1Domain.TOKEN_BURNED_NFT_LOG, unmarshalLogBurn.LogType, "burn log type mismatch")

	burn, err := utils.UnmarshalEvent[tokenV1Domain.BurnNFT](unmarshalLogBurn.Event)
	if err != nil {
		t.Fatalf("UnmarshalEvent (BurnToken.Logs[0]): %v", err)
	}

	assert.Equal(t, tok.Address, burn.TokenAddress, "burn token address mismatch")
	assert.Equal(t, receiver.PublicKey, burn.BurnFrom, "burn from address mismatch")
	assert.Equal(t, tokenType, burn.TokenType, "burn token type mismatch")
	assert.Equal(t, mintedUUID, burn.TokensUUID[0], "burn token UUID mismatch")

	unmarshalLogSupply2, err := utils.UnmarshalLog[log.Log](burnToken.Logs[1])
	if err != nil {
		t.Fatalf("UnmarshalLog (BurnToken.Logs[1]): %v", err)
	}
	assert.Equal(t, tokenV1Domain.TOKEN_TOTAL_SUPPLY_DECREASED_LOG, unmarshalLogSupply2.LogType, "supply log type mismatch")

	supply2, err := utils.UnmarshalEvent[tokenV1Domain.Supply](unmarshalLogSupply2.Event)
	if err != nil {
		t.Fatalf("UnmarshalEvent (BurnToken.Logs[1]): %v", err)
	}
	assert.Equal(t, tok.Address, supply2.TokenAddress, "supply token address mismatch")
	assert.Equal(t, "1", supply2.Amount, "total supply mismatch after burn")

	unmarshalLogBalance2, err := utils.UnmarshalLog[log.Log](burnToken.Logs[2])
	if err != nil {
		t.Fatalf("UnmarshalLog (BurnToken.Logs[2]): %v", err)
	}
	assert.Equal(t, tokenV1Domain.TOKEN_BALANCE_DECREASED_NFT_LOG, unmarshalLogBalance2.LogType, "balance log type mismatch")

	balance2, err := utils.UnmarshalEvent[tokenV1Domain.BalanceNFT](unmarshalLogBalance2.Event)
	if err != nil {
		t.Fatalf("UnmarshalEvent (BurnToken.Logs[2]): %v", err)
	}

	assert.Equal(t, tok.Address, balance2.TokenAddress, "balance token address mismatch")
	assert.Equal(t, receiver.PublicKey, balance2.OwnerAddress, "balance wallet address mismatch")
	assert.Equal(t, tokenType, balance2.TokenType, "balance token type mismatch")
	assert.Len(t, balance2.TokenUUIDList, 1, "balance token UUID list length mismatch after burn")

	getTokenOut3, err := c.GetToken(tok.Address, "", "")
	if err != nil {
		t.Fatalf("GetToken: %v", err)
	}

	var tokenState3 tokenV1Models.TokenStateModel
	err = utils.UnmarshalState[tokenV1Models.TokenStateModel](getTokenOut3.States[0].Object, &tokenState3)
	if err != nil {
		t.Fatalf("UnmarshalState (GetToken.States[0]): %v", err)
	}

	sumTotalSupply, err := utils.AddBigIntStrings(totalSupply, mintAmount)
	if err != nil {
		t.Fatalf("AddBigIntStrings: %v", err)
	}

	totalSupplyAfterBurn, err := utils.SubBigIntStrings(sumTotalSupply, "1")
	if err != nil {
		t.Fatalf("SubBigIntStrings: %v", err)
	}

	assert.Equal(t, totalSupplyAfterBurn, tokenState3.TotalSupply, "token total supply mismatch after burn")

	tokenBalanceNFT, err := c.GetTokenBalanceNFT(tok.Address, receiver.PublicKey, mintedUUID)
	if err != nil {
		t.Fatalf("GetTokenBalanceNFT: %v", err)
	}

	var balanceStateAfterBurn tokenV1Models.BalanceStateModel
	err = utils.UnmarshalState[tokenV1Models.BalanceStateModel](tokenBalanceNFT.States[0].Object, &balanceStateAfterBurn)
	if err != nil {
		t.Fatalf("UnmarshalState (GetTokenBalanceNFT after burn): %v", err)
	}

	assert.Equal(t, "1", balanceStateAfterBurn.Amount, "receiver balance amount mismatch after burn")
	assert.Equal(t, tokenV1Domain.NON_FUNGIBLE, balanceStateAfterBurn.TokenType, "receiver balance token type mismatch after burn")
	assert.Equal(t, mintedUUID, balanceStateAfterBurn.TokenUUID, "receiver balance token UUID mismatch after burn")
	assert.Equal(t, tok.Address, balanceStateAfterBurn.TokenAddress, "receiver balance token address mismatch after burn")
	assert.Equal(t, receiver.PublicKey, balanceStateAfterBurn.OwnerAddress, "receiver balance owner mismatch after burn")
	assert.Equal(t, true, balanceStateAfterBurn.Burned, "receiver balance status mismatch after burn")
	assert.NotNil(t, balanceStateAfterBurn.CreatedAt, "receiver balance created at should not be nil after burn")
	assert.NotNil(t, balanceStateAfterBurn.UpdatedAt, "receiver balance updated at should not be nil after burn")
	assert.NotNil(t, balanceStateAfterBurn.BurnedAt, "receiver balance burned at should not be nil after burn")

	// ------------------
	//    ALLOW USERS
	// ------------------
	useWallet(t, c, ownerSigner.Wallet)

	anotherUserPub, _ := genKey(t, tmpWM)

	allowedUsers, err := c.AddAllowedUsers(tok.Address, map[string]bool{
		anotherUserPub: true,
	})
	if err != nil {
		t.Fatalf("AddAllowedUsers: %v", err)
	}

	unmarshalLogAllow, err := utils.UnmarshalLog[log.Log](allowedUsers.Logs[0])
	if err != nil {
		t.Fatalf("UnmarshalLog (AddAllowedUsers.Logs[0]): %v", err)
	}
	assert.Equal(t, tokenV1Domain.TOKEN_ALLOWED_USERS_ADDED_LOG, unmarshalLogAllow.LogType, "allow users log type mismatch")

	allow, err := utils.UnmarshalEvent[tokenV1Domain.AccessPolicy](unmarshalLogAllow.Event)
	if err != nil {
		t.Fatalf("UnmarshalEvent (AddAllowedUsers.Logs[0]): %v", err)
	}
	assert.Equal(t, tok.Address, allow.Address, "access policy token address mismatch")
	assert.Equal(t, true, allow.AllowedUsers[anotherUserPub], "access policy users mismatch after allow")

	getTokenOut5, err := c.GetToken(tok.Address, "", "")
	if err != nil {
		t.Fatalf("GetToken: %v", err)
	}

	var tokenState5 tokenV1Models.TokenStateModel
	err = utils.UnmarshalState[tokenV1Models.TokenStateModel](getTokenOut5.States[0].Object, &tokenState5)
	if err != nil {
		t.Fatalf("UnmarshalState (GetToken.States[0]): %v", err)
	}
	assert.Equal(t, true, tokenState5.AllowedUsers[anotherUserPub], "token allowed users mismatch after allow")

	removeAllowedUsers, err := c.RemoveAllowedUsers(tok.Address, map[string]bool{
		anotherUserPub: true,
	})
	if err != nil {
		t.Fatalf("RemoveAllowedUsers: %v", err)
	}

	unmarshalLogRemoveAllow, err := utils.UnmarshalLog[log.Log](removeAllowedUsers.Logs[0])
	if err != nil {
		t.Fatalf("UnmarshalLog (RemoveAllowedUsers.Logs[0]): %v", err)
	}
	assert.Equal(t, tokenV1Domain.TOKEN_ALLOWED_USERS_REMOVED_LOG, unmarshalLogRemoveAllow.LogType, "remove allowed users log type mismatch")

	removeAllowed, err := utils.UnmarshalEvent[tokenV1Domain.AccessPolicy](unmarshalLogRemoveAllow.Event)
	if err != nil {
		t.Fatalf("UnmarshalEvent (RemoveAllowedUsers.Logs[0]): %v", err)
	}
	assert.Equal(t, tok.Address, removeAllowed.Address, "access policy token address mismatch")
	assert.True(t, removeAllowed.AllowedUsers[anotherUserPub], "removed user payload should be true in remove event")

	getTokenOut6, err := c.GetToken(tok.Address, "", "")
	if err != nil {
		t.Fatalf("GetToken: %v", err)
	}

	var tokenState6 tokenV1Models.TokenStateModel
	err = utils.UnmarshalState[tokenV1Models.TokenStateModel](getTokenOut6.States[0].Object, &tokenState6)
	if err != nil {
		t.Fatalf("UnmarshalState (GetToken.States[0]): %v", err)
	}

	_, exists := tokenState6.AllowedUsers[anotherUserPub]
	assert.False(t, exists, "user should have been removed from token allowed users")

	// ------------------
	//    BLOCKED USERS
	// ------------------
	useWallet(t, c, ownerSigner.Wallet)

	blockAddUsers, err := c.AddBlockedUsers(tok.Address, map[string]bool{
		receiver.PublicKey: true,
	})
	if err != nil {
		t.Fatalf("AddBlockedUsers: %v", err)
	}

	unmarshalLogBlocked, err := utils.UnmarshalLog[log.Log](blockAddUsers.Logs[0])
	if err != nil {
		t.Fatalf("UnmarshalLog (AddBlockedUsers.Logs[0]): %v", err)
	}
	assert.Equal(t, tokenV1Domain.TOKEN_BLOCKED_USERS_ADDED_LOG, unmarshalLogBlocked.LogType, "add blocked users log type mismatch")

	blocked, err := utils.UnmarshalEvent[tokenV1Domain.AccessPolicy](unmarshalLogBlocked.Event)
	if err != nil {
		t.Fatalf("UnmarshalEvent (AddBlockedUsers.Logs[0]): %v", err)
	}
	assert.Equal(t, tok.Address, blocked.Address, "access policy token address mismatch")
	assert.Equal(t, true, blocked.BlockedUsers[receiver.PublicKey], "access policy users mismatch after add blocked")

	getTokenOut8, err := c.GetToken(tok.Address, "", "")
	if err != nil {
		t.Fatalf("GetToken: %v", err)
	}

	var tokenState8 tokenV1Models.TokenStateModel
	err = utils.UnmarshalState[tokenV1Models.TokenStateModel](getTokenOut8.States[0].Object, &tokenState8)
	if err != nil {
		t.Fatalf("UnmarshalState (GetToken.States[0]): %v", err)
	}
	assert.Equal(t, true, tokenState8.BlockedUsers[receiver.PublicKey], "token blocked users mismatch after add blocked")

	removeBlockedUsers, err := c.RemoveBlockedUsers(tok.Address, map[string]bool{
		receiver.PublicKey: true,
	})
	if err != nil {
		t.Fatalf("RemoveBlockedUsers: %v", err)
	}

	unmarshalLogRemoveBlocked, err := utils.UnmarshalLog[log.Log](removeBlockedUsers.Logs[0])
	if err != nil {
		t.Fatalf("UnmarshalLog (RemoveBlockedUsers.Logs[0]): %v", err)
	}
	assert.Equal(t, tokenV1Domain.TOKEN_BLOCKED_USERS_REMOVED_LOG, unmarshalLogRemoveBlocked.LogType, "remove blocked users log type mismatch")

	removeBlocked, err := utils.UnmarshalEvent[tokenV1Domain.AccessPolicy](unmarshalLogRemoveBlocked.Event)
	if err != nil {
		t.Fatalf("UnmarshalEvent (RemoveBlockedUsers.Logs[0]): %v", err)
	}
	assert.Equal(t, tok.Address, removeBlocked.Address, "access policy token address mismatch")
	assert.Contains(t, removeBlocked.BlockedUsers, receiver.PublicKey, "removed user should be present in remove event payload")
	assert.True(t, removeBlocked.BlockedUsers[receiver.PublicKey], "removed user payload should be true in remove event")

	getTokenOut9, err := c.GetToken(tok.Address, "", "")
	if err != nil {
		t.Fatalf("GetToken: %v", err)
	}

	var tokenState9 tokenV1Models.TokenStateModel
	err = utils.UnmarshalState[tokenV1Models.TokenStateModel](getTokenOut9.States[0].Object, &tokenState9)
	if err != nil {
		t.Fatalf("UnmarshalState (GetToken.States[0]): %v", err)
	}

	_, exists2 := tokenState9.BlockedUsers[receiver.PublicKey]
	assert.False(t, exists2, "user should have been removed from token blocked users")

	// ------------------
	//       PAUSE
	// ------------------
	useWallet(t, c, ownerSigner.Wallet)

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
	useWallet(t, c, ownerSigner.Wallet)

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
	//    FEE ADDRESS
	// ------------------
	useWallet(t, c, ownerSigner.Wallet)

	updateFeeAddress, err := c.UpdateFeeAddress(tok.Address, feeAddress)
	if err != nil {
		t.Fatalf("UpdateFeeAddress: %v", err)
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
	useWallet(t, c, ownerSigner.Wallet)

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
	useWallet(t, c, ownerSigner.Wallet)

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
	useWallet(t, c, ownerSigner.Wallet)

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
	useWallet(t, c, ownerSigner.Wallet)

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
	useWallet(t, c, ownerSigner.Wallet)

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
	uuid := mint.TokenUUIDList[1]

	balanceNFT, err := c.GetTokenBalanceNFT(tok.Address, owner.PublicKey, uuid)
	if err != nil {
		t.Fatalf("GetTokenBalance(owner): %v", err)
	}

	var balanceNFTState tokenV1Models.BalanceStateModel
	err = utils.UnmarshalState[tokenV1Models.BalanceStateModel](balanceNFT.States[0].Object, &balanceNFTState)
	if err != nil {
		t.Fatalf("UnmarshalState (GetTokenBalanceNFT.States[0]): %v", err)
	}

	assert.Equal(t, tok.Address, balanceNFTState.TokenAddress, "balance token address mismatch")
	assert.Equal(t, owner.PublicKey, balanceNFTState.OwnerAddress, "balance wallet address mismatch")
	assert.Equal(t, tokenType, balanceNFTState.TokenType, "balance token type mismatch")
	assert.Equal(t, uuid, balanceNFTState.TokenUUID, "balance token UUID mismatch")
}