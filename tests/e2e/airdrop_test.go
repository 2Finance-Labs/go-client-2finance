package e2e_test

import (
	"testing"
	"time"

	// client2f "github.com/2Finance-Labs/go-client-2finance/client_2finance"

	"gitlab.com/2finance/2finance-network/blockchain/contract/airdropV1"
	airdropDomain "gitlab.com/2finance/2finance-network/blockchain/contract/airdropV1/domain"
	"gitlab.com/2finance/2finance-network/blockchain/contract/contractV1/models"
	tokenDomain "gitlab.com/2finance/2finance-network/blockchain/contract/tokenV1/domain"
)

func TestAirdropFlow(t *testing.T) {
	c := setupClient(t)

	owner, ownerPriv := createWallet(t, c)
	// user, userPriv := createWallet(t, c)
	// oracle, oraclePriv := createWallet(t, c)

	// ─────────────────────────────
	// Token setup
	// ─────────────────────────────
	c.SetPrivateKey(ownerPriv)

	dec := 6
	tokenType := tokenDomain.FUNGIBLE
	tok := createBasicToken(t, c, owner.PublicKey, dec, true, tokenType)

	_, err := c.MintToken(tok.Address, owner.PublicKey, amt(10_000, dec), dec, tok.TokenType)
	if err != nil {
		t.Fatalf("MintToken: %v", err)
	}

	// ─────────────────────────────
	// Deploy Airdrop contract
	// ─────────────────────────────
	contractState := models.ContractStateModel{}
	deployed, err := c.DeployContract1(airdropV1.AIRDROP_CONTRACT_V1)
	if err != nil {
		t.Fatalf("DeployContract: %v", err)
	}
	unmarshalState(t, deployed.States[0].Object, &contractState)
	airdropAddress := contractState.Address

	// ─────────────────────────────
	// Create Airdrop
	// ─────────────────────────────
	start := time.Now().Add(2 * time.Second)
	expire := time.Now().Add(30 * time.Minute)

	out, err := c.NewAirdrop(
		airdropAddress,
		owner.PublicKey,
		tok.Address,
		start,
		expire,
		false,
		1,
		amt(10, dec),
		5,
		"Airdrop E2E",
		"E2E description",
		"Short desc",
		"https://img.png",
		"https://banner.png",
		"airdrop",
		map[string]bool{
			"FOLLOW_X": true,
		},
		[]string{"https://x.com/post"},
		"MANUAL",
		"",
		true,
		1,
	)
	if err != nil {
		t.Fatalf("NewAirdrop: %v", err)
	}

	var ad airdropDomain.Airdrop
	unmarshalState(t, out.States[0].Object, &ad)
	if ad.Address == "" {
		t.Fatalf("airdrop address empty")
	}

	// ─────────────────────────────
	// Allow faucet to receive tokens
	// ─────────────────────────────
	// _, err = c.AllowUsers(tok.Address, map[string]bool{ad.FaucetAddress: true})
	// if err != nil {
	// 	t.Fatalf("AllowUsers faucet: %v", err)
	// }

	// ─────────────────────────────
	// Deposit funds
	// ─────────────────────────────
	// _, err = c.DepositAirdrop(
	// 	ad.Address,
	// 	amt(100, dec),
	// 	tokenType,
	// 	"",
	// )
	// if err != nil {
	// 	t.Fatalf("DepositAirdrop: %v", err)
	// }

	// ─────────────────────────────
	// Allow oracle
	// ─────────────────────────────
	// _, err = c.AllowOracles(
	// 	ad.Address,
	// 	map[string]bool{oracle.PublicKey: true},
	// )
	// if err != nil {
	// 	t.Fatalf("AllowOracles: %v", err)
	// }

	// ─────────────────────────────
	// Oracle attests eligibility
	// ─────────────────────────────
	// c.SetPrivateKey(oraclePriv)

	// _, err = c.AttestParticipantEligibility(
	// 	ad.Address,
	// 	user.PublicKey,
	// 	true,
	// )
	// if err != nil {
	// 	t.Fatalf("AttestParticipantEligibility: %v", err)
	// }

	// ─────────────────────────────
	// Claim Airdrop
	// ─────────────────────────────
	// time.Sleep(3 * time.Second)

	// c.SetPrivateKey(userPriv)

	// _, err = c.ClaimAirdrop(ad.Address)
	// if err != nil {
	// 	t.Fatalf("ClaimAirdrop: %v", err)
	// }

	// ─────────────────────────────
	// Pause / Unpause
	// ─────────────────────────────
	// c.SetPrivateKey(ownerPriv)

	// if _, err := c.PauseAirdrop(ad.Address); err != nil {
	// 	t.Fatalf("PauseAirdrop: %v", err)
	// }
	// if _, err := c.UnpauseAirdrop(ad.Address); err != nil {
	// 	t.Fatalf("UnpauseAirdrop: %v", err)
	// }

	// ─────────────────────────────
	// Get / List
	// ─────────────────────────────
	// if _, err := c.GetAirdrop(ad.Address); err != nil {
	// 	t.Fatalf("GetAirdrop: %v", err)
	// }
	// if _, err := c.ListAirdrops(owner.PublicKey, 1, 10, true); err != nil {
	// 	t.Fatalf("ListAirdrops: %v", err)
	//}
}