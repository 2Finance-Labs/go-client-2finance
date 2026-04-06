package e2e_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/2finance/2finance-network/blockchain/contract/raffleV1"
	raffleV1Domain "gitlab.com/2finance/2finance-network/blockchain/contract/raffleV1/domain"
	raffleV1Models "gitlab.com/2finance/2finance-network/blockchain/contract/raffleV1/models"
	tokenV1Domain "gitlab.com/2finance/2finance-network/blockchain/contract/tokenV1/domain"
	tokenV1Models "gitlab.com/2finance/2finance-network/blockchain/contract/tokenV1/models"
	"gitlab.com/2finance/2finance-network/blockchain/encryption/seed"
	"gitlab.com/2finance/2finance-network/blockchain/log"
	"gitlab.com/2finance/2finance-network/blockchain/utils"
)

func TestRaffleFlowFungible(t *testing.T) {
	c := setupClient(t)

	// ------------------
	//      WALLETS
	// ------------------
	owner, ownerPriv := createWallet(t, c)
	player1, player1Priv := createWallet(t, c)
	player2, player2Priv := createWallet(t, c)

	// ------------------
	//      TOKENS
	// ------------------
	c.SetPrivateKey(ownerPriv)

	payToken := createBasicToken(
		t,
		c,
		owner.PublicKey,
		0,
		false,
		tokenV1Domain.FUNGIBLE,
		false,
	)

	prizeToken := createBasicToken(
		t,
		c,
		owner.PublicKey,
		0,
		false,
		tokenV1Domain.FUNGIBLE,
		false,
	)

	require.Equal(t, tokenV1Domain.FUNGIBLE, payToken.TokenType, "payToken must be fungible")
	require.Equal(t, tokenV1Domain.FUNGIBLE, prizeToken.TokenType, "prizeToken must be fungible")

	// ------------------
	//    DEPLOY RAFFLE
	// ------------------
	deployedContract, err := c.DeployContract1(raffleV1.RAFFLE_CONTRACT_V1)
	if err != nil {
		t.Fatalf("DeployContract: %v", err)
	}
	require.NotEmpty(t, deployedContract.Logs)

	deployLog, err := utils.UnmarshalLog[log.Log](deployedContract.Logs[0])
	if err != nil {
		t.Fatalf("UnmarshalLog (DeployContract.Logs[0]): %v", err)
	}

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
	if err != nil {
		t.Fatalf("AddAllowedUsers payToken: %v", err)
	}

	_, err = c.AddAllowedUsers(prizeToken.Address, map[string]bool{
		owner.PublicKey:   true,
		player1.PublicKey: true,
		player2.PublicKey: true,
		raffleAddress:     true,
	})
	if err != nil {
		t.Fatalf("AddAllowedUsers prizeToken: %v", err)
	}

	_, err = c.TransferToken(payToken.Address, player1.PublicKey, "30", []string{})
	if err != nil {
		t.Fatalf("TransferToken player1: %v", err)
	}

	_, err = c.TransferToken(payToken.Address, player2.PublicKey, "30", []string{})
	if err != nil {
		t.Fatalf("TransferToken player2: %v", err)
	}

	player1PayBeforeOut, err := c.GetTokenBalance(payToken.Address, player1.PublicKey)
	if err != nil {
		t.Fatalf("GetTokenBalance player1 before: %v", err)
	}
	var player1PayBefore tokenV1Models.BalanceStateModel
	err = utils.UnmarshalState[tokenV1Models.BalanceStateModel](player1PayBeforeOut.States[0].Object, &player1PayBefore)
	if err != nil {
		t.Fatalf("UnmarshalState player1PayBefore: %v", err)
	}

	player2PayBeforeOut, err := c.GetTokenBalance(payToken.Address, player2.PublicKey)
	if err != nil {
		t.Fatalf("GetTokenBalance player2 before: %v", err)
	}
	var player2PayBefore tokenV1Models.BalanceStateModel
	err = utils.UnmarshalState[tokenV1Models.BalanceStateModel](player2PayBeforeOut.States[0].Object, &player2PayBefore)
	if err != nil {
		t.Fatalf("UnmarshalState player2PayBefore: %v", err)
	}

	ownerPayBeforeWithdrawOut, err := c.GetTokenBalance(payToken.Address, owner.PublicKey)
	if err != nil {
		t.Fatalf("GetTokenBalance owner pay before withdraw: %v", err)
	}
	var ownerPayBeforeWithdraw tokenV1Models.BalanceStateModel
	err = utils.UnmarshalState[tokenV1Models.BalanceStateModel](ownerPayBeforeWithdrawOut.States[0].Object, &ownerPayBeforeWithdraw)
	if err != nil {
		t.Fatalf("UnmarshalState ownerPayBeforeWithdraw: %v", err)
	}

	ownerPrizeBeforeOut, err := c.GetTokenBalance(prizeToken.Address, owner.PublicKey)
	if err != nil {
		t.Fatalf("GetTokenBalance owner prize before: %v", err)
	}
	var ownerPrizeBefore tokenV1Models.BalanceStateModel
	err = utils.UnmarshalState[tokenV1Models.BalanceStateModel](ownerPrizeBeforeOut.States[0].Object, &ownerPrizeBefore)
	if err != nil {
		t.Fatalf("UnmarshalState ownerPrizeBefore: %v", err)
	}

	// ------------------
	//     ADD RAFFLE
	// ------------------
	revealSeed := "raffle-secret-seed-fungible-e2e"
	seedCommitHex := seed.CommitSeed(revealSeed)

	ticketPrice := "10"
	maxEntries := 10
	maxEntriesPerUser := 3
	startAt := time.Now().Add(-5 * time.Minute)
	expiredAt := time.Now().Add(2 * time.Hour)
	paused := false

	metadata := map[string]string{
		"name":        "Raffle Fungible E2E",
		"description": "raffle flow fungible",
		"image":       "https://example.com/raffle.png",
	}

	c.SetPrivateKey(ownerPriv)
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
	if err != nil {
		t.Fatalf("AddRaffle: %v", err)
	}
	require.NotEmpty(t, addRaffleOut.Logs)

	addRaffleLog, err := utils.UnmarshalLog[log.Log](addRaffleOut.Logs[0])
	if err != nil {
		t.Fatalf("UnmarshalLog (AddRaffle.Logs[0]): %v", err)
	}
	assert.Equal(t, raffleV1Domain.RAFFLE_ADDED_LOG, addRaffleLog.LogType)

	addRaffleEvent, err := utils.UnmarshalEvent[raffleV1Domain.Raffle](addRaffleLog.Event)
	if err != nil {
		t.Fatalf("UnmarshalEvent (AddRaffle.Logs[0]): %v", err)
	}

	assert.Equal(t, raffleAddress, addRaffleEvent.Address)
	assert.Equal(t, owner.PublicKey, addRaffleEvent.Owner)
	assert.Equal(t, payToken.Address, addRaffleEvent.TokenAddress)
	assert.Equal(t, ticketPrice, addRaffleEvent.TicketPrice)
	assert.Equal(t, maxEntries, addRaffleEvent.MaxEntries)
	assert.Equal(t, maxEntriesPerUser, addRaffleEvent.MaxEntriesPerUser)
	assert.Equal(t, paused, addRaffleEvent.Paused)
	assert.Equal(t, seedCommitHex, addRaffleEvent.SeedCommitHex)
	assert.Equal(t, metadata["name"], addRaffleEvent.Metadata["name"])
	assert.Equal(t, metadata["description"], addRaffleEvent.Metadata["description"])
	assert.Equal(t, metadata["image"], addRaffleEvent.Metadata["image"])

	getRaffleOut, err := c.GetRaffle(raffleAddress)
	if err != nil {
		t.Fatalf("GetRaffle: %v", err)
	}
	require.NotEmpty(t, getRaffleOut.States)

	var raffleState raffleV1Models.RaffleStateModel
	err = utils.UnmarshalState[raffleV1Models.RaffleStateModel](getRaffleOut.States[0].Object, &raffleState)
	if err != nil {
		t.Fatalf("UnmarshalState (GetRaffle.States[0]): %v", err)
	}

	assert.Equal(t, raffleAddress, raffleState.Address)
	assert.Equal(t, owner.PublicKey, raffleState.Owner)
	assert.Equal(t, payToken.Address, raffleState.TokenAddress)
	assert.Equal(t, ticketPrice, raffleState.TicketPrice)
	assert.Equal(t, maxEntries, raffleState.MaxEntries)
	assert.Equal(t, maxEntriesPerUser, raffleState.MaxEntriesPerUser)
	assert.Equal(t, paused, raffleState.Paused)
	assert.Equal(t, seedCommitHex, raffleState.SeedCommitHex)
	assert.Empty(t, raffleState.RevealSeed)
	assert.Equal(t, metadata["name"], raffleState.Metadata["name"])
	assert.Equal(t, metadata["description"], raffleState.Metadata["description"])
	assert.Equal(t, metadata["image"], raffleState.Metadata["image"])
	assert.WithinDuration(t, startAt, raffleState.StartAt, time.Second)
	assert.WithinDuration(t, expiredAt, raffleState.ExpiredAt, time.Second)
	assert.NotEmpty(t, raffleState.Hash)

	// ------------------
	//   ADD PRIZE
	// ------------------
	prizeAmount := "50"
	uuidNFTs := []string{}
	addPrizeOut, err := c.AddRafflePrize(
		raffleAddress,
		prizeToken.Address,
		prizeAmount,
		uuidNFTs,
	)
	if err != nil {
		t.Fatalf("AddRafflePrize: %v", err)
	}
	require.NotEmpty(t, addPrizeOut.Logs)

	addPrizeLog, err := utils.UnmarshalLog[log.Log](addPrizeOut.Logs[0])
	if err != nil {
		t.Fatalf("UnmarshalLog (AddRafflePrize.Logs[0]): %v", err)
	}
	assert.Equal(t, raffleV1Domain.RAFFLE_ADDED_PRIZES_LOG, addPrizeLog.LogType)

	addPrizeEvent, err := utils.UnmarshalEvent[raffleV1Domain.RafflePrize](addPrizeLog.Event)
	if err != nil {
		t.Fatalf("UnmarshalEvent (AddRafflePrize.Logs[0]): %v", err)
	}

	assert.Equal(t, raffleAddress, addPrizeEvent.RaffleAddress)
	assert.Equal(t, owner.PublicKey, addPrizeEvent.Sponsor)
	assert.Equal(t, prizeToken.Address, addPrizeEvent.TokenAddress)
	assert.Equal(t, prizeAmount, addPrizeEvent.Amount)
	assert.NotEmpty(t, addPrizeEvent.UUID)

	rafflePrizeBalanceAfterAddOut, err := c.GetTokenBalance(prizeToken.Address, raffleAddress)
	if err != nil {
		t.Fatalf("GetTokenBalance raffle prize after add: %v", err)
	}
	var rafflePrizeBalanceAfterAdd tokenV1Models.BalanceStateModel
	err = utils.UnmarshalState[tokenV1Models.BalanceStateModel](rafflePrizeBalanceAfterAddOut.States[0].Object, &rafflePrizeBalanceAfterAdd)
	if err != nil {
		t.Fatalf("UnmarshalState rafflePrizeBalanceAfterAdd: %v", err)
	}
	assert.Equal(t, prizeAmount, rafflePrizeBalanceAfterAdd.Amount)

	ownerPrizeAfterAddOut, err := c.GetTokenBalance(prizeToken.Address, owner.PublicKey)
	if err != nil {
		t.Fatalf("GetTokenBalance owner prize after add: %v", err)
	}
	var ownerPrizeAfterAdd tokenV1Models.BalanceStateModel
	err = utils.UnmarshalState[tokenV1Models.BalanceStateModel](ownerPrizeAfterAddOut.States[0].Object, &ownerPrizeAfterAdd)
	if err != nil {
		t.Fatalf("UnmarshalState ownerPrizeAfterAdd: %v", err)
	}

	expectedOwnerPrizeAfterAdd, err := utils.SubBigIntStrings(ownerPrizeBefore.Amount, prizeAmount)
	if err != nil {
		t.Fatalf("SubBigIntStrings ownerPrizeBefore - prizeAmount: %v", err)
	}
	assert.Equal(t, expectedOwnerPrizeAfterAdd, ownerPrizeAfterAdd.Amount)

	// ------------------
	//    REMOVE PRIZE
	// ------------------
	removePrizeOut, err := c.RemoveRafflePrize(raffleAddress, addPrizeEvent.UUID)
	if err != nil {
		t.Fatalf("RemoveRafflePrize: %v", err)
	}
	require.NotEmpty(t, removePrizeOut.Logs)

	removePrizeLog, err := utils.UnmarshalLog[log.Log](removePrizeOut.Logs[0])
	if err != nil {
		t.Fatalf("UnmarshalLog (RemoveRafflePrize.Logs[0]): %v", err)
	}
	assert.Equal(t, raffleV1Domain.RAFFLE_REMOVED_PRIZES_LOG, removePrizeLog.LogType)

	removePrizeEvent, err := utils.UnmarshalEvent[raffleV1Domain.RafflePrize](removePrizeLog.Event)
	if err != nil {
		t.Fatalf("UnmarshalEvent (RemoveRafflePrize.Logs[0]): %v", err)
	}

	assert.Equal(t, raffleAddress, removePrizeEvent.RaffleAddress)
	assert.Equal(t, owner.PublicKey, removePrizeEvent.Sponsor)
	assert.Equal(t, prizeToken.Address, removePrizeEvent.TokenAddress)
	assert.Equal(t, prizeAmount, removePrizeEvent.Amount)
	assert.Equal(t, addPrizeEvent.UUID, removePrizeEvent.UUID)

	rafflePrizeAfterRemoveOut, err := c.GetTokenBalance(prizeToken.Address, raffleAddress)
	if err != nil {
		require.Contains(t, err.Error(), "record not found")
	} else {
		var rafflePrizeAfterRemove tokenV1Models.BalanceStateModel
		err = utils.UnmarshalState[tokenV1Models.BalanceStateModel](rafflePrizeAfterRemoveOut.States[0].Object, &rafflePrizeAfterRemove)
		if err != nil {
			t.Fatalf("UnmarshalState rafflePrizeAfterRemove: %v", err)
		}
		assert.Equal(t, "0", rafflePrizeAfterRemove.Amount)
	}

	ownerPrizeAfterRemoveOut, err := c.GetTokenBalance(prizeToken.Address, owner.PublicKey)
	if err != nil {
		t.Fatalf("GetTokenBalance owner prize after remove: %v", err)
	}
	var ownerPrizeAfterRemove tokenV1Models.BalanceStateModel
	err = utils.UnmarshalState[tokenV1Models.BalanceStateModel](ownerPrizeAfterRemoveOut.States[0].Object, &ownerPrizeAfterRemove)
	if err != nil {
		t.Fatalf("UnmarshalState ownerPrizeAfterRemove: %v", err)
	}

	assert.Equal(t, ownerPrizeBefore.Amount, ownerPrizeAfterRemove.Amount)

	// ------------------
	//  ADD PRIZE AGAIN
	// ------------------
	addPrizeOut, err = c.AddRafflePrize(
		raffleAddress,
		prizeToken.Address,
		prizeAmount,
		uuidNFTs,
	)
	if err != nil {
		t.Fatalf("AddRafflePrize again: %v", err)
	}
	require.NotEmpty(t, addPrizeOut.Logs)

	addPrizeLog, err = utils.UnmarshalLog[log.Log](addPrizeOut.Logs[0])
	if err != nil {
		t.Fatalf("UnmarshalLog (AddRafflePrize again.Logs[0]): %v", err)
	}
	assert.Equal(t, raffleV1Domain.RAFFLE_ADDED_PRIZES_LOG, addPrizeLog.LogType)

	addPrizeEvent, err = utils.UnmarshalEvent[raffleV1Domain.RafflePrize](addPrizeLog.Event)
	if err != nil {
		t.Fatalf("UnmarshalEvent (AddRafflePrize again.Logs[0]): %v", err)
	}

	assert.Equal(t, raffleAddress, addPrizeEvent.RaffleAddress)
	assert.Equal(t, owner.PublicKey, addPrizeEvent.Sponsor)
	assert.Equal(t, prizeToken.Address, addPrizeEvent.TokenAddress)
	assert.Equal(t, prizeAmount, addPrizeEvent.Amount)
	assert.NotEmpty(t, addPrizeEvent.UUID)

	// ------------------
	//    UPDATE RAFFLE
	// ------------------
	newTicketPrice := "5"
	newMaxEntries := 20
	newMaxEntriesPerUser := 5
	newStartAt := time.Now().Add(-10 * time.Minute)
	newExpiredAt := time.Now().Add(3 * time.Hour)
	newMetadata := map[string]string{
		"name":        "Raffle Fungible E2E Updated",
		"description": "raffle flow fungible updated",
		"image":       "https://example.com/raffle-updated.png",
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
	if err != nil {
		t.Fatalf("UpdateRaffle: %v", err)
	}
	require.NotEmpty(t, updateRaffleOut.Logs)

	updateRaffleLog, err := utils.UnmarshalLog[log.Log](updateRaffleOut.Logs[0])
	if err != nil {
		t.Fatalf("UnmarshalLog (UpdateRaffle.Logs[0]): %v", err)
	}
	assert.Equal(t, raffleV1Domain.RAFFLE_UPDATED_LOG, updateRaffleLog.LogType)

	updateRaffleEvent, err := utils.UnmarshalEvent[raffleV1Domain.Raffle](updateRaffleLog.Event)
	if err != nil {
		t.Fatalf("UnmarshalEvent (UpdateRaffle.Logs[0]): %v", err)
	}

	assert.Equal(t, raffleAddress, updateRaffleEvent.Address)
	assert.Equal(t, payToken.Address, updateRaffleEvent.TokenAddress)
	assert.Equal(t, newTicketPrice, updateRaffleEvent.TicketPrice)
	assert.Equal(t, newMaxEntries, updateRaffleEvent.MaxEntries)
	assert.Equal(t, newMaxEntriesPerUser, updateRaffleEvent.MaxEntriesPerUser)
	assert.Equal(t, seedCommitHex, updateRaffleEvent.SeedCommitHex)
	assert.Equal(t, newMetadata["name"], updateRaffleEvent.Metadata["name"])
	assert.Equal(t, newMetadata["description"], updateRaffleEvent.Metadata["description"])
	assert.Equal(t, newMetadata["image"], updateRaffleEvent.Metadata["image"])

	getRaffleOut, err = c.GetRaffle(raffleAddress)
	if err != nil {
		t.Fatalf("GetRaffle after update: %v", err)
	}
	require.NotEmpty(t, getRaffleOut.States)

	err = utils.UnmarshalState[raffleV1Models.RaffleStateModel](getRaffleOut.States[0].Object, &raffleState)
	if err != nil {
		t.Fatalf("UnmarshalState (GetRaffle after update): %v", err)
	}

	assert.Equal(t, raffleAddress, raffleState.Address)
	assert.Equal(t, owner.PublicKey, raffleState.Owner)
	assert.Equal(t, payToken.Address, raffleState.TokenAddress)
	assert.Equal(t, newTicketPrice, raffleState.TicketPrice)
	assert.Equal(t, newMaxEntries, raffleState.MaxEntries)
	assert.Equal(t, newMaxEntriesPerUser, raffleState.MaxEntriesPerUser)
	assert.False(t, raffleState.Paused)
	assert.Equal(t, seedCommitHex, raffleState.SeedCommitHex)
	assert.Empty(t, raffleState.RevealSeed)
	assert.Equal(t, newMetadata["name"], raffleState.Metadata["name"])
	assert.Equal(t, newMetadata["description"], raffleState.Metadata["description"])
	assert.Equal(t, newMetadata["image"], raffleState.Metadata["image"])
	assert.WithinDuration(t, newStartAt, raffleState.StartAt, time.Second)
	assert.WithinDuration(t, newExpiredAt, raffleState.ExpiredAt, time.Second)

	// ------------------
	//       PAUSE
	// ------------------
	pauseOut, err := c.PauseRaffle(raffleAddress, true)
	if err != nil {
		t.Fatalf("PauseRaffle: %v", err)
	}
	require.NotEmpty(t, pauseOut.Logs)

	pauseLog, err := utils.UnmarshalLog[log.Log](pauseOut.Logs[0])
	if err != nil {
		t.Fatalf("UnmarshalLog (PauseRaffle.Logs[0]): %v", err)
	}
	assert.Equal(t, raffleV1Domain.RAFFLE_PAUSED_LOG, pauseLog.LogType)

	pauseEvent, err := utils.UnmarshalEvent[raffleV1Domain.Raffle](pauseLog.Event)
	if err != nil {
		t.Fatalf("UnmarshalEvent (PauseRaffle.Logs[0]): %v", err)
	}
	assert.Equal(t, raffleAddress, pauseEvent.Address)
	assert.True(t, pauseEvent.Paused)

	getRaffleOut, err = c.GetRaffle(raffleAddress)
	if err != nil {
		t.Fatalf("GetRaffle after pause: %v", err)
	}
	err = utils.UnmarshalState[raffleV1Models.RaffleStateModel](getRaffleOut.States[0].Object, &raffleState)
	if err != nil {
		t.Fatalf("UnmarshalState (GetRaffle after pause): %v", err)
	}
	assert.True(t, raffleState.Paused)

	// ------------------
	//      UNPAUSE
	// ------------------
	unpauseOut, err := c.UnpauseRaffle(raffleAddress, false)
	if err != nil {
		t.Fatalf("UnpauseRaffle: %v", err)
	}
	require.NotEmpty(t, unpauseOut.Logs)

	unpauseLog, err := utils.UnmarshalLog[log.Log](unpauseOut.Logs[0])
	if err != nil {
		t.Fatalf("UnmarshalLog (UnpauseRaffle.Logs[0]): %v", err)
	}
	assert.Equal(t, raffleV1Domain.RAFFLE_UNPAUSED_LOG, unpauseLog.LogType)

	unpauseEvent, err := utils.UnmarshalEvent[raffleV1Domain.Raffle](unpauseLog.Event)
	if err != nil {
		t.Fatalf("UnmarshalEvent (UnpauseRaffle.Logs[0]): %v", err)
	}
	assert.Equal(t, raffleAddress, unpauseEvent.Address)
	assert.False(t, unpauseEvent.Paused)

	getRaffleOut, err = c.GetRaffle(raffleAddress)
	if err != nil {
		t.Fatalf("GetRaffle after unpause: %v", err)
	}
	err = utils.UnmarshalState[raffleV1Models.RaffleStateModel](getRaffleOut.States[0].Object, &raffleState)
	if err != nil {
		t.Fatalf("UnmarshalState (GetRaffle after unpause): %v", err)
	}
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
		"enter-ft-001",
	)
	if err != nil {
		t.Fatalf("EnterRaffle player1: %v", err)
	}
	require.NotEmpty(t, enter1Out.Logs)

	enter1Log, err := utils.UnmarshalLog[log.Log](enter1Out.Logs[0])
	if err != nil {
		t.Fatalf("UnmarshalLog (EnterRaffle player1.Logs[0]): %v", err)
	}
	assert.Equal(t, raffleV1Domain.RAFFLE_ENTERED_LOG, enter1Log.LogType)

	enter1Event, err := utils.UnmarshalEvent[raffleV1Domain.Entry](enter1Log.Event)
	if err != nil {
		t.Fatalf("UnmarshalEvent (EnterRaffle player1.Logs[0]): %v", err)
	}

	assert.Equal(t, raffleAddress, enter1Event.RaffleAddress)
	assert.Equal(t, player1.PublicKey, enter1Event.Entrant)
	assert.Equal(t, 2, enter1Event.Tickets)
	assert.Equal(t, payToken.Address, enter1Event.PayTokenAddress)
	assert.Equal(t, "10", enter1Event.Paid)
	assert.NotEmpty(t, enter1Event.UUID)

	c.SetPrivateKey(player2Priv)
	enter2Out, err := c.EnterRaffle(
		raffleAddress,
		1,
		payToken.Address,
		payToken.TokenType,
		"enter-ft-002",
	)
	if err != nil {
		t.Fatalf("EnterRaffle player2: %v", err)
	}
	require.NotEmpty(t, enter2Out.Logs)

	enter2Log, err := utils.UnmarshalLog[log.Log](enter2Out.Logs[0])
	if err != nil {
		t.Fatalf("UnmarshalLog (EnterRaffle player2.Logs[0]): %v", err)
	}
	assert.Equal(t, raffleV1Domain.RAFFLE_ENTERED_LOG, enter2Log.LogType)

	enter2Event, err := utils.UnmarshalEvent[raffleV1Domain.Entry](enter2Log.Event)
	if err != nil {
		t.Fatalf("UnmarshalEvent (EnterRaffle player2.Logs[0]): %v", err)
	}

	assert.Equal(t, raffleAddress, enter2Event.RaffleAddress)
	assert.Equal(t, player2.PublicKey, enter2Event.Entrant)
	assert.Equal(t, 1, enter2Event.Tickets)
	assert.Equal(t, payToken.Address, enter2Event.PayTokenAddress)
	assert.Equal(t, "5", enter2Event.Paid)
	assert.NotEmpty(t, enter2Event.UUID)

	player1PayAfterEnterOut, err := c.GetTokenBalance(payToken.Address, player1.PublicKey)
	if err != nil {
		t.Fatalf("GetTokenBalance player1 after enter: %v", err)
	}
	var player1PayAfterEnter tokenV1Models.BalanceStateModel
	err = utils.UnmarshalState[tokenV1Models.BalanceStateModel](player1PayAfterEnterOut.States[0].Object, &player1PayAfterEnter)
	if err != nil {
		t.Fatalf("UnmarshalState player1PayAfterEnter: %v", err)
	}

	player2PayAfterEnterOut, err := c.GetTokenBalance(payToken.Address, player2.PublicKey)
	if err != nil {
		t.Fatalf("GetTokenBalance player2 after enter: %v", err)
	}
	var player2PayAfterEnter tokenV1Models.BalanceStateModel
	err = utils.UnmarshalState[tokenV1Models.BalanceStateModel](player2PayAfterEnterOut.States[0].Object, &player2PayAfterEnter)
	if err != nil {
		t.Fatalf("UnmarshalState player2PayAfterEnter: %v", err)
	}

	rafflePayAfterEnterOut, err := c.GetTokenBalance(payToken.Address, raffleAddress)
	if err != nil {
		t.Fatalf("GetTokenBalance raffle pay after enter: %v", err)
	}
	var rafflePayAfterEnter tokenV1Models.BalanceStateModel
	err = utils.UnmarshalState[tokenV1Models.BalanceStateModel](rafflePayAfterEnterOut.States[0].Object, &rafflePayAfterEnter)
	if err != nil {
		t.Fatalf("UnmarshalState rafflePayAfterEnter: %v", err)
	}

	expectedPlayer1PayAfterEnter, err := utils.SubBigIntStrings(player1PayBefore.Amount, "10")
	if err != nil {
		t.Fatalf("SubBigIntStrings player1PayBefore - 10: %v", err)
	}
	expectedPlayer2PayAfterEnter, err := utils.SubBigIntStrings(player2PayBefore.Amount, "5")
	if err != nil {
		t.Fatalf("SubBigIntStrings player2PayBefore - 5: %v", err)
	}

	assert.Equal(t, expectedPlayer1PayAfterEnter, player1PayAfterEnter.Amount)
	assert.Equal(t, expectedPlayer2PayAfterEnter, player2PayAfterEnter.Amount)
	assert.Equal(t, "15", rafflePayAfterEnter.Amount)

	// ------------------
	//       DRAW
	// ------------------
	c.SetPrivateKey(ownerPriv)
	drawOut, err := c.DrawRaffle(raffleAddress, revealSeed)
	if err != nil {
		t.Fatalf("DrawRaffle: %v", err)
	}
	require.NotEmpty(t, drawOut.Logs)

	drawLog, err := utils.UnmarshalLog[log.Log](drawOut.Logs[0])
	if err != nil {
		t.Fatalf("UnmarshalLog (DrawRaffle.Logs[0]): %v", err)
	}
	assert.Equal(t, raffleV1Domain.RAFFLE_DRAWN_LOG, drawLog.LogType)

	drawEvent, err := utils.UnmarshalEvent[raffleV1Domain.Draw](drawLog.Event)
	if err != nil {
		t.Fatalf("UnmarshalEvent (DrawRaffle.Logs[0]): %v", err)
	}

	assert.Equal(t, raffleAddress, drawEvent.Address)
	assert.Equal(t, revealSeed, drawEvent.RevealSeed)
	assert.Equal(t, seedCommitHex, drawEvent.SeedCommitHex)
	assert.Equal(t, 1, drawEvent.WinnerCount)
	require.Len(t, drawEvent.Winners, 1)

	winnerPrize := drawEvent.Winners[0]
	assert.Contains(t, []string{player1.PublicKey, player2.PublicKey}, winnerPrize.Winner)
	assert.NotEmpty(t, winnerPrize.PrizeUUID)

	getRaffleOut, err = c.GetRaffle(raffleAddress)
	if err != nil {
		t.Fatalf("GetRaffle after draw: %v", err)
	}
	err = utils.UnmarshalState[raffleV1Models.RaffleStateModel](getRaffleOut.States[0].Object, &raffleState)
	if err != nil {
		t.Fatalf("UnmarshalState (GetRaffle after draw): %v", err)
	}
	assert.Empty(t, raffleState.RevealSeed)
	assert.Equal(t, seedCommitHex, raffleState.SeedCommitHex)
	assert.Equal(t, newMetadata["description"], raffleState.Metadata["description"])

	// ------------------
	//       CLAIM
	// ------------------
	_, err = c.GetTokenBalance(prizeToken.Address, winnerPrize.Winner)
	require.Error(t, err)
	require.Contains(t, err.Error(), "record not found")

	rafflePrizeBeforeClaimOut, err := c.GetTokenBalance(prizeToken.Address, raffleAddress)
	if err != nil {
		t.Fatalf("GetTokenBalance raffle prize before claim: %v", err)
	}
	var rafflePrizeBeforeClaim tokenV1Models.BalanceStateModel
	err = utils.UnmarshalState[tokenV1Models.BalanceStateModel](rafflePrizeBeforeClaimOut.States[0].Object, &rafflePrizeBeforeClaim)
	if err != nil {
		t.Fatalf("UnmarshalState rafflePrizeBeforeClaim: %v", err)
	}

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
	if err != nil {
		t.Fatalf("ClaimRaffle: %v", err)
	}
	require.NotEmpty(t, claimOut.Logs)

	claimLog, err := utils.UnmarshalLog[log.Log](claimOut.Logs[0])
	if err != nil {
		t.Fatalf("UnmarshalLog (ClaimRaffle.Logs[0]): %v", err)
	}
	assert.Equal(t, raffleV1Domain.RAFFLE_CLAIMED_LOG, claimLog.LogType)

	claimEvent, err := utils.UnmarshalEvent[raffleV1Domain.Claim](claimLog.Event)
	if err != nil {
		t.Fatalf("UnmarshalEvent (ClaimRaffle.Logs[0]): %v", err)
	}

	assert.Equal(t, raffleAddress, claimEvent.Address)
	assert.Equal(t, winnerPrize.Winner, claimEvent.Winner)

	winnerPrizeAfterOut, err := c.GetTokenBalance(prizeToken.Address, winnerPrize.Winner)
	if err != nil {
		t.Fatalf("GetTokenBalance winner prize after claim: %v", err)
	}
	var winnerPrizeAfter tokenV1Models.BalanceStateModel
	err = utils.UnmarshalState[tokenV1Models.BalanceStateModel](winnerPrizeAfterOut.States[0].Object, &winnerPrizeAfter)
	if err != nil {
		t.Fatalf("UnmarshalState winnerPrizeAfter: %v", err)
	}

	winnerPrizeBefore := "0"
	expectedWinnerPrizeAfter, err := utils.AddBigIntStrings(winnerPrizeBefore, prizeAmount)
	if err != nil {
		t.Fatalf("AddBigIntStrings winnerPrizeBefore + prizeAmount: %v", err)
	}
	assert.Equal(t, expectedWinnerPrizeAfter, winnerPrizeAfter.Amount)

	// ------------------
	//      WITHDRAW
	// ------------------
	c.SetPrivateKey(ownerPriv)

	withdrawAmount := "10"
	withdrawUUID := "withdraw-ft-001"

	withdrawOut, err := c.WithdrawRaffle(
		raffleAddress,
		payToken.Address,
		withdrawAmount,
		payToken.TokenType,
		withdrawUUID,
	)
	if err != nil {
		t.Fatalf("WithdrawRaffle: %v", err)
	}
	require.NotEmpty(t, withdrawOut.Logs)

	withdrawLog, err := utils.UnmarshalLog[log.Log](withdrawOut.Logs[0])
	if err != nil {
		t.Fatalf("UnmarshalLog (WithdrawRaffle.Logs[0]): %v", err)
	}
	assert.Equal(t, raffleV1Domain.RAFFLE_WITHDRAWN_LOG, withdrawLog.LogType)

	withdrawEvent, err := utils.UnmarshalEvent[raffleV1Domain.Withdrawal](withdrawLog.Event)
	if err != nil {
		t.Fatalf("UnmarshalEvent (WithdrawRaffle.Logs[0]): %v", err)
	}

	assert.Equal(t, raffleAddress, withdrawEvent.Address)
	assert.Equal(t, payToken.Address, withdrawEvent.TokenAddress)
	assert.Equal(t, withdrawAmount, withdrawEvent.Amount)

	ownerPayAfterWithdrawOut, err := c.GetTokenBalance(payToken.Address, owner.PublicKey)
	if err != nil {
		t.Fatalf("GetTokenBalance owner pay after withdraw: %v", err)
	}
	var ownerPayAfterWithdraw tokenV1Models.BalanceStateModel
	err = utils.UnmarshalState[tokenV1Models.BalanceStateModel](ownerPayAfterWithdrawOut.States[0].Object, &ownerPayAfterWithdraw)
	if err != nil {
		t.Fatalf("UnmarshalState ownerPayAfterWithdraw: %v", err)
	}

	rafflePayAfterWithdrawOut, err := c.GetTokenBalance(payToken.Address, raffleAddress)
	if err != nil {
		t.Fatalf("GetTokenBalance raffle pay after withdraw: %v", err)
	}
	var rafflePayAfterWithdraw tokenV1Models.BalanceStateModel
	err = utils.UnmarshalState[tokenV1Models.BalanceStateModel](rafflePayAfterWithdrawOut.States[0].Object, &rafflePayAfterWithdraw)
	if err != nil {
		t.Fatalf("UnmarshalState rafflePayAfterWithdraw: %v", err)
	}

	expectedOwnerPayAfterWithdraw, err := utils.AddBigIntStrings(ownerPayBeforeWithdraw.Amount, withdrawAmount)
	if err != nil {
		t.Fatalf("AddBigIntStrings ownerPayBeforeWithdraw + withdrawAmount: %v", err)
	}
	expectedRafflePayAfterWithdraw, err := utils.SubBigIntStrings("15", withdrawAmount)
	if err != nil {
		t.Fatalf("SubBigIntStrings 15 - withdrawAmount: %v", err)
	}

	assert.Equal(t, expectedOwnerPayAfterWithdraw, ownerPayAfterWithdraw.Amount)
	assert.Equal(t, expectedRafflePayAfterWithdraw, rafflePayAfterWithdraw.Amount)

	// ------------------
	//     LIST PRIZES
	// ------------------
	listPrizesOut, err := c.ListPrizes(raffleAddress, 1, 10, true)
	if err != nil {
		t.Fatalf("ListPrizes: %v", err)
	}
	require.NotEmpty(t, listPrizesOut.States)

	var prizes []raffleV1Models.RafflePrizeModel
	err = utils.UnmarshalState[[]raffleV1Models.RafflePrizeModel](listPrizesOut.States[0].Object, &prizes)
	if err != nil {
		t.Fatalf("UnmarshalState (ListPrizes.States[0]): %v", err)
	}

	require.NotEmpty(t, prizes)

	var found bool
	for _, p := range prizes {
		if p.RaffleAddress == raffleAddress && p.TokenAddress == prizeToken.Address && p.Amount == prizeAmount {
			found = true
			assert.Equal(t, addPrizeEvent.UUID, p.UUID)
			assert.Equal(t, owner.PublicKey, p.Sponsor)
			assert.Equal(t, winnerPrize.Winner, p.Winner)
			assert.NotZero(t, p.CreatedAt)
			assert.NotZero(t, p.UpdatedAt)
			break
		}
	}
	assert.True(t, found, "expected to find claimed raffle prize in ListPrizes")

	// ------------------
	//     GET PRIZE
	// ------------------
	getPrizeOut, err := c.GetPrize(raffleAddress, addPrizeEvent.UUID)
	if err != nil {
		t.Fatalf("GetPrize: %v", err)
	}
	require.NotEmpty(t, getPrizeOut.States)

	var prize raffleV1Models.RafflePrizeModel
	err = utils.UnmarshalState[raffleV1Models.RafflePrizeModel](getPrizeOut.States[0].Object, &prize)
	if err != nil {
		t.Fatalf("UnmarshalState (GetPrize.States[0]): %v", err)
	}

	assert.Equal(t, addPrizeEvent.UUID, prize.UUID)
	assert.Equal(t, raffleAddress, prize.RaffleAddress)
	assert.Equal(t, owner.PublicKey, prize.Sponsor)
	assert.Equal(t, prizeToken.Address, prize.TokenAddress)
	assert.Equal(t, prizeAmount, prize.Amount)
	assert.Equal(t, winnerPrize.Winner, prize.Winner)
	assert.NotZero(t, prize.CreatedAt)
	assert.NotZero(t, prize.UpdatedAt)
}
