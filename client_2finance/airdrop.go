package client_2finance

import (
	"fmt"
	"time"

	"gitlab.com/2finance/2finance-network/blockchain/contract/airdropV1"
	"gitlab.com/2finance/2finance-network/blockchain/contract/airdropV1/domain"
	"gitlab.com/2finance/2finance-network/blockchain/encryption/keys"
	"gitlab.com/2finance/2finance-network/blockchain/types"
)

func (c *networkClient) NewAirdrop(
	address string,
	owner string,
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
	eligibleWallets map[string]bool,
	nonce uint64,
) (types.ContractOutput, error) {

	if address == "" {
		return types.ContractOutput{}, fmt.Errorf("airdrop address not set")
	}
	if owner == "" {
		return types.ContractOutput{}, fmt.Errorf("owner not set")
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

	if err := keys.ValidateEDDSAPublicKey(owner); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid owner address: %w", err)
	}
	if err := keys.ValidateEDDSAPublicKey(tokenAddress); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid token address: %w", err)
	}
	if verifierPublicKey != "" {
		if err := keys.ValidateEDDSAPublicKey(verifierPublicKey); err != nil {
			return types.ContractOutput{}, fmt.Errorf("invalid verifier public key: %w", err)
		}
	}

	from := c.publicKey
	if from == "" {
		return types.ContractOutput{}, fmt.Errorf("from address not set")
	}
	if err := keys.ValidateEDDSAPublicKey(from); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid from address: %w", err)
	}

	method := airdropV1.METHOD_NEW_AIRDROP
	data := map[string]interface{}{
		"owner":                    owner,
		"token_address":            tokenAddress,
		"start_time":               startTime,
		"expire_time":              expireTime,
		"paused":                   paused,
		"request_limit":            requestLimit,
		"claim_amount":             claimAmount,
		"claim_interval_seconds":   claimIntervalSeconds,
		"title":                    title,
		"description":              description,
		"short_description":        shortDescription,
		"image_url":                imageURL,
		"banner_url":               bannerURL,
		"category":                 category,
		"social_requirements":      socialRequirements,
		"post_links":               postLinks,
		"verification_type":        verificationType,
		"verifier_public_key":      verifierPublicKey,
		"manual_review_required":   manualReviewRequired,
		"eligible_wallets":         eligibleWallets,
		"nonce":                    nonce,
	}

	return c.SignAndSendTransaction(from, address, method, data)
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
	eligibleWallets map[string]bool,
) (types.ContractOutput, error) {

	if address == "" {
		return types.ContractOutput{}, fmt.Errorf("airdrop address not set")
	}

	from := c.publicKey
	if from == "" {
		return types.ContractOutput{}, fmt.Errorf("from address not set")
	}

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
		"eligible_wallets":       eligibleWallets,
	}

	return c.SignAndSendTransaction(from, address, method, data)
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

	return c.SignAndSendTransaction(from, address, method, map[string]interface{}{
		"oracles": oracles,
	})
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

	return c.SignAndSendTransaction(from, address, method, map[string]interface{}{
		"oracles": oracles,
	})
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

	return c.SignAndSendTransaction(from, address, method, map[string]interface{}{
		"amount":     amount,
		"token_type": tokenType,
		"uuid":       uuid,
	})
}

func (c *networkClient) ClaimAirdrop(address string) (types.ContractOutput, error) {
	if address == "" {
		return types.ContractOutput{}, fmt.Errorf("airdrop address not set")
	}

	from := c.publicKey
	method := airdropV1.METHOD_CLAIM_AIRDROP

	return c.SignAndSendTransaction(from, address, method, nil)
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

	return c.SignAndSendTransaction(from, address, method, map[string]interface{}{
		"amount":     amount,
		"token_type": tokenType,
		"uuid":       uuid,
	})
}

func (c *networkClient) PauseAirdrop(address string) (types.ContractOutput, error) {
	if address == "" {
		return types.ContractOutput{}, fmt.Errorf("airdrop address not set")
	}

	from := c.publicKey
	method := airdropV1.METHOD_PAUSE_AIRDROP

	return c.SignAndSendTransaction(from, address, method, map[string]interface{}{
		"paused": true,
	})
}

func (c *networkClient) UnpauseAirdrop(address string) (types.ContractOutput, error) {
	if address == "" {
		return types.ContractOutput{}, fmt.Errorf("airdrop address not set")
	}

	from := c.publicKey
	method := airdropV1.METHOD_UNPAUSE_AIRDROP

	return c.SignAndSendTransaction(from, address, method, map[string]interface{}{
		"paused": false,
	})
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
	if err := keys.ValidateEDDSAPublicKey(wallet); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid wallet address: %w", err)
	}

	from := c.publicKey
	method := airdropV1.METHOD_ATTEST_ELIGIBILITY

	return c.SignAndSendTransaction(from, address, method, map[string]interface{}{
		"wallet":   wallet,
		"approved": approved,
	})
}

// func (c *networkClient) GetAirdrop(address string) (types.ContractOutput, error) {
// 	if address == "" {
// 		return types.ContractOutput{}, fmt.Errorf("airdrop address not set")
// 	}

// 	from := c.publicKey
// 	method := airdropV1.METHOD_GET_AIRDROP

// 	return c.GetState(address, method, nil)
// }

// func (c *networkClient) ListAirdrops(
// 	owner string,
// 	page int,
// 	limit int,
// 	ascending bool,
// ) (types.ContractOutput, error) {

// 	if owner != "" {
// 		if err := keys.ValidateEDDSAPublicKey(owner); err != nil {
// 			return types.ContractOutput{}, fmt.Errorf("invalid owner address: %w", err)
// 		}
// 	}

// 	from := c.publicKey
// 	method := airdropV1.METHOD_LIST_AIRDROPS

// 	data := map[string]interface{}{
// 		"owner":     owner,
// 		"page":      page,
// 		"limit":     limit,
// 		"ascending": ascending,
// 	}

// 	return c.GetState("", method, data)
// }