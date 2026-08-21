package protocol

import (
	"encoding/json"
	"testing"

	"gitlab.com/2finance/2finance-network/blockchain/transaction"
)

func TestSignedTransactionFromNetworkRoundTrip(t *testing.T) {
	networkTx := &transaction.Transaction{
		ChainID:   1,
		From:      "from",
		To:        "to",
		Method:    "method",
		Data:      json.RawMessage(`{"contract_version":"testV2","amount":"10"}`),
		Version:   1,
		UUID7:     "uuid",
		Hash:      "hash",
		Signature: "signature",
	}

	signed, err := SignedTransactionFromNetwork(networkTx)
	if err != nil {
		t.Fatalf("SignedTransactionFromNetwork: %v", err)
	}
	if signed.Data["contract_version"] != "testV2" {
		t.Fatalf("contract_version = %v, want testV2", signed.Data["contract_version"])
	}

	roundTrip, err := signed.ToNetworkTransaction()
	if err != nil {
		t.Fatalf("ToNetworkTransaction: %v", err)
	}
	if roundTrip.Method != networkTx.Method {
		t.Fatalf("method = %v, want %v", roundTrip.Method, networkTx.Method)
	}
	if string(roundTrip.Data) != `{"amount":"10","contract_version":"testV2"}` {
		t.Fatalf("data = %s", string(roundTrip.Data))
	}
}
