package client_2finance

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"gitlab.com/2finance/2finance-network/blockchain/transaction"
	"gitlab.com/2finance/2finance-network/blockchain/virtualmachine"
)

type v2RoundTripFunc func(*http.Request) (*http.Response, error)

func (fn v2RoundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return fn(request)
}

func TestNewV2RoutesSignedWritesAndQueriesOverHTTP(t *testing.T) {
	hash := strings.Repeat("a", 64)
	requests := make([]string, 0, 2)
	httpClient := &http.Client{Transport: v2RoundTripFunc(func(request *http.Request) (*http.Response, error) {
		requests = append(requests, request.Method+" "+request.URL.Path)
		body, err := io.ReadAll(request.Body)
		if err != nil {
			t.Fatal(err)
		}
		var response any
		if request.URL.Path == "/v2/2finance-network/transactions" {
			var signed transaction.Transaction
			if err := json.Unmarshal(body, &signed); err != nil {
				t.Fatal(err)
			}
			var envelope map[string]any
			if err := json.Unmarshal(signed.Data, &envelope); err != nil {
				t.Fatal(err)
			}
			if envelope["runtime"] != "native_go" {
				t.Fatalf("runtime = %v", envelope["runtime"])
			}
			response = map[string]any{
				"code": 200,
				"msg":  "Transaction committed; finality is pending",
				"data": map[string]any{
					"execution_output": map[string]any{"states": []any{}, "logs": []any{}},
					"transaction": map[string]any{
						"protocol_version":   2,
						"chain_id":           "2finance-dev-v2",
						"transaction_hash":   hash,
						"signed_transaction": signed,
						"commit": map[string]any{
							"sequence":         3,
							"admission_height": 3,
							"status":           "committed",
							"committed_at":     "2026-08-01T01:02:03Z",
						},
						"artifact": map[string]any{"runtime": "native_go", "runtime_version": 1},
					},
				},
			}
		} else {
			var requestEnvelope map[string]any
			if err := json.Unmarshal(body, &requestEnvelope); err != nil {
				t.Fatal(err)
			}
			if requestEnvelope["method"] != virtualmachine.REQUEST_METHOD_GET_STATE {
				t.Fatalf("query method = %v", requestEnvelope["method"])
			}
			response = map[string]any{
				"code": 200,
				"msg":  "Successfully",
				"data": map[string]any{"states": []any{}, "logs": []any{}},
			}
		}
		encoded, err := json.Marshal(response)
		if err != nil {
			t.Fatal(err)
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body:       io.NopCloser(strings.NewReader(string(encoded))),
			Request:    request,
		}, nil
	})}
	client := NewV2("http://2finance-network:9095", httpClient, nil)
	signed := &transaction.Transaction{
		ChainID:   2,
		From:      strings.Repeat("1", 64),
		To:        strings.Repeat("0", 64),
		Method:    "deploy_contract",
		Data:      json.RawMessage(`{"runtime":"native_go","version":1,"payload":{"contract_version":"walletV2"}}`),
		Version:   1,
		UUID7:     "019fba91-223d-7c2e-9825-65eb7c3607e0",
		Hash:      hash,
		Signature: strings.Repeat("b", 128),
	}

	writeOutput, err := client.SendTransaction(virtualmachine.REQUEST_METHOD_SEND, signed, "ignored")
	if err != nil {
		t.Fatal(err)
	}
	if string(writeOutput) != `{"logs":[],"states":[]}` && string(writeOutput) != `{"states":[],"logs":[]}` {
		t.Fatalf("write output = %s", writeOutput)
	}
	queryOutput, err := client.SendTransaction(
		virtualmachine.REQUEST_METHOD_GET_STATE,
		map[string]any{"to": signed.To, "method": "get", "data": map[string]any{}},
		"ignored",
	)
	if err != nil {
		t.Fatal(err)
	}
	if len(queryOutput) == 0 {
		t.Fatal("query output is empty")
	}
	if len(requests) != 2 || requests[0] != "POST /v2/2finance-network/transactions" || requests[1] != "POST /v2/2finance-network/query" {
		t.Fatalf("requests = %#v", requests)
	}
}

func TestNormalizePublicContractVersionsV2Recursively(t *testing.T) {
	normalized, err := normalizePublicContractVersionsV2(map[string]any{
		"contract_version": "walletV2",
		"nested": []any{
			map[string]any{"contract_version": "tokenV2"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := json.Marshal(normalized)
	if err != nil {
		t.Fatal(err)
	}
	if string(encoded) != `{"contract_version":"walletV2","nested":[{"contract_version":"tokenV2"}]}` {
		t.Fatalf("normalized payload = %s", encoded)
	}

	if _, err := normalizePublicContractVersionsV2(map[string]any{
		"contract_version": "unknownV2",
	}); err == nil {
		t.Fatal("expected unsupported legacy contract version to be rejected")
	}
}
