package client_2finance

import (
	"gitlab.com/2finance/2finance-network/blockchain/keys"
	"gitlab.com/2finance/2finance-network/blockchain/contract/tokenV1/domain"
	"gitlab.com/2finance/2finance-network/blockchain/contract/tokenV1"
	"gitlab.com/2finance/2finance-network/blockchain/contract"
	"gitlab.com/2finance/2finance-network/blockchain/types"
	"gitlab.com/2finance/2finance-network/blockchain/utils"
	"fmt"
	"gitlab.com/2finance/2finance-network/blockchain/transaction"
	"time"
	"encoding/json"
)

func (c *networkClient) AddToken(symbol string, 
		name string, 
		decimals int, 
		totalSupply string, 
		description string, 
		owner string,
		image string, 
		website string, 
		tagsSocialMedia map[string]string,
		tagsCategory map[string]string,
		tags map[string]string,
		creator string, 
		creatorWebsite string, 
		allowUsers map[string]bool,
		blockUsers map[string]bool,
		feeTiersList []map[string]interface{},
		feeAddress string,
		freezeAuthorityRevoked bool, 
		mintAuthorityRevoked bool, 
		updateAuthorityRevoked bool, 
		paused bool,
		expired_at time.Time) (types.ContractOutput, error) {

			
	if symbol == "" {
		return types.ContractOutput{}, fmt.Errorf("symbol not set")
	}
	if name == "" {
		return types.ContractOutput{}, fmt.Errorf("name not set")
	}
	if totalSupply == "" {
		return types.ContractOutput{}, fmt.Errorf("total supply not set")
	}
	if owner == "" {
		return types.ContractOutput{}, fmt.Errorf("owner not set")
	}
	if creator == "" {
		return types.ContractOutput{}, fmt.Errorf("creator not set")
	}
	if creatorWebsite == "" {
		return types.ContractOutput{}, fmt.Errorf("creator website not set")
	}
	if image == "" {
		return types.ContractOutput{}, fmt.Errorf("image not set")
	}
	if website == "" {
		return types.ContractOutput{}, fmt.Errorf("website not set")
	}
	if feeAddress == "" {
		return types.ContractOutput{}, fmt.Errorf("fee address not set")
	}
	if err := keys.ValidateEDDSAPublicKey(feeAddress); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid fee address: %w", err)
	}
	
	err := domain.ValidateUserMap(allowUsers, "allow users")
	if err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid allow users: %w", err)
	}

	err = domain.ValidateUserMap(blockUsers, "block users")
	if err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid block users: %w", err)
	}

	from := c.publicKey
	if from == "" {
		return types.ContractOutput{}, fmt.Errorf("from address not set")
	}
	if err := keys.ValidateEDDSAPublicKey(from); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid from address: %w", err)
	}

	to := types.DEPLOY_CONTRACT_ADDRESS
	contractVersion := tokenV1.TOKEN_CONTRACT_V1
	method := tokenV1.METHOD_ADD_TOKEN
	data := map[string]interface{}{
		"symbol":                symbol,
		"name":                  name,
		"decimals":              decimals,
		"total_supply":           totalSupply,
		"description":           description,
		"owner":                 owner,
		"fee_tiers_list":          feeTiersList,
		"fee_address":            feeAddress, // Fee address is the from address
		"image":                 image,
		"website":               website,
		"tags_social_media":       tagsSocialMedia,
		"tags_category":          tagsCategory,
		"tags":                  tags,
		"creator":               creator,
		"creator_website":        creatorWebsite,
		"allow_users":        allowUsers,
		"block_users":        blockUsers,
		"freeze_authority_revoked": freezeAuthorityRevoked,
		"mint_authority_revoked":   mintAuthorityRevoked,
		"update_authority_revoked": updateAuthorityRevoked,
		"paused":                paused,
		"expired_at":           expired_at,
	}

	timestamp := time.Now().UTC()
	
	nonce, err := c.GetNonce(from)
	if err != nil {
		return types.ContractOutput{}, fmt.Errorf("failed to get nonce: %w", err)
	}
	nonce += 1
	
	newTx := transaction.NewTransaction(from, to, timestamp, contractVersion, method, data, nonce)
	tx := newTx.Get()
	// Sign the transaction
	txSigned, err := transaction.SignTransactionHexKey(c.privateKey, tx)
	if err != nil {
		return types.ContractOutput{}, fmt.Errorf("failed to sign transaction: %w", err)
	}

	contractOutputBytes, err := c.HandlerRequest(contract.REQUEST_METHOD_SEND, txSigned, c.replyTo)
	if err != nil {
		return types.ContractOutput{}, fmt.Errorf("failed to send transaction: %w", err)
	}

	var contractOutput types.ContractOutput
	if err := json.Unmarshal(contractOutputBytes, &contractOutput); err != nil {
		return types.ContractOutput{},fmt.Errorf("failed to unmarshal contract output: %w", err)
	}

	return contractOutput, nil
}
//from signer
//to is the token address, we are sending transaction to the token contract
//mintTo is the address that will receive the minted tokens
//amount is the amount of tokens to mint, it should be in the smallest unit (e.g. wei for ETH)
func (c *networkClient) MintToken(to, mintTo, amount string, decimals int) (types.ContractOutput, error) {
	
	from := c.publicKey
	if from == "" {
		return types.ContractOutput{}, fmt.Errorf("from address not set")
	}
	if to == "" {
		return types.ContractOutput{}, fmt.Errorf("token address not set")
	}
	if mintTo == "" {
		return types.ContractOutput{}, fmt.Errorf("mint to address not set")
	}
	if amount == "" {
		return types.ContractOutput{}, fmt.Errorf("amount not set")
	}

	if err := keys.ValidateEDDSAPublicKey(mintTo); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid mint to address: %w", err)
	}

	if decimals != 0 {
		amountConverted, err := utils.RescaleDecimalString(amount, 0, decimals)
		if err != nil {
			return types.ContractOutput{}, fmt.Errorf("failed to convert amount to big int: %w", err)
		}
		amount = amountConverted
	}

	if err := keys.ValidateEDDSAPublicKey(from); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid from address: %w", err)
	}

	if err := keys.ValidateEDDSAPublicKey(to); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid token address: %w", err)
	}

	if err := keys.ValidateEDDSAPublicKey(mintTo); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid mint to address: %w", err)
	}

	contractVersion := tokenV1.TOKEN_CONTRACT_V1
	method := tokenV1.METHOD_MINT_TOKEN
	data := map[string]interface{}{
		"token_address": to,
		"mint_to":       mintTo,
		"amount":        amount,
	}

	contractOutput, err := c.SendTransaction(
		from,
		to,
		contractVersion,
		method,
		data)
	if err != nil {
		return types.ContractOutput{}, fmt.Errorf("failed to send transaction: %w", err)
	}

	return contractOutput, nil
}
		

func (c *networkClient) BurnToken(to, amount string, decimals int) (types.ContractOutput, error) {
	from := c.publicKey
	if from == "" {
		return types.ContractOutput{}, fmt.Errorf("from address not set")
	}
	if to == "" {
		return types.ContractOutput{}, fmt.Errorf("token address not set")
	}
	if amount == "" {
		return types.ContractOutput{}, fmt.Errorf("amount not set")
	}

	if decimals != 0 {
		amountConverted, err := utils.RescaleDecimalString(amount, 0, decimals)
		if err != nil {
			return types.ContractOutput{}, fmt.Errorf("failed to convert amount to big int: %w", err)
		}
		amount = amountConverted
	}

	if err := keys.ValidateEDDSAPublicKey(from); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid from address: %w", err)
	}

	if err := keys.ValidateEDDSAPublicKey(to); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid token address: %w", err)
	}

	contractVersion := tokenV1.TOKEN_CONTRACT_V1
	method := tokenV1.METHOD_BURN_TOKEN
	data := map[string]interface{}{
		"token_address": to,
		"amount":        amount,
	}

	contractOutput, err := c.SendTransaction(
		from,
		to,
		contractVersion,
		method,
		data)
	if err != nil {
		return types.ContractOutput{}, fmt.Errorf("failed to send transaction: %w", err)
	}

	return contractOutput, nil
}

func (c *networkClient) TransferToken(tokenAddress string, transferTo string, amount string, decimals int) (types.ContractOutput, error) {
	from := c.publicKey
	if from == "" {
		return types.ContractOutput{}, fmt.Errorf("from address not set")
	}
	if transferTo == "" {
		return types.ContractOutput{}, fmt.Errorf("to address not set")
	}
	if tokenAddress == "" {
		return types.ContractOutput{}, fmt.Errorf("token address not set")
	}
	if amount == "" {
		return types.ContractOutput{}, fmt.Errorf("amount not set")
	}
	if from == transferTo {
		return types.ContractOutput{}, fmt.Errorf("from and to addresses are the same")
	}

	if decimals != 0 {
		amountConverted, err := utils.RescaleDecimalString(amount, 0, decimals)
		if err != nil {
			return types.ContractOutput{}, fmt.Errorf("failed to convert amount to big int: %w", err)
		}
		amount = amountConverted
	}

	if err := keys.ValidateEDDSAPublicKey(from); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid from address: %w", err)
	}

	if err := keys.ValidateEDDSAPublicKey(transferTo); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid to address: %w", err)
	}

	if err := keys.ValidateEDDSAPublicKey(tokenAddress); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid token address: %w", err)
	}
	
	contractVersion := tokenV1.TOKEN_CONTRACT_V1
	method := tokenV1.METHOD_TRANSFER_TOKEN
	data := map[string]interface{}{
		"token_address": tokenAddress,
		"transfer_to":            transferTo,
		"amount":        amount,
	}

	contractOutput, err := c.SendTransaction(
		from,
		tokenAddress,
		contractVersion,
		method,
		data)
	if err != nil {
		return types.ContractOutput{}, fmt.Errorf("failed to send transaction: %w", err)
	}

	return contractOutput, nil
}

func (c *networkClient) ApproveSpender(tokenAddress, ownerAddress, spenderAddress, amount string, expiredAt time.Time) (types.ContractOutput, error) {
	from := c.publicKey
	if from == "" {
		return types.ContractOutput{}, fmt.Errorf("from address not set")
	}
	if tokenAddress == "" {
		return types.ContractOutput{}, fmt.Errorf("token address not set")
	}
	if ownerAddress == "" {
		return types.ContractOutput{}, fmt.Errorf("owner address not set")
	}
	if spenderAddress == "" {
		return types.ContractOutput{}, fmt.Errorf("spender address not set")
	}
	if amount == "" {
		return types.ContractOutput{}, fmt.Errorf("amount not set")
	}

	if err := keys.ValidateEDDSAPublicKey(from); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid from address: %w", err)
	}

	if err := keys.ValidateEDDSAPublicKey(tokenAddress); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid token address: %w", err)
	}

	if err := keys.ValidateEDDSAPublicKey(ownerAddress); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid owner address: %w", err)
	}

	if err := keys.ValidateEDDSAPublicKey(spenderAddress); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid spender address: %w", err)
	}

	contractVersion := tokenV1.TOKEN_CONTRACT_V1
	method := tokenV1.METHOD_APPROVE_SPENDER
	data := map[string]interface{}{
		"token_address":  tokenAddress,
		"owner_address":  ownerAddress,
		"spender_address": spenderAddress,
		"amount":         amount,
		"expired_at":     expiredAt,
	}

	contractOutput, err := c.SendTransaction(
		from,
		tokenAddress,
		contractVersion,
		method,
		data)
	if err != nil {
		return types.ContractOutput{}, fmt.Errorf("failed to send transaction: %w", err)
	}

	return contractOutput, nil
}

func (c *networkClient) TransferFromApproved(allowanceAddress, toAddress string) (types.ContractOutput, error) {
	from := c.publicKey
	if from == "" {
		return types.ContractOutput{}, fmt.Errorf("from address not set")
	}
	if allowanceAddress == "" {
		return types.ContractOutput{}, fmt.Errorf("allowance address not set")
	}
	if toAddress == "" {
		return types.ContractOutput{}, fmt.Errorf("to address not set")
	}

	if err := keys.ValidateEDDSAPublicKey(from); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid from address: %w", err)
	}
	if err := keys.ValidateEDDSAPublicKey(allowanceAddress); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid allowance address: %w", err)
	}
	if err := keys.ValidateEDDSAPublicKey(toAddress); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid to address: %w", err)
	}

	contractVersion := tokenV1.TOKEN_CONTRACT_V1
	method := tokenV1.METHOD_TRANSFER_FROM_APPROVED
	data := map[string]interface{}{
		"allowance_address": allowanceAddress,
		"to_address":        toAddress,
	}
	contractOutput, err := c.SendTransaction(
		from,
		allowanceAddress,
		contractVersion,
		method,
		data)
	if err != nil {
		return types.ContractOutput{}, fmt.Errorf("failed to send transaction: %w", err)
	}
	return contractOutput, nil
}


func (c *networkClient) AllowUsers(tokenAddress string, users map[string]bool) (types.ContractOutput, error) {
	from := c.publicKey
	if from == "" {
		return types.ContractOutput{}, fmt.Errorf("from address not set")
	}
	if tokenAddress == "" {
		return types.ContractOutput{}, fmt.Errorf("token address not set")
	}
	if len(users) == 0 {
		return types.ContractOutput{}, fmt.Errorf("users map is empty")
	}

	if err := keys.ValidateEDDSAPublicKey(from); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid from address: %w", err)
	}

	if err := keys.ValidateEDDSAPublicKey(tokenAddress); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid token address: %w", err)
	}

	err := domain.ValidateUserMap(users, "allow users")
	if err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid allow users: %w", err)
	}

	contractVersion := tokenV1.TOKEN_CONTRACT_V1
	method := tokenV1.METHOD_ALLOW_USERS
	data := map[string]interface{}{
		"address": tokenAddress,
		"allow_users": users,
	}

	contractOutput, err := c.SendTransaction(
		from,
		tokenAddress,
		contractVersion,
		method,
		data)
	if err != nil {
		return types.ContractOutput{}, fmt.Errorf("failed to send transaction: %w", err)
	}

	return contractOutput, nil

}


func (c *networkClient) DisallowUsers(tokenAddress string, users map[string]bool) (types.ContractOutput, error) {
	from := c.publicKey
	if from == "" {
		return types.ContractOutput{}, fmt.Errorf("from address not set")
	}
	if tokenAddress == "" {
		return types.ContractOutput{}, fmt.Errorf("token address not set")
	}
	if len(users) == 0 {
		return types.ContractOutput{}, fmt.Errorf("users map is empty")
	}

	if err := keys.ValidateEDDSAPublicKey(from); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid from address: %w", err)
	}

	if err := keys.ValidateEDDSAPublicKey(tokenAddress); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid token address: %w", err)
	}

	err := domain.ValidateUserMap(users, "disallow users")
	if err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid disallow users: %w", err)
	}

	contractVersion := tokenV1.TOKEN_CONTRACT_V1
	method := tokenV1.METHOD_DISALLOW_USERS
	data := map[string]interface{}{
		"address": tokenAddress,
		"allow_users": users,
	}

	contractOutput, err := c.SendTransaction(
		from,
		tokenAddress,
		contractVersion,
		method,
		data)
	if err != nil {
		return types.ContractOutput{}, fmt.Errorf("failed to send transaction: %w", err)
	}

	return contractOutput, nil

}

func (c *networkClient) BlockUsers(tokenAddress string, users map[string]bool) (types.ContractOutput, error) {
	from := c.publicKey
	if from == "" {
		return types.ContractOutput{}, fmt.Errorf("from address not set")
	}
	if tokenAddress == "" {
		return types.ContractOutput{}, fmt.Errorf("token address not set")
	}
	if len(users) == 0 {
		return types.ContractOutput{}, fmt.Errorf("users map is empty")
	}

	if err := keys.ValidateEDDSAPublicKey(from); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid from address: %w", err)
	}

	if err := keys.ValidateEDDSAPublicKey(tokenAddress); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid token address: %w", err)
	}

	err := domain.ValidateUserMap(users, "block users")
	if err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid block users: %w", err)
	}

	contractVersion := tokenV1.TOKEN_CONTRACT_V1
	method := tokenV1.METHOD_BLOCK_USERS
	data := map[string]interface{}{
		"address": tokenAddress,
		"block_users": users,
	}

	contractOutput, err := c.SendTransaction(
		from,
		tokenAddress,
		contractVersion,
		method,
		data)
	if err != nil {
		return types.ContractOutput{}, fmt.Errorf("failed to send transaction: %w", err)
	}

	return contractOutput, nil

}

func (c *networkClient) UnblockUsers(tokenAddress string, users map[string]bool) (types.ContractOutput, error) {
	from := c.publicKey
	if from == "" {
		return types.ContractOutput{}, fmt.Errorf("from address not set")
	}
	if tokenAddress == "" {
		return types.ContractOutput{}, fmt.Errorf("token address not set")
	}
	if len(users) == 0 {
		return types.ContractOutput{}, fmt.Errorf("users map is empty")
	}

	if err := keys.ValidateEDDSAPublicKey(from); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid from address: %w", err)
	}

	if err := keys.ValidateEDDSAPublicKey(tokenAddress); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid token address: %w", err)
	}

	err := domain.ValidateUserMap(users, "unblock users")
	if err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid unblock users: %w", err)
	}

	contractVersion := tokenV1.TOKEN_CONTRACT_V1
	method := tokenV1.METHOD_UNBLOCK_USERS
	data := map[string]interface{}{
		"address": tokenAddress,
		"block_users": users,
	}

	contractOutput, err := c.SendTransaction(
		from,
		tokenAddress,
		contractVersion,
		method,
		data)
	if err != nil {
		return types.ContractOutput{}, fmt.Errorf("failed to send transaction: %w", err)
	}

	return contractOutput, nil

}

func (c *networkClient) RevokeFreezeAuthority(tokenAddress string, revoke bool) (types.ContractOutput, error) {
	from := c.publicKey
	if from == "" {
		return types.ContractOutput{}, fmt.Errorf("from address not set")
	}
	if tokenAddress == "" {
		return types.ContractOutput{}, fmt.Errorf("token address not set")
	}

	if err := keys.ValidateEDDSAPublicKey(from); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid from address: %w", err)
	}

	if err := keys.ValidateEDDSAPublicKey(tokenAddress); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid token address: %w", err)
	}

	contractVersion := tokenV1.TOKEN_CONTRACT_V1
	method := tokenV1.METHOD_REVOKE_FREEZE_AUTHORITY
	data := map[string]interface{}{
		"address": tokenAddress,
		"freeze_authority_revoked":  revoke,
	}

	contractOutput, err := c.SendTransaction(
		from,
		tokenAddress,
		contractVersion,
		method,
		data)
	if err != nil {
		return types.ContractOutput{}, fmt.Errorf("failed to send transaction: %w", err)
	}

	return contractOutput, nil
}

func (c *networkClient) RevokeMintAuthority(tokenAddress string, revoke bool) (types.ContractOutput, error) {
	from := c.publicKey
	if from == "" {
		return types.ContractOutput{}, fmt.Errorf("from address not set")
	}
	if tokenAddress == "" {
		return types.ContractOutput{}, fmt.Errorf("token address not set")
	}

	if err := keys.ValidateEDDSAPublicKey(from); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid from address: %w", err)
	}

	if err := keys.ValidateEDDSAPublicKey(tokenAddress); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid token address: %w", err)
	}

	contractVersion := tokenV1.TOKEN_CONTRACT_V1
	method := tokenV1.METHOD_REVOKE_MINT_AUTHORITY
	data := map[string]interface{}{
		"address": tokenAddress,
		"mint_authority_revoked":  revoke,
	}

	contractOutput, err := c.SendTransaction(
		from,
		tokenAddress,
		contractVersion,
		method,
		data)
	if err != nil {
		return types.ContractOutput{}, fmt.Errorf("failed to send transaction: %w", err)
	}

	return contractOutput, nil
}

func (c *networkClient) RevokeUpdateAuthority(tokenAddress string, revoke bool) (types.ContractOutput, error) {
	from := c.publicKey
	if from == "" {
		return types.ContractOutput{}, fmt.Errorf("from address not set")
	}
	if tokenAddress == "" {
		return types.ContractOutput{}, fmt.Errorf("token address not set")
	}

	if err := keys.ValidateEDDSAPublicKey(from); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid from address: %w", err)
	}

	if err := keys.ValidateEDDSAPublicKey(tokenAddress); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid token address: %w", err)
	}

	contractVersion := tokenV1.TOKEN_CONTRACT_V1
	method := tokenV1.METHOD_REVOKE_UPDATE_AUTHORITY
	data := map[string]interface{}{
		"address": tokenAddress,
		"update_authority_revoked":  revoke,
	}

	contractOutput, err := c.SendTransaction(
		from,
		tokenAddress,
		contractVersion,
		method,
		data)
	if err != nil {
		return types.ContractOutput{}, fmt.Errorf("failed to send transaction: %w", err)
	}

	return contractOutput, nil
}

func (c *networkClient) UpdateMetadata(tokenAddress, symbol, name string, decimals int, description, image, website string,
		tagsSocialMedia, tagsCategory, tags map[string]string,
		creator, creatorWebsite string, expired_at time.Time) (types.ContractOutput, error) {
	from := c.publicKey
	if from == "" {
		return types.ContractOutput{}, fmt.Errorf("from address not set")
	}
	if tokenAddress == "" {
		return types.ContractOutput{}, fmt.Errorf("token address not set")
	}
	if symbol == "" {
		return types.ContractOutput{}, fmt.Errorf("symbol not set")
	}
	if name == "" {
		return types.ContractOutput{}, fmt.Errorf("name not set")
	}
	if description == "" {
		return types.ContractOutput{}, fmt.Errorf("description not set")
	}
	if image == "" {
		return types.ContractOutput{}, fmt.Errorf("image not set")
	}
	if website == "" {
		return types.ContractOutput{}, fmt.Errorf("website not set")
	}
	if creator == "" {
		return types.ContractOutput{}, fmt.Errorf("creator not set")
	}
	if creatorWebsite == "" {
		return types.ContractOutput{}, fmt.Errorf("creator website not set")
	}
	
	if err := keys.ValidateEDDSAPublicKey(from); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid from address: %w", err)
	}
	if err := keys.ValidateEDDSAPublicKey(tokenAddress); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid token address: %w", err)
	}
	contractVersion := tokenV1.TOKEN_CONTRACT_V1
	method := tokenV1.METHOD_UPDATE_METADATA
	data := map[string]interface{}{
		"address":               tokenAddress,
		"symbol":                symbol,
		"name":                  name,
		"decimals":              decimals,
		"description":           description,
		"image":                 image,
		"website":               website,
		"tags_social_media":       tagsSocialMedia,
		"tags_category":          tagsCategory,
		"tags":                  tags,
		"creator":               creator,
		"creator_website":        creatorWebsite,
		"expired_at":           expired_at,
	}
	timestamp := time.Now().UTC()
	nonce, err := c.GetNonce(from)
	if err != nil {
		return types.ContractOutput{}, fmt.Errorf("failed to get nonce: %w", err)
	}
	nonce += 1
	newTx := transaction.NewTransaction(from, tokenAddress, timestamp, contractVersion, method, data, nonce)
	tx := newTx.Get()
	// Sign the transaction
	txSigned, err := transaction.SignTransactionHexKey(c.privateKey, tx)
	if err != nil {
		return types.ContractOutput{}, fmt.Errorf("failed to sign transaction: %w", err)
	}

	contractOutputBytes, err := c.HandlerRequest(contract.REQUEST_METHOD_SEND, txSigned, c.replyTo)
	if err != nil {
		return types.ContractOutput{}, fmt.Errorf("failed to send transaction: %w", err)
	}
	var contractOutput types.ContractOutput
	if err := json.Unmarshal(contractOutputBytes, &contractOutput); err != nil {
		return types.ContractOutput{}, fmt.Errorf("failed to unmarshal contract output: %w", err)
	}
	return contractOutput, nil
}

func (c *networkClient) PauseToken(tokenAddress string, paused bool) (types.ContractOutput, error) {
	from := c.publicKey
	if from == "" {
		return types.ContractOutput{}, fmt.Errorf("from address not set")
	}
	if tokenAddress == "" {
		return types.ContractOutput{}, fmt.Errorf("token address not set")
	}

	if paused != true {
		return types.ContractOutput{}, fmt.Errorf("paused must be true to pause token")
	}

	if err := keys.ValidateEDDSAPublicKey(from); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid from address: %w", err)
	}

	if err := keys.ValidateEDDSAPublicKey(tokenAddress); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid token address: %w", err)
	}

	contractVersion := tokenV1.TOKEN_CONTRACT_V1
	method := tokenV1.METHOD_PAUSE_TOKEN
	data := map[string]interface{}{
		"address": tokenAddress,
		"paused":  paused,
	}

	contractOutput, err := c.SendTransaction(
		from,
		tokenAddress,
		contractVersion,
		method,
		data)
	if err != nil {
		return types.ContractOutput{}, fmt.Errorf("failed to send transaction: %w", err)
	}

	return contractOutput, nil
}
func (c *networkClient) UnpauseToken(tokenAddress string, paused bool) (types.ContractOutput, error) {
	from := c.publicKey
	if from == "" {
		return types.ContractOutput{}, fmt.Errorf("from address not set")
	}
	if tokenAddress == "" {
		return types.ContractOutput{}, fmt.Errorf("token address not set")
	}

	if paused != false {
		return types.ContractOutput{}, fmt.Errorf("paused must be false to unpause token")
	}

	if err := keys.ValidateEDDSAPublicKey(from); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid from address: %w", err)
	}

	if err := keys.ValidateEDDSAPublicKey(tokenAddress); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid token address: %w", err)
	}

	contractVersion := tokenV1.TOKEN_CONTRACT_V1
	method := tokenV1.METHOD_UNPAUSE_TOKEN
	data := map[string]interface{}{
		"address": tokenAddress,
		"paused":  paused,
	}

	contractOutput, err := c.SendTransaction(
		from,
		tokenAddress,
		contractVersion,
		method,
		data)
	if err != nil {
		return types.ContractOutput{}, fmt.Errorf("failed to send transaction: %w", err)
	}

	return contractOutput, nil
}

func (c *networkClient) UpdateFeeTiers(tokenAddress string, feeTiersList []map[string]interface{}) (types.ContractOutput, error) {
	from := c.publicKey
	if from == "" {
		return types.ContractOutput{}, fmt.Errorf("from address not set")
	}
	if tokenAddress == "" {
		return types.ContractOutput{}, fmt.Errorf("token address not set")
	}
	if len(feeTiersList) == 0 {
		return types.ContractOutput{}, fmt.Errorf("fee tiers list is empty")
	}

	if err := keys.ValidateEDDSAPublicKey(from); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid from address: %w", err)
	}

	if err := keys.ValidateEDDSAPublicKey(tokenAddress); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid token address: %w", err)
	}

	contractVersion := tokenV1.TOKEN_CONTRACT_V1
	method := tokenV1.METHOD_UPDATE_FEE_TIERS
	data := map[string]interface{}{
		"token_address": tokenAddress,
		"fee_tiers_list": feeTiersList,
	}

	contractOutput, err := c.SendTransaction(
		from,
		tokenAddress,
		contractVersion,
		method,
		data)
	if err != nil {
		return types.ContractOutput{}, fmt.Errorf("failed to send transaction: %w", err)
	}

	return contractOutput, nil
}

func (c *networkClient) UpdateFeeAddress(tokenAddress, feeAddress string) (types.ContractOutput, error) {
	from := c.publicKey
	if from == "" {
		return types.ContractOutput{}, fmt.Errorf("from address not set")
	}
	if tokenAddress == "" {
		return types.ContractOutput{}, fmt.Errorf("token address not set")
	}
	if feeAddress == "" {
		return types.ContractOutput{}, fmt.Errorf("fee address not set")
	}

	if err := keys.ValidateEDDSAPublicKey(from); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid from address: %w", err)
	}

	if err := keys.ValidateEDDSAPublicKey(tokenAddress); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid token address: %w", err)
	}

	contractVersion := tokenV1.TOKEN_CONTRACT_V1
	method := tokenV1.METHOD_UPDATE_FEE_ADDRESS
	data := map[string]interface{}{
		"token_address": tokenAddress,
		"fee_address":   feeAddress,
	}

	contractOutput, err := c.SendTransaction(
		from,
		tokenAddress,
		contractVersion,
		method,
		data)
	if err != nil {
		return types.ContractOutput{}, fmt.Errorf("failed to send transaction: %w", err)
	}

	return contractOutput, nil
}

func (c *networkClient) GetToken(tokenAddress string, symbol string, name string) (types.ContractOutput, error) {
	from := c.publicKey
	if from == "" {
		return types.ContractOutput{}, fmt.Errorf("from address not set")
	}
	
	if tokenAddress == "" && symbol == "" && name == "" {
		return types.ContractOutput{}, fmt.Errorf("token address, symbol or name must be set")
	}

	if err := keys.ValidateEDDSAPublicKey(from); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid from address: %w", err)
	}

	if err := keys.ValidateEDDSAPublicKey(tokenAddress); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid token address: %w", err)
	}

	contractVersion := tokenV1.TOKEN_CONTRACT_V1
	method := tokenV1.METHOD_GET_TOKEN
	data := map[string]interface{}{
		"address": tokenAddress,
		"symbol":  symbol,
		"name":    name,
	}

	contractOutput, err := c.GetState(contractVersion, method, data)
	if err != nil {
		return types.ContractOutput{}, fmt.Errorf("failed to get state: %w", err)
	}

	return contractOutput, nil
}

func (c *networkClient) ListTokens(ownerAddress, symbol, name string, page, limit int, ascending bool) (types.ContractOutput, error) {
	from := c.publicKey
	if from == "" {
		return types.ContractOutput{}, fmt.Errorf("from address not set")
	}

	if ownerAddress != "" {
		if err := keys.ValidateEDDSAPublicKey(ownerAddress); err != nil {
			return types.ContractOutput{}, fmt.Errorf("invalid owner address: %w", err)
		}
	}

	if err := keys.ValidateEDDSAPublicKey(from); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid from address: %w", err)
	}

	contractVersion := tokenV1.TOKEN_CONTRACT_V1
	method := tokenV1.METHOD_LIST_TOKENS
	data := map[string]interface{}{
		"owner":   ownerAddress,
		"symbol":  symbol,
		"name":    name,
		"page":    page,
		"limit":   limit,
		"ascending": ascending,
	}

	contractOutput, err := c.GetState(contractVersion, method, data)
	if err != nil {
		return types.ContractOutput{}, fmt.Errorf("failed to get state: %w", err)
	}

	return contractOutput, nil
}

func (c *networkClient) GetTokenBalance(tokenAddress, ownerAddress string) (types.ContractOutput, error) {
	from := c.publicKey
	if from == "" {
		return types.ContractOutput{}, fmt.Errorf("from address not set")
	}
	if tokenAddress == "" {
		return types.ContractOutput{}, fmt.Errorf("token address not set")
	}
	if ownerAddress == "" {
		return types.ContractOutput{}, fmt.Errorf("owner address not set")
	}

	if err := keys.ValidateEDDSAPublicKey(from); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid from address: %w", err)
	}

	if err := keys.ValidateEDDSAPublicKey(tokenAddress); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid token address: %w", err)
	}

	if err := keys.ValidateEDDSAPublicKey(ownerAddress); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid owner address: %w", err)
	}

	contractVersion := tokenV1.TOKEN_CONTRACT_V1
	method := tokenV1.METHOD_GET_TOKEN_BALANCE
	data := map[string]interface{}{
		"token_address": tokenAddress,
		"owner_address": ownerAddress,
	}

	contractOutput, err := c.GetState(contractVersion, method, data)
	if err != nil {
		return types.ContractOutput{}, fmt.Errorf("failed to get state: %w", err)
	}

	return contractOutput, nil
}

func (c *networkClient) ListTokenBalances(tokenAddress, ownerAddress string, page, limit int, ascending bool) (types.ContractOutput, error) {
	from := c.publicKey
	if from == "" {
		return types.ContractOutput{}, fmt.Errorf("from address not set")
	}

	if err := keys.ValidateEDDSAPublicKey(from); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid from address: %w", err)
	}

	if tokenAddress != "" {

		if err := keys.ValidateEDDSAPublicKey(tokenAddress); err != nil {
			return types.ContractOutput{}, fmt.Errorf("invalid token address: %w", err)
		}
	}
	if ownerAddress != "" {
		if err := keys.ValidateEDDSAPublicKey(ownerAddress); err != nil {
			return types.ContractOutput{}, fmt.Errorf("invalid owner address: %w", err)
		}
	}

	contractVersion := tokenV1.TOKEN_CONTRACT_V1
	method := tokenV1.METHOD_LIST_TOKEN_BALANCES
	data := map[string]interface{}{
		"token_address": tokenAddress,
		"owner_address": ownerAddress,
		"page":          page,
		"limit":         limit,
		"ascending":     ascending,
	}

	contractOutput, err := c.GetState(contractVersion, method, data)
	if err != nil {
		return types.ContractOutput{}, fmt.Errorf("failed to get state: %w", err)
	}

	return contractOutput, nil
}
