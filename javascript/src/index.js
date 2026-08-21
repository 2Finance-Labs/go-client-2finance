"use strict";

const SDK_NAME = "2finance-sdk-client";
const SDK_VERSION = "0.1.0";

const DEFAULT_AUTH_REALM = "2finance";
const DEFAULT_AUTH_CLIENT_ID = "2finance-network";
const DEFAULT_AUTH_PHONE_CLIENT_ID = "2finance-network-phone";
const SERVICE_CATALOG = Object.freeze({
  services: Object.freeze([
    Object.freeze({ name: "auth", env: "TWO_FINANCE_AUTH_URL" }),
    Object.freeze({ name: "network", env: "TWO_FINANCE_NETWORK_URL" }),
    Object.freeze({ name: "analytics", env: "TWO_FINANCE_ANALYTICS_URL" }),
    Object.freeze({ name: "orchestrator", env: "TWO_FINANCE_ORCHESTRATOR_URL" }),
    Object.freeze({ name: "mcp", env: "TWO_FINANCE_MCP_URL" }),
    Object.freeze({ name: "planner", env: "TWO_FINANCE_MCP_URL" }),
    Object.freeze({ name: "tradingcontrol", env: "TWO_FINANCE_TRADING_CONTROL_URL" }),
    Object.freeze({ name: "matchengine", env: "TWO_FINANCE_MATCHENGINE_WS_URL" }),
    Object.freeze({ name: "keystore", env: "TWO_FINANCE_KEYSTORE_URL" }),
    Object.freeze({ name: "hummingbot", env: "TWO_FINANCE_HUMMINGBOT_URL" }),
    Object.freeze({ name: "wise", env: "TWO_FINANCE_WISE_URL" }),
    Object.freeze({ name: "airwallex", env: "TWO_FINANCE_AIRWALLEX_URL" })
  ])
});

function trimTrailingSlash(value) {
  return String(value || "").trim().replace(/\/+$/, "");
}

function joinURL(baseURL, path) {
  const base = trimTrailingSlash(baseURL);
  if (!base) {
    throw new Error("2finance: baseURL is required");
  }
  if (/^https?:\/\//i.test(path)) {
    return path;
  }
  return `${base}/${String(path || "/").replace(/^\/+/, "")}`;
}

function bearerAuthorization(token) {
  const trimmed = String(token || "").trim();
  if (!trimmed) {
    return "";
  }
  if (trimmed.toLowerCase().startsWith("bearer ")) {
    return trimmed;
  }
  return `Bearer ${trimmed}`;
}

class ServiceError extends Error {
  constructor(method, url, statusCode, body) {
    super(`2finance: ${method} ${url} returned ${statusCode}: ${body}`);
    this.name = "ServiceError";
    this.method = method;
    this.url = url;
    this.statusCode = statusCode;
    this.body = body;
  }
}

function parseSDKError(payload) {
  return { ...payload, details: { ...(payload?.details || {}) } };
}

function parsePaginationResponse(payload) {
  return { ...payload, items: [...(payload?.items || [])] };
}

function parseIdempotencyRecord(payload) {
  return { ...payload };
}

function parseServiceCatalog(payload) {
  return { services: [...(payload?.services || [])] };
}

function parseDomainOperationsCatalog(payload) {
  return {
    schema: payload?.schema,
    domains: [...(payload?.domains || [])].map((domain) => ({
      ...domain,
      operations: [...(domain?.operations || [])]
    }))
  };
}

function findDomainOperation(catalog, domainName, operationName) {
  const domain = (catalog?.domains || []).find((item) => serviceKey(item?.name) === serviceKey(domainName));
  return domain?.operations?.find((operation) => operation.name === operationName);
}

function resolveDomainOperation(operation, pathParams = {}, query = {}) {
  let path = operation.path;
  for (const name of operation.path_params || []) {
    if (!(name in pathParams)) {
      throw new Error(`2finance: missing operation path parameter ${name}`);
    }
    path = path.replaceAll(`{${name}}`, encodeURIComponent(String(pathParams[name])));
  }

  const search = new URLSearchParams();
  for (const name of operation.query || []) {
    const value = query?.[name];
    if (value !== undefined && value !== null) {
      search.set(name, String(value));
    }
  }
  const encoded = search.toString();
  if (encoded) {
    path += `${path.includes("?") ? "&" : "?"}${encoded}`;
  }

  return { method: String(operation.method || "").trim().toUpperCase(), path };
}

function resolveCatalogOperation(catalog, domainName, operationName, pathParams = {}, query = {}) {
  const operation = findDomainOperation(catalog, domainName, operationName);
  if (!operation) {
    throw new Error(`2finance: unknown operation ${domainName}.${operationName}`);
  }
  return resolveDomainOperation(operation, pathParams, query);
}

function serviceURL(config, domain) {
  switch (serviceKey(domain)) {
    case "auth":
      return config?.authURL || "";
    case "network":
      return config?.networkURL || "";
    case "analytics":
      return config?.analyticsURL || "";
    case "orchestrator":
      return config?.orchestratorURL || "";
    case "mcp":
    case "planner":
      return config?.mcpURL || "";
    case "tradingcontrol":
      return config?.tradingControlURL || "";
    case "matchengine":
      return config?.matchEngineWSURL || "";
    case "keystore":
      return config?.keyStoreURL || "";
    case "hummingbot":
      return config?.hummingbotURL || "";
    case "wise":
      return config?.wiseURL || "";
    case "airwallex":
      return config?.airwallexURL || "";
    default:
      return "";
  }
}

function serviceURLs(config) {
  const urls = {};
  for (const service of SERVICE_CATALOG.services) {
    const url = serviceURL(config, service.name);
    if (url) {
      urls[service.name] = url;
    }
  }
  return urls;
}

function configuredServices(config) {
  return SERVICE_CATALOG.services
    .map((service) => ({ ...service, url: serviceURL(config, service.name) }))
    .filter((service) => service.url);
}

function missingServiceURLs(config) {
  return SERVICE_CATALOG.services.filter((service) => !serviceURL(config, service.name));
}

function serviceKey(domain) {
  return String(domain || "").trim().toLowerCase().replace(/[-_\s]+/g, "");
}

class StaticTokenSource {
  constructor(token) {
    this.token = token;
  }

  async getToken() {
    return this.token;
  }
}

class ClientCredentialsTokenSource {
  constructor(options) {
    this.options = { ...options };
    this.accessToken = "";
    this.expiresAt = 0;
  }

  async getToken() {
    const now = Date.now();
    const skewMs = this.options.expirySkewMs ?? 30000;
    if (this.accessToken && now < this.expiresAt - skewMs) {
      return this.accessToken;
    }
    if (!this.options.tokenURL || !this.options.clientID || !this.options.clientSecret) {
      throw new Error("2finance auth: tokenURL, clientID and clientSecret are required");
    }
    const body = new URLSearchParams();
    body.set("grant_type", "client_credentials");
    body.set("client_id", this.options.clientID);
    body.set("client_secret", this.options.clientSecret);
    if (this.options.scopes?.length) {
      body.set("scope", this.options.scopes.join(" "));
    }
    const fetchImpl = this.options.fetch || globalThis.fetch;
    if (!fetchImpl) {
      throw new Error("2finance auth: fetch is required");
    }
    const response = await fetchImpl(this.options.tokenURL, {
      method: "POST",
      headers: {
        Accept: "application/json",
        "Content-Type": "application/x-www-form-urlencoded"
      },
      body
    });
    const payload = await response.json();
    if (!response.ok) {
      throw new Error(`2finance auth: token endpoint returned ${response.status}`);
    }
    if (!payload.access_token) {
      throw new Error("2finance auth: token response missing access_token");
    }
    this.accessToken = payload.access_token;
    this.expiresAt = now + Number(payload.expires_in || 300) * 1000;
    return this.accessToken;
  }
}

class ServiceClient {
  constructor(baseURL, options = {}) {
    this.baseURL = trimTrailingSlash(baseURL);
    this.fetch = options.fetch || globalThis.fetch;
    this.tokenSource = options.tokenSource;
  }

  url(path) {
    return joinURL(this.baseURL, path);
  }

  async request(method, path, body, requestOptions = {}) {
    if (!this.fetch) {
      throw new Error("2finance: fetch is required");
    }
    const headers = { Accept: "application/json" };
    let requestBody;
    if (body !== undefined && body !== null) {
      headers["Content-Type"] = "application/json";
      requestBody = typeof body === "string" ? body : JSON.stringify(body);
    }
    if (this.tokenSource) {
      const token = await this.tokenSource.getToken();
      const authorization = bearerAuthorization(token);
      if (authorization) {
        headers.Authorization = authorization;
      }
    }
    Object.assign(headers, requestOptions.headers || {});
    const idempotencyKey = String(requestOptions.idempotencyKey || "").trim();
    if (idempotencyKey) {
      headers["Idempotency-Key"] = idempotencyKey;
    }
    const requestURL = this.urlWithQuery(path, requestOptions.query, requestOptions.pagination);
    const abortController = requestOptions.timeoutMs ? new AbortController() : undefined;
    const timeout = abortController ? setTimeout(() => abortController.abort(), requestOptions.timeoutMs) : undefined;
    const attempts = 1 + Math.max(0, requestOptions.maxRetries || 0);
    try {
      let lastText = "";
      let lastStatus = 0;
      for (let attempt = 0; attempt < attempts; attempt++) {
        const response = await this.fetch(requestURL, {
          method,
          headers,
          body: requestBody,
          signal: abortController?.signal
        });
        const text = await response.text();
        if (response.ok) {
          const payload = text ? JSON.parse(text) : null;
          return payload;
        }
        lastText = text;
        lastStatus = response.status;
        if (attempt + 1 >= attempts || !isRetryableStatus(response.status)) {
          break;
        }
      }
      throw new ServiceError(method, requestURL, lastStatus, lastText);
    } finally {
      if (timeout) {
        clearTimeout(timeout);
      }
    }
  }

  requestOperation(operation, body, requestOptions = {}) {
    return this.request(operation.method, operation.path, body, requestOptions);
  }

  requestCatalogOperation(catalog, domainName, operationName, pathParams = {}, query = {}, body, requestOptions = {}) {
    const operation = resolveCatalogOperation(catalog, domainName, operationName, pathParams, query);
    return this.requestOperation(operation, body, requestOptions);
  }

  urlWithQuery(path, query = {}, pagination = {}) {
    const requestURL = new URL(this.url(path));
    for (const [key, value] of Object.entries(query || {})) {
      if (value !== undefined && value !== null) {
        requestURL.searchParams.set(key, String(value));
      }
    }
    if (pagination?.page !== undefined) {
      requestURL.searchParams.set("page", String(pagination.page));
    }
    if (pagination?.limit !== undefined) {
      requestURL.searchParams.set("limit", String(pagination.limit));
    }
    return requestURL.toString();
  }

  get(path, options) {
    return this.request("GET", path, undefined, options);
  }

  post(path, body, options) {
    return this.request("POST", path, body, options);
  }

  put(path, body, options) {
    return this.request("PUT", path, body, options);
  }

  delete(path, options) {
    return this.request("DELETE", path, undefined, options);
  }
}

function isRetryableStatus(status) {
  return status === 429 || status >= 500;
}

class AuthClient extends ServiceClient {
  constructor(baseURL, options = {}) {
    super(baseURL, options);
    this.realm = options.realm || DEFAULT_AUTH_REALM;
    this.clientID = options.clientID || DEFAULT_AUTH_CLIENT_ID;
    this.phoneClientID = options.phoneClientID || DEFAULT_AUTH_PHONE_CLIENT_ID;
  }

  authPath(clientID, endpoint) {
    return `/v1/2finance-authenticator/${this.realm}/${clientID}/${String(endpoint).replace(/^\/+/, "")}`;
  }

  login(input) {
    return this.post(this.authPath(this.clientID, "/login"), input);
  }

  signUp(input) {
    return this.post(this.authPath(this.clientID, "/signup"), input);
  }

  refreshToken(refreshToken) {
    return this.post(this.authPath(this.clientID, "/refresh"), { refresh_token: refreshToken });
  }

  logout(refreshToken) {
    return this.post(this.authPath(this.clientID, "/logout"), { refresh_token: refreshToken });
  }

  phoneLogin(phoneNumber, code) {
    return this.post(this.authPath(this.phoneClientID, "/phone/sms/login"), {
      phone_number: phoneNumber,
      code
    });
  }

  jwks() {
    return this.get(this.oidcPath("/protocol/openid-connect/certs"));
  }

  validateToken(token) {
    return this.post(this.oidcPath("/protocol/openid-connect/token/introspect"), { token });
  }

  oidcPath(endpoint) {
    return `/realms/${this.realm}/${String(endpoint).replace(/^\/+/, "")}`;
  }
}

class AnalyticsClient extends ServiceClient {
  calculateTechnicalAnalysis(request) {
    return this.post("/analytics/technical-analysis:calculate", request);
  }

  indicators() {
    return this.get("/analytics/indicators");
  }

  upsertCandles(request) {
    return this.post("/analytics/candles:upsert", request);
  }

  optimizePortfolio(request) {
    return this.post("/portfolio-manager/optimizer", request);
  }

  rankings() {
    return this.get("/portfolio-manager/rankings");
  }

  balances(accountID) {
    return this.get(`/portfolio-manager/balances/${encodeURIComponent(accountID)}`);
  }

  blackScholes(query = "") {
    return this.get(`/risk-manager/blackscholes${query ? `?${query}` : ""}`);
  }

  staking() {
    return this.get("/staking");
  }
}

class MCPClient extends ServiceClient {
  constructor(baseURL, options = {}) {
    super(baseURL, options);
    this.nextID = 1;
  }

  call(method, params) {
    return this.post("/mcp", {
      jsonrpc: "2.0",
      id: this.nextID++,
      method,
      params
    });
  }

  callTool(name, arguments_) {
    return this.call("tools/call", { name, arguments: arguments_ || {} });
  }

  listTools() {
    return this.call("tools/list");
  }

  listPrompts() {
    return this.call("prompts/list");
  }

  listResources() {
    return this.call("resources/list");
  }

  readResource(uri) {
    return this.call("resources/read", { uri });
  }

  getPrompt(name, arguments_ = {}) {
    return this.call("prompts/get", { name, arguments: arguments_ });
  }

  conversationPlan(arguments_) {
    return this.callTool("finance_assistant.conversation.plan", arguments_);
  }
}

class OrchestratorClient extends ServiceClient {
  catalog() {
    return this.get("/v1/mcphost/catalog/packages");
  }

  createSession(request) {
    return this.post("/v1/mcphost/sessions", request);
  }

  sendMessage(request) {
    return this.post("/v1/mcphost/messages", request);
  }

  tools() {
    return this.get("/v1/mcphost/tools");
  }

  prompts() {
    return this.get("/v1/mcphost/prompts");
  }

  resources() {
    return this.get("/v1/mcphost/resources");
  }

  providers() {
    return this.get("/v1/mcphost/providers");
  }

  approvals() {
    return this.get("/v1/mcphost/approvals");
  }

  observability() {
    return this.get("/v1/mcphost/observability");
  }

  deleteSession(id) {
    return this.delete(`/v1/mcphost/sessions/${encodeURIComponent(id)}`);
  }
}

class TradingControlClient extends ServiceClient {
  robots() {
    return this.get("/robots");
  }

  createRobot(request) {
    return this.post("/robots", request);
  }

  robot(id) {
    return this.get(`/robots/${encodeURIComponent(id)}`);
  }

  startRobot(id) {
    return this.post(`/robots/${encodeURIComponent(id)}:start`);
  }

  pauseRobot(id) {
    return this.post(`/robots/${encodeURIComponent(id)}:pause`);
  }

  resumeRobot(id) {
    return this.post(`/robots/${encodeURIComponent(id)}:resume`);
  }

  stopRobot(id) {
    return this.post(`/robots/${encodeURIComponent(id)}:stop`);
  }

  riskPolicy(id) {
    return this.get(`/robots/${encodeURIComponent(id)}/risk-policy`);
  }

  setRiskPolicy(id, request) {
    return this.put(`/robots/${encodeURIComponent(id)}/risk-policy`, request);
  }

  riskView(id) {
    return this.get(`/risk-view/${encodeURIComponent(id)}`);
  }

  strategies() {
    return this.get("/strategies");
  }

  createStrategy(request) {
    return this.post("/strategies", request);
  }

  directives() {
    return this.get("/directives");
  }

  createDirective(request) {
    return this.post("/directives", request);
  }

  audit() {
    return this.get("/audit");
  }

  activity() {
    return this.get("/activity");
  }

  mcpTools() {
    return this.get("/mcp/tools");
  }
}

class KeyStoreClient extends ServiceClient {
  health() {
    return this.get("/healthz");
  }

  readiness() {
    return this.get("/readyz");
  }

  startKeygen(request) {
    return this.post("/keystore/keygen/start", request);
  }

  keygenSignature(request) {
    return this.post("/keystore/keygen/signature", request);
  }

  startSigning(request) {
    return this.post("/keystore/signing/start", request);
  }

  signingSignature(request) {
    return this.post("/keystore/signing/signature", request);
  }

  startResharing(request) {
    return this.post("/keystore/resharing/start", request);
  }

  keys(userPublicKey) {
    return this.get(`/keystore/keys/${encodeURIComponent(userPublicKey)}`);
  }

  signatures(userPublicKey) {
    return this.get(`/keystore/signatures/${encodeURIComponent(userPublicKey)}`);
  }

  metrics() {
    return this.get("/keystore/tss/metrics");
  }
}

class NetworkClient extends ServiceClient {
  virtualMachine() {
    return this.post("/v2/2finance-network/query", {
      method: "get_blocks",
      params: { page: 1, limit: 1 },
    });
  }

  marketCandles(market, query = "") {
    return this.get(`/v2/2finance-network/markets/${encodeURIComponent(market)}/candles${query ? `?${query}` : ""}`);
  }

  products(productType) {
    return this.get(`/v2/2finance-network/products/${encodeURIComponent(productType)}`);
  }

  createProduct(productType, request) {
    return this.post(`/v2/2finance-network/products/${encodeURIComponent(productType)}`, request);
  }

  bonds() {
    return this.products("bonds");
  }

  createBond(request) {
    return this.createProduct("bonds", request);
  }

  loans() {
    return this.products("loans");
  }

  createLoan(request) {
    return this.createProduct("loans", request);
  }

  swaps() {
    return this.products("swaps");
  }

  createSwap(request) {
    return this.createProduct("swaps", request);
  }

  stakingProducts() {
    return this.products("staking");
  }

  createStakingProduct(request) {
    return this.createProduct("staking", request);
  }

  syntheticAssets() {
    return this.products("synthetic-assets");
  }

  createSyntheticAsset(request) {
    return this.createProduct("synthetic-assets", request);
  }

  liquidityPools() {
    return this.products("liquidity-pools");
  }

  createLiquidityPool(request) {
    return this.createProduct("liquidity-pools", request);
  }
}

class HummingbotClient extends ServiceClient {
  assets() {
    return this.get("/api/v1/assets");
  }

  symbols() {
    return this.get("/api/v1/symbols");
  }

  balances() {
    return this.get("/api/v1/balances");
  }

  connectorConfig(request) {
    return this.post("/api/v1/connectors/2finance/config", request);
  }
}

class ProviderClient extends ServiceClient {}

class WiseClient extends ProviderClient {
  profiles() {
    return this.get("/v1/profiles");
  }

  profile(profileID) {
    return this.get(`/v1/profiles/${encodeURIComponent(profileID)}`);
  }

  createQuote(profileID, request) {
    return this.post(`/v3/profiles/${encodeURIComponent(profileID)}/quotes`, request);
  }

  createTransfer(request) {
    return this.post("/v1/transfers", request);
  }
}

class AirwallexClient extends ProviderClient {
  accounts() {
    return this.get("/api/v1/accounts");
  }

  payments() {
    return this.get("/api/v1/payments");
  }

  createPayment(request) {
    return this.post("/api/v1/payments", request);
  }

  beneficiaries() {
    return this.get("/api/v1/beneficiaries");
  }

  createBeneficiary(request) {
    return this.post("/api/v1/beneficiaries", request);
  }
}

class MatchEngineClient {
  constructor(webSocketURL) {
    this.webSocketURL = String(webSocketURL || "").trim();
  }

  orderCommand(command) {
    return {
      schema: "matchengine.order_command.v2",
      message_type: "ORDER",
      operation: "ADD",
      ...command
    };
  }

  marketDataSubscribe(request) {
    return {
      schema: "matchengine.market_data_subscribe.v2",
      ...(request || {})
    };
  }

  sendOrder(transport, command) {
    return transport.send(this.orderCommand(command));
  }

  subscribeMarketData(transport, request) {
    return transport.send(this.marketDataSubscribe(request));
  }
}

class PlannerClient {
  constructor({ mcp, orchestrator, analytics, tradingControl }) {
    this.mcp = mcp;
    this.orchestrator = orchestrator;
    this.analytics = analytics;
    this.tradingControl = tradingControl;
  }

  conversationPlan(request) {
    return this.mcp.conversationPlan(request);
  }

  orchestratedPlan(request) {
    return this.orchestrator.sendMessage(request);
  }

  operationalPlan(request) {
    return this.orchestratedPlan(request);
  }

  async tradingPlan(request) {
    const context = { ...(request.context || {}) };
    if (request.useTrading && this.tradingControl) {
      try {
        context.trading_robots = await this.tradingControl.robots();
      } catch (_) {
        // Best-effort enrichment keeps planner usable when trading is unavailable.
      }
    }
    if (request.useAnalytics && this.analytics) {
      try {
        context.analytics_indicators = await this.analytics.indicators();
      } catch (_) {
        // Best-effort enrichment keeps planner usable when analytics is unavailable.
      }
    }
    return this.conversationPlan({ ...request, context });
  }
}

function configFromEnv(env = process.env) {
  return {
    authURL: env.TWO_FINANCE_AUTH_URL || "",
    networkURL: env.TWO_FINANCE_NETWORK_URL || "",
    analyticsURL: env.TWO_FINANCE_ANALYTICS_URL || "",
    orchestratorURL: env.TWO_FINANCE_ORCHESTRATOR_URL || "",
    mcpURL: env.TWO_FINANCE_MCP_URL || "",
    tradingControlURL: env.TWO_FINANCE_TRADING_CONTROL_URL || "",
    matchEngineWSURL: env.TWO_FINANCE_MATCHENGINE_WS_URL || "",
    keyStoreURL: env.TWO_FINANCE_KEYSTORE_URL || "",
    hummingbotURL: env.TWO_FINANCE_HUMMINGBOT_URL || "",
    wiseURL: env.TWO_FINANCE_WISE_URL || "",
    airwallexURL: env.TWO_FINANCE_AIRWALLEX_URL || "",
    authRealm: env.TWO_FINANCE_AUTH_REALM || DEFAULT_AUTH_REALM,
    authClientID: env.TWO_FINANCE_AUTH_CLIENT_ID || DEFAULT_AUTH_CLIENT_ID,
    authPhoneClientID: env.TWO_FINANCE_AUTH_PHONE_CLIENT_ID || DEFAULT_AUTH_PHONE_CLIENT_ID
  };
}

class TwoFinanceClient {
  constructor(config = {}) {
    this.config = config;
    const options = {
      fetch: config.fetch,
      tokenSource: config.tokenSource
    };
    this.auth = new AuthClient(config.authURL, {
      ...options,
      realm: config.authRealm,
      clientID: config.authClientID,
      phoneClientID: config.authPhoneClientID
    });
    this.network = new NetworkClient(config.networkURL, options);
    this.analytics = new AnalyticsClient(config.analyticsURL, options);
    this.orchestrator = new OrchestratorClient(config.orchestratorURL, options);
    this.mcp = new MCPClient(config.mcpURL, options);
    this.tradingControl = new TradingControlClient(config.tradingControlURL, options);
    this.matchEngine = new MatchEngineClient(config.matchEngineWSURL);
    this.keystore = new KeyStoreClient(config.keyStoreURL, options);
    this.hummingbot = new HummingbotClient(config.hummingbotURL, options);
    this.providers = {
      wise: new WiseClient(config.wiseURL, options),
      airwallex: new AirwallexClient(config.airwallexURL, options)
    };
    this.planner = new PlannerClient({
      mcp: this.mcp,
      orchestrator: this.orchestrator,
      analytics: this.analytics,
      tradingControl: this.tradingControl
    });
  }

  static fromEnv(env = process.env, overrides = {}) {
    return new TwoFinanceClient({ ...configFromEnv(env), ...overrides });
  }
}

module.exports = {
  AnalyticsClient,
  AirwallexClient,
  AuthClient,
  ClientCredentialsTokenSource,
  HummingbotClient,
  KeyStoreClient,
  MatchEngineClient,
  MCPClient,
  NetworkClient,
  OrchestratorClient,
  parseIdempotencyRecord,
  parseDomainOperationsCatalog,
  findDomainOperation,
  resolveDomainOperation,
  resolveCatalogOperation,
  parsePaginationResponse,
  parseSDKError,
  parseServiceCatalog,
  PlannerClient,
  ProviderClient,
  SERVICE_CATALOG,
  SDK_NAME,
  SDK_VERSION,
  serviceURL,
  serviceURLs,
  ServiceError,
  ServiceClient,
  StaticTokenSource,
  TradingControlClient,
  TwoFinanceClient,
  WiseClient,
  bearerAuthorization,
  configFromEnv,
  configuredServices,
  joinURL,
  missingServiceURLs
};
