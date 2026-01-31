package client_2finance


import (
	"fmt"
	"time"

	"gitlab.com/2finance/2finance-network/blockchain/contract/reviewV1"
	"gitlab.com/2finance/2finance-network/blockchain/encryption/keys"
	"gitlab.com/2finance/2finance-network/blockchain/types"
)


// AddReview creates a new review (to = DEPLOY address). The tx sender (c.publicKey) becomes owner.
func (c *networkClient) AddReview(
	address, reviewer, reviewee, subjectType, subjectID string,
	rating int,
	comment string,
	tags map[string]string,
	mediaHashes []string,
	startAt, expiredAt time.Time,
	hidden bool,
) (types.ContractOutput, error) {
	from := c.publicKey
	if err := keys.ValidateEDDSAPublicKeyHex(from); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid from address: %w", err)
	}

	if err := keys.ValidateEDDSAPublicKeyHex(address); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid review address: %w", err)
	}

	if reviewer == "" {
		return types.ContractOutput{}, fmt.Errorf("reviewer not set")
	}
	if err := keys.ValidateEDDSAPublicKeyHex(reviewer); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid reviewer address: %w", err)
	}
	if reviewee == "" {
		return types.ContractOutput{}, fmt.Errorf("reviewee not set")
	}
	if err := keys.ValidateEDDSAPublicKeyHex(reviewee); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid reviewee address: %w", err)
	}
	if subjectType == "" {
		return types.ContractOutput{}, fmt.Errorf("subject_type not set")
	}
	if subjectID == "" {
		return types.ContractOutput{}, fmt.Errorf("subject_id not set")
	}
	if rating < 1 || rating > 5 {
		return types.ContractOutput{}, fmt.Errorf("rating must be between 1 and 5")
	}
	if startAt.IsZero() {
		return types.ContractOutput{}, fmt.Errorf("start_at not set")
	}
	if expiredAt.IsZero() {
		return types.ContractOutput{}, fmt.Errorf("expired_at not set")
	}

	to := address
	method := reviewV1.METHOD_ADD_REVIEW

	data := map[string]interface{}{
		"address":       address,
		"reviewer":      reviewer,
		"reviewee":      reviewee,
		"subject_type":  subjectType,
		"subject_id":    subjectID,
		"rating":        rating,
		"comment":       comment,
		"tags":          tags,
		"media_hashes":  mediaHashes,
		"start_at":      startAt,
		"expired_at":    expiredAt,
		"hidden":        hidden,
	}

	return c.SignAndSendTransaction(c.chainId, from, to, method, data)
}

// UpdateReview modifies fields of an existing review.
func (c *networkClient) UpdateReview(
	address, subjectType, subjectID string,
	rating int,
	comment string,
	tags map[string]string,
	mediaHashes []string,
	startAt, expiredAt *time.Time,
) (types.ContractOutput, error) {
	if address == "" {
		return types.ContractOutput{}, fmt.Errorf("address not set")
	}
	if err := keys.ValidateEDDSAPublicKeyHex(address); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid address: %w", err)
	}

	from := c.publicKey

	if err := keys.ValidateEDDSAPublicKeyHex(from); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid from address: %w", err)
	}

	if subjectType == "" {
		return types.ContractOutput{}, fmt.Errorf("subject_type not set")
	}
	if subjectID == "" {
		return types.ContractOutput{}, fmt.Errorf("subject_id not set")
	}
	if rating != 0 && (rating < 1 || rating > 5) { // allow 0 to mean "no change"
		return types.ContractOutput{}, fmt.Errorf("rating must be between 1 and 5")
	}

	to := address
	method := reviewV1.METHOD_UPDATE_REVIEW

	data := map[string]interface{}{
		"address":      address,
		"subject_type": subjectType,
		"subject_id":   subjectID,
		"rating":       rating,
		"comment":      comment,
		"tags":         tags,
		"media_hashes": mediaHashes,
	}
	if startAt != nil {
		data["start_at"] = *startAt
	}
	if expiredAt != nil {
		data["expired_at"] = *expiredAt
	}

	return c.SignAndSendTransaction(c.chainId, from, to, method, data)
}

// HideReview toggles the hidden state. OnlyOwner.
func (c *networkClient) HideReview(address string, hidden bool) (types.ContractOutput, error) {
	if address == "" {
		return types.ContractOutput{}, fmt.Errorf("address not set")
	}
	if err := keys.ValidateEDDSAPublicKeyHex(address); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid address: %w", err)
	}

	from := c.publicKey

	if err := keys.ValidateEDDSAPublicKeyHex(from); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid from address: %w", err)
	}

	to := address
	method := reviewV1.METHOD_HIDE_REVIEW
	data := map[string]interface{}{
		"address": address,
		"hidden":  hidden,
	}

	return c.SignAndSendTransaction(c.chainId, from, to, method, data)
}

// VoteHelpful registers an up/down helpful vote for a review.
func (c *networkClient) VoteHelpful(address, voter string, isHelpful bool) (types.ContractOutput, error) {
	if address == "" {
		return types.ContractOutput{}, fmt.Errorf("address not set")
	}
	if err := keys.ValidateEDDSAPublicKeyHex(address); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid address: %w", err)
	}
	if voter == "" {
		return types.ContractOutput{}, fmt.Errorf("voter not set")
	}
	if err := keys.ValidateEDDSAPublicKeyHex(voter); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid voter address: %w", err)
	}

	from := c.publicKey

	if err := keys.ValidateEDDSAPublicKeyHex(from); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid from address: %w", err)
	}

	to := address
	method := reviewV1.METHOD_VOTE_HELPFUL
	data := map[string]interface{}{
		"address":    address,
		"voter":      voter,
		"is_helpful": isHelpful,
	}

	return c.SignAndSendTransaction(c.chainId, from, to, method, data)
}

// ReportReview flags a review with a reason string by a reporter.
func (c *networkClient) ReportReview(address, reporter, reason string) (types.ContractOutput, error) {
	if address == "" {
		return types.ContractOutput{}, fmt.Errorf("address not set")
	}
	if err := keys.ValidateEDDSAPublicKeyHex(address); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid address: %w", err)
	}
	if reporter == "" {
		return types.ContractOutput{}, fmt.Errorf("reporter not set")
	}
	if err := keys.ValidateEDDSAPublicKeyHex(reporter); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid reporter address: %w", err)
	}
	if reason == "" {
		return types.ContractOutput{}, fmt.Errorf("reason not set")
	}

	from := c.publicKey

	if err := keys.ValidateEDDSAPublicKeyHex(from); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid from address: %w", err)
	}

	to := address
	method := reviewV1.METHOD_REPORT_REVIEW
	data := map[string]interface{}{
		"address":  address,
		"reporter": reporter,
		"reason":   reason,
	}

	return c.SignAndSendTransaction(c.chainId, from, to, method, data)
}

// ModerateReview applies a moderation action (e.g., approve/reject/remove) with an optional note. OnlyModerator/Owner per contract rules.
func (c *networkClient) ModerateReview(address, action, note string) (types.ContractOutput, error) {
	if address == "" {
		return types.ContractOutput{}, fmt.Errorf("address not set")
	}
	if err := keys.ValidateEDDSAPublicKeyHex(address); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid address: %w", err)
	}
	if action == "" {
		return types.ContractOutput{}, fmt.Errorf("action not set")
	}

	from := c.publicKey

	if err := keys.ValidateEDDSAPublicKeyHex(from); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid from address: %w", err)
	}

	to := address
	method := reviewV1.METHOD_MODERATE_REVIEW
	data := map[string]interface{}{
		"address": address,
		"action":  action,
		"note":    note,
	}

	return c.SignAndSendTransaction(c.chainId, from, to, method, data)
}

// GetReview retrieves a single review state.
func (c *networkClient) GetReview(address string) (types.ContractOutput, error) {
	from := c.publicKey

	if err := keys.ValidateEDDSAPublicKeyHex(from); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid from address: %w", err)
	}
	if address == "" {
		return types.ContractOutput{}, fmt.Errorf("review address must be set")
	}
	if err := keys.ValidateEDDSAPublicKeyHex(address); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid review address: %w", err)
	}

	method := reviewV1.METHOD_GET_REVIEW

	return c.GetState(address, method, nil)
}

// ListReviews queries reviews with filters + pagination.
func (c *networkClient) ListReviews(
	owner,
	reviewer, reviewee, subjectType, subjectID string,
	includeHidden *bool,
	minRating, maxRating, page, limit int,
	asc bool,
) (types.ContractOutput, error) {
	from := c.publicKey

	if err := keys.ValidateEDDSAPublicKeyHex(from); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid from address: %w", err)
	}

	// Optional address validations
	if owner != "" {
		if err := keys.ValidateEDDSAPublicKeyHex(owner); err != nil {
			return types.ContractOutput{}, fmt.Errorf("invalid owner address: %w", err)
		}
	}
	if reviewer != "" {
		if err := keys.ValidateEDDSAPublicKeyHex(reviewer); err != nil {
			return types.ContractOutput{}, fmt.Errorf("invalid reviewer address: %w", err)
		}
	}
	if reviewee != "" {
		if err := keys.ValidateEDDSAPublicKeyHex(reviewee); err != nil {
			return types.ContractOutput{}, fmt.Errorf("invalid reviewee address: %w", err)
		}
	}

	if page < 1 {
		return types.ContractOutput{}, fmt.Errorf("page must be greater than 0")
	}
	if limit < 1 {
		return types.ContractOutput{}, fmt.Errorf("limit must be greater than 0")
	}
	if minRating < 0 || minRating > 5 {
		return types.ContractOutput{}, fmt.Errorf("min_rating must be between 0 and 5")
	}
	if maxRating < 0 || maxRating > 5 {
		return types.ContractOutput{}, fmt.Errorf("max_rating must be between 0 and 5")
	}
	if maxRating != 0 && minRating > maxRating {
		return types.ContractOutput{}, fmt.Errorf("min_rating cannot be greater than max_rating")
	}

	method := reviewV1.METHOD_LIST_REVIEWS
	data := map[string]interface{}{
		"reviewer":      reviewer,
		"reviewee":      reviewee,
		"subject_type":  subjectType,
		"subject_id":    subjectID,
		"min_rating":    minRating,
		"max_rating":    maxRating,
		"page":          page,
		"limit":         limit,
		"ascending":     asc,
		"contract_version": reviewV1.REVIEW_CONTRACT_V1,
	}
	if includeHidden != nil {
		data["include_hidden"] = *includeHidden
	}

	return c.GetState("", method, data)
}
