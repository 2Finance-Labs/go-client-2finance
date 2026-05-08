package client_2finance

import (
	"fmt"
	"time"

	"gitlab.com/2finance/2finance-network/blockchain/contract/couponV1"
	"gitlab.com/2finance/2finance-network/blockchain/encryption/keys"
	"gitlab.com/2finance/2finance-network/blockchain/types"
	"gitlab.com/2finance/2finance-network/blockchain/utils"
)

// ---------------------------------------------
// Write methods
// ---------------------------------------------

func (c *networkClient) AddCoupon(
	address string, // optional, depends on your infra
	discountType string, // "percentage" | "fixed-amount"
	percentageBPS string, // required if percentage
	fixedAmount string, // required if fixed-amount
	minOrder string, // optional, "" means none
	startAt time.Time,
	expiredAt time.Time,
	paused bool,
	stackable bool,
	maxRedemptions int,
	perUserLimit int,
	passcodeHash string, // sha256(preimage)
	voucherOwner string,
	symbol string,
	name string,
	amount string,
	description string,
	image string,
	website string,
	tagsSocialMedia map[string]string,
	tagsCategory map[string]string,
	tags map[string]string,
	creator string,
	creatorWebsite string,
	assetGLBUri string,
) (types.ContractOutput, error) {

	// Sender validations
	from := c.walletManager.GetPublicKey()

	if err := keys.ValidateEDDSAPublicKeyHex(from); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid from address: %w", err)
	}
	if address == "" {
		return types.ContractOutput{}, fmt.Errorf("address not set")
	}
	if err := keys.ValidateEDDSAPublicKeyHex(address); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid coupon address: %w", err)
	}
	if !(discountType == "percentage" || discountType == "fixed-amount") {
		return types.ContractOutput{}, fmt.Errorf("invalid discount_type: %s", discountType)
	}
	// Basic param sanity (business rules enforced again in contract/domain)
	if discountType == "percentage" && percentageBPS == "" {
		return types.ContractOutput{}, fmt.Errorf("percentage_bps must be set for discount_type=percentage")
	}
	if discountType == "fixed-amount" && fixedAmount == "" {
		return types.ContractOutput{}, fmt.Errorf("fixed_amount must be set for discount_type=fixed-amount")
	}
	if voucherOwner == "" {
		return types.ContractOutput{}, fmt.Errorf("voucherOwner must be set")
	}
	if symbol == "" {
		return types.ContractOutput{}, fmt.Errorf("symbol must be set")
	}
	if name == "" {
		return types.ContractOutput{}, fmt.Errorf("name must be set")
	}
	if amount == "" {
		return types.ContractOutput{}, fmt.Errorf("amount must be set")
	}
	if description == "" {
		return types.ContractOutput{}, fmt.Errorf("description must be set")
	}
	if image == "" {
		return types.ContractOutput{}, fmt.Errorf("image must be set")
	}
	if website == "" {
		return types.ContractOutput{}, fmt.Errorf("website must be set")
	}
	if tagsSocialMedia == nil {
		return types.ContractOutput{}, fmt.Errorf("tagsSocialMedia must be set")
	}
	if tagsCategory == nil {
		return types.ContractOutput{}, fmt.Errorf("tagsCategory must be set")
	}
	if tags == nil {
		return types.ContractOutput{}, fmt.Errorf("tags must be set")
	}
	if creator == "" {
		return types.ContractOutput{}, fmt.Errorf("creator must be set")
	}
	if creatorWebsite == "" {
		return types.ContractOutput{}, fmt.Errorf("creatorWebsite must be set")
	}
	if assetGLBUri == "" {
		return types.ContractOutput{}, fmt.Errorf("assetGLBUri must be set")
	}
	// Deploy new coupon program
	to := address
	method := couponV1.METHOD_ADD_COUPON
	data := map[string]interface{}{
		"address":           address, // optional, depends on your infra
		"discount_type":     discountType,
		"percentage_bps":    percentageBPS,
		"fixed_amount":      fixedAmount,
		"min_order":         minOrder,
		"start_at":          startAt,
		"expired_at":        expiredAt,
		"paused":            paused,
		"stackable":         stackable,
		"max_redemptions":   maxRedemptions,
		"per_user_limit":    perUserLimit,
		"passcode_hash":     passcodeHash, // sha256(preimage) hex
		"voucher_owner":     voucherOwner,
		"symbol":            symbol,
		"name":              name,
		"amount":            amount,
		"description":       description,
		"image":             image,
		"website":           website,
		"tags_social_media": tagsSocialMedia,
		"tags_category":     tagsCategory,
		"tags":              tags,
		"creator":           creator,
		"creator_website":   creatorWebsite,
		"asset_glb_uri":     assetGLBUri,
	}
	version := uint8(1)
	uuid7, err := utils.NewUUID7()
	if err != nil {
		return types.ContractOutput{}, fmt.Errorf("failed to generate UUIDv7: %w", err)
	}

	return c.SignAndSendTransaction(c.chainId, from, to, method, data, version, uuid7)
}

func (c *networkClient) UpdateCoupon(
	address string,
	tokenAddress string,
	discountType string,
	percentageBPS string,
	fixedAmount string,
	minOrder string,
	startAt time.Time,
	expiredAt time.Time,
	stackable bool,
	maxRedemptions int,
	perUserLimit int,
	passcodeHash string, // optional; pass "" to keep
) (types.ContractOutput, error) {

	if address == "" {
		return types.ContractOutput{}, fmt.Errorf("address not set")
	}
	if err := keys.ValidateEDDSAPublicKeyHex(address); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid coupon address: %w", err)
	}
	if tokenAddress != "" {
		if err := keys.ValidateEDDSAPublicKeyHex(tokenAddress); err != nil {
			return types.ContractOutput{}, fmt.Errorf("invalid token address: %w", err)
		}
	}
	if discountType != "" && !(discountType == "percentage" || discountType == "fixed-amount") {
		return types.ContractOutput{}, fmt.Errorf("invalid discount_type: %s", discountType)
	}

	from := c.walletManager.GetPublicKey()

	if err := keys.ValidateEDDSAPublicKeyHex(from); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid from address: %w", err)
	}

	to := address
	method := couponV1.METHOD_UPDATE_COUPON

	data := map[string]interface{}{
		"address":         address,
		"token_address":   tokenAddress,  // optional; handler may ignore if ""
		"discount_type":   discountType,  // optional; handler may ignore if ""
		"percentage_bps":  percentageBPS, // optional
		"fixed_amount":    fixedAmount,   // optional
		"min_order":       minOrder,      // optional
		"start_at":        startAt,
		"expired_at":      expiredAt,
		"stackable":       stackable,
		"max_redemptions": maxRedemptions,
		"per_user_limit":  perUserLimit,
		"passcode_hash":   passcodeHash, // "" => keep prior hash
	}
	version := uint8(1)
	uuid7, err := utils.NewUUID7()
	if err != nil {
		return types.ContractOutput{}, fmt.Errorf("failed to generate UUIDv7: %w", err)
	}

	return c.SignAndSendTransaction(c.chainId, from, to, method, data, version, uuid7)
}

func (c *networkClient) PauseCoupon(address string, pause bool) (types.ContractOutput, error) {
	if address == "" {
		return types.ContractOutput{}, fmt.Errorf("address not set")
	}
	if err := keys.ValidateEDDSAPublicKeyHex(address); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid coupon address: %w", err)
	}
	if !pause {
		return types.ContractOutput{}, fmt.Errorf("pause must be true: paused=%t", pause)
	}

	from := c.walletManager.GetPublicKey()

	if err := keys.ValidateEDDSAPublicKeyHex(from); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid from address: %w", err)
	}

	to := address
	method := couponV1.METHOD_PAUSE_COUPON

	data := map[string]interface{}{
		"address": address,
		"paused":  pause,
	}
	version := uint8(1)
	uuid7, err := utils.NewUUID7()
	if err != nil {
		return types.ContractOutput{}, fmt.Errorf("failed to generate UUIDv7: %w", err)
	}

	return c.SignAndSendTransaction(c.chainId, from, to, method, data, version, uuid7)
}

func (c *networkClient) UnpauseCoupon(address string, pause bool) (types.ContractOutput, error) {
	if address == "" {
		return types.ContractOutput{}, fmt.Errorf("address not set")
	}
	if err := keys.ValidateEDDSAPublicKeyHex(address); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid coupon address: %w", err)
	}
	if pause {
		return types.ContractOutput{}, fmt.Errorf("pause must be false: paused=%t", pause)
	}

	from := c.walletManager.GetPublicKey()

	if err := keys.ValidateEDDSAPublicKeyHex(from); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid from address: %w", err)
	}

	to := address
	method := couponV1.METHOD_UNPAUSE_COUPON

	data := map[string]interface{}{
		"address": address,
		"paused":  pause,
	}

	version := uint8(1)
	uuid7, err := utils.NewUUID7()
	if err != nil {
		return types.ContractOutput{}, fmt.Errorf("failed to generate UUIDv7: %w", err)
	}
	return c.SignAndSendTransaction(c.chainId, from, to, method, data, version, uuid7)
}

func (c *networkClient) IssueVoucher(
	address string, // coupon address
	toAddress string,
	amount string, // integer string in token base units
) (types.ContractOutput, error) {
	if address == "" {
		return types.ContractOutput{}, fmt.Errorf("address not set")
	}
	if err := keys.ValidateEDDSAPublicKeyHex(address); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid coupon address: %w", err)
	}
	if toAddress == "" {
		return types.ContractOutput{}, fmt.Errorf("to_address not set")
	}
	if err := keys.ValidateEDDSAPublicKeyHex(toAddress); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid to_address: %w", err)
	}
	if amount == "" {
		return types.ContractOutput{}, fmt.Errorf("amount not set")
	}

	from := c.walletManager.GetPublicKey()

	if err := keys.ValidateEDDSAPublicKeyHex(from); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid from address: %w", err)
	}

	to := address
	method := couponV1.METHOD_ISSUE_VOUCHER

	data := map[string]interface{}{
		"address":    address,
		"to_address": toAddress,
		"amount":     amount,
	}

	version := uint8(1)
	uuid7, err := utils.NewUUID7()
	if err != nil {
		return types.ContractOutput{}, fmt.Errorf("failed to generate UUIDv7: %w", err)
	}
	return c.SignAndSendTransaction(c.chainId, from, to, method, data, version, uuid7)
}

// Redeem a coupon for an order amount using a passcode preimage.
// NOTE: If you bind the hash to the redeemer (recommended), your handler
// should validate msg.sender and the recomputed hash.
func (c *networkClient) RedeemVoucher(
	address string, // voucher address
	orderAmount string, // integer string in token base units
	passcode string,
	voucherUUID string,
) (types.ContractOutput, error) {

	if address == "" {
		return types.ContractOutput{}, fmt.Errorf("address not set")
	}
	if err := keys.ValidateEDDSAPublicKeyHex(address); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid voucher address: %w", err)
	}
	if orderAmount == "" {
		return types.ContractOutput{}, fmt.Errorf("order_amount not set")
	}
	if passcode == "" {
		return types.ContractOutput{}, fmt.Errorf("passcode (preimage) not set")
	}
	if voucherUUID == "" {
		return types.ContractOutput{}, fmt.Errorf("voucher_uuid must be set for non-fungible tokens")
	}

	from := c.walletManager.GetPublicKey()

	if err := keys.ValidateEDDSAPublicKeyHex(from); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid from address: %w", err)
	}

	to := address
	method := couponV1.METHOD_REDEEM_VOUCHER

	data := map[string]interface{}{
		"address":      address,
		"order_amount": orderAmount,
		"passcode":     passcode,
		"voucher_uuid": voucherUUID,
	}

	version := uint8(1)
	uuid7, err := utils.NewUUID7()
	if err != nil {
		return types.ContractOutput{}, fmt.Errorf("failed to generate UUIDv7: %w", err)
	}
	return c.SignAndSendTransaction(c.chainId, from, to, method, data, version, uuid7)
}

// ---------------------------------------------
// Read methods
// ---------------------------------------------

func (c *networkClient) GetCoupon(address string) (types.ContractOutput, error) {
	from := c.walletManager.GetPublicKey()

	if err := keys.ValidateEDDSAPublicKeyHex(from); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid from address: %w", err)
	}
	if address == "" {
		return types.ContractOutput{}, fmt.Errorf("coupon address must be set")
	}
	if err := keys.ValidateEDDSAPublicKeyHex(address); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid coupon address: %w", err)
	}

	method := couponV1.METHOD_GET_COUPON

	return c.GetState(address, method, nil)
}

func (c *networkClient) ListCoupons(
	owner string,
	tokenAddress string,
	programType string,
	paused *bool, // tri-state: nil=any, true/false to filter
	page int,
	limit int,
	ascending bool,
) (types.ContractOutput, error) {

	from := c.walletManager.GetPublicKey()

	if err := keys.ValidateEDDSAPublicKeyHex(from); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid from address: %w", err)
	}
	if owner != "" {
		if err := keys.ValidateEDDSAPublicKeyHex(owner); err != nil {
			return types.ContractOutput{}, fmt.Errorf("invalid owner address: %w", err)
		}
	}
	if tokenAddress != "" {
		if err := keys.ValidateEDDSAPublicKeyHex(tokenAddress); err != nil {
			return types.ContractOutput{}, fmt.Errorf("invalid token address: %w", err)
		}
	}
	if programType != "" && !(programType == "percentage" || programType == "fixed-amount") {
		return types.ContractOutput{}, fmt.Errorf("invalid program_type: %s", programType)
	}
	if page < 1 {
		return types.ContractOutput{}, fmt.Errorf("page must be greater than 0")
	}
	if limit < 1 {
		return types.ContractOutput{}, fmt.Errorf("limit must be greater than 0")
	}

	method := couponV1.METHOD_LIST_COUPONS

	data := map[string]interface{}{
		"owner":            owner,
		"program_type":     programType,
		"paused":           paused, // send as pointer; your read handler can interpret tri-state
		"page":             page,
		"limit":            limit,
		"ascending":        ascending,
		"contract_version": couponV1.COUPON_CONTRACT_V1,
		"token_address":    tokenAddress,
	}

	return c.GetState("", method, data)
}
