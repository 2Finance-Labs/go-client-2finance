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
	user, _ := createWallet(t, c)
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

	// IMPORTANTe: agora o faucet precisa existir antes do airdrop
	// Ajuste o nome/assinatura se no seu client for diferente.
	if _, err := c.AddFaucet(
		faucetAddress,
		owner.PublicKey,
		tok.Address,
		start,
		expire,
		false,
		3,
		amt(10, dec),
		2, // claim_interval_seconds
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
		faucetAddress, // <-- NOVO parâmetro obrigatório
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
		verifier.PublicKey, // obrigatório quando manual_review_required = true
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
		ad.FaucetAddress: true, // deve ser o faucetAddress
		user.PublicKey:   true,
	}); err != nil {
		t.Fatalf("AllowUsers(token): %v", err)
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
	// c.SetPrivateKey(userPriv)

	// // Observação: se o seu backend rejeita "data nil" em tx write,
	// // o seu client ClaimAirdrop precisa enviar um payload dummy.
	// // (Se você já corrigiu isso no client, ok.)
	// if _, err := c.ClaimAirdrop(ad.Address); err != nil {
	// 	t.Fatalf("ClaimAirdrop: %v", err)
	// }

	// (Opcional) Double-claim deve falhar
	// if _, err := c.ClaimAirdrop(ad.Address); err == nil {
	// 	t.Fatalf("ClaimAirdrop: expected error on double-claim, got nil")
	// }
}
