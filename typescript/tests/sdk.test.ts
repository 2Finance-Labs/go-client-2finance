import { readFileSync } from "node:fs";
import { dirname, join } from "node:path";
import { fileURLToPath } from "node:url";

import {
  ClientCredentialsTokenSource,
  RequestOptions,
  SERVICE_CATALOG,
  SDK_NAME,
  SDK_VERSION,
  ServiceClient,
  ServiceError,
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
} from "../src/index.js";

type DomainOperation = {
  name: string;
  method: string;
  path: string;
  path_params?: string[];
  request_schema?: string;
  response_schema?: string;
};

type DomainFixture = {
  schema: string;
  domains: {
    name: string;
    env: string;
    transport?: string;
    operations: DomainOperation[];
  }[];
};

type SeenRequest = {
  url: string;
  method: string;
  headers: Record<string, string>;
  body?: BodyInit | null;
};

function assertEqual<T>(actual: T, expected: T, message: string): void {
  if (actual !== expected) {
    throw new Error(`${message}: expected ${String(expected)}, got ${String(actual)}`);
  }
}

function assertDeepEqual(actual: unknown, expected: unknown, message: string): void {
  const actualJSON = JSON.stringify(actual);
  const expectedJSON = JSON.stringify(expected);
  if (actualJSON !== expectedJSON) {
    throw new Error(`${message}: expected ${expectedJSON}, got ${actualJSON}`);
  }
}

assertEqual(SDK_NAME, "2finance-sdk-client", "SDK_NAME should be public");
assertEqual(SDK_VERSION, "0.1.0", "SDK_VERSION should be public");
assertEqual(SERVICE_CATALOG.services.length, 12, "SERVICE_CATALOG should expose canonical domains");
assertEqual(SERVICE_CATALOG.services[0].env, "TWO_FINANCE_AUTH_URL", "SERVICE_CATALOG should expose env vars");
assertEqual(
  serviceURL({ analyticsURL: "https://analytics.example", matchEngineWSURL: "wss://matchengine.example/ws" }, "match_engine"),
  "wss://matchengine.example/ws",
  "serviceURL should resolve service aliases"
);
assertEqual(
  serviceURLs({ analyticsURL: "https://analytics.example", matchEngineWSURL: "wss://matchengine.example/ws" }).matchengine,
  "wss://matchengine.example/ws",
  "serviceURLs should expose configured service URLs"
);
assertEqual(
  configuredServices({ authURL: "https://auth.example", analyticsURL: "https://analytics.example" })[1].name,
  "analytics",
  "configuredServices should preserve catalog order"
);
assertEqual(
  missingServiceURLs({ authURL: "https://auth.example", analyticsURL: "https://analytics.example" })[0].env,
  "TWO_FINANCE_NETWORK_URL",
  "missingServiceURLs should expose missing env vars"
);

function headersRecord(headers: HeadersInit | undefined): Record<string, string> {
  if (!headers) {
    return {};
  }
  if (headers instanceof Headers) {
    const result: Record<string, string> = {};
    headers.forEach((value, key) => {
      result[key] = value;
    });
    return result;
  }
  if (Array.isArray(headers)) {
    return Object.fromEntries(headers);
  }
  return headers;
}

const testDir = dirname(fileURLToPath(import.meta.url));
const contractsDir = join(testDir, "../../../contracts/examples");

function fixture<T>(name: string): T {
  return JSON.parse(readFileSync(join(contractsDir, name), "utf8")) as T;
}

function operation(fixtureValue: DomainFixture, domainName: string, operationName: string): DomainOperation {
  const domain = fixtureValue.domains.find((item) => item.name === domainName);
  if (!domain) {
    throw new Error(`domain ${domainName} should exist`);
  }
  const found = domain.operations.find((item) => item.name === operationName);
  if (!found) {
    throw new Error(`operation ${domainName}.${operationName} should exist`);
  }
  return found;
}

function contractFixtureSmoke(): void {
  const domains = parseDomainOperationsCatalog(fixture<unknown>("domain-operations.json"));
  const error = fixture<{ error: string; code: string }>("error.json");
  const pagination = fixture<{ limit: number; next_cursor: string }>("pagination.json");
  const idempotency = fixture<{ idempotency_key: string }>("idempotency.json");

  assertEqual(domains.schema, "sdk.domain_operations.v2", "domain operations schema should match");
  assertEqual(
    operation(domains, "analytics", "balances").path,
    "/portfolio-manager/balances/{account_id}",
    "analytics balances contract should define path"
  );
  assertDeepEqual(
    operation(domains, "analytics", "balances").path_params,
    ["account_id"],
    "analytics balances contract should define path params"
  );
  assertEqual(
    operation(domains, "planner", "trading_plan").request_schema,
    "planner.trading_plan.request.v1",
    "planner trading contract should define request schema"
  );
  assertEqual(
    operation(domains, "auth", "login").response_schema,
    "auth.token.response.v1",
    "domain operation model should parse response schema"
  );
  assertEqual(
    findDomainOperation(domains, "analytics", "balances")?.path,
    "/portfolio-manager/balances/{account_id}",
    "findDomainOperation should locate operations by domain"
  );
  const balances = findDomainOperation(domains, "analytics", "balances");
  if (!balances) {
    throw new Error("analytics balances operation should exist");
  }
  assertDeepEqual(
    resolveDomainOperation(balances, { account_id: "acct/1 ok" }),
    { method: "GET", path: "/portfolio-manager/balances/acct%2F1%20ok" },
    "resolveDomainOperation should expand and escape path params"
  );
  const risk = findDomainOperation(domains, "analytics", "black_scholes");
  if (!risk) {
    throw new Error("analytics black_scholes operation should exist");
  }
  assertEqual(
    resolveDomainOperation(risk, {}, { symbol: "BTC/USD", strike: 100000, ignored: "drop-me", volatility: 0.5 }).path,
    "/risk-manager/blackscholes?symbol=BTC%2FUSD&strike=100000&volatility=0.5",
    "resolveDomainOperation should keep only contract query params"
  );
  assertDeepEqual(
    resolveCatalogOperation(domains, "analytics", "balances", { account_id: "acct/1 ok" }),
    { method: "GET", path: "/portfolio-manager/balances/acct%2F1%20ok" },
    "resolveCatalogOperation should locate and resolve operations"
  );
  assertEqual(error.code, "HTTP_429", "error fixture should define rate-limit code");
  assertEqual(error.error, "rate_limited", "error fixture should define rate-limit name");
  assertEqual(pagination.limit, 25, "pagination fixture should define limit");
  assertEqual(pagination.next_cursor, "cursor-next", "pagination fixture should define next cursor");
  assertEqual(idempotency.idempotency_key, "idem-001", "idempotency fixture should define canonical key");
}

function sharedModelsSmoke(): void {
  const error = parseSDKError(fixture("error.json"));
  const pagination = parsePaginationResponse(fixture("pagination.json"));
  const idempotency = parseIdempotencyRecord(fixture("idempotency.json"));
  const catalog = parseServiceCatalog(fixture("service-catalog.json"));

  assertEqual(error.code, "HTTP_429", "SDK error model should parse code");
  assertEqual(error.details?.request_id, "req_2finance_001", "SDK error model should parse details");
  assertEqual(pagination.limit, 25, "pagination model should parse limit");
  assertEqual(pagination.next_cursor, "cursor-next", "pagination model should parse next cursor");
  assertEqual(idempotency.idempotency_key, "idem-001", "idempotency model should parse key");
  assertEqual(catalog.services[0].name, "auth", "service catalog model should parse services");
}

const seen: SeenRequest[] = [];
const fetchImpl: typeof fetch = async (input, init = {}) => {
  seen.push({
    url: String(input),
    method: init.method || "GET",
    headers: headersRecord(init.headers),
    body: init.body
  });
  if (String(input) === "https://auth.example/token") {
    return new Response(JSON.stringify({ access_token: "cc-token", expires_in: 3600 }), {
      status: 200,
      headers: { "content-type": "application/json" }
    });
  }
  return new Response(JSON.stringify({ ok: true }), {
    status: 200,
    headers: { "content-type": "application/json" }
  });
};

const requestOptions: RequestOptions = {
  headers: { "X-Trace-ID": "trace-1" },
  idempotencyKey: " idem-1 ",
  query: { symbol: "BTC-USDT" },
  pagination: { page: 2, limit: 25 },
  timeoutMs: 500,
  maxRetries: 1
};

const config = configFromEnv({
  TWO_FINANCE_AUTH_URL: "https://auth.example",
  TWO_FINANCE_NETWORK_URL: "https://network.example",
  TWO_FINANCE_ANALYTICS_URL: "https://analytics.example",
  TWO_FINANCE_ORCHESTRATOR_URL: "https://orchestrator.example",
  TWO_FINANCE_MCP_URL: "https://mcp.example",
  TWO_FINANCE_TRADING_CONTROL_URL: "https://trading.example",
  TWO_FINANCE_MATCHENGINE_WS_URL: "wss://matchengine.example/ws",
  TWO_FINANCE_KEYSTORE_URL: "https://keys.example",
  TWO_FINANCE_HUMMINGBOT_URL: "https://hbot.example",
  TWO_FINANCE_WISE_URL: "https://wise.example",
  TWO_FINANCE_AIRWALLEX_URL: "https://airwallex.example"
});

const client = new TwoFinanceClient({
  ...config,
  fetch: fetchImpl,
  tokenSource: new StaticTokenSource(" token-123 ")
});

async function sdkSmoke(): Promise<void> {
  await client.auth.login({ username: "user", password: "pass" });
  await client.auth.refreshToken("refresh-token");
  await client.auth.phoneLogin("+15555550100", "123456");
  await client.auth.jwks();
  await client.auth.validateToken("token-1");

  await client.analytics.indicators();
  await client.analytics.calculateTechnicalAnalysis({ symbol: "BTC-USDT" });
  await client.analytics.upsertCandles({ symbol: "BTC-USDT" });
  await client.analytics.post("/analytics/candles:upsert", { symbol: "BTC-USDT" }, requestOptions);
  await client.analytics.optimizePortfolio({ account_id: "acct-1" });
  await client.analytics.rankings();
  await client.analytics.balances("acct/1");
  await client.analytics.requestOperation({ method: "GET", path: "/portfolio-manager/balances/acct%2Fresolved" });
  await client.analytics.requestCatalogOperation(
    parseDomainOperationsCatalog(fixture("domain-operations.json")),
    "analytics",
    "balances",
    { account_id: "acct/1 ok" }
  );
  await client.analytics.blackScholes("symbol=BTC");
  await client.analytics.staking();

  await client.network.virtualMachine();
  await client.network.marketCandles("BTC-USDT", "limit=10");
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

  await client.mcp.listTools();
  await client.mcp.listPrompts();
  await client.mcp.listResources();
  await client.mcp.readResource("resource://portfolio");
  await client.mcp.getPrompt("planner.prompt", { symbol: "BTC-USDT" });
  await client.mcp.conversationPlan({ goal: "trade plan" });

  await client.orchestrator.catalog();
  await client.orchestrator.createSession({ user_id: "user-1" });
  await client.orchestrator.sendMessage({ session_id: "session-1", message: "hello" });
  await client.orchestrator.tools();
  await client.orchestrator.prompts();
  await client.orchestrator.resources();
  await client.orchestrator.providers();
  await client.orchestrator.approvals();
  await client.orchestrator.observability();
  await client.orchestrator.deleteSession("session/1");

  await client.tradingControl.robots();
  await client.tradingControl.createRobot({ id: "robot-1" });
  await client.tradingControl.robot("robot/1");
  await client.tradingControl.startRobot("robot/1");
  await client.tradingControl.pauseRobot("robot/1");
  await client.tradingControl.resumeRobot("robot/1");
  await client.tradingControl.stopRobot("robot/1");
  await client.tradingControl.riskPolicy("robot/1");
  await client.tradingControl.setRiskPolicy("robot/1", { max_drawdown: "0.1" });
  await client.tradingControl.riskView("robot/1");
  await client.tradingControl.strategies();
  await client.tradingControl.createStrategy({ name: "mean-reversion" });
  await client.tradingControl.directives();
  await client.tradingControl.createDirective({ action: "rebalance" });
  await client.tradingControl.audit();
  await client.tradingControl.activity();
  await client.tradingControl.mcpTools();

  await client.keystore.health();
  await client.keystore.readiness();
  await client.keystore.startKeygen({ key_id: "key-1" });
  await client.keystore.keygenSignature({ key_id: "key-1" });
  await client.keystore.startSigning({ key_id: "key-1" });
  await client.keystore.signingSignature({ key_id: "key-1" });
  await client.keystore.startResharing({ key_id: "key-1" });
  await client.keystore.keys("pub/1");
  await client.keystore.signatures("pub/1");
  await client.keystore.metrics();

  await client.hummingbot.assets();
  await client.hummingbot.symbols();
  await client.hummingbot.balances();
  await client.hummingbot.connectorConfig({ connector: "2finance" });

  await client.providers.wise.profiles();
  await client.providers.wise.profile("profile/1");
  await client.providers.wise.createQuote("profile/1", { source: "USD" });
  await client.providers.wise.createTransfer({ target: "BRL" });
  await client.providers.airwallex.accounts();
  await client.providers.airwallex.payments();
  await client.providers.airwallex.createPayment({ amount: 10 });
  await client.providers.airwallex.beneficiaries();
  await client.providers.airwallex.createBeneficiary({ name: "beneficiary" });

  await client.planner.conversationPlan({ goal: "trade plan" });
  await client.planner.orchestratedPlan({ session_id: "session-1", message: "plan" });
  await client.planner.operationalPlan({ session_id: "session-1", message: "operate" });
  await client.planner.tradingPlan({ goal: "rebalance", useAnalytics: true, useTrading: true });

  const command = client.matchEngine.orderCommand({
    client_order_id: "co-1",
    idempotency_key: "idem-1",
    order_type: "LIMIT",
    wallet_id: 1,
    symbol_id: 1,
    side: "BUY",
    quantity: "0.01"
  });
  command.schema satisfies string;
  assertEqual(command.schema, "matchengine.order_command.v2", "matchengine schema should default");
  const subscription = client.matchEngine.marketDataSubscribe({
    symbols: ["BTC-USDT"],
    channels: ["book"]
  });
  subscription.schema satisfies string;
  assertEqual(subscription.schema, "matchengine.market_data_subscribe.v2", "matchengine market data schema should default");
  const matchMessages: unknown[] = [];
  const matchTransport = {
    send(message: unknown): unknown {
      matchMessages.push(message);
      return { ok: true };
    }
  };
  assertDeepEqual(
    client.matchEngine.sendOrder(matchTransport, command),
    { ok: true },
    "matchengine sendOrder should return transport result"
  );
  client.matchEngine.subscribeMarketData(matchTransport, subscription);
  assertEqual(
    (matchMessages[0] as { schema: string }).schema,
    "matchengine.order_command.v2",
    "matchengine sendOrder should send order schema"
  );
  assertEqual(
    (matchMessages[1] as { schema: string }).schema,
    "matchengine.market_data_subscribe.v2",
    "matchengine subscribeMarketData should send subscription schema"
  );
}

async function behaviorSmoke(): Promise<void> {
  assertEqual(config.authURL, "https://auth.example", "auth URL should load from env");
  assertEqual(client.matchEngine.webSocketURL, "wss://matchengine.example/ws", "matchengine URL should load");
  assertEqual(bearerAuthorization(" Bearer abc "), "Bearer abc", "bearer token should be trimmed");
  assertEqual(joinURL("https://api.example/", "/healthz"), "https://api.example/healthz", "URL should join");

  const candles = seen.find((entry) => entry.url.startsWith("https://analytics.example/analytics/candles:upsert?"));
  if (!candles) {
    throw new Error("analytics candles request should be sent");
  }
  assertEqual(
    candles.url,
    "https://analytics.example/analytics/candles:upsert?symbol=BTC-USDT&page=2&limit=25",
    "request options should add query params and pagination"
  );
  assertEqual(candles.headers.Authorization, "Bearer token-123", "bearer auth should be trimmed and injected");
  assertEqual(candles.headers["Idempotency-Key"], "idem-1", "idempotency key should be trimmed and injected");
  assertEqual(candles.headers["X-Trace-ID"], "trace-1", "custom header should be injected");

  const encodedURLs = seen.map((entry) => `${entry.method} ${entry.url}`);
  for (const expected of [
    "GET https://analytics.example/portfolio-manager/balances/acct%2F1",
    "GET https://analytics.example/portfolio-manager/balances/acct%2Fresolved",
    "GET https://analytics.example/portfolio-manager/balances/acct%2F1%20ok",
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
    "GET https://auth.example/realms/2finance/protocol/openid-connect/certs",
    "POST https://auth.example/realms/2finance/protocol/openid-connect/token/introspect",
    "DELETE https://orchestrator.example/v1/mcphost/sessions/session%2F1",
    "POST https://trading.example/robots/robot%2F1:pause",
    "GET https://trading.example/strategies",
    "POST https://trading.example/strategies",
    "GET https://trading.example/directives",
    "POST https://trading.example/directives",
    "GET https://trading.example/audit",
    "GET https://trading.example/activity",
    "GET https://trading.example/mcp/tools",
    "GET https://keys.example/keystore/keys/pub%2F1",
    "POST https://hbot.example/api/v1/connectors/2finance/config",
    "GET https://wise.example/v1/profiles",
    "GET https://wise.example/v1/profiles/profile%2F1",
    "POST https://wise.example/v3/profiles/profile%2F1/quotes",
    "POST https://wise.example/v1/transfers",
    "GET https://airwallex.example/api/v1/accounts",
    "GET https://airwallex.example/api/v1/payments",
    "POST https://airwallex.example/api/v1/payments",
    "GET https://airwallex.example/api/v1/beneficiaries",
    "POST https://airwallex.example/api/v1/beneficiaries"
  ]) {
    if (!encodedURLs.includes(expected)) {
      throw new Error(`missing expected TypeScript SDK request: ${expected}`);
    }
  }

  const plannerCall = seen
    .filter((entry) => entry.url === "https://mcp.example/mcp")
    .map((entry) => JSON.parse(String(entry.body)))
    .find((body) => body.params?.arguments?.context?.trading_robots && body.params?.arguments?.context?.analytics_indicators);
  if (!plannerCall) {
    throw new Error("planner tradingPlan should enrich context and call MCP");
  }
  assertEqual(
    plannerCall.params.name,
    "finance_assistant.conversation.plan",
    "planner should call conversation plan tool"
  );

  let retryAttempts = 0;
  const retryService = new ServiceClient("https://analytics.example", {
    fetch: async () => {
      retryAttempts++;
      if (retryAttempts === 1) {
        return new Response("temporary", { status: 500 });
      }
      return new Response(JSON.stringify({ ok: true }), { status: 200 });
    }
  });
  const retryResponse = await retryService.get("/analytics/indicators", { maxRetries: 1 });
  assertDeepEqual(retryResponse, { ok: true }, "retryable response should eventually decode");
  assertEqual(retryAttempts, 2, "retryable response should retry once");

  const failingService = new ServiceClient("https://analytics.example", {
    fetch: async () => new Response("rate limited", { status: 429 })
  });
  try {
    await failingService.get("/analytics/indicators");
    throw new Error("ServiceError should be thrown for non-2xx responses");
  } catch (error) {
    if (!(error instanceof ServiceError)) {
      throw error;
    }
    assertEqual(error.method, "GET", "service error should keep method");
    assertEqual(error.statusCode, 429, "service error should keep status code");
    assertEqual(error.body, "rate limited", "service error should keep body");
  }

  const credentials = new ClientCredentialsTokenSource({
    tokenURL: "https://auth.example/token",
    clientID: "client-id",
    clientSecret: "client-secret",
    scopes: ["analytics:read"],
    fetch: fetchImpl
  });
  assertEqual(await credentials.getToken(), "cc-token", "client credentials token should parse");
  assertEqual(await credentials.getToken(), "cc-token", "client credentials token should cache");
  const tokenRequests = seen.filter((entry) => entry.url === "https://auth.example/token");
  assertEqual(tokenRequests.length, 1, "client credentials token should fetch once while cached");
}

await sdkSmoke();
await behaviorSmoke();
contractFixtureSmoke();
sharedModelsSmoke();
