export type JSONValue =
  | null
  | boolean
  | number
  | string
  | JSONValue[]
  | { [key: string]: JSONValue };

export const SDK_NAME = "2finance-sdk-client";
export const SDK_VERSION = "0.1.0";

export interface TokenSource {
  getToken(): Promise<string> | string;
}

export interface SDKConfig {
  authURL?: string;
  networkURL?: string;
  analyticsURL?: string;
  orchestratorURL?: string;
  mcpURL?: string;
  tradingControlURL?: string;
  matchEngineWSURL?: string;
  keyStoreURL?: string;
  hummingbotURL?: string;
  wiseURL?: string;
  airwallexURL?: string;
  authRealm?: string;
  authClientID?: string;
  authPhoneClientID?: string;
  fetch?: typeof fetch;
  tokenSource?: TokenSource;
}

export interface RequestOptions {
  headers?: Record<string, string>;
  idempotencyKey?: string;
  query?: Record<string, string | number | boolean | null | undefined>;
  pagination?: { page?: number; limit?: number };
  timeoutMs?: number;
  maxRetries?: number;
}

export interface SDKErrorPayload {
  error: string;
  message: string;
  code: string;
  details?: Record<string, unknown>;
}

export interface PaginationResponse<T = Record<string, unknown>> {
  items: T[];
  limit: number;
  cursor?: string;
  next_cursor?: string;
}

export interface IdempotencyRecord {
  idempotency_key: string;
  operation: string;
  scope: string;
  request_id: string;
}

export interface ServiceCatalogEntry {
  name: string;
  env: string;
}

export interface ServiceCatalog {
  services: ServiceCatalogEntry[];
}

export interface ConfiguredServiceEntry extends ServiceCatalogEntry {
  url: string;
}

export interface DomainOperation {
  name: string;
  method: string;
  path: string;
  path_params?: string[];
  query?: string[];
  request_schema?: string;
  response_schema?: string;
  notes?: string;
}

export interface ResolvedOperation {
  method: string;
  path: string;
}

export interface DomainOperationsDomain {
  name: string;
  env: string;
  transport?: "http" | "jsonrpc" | "websocket";
  description?: string;
  operations: DomainOperation[];
}

export interface DomainOperationsCatalog {
  schema: "sdk.domain_operations.v2";
  domains: DomainOperationsDomain[];
}

export const SERVICE_CATALOG: ServiceCatalog = {
  services: [
    { name: "auth", env: "TWO_FINANCE_AUTH_URL" },
    { name: "network", env: "TWO_FINANCE_NETWORK_URL" },
    { name: "analytics", env: "TWO_FINANCE_ANALYTICS_URL" },
    { name: "orchestrator", env: "TWO_FINANCE_ORCHESTRATOR_URL" },
    { name: "mcp", env: "TWO_FINANCE_MCP_URL" },
    { name: "planner", env: "TWO_FINANCE_MCP_URL" },
    { name: "tradingcontrol", env: "TWO_FINANCE_TRADING_CONTROL_URL" },
    { name: "matchengine", env: "TWO_FINANCE_MATCHENGINE_WS_URL" },
    { name: "keystore", env: "TWO_FINANCE_KEYSTORE_URL" },
    { name: "hummingbot", env: "TWO_FINANCE_HUMMINGBOT_URL" },
    { name: "wise", env: "TWO_FINANCE_WISE_URL" },
    { name: "airwallex", env: "TWO_FINANCE_AIRWALLEX_URL" }
  ]
};

export function parseSDKError(payload: unknown): SDKErrorPayload {
  return payload as SDKErrorPayload;
}

export function parsePaginationResponse<T = Record<string, unknown>>(payload: unknown): PaginationResponse<T> {
  return payload as PaginationResponse<T>;
}

export function parseIdempotencyRecord(payload: unknown): IdempotencyRecord {
  return payload as IdempotencyRecord;
}

export function parseServiceCatalog(payload: unknown): ServiceCatalog {
  return payload as ServiceCatalog;
}

export function parseDomainOperationsCatalog(payload: unknown): DomainOperationsCatalog {
  return payload as DomainOperationsCatalog;
}

export function findDomainOperation(
  catalog: DomainOperationsCatalog,
  domainName: string,
  operationName: string
): DomainOperation | undefined {
  const domain = catalog.domains.find((item) => serviceKey(item.name) === serviceKey(domainName));
  return domain?.operations.find((operation) => operation.name === operationName);
}

export function resolveDomainOperation(
  operation: DomainOperation,
  pathParams: Record<string, string | number | boolean> = {},
  query: RequestOptions["query"] = {}
): ResolvedOperation {
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

  return { method: operation.method.trim().toUpperCase(), path };
}

export function resolveCatalogOperation(
  catalog: DomainOperationsCatalog,
  domainName: string,
  operationName: string,
  pathParams: Record<string, string | number | boolean> = {},
  query: RequestOptions["query"] = {}
): ResolvedOperation {
  const operation = findDomainOperation(catalog, domainName, operationName);
  if (!operation) {
    throw new Error(`2finance: unknown operation ${domainName}.${operationName}`);
  }
  return resolveDomainOperation(operation, pathParams, query);
}

export function serviceURL(config: SDKConfig, domain: string): string {
  switch (serviceKey(domain)) {
    case "auth":
      return config.authURL || "";
    case "network":
      return config.networkURL || "";
    case "analytics":
      return config.analyticsURL || "";
    case "orchestrator":
      return config.orchestratorURL || "";
    case "mcp":
    case "planner":
      return config.mcpURL || "";
    case "tradingcontrol":
      return config.tradingControlURL || "";
    case "matchengine":
      return config.matchEngineWSURL || "";
    case "keystore":
      return config.keyStoreURL || "";
    case "hummingbot":
      return config.hummingbotURL || "";
    case "wise":
      return config.wiseURL || "";
    case "airwallex":
      return config.airwallexURL || "";
    default:
      return "";
  }
}

export function serviceURLs(config: SDKConfig): Record<string, string> {
  const urls: Record<string, string> = {};
  for (const service of SERVICE_CATALOG.services) {
    const url = serviceURL(config, service.name);
    if (url) {
      urls[service.name] = url;
    }
  }
  return urls;
}

export function configuredServices(config: SDKConfig): ConfiguredServiceEntry[] {
  return SERVICE_CATALOG.services
    .map((service) => ({ ...service, url: serviceURL(config, service.name) }))
    .filter((service) => service.url);
}

export function missingServiceURLs(config: SDKConfig): ServiceCatalogEntry[] {
  return SERVICE_CATALOG.services.filter((service) => !serviceURL(config, service.name));
}

function serviceKey(domain: string): string {
  return domain.trim().toLowerCase().replace(/[-_\s]+/g, "");
}

export const DEFAULT_AUTH_REALM = "2finance";
export const DEFAULT_AUTH_CLIENT_ID = "2finance-network";
export const DEFAULT_AUTH_PHONE_CLIENT_ID = "2finance-network-phone";

export class StaticTokenSource implements TokenSource {
  constructor(private readonly token: string) {}

  getToken(): string {
    return this.token;
  }
}

export interface ClientCredentialsTokenSourceOptions {
  tokenURL: string;
  clientID: string;
  clientSecret: string;
  scopes?: string[];
  fetch?: typeof fetch;
  expirySkewMs?: number;
}

export class ClientCredentialsTokenSource implements TokenSource {
  private accessToken = "";
  private expiresAt = 0;

  constructor(private readonly options: ClientCredentialsTokenSourceOptions) {}

  async getToken(): Promise<string> {
    const now = Date.now();
    const skewMs = this.options.expirySkewMs ?? 30000;
    if (this.accessToken && now < this.expiresAt - skewMs) {
      return this.accessToken;
    }
    if (!this.options.tokenURL || !this.options.clientID || !this.options.clientSecret) {
      throw new Error("2finance auth: tokenURL, clientID and clientSecret are required");
    }
    const fetchImpl = this.options.fetch || globalThis.fetch;
    if (!fetchImpl) {
      throw new Error("2finance auth: fetch is required");
    }
    const body = new URLSearchParams();
    body.set("grant_type", "client_credentials");
    body.set("client_id", this.options.clientID);
    body.set("client_secret", this.options.clientSecret);
    if (this.options.scopes?.length) {
      body.set("scope", this.options.scopes.join(" "));
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
    if (!payload.access_token || typeof payload.access_token !== "string") {
      throw new Error("2finance auth: token response missing access_token");
    }
    this.accessToken = payload.access_token;
    this.expiresAt = now + Number(payload.expires_in || 300) * 1000;
    return this.accessToken;
  }
}

export function bearerAuthorization(token: string): string {
  const trimmed = token.trim();
  if (!trimmed) {
    return "";
  }
  return trimmed.toLowerCase().startsWith("bearer ") ? trimmed : `Bearer ${trimmed}`;
}

export function joinURL(baseURL: string | undefined, path: string): string {
  const base = (baseURL || "").trim().replace(/\/+$/, "");
  if (!base) {
    throw new Error("2finance: baseURL is required");
  }
  if (/^https?:\/\//i.test(path)) {
    return path;
  }
  return `${base}/${path.replace(/^\/+/, "")}`;
}

export class ServiceError extends Error {
  constructor(
    readonly method: string,
    readonly url: string,
    readonly statusCode: number,
    readonly body: string
  ) {
    super(`2finance: ${method} ${url} returned ${statusCode}: ${body}`);
    this.name = "ServiceError";
  }
}

export class ServiceClient {
  protected readonly fetchImpl: typeof fetch;

  constructor(
    protected readonly baseURL: string | undefined,
    protected readonly options: Pick<SDKConfig, "fetch" | "tokenSource"> = {}
  ) {
    const fetchImpl = options.fetch || globalThis.fetch;
    if (!fetchImpl) {
      throw new Error("2finance: fetch is required");
    }
    this.fetchImpl = fetchImpl;
  }

  url(path: string): string {
    return joinURL(this.baseURL, path);
  }

  async request<T = unknown>(
    method: string,
    path: string,
    body?: unknown,
    requestOptions: RequestOptions = {}
  ): Promise<T> {
    const headers: Record<string, string> = { Accept: "application/json" };
    let requestBody: string | undefined;
    if (body !== undefined && body !== null) {
      headers["Content-Type"] = "application/json";
      requestBody = typeof body === "string" ? body : JSON.stringify(body);
    }
    if (this.options.tokenSource) {
      const token = await this.options.tokenSource.getToken();
      const authorization = bearerAuthorization(token);
      if (authorization) {
        headers.Authorization = authorization;
      }
    }
    Object.assign(headers, requestOptions.headers || {});
    const idempotencyKey = requestOptions.idempotencyKey?.trim();
    if (idempotencyKey) {
      headers["Idempotency-Key"] = idempotencyKey;
    }
    const requestURL = this.urlWithQuery(path, requestOptions.query, requestOptions.pagination);
    const abortController = requestOptions.timeoutMs ? new AbortController() : undefined;
    const timeout = abortController
      ? setTimeout(() => abortController.abort(), requestOptions.timeoutMs)
      : undefined;
    const attempts = 1 + Math.max(0, requestOptions.maxRetries || 0);
    try {
      let lastText = "";
      let lastStatus = 0;
      for (let attempt = 0; attempt < attempts; attempt++) {
        const response = await this.fetchImpl(requestURL, {
          method,
          headers,
          body: requestBody,
          signal: abortController?.signal
        });
        const text = await response.text();
        if (response.ok) {
          const payload = text ? JSON.parse(text) : null;
          return payload as T;
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

  requestOperation<T = unknown>(
    operation: ResolvedOperation,
    body?: unknown,
    requestOptions: RequestOptions = {}
  ): Promise<T> {
    return this.request<T>(operation.method, operation.path, body, requestOptions);
  }

  requestCatalogOperation<T = unknown>(
    catalog: DomainOperationsCatalog,
    domainName: string,
    operationName: string,
    pathParams: Record<string, string | number | boolean> = {},
    query: RequestOptions["query"] = {},
    body?: unknown,
    requestOptions: RequestOptions = {}
  ): Promise<T> {
    const operation = resolveCatalogOperation(catalog, domainName, operationName, pathParams, query);
    return this.requestOperation<T>(operation, body, requestOptions);
  }

  protected urlWithQuery(
    path: string,
    query?: RequestOptions["query"],
    pagination?: RequestOptions["pagination"]
  ): string {
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

  get<T = unknown>(path: string, options?: RequestOptions): Promise<T> {
    return this.request<T>("GET", path, undefined, options);
  }

  post<T = unknown>(path: string, body?: unknown, options?: RequestOptions): Promise<T> {
    return this.request<T>("POST", path, body, options);
  }

  put<T = unknown>(path: string, body?: unknown, options?: RequestOptions): Promise<T> {
    return this.request<T>("PUT", path, body, options);
  }

  delete<T = unknown>(path: string, options?: RequestOptions): Promise<T> {
    return this.request<T>("DELETE", path, undefined, options);
  }
}

function isRetryableStatus(status: number): boolean {
  return status === 429 || status >= 500;
}

export class AuthClient extends ServiceClient {
  private readonly realm: string;
  private readonly clientID: string;
  private readonly phoneClientID: string;

  constructor(baseURL: string | undefined, config: SDKConfig = {}) {
    super(baseURL, config);
    this.realm = config.authRealm || DEFAULT_AUTH_REALM;
    this.clientID = config.authClientID || DEFAULT_AUTH_CLIENT_ID;
    this.phoneClientID = config.authPhoneClientID || DEFAULT_AUTH_PHONE_CLIENT_ID;
  }

  private authPath(clientID: string, endpoint: string): string {
    return `/v1/2finance-authenticator/${this.realm}/${clientID}/${endpoint.replace(/^\/+/, "")}`;
  }

  login(input: unknown): Promise<unknown> {
    return this.post(this.authPath(this.clientID, "/login"), input);
  }

  signUp(input: unknown): Promise<unknown> {
    return this.post(this.authPath(this.clientID, "/signup"), input);
  }

  refreshToken(refreshToken: string): Promise<unknown> {
    return this.post(this.authPath(this.clientID, "/refresh"), { refresh_token: refreshToken });
  }

  phoneLogin(phoneNumber: string, code: string): Promise<unknown> {
    return this.post(this.authPath(this.phoneClientID, "/phone/sms/login"), {
      phone_number: phoneNumber,
      code
    });
  }

  jwks(): Promise<unknown> {
    return this.get(this.oidcPath("/protocol/openid-connect/certs"));
  }

  validateToken(token: string): Promise<unknown> {
    return this.post(this.oidcPath("/protocol/openid-connect/token/introspect"), { token });
  }

  private oidcPath(endpoint: string): string {
    return `/realms/${this.realm}/${endpoint.replace(/^\/+/, "")}`;
  }
}

export class AnalyticsClient extends ServiceClient {
  indicators(): Promise<unknown> {
    return this.get("/analytics/indicators");
  }

  calculateTechnicalAnalysis(request: unknown): Promise<unknown> {
    return this.post("/analytics/technical-analysis:calculate", request);
  }

  optimizePortfolio(request: unknown): Promise<unknown> {
    return this.post("/portfolio-manager/optimizer", request);
  }

  upsertCandles(request: unknown): Promise<unknown> {
    return this.post("/analytics/candles:upsert", request);
  }

  rankings(): Promise<unknown> {
    return this.get("/portfolio-manager/rankings");
  }

  balances(accountID: string): Promise<unknown> {
    return this.get(`/portfolio-manager/balances/${encodeURIComponent(accountID)}`);
  }

  blackScholes(query = ""): Promise<unknown> {
    return this.get(`/risk-manager/blackscholes${query ? `?${query}` : ""}`);
  }

  staking(): Promise<unknown> {
    return this.get("/staking");
  }
}

export class MCPClient extends ServiceClient {
  private nextID = 1;

  call(method: string, params?: unknown): Promise<unknown> {
    return this.post("/mcp", {
      jsonrpc: "2.0",
      id: this.nextID++,
      method,
      params
    });
  }

  callTool(name: string, args: unknown = {}): Promise<unknown> {
    return this.call("tools/call", { name, arguments: args });
  }

  listTools(): Promise<unknown> {
    return this.call("tools/list");
  }

  listPrompts(): Promise<unknown> {
    return this.call("prompts/list");
  }

  listResources(): Promise<unknown> {
    return this.call("resources/list");
  }

  readResource(uri: string): Promise<unknown> {
    return this.call("resources/read", { uri });
  }

  getPrompt(name: string, args: unknown = {}): Promise<unknown> {
    return this.call("prompts/get", { name, arguments: args });
  }

  conversationPlan(args: unknown): Promise<unknown> {
    return this.callTool("finance_assistant.conversation.plan", args);
  }
}

export class OrchestratorClient extends ServiceClient {
  catalog(): Promise<unknown> {
    return this.get("/v1/mcphost/catalog/packages");
  }

  createSession(request: unknown): Promise<unknown> {
    return this.post("/v1/mcphost/sessions", request);
  }

  sendMessage(request: unknown): Promise<unknown> {
    return this.post("/v1/mcphost/messages", request);
  }

  tools(): Promise<unknown> {
    return this.get("/v1/mcphost/tools");
  }

  prompts(): Promise<unknown> {
    return this.get("/v1/mcphost/prompts");
  }

  resources(): Promise<unknown> {
    return this.get("/v1/mcphost/resources");
  }

  providers(): Promise<unknown> {
    return this.get("/v1/mcphost/providers");
  }

  approvals(): Promise<unknown> {
    return this.get("/v1/mcphost/approvals");
  }

  observability(): Promise<unknown> {
    return this.get("/v1/mcphost/observability");
  }

  deleteSession(id: string): Promise<unknown> {
    return this.delete(`/v1/mcphost/sessions/${encodeURIComponent(id)}`);
  }
}

export class TradingControlClient extends ServiceClient {
  robots(): Promise<unknown> {
    return this.get("/robots");
  }

  createRobot(request: unknown): Promise<unknown> {
    return this.post("/robots", request);
  }

  startRobot(id: string): Promise<unknown> {
    return this.post(`/robots/${encodeURIComponent(id)}:start`);
  }

  robot(id: string): Promise<unknown> {
    return this.get(`/robots/${encodeURIComponent(id)}`);
  }

  pauseRobot(id: string): Promise<unknown> {
    return this.post(`/robots/${encodeURIComponent(id)}:pause`);
  }

  resumeRobot(id: string): Promise<unknown> {
    return this.post(`/robots/${encodeURIComponent(id)}:resume`);
  }

  stopRobot(id: string): Promise<unknown> {
    return this.post(`/robots/${encodeURIComponent(id)}:stop`);
  }

  riskPolicy(id: string): Promise<unknown> {
    return this.get(`/robots/${encodeURIComponent(id)}/risk-policy`);
  }

  setRiskPolicy(id: string, request: unknown): Promise<unknown> {
    return this.put(`/robots/${encodeURIComponent(id)}/risk-policy`, request);
  }

  riskView(id: string): Promise<unknown> {
    return this.get(`/risk-view/${encodeURIComponent(id)}`);
  }

  strategies(): Promise<unknown> {
    return this.get("/strategies");
  }

  createStrategy(request: unknown): Promise<unknown> {
    return this.post("/strategies", request);
  }

  directives(): Promise<unknown> {
    return this.get("/directives");
  }

  createDirective(request: unknown): Promise<unknown> {
    return this.post("/directives", request);
  }

  audit(): Promise<unknown> {
    return this.get("/audit");
  }

  activity(): Promise<unknown> {
    return this.get("/activity");
  }

  mcpTools(): Promise<unknown> {
    return this.get("/mcp/tools");
  }
}

export class KeyStoreClient extends ServiceClient {
  health(): Promise<unknown> {
    return this.get("/healthz");
  }

  readiness(): Promise<unknown> {
    return this.get("/readyz");
  }

  startKeygen(request: unknown): Promise<unknown> {
    return this.post("/keystore/keygen/start", request);
  }

  startSigning(request: unknown): Promise<unknown> {
    return this.post("/keystore/signing/start", request);
  }

  keygenSignature(request: unknown): Promise<unknown> {
    return this.post("/keystore/keygen/signature", request);
  }

  signingSignature(request: unknown): Promise<unknown> {
    return this.post("/keystore/signing/signature", request);
  }

  startResharing(request: unknown): Promise<unknown> {
    return this.post("/keystore/resharing/start", request);
  }

  keys(userPublicKey: string): Promise<unknown> {
    return this.get(`/keystore/keys/${encodeURIComponent(userPublicKey)}`);
  }

  signatures(userPublicKey: string): Promise<unknown> {
    return this.get(`/keystore/signatures/${encodeURIComponent(userPublicKey)}`);
  }

  metrics(): Promise<unknown> {
    return this.get("/keystore/tss/metrics");
  }
}

export type MarketRouteStatus = "pending" | "active" | "draining" | "migrating" | "halted" | "retired";

export interface MarketRange {
  min: string | null;
  max: string | null;
}

export interface MarketDefinition {
  schema: "2finance.market_definition.v1";
  id: string;
  market_id: string;
  market: string;
  symbol: string;
  engine_id: string;
  symbol_id: number;
  route_epoch: number;
  chain: { id: string; network: string; address: string };
  address: string;
  endpoints: {
    http: string;
    websocket: string;
    orderbook_snapshot: string;
    trades: string;
    recovery: string;
  };
  status: MarketRouteStatus;
  active: boolean;
  environment: string;
  protocol_version: string;
  base: string;
  quote: string;
  base_id: string;
  quote_id: string;
  type: "spot";
  spot: true;
  precision: { amount: number; price: number; base: number; quote: number };
  tick_size: string;
  step_size: string;
  limits: { amount: MarketRange; price: MarketRange; cost: MarketRange; notional: MarketRange };
  fees: { maker: string; taker: string; asset_policy: "base" | "quote" | "received" | "native" };
  capabilities: {
    order_types: string[];
    time_in_force: string[];
    channels: string[];
    operations: string[];
    snapshot: boolean;
    recovery: boolean;
  };
  freshness: { health: "healthy" | "degraded" | "stale" | "unknown"; last_sequence: number; observed_at: string };
}

export interface MarketDirectory {
  schema: "2finance.market_directory.v1";
  generated_at: string;
  markets: MarketDefinition[];
}

export class NetworkClient extends ServiceClient {
  virtualMachine(): Promise<unknown> {
    return this.post("/v2/2finance-network/query", {
      method: "get_blocks",
      params: { page: 1, limit: 1 },
    });
  }

  marketCandles(market: string, query = ""): Promise<unknown> {
    return this.get(`/v2/2finance-network/markets/${encodeURIComponent(market)}/candles${query ? `?${query}` : ""}`);
  }

  marketDirectory(): Promise<MarketDirectory> {
    return this.get("/api/v2/exchange/market-directory") as Promise<MarketDirectory>;
  }

  products(productType: string): Promise<unknown> {
    return this.get(`/v2/2finance-network/products/${encodeURIComponent(productType)}`);
  }

  createProduct(productType: string, request: unknown): Promise<unknown> {
    return this.post(`/v2/2finance-network/products/${encodeURIComponent(productType)}`, request);
  }

  bonds(): Promise<unknown> {
    return this.products("bonds");
  }

  createBond(request: unknown): Promise<unknown> {
    return this.createProduct("bonds", request);
  }

  loans(): Promise<unknown> {
    return this.products("loans");
  }

  createLoan(request: unknown): Promise<unknown> {
    return this.createProduct("loans", request);
  }

  swaps(): Promise<unknown> {
    return this.products("swaps");
  }

  createSwap(request: unknown): Promise<unknown> {
    return this.createProduct("swaps", request);
  }

  stakingProducts(): Promise<unknown> {
    return this.products("staking");
  }

  createStakingProduct(request: unknown): Promise<unknown> {
    return this.createProduct("staking", request);
  }

  syntheticAssets(): Promise<unknown> {
    return this.products("synthetic-assets");
  }

  createSyntheticAsset(request: unknown): Promise<unknown> {
    return this.createProduct("synthetic-assets", request);
  }

  liquidityPools(): Promise<unknown> {
    return this.products("liquidity-pools");
  }

  createLiquidityPool(request: unknown): Promise<unknown> {
    return this.createProduct("liquidity-pools", request);
  }
}

export class HummingbotClient extends ServiceClient {
  assets(): Promise<unknown> {
    return this.get("/api/v1/assets");
  }

  symbols(): Promise<unknown> {
    return this.get("/api/v1/symbols");
  }

  balances(): Promise<unknown> {
    return this.get("/api/v1/balances");
  }

  connectorConfig(request: unknown): Promise<unknown> {
    return this.post("/api/v1/connectors/2finance/config", request);
  }
}

export class ProviderClient extends ServiceClient {}

export class WiseClient extends ProviderClient {
  profiles(): Promise<unknown> {
    return this.get("/v1/profiles");
  }

  profile(profileID: string): Promise<unknown> {
    return this.get(`/v1/profiles/${encodeURIComponent(profileID)}`);
  }

  createQuote(profileID: string, request: unknown): Promise<unknown> {
    return this.post(`/v3/profiles/${encodeURIComponent(profileID)}/quotes`, request);
  }

  createTransfer(request: unknown): Promise<unknown> {
    return this.post("/v1/transfers", request);
  }
}

export class AirwallexClient extends ProviderClient {
  accounts(): Promise<unknown> {
    return this.get("/api/v1/accounts");
  }

  payments(): Promise<unknown> {
    return this.get("/api/v1/payments");
  }

  createPayment(request: unknown): Promise<unknown> {
    return this.post("/api/v1/payments", request);
  }

  beneficiaries(): Promise<unknown> {
    return this.get("/api/v1/beneficiaries");
  }

  createBeneficiary(request: unknown): Promise<unknown> {
    return this.post("/api/v1/beneficiaries", request);
  }
}

export interface ProvidersClient {
  wise: WiseClient;
  airwallex: AirwallexClient;
}

export interface MatchEngineOrderCommand {
  schema?: string;
  market_id?: string;
  engine_id?: string;
  route_epoch?: number;
  message_type?: "ORDER";
  operation?: "ADD" | "MODIFY" | "REPLACE" | "DELETE" | "MITIGATE";
  order_type?: "LIMIT" | "MARKET" | "STOP" | "STOP_LIMIT" | "TRAILING_STOP" | "TRAILING_STOP_LIMIT";
  client_order_id: string;
  idempotency_key: string;
  wallet_id: number;
  symbol_id: number;
  order_id?: number;
  side: "BUY" | "SELL";
  quantity?: string;
  price?: string;
  stop_price?: string;
  slippage?: string;
  max_visible_quantity?: string;
  trailing_distance?: string;
  trailing_step?: string;
  time_in_force?: string;
  agent_id?: string;
  controller_id?: string;
  executor_id?: string;
  strategy_id?: string;
  metadata?: unknown;
}

export interface MatchEngineMarketDataSubscribe {
  schema?: string;
  symbols?: string[];
  channels?: string[];
  interval?: string;
  account_id?: string;
  metadata?: unknown;
}

export interface MatchEngineTransport {
  send(message: unknown): Promise<unknown> | unknown;
}

export class MatchEngineClient {
  readonly webSocketURL: string;

  constructor(webSocketURL: string | undefined) {
    this.webSocketURL = (webSocketURL || "").trim();
  }

  orderCommand(command: MatchEngineOrderCommand): Required<Pick<MatchEngineOrderCommand, "schema">> & MatchEngineOrderCommand {
    return {
      schema: "matchengine.order_command.v2",
      message_type: "ORDER",
      operation: "ADD",
      ...command
    };
  }

  marketDataSubscribe(request: MatchEngineMarketDataSubscribe): Required<Pick<MatchEngineMarketDataSubscribe, "schema">> & MatchEngineMarketDataSubscribe {
    return {
      schema: "matchengine.market_data_subscribe.v2",
      ...request
    };
  }

  sendOrder(transport: MatchEngineTransport, command: MatchEngineOrderCommand): Promise<unknown> | unknown {
    return transport.send(this.orderCommand(command));
  }

  subscribeMarketData(transport: MatchEngineTransport, request: MatchEngineMarketDataSubscribe): Promise<unknown> | unknown {
    return transport.send(this.marketDataSubscribe(request));
  }
}

export class PlannerClient {
  constructor(
    private readonly mcp: MCPClient,
    private readonly orchestrator: OrchestratorClient,
    private readonly analytics: AnalyticsClient,
    private readonly tradingControl: TradingControlClient
  ) {}

  conversationPlan(request: unknown): Promise<unknown> {
    return this.mcp.conversationPlan(request);
  }

  orchestratedPlan(request: unknown): Promise<unknown> {
    return this.orchestrator.sendMessage(request);
  }

  operationalPlan(request: unknown): Promise<unknown> {
    return this.orchestratedPlan(request);
  }

  async tradingPlan(request: { context?: Record<string, unknown>; useAnalytics?: boolean; useTrading?: boolean; [key: string]: unknown }): Promise<unknown> {
    const context = { ...(request.context || {}) };
    if (request.useTrading) {
      try {
        context.trading_robots = await this.tradingControl.robots();
      } catch {
        // Best-effort enrichment keeps planning usable when trading is unavailable.
      }
    }
    if (request.useAnalytics) {
      try {
        context.analytics_indicators = await this.analytics.indicators();
      } catch {
        // Best-effort enrichment keeps planning usable when analytics is unavailable.
      }
    }
    return this.conversationPlan({ ...request, context });
  }
}

export class TwoFinanceClient {
  readonly auth: AuthClient;
  readonly network: NetworkClient;
  readonly analytics: AnalyticsClient;
  readonly orchestrator: OrchestratorClient;
  readonly mcp: MCPClient;
  readonly tradingControl: TradingControlClient;
  readonly matchEngine: MatchEngineClient;
  readonly keystore: KeyStoreClient;
  readonly hummingbot: HummingbotClient;
  readonly providers: ProvidersClient;
  readonly planner: PlannerClient;

  constructor(readonly config: SDKConfig = {}) {
    this.auth = new AuthClient(config.authURL, config);
    this.network = new NetworkClient(config.networkURL, config);
    this.analytics = new AnalyticsClient(config.analyticsURL, config);
    this.orchestrator = new OrchestratorClient(config.orchestratorURL, config);
    this.mcp = new MCPClient(config.mcpURL, config);
    this.tradingControl = new TradingControlClient(config.tradingControlURL, config);
    this.matchEngine = new MatchEngineClient(config.matchEngineWSURL);
    this.keystore = new KeyStoreClient(config.keyStoreURL, config);
    this.hummingbot = new HummingbotClient(config.hummingbotURL, config);
    this.providers = {
      wise: new WiseClient(config.wiseURL, config),
      airwallex: new AirwallexClient(config.airwallexURL, config)
    };
    this.planner = new PlannerClient(this.mcp, this.orchestrator, this.analytics, this.tradingControl);
  }
}

export function configFromEnv(env: Record<string, string | undefined>): SDKConfig {
  return {
    authURL: env.TWO_FINANCE_AUTH_URL,
    networkURL: env.TWO_FINANCE_NETWORK_URL,
    analyticsURL: env.TWO_FINANCE_ANALYTICS_URL,
    orchestratorURL: env.TWO_FINANCE_ORCHESTRATOR_URL,
    mcpURL: env.TWO_FINANCE_MCP_URL,
    tradingControlURL: env.TWO_FINANCE_TRADING_CONTROL_URL,
    matchEngineWSURL: env.TWO_FINANCE_MATCHENGINE_WS_URL,
    keyStoreURL: env.TWO_FINANCE_KEYSTORE_URL,
    hummingbotURL: env.TWO_FINANCE_HUMMINGBOT_URL,
    wiseURL: env.TWO_FINANCE_WISE_URL,
    airwallexURL: env.TWO_FINANCE_AIRWALLEX_URL,
    authRealm: env.TWO_FINANCE_AUTH_REALM || DEFAULT_AUTH_REALM,
    authClientID: env.TWO_FINANCE_AUTH_CLIENT_ID || DEFAULT_AUTH_CLIENT_ID,
    authPhoneClientID: env.TWO_FINANCE_AUTH_PHONE_CLIENT_ID || DEFAULT_AUTH_PHONE_CLIENT_ID
  };
}
