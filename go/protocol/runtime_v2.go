package protocol

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"gitlab.com/2finance/2finance-network/blockchain/transaction"
	"gitlab.com/2finance/2finance-network/blockchain/utils"
)

const (
	ProtocolVersionV2 = uint32(2)

	RuntimeNativeGoV2 = "native_go"
	RuntimeEVMV2      = "evm"

	// Runtime versions are executor-specific. Native Go remains on its first
	// V2 runtime ABI, while the deterministic EVM executor is ABI version 2.
	NativeGoRuntimeVersionV2 = uint32(1)
	EVMRuntimeVersionV2      = uint32(2)

	TransactionVersionV2 = uint8(1)

	// EVMReservedAddressV2 and EVMExecuteMethodV2 select the deterministic EVM
	// without changing the signed transaction envelope shared with Native Go.
	EVMReservedAddressV2 = "2b6bf3044c330548317fb564120b29a0246eaec4a255145dd350ec288e63264e"
	EVMExecuteMethodV2   = "evm_execute"
)

type NativeGoTransactionV2Input struct {
	ChainID       uint8
	From          string
	To            string
	Method        string
	Payload       any
	UUID7         string
	Authorization *transaction.AuthorizationEnvelope
}

type EVMMessageKindV2 string

const (
	EVMCallV2   EVMMessageKindV2 = "call"
	EVMCreateV2 EVMMessageKindV2 = "create"
)

type EVMAccessTupleV2 struct {
	Address     string   `json:"address"`
	StorageKeys []string `json:"storage_keys"`
}

// EVMMessageV2 is the strict payload carried inside the v2 execution
// envelope. Value is an unsigned 256-bit integer encoded as exactly 64
// lowercase hexadecimal characters. Calldata is lowercase, even-length hex.
type EVMMessageV2 struct {
	Kind       EVMMessageKindV2   `json:"kind"`
	To         string             `json:"to,omitempty"`
	Value      string             `json:"value"`
	GasLimit   uint64             `json:"gas_limit"`
	Calldata   string             `json:"calldata"`
	AccessList []EVMAccessTupleV2 `json:"access_list,omitempty"`
}

type EVMTransactionV2Input struct {
	ChainID       uint8
	From          string
	Message       EVMMessageV2
	UUID7         string
	Authorization *transaction.AuthorizationEnvelope
}

// PrepareNativeGoTransactionV2 creates the unsigned transaction shape that is
// signed by WalletManager.SignPreparedTransaction and submitted to the v2 API.
func PrepareNativeGoTransactionV2(input NativeGoTransactionV2Input) (PreparedTransaction, error) {
	if strings.TrimSpace(input.To) == "" {
		return PreparedTransaction{}, errors.New("protocol v2: native destination is required")
	}
	if strings.TrimSpace(input.Method) == "" {
		return PreparedTransaction{}, errors.New("protocol v2: native method is required")
	}
	data, err := executionEnvelopeV2(RuntimeNativeGoV2, NativeGoRuntimeVersionV2, input.Payload)
	if err != nil {
		return PreparedTransaction{}, err
	}
	return prepareTransactionV2(
		input.ChainID,
		input.From,
		input.To,
		input.Method,
		data,
		input.UUID7,
		input.Authorization,
	)
}

// PrepareEVMTransactionV2 creates the same unsigned transaction envelope, but
// selects the reserved EVM destination and method.
func PrepareEVMTransactionV2(input EVMTransactionV2Input) (PreparedTransaction, error) {
	if input.Message.Kind != EVMCallV2 && input.Message.Kind != EVMCreateV2 {
		return PreparedTransaction{}, fmt.Errorf(
			"protocol v2: unsupported EVM message kind %q",
			input.Message.Kind,
		)
	}
	data, err := executionEnvelopeV2(RuntimeEVMV2, EVMRuntimeVersionV2, input.Message)
	if err != nil {
		return PreparedTransaction{}, err
	}
	return prepareTransactionV2(
		input.ChainID,
		input.From,
		EVMReservedAddressV2,
		EVMExecuteMethodV2,
		data,
		input.UUID7,
		input.Authorization,
	)
}

func prepareTransactionV2(
	chainID uint8,
	from string,
	to string,
	method string,
	data map[string]interface{},
	uuid7 string,
	authorization *transaction.AuthorizationEnvelope,
) (PreparedTransaction, error) {
	if chainID == 0 || chainID > 2 {
		return PreparedTransaction{}, errors.New("protocol v2: chain id must be 1 or 2")
	}
	if strings.TrimSpace(from) == "" {
		return PreparedTransaction{}, errors.New("protocol v2: sender is required")
	}
	if strings.TrimSpace(uuid7) == "" {
		generated, err := utils.NewUUID7()
		if err != nil {
			return PreparedTransaction{}, fmt.Errorf("protocol v2: generate UUIDv7: %w", err)
		}
		uuid7 = generated
	}
	if err := utils.ValidateUUID7(uuid7); err != nil {
		return PreparedTransaction{}, fmt.Errorf("protocol v2: invalid UUIDv7: %w", err)
	}
	return PreparedTransaction{
		ChainID:       chainID,
		From:          from,
		To:            to,
		Method:        method,
		Data:          data,
		Version:       TransactionVersionV2,
		UUID7:         uuid7,
		Authorization: authorization,
	}, nil
}

func executionEnvelopeV2(runtime string, runtimeVersion uint32, payload any) (map[string]interface{}, error) {
	raw, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("protocol v2: encode %s payload: %w", runtime, err)
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	var normalized map[string]interface{}
	if err := decoder.Decode(&normalized); err != nil {
		return nil, fmt.Errorf("protocol v2: %s payload must be a JSON object: %w", runtime, err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		return nil, fmt.Errorf("protocol v2: %s payload contains trailing JSON", runtime)
	}
	if normalized == nil {
		return nil, fmt.Errorf("protocol v2: %s payload must be a JSON object", runtime)
	}
	return map[string]interface{}{
		"runtime": runtime,
		"version": runtimeVersion,
		"payload": normalized,
	}, nil
}
