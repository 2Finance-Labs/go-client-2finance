package e2e_test

import (
	"fmt"
	"log"
	"testing"
	"time"

	"gitlab.com/2finance/2finance-network/blockchain/contract/contractV1/models"
	paymentV1Models "gitlab.com/2finance/2finance-network/blockchain/contract/paymentV1/models"
	"gitlab.com/2finance/2finance-network/blockchain/contract/paymentV1"
	tokenV1Domain "gitlab.com/2finance/2finance-network/blockchain/contract/tokenV1/domain"
)

func TestPaymentFlow(t *testing.T) {
	c := setupClient(t)

	// --------------------------------------------------------------------
	// Owner + base token
	// --------------------------------------------------------------------
	owner, ownerPriv := createWallet(t, c)
	c.SetPrivateKey(ownerPriv)

	dec := 6
	stablecoin := false
	tok := createBasicToken(t, c, owner.PublicKey, dec, true, tokenV1Domain.FUNGIBLE, stablecoin)
	_ = createMint(t, c, tok, owner.PublicKey, "10000", dec, tok.TokenType)

	// Token (mínimo) validate + log
	if tok.Address == "" {
		t.Fatalf("Token address empty")
	}
	if tok.Creator == "" {
		t.Fatalf("Token creator empty")
	}
	if tok.Decimals != dec {
		t.Fatalf("Token decimals mismatch: got %d want %d", tok.Decimals, dec)
	}
	if tok.TokenType != tokenV1Domain.FUNGIBLE {
		t.Fatalf("Token type mismatch: got %s want %s", tok.TokenType, tokenV1Domain.FUNGIBLE)
	}
	if tok.Stablecoin != stablecoin {
		t.Fatalf("Token stablecoin mismatch: got %v want %v", tok.Stablecoin, stablecoin)
	}

	log.Printf("Token Address: %s", tok.Address)
	log.Printf("Token Symbol: %s", tok.Symbol)
	log.Printf("Token Name: %s", tok.Name)
	log.Printf("Token Decimals: %d", tok.Decimals)
	log.Printf("Token Total Supply: %s", tok.TotalSupply)
	log.Printf("Token Type: %s", tok.TokenType)
	log.Printf("Token Stablecoin: %v", tok.Stablecoin)

	// --------------------------------------------------------------------
	// Payer / Payee setup
	// --------------------------------------------------------------------
	payer, payerPriv := createWallet(t, c)
	payee, payeePriv := createWallet(t, c)

	c.SetPrivateKey(ownerPriv)
	allowPayerOut, err := c.AllowUsers(tok.Address, map[string]bool{payer.PublicKey: true})
	if err != nil {
		t.Fatalf("AllowUsers(payer): %v", err)
	}
	if len(allowPayerOut.States) == 0 || allowPayerOut.States[0].Object == nil {
		t.Fatalf("AllowUsers(payer) returned empty/nil state")
	}

	allowPayeeOut, err := c.AllowUsers(tok.Address, map[string]bool{payee.PublicKey: true})
	if err != nil {
		t.Fatalf("AllowUsers(payee): %v", err)
	}
	if len(allowPayeeOut.States) == 0 || allowPayeeOut.States[0].Object == nil {
		t.Fatalf("AllowUsers(payee) returned empty/nil state")
	}

	_ = createTransfer(t, c, tok, payer.PublicKey, "50", tok.Decimals, tok.TokenType, "")

	// --------------------------------------------------------------------
	// Deploy Payment contract
	// --------------------------------------------------------------------
	contractState := models.ContractStateModel{}
	deployedContract, err := c.DeployContract1(paymentV1.PAYMENT_CONTRACT_V1)
	if err != nil {
		t.Fatalf("DeployContract(Payment): %v", err)
	}
	if len(deployedContract.States) == 0 {
		t.Fatalf("DeployContract(Payment) returned empty States")
	}
	if deployedContract.States[0].Object == nil {
		t.Fatalf("DeployContract(Payment) returned nil state object")
	}

	unmarshalState(t, deployedContract.States[0].Object, &contractState)
	address := contractState.Address
	if address == "" {
		t.Fatalf("DeployContract(Payment) returned empty contract address")
	}

	log.Printf("DeployContract(Payment) Output States: %+v", deployedContract.States)
	log.Printf("DeployContract(Payment) Output Logs: %+v", deployedContract.Logs)
	log.Printf("DeployContract(Payment) Output Delegated Call: %+v", deployedContract.DelegatedCall)
	log.Printf("Payment Contract Address: %s", address)

	// --------------------------------------------------------------------
	// Create Payment (payer)
	// --------------------------------------------------------------------
	orderID := fmt.Sprintf("order-%d", time.Now().Unix())
	amount := "10"
	exp := time.Now().Add(30 * time.Minute)

	c.SetPrivateKey(payerPriv)

	created, err := c.CreatePayment(
		address,
		tok.Address,
		orderID,
		payer.PublicKey,
		payee.PublicKey,
		amount,
		exp,
	)
	if err != nil {
		t.Fatalf("CreatePayment: %v", err)
	}
	if len(created.States) == 0 {
		t.Fatalf("CreatePayment returned empty States")
	}
	if created.States[0].Object == nil {
		t.Fatalf("CreatePayment returned nil state object")
	}

	var pay paymentV1Models.PaymentStateModel
	unmarshalState(t, created.States[0].Object, &pay)

	// Field validation (determinísticos)
	if pay.Address == "" {
		t.Fatalf("CreatePayment returned empty payment address")
	}
	if pay.TokenAddress != tok.Address {
		t.Fatalf("CreatePayment TokenAddress mismatch: got %q want %q", pay.TokenAddress, tok.Address)
	}
	if pay.OrderId != orderID {
		t.Fatalf("CreatePayment OrderId mismatch: got %q want %q", pay.OrderId, orderID)
	}
	if pay.Payer != payer.PublicKey {
		t.Fatalf("CreatePayment Payer mismatch: got %q want %q", pay.Payer, payer.PublicKey)
	}
	if pay.Payee != payee.PublicKey {
		t.Fatalf("CreatePayment Payee mismatch: got %q want %q", pay.Payee, payee.PublicKey)
	}
	if pay.Amount != amount {
		t.Fatalf("CreatePayment Amount mismatch: got %q want %q", pay.Amount, amount)
	}
	if pay.Hash == "" {
		t.Fatalf("CreatePayment Hash empty")
	}

	log.Printf("CreatePayment Output States: %+v", created.States)
	log.Printf("CreatePayment Output Logs: %+v", created.Logs)
	log.Printf("CreatePayment Output Delegated Call: %+v", created.DelegatedCall)

	log.Printf("CreatePayment Address: %s", pay.Address)
	log.Printf("CreatePayment Owner: %s", pay.Owner)
	log.Printf("CreatePayment TokenAddress: %s", pay.TokenAddress)
	log.Printf("CreatePayment OrderId: %s", pay.OrderId)
	log.Printf("CreatePayment Payer: %s", pay.Payer)
	log.Printf("CreatePayment Payee: %s", pay.Payee)
	log.Printf("CreatePayment Amount: %s", pay.Amount)
	log.Printf("CreatePayment CapturedAmount: %s", pay.CapturedAmount)
	log.Printf("CreatePayment RefundedAmount: %s", pay.RefundedAmount)
	log.Printf("CreatePayment Status: %s", pay.Status)
	log.Printf("CreatePayment Paused: %v", pay.Paused)
	log.Printf("CreatePayment ExpiredAt: %s", pay.ExpiredAt.String())
	log.Printf("CreatePayment Hash: %s", pay.Hash)

	// --------------------------------------------------------------------
	// Allow payment contract to move token (owner)
	// --------------------------------------------------------------------
	c.SetPrivateKey(ownerPriv)
	allowPaymentOut, err := c.AllowUsers(tok.Address, map[string]bool{pay.Address: true})
	if err != nil {
		t.Fatalf("AllowUsers(payment): %v", err)
	}
	if len(allowPaymentOut.States) == 0 || allowPaymentOut.States[0].Object == nil {
		t.Fatalf("AllowUsers(payment) returned empty/nil state")
	}

	var ap tokenV1Domain.AccessPolicy
	unmarshalState(t, allowPaymentOut.States[0].Object, &ap)
	if ap.Users == nil || !ap.Users[pay.Address] {
		t.Fatalf("AllowUsers(payment) missing payment address in allowlist")
	}

	log.Printf("AllowUsers(payment) Output States: %+v", allowPaymentOut.States)
	log.Printf("AllowUsers(payment) Output Logs: %+v", allowPaymentOut.Logs)
	log.Printf("AllowUsers(payment) Output Delegated Call: %+v", allowPaymentOut.DelegatedCall)

	// --------------------------------------------------------------------
	// Authorize (payer)
	// --------------------------------------------------------------------
	c.SetPrivateKey(payerPriv)

	authOut, err := c.AuthorizePayment(pay.Address, tokenV1Domain.FUNGIBLE, "")
	if err != nil {
		t.Fatalf("AuthorizePayment: %v", err)
	}
	if len(authOut.States) == 0 || authOut.States[0].Object == nil {
		t.Fatalf("AuthorizePayment returned empty/nil state")
	}

	var payAuth paymentV1Models.PaymentStateModel
	unmarshalState(t, authOut.States[0].Object, &payAuth)

	if payAuth.Address != pay.Address {
		t.Fatalf("AuthorizePayment Address mismatch: got %q want %q", payAuth.Address, pay.Address)
	}
	// se seu backend define status determinístico aqui, valide:
	// if payAuth.Status != "authorized" { ... }
	if payAuth.Hash == "" {
		t.Fatalf("AuthorizePayment Hash empty")
	}

	log.Printf("AuthorizePayment Output States: %+v", authOut.States)
	log.Printf("AuthorizePayment Output Logs: %+v", authOut.Logs)
	log.Printf("AuthorizePayment Output Delegated Call: %+v", authOut.DelegatedCall)

	log.Printf("AuthorizePayment Address: %s", payAuth.Address)
	log.Printf("AuthorizePayment Status: %s", payAuth.Status)
	log.Printf("AuthorizePayment CapturedAmount: %s", payAuth.CapturedAmount)
	log.Printf("AuthorizePayment RefundedAmount: %s", payAuth.RefundedAmount)
	log.Printf("AuthorizePayment Paused: %v", payAuth.Paused)
	log.Printf("AuthorizePayment Hash: %s", payAuth.Hash)

	// --------------------------------------------------------------------
	// Capture + Refund (payee)
	// --------------------------------------------------------------------
	c.SetPrivateKey(payeePriv)

	capOut, err := c.CapturePayment(pay.Address, tokenV1Domain.FUNGIBLE, "")
	if err != nil {
		t.Fatalf("CapturePayment: %v", err)
	}
	if len(capOut.States) == 0 || capOut.States[0].Object == nil {
		t.Fatalf("CapturePayment returned empty/nil state")
	}

	var payCap paymentV1Models.PaymentStateModel
	unmarshalState(t, capOut.States[0].Object, &payCap)

	if payCap.Address != pay.Address {
		t.Fatalf("CapturePayment Address mismatch: got %q want %q", payCap.Address, pay.Address)
	}
	if payCap.Hash == "" {
		t.Fatalf("CapturePayment Hash empty")
	}

	log.Printf("CapturePayment Output States: %+v", capOut.States)
	log.Printf("CapturePayment Output Logs: %+v", capOut.Logs)
	log.Printf("CapturePayment Output Delegated Call: %+v", capOut.DelegatedCall)

	log.Printf("CapturePayment Address: %s", payCap.Address)
	log.Printf("CapturePayment Status: %s", payCap.Status)
	log.Printf("CapturePayment CapturedAmount: %s", payCap.CapturedAmount)
	log.Printf("CapturePayment RefundedAmount: %s", payCap.RefundedAmount)
	log.Printf("CapturePayment Hash: %s", payCap.Hash)

	refOut, err := c.RefundPayment(pay.Address, "10", tokenV1Domain.FUNGIBLE, "")
	if err != nil {
		t.Fatalf("RefundPayment: %v", err)
	}
	if len(refOut.States) == 0 || refOut.States[0].Object == nil {
		t.Fatalf("RefundPayment returned empty/nil state")
	}

	var payRef paymentV1Models.PaymentStateModel
	unmarshalState(t, refOut.States[0].Object, &payRef)

	if payRef.Address != pay.Address {
		t.Fatalf("RefundPayment Address mismatch: got %q want %q", payRef.Address, pay.Address)
	}
	if payRef.Hash == "" {
		t.Fatalf("RefundPayment Hash empty")
	}

	log.Printf("RefundPayment Output States: %+v", refOut.States)
	log.Printf("RefundPayment Output Logs: %+v", refOut.Logs)
	log.Printf("RefundPayment Output Delegated Call: %+v", refOut.DelegatedCall)

	log.Printf("RefundPayment Address: %s", payRef.Address)
	log.Printf("RefundPayment Status: %s", payRef.Status)
	log.Printf("RefundPayment CapturedAmount: %s", payRef.CapturedAmount)
	log.Printf("RefundPayment RefundedAmount: %s", payRef.RefundedAmount)
	log.Printf("RefundPayment Hash: %s", payRef.Hash)

	// --------------------------------------------------------------------
	// DirectPay (novo contract) - valida envelope e log
	// --------------------------------------------------------------------
	contractState = models.ContractStateModel{}
	deployedContract2, err := c.DeployContract1(paymentV1.PAYMENT_CONTRACT_V1)
	if err != nil {
		t.Fatalf("DeployContract(Payment #2): %v", err)
	}
	if len(deployedContract2.States) == 0 || deployedContract2.States[0].Object == nil {
		t.Fatalf("DeployContract(Payment #2) returned empty/nil state")
	}
	unmarshalState(t, deployedContract2.States[0].Object, &contractState)
	address2 := contractState.Address
	if address2 == "" {
		t.Fatalf("DeployContract(Payment #2) returned empty address")
	}

	log.Printf("DeployContract(Payment #2) Output States: %+v", deployedContract2.States)
	log.Printf("DeployContract(Payment #2) Output Logs: %+v", deployedContract2.Logs)
	log.Printf("DeployContract(Payment #2) Output Delegated Call: %+v", deployedContract2.DelegatedCall)
	log.Printf("Payment Contract #2 Address: %s", address2)

	c.SetPrivateKey(payerPriv)
	directOut, err := c.DirectPay(
		address2,
		tok.Address,
		orderID+"-direct",
		payer.PublicKey,
		payee.PublicKey,
		amt(2, dec),
		tokenV1Domain.FUNGIBLE,
		"",
	)
	if err != nil {
		t.Fatalf("DirectPay: %v", err)
	}
	if len(directOut.States) == 0 || directOut.States[0].Object == nil {
		t.Fatalf("DirectPay returned empty/nil state")
	}

	var payDirect paymentV1Models.PaymentStateModel
	unmarshalState(t, directOut.States[0].Object, &payDirect)
	if payDirect.Address == "" {
		t.Fatalf("DirectPay returned empty payment address")
	}

	log.Printf("DirectPay Output States: %+v", directOut.States)
	log.Printf("DirectPay Output Logs: %+v", directOut.Logs)
	log.Printf("DirectPay Output Delegated Call: %+v", directOut.DelegatedCall)

	log.Printf("DirectPay Address: %s", payDirect.Address)
	log.Printf("DirectPay OrderId: %s", payDirect.OrderId)
	log.Printf("DirectPay Amount: %s", payDirect.Amount)
	log.Printf("DirectPay Status: %s", payDirect.Status)

	// --------------------------------------------------------------------
	// Pause / Unpause (payer) - valida + log
	// --------------------------------------------------------------------
	c.SetPrivateKey(payerPriv)

	pauseOut, err := c.PausePayment(pay.Address, true)
	if err != nil {
		t.Fatalf("PausePayment(payer): %v", err)
	}
	if len(pauseOut.States) == 0 || pauseOut.States[0].Object == nil {
		t.Fatalf("PausePayment(payer) returned empty/nil state")
	}
	var payPaused paymentV1Models.PaymentStateModel
	unmarshalState(t, pauseOut.States[0].Object, &payPaused)
	if payPaused.Address != pay.Address {
		t.Fatalf("PausePayment address mismatch: got %q want %q", payPaused.Address, pay.Address)
	}
	if !payPaused.Paused {
		t.Fatalf("PausePayment expected Paused=true")
	}

	log.Printf("PausePayment(payer) Output States: %+v", pauseOut.States)
	log.Printf("PausePayment(payer) Output Logs: %+v", pauseOut.Logs)
	log.Printf("PausePayment(payer) Output Delegated Call: %+v", pauseOut.DelegatedCall)
	log.Printf("PausePayment Address: %s", payPaused.Address)
	log.Printf("PausePayment Paused: %v", payPaused.Paused)

	unpauseOut, err := c.UnpausePayment(pay.Address, false)
	if err != nil {
		t.Fatalf("UnpausePayment(payer): %v", err)
	}
	if len(unpauseOut.States) == 0 || unpauseOut.States[0].Object == nil {
		t.Fatalf("UnpausePayment(payer) returned empty/nil state")
	}
	var payUnpaused paymentV1Models.PaymentStateModel
	unmarshalState(t, unpauseOut.States[0].Object, &payUnpaused)
	if payUnpaused.Address != pay.Address {
		t.Fatalf("UnpausePayment address mismatch: got %q want %q", payUnpaused.Address, pay.Address)
	}
	if payUnpaused.Paused {
		t.Fatalf("UnpausePayment expected Paused=false")
	}

	log.Printf("UnpausePayment(payer) Output States: %+v", unpauseOut.States)
	log.Printf("UnpausePayment(payer) Output Logs: %+v", unpauseOut.Logs)
	log.Printf("UnpausePayment(payer) Output Delegated Call: %+v", unpauseOut.DelegatedCall)
	log.Printf("UnpausePayment Address: %s", payUnpaused.Address)
	log.Printf("UnpausePayment Paused: %v", payUnpaused.Paused)

	// --------------------------------------------------------------------
	// Pause / Unpause (payee) - valida + log
	// --------------------------------------------------------------------
	c.SetPrivateKey(payeePriv)

	pauseOut2, err := c.PausePayment(pay.Address, true)
	if err != nil {
		t.Fatalf("PausePayment(payee): %v", err)
	}
	if len(pauseOut2.States) == 0 || pauseOut2.States[0].Object == nil {
		t.Fatalf("PausePayment(payee) returned empty/nil state")
	}

	var payPaused2 paymentV1Models.PaymentStateModel
	unmarshalState(t, pauseOut2.States[0].Object, &payPaused2)
	if payPaused2.Address != pay.Address {
		t.Fatalf("PausePayment(payee) address mismatch: got %q want %q", payPaused2.Address, pay.Address)
	}
	if !payPaused2.Paused {
		t.Fatalf("PausePayment(payee) expected Paused=true")
	}

	log.Printf("PausePayment(payee) Output States: %+v", pauseOut2.States)
	log.Printf("PausePayment(payee) Output Logs: %+v", pauseOut2.Logs)
	log.Printf("PausePayment(payee) Output Delegated Call: %+v", pauseOut2.DelegatedCall)

	unpauseOut2, err := c.UnpausePayment(pay.Address, false)
	if err != nil {
		t.Fatalf("UnpausePayment(payee): %v", err)
	}
	if len(unpauseOut2.States) == 0 || unpauseOut2.States[0].Object == nil {
		t.Fatalf("UnpausePayment(payee) returned empty/nil state")
	}

	var payUnpaused2 paymentV1Models.PaymentStateModel
	unmarshalState(t, unpauseOut2.States[0].Object, &payUnpaused2)
	if payUnpaused2.Address != pay.Address {
		t.Fatalf("UnpausePayment(payee) address mismatch: got %q want %q", payUnpaused2.Address, pay.Address)
	}
	if payUnpaused2.Paused {
		t.Fatalf("UnpausePayment(payee) expected Paused=false")
	}

	log.Printf("UnpausePayment(payee) Output States: %+v", unpauseOut2.States)
	log.Printf("UnpausePayment(payee) Output Logs: %+v", unpauseOut2.Logs)
	log.Printf("UnpausePayment(payee) Output Delegated Call: %+v", unpauseOut2.DelegatedCall)

	// --------------------------------------------------------------------
	// GetPayment (reader) - valida + log
	// --------------------------------------------------------------------
	getOut, err := c.GetPayment(pay.Address)
	if err != nil {
		t.Fatalf("GetPayment: %v", err)
	}
	if len(getOut.States) == 0 || getOut.States[0].Object == nil {
		t.Fatalf("GetPayment returned empty/nil state")
	}

	var got paymentV1Models.PaymentStateModel
	unmarshalState(t, getOut.States[0].Object, &got)
	if got.Address != pay.Address {
		t.Fatalf("GetPayment address mismatch: got %q want %q", got.Address, pay.Address)
	}

	log.Printf("GetPayment Output States: %+v", getOut.States)
	log.Printf("GetPayment Output Logs: %+v", getOut.Logs)
	log.Printf("GetPayment Output Delegated Call: %+v", getOut.DelegatedCall)

	log.Printf("GetPayment Address: %s", got.Address)
	log.Printf("GetPayment Status: %s", got.Status)
	log.Printf("GetPayment CapturedAmount: %s", got.CapturedAmount)
	log.Printf("GetPayment RefundedAmount: %s", got.RefundedAmount)
	log.Printf("GetPayment Paused: %v", got.Paused)
}

func TestPaymentAuthVoidFlow(t *testing.T) {
	c := setupClient(t)

	owner, ownerPriv := createWallet(t, c)
	c.SetPrivateKey(ownerPriv)

	dec := 6
	stablecoin := false

	tok := createBasicToken(t, c, owner.PublicKey, dec, true, tokenV1Domain.FUNGIBLE, stablecoin)

	// Token (mínimo) validate
	if tok.Address == "" {
		t.Fatalf("Token address empty")
	}

	payer, payerPriv := createWallet(t, c)
	payee, _ := createWallet(t, c)

	c.SetPrivateKey(ownerPriv)
	allowPayerOut, err := c.AllowUsers(tok.Address, map[string]bool{payer.PublicKey: true})
	if err != nil {
		t.Fatalf("AllowUsers(payer): %v", err)
	}
	if len(allowPayerOut.States) == 0 || allowPayerOut.States[0].Object == nil {
		t.Fatalf("AllowUsers(payer) returned empty/nil state")
	}

	allowPayeeOut, err := c.AllowUsers(tok.Address, map[string]bool{payee.PublicKey: true})
	if err != nil {
		t.Fatalf("AllowUsers(payee): %v", err)
	}
	if len(allowPayeeOut.States) == 0 || allowPayeeOut.States[0].Object == nil {
		t.Fatalf("AllowUsers(payee) returned empty/nil state")
	}

	_ = createTransfer(t, c, tok, payer.PublicKey, "50", tok.Decimals, tok.TokenType, "")

	// Deploy Payment contract
	contractState := models.ContractStateModel{}
	deployedContract, err := c.DeployContract1(paymentV1.PAYMENT_CONTRACT_V1)
	if err != nil {
		t.Fatalf("DeployContract(Payment): %v", err)
	}
	if len(deployedContract.States) == 0 || deployedContract.States[0].Object == nil {
		t.Fatalf("DeployContract(Payment) returned empty/nil state")
	}
	unmarshalState(t, deployedContract.States[0].Object, &contractState)
	address := contractState.Address
	if address == "" {
		t.Fatalf("DeployContract(Payment) returned empty address")
	}

	log.Printf("DeployContract(Payment) Output States: %+v", deployedContract.States)
	log.Printf("DeployContract(Payment) Output Logs: %+v", deployedContract.Logs)
	log.Printf("DeployContract(Payment) Output Delegated Call: %+v", deployedContract.DelegatedCall)

	// Create payment (payer)
	orderID := fmt.Sprintf("order-%d-void", time.Now().Unix())
	amount := "10"

	c.SetPrivateKey(payerPriv)
	created, err := c.CreatePayment(address, tok.Address, orderID, payer.PublicKey, payee.PublicKey, amount, time.Now().Add(10*time.Minute))
	if err != nil {
		t.Fatalf("CreatePayment: %v", err)
	}
	if len(created.States) == 0 || created.States[0].Object == nil {
		t.Fatalf("CreatePayment returned empty/nil state")
	}

	var pay paymentV1Models.PaymentStateModel
	unmarshalState(t, created.States[0].Object, &pay)
	if pay.Address == "" {
		t.Fatalf("payment addr empty")
	}
	if pay.OrderId != orderID {
		t.Fatalf("CreatePayment OrderId mismatch: got %q want %q", pay.OrderId, orderID)
	}

	log.Printf("CreatePayment(VoidFlow) Output States: %+v", created.States)
	log.Printf("CreatePayment(VoidFlow) Output Logs: %+v", created.Logs)
	log.Printf("CreatePayment(VoidFlow) Output Delegated Call: %+v", created.DelegatedCall)

	// Allow payment
	c.SetPrivateKey(ownerPriv)
	allowPaymentOut, err := c.AllowUsers(tok.Address, map[string]bool{pay.Address: true})
	if err != nil {
		t.Fatalf("AllowUsers(payment): %v", err)
	}
	if len(allowPaymentOut.States) == 0 || allowPaymentOut.States[0].Object == nil {
		t.Fatalf("AllowUsers(payment) returned empty/nil state")
	}

	// Authorize + Void
	c.SetPrivateKey(payerPriv)

	authOut, err := c.AuthorizePayment(pay.Address, tokenV1Domain.FUNGIBLE, "")
	if err != nil {
		t.Fatalf("AuthorizePayment: %v", err)
	}
	if len(authOut.States) == 0 || authOut.States[0].Object == nil {
		t.Fatalf("AuthorizePayment returned empty/nil state")
	}

	var payAuth paymentV1Models.PaymentStateModel
	unmarshalState(t, authOut.States[0].Object, &payAuth)
	if payAuth.Address != pay.Address {
		t.Fatalf("AuthorizePayment address mismatch: got %q want %q", payAuth.Address, pay.Address)
	}

	log.Printf("AuthorizePayment(VoidFlow) Output States: %+v", authOut.States)
	log.Printf("AuthorizePayment(VoidFlow) Output Logs: %+v", authOut.Logs)
	log.Printf("AuthorizePayment(VoidFlow) Output Delegated Call: %+v", authOut.DelegatedCall)

	voidOut, err := c.VoidPayment(pay.Address, tokenV1Domain.FUNGIBLE, "")
	if err != nil {
		t.Fatalf("VoidPayment: %v", err)
	}
	if len(voidOut.States) == 0 || voidOut.States[0].Object == nil {
		t.Fatalf("VoidPayment returned empty/nil state")
	}

	var payVoid paymentV1Models.PaymentStateModel
	unmarshalState(t, voidOut.States[0].Object, &payVoid)
	if payVoid.Address != pay.Address {
		t.Fatalf("VoidPayment address mismatch: got %q want %q", payVoid.Address, pay.Address)
	}

	log.Printf("VoidPayment(VoidFlow) Output States: %+v", voidOut.States)
	log.Printf("VoidPayment(VoidFlow) Output Logs: %+v", voidOut.Logs)
	log.Printf("VoidPayment(VoidFlow) Output Delegated Call: %+v", voidOut.DelegatedCall)
}

func TestPaymentFlow_NonFungible(t *testing.T) {
	c := setupClient(t)

	// --------------------------------------------------------------------
	// Owner + NFT
	// --------------------------------------------------------------------
	owner, ownerPriv := createWallet(t, c)
	c.SetPrivateKey(ownerPriv)

	dec := 0
	tokenType := tokenV1Domain.NON_FUNGIBLE
	stablecoin := false

	tok := createBasicToken(t, c, owner.PublicKey, dec, true, tokenType, stablecoin)

	if tok.Address == "" {
		t.Fatalf("Token address empty")
	}
	if tok.TokenType != tokenType {
		t.Fatalf("Token type mismatch: got %s want %s", tok.TokenType, tokenType)
	}

	log.Printf("Token Address: %s", tok.Address)
	log.Printf("Token Type: %s", tok.TokenType)

	// --------------------------------------------------------------------
	// Mint NFT - padrão mint
	// --------------------------------------------------------------------
	mintOut, err := c.MintToken(tok.Address, owner.PublicKey, "1", dec, tok.TokenType)
	if err != nil {
		t.Fatalf("MintToken NFT: %v", err)
	}
	if len(mintOut.States) == 0 || mintOut.States[0].Object == nil {
		t.Fatalf("MintToken NFT returned empty/nil state")
	}

	var mint tokenV1Domain.Mint
	unmarshalState(t, mintOut.States[0].Object, &mint)

	if mint.TokenAddress != tok.Address {
		t.Fatalf("Mint TokenAddress mismatch: got %s want %s", mint.TokenAddress, tok.Address)
	}
	if mint.MintTo != owner.PublicKey {
		t.Fatalf("Mint ToAddress mismatch: got %s want %s", mint.MintTo, owner.PublicKey)
	}
	if mint.TokenType != tok.TokenType {
		t.Fatalf("Mint TokenType mismatch: got %s want %s", mint.TokenType, tok.TokenType)
	}
	if len(mint.TokenUUIDList) != 1 {
		t.Fatalf("expected 1 uuid, got %d", len(mint.TokenUUIDList))
	}
	nftUUID := mint.TokenUUIDList[0]
	if nftUUID == "" {
		t.Fatalf("minted uuid empty")
	}

	log.Printf("Mint Output States: %+v", mintOut.States)
	log.Printf("Mint Output Logs: %+v", mintOut.Logs)
	log.Printf("Mint Output Delegated Call: %+v", mintOut.DelegatedCall)

	log.Printf("Mint TokenAddress: %s", mint.TokenAddress)
	log.Printf("Mint ToAddress: %s", mint.MintTo)
	log.Printf("Mint Amount: %s", mint.Amount)
	log.Printf("Mint TokenType: %s", mint.TokenType)
	log.Printf("Mint TokenUUIDList: %+v", mint.TokenUUIDList)

	// --------------------------------------------------------------------
	// Payer / Payee
	// --------------------------------------------------------------------
	payer, payerPriv := createWallet(t, c)
	payee, payeePriv := createWallet(t, c)

	// Allow users (owner)
	c.SetPrivateKey(ownerPriv)
	allowOut, err := c.AllowUsers(tok.Address, map[string]bool{
		payer.PublicKey: true,
		payee.PublicKey: true,
	})
	if err != nil {
		t.Fatalf("AllowUsers(payer/payee): %v", err)
	}
	if len(allowOut.States) == 0 || allowOut.States[0].Object == nil {
		t.Fatalf("AllowUsers(payer/payee) returned empty/nil state")
	}

	// Transfer NFT -> payer (owner)
	if _, err := c.TransferToken(tok.Address, payer.PublicKey, "1", dec, tokenType, nftUUID); err != nil {
		t.Fatalf("Transfer NFT to payer: %v", err)
	}

	// --------------------------------------------------------------------
	// Deploy Payment contract
	// --------------------------------------------------------------------
	contractState := models.ContractStateModel{}
	deployedContract, err := c.DeployContract1(paymentV1.PAYMENT_CONTRACT_V1)
	if err != nil {
		t.Fatalf("DeployContract(Payment): %v", err)
	}
	if len(deployedContract.States) == 0 || deployedContract.States[0].Object == nil {
		t.Fatalf("DeployContract(Payment) returned empty/nil state")
	}
	unmarshalState(t, deployedContract.States[0].Object, &contractState)
	paymentContractAddr := contractState.Address
	if paymentContractAddr == "" {
		t.Fatalf("DeployContract(Payment) returned empty contract address")
	}

	log.Printf("DeployContract(Payment) Output States: %+v", deployedContract.States)
	log.Printf("DeployContract(Payment) Output Logs: %+v", deployedContract.Logs)
	log.Printf("DeployContract(Payment) Output Delegated Call: %+v", deployedContract.DelegatedCall)
	log.Printf("Payment Contract Address: %s", paymentContractAddr)

	// --------------------------------------------------------------------
	// Create Payment (payer)
	// --------------------------------------------------------------------
	orderID := fmt.Sprintf("order-nft-%d", time.Now().Unix())
	exp := time.Now().Add(30 * time.Minute)

	c.SetPrivateKey(payerPriv)
	created, err := c.CreatePayment(paymentContractAddr, tok.Address, orderID, payer.PublicKey, payee.PublicKey, "1", exp)
	if err != nil {
		t.Fatalf("CreatePayment NFT: %v", err)
	}
	if len(created.States) == 0 || created.States[0].Object == nil {
		t.Fatalf("CreatePayment NFT returned empty/nil state")
	}

	var pay paymentV1Models.PaymentStateModel
	unmarshalState(t, created.States[0].Object, &pay)

	if pay.Address == "" {
		t.Fatalf("payment addr empty")
	}
	if pay.TokenAddress != tok.Address {
		t.Fatalf("CreatePayment NFT TokenAddress mismatch: got %q want %q", pay.TokenAddress, tok.Address)
	}
	if pay.OrderId != orderID {
		t.Fatalf("CreatePayment NFT OrderId mismatch: got %q want %q", pay.OrderId, orderID)
	}
	if pay.Payer != payer.PublicKey {
		t.Fatalf("CreatePayment NFT Payer mismatch: got %q want %q", pay.Payer, payer.PublicKey)
	}
	if pay.Payee != payee.PublicKey {
		t.Fatalf("CreatePayment NFT Payee mismatch: got %q want %q", pay.Payee, payee.PublicKey)
	}
	if pay.Amount != "1" {
		t.Fatalf("CreatePayment NFT Amount mismatch: got %q want %q", pay.Amount, "1")
	}
	if pay.Hash == "" {
		t.Fatalf("CreatePayment NFT Hash empty")
	}

	log.Printf("CreatePayment NFT Output States: %+v", created.States)
	log.Printf("CreatePayment NFT Output Logs: %+v", created.Logs)
	log.Printf("CreatePayment NFT Output Delegated Call: %+v", created.DelegatedCall)

	// Allow payment contract
	c.SetPrivateKey(ownerPriv)
	allowPayOut, err := c.AllowUsers(tok.Address, map[string]bool{pay.Address: true})
	if err != nil {
		t.Fatalf("AllowUsers(payment): %v", err)
	}
	if len(allowPayOut.States) == 0 || allowPayOut.States[0].Object == nil {
		t.Fatalf("AllowUsers(payment) returned empty/nil state")
	}

	// --------------------------------------------------------------------
	// Authorize / Capture / Refund
	// --------------------------------------------------------------------
	c.SetPrivateKey(payerPriv)
	authOut, err := c.AuthorizePayment(pay.Address, tokenType, nftUUID)
	if err != nil {
		t.Fatalf("AuthorizePayment NFT: %v", err)
	}
	if len(authOut.States) == 0 || authOut.States[0].Object == nil {
		t.Fatalf("AuthorizePayment NFT returned empty/nil state")
	}

	var payAuth paymentV1Models.PaymentStateModel
	unmarshalState(t, authOut.States[0].Object, &payAuth)
	if payAuth.Address != pay.Address {
		t.Fatalf("AuthorizePayment NFT address mismatch: got %q want %q", payAuth.Address, pay.Address)
	}

	log.Printf("AuthorizePayment NFT Output States: %+v", authOut.States)
	log.Printf("AuthorizePayment NFT Output Logs: %+v", authOut.Logs)
	log.Printf("AuthorizePayment NFT Output Delegated Call: %+v", authOut.DelegatedCall)

	c.SetPrivateKey(payeePriv)
	capOut, err := c.CapturePayment(pay.Address, tokenType, nftUUID)
	if err != nil {
		t.Fatalf("CapturePayment NFT: %v", err)
	}
	if len(capOut.States) == 0 || capOut.States[0].Object == nil {
		t.Fatalf("CapturePayment NFT returned empty/nil state")
	}

	var payCap paymentV1Models.PaymentStateModel
	unmarshalState(t, capOut.States[0].Object, &payCap)
	if payCap.Address != pay.Address {
		t.Fatalf("CapturePayment NFT address mismatch: got %q want %q", payCap.Address, pay.Address)
	}

	log.Printf("CapturePayment NFT Output States: %+v", capOut.States)
	log.Printf("CapturePayment NFT Output Logs: %+v", capOut.Logs)
	log.Printf("CapturePayment NFT Output Delegated Call: %+v", capOut.DelegatedCall)

	refOut, err := c.RefundPayment(pay.Address, "1", tokenType, nftUUID)
	if err != nil {
		t.Fatalf("RefundPayment NFT: %v", err)
	}
	if len(refOut.States) == 0 || refOut.States[0].Object == nil {
		t.Fatalf("RefundPayment NFT returned empty/nil state")
	}

	var payRef paymentV1Models.PaymentStateModel
	unmarshalState(t, refOut.States[0].Object, &payRef)
	if payRef.Address != pay.Address {
		t.Fatalf("RefundPayment NFT address mismatch: got %q want %q", payRef.Address, pay.Address)
	}

	log.Printf("RefundPayment NFT Output States: %+v", refOut.States)
	log.Printf("RefundPayment NFT Output Logs: %+v", refOut.Logs)
	log.Printf("RefundPayment NFT Output Delegated Call: %+v", refOut.DelegatedCall)

	// --------------------------------------------------------------------
	// DirectPay (novo contract)
	// --------------------------------------------------------------------
	contractState = models.ContractStateModel{}
	deployedContract2, err := c.DeployContract1(paymentV1.PAYMENT_CONTRACT_V1)
	if err != nil {
		t.Fatalf("DeployContract(Payment #2): %v", err)
	}
	if len(deployedContract2.States) == 0 || deployedContract2.States[0].Object == nil {
		t.Fatalf("DeployContract(Payment #2) returned empty/nil state")
	}
	unmarshalState(t, deployedContract2.States[0].Object, &contractState)
	address2 := contractState.Address
	if address2 == "" {
		t.Fatalf("DeployContract(Payment #2) returned empty address")
	}

	c.SetPrivateKey(payerPriv)
	directOut, err := c.DirectPay(address2, tok.Address, orderID+"-direct", payer.PublicKey, payee.PublicKey, "1", tokenType, nftUUID)
	if err != nil {
		t.Fatalf("DirectPay NFT: %v", err)
	}
	if len(directOut.States) == 0 || directOut.States[0].Object == nil {
		t.Fatalf("DirectPay NFT returned empty/nil state")
	}

	log.Printf("DirectPay NFT Output States: %+v", directOut.States)
	log.Printf("DirectPay NFT Output Logs: %+v", directOut.Logs)
	log.Printf("DirectPay NFT Output Delegated Call: %+v", directOut.DelegatedCall)

	// --------------------------------------------------------------------
	// Pause / Unpause (payer + payee)
	// --------------------------------------------------------------------
	c.SetPrivateKey(payerPriv)
	pauseOut, err := c.PausePayment(pay.Address, true)
	if err != nil {
		t.Fatalf("PausePayment(payer) NFT: %v", err)
	}
	if len(pauseOut.States) == 0 || pauseOut.States[0].Object == nil {
		t.Fatalf("PausePayment(payer) NFT returned empty/nil state")
	}
	unpauseOut, err := c.UnpausePayment(pay.Address, false)
	if err != nil {
		t.Fatalf("UnpausePayment(payer) NFT: %v", err)
	}
	if len(unpauseOut.States) == 0 || unpauseOut.States[0].Object == nil {
		t.Fatalf("UnpausePayment(payer) NFT returned empty/nil state")
	}

	c.SetPrivateKey(payeePriv)
	pauseOut2, err := c.PausePayment(pay.Address, true)
	if err != nil {
		t.Fatalf("PausePayment(payee) NFT: %v", err)
	}
	if len(pauseOut2.States) == 0 || pauseOut2.States[0].Object == nil {
		t.Fatalf("PausePayment(payee) NFT returned empty/nil state")
	}
	unpauseOut2, err := c.UnpausePayment(pay.Address, false)
	if err != nil {
		t.Fatalf("UnpausePayment(payee) NFT: %v", err)
	}
	if len(unpauseOut2.States) == 0 || unpauseOut2.States[0].Object == nil {
		t.Fatalf("UnpausePayment(payee) NFT returned empty/nil state")
	}

	// --------------------------------------------------------------------
	// GetPayment
	// --------------------------------------------------------------------
	getOut, err := c.GetPayment(pay.Address)
	if err != nil {
		t.Fatalf("GetPayment NFT: %v", err)
	}
	if len(getOut.States) == 0 || getOut.States[0].Object == nil {
		t.Fatalf("GetPayment NFT returned empty/nil state")
	}
	var got paymentV1Models.PaymentStateModel
	unmarshalState(t, getOut.States[0].Object, &got)
	if got.Address != pay.Address {
		t.Fatalf("GetPayment NFT address mismatch: got %q want %q", got.Address, pay.Address)
	}

	log.Printf("GetPayment NFT Output States: %+v", getOut.States)
	log.Printf("GetPayment NFT Output Logs: %+v", getOut.Logs)
	log.Printf("GetPayment NFT Output Delegated Call: %+v", getOut.DelegatedCall)
}

func TestPaymentAuthVoidFlow_NonFungible(t *testing.T) {
	c := setupClient(t)

	// Owner + NFT
	owner, ownerPriv := createWallet(t, c)
	c.SetPrivateKey(ownerPriv)

	dec := 0
	tokenType := tokenV1Domain.NON_FUNGIBLE
	stablecoin := false

	tok := createBasicToken(t, c, owner.PublicKey, dec, true, tokenType, stablecoin)

	// Mint NFT - padrão mint
	mintOut, err := c.MintToken(tok.Address, owner.PublicKey, "1", dec, tok.TokenType)
	if err != nil {
		t.Fatalf("MintToken NFT: %v", err)
	}
	if len(mintOut.States) == 0 || mintOut.States[0].Object == nil {
		t.Fatalf("MintToken NFT returned empty/nil state")
	}

	var mint tokenV1Domain.Mint
	unmarshalState(t, mintOut.States[0].Object, &mint)
	if len(mint.TokenUUIDList) != 1 {
		t.Fatalf("expected 1 uuid, got %d", len(mint.TokenUUIDList))
	}
	nftUUID := mint.TokenUUIDList[0]

	// Payer / Payee
	payer, payerPriv := createWallet(t, c)
	payee, _ := createWallet(t, c)

	c.SetPrivateKey(ownerPriv)
	_, err = c.AllowUsers(tok.Address, map[string]bool{
		payer.PublicKey: true,
		payee.PublicKey: true,
	})
	if err != nil {
		t.Fatalf("AllowUsers(payer/payee): %v", err)
	}

	// Transfer NFT -> payer
	if _, err := c.TransferToken(tok.Address, payer.PublicKey, "1", dec, tokenType, nftUUID); err != nil {
		t.Fatalf("Transfer NFT: %v", err)
	}

	// Deploy Payment contract
	contractState := models.ContractStateModel{}
	deployedContract, err := c.DeployContract1(paymentV1.PAYMENT_CONTRACT_V1)
	if err != nil {
		t.Fatalf("DeployContract(Payment): %v", err)
	}
	if len(deployedContract.States) == 0 || deployedContract.States[0].Object == nil {
		t.Fatalf("DeployContract(Payment) returned empty/nil state")
	}
	unmarshalState(t, deployedContract.States[0].Object, &contractState)

	// Create payment (payer)
	orderID := fmt.Sprintf("order-nft-void-%d", time.Now().Unix())

	c.SetPrivateKey(payerPriv)
	created, err := c.CreatePayment(contractState.Address, tok.Address, orderID, payer.PublicKey, payee.PublicKey, "1", time.Now().Add(10*time.Minute))
	if err != nil {
		t.Fatalf("CreatePayment NFT: %v", err)
	}
	if len(created.States) == 0 || created.States[0].Object == nil {
		t.Fatalf("CreatePayment NFT returned empty/nil state")
	}

	var pay paymentV1Models.PaymentStateModel
	unmarshalState(t, created.States[0].Object, &pay)
	if pay.Address == "" {
		t.Fatalf("payment addr empty")
	}

	// Allow payment
	c.SetPrivateKey(ownerPriv)
	_, err = c.AllowUsers(tok.Address, map[string]bool{pay.Address: true})
	if err != nil {
		t.Fatalf("AllowUsers(payment): %v", err)
	}

	// Authorize + Void
	c.SetPrivateKey(payerPriv)

	authOut, err := c.AuthorizePayment(pay.Address, tokenType, nftUUID)
	if err != nil {
		t.Fatalf("AuthorizePayment NFT: %v", err)
	}
	if len(authOut.States) == 0 || authOut.States[0].Object == nil {
		t.Fatalf("AuthorizePayment NFT returned empty/nil state")
	}

	voidOut, err := c.VoidPayment(pay.Address, tokenType, nftUUID)
	if err != nil {
		t.Fatalf("VoidPayment NFT: %v", err)
	}
	if len(voidOut.States) == 0 || voidOut.States[0].Object == nil {
		t.Fatalf("VoidPayment NFT returned empty/nil state")
	}

	log.Printf("AuthorizePayment NFT Output States: %+v", authOut.States)
	log.Printf("AuthorizePayment NFT Output Logs: %+v", authOut.Logs)
	log.Printf("AuthorizePayment NFT Output Delegated Call: %+v", authOut.DelegatedCall)

	log.Printf("VoidPayment NFT Output States: %+v", voidOut.States)
	log.Printf("VoidPayment NFT Output Logs: %+v", voidOut.Logs)
	log.Printf("VoidPayment NFT Output Delegated Call: %+v", voidOut.DelegatedCall)
}
