package client_2finance

import (
	"fmt"
	"time"

	"gitlab.com/2finance/2finance-network/blockchain/contract/faucetV1"
	"gitlab.com/2finance/2finance-network/blockchain/keys"
	"gitlab.com/2finance/2finance-network/blockchain/types"
)

func (c *networkClient) AddFaucet(
	owner string,
	tokenAddress string,
	startTime time.Time,
	expireTime time.Time,
	paused bool,
	requestLimit int,
	amount string,
) (types.ContractOutput, error) {

	from := c.publicKey
	if from == "" {
		return types.ContractOutput{}, fmt.Errorf("from address not set")
	}
	if err := keys.ValidateEDDSAPublicKey(from); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid from address: %w", err)
	}
	if owner == "" {
		return types.ContractOutput{}, fmt.Errorf("owner not set")
	}
	if tokenAddress == "" {
		return types.ContractOutput{}, fmt.Errorf("token address not set")
	}
	if err := keys.ValidateEDDSAPublicKey(owner); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid owner address: %w", err)
	}
	if err := keys.ValidateEDDSAPublicKey(tokenAddress); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid token address: %w", err)
	}
	if amount == "" {
		return types.ContractOutput{}, fmt.Errorf("amount not set")
	}

	to := types.DEPLOY_CONTRACT_ADDRESS
	contractVersion := faucetV1.FAUCET_CONTRACT_V1
	method := faucetV1.METHOD_ADD_FAUCET

	data := map[string]interface{}{
		"owner":            owner,
		"token_address":    tokenAddress,
		"start_time":       startTime,
		"expire_time":      expireTime,
		"paused":           paused,
		"request_limit":    requestLimit,
		"amount":			amount,
	}

	contractOutput, err := c.SendTransaction(
		from,
		to,
		contractVersion,
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
	amount string,
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

	to := address
	contractVersion := faucetV1.FAUCET_CONTRACT_V1
	method := faucetV1.METHOD_UPDATE_FAUCET

	data := map[string]interface{}{
		"address":          address,
		"start_time":       startTime,
		"expire_time":      expireTime,
		"request_limit":    requestLimit,
		"requests_by_user": requestsByUser,
		"amount":			amount,
	}

	contractOutput, err := c.SendTransaction(
		from,
		to,
		contractVersion,
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
	contractVersion := faucetV1.FAUCET_CONTRACT_V1
	method := faucetV1.METHOD_PAUSE_FAUCET

	data := map[string]interface{}{
		"address":       address,
		"paused":         pause,
	}

	contractOutput, err := c.SendTransaction(
		from,
		to,
		contractVersion,
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
	contractVersion := faucetV1.FAUCET_CONTRACT_V1
	method := faucetV1.METHOD_UNPAUSE_FAUCET

	data := map[string]interface{}{
		"address":       address,
		"pause":         pause,
	}

	contractOutput, err := c.SendTransaction(
		from,
		to,
		contractVersion,
		method,
		data,
	)
	if err != nil {
		return types.ContractOutput{}, fmt.Errorf("failed to send transaction: %w", err)
	}

	return contractOutput, nil
}

func (c *networkClient) DepositFunds(address, tokenAddress, amount string) (types.ContractOutput, error) {
	if address == "" {
		return types.ContractOutput{}, fmt.Errorf("address not set")
	}
	if err := keys.ValidateEDDSAPublicKey(address); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid address: %w", err)
	}

	if tokenAddress == "" {
		return types.ContractOutput{}, fmt.Errorf("address not set")
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
	contractVersion := faucetV1.FAUCET_CONTRACT_V1
	method := faucetV1.METHOD_DEPOSIT_FUNDS

	data := map[string]interface{}{
		"address":			address,
		"token_address":	tokenAddress,
		"amount":			amount,
	}

	contractOutput, err := c.SendTransaction(
		from,
		to,
		contractVersion,
		method,
		data,
	)
	if err != nil {
		return types.ContractOutput{}, fmt.Errorf("failed to send transaction: %w", err)
	}

	return contractOutput, nil
}

func (c *networkClient) WithdrawFunds(address, tokenAddress, amount string) (types.ContractOutput, error) {
	if address == "" {
		return types.ContractOutput{}, fmt.Errorf("address not set")
	}
	if err := keys.ValidateEDDSAPublicKey(address); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid address: %w", err)
	}

	if tokenAddress == "" {
		return types.ContractOutput{}, fmt.Errorf("address not set")
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
	contractVersion := faucetV1.FAUCET_CONTRACT_V1
	method := faucetV1.METHOD_WITHDRAW_FUNDS

	data := map[string]interface{}{
		"address":			address,
		"token_address":	tokenAddress,
		"amount":			amount,
	}

	contractOutput, err := c.SendTransaction(
		from,
		to,
		contractVersion,
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
	if err := keys.ValidateEDDSAPublicKey(address); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid address: %w", err)
	}

	if requestLimit < 0 {
		return types.ContractOutput{}, fmt.Errorf("request limit less than zero: %d", requestLimit)
	}

	from := c.publicKey
	if from == "" {
		return types.ContractOutput{}, fmt.Errorf("from address not set")
	}
	if err := keys.ValidateEDDSAPublicKey(from); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid from address: %w", err)
	}

	to := address
	contractVersion := faucetV1.FAUCET_CONTRACT_V1
	method := faucetV1.METHOD_REQUEST_LIMIT_PER_USER

	data := map[string]interface{}{
		"address":			address,
		"request_limit":	requestLimit,
	}

	contractOutput, err := c.SendTransaction(
		from,
		to,
		contractVersion,
		method,
		data,
	)
	if err != nil {
		return types.ContractOutput{}, fmt.Errorf("failed to send transaction: %w", err)
	}

	return contractOutput, nil
}

func (c *networkClient) ClaimFunds(address string) (types.ContractOutput, error) {
	if address == "" {
		return types.ContractOutput{}, fmt.Errorf("address not set")
	}
	if err := keys.ValidateEDDSAPublicKey(address); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid address: %w", err)
	}

	from := c.publicKey
	if from == "" {
		return types.ContractOutput{}, fmt.Errorf("from address not set")
	}
	if err := keys.ValidateEDDSAPublicKey(from); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid from address: %w", err)
	}

	to := address
	contractVersion := faucetV1.FAUCET_CONTRACT_V1
	method := faucetV1.METHOD_CLAIM_FUNDS

	data := map[string]interface{}{
		"address":			address,
	}

	contractOutput, err := c.SendTransaction(
		from,
		to,
		contractVersion,
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
	if from == "" {
		return types.ContractOutput{}, fmt.Errorf("from address not set")
	}
	
	if faucetAddress == "" {
		return types.ContractOutput{}, fmt.Errorf("faucet address must be set")
	}

	if err := keys.ValidateEDDSAPublicKey(from); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid from address: %w", err)
	}

	if err := keys.ValidateEDDSAPublicKey(faucetAddress); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid faucet address: %w", err)
	}

	contractVersion := faucetV1.FAUCET_CONTRACT_V1
	method := faucetV1.METHOD_GET_FAUCET
	data := map[string]interface{}{
		"address": faucetAddress,
	}

	contractOutput, err := c.GetState(contractVersion, method, data)
	if err != nil {
		return types.ContractOutput{}, fmt.Errorf("failed to get state: %w", err)
	}

	return contractOutput, nil
}

func (c *networkClient) ListFaucets(
	address, ownerAddress, tokenAddress string,
	requestLimit int,
	requestsByUser map[string]int,
	page, limit int,
	ascending bool,
) (types.ContractOutput, error) {
	from := c.publicKey
	if from == "" {
		return types.ContractOutput{}, fmt.Errorf("from address not set")
	}

	if err := keys.ValidateEDDSAPublicKey(from); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid from address: %w", err)
	}

	if ownerAddress != "" {
		if err := keys.ValidateEDDSAPublicKey(ownerAddress); err != nil {
			return types.ContractOutput{}, fmt.Errorf("invalid owner address: %w", err)
		}
	}

	if address != "" {
		if err := keys.ValidateEDDSAPublicKey(address); err != nil {
			return types.ContractOutput{}, fmt.Errorf("invalid faucet address: %w", err)
		}
	}

	if tokenAddress != "" {
		if err := keys.ValidateEDDSAPublicKey(tokenAddress); err != nil {
			return types.ContractOutput{}, fmt.Errorf("invalid token address: %w", err)
		}
	}

	if page < 1 {
		return types.ContractOutput{}, fmt.Errorf("page must be greater than 0")
	}
	if limit < 1 {
		return types.ContractOutput{}, fmt.Errorf("limit must be greater than 0")
	}

	contractVersion := faucetV1.FAUCET_CONTRACT_V1
	method := faucetV1.METHOD_LIST_FAUCETS

	data := map[string]interface{}{
		"address":         address,
		"owner":           ownerAddress,
		"token_address":   tokenAddress,
		"request_limit":   requestLimit,
		"requests_by_user": requestsByUser,
		"page":            page,
		"limit":           limit,
		"ascending":       ascending,
	}

	contractOutput, err := c.GetState(contractVersion, method, data)
	if err != nil {
		return types.ContractOutput{}, fmt.Errorf("failed to list faucet states: %w", err)
	}

	return contractOutput, nil
}
