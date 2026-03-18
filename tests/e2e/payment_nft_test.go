package e2e_test

import "testing"

func TestPaymentFlow_NonFungible(t *testing.T) {
	// c := setupClient(t)

	// // --------------------------------------------------------------------
	// // Owner + NFT
	// // --------------------------------------------------------------------
	// owner, ownerPriv := createWallet(t, c)
	// c.SetPrivateKey(ownerPriv)

	// dec := 0
	// tokenType := tokenV1Domain.NON_FUNGIBLE
	// stablecoin := false

	// tok := createBasicToken(t, c, owner.PublicKey, dec, true, tokenType, stablecoin)

	// // Mint 1 NFT
	// mintOut, err := c.MintToken(tok.Address, owner.PublicKey, "1", dec, tok.TokenType)
	// if err != nil {
	// 	t.Fatalf("MintToken NFT: %v", err)
	// }

	// var mint tokenV1Domain.Mint
	// unmarshalState(t, mintOut.States[0].Object, &mint)
	// if len(mint.TokenUUIDList) != 1 {
	// 	t.Fatalf("expected 1 uuid, got %d", len(mint.TokenUUIDList))
	// }
	// nftUUID := mint.TokenUUIDList[0]

	// // --------------------------------------------------------------------
	// // Payer / Payee
	// // --------------------------------------------------------------------
	// payer, payerPriv := createWallet(t, c)
	// payee, payeePriv := createWallet(t, c)

	// // Allow users
	// c.SetPrivateKey(ownerPriv)
	// _, _ = c.AllowUsers(tok.Address, map[string]bool{
	// 	payer.PublicKey: true,
	// 	payee.PublicKey: true,
	// })

	// // Transfer NFT → payer
	// if _, err := c.TransferToken(
	// 	tok.Address,
	// 	payer.PublicKey,
	// 	"1",
	// 	dec,
	// 	tokenType,
	// 	nftUUID,
	// ); err != nil {
	// 	t.Fatalf("Transfer NFT to payer: %v", err)
	// }

	// // --------------------------------------------------------------------
	// // Create Payment
	// // --------------------------------------------------------------------
	// orderID := fmt.Sprintf("order-nft-%d", time.Now().Unix())
	// exp := time.Now().Add(30 * time.Minute)

	// contractState := models.ContractStateModel{}
	// deployedContract, err := c.DeployContract1(paymentV1.PAYMENT_CONTRACT_V1)
	// if err != nil {
	// 	t.Fatalf("DeployContract: %v", err)
	// }
	// unmarshalState(t, deployedContract.States[0].Object, &contractState)

	// c.SetPrivateKey(payerPriv)
	// created, err := c.CreatePayment(
	// 	contractState.Address,
	// 	tok.Address,
	// 	orderID,
	// 	payer.PublicKey,
	// 	payee.PublicKey,
	// 	"1", // NFT sempre 1
	// 	exp,
	// )
	// if err != nil {
	// 	t.Fatalf("CreatePayment NFT: %v", err)
	// }

	// var pay paymentV1Domain.Payment
	// unmarshalState(t, created.States[0].Object, &pay)
	// if pay.Address == "" {
	// 	t.Fatalf("payment addr empty")
	// }

	// // Allow payment contract
	// c.SetPrivateKey(ownerPriv)
	// _, _ = c.AllowUsers(tok.Address, map[string]bool{pay.Address: true})

	// // --------------------------------------------------------------------
	// // Authorize / Capture / Refund
	// // --------------------------------------------------------------------
	// c.SetPrivateKey(payerPriv)
	// if _, err := c.AuthorizePayment(pay.Address, tokenType, nftUUID); err != nil {
	// 	t.Fatalf("AuthorizePayment NFT: %v", err)
	// }

	// c.SetPrivateKey(payeePriv)
	// if _, err := c.CapturePayment(pay.Address, tokenType, nftUUID); err != nil {
	// 	t.Fatalf("CapturePayment NFT: %v", err)
	// }

	// _, _ = c.RefundPayment(pay.Address, "1", tokenType, nftUUID)

	// // --------------------------------------------------------------------
	// // Direct Pay
	// // --------------------------------------------------------------------
	// contractState = models.ContractStateModel{}
	// deployedContract, err = c.DeployContract1(paymentV1.PAYMENT_CONTRACT_V1)
	// if err != nil {
	// 	t.Fatalf("DeployContract: %v", err)
	// }
	// unmarshalState(t, deployedContract.States[0].Object, &contractState)

	// c.SetPrivateKey(payerPriv)
	// _, _ = c.DirectPay(
	// 	contractState.Address,
	// 	tok.Address,
	// 	orderID+"-direct",
	// 	payer.PublicKey,
	// 	payee.PublicKey,
	// 	"1",
	// 	tokenType,
	// 	nftUUID,
	// )

	// // --------------------------------------------------------------------
	// // Pause / Unpause
	// // --------------------------------------------------------------------
	// c.SetPrivateKey(payerPriv)
	// _, _ = c.PausePayment(pay.Address, true)
	// _, _ = c.UnpausePayment(pay.Address, false)

	// c.SetPrivateKey(payeePriv)
	// _, _ = c.PausePayment(pay.Address, true)
	// _, _ = c.UnpausePayment(pay.Address, false)

	// // --------------------------------------------------------------------
	// // Getters
	// // --------------------------------------------------------------------
	// if _, err := c.GetPayment(pay.Address); err != nil {
	// 	t.Fatalf("GetPayment: %v", err)
	// }
}

func TestPaymentAuthVoidFlow_NonFungible(t *testing.T) {
	// c := setupClient(t)

	// // --------------------------------------------------------------------
	// // Owner + NFT
	// // --------------------------------------------------------------------
	// owner, ownerPriv := createWallet(t, c)
	// c.SetPrivateKey(ownerPriv)

	// dec := 0
	// tokenType := tokenV1Domain.NON_FUNGIBLE
	// stablecoin := false

	// tok := createBasicToken(t, c, owner.PublicKey, dec, true, tokenType, stablecoin)

	// mintOut, err := c.MintToken(tok.Address, owner.PublicKey, "1", dec, tok.TokenType)
	// if err != nil {
	// 	t.Fatalf("MintToken NFT: %v", err)
	// }

	// var mint tokenV1Domain.Mint
	// unmarshalState(t, mintOut.States[0].Object, &mint)
	// nftUUID := mint.TokenUUIDList[0]

	// // --------------------------------------------------------------------
	// // Payer / Payee
	// // --------------------------------------------------------------------
	// payer, payerPriv := createWallet(t, c)
	// payee, _ := createWallet(t, c)

	// c.SetPrivateKey(ownerPriv)
	// _, _ = c.AllowUsers(tok.Address, map[string]bool{
	// 	payer.PublicKey: true,
	// 	payee.PublicKey: true,
	// })

	// // Transfer NFT → payer
	// if _, err := c.TransferToken(
	// 	tok.Address,
	// 	payer.PublicKey,
	// 	"1",
	// 	dec,
	// 	tokenType,
	// 	nftUUID,
	// ); err != nil {
	// 	t.Fatalf("Transfer NFT: %v", err)
	// }

	// // --------------------------------------------------------------------
	// // Create Payment
	// // --------------------------------------------------------------------
	// orderID := fmt.Sprintf("order-nft-void-%d", time.Now().Unix())

	// contractState := models.ContractStateModel{}
	// deployedContract, err := c.DeployContract1(paymentV1.PAYMENT_CONTRACT_V1)
	// if err != nil {
	// 	t.Fatalf("DeployContract: %v", err)
	// }
	// unmarshalState(t, deployedContract.States[0].Object, &contractState)

	// c.SetPrivateKey(payerPriv)
	// created, err := c.CreatePayment(
	// 	contractState.Address,
	// 	tok.Address,
	// 	orderID,
	// 	payer.PublicKey,
	// 	payee.PublicKey,
	// 	"1",
	// 	time.Now().Add(10*time.Minute),
	// )
	// if err != nil {
	// 	t.Fatalf("CreatePayment NFT: %v", err)
	// }

	// var pay paymentV1Domain.Payment
	// unmarshalState(t, created.States[0].Object, &pay)

	// // Allow payment contract
	// c.SetPrivateKey(ownerPriv)
	// _, _ = c.AllowUsers(tok.Address, map[string]bool{pay.Address: true})

	// // --------------------------------------------------------------------
	// // Authorize + Void
	// // --------------------------------------------------------------------
	// c.SetPrivateKey(payerPriv)
	// if _, err := c.AuthorizePayment(pay.Address, tokenType, nftUUID); err != nil {
	// 	t.Fatalf("AuthorizePayment NFT: %v", err)
	// }
	// if _, err := c.VoidPayment(pay.Address, tokenType, nftUUID); err != nil {
	// 	t.Fatalf("VoidPayment NFT: %v", err)
	// }
}
