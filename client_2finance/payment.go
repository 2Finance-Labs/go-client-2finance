package client_2finance

import (
	"fmt"

	"gitlab.com/2finance/2finance-network/blockchain/contract/paymentV1"
	"gitlab.com/2finance/2finance-network/blockchain/contract/paymentV1/inputs"
	"gitlab.com/2finance/2finance-network/blockchain/encryption/keys"
	"gitlab.com/2finance/2finance-network/blockchain/types"
	"gitlab.com/2finance/2finance-network/blockchain/utils"
)

// CreatePayment creates a new payment intent.
func (c *networkClient) CreatePayment(in inputs.InputCreate) (types.ContractOutput, error) {
	from := c.publicKey

	if err := keys.ValidateEDDSAPublicKeyHex(from); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid from address: %w", err)
	}

	if in.Address == "" {
		return types.ContractOutput{}, fmt.Errorf("address not set")
	}
	if err := keys.ValidateEDDSAPublicKeyHex(in.Address); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid address: %w", err)
	}

	if in.Owner == "" {
		return types.ContractOutput{}, fmt.Errorf("owner not set")
	}
	if err := keys.ValidateEDDSAPublicKeyHex(in.Owner); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid owner address: %w", err)
	}

	if in.TokenAddress == "" {
		return types.ContractOutput{}, fmt.Errorf("token address not set")
	}
	if err := keys.ValidateEDDSAPublicKeyHex(in.TokenAddress); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid token address: %w", err)
	}

	if in.Payer == "" {
		return types.ContractOutput{}, fmt.Errorf("payer not set")
	}
	if err := keys.ValidateEDDSAPublicKeyHex(in.Payer); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid payer address: %w", err)
	}

	if in.Payee == "" {
		return types.ContractOutput{}, fmt.Errorf("payee not set")
	}
	if err := keys.ValidateEDDSAPublicKeyHex(in.Payee); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid payee address: %w", err)
	}

	if in.Payer == in.Payee {
		return types.ContractOutput{}, fmt.Errorf("payee and payer cannot be the same: %s - %s", in.Payee, in.Payer)
	}

	if in.OrderId == "" {
		return types.ContractOutput{}, fmt.Errorf("order_id not set")
	}
	if in.Amount == "" {
		return types.ContractOutput{}, fmt.Errorf("amount not set")
	}
	if in.ExpiredAt.IsZero() {
		return types.ContractOutput{}, fmt.Errorf("expired_at not set")
	}

	to := in.Address
	method := paymentV1.METHOD_CREATE_PAYMENT

	data := map[string]interface{}{
		"address":       in.Address,
		"owner":         in.Owner,
		"token_address": in.TokenAddress,
		"order_id":      in.OrderId,
		"payer":         in.Payer,
		"payee":         in.Payee,
		"amount":        in.Amount,
		"expired_at":    in.ExpiredAt,
	}

	version := uint8(1)
	uuid7, err := utils.NewUUID7()
	if err != nil {
		return types.ContractOutput{}, fmt.Errorf("failed to generate UUIDv7: %w", err)
	}

	return c.SignAndSendTransaction(c.chainId, from, to, method, data, version, uuid7)
}

// DirectPay creates + authorizes + captures in one step.
func (c *networkClient) DirectPay(in inputs.InputDirectPay) (types.ContractOutput, error) {
	from := c.publicKey

	if err := keys.ValidateEDDSAPublicKeyHex(from); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid from address: %w", err)
	}

	if in.Address == "" {
		return types.ContractOutput{}, fmt.Errorf("address not set")
	}
	if err := keys.ValidateEDDSAPublicKeyHex(in.Address); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid address: %w", err)
	}

	if in.Owner == "" {
		return types.ContractOutput{}, fmt.Errorf("owner not set")
	}
	if err := keys.ValidateEDDSAPublicKeyHex(in.Owner); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid owner address: %w", err)
	}

	if in.TokenAddress == "" {
		return types.ContractOutput{}, fmt.Errorf("token address not set")
	}
	if err := keys.ValidateEDDSAPublicKeyHex(in.TokenAddress); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid token address: %w", err)
	}

	if in.Payer == "" {
		return types.ContractOutput{}, fmt.Errorf("payer not set")
	}
	if err := keys.ValidateEDDSAPublicKeyHex(in.Payer); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid payer address: %w", err)
	}

	if in.Payee == "" {
		return types.ContractOutput{}, fmt.Errorf("payee not set")
	}
	if err := keys.ValidateEDDSAPublicKeyHex(in.Payee); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid payee address: %w", err)
	}

	if in.Payer == in.Payee {
		return types.ContractOutput{}, fmt.Errorf("payee and payer cannot be the same: %s - %s", in.Payee, in.Payer)
	}

	if in.OrderId == "" {
		return types.ContractOutput{}, fmt.Errorf("order_id not set")
	}
	if in.Amount == "" {
		return types.ContractOutput{}, fmt.Errorf("amount not set")
	}
	if in.ExpiredAt.IsZero() {
		return types.ContractOutput{}, fmt.Errorf("expired_at not set")
	}

	to := in.Address
	method := paymentV1.METHOD_DIRECT_PAY

	data := map[string]interface{}{
		"address":       in.Address,
		"owner":         in.Owner,
		"token_address": in.TokenAddress,
		"order_id":      in.OrderId,
		"payer":         in.Payer,
		"payee":         in.Payee,
		"amount":        in.Amount,
		"expired_at":    in.ExpiredAt,
	}

	version := uint8(1)
	uuid7, err := utils.NewUUID7()
	if err != nil {
		return types.ContractOutput{}, fmt.Errorf("failed to generate UUIDv7: %w", err)
	}

	return c.SignAndSendTransaction(c.chainId, from, to, method, data, version, uuid7)
}

// AuthorizePayment places a hold on funds.
func (c *networkClient) AuthorizePayment(in inputs.InputAuthorize) (types.ContractOutput, error) {
	if in.Address == "" {
		return types.ContractOutput{}, fmt.Errorf("address not set")
	}
	if err := keys.ValidateEDDSAPublicKeyHex(in.Address); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid address: %w", err)
	}

	from := c.publicKey
	if err := keys.ValidateEDDSAPublicKeyHex(from); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid from address: %w", err)
	}

	to := in.Address
	method := paymentV1.METHOD_AUTHORIZE_PAYMENT
	data := map[string]interface{}{
		"address": in.Address,
	}

	version := uint8(1)
	uuid7, err := utils.NewUUID7()
	if err != nil {
		return types.ContractOutput{}, fmt.Errorf("failed to generate UUIDv7: %w", err)
	}

	return c.SignAndSendTransaction(c.chainId, from, to, method, data, version, uuid7)
}

// CapturePayment settles funds.
func (c *networkClient) CapturePayment(in inputs.InputCapture) (types.ContractOutput, error) {
	if in.Address == "" {
		return types.ContractOutput{}, fmt.Errorf("address not set")
	}
	if err := keys.ValidateEDDSAPublicKeyHex(in.Address); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid address: %w", err)
	}

	from := c.publicKey
	if err := keys.ValidateEDDSAPublicKeyHex(from); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid from address: %w", err)
	}

	to := in.Address
	method := paymentV1.METHOD_CAPTURE_PAYMENT
	data := map[string]interface{}{
		"address": in.Address,
	}

	version := uint8(1)
	uuid7, err := utils.NewUUID7()
	if err != nil {
		return types.ContractOutput{}, fmt.Errorf("failed to generate UUIDv7: %w", err)
	}

	return c.SignAndSendTransaction(c.chainId, from, to, method, data, version, uuid7)
}

// RefundPayment returns funds from payee back to payer.
func (c *networkClient) RefundPayment(in inputs.InputRefund) (types.ContractOutput, error) {
	if in.Address == "" {
		return types.ContractOutput{}, fmt.Errorf("address not set")
	}
	if err := keys.ValidateEDDSAPublicKeyHex(in.Address); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid address: %w", err)
	}
	if in.Amount == "" {
		return types.ContractOutput{}, fmt.Errorf("amount not set")
	}

	from := c.publicKey
	if err := keys.ValidateEDDSAPublicKeyHex(from); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid from address: %w", err)
	}

	to := in.Address
	method := paymentV1.METHOD_REFUND_PAYMENT
	data := map[string]interface{}{
		"address": in.Address,
		"amount":  in.Amount,
	}

	version := uint8(1)
	uuid7, err := utils.NewUUID7()
	if err != nil {
		return types.ContractOutput{}, fmt.Errorf("failed to generate UUIDv7: %w", err)
	}

	return c.SignAndSendTransaction(c.chainId, from, to, method, data, version, uuid7)
}

// VoidPayment releases an authorization hold.
func (c *networkClient) VoidPayment(in inputs.InputVoidPayment) (types.ContractOutput, error) {
	if in.Address == "" {
		return types.ContractOutput{}, fmt.Errorf("address not set")
	}
	if err := keys.ValidateEDDSAPublicKeyHex(in.Address); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid address: %w", err)
	}

	from := c.publicKey
	if err := keys.ValidateEDDSAPublicKeyHex(from); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid from address: %w", err)
	}

	to := in.Address
	method := paymentV1.METHOD_VOID_PAYMENT
	data := map[string]interface{}{
		"address": in.Address,
	}

	version := uint8(1)
	uuid7, err := utils.NewUUID7()
	if err != nil {
		return types.ContractOutput{}, fmt.Errorf("failed to generate UUIDv7: %w", err)
	}

	return c.SignAndSendTransaction(c.chainId, from, to, method, data, version, uuid7)
}

// PausePayment toggles paused=true.
func (c *networkClient) PausePayment(in inputs.InputPause) (types.ContractOutput, error) {
	if in.Address == "" {
		return types.ContractOutput{}, fmt.Errorf("address not set")
	}
	if err := keys.ValidateEDDSAPublicKeyHex(in.Address); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid address: %w", err)
	}
	if !in.Paused {
		return types.ContractOutput{}, fmt.Errorf("paused must be true: Pause: %t", in.Paused)
	}

	from := c.publicKey
	if err := keys.ValidateEDDSAPublicKeyHex(from); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid from address: %w", err)
	}

	to := in.Address
	method := paymentV1.METHOD_PAUSE_PAYMENT
	data := map[string]interface{}{
		"address": in.Address,
		"paused":  in.Paused,
	}

	version := uint8(1)
	uuid7, err := utils.NewUUID7()
	if err != nil {
		return types.ContractOutput{}, fmt.Errorf("failed to generate UUIDv7: %w", err)
	}

	return c.SignAndSendTransaction(c.chainId, from, to, method, data, version, uuid7)
}

// UnpausePayment toggles paused=false.
func (c *networkClient) UnpausePayment(in inputs.InputPause) (types.ContractOutput, error) {
	if in.Address == "" {
		return types.ContractOutput{}, fmt.Errorf("address not set")
	}
	if err := keys.ValidateEDDSAPublicKeyHex(in.Address); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid address: %w", err)
	}
	if in.Paused {
		return types.ContractOutput{}, fmt.Errorf("paused must be false: Pause: %t", in.Paused)
	}

	from := c.publicKey
	if err := keys.ValidateEDDSAPublicKeyHex(from); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid from address: %w", err)
	}

	to := in.Address
	method := paymentV1.METHOD_UNPAUSE_PAYMENT
	data := map[string]interface{}{
		"address": in.Address,
		"paused":  in.Paused,
	}

	version := uint8(1)
	uuid7, err := utils.NewUUID7()
	if err != nil {
		return types.ContractOutput{}, fmt.Errorf("failed to generate UUIDv7: %w", err)
	}

	return c.SignAndSendTransaction(c.chainId, from, to, method, data, version, uuid7)
}

// GetPayment reads a single payment state.
func (c *networkClient) GetPayment(address string) (types.ContractOutput, error) {
	from := c.publicKey

	if err := keys.ValidateEDDSAPublicKeyHex(from); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid from address: %w", err)
	}

	if address == "" {
		return types.ContractOutput{}, fmt.Errorf("payment address must be set")
	}
	if err := keys.ValidateEDDSAPublicKeyHex(address); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid payment address: %w", err)
	}

	method := paymentV1.METHOD_GET_PAYMENT
	return c.GetState(address, method, nil)
}

// ListPayments queries payments with filters + pagination.
func (c *networkClient) ListPayments(in inputs.InputList) (types.ContractOutput, error) {
	from := c.publicKey

	if err := keys.ValidateEDDSAPublicKeyHex(from); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid from address: %w", err)
	}

	if in.TokenAddress != "" {
		if err := keys.ValidateEDDSAPublicKeyHex(in.TokenAddress); err != nil {
			return types.ContractOutput{}, fmt.Errorf("invalid token address: %w", err)
		}
	}
	if in.Payer != "" {
		if err := keys.ValidateEDDSAPublicKeyHex(in.Payer); err != nil {
			return types.ContractOutput{}, fmt.Errorf("invalid payer address: %w", err)
		}
	}
	if in.Payee != "" {
		if err := keys.ValidateEDDSAPublicKeyHex(in.Payee); err != nil {
			return types.ContractOutput{}, fmt.Errorf("invalid payee address: %w", err)
		}
	}
	if in.Page < 1 {
		return types.ContractOutput{}, fmt.Errorf("page must be greater than 0")
	}
	if in.Limit < 1 {
		return types.ContractOutput{}, fmt.Errorf("limit must be greater than 0")
	}

	method := paymentV1.METHOD_LIST_PAYMENTS
	data := map[string]interface{}{
		"order_id":         in.OrderId,
		"token_address":    in.TokenAddress,
		"status":           in.Status,
		"payer":            in.Payer,
		"payee":            in.Payee,
		"page":             in.Page,
		"limit":            in.Limit,
		"ascending":        in.Ascending,
		"contract_version": paymentV1.PAYMENT_CONTRACT_V1,
	}

	return c.GetState("", method, data)
}