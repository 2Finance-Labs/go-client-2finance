package client_2finance

import (
	"fmt"
	"time"

	"gitlab.com/2finance/2finance-network/blockchain/contract/tokenV1"
	"gitlab.com/2finance/2finance-network/blockchain/contract/tokenV1/domain"
	"gitlab.com/2finance/2finance-network/blockchain/encryption/keys"
	"gitlab.com/2finance/2finance-network/blockchain/types"
	"gitlab.com/2finance/2finance-network/blockchain/utils"
)

func (c *networkClient) AddToken(
	address string,
	symbol string,
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
	accessPolicy domain.AccessPolicy,
	feeTiersList []map[string]interface{},
	feeAddress string,
	freezeAuthorityRevoked bool,
	mintAuthorityRevoked bool,
	updateAuthorityRevoked bool,
	paused bool,
	expired_at time.Time,
	assetGLBUri string,
	tokenType string,
	transferable bool) (types.ContractOutput, error) {

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
	if err := keys.ValidateEDDSAPublicKeyHex(feeAddress); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid fee address: %w", err)
	}
	if assetGLBUri == "" {
		return types.ContractOutput{}, fmt.Errorf("asset GLB URI not set")
	}
	if tokenType == "" {
		return types.ContractOutput{}, fmt.Errorf("token type not set")
	}

	err := domain.ValidateAccessPolicy(accessPolicy.List, "access policy")
	if err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid access policy: %w", err)
	}

	from := c.publicKey

	if err := keys.ValidateEDDSAPublicKeyHex(from); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid from address: %w", err)
	}

	to := address
	method := tokenV1.METHOD_ADD_TOKEN
	data := map[string]interface{}{
		"symbol":                   symbol,
		"name":                     name,
		"decimals":                 decimals,
		"total_supply":             totalSupply,
		"description":              description,
		"owner":                    owner,
		"fee_tiers_list":           feeTiersList,
		"fee_address":              feeAddress, // Fee address is the from address
		"image":                    image,
		"website":                  website,
		"tags_social_media":        tagsSocialMedia,
		"tags_category":            tagsCategory,
		"tags":                     tags,
		"creator":                  creator,
		"creator_website":          creatorWebsite,
		"access_policy_list":       accessPolicy,
		"freeze_authority_revoked": freezeAuthorityRevoked,
		"mint_authority_revoked":   mintAuthorityRevoked,
		"update_authority_revoked": updateAuthorityRevoked,
		"paused":                   paused,
		"expired_at":               expired_at,
		"asset_glb_uri":            assetGLBUri,
		"token_type":               tokenType,
		"transferable":             transferable,
	}

	version := uint8(1)
	uuid7, err := utils.NewUUID7()
	if err != nil {
		return types.ContractOutput{}, fmt.Errorf("failed to generate UUIDv7: %w", err)
	}

	contractOutput, err := c.SignAndSendTransaction(
		c.chainId,
		from,
		to,
		method,
		data,
		version,
		uuid7)
	if err != nil {
		return types.ContractOutput{}, fmt.Errorf("failed to send transaction: %w", err)
	}

	return contractOutput, nil
}

// from signer
// to is the token address, we are sending transaction to the token contract
// mintTo is the address that will receive the minted tokens
// amount is the amount of tokens to mint, it should be in the smallest unit (e.g. wei for ETH)
func (c *networkClient) MintToken(to, mintTo, amount string, decimals int, tokenType string) (types.ContractOutput, error) {

	from := c.publicKey

	if to == "" {
		return types.ContractOutput{}, fmt.Errorf("token address not set")
	}
	if mintTo == "" {
		return types.ContractOutput{}, fmt.Errorf("mint to address not set")
	}
	if amount == "" {
		return types.ContractOutput{}, fmt.Errorf("amount not set")
	}
	if tokenType == "" {
		return types.ContractOutput{}, fmt.Errorf("token type not set")
	}

	if err := keys.ValidateEDDSAPublicKeyHex(mintTo); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid mint to address: %w", err)
	}

	if decimals != 0 {
		amountConverted, err := utils.RescaleDecimalString(amount, 0, decimals)
		if err != nil {
			return types.ContractOutput{}, fmt.Errorf("failed to convert amount to big int: %w", err)
		}
		amount = amountConverted
	}

	if err := keys.ValidateEDDSAPublicKeyHex(from); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid from address: %w", err)
	}

	if err := keys.ValidateEDDSAPublicKeyHex(to); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid token address: %w", err)
	}

	if err := keys.ValidateEDDSAPublicKeyHex(mintTo); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid mint to address: %w", err)
	}

	method := tokenV1.METHOD_MINT_TOKEN
	data := map[string]interface{}{
		"mint_to":    mintTo,
		"amount":     amount,
		"token_type": tokenType,
	}
	version := uint8(1)
	uuid7, err := utils.NewUUID7()
	if err != nil {
		return types.ContractOutput{}, fmt.Errorf("failed to generate UUIDv7: %w", err)
	}

	contractOutput, err := c.SignAndSendTransaction(
		c.chainId,
		from,
		to,
		method,
		data,
		version,
		uuid7)
	if err != nil {
		return types.ContractOutput{}, fmt.Errorf("failed to send transaction: %w", err)
	}

	return contractOutput, nil
}

func (c *networkClient) BurnToken(to, amount string, decimals int, tokenType string, uuid string) (types.ContractOutput, error) {
	from := c.publicKey

	if to == "" {
		return types.ContractOutput{}, fmt.Errorf("token address not set")
	}
	if amount == "" {
		return types.ContractOutput{}, fmt.Errorf("amount not set")
	}
	if tokenType == "" {
		return types.ContractOutput{}, fmt.Errorf("token type not set")
	}
	if tokenType == domain.NON_FUNGIBLE {
		if uuid == "" {
			return types.ContractOutput{}, fmt.Errorf("uuid not set")
		}
	}

	if decimals != 0 {
		amountConverted, err := utils.RescaleDecimalString(amount, 0, decimals)
		if err != nil {
			return types.ContractOutput{}, fmt.Errorf("failed to convert amount to big int: %w", err)
		}
		amount = amountConverted
	}

	if err := keys.ValidateEDDSAPublicKeyHex(from); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid from address: %w", err)
	}

	if err := keys.ValidateEDDSAPublicKeyHex(to); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid token address: %w", err)
	}

	method := tokenV1.METHOD_BURN_TOKEN
	data := map[string]interface{}{
		"amount":     amount,
		"token_type": tokenType,
		"uuid":       uuid,
	}
	version := uint8(1)
	uuid7, err := utils.NewUUID7()
	if err != nil {
		return types.ContractOutput{}, fmt.Errorf("failed to generate UUIDv7: %w", err)
	}

	contractOutput, err := c.SignAndSendTransaction(
		c.chainId,
		from,
		to,
		method,
		data,
		version,
		uuid7)
	if err != nil {
		return types.ContractOutput{}, fmt.Errorf("failed to send transaction: %w", err)
	}

	return contractOutput, nil
}

func (c *networkClient) TransferToken(tokenAddress string, transferTo string, amount string, decimals int, tokenType string, uuid string) (types.ContractOutput, error) {
	from := c.publicKey

	if transferTo == "" {
		return types.ContractOutput{}, fmt.Errorf("to address not set")
	}
	if tokenAddress == "" {
		return types.ContractOutput{}, fmt.Errorf("token address not set")
	}
	if amount == "" {
		return types.ContractOutput{}, fmt.Errorf("amount not set")
	}
	if tokenType == "" {
		return types.ContractOutput{}, fmt.Errorf("token type not set")
	}
	if tokenType == domain.NON_FUNGIBLE {
		if uuid == "" {
			return types.ContractOutput{}, fmt.Errorf("uuid not set")
		}
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

	if err := keys.ValidateEDDSAPublicKeyHex(from); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid from address: %w", err)
	}

	if err := keys.ValidateEDDSAPublicKeyHex(transferTo); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid to address: %w", err)
	}

	if err := keys.ValidateEDDSAPublicKeyHex(tokenAddress); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid token address: %w", err)
	}

	method := tokenV1.METHOD_TRANSFER_TOKEN
	data := map[string]interface{}{
		"transfer_to": transferTo,
		"amount":      amount,
		"token_type":  tokenType,
		"uuid":        uuid,
	}
	version := uint8(1)
	uuid7, err := utils.NewUUID7()
	if err != nil {
		return types.ContractOutput{}, fmt.Errorf("failed to generate UUIDv7: %w", err)
	}
	contractOutput, err := c.SignAndSendTransaction(
		c.chainId,
		from,
		tokenAddress,
		method,
		data,
		version,
		uuid7)
	if err != nil {
		return types.ContractOutput{}, fmt.Errorf("failed to send transaction: %w", err)
	}

	return contractOutput, nil
}

func (c *networkClient) AllowUsers(tokenAddress string, users map[string]bool) (types.ContractOutput, error) {
	from := c.publicKey

	if tokenAddress == "" {
		return types.ContractOutput{}, fmt.Errorf("token address not set")
	}
	if len(users) == 0 {
		return types.ContractOutput{}, fmt.Errorf("users map is empty")
	}

	if err := keys.ValidateEDDSAPublicKeyHex(from); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid from address: %w", err)
	}

	if err := keys.ValidateEDDSAPublicKeyHex(tokenAddress); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid token address: %w", err)
	}

	err := domain.ValidateUserMap(users, "allow users")
	if err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid allow users: %w", err)
	}

	method := tokenV1.METHOD_ALLOW_USERS
	data := map[string]interface{}{
		"access_policy_list": users,
	}
	version := uint8(1)
	uuid7, err := utils.NewUUID7()
	if err != nil {
		return types.ContractOutput{}, fmt.Errorf("failed to generate UUIDv7: %w", err)
	}

	contractOutput, err := c.SignAndSendTransaction(
		c.chainId,
		from,
		tokenAddress,
		method,
		data,
		version,
		uuid7)
	if err != nil {
		return types.ContractOutput{}, fmt.Errorf("failed to send transaction: %w", err)
	}

	return contractOutput, nil

}

func (c *networkClient) DisallowUsers(tokenAddress string, users map[string]bool) (types.ContractOutput, error) {
	from := c.publicKey

	if tokenAddress == "" {
		return types.ContractOutput{}, fmt.Errorf("token address not set")
	}
	if len(users) == 0 {
		return types.ContractOutput{}, fmt.Errorf("users map is empty")
	}

	if err := keys.ValidateEDDSAPublicKeyHex(from); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid from address: %w", err)
	}

	if err := keys.ValidateEDDSAPublicKeyHex(tokenAddress); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid token address: %w", err)
	}

	err := domain.ValidateUserMap(users, "disallow users")
	if err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid disallow users: %w", err)
	}

	method := tokenV1.METHOD_DISALLOW_USERS
	data := map[string]interface{}{
		"access_policy_list": users,
	}
	version := uint8(1)
	uuid7, err := utils.NewUUID7()
	if err != nil {
		return types.ContractOutput{}, fmt.Errorf("failed to generate UUIDv7: %w", err)
	}
	contractOutput, err := c.SignAndSendTransaction(
		c.chainId,
		from,
		tokenAddress,
		method,
		data,
		version,
		uuid7)
	if err != nil {
		return types.ContractOutput{}, fmt.Errorf("failed to send transaction: %w", err)
	}

	return contractOutput, nil

}

func (c *networkClient) BlockUsers(tokenAddress string, users map[string]bool) (types.ContractOutput, error) {
	from := c.publicKey

	if tokenAddress == "" {
		return types.ContractOutput{}, fmt.Errorf("token address not set")
	}
	if len(users) == 0 {
		return types.ContractOutput{}, fmt.Errorf("users map is empty")
	}

	if err := keys.ValidateEDDSAPublicKeyHex(from); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid from address: %w", err)
	}

	if err := keys.ValidateEDDSAPublicKeyHex(tokenAddress); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid token address: %w", err)
	}

	err := domain.ValidateUserMap(users, "block users")
	if err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid block users: %w", err)
	}

	method := tokenV1.METHOD_BLOCK_USERS
	data := map[string]interface{}{
		"access_policy_list": users,
	}
	version := uint8(1)
	uuid7, err := utils.NewUUID7()
	if err != nil {
		return types.ContractOutput{}, fmt.Errorf("failed to generate UUIDv7: %w", err)
	}
	contractOutput, err := c.SignAndSendTransaction(
		c.chainId,
		from,
		tokenAddress,
		method,
		data,
		version,
		uuid7)
	if err != nil {
		return types.ContractOutput{}, fmt.Errorf("failed to send transaction: %w", err)
	}

	return contractOutput, nil

}

func (c *networkClient) UnblockUsers(tokenAddress string, users map[string]bool) (types.ContractOutput, error) {
	from := c.publicKey

	if tokenAddress == "" {
		return types.ContractOutput{}, fmt.Errorf("token address not set")
	}
	if len(users) == 0 {
		return types.ContractOutput{}, fmt.Errorf("users map is empty")
	}

	if err := keys.ValidateEDDSAPublicKeyHex(from); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid from address: %w", err)
	}

	if err := keys.ValidateEDDSAPublicKeyHex(tokenAddress); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid token address: %w", err)
	}

	err := domain.ValidateUserMap(users, "unblock users")
	if err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid unblock users: %w", err)
	}

	method := tokenV1.METHOD_UNBLOCK_USERS
	data := map[string]interface{}{
		"access_policy_list": users,
	}
	version := uint8(1)
	uuid7, err := utils.NewUUID7()
	if err != nil {
		return types.ContractOutput{}, fmt.Errorf("failed to generate UUIDv7: %w", err)
	}
	contractOutput, err := c.SignAndSendTransaction(
		c.chainId,
		from,
		tokenAddress,
		method,
		data,
		version,
		uuid7)
	if err != nil {
		return types.ContractOutput{}, fmt.Errorf("failed to send transaction: %w", err)
	}

	return contractOutput, nil

}

func (c *networkClient) RevokeFreezeAuthority(tokenAddress string, revoke bool) (types.ContractOutput, error) {
	from := c.publicKey

	if tokenAddress == "" {
		return types.ContractOutput{}, fmt.Errorf("token address not set")
	}

	if err := keys.ValidateEDDSAPublicKeyHex(from); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid from address: %w", err)
	}

	if err := keys.ValidateEDDSAPublicKeyHex(tokenAddress); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid token address: %w", err)
	}

	method := tokenV1.METHOD_REVOKE_FREEZE_AUTHORITY
	data := map[string]interface{}{
		"freeze_authority_revoked": revoke,
	}
	version := uint8(1)
	uuid7, err := utils.NewUUID7()
	if err != nil {
		return types.ContractOutput{}, fmt.Errorf("failed to generate UUIDv7: %w", err)
	}
	contractOutput, err := c.SignAndSendTransaction(
		c.chainId,
		from,
		tokenAddress,
		method,
		data,
		version,
		uuid7)
	if err != nil {
		return types.ContractOutput{}, fmt.Errorf("failed to send transaction: %w", err)
	}

	return contractOutput, nil
}

func (c *networkClient) RevokeMintAuthority(tokenAddress string, revoke bool) (types.ContractOutput, error) {
	from := c.publicKey

	if tokenAddress == "" {
		return types.ContractOutput{}, fmt.Errorf("token address not set")
	}

	if err := keys.ValidateEDDSAPublicKeyHex(from); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid from address: %w", err)
	}

	if err := keys.ValidateEDDSAPublicKeyHex(tokenAddress); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid token address: %w", err)
	}

	method := tokenV1.METHOD_REVOKE_MINT_AUTHORITY
	data := map[string]interface{}{
		"mint_authority_revoked": revoke,
	}
	version := uint8(1)
	uuid7, err := utils.NewUUID7()
	if err != nil {
		return types.ContractOutput{}, fmt.Errorf("failed to generate UUIDv7: %w", err)
	}
	contractOutput, err := c.SignAndSendTransaction(
		c.chainId,
		from,
		tokenAddress,
		method,
		data,
		version,
		uuid7)
	if err != nil {
		return types.ContractOutput{}, fmt.Errorf("failed to send transaction: %w", err)
	}

	return contractOutput, nil
}

func (c *networkClient) RevokeUpdateAuthority(tokenAddress string, revoke bool) (types.ContractOutput, error) {
	from := c.publicKey

	if tokenAddress == "" {
		return types.ContractOutput{}, fmt.Errorf("token address not set")
	}

	if err := keys.ValidateEDDSAPublicKeyHex(from); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid from address: %w", err)
	}

	if err := keys.ValidateEDDSAPublicKeyHex(tokenAddress); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid token address: %w", err)
	}

	method := tokenV1.METHOD_REVOKE_UPDATE_AUTHORITY
	data := map[string]interface{}{
		"update_authority_revoked": revoke,
	}
	version := uint8(1)
	uuid7, err := utils.NewUUID7()
	if err != nil {
		return types.ContractOutput{}, fmt.Errorf("failed to generate UUIDv7: %w", err)
	}
	contractOutput, err := c.SignAndSendTransaction(
		c.chainId,
		from,
		tokenAddress,
		method,
		data,
		version,
		uuid7)
	if err != nil {
		return types.ContractOutput{}, fmt.Errorf("failed to send transaction: %w", err)
	}

	return contractOutput, nil
}

func (c *networkClient) UpdateMetadata(tokenAddress, symbol, name string, decimals int, description, image, website string,
	tagsSocialMedia, tagsCategory, tags map[string]string,
	creator, creatorWebsite string, expired_at time.Time) (types.ContractOutput, error) {
	from := c.publicKey

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

	if err := keys.ValidateEDDSAPublicKeyHex(from); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid from address: %w", err)
	}
	if err := keys.ValidateEDDSAPublicKeyHex(tokenAddress); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid token address: %w", err)
	}
	method := tokenV1.METHOD_UPDATE_METADATA
	data := map[string]interface{}{
		"symbol":            symbol,
		"name":              name,
		"decimals":          decimals,
		"description":       description,
		"image":             image,
		"website":           website,
		"tags_social_media": tagsSocialMedia,
		"tags_category":     tagsCategory,
		"tags":              tags,
		"creator":           creator,
		"creator_website":   creatorWebsite,
		"expired_at":        expired_at,
	}
	version := uint8(1)
	uuid7, err := utils.NewUUID7()
	if err != nil {
		return types.ContractOutput{}, fmt.Errorf("failed to generate UUIDv7: %w", err)
	}
	contractOutput, err := c.SignAndSendTransaction(
		c.chainId,
		from,
		tokenAddress,
		method,
		data,
		version,
		uuid7)
	if err != nil {
		return types.ContractOutput{}, fmt.Errorf("failed to send transaction: %w", err)
	}

	return contractOutput, nil
}

func (c *networkClient) PauseToken(tokenAddress string, paused bool) (types.ContractOutput, error) {
	from := c.publicKey

	if tokenAddress == "" {
		return types.ContractOutput{}, fmt.Errorf("token address not set")
	}

	if paused != true {
		return types.ContractOutput{}, fmt.Errorf("paused must be true to pause token")
	}

	if err := keys.ValidateEDDSAPublicKeyHex(from); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid from address: %w", err)
	}

	if err := keys.ValidateEDDSAPublicKeyHex(tokenAddress); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid token address: %w", err)
	}

	method := tokenV1.METHOD_PAUSE_TOKEN
	data := map[string]interface{}{
		"paused": paused,
	}
	version := uint8(1)
	uuid7, err := utils.NewUUID7()
	if err != nil {
		return types.ContractOutput{}, fmt.Errorf("failed to generate UUIDv7: %w", err)
	}
	contractOutput, err := c.SignAndSendTransaction(
		c.chainId,
		from,
		tokenAddress,
		method,
		data,
		version,
		uuid7)
	if err != nil {
		return types.ContractOutput{}, fmt.Errorf("failed to send transaction: %w", err)
	}

	return contractOutput, nil
}
func (c *networkClient) UnpauseToken(tokenAddress string, paused bool) (types.ContractOutput, error) {
	from := c.publicKey

	if tokenAddress == "" {
		return types.ContractOutput{}, fmt.Errorf("token address not set")
	}

	if paused != false {
		return types.ContractOutput{}, fmt.Errorf("paused must be false to unpause token")
	}

	if err := keys.ValidateEDDSAPublicKeyHex(from); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid from address: %w", err)
	}

	if err := keys.ValidateEDDSAPublicKeyHex(tokenAddress); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid token address: %w", err)
	}

	method := tokenV1.METHOD_UNPAUSE_TOKEN
	data := map[string]interface{}{
		"paused": paused,
	}
	version := uint8(1)
	uuid7, err := utils.NewUUID7()
	if err != nil {
		return types.ContractOutput{}, fmt.Errorf("failed to generate UUIDv7: %w", err)
	}
	contractOutput, err := c.SignAndSendTransaction(
		c.chainId,
		from,
		tokenAddress,
		method,
		data,
		version,
		uuid7)
	if err != nil {
		return types.ContractOutput{}, fmt.Errorf("failed to send transaction: %w", err)
	}

	return contractOutput, nil
}

func (c *networkClient) UpdateFeeTiers(tokenAddress string, feeTiersList []map[string]interface{}) (types.ContractOutput, error) {
	from := c.publicKey

	if tokenAddress == "" {
		return types.ContractOutput{}, fmt.Errorf("token address not set")
	}
	if len(feeTiersList) == 0 {
		return types.ContractOutput{}, fmt.Errorf("fee tiers list is empty")
	}

	if err := keys.ValidateEDDSAPublicKeyHex(from); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid from address: %w", err)
	}

	if err := keys.ValidateEDDSAPublicKeyHex(tokenAddress); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid token address: %w", err)
	}

	method := tokenV1.METHOD_UPDATE_FEE_TIERS
	data := map[string]interface{}{
		"fee_tiers_list": feeTiersList,
	}
	version := uint8(1)
	uuid7, err := utils.NewUUID7()
	if err != nil {
		return types.ContractOutput{}, fmt.Errorf("failed to generate UUIDv7: %w", err)
	}
	contractOutput, err := c.SignAndSendTransaction(
		c.chainId,
		from,
		tokenAddress,
		method,
		data,
		version,
		uuid7)
	if err != nil {
		return types.ContractOutput{}, fmt.Errorf("failed to send transaction: %w", err)
	}

	return contractOutput, nil
}

func (c *networkClient) UpdateFeeAddress(tokenAddress, feeAddress string) (types.ContractOutput, error) {
	from := c.publicKey

	if tokenAddress == "" {
		return types.ContractOutput{}, fmt.Errorf("token address not set")
	}
	if feeAddress == "" {
		return types.ContractOutput{}, fmt.Errorf("fee address not set")
	}

	if err := keys.ValidateEDDSAPublicKeyHex(from); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid from address: %w", err)
	}

	if err := keys.ValidateEDDSAPublicKeyHex(tokenAddress); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid token address: %w", err)
	}

	method := tokenV1.METHOD_UPDATE_FEE_ADDRESS
	data := map[string]interface{}{
		"fee_address": feeAddress,
	}
	version := uint8(1)
	uuid7, err := utils.NewUUID7()
	if err != nil {
		return types.ContractOutput{}, fmt.Errorf("failed to generate UUIDv7: %w", err)
	}
	contractOutput, err := c.SignAndSendTransaction(
		c.chainId,
		from,
		tokenAddress,
		method,
		data,
		version,
		uuid7)
	if err != nil {
		return types.ContractOutput{}, fmt.Errorf("failed to send transaction: %w", err)
	}

	return contractOutput, nil
}

func (c *networkClient) UpdateGlbFile(tokenAddress string, newAssetGLBUri string) (types.ContractOutput, error) {
	from := c.publicKey
	if from == "" {
		return types.ContractOutput{}, fmt.Errorf("from address not set")
	}
	if tokenAddress == "" {
		return types.ContractOutput{}, fmt.Errorf("token address not set")
	}
	if newAssetGLBUri == "" {
		return types.ContractOutput{}, fmt.Errorf("new asset GLB URI not set")
	}

	if err := keys.ValidateEDDSAPublicKeyHex(from); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid from address: %w", err)
	}

	if err := keys.ValidateEDDSAPublicKeyHex(tokenAddress); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid token address: %w", err)
	}

	method := tokenV1.METHOD_UPDATE_GLB_FILE
	data := map[string]interface{}{
		"asset_glb_uri": newAssetGLBUri,
	}
	version := uint8(1)
	uuid7, err := utils.NewUUID7()
	if err != nil {
		return types.ContractOutput{}, fmt.Errorf("failed to generate UUIDv7: %w", err)
	}
	contractOutput, err := c.SignAndSendTransaction(
		c.chainId,
		from,
		tokenAddress,
		method,
		data,
		version,
		uuid7)
	if err != nil {
		return types.ContractOutput{}, fmt.Errorf("failed to send transaction: %w", err)
	}

	return contractOutput, nil
}

func (c *networkClient) TransferableToken(tokenAddress string, transferable bool) (types.ContractOutput, error) {
	from := c.publicKey
	if from == "" {
		return types.ContractOutput{}, fmt.Errorf("from address not set")
	}
	if tokenAddress == "" {
		return types.ContractOutput{}, fmt.Errorf("token address not set")
	}

	if err := keys.ValidateEDDSAPublicKeyHex(from); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid from address: %w", err)
	}

	if err := keys.ValidateEDDSAPublicKeyHex(tokenAddress); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid token address: %w", err)
	}

	method := tokenV1.METHOD_TRANSFERABLE_TOKEN
	data := map[string]interface{}{
		"transferable": transferable,
	}
	version := uint8(1)
	uuid7, err := utils.NewUUID7()
	if err != nil {
		return types.ContractOutput{}, fmt.Errorf("failed to generate UUIDv7: %w", err)
	}
	contractOutput, err := c.SignAndSendTransaction(
		c.chainId,
		from,
		tokenAddress,
		method,
		data,
		version,
		uuid7)
	if err != nil {
		return types.ContractOutput{}, fmt.Errorf("failed to send transaction: %w", err)
	}

	return contractOutput, nil
}

func (c *networkClient) UntransferableToken(tokenAddress string, transferable bool) (types.ContractOutput, error) {
	from := c.publicKey
	if from == "" {
		return types.ContractOutput{}, fmt.Errorf("from address not set")
	}
	if tokenAddress == "" {
		return types.ContractOutput{}, fmt.Errorf("token address not set")
	}

	if err := keys.ValidateEDDSAPublicKeyHex(from); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid from address: %w", err)
	}

	if err := keys.ValidateEDDSAPublicKeyHex(tokenAddress); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid token address: %w", err)
	}

	method := tokenV1.METHOD_UNTRANSFERABLE_TOKEN
	data := map[string]interface{}{
		"transferable": transferable,
	}
	version := uint8(1)
	uuid7, err := utils.NewUUID7()
	if err != nil {
		return types.ContractOutput{}, fmt.Errorf("failed to generate UUIDv7: %w", err)
	}

	contractOutput, err := c.SignAndSendTransaction(
		c.chainId,
		from,
		tokenAddress,
		method,
		data,
		version,
		uuid7)
	if err != nil {
		return types.ContractOutput{}, fmt.Errorf("failed to send transaction: %w", err)
	}

	return contractOutput, nil
}

func (c *networkClient) GetToken(tokenAddress string, symbol string, name string) (types.ContractOutput, error) {
	from := c.publicKey

	if tokenAddress == "" && symbol == "" && name == "" {
		return types.ContractOutput{}, fmt.Errorf("token address, symbol or name must be set")
	}

	if err := keys.ValidateEDDSAPublicKeyHex(from); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid from address: %w", err)
	}

	if err := keys.ValidateEDDSAPublicKeyHex(tokenAddress); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid token address: %w", err)
	}

	method := tokenV1.METHOD_GET_TOKEN
	data := map[string]interface{}{
		"symbol": symbol,
		"name":   name,
	}

	contractOutput, err := c.GetState(tokenAddress, method, data)
	if err != nil {
		return types.ContractOutput{}, fmt.Errorf("failed to get state: %w", err)
	}

	return contractOutput, nil
}

func (c *networkClient) ListTokens(ownerAddress, symbol, name string, page, limit int, ascending bool) (types.ContractOutput, error) {
	from := c.publicKey

	if ownerAddress != "" {
		if err := keys.ValidateEDDSAPublicKeyHex(ownerAddress); err != nil {
			return types.ContractOutput{}, fmt.Errorf("invalid owner address: %w", err)
		}
	}

	if err := keys.ValidateEDDSAPublicKeyHex(from); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid from address: %w", err)
	}

	method := tokenV1.METHOD_LIST_TOKENS
	data := map[string]interface{}{
		"owner":            ownerAddress,
		"symbol":           symbol,
		"name":             name,
		"page":             page,
		"limit":            limit,
		"ascending":        ascending,
		"contract_version": tokenV1.TOKEN_CONTRACT_V1,
	}

	contractOutput, err := c.GetState("", method, data)
	if err != nil {
		return types.ContractOutput{}, fmt.Errorf("failed to get state: %w", err)
	}

	return contractOutput, nil
}

func (c *networkClient) GetTokenBalance(tokenAddress, ownerAddress string) (types.ContractOutput, error) {
	from := c.publicKey

	if tokenAddress == "" {
		return types.ContractOutput{}, fmt.Errorf("token address not set")
	}
	if ownerAddress == "" {
		return types.ContractOutput{}, fmt.Errorf("owner address not set")
	}

	if err := keys.ValidateEDDSAPublicKeyHex(from); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid from address: %w", err)
	}

	if err := keys.ValidateEDDSAPublicKeyHex(tokenAddress); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid token address: %w", err)
	}

	if err := keys.ValidateEDDSAPublicKeyHex(ownerAddress); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid owner address: %w", err)
	}

	method := tokenV1.METHOD_GET_TOKEN_BALANCE
	data := map[string]interface{}{
		"owner_address": ownerAddress,
	}

	contractOutput, err := c.GetState(tokenAddress, method, data)
	if err != nil {
		return types.ContractOutput{}, fmt.Errorf("failed to get state: %w", err)
	}

	return contractOutput, nil
}

func (c *networkClient) ListTokenBalances(tokenAddress, ownerAddress string, page, limit int, ascending bool) (types.ContractOutput, error) {
	from := c.publicKey

	if err := keys.ValidateEDDSAPublicKeyHex(from); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid from address: %w", err)
	}

	if tokenAddress != "" {

		if err := keys.ValidateEDDSAPublicKeyHex(tokenAddress); err != nil {
			return types.ContractOutput{}, fmt.Errorf("invalid token address: %w", err)
		}
	}
	if ownerAddress != "" {
		if err := keys.ValidateEDDSAPublicKeyHex(ownerAddress); err != nil {
			return types.ContractOutput{}, fmt.Errorf("invalid owner address: %w", err)
		}
	}

	method := tokenV1.METHOD_LIST_TOKEN_BALANCES
	data := map[string]interface{}{
		"owner_address":    ownerAddress,
		"page":             page,
		"limit":            limit,
		"ascending":        ascending,
		"token_address":    tokenAddress,
		"contract_version": tokenV1.TOKEN_CONTRACT_V1,
	}

	contractOutput, err := c.GetState("", method, data)
	if err != nil {
		return types.ContractOutput{}, fmt.Errorf("failed to get state: %w", err)
	}

	return contractOutput, nil
}
