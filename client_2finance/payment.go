package client_2finance


import (
	"fmt"

	"gitlab.com/2finance/2finance-network/blockchain/contract/paymentV1"
	"gitlab.com/2finance/2finance-network/blockchain/encryption/keys"
	"gitlab.com/2finance/2finance-network/blockchain/types"
	"time"
)

// CreatePayment creates a new payment intent (to = DEPLOY address).
// Server/contract treats the tx sender (c.publicKey) as the owner.
func (c *networkClient) CreatePayment(
	address string,
	tokenAddress string, // ERC-20-like token on your chain
	orderId string,
	payer string,
	payee string,
	amount string, // integer string
	expiredAt time.Time,
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
	if err := keys.ValidateEDDSAPublicKey(address); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid address: %w", err)
	}

	if tokenAddress == "" {
		return types.ContractOutput{}, fmt.Errorf("token address not set")
	}
	if err := keys.ValidateEDDSAPublicKey(tokenAddress); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid token address: %w", err)
	}
	if payer == "" {
		return types.ContractOutput{}, fmt.Errorf("payer not set")
	}
	if err := keys.ValidateEDDSAPublicKey(payer); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid payer address: %w", err)
	}
	if payee == "" {
		return types.ContractOutput{}, fmt.Errorf("payee not set")
	}
	if err := keys.ValidateEDDSAPublicKey(payee); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid payee address: %w", err)
	}

	if payer == payee {
		return types.ContractOutput{}, fmt.Errorf("payee and payer cannot be the same: %s - %s", payee, payer)
	}
	
	if orderId == "" {
		return types.ContractOutput{}, fmt.Errorf("order_id not set")
	}
	if amount == "" {
		return types.ContractOutput{}, fmt.Errorf("amount not set")
	}
	if expiredAt.IsZero() {
		return types.ContractOutput{}, fmt.Errorf("expired_at not set")
	}

	to := address
	contractVersion := paymentV1.PAYMENT_CONTRACT_V1
	method := paymentV1.METHOD_CREATE_PAYMENT

	data := map[string]interface{}{
		"address":       address,
		"token_address": tokenAddress,
		"order_id":      orderId,
		"payer":         payer,
		"payee":         payee,
		"amount":        amount,
		"expired_at":   expiredAt,
	}

	return c.SignAndSendTransaction(from, to, contractVersion, method, data)
}

// DirectPay is a one-step convenience: create + immediate capture.
func (c *networkClient) DirectPay(
	address string,
	tokenAddress string,
	orderId string,
	payer string,
	payee string,
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
	if err := keys.ValidateEDDSAPublicKey(address); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid address: %w", err)
	}

	if tokenAddress == "" {
		return types.ContractOutput{}, fmt.Errorf("token address not set")
	}
	if err := keys.ValidateEDDSAPublicKey(tokenAddress); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid token address: %w", err)
	}
	if payer == "" {
		return types.ContractOutput{}, fmt.Errorf("payer not set")
	}
	if err := keys.ValidateEDDSAPublicKey(payer); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid payer address: %w", err)
	}
	if payee == "" {
		return types.ContractOutput{}, fmt.Errorf("payee not set")
	}
	if err := keys.ValidateEDDSAPublicKey(payee); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid payee address: %w", err)
	}
	if orderId == "" {
		return types.ContractOutput{}, fmt.Errorf("order_id not set")
	}
	if amount == "" {
		return types.ContractOutput{}, fmt.Errorf("amount not set")
	}

	to := address
	contractVersion := paymentV1.PAYMENT_CONTRACT_V1
	method := paymentV1.METHOD_DIRECT_PAY

	data := map[string]interface{}{
		"address":       address,
		"token_address": tokenAddress,
		"order_id":      orderId,
		"payer":         payer,
		"payee":         payee,
		"amount":        amount,
	}

	return c.SignAndSendTransaction(from, to, contractVersion, method, data)
}

// AuthorizePayment places a hold on funds (payer -> payee) for a payment address.
func (c *networkClient) AuthorizePayment(address string) (types.ContractOutput, error) {
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
	contractVersion := paymentV1.PAYMENT_CONTRACT_V1
	method := paymentV1.METHOD_AUTHORIZE_PAYMENT
	data := map[string]interface{}{
		"address": address,
	}

	return c.SignAndSendTransaction(from, to, contractVersion, method, data)
}

// CapturePayment settles funds (full/partial).
func (c *networkClient) CapturePayment(address string) (types.ContractOutput, error) {
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
	contractVersion := paymentV1.PAYMENT_CONTRACT_V1
	method := paymentV1.METHOD_CAPTURE_PAYMENT
	data := map[string]interface{}{
		"address": address,
	}

	return c.SignAndSendTransaction(from, to, contractVersion, method, data)
}

// RefundPayment returns funds (full/partial) from payee back to payer.
func (c *networkClient) RefundPayment(address, amount string) (types.ContractOutput, error) {
	if address == "" {
		return types.ContractOutput{}, fmt.Errorf("address not set")
	}
	if err := keys.ValidateEDDSAPublicKey(address); err != nil {
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

	to := address
	contractVersion := paymentV1.PAYMENT_CONTRACT_V1
	method := paymentV1.METHOD_REFUND_PAYMENT
	data := map[string]interface{}{
		"address": address,
		"amount":  amount,
	}

	return c.SignAndSendTransaction(from, to, contractVersion, method, data)
}

// VoidPayment releases an authorization hold.
func (c *networkClient) VoidPayment(address string) (types.ContractOutput, error) {
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
	contractVersion := paymentV1.PAYMENT_CONTRACT_V1
	method := paymentV1.METHOD_VOID_PAYMENT
	data := map[string]interface{}{
		"address": address,
	}

	return c.SignAndSendTransaction(from, to, contractVersion, method, data)
}

// PausePayment toggles the paused state to true. OnlyOwner.
func (c *networkClient) PausePayment(address string, paused bool) (types.ContractOutput, error) {
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
	contractVersion := paymentV1.PAYMENT_CONTRACT_V1
	method := paymentV1.METHOD_PAUSE_PAYMENT
	data := map[string]interface{}{
		"address": address,
		"paused":  paused,
	}

	return c.SignAndSendTransaction(from, to, contractVersion, method, data)
}

// UnpausePayment toggles the paused state to false. OnlyOwner.
func (c *networkClient) UnpausePayment(address string, paused bool) (types.ContractOutput, error) {
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
	contractVersion := paymentV1.PAYMENT_CONTRACT_V1
	method := paymentV1.METHOD_UNPAUSE_PAYMENT
	data := map[string]interface{}{
		"address": address,
		"paused":  paused,
	}

	return c.SignAndSendTransaction(from, to, contractVersion, method, data)
}

// GetPayment reads a single payment state.
func (c *networkClient) GetPayment(address string) (types.ContractOutput, error) {
	from := c.publicKey
	if from == "" {
		return types.ContractOutput{}, fmt.Errorf("from address not set")
	}
	if err := keys.ValidateEDDSAPublicKey(from); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid from address: %w", err)
	}

	if address == "" {
		return types.ContractOutput{}, fmt.Errorf("payment address must be set")
	}
	if err := keys.ValidateEDDSAPublicKey(address); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid payment address: %w", err)
	}

	contractVersion := paymentV1.PAYMENT_CONTRACT_V1
	method := paymentV1.METHOD_GET_PAYMENT
	data := map[string]interface{}{
		"address": address,
	}

	return c.GetState(contractVersion, method, data)
}

//payer, payee, orderId, tokenAddress string, status []string, page, limit int, ascending bool
// ListPayments queries payments with filters + pagination.
func (c *networkClient) ListPayments(
	payer string,
	payee string,
	orderId string,
	tokenAddress string,
	status []string,
	page int,
	limit int,
	ascending bool,
) (types.ContractOutput, error) {
	
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
	if payer != "" {
		if err := keys.ValidateEDDSAPublicKey(payer); err != nil {
			return types.ContractOutput{}, fmt.Errorf("invalid payer address: %w", err)
		}
	}
	if payee != "" {
		if err := keys.ValidateEDDSAPublicKey(payee); err != nil {
			return types.ContractOutput{}, fmt.Errorf("invalid payee address: %w", err)
		}
	}
	if page < 1 {
		return types.ContractOutput{}, fmt.Errorf("page must be greater than 0")
	}
	if limit < 1 {
		return types.ContractOutput{}, fmt.Errorf("limit must be greater than 0")
	}

	contractVersion := paymentV1.PAYMENT_CONTRACT_V1
	method := paymentV1.METHOD_LIST_PAYMENTS
	data := map[string]interface{}{
		"token_address": tokenAddress,
		"status":        status,
		"payer":         payer,
		"payee":         payee,
		"page":          page,
		"limit":         limit,
		"ascending":     ascending,
	}

	return c.GetState(contractVersion, method, data)
}
