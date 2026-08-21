package twofinance

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/2Finance-Labs/2finance-sdk-client/matchengine"
)

type testTokenSource string

func (s testTokenSource) Token(context.Context) (string, error) {
	return string(s), nil
}

func TestConfigFromEnv(t *testing.T) {
	if SDKName != "2finance-sdk-client" {
		t.Fatalf("SDKName = %q", SDKName)
	}
	if SDKVersion != "0.1.0" {
		t.Fatalf("SDKVersion = %q", SDKVersion)
	}
	catalog := DefaultServiceCatalog()
	if len(catalog.Services) != 12 {
		t.Fatalf("catalog service count = %d", len(catalog.Services))
	}
	if catalog.Services[0].Name != "auth" || catalog.Services[0].Env != "TWO_FINANCE_AUTH_URL" {
		t.Fatalf("unexpected first service catalog entry: %+v", catalog.Services[0])
	}

	t.Setenv("TWO_FINANCE_AUTH_URL", "https://auth.example")
	t.Setenv("TWO_FINANCE_ANALYTICS_URL", "https://analytics.example")
	t.Setenv("TWO_FINANCE_MATCHENGINE_WS_URL", "wss://matchengine.example/ws")

	config := ConfigFromEnv()
	if config.AuthURL != "https://auth.example" {
		t.Fatalf("AuthURL = %q", config.AuthURL)
	}
	if config.AnalyticsURL != "https://analytics.example" {
		t.Fatalf("AnalyticsURL = %q", config.AnalyticsURL)
	}
	if config.ServiceURL("analytics") != "https://analytics.example" {
		t.Fatalf("ServiceURL(analytics) = %q", config.ServiceURL("analytics"))
	}
	if config.ServiceURL("match_engine") != "wss://matchengine.example/ws" {
		t.Fatalf("ServiceURL(match_engine) = %q", config.ServiceURL("match_engine"))
	}
	urls := config.ServiceURLs()
	if urls["analytics"] != "https://analytics.example" || urls["matchengine"] != "wss://matchengine.example/ws" {
		t.Fatalf("ServiceURLs = %#v", urls)
	}
	services := config.ConfiguredServices()
	if len(services) != 3 || services[1].Name != "analytics" || services[1].URL != "https://analytics.example" {
		t.Fatalf("ConfiguredServices = %#v", services)
	}
	missing := config.MissingServiceURLs()
	if len(missing) != 9 || missing[0].Name != "network" || missing[0].Env != "TWO_FINANCE_NETWORK_URL" {
		t.Fatalf("MissingServiceURLs = %#v", missing)
	}
	if config.MatchEngineWSURL != "wss://matchengine.example/ws" {
		t.Fatalf("MatchEngineWSURL = %q", config.MatchEngineWSURL)
	}
	if config.AuthRealm == "" || config.AuthClientID == "" || config.AuthPhoneClientID == "" {
		t.Fatalf("expected auth defaults to be populated")
	}
}

func TestNewBuildsDomainClients(t *testing.T) {
	client := New(Config{
		AuthURL:           "https://auth.example",
		NetworkURL:        "https://network.example",
		AnalyticsURL:      "https://analytics.example",
		OrchestratorURL:   "https://orchestrator.example",
		MCPURL:            "https://mcp.example",
		TradingControlURL: "https://trading.example",
		MatchEngineWSURL:  "wss://matchengine.example/ws",
		KeyStoreURL:       "https://keystore.example",
		WiseURL:           "https://wise.example",
		AirwallexURL:      "https://airwallex.example",
	})

	if client.Auth == nil || client.Network == nil || client.Analytics == nil || client.Planner == nil || client.Providers == nil {
		t.Fatalf("expected core domain clients to be initialized")
	}
	if client.Providers.Wise == nil || client.Providers.Airwallex == nil {
		t.Fatalf("expected provider clients to be initialized")
	}
	if client.MatchEngine.WebSocketURL != "wss://matchengine.example/ws" {
		t.Fatalf("unexpected matchengine URL: %q", client.MatchEngine.WebSocketURL)
	}
}

func TestAuthenticatedHTTPClientInjectsBearer(t *testing.T) {
	var gotAuth string
	base := &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		gotAuth = req.Header.Get("Authorization")
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body:       http.NoBody,
			Request:    req,
		}, nil
	})}
	client := authenticatedHTTPClient(base, testTokenSource("abc123"))

	req, err := http.NewRequest(http.MethodGet, "https://example.test", nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := client.Do(req); err != nil {
		t.Fatal(err)
	}
	if gotAuth != "Bearer abc123" {
		t.Fatalf("Authorization = %q", gotAuth)
	}
}

func TestDomainClientsEscapePathParamsAndExposeOrchestratorEndpoints(t *testing.T) {
	var seen []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen = append(seen, r.Method+" "+r.RequestURI)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	client := New(Config{
		AuthURL:           server.URL,
		NetworkURL:        server.URL,
		AnalyticsURL:      server.URL,
		OrchestratorURL:   server.URL,
		TradingControlURL: server.URL,
		KeyStoreURL:       server.URL,
		HummingbotURL:     server.URL,
		WiseURL:           server.URL,
		AirwallexURL:      server.URL,
	})
	ctx := context.Background()

	if _, err := client.Analytics.Balances(ctx, "acct/1"); err != nil {
		t.Fatal(err)
	}
	if _, err := client.Network.MarketCandles(ctx, "BTC/USDT", "limit=10"); err != nil {
		t.Fatal(err)
	}
	if _, err := client.Network.Bonds(ctx); err != nil {
		t.Fatal(err)
	}
	if _, err := client.Network.CreateBond(ctx, map[string]string{"symbol": "BOND1"}); err != nil {
		t.Fatal(err)
	}
	if _, err := client.Network.Loans(ctx); err != nil {
		t.Fatal(err)
	}
	if _, err := client.Network.CreateLoan(ctx, map[string]string{"loan": "ln1"}); err != nil {
		t.Fatal(err)
	}
	if _, err := client.Network.Swaps(ctx); err != nil {
		t.Fatal(err)
	}
	if _, err := client.Network.CreateSwap(ctx, map[string]string{"pair": "BTC-USDT"}); err != nil {
		t.Fatal(err)
	}
	if _, err := client.Network.StakingProducts(ctx); err != nil {
		t.Fatal(err)
	}
	if _, err := client.Network.CreateStakingProduct(ctx, map[string]string{"asset": "TWO"}); err != nil {
		t.Fatal(err)
	}
	if _, err := client.Network.SyntheticAssets(ctx); err != nil {
		t.Fatal(err)
	}
	if _, err := client.Network.CreateSyntheticAsset(ctx, map[string]string{"asset": "sBTC"}); err != nil {
		t.Fatal(err)
	}
	if _, err := client.Network.LiquidityPools(ctx); err != nil {
		t.Fatal(err)
	}
	if _, err := client.Network.CreateLiquidityPool(ctx, map[string]string{"pool": "BTC-USDT"}); err != nil {
		t.Fatal(err)
	}
	if _, err := client.Auth.JWKS(ctx); err != nil {
		t.Fatal(err)
	}
	if _, err := client.Auth.ValidateToken(ctx, "token-1"); err != nil {
		t.Fatal(err)
	}
	if _, err := client.TradingControl.PauseRobot(ctx, "robot/1"); err != nil {
		t.Fatal(err)
	}
	if _, err := client.TradingControl.RiskView(ctx, "robot/1"); err != nil {
		t.Fatal(err)
	}
	if _, err := client.TradingControl.Strategies(ctx); err != nil {
		t.Fatal(err)
	}
	if _, err := client.TradingControl.CreateStrategy(ctx, map[string]string{"name": "mean-reversion"}); err != nil {
		t.Fatal(err)
	}
	if _, err := client.TradingControl.Directives(ctx); err != nil {
		t.Fatal(err)
	}
	if _, err := client.TradingControl.CreateDirective(ctx, map[string]string{"action": "rebalance"}); err != nil {
		t.Fatal(err)
	}
	if _, err := client.TradingControl.Audit(ctx); err != nil {
		t.Fatal(err)
	}
	if _, err := client.TradingControl.Activity(ctx); err != nil {
		t.Fatal(err)
	}
	if _, err := client.TradingControl.MCPTools(ctx); err != nil {
		t.Fatal(err)
	}
	if _, err := client.KeyStore.Keys(ctx, "pub/1"); err != nil {
		t.Fatal(err)
	}
	if _, err := client.KeyStore.Signatures(ctx, "pub/1"); err != nil {
		t.Fatal(err)
	}
	if _, err := client.Hummingbot.ConnectorConfig(ctx, map[string]string{"connector": "2finance"}); err != nil {
		t.Fatal(err)
	}
	if _, err := client.Providers.Wise.Profiles(ctx); err != nil {
		t.Fatal(err)
	}
	if _, err := client.Providers.Wise.Profile(ctx, "profile/1"); err != nil {
		t.Fatal(err)
	}
	if _, err := client.Providers.Wise.CreateQuote(ctx, "profile/1", map[string]string{"source": "USD"}); err != nil {
		t.Fatal(err)
	}
	if _, err := client.Providers.Wise.CreateTransfer(ctx, map[string]string{"target": "BRL"}); err != nil {
		t.Fatal(err)
	}
	if _, err := client.Providers.Airwallex.Accounts(ctx); err != nil {
		t.Fatal(err)
	}
	if _, err := client.Providers.Airwallex.Payments(ctx); err != nil {
		t.Fatal(err)
	}
	if _, err := client.Providers.Airwallex.CreatePayment(ctx, map[string]int{"amount": 10}); err != nil {
		t.Fatal(err)
	}
	if _, err := client.Providers.Airwallex.Beneficiaries(ctx); err != nil {
		t.Fatal(err)
	}
	if _, err := client.Providers.Airwallex.CreateBeneficiary(ctx, map[string]string{"name": "beneficiary"}); err != nil {
		t.Fatal(err)
	}
	if _, err := client.Orchestrator.Providers(ctx); err != nil {
		t.Fatal(err)
	}
	if _, err := client.Orchestrator.Approvals(ctx); err != nil {
		t.Fatal(err)
	}
	if _, err := client.Orchestrator.Observability(ctx); err != nil {
		t.Fatal(err)
	}
	if _, err := client.Orchestrator.DeleteSession(ctx, "session/1"); err != nil {
		t.Fatal(err)
	}
	if _, err := client.Planner.OperationalPlan(ctx, map[string]string{"message": "operate"}); err != nil {
		t.Fatal(err)
	}

	for _, expected := range []string{
		"GET /portfolio-manager/balances/acct%2F1",
		"GET /v2/2finance-network/markets/BTC%2FUSDT/candles?limit=10",
		"GET /v2/2finance-network/products/bonds",
		"POST /v2/2finance-network/products/bonds",
		"GET /v2/2finance-network/products/loans",
		"POST /v2/2finance-network/products/loans",
		"GET /v2/2finance-network/products/swaps",
		"POST /v2/2finance-network/products/swaps",
		"GET /v2/2finance-network/products/staking",
		"POST /v2/2finance-network/products/staking",
		"GET /v2/2finance-network/products/synthetic-assets",
		"POST /v2/2finance-network/products/synthetic-assets",
		"GET /v2/2finance-network/products/liquidity-pools",
		"POST /v2/2finance-network/products/liquidity-pools",
		"GET /realms/2finance/protocol/openid-connect/certs",
		"POST /realms/2finance/protocol/openid-connect/token/introspect",
		"POST /robots/robot%2F1:pause",
		"GET /risk-view/robot%2F1",
		"GET /strategies",
		"POST /strategies",
		"GET /directives",
		"POST /directives",
		"GET /audit",
		"GET /activity",
		"GET /mcp/tools",
		"GET /keystore/keys/pub%2F1",
		"GET /keystore/signatures/pub%2F1",
		"POST /api/v1/connectors/2finance/config",
		"GET /v1/profiles",
		"GET /v1/profiles/profile%2F1",
		"POST /v3/profiles/profile%2F1/quotes",
		"POST /v1/transfers",
		"GET /api/v1/accounts",
		"GET /api/v1/payments",
		"POST /api/v1/payments",
		"GET /api/v1/beneficiaries",
		"POST /api/v1/beneficiaries",
		"GET /v1/mcphost/providers",
		"GET /v1/mcphost/approvals",
		"GET /v1/mcphost/observability",
		"DELETE /v1/mcphost/sessions/session%2F1",
		"POST /v1/mcphost/messages",
	} {
		if !containsString(seen, expected) {
			t.Fatalf("missing request %q in %#v", expected, seen)
		}
	}
}

func TestMatchEngineMarketDataSubscribeDefaultsSchema(t *testing.T) {
	request := clientMatchEngineSubscribeFixture()
	if request.Schema != "matchengine.market_data_subscribe.v2" {
		t.Fatalf("Schema = %q", request.Schema)
	}
	if len(request.Symbols) != 1 || request.Symbols[0] != "BTC-USDT" {
		t.Fatalf("Symbols = %#v", request.Symbols)
	}
	if len(request.Channels) != 1 || request.Channels[0] != "book" {
		t.Fatalf("Channels = %#v", request.Channels)
	}
}

func TestSharedModelsParseContractFixtures(t *testing.T) {
	var sdkError SDKError
	readFixture(t, "error.json", &sdkError)
	if sdkError.Code != "HTTP_429" || sdkError.Details["request_id"] != "req_2finance_001" {
		t.Fatalf("unexpected SDKError fixture: %#v", sdkError)
	}

	var page PaginationResponse
	readFixture(t, "pagination.json", &page)
	if page.Limit != 25 || page.NextCursor != "cursor-next" || len(page.Items) != 1 {
		t.Fatalf("unexpected PaginationResponse fixture: %#v", page)
	}

	var idempotency IdempotencyRecord
	readFixture(t, "idempotency.json", &idempotency)
	if idempotency.IdempotencyKey != "idem-001" || idempotency.Operation != "matchengine.order_command" {
		t.Fatalf("unexpected IdempotencyRecord fixture: %#v", idempotency)
	}

	var catalog ServiceCatalog
	readFixture(t, "service-catalog.json", &catalog)
	if len(catalog.Services) == 0 || catalog.Services[0].Name != "auth" {
		t.Fatalf("unexpected ServiceCatalog fixture: %#v", catalog)
	}

	var operations DomainOperationsCatalog
	readFixture(t, "domain-operations.json", &operations)
	if operations.Schema != "sdk.domain_operations.v2" || len(operations.Domains) == 0 {
		t.Fatalf("unexpected DomainOperationsCatalog fixture: %#v", operations)
	}
	if operations.Domains[0].Operations[0].RequestSchema != "auth.login.request.v1" {
		t.Fatalf("unexpected first domain operation: %#v", operations.Domains[0].Operations[0])
	}
	balances, ok := operations.Operation("analytics", "balances")
	if !ok || balances.Path != "/portfolio-manager/balances/{account_id}" {
		t.Fatalf("unexpected analytics.balances operation: %#v ok=%v", balances, ok)
	}
	resolvedBalances, err := balances.Resolve(map[string]string{"account_id": "acct/1 ok"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if resolvedBalances.Method != "GET" || resolvedBalances.Path != "/portfolio-manager/balances/acct%2F1%20ok" {
		t.Fatalf("unexpected resolved balances operation: %#v", resolvedBalances)
	}
	resolvedFromCatalog, err := operations.ResolveOperation("analytics", "balances", map[string]string{"account_id": "acct/1 ok"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if resolvedFromCatalog != resolvedBalances {
		t.Fatalf("unexpected catalog resolved operation: %#v", resolvedFromCatalog)
	}
	blackScholes, ok := operations.Operation("analytics", "black_scholes")
	if !ok {
		t.Fatalf("expected analytics.black_scholes operation")
	}
	resolvedBlackScholes, err := blackScholes.Resolve(nil, map[string]string{
		"symbol":     "BTC/USD",
		"strike":     "100000",
		"ignored":    "drop-me",
		"volatility": "0.5",
	})
	if err != nil {
		t.Fatal(err)
	}
	if resolvedBlackScholes.Path != "/risk-manager/blackscholes?strike=100000&symbol=BTC%2FUSD&volatility=0.5" {
		t.Fatalf("unexpected resolved black scholes operation: %#v", resolvedBlackScholes)
	}
}

func TestCallOperationDispatchesResolvedHTTP(t *testing.T) {
	var seen string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen = r.Method + " " + r.RequestURI
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	client := New(Config{AnalyticsURL: server.URL})
	catalog := DomainOperationsCatalog{
		Domains: []DomainOperationsDomain{{
			Name: "analytics",
			Operations: []DomainOperation{{
				Name:       "balances",
				Method:     "GET",
				Path:       "/portfolio-manager/balances/{account_id}",
				PathParams: []string{"account_id"},
			}},
		}},
	}
	response, err := client.CallCatalogOperation(
		context.Background(),
		catalog,
		"analytics",
		"balances",
		map[string]string{"account_id": "acct/1 ok"},
		nil,
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}
	if string(response) != `{"ok":true}` {
		t.Fatalf("unexpected response: %s", response)
	}
	if seen != "GET /portfolio-manager/balances/acct%2F1%20ok" {
		t.Fatalf("unexpected resolved request: %s", seen)
	}
}

func clientMatchEngineSubscribeFixture() matchengine.MarketDataSubscribeRequest {
	return matchengine.NewMarketDataSubscribeRequest(matchengine.MarketDataSubscribeRequest{
		Symbols:  []string{"BTC-USDT"},
		Channels: []string{"book"},
	})
}

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func readFixture(t *testing.T, name string, out any) {
	t.Helper()
	raw, err := os.ReadFile("../contracts/examples/" + name)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(raw, out); err != nil {
		t.Fatal(err)
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}
