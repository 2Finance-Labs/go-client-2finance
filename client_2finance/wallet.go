package client_2finance

import (

	"fmt"
	"gitlab.com/2finance/2finance-network/blockchain/encryption/keys"
	"gitlab.com/2finance/2finance-network/blockchain/types"
	"gitlab.com/2finance/2finance-network/blockchain/contract/walletV1"

	"gitlab.com/2finance/2finance-network/blockchain/utils"

)
// AddWallet creates a new wallet
// and sends a transaction to the network
// Amount is a string representation of the amount to be added
// If amount is empty, it defaults to "0"
// Amount must be sent considering decimals
// For example, if the amount is 100000, and decimals is 18, the amount in database will be 100000000000000000000000
// if the amonut is 0,0000000001, and decimals is 18, the amount in database will be 100000000
func (c *networkClient) AddWallet(address, pubKey string) (types.ContractOutput, error) {
	if pubKey == "" {
		return types.ContractOutput{}, fmt.Errorf("public key not set")
	}
	if err := keys.ValidateEDDSAPublicKey(pubKey); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid public key: %w", err)
	}
	if address == "" {
		return types.ContractOutput{}, fmt.Errorf("contract address not set")
	}
	if err := keys.ValidateEDDSAPublicKey(address); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid contract address: %w", err)
	}

	from := c.publicKey
	if from == "" {
		return types.ContractOutput{}, fmt.Errorf("from address not set")
	}

	to := address
	method := walletV1.METHOD_ADD_WALLET
	data := map[string]interface{}{
		"address":    address,
		"public_key": pubKey,
		//TODO REMOVER
		"amount":     "0",
	}

	contractOutput, err := c.SignAndSendTransaction(
		from,
		to,
		method,
		data)
	if err != nil {
		return types.ContractOutput{}, fmt.Errorf("failed to send transaction: %w", err)
	}

	return contractOutput, nil
}

func (c *networkClient) GetWallet(pubKey string) (types.ContractOutput, error) {
	
	if pubKey == "" {
		return types.ContractOutput{}, fmt.Errorf("public key not set")
	}
	err := keys.ValidateEDDSAPublicKey(pubKey)
	if err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid public key: %w", err)
	}

	method := walletV1.METHOD_GET_WALLET_BY_PUBLIC_KEY
	data := map[string]interface{}{
		"public_key": pubKey,
	}

	contractOutput, err := c.GetState("", method, data)
	if err != nil {
		return types.ContractOutput{}, fmt.Errorf("failed to get state: %w", err)
	}

	return contractOutput, nil
}

func (c *networkClient) TransferWallet(to, amount string, decimals int) (types.ContractOutput, error) {
	if c.publicKey == "" {
		return types.ContractOutput{}, fmt.Errorf("public key not set")
	}
	if to == "" {
		return types.ContractOutput{}, fmt.Errorf("to address not set")
	}
	if to == c.publicKey {
		return types.ContractOutput{}, fmt.Errorf("cannot transfer to the same address")
	}
	if err := keys.ValidateEDDSAPublicKey(to); err != nil {
		return types.ContractOutput{}, fmt.Errorf("invalid to address: %w", err)
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

	method := walletV1.METHOD_TRANSFER_WALLET
	data := map[string]interface{}{
		"from":    c.publicKey,
		"to":      to,
		"amount":  amount,
	}

	contractOutput, err := c.SignAndSendTransaction(
		c.publicKey,
		to,
		method,
		data)
	if err != nil {
		return types.ContractOutput{}, fmt.Errorf("failed to send transaction: %w", err)
	}
	
	//TODOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOO
	// transferOutput.Transfer = contractOutput.States[0].Object.(*domain.Transfer)
	// transferOutput.EventTransfer = &contractOutput.Events[0]
	// transferOutput.LogTypeTransfer = &contractOutput.LogTypes[0]
	
	// transferOutput.WalletSender = contractOutput.States[1].Object.(*domain.Wallet)
	// transferOutput.EventSender = &contractOutput.Events[1]
	// transferOutput.LogTypeSender = &contractOutput.LogTypes[1]

	// transferOutput.WalletReceiver = contractOutput.States[2].Object.(*domain.Wallet)
	// transferOutput.EventReceiver = &contractOutput.Events[2]
	// transferOutput.LogTypeReceiver = &contractOutput.LogTypes[2]

	return contractOutput, nil
}