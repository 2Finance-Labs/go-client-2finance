package client_2finance

import (
	"fmt"
	"gitlab.com/2finance/2finance-network/blockchain/contract/dropV1"
	"gitlab.com/2finance/2finance-network/blockchain/contract/dropV1/inputs"
	"gitlab.com/2finance/2finance-network/blockchain/encryption/keys"
	"gitlab.com/2finance/2finance-network/blockchain/types"
	"gitlab.com/2finance/2finance-network/blockchain/utils"
)

func (c *networkClient) NewDrop(in inputs.InputNewDrop) (types.ContractOutput, error) {

	if in.Address == "" {
		return types.ContractOutput{}, fmt.Errorf("drop address not set")
	}
	if in.Owner == "" {
		return types.ContractOutput{}, fmt.Errorf("owner not set")
	}
	if in.Title == "" {
		return types.ContractOutput{}, fmt.Errorf("title not set")
	}
	if in.VerificationType == "" {
		return types.ContractOutput{}, fmt.Errorf("verification type not set")
	}

	if err := keys.ValidateEDDSAPublicKeyHex(in.Owner); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid owner address: %w", err)
	}

	from := c.publicKey
	if err := keys.ValidateEDDSAPublicKeyHex(from); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid from address: %w", err)
	}

	method := dropV1.METHOD_NEW_DROP
	data := map[string]interface{}{
		"address":                in.Address,
		"program_address":        in.ProgramAddress,
		"token_address":          in.TokenAddress,
		"owner":                  in.Owner,
		"title":                  in.Title,
		"description":            in.Description,
		"short_description":      in.ShortDescription,
		"image_url":              in.ImageURL,
		"banner_url":             in.BannerURL,
		"categories":             in.Categories,
		"social_requirements":    in.SocialRequirements,
		"post_links":             in.PostLinks,
		"verification_type":      in.VerificationType,
		"start_at":               in.StartAt,
		"expire_at":              in.ExpireAt,
		"request_limit":          in.RequestLimit,
		"claim_amount":           in.ClaimAmount,
		"claim_interval_seconds": in.ClaimIntervalSeconds,
	}
	version := uint8(1)
	uuid7, err := utils.NewUUID7()
	if err != nil {
		return types.ContractOutput{}, fmt.Errorf("failed to generate UUIDv7: %w", err)
	}

	return c.SignAndSendTransaction(c.chainId, from, in.Address, method, data, version, uuid7)
}

func (c *networkClient) UpdateDropMetadata(
	in inputs.InputUpdateDropMetadata,
) (types.ContractOutput, error) {

	if in.Address == "" {
		return types.ContractOutput{}, fmt.Errorf("drop address not set")
	}

	from := c.publicKey

	method := dropV1.METHOD_UPDATE_DROP_METADATA
		data := map[string]interface{}{
		"address":                in.Address,
		"program_address":        in.ProgramAddress,
		"token_address":          in.TokenAddress,
		"title":                  in.Title,
		"description":            in.Description,
		"short_description":      in.ShortDescription,
		"image_url":              in.ImageURL,
		"banner_url":             in.BannerURL,
		"categories":             in.Categories,
		"social_requirements":    in.SocialRequirements,
		"post_links":             in.PostLinks,
		"verification_type":      in.VerificationType,
		"start_at":               in.StartAt,
		"expire_at":              in.ExpireAt,
		"request_limit":          in.RequestLimit,
		"claim_amount":           in.ClaimAmount,
		"claim_interval_seconds": in.ClaimIntervalSeconds,
	}
	version := uint8(1)
	uuid7, err := utils.NewUUID7()
	if err != nil {
		return types.ContractOutput{}, fmt.Errorf("failed to generate UUIDv7: %w", err)
	}

	return c.SignAndSendTransaction(c.chainId, from, in.Address, method, data, version, uuid7)
}


func (c *networkClient) AllowOracles(address string, oracles map[string]bool) (types.ContractOutput, error) {
	if address == "" {
		return types.ContractOutput{}, fmt.Errorf("drop address not set")
	}
	if len(oracles) == 0 {
		return types.ContractOutput{}, fmt.Errorf("oracles map is empty")
	}

	from := c.publicKey
	method := dropV1.METHOD_ALLOW_ORACLES
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
		return types.ContractOutput{}, fmt.Errorf("drop address not set")
	}
	if len(oracles) == 0 {
		return types.ContractOutput{}, fmt.Errorf("oracles map is empty")
	}

	from := c.publicKey
	method := dropV1.METHOD_DISALLOW_ORACLES
	version := uint8(1)
	uuid7, err := utils.NewUUID7()
	if err != nil {
		return types.ContractOutput{}, fmt.Errorf("failed to generate UUIDv7: %w", err)
	}

	return c.SignAndSendTransaction(c.chainId, from, address, method, map[string]interface{}{
		"oracles": oracles,
	}, version, uuid7)
}

func (c *networkClient) DepositDrop(
	address string,
	programAddress string,
	tokenAddress string,
	amount string,
	uuid []string,
) (types.ContractOutput, error) {

	if address == "" {
		return types.ContractOutput{}, fmt.Errorf("drop address not set")
	}
	if amount == "" {
		return types.ContractOutput{}, fmt.Errorf("amount not set")
	}

	from := c.publicKey
	method := dropV1.METHOD_DEPOSIT_DROP
	version := uint8(1)
	
	uuid7, err := utils.NewUUID7()
	if err != nil {
		return types.ContractOutput{}, fmt.Errorf("failed to generate UUIDv7: %w", err)
	}

	return c.SignAndSendTransaction(c.chainId, from, address, method, map[string]interface{}{
		"program_address": programAddress,
		"token_address":   tokenAddress,
		"amount":     amount,
		"uuid":       uuid,
	}, version, uuid7)
}

func (c *networkClient) ClaimDrop(address string) (types.ContractOutput, error) {
	if address == "" {
		return types.ContractOutput{}, fmt.Errorf("drop address not set")
	}

	from := c.publicKey
	method := dropV1.METHOD_CLAIM_DROP
	version := uint8(1)

	uuid7, err := utils.NewUUID7()
	if err != nil {
		return types.ContractOutput{}, fmt.Errorf("failed to generate UUIDv7: %w", err)
	}

	return c.SignAndSendTransaction(c.chainId, from, address, method, map[string]interface{}{
	}, version, uuid7)
}

func (c *networkClient) WithdrawDrop(
	address string,
	programAddress string,
	tokenAddress string,
	amount string,
	uuid []string,
) (types.ContractOutput, error) {

	if address == "" {
		return types.ContractOutput{}, fmt.Errorf("drop address not set")
	}
	if amount == "" {
		return types.ContractOutput{}, fmt.Errorf("amount not set")
	}

	from := c.publicKey
	method := dropV1.METHOD_WITHDRAW_DROP
	version := uint8(1)
	uuid7, err := utils.NewUUID7()
	if err != nil {
		return types.ContractOutput{}, fmt.Errorf("failed to generate UUIDv7: %w", err)
	}
	return c.SignAndSendTransaction(c.chainId, from, address, method, map[string]interface{}{
		"program_address": programAddress,
		"token_address":   tokenAddress,
		"amount":     amount,
		"uuid":       uuid,
	}, version, uuid7)
}

func (c *networkClient) PauseDrop(dropAddress string) (types.ContractOutput, error) {
	if dropAddress == "" {
		return types.ContractOutput{}, fmt.Errorf("drop address not set")
	}

	from := c.publicKey
	if from == "" {
		return types.ContractOutput{}, fmt.Errorf("from address not set")
	}
	if err := keys.ValidateEDDSAPublicKeyHex(from); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid from address: %w", err)
	}
	if err := keys.ValidateEDDSAPublicKeyHex(dropAddress); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid drop address: %w", err)
	}

	method := dropV1.METHOD_PAUSE_DROP
	// Evite "data nil" — seu backend já reclamou disso em outros métodos.
	data := map[string]interface{}{
		"address": dropAddress,
		"paused":  true,
	}
	version := uint8(1)
	uuid7, err := utils.NewUUID7()
	if err != nil {
		return types.ContractOutput{}, fmt.Errorf("failed to generate UUIDv7: %w", err)
	}

	return c.SignAndSendTransaction(c.chainId, from, dropAddress, method, data, version, uuid7)
}

func (c *networkClient) UnpauseDrop(dropAddress string) (types.ContractOutput, error) {
	if dropAddress == "" {
		return types.ContractOutput{}, fmt.Errorf("drop address not set")
	}

	from := c.publicKey
	if from == "" {
		return types.ContractOutput{}, fmt.Errorf("from address not set")
	}
	if err := keys.ValidateEDDSAPublicKeyHex(from); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid from address: %w", err)
	}
	if err := keys.ValidateEDDSAPublicKeyHex(dropAddress); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid drop address: %w", err)
	}

	method := dropV1.METHOD_UNPAUSE_DROP
	data := map[string]interface{}{
		"address": dropAddress,
		"paused":  false,
	}
	version := uint8(1)

	uuid7, err := utils.NewUUID7()
	if err != nil {
		return types.ContractOutput{}, fmt.Errorf("failed to generate UUIDv7: %w", err)
	}

	return c.SignAndSendTransaction(c.chainId, from, dropAddress, method, data, version, uuid7)
}

func (c *networkClient) AttestParticipantEligibility(
	address string,
	wallet string,
	approved bool,
) (types.ContractOutput, error) {

	if address == "" {
		return types.ContractOutput{}, fmt.Errorf("drop address not set")
	}
	if wallet == "" {
		return types.ContractOutput{}, fmt.Errorf("wallet not set")
	}
	if err := keys.ValidateEDDSAPublicKeyHex(wallet); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid wallet address: %w", err)
	}

	from := c.publicKey
	method := dropV1.METHOD_ATTEST_ELIGIBILITY
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
	dropAddress string,
	wallet string,
	approved bool,
) (types.ContractOutput, error) {

	from := c.publicKey
	if from == "" {
		return types.ContractOutput{}, fmt.Errorf("from address not set")
	}
	if dropAddress == "" {
		return types.ContractOutput{}, fmt.Errorf("drop address not set")
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

	method := dropV1.METHOD_MANUAL_ATTEST_ELIGIBILITY
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
		dropAddress,
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

func (c *networkClient) GetDrop(address string) (types.ContractOutput, error) {
	from := c.publicKey
	if from == "" {
		return types.ContractOutput{}, fmt.Errorf("from address not set")
	}

	if address == "" {
		return types.ContractOutput{}, fmt.Errorf("drop address must be set")
	}

	if err := keys.ValidateEDDSAPublicKeyHex(from); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid from address: %w", err)
	}

	if err := keys.ValidateEDDSAPublicKeyHex(address); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid drop address: %w", err)
	}

	method := dropV1.METHOD_GET_DROP

	contractOutput, err := c.GetState(address, method, nil)
	if err != nil {
		return types.ContractOutput{}, fmt.Errorf("failed to get state: %w", err)
	}

	return contractOutput, nil
}

func (c *networkClient) ListDrops(
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

	method := dropV1.METHOD_LIST_DROPS

	data := map[string]interface{}{
		"owner":            owner,
		"page":             page,
		"limit":            limit,
		"ascending":        ascending,
		"contract_version": dropV1.DROP_CONTRACT_V1,
	}

	contractOutput, err := c.GetState("", method, data)
	if err != nil {
		return types.ContractOutput{}, fmt.Errorf("failed to list drop states: %w", err)
	}

	return contractOutput, nil
}

func (c *networkClient) LastClaimed(address string, wallet string) (types.ContractOutput, error) {
	from := c.publicKey
	if from == "" {
		return types.ContractOutput{}, fmt.Errorf("from address not set")
	}

	if address == "" {
		return types.ContractOutput{}, fmt.Errorf("drop address must be set")
	}
	if wallet == "" {
		return types.ContractOutput{}, fmt.Errorf("wallet must be set")
	}

	if err := keys.ValidateEDDSAPublicKeyHex(from); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid from address: %w", err)
	}

	if err := keys.ValidateEDDSAPublicKeyHex(address); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid drop address: %w", err)
	}

	if err := keys.ValidateEDDSAPublicKeyHex(wallet); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid wallet address: %w", err)
	}

	method := dropV1.METHOD_LAST_CLAIMED_DROP

	data := map[string]interface{}{
		"wallet": wallet,
	}

	contractOutput, err := c.GetState(address, method, data)
	if err != nil {
		return types.ContractOutput{}, fmt.Errorf("failed to get state: %w", err)
	}

	return contractOutput, nil
}
