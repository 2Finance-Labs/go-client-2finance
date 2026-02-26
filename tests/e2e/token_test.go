package e2e_test

import (
	"testing"
	"time"

	client2f "github.com/2Finance-Labs/go-client-2finance/client_2finance"
	"github.com/2Finance-Labs/go-client-2finance/tests"
	"github.com/stretchr/testify/require"

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
	// Mint (envelope + unmarshal + validate + logs + events)
	// -------------------------
	mintAmt := amt(35, dec)

	mintOut, err := c.MintToken(tok.Address, owner.PublicKey, mintAmt, dec, tok.TokenType)
	require.NoError(t, err, "MintToken")

	// States (Mint, Supply, Balance)
	tests.RequireStateObjectsNotNil(t, mintOut.States, 3)

	var mint tokenV1Domain.Mint
	tests.UnmarshalState(t, mintOut.States[0].Object, &mint)
	require.Equal(t, tok.Address, mint.TokenAddress)
	require.Equal(t, owner.PublicKey, mint.MintTo)
	require.Equal(t, mintAmt, mint.Amount)
	require.Equal(t, tok.TokenType, mint.TokenType)

	var supplyAfterMint tokenV1Domain.Supply
	tests.UnmarshalState(t, mintOut.States[1].Object, &supplyAfterMint)
	require.Equal(t, tok.Address, supplyAfterMint.TokenAddress)
	require.NotEmpty(t, supplyAfterMint.Amount)

	var balAfterMint tokenV1Domain.Balance
	tests.UnmarshalState(t, mintOut.States[2].Object, &balAfterMint)
	require.Equal(t, tok.Address, balAfterMint.TokenAddress)
	require.Equal(t, owner.PublicKey, balAfterMint.OwnerAddress)
	require.Equal(t, tok.TokenType, balAfterMint.TokenType)
	require.NotEmpty(t, balAfterMint.Amount)

	// Logs 
	tests.RequireLogsBase(t, mintOut.Logs, 3)
	tests.RequireLogTypesInOrder(t, mintOut.Logs, []string{
		domain.TOKEN_MINTED_LOG,
		domain.TOKEN_TOTAL_SUPPLY_INCREASED_LOG,
		domain.TOKEN_BALANCE_INCREASED_LOG,
	})

	// Events 
	mintEvent := tests.UnmarshalJSONB[map[string]any](t, mintOut.Logs[0].Event)
	require.Equal(t, tok.Address, tests.RequireMapFieldString(t, mintEvent, "token_address"))
	require.Equal(t, owner.PublicKey, tests.RequireMapFieldString(t, mintEvent, "mint_to"))
	require.Equal(t, mintAmt, tests.RequireMapFieldString(t, mintEvent, "amount"))
	require.Equal(t, tok.TokenType, tests.RequireMapFieldString(t, mintEvent, "token_type"))

	supplyMintEvent := tests.UnmarshalJSONB[map[string]any](t, mintOut.Logs[1].Event)
	require.Equal(t, tok.Address, tests.RequireMapFieldString(t, supplyMintEvent, "token_address"))
	require.Equal(t, mintAmt, tests.RequireMapFieldString(t, supplyMintEvent, "amount_delta"))
	require.NotEmpty(t, tests.RequireMapFieldString(t, supplyMintEvent, "total_supply"))

	balanceMintEvent := tests.UnmarshalJSONB[map[string]any](t, mintOut.Logs[2].Event)
	require.Equal(t, tok.Address, tests.RequireMapFieldString(t, balanceMintEvent, "token_address"))
	require.Equal(t, owner.PublicKey, tests.RequireMapFieldString(t, balanceMintEvent, "owner"))
	require.Equal(t, mintAmt, tests.RequireMapFieldString(t, balanceMintEvent, "amount_delta"))
	require.Equal(t, tok.TokenType, tests.RequireMapFieldString(t, balanceMintEvent, "token_type"))

	// -------------------------
	// Burn (envelope + unmarshal + validate + logs + events)
	// -------------------------
	burnAmt := amt(12, dec)

	burnOut, err := c.BurnToken(tok.Address, burnAmt, dec, tok.TokenType, "")
	require.NoError(t, err, "BurnToken")

	// States (Burn, Supply, Balance)
	tests.RequireStateObjectsNotNil(t, burnOut.States, 3)

	var burn tokenV1Domain.Burn
	tests.UnmarshalState(t, burnOut.States[0].Object, &burn)
	require.Equal(t, tok.Address, burn.TokenAddress)
	require.Equal(t, owner.PublicKey, burn.BurnFrom)
	require.Equal(t, burnAmt, burn.Amount)
	require.Equal(t, tok.TokenType, burn.TokenType)

	var supplyAfterBurn tokenV1Domain.Supply
	tests.UnmarshalState(t, burnOut.States[1].Object, &supplyAfterBurn)
	require.Equal(t, tok.Address, supplyAfterBurn.TokenAddress)
	require.NotEmpty(t, supplyAfterBurn.Amount)

	var balAfterBurn tokenV1Domain.Balance
	tests.UnmarshalState(t, burnOut.States[2].Object, &balAfterBurn)
	require.Equal(t, tok.Address, balAfterBurn.TokenAddress)
	require.Equal(t, owner.PublicKey, balAfterBurn.OwnerAddress)
	require.Equal(t, tok.TokenType, balAfterBurn.TokenType)
	require.NotEmpty(t, balAfterBurn.Amount)

	// Logs 
	tests.RequireLogsBase(t, burnOut.Logs, 3)
	tests.RequireLogTypesInOrder(t, burnOut.Logs, []string{
		domain.TOKEN_BURNED_LOG,
		domain.TOKEN_TOTAL_SUPPLY_DECREASED_LOG,
		domain.TOKEN_BALANCE_DECREASED_LOG,
	})

	// Events 
	burnEvent := tests.UnmarshalJSONB[map[string]any](t, burnOut.Logs[0].Event)
	require.Equal(t, tok.Address, tests.RequireMapFieldString(t, burnEvent, "token_address"))
	require.Equal(t, owner.PublicKey, tests.RequireMapFieldString(t, burnEvent, "burn_from"))
	require.Equal(t, burnAmt, tests.RequireMapFieldString(t, burnEvent, "amount"))
	require.Equal(t, tok.TokenType, tests.RequireMapFieldString(t, burnEvent, "token_type"))

	supplyBurnEvent := tests.UnmarshalJSONB[map[string]any](t, burnOut.Logs[1].Event)
	require.Equal(t, tok.Address, tests.RequireMapFieldString(t, supplyBurnEvent, "token_address"))
	
	if v, ok := supplyBurnEvent["amount_delta"]; ok {
		deltaStr, _ := v.(string)
		require.Contains(t, []string{"-" + burnAmt, burnAmt}, deltaStr)
	}
	if v, ok := supplyBurnEvent["amount"]; ok {
		amtStr, _ := v.(string)
		if amtStr != "" {
			require.Equal(t, burnAmt, amtStr)
		}
	}

	balanceBurnEvent := tests.UnmarshalJSONB[map[string]any](t, burnOut.Logs[2].Event)
	require.Equal(t, tok.Address, tests.RequireMapFieldString(t, balanceBurnEvent, "token_address"))
	require.Equal(t, owner.PublicKey, tests.RequireMapFieldString(t, balanceBurnEvent, "owner"))
	require.Equal(t, tok.TokenType, tests.RequireMapFieldString(t, balanceBurnEvent, "token_type"))
	if v, ok := balanceBurnEvent["balance_after"]; ok {
		afterStr, _ := v.(string)
		if afterStr != "" {
			require.Equal(t, balAfterBurn.Amount, afterStr)
		}
	}

	// -------------------------
	// AllowUsers (envelope + unmarshal + validate + logs)
	// -------------------------
	receiver, _ := createWallet(t, c)
	c.SetPrivateKey(ownerPriv)

	allowOut, err := c.AllowUsers(tok.Address, map[string]bool{receiver.PublicKey: true})
	require.NoError(t, err, "AllowUsers")

	tests.RequireStateObjectsNotNil(t, allowOut.States, 1)

	var accessPolicy tokenV1Domain.AccessPolicy
	tests.UnmarshalState(t, allowOut.States[0].Object, &accessPolicy)
	require.NotEmpty(t, accessPolicy.Mode)
	require.NotNil(t, accessPolicy.Users)
	require.True(t, accessPolicy.Users[receiver.PublicKey])

	// Logs 
	if len(allowOut.Logs) > 0 {
		tests.RequireLogsBase(t, allowOut.Logs, len(allowOut.Logs))
		_ = tests.FindLogByType(t, allowOut.Logs, domain.TOKEN_USERS_ALLOWED_LOG)
	}

	// -------------------------
	// Transfer (envelope + unmarshal + validate + logs)
	// -------------------------
	trAmt := amt(1, dec)

	trOut, err := c.TransferToken(tok.Address, receiver.PublicKey, trAmt, dec, tok.TokenType, "")
	require.NoError(t, err, "TransferToken")

	tests.RequireStateObjectsNotNil(t, trOut.States, 1)

	var tr tokenV1Domain.Transfer
	tests.UnmarshalState(t, trOut.States[0].Object, &tr)
	require.Equal(t, tok.Address, tr.TokenAddress)
	require.Equal(t, owner.PublicKey, tr.FromAddress)
	require.Equal(t, receiver.PublicKey, tr.ToAddress)
	require.Equal(t, trAmt, tr.Amount)
	require.Equal(t, tok.TokenType, tr.TokenType)
	require.Empty(t, tr.UUID, "fungible transfer must not have uuid")

	// Logs 
	if len(trOut.Logs) > 0 {
		tests.RequireLogsBase(t, trOut.Logs, len(trOut.Logs))
		_ = tests.FindLogByType(t, trOut.Logs, domain.TOKEN_TRANSFERRED_LOG)
	}

	// -------------------------
	// Fee tiers (envelope + unmarshal + validate + logs)
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
	require.NoError(t, err, "UpdateFeeTiers")

	tests.RequireStateObjectsNotNil(t, feeTiersOut.States, 1)

	var feeTiers tokenV1Domain.FeeTiers
	tests.UnmarshalState(t, feeTiersOut.States[0].Object, &feeTiers)
	require.NotEmpty(t, feeTiers.FeeTiersList)

	if len(feeTiersOut.Logs) > 0 {
		tests.RequireLogsBase(t, feeTiersOut.Logs, len(feeTiersOut.Logs))
		_ = tests.FindLogByType(t, feeTiersOut.Logs, domain.TOKEN_FEE_UPDATED_LOG)
	}

	// -------------------------
	// Fee address (envelope + unmarshal + validate + logs)
	// -------------------------
	feeOut, err := c.UpdateFeeAddress(tok.Address, owner.PublicKey)
	require.NoError(t, err, "UpdateFeeAddress")

	tests.RequireStateObjectsNotNil(t, feeOut.States, 1)

	var fee tokenV1Domain.Fee
	tests.UnmarshalState(t, feeOut.States[0].Object, &fee)
	require.Equal(t, tok.Address, fee.TokenAddress)
	require.Equal(t, owner.PublicKey, fee.FeeAddress)

	if len(feeOut.Logs) > 0 {
		tests.RequireLogsBase(t, feeOut.Logs, len(feeOut.Logs))
		_ = tests.FindLogByType(t, feeOut.Logs, domain.TOKEN_FEE_ADDRESS_UPDATED_LOG)
	}

	// -------------------------
	// Metadata (envelope + unmarshal + validate + logs)
	// -------------------------
	newSymbol := "2F-NEW" + randSuffix(4)
	newName := "2Finance New"
	expAt := time.Now().Add(30 * 24 * time.Hour)

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
		expAt,
	)
	require.NoError(t, err, "UpdateMetadata")

	tests.RequireStateObjectsNotNil(t, metaOut.States, 1)

	var meta tokenV1Domain.Token
	tests.UnmarshalState(t, metaOut.States[0].Object, &meta)
	require.Equal(t, tok.Address, meta.Address)
	require.Equal(t, newSymbol, meta.Symbol)
	require.Equal(t, newName, meta.Name)
	require.Equal(t, dec, meta.Decimals)

	if len(metaOut.Logs) > 0 {
		tests.RequireLogsBase(t, metaOut.Logs, len(metaOut.Logs))
		_ = tests.FindLogByType(t, metaOut.Logs, domain.TOKEN_METADATA_UPDATED_LOG)
	}

	// -------------------------
	// Revoke Mint Authority (envelope + unmarshal + validate + logs)
	// -------------------------
	revMintOut, err := c.RevokeMintAuthority(tok.Address, true)
	require.NoError(t, err, "RevokeMintAuthority")

	tests.RequireStateObjectsNotNil(t, revMintOut.States, 1)

	var revMint tokenV1Domain.Token
	tests.UnmarshalState(t, revMintOut.States[0].Object, &revMint)
	require.Equal(t, tok.Address, revMint.Address)
	require.True(t, revMint.MintAuthorityRevoked)

	if len(revMintOut.Logs) > 0 {
		tests.RequireLogsBase(t, revMintOut.Logs, len(revMintOut.Logs))
		_ = tests.FindLogByType(t, revMintOut.Logs, domain.TOKEN_MINT_AUTHORITY_REVOKED_LOG)
	}

	// -------------------------
	// Revoke Update Authority (envelope + unmarshal + validate + logs)
	// -------------------------
	revUpdOut, err := c.RevokeUpdateAuthority(tok.Address, true)
	require.NoError(t, err, "RevokeUpdateAuthority")

	tests.RequireStateObjectsNotNil(t, revUpdOut.States, 1)

	var revUpd tokenV1Domain.Token
	tests.UnmarshalState(t, revUpdOut.States[0].Object, &revUpd)
	require.Equal(t, tok.Address, revUpd.Address)
	require.True(t, revUpd.UpdateAuthorityRevoked)

	if len(revUpdOut.Logs) > 0 {
		tests.RequireLogsBase(t, revUpdOut.Logs, len(revUpdOut.Logs))
		_ = tests.FindLogByType(t, revUpdOut.Logs, domain.TOKEN_UPDATE_AUTHORITY_REVOKED_LOG)
	}

	// -------------------------
	// Pause (envelope + unmarshal + validate + logs)
	// -------------------------
	pauseOut, err := c.PauseToken(tok.Address, true)
	require.NoError(t, err, "PauseToken")

	tests.RequireStateObjectsNotNil(t, pauseOut.States, 1)

	var pause tokenV1Domain.Token
	tests.UnmarshalState(t, pauseOut.States[0].Object, &pause)
	require.Equal(t, tok.Address, pause.Address)
	require.True(t, pause.Paused)

	if len(pauseOut.Logs) > 0 {
		tests.RequireLogsBase(t, pauseOut.Logs, len(pauseOut.Logs))
		_ = tests.FindLogByType(t, pauseOut.Logs, domain.TOKEN_PAUSED_LOG)
	}

	// -------------------------
	// Unpause (envelope + unmarshal + validate + logs)
	// -------------------------
	unpauseOut, err := c.UnpauseToken(tok.Address, false)
	require.NoError(t, err, "UnpauseToken")

	tests.RequireStateObjectsNotNil(t, unpauseOut.States, 1)

	var unpause tokenV1Domain.Token
	tests.UnmarshalState(t, unpauseOut.States[0].Object, &unpause)
	require.Equal(t, tok.Address, unpause.Address)
	require.False(t, unpause.Paused)

	if len(unpauseOut.Logs) > 0 {
		tests.RequireLogsBase(t, unpauseOut.Logs, len(unpauseOut.Logs))
		_ = tests.FindLogByType(t, unpauseOut.Logs, domain.TOKEN_UNPAUSED_LOG)
	}

	// -------------------------
	// Freeze wallet (envelope + unmarshal + validate + logs)
	// -------------------------
	freezeOut, err := c.FreezeWallet(tok.Address, owner.PublicKey)
	require.NoError(t, err, "FreezeWallet")

	tests.RequireStateObjectsNotNil(t, freezeOut.States, 1)

	var freeze tokenV1Domain.Token
	tests.UnmarshalState(t, freezeOut.States[0].Object, &freeze)
	require.Equal(t, tok.Address, freeze.Address)
	require.NotNil(t, freeze.FrozenAccounts)
	require.True(t, freeze.FrozenAccounts[owner.PublicKey])

	if len(freezeOut.Logs) > 0 {
		tests.RequireLogsBase(t, freezeOut.Logs, len(freezeOut.Logs))
		_ = tests.FindLogByType(t, freezeOut.Logs, domain.TOKEN_FREEZE_ACCOUNT_LOG)
	}

	// -------------------------
	// Unfreeze wallet (envelope + unmarshal + validate + logs)
	// -------------------------
	unfreezeOut, err := c.UnfreezeWallet(tok.Address, owner.PublicKey)
	require.NoError(t, err, "UnfreezeWallet")

	tests.RequireStateObjectsNotNil(t, unfreezeOut.States, 1)

	var unfreeze tokenV1Domain.Token
	tests.UnmarshalState(t, unfreezeOut.States[0].Object, &unfreeze)
	require.Equal(t, tok.Address, unfreeze.Address)
	require.NotNil(t, unfreeze.FrozenAccounts)
	require.False(t, unfreeze.FrozenAccounts[owner.PublicKey])

	if len(unfreezeOut.Logs) > 0 {
		tests.RequireLogsBase(t, unfreezeOut.Logs, len(unfreezeOut.Logs))
		_ = tests.FindLogByType(t, unfreezeOut.Logs, domain.TOKEN_UNFREEZE_ACCOUNT_LOG)
	}

	// -------------------------
	// GetTokenBalance (envelope + unmarshal + validate + logs mínimos)
	// -------------------------
	getBalOut, err := c.GetTokenBalance(tok.Address, owner.PublicKey)
	require.NoError(t, err, "GetTokenBalance(owner)")

	tests.RequireStateObjectsNotNil(t, getBalOut.States, 1)

	var gotBal tokenV1Domain.Balance
	tests.UnmarshalState(t, getBalOut.States[0].Object, &gotBal)
	require.Equal(t, tok.Address, gotBal.TokenAddress)
	require.Equal(t, owner.PublicKey, gotBal.OwnerAddress)
	require.NotEmpty(t, gotBal.Amount)

	if len(getBalOut.Logs) > 0 {
		tests.RequireLogsBase(t, getBalOut.Logs, len(getBalOut.Logs))
	}

	// -------------------------
	// ListTokenBalances (envelope + unmarshal + validate + logs mínimos)
	// -------------------------
	listBalOut, err := c.ListTokenBalances(tok.Address, "", 1, 10, true)
	require.NoError(t, err, "ListTokenBalances")

	tests.RequireStateObjectsNotNil(t, listBalOut.States, 1)

	var balList []tokenV1Domain.Balance
	tests.UnmarshalState(t, listBalOut.States[0].Object, &balList)
	require.NotEmpty(t, balList)
	require.NotEmpty(t, balList[0].TokenAddress)

	if len(listBalOut.Logs) > 0 {
		tests.RequireLogsBase(t, listBalOut.Logs, len(listBalOut.Logs))
	}

	// -------------------------
	// GetToken (envelope + unmarshal + validate + logs mínimos)
	// -------------------------
	getTokOut, err := c.GetToken(tok.Address, "", "")
	require.NoError(t, err, "GetToken")

	tests.RequireStateObjectsNotNil(t, getTokOut.States, 1)

	var got tokenV1Domain.Token
	tests.UnmarshalState(t, getTokOut.States[0].Object, &got)
	require.Equal(t, tok.Address, got.Address)
	require.NotEmpty(t, got.Symbol)
	require.NotEmpty(t, got.Name)
	require.Equal(t, tok.TokenType, got.TokenType)

	if len(getTokOut.Logs) > 0 {
		tests.RequireLogsBase(t, getTokOut.Logs, len(getTokOut.Logs))
	}

	// -------------------------
	// ListTokens (envelope + unmarshal + validate + logs mínimos)
	// -------------------------
	listTokOut, err := c.ListTokens("", "", "", 1, 10, true)
	require.NoError(t, err, "ListTokens")

	tests.RequireStateObjectsNotNil(t, listTokOut.States, 1)

	var tokList []tokenV1Domain.Token
	tests.UnmarshalState(t, listTokOut.States[0].Object, &tokList)
	require.NotEmpty(t, tokList)
	require.NotEmpty(t, tokList[0].Address)

	if len(listTokOut.Logs) > 0 {
		tests.RequireLogsBase(t, listTokOut.Logs, len(listTokOut.Logs))
	}
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
	// Mint NFT (envelope + unmarshal + validate + logs + events)
	// -------------------------
	mintAmt := amt(35, dec)

	mintOut, err := c.MintToken(tok.Address, owner.PublicKey, mintAmt, dec, tok.TokenType)
	require.NoError(t, err, "MintToken NFT")

	// States (Mint, Supply, Balance)
	tests.RequireStateObjectsNotNil(t, mintOut.States, 3)

	var mint tokenV1Domain.Mint
	tests.UnmarshalState(t, mintOut.States[0].Object, &mint)
	require.Equal(t, tok.Address, mint.TokenAddress)
	require.Equal(t, owner.PublicKey, mint.MintTo)
	require.Equal(t, mintAmt, mint.Amount)
	require.Equal(t, tok.TokenType, mint.TokenType)

	require.Len(t, mint.TokenUUIDList, 35, "expected 35 uuids")
	for i, u := range mint.TokenUUIDList {
		require.NotEmpty(t, u, "mint uuid[%d] empty", i)
	}

	var supplyAfterMint tokenV1Domain.Supply
	tests.UnmarshalState(t, mintOut.States[1].Object, &supplyAfterMint)
	require.Equal(t, tok.Address, supplyAfterMint.TokenAddress)
	require.NotEmpty(t, supplyAfterMint.Amount)

	var balAfterMint tokenV1Domain.Balance
	tests.UnmarshalState(t, mintOut.States[2].Object, &balAfterMint)
	require.Equal(t, tok.Address, balAfterMint.TokenAddress)
	require.Equal(t, owner.PublicKey, balAfterMint.OwnerAddress)
	require.Equal(t, tok.TokenType, balAfterMint.TokenType)
	require.NotEmpty(t, balAfterMint.TokenUUIDList, "balance uuid list empty after mint")

	// Logs 
	tests.RequireLogsBase(t, mintOut.Logs, 3)
	tests.RequireLogTypesInOrder(t, mintOut.Logs, []string{
		domain.TOKEN_MINTED_LOG,
		domain.TOKEN_TOTAL_SUPPLY_INCREASED_LOG,
		domain.TOKEN_BALANCE_INCREASED_LOG,
	})

	// Events 
	mintEvent := tests.UnmarshalJSONB[map[string]any](t, mintOut.Logs[0].Event)
	require.Equal(t, tok.Address, tests.RequireMapFieldString(t, mintEvent, "token_address"))
	require.Equal(t, owner.PublicKey, tests.RequireMapFieldString(t, mintEvent, "mint_to"))
	require.Equal(t, mintAmt, tests.RequireMapFieldString(t, mintEvent, "amount"))
	require.Equal(t, tok.TokenType, tests.RequireMapFieldString(t, mintEvent, "token_type"))

	supplyMintEvent := tests.UnmarshalJSONB[map[string]any](t, mintOut.Logs[1].Event)
	require.Equal(t, tok.Address, tests.RequireMapFieldString(t, supplyMintEvent, "token_address"))
	require.Equal(t, mintAmt, tests.RequireMapFieldString(t, supplyMintEvent, "amount_delta"))
	require.NotEmpty(t, tests.RequireMapFieldString(t, supplyMintEvent, "total_supply"))

	balanceMintEvent := tests.UnmarshalJSONB[map[string]any](t, mintOut.Logs[2].Event)
	require.Equal(t, tok.Address, tests.RequireMapFieldString(t, balanceMintEvent, "token_address"))
	require.Equal(t, owner.PublicKey, tests.RequireMapFieldString(t, balanceMintEvent, "owner"))
	require.Equal(t, mintAmt, tests.RequireMapFieldString(t, balanceMintEvent, "amount_delta"))
	require.Equal(t, tok.TokenType, tests.RequireMapFieldString(t, balanceMintEvent, "token_type"))

	// -------------------------
	// Burn 1 NFT (envelope + unmarshal + validate + logs + events)
	// -------------------------
	burnUUID := mint.TokenUUIDList[0]
	require.NotEmpty(t, burnUUID, "burnUUID empty")

	burnAmt := amt(1, dec) // "1"
	burnOut, err := c.BurnToken(tok.Address, burnAmt, dec, tok.TokenType, burnUUID)
	require.NoError(t, err, "BurnToken NFT")

	// States (Burn, Supply, Balance)
	tests.RequireStateObjectsNotNil(t, burnOut.States, 3)

	var burn tokenV1Domain.Burn
	tests.UnmarshalState(t, burnOut.States[0].Object, &burn)
	require.Equal(t, tok.Address, burn.TokenAddress)
	require.Equal(t, owner.PublicKey, burn.BurnFrom) // signer
	require.Equal(t, burnAmt, burn.Amount)
	require.Equal(t, tok.TokenType, burn.TokenType)
	require.Equal(t, burnUUID, burn.UUID)

	var supplyAfterBurn tokenV1Domain.Supply
	tests.UnmarshalState(t, burnOut.States[1].Object, &supplyAfterBurn)
	require.Equal(t, tok.Address, supplyAfterBurn.TokenAddress)
	require.NotEmpty(t, supplyAfterBurn.Amount)

	var balAfterBurn tokenV1Domain.Balance
	tests.UnmarshalState(t, burnOut.States[2].Object, &balAfterBurn)
	require.Equal(t, tok.Address, balAfterBurn.TokenAddress)
	require.Equal(t, owner.PublicKey, balAfterBurn.OwnerAddress)
	require.Equal(t, tok.TokenType, balAfterBurn.TokenType)
	
	for _, u := range balAfterBurn.TokenUUIDList {
		require.NotEqual(t, burnUUID, u, "burned uuid must not remain in balance")
	}

	// Logs 
	tests.RequireLogsBase(t, burnOut.Logs, 3)
	tests.RequireLogTypesInOrder(t, burnOut.Logs, []string{
		domain.TOKEN_BURNED_LOG,
		domain.TOKEN_TOTAL_SUPPLY_DECREASED_LOG,
		domain.TOKEN_BALANCE_DECREASED_LOG,
	})

	// Events
	burnEvent := tests.UnmarshalJSONB[map[string]any](t, burnOut.Logs[0].Event)
	require.Equal(t, tok.Address, tests.RequireMapFieldString(t, burnEvent, "token_address"))
	require.Equal(t, owner.PublicKey, tests.RequireMapFieldString(t, burnEvent, "burn_from"))
	require.Equal(t, burnAmt, tests.RequireMapFieldString(t, burnEvent, "amount"))
	require.Equal(t, tok.TokenType, tests.RequireMapFieldString(t, burnEvent, "token_type"))
	
	if v, ok := burnEvent["uuid"]; ok {
		if s, _ := v.(string); s != "" {
			require.Equal(t, burnUUID, s)
		}
	}

	supplyBurnEvent := tests.UnmarshalJSONB[map[string]any](t, burnOut.Logs[1].Event)
	require.Equal(t, tok.Address, tests.RequireMapFieldString(t, supplyBurnEvent, "token_address"))
	if v, ok := supplyBurnEvent["amount_delta"]; ok {
		deltaStr, _ := v.(string)
		require.Contains(t, []string{"-" + burnAmt, burnAmt}, deltaStr)
	}
	if v, ok := supplyBurnEvent["amount"]; ok {
		amtStr, _ := v.(string)
		if amtStr != "" {
			require.Equal(t, burnAmt, amtStr)
		}
	}

	balanceBurnEvent := tests.UnmarshalJSONB[map[string]any](t, burnOut.Logs[2].Event)
	require.Equal(t, tok.Address, tests.RequireMapFieldString(t, balanceBurnEvent, "token_address"))
	require.Equal(t, owner.PublicKey, tests.RequireMapFieldString(t, balanceBurnEvent, "owner"))
	require.Equal(t, tok.TokenType, tests.RequireMapFieldString(t, balanceBurnEvent, "token_type"))
	
	if v, ok := balanceBurnEvent["uuid"]; ok {
		if s, _ := v.(string); s != "" {
			require.Equal(t, burnUUID, s)
		}
	}

	// -------------------------
	// AllowUsers (envelope + unmarshal + validate + logs)
	// -------------------------
	receiver, _ := createWallet(t, c)
	c.SetPrivateKey(ownerPriv)

	allowOut, err := c.AllowUsers(tok.Address, map[string]bool{receiver.PublicKey: true})
	require.NoError(t, err, "AllowUsers")

	tests.RequireStateObjectsNotNil(t, allowOut.States, 1)

	var accessPolicy tokenV1Domain.AccessPolicy
	tests.UnmarshalState(t, allowOut.States[0].Object, &accessPolicy)
	require.NotEmpty(t, accessPolicy.Mode)
	require.NotNil(t, accessPolicy.Users)
	require.True(t, accessPolicy.Users[receiver.PublicKey])

	if len(allowOut.Logs) > 0 {
		tests.RequireLogsBase(t, allowOut.Logs, len(allowOut.Logs))
		_ = tests.FindLogByType(t, allowOut.Logs, domain.TOKEN_USERS_ALLOWED_LOG)
	}

	// -------------------------
	// Transfer NFT (envelope + unmarshal + validate + logs)
	// -------------------------
	transferUUID := mint.TokenUUIDList[1]
	require.NotEmpty(t, transferUUID, "transferUUID empty")
	require.NotEqual(t, burnUUID, transferUUID, "transferUUID equals burned uuid")

	trAmt := amt(1, dec) // "1"
	trOut, err := c.TransferToken(tok.Address, receiver.PublicKey, trAmt, dec, tok.TokenType, transferUUID)
	require.NoError(t, err, "TransferToken NFT")

	tests.RequireStateObjectsNotNil(t, trOut.States, 1)

	var tr tokenV1Domain.Transfer
	tests.UnmarshalState(t, trOut.States[0].Object, &tr)
	require.Equal(t, tok.Address, tr.TokenAddress)
	require.Equal(t, owner.PublicKey, tr.FromAddress)
	require.Equal(t, receiver.PublicKey, tr.ToAddress)
	require.Equal(t, trAmt, tr.Amount)
	require.Equal(t, tok.TokenType, tr.TokenType)
	require.Equal(t, transferUUID, tr.UUID)

	if len(trOut.Logs) > 0 {
		tests.RequireLogsBase(t, trOut.Logs, len(trOut.Logs))
		_ = tests.FindLogByType(t, trOut.Logs, domain.TOKEN_TRANSFERRED_LOG)
	}

	// -------------------------
	// Fee tiers (envelope + unmarshal + validate + logs)
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
	require.NoError(t, err, "UpdateFeeTiers")

	tests.RequireStateObjectsNotNil(t, feeTiersOut.States, 1)

	var feeTiers tokenV1Domain.FeeTiers
	tests.UnmarshalState(t, feeTiersOut.States[0].Object, &feeTiers)
	require.NotEmpty(t, feeTiers.FeeTiersList)

	if len(feeTiersOut.Logs) > 0 {
		tests.RequireLogsBase(t, feeTiersOut.Logs, len(feeTiersOut.Logs))
		_ = tests.FindLogByType(t, feeTiersOut.Logs, domain.TOKEN_FEE_UPDATED_LOG)
	}

	// -------------------------
	// Fee address (envelope + unmarshal + validate + logs)
	// -------------------------
	feeOut, err := c.UpdateFeeAddress(tok.Address, owner.PublicKey)
	require.NoError(t, err, "UpdateFeeAddress")

	tests.RequireStateObjectsNotNil(t, feeOut.States, 1)

	var fee tokenV1Domain.Fee
	tests.UnmarshalState(t, feeOut.States[0].Object, &fee)
	require.Equal(t, tok.Address, fee.TokenAddress)
	require.Equal(t, owner.PublicKey, fee.FeeAddress)

	if len(feeOut.Logs) > 0 {
		tests.RequireLogsBase(t, feeOut.Logs, len(feeOut.Logs))
		_ = tests.FindLogByType(t, feeOut.Logs, domain.TOKEN_FEE_ADDRESS_UPDATED_LOG)
	}

	// -------------------------
	// Metadata (envelope + unmarshal + validate + logs)
	// -------------------------
	newSymbol := "2F-NEW" + randSuffix(4)
	newName := "2Finance New"
	expAt := time.Now().Add(30 * 24 * time.Hour)

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
		expAt,
	)
	require.NoError(t, err, "UpdateMetadata")

	tests.RequireStateObjectsNotNil(t, metaOut.States, 1)

	var meta tokenV1Domain.Token
	tests.UnmarshalState(t, metaOut.States[0].Object, &meta)
	require.Equal(t, tok.Address, meta.Address)
	require.Equal(t, newSymbol, meta.Symbol)
	require.Equal(t, newName, meta.Name)
	require.Equal(t, dec, meta.Decimals)

	if len(metaOut.Logs) > 0 {
		tests.RequireLogsBase(t, metaOut.Logs, len(metaOut.Logs))
		_ = tests.FindLogByType(t, metaOut.Logs, domain.TOKEN_METADATA_UPDATED_LOG)
	}

	// -------------------------
	// Revoke Mint Authority (envelope + unmarshal + validate + logs)
	// -------------------------
	revMintOut, err := c.RevokeMintAuthority(tok.Address, true)
	require.NoError(t, err, "RevokeMintAuthority")

	tests.RequireStateObjectsNotNil(t, revMintOut.States, 1)

	var revMint tokenV1Domain.Token
	tests.UnmarshalState(t, revMintOut.States[0].Object, &revMint)
	require.Equal(t, tok.Address, revMint.Address)
	require.True(t, revMint.MintAuthorityRevoked)

	if len(revMintOut.Logs) > 0 {
		tests.RequireLogsBase(t, revMintOut.Logs, len(revMintOut.Logs))
		_ = tests.FindLogByType(t, revMintOut.Logs, domain.TOKEN_MINT_AUTHORITY_REVOKED_LOG)
	}

	// -------------------------
	// Revoke Update Authority (envelope + unmarshal + validate + logs)
	// -------------------------
	revUpdOut, err := c.RevokeUpdateAuthority(tok.Address, true)
	require.NoError(t, err, "RevokeUpdateAuthority")

	tests.RequireStateObjectsNotNil(t, revUpdOut.States, 1)

	var revUpd tokenV1Domain.Token
	tests.UnmarshalState(t, revUpdOut.States[0].Object, &revUpd)
	require.Equal(t, tok.Address, revUpd.Address)
	require.True(t, revUpd.UpdateAuthorityRevoked)

	if len(revUpdOut.Logs) > 0 {
		tests.RequireLogsBase(t, revUpdOut.Logs, len(revUpdOut.Logs))
		_ = tests.FindLogByType(t, revUpdOut.Logs, domain.TOKEN_UPDATE_AUTHORITY_REVOKED_LOG)
	}

	// -------------------------
	// Pause (envelope + unmarshal + validate + logs)
	// -------------------------
	pauseOut, err := c.PauseToken(tok.Address, true)
	require.NoError(t, err, "PauseToken")

	tests.RequireStateObjectsNotNil(t, pauseOut.States, 1)

	var pause tokenV1Domain.Token
	tests.UnmarshalState(t, pauseOut.States[0].Object, &pause)
	require.Equal(t, tok.Address, pause.Address)
	require.True(t, pause.Paused)

	if len(pauseOut.Logs) > 0 {
		tests.RequireLogsBase(t, pauseOut.Logs, len(pauseOut.Logs))
		_ = tests.FindLogByType(t, pauseOut.Logs, domain.TOKEN_PAUSED_LOG)
	}

	// -------------------------
	// Unpause (envelope + unmarshal + validate + logs)
	// -------------------------
	unpauseOut, err := c.UnpauseToken(tok.Address, false)
	require.NoError(t, err, "UnpauseToken")

	tests.RequireStateObjectsNotNil(t, unpauseOut.States, 1)

	var unpause tokenV1Domain.Token
	tests.UnmarshalState(t, unpauseOut.States[0].Object, &unpause)
	require.Equal(t, tok.Address, unpause.Address)
	require.False(t, unpause.Paused)

	if len(unpauseOut.Logs) > 0 {
		tests.RequireLogsBase(t, unpauseOut.Logs, len(unpauseOut.Logs))
		_ = tests.FindLogByType(t, unpauseOut.Logs, domain.TOKEN_UNPAUSED_LOG)
	}

	// -------------------------
	// Freeze wallet (envelope + unmarshal + validate + logs)
	// -------------------------
	freezeOut, err := c.FreezeWallet(tok.Address, owner.PublicKey)
	require.NoError(t, err, "FreezeWallet")

	tests.RequireStateObjectsNotNil(t, freezeOut.States, 1)

	var freeze tokenV1Domain.Token
	tests.UnmarshalState(t, freezeOut.States[0].Object, &freeze)
	require.Equal(t, tok.Address, freeze.Address)
	require.NotNil(t, freeze.FrozenAccounts)
	require.True(t, freeze.FrozenAccounts[owner.PublicKey])

	if len(freezeOut.Logs) > 0 {
		tests.RequireLogsBase(t, freezeOut.Logs, len(freezeOut.Logs))
		_ = tests.FindLogByType(t, freezeOut.Logs, domain.TOKEN_FREEZE_ACCOUNT_LOG)
	}

	// -------------------------
	// Unfreeze wallet (envelope + unmarshal + validate + logs)
	// -------------------------
	unfreezeOut, err := c.UnfreezeWallet(tok.Address, owner.PublicKey)
	require.NoError(t, err, "UnfreezeWallet")

	tests.RequireStateObjectsNotNil(t, unfreezeOut.States, 1)

	var unfreeze tokenV1Domain.Token
	tests.UnmarshalState(t, unfreezeOut.States[0].Object, &unfreeze)
	require.Equal(t, tok.Address, unfreeze.Address)
	require.NotNil(t, unfreeze.FrozenAccounts)
	require.False(t, unfreeze.FrozenAccounts[owner.PublicKey])

	if len(unfreezeOut.Logs) > 0 {
		tests.RequireLogsBase(t, unfreezeOut.Logs, len(unfreezeOut.Logs))
		_ = tests.FindLogByType(t, unfreezeOut.Logs, domain.TOKEN_UNFREEZE_ACCOUNT_LOG)
	}

	// -------------------------
	// GetTokenBalance (envelope + unmarshal + validate + logs mínimos)
	// -------------------------
	getBalOut, err := c.GetTokenBalance(tok.Address, owner.PublicKey)
	require.NoError(t, err, "GetTokenBalance(owner)")

	tests.RequireStateObjectsNotNil(t, getBalOut.States, 1)

	var gotBal tokenV1Domain.Balance
	tests.UnmarshalState(t, getBalOut.States[0].Object, &gotBal)
	require.Equal(t, tok.Address, gotBal.TokenAddress)
	require.Equal(t, owner.PublicKey, gotBal.OwnerAddress)

	if len(getBalOut.Logs) > 0 {
		tests.RequireLogsBase(t, getBalOut.Logs, len(getBalOut.Logs))
	}

	// -------------------------
	// ListTokenBalances (envelope + unmarshal + validate + logs mínimos)
	// -------------------------
	listBalOut, err := c.ListTokenBalances(tok.Address, "", 1, 10, true)
	require.NoError(t, err, "ListTokenBalances")

	tests.RequireStateObjectsNotNil(t, listBalOut.States, 1)

	var balList []tokenV1Domain.Balance
	tests.UnmarshalState(t, listBalOut.States[0].Object, &balList)
	require.NotEmpty(t, balList)
	require.NotEmpty(t, balList[0].TokenAddress)

	if len(listBalOut.Logs) > 0 {
		tests.RequireLogsBase(t, listBalOut.Logs, len(listBalOut.Logs))
	}

	// -------------------------
	// GetToken (envelope + unmarshal + validate + logs mínimos)
	// -------------------------
	getTokOut, err := c.GetToken(tok.Address, "", "")
	require.NoError(t, err, "GetToken")

	tests.RequireStateObjectsNotNil(t, getTokOut.States, 1)

	var got tokenV1Domain.Token
	tests.UnmarshalState(t, getTokOut.States[0].Object, &got)
	require.Equal(t, tok.Address, got.Address)
	require.NotEmpty(t, got.Symbol)
	require.NotEmpty(t, got.Name)
	require.Equal(t, tok.TokenType, got.TokenType)

	if len(getTokOut.Logs) > 0 {
		tests.RequireLogsBase(t, getTokOut.Logs, len(getTokOut.Logs))
	}

	// -------------------------
	// ListTokens (envelope + unmarshal + validate + logs mínimos)
	// -------------------------
	listTokOut, err := c.ListTokens("", "", "", 1, 10, true)
	require.NoError(t, err, "ListTokens")

	tests.RequireStateObjectsNotNil(t, listTokOut.States, 1)

	var tokList []tokenV1Domain.Token
	tests.UnmarshalState(t, listTokOut.States[0].Object, &tokList)
	require.NotEmpty(t, tokList)
	require.NotEmpty(t, tokList[0].Address)

	if len(listTokOut.Logs) > 0 {
		tests.RequireLogsBase(t, listTokOut.Logs, len(listTokOut.Logs))
	}
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

	require.NoError(t, err, "AddToken")

	require.GreaterOrEqual(t, len(out.States), 1, "AddToken must return at least 1 state")
	require.NotNil(t, out.States[0].Object, "AddToken states[0].Object nil")

	if len(out.Logs) > 0 {
		tests.RequireLogsBase(t, out.Logs, len(out.Logs))
		ev := tests.UnmarshalJSONB[map[string]any](t, out.Logs[0].Event)
		_ = ev
	}

	var tok tokenV1Domain.Token
	unmarshalState(t, out.States[0].Object, &tok)

	// validações fortes do token criado
	require.NotEmpty(t, tok.Address, "token address empty")
	require.Equal(t, ownerPub, tok.Owner, "token owner mismatch")
	require.Equal(t, decimals, tok.Decimals, "token decimals mismatch")
	require.Equal(t, tokenType, tok.TokenType, "token type mismatch")
	require.Equal(t, stablecoin, tok.Stablecoin, "token stablecoin mismatch")

	require.NotEmpty(t, tok.Symbol)
	require.NotEmpty(t, tok.Name)
	require.Equal(t, symbol, tok.Symbol)
	require.Equal(t, name, tok.Name)

	if tokenType == tokenV1Domain.NON_FUNGIBLE {
		require.Equal(t, "1", tok.TotalSupply)
	} else {
		require.Equal(t, amt(1_000_000, decimals), tok.TotalSupply)
	}

	if tok.AccessPolicy.Mode != "" {
		require.Equal(t, accessPolicy.Mode, tok.AccessPolicy.Mode)
		require.NotNil(t, tok.AccessPolicy.Users)
		require.True(t, tok.AccessPolicy.Users[ownerPub], "access policy must include owner")
	}

	if requireFee {
		require.NotEmpty(t, tok.FeeTiersList, "fee tiers list empty")
		require.Equal(t, feeAddress, tok.FeeAddress, "fee address mismatch")
	}

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
