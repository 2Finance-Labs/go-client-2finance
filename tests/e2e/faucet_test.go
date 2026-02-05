package e2e_test

import (
	"testing"
	"time"

	"gitlab.com/2finance/2finance-network/blockchain/contract/contractV1/models"
	"gitlab.com/2finance/2finance-network/blockchain/contract/faucetV1"
	faucetV1Domain "gitlab.com/2finance/2finance-network/blockchain/contract/faucetV1/domain"
	tokenV1Domain "gitlab.com/2finance/2finance-network/blockchain/contract/tokenV1/domain"
)

// FAILING TESTS
func TestFaucetFlow(t *testing.T) {

	c := setupClient(t)
	owner, ownerPriv := createWallet(t, c)

	c.SetPrivateKey(ownerPriv)
	dec := 5
	tok := createBasicToken(t, c, owner.PublicKey, dec, true, tokenV1Domain.FUNGIBLE)
	_ = createMint(t, c, tok, owner.PublicKey, "10000", tok.Decimals, tok.TokenType)

	merchant, merchPriv := createWallet(t, c)

	c.SetPrivateKey(ownerPriv)
	_ = createTransfer(t, c, tok, merchant.PublicKey, "50", tok.Decimals, tok.TokenType, "")

	start := time.Now().Add(2 * time.Second)
	exp := time.Now().Add(20 * time.Minute)

	amount := "4"

	contractState := models.ContractStateModel{}
	deployedContract, err := c.DeployContract1(faucetV1.FAUCET_CONTRACT_V1)
	if err != nil {
		t.Fatalf("DeployContract: %v", err)
	}
	unmarshalState(t, deployedContract.States[0].Object, &contractState)
	address := contractState.Address

	out, err := c.AddFaucet(address, merchant.PublicKey, tok.Address, start, exp, false, 3, amount, 2*time.Second)
	if err != nil {
		t.Fatalf("AddFaucet: %v", err)
	}
	var f faucetV1Domain.Faucet
	unmarshalState(t, out.States[0].Object, &f)
	if f.Address == "" {
		t.Fatalf("faucet addr empty")
	}

	// allow faucet (if token is allow-listed in your impl)
	_, err = c.AllowUsers(tok.Address, map[string]bool{f.Address: true})
	if err != nil {
		t.Fatalf("AllowUsers: %v", err)
	}

	c.SetPrivateKey(merchPriv)
	depositAmount := "569"
	if _, err := c.DepositFunds(f.Address, tok.Address, depositAmount, tokenV1Domain.FUNGIBLE, ""); err != nil {
		t.Fatalf("DepositFunds: %v", err)
	}

	// wait for start and claim as a user
	user, userPriv := createWallet(t, c)
	_ = user

	time.Sleep(5 * time.Second)
	c.SetPrivateKey(ownerPriv)
	_, err = c.AllowUsers(tok.Address, map[string]bool{user.PublicKey: true})
	if err != nil {
		t.Fatalf("AllowUsers: %v", err)
	}
	c.SetPrivateKey(userPriv)
	if _, err := c.ClaimFunds(f.Address, tokenV1Domain.FUNGIBLE, ""); err != nil {
		t.Fatalf("ClaimFunds warning: %v", err)
	}

	// pause/unpause & getters
	c.SetPrivateKey(merchPriv)
	if _, err = c.PauseFaucet(f.Address, true); err != nil {
		t.Fatalf("PauseFaucet: %v", err)
	}
	if _, err = c.UnpauseFaucet(f.Address, false); err != nil {
		t.Fatalf("UnpauseFaucet: %v", err)
	}
	if _, err := c.GetFaucet(f.Address); err != nil {
		t.Fatalf("GetFaucet: %v", err)
	}
	if _, err := c.ListFaucets(merchant.PublicKey, 1, 10, true); err != nil {
		t.Fatalf("ListFaucets: %v", err)
	}
}

func TestFaucetFlow_NonFungible(t *testing.T) {
	c := setupClient(t)

	// =========================
	// Owner / Token
	// =========================
	owner, ownerPriv := createWallet(t, c)
	c.SetPrivateKey(ownerPriv)

	dec := 0
	tokenType := tokenV1Domain.NON_FUNGIBLE

	tok := createBasicToken(
		t,
		c,
		owner.PublicKey,
		dec,
		true, // faucet = true
		tokenType,
	)

	// =========================
	// Mint NFT (owner)
	// =========================
	mintOut, err := c.MintToken(
		tok.Address,
		owner.PublicKey,
		"1",
		dec,
		tok.TokenType,
	)
	if err != nil {
		t.Fatalf("MintToken NFT: %v", err)
	}

	var mint tokenV1Domain.Mint
	unmarshalState(t, mintOut.States[0].Object, &mint)

	if len(mint.TokenUUIDList) != 1 {
		t.Fatalf("expected 1 NFT uuid, got %d", len(mint.TokenUUIDList))
	}
	nftUUID := mint.TokenUUIDList[0]

	// =========================
	// Merchant
	// =========================
	merchant, merchPriv := createWallet(t, c)

	// =========================
	// Transfer NFT OWNER → MERCHANT
	// =========================
	c.SetPrivateKey(ownerPriv)

	if _, err := c.TransferToken(
		tok.Address,
		merchant.PublicKey,
		"1",
		dec,
		tokenType,
		nftUUID,
	); err != nil {
		t.Fatalf("Transfer NFT to merchant: %v", err)
	}

	// =========================
	// Deploy Faucet Contract
	// =========================
	start := time.Now().Add(2 * time.Second)
	exp := time.Now().Add(20 * time.Minute)

	contractState := models.ContractStateModel{}
	deployedContract, err := c.DeployContract1(faucetV1.FAUCET_CONTRACT_V1)
	if err != nil {
		t.Fatalf("DeployContract: %v", err)
	}
	unmarshalState(t, deployedContract.States[0].Object, &contractState)
	address := contractState.Address

	// =========================
	// Add Faucet (merchant = owner)
	// =========================
	c.SetPrivateKey(merchPriv)

	out, err := c.AddFaucet(
		address,
		merchant.PublicKey,
		tok.Address,
		start,
		exp,
		false,
		1,      // max claims
		"1",    // amount (NFT sempre 1)
		2*time.Second,
	)
	if err != nil {
		t.Fatalf("AddFaucet NFT: %v", err)
	}

	var f faucetV1Domain.Faucet
	unmarshalState(t, out.States[0].Object, &f)
	if f.Address == "" {
		t.Fatalf("faucet addr empty")
	}

	// =========================
	// Allow faucet to interact with token
	// =========================
	c.SetPrivateKey(ownerPriv)
	if _, err := c.AllowUsers(
		tok.Address,
		map[string]bool{f.Address: true},
	); err != nil {
		t.Fatalf("AllowUsers (faucet): %v", err)
	}

	// =========================
	// Deposit NFT into Faucet (merchant signs)
	// =========================
	c.SetPrivateKey(merchPriv)

	if _, err := c.DepositFunds(
		f.Address,
		tok.Address,
		"1",
		tokenType,
		nftUUID,
	); err != nil {
		t.Fatalf("DepositFunds NFT: %v", err)
	}

	// =========================
	// Wait start
	// =========================
	time.Sleep(5 * time.Second)

	// =========================
	// Create User
	// =========================
	user, userPriv := createWallet(t, c)

	// Allow user to interact with token
	c.SetPrivateKey(ownerPriv)
	if _, err := c.AllowUsers(
		tok.Address,
		map[string]bool{user.PublicKey: true},
	); err != nil {
		t.Fatalf("AllowUsers (user): %v", err)
	}

	// =========================
	// Claim NFT from Faucet (user)
	// =========================
	c.SetPrivateKey(userPriv)

	if _, err := c.ClaimFunds(
		f.Address,
		tokenType,
		nftUUID,
	); err != nil {
		t.Fatalf("ClaimFunds NFT: %v", err)
	}

	// =========================
	// Pause / Unpause (merchant)
	// =========================
	c.SetPrivateKey(merchPriv)

	if _, err := c.PauseFaucet(f.Address, true); err != nil {
		t.Fatalf("PauseFaucet: %v", err)
	}
	if _, err := c.UnpauseFaucet(f.Address, false); err != nil {
		t.Fatalf("UnpauseFaucet: %v", err)
	}

	// =========================
	// Getters
	// =========================
	if _, err := c.GetFaucet(f.Address); err != nil {
		t.Fatalf("GetFaucet: %v", err)
	}
	if _, err := c.ListFaucets(merchant.PublicKey, 1, 10, true); err != nil {
		t.Fatalf("ListFaucets: %v", err)
	}
}
