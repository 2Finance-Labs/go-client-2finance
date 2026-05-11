package e2e_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/2finance/2finance-network/blockchain/contract/cashbackV1"
	cashbackV1Domain "gitlab.com/2finance/2finance-network/blockchain/contract/cashbackV1/domain"
	cashbackV1Models "gitlab.com/2finance/2finance-network/blockchain/contract/cashbackV1/models"
	tokenV1Domain "gitlab.com/2finance/2finance-network/blockchain/contract/tokenV1/domain"
	tokenV1Models "gitlab.com/2finance/2finance-network/blockchain/contract/tokenV1/models"
	"gitlab.com/2finance/2finance-network/blockchain/log"
	"gitlab.com/2finance/2finance-network/blockchain/utils"
)

func TestCashbackFlow(t *testing.T) {
	wm := setupWalletManager(t)
	c := setupClient(t, wm)

	// ------------------
	//      WALLETS
	// ------------------
	owner, ownerPriv := createWallet(t, c, wm)
	customer, customerPriv := createWallet(t, c, wm)
	if err := wm.SetOwner(owner.PublicKey); err != nil {
		t.Fatalf("SetOwner: %v", err)
	}
	wm.SetPrivateKey(ownerPriv)

	// ------------------
	//      TOKEN
	// ------------------
	cashbackToken := createBasicToken(
		t,
		c,
		owner.PublicKey,
		6,
		false,
		tokenV1Domain.FUNGIBLE,
		false,
	)

	require.Equal(t, tokenV1Domain.FUNGIBLE, cashbackToken.TokenType, "cashback token must be fungible")

	// ------------------
	//   DEPLOY CASHBACK
	// ------------------
	wm.SetPrivateKey(ownerPriv)
	deployedContract, err := c.DeployContract1(cashbackV1.CASHBACK_CONTRACT_V1)
	if err != nil {
		t.Fatalf("DeployContract cashback: %v", err)
	}
	require.NotEmpty(t, deployedContract.Logs)

	deployLog, err := utils.UnmarshalLog[log.Log](deployedContract.Logs[0])
	if err != nil {
		t.Fatalf("UnmarshalLog (DeployContract cashback.Logs[0]): %v", err)
	}

	cashbackAddress := deployLog.ContractAddress
	require.NotEmpty(t, cashbackAddress)

	// ------------------
	//   ALLOW USERS
	// ------------------
	if err := wm.SetOwner(owner.PublicKey); err != nil {
		t.Fatalf("SetOwner: %v", err)
	}
	wm.SetPrivateKey(ownerPriv)

	_, err = c.AddAllowedUsers(cashbackToken.Address, map[string]bool{
		owner.PublicKey:    true,
		customer.PublicKey: true,
		cashbackAddress:    true,
	})
	if err != nil {
		t.Fatalf("AddAllowedUsers cashback token: %v", err)
	}

	// ------------------
	//   FUND OWNER
	// ------------------
	ownerBalanceBeforeOut, err := c.GetTokenBalance(cashbackToken.Address, owner.PublicKey)
	if err != nil {
		t.Fatalf("GetTokenBalance owner before: %v", err)
	}
	var ownerBalanceBefore tokenV1Models.BalanceStateModel
	err = utils.UnmarshalState[tokenV1Models.BalanceStateModel](ownerBalanceBeforeOut.States[0].Object, &ownerBalanceBefore)
	if err != nil {
		t.Fatalf("UnmarshalState ownerBalanceBefore: %v", err)
	}

	var customerBalanceBefore tokenV1Models.BalanceStateModel
	customerBalanceBeforeOut, err := c.GetTokenBalance(cashbackToken.Address, customer.PublicKey)
	if err != nil {
		require.Contains(t, err.Error(), "record not found")
		customerBalanceBefore = tokenV1Models.BalanceStateModel{
			TokenAddress: cashbackToken.Address,
			OwnerAddress: customer.PublicKey,
			Amount:       "0",
			TokenType:    tokenV1Domain.FUNGIBLE,
		}
	} else {
		err = utils.UnmarshalState[tokenV1Models.BalanceStateModel](customerBalanceBeforeOut.States[0].Object, &customerBalanceBefore)
		if err != nil {
			t.Fatalf("UnmarshalState customerBalanceBefore: %v", err)
		}
	}

	// ------------------
	//   ADD CASHBACK
	// ------------------
	if err := wm.SetOwner(owner.PublicKey); err != nil {
		t.Fatalf("SetOwner: %v", err)
	}

	startAt := time.Now().Add(2 * time.Second)
	expiredAt := time.Now().Add(24 * time.Hour)
	programType := "fixed-percentage"
	percentage := "1000"

	wm.SetPrivateKey(ownerPriv)
	addOut, err := c.AddCashback(
		cashbackAddress,
		owner.PublicKey,
		cashbackToken.Address,
		programType,
		percentage,
		startAt,
		expiredAt,
		false,
	)
	if err != nil {
		t.Fatalf("AddCashback: %v", err)
	}
	require.NotEmpty(t, addOut.Logs)

	addLog, err := utils.UnmarshalLog[log.Log](addOut.Logs[0])
	if err != nil {
		t.Fatalf("UnmarshalLog (AddCashback.Logs[0]): %v", err)
	}

	// Ajuste este log se o nome exato no domínio estiver diferente
	assert.Equal(t, cashbackV1Domain.CASHBACK_CREATED_LOG, addLog.LogType)

	addEvent, err := utils.UnmarshalEvent[cashbackV1Domain.Cashback](addLog.Event)
	if err != nil {
		t.Fatalf("UnmarshalEvent (AddCashback.Logs[0]): %v", err)
	}

	assert.Equal(t, cashbackAddress, addEvent.Address)
	assert.Equal(t, owner.PublicKey, addEvent.Owner)
	assert.Equal(t, cashbackToken.Address, addEvent.TokenAddress)
	assert.Equal(t, programType, addEvent.ProgramType)
	assert.Equal(t, percentage, addEvent.Percentage)
	assert.False(t, addEvent.Paused)
	if !addEvent.StartAt.IsZero() {
		assert.WithinDuration(t, startAt, addEvent.StartAt, time.Second)
	}
	if !addEvent.ExpiredAt.IsZero() {
		assert.WithinDuration(t, expiredAt, addEvent.ExpiredAt, time.Second)
	}

	getCashbackOut, err := c.GetCashback(cashbackAddress)
	if err != nil {
		t.Fatalf("GetCashback after add: %v", err)
	}
	require.NotEmpty(t, getCashbackOut.States)

	var cashbackState cashbackV1Models.CashbackStateModel
	err = utils.UnmarshalState[cashbackV1Models.CashbackStateModel](getCashbackOut.States[0].Object, &cashbackState)
	if err != nil {
		t.Fatalf("UnmarshalState (GetCashback.States[0]): %v", err)
	}

	assert.Equal(t, cashbackAddress, cashbackState.Address)
	assert.Equal(t, owner.PublicKey, cashbackState.Owner)
	assert.Equal(t, cashbackToken.Address, cashbackState.TokenAddress)
	assert.Equal(t, programType, cashbackState.ProgramType)
	assert.Equal(t, percentage, cashbackState.Percentage)
	assert.False(t, cashbackState.Paused)
	if cashbackState.StartAt != nil && !cashbackState.StartAt.IsZero() {
		assert.WithinDuration(t, startAt, *cashbackState.StartAt, time.Second)
	}
	if cashbackState.ExpiredAt != nil && !cashbackState.ExpiredAt.IsZero() {
		assert.WithinDuration(t, expiredAt, *cashbackState.ExpiredAt, time.Second)
	}

	// ------------------
	//      UPDATE
	// ------------------
	if err := wm.SetOwner(owner.PublicKey); err != nil {
		t.Fatalf("SetOwner: %v", err)
	}

	updatedPercentage := "1500"
	updatedExpiredAt := time.Now().Add(48 * time.Hour)

	wm.SetPrivateKey(ownerPriv)
	updateOut, err := c.UpdateCashback(
		cashbackAddress,
		cashbackToken.Address,
		programType,
		updatedPercentage,
		startAt,
		updatedExpiredAt,
	)
	if err != nil {
		t.Fatalf("UpdateCashback: %v", err)
	}
	require.NotEmpty(t, updateOut.Logs)

	updateLog, err := utils.UnmarshalLog[log.Log](updateOut.Logs[0])
	if err != nil {
		t.Fatalf("UnmarshalLog (UpdateCashback.Logs[0]): %v", err)
	}

	assert.Equal(t, cashbackV1Domain.CASHBACK_UPDATED_LOG, updateLog.LogType)

	updateEvent, err := utils.UnmarshalEvent[cashbackV1Domain.Cashback](updateLog.Event)
	if err != nil {
		t.Fatalf("UnmarshalEvent (UpdateCashback.Logs[0]): %v", err)
	}

	assert.Equal(t, cashbackAddress, updateEvent.Address)
	assert.Equal(t, updatedPercentage, updateEvent.Percentage)

	getCashbackOut, err = c.GetCashback(cashbackAddress)
	if err != nil {
		t.Fatalf("GetCashback after update: %v", err)
	}
	err = utils.UnmarshalState[cashbackV1Models.CashbackStateModel](getCashbackOut.States[0].Object, &cashbackState)
	if err != nil {
		t.Fatalf("UnmarshalState (GetCashback after update): %v", err)
	}

	assert.Equal(t, updatedPercentage, cashbackState.Percentage)
	if cashbackState.ExpiredAt != nil && !cashbackState.ExpiredAt.IsZero() {
		assert.WithinDuration(t, updatedExpiredAt, *cashbackState.ExpiredAt, time.Second)
	}

	// ------------------
	//       PAUSE
	// ------------------
	if err := wm.SetOwner(owner.PublicKey); err != nil {
		t.Fatalf("SetOwner: %v", err)
	}

	wm.SetPrivateKey(ownerPriv)
	pauseOut, err := c.PauseCashback(cashbackAddress, true)
	if err != nil {
		t.Fatalf("PauseCashback: %v", err)
	}
	require.NotEmpty(t, pauseOut.Logs)

	pauseLog, err := utils.UnmarshalLog[log.Log](pauseOut.Logs[0])
	if err != nil {
		t.Fatalf("UnmarshalLog (PauseCashback.Logs[0]): %v", err)
	}

	assert.Equal(t, cashbackV1Domain.CASHBACK_PAUSED_LOG, pauseLog.LogType)

	pauseEvent, err := utils.UnmarshalEvent[cashbackV1Domain.Cashback](pauseLog.Event)
	if err != nil {
		t.Fatalf("UnmarshalEvent (PauseCashback.Logs[0]): %v", err)
	}

	assert.Equal(t, cashbackAddress, pauseEvent.Address)
	assert.True(t, pauseEvent.Paused)

	getCashbackOut, err = c.GetCashback(cashbackAddress)
	if err != nil {
		t.Fatalf("GetCashback after pause: %v", err)
	}
	err = utils.UnmarshalState[cashbackV1Models.CashbackStateModel](getCashbackOut.States[0].Object, &cashbackState)
	if err != nil {
		t.Fatalf("UnmarshalState (GetCashback after pause): %v", err)
	}
	assert.True(t, cashbackState.Paused)

	// ------------------
	//      UNPAUSE
	// ------------------
	if err := wm.SetOwner(owner.PublicKey); err != nil {
		t.Fatalf("SetOwner: %v", err)
	}

	wm.SetPrivateKey(ownerPriv)
	unpauseOut, err := c.UnpauseCashback(cashbackAddress, false)
	if err != nil {
		t.Fatalf("UnpauseCashback: %v", err)
	}
	require.NotEmpty(t, unpauseOut.Logs)

	unpauseLog, err := utils.UnmarshalLog[log.Log](unpauseOut.Logs[0])
	if err != nil {
		t.Fatalf("UnmarshalLog (UnpauseCashback.Logs[0]): %v", err)
	}

	assert.Equal(t, cashbackV1Domain.CASHBACK_UNPAUSED_LOG, unpauseLog.LogType)

	unpauseEvent, err := utils.UnmarshalEvent[cashbackV1Domain.Cashback](unpauseLog.Event)
	if err != nil {
		t.Fatalf("UnmarshalEvent (UnpauseCashback.Logs[0]): %v", err)
	}

	assert.Equal(t, cashbackAddress, unpauseEvent.Address)
	assert.False(t, unpauseEvent.Paused)

	getCashbackOut, err = c.GetCashback(cashbackAddress)
	if err != nil {
		t.Fatalf("GetCashback after unpause: %v", err)
	}
	err = utils.UnmarshalState[cashbackV1Models.CashbackStateModel](getCashbackOut.States[0].Object, &cashbackState)
	if err != nil {
		t.Fatalf("UnmarshalState (GetCashback after unpause): %v", err)
	}
	assert.False(t, cashbackState.Paused)

	// ------------------
	//      DEPOSIT
	// ------------------
	if err := wm.SetOwner(owner.PublicKey); err != nil {
		t.Fatalf("SetOwner: %v", err)
	}

	depositAmount := "200"

	wm.SetPrivateKey(ownerPriv)
	depositOut, err := c.DepositCashbackFunds(
		cashbackAddress,
		cashbackToken.Address,
		depositAmount,
		tokenV1Domain.FUNGIBLE,
		"",
	)
	if err != nil {
		t.Fatalf("DepositCashbackFunds: %v", err)
	}
	require.NotEmpty(t, depositOut.Logs)

	depositLog, err := utils.UnmarshalLog[log.Log](depositOut.Logs[0])
	if err != nil {
		t.Fatalf("UnmarshalLog (DepositCashbackFunds.Logs[0]): %v", err)
	}

	assert.Equal(t, cashbackV1Domain.CASHBACK_DEPOSITED_LOG, depositLog.LogType)

	depositEvent, err := utils.UnmarshalEvent[cashbackV1Domain.Cashback](depositLog.Event)
	if err != nil {
		t.Fatalf("UnmarshalEvent (DepositCashbackFunds.Logs[0]): %v", err)
	}

	assert.Equal(t, cashbackAddress, depositEvent.Address)

	ownerBalanceAfterDepositOut, err := c.GetTokenBalance(cashbackToken.Address, owner.PublicKey)
	if err != nil {
		t.Fatalf("GetTokenBalance owner after deposit: %v", err)
	}
	var ownerBalanceAfterDeposit tokenV1Models.BalanceStateModel
	err = utils.UnmarshalState[tokenV1Models.BalanceStateModel](ownerBalanceAfterDepositOut.States[0].Object, &ownerBalanceAfterDeposit)
	if err != nil {
		t.Fatalf("UnmarshalState ownerBalanceAfterDeposit: %v", err)
	}

	expectedOwnerBalanceAfterDeposit, err := utils.SubBigIntStrings(ownerBalanceBefore.Amount, depositAmount)
	if err != nil {
		t.Fatalf("SubBigIntStrings ownerBalanceBefore - depositAmount: %v", err)
	}
	assert.Equal(t, expectedOwnerBalanceAfterDeposit, ownerBalanceAfterDeposit.Amount)

	// ------------------
	//       CLAIM
	// ------------------
	if err := wm.SetOwner(customer.PublicKey); err != nil {
		t.Fatalf("SetOwner: %v", err)
	}

	wait := time.Until(startAt) + 500*time.Millisecond
	if wait > 0 {
		time.Sleep(wait)
	}

	claimAmount := "100"

	wm.SetPrivateKey(customerPriv)
	claimOut, err := c.ClaimCashback(
		cashbackAddress,
		claimAmount,
		tokenV1Domain.FUNGIBLE,
		"",
	)
	if err != nil {
		t.Fatalf("ClaimCashback: %v", err)
	}
	require.NotEmpty(t, claimOut.Logs)

	claimLog, err := utils.UnmarshalLog[log.Log](claimOut.Logs[0])
	if err != nil {
		t.Fatalf("UnmarshalLog (ClaimCashback.Logs[0]): %v", err)
	}

	assert.Equal(t, cashbackV1Domain.CASHBACK_CLAIMED_LOG, claimLog.LogType)

	claimEvent, err := utils.UnmarshalEvent[cashbackV1Domain.Cashback](claimLog.Event)
	if err != nil {
		t.Fatalf("UnmarshalEvent (ClaimCashback.Logs[0]): %v", err)
	}

	assert.Equal(t, cashbackAddress, claimEvent.Address)

	customerBalanceAfterClaimOut, err := c.GetTokenBalance(cashbackToken.Address, customer.PublicKey)
	if err != nil {
		t.Fatalf("GetTokenBalance customer after claim: %v", err)
	}
	var customerBalanceAfterClaim tokenV1Models.BalanceStateModel
	err = utils.UnmarshalState[tokenV1Models.BalanceStateModel](customerBalanceAfterClaimOut.States[0].Object, &customerBalanceAfterClaim)
	if err != nil {
		t.Fatalf("UnmarshalState customerBalanceAfterClaim: %v", err)
	}

	cashbackReceived := "15"

	expectedCustomerBalanceAfterClaim, err := utils.AddBigIntStrings(customerBalanceBefore.Amount, cashbackReceived)
	if err != nil {
		t.Fatalf("AddBigIntStrings customerBalanceBefore + cashbackReceived: %v", err)
	}
	assert.Equal(t, expectedCustomerBalanceAfterClaim, customerBalanceAfterClaim.Amount)

	// ------------------
	//      WITHDRAW
	// ------------------
	if err := wm.SetOwner(owner.PublicKey); err != nil {
		t.Fatalf("SetOwner: %v", err)
	}

	withdrawAmount := "100"

	wm.SetPrivateKey(ownerPriv)
	withdrawOut, err := c.WithdrawCashbackFunds(
		cashbackAddress,
		cashbackToken.Address,
		withdrawAmount,
		tokenV1Domain.FUNGIBLE,
		"",
	)
	if err != nil {
		t.Fatalf("WithdrawCashbackFunds: %v", err)
	}
	require.NotEmpty(t, withdrawOut.Logs)

	withdrawLog, err := utils.UnmarshalLog[log.Log](withdrawOut.Logs[0])
	if err != nil {
		t.Fatalf("UnmarshalLog (WithdrawCashbackFunds.Logs[0]): %v", err)
	}

	assert.Equal(t, cashbackV1Domain.CASHBACK_WITHDRAWN_LOG, withdrawLog.LogType)

	withdrawEvent, err := utils.UnmarshalEvent[cashbackV1Domain.Cashback](withdrawLog.Event)
	if err != nil {
		t.Fatalf("UnmarshalEvent (WithdrawCashbackFunds.Logs[0]): %v", err)
	}

	assert.Equal(t, cashbackAddress, withdrawEvent.Address)

	ownerBalanceAfterWithdrawOut, err := c.GetTokenBalance(cashbackToken.Address, owner.PublicKey)
	if err != nil {
		t.Fatalf("GetTokenBalance owner after withdraw: %v", err)
	}
	var ownerBalanceAfterWithdraw tokenV1Models.BalanceStateModel
	err = utils.UnmarshalState[tokenV1Models.BalanceStateModel](ownerBalanceAfterWithdrawOut.States[0].Object, &ownerBalanceAfterWithdraw)
	if err != nil {
		t.Fatalf("UnmarshalState ownerBalanceAfterWithdraw: %v", err)
	}

	expectedOwnerBalanceAfterWithdraw, err := utils.AddBigIntStrings(ownerBalanceAfterDeposit.Amount, withdrawAmount)
	if err != nil {
		t.Fatalf("AddBigIntStrings ownerBalanceAfterDeposit + withdrawAmount: %v", err)
	}
	assert.Equal(t, expectedOwnerBalanceAfterWithdraw, ownerBalanceAfterWithdraw.Amount)

	// ------------------
	//   LIST CASHBACKS
	// ------------------
	listOut, err := c.ListCashbacks(
		owner.PublicKey,
		cashbackToken.Address,
		programType,
		false,
		1,
		10,
		true,
	)
	if err != nil {
		t.Fatalf("ListCashbacks: %v", err)
	}
	require.NotEmpty(t, listOut.States)

	var cashbacks []cashbackV1Models.CashbackStateModel
	err = utils.UnmarshalState[[]cashbackV1Models.CashbackStateModel](listOut.States[0].Object, &cashbacks)
	if err != nil {
		t.Fatalf("UnmarshalState (ListCashbacks.States[0]): %v", err)
	}

	require.NotEmpty(t, cashbacks)

	var found bool
	for _, cb := range cashbacks {
		if cb.Address == cashbackAddress {
			found = true
			assert.Equal(t, owner.PublicKey, cb.Owner)
			assert.Equal(t, cashbackToken.Address, cb.TokenAddress)
			assert.Equal(t, programType, cb.ProgramType)
			assert.Equal(t, updatedPercentage, cb.Percentage)
			assert.False(t, cb.Paused)
			break
		}
	}

	assert.True(t, found, "expected to find cashback program in ListCashbacks")
}
