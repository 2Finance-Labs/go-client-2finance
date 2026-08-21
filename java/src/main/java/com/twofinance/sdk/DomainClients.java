package com.twofinance.sdk;

import java.io.IOException;
import java.net.URLEncoder;
import java.net.http.HttpClient;
import java.nio.charset.StandardCharsets;
import java.util.function.Function;

public final class DomainClients {
    private DomainClients() {}

    public static class AuthClient extends ServiceClient {
        private final String realm;
        private final String clientId;
        private final String phoneClientId;

        public AuthClient(String baseUrl, HttpClient httpClient, TokenSource tokenSource, SDKConfig config) {
            super(baseUrl, httpClient, tokenSource);
            this.realm = config.authRealm;
            this.clientId = config.authClientId;
            this.phoneClientId = config.authPhoneClientId;
        }

        public String login(String jsonBody) throws IOException, InterruptedException {
            return post(authPath(clientId, "/login"), jsonBody);
        }

        public String refreshToken(String jsonBody) throws IOException, InterruptedException {
            return post(authPath(clientId, "/refresh"), jsonBody);
        }

        public String phoneLogin(String jsonBody) throws IOException, InterruptedException {
            return post(authPath(phoneClientId, "/phone/sms/login"), jsonBody);
        }

        public String jwks() throws IOException, InterruptedException {
            return get(oidcPath("/protocol/openid-connect/certs"));
        }

        public String validateToken(String jsonBody) throws IOException, InterruptedException {
            return post(oidcPath("/protocol/openid-connect/token/introspect"), jsonBody);
        }

        private String authPath(String selectedClientId, String endpoint) {
            return "/v1/2finance-authenticator/" + realm + "/" + selectedClientId + "/" + endpoint.replaceFirst("^/+", "");
        }

        private String oidcPath(String endpoint) {
            return "/realms/" + realm + "/" + endpoint.replaceFirst("^/+", "");
        }
    }

    public static class AnalyticsClient extends ServiceClient {
        public AnalyticsClient(String baseUrl, HttpClient httpClient, TokenSource tokenSource) {
            super(baseUrl, httpClient, tokenSource);
        }

        public String indicators() throws IOException, InterruptedException {
            return get("/analytics/indicators");
        }

        public String calculateTechnicalAnalysis(String jsonBody) throws IOException, InterruptedException {
            return post("/analytics/technical-analysis:calculate", jsonBody);
        }

        public String optimizePortfolio(String jsonBody) throws IOException, InterruptedException {
            return post("/portfolio-manager/optimizer", jsonBody);
        }

        public String upsertCandles(String jsonBody) throws IOException, InterruptedException {
            return post("/analytics/candles:upsert", jsonBody);
        }

        public String rankings() throws IOException, InterruptedException {
            return get("/portfolio-manager/rankings");
        }

        public String balances(String accountId) throws IOException, InterruptedException {
            return get("/portfolio-manager/balances/" + encode(accountId));
        }

        public String blackScholes(String query) throws IOException, InterruptedException {
            return get("/risk-manager/blackscholes" + (query == null || query.isBlank() ? "" : "?" + query));
        }

        public String staking() throws IOException, InterruptedException {
            return get("/staking");
        }
    }

    public static class MCPClient extends ServiceClient {
        private long nextId = 1;

        public MCPClient(String baseUrl, HttpClient httpClient, TokenSource tokenSource) {
            super(baseUrl, httpClient, tokenSource);
        }

        public String call(String method, String paramsJson) throws IOException, InterruptedException {
            String params = paramsJson == null || paramsJson.isBlank() ? "null" : paramsJson;
            String body = "{\"jsonrpc\":\"2.0\",\"id\":" + nextId++ + ",\"method\":\"" + method + "\",\"params\":" + params + "}";
            return post("/mcp", body);
        }

        public String conversationPlan(String argumentsJson) throws IOException, InterruptedException {
            String args = argumentsJson == null || argumentsJson.isBlank() ? "{}" : argumentsJson;
            return call("tools/call", "{\"name\":\"finance_assistant.conversation.plan\",\"arguments\":" + args + "}");
        }

        public String listTools() throws IOException, InterruptedException {
            return call("tools/list", null);
        }

        public String listPrompts() throws IOException, InterruptedException {
            return call("prompts/list", null);
        }

        public String listResources() throws IOException, InterruptedException {
            return call("resources/list", null);
        }

        public String readResource(String uri) throws IOException, InterruptedException {
            return call("resources/read", "{\"uri\":\"" + escape(uri) + "\"}");
        }

        public String callTool(String name, String argumentsJson) throws IOException, InterruptedException {
            String args = argumentsJson == null || argumentsJson.isBlank() ? "{}" : argumentsJson;
            return call("tools/call", "{\"name\":\"" + escape(name) + "\",\"arguments\":" + args + "}");
        }

        public String getPrompt(String name, String argumentsJson) throws IOException, InterruptedException {
            String args = argumentsJson == null || argumentsJson.isBlank() ? "{}" : argumentsJson;
            return call("prompts/get", "{\"name\":\"" + escape(name) + "\",\"arguments\":" + args + "}");
        }
    }

    public static class OrchestratorClient extends ServiceClient {
        public OrchestratorClient(String baseUrl, HttpClient httpClient, TokenSource tokenSource) {
            super(baseUrl, httpClient, tokenSource);
        }

        public String catalog() throws IOException, InterruptedException {
            return get("/v1/mcphost/catalog/packages");
        }

        public String createSession(String jsonBody) throws IOException, InterruptedException {
            return post("/v1/mcphost/sessions", jsonBody);
        }

        public String sendMessage(String jsonBody) throws IOException, InterruptedException {
            return post("/v1/mcphost/messages", jsonBody);
        }

        public String tools() throws IOException, InterruptedException {
            return get("/v1/mcphost/tools");
        }

        public String prompts() throws IOException, InterruptedException {
            return get("/v1/mcphost/prompts");
        }

        public String resources() throws IOException, InterruptedException {
            return get("/v1/mcphost/resources");
        }

        public String providers() throws IOException, InterruptedException {
            return get("/v1/mcphost/providers");
        }

        public String approvals() throws IOException, InterruptedException {
            return get("/v1/mcphost/approvals");
        }

        public String observability() throws IOException, InterruptedException {
            return get("/v1/mcphost/observability");
        }

        public String deleteSession(String id) throws IOException, InterruptedException {
            return delete("/v1/mcphost/sessions/" + encode(id));
        }
    }

    public static class NetworkClient extends ServiceClient {
        public NetworkClient(String baseUrl, HttpClient httpClient, TokenSource tokenSource) {
            super(baseUrl, httpClient, tokenSource);
        }

        public String virtualMachine() throws IOException, InterruptedException {
            return post(
                "/v2/2finance-network/query",
                "{\"method\":\"get_blocks\",\"params\":{\"page\":1,\"limit\":1}}"
            );
        }

        public String marketCandles(String market, String query) throws IOException, InterruptedException {
            return get("/v2/2finance-network/markets/" + encode(market) + "/candles" + (query == null || query.isBlank() ? "" : "?" + query));
        }

        public String products(String productType) throws IOException, InterruptedException {
            return get("/v2/2finance-network/products/" + encode(productType));
        }

        public String createProduct(String productType, String jsonBody) throws IOException, InterruptedException {
            return post("/v2/2finance-network/products/" + encode(productType), jsonBody);
        }

        public String bonds() throws IOException, InterruptedException {
            return products("bonds");
        }

        public String createBond(String jsonBody) throws IOException, InterruptedException {
            return createProduct("bonds", jsonBody);
        }

        public String loans() throws IOException, InterruptedException {
            return products("loans");
        }

        public String createLoan(String jsonBody) throws IOException, InterruptedException {
            return createProduct("loans", jsonBody);
        }

        public String swaps() throws IOException, InterruptedException {
            return products("swaps");
        }

        public String createSwap(String jsonBody) throws IOException, InterruptedException {
            return createProduct("swaps", jsonBody);
        }

        public String stakingProducts() throws IOException, InterruptedException {
            return products("staking");
        }

        public String createStakingProduct(String jsonBody) throws IOException, InterruptedException {
            return createProduct("staking", jsonBody);
        }

        public String syntheticAssets() throws IOException, InterruptedException {
            return products("synthetic-assets");
        }

        public String createSyntheticAsset(String jsonBody) throws IOException, InterruptedException {
            return createProduct("synthetic-assets", jsonBody);
        }

        public String liquidityPools() throws IOException, InterruptedException {
            return products("liquidity-pools");
        }

        public String createLiquidityPool(String jsonBody) throws IOException, InterruptedException {
            return createProduct("liquidity-pools", jsonBody);
        }
    }

    public static class TradingControlClient extends ServiceClient {
        public TradingControlClient(String baseUrl, HttpClient httpClient, TokenSource tokenSource) {
            super(baseUrl, httpClient, tokenSource);
        }

        public String robots() throws IOException, InterruptedException {
            return get("/robots");
        }

        public String createRobot(String jsonBody) throws IOException, InterruptedException {
            return post("/robots", jsonBody);
        }

        public String robot(String id) throws IOException, InterruptedException {
            return get("/robots/" + encode(id));
        }

        public String startRobot(String id) throws IOException, InterruptedException {
            return post("/robots/" + encode(id) + ":start", null);
        }

        public String pauseRobot(String id) throws IOException, InterruptedException {
            return post("/robots/" + encode(id) + ":pause", null);
        }

        public String resumeRobot(String id) throws IOException, InterruptedException {
            return post("/robots/" + encode(id) + ":resume", null);
        }

        public String stopRobot(String id) throws IOException, InterruptedException {
            return post("/robots/" + encode(id) + ":stop", null);
        }

        public String riskPolicy(String id) throws IOException, InterruptedException {
            return get("/robots/" + encode(id) + "/risk-policy");
        }

        public String setRiskPolicy(String id, String jsonBody) throws IOException, InterruptedException {
            return put("/robots/" + encode(id) + "/risk-policy", jsonBody);
        }

        public String riskView(String id) throws IOException, InterruptedException {
            return get("/risk-view/" + encode(id));
        }

        public String strategies() throws IOException, InterruptedException {
            return get("/strategies");
        }

        public String createStrategy(String jsonBody) throws IOException, InterruptedException {
            return post("/strategies", jsonBody);
        }

        public String directives() throws IOException, InterruptedException {
            return get("/directives");
        }

        public String createDirective(String jsonBody) throws IOException, InterruptedException {
            return post("/directives", jsonBody);
        }

        public String audit() throws IOException, InterruptedException {
            return get("/audit");
        }

        public String activity() throws IOException, InterruptedException {
            return get("/activity");
        }

        public String mcpTools() throws IOException, InterruptedException {
            return get("/mcp/tools");
        }
    }

    public static class KeyStoreClient extends ServiceClient {
        public KeyStoreClient(String baseUrl, HttpClient httpClient, TokenSource tokenSource) {
            super(baseUrl, httpClient, tokenSource);
        }

        public String health() throws IOException, InterruptedException {
            return get("/healthz");
        }

        public String readiness() throws IOException, InterruptedException {
            return get("/readyz");
        }

        public String startKeygen(String jsonBody) throws IOException, InterruptedException {
            return post("/keystore/keygen/start", jsonBody);
        }

        public String keygenSignature(String jsonBody) throws IOException, InterruptedException {
            return post("/keystore/keygen/signature", jsonBody);
        }

        public String startSigning(String jsonBody) throws IOException, InterruptedException {
            return post("/keystore/signing/start", jsonBody);
        }

        public String signingSignature(String jsonBody) throws IOException, InterruptedException {
            return post("/keystore/signing/signature", jsonBody);
        }

        public String startResharing(String jsonBody) throws IOException, InterruptedException {
            return post("/keystore/resharing/start", jsonBody);
        }

        public String keys(String userPublicKey) throws IOException, InterruptedException {
            return get("/keystore/keys/" + encode(userPublicKey));
        }

        public String signatures(String userPublicKey) throws IOException, InterruptedException {
            return get("/keystore/signatures/" + encode(userPublicKey));
        }

        public String metrics() throws IOException, InterruptedException {
            return get("/keystore/tss/metrics");
        }
    }

    public static class HummingbotClient extends ServiceClient {
        public HummingbotClient(String baseUrl, HttpClient httpClient, TokenSource tokenSource) {
            super(baseUrl, httpClient, tokenSource);
        }

        public String assets() throws IOException, InterruptedException {
            return get("/api/v1/assets");
        }

        public String symbols() throws IOException, InterruptedException {
            return get("/api/v1/symbols");
        }

        public String balances() throws IOException, InterruptedException {
            return get("/api/v1/balances");
        }

        public String connectorConfig(String requestJson) throws IOException, InterruptedException {
            return post("/api/v1/connectors/2finance/config", requestJson);
        }
    }

    public static class ProviderClient extends ServiceClient {
        public ProviderClient(String baseUrl, HttpClient httpClient, TokenSource tokenSource) {
            super(baseUrl, httpClient, tokenSource);
        }
    }

    public static class WiseClient extends ProviderClient {
        public WiseClient(String baseUrl, HttpClient httpClient, TokenSource tokenSource) {
            super(baseUrl, httpClient, tokenSource);
        }

        public String profiles() throws IOException, InterruptedException {
            return get("/v1/profiles");
        }

        public String profile(String profileId) throws IOException, InterruptedException {
            return get("/v1/profiles/" + encode(profileId));
        }

        public String createQuote(String profileId, String jsonBody) throws IOException, InterruptedException {
            return post("/v3/profiles/" + encode(profileId) + "/quotes", jsonBody);
        }

        public String createTransfer(String jsonBody) throws IOException, InterruptedException {
            return post("/v1/transfers", jsonBody);
        }
    }

    public static class AirwallexClient extends ProviderClient {
        public AirwallexClient(String baseUrl, HttpClient httpClient, TokenSource tokenSource) {
            super(baseUrl, httpClient, tokenSource);
        }

        public String accounts() throws IOException, InterruptedException {
            return get("/api/v1/accounts");
        }

        public String payments() throws IOException, InterruptedException {
            return get("/api/v1/payments");
        }

        public String createPayment(String jsonBody) throws IOException, InterruptedException {
            return post("/api/v1/payments", jsonBody);
        }

        public String beneficiaries() throws IOException, InterruptedException {
            return get("/api/v1/beneficiaries");
        }

        public String createBeneficiary(String jsonBody) throws IOException, InterruptedException {
            return post("/api/v1/beneficiaries", jsonBody);
        }
    }

    public static class MatchEngineClient {
        public final String webSocketUrl;

        public MatchEngineClient(String webSocketUrl) {
            this.webSocketUrl = webSocketUrl == null ? "" : webSocketUrl.trim();
        }

        public String orderCommand(String commandJson) {
            String body = commandJson == null || commandJson.isBlank() ? "{}" : commandJson.trim();
            if (body.equals("{}")) {
                return "{\"schema\":\"matchengine.order_command.v2\",\"message_type\":\"ORDER\",\"operation\":\"ADD\"}";
            }
            return "{\"schema\":\"matchengine.order_command.v2\",\"message_type\":\"ORDER\",\"operation\":\"ADD\"," + body.replaceFirst("^\\{", "");
        }

        public String marketDataSubscribe(String requestJson) {
            String body = requestJson == null || requestJson.isBlank() ? "{}" : requestJson.trim();
            if (body.equals("{}")) {
                return "{\"schema\":\"matchengine.market_data_subscribe.v2\"}";
            }
            return "{\"schema\":\"matchengine.market_data_subscribe.v2\"," + body.replaceFirst("^\\{", "");
        }

        public String sendOrder(Function<String, String> sender, String commandJson) {
            return sender.apply(orderCommand(commandJson));
        }

        public String subscribeMarketData(Function<String, String> sender, String requestJson) {
            return sender.apply(marketDataSubscribe(requestJson));
        }
    }

    public static class PlannerClient {
        private final MCPClient mcp;
        private final OrchestratorClient orchestrator;
        private final AnalyticsClient analytics;
        private final TradingControlClient tradingControl;

        public PlannerClient(
                MCPClient mcp,
                OrchestratorClient orchestrator,
                AnalyticsClient analytics,
                TradingControlClient tradingControl
        ) {
            this.mcp = mcp;
            this.orchestrator = orchestrator;
            this.analytics = analytics;
            this.tradingControl = tradingControl;
        }

        public String conversationPlan(String argumentsJson) throws IOException, InterruptedException {
            return mcp.conversationPlan(argumentsJson);
        }

        public String orchestratedPlan(String jsonBody) throws IOException, InterruptedException {
            return orchestrator.sendMessage(jsonBody);
        }

        public String operationalPlan(String jsonBody) throws IOException, InterruptedException {
            return orchestratedPlan(jsonBody);
        }

        public String tradingPlan(String requestJson, boolean useAnalytics, boolean useTrading)
                throws IOException, InterruptedException {
            StringBuilder context = new StringBuilder("{");
            boolean[] hasContext = {false};
            if (useTrading && tradingControl != null) {
                try {
                    appendContextValue(context, hasContext, "trading_robots", tradingControl.robots());
                } catch (IOException | InterruptedException error) {
                    if (error instanceof InterruptedException) {
                        Thread.currentThread().interrupt();
                    }
                }
            }
            if (useAnalytics && analytics != null) {
                try {
                    appendContextValue(context, hasContext, "analytics_indicators", analytics.indicators());
                } catch (IOException | InterruptedException error) {
                    if (error instanceof InterruptedException) {
                        Thread.currentThread().interrupt();
                    }
                }
            }
            context.append("}");
            return conversationPlan("{\"request\":" + jsonValueOrString(requestJson) + ",\"context\":" + context + "}");
        }

        private static void appendContextValue(
                StringBuilder context,
                boolean[] hasContext,
                String key,
                String value
        ) {
            if (hasContext[0]) {
                context.append(",");
            }
            context.append("\"").append(escape(key)).append("\":").append(jsonValueOrString(value));
            hasContext[0] = true;
        }

        private static String jsonValueOrString(String value) {
            String trimmed = value == null ? "" : value.trim();
            if (trimmed.isEmpty()) {
                return "null";
            }
            char first = trimmed.charAt(0);
            if (first == '{'
                    || first == '['
                    || first == '"'
                    || first == '-'
                    || Character.isDigit(first)
                    || trimmed.equals("true")
                    || trimmed.equals("false")
                    || trimmed.equals("null")) {
                return trimmed;
            }
            return "\"" + escape(trimmed) + "\"";
        }
    }

    private static String encode(String value) {
        return URLEncoder.encode(value == null ? "" : value, StandardCharsets.UTF_8).replace("+", "%20");
    }

    private static String escape(String value) {
        return (value == null ? "" : value).replace("\\", "\\\\").replace("\"", "\\\"");
    }
}
