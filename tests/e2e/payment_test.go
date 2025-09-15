package e2e_test

import (
	"fmt"
	"testing"
	"time"

	paymentV1Domain "gitlab.com/2finance/2finance-network/blockchain/contract/paymentV1/domain"
)

func TestPaymentFlow(t *testing.T) {
	c := setupClient(t)
	owner, ownerPriv := createWallet(t, c)
	c.SetPrivateKey(ownerPriv)
	dec := 6

	tok := createBasicToken(t, c, owner.PublicKey, dec, true)
	_ = createMint(t, c, tok, owner.PublicKey, "10000", dec)

	payer, payerPriv := createWallet(t, c)
	payee, payeePriv := createWallet(t, c)

	c.SetPrivateKey(ownerPriv)
	_, _ = c.AllowUsers(tok.Address, map[string]bool{payer.PublicKey: true})
	_, _ = c.AllowUsers(tok.Address, map[string]bool{payee.PublicKey: true})
	_ = createTransfer(t, c, tok, payer.PublicKey, "50", tok.Decimals)

	orderID := fmt.Sprintf("order-%d", time.Now().Unix())
	c.SetPrivateKey(payerPriv)
	amount := "10"
	exp := time.Now().Add(30 * time.Minute)

	created, err := c.CreatePayment(tok.Address, orderID, payer.PublicKey, payee.PublicKey, amount, exp)
	if err != nil { t.Fatalf("CreatePayment: %v", err) }
	var pay paymentV1Domain.Payment
	unmarshalState(t, created.States[0].Object, &pay)
	
	if pay.Address == "" { t.Fatalf("payment addr empty") }
	c.SetPrivateKey(ownerPriv)
	_, _ = c.AllowUsers(tok.Address, map[string]bool{pay.Address: true})
	
	c.SetPrivateKey(payerPriv)
	if _, err := c.AuthorizePayment(pay.Address); err != nil { t.Fatalf("AuthorizePayment: %v", err) }

	// capture and refund by payee
	c.SetPrivateKey(payeePriv)
	if _, err := c.CapturePayment(pay.Address); err != nil { t.Fatalf("CapturePayment warning: %v", err) }
	_, _ = c.RefundPayment(pay.Address, "10")

	// direct pay (no auth/capture)
	_, _ = c.DirectPay(tok.Address, orderID+"-direct", payer.PublicKey, payee.PublicKey, amt(2, dec))

	// pause/unpause by owner/admin (if applicable)
	c.SetPrivateKey(ownerPriv)
	_, _ = c.PausePayment(pay.Address, true)
	_, _ = c.UnpausePayment(pay.Address, false)

	// getters
	if _, err := c.GetPayment(pay.Address); err != nil { t.Fatalf("GetPayment: %v", err) }
	//if _, err := c.ListPayments(payer.PublicKey, payee.PublicKey, orderID, tok.Address, []string{}, 1, 10, true); err != nil { t.Fatalf("ListPayments: %v", err) }
}

// More payment scenarios
func TestPaymentAuthVoidFlow(t *testing.T) {
	c := setupClient(t)
	owner, ownerPriv := createWallet(t, c)
	c.SetPrivateKey(ownerPriv)
	dec := 6
	tok := createBasicToken(t, c, owner.PublicKey, dec, true)
	payer, payerPriv := createWallet(t, c)
	payee, _ := createWallet(t, c)
	c.SetPrivateKey(ownerPriv)
	_, err := c.AllowUsers(tok.Address, map[string]bool{payer.PublicKey: true})
	if err != nil { t.Fatalf("AllowUsers(payer): %v", err) }
	_, err = c.AllowUsers(tok.Address, map[string]bool{payee.PublicKey: true})
	if err != nil { t.Fatalf("AllowUsers(payee): %v", err) }
	_ = createTransfer(t, c, tok, payer.PublicKey, "50", tok.Decimals)


	orderID := fmt.Sprintf("order-%d-void", time.Now().Unix())
	c.SetPrivateKey(payerPriv)
	amount := "10"
	created, err := c.CreatePayment(tok.Address, orderID, payer.PublicKey, payee.PublicKey, amount, time.Now().Add(10*time.Minute))
	if err != nil { t.Fatalf("CreatePayment: %v", err) }
	var pay paymentV1Domain.Payment
	unmarshalState(t, created.States[0].Object, &pay)

	c.SetPrivateKey(ownerPriv)
	_, err = c.AllowUsers(tok.Address, map[string]bool{pay.Address: true})
	if err != nil { t.Fatalf("AllowUsers(payment): %v", err) }


	c.SetPrivateKey(payerPriv)
	if _, err := c.AuthorizePayment(pay.Address); err != nil { t.Fatalf("AuthorizePayment: %v", err) }
	if _, err := c.VoidPayment(pay.Address); err != nil { t.Logf("VoidPayment warning: %v", err) }
}

// func TestPaymentStatusQueries(t *testing.T) {
// 	c := setupClient(t)
// 	// TODO FIX
// 	// Best-effort: ensure the call works with a status filter
// 	//_, _ = c.ListPayments("", "", "", "", []string{"created","authorized","captured"}, 1, 10, true)
// }