package com.twofinance.sdk;

import com.sun.net.httpserver.HttpServer;
import java.net.InetSocketAddress;
import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.ArrayList;
import java.util.Collections;
import java.util.List;
import java.util.Map;
import java.util.concurrent.atomic.AtomicInteger;

public final class SdkClientTest {
    public static void main(String[] args) throws Exception {
        if (!"2finance-sdk-client".equals(SDKMetadata.SDK_NAME)) {
            throw new AssertionError("SDK name should be public");
        }
        if (!"0.1.0".equals(SDKMetadata.SDK_VERSION)) {
            throw new AssertionError("SDK version should be public");
        }
        if (SDKMetadata.serviceCatalog().services().size() != 12) {
            throw new AssertionError("SDK service catalog should expose canonical services");
        }
        if (!"TWO_FINANCE_AUTH_URL".equals(SDKMetadata.serviceCatalog().services().get(0).env())) {
            throw new AssertionError("SDK service catalog should expose env vars");
        }
        Models.DomainOperationsCatalog operationsCatalog = new Models.DomainOperationsCatalog(
                "sdk.domain_operations.v2",
                List.of(new Models.DomainOperationsDomain(
                        "auth",
                        "TWO_FINANCE_AUTH_URL",
                        "http",
                        "User auth and client credentials token flows.",
                        List.of(new Models.DomainOperation(
                                "login",
                                "POST",
                                "/v1/2finance-authenticator/{realm}/{client_id}/login",
                                List.of("realm", "client_id"),
                                List.of(),
                                "auth.login.request.v1",
                                "auth.token.response.v1",
                                null
                        ))
                ))
        );
        if (!"auth.login.request.v1".equals(operationsCatalog.domains().get(0).operations().get(0).requestSchema())) {
            throw new AssertionError("domain operations model should expose request schema");
        }
        if (!operationsCatalog.operation("auth", "login").isPresent()) {
            throw new AssertionError("domain operations model should locate operation");
        }
        Models.ResolvedOperation resolvedLogin = operationsCatalog.operation("auth", "login").get().resolve(
                Map.of("realm", "2finance", "client_id", "client/1"),
                Map.of("ignored", "drop-me")
        );
        if (!"POST".equals(resolvedLogin.method())
                || !"/v1/2finance-authenticator/2finance/client%2F1/login".equals(resolvedLogin.path())) {
            throw new AssertionError("domain operations model should resolve path params");
        }
        if (!resolvedLogin.path().equals(operationsCatalog.resolveOperation(
                "auth",
                "login",
                Map.of("realm", "2finance", "client_id", "client/1"),
                Map.of()
        ).path())) {
            throw new AssertionError("domain operations catalog should resolve operation by name");
        }
        Models.DomainOperation riskOperation = new Models.DomainOperation(
                "black_scholes",
                "get",
                "/risk-manager/blackscholes",
                List.of(),
                List.of("symbol", "strike", "volatility"),
                null,
                null,
                null
        );
        if (!"/risk-manager/blackscholes?symbol=BTC%2FUSD&strike=100000&volatility=0.5".equals(
                riskOperation.resolve(Map.of(), Map.of(
                        "symbol", "BTC/USD",
                        "strike", 100000,
                        "ignored", "drop-me",
                        "volatility", 0.5
                )).path())) {
            throw new AssertionError("domain operations model should resolve contract query params");
        }
        assertSharedContractFixtures();
        if (!"Bearer abc".equals(Auth.bearerAuthorization("abc"))) {
            throw new AssertionError("bearer token was not normalized");
        }
        SDKConfig config = SDKConfig.fromEnv(Map.of(
                "TWO_FINANCE_AUTH_URL", "https://auth.example",
                "TWO_FINANCE_ANALYTICS_URL", "https://analytics.example",
                "TWO_FINANCE_KEYSTORE_URL", "https://keys.example",
                "TWO_FINANCE_MATCHENGINE_WS_URL", "wss://matchengine.example/ws"
        ));
        if (!"https://auth.example".equals(config.authUrl)) {
            throw new AssertionError("auth URL not loaded from env map");
        }
        if (!"https://analytics.example".equals(config.serviceUrl("analytics"))) {
            throw new AssertionError("analytics URL not resolved by serviceUrl");
        }
        if (!"wss://matchengine.example/ws".equals(config.serviceUrl("match_engine"))) {
            throw new AssertionError("matchengine URL not resolved by serviceUrl");
        }
        if (!"wss://matchengine.example/ws".equals(config.serviceUrls().get("matchengine"))) {
            throw new AssertionError("matchengine URL not listed by serviceUrls");
        }
        if (!"matchengine".equals(config.configuredServices().get(1).name())) {
            throw new AssertionError("configured services should preserve catalog order");
        }
        if (!"network".equals(config.missingServiceUrls().get(0).name())) {
            throw new AssertionError("missing services should preserve catalog order");
        }
        SdkClient client = new SdkClient(config);
        if (!"https://analytics.example/analytics/indicators".equals(client.analytics.url("/analytics/indicators"))) {
            throw new AssertionError("analytics URL was not resolved correctly");
        }
        if (!"wss://matchengine.example/ws".equals(client.matchEngine.webSocketUrl)) {
            throw new AssertionError("matchengine URL was not loaded");
        }
        if (!client.matchEngine.orderCommand("{\"symbol\":\"BTC-USDT\"}").contains("matchengine.order_command.v2")) {
            throw new AssertionError("matchengine schema was not defaulted");
        }
        if (!client.matchEngine.marketDataSubscribe("{\"symbols\":[\"BTC-USDT\"]}").contains("matchengine.market_data_subscribe.v2")) {
            throw new AssertionError("matchengine market data schema was not defaulted");
        }
        String sentOrder = client.matchEngine.sendOrder(message -> "sent:" + message, "{\"symbol\":\"BTC-USDT\"}");
        if (!sentOrder.contains("matchengine.order_command.v2")) {
            throw new AssertionError("matchengine sendOrder should send defaulted order payload");
        }
        String sentSubscription = client.matchEngine.subscribeMarketData(message -> "sent:" + message, "{\"symbols\":[\"BTC-USDT\"]}");
        if (!sentSubscription.contains("matchengine.market_data_subscribe.v2")) {
            throw new AssertionError("matchengine subscribeMarketData should send defaulted subscription payload");
        }
        if (!"https://analytics.example/analytics/candles:upsert".equals(client.analytics.url("/analytics/candles:upsert"))) {
            throw new AssertionError("analytics candles URL was not resolved correctly");
        }
        if (!"https://analytics.example/portfolio-manager/rankings".equals(client.analytics.url("/portfolio-manager/rankings"))) {
            throw new AssertionError("analytics rankings URL was not resolved correctly");
        }
        if (!"https://keys.example/healthz".equals(client.keystore.url("/healthz"))) {
            throw new AssertionError("keystore health URL was not resolved correctly");
        }
        if (!"https://keys.example/readyz".equals(client.keystore.url("/readyz"))) {
            throw new AssertionError("keystore readiness URL was not resolved correctly");
        }
        if (!"https://keys.example/keystore/tss/metrics".equals(client.keystore.url("/keystore/tss/metrics"))) {
            throw new AssertionError("keystore metrics URL was not resolved correctly");
        }
        RequestOptions requestOptions = new RequestOptions(
                Map.of("X-Trace-ID", "trace-1"),
                " idem-1 ",
                Map.of("symbol", "BTC-USDT"),
                java.time.Duration.ofSeconds(2),
                1,
                2,
                25
        );
        if (!"trace-1".equals(requestOptions.headers().get("X-Trace-ID"))) {
            throw new AssertionError("request options should carry custom headers");
        }
        if (!" idem-1 ".equals(requestOptions.idempotencyKey())) {
            throw new AssertionError("request options should carry idempotency key");
        }
        if (!"BTC-USDT".equals(requestOptions.query().get("symbol"))) {
            throw new AssertionError("request options should carry query params");
        }
        if (requestOptions.page() != 2 || requestOptions.limit() != 25) {
            throw new AssertionError("request options should carry pagination");
        }
        if (!java.time.Duration.ofSeconds(2).equals(requestOptions.timeout())) {
            throw new AssertionError("request options should carry timeout");
        }
        if (requestOptions.maxRetries() != 1) {
            throw new AssertionError("request options should carry max retries");
        }
        ServiceException serviceException = new ServiceException(
                "GET",
                "https://analytics.example/analytics/indicators",
                429,
                "rate limited"
        );
        if (serviceException.statusCode() != 429) {
            throw new AssertionError("service exception should keep status code");
        }
        if (!"rate limited".equals(serviceException.body())) {
            throw new AssertionError("service exception should keep body");
        }
        Models.SdkErrorPayload errorPayload = new Models.SdkErrorPayload(
                "rate_limited",
                "Too many requests",
                "HTTP_429",
                Map.of("request_id", "req_2finance_001")
        );
        Models.PaginationResponse pagination = new Models.PaginationResponse(
                List.of(Map.of("id", "robot-001", "status", "running")),
                25,
                "cursor-current",
                "cursor-next"
        );
        Models.IdempotencyRecord idempotency = new Models.IdempotencyRecord(
                "idem-001",
                "matchengine.order_command",
                "client_order_id",
                "req_2finance_001"
        );
        Models.ServiceCatalog catalog = new Models.ServiceCatalog(List.of(
                new Models.ServiceCatalogEntry("auth", "TWO_FINANCE_AUTH_URL")
        ));
        if (!"HTTP_429".equals(errorPayload.code()) || !"req_2finance_001".equals(errorPayload.details().get("request_id"))) {
            throw new AssertionError("SDK error model should keep canonical fields");
        }
        if (pagination.limit() != 25 || !"cursor-next".equals(pagination.nextCursor())) {
            throw new AssertionError("pagination model should keep canonical fields");
        }
        if (!"idem-001".equals(idempotency.idempotencyKey())) {
            throw new AssertionError("idempotency model should keep canonical key");
        }
        if (!"auth".equals(catalog.services().get(0).name())) {
            throw new AssertionError("service catalog model should keep service names");
        }
        assertServiceClientRequestBehavior();
        assertAuthOIDCHelperBehavior();
        assertDomainPathEncodingBehavior();
        assertHummingbotConnectorConfigBehavior();
        assertProviderClientBehavior();
        assertPlannerOperationalPlanBehavior();
        SDKConfig orchestratorConfig = SDKConfig.fromEnv(Map.of(
                "TWO_FINANCE_ORCHESTRATOR_URL", "https://orchestrator.example",
                "TWO_FINANCE_WISE_URL", "https://wise.example",
                "TWO_FINANCE_AIRWALLEX_URL", "https://airwallex.example"
        ));
        SdkClient orchestratorClient = new SdkClient(orchestratorConfig);
        if (!"https://orchestrator.example/v1/mcphost/tools".equals(orchestratorClient.orchestrator.url("/v1/mcphost/tools"))) {
            throw new AssertionError("orchestrator tools URL was not resolved correctly");
        }
        if (!"https://orchestrator.example/v1/mcphost/sessions/session-1".equals(orchestratorClient.orchestrator.url("/v1/mcphost/sessions/session-1"))) {
            throw new AssertionError("orchestrator session URL was not resolved correctly");
        }
        if (!"https://wise.example/v1/profiles".equals(orchestratorClient.wise.url("/v1/profiles"))) {
            throw new AssertionError("wise URL was not resolved correctly");
        }
        if (!"https://wise.example/v3/profiles/profile-1/quotes".equals(orchestratorClient.wise.url("/v3/profiles/profile-1/quotes"))) {
            throw new AssertionError("wise quote URL was not resolved correctly");
        }
        if (!"https://airwallex.example/api/v1/payments".equals(orchestratorClient.airwallex.url("/api/v1/payments"))) {
            throw new AssertionError("airwallex URL was not resolved correctly");
        }
        StaticTokenSource tokenSource = new StaticTokenSource("token-123");
        if (!"token-123".equals(tokenSource.token())) {
            throw new AssertionError("static token source did not return token");
        }
        ClientCredentialsTokenSource credentials = new ClientCredentialsTokenSource(
                "https://auth.example/token",
                "client-id",
                "client-secret",
                java.util.List.of("analytics:read", "mcp:invoke")
        );
        if (credentials == null) {
            throw new AssertionError("client credentials source should instantiate");
        }
    }

    private static void assertSharedContractFixtures() throws Exception {
        String domains = Files.readString(Path.of("..", "contracts", "examples", "domain-operations.json"));
        String error = Files.readString(Path.of("..", "contracts", "examples", "error.json"));
        String pagination = Files.readString(Path.of("..", "contracts", "examples", "pagination.json"));
        String idempotency = Files.readString(Path.of("..", "contracts", "examples", "idempotency.json"));

        assertContains(domains, "\"schema\": \"sdk.domain_operations.v2\"");
        assertContains(domains, "\"name\": \"planner\"");
        assertContains(domains, "\"name\": \"trading_plan\"");
        assertContains(domains, "\"path\": \"/portfolio-manager/balances/{account_id}\"");
        assertContains(domains, "\"path_params\": [\"account_id\"]");
        assertContains(error, "\"error\": \"rate_limited\"");
        assertContains(error, "\"code\": \"HTTP_429\"");
        assertContains(pagination, "\"next_cursor\": \"cursor-next\"");
        assertContains(idempotency, "\"idempotency_key\": \"idem-001\"");
    }

    private static void assertContains(String source, String expected) {
        if (!source.contains(expected)) {
            throw new AssertionError("missing contract fixture content: " + expected);
        }
    }

    private static void assertAuthOIDCHelperBehavior() throws Exception {
        List<String> seen = Collections.synchronizedList(new ArrayList<>());
        HttpServer server = HttpServer.create(new InetSocketAddress("127.0.0.1", 0), 0);
        server.createContext("/", exchange -> {
            String body = new String(exchange.getRequestBody().readAllBytes(), StandardCharsets.UTF_8);
            seen.add(exchange.getRequestMethod() + " " + exchange.getRequestURI().getRawPath() + " " + body);
            byte[] bytes = "{\"ok\":true}".getBytes(StandardCharsets.UTF_8);
            exchange.getResponseHeaders().set("Content-Type", "application/json");
            exchange.sendResponseHeaders(200, bytes.length);
            exchange.getResponseBody().write(bytes);
            exchange.close();
        });
        server.start();
        try {
            SDKConfig config = new SDKConfig();
            config.authUrl = "http://127.0.0.1:" + server.getAddress().getPort();
            SdkClient client = new SdkClient(config);
            client.auth.jwks();
            client.auth.validateToken("{\"token\":\"token-1\"}");
            assertSeen(seen, "GET /realms/2finance/protocol/openid-connect/certs ");
            assertSeen(seen, "POST /realms/2finance/protocol/openid-connect/token/introspect {\"token\":\"token-1\"}");
        } finally {
            server.stop(0);
        }
    }

    private static void assertPlannerOperationalPlanBehavior() throws Exception {
        List<String> seen = Collections.synchronizedList(new ArrayList<>());
        HttpServer server = HttpServer.create(new InetSocketAddress("127.0.0.1", 0), 0);
        server.createContext("/", exchange -> {
            seen.add(exchange.getRequestMethod() + " " + exchange.getRequestURI().getRawPath());
            byte[] bytes = "{\"ok\":true}".getBytes(StandardCharsets.UTF_8);
            exchange.getResponseHeaders().set("Content-Type", "application/json");
            exchange.sendResponseHeaders(200, bytes.length);
            exchange.getResponseBody().write(bytes);
            exchange.close();
        });
        server.start();
        try {
            SDKConfig config = new SDKConfig();
            config.orchestratorUrl = "http://127.0.0.1:" + server.getAddress().getPort();
            SdkClient client = new SdkClient(config);
            client.planner.operationalPlan("{\"message\":\"operate\"}");
            assertSeen(seen, "POST /v1/mcphost/messages");
        } finally {
            server.stop(0);
        }
    }

    private static void assertHummingbotConnectorConfigBehavior() throws Exception {
        List<String> seen = Collections.synchronizedList(new ArrayList<>());
        HttpServer server = HttpServer.create(new InetSocketAddress("127.0.0.1", 0), 0);
        server.createContext("/", exchange -> {
            seen.add(exchange.getRequestMethod() + " " + exchange.getRequestURI().getRawPath());
            byte[] bytes = "{\"ok\":true}".getBytes(StandardCharsets.UTF_8);
            exchange.getResponseHeaders().set("Content-Type", "application/json");
            exchange.sendResponseHeaders(200, bytes.length);
            exchange.getResponseBody().write(bytes);
            exchange.close();
        });
        server.start();
        try {
            SDKConfig config = new SDKConfig();
            config.hummingbotUrl = "http://127.0.0.1:" + server.getAddress().getPort();
            SdkClient client = new SdkClient(config);
            client.hummingbot.connectorConfig("{\"connector\":\"2finance\"}");
            assertSeen(seen, "POST /api/v1/connectors/2finance/config");
        } finally {
            server.stop(0);
        }
    }

    private static void assertProviderClientBehavior() throws Exception {
        List<String> seen = Collections.synchronizedList(new ArrayList<>());
        HttpServer server = HttpServer.create(new InetSocketAddress("127.0.0.1", 0), 0);
        server.createContext("/", exchange -> {
            seen.add(exchange.getRequestMethod() + " " + exchange.getRequestURI().getRawPath());
            byte[] bytes = "{\"ok\":true}".getBytes(StandardCharsets.UTF_8);
            exchange.getResponseHeaders().set("Content-Type", "application/json");
            exchange.sendResponseHeaders(200, bytes.length);
            exchange.getResponseBody().write(bytes);
            exchange.close();
        });
        server.start();
        try {
            String baseURL = "http://127.0.0.1:" + server.getAddress().getPort();
            SDKConfig config = new SDKConfig();
            config.wiseUrl = baseURL;
            config.airwallexUrl = baseURL;
            SdkClient client = new SdkClient(config);
            client.wise.profiles();
            client.wise.profile("profile/1");
            client.wise.createQuote("profile/1", "{\"source\":\"USD\"}");
            client.wise.createTransfer("{\"target\":\"BRL\"}");
            client.airwallex.accounts();
            client.airwallex.payments();
            client.airwallex.createPayment("{\"amount\":10}");
            client.airwallex.beneficiaries();
            client.airwallex.createBeneficiary("{\"name\":\"beneficiary\"}");
            assertSeen(seen, "GET /v1/profiles");
            assertSeen(seen, "GET /v1/profiles/profile%2F1");
            assertSeen(seen, "POST /v3/profiles/profile%2F1/quotes");
            assertSeen(seen, "POST /v1/transfers");
            assertSeen(seen, "GET /api/v1/accounts");
            assertSeen(seen, "GET /api/v1/payments");
            assertSeen(seen, "POST /api/v1/payments");
            assertSeen(seen, "GET /api/v1/beneficiaries");
            assertSeen(seen, "POST /api/v1/beneficiaries");
        } finally {
            server.stop(0);
        }
    }

    private static void assertServiceClientRequestBehavior() throws Exception {
        AtomicInteger attempts = new AtomicInteger();
        List<String> failures = Collections.synchronizedList(new ArrayList<>());
        HttpServer server = HttpServer.create(new InetSocketAddress("127.0.0.1", 0), 0);
        server.createContext("/analytics/indicators", exchange -> {
            attempts.incrementAndGet();
            if (!"/analytics/indicators".equals(exchange.getRequestURI().getPath())) {
                failures.add("unexpected path " + exchange.getRequestURI().getPath());
            }
            String query = exchange.getRequestURI().getRawQuery();
            if (attempts.get() <= 2 && (query == null || !query.contains("symbol=BTC-USDT") || !query.contains("page=2") || !query.contains("limit=25"))) {
                failures.add("request query did not include expected params: " + query);
            }
            if (!"Bearer token-123".equals(exchange.getRequestHeaders().getFirst("Authorization"))) {
                failures.add("authorization header was not injected");
            }
            if (attempts.get() <= 2 && !"idem-1".equals(exchange.getRequestHeaders().getFirst("Idempotency-Key"))) {
                failures.add("idempotency key was not trimmed and injected");
            }
            if (attempts.get() <= 2 && !"trace-1".equals(exchange.getRequestHeaders().getFirst("X-Trace-ID"))) {
                failures.add("custom header was not injected");
            }
            String body = attempts.get() == 1 ? "{\"error\":\"retry\"}" : "{\"ok\":true}";
            int status = attempts.get() == 1 ? 500 : 200;
            byte[] bytes = body.getBytes(StandardCharsets.UTF_8);
            exchange.getResponseHeaders().set("Content-Type", "application/json");
            exchange.sendResponseHeaders(status, bytes.length);
            exchange.getResponseBody().write(bytes);
            exchange.close();
        });
        server.start();
        try {
            SDKConfig config = new SDKConfig();
            config.analyticsUrl = "http://127.0.0.1:" + server.getAddress().getPort();
            config.tokenSource = new StaticTokenSource("token-123");
            SdkClient client = new SdkClient(config);
            String response = client.analytics.get("/analytics/indicators", new RequestOptions(
                    Map.of("X-Trace-ID", "trace-1"),
                    " idem-1 ",
                    Map.of("symbol", "BTC-USDT"),
                    java.time.Duration.ofSeconds(2),
                    1,
                    2,
                    25
            ));
            if (!"{\"ok\":true}".equals(response)) {
                throw new AssertionError("service client should return successful retry body");
            }
            if (attempts.get() != 2) {
                throw new AssertionError("service client should retry one retryable response");
            }
            String resolvedResponse = client.analytics.requestOperation(
                    new Models.ResolvedOperation("GET", "/analytics/indicators"),
                    null
            );
            if (!"{\"ok\":true}".equals(resolvedResponse)) {
                throw new AssertionError("service client should execute resolved operations");
            }
            String catalogResponse = client.analytics.requestCatalogOperation(
                    new Models.DomainOperationsCatalog(
                            "sdk.domain_operations.v2",
                            List.of(new Models.DomainOperationsDomain(
                                    "analytics",
                                    "TWO_FINANCE_ANALYTICS_URL",
                                    "http",
                                    null,
                                    List.of(new Models.DomainOperation(
                                            "indicators",
                                            "GET",
                                            "/analytics/indicators",
                                            List.of(),
                                            List.of(),
                                            null,
                                            null,
                                            null
                                    ))
                            ))
                    ),
                    "analytics",
                    "indicators",
                    Map.of(),
                    Map.of(),
                    null,
                    null
            );
            if (!"{\"ok\":true}".equals(catalogResponse)) {
                throw new AssertionError("service client should execute catalog operations");
            }
            if (!failures.isEmpty()) {
                throw new AssertionError(String.join("; ", failures));
            }
        } finally {
            server.stop(0);
        }
    }

    private static void assertDomainPathEncodingBehavior() throws Exception {
        List<String> seen = Collections.synchronizedList(new ArrayList<>());
        HttpServer server = HttpServer.create(new InetSocketAddress("127.0.0.1", 0), 0);
        server.createContext("/", exchange -> {
            String target = exchange.getRequestMethod() + " " + exchange.getRequestURI().getRawPath();
            String query = exchange.getRequestURI().getRawQuery();
            if (query != null && !query.isBlank()) {
                target += "?" + query;
            }
            seen.add(target);
            byte[] bytes = "{\"ok\":true}".getBytes(StandardCharsets.UTF_8);
            exchange.getResponseHeaders().set("Content-Type", "application/json");
            exchange.sendResponseHeaders(200, bytes.length);
            exchange.getResponseBody().write(bytes);
            exchange.close();
        });
        server.start();
        try {
            String baseURL = "http://127.0.0.1:" + server.getAddress().getPort();
            SDKConfig config = new SDKConfig();
            config.analyticsUrl = baseURL;
            config.networkUrl = baseURL;
            config.orchestratorUrl = baseURL;
            config.tradingControlUrl = baseURL;
            config.keyStoreUrl = baseURL;
            SdkClient client = new SdkClient(config);

            client.analytics.balances("acct/1 ok");
            client.network.marketCandles("BTC/USDT spot", "limit=10");
            client.network.bonds();
            client.network.createBond("{\"symbol\":\"BOND1\"}");
            client.network.loans();
            client.network.createLoan("{\"loan\":\"ln1\"}");
            client.network.swaps();
            client.network.createSwap("{\"pair\":\"BTC-USDT\"}");
            client.network.stakingProducts();
            client.network.createStakingProduct("{\"asset\":\"TWO\"}");
            client.network.syntheticAssets();
            client.network.createSyntheticAsset("{\"asset\":\"sBTC\"}");
            client.network.liquidityPools();
            client.network.createLiquidityPool("{\"pool\":\"BTC-USDT\"}");
            client.tradingControl.pauseRobot("robot/1 ok");
            client.tradingControl.riskView("robot/1 ok");
            client.tradingControl.strategies();
            client.tradingControl.createStrategy("{\"name\":\"mean-reversion\"}");
            client.tradingControl.directives();
            client.tradingControl.createDirective("{\"action\":\"rebalance\"}");
            client.tradingControl.audit();
            client.tradingControl.activity();
            client.tradingControl.mcpTools();
            client.keystore.keys("pub/1 ok");
            client.keystore.signatures("pub/1 ok");
            client.orchestrator.deleteSession("session/1 ok");

            assertSeen(seen, "GET /portfolio-manager/balances/acct%2F1%20ok");
            assertSeen(seen, "GET /v2/2finance-network/markets/BTC%2FUSDT%20spot/candles?limit=10");
            assertSeen(seen, "GET /v2/2finance-network/products/bonds");
            assertSeen(seen, "POST /v2/2finance-network/products/bonds");
            assertSeen(seen, "GET /v2/2finance-network/products/loans");
            assertSeen(seen, "POST /v2/2finance-network/products/loans");
            assertSeen(seen, "GET /v2/2finance-network/products/swaps");
            assertSeen(seen, "POST /v2/2finance-network/products/swaps");
            assertSeen(seen, "GET /v2/2finance-network/products/staking");
            assertSeen(seen, "POST /v2/2finance-network/products/staking");
            assertSeen(seen, "GET /v2/2finance-network/products/synthetic-assets");
            assertSeen(seen, "POST /v2/2finance-network/products/synthetic-assets");
            assertSeen(seen, "GET /v2/2finance-network/products/liquidity-pools");
            assertSeen(seen, "POST /v2/2finance-network/products/liquidity-pools");
            assertSeen(seen, "POST /robots/robot%2F1%20ok:pause");
            assertSeen(seen, "GET /risk-view/robot%2F1%20ok");
            assertSeen(seen, "GET /strategies");
            assertSeen(seen, "POST /strategies");
            assertSeen(seen, "GET /directives");
            assertSeen(seen, "POST /directives");
            assertSeen(seen, "GET /audit");
            assertSeen(seen, "GET /activity");
            assertSeen(seen, "GET /mcp/tools");
            assertSeen(seen, "GET /keystore/keys/pub%2F1%20ok");
            assertSeen(seen, "GET /keystore/signatures/pub%2F1%20ok");
            assertSeen(seen, "DELETE /v1/mcphost/sessions/session%2F1%20ok");
        } finally {
            server.stop(0);
        }
    }

    private static void assertSeen(List<String> seen, String expected) {
        if (!seen.contains(expected)) {
            throw new AssertionError("missing request " + expected + " in " + seen);
        }
    }
}
