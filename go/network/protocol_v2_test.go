package network

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/2Finance-Labs/2finance-sdk-client/internal/service"
	"github.com/2Finance-Labs/2finance-sdk-client/protocol"
)

type protocolV2RoundTripFunc func(*http.Request) (*http.Response, error)

func (fn protocolV2RoundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return fn(request)
}

func TestSubmitAndQueryProtocolV2Transaction(t *testing.T) {
	hash := strings.Repeat("a", 64)
	signature := strings.Repeat("b", 128)
	signed := protocol.SignedTransaction{
		ChainID:   2,
		From:      strings.Repeat("1", 64),
		To:        strings.Repeat("0", 64),
		Method:    "deploy_contract",
		Data:      map[string]interface{}{"runtime": "native_go", "version": 1, "payload": map[string]interface{}{"contract_version": "walletV2"}},
		Version:   1,
		UUID7:     "019fba91-223d-7c2e-9825-65eb7c3607e0",
		Hash:      hash,
		Signature: signature,
	}
	committedAt := time.Date(2026, 8, 1, 1, 2, 3, 0, time.UTC)
	status := ProtocolV2TransactionStatus{
		ProtocolVersion:   2,
		ChainID:           "2finance-dev-v2",
		TransactionHash:   hash,
		SignedTransaction: mustProtocolV2JSON(t, signed),
		Commit: ProtocolV2Commit{
			Sequence:        1,
			AdmissionHeight: 2,
			Status:          "committed",
			CommittedAt:     committedAt,
		},
		Artifact: ProtocolV2Artifact{Runtime: "native_go", RuntimeVersion: 1},
	}
	requestCount := 0
	httpClient := &http.Client{Transport: protocolV2RoundTripFunc(func(request *http.Request) (*http.Response, error) {
		requestCount++
		var payload any
		if request.Method == http.MethodPost {
			if request.URL.Path != transactionsV2Path {
				t.Fatalf("POST path = %q", request.URL.Path)
			}
			body, err := io.ReadAll(request.Body)
			if err != nil {
				t.Fatal(err)
			}
			var actual protocol.SignedTransaction
			if err := json.Unmarshal(body, &actual); err != nil {
				t.Fatal(err)
			}
			if actual.Hash != hash || actual.Signature != signature {
				t.Fatalf("signed body changed: %+v", actual)
			}
			payload = protocolV2Response[ProtocolV2SubmitResult]{
				Code: 200,
				Msg:  "Transaction committed; finality is pending",
				Data: ProtocolV2SubmitResult{
					ExecutionOutput: json.RawMessage(`{"states":[],"logs":[]}`),
					Transaction:     status,
				},
			}
		} else {
			if request.URL.Path != transactionsV2Path+"/"+hash+"/finality" {
				t.Fatalf("GET path = %q", request.URL.Path)
			}
			payload = protocolV2Response[ProtocolV2TransactionStatus]{Code: 200, Msg: "Successfully", Data: status}
		}
		encoded, err := json.Marshal(payload)
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
	client := New("http://2finance-network:9095", httpClient)

	submitted, err := client.SubmitSignedTransactionV2(context.Background(), signed)
	if err != nil {
		t.Fatal(err)
	}
	if submitted.Transaction.Commit.Sequence != 1 || submitted.Transaction.Artifact.Runtime != "native_go" {
		t.Fatalf("unexpected submission response: %+v", submitted)
	}
	queried, err := client.TransactionFinalityV2(context.Background(), strings.ToUpper(hash))
	if err != nil {
		t.Fatal(err)
	}
	if queried.TransactionHash != hash || string(queried.SignedTransaction) != string(status.SignedTransaction) {
		t.Fatalf("unexpected finality response: %+v", queried)
	}
	decoded, err := queried.DecodeSignedTransaction()
	if err != nil {
		t.Fatal(err)
	}
	if decoded.Signature != signature {
		t.Fatalf("decoded signature = %q", decoded.Signature)
	}
	if requestCount != 2 {
		t.Fatalf("request count = %d", requestCount)
	}
}

func TestTransactionFinalityV2RejectsMalformedHashBeforeHTTP(t *testing.T) {
	client := New("http://2finance-network:9095", &http.Client{Transport: protocolV2RoundTripFunc(func(request *http.Request) (*http.Response, error) {
		t.Fatal("HTTP must not be called")
		return nil, nil
	})})
	_, err := client.TransactionFinalityV2(context.Background(), "bad")
	if err == nil {
		t.Fatal("expected malformed hash error")
	}
}

func TestSubmitProtocolV2RetriesTransientConflictWithIdenticalSignedBody(t *testing.T) {
	hash := strings.Repeat("a", 64)
	signed := protocol.SignedTransaction{
		ChainID:   2,
		From:      strings.Repeat("1", 64),
		To:        strings.Repeat("0", 64),
		Method:    "deploy_contract",
		Data:      map[string]interface{}{"runtime": "native_go", "version": 1},
		Version:   1,
		UUID7:     "019fba91-223d-7c2e-9825-65eb7c3607e0",
		Hash:      hash,
		Signature: strings.Repeat("b", 128),
	}
	var bodies [][]byte
	httpClient := &http.Client{Transport: protocolV2RoundTripFunc(func(request *http.Request) (*http.Response, error) {
		body, err := io.ReadAll(request.Body)
		if err != nil {
			t.Fatal(err)
		}
		bodies = append(bodies, body)
		if len(bodies) == 1 {
			return &http.Response{
				StatusCode: http.StatusUnprocessableEntity,
				Header:     http.Header{"Content-Type": []string{"application/json"}},
				Body:       io.NopCloser(strings.NewReader(`{"error":"can't serialize access for this transaction"}`)),
				Request:    request,
			}, nil
		}
		response := protocolV2Response[ProtocolV2SubmitResult]{
			Code: 200,
			Data: ProtocolV2SubmitResult{Transaction: ProtocolV2TransactionStatus{
				TransactionHash: hash,
			}},
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

	result, err := New("http://2finance-network:9095", httpClient).
		SubmitSignedTransactionV2(context.Background(), signed)
	if err != nil {
		t.Fatal(err)
	}
	if result.Transaction.TransactionHash != hash {
		t.Fatalf("transaction hash = %q", result.Transaction.TransactionHash)
	}
	if len(bodies) != 2 {
		t.Fatalf("request count = %d", len(bodies))
	}
	if string(bodies[0]) != string(bodies[1]) {
		t.Fatalf("signed body changed between attempts:\nfirst:  %s\nsecond: %s", bodies[0], bodies[1])
	}
}

func TestProtocolV2ClassifiesDatabaseDeadlineAsTransient(t *testing.T) {
	err := &service.HTTPError{
		StatusCode: http.StatusUnprocessableEntity,
		Body:       []byte(`{"error":"failed to insert transaction: context deadline exceeded"}`),
	}
	if !isTransientProtocolV2SubmissionError(err) {
		t.Fatal("database transaction deadline must be retried with the identical signed body")
	}
}

func TestExecutionLogsV2UsesUnifiedEndpoint(t *testing.T) {
	client := New("http://2finance-network:9095", &http.Client{Transport: protocolV2RoundTripFunc(func(request *http.Request) (*http.Response, error) {
		if request.Method != http.MethodGet || request.URL.Path != "/v2/2finance-network/logs" {
			t.Fatalf("request = %s %s", request.Method, request.URL.Path)
		}
		if request.URL.Query().Get("runtime") != "evm" {
			t.Fatalf("query = %v", request.URL.Query())
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body: io.NopCloser(strings.NewReader(`{
				"code":200,
				"data":{"page":1,"limit":100,"logs":[{
					"chain_id":"chain-v2","transaction_hash":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
					"log_index":0,"source_log_index":0,"runtime":"evm","contract_address":"0x1234",
					"contract_version":"evmV2","event_signature":"0xabcd","topics":["0xabcd"],
					"data":{"hex":"0xdead"},"raw_data":"0xdead","leaf_hash":"bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
					"created_at":"2026-08-02T00:00:00Z"
				}]}
			}`)),
			Request: request,
		}, nil
	})})
	page, err := client.ExecutionLogsV2(context.Background(), url.Values{"runtime": {"evm"}})
	if err != nil {
		t.Fatal(err)
	}
	if len(page.Logs) != 1 || page.Logs[0].Runtime != "evm" || page.Logs[0].ContractVersion != "evmV2" {
		t.Fatalf("logs = %+v", page.Logs)
	}
}

func mustProtocolV2JSON(t *testing.T, value any) json.RawMessage {
	t.Helper()
	encoded, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return encoded
}
