package client_2finance

import (
	"fmt"
	"time"

	"gitlab.com/2finance/2finance-network/blockchain/contract/faucetV1"
	"gitlab.com/2finance/2finance-network/blockchain/contract/tokenV1/domain"
	"gitlab.com/2finance/2finance-network/blockchain/encryption/keys"
	"gitlab.com/2finance/2finance-network/blockchain/types"
)

func (c *networkClient) AddFaucet(
	address string,
	owner string,
	tokenAddress string,
	startTime time.Time,
	expireTime time.Time,
	paused bool,
	requestLimit int,
	claimAmount string,
	claimIntervalDuration time.Duration,
) (types.ContractOutput, error) {

	from := c.publicKey

	if err := keys.ValidateEDDSAPublicKeyHex(from); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid from address: %w", err)
	}
	if address == "" {
		return types.ContractOutput{}, fmt.Errorf("address not set")
	}
	if err := keys.ValidateEDDSAPublicKeyHex(address); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid address: %w", err)
	}
	if owner == "" {
		return types.ContractOutput{}, fmt.Errorf("owner not set")
	}
	if tokenAddress == "" {
		return types.ContractOutput{}, fmt.Errorf("token address not set")
	}
	if err := keys.ValidateEDDSAPublicKeyHex(owner); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid owner address: %w", err)
	}
	if err := keys.ValidateEDDSAPublicKeyHex(tokenAddress); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid token address: %w", err)
	}
	if claimAmount == "" {
		return types.ContractOutput{}, fmt.Errorf("amount not set")
	}

	to := address
	method := faucetV1.METHOD_ADD_FAUCET

	data := map[string]interface{}{
		"address":                 address,
		"owner":                   owner,
		"token_address":           tokenAddress,
		"start_time":              startTime,
		"expire_time":             expireTime,
		"paused":                  paused,
		"request_limit":           requestLimit,
		"claim_amount":            claimAmount,
		"claim_interval_duration": claimIntervalDuration,
	}

	contractOutput, err := c.SignAndSendTransaction(
		c.chainId,
		from,
		to,
		method,
		data,
	)
	if err != nil {
		return types.ContractOutput{}, fmt.Errorf("failed to send transaction: %w", err)
	}

	return contractOutput, nil
}

func (c *networkClient) UpdateFaucet(
	address string,
	startTime time.Time,
	expireTime time.Time,
	requestLimit int,
	requestsByUser map[string]int,
	claimAmount string,
	claimIntervalDuration time.Duration,
	lastClaimByUser map[string]time.Time,
) (types.ContractOutput, error) {

	from := c.publicKey

	if err := keys.ValidateEDDSAPublicKeyHex(from); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid from address: %w", err)
	}
	if address == "" {
		return types.ContractOutput{}, fmt.Errorf("address not set")
	}

	to := address
	method := faucetV1.METHOD_UPDATE_FAUCET

	data := map[string]interface{}{
		"address":                 address,
		"start_time":              startTime,
		"expire_time":             expireTime,
		"request_limit":           requestLimit,
		"requests_by_user":        requestsByUser,
		"claim_amount":            claimAmount,
		"claim_interval_duration": claimIntervalDuration,
		"last_claim_by_user":      lastClaimByUser,
	}

	contractOutput, err := c.SignAndSendTransaction(
		c.chainId,
		from,
		to,
		method,
		data,
	)
	if err != nil {
		return types.ContractOutput{}, fmt.Errorf("failed to send transaction: %w", err)
	}

	return contractOutput, nil
}

func (c *networkClient) PauseFaucet(
	address string,
	pause bool,
) (types.ContractOutput, error) {

	if address == "" {
		return types.ContractOutput{}, fmt.Errorf("address not set")
	}
	if err := keys.ValidateEDDSAPublicKeyHex(address); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid address: %w", err)
	}

	if !pause {
		return types.ContractOutput{}, fmt.Errorf("pause must be true: Pause: %t", pause)
	}

	from := c.publicKey

	if err := keys.ValidateEDDSAPublicKeyHex(from); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid from address: %w", err)
	}

	to := address
	method := faucetV1.METHOD_PAUSE_FAUCET

	data := map[string]interface{}{
		"address": address,
		"paused":  pause,
	}

	contractOutput, err := c.SignAndSendTransaction(
		c.chainId,
		from,
		to,
		method,
		data,
	)
	if err != nil {
		return types.ContractOutput{}, fmt.Errorf("failed to send transaction: %w", err)
	}

	return contractOutput, nil
}

func (c *networkClient) UnpauseFaucet(
	address string,
	pause bool,
) (types.ContractOutput, error) {

	if address == "" {
		return types.ContractOutput{}, fmt.Errorf("address not set")
	}
	if err := keys.ValidateEDDSAPublicKeyHex(address); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid address: %w", err)
	}

	if pause {
		return types.ContractOutput{}, fmt.Errorf("pause must be false: Pause: %t", pause)
	}

	from := c.publicKey

	if err := keys.ValidateEDDSAPublicKeyHex(from); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid from address: %w", err)
	}

	to := address
	method := faucetV1.METHOD_UNPAUSE_FAUCET

	data := map[string]interface{}{
		"address": address,
		"pause":   pause,
	}

	contractOutput, err := c.SignAndSendTransaction(
		c.chainId,
		from,
		to,
		method,
		data,
	)
	if err != nil {
		return types.ContractOutput{}, fmt.Errorf("failed to send transaction: %w", err)
	}

	return contractOutput, nil
}

func (c *networkClient) DepositFunds(address, tokenAddress, amount, tokenType, uuid string) (types.ContractOutput, error) {
	if address == "" {
		return types.ContractOutput{}, fmt.Errorf("address not set")
	}
	if err := keys.ValidateEDDSAPublicKeyHex(address); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid address: %w", err)
	}

	if tokenAddress == "" {
		return types.ContractOutput{}, fmt.Errorf("address not set")
	}
	if err := keys.ValidateEDDSAPublicKeyHex(tokenAddress); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid token address: %w", err)
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

	if err := keys.ValidateEDDSAPublicKeyHex(from); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid from address: %w", err)
	}

	to := address
	method := faucetV1.METHOD_DEPOSIT_FUNDS

	data := map[string]interface{}{
		"address":       address,
		"token_address": tokenAddress,
		"amount":        amount,
		"token_type":    tokenType,
		"uuid":          uuid,
	}

	contractOutput, err := c.SignAndSendTransaction(
		c.chainId,
		from,
		to,
		method,
		data,
	)
	if err != nil {
		return types.ContractOutput{}, fmt.Errorf("failed to send transaction: %w", err)
	}

	return contractOutput, nil
}

func (c *networkClient) WithdrawFunds(address, tokenAddress, amount, tokenType, uuid string) (types.ContractOutput, error) {
	if address == "" {
		return types.ContractOutput{}, fmt.Errorf("address not set")
	}
	if err := keys.ValidateEDDSAPublicKeyHex(address); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid address: %w", err)
	}

	if tokenAddress == "" {
		return types.ContractOutput{}, fmt.Errorf("address not set")
	}
	if err := keys.ValidateEDDSAPublicKeyHex(tokenAddress); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid token address: %w", err)
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

	if err := keys.ValidateEDDSAPublicKeyHex(from); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid from address: %w", err)
	}

	to := address
	method := faucetV1.METHOD_WITHDRAW_FUNDS

	data := map[string]interface{}{
		"address":       address,
		"token_address": tokenAddress,
		"amount":        amount,
		"token_type":    tokenType,
		"uuid":          uuid,
	}

	contractOutput, err := c.SignAndSendTransaction(
		c.chainId,
		from,
		to,
		method,
		data,
	)
	if err != nil {
		return types.ContractOutput{}, fmt.Errorf("failed to send transaction: %w", err)
	}

	return contractOutput, nil
}

func (c *networkClient) UpdateRequestLimitPerUser(address string, requestLimit int) (types.ContractOutput, error) {
	if address == "" {
		return types.ContractOutput{}, fmt.Errorf("address not set")
	}
	if err := keys.ValidateEDDSAPublicKeyHex(address); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid address: %w", err)
	}

	if requestLimit < 0 {
		return types.ContractOutput{}, fmt.Errorf("request limit less than zero: %d", requestLimit)
	}

	from := c.publicKey

	if err := keys.ValidateEDDSAPublicKeyHex(from); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid from address: %w", err)
	}

	to := address
	method := faucetV1.METHOD_REQUEST_LIMIT_PER_USER

	data := map[string]interface{}{
		"address":       address,
		"request_limit": requestLimit,
	}

	contractOutput, err := c.SignAndSendTransaction(
		c.chainId,
		from,
		to,
		method,
		data,
	)
	if err != nil {
		return types.ContractOutput{}, fmt.Errorf("failed to send transaction: %w", err)
	}

	return contractOutput, nil
}

func (c *networkClient) ClaimFunds(address, tokenType, uuid string) (types.ContractOutput, error) {
	if address == "" {
		return types.ContractOutput{}, fmt.Errorf("address not set")
	}
	if err := keys.ValidateEDDSAPublicKeyHex(address); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid address: %w", err)
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

	if err := keys.ValidateEDDSAPublicKeyHex(from); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid from address: %w", err)
	}
	method := faucetV1.METHOD_CLAIM_FUNDS
	to := address
	data := map[string]interface{}{
		"address": address,
		"token_type": tokenType,
		"uuid": uuid,
	}

	contractOutput, err := c.SignAndSendTransaction(
		c.chainId,
		from,
		to,
		method,
		data,
	)
	if err != nil {
		return types.ContractOutput{}, fmt.Errorf("failed to send transaction: %w", err)
	}

	return contractOutput, nil
}

func (c *networkClient) GetFaucet(faucetAddress string) (types.ContractOutput, error) {
	from := c.publicKey


	if faucetAddress == "" {
		return types.ContractOutput{}, fmt.Errorf("faucet address must be set")
	}

	if err := keys.ValidateEDDSAPublicKeyHex(from); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid from address: %w", err)
	}

	if err := keys.ValidateEDDSAPublicKeyHex(faucetAddress); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid faucet address: %w", err)
	}

	method := faucetV1.METHOD_GET_FAUCET

	contractOutput, err := c.GetState(faucetAddress, method, nil)
	if err != nil {
		return types.ContractOutput{}, fmt.Errorf("failed to get state: %w", err)
	}

	return contractOutput, nil
}

func (c *networkClient) ListFaucets(
	ownerAddress string,
	page, limit int,
	ascending bool,
) (types.ContractOutput, error) {
	from := c.publicKey


	if err := keys.ValidateEDDSAPublicKeyHex(from); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid from address: %w", err)
	}

	if ownerAddress != "" {
		if err := keys.ValidateEDDSAPublicKeyHex(ownerAddress); err != nil {
			return types.ContractOutput{}, fmt.Errorf("invalid owner address: %w", err)
		}
	}

	if page < 1 {
		return types.ContractOutput{}, fmt.Errorf("page must be greater than 0")
	}
	if limit < 1 {
		return types.ContractOutput{}, fmt.Errorf("limit must be greater than 0")
	}

	method := faucetV1.METHOD_LIST_FAUCETS

	data := map[string]interface{}{
		"owner":            ownerAddress,
		"page":             page,
		"limit":            limit,
		"ascending":        ascending,
		"contract_version": faucetV1.FAUCET_CONTRACT_V1,
	}

	contractOutput, err := c.GetState("", method, data)
	if err != nil {
		return types.ContractOutput{}, fmt.Errorf("failed to list faucet states: %w", err)
	}

	return contractOutput, nil
}
