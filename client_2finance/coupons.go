package client_2finance

import (
	"fmt"
	"time"

	"gitlab.com/2finance/2finance-network/blockchain/contract/couponV1"
	"gitlab.com/2finance/2finance-network/blockchain/contract/tokenV1/domain"

	"gitlab.com/2finance/2finance-network/blockchain/encryption/keys"
	"gitlab.com/2finance/2finance-network/blockchain/types"
)

// ---------------------------------------------
// Write methods
// ---------------------------------------------

func (c *networkClient) AddCoupon(
	address string, // optional, depends on your infra
	tokenAddress string,
	programType string,   // "percentage" | "fixed-amount"
	percentageBPS string, // required if percentage
	fixedAmount string,   // required if fixed-amount
	minOrder string,      // optional, "" means none
	startAt time.Time,
	expiredAt time.Time,
	paused bool,
	stackable bool,
	maxRedemptions int,
	perUserLimit int,
	passcodeHash string, // sha256(preimage)
) (types.ContractOutput, error) {

	// Sender validations
	from := c.publicKey

	if err := keys.ValidateEDDSAPublicKey(from); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid from address: %w", err)
	}
	if address == "" {
		return types.ContractOutput{}, fmt.Errorf("address not set")
	}
	if err := keys.ValidateEDDSAPublicKey(address); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid coupon address: %w", err)
	}
	if tokenAddress == "" {
		return types.ContractOutput{}, fmt.Errorf("token address not set")
	}
	if err := keys.ValidateEDDSAPublicKey(tokenAddress); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid token address: %w", err)
	}
	if !(programType == "percentage" || programType == "fixed-amount") {
		return types.ContractOutput{}, fmt.Errorf("invalid program_type: %s", programType)
	}
	// Basic param sanity (business rules enforced again in contract/domain)
	if programType == "percentage" && percentageBPS == "" {
		return types.ContractOutput{}, fmt.Errorf("percentage_bps must be set for program_type=percentage")
	}
	if programType == "fixed-amount" && fixedAmount == "" {
		return types.ContractOutput{}, fmt.Errorf("fixed_amount must be set for program_type=fixed-amount")
	}
	// Deploy new coupon program
	to := address
	method := couponV1.METHOD_ADD_COUPON
	data := map[string]interface{}{
		"address":          address,       // optional, depends on your infra
		"token_address":    tokenAddress,
		"program_type":     programType,
		"percentage_bps":   percentageBPS,
		"fixed_amount":     fixedAmount,
		"min_order":        minOrder,
		"start_at":         startAt,
		"expired_at":       expiredAt,
		"paused":           paused,
		"stackable":        stackable,
		"max_redemptions":  maxRedemptions,
		"per_user_limit":   perUserLimit,
		"passcode_hash":    passcodeHash, // sha256(preimage) hex
	}

	return c.SignAndSendTransaction(c.chainId, from, to, method, data)
}

func (c *networkClient) UpdateCoupon(
	address string,
	tokenAddress string,
	programType string,
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
	if err := keys.ValidateEDDSAPublicKey(address); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid coupon address: %w", err)
	}
	if tokenAddress != "" {
		if err := keys.ValidateEDDSAPublicKey(tokenAddress); err != nil {
			return types.ContractOutput{}, fmt.Errorf("invalid token address: %w", err)
		}
	}
	if programType != "" && !(programType == "percentage" || programType == "fixed-amount") {
		return types.ContractOutput{}, fmt.Errorf("invalid program_type: %s", programType)
	}

	from := c.publicKey

	if err := keys.ValidateEDDSAPublicKey(from); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid from address: %w", err)
	}

	to := address
	method := couponV1.METHOD_UPDATE_COUPON

	data := map[string]interface{}{
		"address":         address,
		"token_address":   tokenAddress,   // optional; handler may ignore if ""
		"program_type":    programType,    // optional; handler may ignore if ""
		"percentage_bps":  percentageBPS,  // optional
		"fixed_amount":    fixedAmount,    // optional
		"min_order":       minOrder,       // optional
		"start_at":        startAt,
		"expired_at":      expiredAt,
		"stackable":       stackable,
		"max_redemptions": maxRedemptions,
		"per_user_limit":  perUserLimit,
		"passcode_hash":   passcodeHash,   // "" => keep prior hash
	}

	return c.SignAndSendTransaction(c.chainId, from, to, method, data)
}

func (c *networkClient) PauseCoupon(address string, pause bool) (types.ContractOutput, error) {
	if address == "" {
		return types.ContractOutput{}, fmt.Errorf("address not set")
	}
	if err := keys.ValidateEDDSAPublicKey(address); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid coupon address: %w", err)
	}
	if !pause {
		return types.ContractOutput{}, fmt.Errorf("pause must be true: paused=%t", pause)
	}

	from := c.publicKey

	if err := keys.ValidateEDDSAPublicKey(from); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid from address: %w", err)
	}

	to := address
	method := couponV1.METHOD_PAUSE_COUPON

	data := map[string]interface{}{
		"address": address,
		"paused":  pause,
	}

	return c.SignAndSendTransaction(c.chainId, from, to, method, data)
}

func (c *networkClient) UnpauseCoupon(address string, pause bool) (types.ContractOutput, error) {
	if address == "" {
		return types.ContractOutput{}, fmt.Errorf("address not set")
	}
	if err := keys.ValidateEDDSAPublicKey(address); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid coupon address: %w", err)
	}
	if pause {
		return types.ContractOutput{}, fmt.Errorf("pause must be false: paused=%t", pause)
	}

	from := c.publicKey

	if err := keys.ValidateEDDSAPublicKey(from); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid from address: %w", err)
	}

	to := address
	method := couponV1.METHOD_UNPAUSE_COUPON

	data := map[string]interface{}{
		"address": address,
		"paused":  pause,
	}

	return c.SignAndSendTransaction(c.chainId, from, to, method, data)
}

// Redeem a coupon for an order amount using a passcode preimage.
// NOTE: If you bind the hash to the redeemer (recommended), your handler
// should validate msg.sender and the recomputed hash.
func (c *networkClient) RedeemCoupon(
	address string,     // coupon address
	orderAmount string, // integer string in token base units
	passcode string,
	tokenType string,
	uuid string,
) (types.ContractOutput, error) {

	if address == "" {
		return types.ContractOutput{}, fmt.Errorf("address not set")
	}
	if err := keys.ValidateEDDSAPublicKey(address); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid coupon address: %w", err)
	}
	if orderAmount == "" {
		return types.ContractOutput{}, fmt.Errorf("order_amount not set")
	}
	if passcode == "" {
		return types.ContractOutput{}, fmt.Errorf("passcode (preimage) not set")
	}
	if tokenType == "" {
		return types.ContractOutput{}, fmt.Errorf("tokenType not set")
	}
	if tokenType == domain.NON_FUNGIBLE {
		if uuid == "" {
			return types.ContractOutput{}, fmt.Errorf("uuid must be set for non-fungible tokens")
		}
	}

	from := c.publicKey

	if err := keys.ValidateEDDSAPublicKey(from); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid from address: %w", err)
	}

	to := address
	method := couponV1.METHOD_REDEEM_COUPON

	data := map[string]interface{}{
		"address":       address,
		"order_amount":  orderAmount,
		"passcode": passcode,
		"token_type":    tokenType,
		"uuid":          uuid,
	}

	return c.SignAndSendTransaction(c.chainId, from, to, method, data)
}

// ---------------------------------------------
// Read methods
// ---------------------------------------------

func (c *networkClient) GetCoupon(address string) (types.ContractOutput, error) {
	from := c.publicKey

	if err := keys.ValidateEDDSAPublicKey(from); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid from address: %w", err)
	}
	if address == "" {
		return types.ContractOutput{}, fmt.Errorf("coupon address must be set")
	}
	if err := keys.ValidateEDDSAPublicKey(address); err != nil {
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

	from := c.publicKey

	if err := keys.ValidateEDDSAPublicKey(from); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid from address: %w", err)
	}
	if owner != "" {
		if err := keys.ValidateEDDSAPublicKey(owner); err != nil {
			return types.ContractOutput{}, fmt.Errorf("invalid owner address: %w", err)
		}
	}
	if tokenAddress != "" {
		if err := keys.ValidateEDDSAPublicKey(tokenAddress); err != nil {
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
		"owner":         owner,
		"program_type":  programType,
		"paused":        paused,   // send as pointer; your read handler can interpret tri-state
		"page":          page,
		"limit":         limit,
		"ascending":     ascending,
		"contract_version": couponV1.COUPON_CONTRACT_V1,
		"token_address": tokenAddress,
	}

	return c.GetState("", method, data)
}
