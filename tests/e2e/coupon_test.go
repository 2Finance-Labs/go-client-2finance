package e2e_test

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"
	"time"
	"gitlab.com/2finance/2finance-network/blockchain/utils"
	"gitlab.com/2finance/2finance-network/blockchain/log"
	"github.com/stretchr/testify/assert"
	"fmt"
	
	couponV1 "gitlab.com/2finance/2finance-network/blockchain/contract/couponV1"
	couponV1Domain "gitlab.com/2finance/2finance-network/blockchain/contract/couponV1/domain"
	tokenV1Domain "gitlab.com/2finance/2finance-network/blockchain/contract/tokenV1/domain"
	tokenV1Models "gitlab.com/2finance/2finance-network/blockchain/contract/tokenV1/models"

)

func TestCouponFlow_NonFungible(t *testing.T) {
	c := setupClient(t)
	ownerPub, ownerPriv := genKey(t, c)
	c.SetPrivateKey(ownerPriv)
	fmt.Printf("Owner Public Key: %s\n", ownerPub)
	fmt.Printf("Owner Private Key: %s\n", ownerPriv)
	deployedContract, err := c.DeployContract1(couponV1.COUPON_CONTRACT_V1)
	if err != nil {
		t.Fatalf("DeployContract: %v", err)
	}
	contractLog, err := utils.UnmarshalLog[log.Log](deployedContract.Logs[0])
	if err != nil {
		t.Fatalf("UnmarshalLog (AddWallet.Logs[0]): %v", err)
	}

	address := contractLog.ContractAddress

	start := time.Now().Add(2 * time.Second)
	exp := time.Now().Add(25 * time.Minute)
	raw := sha256.Sum256([]byte("e2e-passcode"))
	pcHash := hex.EncodeToString(raw[:])

	programType := couponV1Domain.DISCOUNT_TYPE_PERCENTAGE
	percentageBPS := "1000" 
	fixedAmount := ""
	minOrder := "50"
	startAt := start
	expiredAt := exp
	paused := false
	stackable := true
	maxRedemptions := 100
	perUserLimit := 5
	passcodeHash := pcHash

	symbol := "TST" + randSuffix(4)
	name := "Test Token"
	amount := "1000"
	description := "This is a test token"
	image := "https://example.com/image.png"
	website := "https://example.com"
	tagsSocialMedia := map[string]string{"twitter": "@example", "discord": "example#1234"}
	tagsCategory := map[string]string{"type": "utility", "industry": "gaming"}
	tags := map[string]string{"tag1": "value1", "tag2": "value2"}
	creator := "Example Creator"
	creatorWebsite := "https://creator.example.com"
	assetGLBUri := "https://example.com/asset.glb"

	outAddCoupon, err := c.AddCoupon(
		address, 
		programType, 
		percentageBPS, 
		fixedAmount, 
		minOrder, 
		startAt, 
		expiredAt, 
		paused, 
		stackable, 
		maxRedemptions, 
		perUserLimit, 
		passcodeHash,
		symbol,
		name,
		amount,
		description,
		image,
		website,
		tagsSocialMedia,
		tagsCategory,
		tags,
		creator,
		creatorWebsite,
		assetGLBUri,
	)
	if err != nil {
		t.Fatalf("AddCoupon: %v", err)
	}
	
	couponLog, err := utils.UnmarshalLog[log.Log](outAddCoupon.Logs[0])
	if err != nil {
		t.Fatalf("UnmarshalLog (AddCoupon.Logs[0]): %v", err)
	}

	assert.Equal(t, couponV1Domain.COUPON_CREATED_LOG, couponLog.LogType, "add-token log type mismatch")
	coupon, err := utils.UnmarshalEvent[couponV1Domain.Coupon](couponLog.Event)
	if err != nil {
		t.Fatalf("UnmarshalEvent (AddToken.Logs[0]): %v", err)
	}
	
	assert.Equal(t, address, coupon.Address, "coupon address empty")
	assert.NotNil(t, coupon.TokenAddress, "coupon token address mismatch")
	assert.Equal(t, programType, coupon.ProgramType, "coupon discount type mismatch")
	assert.Equal(t, percentageBPS, coupon.PercentageBPS, "coupon percentage BPS mismatch")
	assert.Equal(t, fixedAmount, coupon.FixedAmount, "coupon fixed amount mismatch")
	assert.Equal(t, minOrder, coupon.MinOrder, "coupon min order mismatch")
	assert.WithinDuration(t, startAt, coupon.StartAt, time.Second, "coupon start time mismatch")
	assert.WithinDuration(t, expiredAt, coupon.ExpiredAt, time.Second, "coupon expired at mismatch")
	assert.Equal(t, paused, coupon.Paused, "coupon paused mismatch")
	assert.Equal(t, stackable, coupon.Stackable, "coupon stackable mismatch")
	assert.Equal(t, maxRedemptions, coupon.MaxRedemptions, "coupon max redemptions mismatch")
	assert.Equal(t, perUserLimit, coupon.PerUserLimit, "coupon per user limit mismatch")
	assert.Equal(t, passcodeHash, coupon.PasscodeHash, "coupon passcode hash mismatch")	

	outBalance, err := c.ListTokenBalances(coupon.TokenAddress, address, tokenV1Domain.NON_FUNGIBLE, 1, 3, true)
	if err != nil {
		t.Fatalf("GetTokenBalance: %v", err)
	}
	var balanceStates []tokenV1Models.BalanceStateModel
	err = utils.UnmarshalState[[]tokenV1Models.BalanceStateModel](outBalance.States[0].Object, &balanceStates)
	if err != nil {
		t.Fatalf("UnmarshalState (GetTokenBalance.States[0]): %v", err)
	}

	assert.Equal(t, coupon.TokenAddress, balanceStates[0].TokenAddress, "balance token address mismatch")
	assert.Equal(t, address, balanceStates[0].OwnerAddress, "balance wallet address mismatch")
	assert.Equal(t, "1", balanceStates[0].Amount, "balance amount mismatch for NFT token")
	assert.NotNil(t, balanceStates[0].TokenUUID, "balance token uuid mismatch for NFT token")
	fmt.Printf("Token Address: %s\n", balanceStates[0].TokenAddress)
	fmt.Printf("Owner Address: %s\n", balanceStates[0].OwnerAddress)
	fmt.Printf("Balance Amount: %s\n", balanceStates[0].Amount)
	fmt.Printf("Token UUID: %s\n", balanceStates[0].TokenUUID)
	// Update coupon to fixed amount
	updatedProgramType := couponV1Domain.DISCOUNT_TYPE_FIXED
	updatedPercentageBPS := "" 
	updatedFixedAmount := "10000"
	updatedMinOrder := "50000"
	updatedStart := time.Now().Add(1 * time.Second)
	updatedExp := time.Now().Add(10 * time.Minute)
	raw2 := sha256.Sum256([]byte("e2e-passcode-2"))
	updatedStackable := false
	updatedMaxRedemptions := 50
	updatedPerUserLimit := 2
	updatedPcHash := hex.EncodeToString(raw2[:])
	

	outUpdateCoupon, err := c.UpdateCoupon(
		coupon.Address, 
		coupon.TokenAddress, 
		updatedProgramType, 
		updatedPercentageBPS, 
		updatedFixedAmount, 
		updatedMinOrder, 
		updatedStart, 
		updatedExp, 
		updatedStackable, 
		updatedMaxRedemptions, 
		updatedPerUserLimit, 
		updatedPcHash,
	) 
	if err != nil {
		t.Fatalf("UpdateCoupon: %v", err)
	}

	updateCouponLog, err := utils.UnmarshalLog[log.Log](outUpdateCoupon.Logs[0])
	if err != nil {
		t.Fatalf("UnmarshalLog (UpdateCoupon.Logs[0]): %v", err)
	}

	assert.Equal(t, couponV1Domain.COUPON_UPDATED_LOG, updateCouponLog.LogType, "update-coupon log type mismatch")
	updatedCoupon, err := utils.UnmarshalEvent[couponV1Domain.Coupon](updateCouponLog.Event)
	if err != nil {
		t.Fatalf("UnmarshalEvent (UpdateCoupon.Logs[0]): %v", err)
	}

	assert.Equal(t, coupon.Address, updatedCoupon.Address, "updated coupon address mismatch")
	assert.Equal(t, coupon.TokenAddress, updatedCoupon.TokenAddress, "updated coupon token address mismatch")
	assert.Equal(t, updatedProgramType, updatedCoupon.ProgramType, "updated coupon discount type mismatch")
	assert.Equal(t, updatedPercentageBPS, updatedCoupon.PercentageBPS, "updated coupon percentage BPS mismatch")
	assert.Equal(t, updatedFixedAmount, updatedCoupon.FixedAmount, "updated coupon fixed amount mismatch")
	assert.Equal(t, updatedMinOrder, updatedCoupon.MinOrder, "updated coupon min order mismatch")
	assert.WithinDuration(t, updatedStart, updatedCoupon.StartAt, time.Second, "updated coupon start time mismatch")
	assert.WithinDuration(t, updatedExp, updatedCoupon.ExpiredAt, time.Second, "updated coupon expired at mismatch")
	assert.Equal(t, paused, updatedCoupon.Paused, "updated coupon paused mismatch") // paused unchanged
	assert.Equal(t, updatedStackable, updatedCoupon.Stackable, "updated coupon stackable mismatch")
	assert.Equal(t, updatedMaxRedemptions, updatedCoupon.MaxRedemptions, "updated coupon max redemptions mismatch")
	assert.Equal(t, updatedPerUserLimit, updatedCoupon.PerUserLimit, "updated coupon per user limit mismatch")
	assert.Equal(t, updatedPcHash, updatedCoupon.PasscodeHash, "updated coupon passcode hash mismatch")

	issueAmount := "101"
	issueOutput, err := c.IssueCoupon(coupon.Address, ownerPub, issueAmount)
	if err != nil {
		t.Fatalf("IssueCoupon: %v", err)
	}

	issueLog, err := utils.UnmarshalLog[log.Log](issueOutput.Logs[0])
	if err != nil {
		t.Fatalf("UnmarshalLog (IssueCoupon.Logs[0]): %v", err)
	}

	assert.Equal(t, couponV1Domain.COUPON_ISSUED_LOG, issueLog.LogType, "issue-coupon log type mismatch")
	issuedCoupon, err := utils.UnmarshalEvent[couponV1Domain.IssueCoupon](issueLog.Event)
	if err != nil {
		t.Fatalf("UnmarshalEvent (IssueCoupon.Logs[0]): %v", err)
	}

	assert.Equal(t, coupon.Address, issuedCoupon.Address, "issued coupon address mismatch")
	assert.Equal(t, ownerPub, issuedCoupon.ToAddress, "issued coupon recipient address mismatch")
	assert.Equal(t, issueAmount, issuedCoupon.Amount, "issued coupon amount mismatch")

	// _, err = c.RedeemCoupon(coupon.Address, "100", "e2e-passcode-2", "")
	// if err != nil {
	// 	t.Fatalf("RedeemCoupon: %v", err)
	// }

	// // pause/unpause & getters
	// if _, err := c.PauseCoupon(cp.Address, true); err != nil {
	// 	t.Fatalf("PauseCoupon: %v", err)
	// }
	// if _, err = c.UnpauseCoupon(cp.Address, false); err != nil {
	// 	t.Fatalf("UnpauseCoupon: %v", err)
	// }
	// // getters
	// if _, err := c.GetCoupon(cp.Address); err != nil {
	// 	t.Fatalf("GetCoupon: %v", err)
	// }
	// if _, err := c.ListCoupons("", tok.Address, "", nil, 1, 10, true); err != nil {
	// 	t.Fatalf("ListCoupons: %v", err)
	// }
}

// func TestCouponFlow_NonFungible(t *testing.T) {
	// c := setupClient(t)

	// // =========================
	// // Owner / Token
	// // =========================
	// owner, ownerPriv := createWallet(t, c)
	// c.SetPrivateKey(ownerPriv)
	// stablecoin := false

	// dec := 0
	// tok := createBasicToken(
	// 	t,
	// 	c,
	// 	owner.PublicKey,
	// 	dec,
	// 	false,
	// 	tokenV1Domain.NON_FUNGIBLE,
	// 	stablecoin,
	// )

	// // =========================
	// // Mint NFT (owner)
	// // =========================
	// mintOut, err := c.MintToken(
	// 	tok.Address,
	// 	owner.PublicKey,
	// 	"1",
	// 	dec,
	// 	tok.TokenType,
	// )
	// if err != nil {
	// 	t.Fatalf("MintToken NFT: %v", err)
	// }

	// var mint tokenV1Domain.Mint
	// unmarshalState(t, mintOut.States[0].Object, &mint)

	// if len(mint.TokenUUIDList) != 1 {
	// 	t.Fatalf("expected 1 NFT uuid, got %d", len(mint.TokenUUIDList))
	// }
	// nftUUID := mint.TokenUUIDList[0]

	// // =========================
	// // Deploy Coupon Contract
	// // =========================
	// start := time.Now().Add(2 * time.Second)
	// exp := time.Now().Add(25 * time.Minute)

	// raw := sha256.Sum256([]byte("e2e-passcode-nft"))
	// pcHash := hex.EncodeToString(raw[:])

	// var contractState models.ContractStateModel
	// deployedContract, err := c.DeployContract1(couponV1.COUPON_CONTRACT_V1)
	// if err != nil {
	// 	t.Fatalf("DeployContract: %v", err)
	// }
	// unmarshalState(t, deployedContract.States[0].Object, &contractState)

	// // =========================
	// // Add Coupon (owner)
	// // =========================
	// out, err := c.AddCoupon(
	// 	contractState.Address,
	// 	tok.Address,
	// 	couponV1Domain.DISCOUNT_TYPE_FIXED,
	// 	"",      // percentage
	// 	"10000", // fixed_amount
	// 	"",
	// 	start,
	// 	exp,
	// 	false,
	// 	true,
	// 	1,
	// 	1,
	// 	pcHash,
	// )
	// if err != nil {
	// 	t.Fatalf("AddCoupon NFT: %v", err)
	// }

	// var cp couponV1Domain.Coupon
	// unmarshalState(t, out.States[0].Object, &cp)
	// if cp.Address == "" {
	// 	t.Fatalf("coupon addr empty")
	// }

	// // =========================
	// // Allow coupon address
	// // =========================
	// if _, err := c.AllowUsers(
	// 	tok.Address,
	// 	map[string]bool{cp.Address: true},
	// ); err != nil {
	// 	t.Fatalf("AllowUsers (coupon): %v", err)
	// }

	// // =========================
	// // Update Coupon (owner)
	// // =========================
	// start2 := time.Now().Add(1 * time.Second)
	// exp2 := time.Now().Add(10 * time.Minute)

	// raw2 := sha256.Sum256([]byte("e2e-passcode-nft-2"))
	// pcHash2 := hex.EncodeToString(raw2[:])

	// if _, err := c.UpdateCoupon(
	// 	cp.Address,
	// 	tok.Address,
	// 	couponV1Domain.DISCOUNT_TYPE_FIXED,
	// 	"",
	// 	"10000",
	// 	"",
	// 	start2,
	// 	exp2,
	// 	false,
	// 	1,
	// 	1,
	// 	pcHash2,
	// ); err != nil {
	// 	t.Fatalf("UpdateCoupon NFT: %v", err)
	// }

	// waitUntil(t, 20*time.Second, func() bool {
	// 	return time.Now().After(start2)
	// })

	// // =========================
	// // Create User
	// // =========================
	// user, userPriv := createWallet(t, c)

	// // Allow user to interact with token (owner signs)
	// c.SetPrivateKey(ownerPriv)
	// if _, err := c.AllowUsers(
	// 	tok.Address,
	// 	map[string]bool{user.PublicKey: true},
	// ); err != nil {
	// 	t.Fatalf("AllowUsers (user): %v", err)
	// }

	// // =========================
	// // Transfer NFT owner -> user
	// // =========================
	// if _, err := c.TransferToken(
	// 	tok.Address,
	// 	user.PublicKey, // transferTo
	// 	"1",
	// 	dec,
	// 	tokenV1Domain.NON_FUNGIBLE,
	// 	nftUUID,
	// ); err != nil {
	// 	t.Fatalf("Transfer NFT to user: %v", err)
	// }

	// // =========================
	// // Redeem Coupon (user)
	// // =========================
	// c.SetPrivateKey(userPriv)

	// if _, err := c.RedeemCoupon(
	// 	cp.Address,
	// 	"1",
	// 	"e2e-passcode-nft-2",
	// 	tokenV1Domain.NON_FUNGIBLE,
	// 	nftUUID,
	// ); err != nil {
	// 	t.Fatalf("RedeemCoupon NFT: %v", err)
	// }

	// // =========================
	// // Pause / Unpause (owner)
	// // =========================
	// c.SetPrivateKey(ownerPriv)

	// if _, err := c.PauseCoupon(cp.Address, true); err != nil {
	// 	t.Fatalf("PauseCoupon: %v", err)
	// }
	// if _, err := c.UnpauseCoupon(cp.Address, false); err != nil {
	// 	t.Fatalf("UnpauseCoupon: %v", err)
	// }

	// // =========================
	// // Getters
	// // =========================
	// if _, err := c.GetCoupon(cp.Address); err != nil {
	// 	t.Fatalf("GetCoupon: %v", err)
	// }
	// if _, err := c.ListCoupons("", tok.Address, "", nil, 1, 10, true); err != nil {
	// 	t.Fatalf("ListCoupons: %v", err)
	// }
// }
