"use strict";

const test = require("node:test");
const assert = require("node:assert/strict");
const fs = require("node:fs");
const path = require("node:path");
const {
  ClientCredentialsTokenSource,
  SERVICE_CATALOG,
  SDK_NAME,
  SDK_VERSION,
  ServiceError,
  ServiceClient,
  StaticTokenSource,
  TwoFinanceClient,
  bearerAuthorization,
  configFromEnv,
  configuredServices,
  findDomainOperation,
  joinURL,
  missingServiceURLs,
  parseDomainOperationsCatalog,
  parseIdempotencyRecord,
  parsePaginationResponse,
  parseSDKError,
  parseServiceCatalog,
  resolveDomainOperation,
  resolveCatalogOperation,
  serviceURL,
  serviceURLs
} = require("../src");

assert.equal(SDK_NAME, "2finance-sdk-client");
assert.equal(SDK_VERSION, "0.1.0");
assert.equal(SERVICE_CATALOG.services.length, 12);
assert.equal(SERVICE_CATALOG.services[0].env, "TWO_FINANCE_AUTH_URL");
assert.equal(
  serviceURL({ analyticsURL: "https://analytics.example", matchEngineWSURL: "wss://matchengine.example/ws" }, "match_engine"),
  "wss://matchengine.example/ws"
);
assert.equal(
  serviceURLs({ analyticsURL: "https://analytics.example", matchEngineWSURL: "wss://matchengine.example/ws" }).matchengine,
  "wss://matchengine.example/ws"
);
assert.equal(
  configuredServices({ authURL: "https://auth.example", analyticsURL: "https://analytics.example" })[1].name,
  "analytics"
);
assert.equal(
  missingServiceURLs({ authURL: "https://auth.example", analyticsURL: "https://analytics.example" })[0].env,
  "TWO_FINANCE_NETWORK_URL"
);

const requestOptionsFixture = JSON.parse(
  fs.readFileSync(path.join(__dirname, "../../contracts/examples/request-options.json"), "utf8")
);
const domainOperationsFixture = JSON.parse(
  fs.readFileSync(path.join(__dirname, "../../contracts/examples/domain-operations.json"), "utf8")
);
const errorFixture = JSON.parse(fs.readFileSync(path.join(__dirname, "../../contracts/examples/error.json"), "utf8"));
const paginationFixture = JSON.parse(
  fs.readFileSync(path.join(__dirname, "../../contracts/examples/pagination.json"), "utf8")
);
const idempotencyFixture = JSON.parse(
  fs.readFileSync(path.join(__dirname, "../../contracts/examples/idempotency.json"), "utf8")
);
const serviceCatalogFixture = JSON.parse(
  fs.readFileSync(path.join(__dirname, "../../contracts/examples/service-catalog.json"), "utf8")
);

function operation(domainName, operationName) {
  const domain = domainOperationsFixture.domains.find((item) => item.name === domainName);
  assert.ok(domain, `domain ${domainName} should exist`);
  const found = domain.operations.find((item) => item.name === operationName);
  assert.ok(found, `operation ${domainName}.${operationName} should exist`);
  return found;
}

test("configFromEnv loads standard 2Finance URLs", () => {
  const config = configFromEnv({
    TWO_FINANCE_AUTH_URL: "https://auth.example",
    TWO_FINANCE_ANALYTICS_URL: "https://analytics.example",
    TWO_FINANCE_MATCHENGINE_WS_URL: "wss://matchengine.example/ws"
  });
  assert.equal(config.authURL, "https://auth.example");
  assert.equal(config.analyticsURL, "https://analytics.example");
  assert.equal(config.matchEngineWSURL, "wss://matchengine.example/ws");
  assert.equal(config.authRealm, "2finance");
});

test("shared contract fixtures describe public SDK operations", () => {
  const parsedDomains = parseDomainOperationsCatalog(domainOperationsFixture);
  assert.equal(domainOperationsFixture.schema, "sdk.domain_operations.v2");
  assert.equal(parsedDomains.domains[0].operations[0].request_schema, "auth.login.request.v1");
  assert.equal(findDomainOperation(parsedDomains, "analytics", "balances").path, "/portfolio-manager/balances/{account_id}");
  assert.deepEqual(resolveDomainOperation(findDomainOperation(parsedDomains, "analytics", "balances"), { account_id: "acct/1 ok" }), {
    method: "GET",
    path: "/portfolio-manager/balances/acct%2F1%20ok"
  });
  assert.equal(
    resolveDomainOperation(findDomainOperation(parsedDomains, "analytics", "black_scholes"), {}, {
      symbol: "BTC/USD",
      strike: 100000,
      ignored: "drop-me",
      volatility: 0.5
    }).path,
    "/risk-manager/blackscholes?symbol=BTC%2FUSD&strike=100000&volatility=0.5"
  );
  assert.deepEqual(resolveCatalogOperation(parsedDomains, "analytics", "balances", { account_id: "acct/1 ok" }), {
    method: "GET",
    path: "/portfolio-manager/balances/acct%2F1%20ok"
  });
  assert.equal(operation("analytics", "balances").path, "/portfolio-manager/balances/{account_id}");
  assert.deepEqual(operation("analytics", "balances").path_params, ["account_id"]);
  assert.equal(operation("planner", "trading_plan").request_schema, "planner.trading_plan.request.v1");
  assert.equal(operation("matchengine", "order_command").transport, undefined);
  assert.equal(domainOperationsFixture.domains.find((item) => item.name === "matchengine").transport, "websocket");
  assert.equal(errorFixture.error, "rate_limited");
  assert.equal(errorFixture.code, "HTTP_429");
  assert.equal(paginationFixture.next_cursor, "cursor-next");
  assert.equal(idempotencyFixture.idempotency_key, "idem-001");
});

test("shared SDK models parse contract fixtures", () => {
  const error = parseSDKError(errorFixture);
  const pagination = parsePaginationResponse(paginationFixture);
  const idempotency = parseIdempotencyRecord(idempotencyFixture);
  const catalog = parseServiceCatalog(serviceCatalogFixture);

  assert.equal(error.code, "HTTP_429");
  assert.equal(error.details.request_id, "req_2finance_001");
  assert.equal(pagination.limit, 25);
  assert.equal(pagination.next_cursor, "cursor-next");
  assert.equal(idempotency.idempotency_key, "idem-001");
  assert.equal(catalog.services[0].name, "auth");
});

test("MatchEngine client prepares order commands", () => {
  const client = new TwoFinanceClient({
    matchEngineWSURL: "wss://matchengine.example/ws"
  });
  const command = client.matchEngine.orderCommand({
    client_order_id: "co-1",
    idempotency_key: "idem-1",
    symbol: "BTC-USDT",
    side: "buy",
    type: "limit",
    quantity: "0.01"
  });

  assert.equal(client.matchEngine.webSocketURL, "wss://matchengine.example/ws");
  assert.equal(command.schema, "matchengine.order_command.v2");
  assert.equal(command.symbol, "BTC-USDT");
  const subscription = client.matchEngine.marketDataSubscribe({
    symbols: ["BTC-USDT"],
    channels: ["book"]
  });
  assert.equal(subscription.schema, "matchengine.market_data_subscribe.v2");
  assert.deepEqual(subscription.symbols, ["BTC-USDT"]);
  const messages = [];
  const transport = {
    send(message) {
      messages.push(message);
      return { ok: true };
    }
  };
  assert.deepEqual(client.matchEngine.sendOrder(transport, command), { ok: true });
  assert.deepEqual(client.matchEngine.subscribeMarketData(transport, subscription), { ok: true });
  assert.equal(messages[0].schema, "matchengine.order_command.v2");
  assert.equal(messages[1].schema, "matchengine.market_data_subscribe.v2");
});

test("bearerAuthorization normalizes tokens", () => {
  assert.equal(bearerAuthorization("abc"), "Bearer abc");
  assert.equal(bearerAuthorization("Bearer abc"), "Bearer abc");
  assert.equal(bearerAuthorization(""), "");
});

test("Service clients build requests with bearer auth", async () => {
  const seen = [];
  const client = new TwoFinanceClient({
    analyticsURL: "https://analytics.example",
    tokenSource: new StaticTokenSource("token-123"),
    fetch: async (url, init) => {
      seen.push({ url, init });
      return new Response(JSON.stringify({ ok: true }), {
        status: 200,
        headers: { "content-type": "application/json" }
      });
    }
  });

  const response = await client.analytics.indicators();
  assert.deepEqual(response, { ok: true });
  assert.equal(seen[0].url, "https://analytics.example/analytics/indicators");
  assert.equal(seen[0].init.headers.Authorization, "Bearer token-123");
});

test("Auth client exposes JWKS and token validation helpers", async () => {
  const seen = [];
  const client = new TwoFinanceClient({
    authURL: "https://auth.example",
    fetch: async (url, init) => {
      seen.push({ url, init });
      return new Response(JSON.stringify({ ok: true }), {
        status: 200,
        headers: { "content-type": "application/json" }
      });
    }
  });

  await client.auth.jwks();
  await client.auth.validateToken("token-1");

  assert.equal(seen[0].url, "https://auth.example/realms/2finance/protocol/openid-connect/certs");
  assert.equal(seen[0].init.method, "GET");
  assert.equal(seen[1].url, "https://auth.example/realms/2finance/protocol/openid-connect/token/introspect");
  assert.equal(seen[1].init.method, "POST");
  assert.equal(JSON.parse(seen[1].init.body).token, "token-1");
});

test("Service clients apply request options and idempotency key", async () => {
  const seen = [];
  const fixture = requestOptionsFixture;
  const service = new ServiceClient(fixture.request.base_url, {
    fetch: async (url, init) => {
      seen.push({ url, init });
      return new Response(JSON.stringify({ ok: true }), {
        status: 200,
        headers: { "content-type": "application/json" }
      });
    }
  });

  await service.post(
    fixture.request.path,
    { symbol: "BTC-USDT" },
    {
      headers: fixture.request.headers,
      idempotencyKey: ` ${fixture.request.idempotency_key} `,
      query: fixture.request.query,
      pagination: fixture.request.pagination,
      timeoutMs: fixture.request.timeout_ms,
      maxRetries: fixture.request.max_retries
    }
  );

  assert.equal(seen[0].url, fixture.expected.url);
  assert.equal(seen[0].init.headers["X-Trace-ID"], fixture.expected.headers["X-Trace-ID"]);
  assert.equal(seen[0].init.headers["Idempotency-Key"], fixture.expected.headers["Idempotency-Key"]);
  assert.ok(seen[0].init.signal instanceof AbortSignal);
});

test("Service clients throw rich errors with status and body", async () => {
  const service = new ServiceClient("https://analytics.example", {
    fetch: async () => new Response("rate limited", { status: 429 })
  });

  await assert.rejects(
    () => service.get("/analytics/indicators"),
    (error) => {
      assert.ok(error instanceof ServiceError);
      assert.equal(error.method, "GET");
      assert.equal(error.url, "https://analytics.example/analytics/indicators");
      assert.equal(error.statusCode, 429);
      assert.equal(error.body, "rate limited");
      return true;
    }
  );
});

test("Service clients retry retryable responses", async () => {
  let attempts = 0;
  const service = new ServiceClient("https://analytics.example", {
    fetch: async () => {
      attempts++;
      if (attempts === 1) {
        return new Response("temporary", { status: 500 });
      }
      return new Response(JSON.stringify({ ok: true }), {
        status: 200,
        headers: { "content-type": "application/json" }
      });
    }
  });

  const response = await service.get("/analytics/indicators", { maxRetries: 1 });

  assert.equal(attempts, 2);
  assert.deepEqual(response, { ok: true });
});

test("Provider clients expose Wise and Airwallex APIs", async () => {
  const seen = [];
  const client = new TwoFinanceClient({
    wiseURL: "https://wise.example",
    airwallexURL: "https://airwallex.example",
    hummingbotURL: "https://hbot.example",
    fetch: async (url, init) => {
      seen.push({ url, init });
      return new Response(JSON.stringify({ ok: true }), {
        status: 200,
        headers: { "content-type": "application/json" }
      });
    }
  });

  await client.providers.wise.profiles();
  await client.providers.wise.profile("profile/1");
  await client.providers.wise.createQuote("profile/1", { source: "USD" });
  await client.providers.wise.createTransfer({ target: "BRL" });
  await client.providers.airwallex.accounts();
  await client.providers.airwallex.payments();
  await client.providers.airwallex.createPayment({ amount: 10 });
  await client.providers.airwallex.beneficiaries();
  await client.providers.airwallex.createBeneficiary({ name: "beneficiary" });
  await client.hummingbot.connectorConfig({ connector: "2finance" });

  assert.equal(seen[0].url, "https://wise.example/v1/profiles");
  assert.equal(seen[0].init.method, "GET");
  assert.equal(seen[1].url, "https://wise.example/v1/profiles/profile%2F1");
  assert.equal(seen[1].init.method, "GET");
  assert.equal(seen[2].url, "https://wise.example/v3/profiles/profile%2F1/quotes");
  assert.equal(seen[2].init.method, "POST");
  assert.equal(seen[3].url, "https://wise.example/v1/transfers");
  assert.equal(seen[3].init.method, "POST");
  assert.equal(seen[4].url, "https://airwallex.example/api/v1/accounts");
  assert.equal(seen[4].init.method, "GET");
  assert.equal(seen[5].url, "https://airwallex.example/api/v1/payments");
  assert.equal(seen[5].init.method, "GET");
  assert.equal(seen[6].url, "https://airwallex.example/api/v1/payments");
  assert.equal(seen[6].init.method, "POST");
  assert.equal(seen[7].url, "https://airwallex.example/api/v1/beneficiaries");
  assert.equal(seen[7].init.method, "GET");
  assert.equal(seen[8].url, "https://airwallex.example/api/v1/beneficiaries");
  assert.equal(seen[8].init.method, "POST");
  assert.equal(seen[9].url, "https://hbot.example/api/v1/connectors/2finance/config");
  assert.equal(seen[9].init.method, "POST");
});

test("Analytics and TradingControl clients expose full core endpoints", async () => {
  const seen = [];
  const client = new TwoFinanceClient({
    analyticsURL: "https://analytics.example",
    networkURL: "https://network.example",
    tradingControlURL: "https://trading.example",
    fetch: async (url, init) => {
      seen.push({ url, init });
      return new Response(JSON.stringify({ ok: true }), {
        status: 200,
        headers: { "content-type": "application/json" }
      });
    }
  });

  await client.analytics.balances("acct/1");
  await client.analytics.requestOperation({ method: "GET", path: "/portfolio-manager/balances/acct%2Fresolved" });
  await client.analytics.requestCatalogOperation(domainOperationsFixture, "analytics", "balances", { account_id: "acct/1 ok" });
  await client.analytics.blackScholes("symbol=BTC");
  await client.analytics.staking();
  await client.network.marketCandles("BTC/USDT", "limit=10");
  await client.network.bonds();
  await client.network.createBond({ symbol: "BOND1" });
  await client.network.loans();
  await client.network.createLoan({ loan: "ln1" });
  await client.network.swaps();
  await client.network.createSwap({ pair: "BTC-USDT" });
  await client.network.stakingProducts();
  await client.network.createStakingProduct({ asset: "TWO" });
  await client.network.syntheticAssets();
  await client.network.createSyntheticAsset({ asset: "sBTC" });
  await client.network.liquidityPools();
  await client.network.createLiquidityPool({ pool: "BTC-USDT" });
  await client.tradingControl.pauseRobot("robot/1");
  await client.tradingControl.resumeRobot("robot/1");
  await client.tradingControl.stopRobot("robot/1");
  await client.tradingControl.setRiskPolicy("robot/1", { max_drawdown: "0.1" });
  await client.tradingControl.riskView("robot/1");
  await client.tradingControl.strategies();
  await client.tradingControl.createStrategy({ name: "mean-reversion" });
  await client.tradingControl.directives();
  await client.tradingControl.createDirective({ action: "rebalance" });
  await client.tradingControl.audit();
  await client.tradingControl.activity();
  await client.tradingControl.mcpTools();

  assert.deepEqual(
    seen.map((entry) => `${entry.init.method} ${entry.url}`),
    [
      "GET https://analytics.example/portfolio-manager/balances/acct%2F1",
      "GET https://analytics.example/portfolio-manager/balances/acct%2Fresolved",
      "GET https://analytics.example/portfolio-manager/balances/acct%2F1%20ok",
      "GET https://analytics.example/risk-manager/blackscholes?symbol=BTC",
      "GET https://analytics.example/staking",
      "GET https://network.example/v2/2finance-network/markets/BTC%2FUSDT/candles?limit=10",
      "GET https://network.example/v2/2finance-network/products/bonds",
      "POST https://network.example/v2/2finance-network/products/bonds",
      "GET https://network.example/v2/2finance-network/products/loans",
      "POST https://network.example/v2/2finance-network/products/loans",
      "GET https://network.example/v2/2finance-network/products/swaps",
      "POST https://network.example/v2/2finance-network/products/swaps",
      "GET https://network.example/v2/2finance-network/products/staking",
      "POST https://network.example/v2/2finance-network/products/staking",
      "GET https://network.example/v2/2finance-network/products/synthetic-assets",
      "POST https://network.example/v2/2finance-network/products/synthetic-assets",
      "GET https://network.example/v2/2finance-network/products/liquidity-pools",
      "POST https://network.example/v2/2finance-network/products/liquidity-pools",
      "POST https://trading.example/robots/robot%2F1:pause",
      "POST https://trading.example/robots/robot%2F1:resume",
      "POST https://trading.example/robots/robot%2F1:stop",
      "PUT https://trading.example/robots/robot%2F1/risk-policy",
      "GET https://trading.example/risk-view/robot%2F1",
      "GET https://trading.example/strategies",
      "POST https://trading.example/strategies",
      "GET https://trading.example/directives",
      "POST https://trading.example/directives",
      "GET https://trading.example/audit",
      "GET https://trading.example/activity",
      "GET https://trading.example/mcp/tools"
    ]
  );
});

test("KeyStore client exposes health readiness signatures and metrics", async () => {
  const seen = [];
  const client = new TwoFinanceClient({
    keyStoreURL: "https://keys.example",
    fetch: async (url, init) => {
      seen.push({ url, init });
      return new Response(JSON.stringify({ ok: true }), {
        status: 200,
        headers: { "content-type": "application/json" }
      });
    }
  });

  await client.keystore.health();
  await client.keystore.readiness();
  await client.keystore.keygenSignature({ key: "k1" });
  await client.keystore.signingSignature({ signature: "s1" });
  await client.keystore.startResharing({ key: "k1" });
  await client.keystore.signatures("pub-1");
  await client.keystore.metrics();

  assert.deepEqual(
    seen.map((entry) => `${entry.init.method} ${entry.url}`),
    [
      "GET https://keys.example/healthz",
      "GET https://keys.example/readyz",
      "POST https://keys.example/keystore/keygen/signature",
      "POST https://keys.example/keystore/signing/signature",
      "POST https://keys.example/keystore/resharing/start",
      "GET https://keys.example/keystore/signatures/pub-1",
      "GET https://keys.example/keystore/tss/metrics"
    ]
  );
});

test("MCP and Orchestrator clients expose prompts resources providers approvals and sessions", async () => {
  const seen = [];
  const client = new TwoFinanceClient({
    mcpURL: "https://mcp.example",
    orchestratorURL: "https://orchestrator.example",
    fetch: async (url, init) => {
      seen.push({ url, init, body: init.body ? String(init.body) : "" });
      return new Response(JSON.stringify({ ok: true }), {
        status: 200,
        headers: { "content-type": "application/json" }
      });
    }
  });

  await client.mcp.listTools();
  await client.mcp.listPrompts();
  await client.mcp.listResources();
  await client.mcp.readResource("resource://portfolio");
  await client.mcp.getPrompt("planner.prompt", { symbol: "BTC-USDT" });
  await client.orchestrator.prompts();
  await client.orchestrator.resources();
  await client.orchestrator.providers();
  await client.orchestrator.approvals();
  await client.orchestrator.observability();
  await client.orchestrator.deleteSession("session-1");
  await client.planner.operationalPlan({ session_id: "session-1", message: "operate" });

  assert.ok(seen.some((entry) => entry.url === "https://mcp.example/mcp" && entry.body.includes('"method":"tools/list"')));
  assert.ok(seen.some((entry) => entry.url === "https://mcp.example/mcp" && entry.body.includes('"method":"prompts/list"')));
  assert.ok(seen.some((entry) => entry.url === "https://mcp.example/mcp" && entry.body.includes('"method":"resources/list"')));
  assert.ok(seen.some((entry) => entry.body.includes("resource://portfolio")));
  assert.ok(seen.some((entry) => entry.body.includes("planner.prompt")));
  assert.deepEqual(
    seen.filter((entry) => entry.url.startsWith("https://orchestrator.example")).map((entry) => `${entry.init.method} ${entry.url}`),
    [
      "GET https://orchestrator.example/v1/mcphost/prompts",
      "GET https://orchestrator.example/v1/mcphost/resources",
      "GET https://orchestrator.example/v1/mcphost/providers",
      "GET https://orchestrator.example/v1/mcphost/approvals",
      "GET https://orchestrator.example/v1/mcphost/observability",
      "DELETE https://orchestrator.example/v1/mcphost/sessions/session-1",
      "POST https://orchestrator.example/v1/mcphost/messages"
    ]
  );
});

test("ClientCredentialsTokenSource fetches and caches tokens", async () => {
  const seen = [];
  const source = new ClientCredentialsTokenSource({
    tokenURL: "https://auth.example/token",
    clientID: "client-id",
    clientSecret: "client-secret",
    scopes: ["analytics:read", "mcp:invoke"],
    fetch: async (url, init) => {
      seen.push({ url, init });
      return new Response(JSON.stringify({ access_token: "cc-token", expires_in: 3600 }), {
        status: 200,
        headers: { "content-type": "application/json" }
      });
    }
  });

  assert.equal(await source.getToken(), "cc-token");
  assert.equal(await source.getToken(), "cc-token");
  assert.equal(seen.length, 1);
  assert.equal(seen[0].url, "https://auth.example/token");
  assert.equal(seen[0].init.method, "POST");
  assert.equal(seen[0].init.body.get("grant_type"), "client_credentials");
  assert.equal(seen[0].init.body.get("scope"), "analytics:read mcp:invoke");
});

test("joinURL resolves service paths", () => {
  assert.equal(joinURL("https://api.example/", "/healthz"), "https://api.example/healthz");
});
