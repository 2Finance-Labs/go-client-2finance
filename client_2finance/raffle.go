package client_2finance

import (
	"fmt"
	"time"

	"gitlab.com/2finance/2finance-network/blockchain/contract/raffleV1"
	"gitlab.com/2finance/2finance-network/blockchain/contract/tokenV1/domain"
	"gitlab.com/2finance/2finance-network/blockchain/encryption/keys"
	"gitlab.com/2finance/2finance-network/blockchain/types"
)

// AddRaffle creates a new raffle instance (to = DEPLOY address). The tx sender becomes the owner, but an explicit owner is also recorded.
func (c *networkClient) AddRaffle(
	address, owner, tokenAddress, ticketPrice string,
	maxEntries, maxEntriesPerUser int,
	startAt, expiredAt time.Time,
	paused bool,
	seedCommitHex string,
	metadata map[string]string,
) (types.ContractOutput, error) {
	from := c.publicKey

	if err := keys.ValidateEDDSAPublicKey(from); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid from address: %w", err)
	}
	if address == "" {
		return types.ContractOutput{}, fmt.Errorf("address not set")
	}
	if err := keys.ValidateEDDSAPublicKey(address); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid raffle address: %w", err)
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
	if ticketPrice == "" {
		return types.ContractOutput{}, fmt.Errorf("ticket_price not set")
	}
	if maxEntries <= 0 {
		return types.ContractOutput{}, fmt.Errorf("max_entries must be > 0")
	}
	if maxEntriesPerUser <= 0 {
		return types.ContractOutput{}, fmt.Errorf("max_entries_per_user must be > 0")
	}
	if maxEntriesPerUser > maxEntries {
		return types.ContractOutput{}, fmt.Errorf("max_entries_per_user cannot exceed max_entries")
	}
	if startAt.IsZero() {
		return types.ContractOutput{}, fmt.Errorf("start_at not set")
	}
	if expiredAt.IsZero() {
		return types.ContractOutput{}, fmt.Errorf("expired_at not set")
	}

	to := address
	method := raffleV1.METHOD_ADD_RAFFLE

	data := map[string]interface{}{
		"address":              address,
		"owner":                owner,
		"token_address":        tokenAddress,
		"ticket_price":         ticketPrice,
		"max_entries":          maxEntries,
		"max_entries_per_user": maxEntriesPerUser,
		"start_at":             startAt,
		"expired_at":           expiredAt,
		"paused":               paused,
		"seed_commit_hex":      seedCommitHex,
		"metadata":             metadata,
	}

	return c.SignAndSendTransaction(c.chainId, from, to, method, data)
}

// UpdateRaffle updates mutable fields of an existing raffle.
func (c *networkClient) UpdateRaffle(
	address, tokenAddress, ticketPrice string,
	maxEntries, maxEntriesPerUser int,
	startAt, expiredAt *time.Time,
	seedCommitHex string,
	metadata map[string]string,
) (types.ContractOutput, error) {
	if address == "" {
		return types.ContractOutput{}, fmt.Errorf("address not set")
	}
	if err := keys.ValidateEDDSAPublicKey(address); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid address: %w", err)
	}

	from := c.publicKey

	if err := keys.ValidateEDDSAPublicKey(from); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid from address: %w", err)
	}

	if tokenAddress != "" { // optional change
		if err := keys.ValidateEDDSAPublicKey(tokenAddress); err != nil {
			return types.ContractOutput{}, fmt.Errorf("invalid token address: %w", err)
		}
	}
	if ticketPrice == "" && maxEntries == 0 && maxEntriesPerUser == 0 && startAt == nil && expiredAt == nil && seedCommitHex == "" && len(metadata) == 0 {
		return types.ContractOutput{}, fmt.Errorf("no fields to update")
	}
	if maxEntries < 0 || maxEntriesPerUser < 0 {
		return types.ContractOutput{}, fmt.Errorf("max entries must be >= 0")
	}
	if maxEntries > 0 && maxEntriesPerUser > maxEntries {
		return types.ContractOutput{}, fmt.Errorf("max_entries_per_user cannot exceed max_entries")
	}

	to := address
	method := raffleV1.METHOD_UPDATE_RAFFLE
	data := map[string]interface{}{
		"address":              address,
		"token_address":        tokenAddress,
		"ticket_price":         ticketPrice,
		"max_entries":          maxEntries,
		"max_entries_per_user": maxEntriesPerUser,
		"seed_commit_hex":      seedCommitHex,
		"metadata":             metadata,
	}
	if startAt != nil {
		data["start_at"] = *startAt
	}
	if expiredAt != nil {
		data["expired_at"] = *expiredAt
	}

	return c.SignAndSendTransaction(c.chainId, from, to, method, data)
}

// PauseRaffle sets paused=true. OnlyOwner.
func (c *networkClient) PauseRaffle(address string, paused bool) (types.ContractOutput, error) {
	if address == "" {
		return types.ContractOutput{}, fmt.Errorf("address not set")
	}
	if err := keys.ValidateEDDSAPublicKey(address); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid address: %w", err)
	}
	if !paused {
		return types.ContractOutput{}, fmt.Errorf("paused must be true: Pause: %t", paused)
	}

	from := c.publicKey
	if from == "" {
		return types.ContractOutput{}, fmt.Errorf("from address not set")
	}
	if err := keys.ValidateEDDSAPublicKey(from); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid from address: %w", err)
	}

	to := address
	method := raffleV1.METHOD_PAUSE_RAFFLE
	data := map[string]interface{}{"address": address, "paused": paused}
	return c.SignAndSendTransaction(c.chainId, from, to, method, data)
}

// UnpauseRaffle sets paused=false. OnlyOwner.
func (c *networkClient) UnpauseRaffle(address string, paused bool) (types.ContractOutput, error) {
	if address == "" {
		return types.ContractOutput{}, fmt.Errorf("address not set")
	}
	if err := keys.ValidateEDDSAPublicKey(address); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid address: %w", err)
	}
	if paused {
		return types.ContractOutput{}, fmt.Errorf("paused must be false: Pause: %t", paused)
	}

	from := c.publicKey
	if from == "" {
		return types.ContractOutput{}, fmt.Errorf("from address not set")
	}
	if err := keys.ValidateEDDSAPublicKey(from); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid from address: %w", err)
	}

	to := address
	method := raffleV1.METHOD_UNPAUSE_RAFFLE
	data := map[string]interface{}{"address": address, "paused": paused}
	return c.SignAndSendTransaction(c.chainId, from, to, method, data)
}

func (c *networkClient) EnterRaffle(address string, tickets int, payTokenAddress string) (types.ContractOutput, error) {
    // Pre-check client state
    from := c.publicKey
	if from == "" { return types.ContractOutput{}, fmt.Errorf("from address not set") }

	// Validate inputs (server/domain will also validate)
	if err := keys.ValidateEDDSAPublicKey(c.publicKey); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid client public key: %w", err)
	}
	if address == "" {
		return types.ContractOutput{}, fmt.Errorf("address is required")
	}
	if err := keys.ValidateEDDSAPublicKey(address); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid raffle address: %w", err)
	}
	if tickets <= 0 {
		return types.ContractOutput{}, fmt.Errorf("tickets must be > 0")
	}
	if payTokenAddress == "" {
		return types.ContractOutput{}, fmt.Errorf("pay_token_address is required")
	}
	if err := keys.ValidateEDDSAPublicKey(payTokenAddress); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid pay_token_address: %w", err)
	}
	if tokenType == "" {
		return types.ContractOutput{}, fmt.Errorf("tokenType not set")
	}
	if tokenType == domain.NON_FUNGIBLE {
		if uuid == "" {
			return types.ContractOutput{}, fmt.Errorf("uuid must be set for non-fungible tokens")
		}
	}
	// Exact payload fields expected by the refactored EnterRaffle handler
	data := map[string]interface{}{
		"address":           address,
		"entrant":           c.publicKey,
		"tickets":           tickets,
		"pay_token_address": payTokenAddress,
		"token_type":        tokenType,
		"uuid":              uuid,
	}

    // Send: from = caller (client public key), to = raffle instance address
    return c.SignAndSendTransaction(
        c.chainId,
        from,
        address,
        raffleV1.METHOD_ENTER_RAFFLE,    // method constant
        data,
    )
}

// DrawRaffle reveals the seed and draws winners (commit-reveal). OnlyOwner/Moderator.
func (c *networkClient) DrawRaffle(address, revealSeed string) (types.ContractOutput, error) {
	if address == "" {
		return types.ContractOutput{}, fmt.Errorf("address not set")
	}
	if err := keys.ValidateEDDSAPublicKey(address); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid address: %w", err)
	}
	if revealSeed == "" {
		return types.ContractOutput{}, fmt.Errorf("reveal_seed not set")
	}

	from := c.publicKey
	if from == "" {
		return types.ContractOutput{}, fmt.Errorf("from address not set")
	}
	if err := keys.ValidateEDDSAPublicKey(from); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid from address: %w", err)
	}

	to := address
	method := raffleV1.METHOD_DRAW_RAFFLE
	data := map[string]interface{}{
		"address":     address,
		"reveal_seed": revealSeed,
	}
	return c.SignAndSendTransaction(c.chainId, from, to, method, data)
}

// ClaimRaffle allows a winner to claim their prize.
func (c *networkClient) ClaimRaffle(address, winner, tokenType, uuid string) (types.ContractOutput, error) {
	if address == "" {
		return types.ContractOutput{}, fmt.Errorf("address not set")
	}
	if err := keys.ValidateEDDSAPublicKey(address); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid address: %w", err)
	}
	if winner == "" {
		return types.ContractOutput{}, fmt.Errorf("winner not set")
	}
	if err := keys.ValidateEDDSAPublicKey(winner); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid winner address: %w", err)
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
	if from == "" {
		return types.ContractOutput{}, fmt.Errorf("from address not set")
	}
	if err := keys.ValidateEDDSAPublicKey(from); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid from address: %w", err)
	}

	to := address
	method := raffleV1.METHOD_CLAIM_RAFFLE
	data := map[string]interface{}{"address": address, "winner": winner}
	return c.SignAndSendTransaction(c.chainId, from, to, method, data)
}

// WithdrawRaffle withdraws unused/prize funds from the raffle pool.
func (c *networkClient) WithdrawRaffle(address, tokenAddress, amount, tokenType, uuid string) (types.ContractOutput, error) {
	if address == "" {
		return types.ContractOutput{}, fmt.Errorf("address not set")
	}
	if err := keys.ValidateEDDSAPublicKey(address); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid address: %w", err)
	}
	if tokenAddress == "" {
		return types.ContractOutput{}, fmt.Errorf("token address not set")
	}
	if err := keys.ValidateEDDSAPublicKey(tokenAddress); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid token address: %w", err)
	}
	if amount == "" {
		return types.ContractOutput{}, fmt.Errorf("amount not set")
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
	if from == "" {
		return types.ContractOutput{}, fmt.Errorf("from address not set")
	}
	if err := keys.ValidateEDDSAPublicKey(from); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid from address: %w", err)
	}

	to := address
	method := raffleV1.METHOD_WITHDRAW_RAFFLE
	data := map[string]interface{}{"address": address, "token_address": tokenAddress, "amount": amount}
	return c.SignAndSendTransaction(c.chainId, from, to, method, data)
}

func (c *networkClient) AddRafflePrize(raffleAddress string, tokenAddress string, amount string, tokenType string, uuid string) (types.ContractOutput, error) {
	if raffleAddress == "" {
		return types.ContractOutput{}, fmt.Errorf("raffle address not set")
	}
	if err := keys.ValidateEDDSAPublicKey(raffleAddress); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid raffle address: %w", err)
	}
	if tokenAddress == "" {
		return types.ContractOutput{}, fmt.Errorf("token address not set")
	}
	if err := keys.ValidateEDDSAPublicKey(tokenAddress); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid token address: %w", err)
	}
	if amount == "" {
		return types.ContractOutput{}, fmt.Errorf("amount not set")
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
	if from == "" {
		return types.ContractOutput{}, fmt.Errorf("from address not set")
	}
	if err := keys.ValidateEDDSAPublicKey(from); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid from address: %w", err)
	}

	to := raffleAddress
	method := raffleV1.METHOD_ADD_RAFFLE_PRIZE
	data := map[string]interface{}{"amount": amount, "raffle_address": raffleAddress, "token_address": tokenAddress}
	return c.SignAndSendTransaction(c.chainId, from, to, method, data)
}

func (c *networkClient) RemoveRafflePrize(raffleAddress string, tokenType string, uuid string) (types.ContractOutput, error) {
	if raffleAddress == "" {
		return types.ContractOutput{}, fmt.Errorf("raffle address not set")
	}
	if err := keys.ValidateEDDSAPublicKey(raffleAddress); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid raffle address: %w", err)
	}
	if tokenType == "" {
		return types.ContractOutput{}, fmt.Errorf("tokenType not set")
	}
	if uuid == "" {
		return types.ContractOutput{}, fmt.Errorf("uuid not set")
	}

	from := c.publicKey
	if from == "" {
		return types.ContractOutput{}, fmt.Errorf("from address not set")
	}
	if err := keys.ValidateEDDSAPublicKey(from); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid from address: %w", err)
	}

	to := raffleAddress
	method := raffleV1.METHOD_REMOVE_RAFFLE_PRIZE
	data := map[string]interface{}{"raffle_address": raffleAddress, "uuid": uuid}
	return c.SignAndSendTransaction(c.chainId, from, to, method, data)
}

// GetRaffle reads a single raffle state.
func (c *networkClient) GetRaffle(address string) (types.ContractOutput, error) {
	from := c.publicKey
	if from == "" {
		return types.ContractOutput{}, fmt.Errorf("from address not set")
	}
	if err := keys.ValidateEDDSAPublicKey(from); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid from address: %w", err)
	}
	if address == "" {
		return types.ContractOutput{}, fmt.Errorf("raffle address must be set")
	}
	if err := keys.ValidateEDDSAPublicKey(address); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid raffle address: %w", err)
	}

	method := raffleV1.METHOD_GET_RAFFLE

	return c.GetState(address, method, nil)
}

// ListRaffles queries raffles with filters + pagination.
func (c *networkClient) ListRaffles(owner, tokenAddress string, paused *bool, activeOnly *bool, page, limit int, asc bool) (types.ContractOutput, error) {
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
	if page < 1 {
		return types.ContractOutput{}, fmt.Errorf("page must be greater than 0")
	}
	if limit < 1 {
		return types.ContractOutput{}, fmt.Errorf("limit must be greater than 0")
	}

	method := raffleV1.METHOD_LIST_RAFFLES
	data := map[string]interface{}{
		"owner":            owner,
		"page":             page,
		"limit":            limit,
		"ascending":        asc,
		"token_address":    tokenAddress,
		"contract_version": raffleV1.RAFFLE_CONTRACT_V1,
	}
	if paused != nil {
		data["paused"] = *paused
	}
	if activeOnly != nil {
		data["active_only"] = *activeOnly
	}

	return c.GetState("", method, data)
}

func (c *networkClient) ListPrizes(raffleAddress string, page, limit int, asc bool) (types.ContractOutput, error) {
	from := c.publicKey
	if from == "" {
		return types.ContractOutput{}, fmt.Errorf("from address not set")
	}
	if err := keys.ValidateEDDSAPublicKey(from); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid from address: %w", err)
	}
	if raffleAddress == "" {
		return types.ContractOutput{}, fmt.Errorf("raffle address must be set")
	}
	if err := keys.ValidateEDDSAPublicKey(raffleAddress); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid raffle address: %w", err)
	}
	if page < 1 {
		return types.ContractOutput{}, fmt.Errorf("page must be greater than 0")
	}
	if limit < 1 {
		return types.ContractOutput{}, fmt.Errorf("limit must be greater than 0")
	}

	method := raffleV1.METHOD_LIST_PRIZES
	data := map[string]interface{}{
		"raffle_address":   raffleAddress,
		"page":             page,
		"limit":            limit,
		"ascending":        asc,
		"contract_version": raffleV1.RAFFLE_CONTRACT_V1,
	}

	return c.GetState("", method, data)
}
