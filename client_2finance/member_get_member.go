package client_2finance

import (
	"fmt"
	"time"

	"gitlab.com/2finance/2finance-network/blockchain/contract/memberGetMemberV1"
	"gitlab.com/2finance/2finance-network/blockchain/keys"
	"gitlab.com/2finance/2finance-network/blockchain/types"
)

func (c *networkClient) AddMgM(
	owner string,
	tokenAddress string,
	faucetAddress string,
	amount string,
	startAt time.Time,
	expireAt time.Time,
	paused bool,
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
	if faucetAddress == "" {
		return types.ContractOutput{}, fmt.Errorf("faucet address not set")
	}
	if err := keys.ValidateEDDSAPublicKey(owner); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid owner address: %w", err)
	}
	if err := keys.ValidateEDDSAPublicKey(tokenAddress); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid token address: %w", err)
	}
	if err := keys.ValidateEDDSAPublicKey(faucetAddress); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid faucet address: %w", err)
	}

	to := types.DEPLOY_CONTRACT_ADDRESS
	contractVersion := memberGetMemberV1.MEMBER_GET_MEMBER_CONTRACT_V1
	method := memberGetMemberV1.METHOD_ADD_MGM

	data := map[string]interface{}{
		"owner":          owner,
		"token_address":  tokenAddress,
		"faucet_address": faucetAddress,
		"amount":         amount,
		"start_at":       startAt,
		"expire_at":      expireAt,
		"paused":         paused,
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

func (c *networkClient) UpdateMgM(
	mgmAddress string,
	amount string,
	startAt time.Time,
	expireAt time.Time,
) (types.ContractOutput, error) {
	from := c.publicKey
	if from == "" {
		return types.ContractOutput{}, fmt.Errorf("from address not set")
	}
	if err := keys.ValidateEDDSAPublicKey(from); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid from address: %w", err)
	}
	if mgmAddress == "" {
		return types.ContractOutput{}, fmt.Errorf("address not set")
	}

	to := mgmAddress
	contractVersion := memberGetMemberV1.MEMBER_GET_MEMBER_CONTRACT_V1
	method := memberGetMemberV1.METHOD_UPDATE_MGM

	data := map[string]interface{}{
		"mgm_address": mgmAddress,
		"amount":      amount,
		"start_at":    startAt,
		"expire_at":   expireAt,
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

func (c *networkClient) PauseMgM(mgmAddress string, pause bool) (types.ContractOutput, error) {
	if mgmAddress == "" {
		return types.ContractOutput{}, fmt.Errorf("address not set")
	}
	if err := keys.ValidateEDDSAPublicKey(mgmAddress); err != nil {
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

	to := mgmAddress
	contractVersion := memberGetMemberV1.MEMBER_GET_MEMBER_CONTRACT_V1
	method := memberGetMemberV1.METHOD_PAUSE_MGM

	data := map[string]interface{}{
		"mgm_address": mgmAddress,
		"paused":      pause,
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

func (c *networkClient) UnpauseMgM(mgmAddress string, pause bool) (types.ContractOutput, error) {
	if mgmAddress == "" {
		return types.ContractOutput{}, fmt.Errorf("address not set")
	}
	if err := keys.ValidateEDDSAPublicKey(mgmAddress); err != nil {
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

	to := mgmAddress
	contractVersion := memberGetMemberV1.MEMBER_GET_MEMBER_CONTRACT_V1
	method := memberGetMemberV1.METHOD_UNPAUSE_MGM

	data := map[string]interface{}{
		"mgm_address": mgmAddress,
		"paused":      pause,
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

func (c *networkClient) DepositMgM(
	mgmAddress string,
	amount string,
) (types.ContractOutput, error) {
	if mgmAddress == "" {
		return types.ContractOutput{}, fmt.Errorf("address not set")
	}
	if err := keys.ValidateEDDSAPublicKey(mgmAddress); err != nil {
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

	to := mgmAddress
	contractVersion := memberGetMemberV1.MEMBER_GET_MEMBER_CONTRACT_V1
	method := memberGetMemberV1.METHOD_DEPOSIT_MGM

	data := map[string]interface{}{
		"mgm_address": mgmAddress,
		"amount":      amount,
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

func (c *networkClient) WithdrawMgM(
	mgmAddress string,
	amount string,
) (types.ContractOutput, error) {
	if mgmAddress == "" {
		return types.ContractOutput{}, fmt.Errorf("address not set")
	}
	if err := keys.ValidateEDDSAPublicKey(mgmAddress); err != nil {
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

	to := mgmAddress
	contractVersion := memberGetMemberV1.MEMBER_GET_MEMBER_CONTRACT_V1
	method := memberGetMemberV1.METHOD_WITHDRAW_MGM

	data := map[string]interface{}{
		"mgm_address": mgmAddress,
		"amount":      amount,
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
