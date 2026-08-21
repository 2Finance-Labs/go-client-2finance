package protocol

import (
	"encoding/json"
	"strings"
	"testing"

	"gitlab.com/2finance/2finance-network/blockchain/utils"
)

func TestPrepareNativeGoAndEVMTransactionsUseOneSignedShape(t *testing.T) {
	uuidNative, err := utils.NewUUID7()
	if err != nil {
		t.Fatal(err)
	}
	uuidEVM, err := utils.NewUUID7()
	if err != nil {
		t.Fatal(err)
	}
	from := strings.Repeat("1", 64)
	native, err := PrepareNativeGoTransactionV2(NativeGoTransactionV2Input{
		ChainID: 2,
		From:    from,
		To:      strings.Repeat("0", 64),
		Method:  "deploy_contract",
		UUID7:   uuidNative,
		Payload: map[string]any{"contract_version": "walletV2"},
	})
	if err != nil {
		t.Fatal(err)
	}
	evm, err := PrepareEVMTransactionV2(EVMTransactionV2Input{
		ChainID: 2,
		From:    from,
		UUID7:   uuidEVM,
		Message: EVMMessageV2{
			Kind:     EVMCreateV2,
			Value:    strings.Repeat("0", 64),
			GasLimit: 300_000,
			Calldata: "6000",
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	if native.Data["runtime"] != RuntimeNativeGoV2 || evm.Data["runtime"] != RuntimeEVMV2 {
		t.Fatalf("unexpected runtime selectors: native=%v evm=%v", native.Data, evm.Data)
	}
	if native.Data["version"] != NativeGoRuntimeVersionV2 || evm.Data["version"] != EVMRuntimeVersionV2 {
		t.Fatalf("unexpected runtime versions: native=%v evm=%v", native.Data, evm.Data)
	}
	if native.Version != TransactionVersionV2 || evm.Version != TransactionVersionV2 {
		t.Fatalf("unexpected transaction versions: native=%d evm=%d", native.Version, evm.Version)
	}
	if evm.To != EVMReservedAddressV2 || evm.Method != EVMExecuteMethodV2 {
		t.Fatalf("unexpected EVM selector: %s %s", evm.To, evm.Method)
	}
	encoded, err := json.Marshal(evm.Data)
	if err != nil {
		t.Fatal(err)
	}
	if string(encoded) != `{"payload":{"calldata":"6000","gas_limit":300000,"kind":"create","value":"0000000000000000000000000000000000000000000000000000000000000000"},"runtime":"evm","version":2}` {
		t.Fatalf("unexpected EVM envelope: %s", encoded)
	}
}

func TestPrepareV2GeneratesUUIDAndRejectsNonObjectPayload(t *testing.T) {
	prepared, err := PrepareNativeGoTransactionV2(NativeGoTransactionV2Input{
		ChainID: 2,
		From:    strings.Repeat("1", 64),
		To:      strings.Repeat("0", 64),
		Method:  "deploy_contract",
		Payload: map[string]any{"contract_version": "walletV2"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := utils.ValidateUUID7(prepared.UUID7); err != nil {
		t.Fatalf("generated UUIDv7 is invalid: %v", err)
	}
	_, err = PrepareNativeGoTransactionV2(NativeGoTransactionV2Input{
		ChainID: 2,
		From:    strings.Repeat("1", 64),
		To:      strings.Repeat("0", 64),
		Method:  "deploy_contract",
		Payload: []string{"not", "an", "object"},
	})
	if err == nil {
		t.Fatal("expected non-object payload to fail")
	}
}
