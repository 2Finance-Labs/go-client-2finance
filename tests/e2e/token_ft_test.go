package e2e_test

import (
	"testing"
	"time"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	userPubAccess, _ := genKey(t, c)
	accessUsers := map[string]bool{
		userPubAccess: true,
	}
	userPubFrozenAccount, _ := genKey(t, c)
	frozenAccounts := map[string]bool{
		userPubFrozenAccount: true,
	}
	
	feeTiers := []map[string]interface{}{
		{
			"min_amount": "0",
			"max_amount": "1000000",
			"min_volume": "0",
			"max_volume": "2000000",
			"fee_bps":    50,
		},
	}

	feeAddress, _ := genKey(t, c)
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
	assert.Equal(t, accessMode, tok.AccessMode, "token access policy mode mismatch")
	assert.Equal(t, accessUsers[owner.PublicKey], tok.AccessUsers[owner.PublicKey], "token access policy users mismatch")
	assert.Equal(t, frozenAccounts[owner.PublicKey], tok.FrozenAccounts[owner.PublicKey], "token frozen accounts mismatch")
	assert.Equal(t, len(feeTiers), len(tok.FeeTiersList), "token fee tiers length mismatch")
	assert.Equal(t, feeAddress, tok.FeeAddress, "token fee address mismatch")
	assert.Equal(t, freezeAuthorityRevoked, tok.FreezeAuthorityRevoked, "token freeze authority revoked mismatch")
	assert.Equal(t, mintAuthorityRevoked, tok.MintAuthorityRevoked, "token mint authority revoked mismatch")
	assert.Equal(t, updateAuthorityRevoked, tok.UpdateAuthorityRevoked, "token update authority revoked mismatch")
	assert.Equal(t, paused, tok.Paused, "token paused mismatch")
	// Skipping expiredAt deep equality for simplicity
	assert.Equal(t, assetGLBUri, tok.AssetGLBUri, "token asset GLB URI mismatch")
	assert.Equal(t, tokenType, tok.TokenType, "token type mismatch")
	assert.Equal(t, transferable, tok.Transferable, "token transferable mismatch")
	assert.Equal(t, stablecoin, tok.Stablecoin, "token stablecoin mismatch")

	unmarshalLogMint, err := utils.UnmarshalLog[log.Log](out.Logs[1])
	if err != nil {
		t.Fatalf("UnmarshalLog (AddToken.Logs[1]): %v", err)
	}
	assert.Equal(t, unmarshalLogMint.LogType, tokenV1Domain.TOKEN_MINTED_FT_LOG, "mint log type mismatch")
	mint, err := utils.UnmarshalEvent[tokenV1Domain.MintFT](unmarshalLogMint.Event)
	if err != nil {
		t.Fatalf("UnmarshalEvent (AddToken.Logs[1]): %v", err)
	}

	assert.Equal(t, tok.Address, mint.TokenAddress, "mint to address mismatch")
	assert.Equal(t, owner.PublicKey, mint.MintTo, "mint to address mismatch")
	assert.Equal(t, totalSupply, mint.Amount, "mint amount mismatch")
	assert.Equal(t, tokenType, mint.TokenType, "mint token type mismatch")

	unmarshalLogBalance, err := utils.UnmarshalLog[log.Log](out.Logs[2])
	if err != nil {
		t.Fatalf("UnmarshalLog (AddToken.Logs[2]): %v", err)
	}
	assert.Equal(t, unmarshalLogBalance.LogType, tokenV1Domain.TOKEN_BALANCE_INCREASED_FT_LOG, "balance log type mismatch")

	balance, err := utils.UnmarshalEvent[tokenV1Domain.BalanceFT](unmarshalLogBalance.Event)
	if err != nil {
		t.Fatalf("UnmarshalEvent (AddToken.Logs[2]): %v", err)
	}

	assert.Equal(t, tok.Address, balance.TokenAddress,  "balance token address mismatch")
	assert.Equal(t, owner.PublicKey, balance.OwnerAddress, "balance wallet address mismatch")
	assert.Equal(t, totalSupply, balance.Amount, "balance amount mismatch")
	assert.Equal(t, tokenType, balance.TokenType, "balance token type mismatch")

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
	assert.Equal(t, accessMode, tokenState.AccessMode, "token access policy mode mismatch")
	assert.Equal(t, accessUsers[owner.PublicKey], tokenState.AccessUsers[owner.PublicKey], "token access policy users mismatch")
	assert.Equal(t, frozenAccounts[owner.PublicKey], tokenState.FrozenAccounts[owner.PublicKey], "token frozen accounts mismatch")
	// Skipping fee tiers deep equality for simplicity
	assert.Equal(t, feeAddress, tokenState.FeeAddress, "token fee address mismatch")
	assert.Equal(t, freezeAuthorityRevoked, tokenState.FreezeAuthorityRevoked, "token freeze authority revoked mismatch")
	assert.Equal(t, mintAuthorityRevoked, tokenState.MintAuthorityRevoked, "token mint authority revoked mismatch")
	assert.Equal(t, updateAuthorityRevoked, tokenState.UpdateAuthorityRevoked, "token update authority revoked mismatch")
	assert.Equal(t, paused, tokenState.Paused, "token paused mismatch")
	// Skipping expiredAt deep equality for simplicity
	assert.Equal(t, assetGLBUri, tokenState.AssetGLBUri, "token asset GLB URI mismatch")
	assert.Equal(t, tokenType, tokenState.TokenType, "token type mismatch")
	assert.Equal(t, transferable, tokenState.Transferable, "token transferable mismatch")
	assert.Equal(t, stablecoin, tokenState.Stablecoin, "token stablecoin mismatch")

	getBalanceOut, err := c.GetTokenBalance(tok.Address, owner.PublicKey)
	if err != nil {
		t.Fatalf("GetTokenBalance: %v", err)
	}
	var balanceState tokenV1Models.BalanceStateModel
	err = utils.UnmarshalState[tokenV1Models.BalanceStateModel](getBalanceOut.States[0].Object, &balanceState)
	if err != nil {
		t.Fatalf("UnmarshalState (GetTokenBalance.States[0]): %v", err)
	}
	assert.Equal(t, tok.Address, balanceState.TokenAddress, "balance token address mismatch")
	assert.Equal(t, owner.PublicKey, balanceState.OwnerAddress, "balance wallet address mismatch")
	assert.Equal(t, totalSupply, balanceState.Amount, "balance amount mismatch")
	assert.Equal(t, tokenType, balanceState.TokenType, "balance token type mismatch")
	assert.Equal(t, "", balanceState.TokenUUID, "balance token UUID mismatch")
	assert.NotNil(t, balanceState.CreatedAt, "balance token created at mismatch")
	assert.NotNil(t, balanceState.UpdatedAt, "balance token updated at mismatch")
	// ------------------
	//        MINT
	// ------------------

	mintAmount := "1000000"
	mintToken, err := c.MintToken(tok.Address, owner.PublicKey, mintAmount)
	if err != nil {
		t.Fatalf("MintToken: %v", err)
	}

	unmarshalLogMint2, err := utils.UnmarshalLog[log.Log](mintToken.Logs[0])
	if err != nil {
		t.Fatalf("UnmarshalLog (MintToken.Logs[0]): %v", err)
	}
	assert.Equal(t, tokenV1Domain.TOKEN_MINTED_FT_LOG, unmarshalLogMint2.LogType, "mint log type mismatch")
	mint2, err := utils.UnmarshalEvent[tokenV1Domain.MintFT](unmarshalLogMint2.Event)
	if err != nil {
		t.Fatalf("UnmarshalEvent (MintToken.Logs[0]): %v", err)
	}

	assert.Equal(t, tok.Address, mint2.TokenAddress, "mint to address mismatch")
	assert.Equal(t, owner.PublicKey, mint2.MintTo, "mint to address mismatch")
	assert.Equal(t, mintAmount, mint2.Amount, "mint amount mismatch")

	unmmarshalLogSupply, err := utils.UnmarshalLog[log.Log](mintToken.Logs[1])
	if err != nil {
		t.Fatalf("UnmarshalLog (MintToken.Logs[1]): %v", err)
	}
	assert.Equal(t, tokenV1Domain.TOKEN_TOTAL_SUPPLY_INCREASED_LOG, unmmarshalLogSupply.LogType, "supply log type mismatch")
	supply, err := utils.UnmarshalEvent[tokenV1Domain.Supply](unmmarshalLogSupply.Event)
	if err != nil {
		t.Fatalf("UnmarshalEvent (MintToken.Logs[1]): %v", err)
	}
	assert.Equal(t, tok.Address, supply.TokenAddress, "supply token address mismatch")
	assert.Equal(t, mintAmount, supply.Amount, "total supply mismatch after mint")

	unmarshalLogBalanceMint, err := utils.UnmarshalLog[log.Log](mintToken.Logs[2])
	if err != nil {
		t.Fatalf("UnmarshalLog (MintToken.Logs[2]): %v", err)
	}
	assert.Equal(t, tokenV1Domain.TOKEN_BALANCE_INCREASED_FT_LOG, unmarshalLogBalanceMint.LogType, "balance log type mismatch")
	balanceMint, err := utils.UnmarshalEvent[tokenV1Domain.BalanceFT](unmarshalLogBalanceMint.Event)
	if err != nil {
		t.Fatalf("UnmarshalEvent (MintToken.Logs[2]): %v", err)
	}

	assert.Equal(t, tok.Address, balanceMint.TokenAddress, "balance token address mismatch")
	assert.Equal(t, owner.PublicKey, balanceMint.OwnerAddress, "balance wallet address mismatch")
	assert.Equal(t, mintAmount, balanceMint.Amount, "balance amount mismatch after mint")
	assert.Equal(t, tokenType, balanceMint.TokenType, "balance token type mismatch")

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
	assert.Equal(t, sumTotalSupply, tokenState2.TotalSupply, "token total supply mismatch after mint")

	getTokenOutBalanceMint, err := c.GetTokenBalance(tok.Address, owner.PublicKey)
	if err != nil {
		t.Fatalf("GetTokenBalance: %v", err)
	}
	var balanceStateMint tokenV1Models.BalanceStateModel
	err = utils.UnmarshalState[tokenV1Models.BalanceStateModel](getTokenOutBalanceMint.States[0].Object, &balanceStateMint)
	if err != nil {
		t.Fatalf("UnmarshalState (GetTokenBalance.States[0]): %v", err)
	}
	expectedBalanceMint, err := utils.AddBigIntStrings(totalSupply, mintAmount)
	if err != nil {
		t.Fatalf("AddBigIntStrings: %v", err)
	}
	assert.Equal(t, tok.Address, balanceStateMint.TokenAddress, "balance token address mismatch")
	assert.Equal(t, owner.PublicKey, balanceStateMint.OwnerAddress, "balance wallet address mismatch")
	assert.Equal(t, expectedBalanceMint, balanceStateMint.Amount, "balance amount mismatch after mint")
	assert.Equal(t, tokenType, balanceStateMint.TokenType, "balance token type mismatch")
	// ------------------
	//   TRANSFER TOKEN
	// ------------------
	receiver, _ := createWallet(t, c)
	c.SetPrivateKey(ownerPriv)

	transferAmount := "5000000000000000"
	transferToken, err := c.TransferToken(tok.Address, receiver.PublicKey, transferAmount, []string{})
	assert.Error(t, err, "insufficient balance:")
	
	transferAmount = "600000"
	transferToken, err = c.TransferToken(tok.Address, receiver.PublicKey, transferAmount, []string{})
	if err != nil {
		t.Fatalf("TransferToken: %v", err)
	}

	unmarshalLogTransfer, err := utils.UnmarshalLog[log.Log](transferToken.Logs[0])
	if err != nil {
		t.Fatalf("UnmarshalLog (TransferToken.Logs[0]): %v", err)
	}
	assert.Equal(t, unmarshalLogTransfer.LogType, tokenV1Domain.TOKEN_TRANSFERRED_FT_LOG, "transfer log type mismatch")
	transfer, err := utils.UnmarshalEvent[tokenV1Domain.TransferFT](unmarshalLogTransfer.Event)
	if err != nil {
		t.Fatalf("UnmarshalEvent (TransferToken.Logs[0]): %v", err)
	}

	assert.Equal(t, tok.Address, transfer.TokenAddress, "transfer token address mismatch")
	assert.Equal(t, owner.PublicKey, transfer.FromAddress, "transfer from address mismatch")
	assert.Equal(t, receiver.PublicKey, transfer.ToAddress, "transfer to address mismatch")
	assert.Equal(t, transferAmount, transfer.Amount, "transfer amount mismatch")
	assert.Equal(t, tokenType, transfer.TokenType, "transfer token type mismatch")

	unmarshalLogBalanceSender, err := utils.UnmarshalLog[log.Log](transferToken.Logs[1])
	if err != nil {
		t.Fatalf("UnmarshalLog (TransferToken.Logs[1]): %v", err)
	}
	assert.Equal(t, tokenV1Domain.TOKEN_BALANCE_DECREASED_FT_LOG, unmarshalLogBalanceSender.LogType, "balance decreased log type mismatch")

	balanceSender, err := utils.UnmarshalEvent[tokenV1Domain.BalanceFT](unmarshalLogBalanceSender.Event)
	if err != nil {
		t.Fatalf("UnmarshalEvent (TransferToken.Logs[1]): %v", err)
	}

	assert.Equal(t, tok.Address, balanceSender.TokenAddress, "balance token address mismatch")
	assert.Equal(t, owner.PublicKey, balanceSender.OwnerAddress, "balance wallet address mismatch")
	assert.Equal(t, transferAmount, balanceSender.Amount, "balance amount mismatch after transfer")
	assert.Equal(t, tokenType, balanceSender.TokenType, "balance token type mismatch")

	unmarshalLogBalanceReceiver, err := utils.UnmarshalLog[log.Log](transferToken.Logs[2])
	if err != nil {
		t.Fatalf("UnmarshalLog (TransferToken.Logs[2]): %v", err)
	}

	balanceReceiver, err := utils.UnmarshalEvent[tokenV1Domain.BalanceFT](unmarshalLogBalanceReceiver.Event)
	if err != nil {
		t.Fatalf("UnmarshalEvent (TransferToken.Logs[2]): %v", err)
	}
	feeAmount := "3000"
	transferAfterFee, err := utils.SubBigIntStrings(transferAmount, feeAmount)
	if err != nil {
		t.Fatalf("SubBigIntStrings: %v", err)
	}
	assert.Equal(t, tok.Address, balanceReceiver.TokenAddress, "balance token address mismatch")
	assert.Equal(t, receiver.PublicKey, balanceReceiver.OwnerAddress, "balance wallet address mismatch")
	assert.Equal(t, transferAfterFee, balanceReceiver.Amount, "balance amount mismatch after transfer")
	assert.Equal(t, tokenType, balanceReceiver.TokenType, "balance token type mismatch")

	unmarshalLogFee, err := utils.UnmarshalLog[log.Log](transferToken.Logs[3])
	if err != nil {
		t.Fatalf("UnmarshalLog (TransferToken.Logs[3]): %v", err)
	}
	assert.Equal(t, tokenV1Domain.TOKEN_FEE_LOG, unmarshalLogFee.LogType, "fee log type mismatch")

	fee, err := utils.UnmarshalEvent[tokenV1Domain.Fee](unmarshalLogFee.Event)
	if err != nil {
		t.Fatalf("UnmarshalEvent (TransferToken.Logs[3]): %v", err)
	}
	assert.Equal(t, tok.Address, fee.TokenAddress, "fee token address mismatch")
	assert.Equal(t, feeAmount, fee.Amount, "fee amount mismatch")
	assert.Equal(t, feeAddress, fee.FeeAddress, "fee address mismatch")
	assert.Equal(t, transferAfterFee, fee.AmountAfterFee, "amount after fee mismatch")

	unmarshalLogReceiverFee, err := utils.UnmarshalLog[log.Log](transferToken.Logs[4])
	if err != nil {
		t.Fatalf("UnmarshalLog (TransferToken.Logs[4]): %v", err)
	}

	feeReceiver, err := utils.UnmarshalEvent[tokenV1Domain.BalanceFT](unmarshalLogReceiverFee.Event)
	if err != nil {
		t.Fatalf("UnmarshalEvent (TransferToken.Logs[4]): %v", err)
	}
	assert.Equal(t, tok.Address, feeReceiver.TokenAddress, "fee receiver token address mismatch")
	assert.Equal(t, feeAddress, feeReceiver.OwnerAddress, "fee receiver wallet address mismatch")
	assert.Equal(t, feeAmount, feeReceiver.Amount, "fee receiver balance amount mismatch after transfer")
	assert.Equal(t, tokenType, feeReceiver.TokenType, "fee receiver token type mismatch")

	getTokenOut10, err := c.GetToken(tok.Address, "", "")
	if err != nil {
		t.Fatalf("GetToken: %v", err)
	}

	var tokenState3 tokenV1Models.TokenStateModel
	err = utils.UnmarshalState[tokenV1Models.TokenStateModel](getTokenOut10.States[0].Object, &tokenState3)
	if err != nil {
		t.Fatalf("UnmarshalState (GetToken.States[0]): %v", err)
	}

	assert.Equal(t, tokenState3.TotalSupply, sumTotalSupply, "token total supply mismatch after transfer")

	getTokenOutBalanceTransferSender, err := c.GetTokenBalance(tok.Address, owner.PublicKey)
	if err != nil {
		t.Fatalf("GetTokenBalance: %v", err)
	}
	var balanceStateTransferSender tokenV1Models.BalanceStateModel
	err = utils.UnmarshalState[tokenV1Models.BalanceStateModel](getTokenOutBalanceTransferSender.States[0].Object, &balanceStateTransferSender)
	if err != nil {
		t.Fatalf("UnmarshalState (GetTokenBalance.States[0]): %v", err)
	}
	expectedBalance := "100400000"
	assert.Equal(t, tok.Address, balanceStateTransferSender.TokenAddress, "balance token address mismatch")
	assert.Equal(t, owner.PublicKey, balanceStateTransferSender.OwnerAddress, "balance wallet address mismatch")
	assert.Equal(t, expectedBalance, balanceStateTransferSender.Amount, "balance amount mismatch after transfer")
	assert.Equal(t, tokenType, balanceStateTransferSender.TokenType, "balance token type mismatch")

	getTokenOutBalanceTransferReceiver, err := c.GetTokenBalance(tok.Address, receiver.PublicKey)
	if err != nil {
		t.Fatalf("GetTokenBalance: %v", err)
	}
	var balanceStateTransferReceiver tokenV1Models.BalanceStateModel
	err = utils.UnmarshalState[tokenV1Models.BalanceStateModel](getTokenOutBalanceTransferReceiver.States[0].Object, &balanceStateTransferReceiver)
	if err != nil {
		t.Fatalf("UnmarshalState (GetTokenBalance.States[0]): %v", err)
	}
	expectedBalanceTransferReceiver := "597000"
	assert.Equal(t, expectedBalanceTransferReceiver, balanceStateTransferReceiver.Amount, "balance amount mismatch after transfer")

	getTokenOutBalanceFeeReceiver, err := c.GetTokenBalance(tok.Address, feeAddress)
	if err != nil {
		t.Fatalf("GetTokenBalance: %v", err)
	}
	var balanceStateFeeReceiver tokenV1Models.BalanceStateModel
	err = utils.UnmarshalState[tokenV1Models.BalanceStateModel](getTokenOutBalanceFeeReceiver.States[0].Object, &balanceStateFeeReceiver)
	if err != nil {
		t.Fatalf("UnmarshalState (GetTokenBalance.States[0]): %v", err)
	}
	assert.Equal(t, feeAmount, balanceStateFeeReceiver.Amount, "fee receiver balance amount mismatch after transfer")

	// ------------------
	//        BURN
	// ------------------
	burnAmount := "5000000000000000000000"
	burnToken, err := c.BurnToken(tok.Address, burnAmount, []string{})
	require.Error(t, err)
	assert.ErrorContains(t, err, "insufficient balance")

	
	burnAmount = "1500"
	burnToken, err = c.BurnToken(tok.Address, burnAmount, []string{})
	if err != nil {
		t.Fatalf("BurnToken: %v", err)
	}

	unmarshalLogBurn, err := utils.UnmarshalLog[log.Log](burnToken.Logs[0])
	if err != nil {
		t.Fatalf("UnmarshalLog (BurnToken.Logs[0]): %v", err)
	}
	assert.Equal(t, unmarshalLogBurn.LogType, tokenV1Domain.TOKEN_BURNED_FT_LOG, "burn log type mismatch")
	burn, err := utils.UnmarshalEvent[tokenV1Domain.BurnFT](unmarshalLogBurn.Event)
	if err != nil {
		t.Fatalf("UnmarshalEvent (BurnToken.Logs[0]): %v", err)
	}

	assert.Equal(t, tok.Address, burn.TokenAddress, "burn token address mismatch")
	assert.Equal(t, owner.PublicKey, burn.BurnFrom, "burn from address mismatch")
	assert.Equal(t, burnAmount, burn.Amount, "burn amount mismatch")
	assert.Equal(t, tokenType, burn.TokenType, "burn token type mismatch")

	unmmarshalLogSupply2, err := utils.UnmarshalLog[log.Log](burnToken.Logs[1])
	if err != nil {
		t.Fatalf("UnmarshalLog (BurnToken.Logs[1]): %v", err)
	}
	assert.Equal(t, tokenV1Domain.TOKEN_TOTAL_SUPPLY_DECREASED_LOG, unmmarshalLogSupply2.LogType, "supply log type mismatch")
	supply2, err := utils.UnmarshalEvent[tokenV1Domain.Supply](unmmarshalLogSupply2.Event)
	if err != nil {
		t.Fatalf("UnmarshalEvent (BurnToken.Logs[1]): %v", err)
	}
	assert.Equal(t, tok.Address, supply2.TokenAddress, "supply token address mismatch")
	assert.Equal(t, burnAmount, supply2.Amount, "total supply mismatch after burn")

	unmarshalLogBalance3, err := utils.UnmarshalLog[log.Log](burnToken.Logs[2])
	if err != nil {
		t.Fatalf("UnmarshalLog (BurnToken.Logs[2]): %v", err)
	}
	assert.Equal(t, tokenV1Domain.TOKEN_BALANCE_DECREASED_FT_LOG, unmarshalLogBalance3.LogType, "balance log type mismatch")
	balance3, err := utils.UnmarshalEvent[tokenV1Domain.BalanceFT](unmarshalLogBalance3.Event)
	if err != nil {
		t.Fatalf("UnmarshalEvent (BurnToken.Logs[2]): %v", err)
	}

	assert.Equal(t, tok.Address, balance3.TokenAddress, "balance token address mismatch")
	assert.Equal(t, owner.PublicKey, balance3.OwnerAddress, "balance wallet address mismatch")
	assert.Equal(t, burnAmount, balance3.Amount, "balance amount mismatch after burn")
	assert.Equal(t, tokenType, balance3.TokenType, "balance token type mismatch")

	getTokenOut4, err := c.GetToken(tok.Address, "", "")
	if err != nil {
		t.Fatalf("GetToken: %v", err)
	}

	var tokenState4 tokenV1Models.TokenStateModel
	err = utils.UnmarshalState[tokenV1Models.TokenStateModel](getTokenOut4.States[0].Object, &tokenState4)
	if err != nil {
		t.Fatalf("UnmarshalState (GetToken.States[0]): %v", err)
	}

	subTotalSupply, err := utils.SubBigIntStrings(sumTotalSupply, burnAmount)
	if err != nil {
		t.Fatalf("SubBigIntStrings: %v", err)
	}
	assert.Equal(t, tokenState4.TotalSupply, subTotalSupply, "token total supply mismatch after burn")

	getTokenOutBalanceBurn, err := c.GetTokenBalance(tok.Address, owner.PublicKey)
	if err != nil {
		t.Fatalf("GetTokenBalance: %v", err)
	}
	var balanceStateBurn tokenV1Models.BalanceStateModel
	err = utils.UnmarshalState[tokenV1Models.BalanceStateModel](getTokenOutBalanceBurn.States[0].Object, &balanceStateBurn)
	if err != nil {
		t.Fatalf("UnmarshalState (GetTokenBalance.States[0]): %v", err)
	}
	expectedBalanceAfterBurn := "100398500"
	assert.Equal(t, expectedBalanceAfterBurn, balanceStateBurn.Amount, "balance amount mismatch after burn")
	// ------------------
	// CHANGE ACCESS MODE
	// ------------------

	// Remove allow users
	removeAllowUsers, err := c.RemoveAllowUsers(tok.Address, domain.DENY_ACCESS_MODE, map[string]bool{
		userPubAccess: true,
	})
	if err != nil {
		t.Fatalf("RemoveAllowUsers: %v", err)
	}

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
	assert.Equal(t, tok.Address, changeAccess.Address, "access policy token address mismatch")
	assert.Equal(t, tokenV1Domain.ALLOW_ACCESS_MODE, changeAccess.AccessMode, "access policy mode mismatch after change access mode")

	getTokenOut5, err := c.GetToken(tok.Address, "", "")
	if err != nil {
		t.Fatalf("GetToken: %v", err)
	}

	var tokenState5 tokenV1Models.TokenStateModel
	err = utils.UnmarshalState[tokenV1Models.TokenStateModel](getTokenOut5.States[0].Object, &tokenState5)
	if err != nil {
		t.Fatalf("UnmarshalState (GetToken.States[0]): %v", err)
	}
	assert.Equal(t, tokenV1Domain.ALLOW_ACCESS_MODE, tokenState5.AccessMode, "token access mode mismatch after change access mode")
	assert.Empty(t, tokenState5.AccessUsers, "access users should be empty")	// ------------------
	//    ALLOW USERS
	// ------------------

	// Add allow users
	receiverPub, _ := genKey(t, c)
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
	assert.Equal(t, tok.Address, allow.Address, "access policy token address mismatch")
	assert.Equal(t, tokenV1Domain.ALLOW_ACCESS_MODE, allow.AccessMode, "access policy mode mismatch after allow")
	assert.Equal(t, true, allow.AccessUsers[receiverPub], "access policy users mismatch after allow")

	getTokenOutAllow, err := c.GetToken(tok.Address, "", "")
	if err != nil {
		t.Fatalf("GetToken: %v", err)
	}

	var tokenStateAllow tokenV1Models.TokenStateModel
	err = utils.UnmarshalState[tokenV1Models.TokenStateModel](getTokenOutAllow.States[0].Object, &tokenStateAllow)
	if err != nil {
		t.Fatalf("UnmarshalState (GetToken.States[0]): %v", err)
	}

	assert.Equal(t, tokenV1Domain.ALLOW_ACCESS_MODE, tokenStateAllow.AccessMode, "token access mode mismatch after allow")
	assert.Equal(t, true, tokenStateAllow.AccessUsers[receiverPub], "token access users mismatch after allow")

	// Remove allow users
	removeAllowUsers, err = c.RemoveAllowUsers(tok.Address, domain.ALLOW_ACCESS_MODE, map[string]bool{
		receiverPub: true,
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
	assert.Contains(t, removeAllow.AccessUsers, receiverPub, "removed user should be present in remove event payload")
	assert.True(t, removeAllow.AccessUsers[receiverPub], "removed user payload should be true in remove event")

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

	feeTiersTest := []map[string]interface{}{
		{
			"min_amount": "0",
			"max_amount": "200000",
			"min_volume": "0",
			"max_volume": "2000000",
			"fee_bps":    50,
		},
	}

	updateFeeTiers, er := c.UpdateFeeTiers(tok.Address, feeTiersTest)
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
	assert.Equal(t, feeTiersTest[0]["min_amount"], feeTiersEvent.FeeTiersList[0].MinAmount, "fee tiers min amount mismatch")
	assert.Equal(t, feeTiersTest[0]["max_amount"], feeTiersEvent.FeeTiersList[0].MaxAmount, "fee tiers max amount mismatch")
	assert.Equal(t, feeTiersTest[0]["min_volume"], feeTiersEvent.FeeTiersList[0].MinVolume, "fee tiers min volume mismatch")
	assert.Equal(t, feeTiersTest[0]["max_volume"], feeTiersEvent.FeeTiersList[0].MaxVolume, "fee tiers max volume mismatch")
	assert.Equal(t, feeTiersTest[0]["fee_bps"], feeTiersEvent.FeeTiersList[0].FeeBps, "fee tiers bps mismatch")

	getTokenOut13, err := c.GetToken(tok.Address, "", "")
	if err != nil {
		t.Fatalf("GetToken: %v", err)
	}

	var tokenState13 tokenV1Models.TokenStateModel
	err = utils.UnmarshalState[tokenV1Models.TokenStateModel](getTokenOut13.States[0].Object, &tokenState13)
	if err != nil {
		t.Fatalf("UnmarshalState (GetToken.States[0]): %v", err)
	}

	assert.Equal(t, feeTiersTest[0]["min_amount"], tokenState13.FeeTiersList[0].MinAmount, "token state fee tiers min amount mismatch")
	assert.Equal(t, feeTiersTest[0]["max_amount"], tokenState13.FeeTiersList[0].MaxAmount, "token state fee tiers max amount mismatch")
	assert.Equal(t, feeTiersTest[0]["min_volume"], tokenState13.FeeTiersList[0].MinVolume, "token state fee tiers min volume mismatch")
	assert.Equal(t, feeTiersTest[0]["max_volume"], tokenState13.FeeTiersList[0].MaxVolume, "token state fee tiers max volume mismatch")
	assert.Equal(t, feeTiersTest[0]["fee_bps"], tokenState13.FeeTiersList[0].FeeBps, "token state fee tiers bps mismatch")

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

	getTokenOut14, err := c.GetToken(tok.Address, "", "")
	if err != nil {
		t.Fatalf("GetToken: %v", err)
	}

	var tokenState14 tokenV1Models.TokenStateModel
	err = utils.UnmarshalState[tokenV1Models.TokenStateModel](getTokenOut14.States[0].Object, &tokenState14)
	if err != nil {
		t.Fatalf("UnmarshalState (GetToken.States[0]): %v", err)
	}

	assert.Equal(t, feeAddress, tokenState14.FeeAddress, "token state fee address mismatch")

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
	//    TRANSFERABLE/UNTRANSFERABLE
	// ------------------
	untransferableToken, err := c.UntransferableToken(tok.Address, false)
	if err != nil {
		t.Fatalf("UntransferableToken: %v", err)
	}

	unmarshalLogUntransferableToken, err := utils.UnmarshalLog[log.Log](untransferableToken.Logs[0])
	if err != nil {
		t.Fatalf("UnmarshalLog (UntransferableToken.Logs[0]): %v", err)
	}
	assert.Equal(t, tokenV1Domain.TOKEN_UNTRANSFERABLE_LOG, unmarshalLogUntransferableToken.LogType, "untransferable token log type mismatch")
	untransferableTokenEvent, err := utils.UnmarshalEvent[tokenV1Domain.TransferPolicy](unmarshalLogUntransferableToken.Event)
	if err != nil {
		t.Fatalf("UnmarshalEvent (UntransferableToken.Logs[0]): %v", err)
	}
	assert.Equal(t, tok.Address, untransferableTokenEvent.TokenAddress, "untransferable token address mismatch")
	assert.Equal(t, false, untransferableTokenEvent.Enable, "token transferable state mismatch after untransferable")

	getTokenOut19, err := c.GetToken(tok.Address, "", "")
	if err != nil {
		t.Fatalf("GetToken: %v", err)
	}

	var tokenState19 tokenV1Models.TokenStateModel
	err = utils.UnmarshalState[tokenV1Models.TokenStateModel](getTokenOut19.States[0].Object, &tokenState19)
	if err != nil {
		t.Fatalf("UnmarshalState (GetToken.States[0]): %v", err)
	}
	assert.Equal(t, false, tokenState19.Transferable, "token transferable mismatch after transferable update")

	transferAmount = "500000"
	transferToken, err = c.TransferToken(tok.Address, receiver.PublicKey, transferAmount, []string{})
	assert.Error(t, err, "expected error when transferring untransferable token")
	assert.Contains(t, err.Error(), "token is not transferable", "unexpected error message when transferring untransferable token")

	transferableToken, err := c.TransferableToken(tok.Address, true)
	if err != nil {
		t.Fatalf("TransferableToken: %v", err)
	}

	unmarshalLogTransferableToken, err := utils.UnmarshalLog[log.Log](transferableToken.Logs[0])
	if err != nil {
		t.Fatalf("UnmarshalLog (TransferableToken.Logs[0]): %v", err)
	}
	assert.Equal(t, tokenV1Domain.TOKEN_TRANSFERABLE_LOG, unmarshalLogTransferableToken.LogType, "transferable token log type mismatch")

	transferableTokenEvent, err := utils.UnmarshalEvent[tokenV1Domain.TransferPolicy](unmarshalLogTransferableToken.Event)
	if err != nil {
		t.Fatalf("UnmarshalEvent (TransferableToken.Logs[0]): %v", err)
	}
	assert.Equal(t, tok.Address, transferableTokenEvent.TokenAddress, "transferable token address mismatch")
	assert.Equal(t, true, transferableTokenEvent.Enable, "token transferable state mismatch after transferable")

	getTokenOut20, err := c.GetToken(tok.Address, "", "")
	if err != nil {
		t.Fatalf("GetToken: %v", err)
	}

	var tokenStateTransferable tokenV1Models.TokenStateModel
	err = utils.UnmarshalState[tokenV1Models.TokenStateModel](getTokenOut20.States[0].Object, &tokenStateTransferable)
	if err != nil {
		t.Fatalf("UnmarshalState (GetToken.States[0]): %v", err)
	}
	assert.Equal(t, true, tokenStateTransferable.Transferable, "token transferable mismatch after transferable update")

	transferToken, err = c.TransferToken(tok.Address, receiver.PublicKey, transferAmount, []string{})
	if err != nil {
		t.Fatalf("TransferToken: %v", err)
	}

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

	getTokenOut20, err = c.GetToken(tok.Address, "", "")
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
	assert.Equal(t, tokenV1Domain.TOKEN_MINT_AUTHORITY_REVOKED_LOG, unmarshalLogRevokeMint.LogType, "revoke mint authority log type mismatch")

	revokeMintEvent, err := utils.UnmarshalEvent[tokenV1Domain.Token](unmarshalLogRevokeMint.Event)
	if err != nil {
		t.Fatalf("UnmarshalEvent (RevokeMintAuthority.Logs[0]): %v", err)
	}
	assert.Equal(t, tok.Address, revokeMintEvent.Address, "revoke mint authority event token address mismatch")
	assert.Equal(t, true, revokeMintEvent.MintAuthorityRevoked, "revoke mint authority event mint authority revoked mismatch")

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
	assert.Equal(t, tokenV1Domain.TOKEN_UPDATE_AUTHORITY_REVOKED_LOG, unmarshalLogRevokeUpdate.LogType, "revoke update authority log type mismatch")

	revokeUpdateEvent, err := utils.UnmarshalEvent[tokenV1Domain.Token](unmarshalLogRevokeUpdate.Event)
	if err != nil {
		t.Fatalf("UnmarshalEvent (RevokeUpdateAuthority.Logs[0]): %v", err)
	}
	assert.Equal(t, tok.Address, revokeUpdateEvent.Address, "revoke update authority event token address mismatch")
	assert.Equal(t, true, revokeUpdateEvent.UpdateAuthorityRevoked, "revoke update authority event update authority revoked mismatch")

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

	getTokenOutBalanceOwner, err := c.GetTokenBalance(tok.Address, owner.PublicKey)
	if err != nil {
		t.Fatalf("GetTokenBalance: %v", err)
	}
	var balanceStateOwner tokenV1Models.BalanceStateModel
	err = utils.UnmarshalState[tokenV1Models.BalanceStateModel](getTokenOutBalanceOwner.States[0].Object, &balanceStateOwner)
	if err != nil {
		t.Fatalf("UnmarshalState (GetTokenBalance.States[0]): %v", err)
	}

	amount := "99898500"
	assert.Equal(t, balanceStateOwner.TokenAddress, tok.Address, "token address mismatch in balance state for owner")
	assert.Equal(t, balanceStateOwner.OwnerAddress, owner.PublicKey, "owner address mismatch in balance state for owner")
	assert.Equal(t, balanceStateOwner.Amount, amount, "token amount mismatch in balance state for owner")
	assert.NotNil(t, balanceStateOwner.CreatedAt, "created at is nil for owner")
	assert.NotNil(t, balanceStateOwner.UpdatedAt, "updated at is nil for owner")
	assert.Equal(t, balanceStateOwner.TokenUUID, "", "token uuid mismatch in balance state for owner")

	listOfBalances, err := c.ListTokenBalances("", "", domain.FUNGIBLE, 1, 10, true)
	if err != nil {
		t.Fatalf("ListTokenBalances: %v", err)
	}

	require.NotEmpty(t, listOfBalances.States, "expected at least one state in ListTokenBalances response")
	var balanceStateList []tokenV1Models.BalanceStateModel
	err = utils.UnmarshalState[[]tokenV1Models.BalanceStateModel](listOfBalances.States[0].Object, &balanceStateList)
	if err != nil {
		t.Fatalf("UnmarshalState (ListTokenBalances.States[0].Object): %v", err)
	}

	require.NotEmpty(t, balanceStateList, "expected at least one balance in list")
	require.Equal(t, 10, len(balanceStateList), "expected exactly ten balances in list")
	require.NotNil(t, balanceStateList[0].TokenAddress, "token address is nil for balance in list")
	require.NotNil(t, balanceStateList[0].OwnerAddress, "owner address is nil for balance in list")
	require.NotNil(t, balanceStateList[0].Amount, "amount is nil for balance in list")
	require.NotNil(t, balanceStateList[0].TokenUUID, "token uuid is nil for balance in list")
	require.NotNil(t, balanceStateList[0].CreatedAt, "created at is nil for balance in list")
	require.NotNil(t, balanceStateList[0].UpdatedAt, "updated at is nil for balance in list")


	listTokens, err := c.ListTokens("", "", "", domain.FUNGIBLE, 1, 10, true);
	if err != nil {
		t.Fatalf("ListTokens: %v", err)
	}

	require.NotEmpty(t, listTokens.States, "expected at least one state in ListTokens response")
	var tokenStateList []tokenV1Models.TokenStateModel
	err = utils.UnmarshalState[[]tokenV1Models.TokenStateModel](listTokens.States[0].Object, &tokenStateList)
	if err != nil {
		t.Fatalf("UnmarshalState (ListTokens.States[0].Object): %v", err)
	}

	require.NotEmpty(t, tokenStateList, "expected at least one token in list")
	require.Equal(t, 10, len(tokenStateList), "expected exactly ten tokens in list")
	require.NotNil(t, tokenStateList[0].Address, "token address is nil for token in list")
	require.NotNil(t, tokenStateList[0].Symbol, "token symbol is nil for token in list")
	require.NotNil(t, tokenStateList[0].Name, "token name is nil for token in list")
	require.NotNil(t, tokenStateList[0].Decimals, "token decimals is nil for token in list")
	require.NotNil(t, tokenStateList[0].TotalSupply, "token total supply is nil for token in list")
	require.NotNil(t, tokenStateList[0].Description, "token description is nil for token in list")
	require.NotNil(t, tokenStateList[0].Image, "token image is nil for token in list")
	require.NotNil(t, tokenStateList[0].Website, "token website is nil for token in list")
	require.NotNil(t, tokenStateList[0].TagsSocialMedia, "token tags social media is nil for token in list")
	require.NotNil(t, tokenStateList[0].TagsCategory, "token tags category is nil for token in list")
	require.NotNil(t, tokenStateList[0].Tags, "token tags is nil for token in list")
	require.NotNil(t, tokenStateList[0].Creator, "token creator is nil for token in list")
	require.NotNil(t, tokenStateList[0].CreatorWebsite, "token creator website is nil for token in list")
	require.NotNil(t, tokenStateList[0].AccessMode, "token access mode is nil for token in list")
	require.Nil(t, tokenStateList[0].AccessUsers, "token access users is nil for token in list")
	require.NotNil(t, tokenStateList[0].FrozenAccounts, "token frozen accounts is nil for token in list")
	require.NotNil(t, tokenStateList[0].FeeTiersList, "token fee tiers list is nil for token in list")
	require.NotNil(t, tokenStateList[0].FeeAddress, "token fee address is nil for token in list")
	require.NotNil(t, tokenStateList[0].FreezeAuthorityRevoked, "token freeze authority revoked is nil for token in list")
	require.NotNil(t, tokenStateList[0].MintAuthorityRevoked, "token mint authority revoked is nil for token in list")
	require.NotNil(t, tokenStateList[0].UpdateAuthorityRevoked, "token update authority revoked is nil for token in list")
	require.NotNil(t, tokenStateList[0].Paused, "token paused is nil for token in list")
	require.NotNil(t, tokenStateList[0].AssetGLBUri, "token asset GLB URI is nil for token in list")
	require.NotNil(t, tokenStateList[0].TokenType, "token type is nil for token in list")
	require.NotNil(t, tokenStateList[0].Transferable, "token transferable is nil for token in list")
	require.NotNil(t, tokenStateList[0].Stablecoin, "token stablecoin is nil for token in list")
	require.NotNil(t, tokenStateList[0].CreatedAt, "created at is nil for token in list")
	require.NotNil(t, tokenStateList[0].UpdatedAt, "updated at is nil for token in list")

}