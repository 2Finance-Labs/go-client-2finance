package client_2finance


import (
	"time"
	"gitlab.com/2finance/2finance-network/blockchain/contract/cashbackV1"

	"gitlab.com/2finance/2finance-network/blockchain/encryption/keys"
	"fmt"
	"gitlab.com/2finance/2finance-network/blockchain/types"
)


// AddCashBack deploys a new cashback program (to = DEPLOY address).
func (c *networkClient) AddCashback(
	address string,
	owner string,
	tokenAddress string,
	programType string,  // "fixed-percentage" | "variable-percentage"
	percentage string,   // string-encoded number
	startAt time.Time,
	expiredAt time.Time,
	paused bool,
) (types.ContractOutput, error) {

	from := c.publicKey
	if from == "" {
		return types.ContractOutput{}, fmt.Errorf("from address not set")
	}
	if err := keys.ValidateEDDSAPublicKey(from); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid from address: %w", err)
	}

	if address == "" {
		return types.ContractOutput{}, fmt.Errorf("address not set")
	}
	if err := keys.ValidateEDDSAPublicKey(address); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid address: %w", err)
	}

	if owner == "" {
		return types.ContractOutput{}, fmt.Errorf("owner not set")
	}
	if err := keys.ValidateEDDSAPublicKey(owner); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid owner address: %w", err)
	}
	if tokenAddress == "" {
		return types.ContractOutput{}, fmt.Errorf("token address not set")
	}
	if err := keys.ValidateEDDSAPublicKey(tokenAddress); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid token address: %w", err)
	}
	if programType != "fixed-percentage" && programType != "variable-percentage" {
		return types.ContractOutput{}, fmt.Errorf("invalid program_type: %s", programType)
	}
	if percentage == "" {
		return types.ContractOutput{}, fmt.Errorf("percentage not set")
	}

	to := address
	contractVersion := cashbackV1.CASHBACK_CONTRACT_V1
	method := cashbackV1.METHOD_ADD_CASHBACK
	data := map[string]interface{}{
		"address":       address,
		"owner":         owner,
		"token_address": tokenAddress,
		"program_type":  programType,
		"percentage":    percentage,
		"start_at":      startAt,
		"expired_at":    expiredAt,
		"paused":        paused,
	}

	cashback, err := c.SignAndSendTransaction(from, to, contractVersion, method, data)
	if err != nil {
		return types.ContractOutput{}, fmt.Errorf("failed to add cashback: %w", err)
	}
	return cashback, nil
}

// UpdateCashback updates an existing cashback program (to = program address). OnlyOwner.
func (c *networkClient) UpdateCashback(
	address string,
	tokenAddress string,
	programType string, // "fixed-percentage" | "variable-percentage"
	percentage string,
	startAt time.Time,
	expiredAt time.Time,
) (types.ContractOutput, error) {

	if address == "" {
		return types.ContractOutput{}, fmt.Errorf("address not set")
	}
	if err := keys.ValidateEDDSAPublicKey(address); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid address: %w", err)
	}
	if tokenAddress != "" {
		if err := keys.ValidateEDDSAPublicKey(tokenAddress); err != nil {
			return types.ContractOutput{}, fmt.Errorf("invalid token address: %w", err)
		}
	}
	if programType != "fixed-percentage" && programType != "variable-percentage" {
		return types.ContractOutput{}, fmt.Errorf("invalid program_type: %s", programType)
	}
	if percentage == "" {
		return types.ContractOutput{}, fmt.Errorf("percentage not set")
	}

	from := c.publicKey
	if from == "" {
		return types.ContractOutput{}, fmt.Errorf("from address not set")
	}
	if err := keys.ValidateEDDSAPublicKey(from); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid from address: %w", err)
	}

	to := address
	contractVersion := cashbackV1.CASHBACK_CONTRACT_V1
	method := cashbackV1.METHOD_UPDATE_CASHBACK

	data := map[string]interface{}{
		"address":       address,
		"token_address": tokenAddress, // optional: include if your handler supports it
		"program_type":  programType,
		"percentage":    percentage,
		"start_at":      startAt,
		"expired_at":    expiredAt,
	}

	return c.SignAndSendTransaction(from, to, contractVersion, method, data)
}

// PauseCashBack pauses a cashback program. OnlyOwner.
func (c *networkClient) PauseCashback(address string, pause bool) (types.ContractOutput, error) {
	if address == "" {
		return types.ContractOutput{}, fmt.Errorf("address not set")
	}
	if err := keys.ValidateEDDSAPublicKey(address); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid address: %w", err)
	}
	if !pause {
		return types.ContractOutput{}, fmt.Errorf("pause must be true: Pause: %t", pause)
	}

	from := c.publicKey
	if from == "" {
		return types.ContractOutput{}, fmt.Errorf("from address not set")
	}
	if err := keys.ValidateEDDSAPublicKey(from); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid from address: %w", err)
	}

	to := address
	contractVersion := cashbackV1.CASHBACK_CONTRACT_V1
	method := cashbackV1.METHOD_PAUSE_CASHBACK

	data := map[string]interface{}{
		"address": address,
		"paused":  pause,
	}

	return c.SignAndSendTransaction(from, to, contractVersion, method, data)
}

// UnpauseCashback unpauses a cashback program. OnlyOwner.
func (c *networkClient) UnpauseCashback(address string, pause bool) (types.ContractOutput, error) {
	if address == "" {
		return types.ContractOutput{}, fmt.Errorf("address not set")
	}
	if err := keys.ValidateEDDSAPublicKey(address); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid address: %w", err)
	}
	if pause {
		return types.ContractOutput{}, fmt.Errorf("pause must be false: Pause: %t", pause)
	}

	from := c.publicKey
	if from == "" {
		return types.ContractOutput{}, fmt.Errorf("from address not set")
	}
	if err := keys.ValidateEDDSAPublicKey(from); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid from address: %w", err)
	}

	to := address
	contractVersion := cashbackV1.CASHBACK_CONTRACT_V1
	method := cashbackV1.METHOD_UNPAUSE_CASHBACK

	data := map[string]interface{}{
		"address": address,
		"paused":  pause,
	}

	return c.SignAndSendTransaction(from, to, contractVersion, method, data)
}

// DepositCashBack funds the cashback pool (token inferred from state).
func (c *networkClient) DepositCashbackFunds(address, tokenAddress, amount string) (types.ContractOutput, error) {
	if address == "" {
		return types.ContractOutput{}, fmt.Errorf("address not set")
	}
	if err := keys.ValidateEDDSAPublicKey(address); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid address: %w", err)
	}
	if amount == "" {
		return types.ContractOutput{}, fmt.Errorf("amount not set")
	}

	if tokenAddress == "" {
		return types.ContractOutput{}, fmt.Errorf("token address not set")
	}
	if err := keys.ValidateEDDSAPublicKey(tokenAddress); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid token address: %w", err)
	}
	from := c.publicKey
	if from == "" {
		return types.ContractOutput{}, fmt.Errorf("from address not set")
	}
	if err := keys.ValidateEDDSAPublicKey(from); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid from address: %w", err)
	}

	to := address
	contractVersion := cashbackV1.CASHBACK_CONTRACT_V1
	method := cashbackV1.METHOD_DEPOSIT_CASHBACK

	data := map[string]interface{}{
		"address": address,
		"token_address": tokenAddress, // token address inferred from state
		"amount":  amount,
	}

	contractOutput, err := c.SignAndSendTransaction(from, to, contractVersion, method, data)
	if err != nil {
		return types.ContractOutput{}, fmt.Errorf("failed to deposit cashback: %w", err)
	}
	return contractOutput, nil
}

// WithdrawCashback withdraws funds from the cashback pool. OnlyOwner.
func (c *networkClient) WithdrawCashbackFunds(address, tokenAddress, amount string) (types.ContractOutput, error) {
	if address == "" {
		return types.ContractOutput{}, fmt.Errorf("address not set")
	}
	if err := keys.ValidateEDDSAPublicKey(address); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid address: %w", err)
	}
	if amount == "" {
		return types.ContractOutput{}, fmt.Errorf("amount not set")
	}

	if tokenAddress == "" {
		return types.ContractOutput{}, fmt.Errorf("token address not set")
	}
	if err := keys.ValidateEDDSAPublicKey(tokenAddress); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid token address: %w", err)
	}

	from := c.publicKey
	if from == "" {
		return types.ContractOutput{}, fmt.Errorf("from address not set")
	}
	if err := keys.ValidateEDDSAPublicKey(from); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid from address: %w", err)
	}

	to := address
	contractVersion := cashbackV1.CASHBACK_CONTRACT_V1
	method := cashbackV1.METHOD_WITHDRAW_CASHBACK

	data := map[string]interface{}{
		"address": address,
		"amount":  amount,
		"token_address": tokenAddress, // token address inferred from state
	}

	return c.SignAndSendTransaction(from, to, contractVersion, method, data)
}

// GetCashBack reads a single cashback state.
func (c *networkClient) GetCashback(address string) (types.ContractOutput, error) {
	from := c.publicKey
	if from == "" {
		return types.ContractOutput{}, fmt.Errorf("from address not set")
	}
	if address == "" {
		return types.ContractOutput{}, fmt.Errorf("cashback address must be set")
	}
	if err := keys.ValidateEDDSAPublicKey(from); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid from address: %w", err)
	}
	if err := keys.ValidateEDDSAPublicKey(address); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid cashback address: %w", err)
	}

	contractVersion := cashbackV1.CASHBACK_CONTRACT_V1
	method := cashbackV1.METHOD_GET_CASHBACK
	data := map[string]interface{}{
		"address": address,
	}

	return c.GetState(contractVersion, method, data)
}

// ListCashBack queries cashback programs with filters + pagination.
func (c *networkClient) ListCashbacks(
	owner string,
	tokenAddress string,
	programType string,
	paused bool,
	page int,
	limit int,
	ascending bool,
) (types.ContractOutput, error) {
	from := c.publicKey
	if from == "" {
		return types.ContractOutput{}, fmt.Errorf("from address not set")
	}
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
	if programType != "" && programType != "fixed-percentage" && programType != "variable-percentage" {
		return types.ContractOutput{}, fmt.Errorf("invalid program_type: %s", programType)
	}
	if page < 1 {
		return types.ContractOutput{}, fmt.Errorf("page must be greater than 0")
	}
	if limit < 1 {
		return types.ContractOutput{}, fmt.Errorf("limit must be greater than 0")
	}

	contractVersion := cashbackV1.CASHBACK_CONTRACT_V1
	method := cashbackV1.METHOD_LIST_CASHBACKS
	data := map[string]interface{}{
		"owner":         owner,
		"token_address": tokenAddress,
		"program_type":  programType,
		"paused":        paused,
		"page":          page,
		"limit":         limit,
		"ascending":     ascending,
	}

	return c.GetState(contractVersion, method, data)
}

func (c *networkClient) ClaimCashback(address, amount string) (types.ContractOutput, error) {
	if address == "" {
		return types.ContractOutput{}, fmt.Errorf("address not set")
	}
	if err := keys.ValidateEDDSAPublicKey(address); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid address: %w", err)
	}
	if amount == "" {
		return types.ContractOutput{}, fmt.Errorf("amount not set")
	}

	from := c.publicKey
	if from == "" {
		return types.ContractOutput{}, fmt.Errorf("from address not set")
	}
	if err := keys.ValidateEDDSAPublicKey(from); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid from address: %w", err)
	}

	to := address
	contractVersion := cashbackV1.CASHBACK_CONTRACT_V1
	method := cashbackV1.METHOD_CLAIM_CASHBACK

	data := map[string]interface{}{
		"address": address,
		"amount":  amount,
	}

	return c.SignAndSendTransaction(from, to, contractVersion, method, data)
}