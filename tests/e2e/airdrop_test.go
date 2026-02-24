package e2e_test

import (
	"log"
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
	// Token setup (validate + log)
	// --------------------------------------------------------------------
	c.SetPrivateKey(ownerPriv)

	dec := 6
	tokenType := tokenDomain.FUNGIBLE
	stablecoin := false
	tok := createBasicToken(t, c, owner.PublicKey, dec, true, tokenType, stablecoin)

	if tok.Address == "" {
		t.Fatalf("token address empty")
	}
	if tok.Symbol == "" {
		t.Fatalf("token symbol empty")
	}
	if tok.Name == "" {
		t.Fatalf("token name empty")
	}
	if tok.Decimals != dec {
		t.Fatalf("token decimals mismatch: got %d want %d", tok.Decimals, dec)
	}
	if tok.TokenType != tokenType {
		t.Fatalf("token type mismatch: got %s want %s", tok.TokenType, tokenType)
	}
	if tok.Stablecoin != stablecoin {
		t.Fatalf("token stablecoin mismatch: got %v want %v", tok.Stablecoin, stablecoin)
	}
	if tok.Creator == "" {
		t.Fatalf("token creator empty")
	}

	log.Printf("Token Address: %s", tok.Address)
	log.Printf("Token Symbol: %s", tok.Symbol)
	log.Printf("Token Name: %s", tok.Name)
	log.Printf("Token Decimals: %d", tok.Decimals)
	log.Printf("Token Total Supply: %s", tok.TotalSupply)
	log.Printf("Token Description: %s", tok.Description)
	log.Printf("Token Image: %s", tok.Image)
	log.Printf("Token Website: %s", tok.Website)
	log.Printf("Token Tags Social: %+v", tok.TagsSocialMedia)
	log.Printf("Token Tags Category: %+v", tok.TagsCategory)
	log.Printf("Token Tags: %+v", tok.Tags)
	log.Printf("Token Creator: %s", tok.Creator)
	log.Printf("Token Creator Website: %s", tok.CreatorWebsite)
	log.Printf("Token Access Policy Mode: %s", tok.AccessPolicy.Mode)
	log.Printf("Token Access Policy Users: %+v", tok.AccessPolicy.Users)
	log.Printf("Token Frozen Accounts: %+v", tok.FrozenAccounts)
	log.Printf("Token Fee Tiers: %+v", tok.FeeTiersList)
	log.Printf("Token Fee Address: %s", tok.FeeAddress)
	log.Printf("Token Freeze Authority Revoked: %v", tok.FreezeAuthorityRevoked)
	log.Printf("Token Mint Authority Revoked: %v", tok.MintAuthorityRevoked)
	log.Printf("Token Update Authority Revoked: %v", tok.UpdateAuthorityRevoked)
	log.Printf("Token Paused: %v", tok.Paused)
	log.Printf("Token Expired At: %s", tok.ExpiredAt.String())
	log.Printf("Token Asset GLB URI: %s", tok.AssetGLBUri)
	log.Printf("Token Type: %s", tok.TokenType)
	log.Printf("Token Transferable: %v", tok.Transferable)
	log.Printf("Token Stablecoin: %v", tok.Stablecoin)

	// --------------------------------------------------------------------
	// Mint (envelope + validate + log)
	// --------------------------------------------------------------------
	mintOut, err := c.MintToken(tok.Address, owner.PublicKey, amt(10_000, dec), dec, tok.TokenType)
	if err != nil {
		t.Fatalf("MintToken: %v", err)
	}
	if len(mintOut.States) == 0 {
		t.Fatalf("MintToken returned empty States")
	}
	if mintOut.States[0].Object == nil {
		t.Fatalf("MintToken returned nil state object")
	}

	// opcional: se você tiver struct de Mint no tokenDomain, dá para validar também.
	// aqui eu mantenho no padrão "envelope" + logs, sem inventar campos.

	log.Printf("Mint Output States: %+v", mintOut.States)
	log.Printf("Mint Output Logs: %+v", mintOut.Logs)
	log.Printf("Mint Output Delegated Call: %+v", mintOut.DelegatedCall)

	// --------------------------------------------------------------------
	// Deploy Faucet contract (envelope + unmarshal + validate + log)
	// --------------------------------------------------------------------
	faucetContractState := models.ContractStateModel{}
	faucetDeployed, err := c.DeployContract1(faucetV1.FAUCET_CONTRACT_V1)
	if err != nil {
		t.Fatalf("DeployContract(Faucet): %v", err)
	}
	if len(faucetDeployed.States) == 0 {
		t.Fatalf("DeployContract(Faucet) returned empty States")
	}
	if faucetDeployed.States[0].Object == nil {
		t.Fatalf("DeployContract(Faucet) returned nil state object")
	}

	unmarshalState(t, faucetDeployed.States[0].Object, &faucetContractState)

	if faucetContractState.Address == "" {
		t.Fatalf("DeployContract(Faucet) returned empty contract address")
	}
	faucetAddress := faucetContractState.Address

	log.Printf("DeployContract(Faucet) Output States: %+v", faucetDeployed.States)
	log.Printf("DeployContract(Faucet) Output Logs: %+v", faucetDeployed.Logs)
	log.Printf("DeployContract(Faucet) Output Delegated Call: %+v", faucetDeployed.DelegatedCall)

	log.Printf("Faucet Contract Address: %s", faucetAddress)

	// --------------------------------------------------------------------
	// Create Faucet (envelope + validate + log)
	// --------------------------------------------------------------------
	start := time.Now().Add(2 * time.Second)
	expire := time.Now().Add(30 * time.Minute)

	addFaucetOut, err := c.AddFaucet(
		faucetAddress,
		owner.PublicKey,
		tok.Address,
		start,
		expire,
		false,
		3,
		amt(10, dec),
		2,
	)
	if err != nil {
		t.Fatalf("NewFaucet: %v", err)
	}
	if len(addFaucetOut.States) == 0 {
		t.Fatalf("AddFaucet returned empty States")
	}
	if addFaucetOut.States[0].Object == nil {
		t.Fatalf("AddFaucet returned nil state object")
	}

	log.Printf("AddFaucet Output States: %+v", addFaucetOut.States)
	log.Printf("AddFaucet Output Logs: %+v", addFaucetOut.Logs)
	log.Printf("AddFaucet Output Delegated Call: %+v", addFaucetOut.DelegatedCall)

	// --------------------------------------------------------------------
	// Deploy Airdrop contract (envelope + unmarshal + validate + log)
	// --------------------------------------------------------------------
	airdropContractState := models.ContractStateModel{}
	airdropDeployed, err := c.DeployContract1(airdropV1.AIRDROP_CONTRACT_V1)
	if err != nil {
		t.Fatalf("DeployContract(Airdrop): %v", err)
	}
	if len(airdropDeployed.States) == 0 {
		t.Fatalf("DeployContract(Airdrop) returned empty States")
	}
	if airdropDeployed.States[0].Object == nil {
		t.Fatalf("DeployContract(Airdrop) returned nil state object")
	}

	unmarshalState(t, airdropDeployed.States[0].Object, &airdropContractState)

	if airdropContractState.Address == "" {
		t.Fatalf("DeployContract(Airdrop) returned empty contract address")
	}
	airdropAddress := airdropContractState.Address

	log.Printf("DeployContract(Airdrop) Output States: %+v", airdropDeployed.States)
	log.Printf("DeployContract(Airdrop) Output Logs: %+v", airdropDeployed.Logs)
	log.Printf("DeployContract(Airdrop) Output Delegated Call: %+v", airdropDeployed.DelegatedCall)

	log.Printf("Airdrop Contract Address: %s", airdropAddress)

	// --------------------------------------------------------------------
	// Create Airdrop (envelope + unmarshal + validate + log)
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
	if len(out.States) == 0 {
		t.Fatalf("NewAirdrop returned empty States")
	}
	if out.States[0].Object == nil {
		t.Fatalf("NewAirdrop returned nil state object")
	}

	var ad airdropModels.AirdropStateModel
	unmarshalState(t, out.States[0].Object, &ad)

	if ad.Address == "" {
		t.Fatalf("NewAirdrop returned empty airdrop address")
	}
	if ad.Owner != owner.PublicKey {
		t.Fatalf("NewAirdrop owner mismatch: got %q want %q", ad.Owner, owner.PublicKey)
	}
	if ad.FaucetAddress != faucetAddress {
		t.Fatalf("NewAirdrop faucet_address mismatch: got %q want %q", ad.FaucetAddress, faucetAddress)
	}
	if ad.TokenAddress != tok.Address {
		t.Fatalf("NewAirdrop token_address mismatch: got %q want %q", ad.TokenAddress, tok.Address)
	}

	log.Printf("NewAirdrop Output States: %+v", out.States)
	log.Printf("NewAirdrop Output Logs: %+v", out.Logs)
	log.Printf("NewAirdrop Output Delegated Call: %+v", out.DelegatedCall)

	log.Printf("Airdrop Address: %s", ad.Address)
	log.Printf("Airdrop Owner: %s", ad.Owner)
	log.Printf("Airdrop FaucetAddress: %s", ad.FaucetAddress)
	log.Printf("Airdrop TokenAddress: %s", ad.TokenAddress)
	log.Printf("Airdrop Title: %s", ad.Title)
	log.Printf("Airdrop ShortDescription: %s", ad.ShortDescription)
	log.Printf("Airdrop VerificationType: %s", ad.VerificationType)
	log.Printf("Airdrop Verifier: %s", ad.VerifierPublicKey)

	// --------------------------------------------------------------------
	// GET Airdrop (envelope + unmarshal + validate + log)
	// --------------------------------------------------------------------
	gotOut, err := c.GetAirdrop(ad.Address)
	if err != nil {
		t.Fatalf("GetAirdrop: %v", err)
	}
	if len(gotOut.States) == 0 {
		t.Fatalf("GetAirdrop returned empty States")
	}
	if gotOut.States[0].Object == nil {
		t.Fatalf("GetAirdrop returned nil state object")
	}

	var adGet airdropModels.AirdropStateModel
	unmarshalState(t, gotOut.States[0].Object, &adGet)

	if adGet.Address != ad.Address {
		t.Fatalf("GetAirdrop mismatch: address=%q want=%q", adGet.Address, ad.Address)
	}
	if adGet.FaucetAddress != ad.FaucetAddress {
		t.Fatalf("GetAirdrop mismatch: faucet_address=%q want=%q", adGet.FaucetAddress, ad.FaucetAddress)
	}
	if adGet.TokenAddress != ad.TokenAddress {
		t.Fatalf("GetAirdrop mismatch: token_address=%q want=%q", adGet.TokenAddress, ad.TokenAddress)
	}

	log.Printf("GetAirdrop Output States: %+v", gotOut.States)
	log.Printf("GetAirdrop Output Logs: %+v", gotOut.Logs)
	log.Printf("GetAirdrop Output Delegated Call: %+v", gotOut.DelegatedCall)

	log.Printf("GetAirdrop Address: %s", adGet.Address)
	log.Printf("GetAirdrop FaucetAddress: %s", adGet.FaucetAddress)
	log.Printf("GetAirdrop TokenAddress: %s", adGet.TokenAddress)

	// --------------------------------------------------------------------
	// LIST Airdrops (envelope + unmarshal + validate + log)
	// --------------------------------------------------------------------
	listOut, err := c.ListAirdrops(owner.PublicKey, 1, 50, false)
	if err != nil {
		t.Fatalf("ListAirdrops: %v", err)
	}
	if len(listOut.States) == 0 {
		t.Fatalf("ListAirdrops returned empty States")
	}
	if listOut.States[0].Object == nil {
		t.Fatalf("ListAirdrops returned nil state object")
	}

	var list []airdropModels.AirdropStateModel
	unmarshalState(t, listOut.States[0].Object, &list)

	found := false
	for _, it := range list {
		if it.Address == ad.Address {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("ListAirdrops: created airdrop %s not found in list", ad.Address)
	}

	log.Printf("ListAirdrops Output States: %+v", listOut.States)
	log.Printf("ListAirdrops Output Logs: %+v", listOut.Logs)
	log.Printf("ListAirdrops Output Delegated Call: %+v", listOut.DelegatedCall)

	log.Printf("ListAirdrops Count: %d", len(list))

	// --------------------------------------------------------------------
	// Allowlist token: owner + faucet + user (envelope + validate + log)
	// --------------------------------------------------------------------
	allowTokOut, err := c.AllowUsers(tok.Address, map[string]bool{
		owner.PublicKey:  true,
		ad.FaucetAddress: true,
		user.PublicKey:   true,
	})
	if err != nil {
		t.Fatalf("AllowUsers(token): %v", err)
	}
	if len(allowTokOut.States) == 0 {
		t.Fatalf("AllowUsers(token) returned empty States")
	}
	if allowTokOut.States[0].Object == nil {
		t.Fatalf("AllowUsers(token) returned nil state object")
	}

	// se o retorno for AccessPolicy, valida e loga no padrão (unmarshal + campos)
	var apTok tokenDomain.AccessPolicy
	unmarshalState(t, allowTokOut.States[0].Object, &apTok)

	if apTok.Mode == "" {
		t.Fatalf("AllowUsers(token) returned empty mode")
	}
	if apTok.Users == nil {
		t.Fatalf("AllowUsers(token) returned nil users")
	}
	if !apTok.Users[owner.PublicKey] || !apTok.Users[user.PublicKey] || !apTok.Users[ad.FaucetAddress] {
		t.Fatalf("AllowUsers(token) missing expected users in allowlist")
	}

	log.Printf("AllowUsers(token) Output States: %+v", allowTokOut.States)
	log.Printf("AllowUsers(token) Output Logs: %+v", allowTokOut.Logs)
	log.Printf("AllowUsers(token) Output Delegated Call: %+v", allowTokOut.DelegatedCall)

	log.Printf("AllowUsers(token) Mode: %s", apTok.Mode)
	log.Printf("AllowUsers(token) Users: %+v", apTok.Users)

	// --------------------------------------------------------------------
	// Pause / Unpause (owner) (envelope + validate + log)
	// --------------------------------------------------------------------
	c.SetPrivateKey(ownerPriv)

	pauseOut, err := c.PauseAirdrop(ad.Address)
	if err != nil {
		t.Fatalf("PauseAirdrop: %v", err)
	}
	if len(pauseOut.States) == 0 {
		t.Fatalf("PauseAirdrop returned empty States")
	}
	if pauseOut.States[0].Object == nil {
		t.Fatalf("PauseAirdrop returned nil state object")
	}

	var adPaused airdropModels.AirdropStateModel
	unmarshalState(t, pauseOut.States[0].Object, &adPaused)

	if adPaused.Address != ad.Address {
		t.Fatalf("PauseAirdrop address mismatch: got %q want %q", adPaused.Address, ad.Address)
	}
	if !adPaused.Paused {
		t.Fatalf("PauseAirdrop expected Paused=true")
	}

	log.Printf("PauseAirdrop Output States: %+v", pauseOut.States)
	log.Printf("PauseAirdrop Output Logs: %+v", pauseOut.Logs)
	log.Printf("PauseAirdrop Output Delegated Call: %+v", pauseOut.DelegatedCall)

	log.Printf("PauseAirdrop Address: %s", adPaused.Address)
	log.Printf("PauseAirdrop Paused: %v", adPaused.Paused)

	unpauseOut, err := c.UnpauseAirdrop(ad.Address)
	if err != nil {
		t.Fatalf("UnpauseAirdrop: %v", err)
	}
	if len(unpauseOut.States) == 0 {
		t.Fatalf("UnpauseAirdrop returned empty States")
	}
	if unpauseOut.States[0].Object == nil {
		t.Fatalf("UnpauseAirdrop returned nil state object")
	}

	var adUnpaused airdropModels.AirdropStateModel
	unmarshalState(t, unpauseOut.States[0].Object, &adUnpaused)

	if adUnpaused.Address != ad.Address {
		t.Fatalf("UnpauseAirdrop address mismatch: got %q want %q", adUnpaused.Address, ad.Address)
	}
	if adUnpaused.Paused {
		t.Fatalf("UnpauseAirdrop expected Paused=false")
	}

	log.Printf("UnpauseAirdrop Output States: %+v", unpauseOut.States)
	log.Printf("UnpauseAirdrop Output Logs: %+v", unpauseOut.Logs)
	log.Printf("UnpauseAirdrop Output Delegated Call: %+v", unpauseOut.DelegatedCall)

	log.Printf("UnpauseAirdrop Address: %s", adUnpaused.Address)
	log.Printf("UnpauseAirdrop Paused: %v", adUnpaused.Paused)

	// --------------------------------------------------------------------
	// Deposit funds (owner) (envelope + validate + log)
	// --------------------------------------------------------------------
	depOut, err := c.DepositAirdrop(ad.Address, amt(200, dec), tokenType, "")
	if err != nil {
		t.Fatalf("DepositAirdrop: %v", err)
	}
	if len(depOut.States) == 0 {
		t.Fatalf("DepositAirdrop returned empty States")
	}
	if depOut.States[0].Object == nil {
		t.Fatalf("DepositAirdrop returned nil state object")
	}

	// se existir struct específica (ex.: Deposit), você pode trocar aqui.
	var adAfterDeposit airdropModels.AirdropStateModel
	unmarshalState(t, depOut.States[0].Object, &adAfterDeposit)

	if adAfterDeposit.Address != ad.Address {
		t.Fatalf("DepositAirdrop address mismatch: got %q want %q", adAfterDeposit.Address, ad.Address)
	}

	log.Printf("DepositAirdrop Output States: %+v", depOut.States)
	log.Printf("DepositAirdrop Output Logs: %+v", depOut.Logs)
	log.Printf("DepositAirdrop Output Delegated Call: %+v", depOut.DelegatedCall)

	log.Printf("DepositAirdrop Address: %s", adAfterDeposit.Address)

	// --------------------------------------------------------------------
	// Manual attest (owner) (envelope + validate + log)
	// --------------------------------------------------------------------
	attManualOut, err := c.ManuallyAttestParticipantEligibility(ad.Address, user.PublicKey, true)
	if err != nil {
		t.Fatalf("ManuallyAttestParticipantEligibility: %v", err)
	}
	if len(attManualOut.States) == 0 {
		t.Fatalf("ManuallyAttestParticipantEligibility returned empty States")
	}
	if attManualOut.States[0].Object == nil {
		t.Fatalf("ManuallyAttestParticipantEligibility returned nil state object")
	}

	var attManual airdropModels.AirdropStateModel
	unmarshalState(t, attManualOut.States[0].Object, &attManual)

	if attManual.Address != ad.Address {
		t.Fatalf("ManualAttest airdrop mismatch: got %q want %q", attManual.Address, ad.Address)
	}
	if attManual.Owner != owner.PublicKey {
		t.Fatalf("ManualAttest wallet mismatch: got %q want %q", attManual.Owner, owner.PublicKey)
	}
	if !attManual.EligibleWallets[owner.PublicKey] {
		t.Fatalf("ManualAttest expected Eligible=true")
	}

	log.Printf("ManuallyAttest Output States: %+v", attManualOut.States)
	log.Printf("ManuallyAttest Output Logs: %+v", attManualOut.Logs)
	log.Printf("ManuallyAttest Output Delegated Call: %+v", attManualOut.DelegatedCall)

	log.Printf("ManuallyAttest AirdropAddress: %s", attManual.Address)
	log.Printf("ManuallyAttest WalletAddress: %s", attManual.Owner)
	log.Printf("ManuallyAttest Eligible: %v", attManual.EligibleWallets[owner.PublicKey])

	// --------------------------------------------------------------------
	// Wait start
	// --------------------------------------------------------------------
	waitUntil(t, 15*time.Second, func() bool {
		return time.Now().After(start)
	})

	// --------------------------------------------------------------------
	// Claim (user) (envelope + validate + log)
	// --------------------------------------------------------------------
	c.SetPrivateKey(userPriv)

	claimOut, err := c.ClaimAirdrop(ad.Address, tok.TokenType)
	if err != nil {
		t.Fatalf("ClaimAirdrop: %v", err)
	}
	if len(claimOut.States) == 0 {
		t.Fatalf("ClaimAirdrop returned empty States")
	}
	if claimOut.States[0].Object == nil {
		t.Fatalf("ClaimAirdrop returned nil state object")
	}

	// geralmente Claim retorna estado do airdrop ou algo de claim; aqui mantenho genérico
	var adAfterClaim airdropModels.AirdropStateModel
	unmarshalState(t, claimOut.States[0].Object, &adAfterClaim)

	if adAfterClaim.Address != ad.Address {
		t.Fatalf("ClaimAirdrop address mismatch: got %q want %q", adAfterClaim.Address, ad.Address)
	}

	log.Printf("ClaimAirdrop Output States: %+v", claimOut.States)
	log.Printf("ClaimAirdrop Output Logs: %+v", claimOut.Logs)
	log.Printf("ClaimAirdrop Output Delegated Call: %+v", claimOut.DelegatedCall)

	log.Printf("ClaimAirdrop Address: %s", adAfterClaim.Address)

	// Double-claim deve falhar
	if _, err := c.ClaimAirdrop(ad.Address, tok.TokenType); err == nil {
		t.Fatalf("ClaimAirdrop: expected error on double-claim, got nil")
	}

	// --------------------------------------------------------------------
	// Withdraw remaining funds (owner) (envelope + validate + log)
	// --------------------------------------------------------------------
	time.Sleep(2 * time.Second)

	c.SetPrivateKey(ownerPriv)

	withOut, err := c.WithdrawAirdropFunds(ad.Address, amt(50, dec), tokenType, "")
	if err != nil {
		t.Fatalf("WithdrawAirdropFunds: %v", err)
	}
	if len(withOut.States) == 0 {
		t.Fatalf("WithdrawAirdropFunds returned empty States")
	}
	if withOut.States[0].Object == nil {
		t.Fatalf("WithdrawAirdropFunds returned nil state object")
	}

	var adAfterWithdraw airdropModels.AirdropStateModel
	unmarshalState(t, withOut.States[0].Object, &adAfterWithdraw)

	if adAfterWithdraw.Address != ad.Address {
		t.Fatalf("WithdrawAirdropFunds address mismatch: got %q want %q", adAfterWithdraw.Address, ad.Address)
	}

	log.Printf("WithdrawAirdropFunds Output States: %+v", withOut.States)
	log.Printf("WithdrawAirdropFunds Output Logs: %+v", withOut.Logs)
	log.Printf("WithdrawAirdropFunds Output Delegated Call: %+v", withOut.DelegatedCall)

	log.Printf("WithdrawAirdropFunds Address: %s", adAfterWithdraw.Address)

	// --------------------------------------------------------------------
	// Update metadata (owner) (envelope + unmarshal + validate + log)
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
	if len(outMeta.States) == 0 {
		t.Fatalf("UpdateAirdropMetadata returned empty States")
	}
	if outMeta.States[0].Object == nil {
		t.Fatalf("UpdateAirdropMetadata returned nil state object")
	}

	var adUpdated airdropModels.AirdropStateModel
	unmarshalState(t, outMeta.States[0].Object, &adUpdated)

	if adUpdated.Address != ad.Address {
		t.Fatalf("UpdateAirdropMetadata address mismatch: got %q want %q", adUpdated.Address, ad.Address)
	}
	if adUpdated.Title != newTitle {
		t.Fatalf("UpdateAirdropMetadata title mismatch: got %q want %q", adUpdated.Title, newTitle)
	}
	if adUpdated.ShortDescription != newShort {
		t.Fatalf("UpdateAirdropMetadata short mismatch: got %q want %q", adUpdated.ShortDescription, newShort)
	}
	if adUpdated.VerificationType != newVerificationType {
		t.Fatalf("UpdateAirdropMetadata verification_type mismatch: got %q want %q", adUpdated.VerificationType, newVerificationType)
	}
	if adUpdated.VerifierPublicKey != newVerifier {
		t.Fatalf("UpdateAirdropMetadata verifier mismatch: got %q want %q", adUpdated.VerifierPublicKey, newVerifier)
	}
	if adUpdated.ManualReviewRequired != newManualReviewRequired {
		t.Fatalf("UpdateAirdropMetadata manual_review_required mismatch: got %v want %v", adUpdated.ManualReviewRequired, newManualReviewRequired)
	}

	log.Printf("UpdateAirdropMetadata Output States: %+v", outMeta.States)
	log.Printf("UpdateAirdropMetadata Output Logs: %+v", outMeta.Logs)
	log.Printf("UpdateAirdropMetadata Output Delegated Call: %+v", outMeta.DelegatedCall)

	log.Printf("UpdateAirdropMetadata Address: %s", adUpdated.Address)
	log.Printf("UpdateAirdropMetadata Title: %s", adUpdated.Title)
	log.Printf("UpdateAirdropMetadata ShortDescription: %s", adUpdated.ShortDescription)
	log.Printf("UpdateAirdropMetadata VerificationType: %s", adUpdated.VerificationType)
	log.Printf("UpdateAirdropMetadata Verifier: %s", adUpdated.VerifierPublicKey)
	log.Printf("UpdateAirdropMetadata ManualReviewRequired: %v", adUpdated.ManualReviewRequired)

	// --------------------------------------------------------------------
	// GET pós-update (envelope + unmarshal + validate + log)
	// --------------------------------------------------------------------
	gotOut2, err := c.GetAirdrop(ad.Address)
	if err != nil {
		t.Fatalf("GetAirdrop(post-update): %v", err)
	}
	if len(gotOut2.States) == 0 {
		t.Fatalf("GetAirdrop(post-update) returned empty States")
	}
	if gotOut2.States[0].Object == nil {
		t.Fatalf("GetAirdrop(post-update) returned nil state object")
	}

	var adGet2 airdropModels.AirdropStateModel
	unmarshalState(t, gotOut2.States[0].Object, &adGet2)

	if adGet2.Title != newTitle {
		t.Fatalf("GetAirdrop(post-update) mismatch: title=%q want=%q", adGet2.Title, newTitle)
	}
	if adGet2.ShortDescription != newShort {
		t.Fatalf("GetAirdrop(post-update) mismatch: short=%q want=%q", adGet2.ShortDescription, newShort)
	}
	if adGet2.VerificationType != newVerificationType {
		t.Fatalf("GetAirdrop(post-update) mismatch: verification_type=%q want=%q", adGet2.VerificationType, newVerificationType)
	}

	log.Printf("GetAirdrop(post-update) Output States: %+v", gotOut2.States)
	log.Printf("GetAirdrop(post-update) Output Logs: %+v", gotOut2.Logs)
	log.Printf("GetAirdrop(post-update) Output Delegated Call: %+v", gotOut2.DelegatedCall)

	log.Printf("GetAirdrop(post-update) Title: %s", adGet2.Title)
	log.Printf("GetAirdrop(post-update) ShortDescription: %s", adGet2.ShortDescription)
	log.Printf("GetAirdrop(post-update) VerificationType: %s", adGet2.VerificationType)

	// --------------------------------------------------------------------
	// Allow oracle (owner) (envelope + validate + log)
	// --------------------------------------------------------------------
	oracle, oraclePriv := createWallet(t, c)
	userOracle, userOraclePriv := createWallet(t, c)

	c.SetPrivateKey(ownerPriv)

	allowOracleOut, err := c.AllowOracles(ad.Address, map[string]bool{
		oracle.PublicKey: true,
	})
	if err != nil {
		t.Fatalf("AllowOracles: %v", err)
	}
	if len(allowOracleOut.States) == 0 {
		t.Fatalf("AllowOracles returned empty States")
	}
	if allowOracleOut.States[0].Object == nil {
		t.Fatalf("AllowOracles returned nil state object")
	}

	// se houver model específico de oracle allowlist, troque; aqui valido pelo menos leitura básica
	var adAfterAllowOracle airdropModels.AirdropStateModel
	unmarshalState(t, allowOracleOut.States[0].Object, &adAfterAllowOracle)

	if adAfterAllowOracle.Address != ad.Address {
		t.Fatalf("AllowOracles address mismatch: got %q want %q", adAfterAllowOracle.Address, ad.Address)
	}

	log.Printf("AllowOracles Output States: %+v", allowOracleOut.States)
	log.Printf("AllowOracles Output Logs: %+v", allowOracleOut.Logs)
	log.Printf("AllowOracles Output Delegated Call: %+v", allowOracleOut.DelegatedCall)

	log.Printf("AllowOracles AirdropAddress: %s", adAfterAllowOracle.Address)
	log.Printf("AllowOracles Oracle: %s", oracle.PublicKey)

	allowUserOracleTokOut, err := c.AllowUsers(tok.Address, map[string]bool{
		userOracle.PublicKey: true,
	})
	if err != nil {
		t.Fatalf("AllowUsers(token userOracle): %v", err)
	}
	if len(allowUserOracleTokOut.States) == 0 {
		t.Fatalf("AllowUsers(token userOracle) returned empty States")
	}
	if allowUserOracleTokOut.States[0].Object == nil {
		t.Fatalf("AllowUsers(token userOracle) returned nil state object")
	}

	var apTok2 tokenDomain.AccessPolicy
	unmarshalState(t, allowUserOracleTokOut.States[0].Object, &apTok2)

	if apTok2.Users == nil || !apTok2.Users[userOracle.PublicKey] {
		t.Fatalf("AllowUsers(token userOracle) missing userOracle in allowlist")
	}

	log.Printf("AllowUsers(token userOracle) Output States: %+v", allowUserOracleTokOut.States)
	log.Printf("AllowUsers(token userOracle) Output Logs: %+v", allowUserOracleTokOut.Logs)
	log.Printf("AllowUsers(token userOracle) Output Delegated Call: %+v", allowUserOracleTokOut.DelegatedCall)

	log.Printf("AllowUsers(token userOracle) Mode: %s", apTok2.Mode)
	log.Printf("AllowUsers(token userOracle) Users: %+v", apTok2.Users)

	// --------------------------------------------------------------------
	// Attest eligibility (oracle) (envelope + validate + log)
	// --------------------------------------------------------------------
	c.SetPrivateKey(oraclePriv)

	attOut, err := c.AttestParticipantEligibility(ad.Address, userOracle.PublicKey, true)
	if err != nil {
		t.Fatalf("AttestParticipantEligibility: %v", err)
	}
	if len(attOut.States) == 0 {
		t.Fatalf("AttestParticipantEligibility returned empty States")
	}
	if attOut.States[0].Object == nil {
		t.Fatalf("AttestParticipantEligibility returned nil state object")
	}

	var att airdropModels.AirdropStateModel
	unmarshalState(t, attOut.States[0].Object, &att)

	if att.Address != ad.Address {
		t.Fatalf("Attest airdrop mismatch: got %q want %q", att.Address, ad.Address)
	}
	if att.Owner != userOracle.PublicKey {
		t.Fatalf("Attest wallet mismatch: got %q want %q", att.Owner, userOracle.PublicKey)
	}
	if !att.EligibleWallets[userOracle.PublicKey] {
		t.Fatalf("Attest expected Eligible=true")
	}

	log.Printf("AttestParticipantEligibility Output States: %+v", attOut.States)
	log.Printf("AttestParticipantEligibility Output Logs: %+v", attOut.Logs)
	log.Printf("AttestParticipantEligibility Output Delegated Call: %+v", attOut.DelegatedCall)

	log.Printf("AttestParticipantEligibility AirdropAddress: %s", att.Address)
	log.Printf("AttestParticipantEligibility WalletAddress: %s", att.Owner)
	log.Printf("AttestParticipantEligibility Eligible: %v", att.EligibleWallets[userOracle.PublicKey])

	// --------------------------------------------------------------------
	// Claim (userOracle) (envelope + validate + log)
	// --------------------------------------------------------------------
	c.SetPrivateKey(userOraclePriv)

	claim2Out, err := c.ClaimAirdrop(ad.Address, tok.TokenType)
	if err != nil {
		t.Fatalf("ClaimAirdrop: %v", err)
	}
	if len(claim2Out.States) == 0 {
		t.Fatalf("ClaimAirdrop(userOracle) returned empty States")
	}
	if claim2Out.States[0].Object == nil {
		t.Fatalf("ClaimAirdrop(userOracle) returned nil state object")
	}

	var adAfterClaim2 airdropModels.AirdropStateModel
	unmarshalState(t, claim2Out.States[0].Object, &adAfterClaim2)

	if adAfterClaim2.Address != ad.Address {
		t.Fatalf("ClaimAirdrop(userOracle) address mismatch: got %q want %q", adAfterClaim2.Address, ad.Address)
	}

	log.Printf("ClaimAirdrop(userOracle) Output States: %+v", claim2Out.States)
	log.Printf("ClaimAirdrop(userOracle) Output Logs: %+v", claim2Out.Logs)
	log.Printf("ClaimAirdrop(userOracle) Output Delegated Call: %+v", claim2Out.DelegatedCall)

	log.Printf("ClaimAirdrop(userOracle) Address: %s", adAfterClaim2.Address)

	// Double-claim deve falhar
	if _, err := c.ClaimAirdrop(ad.Address, tok.TokenType); err == nil {
		t.Fatalf("ClaimAirdrop: expected error on double-claim, got nil")
	}

	// --------------------------------------------------------------------
	// Withdraw remaining funds (owner) (envelope + validate + log)
	// --------------------------------------------------------------------
	time.Sleep(2 * time.Second)

	c.SetPrivateKey(ownerPriv)

	with2Out, err := c.WithdrawAirdropFunds(ad.Address, amt(50, dec), tokenType, "")
	if err != nil {
		t.Fatalf("WithdrawAirdropFunds: %v", err)
	}
	if len(with2Out.States) == 0 {
		t.Fatalf("WithdrawAirdropFunds(2) returned empty States")
	}
	if with2Out.States[0].Object == nil {
		t.Fatalf("WithdrawAirdropFunds(2) returned nil state object")
	}

	var adAfterWithdraw2 airdropModels.AirdropStateModel
	unmarshalState(t, with2Out.States[0].Object, &adAfterWithdraw2)

	if adAfterWithdraw2.Address != ad.Address {
		t.Fatalf("WithdrawAirdropFunds(2) address mismatch: got %q want %q", adAfterWithdraw2.Address, ad.Address)
	}

	log.Printf("WithdrawAirdropFunds(2) Output States: %+v", with2Out.States)
	log.Printf("WithdrawAirdropFunds(2) Output Logs: %+v", with2Out.Logs)
	log.Printf("WithdrawAirdropFunds(2) Output Delegated Call: %+v", with2Out.DelegatedCall)

	log.Printf("WithdrawAirdropFunds(2) Address: %s", adAfterWithdraw2.Address)
}
