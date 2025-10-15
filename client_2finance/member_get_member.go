package client_2finance

import (
	"fmt"
	"time"
	memberGetMemberV1 "gitlab.com/2finance/2finance-network/blockchain/contract/memberGetMemberV1"
	"gitlab.com/2finance/2finance-network/blockchain/encryption/keys"
	"gitlab.com/2finance/2finance-network/blockchain/types"
)

func (c *networkClient) AddMgM(
	address string,
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
	if address == "" {
		return types.ContractOutput{}, fmt.Errorf("address not set")
	}
	if err := keys.ValidateEDDSAPublicKey(address); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid address: %w", err)
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

	to := address
	method := memberGetMemberV1.METHOD_ADD_MGM
	data := map[string]interface{}{
		"address":       address,
		"owner":          owner,
		"token_address":  tokenAddress,
		"faucet_address": faucetAddress,
		"amount":         amount,
		"start_at":       startAt,
		"expire_at":      expireAt,
		"paused":         paused,
	}

	contractOutput, err := c.SignAndSendTransaction(
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
	method := memberGetMemberV1.METHOD_UPDATE_MGM

	data := map[string]interface{}{
		"mgm_address": mgmAddress,
		"amount":      amount,
		"start_at":    startAt,
		"expire_at":   expireAt,
	}

	contractOutput, err := c.SignAndSendTransaction(
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
	method := memberGetMemberV1.METHOD_PAUSE_MGM

	data := map[string]interface{}{
		"mgm_address": mgmAddress,
		"paused":      pause,
	}

	contractOutput, err := c.SignAndSendTransaction(
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
	method := memberGetMemberV1.METHOD_UNPAUSE_MGM

	data := map[string]interface{}{
		"mgm_address": mgmAddress,
		"paused":      pause,
	}

	contractOutput, err := c.SignAndSendTransaction(
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
	method := memberGetMemberV1.METHOD_DEPOSIT_MGM

	data := map[string]interface{}{
		"mgm_address": mgmAddress,
		"amount":      amount,
	}

	contractOutput, err := c.SignAndSendTransaction(
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
	method := memberGetMemberV1.METHOD_WITHDRAW_MGM

	data := map[string]interface{}{
		"mgm_address": mgmAddress,
		"amount":      amount,
	}

	contractOutput, err := c.SignAndSendTransaction(
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

func (c *networkClient) AddInviterMember(mgmAddress string, inviterAddress string, password string ) (types.ContractOutput, error) {
	if mgmAddress == "" {
		return types.ContractOutput{}, fmt.Errorf("address not set")
	}
	if err := keys.ValidateEDDSAPublicKey(mgmAddress); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid address: %w", err)
	}

	if password == "" {
		return types.ContractOutput{}, fmt.Errorf("password not set")
	}

	from := c.publicKey
	if from == "" {
		return types.ContractOutput{}, fmt.Errorf("from address not set")
	}
	if err := keys.ValidateEDDSAPublicKey(from); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid from address: %w", err)
	}

	to := mgmAddress
	method := memberGetMemberV1.METHOD_ADD_INVITER_MEMBER

	data := map[string]interface{}{
		"mgm_address":   mgmAddress,
		"inviter_address": inviterAddress,
		"password":    password,
	}

	contractOutput, err := c.SignAndSendTransaction(
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

func (c *networkClient) UpdateInviterPassword(mgmAddress string, inviterAddress string, newPassword string) (types.ContractOutput, error) {
	if mgmAddress == "" {
		return types.ContractOutput{}, fmt.Errorf("address not set")
	}
	if err := keys.ValidateEDDSAPublicKey(mgmAddress); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid address: %w", err)
	}
	if newPassword == "" {
		return types.ContractOutput{}, fmt.Errorf("new password not set")
	}

	from := c.publicKey
	if from == "" {
		return types.ContractOutput{}, fmt.Errorf("from address not set")
	}
	if err := keys.ValidateEDDSAPublicKey(from); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid from address: %w", err)
	}

	method := memberGetMemberV1.METHOD_UPDATE_INVITER_PASSWORD

	data := map[string]interface{}{
		"mgm_address":   mgmAddress,
		"inviter_address": inviterAddress,
		"new_password":  newPassword,
	}

	to := mgmAddress
	contractOutput, err := c.SignAndSendTransaction(
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

func (c *networkClient) DeleteInviterMember(mgmAddress string, inviterAddress string) (types.ContractOutput, error) {
	if mgmAddress == "" {
		return types.ContractOutput{}, fmt.Errorf("address not set")
	}
	if err := keys.ValidateEDDSAPublicKey(mgmAddress); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid address: %w", err)
	}

	from := c.publicKey
	if from == "" {
		return types.ContractOutput{}, fmt.Errorf("from address not set")
	}
	if err := keys.ValidateEDDSAPublicKey(from); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid from address: %w", err)
	}

	to := mgmAddress
	method := memberGetMemberV1.METHOD_DELETE_INVITER_MEMBER

	data := map[string]interface{}{
		"mgm_address":   mgmAddress,
		"inviter_address": inviterAddress,
	}

	contractOutput, err := c.SignAndSendTransaction(
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

func (c *networkClient) ClaimReward(mgmAddress, invitedAddress, password string) (types.ContractOutput, error) {
	if mgmAddress == "" {
		return types.ContractOutput{}, fmt.Errorf("address not set")
	}
	if err := keys.ValidateEDDSAPublicKey(mgmAddress); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid address: %w", err)
	}

	if password == "" {
		return types.ContractOutput{}, fmt.Errorf("password not set")
	}

	from := c.publicKey
	if from == "" {
		return types.ContractOutput{}, fmt.Errorf("from address not set")
	}
	if err := keys.ValidateEDDSAPublicKey(from); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid from address: %w", err)
	}

	to := mgmAddress
	method := memberGetMemberV1.METHOD_CLAIM_REWARD

	data := map[string]interface{}{
		"mgm_address":     mgmAddress,
		"password":        password,
		"invited_address": invitedAddress,
	}

	contractOutput, err := c.SignAndSendTransaction(
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

func (c *networkClient) GetMgM(mgmAddress string) (types.ContractOutput, error) {
	if mgmAddress == "" {
		return types.ContractOutput{}, fmt.Errorf("address not set")
	}
	if err := keys.ValidateEDDSAPublicKey(mgmAddress); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid address: %w", err)
	}

	from := c.publicKey
	if from == "" {
		return types.ContractOutput{}, fmt.Errorf("from address not set")
	}
	if err := keys.ValidateEDDSAPublicKey(from); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid from address: %w", err)
	}

	method := memberGetMemberV1.METHOD_GET_MGM

	contractOutput, err := c.GetState(mgmAddress, method, nil)
	if err != nil {
		return types.ContractOutput{}, fmt.Errorf("failed to get state: %w", err)
	}

	return contractOutput, nil
}

func (c *networkClient) GetInviterMember(mgmAddress string, inviterAddress string) (types.ContractOutput, error) {
	if mgmAddress == "" {
		return types.ContractOutput{}, fmt.Errorf("address not set")
	}
	if err := keys.ValidateEDDSAPublicKey(mgmAddress); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid address: %w", err)
	}

	if inviterAddress == "" {
		return types.ContractOutput{}, fmt.Errorf("inviter address not set")
	}
	if err := keys.ValidateEDDSAPublicKey(inviterAddress); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid inviter address: %w", err)
	}

	from := c.publicKey
	if from == "" {
		return types.ContractOutput{}, fmt.Errorf("from address not set")
	}
	if err := keys.ValidateEDDSAPublicKey(from); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid from address: %w", err)
	}

	method := memberGetMemberV1.METHOD_GET_INVITER_MEMBER

	data := map[string]interface{}{
		"inviter_address": inviterAddress,
	}

	contractOutput, err := c.GetState(mgmAddress, method, data)
	if err != nil {
		return types.ContractOutput{}, fmt.Errorf("failed to get state: %w", err)
	}

	return contractOutput, nil
}

func (c *networkClient) GetClaimInviter(mgmAddress string, inviterAddress string) (types.ContractOutput, error) {
	if mgmAddress == "" {
		return types.ContractOutput{}, fmt.Errorf("address not set")
	}
	if err := keys.ValidateEDDSAPublicKey(mgmAddress); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid address: %w", err)
	}

	if inviterAddress == "" {
		return types.ContractOutput{}, fmt.Errorf("inviter address not set")
	}
	if err := keys.ValidateEDDSAPublicKey(inviterAddress); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid inviter address: %w", err)
	}

	from := c.publicKey
	if from == "" {
		return types.ContractOutput{}, fmt.Errorf("from address not set")
	}
	if err := keys.ValidateEDDSAPublicKey(from); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid from address: %w", err)
	}

	method := memberGetMemberV1.METHOD_GET_CLAIM_INVITER

	data := map[string]interface{}{
		"inviter_address": inviterAddress,
	}

	contractOutput, err := c.GetState(mgmAddress, method, data)
	if err != nil {
		return types.ContractOutput{}, fmt.Errorf("failed to get state: %w", err)
	}

	return contractOutput, nil
}

func (c *networkClient) GetClaimInvited(mgmAddress string, invitedAddress string) (types.ContractOutput, error) {
	if mgmAddress == "" {
		return types.ContractOutput{}, fmt.Errorf("address not set")
	}
	if err := keys.ValidateEDDSAPublicKey(mgmAddress); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid address: %w", err)
	}

	if invitedAddress == "" {
		return types.ContractOutput{}, fmt.Errorf("invited address not set")
	}
	if err := keys.ValidateEDDSAPublicKey(invitedAddress); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid invited address: %w", err)
	}

	from := c.publicKey
	if from == "" {
		return types.ContractOutput{}, fmt.Errorf("from address not set")
	}
	if err := keys.ValidateEDDSAPublicKey(from); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid from address: %w", err)
	}

	method := memberGetMemberV1.METHOD_GET_CLAIM_INVITED

	data := map[string]interface{}{
		"invited_address": invitedAddress,
	}

	contractOutput, err := c.GetState(mgmAddress, method, data)
	if err != nil {
		return types.ContractOutput{}, fmt.Errorf("failed to get state: %w", err)
	}

	return contractOutput, nil
}