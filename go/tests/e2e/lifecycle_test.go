package e2e_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	clientauth "github.com/2Finance-Labs/2finance-sdk-client/auth"
	"github.com/2Finance-Labs/2finance-sdk-client/protocol"
	"gitlab.com/2finance/2finance-network/blockchain/contract/fxLifecycleV2"
	"gitlab.com/2finance/2finance-network/blockchain/contract/lifecycleCommonV2"
	"gitlab.com/2finance/2finance-network/blockchain/contract/multiCurrencyLifecycleV2"
	"gitlab.com/2finance/2finance-network/blockchain/contract/onboardingLifecycleV2"
	"gitlab.com/2finance/2finance-network/blockchain/contract/receivingLifecycleV2"
	"gitlab.com/2finance/2finance-network/blockchain/contract/sendingLifecycleV2"
	"gitlab.com/2finance/2finance-network/blockchain/log"
	"gitlab.com/2finance/2finance-network/blockchain/utils"
)

func TestLifecyclePreparedTransactions_CanBeSignedByWallet(t *testing.T) {
	manager, owner := newImportedWalletForUnlockTest(t)
	if err := manager.UnlockWithPassword(testWalletPassword); err != nil {
		t.Fatalf("UnlockWithPassword error: %v", err)
	}

	contractAddress, _ := newUnlockTestKeyPair(t)
	tests := []struct {
		name            string
		contractVersion string
		method          string
		data            map[string]interface{}
	}{
		{
			name:            "fx",
			contractVersion: fxLifecycleV2.FX_LIFECYCLE_CONTRACT_V2,
			method:          fxLifecycleV2.METHOD_START_FX,
			data: map[string]interface{}{
				"address":         contractAddress,
				"owner":           owner,
				"request_id":      "fx-e2e-001",
				"provider":        "wise",
				"source_currency": "USD",
				"target_currency": "BRL",
			},
		},
		{
			name:            "onboarding",
			contractVersion: onboardingLifecycleV2.ONBOARDING_LIFECYCLE_CONTRACT_V2,
			method:          onboardingLifecycleV2.METHOD_START_ONBOARDING,
			data: map[string]interface{}{
				"address":             contractAddress,
				"owner":               owner,
				"request_id":          "onboarding-e2e-001",
				"provider":            "wise",
				"provider_account_id": "account-1",
				"profile_id":          "profile-1",
				"currency":            "USD",
			},
		},
		{
			name:            "receiving",
			contractVersion: receivingLifecycleV2.RECEIVING_LIFECYCLE_CONTRACT_V2,
			method:          receivingLifecycleV2.METHOD_START_RECEIVING,
			data: map[string]interface{}{
				"address":             contractAddress,
				"owner":               owner,
				"request_id":          "receiving-e2e-001",
				"provider":            "wise",
				"provider_account_id": "account-1",
				"external_auth_id":    "auth-1",
				"profile_id":          "profile-1",
				"external_wallet_id":  "wallet-1",
				"currency":            "USD",
			},
		},
		{
			name:            "receiving_codexa_brl_pix",
			contractVersion: receivingLifecycleV2.RECEIVING_LIFECYCLE_CONTRACT_V2,
			method:          receivingLifecycleV2.METHOD_START_RECEIVING,
			data: map[string]interface{}{
				"address":             contractAddress,
				"owner":               owner,
				"request_id":          "receiving-codexa-brl-e2e-001",
				"provider":            "codexa",
				"provider_account_id": "account-1",
				"external_auth_id":    "auth-1",
				"profile_id":          "profile-1",
				"currency":            "BRL",
				"amount":              "50.00",
				"details":             map[string]interface{}{"payment_method": "pix"},
			},
		},
		{
			name:            "sending",
			contractVersion: sendingLifecycleV2.SENDING_LIFECYCLE_CONTRACT_V2,
			method:          sendingLifecycleV2.METHOD_START_SENDING,
			data: map[string]interface{}{
				"address":                 contractAddress,
				"owner":                   owner,
				"request_id":              "sending-e2e-001",
				"provider":                "wise",
				"provider_account_id":     "account-1",
				"external_auth_id":        "auth-1",
				"profile_id":              "profile-1",
				"external_quote_id":       "quote-1",
				"external_beneficiary_id": "beneficiary-1",
				"currency":                "USD",
				"amount":                  "25.50",
			},
		},
		{
			name:            "sending_codexa_quote_first",
			contractVersion: sendingLifecycleV2.SENDING_LIFECYCLE_CONTRACT_V2,
			method:          sendingLifecycleV2.METHOD_START_SENDING,
			data: map[string]interface{}{
				"address":             contractAddress,
				"owner":               owner,
				"request_id":          "sending-codexa-e2e-001",
				"provider":            "codexa",
				"provider_account_id": "account-1",
				"external_auth_id":    "auth-1",
				"profile_id":          "profile-1",
				"source_currency":     "BRL",
				"target_currency":     "USD",
				"currency":            "USD",
				"amount":              "1000.00",
				"beneficiary_name":    "John Smith",
				"beneficiary_type":    "individual",
				"beneficiary_country": "US",
			},
		},
		{
			name:            "multi_currency",
			contractVersion: multiCurrencyLifecycleV2.MULTI_CURRENCY_LIFECYCLE_CONTRACT_V2,
			method:          multiCurrencyLifecycleV2.METHOD_START_MULTI_CURRENCY,
			data: map[string]interface{}{
				"address":             contractAddress,
				"owner":               owner,
				"request_id":          "multi-currency-e2e-001",
				"provider":            "wise",
				"provider_account_id": "account-1",
				"external_auth_id":    "auth-1",
				"profile_id":          "profile-1",
				"currency":            "USD",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uuid7, err := utils.NewUUID7()
			if err != nil {
				t.Fatalf("NewUUID7 error: %v", err)
			}
			tt.data["contract_version"] = tt.contractVersion

			signed, err := manager.SignPreparedTransaction(protocol.PreparedTransaction{
				ChainID: 1,
				From:    owner,
				To:      contractAddress,
				Method:  tt.method,
				Data:    tt.data,
				Version: 1,
				UUID7:   uuid7,
			})
			if err != nil {
				t.Fatalf("SignPreparedTransaction error: %v", err)
			}
			if signed.Hash == "" || signed.Signature == "" {
				t.Fatalf("hash/signature should not be empty")
			}
			if signed.Method != tt.method {
				t.Fatalf("method = %v, want %v", signed.Method, tt.method)
			}
			if signed.Data["contract_version"] != tt.contractVersion {
				t.Fatalf("contract_version = %v, want %v", signed.Data["contract_version"], tt.contractVersion)
			}
			if signed.Data["request_id"] == "" {
				t.Fatalf("request_id should be preserved")
			}
		})
	}
}

func TestSendingLifecycle_StartViaProtocolV2(t *testing.T) {
	signer := setupSignerWallet(t)
	c := setupClient(t, signer.Wallet)

	useWallet(t, c, signer.Wallet)

	tmpWM := setupWalletManager(t)
	contractAddress, _ := genKey(t, tmpWM)

	if _, err := c.DeployContract2(sendingLifecycleV2.SENDING_LIFECYCLE_CONTRACT_V2, contractAddress); err != nil {
		t.Fatalf("DeployContract2 sending lifecycle: %v", err)
	}

	requestID, err := utils.NewUUID7()
	if err != nil {
		t.Fatalf("NewUUID7: %v", err)
	}

	out, err := c.StartSending(lifecycleCommonV2.StartInput{
		Address:               contractAddress,
		Owner:                 signer.PublicKey,
		RequestID:             requestID,
		Provider:              "wise",
		ProviderAccountID:     "account-1",
		ExternalAuthID:        "auth-1",
		ProfileID:             "profile-1",
		ExternalQuoteID:       "quote-1",
		ExternalBeneficiaryID: "beneficiary-1",
		Currency:              "USD",
		Amount:                "25.50",
	})
	if err != nil {
		t.Fatalf("StartSending: %v", err)
	}
	if len(out.Logs) == 0 {
		t.Fatal("expected StartSending to return logs")
	}

	startLog, err := utils.UnmarshalLog[log.Log](out.Logs[0])
	if err != nil {
		t.Fatalf("UnmarshalLog (StartSending.Logs[0]): %v", err)
	}
	if startLog.LogType != "SendingLifecycle_Started" {
		t.Fatalf("log type = %s, want SendingLifecycle_Started", startLog.LogType)
	}

	stateOut, err := c.GetSending(contractAddress, requestID)
	if err != nil {
		t.Fatalf("GetSending: %v", err)
	}
	if len(stateOut.States) == 0 {
		t.Fatal("expected GetSending to return state")
	}
}

func TestSendingLifecycle_MCPPrepareSignSubmitGet(t *testing.T) {
	if !strings.EqualFold(os.Getenv("MCP_E2E"), "true") {
		t.Skip("set MCP_E2E=true and MCP_URL to run against a live MCP HTTP server")
	}

	mcpURL := os.Getenv("MCP_URL")
	if mcpURL == "" {
		mcpURL = "http://127.0.0.1:8089/mcp"
	}
	t.Logf("MCP endpoint: %s", mcpURL)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	mcpClient := newE2EMCPClient(t, mcpURL)
	if _, err := mcpClient.call(ctx, "initialize", map[string]any{"protocolVersion": "2025-06-18"}); err != nil {
		t.Fatalf("mcp initialize: %v", err)
	}

	signer := setupSignerWallet(t)
	tmpWM := setupWalletManager(t)
	contractAddress, _ := genKey(t, tmpWM)

	deployUUID, err := utils.NewUUID7()
	if err != nil {
		t.Fatalf("NewUUID7 deploy: %v", err)
	}
	deployResult := mcpCallPreparedTransaction(t, ctx, mcpClient, "finance.contract.deploy.prepare", map[string]any{
		"from":             signer.PublicKey,
		"contract_address": contractAddress,
		"contract_version": sendingLifecycleV2.SENDING_LIFECYCLE_CONTRACT_V2,
		"uuid7":            deployUUID,
	})
	t.Logf("MCP finance.contract.deploy.prepare -> workflow=%s method=%s to=%s contract_version=%v",
		deployResult.Workflow,
		deployResult.UnsignedTransaction.Method,
		deployResult.UnsignedTransaction.To,
		deployResult.UnsignedTransaction.Data["contract_version"],
	)
	if deployResult.UnsignedTransaction.Method != "deploy_contract2" {
		t.Fatalf("deploy method = %s, want deploy_contract2", deployResult.UnsignedTransaction.Method)
	}
	deploySubmit := mcpSignAndSubmit(t, ctx, mcpClient, signer.Wallet, deployResult.UnsignedTransaction)
	t.Logf("MCP finance.transaction.submit_signed deploy -> logs=%d states=%d", mcpLen(deploySubmit["logs"]), mcpLen(deploySubmit["states"]))

	requestID, err := utils.NewUUID7()
	if err != nil {
		t.Fatalf("NewUUID7 request: %v", err)
	}
	startUUID, err := utils.NewUUID7()
	if err != nil {
		t.Fatalf("NewUUID7 start: %v", err)
	}
	startResult := mcpCallPreparedTransaction(t, ctx, mcpClient, "finance.send.start.prepare", map[string]any{
		"from":                    signer.PublicKey,
		"to":                      contractAddress,
		"address":                 contractAddress,
		"owner":                   signer.PublicKey,
		"request_id":              requestID,
		"provider":                "wise",
		"provider_account_id":     "account-1",
		"external_auth_id":        "auth-1",
		"profile_id":              "profile-1",
		"external_quote_id":       "quote-1",
		"external_beneficiary_id": "beneficiary-1",
		"currency":                "USD",
		"amount":                  "25.50",
		"uuid7":                   startUUID,
	})
	t.Logf("MCP finance.send.start.prepare -> workflow=%s method=%s to=%s request_id=%v",
		startResult.Workflow,
		startResult.UnsignedTransaction.Method,
		startResult.UnsignedTransaction.To,
		startResult.UnsignedTransaction.Data["request_id"],
	)
	if startResult.Workflow != "send.start" {
		t.Fatalf("workflow = %s, want send.start", startResult.Workflow)
	}
	if startResult.UnsignedTransaction.Method != sendingLifecycleV2.METHOD_START_SENDING {
		t.Fatalf("start method = %s, want %s", startResult.UnsignedTransaction.Method, sendingLifecycleV2.METHOD_START_SENDING)
	}
	startSubmit := mcpSignAndSubmit(t, ctx, mcpClient, signer.Wallet, startResult.UnsignedTransaction)
	t.Logf("MCP finance.transaction.submit_signed start -> logs=%d states=%d", mcpLen(startSubmit["logs"]), mcpLen(startSubmit["states"]))

	getRaw, err := mcpClient.callTool(ctx, "finance.send.get", map[string]any{
		"to":         contractAddress,
		"request_id": requestID,
	})
	if err != nil {
		t.Fatalf("finance.send.get: %v", err)
	}
	stateContent := mcpRequireStructuredContent(t, getRaw)
	state := mcpUnwrapLifecycleState(stateContent)
	if state == nil {
		t.Fatalf("expected lifecycle state in MCP get response, got: %#v", stateContent)
	}
	t.Logf("MCP finance.send.get -> request_id=%v status=%v current_step=%v next_command_present=%t",
		state["request_id"],
		state["status"],
		state["current_step"],
		mcpHasMap(state["next_command"]),
	)
	if state["request_id"] != requestID {
		t.Fatalf("request_id = %v, want %s", state["request_id"], requestID)
	}
}

type e2eMCPClient struct {
	url       string
	sessionID string
	http      *http.Client
	nextID    int
}

func newE2EMCPClient(t testing.TB, url string) *e2eMCPClient {
	t.Helper()

	httpClient := &http.Client{Timeout: 15 * time.Second}
	tokenSource, err := mcpE2ETokenSource()
	if err != nil {
		t.Fatalf("MCP auth config: %v", err)
	}
	if tokenSource != nil {
		httpClient.Transport = clientauth.AuthTransport{
			Source: tokenSource,
			Base:   http.DefaultTransport,
		}
	}
	return &e2eMCPClient{url: url, http: httpClient}
}

func mcpE2ETokenSource() (clientauth.TokenSource, error) {
	if token := os.Getenv("MCP_ACCESS_TOKEN"); strings.TrimSpace(token) != "" {
		return clientauth.StaticTokenSource(token), nil
	}

	tokenURL := firstNonEmptyEnv("MCP_OIDC_TOKEN_URL", "AUTH_E2E_TOKEN_URL", "TWO_FINANCE_OIDC_TOKEN_URL")
	clientID := firstNonEmptyEnv("MCP_OIDC_CLIENT_ID", "AUTH_E2E_CLIENT_ID", "TWO_FINANCE_OIDC_CLIENT_ID")
	clientSecret := firstNonEmptyEnv("MCP_OIDC_CLIENT_SECRET", "AUTH_E2E_CLIENT_SECRET", "TWO_FINANCE_OIDC_CLIENT_SECRET")
	if tokenURL == "" && clientID == "" && clientSecret == "" {
		return nil, nil
	}
	if tokenURL == "" || clientID == "" || clientSecret == "" {
		return nil, fmt.Errorf("set MCP_ACCESS_TOKEN or provide MCP_OIDC_TOKEN_URL/MCP_OIDC_CLIENT_ID/MCP_OIDC_CLIENT_SECRET")
	}

	return clientauth.NewClientCredentialsTokenSource(clientauth.ClientCredentialsConfig{
		TokenURL:     tokenURL,
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Scopes:       strings.Fields(firstNonEmptyEnv("MCP_OIDC_SCOPES", "AUTH_E2E_SCOPE", "TWO_FINANCE_OIDC_SCOPES")),
	})
}

func firstNonEmptyEnv(keys ...string) string {
	for _, key := range keys {
		if value := strings.TrimSpace(os.Getenv(key)); value != "" {
			return value
		}
	}
	return ""
}

func TestMCPRedactedErrorBodyMasksOAuthValues(t *testing.T) {
	redacted := mcpRedactedErrorBody([]byte(`{"error":"Bearer token-123 client_secret=s3cr3t-value access_token=token-456"}`))
	for _, forbidden := range []string{"token-123", "s3cr3t-value", "token-456"} {
		if strings.Contains(redacted, forbidden) {
			t.Fatalf("redacted MCP error leaked %q: %s", forbidden, redacted)
		}
	}
}

func TestNewE2EMCPClientUsesClientCredentialsEnv(t *testing.T) {
	t.Setenv("MCP_OIDC_TOKEN_URL", "http://127.0.0.1:8080/realms/2Finance/protocol/openid-connect/token")
	t.Setenv("MCP_OIDC_CLIENT_ID", "2finance-mcp")
	t.Setenv("MCP_OIDC_CLIENT_SECRET", "local-secret")
	t.Setenv("MCP_OIDC_SCOPES", "mcp:invoke finance:prepare")

	client := newE2EMCPClient(t, "http://127.0.0.1:8089/mcp")
	transport, ok := client.http.Transport.(clientauth.AuthTransport)
	if !ok {
		t.Fatalf("transport = %T, want auth.AuthTransport", client.http.Transport)
	}
	if transport.Source == nil {
		t.Fatal("expected client credentials token source")
	}
}

func (c *e2eMCPClient) callTool(ctx context.Context, name string, args map[string]any) (json.RawMessage, error) {
	return c.call(ctx, "tools/call", map[string]any{"name": name, "arguments": args})
}

func (c *e2eMCPClient) call(ctx context.Context, method string, params any) (json.RawMessage, error) {
	c.nextID++
	body, err := json.Marshal(map[string]any{
		"jsonrpc": "2.0",
		"id":      c.nextID,
		"method":  method,
		"params":  params,
	})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if c.sessionID != "" {
		req.Header.Set("Mcp-Session-Id", c.sessionID)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if sessionID := resp.Header.Get("Mcp-Session-Id"); sessionID != "" {
		c.sessionID = sessionID
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("mcp HTTP %d: %s", resp.StatusCode, mcpRedactedErrorBody(raw))
	}

	var decoded struct {
		Result json.RawMessage `json:"result"`
		Error  *struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(raw, &decoded); err != nil {
		return nil, fmt.Errorf("decode MCP response: %w; raw=%s", err, mcpRedactedErrorBody(raw))
	}
	if decoded.Error != nil {
		return nil, fmt.Errorf("mcp jsonrpc error %d: %s", decoded.Error.Code, clientauth.RedactSensitive(decoded.Error.Message))
	}
	return decoded.Result, nil
}

func mcpRedactedErrorBody(raw []byte) string {
	return clientauth.RedactSensitive(string(raw))
}

func mcpCallPreparedTransaction(t *testing.T, ctx context.Context, client *e2eMCPClient, tool string, args map[string]any) protocol.PreparedTransactionResult {
	t.Helper()

	raw, err := client.callTool(ctx, tool, args)
	if err != nil {
		t.Fatalf("%s: %v", tool, err)
	}
	content := mcpRequireStructuredContent(t, raw)
	encoded, err := json.Marshal(content)
	if err != nil {
		t.Fatalf("marshal structuredContent: %v", err)
	}

	var result protocol.PreparedTransactionResult
	if err := json.Unmarshal(encoded, &result); err != nil {
		t.Fatalf("decode prepared transaction result: %v; content=%s", err, string(encoded))
	}
	if result.UnsignedTransaction.Method == "" {
		t.Fatalf("%s returned empty unsigned_transaction.method", tool)
	}
	return result
}

func mcpSignAndSubmit(t *testing.T, ctx context.Context, client *e2eMCPClient, wallet interface {
	SignPreparedTransaction(protocol.PreparedTransaction) (protocol.SignedTransaction, error)
}, tx protocol.PreparedTransaction) map[string]any {
	t.Helper()

	signed, err := wallet.SignPreparedTransaction(tx)
	if err != nil {
		t.Fatalf("SignPreparedTransaction: %v", err)
	}
	if signed.Hash == "" || signed.Signature == "" {
		t.Fatalf("signed transaction missing hash/signature")
	}

	raw, err := client.callTool(ctx, "finance.transaction.submit_signed", map[string]any{"transaction": signed})
	if err != nil {
		t.Fatalf("finance.transaction.submit_signed: %v", err)
	}
	return mcpRequireStructuredContent(t, raw)
}

func mcpRequireStructuredContent(t *testing.T, raw json.RawMessage) map[string]any {
	t.Helper()

	var payload struct {
		StructuredContent map[string]any `json:"structuredContent"`
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	if err := decoder.Decode(&payload); err != nil {
		t.Fatalf("decode tool result: %v; raw=%s", err, string(raw))
	}
	if payload.StructuredContent == nil {
		t.Fatalf("missing structuredContent; raw=%s", string(raw))
	}
	return payload.StructuredContent
}

func mcpUnwrapLifecycleState(value any) map[string]any {
	switch v := value.(type) {
	case nil:
		return nil
	case map[string]any:
		if mcpIsLifecycleState(v) {
			return v
		}
		if object, ok := v["object"].(map[string]any); ok && mcpIsLifecycleState(object) {
			return object
		}
		for _, key := range []string{"state", "lifecycle_state", "result", "data"} {
			if found := mcpUnwrapLifecycleState(v[key]); found != nil {
				return found
			}
		}
		for _, item := range v {
			if found := mcpUnwrapLifecycleState(item); found != nil {
				return found
			}
		}
	case []any:
		for _, item := range v {
			if found := mcpUnwrapLifecycleState(item); found != nil {
				return found
			}
		}
	}
	return nil
}

func mcpIsLifecycleState(state map[string]any) bool {
	if state == nil {
		return false
	}
	if _, ok := state["next_command"].(map[string]any); ok {
		return true
	}
	return state["request_id"] != nil || state["status"] != nil || state["current_step"] != nil
}

func mcpLen(value any) int {
	switch v := value.(type) {
	case []any:
		return len(v)
	default:
		return 0
	}
}

func mcpHasMap(value any) bool {
	_, ok := value.(map[string]any)
	return ok
}
