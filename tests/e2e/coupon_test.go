package e2e_test

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	couponV1 "gitlab.com/2finance/2finance-network/blockchain/contract/couponV1"
	couponV1Domain "gitlab.com/2finance/2finance-network/blockchain/contract/couponV1/domain"
	couponV1Models "gitlab.com/2finance/2finance-network/blockchain/contract/couponV1/models"
	tokenV1Domain "gitlab.com/2finance/2finance-network/blockchain/contract/tokenV1/domain"
	tokenV1Models "gitlab.com/2finance/2finance-network/blockchain/contract/tokenV1/models"
	"gitlab.com/2finance/2finance-network/blockchain/log"
	"gitlab.com/2finance/2finance-network/blockchain/utils"
)

func TestCouponFlow_NonFungible(t *testing.T) {
	wm := setupWalletManager(t)
	c := setupClient(t, wm)
	ownerPub, ownerPriv := genKey(t, wm)
	wm.SetPrivateKey(ownerPriv)

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

	discountType := couponV1Domain.DISCOUNT_TYPE_PERCENTAGE
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

	voucherOwner := ownerPub
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
		discountType,
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
		voucherOwner,
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
	assert.Equal(t, discountType, coupon.DiscountType, "coupon discount type mismatch")
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

	delegatedCreatedToken := outAddCoupon.DelegatedCall[0]
	assert.NotNil(t, delegatedCreatedToken)
	delegatedCreatedTokenLog, err := utils.UnmarshalLog[log.Log](delegatedCreatedToken.Logs[0])
	if err != nil {
		t.Fatalf("UnmarshalLog (AddCoupon.DelegatedCall[0].Logs[0]): %v", err)
	}
	assert.Equal(t, tokenV1Domain.TOKEN_CREATED_LOG, delegatedCreatedTokenLog.LogType, "delegated call log type mismatch")
	delegatedTokenTransfer := outAddCoupon.DelegatedCall[1]
	assert.NotNil(t, delegatedTokenTransfer)
	delegatedTokenTransferLog, err := utils.UnmarshalLog[log.Log](delegatedTokenTransfer.Logs[0])
	if err != nil {
		t.Fatalf("UnmarshalLog (AddCoupon.DelegatedCall[1].Logs[0]): %v", err)
	}
	assert.Equal(t, tokenV1Domain.TOKEN_TRANSFERRED_NFT_LOG, delegatedTokenTransferLog.LogType, "delegated token transfer log type mismatch")

	outBalance, err := c.ListTokenBalances(coupon.TokenAddress, voucherOwner, tokenV1Domain.NON_FUNGIBLE, 1, 3, true)
	if err != nil {
		t.Fatalf("GetTokenBalance: %v", err)
	}
	var balanceStates []tokenV1Models.BalanceStateModel
	err = utils.UnmarshalState[[]tokenV1Models.BalanceStateModel](outBalance.States[0].Object, &balanceStates)
	if err != nil {
		t.Fatalf("UnmarshalState (GetTokenBalance.States[0]): %v", err)
	}

	assert.Equal(t, coupon.TokenAddress, balanceStates[0].TokenAddress, "balance token address mismatch")
	assert.Equal(t, voucherOwner, balanceStates[0].OwnerAddress, "balance wallet address mismatch")
	assert.Equal(t, "1", balanceStates[0].Amount, "balance amount mismatch for NFT token")
	assert.NotNil(t, balanceStates[0].TokenUUID, "balance token uuid mismatch for NFT token")

	// Update coupon to fixed amount
	updatedDiscountType := couponV1Domain.DISCOUNT_TYPE_FIXED
	updatedPercentageBPS := ""
	updatedFixedAmount := "10000"
	updatedMinOrder := "50"
	updatedStart := time.Now().Add(1 * time.Second)
	updatedExp := time.Now().Add(10 * time.Minute)
	raw2 := sha256.Sum256([]byte("e2e-passcode-2"))
	updatedPcHash := hex.EncodeToString(raw2[:])
	updatedStackable := false
	updatedMaxRedemptions := 50
	updatedPerUserLimit := 2

	outUpdateCoupon, err := c.UpdateCoupon(
		coupon.Address,
		coupon.TokenAddress,
		updatedDiscountType,
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
	assert.Equal(t, updatedDiscountType, updatedCoupon.DiscountType, "updated coupon discount type mismatch")
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
	issueOutput, err := c.IssueVoucher(coupon.Address, ownerPub, issueAmount)
	if err != nil {
		t.Fatalf("IssueVoucher: %v", err)
	}

	issueLog, err := utils.UnmarshalLog[log.Log](issueOutput.Logs[0])
	if err != nil {
		t.Fatalf("UnmarshalLog (IssueVoucher.Logs[0]): %v", err)
	}

	assert.Equal(t, couponV1Domain.VOUCHER_ISSUED_LOG, issueLog.LogType, "issue-voucher log type mismatch")
	issuedVoucher, err := utils.UnmarshalEvent[couponV1Domain.IssueVoucher](issueLog.Event)
	if err != nil {
		t.Fatalf("UnmarshalEvent (IssueVoucher.Logs[0]): %v", err)
	}

	assert.Equal(t, coupon.Address, issuedVoucher.Address, "issued voucher address mismatch")
	assert.Equal(t, ownerPub, issuedVoucher.ToAddress, "issued voucher recipient address mismatch")
	assert.Equal(t, issueAmount, issuedVoucher.Amount, "issued voucher amount mismatch")

	delegatedIssue := issueOutput.DelegatedCall[0]
	assert.NotNil(t, delegatedIssue)
	delegatedIssueLog, err := utils.UnmarshalLog[log.Log](delegatedIssue.Logs[0])
	if err != nil {
		t.Fatalf("UnmarshalLog (IssueVoucher.DelegatedCall[0].Logs[0]): %v", err)
	}
	assert.Equal(t, tokenV1Domain.TOKEN_MINTED_NFT_LOG, delegatedIssueLog.LogType, "delegated issue voucher log type mismatch")

	outBalanceOwner, err := c.ListTokenBalances(coupon.TokenAddress, ownerPub, tokenV1Domain.NON_FUNGIBLE, 1, 3, true)
	if err != nil {
		t.Fatalf("GetTokenBalance: %v", err)
	}
	var balanceStatesOwner []tokenV1Models.BalanceStateModel
	err = utils.UnmarshalState[[]tokenV1Models.BalanceStateModel](outBalanceOwner.States[0].Object, &balanceStatesOwner)
	if err != nil {
		t.Fatalf("UnmarshalState (GetTokenBalance.States[0]): %v", err)
	}

	time.Sleep(25 * time.Second) // wait for state to update before redeeming

	orderAmount := "100"
	passcode := "e2e-passcode-2"
	tokenUUID := balanceStatesOwner[0].TokenUUID
	redeemVoucherOutput, err := c.RedeemVoucher(coupon.Address, orderAmount, passcode, tokenUUID)
	if err != nil {
		t.Fatalf("RedeemVoucher: %v", err)
	}

	redeemLog, err := utils.UnmarshalLog[log.Log](redeemVoucherOutput.Logs[0])
	if err != nil {
		t.Fatalf("UnmarshalLog (RedeemVoucher.Logs[0]): %v", err)
	}

	assert.Equal(t, couponV1Domain.VOUCHER_REDEEMED_LOG, redeemLog.LogType, "redeem-voucher log type mismatch")
	redeemedVoucher, err := utils.UnmarshalEvent[couponV1Domain.RedeemVoucher](redeemLog.Event)
	if err != nil {
		t.Fatalf("UnmarshalEvent (RedeemVoucher.Logs[0]): %v", err)
	}

	assert.Equal(t, coupon.Address, redeemedVoucher.CouponAddress, "redeemed voucher address mismatch")
	assert.Equal(t, coupon.TokenAddress, redeemedVoucher.TokenAddress, "redeemed voucher token address mismatch")
	assert.Equal(t, ownerPub, redeemedVoucher.UserAddress, "redeemed voucher sender address mismatch")
	assert.Equal(t, orderAmount, redeemedVoucher.OrderAmount, "redeemed voucher amount mismatch")
	assert.NotEmpty(t, redeemedVoucher.DiscountAmount, "redeemed voucher discount amount empty")
	assert.Equal(t, tokenUUID, redeemedVoucher.VoucherUUID, "redeemed voucher token UUID mismatch")

	// pause/unpause & getters
	pausedOutput, err := c.PauseCoupon(coupon.Address, true)
	if err != nil {
		t.Fatalf("PauseCoupon: %v", err)
	}
	pausedLog, err := utils.UnmarshalLog[log.Log](pausedOutput.Logs[0])
	if err != nil {
		t.Fatalf("UnmarshalLog (PauseCoupon.Logs[0]): %v", err)
	}
	assert.Equal(t, couponV1Domain.COUPON_PAUSED_LOG, pausedLog.LogType, "pause-coupon log type mismatch")
	pausedCoupon, err := utils.UnmarshalEvent[couponV1Domain.Pause](pausedLog.Event)
	if err != nil {
		t.Fatalf("UnmarshalEvent (PauseCoupon.Logs[0]): %v", err)
	}
	assert.Equal(t, coupon.Address, pausedCoupon.Address, "paused coupon address mismatch")
	assert.Equal(t, true, pausedCoupon.Paused, "coupon paused state mismatch after pausing")
	unpausedOutput, err := c.UnpauseCoupon(coupon.Address, false)
	if err != nil {
		t.Fatalf("UnpauseCoupon: %v", err)
	}
	unpausedLog, err := utils.UnmarshalLog[log.Log](unpausedOutput.Logs[0])
	if err != nil {
		t.Fatalf("UnmarshalLog (UnpauseCoupon.Logs[0]): %v", err)
	}
	assert.Equal(t, couponV1Domain.COUPON_UNPAUSED_LOG, unpausedLog.LogType, "unpause-coupon log type mismatch")
	unpausedCoupon, err := utils.UnmarshalEvent[couponV1Domain.Pause](unpausedLog.Event)
	if err != nil {
		t.Fatalf("UnmarshalEvent (UnpauseCoupon.Logs[0]): %v", err)
	}
	assert.Equal(t, coupon.Address, unpausedCoupon.Address, "unpaused coupon address mismatch")
	assert.Equal(t, false, unpausedCoupon.Paused, "coupon paused state mismatch after unpausing")

	// getters
	couponStateOutput, err := c.GetCoupon(coupon.Address)
	if err != nil {
		t.Fatalf("GetCoupon: %v", err)
	}
	var couponState couponV1Models.CouponStateModel
	err = utils.UnmarshalState[couponV1Models.CouponStateModel](couponStateOutput.States[0].Object, &couponState)
	if err != nil {
		t.Fatalf("UnmarshalState (GetCoupon.State): %v", err)
	}
	assert.Equal(t, coupon.Address, couponState.Address, "GetCoupon address mismatch")
	assert.Equal(t, coupon.TokenAddress, couponState.TokenAddress, "GetCoupon token address mismatch")
	assert.Equal(t, updatedCoupon.DiscountType, couponState.DiscountType, "GetCoupon discount type mismatch")
	assert.Equal(t, updatedCoupon.PercentageBPS, couponState.PercentageBPS, "GetCoupon percentage BPS mismatch")
	assert.Equal(t, updatedCoupon.FixedAmount, couponState.FixedAmount, "GetCoupon fixed amount mismatch")
	assert.Equal(t, updatedCoupon.MinOrder, couponState.MinOrder, "GetCoupon min order mismatch")
	assert.WithinDuration(t, updatedCoupon.StartAt, *couponState.StartAt, time.Second, "GetCoupon start time mismatch")
	assert.WithinDuration(t, updatedCoupon.ExpiredAt, *couponState.ExpiredAt, time.Second, "GetCoupon expired at mismatch")
	assert.Equal(t, updatedCoupon.Paused, couponState.Paused, "GetCoupon paused state mismatch")
	assert.Equal(t, updatedCoupon.Stackable, couponState.Stackable, "GetCoupon stackable mismatch")
	assert.Equal(t, updatedCoupon.MaxRedemptions, couponState.MaxRedemptions, "GetCoupon max redemptions mismatch")
	assert.Equal(t, updatedCoupon.PerUserLimit, couponState.PerUserLimit, "GetCoupon per user limit mismatch")
	assert.Equal(t, updatedCoupon.PasscodeHash, couponState.PasscodeHash, "GetCoupon passcode hash mismatch")
	assert.Equal(t, 1, couponState.TotalRedemptions, "GetCoupon total redemptions mismatch")
	assert.NotNil(t, couponState.CreatedAt, "GetCoupon created at nil")
	assert.NotNil(t, couponState.UpdatedAt, "GetCoupon updated at nil")

	couponsListOutput, err := c.ListCoupons("", coupon.TokenAddress, "", nil, 1, 10, true)
	if err != nil {
		t.Fatalf("ListCoupons: %v", err)
	}
	var couponsList []couponV1Models.CouponStateModel
	err = utils.UnmarshalState[[]couponV1Models.CouponStateModel](couponsListOutput.States[0].Object, &couponsList)
	if err != nil {
		t.Fatalf("UnmarshalState (ListCoupons.States[0]): %v", err)
	}
	assert.GreaterOrEqual(t, len(couponsList), 1, "ListCoupons returned empty list")
	assert.Equal(t, coupon.Address, couponsList[0].Address, "ListCoupons first coupon address mismatch")
	assert.Equal(t, coupon.TokenAddress, couponsList[0].TokenAddress, "ListCoupons first coupon token address mismatch")
	assert.Equal(t, updatedCoupon.DiscountType, couponsList[0].DiscountType, "ListCoupons first coupon discount type mismatch")
	assert.Equal(t, updatedCoupon.PercentageBPS, couponsList[0].PercentageBPS, "ListCoupons first coupon percentage BPS mismatch")
	assert.Equal(t, updatedCoupon.FixedAmount, couponsList[0].FixedAmount, "ListCoupons first coupon fixed amount mismatch")
	assert.Equal(t, updatedCoupon.MinOrder, couponsList[0].MinOrder, "ListCoupons first coupon min order mismatch")
	assert.WithinDuration(t, updatedCoupon.StartAt, *couponsList[0].StartAt, time.Second, "ListCoupons first coupon start time mismatch")
	assert.WithinDuration(t, updatedCoupon.ExpiredAt, *couponsList[0].ExpiredAt, time.Second, "ListCoupons first coupon expired at mismatch")
	assert.Equal(t, updatedCoupon.Paused, couponsList[0].Paused, "ListCoupons first coupon paused state mismatch")
	assert.Equal(t, updatedCoupon.Stackable, couponsList[0].Stackable, "ListCoupons first coupon stackable mismatch")
	assert.Equal(t, updatedCoupon.MaxRedemptions, couponsList[0].MaxRedemptions, "ListCoupons first coupon max redemptions mismatch")
	assert.Equal(t, updatedCoupon.PerUserLimit, couponsList[0].PerUserLimit, "ListCoupons first coupon per user limit mismatch")
	assert.Equal(t, updatedCoupon.PasscodeHash, couponsList[0].PasscodeHash, "ListCoupons first coupon passcode hash mismatch")
	assert.Equal(t, 1, couponsList[0].TotalRedemptions, "ListCoupons first coupon total redemptions mismatch")
	assert.NotNil(t, couponsList[0].CreatedAt, "ListCoupons first coupon created at nil")
	assert.NotNil(t, couponsList[0].UpdatedAt, "ListCoupons first coupon updated at nil")
}
