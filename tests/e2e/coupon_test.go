package e2e_test

import (
	"crypto/sha256"
	"encoding/hex"
	"log"
	"testing"
	"time"

	"gitlab.com/2finance/2finance-network/blockchain/contract/contractV1/models"
	couponV1Models "gitlab.com/2finance/2finance-network/blockchain/contract/couponV1/models"
	couponV1 "gitlab.com/2finance/2finance-network/blockchain/contract/couponV1"
	couponV1Domain "gitlab.com/2finance/2finance-network/blockchain/contract/couponV1/domain"
	tokenV1Domain "gitlab.com/2finance/2finance-network/blockchain/contract/tokenV1/domain"
)

func TestCouponFlow(t *testing.T) {
	c := setupClient(t)

	owner, ownerPriv := createWallet(t, c)
	c.SetPrivateKey(ownerPriv)

	dec := 6
	stablecoin := true
	tok := createBasicToken(t, c, owner.PublicKey, dec, true, tokenV1Domain.FUNGIBLE, stablecoin)

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

	log.Printf("Token Address: %s", tok.Address)
	log.Printf("Token Symbol: %s", tok.Symbol)
	log.Printf("Token Name: %s", tok.Name)
	log.Printf("Token Decimals: %d", tok.Decimals)
	log.Printf("Token Total Supply: %s", tok.TotalSupply)
	log.Printf("Token Type: %s", tok.TokenType)
	log.Printf("Token Stablecoin: %v", tok.Stablecoin)

	start := time.Now().Add(2 * time.Second)
	exp := time.Now().Add(25 * time.Minute)

	raw := sha256.Sum256([]byte("e2e-passcode"))
	pcHash := hex.EncodeToString(raw[:])

	// --------------------------------------------------------------------
	// Deploy Coupon contract (envelope + unmarshal + validate + log)
	// --------------------------------------------------------------------
	contractState := models.ContractStateModel{}
	deployedContract, err := c.DeployContract1(couponV1.COUPON_CONTRACT_V1)
	if err != nil {
		t.Fatalf("DeployContract: %v", err)
	}
	if len(deployedContract.States) == 0 {
		t.Fatalf("DeployContract returned empty States")
	}
	if deployedContract.States[0].Object == nil {
		t.Fatalf("DeployContract returned nil state object")
	}

	unmarshalState(t, deployedContract.States[0].Object, &contractState)

	if contractState.Address == "" {
		t.Fatalf("DeployContract returned empty contract address")
	}
	address := contractState.Address

	log.Printf("DeployContract(Coupon) Output States: %+v", deployedContract.States)
	log.Printf("DeployContract(Coupon) Output Logs: %+v", deployedContract.Logs)
	log.Printf("DeployContract(Coupon) Output Delegated Call: %+v", deployedContract.DelegatedCall)

	log.Printf("Coupon Contract Address: %s", address)

	// --------------------------------------------------------------------
	// AddCoupon (envelope + unmarshal + validate + log)
	// --------------------------------------------------------------------
	addOut, err := c.AddCoupon(
		address,
		tok.Address,
		couponV1Domain.DISCOUNT_TYPE_PERCENTAGE,
		"1000", // percentage_bps
		"",     // fixed_amount
		"",     // min_order
		start,
		exp,
		false, // paused
		true,  // stackable
		100,   // max_redemptions
		5,     // per_user_limit
		pcHash,
	)
	if err != nil {
		t.Fatalf("AddCoupon: %v", err)
	}
	if len(addOut.States) == 0 {
		t.Fatalf("AddCoupon returned empty States")
	}
	if addOut.States[0].Object == nil {
		t.Fatalf("AddCoupon returned nil state object")
	}

	var cp couponV1Models.CouponStateModel
	unmarshalState(t, addOut.States[0].Object, &cp)

	// Field validation (todos os campos relevantes do state)
	if cp.Address == "" {
		t.Fatalf("AddCoupon Address empty")
	}
	// Owner pode ser o dono do contrato/caller; como você não passa owner no request, valide “não-vazio”
	if cp.Owner == "" {
		t.Fatalf("AddCoupon Owner empty")
	}
	if cp.TokenAddress != tok.Address {
		t.Fatalf("AddCoupon TokenAddress mismatch: got %q want %q", cp.TokenAddress, tok.Address)
	}
	if cp.ProgramType != couponV1Domain.DISCOUNT_TYPE_PERCENTAGE {
		t.Fatalf("AddCoupon ProgramType mismatch: got %q want %q", cp.ProgramType, couponV1Domain.DISCOUNT_TYPE_PERCENTAGE)
	}
	if cp.PercentageBPS != "1000" {
		t.Fatalf("AddCoupon PercentageBPS mismatch: got %q want %q", cp.PercentageBPS, "1000")
	}
	if cp.FixedAmount != "" {
		t.Fatalf("AddCoupon FixedAmount expected empty for percentage, got %q", cp.FixedAmount)
	}
	if cp.MinOrder != "" {
		t.Fatalf("AddCoupon MinOrder expected empty, got %q", cp.MinOrder)
	}
	if cp.StartAt == nil {
		t.Fatalf("AddCoupon StartAt nil")
	}
	if cp.ExpiredAt == nil {
		t.Fatalf("AddCoupon ExpiredAt nil")
	}
	if cp.Paused != false {
		t.Fatalf("AddCoupon Paused mismatch: got %v want %v", cp.Paused, false)
	}
	if cp.Stackable != true {
		t.Fatalf("AddCoupon Stackable mismatch: got %v want %v", cp.Stackable, true)
	}
	if cp.MaxRedemptions != 100 {
		t.Fatalf("AddCoupon MaxRedemptions mismatch: got %d want %d", cp.MaxRedemptions, 100)
	}
	if cp.PerUserLimit != 5 {
		t.Fatalf("AddCoupon PerUserLimit mismatch: got %d want %d", cp.PerUserLimit, 5)
	}
	if cp.PasscodeHash != pcHash {
		t.Fatalf("AddCoupon PasscodeHash mismatch: got %q want %q", cp.PasscodeHash, pcHash)
	}
	// TotalRedemptions começa em 0
	if cp.TotalRedemptions != 0 {
		t.Fatalf("AddCoupon TotalRedemptions mismatch: got %d want %d", cp.TotalRedemptions, 0)
	}
	if cp.Hash == "" {
		t.Fatalf("AddCoupon Hash empty")
	}
	if cp.CreatedAt.IsZero() {
		t.Fatalf("AddCoupon CreatedAt is zero")
	}
	if cp.UpdatedAt.IsZero() {
		t.Fatalf("AddCoupon UpdatedAt is zero")
	}

	log.Printf("AddCoupon Output States: %+v", addOut.States)
	log.Printf("AddCoupon Output Logs: %+v", addOut.Logs)
	log.Printf("AddCoupon Output Delegated Call: %+v", addOut.DelegatedCall)

	log.Printf("AddCoupon Address: %s", cp.Address)
	log.Printf("AddCoupon Owner: %s", cp.Owner)
	log.Printf("AddCoupon TokenAddress: %s", cp.TokenAddress)
	log.Printf("AddCoupon ProgramType: %s", cp.ProgramType)
	log.Printf("AddCoupon PercentageBPS: %s", cp.PercentageBPS)
	log.Printf("AddCoupon FixedAmount: %s", cp.FixedAmount)
	log.Printf("AddCoupon MinOrder: %s", cp.MinOrder)
	log.Printf("AddCoupon StartAt: %s", cp.StartAt.String())
	log.Printf("AddCoupon ExpiredAt: %s", cp.ExpiredAt.String())
	log.Printf("AddCoupon Paused: %v", cp.Paused)
	log.Printf("AddCoupon Stackable: %v", cp.Stackable)
	log.Printf("AddCoupon MaxRedemptions: %d", cp.MaxRedemptions)
	log.Printf("AddCoupon PerUserLimit: %d", cp.PerUserLimit)
	log.Printf("AddCoupon PasscodeHash: %s", cp.PasscodeHash)
	log.Printf("AddCoupon TotalRedemptions: %d", cp.TotalRedemptions)
	log.Printf("AddCoupon Hash: %s", cp.Hash)
	log.Printf("AddCoupon CreatedAt: %s", cp.CreatedAt.String())
	log.Printf("AddCoupon UpdatedAt: %s", cp.UpdatedAt.String())

	// --------------------------------------------------------------------
	// Allow coupon address (token) (envelope + validate + log)
	// --------------------------------------------------------------------
	allowCouponOut, err := c.AllowUsers(tok.Address, map[string]bool{cp.Address: true})
	if err != nil {
		t.Fatalf("AllowUsers (coupon): %v", err)
	}
	if len(allowCouponOut.States) == 0 {
		t.Fatalf("AllowUsers(coupon) returned empty States")
	}
	if allowCouponOut.States[0].Object == nil {
		t.Fatalf("AllowUsers(coupon) returned nil state object")
	}

	var ap tokenV1Domain.AccessPolicy
	unmarshalState(t, allowCouponOut.States[0].Object, &ap)
	if ap.Users == nil || !ap.Users[cp.Address] {
		t.Fatalf("AllowUsers(coupon) missing coupon address in users")
	}

	log.Printf("AllowUsers(coupon) Output States: %+v", allowCouponOut.States)
	log.Printf("AllowUsers(coupon) Output Logs: %+v", allowCouponOut.Logs)
	log.Printf("AllowUsers(coupon) Output Delegated Call: %+v", allowCouponOut.DelegatedCall)

	log.Printf("AllowUsers Mode: %s", ap.Mode)
	log.Printf("AllowUsers Users: %+v", ap.Users)

	// --------------------------------------------------------------------
	// Update coupon to fixed amount (envelope + unmarshal + validate + log)
	// --------------------------------------------------------------------
	start2 := time.Now().Add(1 * time.Second)
	exp2 := time.Now().Add(10 * time.Minute)

	raw2 := sha256.Sum256([]byte("e2e-passcode-2"))
	pcHash2 := hex.EncodeToString(raw2[:])

	updOut, err := c.UpdateCoupon(
		cp.Address,
		tok.Address,
		couponV1Domain.DISCOUNT_TYPE_FIXED,
		"",           // percentage_bps
		amt(2, dec),  // fixed_amount
		amt(10, dec), // min_order
		start2,
		exp2,
		false, // paused
		10,    // max_redemptions
		2,     // per_user_limit
		pcHash2,
	)
	if err != nil {
		t.Fatalf("UpdateCoupon: %v", err)
	}
	if len(updOut.States) == 0 {
		t.Fatalf("UpdateCoupon returned empty States")
	}
	if updOut.States[0].Object == nil {
		t.Fatalf("UpdateCoupon returned nil state object")
	}

	var cpUpdated couponV1Models.CouponStateModel
	unmarshalState(t, updOut.States[0].Object, &cpUpdated)

	if cpUpdated.Address != cp.Address {
		t.Fatalf("UpdateCoupon Address mismatch: got %q want %q", cpUpdated.Address, cp.Address)
	}
	if cpUpdated.TokenAddress != tok.Address {
		t.Fatalf("UpdateCoupon TokenAddress mismatch: got %q want %q", cpUpdated.TokenAddress, tok.Address)
	}
	if cpUpdated.ProgramType != couponV1Domain.DISCOUNT_TYPE_FIXED {
		t.Fatalf("UpdateCoupon ProgramType mismatch: got %q want %q", cpUpdated.ProgramType, couponV1Domain.DISCOUNT_TYPE_FIXED)
	}
	if cpUpdated.PercentageBPS != "" {
		t.Fatalf("UpdateCoupon PercentageBPS expected empty for fixed, got %q", cpUpdated.PercentageBPS)
	}
	if cpUpdated.FixedAmount != amt(2, dec) {
		t.Fatalf("UpdateCoupon FixedAmount mismatch: got %q want %q", cpUpdated.FixedAmount, amt(2, dec))
	}
	if cpUpdated.MinOrder != amt(10, dec) {
		t.Fatalf("UpdateCoupon MinOrder mismatch: got %q want %q", cpUpdated.MinOrder, amt(10, dec))
	}
	if cpUpdated.StartAt == nil || cpUpdated.ExpiredAt == nil {
		t.Fatalf("UpdateCoupon StartAt/ExpiredAt nil: start=%v exp=%v", cpUpdated.StartAt, cpUpdated.ExpiredAt)
	}
	if cpUpdated.Paused != false {
		t.Fatalf("UpdateCoupon Paused mismatch: got %v want %v", cpUpdated.Paused, false)
	}
	if cpUpdated.MaxRedemptions != 10 {
		t.Fatalf("UpdateCoupon MaxRedemptions mismatch: got %d want %d", cpUpdated.MaxRedemptions, 10)
	}
	if cpUpdated.PerUserLimit != 2 {
		t.Fatalf("UpdateCoupon PerUserLimit mismatch: got %d want %d", cpUpdated.PerUserLimit, 2)
	}
	if cpUpdated.PasscodeHash != pcHash2 {
		t.Fatalf("UpdateCoupon PasscodeHash mismatch: got %q want %q", cpUpdated.PasscodeHash, pcHash2)
	}
	if cpUpdated.Hash == "" {
		t.Fatalf("UpdateCoupon Hash empty")
	}

	log.Printf("UpdateCoupon Output States: %+v", updOut.States)
	log.Printf("UpdateCoupon Output Logs: %+v", updOut.Logs)
	log.Printf("UpdateCoupon Output Delegated Call: %+v", updOut.DelegatedCall)

	log.Printf("UpdateCoupon Address: %s", cpUpdated.Address)
	log.Printf("UpdateCoupon Owner: %s", cpUpdated.Owner)
	log.Printf("UpdateCoupon TokenAddress: %s", cpUpdated.TokenAddress)
	log.Printf("UpdateCoupon ProgramType: %s", cpUpdated.ProgramType)
	log.Printf("UpdateCoupon PercentageBPS: %s", cpUpdated.PercentageBPS)
	log.Printf("UpdateCoupon FixedAmount: %s", cpUpdated.FixedAmount)
	log.Printf("UpdateCoupon MinOrder: %s", cpUpdated.MinOrder)
	log.Printf("UpdateCoupon StartAt: %s", cpUpdated.StartAt.String())
	log.Printf("UpdateCoupon ExpiredAt: %s", cpUpdated.ExpiredAt.String())
	log.Printf("UpdateCoupon Paused: %v", cpUpdated.Paused)
	log.Printf("UpdateCoupon Stackable: %v", cpUpdated.Stackable)
	log.Printf("UpdateCoupon MaxRedemptions: %d", cpUpdated.MaxRedemptions)
	log.Printf("UpdateCoupon PerUserLimit: %d", cpUpdated.PerUserLimit)
	log.Printf("UpdateCoupon PasscodeHash: %s", cpUpdated.PasscodeHash)
	log.Printf("UpdateCoupon TotalRedemptions: %d", cpUpdated.TotalRedemptions)
	log.Printf("UpdateCoupon Hash: %s", cpUpdated.Hash)
	log.Printf("UpdateCoupon CreatedAt: %s", cpUpdated.CreatedAt.String())
	log.Printf("UpdateCoupon UpdatedAt: %s", cpUpdated.UpdatedAt.String())

	// --------------------------------------------------------------------
	// wait and redeem (user) (envelope + validate + log)
	// --------------------------------------------------------------------
	waitUntil(t, 20*time.Second, func() bool { return time.Now().After(start2) })

	user, userPriv := createWallet(t, c)

	// allow user (token) - owner signs
	c.SetPrivateKey(ownerPriv)
	allowUserOut, err := c.AllowUsers(tok.Address, map[string]bool{user.PublicKey: true})
	if err != nil {
		t.Fatalf("AllowUsers(user): %v", err)
	}
	if len(allowUserOut.States) == 0 || allowUserOut.States[0].Object == nil {
		t.Fatalf("AllowUsers(user) returned empty/nil state")
	}

	c.SetPrivateKey(userPriv)
	redeemOut, err := c.RedeemCoupon(cp.Address, amt(20, dec), "e2e-passcode-2", tokenV1Domain.FUNGIBLE, "")
	if err != nil {
		t.Fatalf("RedeemCoupon: %v", err)
	}
	if len(redeemOut.States) == 0 {
		t.Fatalf("RedeemCoupon returned empty States")
	}
	if redeemOut.States[0].Object == nil {
		t.Fatalf("RedeemCoupon returned nil state object")
	}

	// Preferência: Redeem retornar CouponStateModel atualizado
	var cpAfterRedeem couponV1Models.CouponStateModel
	unmarshalState(t, redeemOut.States[0].Object, &cpAfterRedeem)

	if cpAfterRedeem.Address != cp.Address {
		t.Fatalf("RedeemCoupon Address mismatch: got %q want %q", cpAfterRedeem.Address, cp.Address)
	}

	log.Printf("RedeemCoupon Output States: %+v", redeemOut.States)
	log.Printf("RedeemCoupon Output Logs: %+v", redeemOut.Logs)
	log.Printf("RedeemCoupon Output Delegated Call: %+v", redeemOut.DelegatedCall)

	log.Printf("RedeemCoupon Address: %s", cpAfterRedeem.Address)
	log.Printf("RedeemCoupon TotalRedemptions: %d", cpAfterRedeem.TotalRedemptions)
	log.Printf("RedeemCoupon UpdatedAt: %s", cpAfterRedeem.UpdatedAt.String())

	// --------------------------------------------------------------------
	// pause/unpause (owner) (envelope + unmarshal + validate + log)
	// --------------------------------------------------------------------
	c.SetPrivateKey(ownerPriv)

	pauseOut, err := c.PauseCoupon(cp.Address, true)
	if err != nil {
		t.Fatalf("PauseCoupon: %v", err)
	}
	if len(pauseOut.States) == 0 || pauseOut.States[0].Object == nil {
		t.Fatalf("PauseCoupon returned empty/nil state")
	}

	var cpPaused couponV1Models.CouponStateModel
	unmarshalState(t, pauseOut.States[0].Object, &cpPaused)

	if cpPaused.Address != cp.Address {
		t.Fatalf("PauseCoupon Address mismatch: got %q want %q", cpPaused.Address, cp.Address)
	}
	if !cpPaused.Paused {
		t.Fatalf("PauseCoupon expected Paused=true")
	}

	log.Printf("PauseCoupon Output States: %+v", pauseOut.States)
	log.Printf("PauseCoupon Output Logs: %+v", pauseOut.Logs)
	log.Printf("PauseCoupon Output Delegated Call: %+v", pauseOut.DelegatedCall)

	log.Printf("PauseCoupon Address: %s", cpPaused.Address)
	log.Printf("PauseCoupon Paused: %v", cpPaused.Paused)

	unpauseOut, err := c.UnpauseCoupon(cp.Address, false)
	if err != nil {
		t.Fatalf("UnpauseCoupon: %v", err)
	}
	if len(unpauseOut.States) == 0 || unpauseOut.States[0].Object == nil {
		t.Fatalf("UnpauseCoupon returned empty/nil state")
	}

	var cpUnpaused couponV1Models.CouponStateModel
	unmarshalState(t, unpauseOut.States[0].Object, &cpUnpaused)

	if cpUnpaused.Address != cp.Address {
		t.Fatalf("UnpauseCoupon Address mismatch: got %q want %q", cpUnpaused.Address, cp.Address)
	}
	if cpUnpaused.Paused {
		t.Fatalf("UnpauseCoupon expected Paused=false")
	}

	log.Printf("UnpauseCoupon Output States: %+v", unpauseOut.States)
	log.Printf("UnpauseCoupon Output Logs: %+v", unpauseOut.Logs)
	log.Printf("UnpauseCoupon Output Delegated Call: %+v", unpauseOut.DelegatedCall)

	log.Printf("UnpauseCoupon Address: %s", cpUnpaused.Address)
	log.Printf("UnpauseCoupon Paused: %v", cpUnpaused.Paused)

	// --------------------------------------------------------------------
	// Getters (envelope + unmarshal + validate + log)
	// --------------------------------------------------------------------
	getOut, err := c.GetCoupon(cp.Address)
	if err != nil {
		t.Fatalf("GetCoupon: %v", err)
	}
	if len(getOut.States) == 0 || getOut.States[0].Object == nil {
		t.Fatalf("GetCoupon returned empty/nil state")
	}

	var cpGet couponV1Models.CouponStateModel
	unmarshalState(t, getOut.States[0].Object, &cpGet)

	if cpGet.Address != cp.Address {
		t.Fatalf("GetCoupon Address mismatch: got %q want %q", cpGet.Address, cp.Address)
	}

	log.Printf("GetCoupon Output States: %+v", getOut.States)
	log.Printf("GetCoupon Output Logs: %+v", getOut.Logs)
	log.Printf("GetCoupon Output Delegated Call: %+v", getOut.DelegatedCall)

	log.Printf("GetCoupon Address: %s", cpGet.Address)
	log.Printf("GetCoupon Owner: %s", cpGet.Owner)
	log.Printf("GetCoupon TokenAddress: %s", cpGet.TokenAddress)
	log.Printf("GetCoupon ProgramType: %s", cpGet.ProgramType)
	log.Printf("GetCoupon PercentageBPS: %s", cpGet.PercentageBPS)
	log.Printf("GetCoupon FixedAmount: %s", cpGet.FixedAmount)
	log.Printf("GetCoupon MinOrder: %s", cpGet.MinOrder)
	log.Printf("GetCoupon StartAt: %s", cpGet.StartAt.String())
	log.Printf("GetCoupon ExpiredAt: %s", cpGet.ExpiredAt.String())
	log.Printf("GetCoupon Paused: %v", cpGet.Paused)
	log.Printf("GetCoupon Stackable: %v", cpGet.Stackable)
	log.Printf("GetCoupon MaxRedemptions: %d", cpGet.MaxRedemptions)
	log.Printf("GetCoupon PerUserLimit: %d", cpGet.PerUserLimit)
	log.Printf("GetCoupon PasscodeHash: %s", cpGet.PasscodeHash)
	log.Printf("GetCoupon TotalRedemptions: %d", cpGet.TotalRedemptions)
	log.Printf("GetCoupon Hash: %s", cpGet.Hash)
	log.Printf("GetCoupon CreatedAt: %s", cpGet.CreatedAt.String())
	log.Printf("GetCoupon UpdatedAt: %s", cpGet.UpdatedAt.String())

	listOut, err := c.ListCoupons("", tok.Address, "", nil, 1, 10, true)
	if err != nil {
		t.Fatalf("ListCoupons: %v", err)
	}
	if len(listOut.States) == 0 || listOut.States[0].Object == nil {
		t.Fatalf("ListCoupons returned empty/nil state")
	}

	var list []couponV1Models.CouponStateModel
	unmarshalState(t, listOut.States[0].Object, &list)

	found := false
	for _, it := range list {
		if it.Address == cp.Address {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("ListCoupons: coupon %s not found", cp.Address)
	}

	log.Printf("ListCoupons Output States: %+v", listOut.States)
	log.Printf("ListCoupons Output Logs: %+v", listOut.Logs)
	log.Printf("ListCoupons Output Delegated Call: %+v", listOut.DelegatedCall)

	log.Printf("ListCoupons Count: %d", len(list))
}

func TestCouponFlow_NonFungible(t *testing.T) {
	c := setupClient(t)

	owner, ownerPriv := createWallet(t, c)
	c.SetPrivateKey(ownerPriv)

	stablecoin := false
	dec := 0

	tok := createBasicToken(t, c, owner.PublicKey, dec, false, tokenV1Domain.NON_FUNGIBLE, stablecoin)

	// Token (mínimo) validate + log
	if tok.Address == "" {
		t.Fatalf("Token address empty")
	}
	if tok.TokenType != tokenV1Domain.NON_FUNGIBLE {
		t.Fatalf("Token type mismatch: got %s want %s", tok.TokenType, tokenV1Domain.NON_FUNGIBLE)
	}

	log.Printf("Token Address: %s", tok.Address)
	log.Printf("Token Type: %s", tok.TokenType)

	// --------------------------------------------------------------------
	// Mint NFT (envelope + validate + log)
	// --------------------------------------------------------------------
	mintOut, err := c.MintToken(tok.Address, owner.PublicKey, "1", dec, tok.TokenType)
	if err != nil {
		t.Fatalf("MintToken NFT: %v", err)
	}
	if len(mintOut.States) == 0 {
		t.Fatalf("MintToken NFT returned empty States")
	}
	if mintOut.States[0].Object == nil {
		t.Fatalf("MintToken NFT returned nil state object")
	}

	var mint tokenV1Domain.Mint
	unmarshalState(t, mintOut.States[0].Object, &mint)

	if mint.TokenAddress != tok.Address {
		t.Fatalf("Mint NFT TokenAddress mismatch: got %q want %q", mint.TokenAddress, tok.Address)
	}
	if mint.MintTo != owner.PublicKey {
		t.Fatalf("Mint NFT MintTo mismatch: got %q want %q", mint.MintTo, owner.PublicKey)
	}
	if mint.TokenType != tok.TokenType {
		t.Fatalf("Mint NFT TokenType mismatch: got %q want %q", mint.TokenType, tok.TokenType)
	}
	if len(mint.TokenUUIDList) != 1 {
		t.Fatalf("expected 1 NFT uuid, got %d", len(mint.TokenUUIDList))
	}
	nftUUID := mint.TokenUUIDList[0]
	if nftUUID == "" {
		t.Fatalf("minted nft uuid empty")
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
	// Deploy Coupon Contract (envelope + unmarshal + validate + log)
	// --------------------------------------------------------------------
	start := time.Now().Add(2 * time.Second)
	exp := time.Now().Add(25 * time.Minute)

	raw := sha256.Sum256([]byte("e2e-passcode-nft"))
	pcHash := hex.EncodeToString(raw[:])

	var contractState models.ContractStateModel
	deployedContract, err := c.DeployContract1(couponV1.COUPON_CONTRACT_V1)
	if err != nil {
		t.Fatalf("DeployContract: %v", err)
	}
	if len(deployedContract.States) == 0 || deployedContract.States[0].Object == nil {
		t.Fatalf("DeployContract returned empty/nil state")
	}
	unmarshalState(t, deployedContract.States[0].Object, &contractState)
	if contractState.Address == "" {
		t.Fatalf("DeployContract returned empty contract address")
	}

	log.Printf("DeployContract(Coupon) Output States: %+v", deployedContract.States)
	log.Printf("DeployContract(Coupon) Output Logs: %+v", deployedContract.Logs)
	log.Printf("DeployContract(Coupon) Output Delegated Call: %+v", deployedContract.DelegatedCall)

	log.Printf("Coupon Contract Address: %s", contractState.Address)

	// --------------------------------------------------------------------
	// Add Coupon (owner) (envelope + unmarshal + validate + log)
	// --------------------------------------------------------------------
	addOut, err := c.AddCoupon(
		contractState.Address,
		tok.Address,
		couponV1Domain.DISCOUNT_TYPE_FIXED,
		"",      // percentage_bps
		"10000", // fixed_amount
		"",      // min_order
		start,
		exp,
		false, // paused
		true,  // stackable
		1,     // max_redemptions
		1,     // per_user_limit
		pcHash,
	)
	if err != nil {
		t.Fatalf("AddCoupon NFT: %v", err)
	}
	if len(addOut.States) == 0 || addOut.States[0].Object == nil {
		t.Fatalf("AddCoupon NFT returned empty/nil state")
	}

	var cp couponV1Models.CouponStateModel
	unmarshalState(t, addOut.States[0].Object, &cp)

	if cp.Address == "" {
		t.Fatalf("AddCoupon NFT Address empty")
	}
	if cp.Owner == "" {
		t.Fatalf("AddCoupon NFT Owner empty")
	}
	if cp.TokenAddress != tok.Address {
		t.Fatalf("AddCoupon NFT TokenAddress mismatch: got %q want %q", cp.TokenAddress, tok.Address)
	}
	if cp.ProgramType != couponV1Domain.DISCOUNT_TYPE_FIXED {
		t.Fatalf("AddCoupon NFT ProgramType mismatch: got %q want %q", cp.ProgramType, couponV1Domain.DISCOUNT_TYPE_FIXED)
	}
	if cp.FixedAmount != "10000" {
		t.Fatalf("AddCoupon NFT FixedAmount mismatch: got %q want %q", cp.FixedAmount, "10000")
	}
	if cp.StartAt == nil || cp.ExpiredAt == nil {
		t.Fatalf("AddCoupon NFT StartAt/ExpiredAt nil: start=%v exp=%v", cp.StartAt, cp.ExpiredAt)
	}
	if cp.Stackable != true {
		t.Fatalf("AddCoupon NFT Stackable mismatch: got %v want %v", cp.Stackable, true)
	}
	if cp.MaxRedemptions != 1 {
		t.Fatalf("AddCoupon NFT MaxRedemptions mismatch: got %d want %d", cp.MaxRedemptions, 1)
	}
	if cp.PerUserLimit != 1 {
		t.Fatalf("AddCoupon NFT PerUserLimit mismatch: got %d want %d", cp.PerUserLimit, 1)
	}
	if cp.PasscodeHash != pcHash {
		t.Fatalf("AddCoupon NFT PasscodeHash mismatch: got %q want %q", cp.PasscodeHash, pcHash)
	}
	if cp.Hash == "" {
		t.Fatalf("AddCoupon NFT Hash empty")
	}

	log.Printf("AddCoupon NFT Output States: %+v", addOut.States)
	log.Printf("AddCoupon NFT Output Logs: %+v", addOut.Logs)
	log.Printf("AddCoupon NFT Output Delegated Call: %+v", addOut.DelegatedCall)

	log.Printf("AddCoupon Address: %s", cp.Address)
	log.Printf("AddCoupon Owner: %s", cp.Owner)
	log.Printf("AddCoupon TokenAddress: %s", cp.TokenAddress)
	log.Printf("AddCoupon ProgramType: %s", cp.ProgramType)
	log.Printf("AddCoupon FixedAmount: %s", cp.FixedAmount)
	log.Printf("AddCoupon StartAt: %s", cp.StartAt.String())
	log.Printf("AddCoupon ExpiredAt: %s", cp.ExpiredAt.String())
	log.Printf("AddCoupon Paused: %v", cp.Paused)
	log.Printf("AddCoupon Stackable: %v", cp.Stackable)
	log.Printf("AddCoupon MaxRedemptions: %d", cp.MaxRedemptions)
	log.Printf("AddCoupon PerUserLimit: %d", cp.PerUserLimit)
	log.Printf("AddCoupon PasscodeHash: %s", cp.PasscodeHash)
	log.Printf("AddCoupon TotalRedemptions: %d", cp.TotalRedemptions)
	log.Printf("AddCoupon Hash: %s", cp.Hash)

	// --------------------------------------------------------------------
	// Allow coupon address (token) (envelope + validate + log)
	// --------------------------------------------------------------------
	allowCouponOut, err := c.AllowUsers(tok.Address, map[string]bool{cp.Address: true})
	if err != nil {
		t.Fatalf("AllowUsers (coupon): %v", err)
	}
	if len(allowCouponOut.States) == 0 || allowCouponOut.States[0].Object == nil {
		t.Fatalf("AllowUsers(coupon) returned empty/nil state")
	}

	var ap tokenV1Domain.AccessPolicy
	unmarshalState(t, allowCouponOut.States[0].Object, &ap)
	if ap.Users == nil || !ap.Users[cp.Address] {
		t.Fatalf("AllowUsers(coupon) missing coupon address in users")
	}

	log.Printf("AllowUsers(coupon) Output States: %+v", allowCouponOut.States)
	log.Printf("AllowUsers(coupon) Output Logs: %+v", allowCouponOut.Logs)
	log.Printf("AllowUsers(coupon) Output Delegated Call: %+v", allowCouponOut.DelegatedCall)

	log.Printf("AllowUsers Mode: %s", ap.Mode)
	log.Printf("AllowUsers Users: %+v", ap.Users)

	// --------------------------------------------------------------------
	// Update Coupon (owner) (envelope + unmarshal + validate + log)
	// --------------------------------------------------------------------
	start2 := time.Now().Add(1 * time.Second)
	exp2 := time.Now().Add(10 * time.Minute)

	raw2 := sha256.Sum256([]byte("e2e-passcode-nft-2"))
	pcHash2 := hex.EncodeToString(raw2[:])

	updOut, err := c.UpdateCoupon(
		cp.Address,
		tok.Address,
		couponV1Domain.DISCOUNT_TYPE_FIXED,
		"",
		"10000",
		"",
		start2,
		exp2,
		false,
		1,
		1,
		pcHash2,
	)
	if err != nil {
		t.Fatalf("UpdateCoupon NFT: %v", err)
	}
	if len(updOut.States) == 0 || updOut.States[0].Object == nil {
		t.Fatalf("UpdateCoupon NFT returned empty/nil state")
	}

	var cpUpdated couponV1Models.CouponStateModel
	unmarshalState(t, updOut.States[0].Object, &cpUpdated)

	if cpUpdated.Address != cp.Address {
		t.Fatalf("UpdateCoupon NFT Address mismatch: got %q want %q", cpUpdated.Address, cp.Address)
	}
	if cpUpdated.TokenAddress != tok.Address {
		t.Fatalf("UpdateCoupon NFT TokenAddress mismatch: got %q want %q", cpUpdated.TokenAddress, tok.Address)
	}
	if cpUpdated.ProgramType != couponV1Domain.DISCOUNT_TYPE_FIXED {
		t.Fatalf("UpdateCoupon NFT ProgramType mismatch: got %q want %q", cpUpdated.ProgramType, couponV1Domain.DISCOUNT_TYPE_FIXED)
	}
	if cpUpdated.FixedAmount != "10000" {
		t.Fatalf("UpdateCoupon NFT FixedAmount mismatch: got %q want %q", cpUpdated.FixedAmount, "10000")
	}
	if cpUpdated.PasscodeHash != pcHash2 {
		t.Fatalf("UpdateCoupon NFT PasscodeHash mismatch: got %q want %q", cpUpdated.PasscodeHash, pcHash2)
	}

	log.Printf("UpdateCoupon NFT Output States: %+v", updOut.States)
	log.Printf("UpdateCoupon NFT Output Logs: %+v", updOut.Logs)
	log.Printf("UpdateCoupon NFT Output Delegated Call: %+v", updOut.DelegatedCall)

	log.Printf("UpdateCoupon Address: %s", cpUpdated.Address)
	log.Printf("UpdateCoupon ProgramType: %s", cpUpdated.ProgramType)
	log.Printf("UpdateCoupon FixedAmount: %s", cpUpdated.FixedAmount)
	log.Printf("UpdateCoupon PasscodeHash: %s", cpUpdated.PasscodeHash)

	// --------------------------------------------------------------------
	// Wait start2
	// --------------------------------------------------------------------
	waitUntil(t, 20*time.Second, func() bool { return time.Now().After(start2) })

	// --------------------------------------------------------------------
	// Create User + allow user + transfer NFT owner -> user
	// --------------------------------------------------------------------
	user, userPriv := createWallet(t, c)

	c.SetPrivateKey(ownerPriv)
	if _, err := c.AllowUsers(tok.Address, map[string]bool{user.PublicKey: true}); err != nil {
		t.Fatalf("AllowUsers (user): %v", err)
	}

	trOut, err := c.TransferToken(tok.Address, user.PublicKey, "1", dec, tokenV1Domain.NON_FUNGIBLE, nftUUID)
	if err != nil {
		t.Fatalf("Transfer NFT to user: %v", err)
	}
	if len(trOut.States) == 0 || trOut.States[0].Object == nil {
		t.Fatalf("Transfer NFT returned empty/nil state")
	}

	var tr tokenV1Domain.Transfer
	unmarshalState(t, trOut.States[0].Object, &tr)

	if tr.ToAddress != user.PublicKey {
		t.Fatalf("Transfer NFT to mismatch: got %q want %q", tr.ToAddress, user.PublicKey)
	}
	if tr.UUID != nftUUID {
		t.Fatalf("Transfer NFT uuid mismatch: got %q want %q", tr.UUID, nftUUID)
	}

	log.Printf("Transfer NFT Output States: %+v", trOut.States)
	log.Printf("Transfer NFT Output Logs: %+v", trOut.Logs)
	log.Printf("Transfer NFT Output Delegated Call: %+v", trOut.DelegatedCall)

	log.Printf("Transfer TokenAddress: %s", tr.TokenAddress)
	log.Printf("Transfer FromAddress: %s", tr.FromAddress)
	log.Printf("Transfer ToAddress: %s", tr.ToAddress)
	log.Printf("Transfer Amount: %s", tr.Amount)
	log.Printf("Transfer TokenType: %s", tr.TokenType)
	log.Printf("Transfer UUID: %s", tr.UUID)

	// --------------------------------------------------------------------
	// Redeem Coupon (user) (envelope + validate + log)
	// --------------------------------------------------------------------
	c.SetPrivateKey(userPriv)

	redeemOut, err := c.RedeemCoupon(cp.Address, "1", "e2e-passcode-nft-2", tokenV1Domain.NON_FUNGIBLE, nftUUID)
	if err != nil {
		t.Fatalf("RedeemCoupon NFT: %v", err)
	}
	if len(redeemOut.States) == 0 || redeemOut.States[0].Object == nil {
		t.Fatalf("RedeemCoupon NFT returned empty/nil state")
	}

	var cpAfterRedeem couponV1Models.CouponStateModel
	unmarshalState(t, redeemOut.States[0].Object, &cpAfterRedeem)

	if cpAfterRedeem.Address != cp.Address {
		t.Fatalf("RedeemCoupon NFT Address mismatch: got %q want %q", cpAfterRedeem.Address, cp.Address)
	}

	log.Printf("RedeemCoupon NFT Output States: %+v", redeemOut.States)
	log.Printf("RedeemCoupon NFT Output Logs: %+v", redeemOut.Logs)
	log.Printf("RedeemCoupon NFT Output Delegated Call: %+v", redeemOut.DelegatedCall)

	log.Printf("RedeemCoupon Address: %s", cpAfterRedeem.Address)
	log.Printf("RedeemCoupon TotalRedemptions: %d", cpAfterRedeem.TotalRedemptions)
	log.Printf("RedeemCoupon UpdatedAt: %s", cpAfterRedeem.UpdatedAt.String())

	// --------------------------------------------------------------------
	// Pause / Unpause (owner) (envelope + unmarshal + validate + log)
	// --------------------------------------------------------------------
	c.SetPrivateKey(ownerPriv)

	pauseOut, err := c.PauseCoupon(cp.Address, true)
	if err != nil {
		t.Fatalf("PauseCoupon: %v", err)
	}
	if len(pauseOut.States) == 0 || pauseOut.States[0].Object == nil {
		t.Fatalf("PauseCoupon returned empty/nil state")
	}

	var cpPaused couponV1Models.CouponStateModel
	unmarshalState(t, pauseOut.States[0].Object, &cpPaused)

	if cpPaused.Address != cp.Address {
		t.Fatalf("PauseCoupon Address mismatch: got %q want %q", cpPaused.Address, cp.Address)
	}
	if !cpPaused.Paused {
		t.Fatalf("PauseCoupon expected Paused=true")
	}

	log.Printf("PauseCoupon Output States: %+v", pauseOut.States)
	log.Printf("PauseCoupon Output Logs: %+v", pauseOut.Logs)
	log.Printf("PauseCoupon Output Delegated Call: %+v", pauseOut.DelegatedCall)

	log.Printf("PauseCoupon Address: %s", cpPaused.Address)
	log.Printf("PauseCoupon Paused: %v", cpPaused.Paused)

	unpauseOut, err := c.UnpauseCoupon(cp.Address, false)
	if err != nil {
		t.Fatalf("UnpauseCoupon: %v", err)
	}
	if len(unpauseOut.States) == 0 || unpauseOut.States[0].Object == nil {
		t.Fatalf("UnpauseCoupon returned empty/nil state")
	}

	var cpUnpaused couponV1Models.CouponStateModel
	unmarshalState(t, unpauseOut.States[0].Object, &cpUnpaused)

	if cpUnpaused.Address != cp.Address {
		t.Fatalf("UnpauseCoupon Address mismatch: got %q want %q", cpUnpaused.Address, cp.Address)
	}
	if cpUnpaused.Paused {
		t.Fatalf("UnpauseCoupon expected Paused=false")
	}

	log.Printf("UnpauseCoupon Output States: %+v", unpauseOut.States)
	log.Printf("UnpauseCoupon Output Logs: %+v", unpauseOut.Logs)
	log.Printf("UnpauseCoupon Output Delegated Call: %+v", unpauseOut.DelegatedCall)

	log.Printf("UnpauseCoupon Address: %s", cpUnpaused.Address)
	log.Printf("UnpauseCoupon Paused: %v", cpUnpaused.Paused)

	// --------------------------------------------------------------------
	// Getters (envelope + unmarshal + validate + log)
	// --------------------------------------------------------------------
	getOut, err := c.GetCoupon(cp.Address)
	if err != nil {
		t.Fatalf("GetCoupon: %v", err)
	}
	if len(getOut.States) == 0 || getOut.States[0].Object == nil {
		t.Fatalf("GetCoupon returned empty/nil state")
	}

	var cpGet couponV1Models.CouponStateModel
	unmarshalState(t, getOut.States[0].Object, &cpGet)

	if cpGet.Address != cp.Address {
		t.Fatalf("GetCoupon Address mismatch: got %q want %q", cpGet.Address, cp.Address)
	}

	log.Printf("GetCoupon Output States: %+v", getOut.States)
	log.Printf("GetCoupon Output Logs: %+v", getOut.Logs)
	log.Printf("GetCoupon Output Delegated Call: %+v", getOut.DelegatedCall)

	log.Printf("GetCoupon Address: %s", cpGet.Address)
	log.Printf("GetCoupon Owner: %s", cpGet.Owner)
	log.Printf("GetCoupon TokenAddress: %s", cpGet.TokenAddress)
	log.Printf("GetCoupon ProgramType: %s", cpGet.ProgramType)
	log.Printf("GetCoupon PercentageBPS: %s", cpGet.PercentageBPS)
	log.Printf("GetCoupon FixedAmount: %s", cpGet.FixedAmount)
	log.Printf("GetCoupon MinOrder: %s", cpGet.MinOrder)
	log.Printf("GetCoupon StartAt: %s", cpGet.StartAt.String())
	log.Printf("GetCoupon ExpiredAt: %s", cpGet.ExpiredAt.String())
	log.Printf("GetCoupon Paused: %v", cpGet.Paused)
	log.Printf("GetCoupon Stackable: %v", cpGet.Stackable)
	log.Printf("GetCoupon MaxRedemptions: %d", cpGet.MaxRedemptions)
	log.Printf("GetCoupon PerUserLimit: %d", cpGet.PerUserLimit)
	log.Printf("GetCoupon PasscodeHash: %s", cpGet.PasscodeHash)
	log.Printf("GetCoupon TotalRedemptions: %d", cpGet.TotalRedemptions)
	log.Printf("GetCoupon Hash: %s", cpGet.Hash)
	log.Printf("GetCoupon CreatedAt: %s", cpGet.CreatedAt.String())
	log.Printf("GetCoupon UpdatedAt: %s", cpGet.UpdatedAt.String())

	listOut, err := c.ListCoupons("", tok.Address, "", nil, 1, 10, true)
	if err != nil {
		t.Fatalf("ListCoupons: %v", err)
	}
	if len(listOut.States) == 0 || listOut.States[0].Object == nil {
		t.Fatalf("ListCoupons returned empty/nil state")
	}

	var list []couponV1Models.CouponStateModel
	unmarshalState(t, listOut.States[0].Object, &list)

	found := false
	for _, it := range list {
		if it.Address == cp.Address {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("ListCoupons: coupon %s not found", cp.Address)
	}

	log.Printf("ListCoupons Output States: %+v", listOut.States)
	log.Printf("ListCoupons Output Logs: %+v", listOut.Logs)
	log.Printf("ListCoupons Output Delegated Call: %+v", listOut.DelegatedCall)

	log.Printf("ListCoupons Count: %d", len(list))
}
