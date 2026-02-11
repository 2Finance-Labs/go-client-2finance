package e2e_test

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"
	"time"

	"gitlab.com/2finance/2finance-network/blockchain/contract/contractV1/models"
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

	start := time.Now().Add(2 * time.Second)
	exp := time.Now().Add(25 * time.Minute)
	raw := sha256.Sum256([]byte("e2e-passcode"))
	pcHash := hex.EncodeToString(raw[:])

	contractState := models.ContractStateModel{}
	deployedContract, err := c.DeployContract1(couponV1.COUPON_CONTRACT_V1)
	if err != nil {
		t.Fatalf("DeployContract: %v", err)
	}
	unmarshalState(t, deployedContract.States[0].Object, &contractState)
	address := contractState.Address

	out, err := c.AddCoupon(address, tok.Address, couponV1Domain.DISCOUNT_TYPE_PERCENTAGE, "1000", "", "", start, exp, false, true, 100, 5, pcHash)
	if err != nil {
		t.Fatalf("AddCoupon: %v", err)
	}
	var cp couponV1Domain.Coupon
	unmarshalState(t, out.States[0].Object, &cp)
	if cp.Address == "" {
		t.Fatalf("coupon addr empty")
	}

	// allow coupon to spend if required
	_, _ = c.AllowUsers(tok.Address, map[string]bool{cp.Address: true})

	// Update coupon to fixed amount
	start2 := time.Now().Add(1 * time.Second)
	exp2 := time.Now().Add(10 * time.Minute)
	raw2 := sha256.Sum256([]byte("e2e-passcode-2"))
	pcHash2 := hex.EncodeToString(raw2[:])
	if _, err := c.UpdateCoupon(cp.Address, tok.Address, couponV1Domain.DISCOUNT_TYPE_FIXED, "", amt(2, dec), amt(10, dec), start2, exp2, false, 10, 2, pcHash2); err != nil {
		t.Fatalf("UpdateCoupon: %v", err)
	}

	// wait and redeem
	waitUntil(t, 20*time.Second, func() bool { return time.Now().After(start2) })
	if _, err := c.RedeemCoupon(cp.Address, amt(20, dec), "e2e-passcode-2", tokenV1Domain.FUNGIBLE, ""); err != nil {
		t.Fatalf("RedeemCoupon warning: %v", err) // allow policy-gate differences
	}

	// pause/unpause & getters
	if _, err := c.PauseCoupon(cp.Address, true); err != nil {
		t.Fatalf("PauseCoupon: %v", err)
	}
	if _, err = c.UnpauseCoupon(cp.Address, false); err != nil {
		t.Fatalf("UnpauseCoupon: %v", err)
	}
	// getters
	if _, err := c.GetCoupon(cp.Address); err != nil {
		t.Fatalf("GetCoupon: %v", err)
	}
	if _, err := c.ListCoupons("", tok.Address, "", nil, 1, 10, true); err != nil {
		t.Fatalf("ListCoupons: %v", err)
	}
}

func TestCouponFlow_NonFungible(t *testing.T) {
	c := setupClient(t)

	// =========================
	// Owner / Token
	// =========================
	owner, ownerPriv := createWallet(t, c)
	c.SetPrivateKey(ownerPriv)
	stablecoin := false

	dec := 0
	tok := createBasicToken(
		t,
		c,
		owner.PublicKey,
		dec,
		false,
		tokenV1Domain.NON_FUNGIBLE,
		stablecoin,
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
	// Deploy Coupon Contract
	// =========================
	start := time.Now().Add(2 * time.Second)
	exp := time.Now().Add(25 * time.Minute)

	raw := sha256.Sum256([]byte("e2e-passcode-nft"))
	pcHash := hex.EncodeToString(raw[:])

	var contractState models.ContractStateModel
	deployedContract, err := c.DeployContract1(couponV1.COUPON_CONTRACT_V1)
	if err != nil {
		t.Fatalf("DeployContract: %v", err)
	}
	unmarshalState(t, deployedContract.States[0].Object, &contractState)

	// =========================
	// Add Coupon (owner)
	// =========================
	out, err := c.AddCoupon(
		contractState.Address,
		tok.Address,
		couponV1Domain.DISCOUNT_TYPE_FIXED,
		"",      // percentage
		"10000", // fixed_amount
		"",
		start,
		exp,
		false,
		true,
		1,
		1,
		pcHash,
	)
	if err != nil {
		t.Fatalf("AddCoupon NFT: %v", err)
	}

	var cp couponV1Domain.Coupon
	unmarshalState(t, out.States[0].Object, &cp)
	if cp.Address == "" {
		t.Fatalf("coupon addr empty")
	}

	// =========================
	// Allow coupon address
	// =========================
	if _, err := c.AllowUsers(
		tok.Address,
		map[string]bool{cp.Address: true},
	); err != nil {
		t.Fatalf("AllowUsers (coupon): %v", err)
	}

	// =========================
	// Update Coupon (owner)
	// =========================
	start2 := time.Now().Add(1 * time.Second)
	exp2 := time.Now().Add(10 * time.Minute)

	raw2 := sha256.Sum256([]byte("e2e-passcode-nft-2"))
	pcHash2 := hex.EncodeToString(raw2[:])

	if _, err := c.UpdateCoupon(
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
	); err != nil {
		t.Fatalf("UpdateCoupon NFT: %v", err)
	}

	waitUntil(t, 20*time.Second, func() bool {
		return time.Now().After(start2)
	})

	// =========================
	// Create User
	// =========================
	user, userPriv := createWallet(t, c)

	// Allow user to interact with token (owner signs)
	c.SetPrivateKey(ownerPriv)
	if _, err := c.AllowUsers(
		tok.Address,
		map[string]bool{user.PublicKey: true},
	); err != nil {
		t.Fatalf("AllowUsers (user): %v", err)
	}

	// =========================
	// Transfer NFT owner -> user
	// =========================
	if _, err := c.TransferToken(
		tok.Address,
		user.PublicKey, // transferTo
		"1",
		dec,
		tokenV1Domain.NON_FUNGIBLE,
		nftUUID,
	); err != nil {
		t.Fatalf("Transfer NFT to user: %v", err)
	}

	// =========================
	// Redeem Coupon (user)
	// =========================
	c.SetPrivateKey(userPriv)

	if _, err := c.RedeemCoupon(
		cp.Address,
		"1",
		"e2e-passcode-nft-2",
		tokenV1Domain.NON_FUNGIBLE,
		nftUUID,
	); err != nil {
		t.Fatalf("RedeemCoupon NFT: %v", err)
	}

	// =========================
	// Pause / Unpause (owner)
	// =========================
	c.SetPrivateKey(ownerPriv)

	if _, err := c.PauseCoupon(cp.Address, true); err != nil {
		t.Fatalf("PauseCoupon: %v", err)
	}
	if _, err := c.UnpauseCoupon(cp.Address, false); err != nil {
		t.Fatalf("UnpauseCoupon: %v", err)
	}

	// =========================
	// Getters
	// =========================
	if _, err := c.GetCoupon(cp.Address); err != nil {
		t.Fatalf("GetCoupon: %v", err)
	}
	if _, err := c.ListCoupons("", tok.Address, "", nil, 1, 10, true); err != nil {
		t.Fatalf("ListCoupons: %v", err)
	}
}
