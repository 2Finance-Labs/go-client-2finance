package e2e_test

import (
	"testing"
	"time"

	"gitlab.com/2finance/2finance-network/blockchain/contract/airdropV1"
	airdropModels "gitlab.com/2finance/2finance-network/blockchain/contract/airdropV1/models"
	"gitlab.com/2finance/2finance-network/blockchain/contract/contractV1/models"
	tokenDomain "gitlab.com/2finance/2finance-network/blockchain/contract/tokenV1/domain"
)

func nonceSafe() uint64 {
	// 28 bits => nunca perde precisão se virar float64 no caminho
	return uint64(time.Now().UnixNano() & 0x0FFFFFFF)
}

func TestAirdropFlow(t *testing.T) {
	c := setupClient(t)

	owner, ownerPriv := createWallet(t, c)
	user, _ := createWallet(t, c)

	// Verifier (obrigatório quando manualReviewRequired=true)
	verifier, _ := createWallet(t, c)

	c.SetPrivateKey(ownerPriv)

	dec := 6
	tokenType := tokenDomain.FUNGIBLE
	tok := createBasicToken(t, c, owner.PublicKey, dec, true, tokenType)

	if _, err := c.MintToken(tok.Address, owner.PublicKey, amt(10_000, dec), dec, tok.TokenType); err != nil {
		t.Fatalf("MintToken: %v", err)
	}

	contractState := models.ContractStateModel{}
	deployed, err := c.DeployContract1(airdropV1.AIRDROP_CONTRACT_V1)
	if err != nil {
		t.Fatalf("DeployContract: %v", err)
	}
	unmarshalState(t, deployed.States[0].Object, &contractState)
	airdropAddress := contractState.Address

	start := time.Now().Add(2 * time.Second)
	expire := time.Now().Add(30 * time.Minute)

	out, err := c.NewAirdrop(
		airdropAddress,
		owner.PublicKey,
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
		verifier.PublicKey, // <<< CORREÇÃO: obrigatório
		true,               // manualReviewRequired=true
		nonceSafe(),
	)
	if err != nil {
		t.Fatalf("NewAirdrop: %v", err)
	}

	var ad airdropModels.AirdropStateModel
	unmarshalState(t, out.States[0].Object, &ad)

	// allowlist token: owner + faucet + user
	if _, err := c.AllowUsers(tok.Address, map[string]bool{
		owner.PublicKey:  true,
		ad.FaucetAddress: true,
		user.PublicKey:   true,
	}); err != nil {
		t.Fatalf("AllowUsers(token): %v", err)
	} 

	// deposit
	if _, err := c.DepositAirdrop(ad.Address, amt(200, dec), tokenType, ""); err != nil {
		t.Fatalf("DepositAirdrop: %v", err)
	}

	// // manual attest (owner é quem pode no seu contract)
	if _, err := c.ManuallyAttestParticipantEligibility(ad.Address, user.PublicKey, true); err != nil {
		t.Fatalf("ManuallyAttestParticipantEligibility: %v", err)
	}

	// // wait start
	waitUntil(t, 15*time.Second, func() bool {
		return time.Now().After(start)
	}) 
	// * Até aqui tudo ok *

	// // claim (user)
	// c.SetPrivateKey(userPriv)
	// if _, err := c.ClaimAirdrop(ad.Address); err != nil {
	// 	t.Fatalf("ClaimAirdrop: %v", err)
	// }

	// // double-claim deve falhar
	// if _, err := c.ClaimAirdrop(ad.Address); err == nil {
	// 	t.Fatalf("ClaimAirdrop: expected error on double-claim, got nil")
	// }

	// --------------------------------------------------------------------
	// Pause / Unpause (owner)
	// --------------------------------------------------------------------
	// c.SetPrivateKey(ownerPriv)
	// if _, err := c.PauseAirdrop(ad.Address); err != nil {
	// 	t.Fatalf("PauseAirdrop: %v", err)
	// }
	// if _, err := c.UnpauseAirdrop(ad.Address); err != nil {
	// 	t.Fatalf("UnpauseAirdrop: %v", err)
	// }

	// --------------------------------------------------------------------
	// List (sanity)
	// --------------------------------------------------------------------
	// if _, err := c.ListAirdrops(owner.PublicKey, 1, 10, true); err != nil {
	// 	t.Fatalf("ListAirdrops: %v", err)
	// }
}
