package client_2finance

import (
	"fmt"
	"time"

	"gitlab.com/2finance/2finance-network/blockchain/contract/airdropV1"
	"gitlab.com/2finance/2finance-network/blockchain/encryption/keys"
	"gitlab.com/2finance/2finance-network/blockchain/types"
	"gitlab.com/2finance/2finance-network/blockchain/utils"
)

func (c *networkClient) NewAirdrop(
	address string,
	owner string,
	faucetAddress string,
	tokenAddress string,
	startTime time.Time,
	expireTime time.Time,
	paused bool,
	requestLimit int,
	claimAmount string,
	claimIntervalSeconds int64,
	title string,
	description string,
	shortDescription string,
	imageURL string,
	bannerURL string,
	category string,
	socialRequirements map[string]bool,
	postLinks []string,
	verificationType string,
	verifierPublicKey string,
	manualReviewRequired bool,
) (types.ContractOutput, error) {

	if address == "" {
		return types.ContractOutput{}, fmt.Errorf("airdrop address not set")
	}
	if owner == "" {
		return types.ContractOutput{}, fmt.Errorf("owner not set")
	}
	if faucetAddress == "" {
		return types.ContractOutput{}, fmt.Errorf("faucet address not set")
	}
	if tokenAddress == "" {
		return types.ContractOutput{}, fmt.Errorf("token address not set")
	}
	if title == "" {
		return types.ContractOutput{}, fmt.Errorf("title not set")
	}
	if claimAmount == "" {
		return types.ContractOutput{}, fmt.Errorf("claim amount not set")
	}
	if verificationType == "" {
		return types.ContractOutput{}, fmt.Errorf("verification type not set")
	}

	if err := keys.ValidateEDDSAPublicKeyHex(owner); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid owner address: %w", err)
	}
	if err := keys.ValidateEDDSAPublicKeyHex(tokenAddress); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid token address: %w", err)
	}
	if verifierPublicKey != "" {
		if err := keys.ValidateEDDSAPublicKeyHex(verifierPublicKey); err != nil {
			return types.ContractOutput{}, fmt.Errorf("invalid verifier public key: %w", err)
		}
	}

	from := c.publicKey
	if err := keys.ValidateEDDSAPublicKeyHex(from); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid from address: %w", err)
	}

	method := airdropV1.METHOD_NEW_AIRDROP
	data := map[string]interface{}{
		"owner":                  owner,
		"faucet_address":         faucetAddress,
		"token_address":          tokenAddress,
		"start_time":             startTime,
		"expire_time":            expireTime,
		"paused":                 paused,
		"request_limit":          requestLimit,
		"claim_amount":           claimAmount,
		"claim_interval_seconds": claimIntervalSeconds,
		"title":                  title,
		"description":            description,
		"short_description":      shortDescription,
		"image_url":              imageURL,
		"banner_url":             bannerURL,
		"category":               category,
		"social_requirements":    socialRequirements,
		"post_links":             postLinks,
		"verification_type":      verificationType,
		"verifier_public_key":    verifierPublicKey,
		"manual_review_required": manualReviewRequired,
	}
	version := uint8(1)
	uuid7, err := utils.NewUUID7()
	if err != nil {
		return types.ContractOutput{}, fmt.Errorf("failed to generate UUIDv7: %w", err)
	}

	return c.SignAndSendTransaction(c.chainId, from, address, method, data, version, uuid7)
}

func (c *networkClient) UpdateAirdropMetadata(
	address string,
	title string,
	description string,
	shortDescription string,
	imageURL string,
	bannerURL string,
	category string,
	socialRequirements map[string]bool,
	postLinks []string,
	verificationType string,
	verifierPublicKey string,
	manualReviewRequired bool,
) (types.ContractOutput, error) {

	if address == "" {
		return types.ContractOutput{}, fmt.Errorf("airdrop address not set")
	}

	from := c.publicKey

	method := airdropV1.METHOD_UPDATE_AIRDROP_METADATA
	data := map[string]interface{}{
		"title":                  title,
		"description":            description,
		"short_description":      shortDescription,
		"image_url":              imageURL,
		"banner_url":             bannerURL,
		"category":               category,
		"social_requirements":    socialRequirements,
		"post_links":             postLinks,
		"verification_type":      verificationType,
		"verifier_public_key":    verifierPublicKey,
		"manual_review_required": manualReviewRequired,
	}
	version := uint8(1)
	uuid7, err := utils.NewUUID7()
	if err != nil {
		return types.ContractOutput{}, fmt.Errorf("failed to generate UUIDv7: %w", err)
	}

	return c.SignAndSendTransaction(c.chainId, from, address, method, data, version, uuid7)
}

func (c *networkClient) AllowOracles(address string, oracles map[string]bool) (types.ContractOutput, error) {
	if address == "" {
		return types.ContractOutput{}, fmt.Errorf("airdrop address not set")
	}
	if len(oracles) == 0 {
		return types.ContractOutput{}, fmt.Errorf("oracles map is empty")
	}

	from := c.publicKey
	method := airdropV1.METHOD_ALLOW_ORACLES
	version := uint8(1)
	uuid7, err := utils.NewUUID7()
	if err != nil {
		return types.ContractOutput{}, fmt.Errorf("failed to generate UUIDv7: %w", err)
	}

	return c.SignAndSendTransaction(c.chainId, from, address, method, map[string]interface{}{
		"oracles": oracles,
	}, version, uuid7)
}

func (c *networkClient) DisallowOracles(address string, oracles map[string]bool) (types.ContractOutput, error) {
	if address == "" {
		return types.ContractOutput{}, fmt.Errorf("airdrop address not set")
	}
	if len(oracles) == 0 {
		return types.ContractOutput{}, fmt.Errorf("oracles map is empty")
	}

	from := c.publicKey
	method := airdropV1.METHOD_DISALLOW_ORACLES
	version := uint8(1)
	uuid7, err := utils.NewUUID7()
	if err != nil {
		return types.ContractOutput{}, fmt.Errorf("failed to generate UUIDv7: %w", err)
	}

	return c.SignAndSendTransaction(c.chainId, from, address, method, map[string]interface{}{
		"oracles": oracles,
	}, version, uuid7)
}

func (c *networkClient) DepositAirdrop(
	address string,
	amount string,
	tokenType string,
	uuid string,
) (types.ContractOutput, error) {

	if address == "" {
		return types.ContractOutput{}, fmt.Errorf("airdrop address not set")
	}
	if amount == "" {
		return types.ContractOutput{}, fmt.Errorf("amount not set")
	}
	if tokenType == "" {
		return types.ContractOutput{}, fmt.Errorf("token type not set")
	}

	from := c.publicKey
	method := airdropV1.METHOD_DEPOSIT_AIRDROP
	version := uint8(1)
	
	uuid7, err := utils.NewUUID7()
	if err != nil {
		return types.ContractOutput{}, fmt.Errorf("failed to generate UUIDv7: %w", err)
	}

	return c.SignAndSendTransaction(c.chainId, from, address, method, map[string]interface{}{
		"amount":     amount,
		"token_type": tokenType,
		"uuid":       uuid,
	}, version, uuid7)
}

func (c *networkClient) ClaimAirdrop(address, tokenType string) (types.ContractOutput, error) {
	if address == "" {
		return types.ContractOutput{}, fmt.Errorf("airdrop address not set")
	}
	if tokenType == "" {
		return types.ContractOutput{}, fmt.Errorf("token type not set")
	}

	from := c.publicKey
	method := airdropV1.METHOD_CLAIM_AIRDROP
	version := uint8(1)

	uuid7, err := utils.NewUUID7()
	if err != nil {
		return types.ContractOutput{}, fmt.Errorf("failed to generate UUIDv7: %w", err)
	}

	return c.SignAndSendTransaction(c.chainId, from, address, method, map[string]interface{}{
		"address":    address,
		"token_type": tokenType,
	}, version, uuid7)
}

func (c *networkClient) WithdrawAirdropFunds(
	address string,
	amount string,
	tokenType string,
	uuid string,
) (types.ContractOutput, error) {

	if address == "" {
		return types.ContractOutput{}, fmt.Errorf("airdrop address not set")
	}
	if amount == "" {
		return types.ContractOutput{}, fmt.Errorf("amount not set")
	}

	from := c.publicKey
	method := airdropV1.METHOD_WITHDRAW_AIRDROP
	version := uint8(1)
	uuid7, err := utils.NewUUID7()
	if err != nil {
		return types.ContractOutput{}, fmt.Errorf("failed to generate UUIDv7: %w", err)
	}
	return c.SignAndSendTransaction(c.chainId, from, address, method, map[string]interface{}{
		"amount":     amount,
		"token_type": tokenType,
		"uuid":       uuid,
	}, version, uuid7)
}

func (c *networkClient) PauseAirdrop(airdropAddress string) (types.ContractOutput, error) {
	if airdropAddress == "" {
		return types.ContractOutput{}, fmt.Errorf("airdrop address not set")
	}

	from := c.publicKey
	if from == "" {
		return types.ContractOutput{}, fmt.Errorf("from address not set")
	}
	if err := keys.ValidateEDDSAPublicKeyHex(from); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid from address: %w", err)
	}
	if err := keys.ValidateEDDSAPublicKeyHex(airdropAddress); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid airdrop address: %w", err)
	}

	method := airdropV1.METHOD_PAUSE_AIRDROP
	// Evite "data nil" — seu backend já reclamou disso em outros métodos.
	data := map[string]interface{}{
		"address": airdropAddress,
		"paused":  true,
	}
	version := uint8(1)
	uuid7, err := utils.NewUUID7()
	if err != nil {
		return types.ContractOutput{}, fmt.Errorf("failed to generate UUIDv7: %w", err)
	}

	return c.SignAndSendTransaction(c.chainId, from, airdropAddress, method, data, version, uuid7)
}

func (c *networkClient) UnpauseAirdrop(airdropAddress string) (types.ContractOutput, error) {
	if airdropAddress == "" {
		return types.ContractOutput{}, fmt.Errorf("airdrop address not set")
	}

	from := c.publicKey
	if from == "" {
		return types.ContractOutput{}, fmt.Errorf("from address not set")
	}
	if err := keys.ValidateEDDSAPublicKeyHex(from); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid from address: %w", err)
	}
	if err := keys.ValidateEDDSAPublicKeyHex(airdropAddress); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid airdrop address: %w", err)
	}

	method := airdropV1.METHOD_UNPAUSE_AIRDROP
	data := map[string]interface{}{
		"address": airdropAddress,
		"paused":  false,
	}
	version := uint8(1)

	uuid7, err := utils.NewUUID7()
	if err != nil {
		return types.ContractOutput{}, fmt.Errorf("failed to generate UUIDv7: %w", err)
	}

	return c.SignAndSendTransaction(c.chainId, from, airdropAddress, method, data, version, uuid7)
}

func (c *networkClient) AttestParticipantEligibility(
	address string,
	wallet string,
	approved bool,
) (types.ContractOutput, error) {

	if address == "" {
		return types.ContractOutput{}, fmt.Errorf("airdrop address not set")
	}
	if wallet == "" {
		return types.ContractOutput{}, fmt.Errorf("wallet not set")
	}
	if err := keys.ValidateEDDSAPublicKeyHex(wallet); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid wallet address: %w", err)
	}

	from := c.publicKey
	method := airdropV1.METHOD_ATTEST_ELIGIBILITY
	version := uint8(1)

	uuid7, err := utils.NewUUID7()
	if err != nil {
		return types.ContractOutput{}, fmt.Errorf("failed to generate UUIDv7: %w", err)
	}

	return c.SignAndSendTransaction(c.chainId, from, address, method, map[string]interface{}{
		"wallet":   wallet,
		"approved": approved,
	}, version, uuid7)
}

func (c *networkClient) ManuallyAttestParticipantEligibility(
	airdropAddress string,
	wallet string,
	approved bool,
) (types.ContractOutput, error) {

	from := c.publicKey
	if from == "" {
		return types.ContractOutput{}, fmt.Errorf("from address not set")
	}
	if airdropAddress == "" {
		return types.ContractOutput{}, fmt.Errorf("airdrop address not set")
	}
	if wallet == "" {
		return types.ContractOutput{}, fmt.Errorf("wallet not set")
	}

	if err := keys.ValidateEDDSAPublicKeyHex(from); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid from address: %w", err)
	}
	if err := keys.ValidateEDDSAPublicKeyHex(wallet); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid wallet address: %w", err)
	}

	method := airdropV1.METHOD_MANUAL_ATTEST_ELIGIBILITY
	data := map[string]interface{}{
		"wallet":   wallet,
		"approved": approved,
	}
	version := uint8(1)
	uuid7, err := utils.NewUUID7()
	if err != nil {
		return types.ContractOutput{}, fmt.Errorf("failed to generate UUIDv7: %w", err)
	}

	out, err := c.SignAndSendTransaction(
		c.chainId,
		from,
		airdropAddress,
		method,
		data,
		version,
		uuid7,
	)
	if err != nil {
		return types.ContractOutput{}, fmt.Errorf("failed to send transaction: %w", err)
	}

	return out, nil
}

func (c *networkClient) GetAirdrop(address string) (types.ContractOutput, error) {
	from := c.publicKey
	if from == "" {
		return types.ContractOutput{}, fmt.Errorf("from address not set")
	}

	if address == "" {
		return types.ContractOutput{}, fmt.Errorf("airdrop address must be set")
	}

	if err := keys.ValidateEDDSAPublicKeyHex(from); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid from address: %w", err)
	}

	if err := keys.ValidateEDDSAPublicKeyHex(address); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid airdrop address: %w", err)
	}

	method := airdropV1.METHOD_GET_AIRDROP

	contractOutput, err := c.GetState(address, method, nil)
	if err != nil {
		return types.ContractOutput{}, fmt.Errorf("failed to get state: %w", err)
	}

	return contractOutput, nil
}


func (c *networkClient) ListAirdrops(
	owner string,
	page, limit int,
	ascending bool,
) (types.ContractOutput, error) {
	from := c.publicKey
	if from == "" {
		return types.ContractOutput{}, fmt.Errorf("from address not set")
	}

	if err := keys.ValidateEDDSAPublicKeyHex(from); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid from address: %w", err)
	}

	if owner != "" {
		if err := keys.ValidateEDDSAPublicKeyHex(owner); err != nil {
			return types.ContractOutput{}, fmt.Errorf("invalid owner address: %w", err)
		}
	}

	if page < 1 {
		return types.ContractOutput{}, fmt.Errorf("page must be greater than 0")
	}
	if limit < 1 {
		return types.ContractOutput{}, fmt.Errorf("limit must be greater than 0")
	}

	method := airdropV1.METHOD_LIST_AIRDROPS

	data := map[string]interface{}{
		"owner":            owner,
		"page":             page,
		"limit":            limit,
		"ascending":        ascending,
		"contract_version": airdropV1.AIRDROP_CONTRACT_V1,
	}

	contractOutput, err := c.GetState("", method, data)
	if err != nil {
		return types.ContractOutput{}, fmt.Errorf("failed to list airdrop states: %w", err)
	}

	return contractOutput, nil
}


