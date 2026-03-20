package e2e_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/2finance/2finance-network/blockchain/contract/paymentV1"
	paymentV1Domain "gitlab.com/2finance/2finance-network/blockchain/contract/paymentV1/domain"
	"gitlab.com/2finance/2finance-network/blockchain/contract/paymentV1/inputs"
	paymentV1Models "gitlab.com/2finance/2finance-network/blockchain/contract/paymentV1/models"
	tokenV1Domain "gitlab.com/2finance/2finance-network/blockchain/contract/tokenV1/domain"
	tokenV1Models "gitlab.com/2finance/2finance-network/blockchain/contract/tokenV1/models"
	"gitlab.com/2finance/2finance-network/blockchain/log"
	"gitlab.com/2finance/2finance-network/blockchain/utils"
)

func TestPaymentFlow(t *testing.T) {
	c := setupClient(t)

	// ------------------
	//      WALLETS
	// ------------------
	owner, ownerPriv := createWallet(t, c)
	payer, payerPriv := createWallet(t, c)
	payee, payeePriv := createWallet(t, c)

	// ------------------
	//      TOKEN
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

	require.Equal(t, tokenV1Domain.FUNGIBLE, payToken.TokenType, "payToken must be fungible")

	// ------------------
	//   DEPLOY PAYMENT
	// ------------------
	deployedContract, err := c.DeployContract1(paymentV1.PAYMENT_CONTRACT_V1)
	if err != nil {
		t.Fatalf("DeployContract: %v", err)
	}
	require.NotEmpty(t, deployedContract.Logs)

	deployLog, err := utils.UnmarshalLog[log.Log](deployedContract.Logs[0])
	if err != nil {
		t.Fatalf("UnmarshalLog (DeployContract.Logs[0]): %v", err)
	}

	paymentAddress := deployLog.ContractAddress
	require.NotEmpty(t, paymentAddress)

	// ------------------
	//   ALLOW USERS
	// ------------------
	c.SetPrivateKey(ownerPriv)

	_, err = c.AddAllowedUsers(payToken.Address, map[string]bool{
		owner.PublicKey: true,
		payer.PublicKey: true,
		payee.PublicKey: true,
		paymentAddress:  true,
	})
	if err != nil {
		t.Fatalf("AddAllowedUsers: %v", err)
	}

	// ------------------
	//  FUND THE PAYER
	// ------------------
	fundAmount := "500"
	_, err = c.TransferToken(payToken.Address, payer.PublicKey, fundAmount, []string{})
	if err != nil {
		t.Fatalf("TransferToken payer funding: %v", err)
	}

	payerBalanceBeforeOut, err := c.GetTokenBalance(payToken.Address, payer.PublicKey)
	if err != nil {
		t.Fatalf("GetTokenBalance payer before: %v", err)
	}
	var payerBalanceBefore tokenV1Models.BalanceStateModel
	err = utils.UnmarshalState[tokenV1Models.BalanceStateModel](payerBalanceBeforeOut.States[0].Object, &payerBalanceBefore)
	if err != nil {
		t.Fatalf("UnmarshalState payerBalanceBefore: %v", err)
	}

	var payeeBalanceBefore tokenV1Models.BalanceStateModel
	payeeBalanceBeforeOut, err := c.GetTokenBalance(payToken.Address, payee.PublicKey)
	if err != nil {
		require.Contains(t, err.Error(), "record not found")
		payeeBalanceBefore = tokenV1Models.BalanceStateModel{
			TokenAddress: payToken.Address,
			OwnerAddress: payee.PublicKey,
			Amount:       "0",
			TokenType:    tokenV1Domain.FUNGIBLE,
		}
	} else {
		err = utils.UnmarshalState[tokenV1Models.BalanceStateModel](payeeBalanceBeforeOut.States[0].Object, &payeeBalanceBefore)
		if err != nil {
			t.Fatalf("UnmarshalState payeeBalanceBefore: %v", err)
		}
	}

	// ------------------
	//   CREATE PAYMENT
	// ------------------
	orderId := "order-payment-e2e-001"
	amount := "300"
	expiredAt := time.Now().Add(2 * time.Hour)

	c.SetPrivateKey(ownerPriv)
	createPaymentOut, err := c.CreatePayment(inputs.InputCreate{
		Address:      paymentAddress,
		Owner:        owner.PublicKey,
		TokenAddress: payToken.Address,
		OrderId:      orderId,
		Payer:        payer.PublicKey,
		Payee:        payee.PublicKey,
		Amount:       amount,
		ExpiredAt:    expiredAt,
	})
	if err != nil {
		t.Fatalf("CreatePayment: %v", err)
	}
	require.NotEmpty(t, createPaymentOut.Logs)

	createPaymentLog, err := utils.UnmarshalLog[log.Log](createPaymentOut.Logs[0])
	if err != nil {
		t.Fatalf("UnmarshalLog (CreatePayment.Logs[0]): %v", err)
	}
	assert.Equal(t, paymentV1Domain.PAYMENT_CREATED_LOG, createPaymentLog.LogType)

	createPaymentEvent, err := utils.UnmarshalEvent[paymentV1Domain.Payment](createPaymentLog.Event)
	if err != nil {
		t.Fatalf("UnmarshalEvent (CreatePayment.Logs[0]): %v", err)
	}

	assert.Equal(t, paymentAddress, createPaymentEvent.Address)
	assert.Equal(t, owner.PublicKey, createPaymentEvent.Owner)
	assert.Equal(t, payToken.Address, createPaymentEvent.TokenAddress)
	assert.Equal(t, orderId, createPaymentEvent.OrderId)
	assert.Equal(t, payer.PublicKey, createPaymentEvent.Payer)
	assert.Equal(t, payee.PublicKey, createPaymentEvent.Payee)
	assert.Equal(t, amount, createPaymentEvent.Amount)
	assert.Equal(t, paymentV1Domain.STATUS_CREATED, createPaymentEvent.Status)
	assert.False(t, createPaymentEvent.Paused)
	if !createPaymentEvent.ExpiredAt.IsZero() {
		assert.WithinDuration(t, expiredAt, createPaymentEvent.ExpiredAt, time.Second)
	}

	getPaymentOut, err := c.GetPayment(paymentAddress)
	if err != nil {
		t.Fatalf("GetPayment: %v", err)
	}
	require.NotEmpty(t, getPaymentOut.States)

	var paymentState paymentV1Models.PaymentStateModel
	err = utils.UnmarshalState[paymentV1Models.PaymentStateModel](getPaymentOut.States[0].Object, &paymentState)
	if err != nil {
		t.Fatalf("UnmarshalState (GetPayment.States[0]): %v", err)
	}

	assert.Equal(t, paymentAddress, paymentState.Address)
	assert.Equal(t, owner.PublicKey, paymentState.Owner)
	assert.Equal(t, payToken.Address, paymentState.TokenAddress)
	assert.Equal(t, orderId, paymentState.OrderId)
	assert.Equal(t, payer.PublicKey, paymentState.Payer)
	assert.Equal(t, payee.PublicKey, paymentState.Payee)
	assert.Equal(t, amount, paymentState.Amount)
	assert.Equal(t, paymentV1Domain.STATUS_CREATED, paymentState.Status)
	assert.False(t, paymentState.Paused)
	if !paymentState.ExpiredAt.IsZero() {
		assert.WithinDuration(t, expiredAt, paymentState.ExpiredAt, time.Second)
	}
	assert.NotEmpty(t, paymentState.Hash)

	// ------------------
	//       PAUSE
	// ------------------
	c.SetPrivateKey(payerPriv)
	pauseOut, err := c.PausePayment(inputs.InputPause{
		Address: paymentAddress,
		Paused:  true,
	})
	if err != nil {
		t.Fatalf("PausePayment: %v", err)
	}
	require.NotEmpty(t, pauseOut.Logs)

	pauseLog, err := utils.UnmarshalLog[log.Log](pauseOut.Logs[0])
	if err != nil {
		t.Fatalf("UnmarshalLog (PausePayment.Logs[0]): %v", err)
	}
	assert.Equal(t, paymentV1Domain.PAYMENT_PAUSED_LOG, pauseLog.LogType)

	pauseEvent, err := utils.UnmarshalEvent[paymentV1Domain.Payment](pauseLog.Event)
	if err != nil {
		t.Fatalf("UnmarshalEvent (PausePayment.Logs[0]): %v", err)
	}
	assert.Equal(t, paymentAddress, pauseEvent.Address)
	assert.True(t, pauseEvent.Paused)

	getPaymentOut, err = c.GetPayment(paymentAddress)
	if err != nil {
		t.Fatalf("GetPayment after pause: %v", err)
	}
	err = utils.UnmarshalState[paymentV1Models.PaymentStateModel](getPaymentOut.States[0].Object, &paymentState)
	if err != nil {
		t.Fatalf("UnmarshalState (GetPayment after pause): %v", err)
	}
	assert.True(t, paymentState.Paused)

	// ------------------
	//      UNPAUSE
	// ------------------
	c.SetPrivateKey(payerPriv)
	unpauseOut, err := c.UnpausePayment(inputs.InputPause{
		Address: paymentAddress,
		Paused:  false,
	})
	if err != nil {
		t.Fatalf("UnpausePayment: %v", err)
	}
	require.NotEmpty(t, unpauseOut.Logs)

	unpauseLog, err := utils.UnmarshalLog[log.Log](unpauseOut.Logs[0])
	if err != nil {
		t.Fatalf("UnmarshalLog (UnpausePayment.Logs[0]): %v", err)
	}
	assert.Equal(t, paymentV1Domain.PAYMENT_UNPAUSED_LOG, unpauseLog.LogType)

	unpauseEvent, err := utils.UnmarshalEvent[paymentV1Domain.Payment](unpauseLog.Event)
	if err != nil {
		t.Fatalf("UnmarshalEvent (UnpausePayment.Logs[0]): %v", err)
	}
	assert.Equal(t, paymentAddress, unpauseEvent.Address)
	assert.False(t, unpauseEvent.Paused)

	getPaymentOut, err = c.GetPayment(paymentAddress)
	if err != nil {
		t.Fatalf("GetPayment after unpause: %v", err)
	}
	err = utils.UnmarshalState[paymentV1Models.PaymentStateModel](getPaymentOut.States[0].Object, &paymentState)
	if err != nil {
		t.Fatalf("UnmarshalState (GetPayment after unpause): %v", err)
	}
	assert.False(t, paymentState.Paused)

	// ------------------
	//     AUTHORIZE
	// ------------------
	c.SetPrivateKey(payerPriv)
	authorizeOut, err := c.AuthorizePayment(inputs.InputAuthorize{
		Address: paymentAddress,
	})
	if err != nil {
		t.Fatalf("AuthorizePayment: %v", err)
	}
	require.NotEmpty(t, authorizeOut.Logs)

	authorizeLog, err := utils.UnmarshalLog[log.Log](authorizeOut.Logs[0])
	if err != nil {
		t.Fatalf("UnmarshalLog (AuthorizePayment.Logs[0]): %v", err)
	}
	assert.Equal(t, paymentV1Domain.PAYMENT_AUTHORIZED_LOG, authorizeLog.LogType)

	authorizeEvent, err := utils.UnmarshalEvent[paymentV1Domain.Payment](authorizeLog.Event)
	if err != nil {
		t.Fatalf("UnmarshalEvent (AuthorizePayment.Logs[0]): %v", err)
	}

	assert.Equal(t, paymentAddress, authorizeEvent.Address)
	assert.Equal(t, paymentV1Domain.STATUS_AUTHORIZED, authorizeEvent.Status)

	getPaymentOut, err = c.GetPayment(paymentAddress)
	if err != nil {
		t.Fatalf("GetPayment after authorize: %v", err)
	}
	err = utils.UnmarshalState[paymentV1Models.PaymentStateModel](getPaymentOut.States[0].Object, &paymentState)
	if err != nil {
		t.Fatalf("UnmarshalState (GetPayment after authorize): %v", err)
	}

	assert.Equal(t, paymentV1Domain.STATUS_AUTHORIZED, paymentState.Status)

	// ------------------
	//      CAPTURE
	// ------------------
	c.SetPrivateKey(payeePriv)
	captureOut, err := c.CapturePayment(inputs.InputCapture{
		Address: paymentAddress,
	})
	if err != nil {
		t.Fatalf("CapturePayment: %v", err)
	}
	require.NotEmpty(t, captureOut.Logs)

	captureLog, err := utils.UnmarshalLog[log.Log](captureOut.Logs[0])
	if err != nil {
		t.Fatalf("UnmarshalLog (CapturePayment.Logs[0]): %v", err)
	}
	assert.Equal(t, paymentV1Domain.PAYMENT_CAPTURED_LOG, captureLog.LogType)

	captureEvent, err := utils.UnmarshalEvent[paymentV1Domain.Payment](captureLog.Event)
	if err != nil {
		t.Fatalf("UnmarshalEvent (CapturePayment.Logs[0]): %v", err)
	}

	assert.Equal(t, paymentAddress, captureEvent.Address)
	assert.Equal(t, paymentV1Domain.STATUS_CAPTURED, captureEvent.Status)

	getPaymentOut, err = c.GetPayment(paymentAddress)
	if err != nil {
		t.Fatalf("GetPayment after capture: %v", err)
	}
	err = utils.UnmarshalState[paymentV1Models.PaymentStateModel](getPaymentOut.States[0].Object, &paymentState)
	if err != nil {
		t.Fatalf("UnmarshalState (GetPayment after capture): %v", err)
	}

	assert.Equal(t, paymentV1Domain.STATUS_CAPTURED, paymentState.Status)
	assert.Equal(t, amount, paymentState.CapturedAmount)

	payerBalanceAfterCaptureOut, err := c.GetTokenBalance(payToken.Address, payer.PublicKey)
	if err != nil {
		t.Fatalf("GetTokenBalance payer after capture: %v", err)
	}
	var payerBalanceAfterCapture tokenV1Models.BalanceStateModel
	err = utils.UnmarshalState[tokenV1Models.BalanceStateModel](payerBalanceAfterCaptureOut.States[0].Object, &payerBalanceAfterCapture)
	if err != nil {
		t.Fatalf("UnmarshalState payerBalanceAfterCapture: %v", err)
	}

	payeeBalanceAfterCaptureOut, err := c.GetTokenBalance(payToken.Address, payee.PublicKey)
	if err != nil {
		t.Fatalf("GetTokenBalance payee after capture: %v", err)
	}
	var payeeBalanceAfterCapture tokenV1Models.BalanceStateModel
	err = utils.UnmarshalState[tokenV1Models.BalanceStateModel](payeeBalanceAfterCaptureOut.States[0].Object, &payeeBalanceAfterCapture)
	if err != nil {
		t.Fatalf("UnmarshalState payeeBalanceAfterCapture: %v", err)
	}

	expectedPayerBalanceAfterCapture, err := utils.SubBigIntStrings(payerBalanceBefore.Amount, amount)
	if err != nil {
		t.Fatalf("SubBigIntStrings payerBalanceBefore - amount: %v", err)
	}
	expectedPayeeBalanceAfterCapture, err := utils.AddBigIntStrings(payeeBalanceBefore.Amount, amount)
	if err != nil {
		t.Fatalf("AddBigIntStrings payeeBalanceBefore + amount: %v", err)
	}

	assert.Equal(t, expectedPayerBalanceAfterCapture, payerBalanceAfterCapture.Amount)
	assert.Equal(t, expectedPayeeBalanceAfterCapture, payeeBalanceAfterCapture.Amount)

	// ------------------
	//       REFUND
	// ------------------
	refundAmount := "100"

	c.SetPrivateKey(payeePriv)
	refundOut, err := c.RefundPayment(inputs.InputRefund{
		Address: paymentAddress,
		Amount:  refundAmount,
	})
	if err != nil {
		t.Fatalf("RefundPayment: %v", err)
	}
	require.NotEmpty(t, refundOut.Logs)

	refundLog, err := utils.UnmarshalLog[log.Log](refundOut.Logs[0])
	if err != nil {
		t.Fatalf("UnmarshalLog (RefundPayment.Logs[0]): %v", err)
	}
	assert.Equal(t, paymentV1Domain.PAYMENT_REFUNDED_LOG, refundLog.LogType)

	refundEvent, err := utils.UnmarshalEvent[paymentV1Domain.Payment](refundLog.Event)
	if err != nil {
		t.Fatalf("UnmarshalEvent (RefundPayment.Logs[0]): %v", err)
	}

	assert.Equal(t, paymentAddress, refundEvent.Address)
	assert.Equal(t, paymentV1Domain.STATUS_REFUNDED, refundEvent.Status)
	assert.Equal(t, refundAmount, refundEvent.RefundedAmount)

	getPaymentOut, err = c.GetPayment(paymentAddress)
	if err != nil {
		t.Fatalf("GetPayment after refund: %v", err)
	}
	err = utils.UnmarshalState[paymentV1Models.PaymentStateModel](getPaymentOut.States[0].Object, &paymentState)
	if err != nil {
		t.Fatalf("UnmarshalState (GetPayment after refund): %v", err)
	}

	assert.Equal(t, paymentV1Domain.STATUS_REFUNDED, paymentState.Status)
	assert.Equal(t, refundAmount, paymentState.RefundedAmount)

	payerBalanceAfterRefundOut, err := c.GetTokenBalance(payToken.Address, payer.PublicKey)
	if err != nil {
		t.Fatalf("GetTokenBalance payer after refund: %v", err)
	}
	var payerBalanceAfterRefund tokenV1Models.BalanceStateModel
	err = utils.UnmarshalState[tokenV1Models.BalanceStateModel](payerBalanceAfterRefundOut.States[0].Object, &payerBalanceAfterRefund)
	if err != nil {
		t.Fatalf("UnmarshalState payerBalanceAfterRefund: %v", err)
	}

	payeeBalanceAfterRefundOut, err := c.GetTokenBalance(payToken.Address, payee.PublicKey)
	if err != nil {
		t.Fatalf("GetTokenBalance payee after refund: %v", err)
	}
	var payeeBalanceAfterRefund tokenV1Models.BalanceStateModel
	err = utils.UnmarshalState[tokenV1Models.BalanceStateModel](payeeBalanceAfterRefundOut.States[0].Object, &payeeBalanceAfterRefund)
	if err != nil {
		t.Fatalf("UnmarshalState payeeBalanceAfterRefund: %v", err)
	}

	expectedPayerBalanceAfterRefund, err := utils.AddBigIntStrings(payerBalanceAfterCapture.Amount, refundAmount)
	if err != nil {
		t.Fatalf("AddBigIntStrings payerBalanceAfterCapture + refundAmount: %v", err)
	}
	expectedPayeeBalanceAfterRefund, err := utils.SubBigIntStrings(payeeBalanceAfterCapture.Amount, refundAmount)
	if err != nil {
		t.Fatalf("SubBigIntStrings payeeBalanceAfterCapture - refundAmount: %v", err)
	}

	assert.Equal(t, expectedPayerBalanceAfterRefund, payerBalanceAfterRefund.Amount)
	assert.Equal(t, expectedPayeeBalanceAfterRefund, payeeBalanceAfterRefund.Amount)

	// ------------------
	//     VOID FLOW
	// ------------------
	deployedContract2, err := c.DeployContract1(paymentV1.PAYMENT_CONTRACT_V1)
	if err != nil {
		t.Fatalf("DeployContract second payment: %v", err)
	}
	require.NotEmpty(t, deployedContract2.Logs)

	deployLog2, err := utils.UnmarshalLog[log.Log](deployedContract2.Logs[0])
	if err != nil {
		t.Fatalf("UnmarshalLog (DeployContract second payment.Logs[0]): %v", err)
	}

	voidPaymentAddress := deployLog2.ContractAddress
	require.NotEmpty(t, voidPaymentAddress)

	_, err = c.AddAllowedUsers(payToken.Address, map[string]bool{
		voidPaymentAddress: true,
	})
	if err != nil {
		t.Fatalf("AddAllowedUsers second payment address: %v", err)
	}

	voidOrderId := "order-payment-e2e-void-001"
	voidAmount := "80"

	c.SetPrivateKey(ownerPriv)
	_, err = c.CreatePayment(inputs.InputCreate{
		Address:      paymentAddress,
		Owner:        owner.PublicKey,
		TokenAddress: payToken.Address,
		OrderId:      orderId,
		Payer:        payer.PublicKey,
		Payee:        payee.PublicKey,
		Amount:       amount,
		ExpiredAt:    expiredAt,
	})
	if err != nil {
		t.Fatalf("CreatePayment void flow: %v", err)
	}

	c.SetPrivateKey(payerPriv)
	_, err = c.AuthorizePayment(inputs.InputAuthorize{
		Address: paymentAddress,
	})
	if err != nil {
		t.Fatalf("AuthorizePayment void flow: %v", err)
	}

	voidOut, err := c.VoidPayment(inputs.InputVoidPayment{
		Address: voidPaymentAddress,
	})
	if err != nil {
		t.Fatalf("VoidPayment: %v", err)
	}
	require.NotEmpty(t, voidOut.Logs)

	voidLog, err := utils.UnmarshalLog[log.Log](voidOut.Logs[0])
	if err != nil {
		t.Fatalf("UnmarshalLog (VoidPayment.Logs[0]): %v", err)
	}
	assert.Equal(t, paymentV1Domain.PAYMENT_VOIDED_LOG, voidLog.LogType)

	voidEvent, err := utils.UnmarshalEvent[paymentV1Domain.Payment](voidLog.Event)
	if err != nil {
		t.Fatalf("UnmarshalEvent (VoidPayment.Logs[0]): %v", err)
	}

	assert.Equal(t, voidPaymentAddress, voidEvent.Address)
	assert.Equal(t, paymentV1Domain.STATUS_VOIDED, voidEvent.Status)

	getVoidPaymentOut, err := c.GetPayment(voidPaymentAddress)
	if err != nil {
		t.Fatalf("GetPayment void flow: %v", err)
	}
	require.NotEmpty(t, getVoidPaymentOut.States)

	var voidPaymentState paymentV1Models.PaymentStateModel
	err = utils.UnmarshalState[paymentV1Models.PaymentStateModel](getVoidPaymentOut.States[0].Object, &voidPaymentState)
	if err != nil {
		t.Fatalf("UnmarshalState (GetPayment void flow): %v", err)
	}

	assert.Equal(t, paymentV1Domain.STATUS_VOIDED, voidPaymentState.Status)

	// ------------------
	//     DIRECT PAY
	// ------------------
	deployedContract3, err := c.DeployContract1(paymentV1.PAYMENT_CONTRACT_V1)
	if err != nil {
		t.Fatalf("DeployContract direct pay: %v", err)
	}
	require.NotEmpty(t, deployedContract3.Logs)

	deployLog3, err := utils.UnmarshalLog[log.Log](deployedContract3.Logs[0])
	if err != nil {
		t.Fatalf("UnmarshalLog (DeployContract direct pay.Logs[0]): %v", err)
	}

	directPaymentAddress := deployLog3.ContractAddress
	require.NotEmpty(t, directPaymentAddress)

	_, err = c.AddAllowedUsers(payToken.Address, map[string]bool{
		directPaymentAddress: true,
	})
	if err != nil {
		t.Fatalf("AddAllowedUsers direct payment address: %v", err)
	}

	directPayAmount := "50"
	directPayOrderId := "order-payment-e2e-direct-001"

	payerBalanceBeforeDirectOut, err := c.GetTokenBalance(payToken.Address, payer.PublicKey)
	if err != nil {
		t.Fatalf("GetTokenBalance payer before direct pay: %v", err)
	}
	var payerBalanceBeforeDirect tokenV1Models.BalanceStateModel
	err = utils.UnmarshalState[tokenV1Models.BalanceStateModel](payerBalanceBeforeDirectOut.States[0].Object, &payerBalanceBeforeDirect)
	if err != nil {
		t.Fatalf("UnmarshalState payerBalanceBeforeDirect: %v", err)
	}

	payeeBalanceBeforeDirectOut, err := c.GetTokenBalance(payToken.Address, payee.PublicKey)
	if err != nil {
		t.Fatalf("GetTokenBalance payee before direct pay: %v", err)
	}
	var payeeBalanceBeforeDirect tokenV1Models.BalanceStateModel
	err = utils.UnmarshalState[tokenV1Models.BalanceStateModel](payeeBalanceBeforeDirectOut.States[0].Object, &payeeBalanceBeforeDirect)
	if err != nil {
		t.Fatalf("UnmarshalState payeeBalanceBeforeDirect: %v", err)
	}

	c.SetPrivateKey(payerPriv)
	directPayOut, err := c.DirectPay(inputs.InputDirectPay{
		Address:      directPaymentAddress,
		Owner:        payer.PublicKey,
		TokenAddress: payToken.Address,
		OrderId:      directPayOrderId,
		Payer:        payer.PublicKey,
		Payee:        payee.PublicKey,
		Amount:       directPayAmount,
		ExpiredAt:    time.Now().Add(2 * time.Hour),
	})
	if err != nil {
		t.Fatalf("DirectPay: %v", err)
	}

	require.Len(t, directPayOut.Logs, 3, "direct pay should generate created, authorized and captured logs")

	directCreatedLog, err := utils.UnmarshalLog[log.Log](directPayOut.Logs[0])
	if err != nil {
		t.Fatalf("UnmarshalLog (DirectPay.Logs[0]): %v", err)
	}
	assert.Equal(t, paymentV1Domain.PAYMENT_CREATED_LOG, directCreatedLog.LogType)

	directCreatedEvent, err := utils.UnmarshalEvent[paymentV1Domain.Payment](directCreatedLog.Event)
	if err != nil {
		t.Fatalf("UnmarshalEvent (DirectPay.Logs[0]): %v", err)
	}
	assert.Equal(t, directPaymentAddress, directCreatedEvent.Address)
	assert.Equal(t, payer.PublicKey, directCreatedEvent.Owner)
	assert.Equal(t, payToken.Address, directCreatedEvent.TokenAddress)
	assert.Equal(t, directPayOrderId, directCreatedEvent.OrderId)
	assert.Equal(t, payer.PublicKey, directCreatedEvent.Payer)
	assert.Equal(t, payee.PublicKey, directCreatedEvent.Payee)
	assert.Equal(t, directPayAmount, directCreatedEvent.Amount)
	assert.Equal(t, paymentV1Domain.STATUS_CREATED, directCreatedEvent.Status)
	assert.False(t, directCreatedEvent.Paused)

	directAuthorizedLog, err := utils.UnmarshalLog[log.Log](directPayOut.Logs[1])
	if err != nil {
		t.Fatalf("UnmarshalLog (DirectPay.Logs[1]): %v", err)
	}
	assert.Equal(t, paymentV1Domain.PAYMENT_AUTHORIZED_LOG, directAuthorizedLog.LogType)

	directAuthorizedEvent, err := utils.UnmarshalEvent[paymentV1Domain.Payment](directAuthorizedLog.Event)
	if err != nil {
		t.Fatalf("UnmarshalEvent (DirectPay.Logs[1]): %v", err)
	}
	assert.Equal(t, directPaymentAddress, directAuthorizedEvent.Address)
	assert.Equal(t, paymentV1Domain.STATUS_AUTHORIZED, directAuthorizedEvent.Status)

	directCapturedLog, err := utils.UnmarshalLog[log.Log](directPayOut.Logs[2])
	if err != nil {
		t.Fatalf("UnmarshalLog (DirectPay.Logs[2]): %v", err)
	}
	assert.Equal(t, paymentV1Domain.PAYMENT_CAPTURED_LOG, directCapturedLog.LogType)

	directCapturedEvent, err := utils.UnmarshalEvent[paymentV1Domain.Payment](directCapturedLog.Event)
	if err != nil {
		t.Fatalf("UnmarshalEvent (DirectPay.Logs[2]): %v", err)
	}
	assert.Equal(t, directPaymentAddress, directCapturedEvent.Address)
	assert.Equal(t, directPayAmount, directCapturedEvent.CapturedAmount)
	assert.Equal(t, paymentV1Domain.STATUS_CAPTURED, directCapturedEvent.Status)

	getDirectPaymentOut, err := c.GetPayment(directPaymentAddress)
	if err != nil {
		t.Fatalf("GetPayment direct pay: %v", err)
	}
	require.NotEmpty(t, getDirectPaymentOut.States)

	var directPaymentState paymentV1Models.PaymentStateModel
	err = utils.UnmarshalState[paymentV1Models.PaymentStateModel](getDirectPaymentOut.States[0].Object, &directPaymentState)
	if err != nil {
		t.Fatalf("UnmarshalState (GetPayment direct pay): %v", err)
	}

	assert.Equal(t, directPaymentAddress, directPaymentState.Address)
	assert.Equal(t, payer.PublicKey, directPaymentState.Owner)
	assert.Equal(t, payToken.Address, directPaymentState.TokenAddress)
	assert.Equal(t, directPayOrderId, directPaymentState.OrderId)
	assert.Equal(t, payer.PublicKey, directPaymentState.Payer)
	assert.Equal(t, payee.PublicKey, directPaymentState.Payee)
	assert.Equal(t, directPayAmount, directPaymentState.Amount)
	assert.Equal(t, directPayAmount, directPaymentState.CapturedAmount)
	assert.Equal(t, "0", directPaymentState.RefundedAmount)
	assert.Equal(t, paymentV1Domain.STATUS_CAPTURED, directPaymentState.Status)
	assert.False(t, directPaymentState.Paused)
	assert.NotEmpty(t, directPaymentState.Hash)

	payerBalanceAfterDirectOut, err := c.GetTokenBalance(payToken.Address, payer.PublicKey)
	if err != nil {
		t.Fatalf("GetTokenBalance payer after direct pay: %v", err)
	}
	var payerBalanceAfterDirect tokenV1Models.BalanceStateModel
	err = utils.UnmarshalState[tokenV1Models.BalanceStateModel](payerBalanceAfterDirectOut.States[0].Object, &payerBalanceAfterDirect)
	if err != nil {
		t.Fatalf("UnmarshalState payerBalanceAfterDirect: %v", err)
	}

	payeeBalanceAfterDirectOut, err := c.GetTokenBalance(payToken.Address, payee.PublicKey)
	if err != nil {
		t.Fatalf("GetTokenBalance payee after direct pay: %v", err)
	}
	var payeeBalanceAfterDirect tokenV1Models.BalanceStateModel
	err = utils.UnmarshalState[tokenV1Models.BalanceStateModel](payeeBalanceAfterDirectOut.States[0].Object, &payeeBalanceAfterDirect)
	if err != nil {
		t.Fatalf("UnmarshalState payeeBalanceAfterDirect: %v", err)
	}

	expectedPayerBalanceAfterDirect, err := utils.SubBigIntStrings(payerBalanceBeforeDirect.Amount, directPayAmount)
	if err != nil {
		t.Fatalf("SubBigIntStrings payerBalanceBeforeDirect - directPayAmount: %v", err)
	}
	expectedPayeeBalanceAfterDirect, err := utils.AddBigIntStrings(payeeBalanceBeforeDirect.Amount, directPayAmount)
	if err != nil {
		t.Fatalf("AddBigIntStrings payeeBalanceBeforeDirect + directPayAmount: %v", err)
	}

	assert.Equal(t, expectedPayerBalanceAfterDirect, payerBalanceAfterDirect.Amount)
	assert.Equal(t, expectedPayeeBalanceAfterDirect, payeeBalanceAfterDirect.Amount)

	// ------------------
	//     LIST PAYMENTS
	// ------------------
	listPaymentsOut, err := c.ListPayments(inputs.InputList{
		OrderId:      "",
		TokenAddress: payToken.Address,
		Status:       []string{},
		Payer:        payer.PublicKey,
		Payee:        payee.PublicKey,
		Page:         1,
		Limit:        10,
		Ascending:    true,
	})
	if err != nil {
		t.Fatalf("ListPayments: %v", err)
	}
	require.NotEmpty(t, listPaymentsOut.States)

	var payments []paymentV1Models.PaymentStateModel
	err = utils.UnmarshalState[[]paymentV1Models.PaymentStateModel](listPaymentsOut.States[0].Object, &payments)
	if err != nil {
		t.Fatalf("UnmarshalState (ListPayments.States[0]): %v", err)
	}

	require.NotEmpty(t, payments)

	var foundCreatedFlow bool
	var foundVoidFlow bool
	var foundDirectFlow bool

	for _, p := range payments {
		if p.Address == paymentAddress {
			foundCreatedFlow = true
			assert.Equal(t, orderId, p.OrderId)
			assert.Equal(t, payer.PublicKey, p.Payer)
			assert.Equal(t, payee.PublicKey, p.Payee)
			assert.Equal(t, amount, p.Amount)
			assert.Equal(t, paymentV1Domain.STATUS_REFUNDED, p.Status)
			assert.Equal(t, amount, p.CapturedAmount)
			assert.Equal(t, refundAmount, p.RefundedAmount)
			assert.NotZero(t, p.CreatedAt)
			assert.NotZero(t, p.UpdatedAt)
		}

		if p.Address == voidPaymentAddress {
			foundVoidFlow = true
			assert.Equal(t, voidOrderId, p.OrderId)
			assert.Equal(t, payer.PublicKey, p.Payer)
			assert.Equal(t, payee.PublicKey, p.Payee)
			assert.Equal(t, voidAmount, p.Amount)
			assert.Equal(t, paymentV1Domain.STATUS_VOIDED, p.Status)
			assert.NotZero(t, p.CreatedAt)
			assert.NotZero(t, p.UpdatedAt)
		}

		if p.Address == directPaymentAddress {
			foundDirectFlow = true
			assert.Equal(t, directPayOrderId, p.OrderId)
			assert.Equal(t, payer.PublicKey, p.Payer)
			assert.Equal(t, payee.PublicKey, p.Payee)
			assert.Equal(t, directPayAmount, p.Amount)
			assert.Equal(t, directPayAmount, p.CapturedAmount)
			assert.Equal(t, "0", p.RefundedAmount)
			assert.Equal(t, paymentV1Domain.STATUS_CAPTURED, p.Status)
			assert.NotZero(t, p.CreatedAt)
			assert.NotZero(t, p.UpdatedAt)
		}
	}

	assert.True(t, foundCreatedFlow, "expected to find refunded payment in ListPayments")
	assert.True(t, foundVoidFlow, "expected to find voided payment in ListPayments")
	assert.True(t, foundDirectFlow, "expected to find direct payment in ListPayments")
}

// More payment scenarios
func TestPaymentAuthVoidFlow(t *testing.T) {
	c := setupClient(t)

	// ------------------
	//      WALLETS
	// ------------------
	owner, ownerPriv := createWallet(t, c)
	payer, payerPriv := createWallet(t, c)
	payee, _ := createWallet(t, c)

	// ------------------
	//      TOKEN
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

	require.Equal(t, tokenV1Domain.FUNGIBLE, payToken.TokenType, "payToken must be fungible")

	// ------------------
	//   DEPLOY PAYMENT
	// ------------------
	deployedContract, err := c.DeployContract1(paymentV1.PAYMENT_CONTRACT_V1)
	if err != nil {
		t.Fatalf("DeployContract: %v", err)
	}
	require.NotEmpty(t, deployedContract.Logs)

	deployLog, err := utils.UnmarshalLog[log.Log](deployedContract.Logs[0])
	if err != nil {
		t.Fatalf("UnmarshalLog (DeployContract.Logs[0]): %v", err)
	}

	paymentAddress := deployLog.ContractAddress
	require.NotEmpty(t, paymentAddress)

	// ------------------
	//   ALLOW USERS
	// ------------------
	c.SetPrivateKey(ownerPriv)

	_, err = c.AddAllowedUsers(payToken.Address, map[string]bool{
		owner.PublicKey: true,
		payer.PublicKey: true,
		payee.PublicKey: true,
		paymentAddress:  true,
	})
	if err != nil {
		t.Fatalf("AddAllowedUsers: %v", err)
	}

	// ------------------
	//   FUND PAYER
	// ------------------
	fundAmount := "500"
	_, err = c.TransferToken(payToken.Address, payer.PublicKey, fundAmount, []string{})
	if err != nil {
		t.Fatalf("TransferToken payer funding: %v", err)
	}

	payerBalanceBeforeOut, err := c.GetTokenBalance(payToken.Address, payer.PublicKey)
	if err != nil {
		t.Fatalf("GetTokenBalance payer before: %v", err)
	}
	var payerBalanceBefore tokenV1Models.BalanceStateModel
	err = utils.UnmarshalState[tokenV1Models.BalanceStateModel](payerBalanceBeforeOut.States[0].Object, &payerBalanceBefore)
	if err != nil {
		t.Fatalf("UnmarshalState payerBalanceBefore: %v", err)
	}

	var payeeBalanceBefore tokenV1Models.BalanceStateModel
	payeeBalanceBeforeOut, err := c.GetTokenBalance(payToken.Address, payee.PublicKey)
	if err != nil {
		require.Contains(t, err.Error(), "record not found")
		payeeBalanceBefore = tokenV1Models.BalanceStateModel{
			TokenAddress: payToken.Address,
			OwnerAddress: payee.PublicKey,
			Amount:       "0",
			TokenType:    tokenV1Domain.FUNGIBLE,
		}
	} else {
		err = utils.UnmarshalState[tokenV1Models.BalanceStateModel](payeeBalanceBeforeOut.States[0].Object, &payeeBalanceBefore)
		if err != nil {
			t.Fatalf("UnmarshalState payeeBalanceBefore: %v", err)
		}
	}

	// ------------------
	//   CREATE PAYMENT
	// ------------------
	orderId := "order-payment-auth-void-e2e-001"
	amount := "300"
	expiredAt := time.Now().Add(2 * time.Hour)

	c.SetPrivateKey(ownerPriv)
	createPaymentOut, err := c.CreatePayment(inputs.InputCreate{
		Address:      paymentAddress,
		Owner:        owner.PublicKey,
		TokenAddress: payToken.Address,
		OrderId:      orderId,
		Payer:        payer.PublicKey,
		Payee:        payee.PublicKey,
		Amount:       amount,
		ExpiredAt:    expiredAt,
	})
	if err != nil {
		t.Fatalf("CreatePayment: %v", err)
	}
	require.NotEmpty(t, createPaymentOut.Logs)

	createPaymentLog, err := utils.UnmarshalLog[log.Log](createPaymentOut.Logs[0])
	if err != nil {
		t.Fatalf("UnmarshalLog (CreatePayment.Logs[0]): %v", err)
	}
	assert.Equal(t, paymentV1Domain.PAYMENT_CREATED_LOG, createPaymentLog.LogType)

	createPaymentEvent, err := utils.UnmarshalEvent[paymentV1Domain.Payment](createPaymentLog.Event)
	if err != nil {
		t.Fatalf("UnmarshalEvent (CreatePayment.Logs[0]): %v", err)
	}

	assert.Equal(t, paymentAddress, createPaymentEvent.Address, "payment address mismatch")
	assert.Equal(t, owner.PublicKey, createPaymentEvent.Owner, "payment owner mismatch")
	assert.Equal(t, payToken.Address, createPaymentEvent.TokenAddress, "payment token address mismatch")
	assert.Equal(t, orderId, createPaymentEvent.OrderId, "payment order id mismatch")
	assert.Equal(t, payer.PublicKey, createPaymentEvent.Payer, "payment payer mismatch")
	assert.Equal(t, payee.PublicKey, createPaymentEvent.Payee, "payment payee mismatch")
	assert.Equal(t, amount, createPaymentEvent.Amount, "payment amount mismatch")
	assert.Equal(t, paymentV1Domain.STATUS_CREATED, createPaymentEvent.Status, "payment status mismatch after create")
	assert.False(t, createPaymentEvent.Paused, "payment paused mismatch after create")

	getPaymentOut, err := c.GetPayment(paymentAddress)
	if err != nil {
		t.Fatalf("GetPayment after create: %v", err)
	}
	require.NotEmpty(t, getPaymentOut.States)

	var paymentState paymentV1Models.PaymentStateModel
	err = utils.UnmarshalState[paymentV1Models.PaymentStateModel](getPaymentOut.States[0].Object, &paymentState)
	if err != nil {
		t.Fatalf("UnmarshalState (GetPayment after create): %v", err)
	}

	assert.Equal(t, paymentAddress, paymentState.Address, "payment state address mismatch after create")
	assert.Equal(t, owner.PublicKey, paymentState.Owner, "payment state owner mismatch after create")
	assert.Equal(t, payToken.Address, paymentState.TokenAddress, "payment state token address mismatch after create")
	assert.Equal(t, orderId, paymentState.OrderId, "payment state order id mismatch after create")
	assert.Equal(t, payer.PublicKey, paymentState.Payer, "payment state payer mismatch after create")
	assert.Equal(t, payee.PublicKey, paymentState.Payee, "payment state payee mismatch after create")
	assert.Equal(t, amount, paymentState.Amount, "payment state amount mismatch after create")
	assert.Equal(t, paymentV1Domain.STATUS_CREATED, paymentState.Status, "payment state status mismatch after create")
	assert.False(t, paymentState.Paused, "payment state paused mismatch after create")
	assert.NotEmpty(t, paymentState.Hash, "payment hash should not be empty after create")

	// ------------------
	//     AUTHORIZE
	// ------------------
	c.SetPrivateKey(payerPriv)
	authorizeOut, err := c.AuthorizePayment(inputs.InputAuthorize{
		Address: paymentAddress,
	})
	if err != nil {
		t.Fatalf("AuthorizePayment: %v", err)
	}
	require.NotEmpty(t, authorizeOut.Logs)

	authorizeLog, err := utils.UnmarshalLog[log.Log](authorizeOut.Logs[0])
	if err != nil {
		t.Fatalf("UnmarshalLog (AuthorizePayment.Logs[0]): %v", err)
	}
	assert.Equal(t, paymentV1Domain.PAYMENT_AUTHORIZED_LOG, authorizeLog.LogType)

	authorizeEvent, err := utils.UnmarshalEvent[paymentV1Domain.Payment](authorizeLog.Event)
	if err != nil {
		t.Fatalf("UnmarshalEvent (AuthorizePayment.Logs[0]): %v", err)
	}

	assert.Equal(t, paymentAddress, authorizeEvent.Address, "payment address mismatch after authorize")
	assert.Equal(t, paymentV1Domain.STATUS_AUTHORIZED, authorizeEvent.Status, "payment status mismatch after authorize")

	getPaymentOut, err = c.GetPayment(paymentAddress)
	if err != nil {
		t.Fatalf("GetPayment after authorize: %v", err)
	}
	require.NotEmpty(t, getPaymentOut.States)

	err = utils.UnmarshalState[paymentV1Models.PaymentStateModel](getPaymentOut.States[0].Object, &paymentState)
	if err != nil {
		t.Fatalf("UnmarshalState (GetPayment after authorize): %v", err)
	}

	assert.Equal(t, paymentAddress, paymentState.Address, "payment state address mismatch after authorize")
	assert.Equal(t, paymentV1Domain.STATUS_AUTHORIZED, paymentState.Status, "payment state status mismatch after authorize")
	assert.Equal(t, amount, paymentState.Amount, "payment state amount mismatch after authorize")
	assert.False(t, paymentState.Paused, "payment state paused mismatch after authorize")

	// ------------------
	//        VOID
	// ------------------
	c.SetPrivateKey(payerPriv)
	voidOut, err := c.VoidPayment(inputs.InputVoidPayment{
		Address: paymentAddress,
	})
	if err != nil {
		t.Fatalf("VoidPayment: %v", err)
	}
	require.NotEmpty(t, voidOut.Logs)

	voidLog, err := utils.UnmarshalLog[log.Log](voidOut.Logs[0])
	if err != nil {
		t.Fatalf("UnmarshalLog (VoidPayment.Logs[0]): %v", err)
	}
	assert.Equal(t, paymentV1Domain.PAYMENT_VOIDED_LOG, voidLog.LogType)

	voidEvent, err := utils.UnmarshalEvent[paymentV1Domain.Payment](voidLog.Event)
	if err != nil {
		t.Fatalf("UnmarshalEvent (VoidPayment.Logs[0]): %v", err)
	}

	assert.Equal(t, paymentAddress, voidEvent.Address, "payment address mismatch after void")
	assert.Equal(t, paymentV1Domain.STATUS_VOIDED, voidEvent.Status, "payment status mismatch after void")

	getPaymentOut, err = c.GetPayment(paymentAddress)
	if err != nil {
		t.Fatalf("GetPayment after void: %v", err)
	}
	require.NotEmpty(t, getPaymentOut.States)

	err = utils.UnmarshalState[paymentV1Models.PaymentStateModel](getPaymentOut.States[0].Object, &paymentState)
	if err != nil {
		t.Fatalf("UnmarshalState (GetPayment after void): %v", err)
	}

	assert.Equal(t, paymentAddress, paymentState.Address, "payment state address mismatch after void")
	assert.Equal(t, owner.PublicKey, paymentState.Owner, "payment state owner mismatch after void")
	assert.Equal(t, payToken.Address, paymentState.TokenAddress, "payment state token address mismatch after void")
	assert.Equal(t, orderId, paymentState.OrderId, "payment state order id mismatch after void")
	assert.Equal(t, payer.PublicKey, paymentState.Payer, "payment state payer mismatch after void")
	assert.Equal(t, payee.PublicKey, paymentState.Payee, "payment state payee mismatch after void")
	assert.Equal(t, amount, paymentState.Amount, "payment state amount mismatch after void")
	assert.Equal(t, paymentV1Domain.STATUS_VOIDED, paymentState.Status, "payment state status mismatch after void")
	assert.False(t, paymentState.Paused, "payment state paused mismatch after void")
	assert.NotEmpty(t, paymentState.Hash, "payment hash should not be empty after void")

	// ------------------
	//    BALANCE CHECK
	// ------------------
	payerBalanceAfterVoidOut, err := c.GetTokenBalance(payToken.Address, payer.PublicKey)
	if err != nil {
		t.Fatalf("GetTokenBalance payer after void: %v", err)
	}
	var payerBalanceAfterVoid tokenV1Models.BalanceStateModel
	err = utils.UnmarshalState[tokenV1Models.BalanceStateModel](payerBalanceAfterVoidOut.States[0].Object, &payerBalanceAfterVoid)
	if err != nil {
		t.Fatalf("UnmarshalState payerBalanceAfterVoid: %v", err)
	}

	var payeeBalanceAfterVoid tokenV1Models.BalanceStateModel
	payeeBalanceAfterVoidOut, err := c.GetTokenBalance(payToken.Address, payee.PublicKey)
	if err != nil {
		require.Contains(t, err.Error(), "record not found")
		payeeBalanceAfterVoid = tokenV1Models.BalanceStateModel{
			TokenAddress: payToken.Address,
			OwnerAddress: payee.PublicKey,
			Amount:       "0",
			TokenType:    tokenV1Domain.FUNGIBLE,
		}
	} else {
		err = utils.UnmarshalState[tokenV1Models.BalanceStateModel](payeeBalanceAfterVoidOut.States[0].Object, &payeeBalanceAfterVoid)
		if err != nil {
			t.Fatalf("UnmarshalState payeeBalanceAfterVoid: %v", err)
		}
	}

	assert.Equal(t, payerBalanceBefore.Amount, payerBalanceAfterVoid.Amount, "payer balance should not change after auth+void flow")
	assert.Equal(t, payeeBalanceBefore.Amount, payeeBalanceAfterVoid.Amount, "payee balance should not change after auth+void flow")

	// ------------------
	//    LIST PAYMENTS
	// ------------------
	listPaymentsOut, err := c.ListPayments(inputs.InputList{
		OrderId:      orderId,
		TokenAddress: payToken.Address,
		Status:       []string{paymentV1Domain.STATUS_VOIDED},
		Payer:        payer.PublicKey,
		Payee:        payee.PublicKey,
		Page:         1,
		Limit:        10,
		Ascending:    true,
	})
	if err != nil {
		t.Fatalf("ListPayments: %v", err)
	}
	require.NotEmpty(t, listPaymentsOut.States)

	var payments []paymentV1Models.PaymentStateModel
	err = utils.UnmarshalState[[]paymentV1Models.PaymentStateModel](listPaymentsOut.States[0].Object, &payments)
	if err != nil {
		t.Fatalf("UnmarshalState (ListPayments.States[0]): %v", err)
	}

	require.NotEmpty(t, payments, "expected at least one payment in list")

	var found bool
	for _, p := range payments {
		if p.Address == paymentAddress {
			found = true
			assert.Equal(t, owner.PublicKey, p.Owner, "listed payment owner mismatch")
			assert.Equal(t, payToken.Address, p.TokenAddress, "listed payment token address mismatch")
			assert.Equal(t, orderId, p.OrderId, "listed payment order id mismatch")
			assert.Equal(t, payer.PublicKey, p.Payer, "listed payment payer mismatch")
			assert.Equal(t, payee.PublicKey, p.Payee, "listed payment payee mismatch")
			assert.Equal(t, amount, p.Amount, "listed payment amount mismatch")
			assert.Equal(t, paymentV1Domain.STATUS_VOIDED, p.Status, "listed payment status mismatch")
			assert.NotZero(t, p.CreatedAt, "listed payment createdAt should not be zero")
			assert.NotZero(t, p.UpdatedAt, "listed payment updatedAt should not be zero")
			break
		}
	}

	assert.True(t, found, "expected to find voided payment in ListPayments")
}
