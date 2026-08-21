package client_2finance

import (
	"encoding/json"
	"fmt"

	"gitlab.com/2finance/2finance-network/blockchain/contract/fxLifecycleV2"
	inputsFXLifecycleV2 "gitlab.com/2finance/2finance-network/blockchain/contract/fxLifecycleV2/inputs"
	"gitlab.com/2finance/2finance-network/blockchain/contract/lifecycleCommonV2"
	"gitlab.com/2finance/2finance-network/blockchain/contract/multiCurrencyLifecycleV2"
	"gitlab.com/2finance/2finance-network/blockchain/contract/onboardingLifecycleV2"
	"gitlab.com/2finance/2finance-network/blockchain/contract/receivingLifecycleV2"
	"gitlab.com/2finance/2finance-network/blockchain/contract/sendingLifecycleV2"
	"gitlab.com/2finance/2finance-network/blockchain/encryption/keys"
	"gitlab.com/2finance/2finance-network/blockchain/types"
	"gitlab.com/2finance/2finance-network/blockchain/utils"
)

func (c *NetworkClient) StartFX(in inputsFXLifecycleV2.InputStartFX) (types.ContractOutput, error) {
	return c.sendLifecycleTransaction(in.Address, fxLifecycleV2.METHOD_START_FX, in)
}

func (c *NetworkClient) AdvanceFX(in inputsFXLifecycleV2.InputAdvanceFX) (types.ContractOutput, error) {
	return c.sendLifecycleTransaction(in.Address, fxLifecycleV2.METHOD_ADVANCE_FX, in)
}

func (c *NetworkClient) FailFX(in inputsFXLifecycleV2.InputFailFX) (types.ContractOutput, error) {
	return c.sendLifecycleTransaction(in.Address, fxLifecycleV2.METHOD_FAIL_FX, in)
}

func (c *NetworkClient) GetFX(address, requestID string) (types.ContractOutput, error) {
	return c.getLifecycleState(address, requestID, fxLifecycleV2.METHOD_GET_FX)
}

func (c *NetworkClient) StartOnboarding(in lifecycleCommonV2.StartInput) (types.ContractOutput, error) {
	return c.sendLifecycleTransaction(in.Address, onboardingLifecycleV2.METHOD_START_ONBOARDING, in)
}

func (c *NetworkClient) AdvanceOnboarding(in lifecycleCommonV2.AdvanceInput) (types.ContractOutput, error) {
	return c.sendLifecycleTransaction(in.Address, onboardingLifecycleV2.METHOD_ADVANCE_ONBOARDING, in)
}

func (c *NetworkClient) FailOnboarding(in lifecycleCommonV2.FailInput) (types.ContractOutput, error) {
	return c.sendLifecycleTransaction(in.Address, onboardingLifecycleV2.METHOD_FAIL_ONBOARDING, in)
}

func (c *NetworkClient) GetOnboarding(address, requestID string) (types.ContractOutput, error) {
	return c.getLifecycleState(address, requestID, onboardingLifecycleV2.METHOD_GET_ONBOARDING)
}

func (c *NetworkClient) StartReceiving(in lifecycleCommonV2.StartInput) (types.ContractOutput, error) {
	return c.sendLifecycleTransaction(in.Address, receivingLifecycleV2.METHOD_START_RECEIVING, in)
}

func (c *NetworkClient) AdvanceReceiving(in lifecycleCommonV2.AdvanceInput) (types.ContractOutput, error) {
	return c.sendLifecycleTransaction(in.Address, receivingLifecycleV2.METHOD_ADVANCE_RECEIVING, in)
}

func (c *NetworkClient) FailReceiving(in lifecycleCommonV2.FailInput) (types.ContractOutput, error) {
	return c.sendLifecycleTransaction(in.Address, receivingLifecycleV2.METHOD_FAIL_RECEIVING, in)
}

func (c *NetworkClient) GetReceiving(address, requestID string) (types.ContractOutput, error) {
	return c.getLifecycleState(address, requestID, receivingLifecycleV2.METHOD_GET_RECEIVING)
}

func (c *NetworkClient) StartSending(in lifecycleCommonV2.StartInput) (types.ContractOutput, error) {
	return c.sendLifecycleTransaction(in.Address, sendingLifecycleV2.METHOD_START_SENDING, in)
}

func (c *NetworkClient) AdvanceSending(in lifecycleCommonV2.AdvanceInput) (types.ContractOutput, error) {
	return c.sendLifecycleTransaction(in.Address, sendingLifecycleV2.METHOD_ADVANCE_SENDING, in)
}

func (c *NetworkClient) FailSending(in lifecycleCommonV2.FailInput) (types.ContractOutput, error) {
	return c.sendLifecycleTransaction(in.Address, sendingLifecycleV2.METHOD_FAIL_SENDING, in)
}

func (c *NetworkClient) GetSending(address, requestID string) (types.ContractOutput, error) {
	return c.getLifecycleState(address, requestID, sendingLifecycleV2.METHOD_GET_SENDING)
}

func (c *NetworkClient) StartMultiCurrency(in lifecycleCommonV2.StartInput) (types.ContractOutput, error) {
	return c.sendLifecycleTransaction(in.Address, multiCurrencyLifecycleV2.METHOD_START_MULTI_CURRENCY, in)
}

func (c *NetworkClient) AdvanceMultiCurrency(in lifecycleCommonV2.AdvanceInput) (types.ContractOutput, error) {
	return c.sendLifecycleTransaction(in.Address, multiCurrencyLifecycleV2.METHOD_ADVANCE_MULTI_CURRENCY, in)
}

func (c *NetworkClient) FailMultiCurrency(in lifecycleCommonV2.FailInput) (types.ContractOutput, error) {
	return c.sendLifecycleTransaction(in.Address, multiCurrencyLifecycleV2.METHOD_FAIL_MULTI_CURRENCY, in)
}

func (c *NetworkClient) GetMultiCurrency(address, requestID string) (types.ContractOutput, error) {
	return c.getLifecycleState(address, requestID, multiCurrencyLifecycleV2.METHOD_GET_MULTI_CURRENCY)
}

func (c *NetworkClient) sendLifecycleTransaction(address, method string, input any) (types.ContractOutput, error) {
	if address == "" {
		return types.ContractOutput{}, fmt.Errorf("address not set")
	}
	if err := keys.ValidateEDDSAPublicKeyHex(address); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid lifecycle address: %w", err)
	}
	if c.walletManager == nil {
		return types.ContractOutput{}, fmt.Errorf("wallet manager is required")
	}

	from := c.walletManager.OwnerAddress()
	if err := keys.ValidateEDDSAPublicKeyHex(from); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid from address: %w", err)
	}

	data, err := structToLifecycleMap(input)
	if err != nil {
		return types.ContractOutput{}, err
	}

	uuid7, err := utils.NewUUID7()
	if err != nil {
		return types.ContractOutput{}, fmt.Errorf("failed to generate UUIDv7: %w", err)
	}

	return c.SignAndSendTransaction(c.chainId, from, address, method, data, uint8(1), uuid7)
}

func (c *NetworkClient) getLifecycleState(address, requestID, method string) (types.ContractOutput, error) {
	if address == "" {
		return types.ContractOutput{}, fmt.Errorf("address not set")
	}
	if err := keys.ValidateEDDSAPublicKeyHex(address); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid lifecycle address: %w", err)
	}
	if requestID == "" {
		return types.ContractOutput{}, fmt.Errorf("request_id not set")
	}

	return c.GetState(address, method, map[string]interface{}{"request_id": requestID})
}

func structToLifecycleMap(input any) (map[string]interface{}, error) {
	buf, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal lifecycle input: %w", err)
	}

	var data map[string]interface{}
	if err := json.Unmarshal(buf, &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal lifecycle input: %w", err)
	}
	return data, nil
}
