package e2e_test

import (
	"testing"
	"time"

	"gitlab.com/2finance/2finance-network/blockchain/contract/airdropV1"
	airdropModels "gitlab.com/2finance/2finance-network/blockchain/contract/airdropV1/models"
	"gitlab.com/2finance/2finance-network/blockchain/contract/contractV1/models"
	"gitlab.com/2finance/2finance-network/blockchain/contract/faucetV1"
	tokenDomain "gitlab.com/2finance/2finance-network/blockchain/contract/tokenV1/domain"
)

func TestAirdropFlow(t *testing.T) {
	c := setupClient(t)

	owner, ownerPriv := createWallet(t, c)
	user, userPriv := createWallet(t, c)
	verifier, _ := createWallet(t, c)

	// --------------------------------------------------------------------
	// Token setup
	// --------------------------------------------------------------------
	c.SetPrivateKey(ownerPriv)

	dec := 6
	tokenType := tokenDomain.FUNGIBLE
	tok := createBasicToken(t, c, owner.PublicKey, dec, true, tokenType)

	if _, err := c.MintToken(tok.Address, owner.PublicKey, amt(10_000, dec), dec, tok.TokenType); err != nil {
		t.Fatalf("MintToken: %v", err)
	}

	// --------------------------------------------------------------------
	// Deploy Faucet contract + Create Faucet (obrigatório agora)
	// --------------------------------------------------------------------
	faucetContractState := models.ContractStateModel{}
	faucetDeployed, err := c.DeployContract1(faucetV1.FAUCET_CONTRACT_V1)
	if err != nil {
		t.Fatalf("DeployContract(Faucet): %v", err)
	}
	unmarshalState(t, faucetDeployed.States[0].Object, &faucetContractState)
	faucetAddress := faucetContractState.Address

	start := time.Now().Add(2 * time.Second)
	expire := time.Now().Add(30 * time.Minute)

	if _, err := c.AddFaucet(
		faucetAddress,
		owner.PublicKey,
		tok.Address,
		start,
		expire,
		false,
		3,
		amt(10, dec),
		2,
	); err != nil {
		t.Fatalf("NewFaucet: %v", err)
	}

	// --------------------------------------------------------------------
	// Deploy Airdrop contract
	// --------------------------------------------------------------------
	airdropContractState := models.ContractStateModel{}
	airdropDeployed, err := c.DeployContract1(airdropV1.AIRDROP_CONTRACT_V1)
	if err != nil {
		t.Fatalf("DeployContract(Airdrop): %v", err)
	}
	unmarshalState(t, airdropDeployed.States[0].Object, &airdropContractState)
	airdropAddress := airdropContractState.Address

	// --------------------------------------------------------------------
	// Create Airdrop (agora passa faucetAddress e NÃO tem nonce)
	// --------------------------------------------------------------------
	out, err := c.NewAirdrop(
		airdropAddress,
		owner.PublicKey,
		faucetAddress,
		tok.Address,
		start,
		expire,
		false,
		3,
		amt(10, dec),
		2,
		"Airdrop E2E",
		"E2E description",
		"Short desc",
		"https://img.png",
		"https://banner.png",
		"airdrop",
		map[string]bool{"FOLLOW_X": true},
		[]string{"https://x.com/post"},
		"MANUAL",
		verifier.PublicKey,
		true,
	)
	if err != nil {
		t.Fatalf("NewAirdrop: %v", err)
	}

	var ad airdropModels.AirdropStateModel
	unmarshalState(t, out.States[0].Object, &ad)

	if ad.Address == "" {
		t.Fatalf("airdrop address empty")
	}
	if ad.FaucetAddress == "" {
		t.Fatalf("faucet address empty in airdrop state")
	}

	// --------------------------------------------------------------------
	// Allowlist token: owner + faucet + user
	// --------------------------------------------------------------------
	if _, err := c.AllowUsers(tok.Address, map[string]bool{
		owner.PublicKey:  true,
		ad.FaucetAddress: true,
		user.PublicKey:   true,
	}); err != nil {
		t.Fatalf("AllowUsers(token): %v", err)
	}

	// --------------------------------------------------------------------
	// Pause / Unpause (owner)
	// --------------------------------------------------------------------
	c.SetPrivateKey(ownerPriv)
	if _, err := c.PauseAirdrop(ad.Address); err != nil {
		t.Fatalf("PauseAirdrop: %v", err)
	}
	if _, err := c.UnpauseAirdrop(ad.Address); err != nil {
		t.Fatalf("UnpauseAirdrop: %v", err)
	}

	// --------------------------------------------------------------------
	// Deposit funds (owner)
	// --------------------------------------------------------------------
	if _, err := c.DepositAirdrop(ad.Address, amt(200, dec), tokenType, ""); err != nil {
		t.Fatalf("DepositAirdrop: %v", err)
	}

	// --------------------------------------------------------------------
	// Manual attest (owner)
	// --------------------------------------------------------------------
	if _, err := c.ManuallyAttestParticipantEligibility(ad.Address, user.PublicKey, true); err != nil {
		t.Fatalf("ManuallyAttestParticipantEligibility: %v", err)
	}

	// --------------------------------------------------------------------
	// Wait start
	// --------------------------------------------------------------------
	waitUntil(t, 15*time.Second, func() bool {
		return time.Now().After(start)
	})

	// --------------------------------------------------------------------
	// Claim (user)
	// --------------------------------------------------------------------
	c.SetPrivateKey(userPriv)

	if _, err := c.ClaimAirdrop(ad.Address, tok.TokenType); err != nil {
		t.Fatalf("ClaimAirdrop: %v", err)
	}

	// Double-claim deve falhar
	if _, err := c.ClaimAirdrop(ad.Address, tok.TokenType); err == nil {
		t.Fatalf("ClaimAirdrop: expected error on double-claim, got nil")
	}

	// --------------------------------------------------------------------
	// Withdraw remaining funds (owner)
	// --------------------------------------------------------------------
	time.Sleep(2 * time.Second)

	c.SetPrivateKey(ownerPriv)

	if _, err := c.WithdrawAirdropFunds(ad.Address, amt(50, dec), tokenType, ""); err != nil {
		t.Fatalf("WithdrawAirdropFunds: %v", err)
	}

	// --------------------------------------------------------------------
	// Update metadata (owner) - cobre METHOD_UPDATE_AIRDROP_METADATA
	// --------------------------------------------------------------------
	c.SetPrivateKey(ownerPriv) 

	newTitle := "Airdrop E2E (UPDATED)"
	newDesc := "Updated description"
	newShort := "Updated short"
	newImg := "https://img-updated.png"
	newBanner := "https://banner-updated.png"
	newCategory := "airdrop"

	newReq := map[string]bool{"FOLLOW_X": true, "LIKE_X": true}
	newLinks := []string{"https://x.com/post/updated"}

	newVerificationType := "ORACLE"

	newManualReviewRequired := true
	newVerifier := verifier.PublicKey

	outMeta, err := c.UpdateAirdropMetadata(
		ad.Address,
		newTitle,
		newDesc,
		newShort,
		newImg,
		newBanner,
		newCategory,
		newReq,
		newLinks,
		newVerificationType,
		newVerifier,
		newManualReviewRequired,
	)
	if err != nil {
		t.Fatalf("UpdateAirdropMetadata: %v", err)
	}

	var adUpdated airdropModels.AirdropStateModel
	unmarshalState(t, outMeta.States[0].Object, &adUpdated)

	if adUpdated.Title != newTitle {
		t.Fatalf("metadata not updated: title=%q want=%q", adUpdated.Title, newTitle)
	}
	if adUpdated.ShortDescription != newShort {
		t.Fatalf("metadata not updated: short=%q want=%q", adUpdated.ShortDescription, newShort)
	}
	if adUpdated.VerificationType != newVerificationType {
		t.Fatalf("metadata not updated: verification_type=%q want=%q", adUpdated.VerificationType, newVerificationType)
	}

	// --------------------------------------------------------------------
	// Allow oracle (owner)
	// --------------------------------------------------------------------
	oracle, oraclePriv := createWallet(t, c)
	userOracle, userOraclePriv := createWallet(t, c)

	c.SetPrivateKey(ownerPriv)
	if _, err := c.AllowOracles(ad.Address, map[string]bool{
		oracle.PublicKey: true,
	}); err != nil {
		t.Fatalf("AllowOracles: %v", err)
	}

	if _, err := c.AllowUsers(tok.Address, map[string]bool{
		userOracle.PublicKey: true,
	}); err != nil {
		t.Fatalf("AllowUsers(token userOracle): %v", err)
	}

	// --------------------------------------------------------------------
	// Attest eligibility (oracle)
	// --------------------------------------------------------------------
	c.SetPrivateKey(oraclePriv)
	if _, err := c.AttestParticipantEligibility(ad.Address, userOracle.PublicKey, true); err != nil {
		t.Fatalf("AttestParticipantEligibility: %v", err)
	}

	// --------------------------------------------------------------------
	// Claim (user)
	// --------------------------------------------------------------------
	c.SetPrivateKey(userOraclePriv)

	if _, err := c.ClaimAirdrop(ad.Address, tok.TokenType); err != nil {
		t.Fatalf("ClaimAirdrop: %v", err)
	}

	// Double-claim deve falhar
	if _, err := c.ClaimAirdrop(ad.Address, tok.TokenType); err == nil {
		t.Fatalf("ClaimAirdrop: expected error on double-claim, got nil")
	}

	// --------------------------------------------------------------------
	// Withdraw remaining funds (owner)
	// --------------------------------------------------------------------
	time.Sleep(2 * time.Second)

	c.SetPrivateKey(ownerPriv)

	if _, err := c.WithdrawAirdropFunds(ad.Address, amt(50, dec), tokenType, ""); err != nil {
		t.Fatalf("WithdrawAirdropFunds: %v", err)
	}
}
