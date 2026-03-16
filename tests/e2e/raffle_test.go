package e2e_test

import (
	"testing"
	"time"

	"github.com/2Finance-Labs/go-client-2finance/client_2finance"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	raffleV1 "gitlab.com/2finance/2finance-network/blockchain/contract/raffleV1"
	raffleV1Domain "gitlab.com/2finance/2finance-network/blockchain/contract/raffleV1/domain"
	tokenV1Domain "gitlab.com/2finance/2finance-network/blockchain/contract/tokenV1/domain"
	tokenV1Models "gitlab.com/2finance/2finance-network/blockchain/contract/tokenV1/models"
	"gitlab.com/2finance/2finance-network/blockchain/encryption/seed"
	"gitlab.com/2finance/2finance-network/blockchain/log"
	"gitlab.com/2finance/2finance-network/blockchain/utils"
)

func getFTBalance(
	t *testing.T,
	c client_2finance.Client2FinanceNetwork,
	tokenAddress string,
	ownerAddress string,
) tokenV1Models.BalanceStateModel {
	t.Helper()

	out, err := c.GetTokenBalance(tokenAddress, ownerAddress)
	require.NoError(t, err)
	require.NotEmpty(t, out.States)

	var state tokenV1Models.BalanceStateModel
	err = utils.UnmarshalState[tokenV1Models.BalanceStateModel](out.States[0].Object, &state)
	require.NoError(t, err)

	return state
}

func getFTBalanceOrZero(
	t *testing.T,
	c client_2finance.Client2FinanceNetwork,
	tokenAddress string,
	ownerAddress string,
) tokenV1Models.BalanceStateModel {
	t.Helper()

	out, err := c.GetTokenBalance(tokenAddress, ownerAddress)
	if err != nil {
		require.Contains(t, err.Error(), "record not found")
		return tokenV1Models.BalanceStateModel{
			TokenAddress: tokenAddress,
			OwnerAddress: ownerAddress,
			Amount:       "0",
		}
	}

	if len(out.States) == 0 {
		return tokenV1Models.BalanceStateModel{
			TokenAddress: tokenAddress,
			OwnerAddress: ownerAddress,
			Amount:       "0",
		}
	}

	var state tokenV1Models.BalanceStateModel
	err = utils.UnmarshalState[tokenV1Models.BalanceStateModel](out.States[0].Object, &state)
	require.NoError(t, err)

	if state.Amount == "" {
		state.Amount = "0"
	}

	return state
}

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
		6,
		false,
		tokenV1Domain.FUNGIBLE,
		false,
	)

	prizeToken := createBasicToken(
		t,
		c,
		owner.PublicKey,
		6,
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

	_, err = c.AddAllowedUsers(prizeToken.Address, map[string]bool{
		owner.PublicKey:   true,
		player1.PublicKey: true,
		player2.PublicKey: true,
		raffleAddress:     true,
	})
	require.NoError(t, err)

	// distribui payToken para os participantes
	_, err = c.TransferToken(payToken.Address, player1.PublicKey, "3000000", []string{})
	require.NoError(t, err)

	_, err = c.TransferToken(payToken.Address, player2.PublicKey, "3000000", []string{})
	require.NoError(t, err)

	player1PayBefore := getFTBalance(t, c, payToken.Address, player1.PublicKey)
	player2PayBefore := getFTBalance(t, c, payToken.Address, player2.PublicKey)
	ownerPayBeforeWithdraw := getFTBalance(t, c, payToken.Address, owner.PublicKey)
	ownerPrizeBefore := getFTBalance(t, c, prizeToken.Address, owner.PublicKey)

	// ------------------
	//     ADD RAFFLE
	// ------------------
	revealSeed := "raffle-secret-seed-fungible-e2e"
	seedCommitHex := seed.CommitSeed(revealSeed)

	ticketPrice := "1000000"
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
	assert.Equal(t, metadata["name"], addRaffleEvent.Metadata["name"])
	assert.Equal(t, metadata["description"], addRaffleEvent.Metadata["description"])

	// ------------------
	//   ADD PRIZE
	// ------------------
	prizeAmount := "2500000"
	prizeUUID := "prize-ft-001"

	addPrizeOut, err := c.AddRafflePrize(
		raffleAddress,
		prizeToken.Address,
		prizeAmount,
		prizeToken.TokenType,
		prizeUUID,
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
	assert.Equal(t, prizeToken.Address, addPrizeEvent.TokenAddress)
	assert.Equal(t, prizeAmount, addPrizeEvent.Amount)
	assert.NotEmpty(t, addPrizeEvent.UUID)

	rafflePrizeBalanceAfterAdd := getFTBalance(t, c, prizeToken.Address, raffleAddress)
	assert.Equal(t, prizeAmount, rafflePrizeBalanceAfterAdd.Amount)

	ownerPrizeAfterAdd := getFTBalance(t, c, prizeToken.Address, owner.PublicKey)
	expectedOwnerPrizeAfterAdd, err := utils.SubBigIntStrings(ownerPrizeBefore.Amount, prizeAmount)
	require.NoError(t, err)
	assert.Equal(t, expectedOwnerPrizeAfterAdd, ownerPrizeAfterAdd.Amount)

	// ------------------
	//    UPDATE RAFFLE
	// ------------------
	newTicketPrice := "500000"
	newMaxEntries := 20
	newMaxEntriesPerUser := 5
	newStartAt := time.Now().Add(-10 * time.Minute)
	newExpiredAt := time.Now().Add(3 * time.Hour)
	newMetadata := map[string]string{
		"name":        "Raffle Fungible E2E Updated",
		"description": "raffle flow fungible updated",
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

	updateRaffleEvent, err := utils.UnmarshalEvent[raffleV1Domain.Raffle](updateRaffleLog.Event)
	require.NoError(t, err)

	assert.Equal(t, raffleAddress, updateRaffleEvent.Address)
	assert.Equal(t, payToken.Address, updateRaffleEvent.TokenAddress)
	assert.Equal(t, newTicketPrice, updateRaffleEvent.TicketPrice)
	assert.Equal(t, newMaxEntries, updateRaffleEvent.MaxEntries)
	assert.Equal(t, newMaxEntriesPerUser, updateRaffleEvent.MaxEntriesPerUser)
	assert.Equal(t, seedCommitHex, updateRaffleEvent.SeedCommitHex)
	assert.Equal(t, newMetadata["name"], updateRaffleEvent.Metadata["name"])
	assert.Equal(t, newMetadata["description"], updateRaffleEvent.Metadata["description"])

	// ------------------
	//       PAUSE
	// ------------------
	pauseOut, err := c.PauseRaffle(raffleAddress, true)
	require.NoError(t, err)
	require.NotEmpty(t, pauseOut.Logs)

	pauseLog, err := utils.UnmarshalLog[log.Log](pauseOut.Logs[0])
	require.NoError(t, err)
	assert.Equal(t, raffleV1Domain.RAFFLE_PAUSED_LOG, pauseLog.LogType)

	pauseEvent, err := utils.UnmarshalEvent[raffleV1Domain.Raffle](pauseLog.Event)
	require.NoError(t, err)
	assert.Equal(t, raffleAddress, pauseEvent.Address)
	assert.Equal(t, true, pauseEvent.Paused)

	// ------------------
	//      UNPAUSE
	// ------------------
	unpauseOut, err := c.UnpauseRaffle(raffleAddress, false)
	require.NoError(t, err)
	require.NotEmpty(t, unpauseOut.Logs)

	unpauseLog, err := utils.UnmarshalLog[log.Log](unpauseOut.Logs[0])
	require.NoError(t, err)
	assert.Equal(t, raffleV1Domain.RAFFLE_UNPAUSED_LOG, unpauseLog.LogType)

	unpauseEvent, err := utils.UnmarshalEvent[raffleV1Domain.Raffle](unpauseLog.Event)
	require.NoError(t, err)
	assert.Equal(t, raffleAddress, unpauseEvent.Address)
	assert.Equal(t, false, unpauseEvent.Paused)

	// ------------------
	//    ENTER RAFFLE
	// ------------------
	c.SetPrivateKey(player1Priv)
	enter1Out, err := c.EnterRaffle(
		raffleAddress,
		2,
		payToken.Address,
		payToken.TokenType,
		"",
	)
	require.NoError(t, err)
	require.NotEmpty(t, enter1Out.Logs)

	enter1Log, err := utils.UnmarshalLog[log.Log](enter1Out.Logs[0])
	require.NoError(t, err)
	assert.Equal(t, raffleV1Domain.RAFFLE_ENTERED_LOG, enter1Log.LogType)

	enter1Event, err := utils.UnmarshalEvent[raffleV1Domain.Entry](enter1Log.Event)
	require.NoError(t, err)

	assert.Equal(t, raffleAddress, enter1Event.RaffleAddress)
	assert.Equal(t, player1.PublicKey, enter1Event.Entrant)
	assert.Equal(t, 2, enter1Event.Tickets)
	assert.Equal(t, payToken.Address, enter1Event.PayTokenAddress)
	assert.Equal(t, "1000000", enter1Event.Paid)
	assert.NotEmpty(t, enter1Event.UUID)

	c.SetPrivateKey(player2Priv)
	enter2Out, err := c.EnterRaffle(
		raffleAddress,
		1,
		payToken.Address,
		payToken.TokenType,
		"",
	)
	require.NoError(t, err)
	require.NotEmpty(t, enter2Out.Logs)

	enter2Log, err := utils.UnmarshalLog[log.Log](enter2Out.Logs[0])
	require.NoError(t, err)
	assert.Equal(t, raffleV1Domain.RAFFLE_ENTERED_LOG, enter2Log.LogType)

	enter2Event, err := utils.UnmarshalEvent[raffleV1Domain.Entry](enter2Log.Event)
	require.NoError(t, err)

	assert.Equal(t, raffleAddress, enter2Event.RaffleAddress)
	assert.Equal(t, player2.PublicKey, enter2Event.Entrant)
	assert.Equal(t, 1, enter2Event.Tickets)
	assert.Equal(t, payToken.Address, enter2Event.PayTokenAddress)
	assert.Equal(t, "500000", enter2Event.Paid)
	assert.NotEmpty(t, enter2Event.UUID)

	player1PayAfterEnter := getFTBalance(t, c, payToken.Address, player1.PublicKey)
	player2PayAfterEnter := getFTBalance(t, c, payToken.Address, player2.PublicKey)
	rafflePayAfterEnter := getFTBalance(t, c, payToken.Address, raffleAddress)

	expectedPlayer1PayAfterEnter, err := utils.SubBigIntStrings(player1PayBefore.Amount, "1000000")
	require.NoError(t, err)
	expectedPlayer2PayAfterEnter, err := utils.SubBigIntStrings(player2PayBefore.Amount, "500000")
	require.NoError(t, err)

	assert.Equal(t, expectedPlayer1PayAfterEnter, player1PayAfterEnter.Amount)
	assert.Equal(t, expectedPlayer2PayAfterEnter, player2PayAfterEnter.Amount)
	assert.Equal(t, "1500000", rafflePayAfterEnter.Amount)

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

	winner := drawEvent.Winners[0]
	assert.Contains(t, []string{player1.PublicKey, player2.PublicKey}, winner)

	// ------------------
	//       CLAIM
	// ------------------
	winnerPrizeBefore := getFTBalanceOrZero(t, c, prizeToken.Address, winner)
	rafflePrizeBeforeClaim := getFTBalance(t, c, prizeToken.Address, raffleAddress)

	switch winner {
	case player1.PublicKey:
		c.SetPrivateKey(player1Priv)
	case player2.PublicKey:
		c.SetPrivateKey(player2Priv)
	default:
		t.Fatalf("unexpected winner: %s", winner)
	}

	claimOut, err := c.ClaimRaffle(
		raffleAddress,
		winner,
		prizeToken.TokenType,
		prizeUUID,
	)
	require.NoError(t, err)
	require.NotEmpty(t, claimOut.Logs)

	claimLog, err := utils.UnmarshalLog[log.Log](claimOut.Logs[0])
	require.NoError(t, err)
	assert.Equal(t, raffleV1Domain.RAFFLE_CLAIMED_LOG, claimLog.LogType)

	claimEvent, err := utils.UnmarshalEvent[raffleV1Domain.Claim](claimLog.Event)
	require.NoError(t, err)

	assert.Equal(t, raffleAddress, claimEvent.Address)
	assert.Equal(t, winner, claimEvent.Winner)

	winnerPrizeAfter := getFTBalance(t, c, prizeToken.Address, winner)
	rafflePrizeAfterClaim := getFTBalance(t, c, prizeToken.Address, raffleAddress)

	expectedWinnerPrizeAfter, err := utils.AddBigIntStrings(winnerPrizeBefore.Amount, prizeAmount)
	require.NoError(t, err)
	expectedRafflePrizeAfterClaim, err := utils.SubBigIntStrings(rafflePrizeBeforeClaim.Amount, prizeAmount)
	require.NoError(t, err)

	assert.Equal(t, expectedWinnerPrizeAfter, winnerPrizeAfter.Amount)
	assert.Equal(t, expectedRafflePrizeAfterClaim, rafflePrizeAfterClaim.Amount)

	winnerPrizeAfter = getFTBalance(t, c, prizeToken.Address, winner)
	rafflePrizeAfterClaim = getFTBalanceOrZero(t, c, prizeToken.Address, raffleAddress)

	// ------------------
	//      WITHDRAW
	// ------------------
	c.SetPrivateKey(ownerPriv)

	withdrawAmount := "500000"
	withdrawUUID := "withdraw-ft-001"

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

	withdrawEvent, err := utils.UnmarshalEvent[raffleV1Domain.Treasury](withdrawLog.Event)
	require.NoError(t, err)

	assert.Equal(t, raffleAddress, withdrawEvent.Address)
	assert.Equal(t, payToken.Address, withdrawEvent.TokenAddress)
	assert.Equal(t, withdrawAmount, withdrawEvent.Amount)
	assert.Equal(t, raffleV1Domain.ACTION_WITHDRAW, withdrawEvent.Action)
	assert.NotZero(t, withdrawEvent.Time)

	ownerPayAfterWithdraw := getFTBalance(t, c, payToken.Address, owner.PublicKey)
	rafflePayAfterWithdraw := getFTBalance(t, c, payToken.Address, raffleAddress)

	expectedOwnerPayAfterWithdraw, err := utils.AddBigIntStrings(ownerPayBeforeWithdraw.Amount, withdrawAmount)
	require.NoError(t, err)
	expectedRafflePayAfterWithdraw, err := utils.SubBigIntStrings("1500000", withdrawAmount)
	require.NoError(t, err)

	assert.Equal(t, expectedOwnerPayAfterWithdraw, ownerPayAfterWithdraw.Amount)
	assert.Equal(t, expectedRafflePayAfterWithdraw, rafflePayAfterWithdraw.Amount)

	// ------------------
	//      GET RAFFLE
	// ------------------
	// getRaffleOut, err := c.GetRaffle(raffleAddress)
	// require.NoError(t, err)
	// require.NotEmpty(t, getRaffleOut.States)

	// var raffleState raffleV1Models.RaffleStateModel
	// err = utils.UnmarshalState[raffleV1Models.RaffleStateModel](getRaffleOut.States[0].Object, &raffleState)
	// require.NoError(t, err)

	// assert.Equal(t, raffleAddress, raffleState.Address)
	// assert.Equal(t, owner.PublicKey, raffleState.Owner)
	// assert.Equal(t, payToken.Address, raffleState.TokenAddress)
	// assert.Equal(t, newTicketPrice, raffleState.TicketPrice)
	// assert.Equal(t, newMaxEntries, raffleState.MaxEntries)
	// assert.Equal(t, newMaxEntriesPerUser, raffleState.MaxEntriesPerUser)
	// assert.Equal(t, false, raffleState.Paused)
	// assert.Equal(t, seedCommitHex, raffleState.SeedCommitHex)
	// assert.Equal(t, newMetadata["name"], raffleState.Metadata["name"])
	// assert.Equal(t, newMetadata["description"], raffleState.Metadata["description"])

	// ------------------
	//     LIST PRIZES
	// ------------------
	// listPrizesOut, err := c.ListRafflePrizes(raffleAddress, 1, 10, true)
	// require.NoError(t, err)
	// require.NotEmpty(t, listPrizesOut.States)

	// var prizes []raffleV1Models.RafflePrizeModel
	// err = utils.UnmarshalState[[]raffleV1Models.RafflePrizeModel](listPrizesOut.States[0].Object, &prizes)
	// require.NoError(t, err)

	// require.NotEmpty(t, prizes)

	// var found bool
	// for _, p := range prizes {
	// 	if p.RaffleAddress == raffleAddress && p.TokenAddress == prizeToken.Address && p.Amount == prizeAmount {
	// 		found = true
	// 		assert.Equal(t, winner, p.Winner)
	// 		assert.Equal(t, true, p.Claimed)
	// 		assert.NotEmpty(t, p.UUID)
	// 		assert.Equal(t, owner.PublicKey, p.Sponsor)
	// 		break
	// 	}
	// }
	// assert.True(t, found, "expected to find claimed raffle prize in ListRafflePrizes")
}
