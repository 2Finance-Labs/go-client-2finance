package e2e_test

import (
	"fmt"
	"testing"
	"time"

	"gitlab.com/2finance/2finance-network/blockchain/contract/contractV1/models"
	"gitlab.com/2finance/2finance-network/blockchain/contract/paymentV1"
	paymentV1Domain "gitlab.com/2finance/2finance-network/blockchain/contract/paymentV1/domain"
	tokenV1Domain "gitlab.com/2finance/2finance-network/blockchain/contract/tokenV1/domain"
)

func TestPaymentFlow(t *testing.T) {
	c := setupClient(t)
	owner, ownerPriv := createWallet(t, c)
	c.SetPrivateKey(ownerPriv)
	dec := 6

	tok := createBasicToken(t, c, owner.PublicKey, dec, true, tokenV1Domain.FUNGIBLE)
	_ = createMint(t, c, tok, owner.PublicKey, "10000", dec, tok.TokenType)

	payer, payerPriv := createWallet(t, c)
	payee, payeePriv := createWallet(t, c)

	c.SetPrivateKey(ownerPriv)
	_, _ = c.AllowUsers(tok.Address, map[string]bool{payer.PublicKey: true})
	_, _ = c.AllowUsers(tok.Address, map[string]bool{payee.PublicKey: true})
	_ = createTransfer(t, c, tok, payer.PublicKey, "50", tok.Decimals, tok.TokenType, "")

	orderID := fmt.Sprintf("order-%d", time.Now().Unix())
	c.SetPrivateKey(payerPriv)
	amount := "10"
	exp := time.Now().Add(30 * time.Minute)

	contractState := models.ContractStateModel{}
	deployedContract, err := c.DeployContract1(paymentV1.PAYMENT_CONTRACT_V1)
	if err != nil {
		t.Fatalf("DeployContract: %v", err)
	}
	unmarshalState(t, deployedContract.States[0].Object, &contractState)
	address := contractState.Address

	created, err := c.CreatePayment(address, tok.Address, orderID, payer.PublicKey, payee.PublicKey, amount, exp)
	if err != nil {
		t.Fatalf("CreatePayment: %v", err)
	}
	var pay paymentV1Domain.Payment
	unmarshalState(t, created.States[0].Object, &pay)

	if pay.Address == "" {
		t.Fatalf("payment addr empty")
	}
	c.SetPrivateKey(ownerPriv)
	_, _ = c.AllowUsers(tok.Address, map[string]bool{pay.Address: true})

	c.SetPrivateKey(payerPriv)
	if _, err := c.AuthorizePayment(pay.Address, tokenV1Domain.FUNGIBLE, ""); err != nil {
		t.Fatalf("AuthorizePayment: %v", err)
	}

	// capture and refund by payee
	c.SetPrivateKey(payeePriv)
	if _, err := c.CapturePayment(pay.Address, tokenV1Domain.FUNGIBLE, ""); err != nil {
		t.Fatalf("CapturePayment warning: %v", err)
	}
	_, _ = c.RefundPayment(pay.Address, "10", tokenV1Domain.FUNGIBLE, "")

	// direct pay (no auth/capture)
	contractState = models.ContractStateModel{}
	deployedContract, err = c.DeployContract1(paymentV1.PAYMENT_CONTRACT_V1)
	if err != nil {
		t.Fatalf("DeployContract: %v", err)
	}
	unmarshalState(t, deployedContract.States[0].Object, &contractState)
	address = contractState.Address

	_, _ = c.DirectPay(address, tok.Address, orderID+"-direct", payer.PublicKey, payee.PublicKey, amt(2, dec), tokenV1Domain.FUNGIBLE, "")

	// pause/unpause by owner/admin (if applicable)
	c.SetPrivateKey(payerPriv)
	_, err = c.PausePayment(pay.Address, true)
	if err != nil {
		t.Fatalf("PausePayment: %v", err)
	}
	_, err = c.UnpausePayment(pay.Address, false)
	if err != nil {
		t.Fatalf("UnpausePayment: %v", err)
	}

	c.SetPrivateKey(payeePriv)
	_, err = c.PausePayment(pay.Address, true)
	if err != nil {
		t.Fatalf("PausePayment: %v", err)
	}
	_, err = c.UnpausePayment(pay.Address, false)
	if err != nil {
		t.Fatalf("UnpausePayment: %v", err)
	}

	// getters
	if _, err := c.GetPayment(pay.Address); err != nil {
		t.Fatalf("GetPayment: %v", err)
	}
	//if _, err := c.ListPayments(payer.PublicKey, payee.PublicKey, orderID, tok.Address, []string{}, 1, 10, true); err != nil { t.Fatalf("ListPayments: %v", err) }
}

// More payment scenarios
func TestPaymentAuthVoidFlow(t *testing.T) {
	c := setupClient(t)
	owner, ownerPriv := createWallet(t, c)

	c.SetPrivateKey(ownerPriv)
	dec := 6

	tok := createBasicToken(t, c, owner.PublicKey, dec, true, tokenV1Domain.FUNGIBLE)
	payer, payerPriv := createWallet(t, c)
	payee, _ := createWallet(t, c)
	c.SetPrivateKey(ownerPriv)
	_, err := c.AllowUsers(tok.Address, map[string]bool{payer.PublicKey: true})
	if err != nil {
		t.Fatalf("AllowUsers(payer): %v", err)
	}
	_, err = c.AllowUsers(tok.Address, map[string]bool{payee.PublicKey: true})
	if err != nil {
		t.Fatalf("AllowUsers(payee): %v", err)
	}
	_ = createTransfer(t, c, tok, payer.PublicKey, "50", tok.Decimals, tok.TokenType, "")

	orderID := fmt.Sprintf("order-%d-void", time.Now().Unix())
	c.SetPrivateKey(payerPriv)
	amount := "10"

	contractState := models.ContractStateModel{}
	deployedContract, err := c.DeployContract1(paymentV1.PAYMENT_CONTRACT_V1)
	if err != nil {
		t.Fatalf("DeployContract: %v", err)
	}
	unmarshalState(t, deployedContract.States[0].Object, &contractState)
	address := contractState.Address

	created, err := c.CreatePayment(address, tok.Address, orderID, payer.PublicKey, payee.PublicKey, amount, time.Now().Add(10*time.Minute))
	if err != nil {
		t.Fatalf("CreatePayment: %v", err)
	}
	var pay paymentV1Domain.Payment
	unmarshalState(t, created.States[0].Object, &pay)

	c.SetPrivateKey(ownerPriv)
	_, err = c.AllowUsers(tok.Address, map[string]bool{pay.Address: true})
	if err != nil {
		t.Fatalf("AllowUsers(payment): %v", err)
	}

	c.SetPrivateKey(payerPriv)
	if _, err := c.AuthorizePayment(pay.Address, tokenV1Domain.FUNGIBLE, ""); err != nil {
		t.Fatalf("AuthorizePayment: %v", err)
	}
	if _, err := c.VoidPayment(pay.Address, tokenV1Domain.FUNGIBLE, ""); err != nil {
		t.Fatalf("VoidPayment warning: %v", err)
	}
}

// func TestPaymentStatusQueries(t *testing.T) {
// 	c := setupClient(t)
// 	// TODO FIX
// 	// Best-effort: ensure the call works with a status filter
// 	//_, _ = c.ListPayments("", "", "", "", []string{"created","authorized","captured"}, 1, 10, true)
// }

func TestPaymentFlow_NonFungible(t *testing.T) {
	c := setupClient(t)

	// --------------------------------------------------------------------
	// Owner + NFT
	// --------------------------------------------------------------------
	owner, ownerPriv := createWallet(t, c)
	c.SetPrivateKey(ownerPriv)

	dec := 0
	tokenType := tokenV1Domain.NON_FUNGIBLE

	tok := createBasicToken(t, c, owner.PublicKey, dec, true, tokenType)

	// Mint 1 NFT
	mintOut, err := c.MintToken(tok.Address, owner.PublicKey, "1", dec, tok.TokenType)
	if err != nil {
		t.Fatalf("MintToken NFT: %v", err)
	}

	var mint tokenV1Domain.Mint
	unmarshalState(t, mintOut.States[0].Object, &mint)
	if len(mint.TokenUUIDList) != 1 {
		t.Fatalf("expected 1 uuid, got %d", len(mint.TokenUUIDList))
	}
	nftUUID := mint.TokenUUIDList[0]

	// --------------------------------------------------------------------
	// Payer / Payee
	// --------------------------------------------------------------------
	payer, payerPriv := createWallet(t, c)
	payee, payeePriv := createWallet(t, c)

	// Allow users
	c.SetPrivateKey(ownerPriv)
	_, _ = c.AllowUsers(tok.Address, map[string]bool{
		payer.PublicKey: true,
		payee.PublicKey: true,
	})

	// Transfer NFT → payer
	if _, err := c.TransferToken(
		tok.Address,
		payer.PublicKey,
		"1",
		dec,
		tokenType,
		nftUUID,
	); err != nil {
		t.Fatalf("Transfer NFT to payer: %v", err)
	}

	// --------------------------------------------------------------------
	// Create Payment
	// --------------------------------------------------------------------
	orderID := fmt.Sprintf("order-nft-%d", time.Now().Unix())
	exp := time.Now().Add(30 * time.Minute)

	contractState := models.ContractStateModel{}
	deployedContract, err := c.DeployContract1(paymentV1.PAYMENT_CONTRACT_V1)
	if err != nil {
		t.Fatalf("DeployContract: %v", err)
	}
	unmarshalState(t, deployedContract.States[0].Object, &contractState)

	c.SetPrivateKey(payerPriv)
	created, err := c.CreatePayment(
		contractState.Address,
		tok.Address,
		orderID,
		payer.PublicKey,
		payee.PublicKey,
		"1", // NFT sempre 1
		exp,
	)
	if err != nil {
		t.Fatalf("CreatePayment NFT: %v", err)
	}

	var pay paymentV1Domain.Payment
	unmarshalState(t, created.States[0].Object, &pay)
	if pay.Address == "" {
		t.Fatalf("payment addr empty")
	}

	// Allow payment contract
	c.SetPrivateKey(ownerPriv)
	_, _ = c.AllowUsers(tok.Address, map[string]bool{pay.Address: true})

	// --------------------------------------------------------------------
	// Authorize / Capture / Refund
	// --------------------------------------------------------------------
	c.SetPrivateKey(payerPriv)
	if _, err := c.AuthorizePayment(pay.Address, tokenType, nftUUID); err != nil {
		t.Fatalf("AuthorizePayment NFT: %v", err)
	}

	c.SetPrivateKey(payeePriv)
	if _, err := c.CapturePayment(pay.Address, tokenType, nftUUID); err != nil {
		t.Fatalf("CapturePayment NFT: %v", err)
	}

	_, _ = c.RefundPayment(pay.Address, "1", tokenType, nftUUID)

	// --------------------------------------------------------------------
	// Direct Pay
	// --------------------------------------------------------------------
	contractState = models.ContractStateModel{}
	deployedContract, err = c.DeployContract1(paymentV1.PAYMENT_CONTRACT_V1)
	if err != nil {
		t.Fatalf("DeployContract: %v", err)
	}
	unmarshalState(t, deployedContract.States[0].Object, &contractState)

	c.SetPrivateKey(payerPriv)
	_, _ = c.DirectPay(
		contractState.Address,
		tok.Address,
		orderID+"-direct",
		payer.PublicKey,
		payee.PublicKey,
		"1",
		tokenType,
		nftUUID,
	)

	// --------------------------------------------------------------------
	// Pause / Unpause
	// --------------------------------------------------------------------
	c.SetPrivateKey(payerPriv)
	_, _ = c.PausePayment(pay.Address, true)
	_, _ = c.UnpausePayment(pay.Address, false)

	c.SetPrivateKey(payeePriv)
	_, _ = c.PausePayment(pay.Address, true)
	_, _ = c.UnpausePayment(pay.Address, false)

	// --------------------------------------------------------------------
	// Getters
	// --------------------------------------------------------------------
	if _, err := c.GetPayment(pay.Address); err != nil {
		t.Fatalf("GetPayment: %v", err)
	}
}

func TestPaymentAuthVoidFlow_NonFungible(t *testing.T) {
	c := setupClient(t)

	// --------------------------------------------------------------------
	// Owner + NFT
	// --------------------------------------------------------------------
	owner, ownerPriv := createWallet(t, c)
	c.SetPrivateKey(ownerPriv)

	dec := 0
	tokenType := tokenV1Domain.NON_FUNGIBLE

	tok := createBasicToken(t, c, owner.PublicKey, dec, true, tokenType)

	mintOut, err := c.MintToken(tok.Address, owner.PublicKey, "1", dec, tok.TokenType)
	if err != nil {
		t.Fatalf("MintToken NFT: %v", err)
	}

	var mint tokenV1Domain.Mint
	unmarshalState(t, mintOut.States[0].Object, &mint)
	nftUUID := mint.TokenUUIDList[0]

	// --------------------------------------------------------------------
	// Payer / Payee
	// --------------------------------------------------------------------
	payer, payerPriv := createWallet(t, c)
	payee, _ := createWallet(t, c)

	c.SetPrivateKey(ownerPriv)
	_, _ = c.AllowUsers(tok.Address, map[string]bool{
		payer.PublicKey: true,
		payee.PublicKey: true,
	})

	// Transfer NFT → payer
	if _, err := c.TransferToken(
		tok.Address,
		payer.PublicKey,
		"1",
		dec,
		tokenType,
		nftUUID,
	); err != nil {
		t.Fatalf("Transfer NFT: %v", err)
	}

	// --------------------------------------------------------------------
	// Create Payment
	// --------------------------------------------------------------------
	orderID := fmt.Sprintf("order-nft-void-%d", time.Now().Unix())

	contractState := models.ContractStateModel{}
	deployedContract, err := c.DeployContract1(paymentV1.PAYMENT_CONTRACT_V1)
	if err != nil {
		t.Fatalf("DeployContract: %v", err)
	}
	unmarshalState(t, deployedContract.States[0].Object, &contractState)

	c.SetPrivateKey(payerPriv)
	created, err := c.CreatePayment(
		contractState.Address,
		tok.Address,
		orderID,
		payer.PublicKey,
		payee.PublicKey,
		"1",
		time.Now().Add(10*time.Minute),
	)
	if err != nil {
		t.Fatalf("CreatePayment NFT: %v", err)
	}

	var pay paymentV1Domain.Payment
	unmarshalState(t, created.States[0].Object, &pay)

	// Allow payment contract
	c.SetPrivateKey(ownerPriv)
	_, _ = c.AllowUsers(tok.Address, map[string]bool{pay.Address: true})

	// --------------------------------------------------------------------
	// Authorize + Void
	// --------------------------------------------------------------------
	c.SetPrivateKey(payerPriv)
	if _, err := c.AuthorizePayment(pay.Address, tokenType, nftUUID); err != nil {
		t.Fatalf("AuthorizePayment NFT: %v", err)
	}
	if _, err := c.VoidPayment(pay.Address, tokenType, nftUUID); err != nil {
		t.Fatalf("VoidPayment NFT: %v", err)
	}
}
