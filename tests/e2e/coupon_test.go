package e2e_test

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"
	"time"

	"gitlab.com/2finance/2finance-network/blockchain/contract/contractV1/models"
	couponV1 "gitlab.com/2finance/2finance-network/blockchain/contract/couponV1"
	couponV1Domain "gitlab.com/2finance/2finance-network/blockchain/contract/couponV1/domain"
)

func TestCouponFlow(t *testing.T) {
	c := setupClient(t)
	owner, ownerPriv := createWallet(t, c)
	c.SetPrivateKey(ownerPriv)
	dec := 6
	tok := createBasicToken(t, c, owner.PublicKey, dec, true)

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
	if _, err := c.RedeemCoupon(cp.Address, amt(20, dec), "e2e-passcode-2"); err != nil {
		t.Fatalf("RedeemCoupon warning: %v", err) // allow policy-gate differences
	}

	// pause/unpause & getters
	//TODO -  verify paused state error
	_, _ = c.PauseCoupon(cp.Address, true)
	_, _ = c.UnpauseCoupon(cp.Address, false)
	if _, err := c.GetCoupon(cp.Address); err != nil {
		t.Fatalf("GetCoupon: %v", err)
	}
	if _, err := c.ListCoupons("", tok.Address, "", nil, 1, 10, true); err != nil {
		t.Fatalf("ListCoupons: %v", err)
	}
}
