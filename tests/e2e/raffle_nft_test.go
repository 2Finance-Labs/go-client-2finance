package e2e_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/2finance/2finance-network/blockchain/contract/raffleV1"
	raffleV1Domain "gitlab.com/2finance/2finance-network/blockchain/contract/raffleV1/domain"
	raffleV1Models "gitlab.com/2finance/2finance-network/blockchain/contract/raffleV1/models"
	"gitlab.com/2finance/2finance-network/blockchain/contract/tokenV1"
	tokenV1Domain "gitlab.com/2finance/2finance-network/blockchain/contract/tokenV1/domain"
	tokenV1Models "gitlab.com/2finance/2finance-network/blockchain/contract/tokenV1/models"
	"gitlab.com/2finance/2finance-network/blockchain/encryption/seed"
	"gitlab.com/2finance/2finance-network/blockchain/log"
	"gitlab.com/2finance/2finance-network/blockchain/utils"
)

func TestRaffleFlowNonFungible(t *testing.T) {
	c := setupClient(t)

	// ------------------
	//      WALLETS
	// ------------------
	owner, ownerPriv := createWallet(t, c)
	player1, player1Priv := createWallet(t, c)
	player2, player2Priv := createWallet(t, c)

	// ------------------
	//    PAY TOKEN FT
	// ------------------
	c.SetPrivateKey(ownerPriv)

	payToken := createBasicToken(
		t,
		c,
		owner.PublicKey,
		6,
		false,
		tokenV1Domain.FUNGIBLE,
		false,
	)
	require.Equal(t, tokenV1Domain.FUNGIBLE, payToken.TokenType)

	// ------------------
	//   PRIZE TOKEN NFT
	// ------------------
	deployedPrizeContract, err := c.DeployContract1(tokenV1.TOKEN_CONTRACT_V1)
	require.NoError(t, err)
	require.NotEmpty(t, deployedPrizeContract.Logs)

	prizeContractLog, err := utils.UnmarshalLog[log.Log](deployedPrizeContract.Logs[0])
	require.NoError(t, err)

	prizeTokenAddress := prizeContractLog.ContractAddress
	require.NotEmpty(t, prizeTokenAddress)

	prizeSymbol := "RFNFT" + randSuffix(4)
	prizeName := "Raffle Prize NFT"
	prizeDescription := "raffle nft prize e2e"
	prizeImage := "https://example.com/raffle-prize-nft.png"
	prizeWebsite := "https://example.com"
	prizeTagsSocial := map[string]string{"twitter": "https://twitter.com/2finance"}
	prizeTagsCat := map[string]string{"category": "Collectibles"}
	prizeTags := map[string]string{"tag1": "NFT", "tag2": "Raffle"}
	prizeCreator := "2Finance Test"
	prizeCreatorWebsite := "https://creator.example"
	prizeAllowUsers := map[string]bool{}
	prizeBlockedUsers := map[string]bool{}
	prizeFrozenAccounts := map[string]bool{}
	prizeFeeTiers := []map[string]interface{}{}
	prizeFeeAddress, _ := genKey(t, c)

	prizeAddOut, err := c.AddToken(
		prizeTokenAddress,
		prizeSymbol,
		prizeName,
		0,
		"1",
		prizeDescription,
		owner.PublicKey,
		prizeImage,
		prizeWebsite,
		prizeTagsSocial,
		prizeTagsCat,
		prizeTags,
		prizeCreator,
		prizeCreatorWebsite,
		prizeAllowUsers,
		prizeBlockedUsers,
		prizeFrozenAccounts,
		prizeFeeTiers,
		prizeFeeAddress,
		false,
		false,
		false,
		false,
		time.Time{},
		"https://example.com/prize.glb",
		tokenV1Domain.NON_FUNGIBLE,
		true,
		tokenV1Domain.TOKEN_ASSET_TYPE,
	)
	require.NoError(t, err)
	require.NotEmpty(t, prizeAddOut.Logs)

	prizeAddLog, err := utils.UnmarshalLog[log.Log](prizeAddOut.Logs[0])
	require.NoError(t, err)
	assert.Equal(t, tokenV1Domain.TOKEN_CREATED_LOG, prizeAddLog.LogType)

	prizeTokenEvent, err := utils.UnmarshalEvent[tokenV1Domain.Token](prizeAddLog.Event)
	require.NoError(t, err)
	assert.Equal(t, tokenV1Domain.NON_FUNGIBLE, prizeTokenEvent.TokenType)

	// Mint real NFT(s) for owner
	mintPrizeAmount := "2"
	mintPrizeOut, err := c.MintToken(prizeTokenAddress, owner.PublicKey, mintPrizeAmount)
	require.NoError(t, err)
	require.NotEmpty(t, mintPrizeOut.Logs)

	mintPrizeLog, err := utils.UnmarshalLog[log.Log](mintPrizeOut.Logs[0])
	require.NoError(t, err)
	assert.Equal(t, tokenV1Domain.TOKEN_MINTED_NFT_LOG, mintPrizeLog.LogType)

	mintPrizeEvent, err := utils.UnmarshalEvent[tokenV1Domain.MintNFT](mintPrizeLog.Event)
	require.NoError(t, err)
	require.Len(t, mintPrizeEvent.TokenUUIDList, 2)

	firstTokenUUID := mintPrizeEvent.TokenUUIDList[0]
	secondTokenUUID := mintPrizeEvent.TokenUUIDList[1]
	require.NotEmpty(t, firstTokenUUID)
	require.NotEmpty(t, secondTokenUUID)

	ownerPrizeBeforeOut, err := c.GetTokenBalanceNFT(prizeTokenAddress, owner.PublicKey, firstTokenUUID)
	require.NoError(t, err)

	var ownerPrizeBefore tokenV1Models.BalanceStateModel
	err = utils.UnmarshalState[tokenV1Models.BalanceStateModel](ownerPrizeBeforeOut.States[0].Object, &ownerPrizeBefore)
	require.NoError(t, err)

	assert.Equal(t, prizeTokenAddress, ownerPrizeBefore.TokenAddress)
	assert.Equal(t, owner.PublicKey, ownerPrizeBefore.OwnerAddress)
	assert.Equal(t, firstTokenUUID, ownerPrizeBefore.TokenUUID)
	assert.Equal(t, "1", ownerPrizeBefore.Amount)
	assert.Equal(t, tokenV1Domain.NON_FUNGIBLE, ownerPrizeBefore.TokenType)
	assert.False(t, ownerPrizeBefore.Burned)

	// ------------------
	//    DEPLOY RAFFLE
	// ------------------
	deployedContract, err := c.DeployContract1(raffleV1.RAFFLE_CONTRACT_V1)
	require.NoError(t, err)
	require.NotEmpty(t, deployedContract.Logs)

	deployLog, err := utils.UnmarshalLog[log.Log](deployedContract.Logs[0])
	require.NoError(t, err)

	raffleAddress := deployLog.ContractAddress
	require.NotEmpty(t, raffleAddress)

	// ------------------
	//   ALLOW USERS
	// ------------------
	c.SetPrivateKey(ownerPriv)

	_, err = c.AddAllowedUsers(payToken.Address, map[string]bool{
		owner.PublicKey:   true,
		player1.PublicKey: true,
		player2.PublicKey: true,
		raffleAddress:     true,
	})
	require.NoError(t, err)

	_, err = c.AddAllowedUsers(prizeTokenAddress, map[string]bool{
		owner.PublicKey:   true,
		player1.PublicKey: true,
		player2.PublicKey: true,
		raffleAddress:     true,
	})
	require.NoError(t, err)

	_, err = c.TransferToken(payToken.Address, player1.PublicKey, "30", []string{})
	require.NoError(t, err)

	_, err = c.TransferToken(payToken.Address, player2.PublicKey, "30", []string{})
	require.NoError(t, err)

	player1PayBeforeOut, err := c.GetTokenBalance(payToken.Address, player1.PublicKey)
	require.NoError(t, err)
	var player1PayBefore tokenV1Models.BalanceStateModel
	err = utils.UnmarshalState[tokenV1Models.BalanceStateModel](player1PayBeforeOut.States[0].Object, &player1PayBefore)
	require.NoError(t, err)

	player2PayBeforeOut, err := c.GetTokenBalance(payToken.Address, player2.PublicKey)
	require.NoError(t, err)
	var player2PayBefore tokenV1Models.BalanceStateModel
	err = utils.UnmarshalState[tokenV1Models.BalanceStateModel](player2PayBeforeOut.States[0].Object, &player2PayBefore)
	require.NoError(t, err)

	ownerPayBeforeWithdrawOut, err := c.GetTokenBalance(payToken.Address, owner.PublicKey)
	require.NoError(t, err)
	var ownerPayBeforeWithdraw tokenV1Models.BalanceStateModel
	err = utils.UnmarshalState[tokenV1Models.BalanceStateModel](ownerPayBeforeWithdrawOut.States[0].Object, &ownerPayBeforeWithdraw)
	require.NoError(t, err)

	// ------------------
	//     ADD RAFFLE
	// ------------------
	revealSeed := "raffle-secret-seed-non-fungible-e2e"
	seedCommitHex := seed.CommitSeed(revealSeed)

	ticketPrice := "10"
	maxEntries := 10
	maxEntriesPerUser := 3
	startAt := time.Now().Add(-5 * time.Minute)
	expiredAt := time.Now().Add(2 * time.Hour)
	paused := false

	metadata := map[string]string{
		"name":        "Raffle Non Fungible E2E",
		"description": "raffle flow non fungible",
		"image":       "https://example.com/raffle-nft.png",
	}

	addRaffleOut, err := c.AddRaffle(
		raffleAddress,
		owner.PublicKey,
		payToken.Address,
		ticketPrice,
		maxEntries,
		maxEntriesPerUser,
		startAt,
		expiredAt,
		paused,
		seedCommitHex,
		metadata,
	)
	require.NoError(t, err)
	require.NotEmpty(t, addRaffleOut.Logs)

	addRaffleLog, err := utils.UnmarshalLog[log.Log](addRaffleOut.Logs[0])
	require.NoError(t, err)
	assert.Equal(t, raffleV1Domain.RAFFLE_ADDED_LOG, addRaffleLog.LogType)

	addRaffleEvent, err := utils.UnmarshalEvent[raffleV1Domain.Raffle](addRaffleLog.Event)
	require.NoError(t, err)

	assert.Equal(t, raffleAddress, addRaffleEvent.Address)
	assert.Equal(t, owner.PublicKey, addRaffleEvent.Owner)
	assert.Equal(t, payToken.Address, addRaffleEvent.TokenAddress)
	assert.Equal(t, ticketPrice, addRaffleEvent.TicketPrice)
	assert.Equal(t, maxEntries, addRaffleEvent.MaxEntries)
	assert.Equal(t, maxEntriesPerUser, addRaffleEvent.MaxEntriesPerUser)
	assert.Equal(t, paused, addRaffleEvent.Paused)
	assert.Equal(t, seedCommitHex, addRaffleEvent.SeedCommitHex)

	getRaffleOut, err := c.GetRaffle(raffleAddress)
	require.NoError(t, err)
	require.NotEmpty(t, getRaffleOut.States)

	var raffleState raffleV1Models.RaffleStateModel
	err = utils.UnmarshalState[raffleV1Models.RaffleStateModel](getRaffleOut.States[0].Object, &raffleState)
	require.NoError(t, err)

	assert.Equal(t, raffleAddress, raffleState.Address)
	assert.Equal(t, owner.PublicKey, raffleState.Owner)
	assert.Equal(t, payToken.Address, raffleState.TokenAddress)
	assert.Equal(t, ticketPrice, raffleState.TicketPrice)
	assert.Equal(t, maxEntries, raffleState.MaxEntries)
	assert.Equal(t, maxEntriesPerUser, raffleState.MaxEntriesPerUser)
	assert.WithinDuration(t, startAt, raffleState.StartAt, time.Second)
	assert.WithinDuration(t, expiredAt, raffleState.ExpiredAt, time.Second)

	// ------------------
	//     ADD PRIZE NFT
	// ------------------
	prizeAmount := "1"
	UUIDNFTs := []string{firstTokenUUID}
	addPrizeOut, err := c.AddRafflePrize(
		raffleAddress,
		prizeTokenAddress,
		prizeAmount,
		UUIDNFTs,
	)
	require.NoError(t, err)
	require.NotEmpty(t, addPrizeOut.Logs)

	addPrizeLog, err := utils.UnmarshalLog[log.Log](addPrizeOut.Logs[0])
	require.NoError(t, err)
	assert.Equal(t, raffleV1Domain.RAFFLE_ADDED_PRIZES_LOG, addPrizeLog.LogType)

	addPrizeEvent, err := utils.UnmarshalEvent[raffleV1Domain.RafflePrize](addPrizeLog.Event)
	require.NoError(t, err)

	assert.Equal(t, raffleAddress, addPrizeEvent.RaffleAddress)
	assert.Equal(t, owner.PublicKey, addPrizeEvent.Sponsor)
	assert.Equal(t, prizeTokenAddress, addPrizeEvent.TokenAddress)
	assert.Equal(t, prizeAmount, addPrizeEvent.Amount)
	assert.NotEmpty(t, addPrizeEvent.UUID)

	rafflePrizeUUID := addPrizeEvent.UUID

	// raffle now owns that NFT
	rafflePrizeAfterAddOut, err := c.GetTokenBalanceNFT(prizeTokenAddress, raffleAddress, firstTokenUUID)
	require.NoError(t, err)

	var rafflePrizeAfterAdd tokenV1Models.BalanceStateModel
	err = utils.UnmarshalState[tokenV1Models.BalanceStateModel](rafflePrizeAfterAddOut.States[0].Object, &rafflePrizeAfterAdd)
	require.NoError(t, err)

	assert.Equal(t, prizeTokenAddress, rafflePrizeAfterAdd.TokenAddress)
	assert.Equal(t, raffleAddress, rafflePrizeAfterAdd.OwnerAddress)
	assert.Equal(t, firstTokenUUID, rafflePrizeAfterAdd.TokenUUID)
	assert.Equal(t, "1", rafflePrizeAfterAdd.Amount)
	assert.Equal(t, tokenV1Domain.NON_FUNGIBLE, rafflePrizeAfterAdd.TokenType)
	assert.False(t, rafflePrizeAfterAdd.Burned)

	// ------------------
	//    REMOVE PRIZE
	// ------------------
	removePrizeOut, err := c.RemoveRafflePrize(raffleAddress, addPrizeEvent.UUID)
	require.NoError(t, err)
	require.NotEmpty(t, removePrizeOut.Logs)

	removePrizeLog, err := utils.UnmarshalLog[log.Log](removePrizeOut.Logs[0])
	require.NoError(t, err)
	assert.Equal(t, raffleV1Domain.RAFFLE_REMOVED_PRIZES_LOG, removePrizeLog.LogType)

	removePrizeEvent, err := utils.UnmarshalEvent[raffleV1Domain.RafflePrize](removePrizeLog.Event)
	require.NoError(t, err)

	assert.Equal(t, raffleAddress, removePrizeEvent.RaffleAddress)
	assert.Equal(t, owner.PublicKey, removePrizeEvent.Sponsor)
	assert.Equal(t, prizeTokenAddress, removePrizeEvent.TokenAddress)
	assert.Equal(t, "1", removePrizeEvent.Amount)
	assert.Equal(t, addPrizeEvent.UUID, removePrizeEvent.UUID)

	ownerPrizeAfterRemoveOut, err := c.GetTokenBalanceNFT(prizeTokenAddress, owner.PublicKey, firstTokenUUID)
	require.NoError(t, err)

	var ownerPrizeAfterRemove tokenV1Models.BalanceStateModel
	err = utils.UnmarshalState[tokenV1Models.BalanceStateModel](ownerPrizeAfterRemoveOut.States[0].Object, &ownerPrizeAfterRemove)
	require.NoError(t, err)

	assert.Equal(t, firstTokenUUID, ownerPrizeAfterRemove.TokenUUID)
	assert.Equal(t, owner.PublicKey, ownerPrizeAfterRemove.OwnerAddress)
	assert.Equal(t, "1", ownerPrizeAfterRemove.Amount)
	assert.Equal(t, tokenV1Domain.NON_FUNGIBLE, ownerPrizeAfterRemove.TokenType)

	// ------------------
	//  ADD PRIZE AGAIN
	// ------------------
	secondUUIDNFTs := []string{secondTokenUUID}
	addPrizeOut, err = c.AddRafflePrize(
		raffleAddress,
		prizeTokenAddress,
		prizeAmount,
		secondUUIDNFTs,
	)
	require.NoError(t, err)
	require.NotEmpty(t, addPrizeOut.Logs)

	addPrizeLog, err = utils.UnmarshalLog[log.Log](addPrizeOut.Logs[0])
	require.NoError(t, err)
	assert.Equal(t, raffleV1Domain.RAFFLE_ADDED_PRIZES_LOG, addPrizeLog.LogType)

	addPrizeEvent, err = utils.UnmarshalEvent[raffleV1Domain.RafflePrize](addPrizeLog.Event)
	require.NoError(t, err)

	rafflePrizeUUID = addPrizeEvent.UUID
	activePrizeTokenUUID := secondTokenUUID

	// ------------------
	//    UPDATE RAFFLE
	// ------------------
	newTicketPrice := "5"
	newMaxEntries := 20
	newMaxEntriesPerUser := 5
	newStartAt := time.Now().Add(-10 * time.Minute)
	newExpiredAt := time.Now().Add(3 * time.Hour)
	newMetadata := map[string]string{
		"name":        "Raffle Non Fungible E2E Updated",
		"description": "raffle flow non fungible updated",
		"image":       "https://example.com/raffle-nft-updated.png",
	}

	updateRaffleOut, err := c.UpdateRaffle(
		raffleAddress,
		payToken.Address,
		newTicketPrice,
		newMaxEntries,
		newMaxEntriesPerUser,
		&newStartAt,
		&newExpiredAt,
		seedCommitHex,
		newMetadata,
	)
	require.NoError(t, err)
	require.NotEmpty(t, updateRaffleOut.Logs)

	updateRaffleLog, err := utils.UnmarshalLog[log.Log](updateRaffleOut.Logs[0])
	require.NoError(t, err)
	assert.Equal(t, raffleV1Domain.RAFFLE_UPDATED_LOG, updateRaffleLog.LogType)

	getRaffleOut, err = c.GetRaffle(raffleAddress)
	require.NoError(t, err)
	err = utils.UnmarshalState[raffleV1Models.RaffleStateModel](getRaffleOut.States[0].Object, &raffleState)
	require.NoError(t, err)

	assert.Equal(t, newTicketPrice, raffleState.TicketPrice)
	assert.Equal(t, newMaxEntries, raffleState.MaxEntries)
	assert.Equal(t, newMaxEntriesPerUser, raffleState.MaxEntriesPerUser)
	assert.WithinDuration(t, newStartAt, raffleState.StartAt, time.Second)
	assert.WithinDuration(t, newExpiredAt, raffleState.ExpiredAt, time.Second)

	// ------------------
	//       PAUSE
	// ------------------
	pauseOut, err := c.PauseRaffle(raffleAddress, true)
	require.NoError(t, err)
	require.NotEmpty(t, pauseOut.Logs)

	pauseLog, err := utils.UnmarshalLog[log.Log](pauseOut.Logs[0])
	require.NoError(t, err)
	assert.Equal(t, raffleV1Domain.RAFFLE_PAUSED_LOG, pauseLog.LogType)

	getRaffleOut, err = c.GetRaffle(raffleAddress)
	require.NoError(t, err)
	err = utils.UnmarshalState[raffleV1Models.RaffleStateModel](getRaffleOut.States[0].Object, &raffleState)
	require.NoError(t, err)
	assert.True(t, raffleState.Paused)

	// ------------------
	//      UNPAUSE
	// ------------------
	unpauseOut, err := c.UnpauseRaffle(raffleAddress, false)
	require.NoError(t, err)
	require.NotEmpty(t, unpauseOut.Logs)

	unpauseLog, err := utils.UnmarshalLog[log.Log](unpauseOut.Logs[0])
	require.NoError(t, err)
	assert.Equal(t, raffleV1Domain.RAFFLE_UNPAUSED_LOG, unpauseLog.LogType)

	getRaffleOut, err = c.GetRaffle(raffleAddress)
	require.NoError(t, err)
	err = utils.UnmarshalState[raffleV1Models.RaffleStateModel](getRaffleOut.States[0].Object, &raffleState)
	require.NoError(t, err)
	assert.False(t, raffleState.Paused)

	// ------------------
	//    ENTER RAFFLE
	// ------------------
	c.SetPrivateKey(player1Priv)
	enter1Out, err := c.EnterRaffle(
		raffleAddress,
		2,
		payToken.Address,
		payToken.TokenType,
		"enter-nft-001",
	)
	require.NoError(t, err)
	require.NotEmpty(t, enter1Out.Logs)

	enter1Log, err := utils.UnmarshalLog[log.Log](enter1Out.Logs[0])
	require.NoError(t, err)
	assert.Equal(t, raffleV1Domain.RAFFLE_ENTERED_LOG, enter1Log.LogType)

	enter1Event, err := utils.UnmarshalEvent[raffleV1Domain.Entry](enter1Log.Event)
	require.NoError(t, err)
	assert.Equal(t, player1.PublicKey, enter1Event.Entrant)
	assert.Equal(t, 2, enter1Event.Tickets)
	assert.Equal(t, "10", enter1Event.Paid)

	c.SetPrivateKey(player2Priv)
	enter2Out, err := c.EnterRaffle(
		raffleAddress,
		1,
		payToken.Address,
		payToken.TokenType,
		"enter-nft-002",
	)
	require.NoError(t, err)
	require.NotEmpty(t, enter2Out.Logs)

	enter2Log, err := utils.UnmarshalLog[log.Log](enter2Out.Logs[0])
	require.NoError(t, err)
	assert.Equal(t, raffleV1Domain.RAFFLE_ENTERED_LOG, enter2Log.LogType)

	enter2Event, err := utils.UnmarshalEvent[raffleV1Domain.Entry](enter2Log.Event)
	require.NoError(t, err)
	assert.Equal(t, player2.PublicKey, enter2Event.Entrant)
	assert.Equal(t, 1, enter2Event.Tickets)
	assert.Equal(t, "5", enter2Event.Paid)

	player1PayAfterEnterOut, err := c.GetTokenBalance(payToken.Address, player1.PublicKey)
	require.NoError(t, err)
	var player1PayAfterEnter tokenV1Models.BalanceStateModel
	err = utils.UnmarshalState[tokenV1Models.BalanceStateModel](player1PayAfterEnterOut.States[0].Object, &player1PayAfterEnter)
	require.NoError(t, err)

	player2PayAfterEnterOut, err := c.GetTokenBalance(payToken.Address, player2.PublicKey)
	require.NoError(t, err)
	var player2PayAfterEnter tokenV1Models.BalanceStateModel
	err = utils.UnmarshalState[tokenV1Models.BalanceStateModel](player2PayAfterEnterOut.States[0].Object, &player2PayAfterEnter)
	require.NoError(t, err)

	rafflePayAfterEnterOut, err := c.GetTokenBalance(payToken.Address, raffleAddress)
	require.NoError(t, err)
	var rafflePayAfterEnter tokenV1Models.BalanceStateModel
	err = utils.UnmarshalState[tokenV1Models.BalanceStateModel](rafflePayAfterEnterOut.States[0].Object, &rafflePayAfterEnter)
	require.NoError(t, err)

	expectedPlayer1PayAfterEnter, err := utils.SubBigIntStrings(player1PayBefore.Amount, "10")
	require.NoError(t, err)
	expectedPlayer2PayAfterEnter, err := utils.SubBigIntStrings(player2PayBefore.Amount, "5")
	require.NoError(t, err)

	assert.Equal(t, expectedPlayer1PayAfterEnter, player1PayAfterEnter.Amount)
	assert.Equal(t, expectedPlayer2PayAfterEnter, player2PayAfterEnter.Amount)
	assert.Equal(t, "15", rafflePayAfterEnter.Amount)

	// ------------------
	//       DRAW
	// ------------------
	c.SetPrivateKey(ownerPriv)
	drawOut, err := c.DrawRaffle(raffleAddress, revealSeed)
	require.NoError(t, err)
	require.NotEmpty(t, drawOut.Logs)

	drawLog, err := utils.UnmarshalLog[log.Log](drawOut.Logs[0])
	require.NoError(t, err)
	assert.Equal(t, raffleV1Domain.RAFFLE_DRAWN_LOG, drawLog.LogType)

	drawEvent, err := utils.UnmarshalEvent[raffleV1Domain.Draw](drawLog.Event)
	require.NoError(t, err)

	assert.Equal(t, raffleAddress, drawEvent.Address)
	assert.Equal(t, revealSeed, drawEvent.RevealSeed)
	assert.Equal(t, seedCommitHex, drawEvent.SeedCommitHex)
	assert.Equal(t, 1, drawEvent.WinnerCount)
	require.Len(t, drawEvent.Winners, 1)

	winnerPrize := drawEvent.Winners[0]
	assert.Contains(t, []string{player1.PublicKey, player2.PublicKey}, winnerPrize.Winner)
	assert.NotEmpty(t, winnerPrize.PrizeUUID)

	// ------------------
	//       CLAIM NFT
	// ------------------
	_, err = c.GetTokenBalanceNFT(prizeTokenAddress, winnerPrize.Winner, winnerPrize.PrizeUUID)
	require.Error(t, err)
	require.Contains(t, err.Error(), "record not found")

	switch winnerPrize.Winner {
	case player1.PublicKey:
		c.SetPrivateKey(player1Priv)
	case player2.PublicKey:
		c.SetPrivateKey(player2Priv)
	default:
		t.Fatalf("unexpected winner: %s", winnerPrize.Winner)
	}

	claimOut, err := c.ClaimRaffle(
		raffleAddress,
		winnerPrize.PrizeUUID,
	)
	require.NoError(t, err)
	require.NotEmpty(t, claimOut.Logs)

	claimLog, err := utils.UnmarshalLog[log.Log](claimOut.Logs[0])
	require.NoError(t, err)
	assert.Equal(t, raffleV1Domain.RAFFLE_CLAIMED_LOG, claimLog.LogType)

	claimEvent, err := utils.UnmarshalEvent[raffleV1Domain.Claim](claimLog.Event)
	require.NoError(t, err)
	assert.Equal(t, raffleAddress, claimEvent.Address)
	assert.Equal(t, winnerPrize.Winner, claimEvent.Winner)
	assert.Equal(t, winnerPrize.PrizeUUID, claimEvent.PrizeUUID)

	winnerPrizeAfterOut, err := c.GetTokenBalanceNFT(prizeTokenAddress, winnerPrize.Winner, activePrizeTokenUUID)
	require.NoError(t, err)

	var winnerPrizeAfter tokenV1Models.BalanceStateModel
	err = utils.UnmarshalState[tokenV1Models.BalanceStateModel](winnerPrizeAfterOut.States[0].Object, &winnerPrizeAfter)
	require.NoError(t, err)

	assert.Equal(t, activePrizeTokenUUID, winnerPrizeAfter.TokenUUID)
	assert.Equal(t, winnerPrize.Winner, winnerPrizeAfter.OwnerAddress)
	assert.Equal(t, "1", winnerPrizeAfter.Amount)
	assert.Equal(t, tokenV1Domain.NON_FUNGIBLE, winnerPrizeAfter.TokenType)

	// ------------------
	//      WITHDRAW
	// ------------------
	c.SetPrivateKey(ownerPriv)

	withdrawAmount := "10"
	withdrawUUID := "withdraw-nft-001"

	withdrawOut, err := c.WithdrawRaffle(
		raffleAddress,
		payToken.Address,
		withdrawAmount,
		payToken.TokenType,
		withdrawUUID,
	)
	require.NoError(t, err)
	require.NotEmpty(t, withdrawOut.Logs)

	withdrawLog, err := utils.UnmarshalLog[log.Log](withdrawOut.Logs[0])
	require.NoError(t, err)
	assert.Equal(t, raffleV1Domain.RAFFLE_WITHDRAWN_LOG, withdrawLog.LogType)

	withdrawEvent, err := utils.UnmarshalEvent[raffleV1Domain.Withdrawal](withdrawLog.Event)
	require.NoError(t, err)

	assert.Equal(t, raffleAddress, withdrawEvent.Address)
	assert.Equal(t, payToken.Address, withdrawEvent.TokenAddress)
	assert.Equal(t, withdrawAmount, withdrawEvent.Amount)

	ownerPayAfterWithdrawOut, err := c.GetTokenBalance(payToken.Address, owner.PublicKey)
	require.NoError(t, err)
	var ownerPayAfterWithdraw tokenV1Models.BalanceStateModel
	err = utils.UnmarshalState[tokenV1Models.BalanceStateModel](ownerPayAfterWithdrawOut.States[0].Object, &ownerPayAfterWithdraw)
	require.NoError(t, err)

	rafflePayAfterWithdrawOut, err := c.GetTokenBalance(payToken.Address, raffleAddress)
	require.NoError(t, err)
	var rafflePayAfterWithdraw tokenV1Models.BalanceStateModel
	err = utils.UnmarshalState[tokenV1Models.BalanceStateModel](rafflePayAfterWithdrawOut.States[0].Object, &rafflePayAfterWithdraw)
	require.NoError(t, err)

	expectedOwnerPayAfterWithdraw, err := utils.AddBigIntStrings(ownerPayBeforeWithdraw.Amount, withdrawAmount)
	require.NoError(t, err)
	expectedRafflePayAfterWithdraw, err := utils.SubBigIntStrings("15", withdrawAmount)
	require.NoError(t, err)

	assert.Equal(t, expectedOwnerPayAfterWithdraw, ownerPayAfterWithdraw.Amount)
	assert.Equal(t, expectedRafflePayAfterWithdraw, rafflePayAfterWithdraw.Amount)

	// ------------------
	//     LIST PRIZES
	// ------------------
	listPrizesOut, err := c.ListPrizes(raffleAddress, 1, 10, true)
	require.NoError(t, err)
	require.NotEmpty(t, listPrizesOut.States)

	var prizes []raffleV1Models.RafflePrizeModel
	err = utils.UnmarshalState[[]raffleV1Models.RafflePrizeModel](listPrizesOut.States[0].Object, &prizes)
	require.NoError(t, err)

	require.NotEmpty(t, prizes)

	var found bool
	for _, p := range prizes {
		if p.RaffleAddress == raffleAddress && p.TokenAddress == prizeTokenAddress && p.UUID == rafflePrizeUUID {
			found = true
			assert.Equal(t, "1", p.Amount)
			assert.Equal(t, owner.PublicKey, p.Sponsor)
			assert.Equal(t, winnerPrize.Winner, p.Winner)
			assert.Equal(t, rafflePrizeUUID, p.UUID)
			// assert.True(t, p.Claimed)
			assert.NotZero(t, p.CreatedAt)
			break
		}
	}
	assert.True(t, found, "expected to find claimed raffle NFT prize in ListPrizes")
}
