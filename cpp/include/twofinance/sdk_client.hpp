#pragma once

#include <cstdlib>
#include <functional>
#include <iomanip>
#include <map>
#include <sstream>
#include <stdexcept>
#include <string>
#include <utility>
#include <vector>

namespace twofinance {

inline constexpr const char* SDK_NAME = "2finance-sdk-client";
inline constexpr const char* SDK_VERSION = "0.1.0";

struct SdkConfig {
  std::string auth_url;
  std::string network_url;
  std::string analytics_url;
  std::string orchestrator_url;
  std::string mcp_url;
  std::string trading_control_url;
  std::string matchengine_ws_url;
  std::string keystore_url;
  std::string hummingbot_url;
  std::string wise_url;
  std::string airwallex_url;
  std::string auth_realm = "2finance";
  std::string auth_client_id = "2finance-network";
  std::string auth_phone_client_id = "2finance-network-phone";
};

inline std::string env_value(const std::map<std::string, std::string>& env, const std::string& name,
                             const std::string& fallback = "") {
  auto value = env.find(name);
  return value == env.end() || value->second.empty() ? fallback : value->second;
}

inline SdkConfig config_from_env(const std::map<std::string, std::string>& env) {
  SdkConfig config;
  config.auth_url = env_value(env, "TWO_FINANCE_AUTH_URL");
  config.network_url = env_value(env, "TWO_FINANCE_NETWORK_URL");
  config.analytics_url = env_value(env, "TWO_FINANCE_ANALYTICS_URL");
  config.orchestrator_url = env_value(env, "TWO_FINANCE_ORCHESTRATOR_URL");
  config.mcp_url = env_value(env, "TWO_FINANCE_MCP_URL");
  config.trading_control_url = env_value(env, "TWO_FINANCE_TRADING_CONTROL_URL");
  config.matchengine_ws_url = env_value(env, "TWO_FINANCE_MATCHENGINE_WS_URL");
  config.keystore_url = env_value(env, "TWO_FINANCE_KEYSTORE_URL");
  config.hummingbot_url = env_value(env, "TWO_FINANCE_HUMMINGBOT_URL");
  config.wise_url = env_value(env, "TWO_FINANCE_WISE_URL");
  config.airwallex_url = env_value(env, "TWO_FINANCE_AIRWALLEX_URL");
  config.auth_realm = env_value(env, "TWO_FINANCE_AUTH_REALM", "2finance");
  config.auth_client_id = env_value(env, "TWO_FINANCE_AUTH_CLIENT_ID", "2finance-network");
  config.auth_phone_client_id = env_value(env, "TWO_FINANCE_AUTH_PHONE_CLIENT_ID", "2finance-network-phone");
  return config;
}

inline std::string getenv_string(const char* name) {
  const char* value = std::getenv(name);
  return value == nullptr ? "" : std::string(value);
}

inline SdkConfig config_from_environment() {
  return config_from_env({{"TWO_FINANCE_AUTH_URL", getenv_string("TWO_FINANCE_AUTH_URL")},
                          {"TWO_FINANCE_NETWORK_URL", getenv_string("TWO_FINANCE_NETWORK_URL")},
                          {"TWO_FINANCE_ANALYTICS_URL", getenv_string("TWO_FINANCE_ANALYTICS_URL")},
                          {"TWO_FINANCE_ORCHESTRATOR_URL", getenv_string("TWO_FINANCE_ORCHESTRATOR_URL")},
                          {"TWO_FINANCE_MCP_URL", getenv_string("TWO_FINANCE_MCP_URL")},
                          {"TWO_FINANCE_TRADING_CONTROL_URL", getenv_string("TWO_FINANCE_TRADING_CONTROL_URL")},
                          {"TWO_FINANCE_MATCHENGINE_WS_URL", getenv_string("TWO_FINANCE_MATCHENGINE_WS_URL")},
                          {"TWO_FINANCE_KEYSTORE_URL", getenv_string("TWO_FINANCE_KEYSTORE_URL")},
                          {"TWO_FINANCE_HUMMINGBOT_URL", getenv_string("TWO_FINANCE_HUMMINGBOT_URL")},
                          {"TWO_FINANCE_WISE_URL", getenv_string("TWO_FINANCE_WISE_URL")},
                          {"TWO_FINANCE_AIRWALLEX_URL", getenv_string("TWO_FINANCE_AIRWALLEX_URL")},
                          {"TWO_FINANCE_AUTH_REALM", getenv_string("TWO_FINANCE_AUTH_REALM")},
                          {"TWO_FINANCE_AUTH_CLIENT_ID", getenv_string("TWO_FINANCE_AUTH_CLIENT_ID")},
                          {"TWO_FINANCE_AUTH_PHONE_CLIENT_ID", getenv_string("TWO_FINANCE_AUTH_PHONE_CLIENT_ID")}});
}

inline std::string service_key(std::string domain) {
  std::string key;
  for (char c : domain) {
    if (c == '-' || c == '_' || c == ' ') {
      continue;
    }
    if (c >= 'A' && c <= 'Z') {
      key.push_back(static_cast<char>(c - 'A' + 'a'));
    } else {
      key.push_back(c);
    }
  }
  return key;
}

inline std::string service_url(const SdkConfig& config, const std::string& domain) {
  std::string key = service_key(domain);
  if (key == "auth") return config.auth_url;
  if (key == "network") return config.network_url;
  if (key == "analytics") return config.analytics_url;
  if (key == "orchestrator") return config.orchestrator_url;
  if (key == "mcp" || key == "planner") return config.mcp_url;
  if (key == "tradingcontrol") return config.trading_control_url;
  if (key == "matchengine") return config.matchengine_ws_url;
  if (key == "keystore") return config.keystore_url;
  if (key == "hummingbot") return config.hummingbot_url;
  if (key == "wise") return config.wise_url;
  if (key == "airwallex") return config.airwallex_url;
  return "";
}

struct HttpRequest {
  std::string method;
  std::string url;
  std::map<std::string, std::string> headers;
  std::string body;
  int timeout_ms = 0;
};

struct HttpResponse {
  int status_code = 0;
  std::string body;
};

struct RequestOptions {
  std::map<std::string, std::string> headers;
  std::string idempotency_key;
  std::map<std::string, std::string> query;
  int timeout_ms = 0;
  int max_retries = 0;
  int page = 0;
  int limit = 0;
};

struct SDKErrorPayload {
  std::string error;
  std::string message;
  std::string code;
  std::map<std::string, std::string> details;
};

struct PaginationResponse {
  int limit = 0;
  std::string cursor;
  std::string next_cursor;
};

struct IdempotencyRecord {
  std::string idempotency_key;
  std::string operation;
  std::string scope;
  std::string request_id;
};

struct ServiceCatalogEntry {
  std::string name;
  std::string env;
};

struct ServiceCatalog {
  std::vector<ServiceCatalogEntry> services;
};

struct ConfiguredServiceEntry {
  std::string name;
  std::string env;
  std::string url;
};

struct DomainOperation {
  std::string name;
  std::string method;
  std::string path;
  std::vector<std::string> path_params;
  std::vector<std::string> query;
  std::string request_schema;
  std::string response_schema;
  std::string notes;
};

struct ResolvedOperation {
  std::string method;
  std::string path;
};

struct DomainOperationsDomain {
  std::string name;
  std::string env;
  std::string transport;
  std::string description;
  std::vector<DomainOperation> operations;
};

struct DomainOperationsCatalog {
  std::string schema;
  std::vector<DomainOperationsDomain> domains;
};

inline const DomainOperation* find_domain_operation(const DomainOperationsCatalog& catalog, const std::string& domain_name,
                                                    const std::string& operation_name) {
  std::string domain_key = service_key(domain_name);
  for (const auto& domain : catalog.domains) {
    if (service_key(domain.name) != domain_key) {
      continue;
    }
    for (const auto& operation : domain.operations) {
      if (operation.name == operation_name) {
        return &operation;
      }
    }
    return nullptr;
  }
  return nullptr;
}

inline ServiceCatalog default_service_catalog() {
  return ServiceCatalog{{{"auth", "TWO_FINANCE_AUTH_URL"},
                         {"network", "TWO_FINANCE_NETWORK_URL"},
                         {"analytics", "TWO_FINANCE_ANALYTICS_URL"},
                         {"orchestrator", "TWO_FINANCE_ORCHESTRATOR_URL"},
                         {"mcp", "TWO_FINANCE_MCP_URL"},
                         {"planner", "TWO_FINANCE_MCP_URL"},
                         {"tradingcontrol", "TWO_FINANCE_TRADING_CONTROL_URL"},
                         {"matchengine", "TWO_FINANCE_MATCHENGINE_WS_URL"},
                         {"keystore", "TWO_FINANCE_KEYSTORE_URL"},
                         {"hummingbot", "TWO_FINANCE_HUMMINGBOT_URL"},
                         {"wise", "TWO_FINANCE_WISE_URL"},
                         {"airwallex", "TWO_FINANCE_AIRWALLEX_URL"}}};
}

inline std::map<std::string, std::string> service_urls(const SdkConfig& config) {
  std::map<std::string, std::string> urls;
  for (const auto& service : default_service_catalog().services) {
    std::string url = service_url(config, service.name);
    if (!url.empty()) {
      urls[service.name] = url;
    }
  }
  return urls;
}

inline std::vector<ConfiguredServiceEntry> configured_services(const SdkConfig& config) {
  std::vector<ConfiguredServiceEntry> services;
  for (const auto& service : default_service_catalog().services) {
    std::string url = service_url(config, service.name);
    if (!url.empty()) {
      services.push_back(ConfiguredServiceEntry{service.name, service.env, url});
    }
  }
  return services;
}

inline std::vector<ServiceCatalogEntry> missing_service_urls(const SdkConfig& config) {
  std::vector<ServiceCatalogEntry> services;
  for (const auto& service : default_service_catalog().services) {
    if (service_url(config, service.name).empty()) {
      services.push_back(service);
    }
  }
  return services;
}

class ServiceError : public std::runtime_error {
 public:
  ServiceError(std::string method, std::string url, int status_code, std::string body)
      : std::runtime_error("2finance: " + method + " " + url + " returned " + std::to_string(status_code) + ": " + body),
        method_(std::move(method)),
        url_(std::move(url)),
        status_code_(status_code),
        body_(std::move(body)) {}

  const std::string& method() const { return method_; }
  const std::string& url() const { return url_; }
  int status_code() const { return status_code_; }
  const std::string& body() const { return body_; }

 private:
  std::string method_;
  std::string url_;
  int status_code_;
  std::string body_;
};

using TokenSource = std::function<std::string()>;
using Transport = std::function<HttpResponse(const HttpRequest&)>;

inline TokenSource static_token_source(std::string token) {
  return [token = std::move(token)] { return token; };
}

inline std::string trim_copy(std::string value) {
  while (!value.empty() && (value.front() == ' ' || value.front() == '\n' || value.front() == '\r' || value.front() == '\t')) {
    value.erase(value.begin());
  }
  while (!value.empty() && (value.back() == ' ' || value.back() == '\n' || value.back() == '\r' || value.back() == '\t')) {
    value.pop_back();
  }
  return value;
}

inline std::string bearer_authorization(const std::string& token) {
  std::string trimmed = trim_copy(token);
  if (trimmed.empty()) {
    return "";
  }
  if (trimmed.size() >= 7) {
    std::string prefix = trimmed.substr(0, 7);
    for (char& c : prefix) {
      if (c >= 'A' && c <= 'Z') {
        c = static_cast<char>(c - 'A' + 'a');
      }
    }
    if (prefix == "bearer ") {
      return trimmed;
    }
  }
  return "Bearer " + trimmed;
}

inline std::string trim_trailing_slash(std::string value) {
  while (!value.empty() && value.back() == '/') {
    value.pop_back();
  }
  return value;
}

inline std::string join_url(const std::string& base_url, const std::string& path) {
  if (path.rfind("http://", 0) == 0 || path.rfind("https://", 0) == 0) {
    return path;
  }
  std::string base = trim_trailing_slash(base_url);
  if (base.empty()) {
    throw std::invalid_argument("2finance: base_url is required");
  }
  std::string normalized_path = path;
  while (!normalized_path.empty() && normalized_path.front() == '/') {
    normalized_path.erase(normalized_path.begin());
  }
  return base + "/" + normalized_path;
}

inline std::string json_escape(const std::string& value) {
  std::string escaped;
  escaped.reserve(value.size());
  for (char c : value) {
    if (c == '\\' || c == '"') {
      escaped.push_back('\\');
    }
    escaped.push_back(c);
  }
  return escaped;
}

inline std::string percent_encode(const std::string& value) {
  std::ostringstream encoded;
  encoded << std::uppercase << std::hex;
  for (unsigned char c : value) {
    if ((c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-' || c == '_' ||
        c == '.' || c == '~') {
      encoded << static_cast<char>(c);
    } else {
      encoded << '%' << std::setw(2) << std::setfill('0') << static_cast<int>(c);
    }
  }
  return encoded.str();
}

inline ResolvedOperation resolve_domain_operation(const DomainOperation& operation,
                                                  const std::map<std::string, std::string>& path_params = {},
                                                  const std::map<std::string, std::string>& query = {}) {
  std::string path = operation.path;
  for (const auto& name : operation.path_params) {
    auto value = path_params.find(name);
    if (value == path_params.end()) {
      throw std::invalid_argument("2finance: missing operation path parameter " + name);
    }
    std::string placeholder = "{" + name + "}";
    std::string encoded = percent_encode(value->second);
    std::size_t pos = 0;
    while ((pos = path.find(placeholder, pos)) != std::string::npos) {
      path.replace(pos, placeholder.size(), encoded);
      pos += encoded.size();
    }
  }

  bool first = path.find('?') == std::string::npos;
  for (const auto& name : operation.query) {
    auto value = query.find(name);
    if (value == query.end()) {
      continue;
    }
    path += first ? '?' : '&';
    path += percent_encode(name);
    path += '=';
    path += percent_encode(value->second);
    first = false;
  }

  std::string method = operation.method;
  while (!method.empty() && (method.front() == ' ' || method.front() == '\n' || method.front() == '\r' || method.front() == '\t')) {
    method.erase(method.begin());
  }
  while (!method.empty() && (method.back() == ' ' || method.back() == '\n' || method.back() == '\r' || method.back() == '\t')) {
    method.pop_back();
  }
  for (char& c : method) {
    if (c >= 'a' && c <= 'z') {
      c = static_cast<char>(c - 'a' + 'A');
    }
  }
  return ResolvedOperation{method, path};
}

inline ResolvedOperation resolve_catalog_operation(const DomainOperationsCatalog& catalog, const std::string& domain_name,
                                                   const std::string& operation_name,
                                                   const std::map<std::string, std::string>& path_params = {},
                                                   const std::map<std::string, std::string>& query = {}) {
  const DomainOperation* operation = find_domain_operation(catalog, domain_name, operation_name);
  if (operation == nullptr) {
    throw std::invalid_argument("2finance: unknown operation " + domain_name + "." + operation_name);
  }
  return resolve_domain_operation(*operation, path_params, query);
}

inline std::string append_query(std::string url, const RequestOptions& options) {
  bool has_query = url.find('?') != std::string::npos;
  bool first = !has_query;
  for (const auto& entry : options.query) {
    url += first ? '?' : '&';
    url += percent_encode(entry.first);
    url += '=';
    url += percent_encode(entry.second);
    first = false;
  }
  if (options.page > 0) {
    url += first ? '?' : '&';
    url += "page=" + percent_encode(std::to_string(options.page));
    first = false;
  }
  if (options.limit > 0) {
    url += first ? '?' : '&';
    url += "limit=" + percent_encode(std::to_string(options.limit));
  }
  return url;
}

inline bool is_retryable_status(int status_code) {
  return status_code == 429 || status_code >= 500;
}

class ServiceClient {
 public:
  ServiceClient(std::string base_url, Transport transport, TokenSource token_source = {})
      : base_url_(std::move(base_url)), transport_(std::move(transport)), token_source_(std::move(token_source)) {}

  std::string url(const std::string& path) const {
    return join_url(base_url_, path);
  }

  HttpResponse request(const std::string& method, const std::string& path, const std::string& body = "",
                       const RequestOptions& options = {}) const {
    if (!transport_) {
      throw std::invalid_argument("2finance: transport is required");
    }
    HttpRequest request;
    request.method = method;
    request.url = append_query(url(path), options);
    request.body = body;
    request.timeout_ms = options.timeout_ms;
    request.headers["Accept"] = "application/json";
    if (!body.empty()) {
      request.headers["Content-Type"] = "application/json";
    }
    if (token_source_) {
      std::string authorization = bearer_authorization(token_source_());
      if (!authorization.empty()) {
        request.headers["Authorization"] = authorization;
      }
    }
    for (const auto& header : options.headers) {
      request.headers[header.first] = header.second;
    }
    std::string idempotency_key = trim_copy(options.idempotency_key);
    if (!idempotency_key.empty()) {
      request.headers["Idempotency-Key"] = idempotency_key;
    }
    int attempts = options.max_retries < 0 ? 1 : options.max_retries + 1;
    HttpResponse response;
    for (int attempt = 0; attempt < attempts; ++attempt) {
      response = transport_(request);
      if (response.status_code >= 200 && response.status_code < 300) {
        return response;
      }
      if (attempt + 1 >= attempts || !is_retryable_status(response.status_code)) {
        throw ServiceError(method, request.url, response.status_code, response.body);
      }
    }
    throw ServiceError(method, request.url, response.status_code, response.body);
  }

  HttpResponse request_operation(const ResolvedOperation& operation, const std::string& body = "",
                                 const RequestOptions& options = {}) const {
    return request(operation.method, operation.path, body, options);
  }

  HttpResponse request_catalog_operation(const DomainOperationsCatalog& catalog, const std::string& domain_name,
                                         const std::string& operation_name,
                                         const std::map<std::string, std::string>& path_params = {},
                                         const std::map<std::string, std::string>& query = {},
                                         const std::string& body = "", const RequestOptions& options = {}) const {
    return request_operation(resolve_catalog_operation(catalog, domain_name, operation_name, path_params, query), body, options);
  }

  HttpResponse get(const std::string& path, const RequestOptions& options = {}) const {
    return request("GET", path, "", options);
  }

  HttpResponse post(const std::string& path, const std::string& body = "", const RequestOptions& options = {}) const {
    return request("POST", path, body, options);
  }

  HttpResponse put(const std::string& path, const std::string& body = "", const RequestOptions& options = {}) const {
    return request("PUT", path, body, options);
  }

  HttpResponse del(const std::string& path, const RequestOptions& options = {}) const {
    return request("DELETE", path, "", options);
  }

 protected:
  std::string base_url_;
  Transport transport_;
  TokenSource token_source_;
};

class AuthClient : public ServiceClient {
 public:
  AuthClient(std::string base_url, Transport transport, TokenSource token_source, std::string realm,
             std::string client_id, std::string phone_client_id)
      : ServiceClient(std::move(base_url), std::move(transport), std::move(token_source)),
        realm_(std::move(realm)),
        client_id_(std::move(client_id)),
        phone_client_id_(std::move(phone_client_id)) {}

  HttpResponse login(const std::string& body) const {
    return post(auth_path(client_id_, "/login"), body);
  }

  HttpResponse refresh_token(const std::string& body) const {
    return post(auth_path(client_id_, "/refresh"), body);
  }

  HttpResponse phone_login(const std::string& body) const {
    return post(auth_path(phone_client_id_, "/phone/sms/login"), body);
  }

  HttpResponse jwks() const {
    return get(oidc_path("/protocol/openid-connect/certs"));
  }

  HttpResponse validate_token(const std::string& token) const {
    return post(oidc_path("/protocol/openid-connect/token/introspect"),
                "{\"token\":\"" + json_escape(token) + "\"}");
  }

 private:
  std::string auth_path(const std::string& client_id, std::string endpoint) const {
    while (!endpoint.empty() && endpoint.front() == '/') {
      endpoint.erase(endpoint.begin());
    }
    return "/v1/2finance-authenticator/" + realm_ + "/" + client_id + "/" + endpoint;
  }

  std::string oidc_path(std::string endpoint) const {
    while (!endpoint.empty() && endpoint.front() == '/') {
      endpoint.erase(endpoint.begin());
    }
    return "/realms/" + realm_ + "/" + endpoint;
  }

  std::string realm_;
  std::string client_id_;
  std::string phone_client_id_;
};

class AnalyticsClient : public ServiceClient {
 public:
  using ServiceClient::ServiceClient;

  HttpResponse indicators() const {
    return get("/analytics/indicators");
  }

  HttpResponse calculate_technical_analysis(const std::string& body) const {
    return post("/analytics/technical-analysis:calculate", body);
  }

  HttpResponse upsert_candles(const std::string& body) const {
    return post("/analytics/candles:upsert", body);
  }

  HttpResponse optimize_portfolio(const std::string& body) const {
    return post("/portfolio-manager/optimizer", body);
  }

  HttpResponse rankings() const {
    return get("/portfolio-manager/rankings");
  }

  HttpResponse balances(const std::string& account_id) const {
    return get("/portfolio-manager/balances/" + percent_encode(account_id));
  }

  HttpResponse black_scholes(const std::string& query = "") const {
    return get("/risk-manager/blackscholes" + (query.empty() ? "" : "?" + query));
  }

  HttpResponse staking() const {
    return get("/staking");
  }
};

class MCPClient : public ServiceClient {
 public:
  using ServiceClient::ServiceClient;

  HttpResponse call(const std::string& method, const std::string& params_json = "null") const {
    std::string params = params_json.empty() ? "null" : params_json;
    return post("/mcp", "{\"jsonrpc\":\"2.0\",\"id\":" + std::to_string(next_id_++) + ",\"method\":\"" +
                            json_escape(method) + "\",\"params\":" + params + "}");
  }

  HttpResponse list_tools() const {
    return call("tools/list");
  }

  HttpResponse list_prompts() const {
    return call("prompts/list");
  }

  HttpResponse list_resources() const {
    return call("resources/list");
  }

  HttpResponse read_resource(const std::string& uri) const {
    return call("resources/read", "{\"uri\":\"" + json_escape(uri) + "\"}");
  }

  HttpResponse call_tool(const std::string& name, const std::string& arguments_json = "{}") const {
    std::string args = arguments_json.empty() ? "{}" : arguments_json;
    return call("tools/call", "{\"name\":\"" + json_escape(name) + "\",\"arguments\":" + args + "}");
  }

  HttpResponse get_prompt(const std::string& name, const std::string& arguments_json = "{}") const {
    std::string args = arguments_json.empty() ? "{}" : arguments_json;
    return call("prompts/get", "{\"name\":\"" + json_escape(name) + "\",\"arguments\":" + args + "}");
  }

  HttpResponse conversation_plan(const std::string& arguments_json) const {
    return call_tool("finance_assistant.conversation.plan", arguments_json);
  }

 private:
  mutable long long next_id_ = 1;
};

class OrchestratorClient : public ServiceClient {
 public:
  using ServiceClient::ServiceClient;

  HttpResponse catalog() const {
    return get("/v1/mcphost/catalog/packages");
  }

  HttpResponse create_session(const std::string& body) const {
    return post("/v1/mcphost/sessions", body);
  }

  HttpResponse send_message(const std::string& body) const {
    return post("/v1/mcphost/messages", body);
  }

  HttpResponse tools() const {
    return get("/v1/mcphost/tools");
  }

  HttpResponse prompts() const {
    return get("/v1/mcphost/prompts");
  }

  HttpResponse resources() const {
    return get("/v1/mcphost/resources");
  }

  HttpResponse providers() const {
    return get("/v1/mcphost/providers");
  }

  HttpResponse approvals() const {
    return get("/v1/mcphost/approvals");
  }

  HttpResponse observability() const {
    return get("/v1/mcphost/observability");
  }

  HttpResponse delete_session(const std::string& session_id) const {
    return del("/v1/mcphost/sessions/" + percent_encode(session_id));
  }
};

class NetworkClient : public ServiceClient {
 public:
  using ServiceClient::ServiceClient;

  HttpResponse virtual_machine() const {
    return post(
        "/v2/2finance-network/query",
        "{\"method\":\"get_blocks\",\"params\":{\"page\":1,\"limit\":1}}");
  }

  HttpResponse market_candles(const std::string& market, const std::string& query = "") const {
    return get("/v2/2finance-network/markets/" + percent_encode(market) + "/candles" +
               (query.empty() ? "" : "?" + query));
  }

  HttpResponse products(const std::string& product_type) const {
    return get("/v2/2finance-network/products/" + percent_encode(product_type));
  }

  HttpResponse create_product(const std::string& product_type, const std::string& body) const {
    return post("/v2/2finance-network/products/" + percent_encode(product_type), body);
  }

  HttpResponse bonds() const {
    return products("bonds");
  }

  HttpResponse create_bond(const std::string& body) const {
    return create_product("bonds", body);
  }

  HttpResponse loans() const {
    return products("loans");
  }

  HttpResponse create_loan(const std::string& body) const {
    return create_product("loans", body);
  }

  HttpResponse swaps() const {
    return products("swaps");
  }

  HttpResponse create_swap(const std::string& body) const {
    return create_product("swaps", body);
  }

  HttpResponse staking_products() const {
    return products("staking");
  }

  HttpResponse create_staking_product(const std::string& body) const {
    return create_product("staking", body);
  }

  HttpResponse synthetic_assets() const {
    return products("synthetic-assets");
  }

  HttpResponse create_synthetic_asset(const std::string& body) const {
    return create_product("synthetic-assets", body);
  }

  HttpResponse liquidity_pools() const {
    return products("liquidity-pools");
  }

  HttpResponse create_liquidity_pool(const std::string& body) const {
    return create_product("liquidity-pools", body);
  }
};

class TradingControlClient : public ServiceClient {
 public:
  using ServiceClient::ServiceClient;

  HttpResponse robots() const {
    return get("/robots");
  }

  HttpResponse create_robot(const std::string& body) const {
    return post("/robots", body);
  }

  HttpResponse robot(const std::string& id) const {
    return get("/robots/" + percent_encode(id));
  }

  HttpResponse start_robot(const std::string& id) const {
    return post("/robots/" + percent_encode(id) + ":start");
  }

  HttpResponse pause_robot(const std::string& id) const {
    return post("/robots/" + percent_encode(id) + ":pause");
  }

  HttpResponse resume_robot(const std::string& id) const {
    return post("/robots/" + percent_encode(id) + ":resume");
  }

  HttpResponse stop_robot(const std::string& id) const {
    return post("/robots/" + percent_encode(id) + ":stop");
  }

  HttpResponse risk_policy(const std::string& id) const {
    return get("/robots/" + percent_encode(id) + "/risk-policy");
  }

  HttpResponse set_risk_policy(const std::string& id, const std::string& body) const {
    return put("/robots/" + percent_encode(id) + "/risk-policy", body);
  }

  HttpResponse risk_view(const std::string& id) const {
    return get("/risk-view/" + percent_encode(id));
  }

  HttpResponse strategies() const {
    return get("/strategies");
  }

  HttpResponse create_strategy(const std::string& body) const {
    return post("/strategies", body);
  }

  HttpResponse directives() const {
    return get("/directives");
  }

  HttpResponse create_directive(const std::string& body) const {
    return post("/directives", body);
  }

  HttpResponse audit() const {
    return get("/audit");
  }

  HttpResponse activity() const {
    return get("/activity");
  }

  HttpResponse mcp_tools() const {
    return get("/mcp/tools");
  }
};

class KeyStoreClient : public ServiceClient {
 public:
  using ServiceClient::ServiceClient;

  HttpResponse health() const {
    return get("/healthz");
  }

  HttpResponse readiness() const {
    return get("/readyz");
  }

  HttpResponse start_keygen(const std::string& body) const {
    return post("/keystore/keygen/start", body);
  }

  HttpResponse keygen_signature(const std::string& body) const {
    return post("/keystore/keygen/signature", body);
  }

  HttpResponse start_signing(const std::string& body) const {
    return post("/keystore/signing/start", body);
  }

  HttpResponse signing_signature(const std::string& body) const {
    return post("/keystore/signing/signature", body);
  }

  HttpResponse start_resharing(const std::string& body) const {
    return post("/keystore/resharing/start", body);
  }

  HttpResponse keys(const std::string& user_public_key) const {
    return get("/keystore/keys/" + percent_encode(user_public_key));
  }

  HttpResponse signatures(const std::string& user_public_key) const {
    return get("/keystore/signatures/" + percent_encode(user_public_key));
  }

  HttpResponse metrics() const {
    return get("/keystore/tss/metrics");
  }
};

class HummingbotClient : public ServiceClient {
 public:
  using ServiceClient::ServiceClient;

  HttpResponse assets() const {
    return get("/api/v1/assets");
  }

  HttpResponse symbols() const {
    return get("/api/v1/symbols");
  }

  HttpResponse balances() const {
    return get("/api/v1/balances");
  }

  HttpResponse connector_config(const std::string& body) const {
    return post("/api/v1/connectors/2finance/config", body);
  }
};

class ProviderClient : public ServiceClient {
 public:
  using ServiceClient::ServiceClient;
};

class WiseClient : public ProviderClient {
 public:
  using ProviderClient::ProviderClient;

  HttpResponse profiles() const {
    return get("/v1/profiles");
  }

  HttpResponse profile(const std::string& profile_id) const {
    return get("/v1/profiles/" + percent_encode(profile_id));
  }

  HttpResponse create_quote(const std::string& profile_id, const std::string& body) const {
    return post("/v3/profiles/" + percent_encode(profile_id) + "/quotes", body);
  }

  HttpResponse create_transfer(const std::string& body) const {
    return post("/v1/transfers", body);
  }
};

class AirwallexClient : public ProviderClient {
 public:
  using ProviderClient::ProviderClient;

  HttpResponse accounts() const {
    return get("/api/v1/accounts");
  }

  HttpResponse payments() const {
    return get("/api/v1/payments");
  }

  HttpResponse create_payment(const std::string& body) const {
    return post("/api/v1/payments", body);
  }

  HttpResponse beneficiaries() const {
    return get("/api/v1/beneficiaries");
  }

  HttpResponse create_beneficiary(const std::string& body) const {
    return post("/api/v1/beneficiaries", body);
  }
};

class MatchEngineClient {
 public:
  using MessageSender = std::function<HttpResponse(const std::string&)>;

  explicit MatchEngineClient(std::string websocket_url) : websocket_url_(std::move(websocket_url)) {}

  const std::string& websocket_url() const {
    return websocket_url_;
  }

  std::string order_command(const std::string& command_json) const {
    if (command_json.empty() || command_json == "{}") {
      return "{\"schema\":\"matchengine.order_command.v2\",\"message_type\":\"ORDER\",\"operation\":\"ADD\"}";
    }
    if (command_json.front() == '{') {
      return "{\"schema\":\"matchengine.order_command.v2\",\"message_type\":\"ORDER\",\"operation\":\"ADD\"," + command_json.substr(1);
    }
    return "{\"schema\":\"matchengine.order_command.v2\",\"message_type\":\"ORDER\",\"operation\":\"ADD\",\"payload\":" + command_json + "}";
  }

  std::string market_data_subscribe(const std::string& request_json) const {
    if (request_json.empty() || request_json == "{}") {
      return "{\"schema\":\"matchengine.market_data_subscribe.v2\"}";
    }
    if (request_json.front() == '{') {
      return "{\"schema\":\"matchengine.market_data_subscribe.v2\"," + request_json.substr(1);
    }
    return "{\"schema\":\"matchengine.market_data_subscribe.v2\",\"payload\":" + request_json + "}";
  }

  HttpResponse send_order(MessageSender sender, const std::string& command_json) const {
    return sender(order_command(command_json));
  }

  HttpResponse subscribe_market_data(MessageSender sender, const std::string& request_json) const {
    return sender(market_data_subscribe(request_json));
  }

 private:
  std::string websocket_url_;
};

class PlannerClient {
 public:
  PlannerClient(MCPClient* mcp, OrchestratorClient* orchestrator, AnalyticsClient* analytics,
                TradingControlClient* trading_control)
      : mcp_(mcp), orchestrator_(orchestrator), analytics_(analytics), trading_control_(trading_control) {}

  HttpResponse conversation_plan(const std::string& arguments_json) const {
    if (mcp_ == nullptr) {
      throw std::invalid_argument("2finance planner: mcp client is required");
    }
    return mcp_->conversation_plan(arguments_json);
  }

  HttpResponse orchestrated_plan(const std::string& request_json) const {
    if (orchestrator_ == nullptr) {
      throw std::invalid_argument("2finance planner: orchestrator client is required");
    }
    return orchestrator_->send_message(request_json);
  }

  HttpResponse operational_plan(const std::string& request_json) const {
    return orchestrated_plan(request_json);
  }

  HttpResponse trading_plan(const std::string& request_json, bool use_analytics = false, bool use_trading = false) const {
    std::string context = "{";
    bool has_context = false;
    if (use_trading && trading_control_ != nullptr) {
      try {
        HttpResponse robots = trading_control_->robots();
        append_context_value(context, has_context, "trading_robots", robots.body);
      } catch (...) {
        // Best-effort enrichment keeps planning usable when trading is unavailable.
      }
    }
    if (use_analytics && analytics_ != nullptr) {
      try {
        HttpResponse indicators = analytics_->indicators();
        append_context_value(context, has_context, "analytics_indicators", indicators.body);
      } catch (...) {
        // Best-effort enrichment keeps planning usable when analytics is unavailable.
      }
    }
    context += "}";
    return conversation_plan("{\"request\":" + json_value_or_string(request_json) + ",\"context\":" + context + "}");
  }

 private:
  static void append_context_value(std::string& context, bool& has_context, const std::string& key,
                                   const std::string& value) {
    if (has_context) {
      context += ",";
    }
    context += "\"" + json_escape(key) + "\":" + json_value_or_string(value);
    has_context = true;
  }

  static std::string json_value_or_string(const std::string& value) {
    std::string trimmed = twofinance::trim_copy(value);
    if (trimmed.empty()) {
      return "null";
    }
    char first = trimmed.front();
    if (first == '{' || first == '[' || first == '"' || first == '-' || (first >= '0' && first <= '9') ||
        trimmed == "true" || trimmed == "false" || trimmed == "null") {
      return trimmed;
    }
    return "\"" + json_escape(trimmed) + "\"";
  }

  MCPClient* mcp_;
  OrchestratorClient* orchestrator_;
  AnalyticsClient* analytics_;
  TradingControlClient* trading_control_;
};

class SdkClient {
 public:
  SdkClient(SdkConfig cfg, Transport transport, TokenSource token_source = {})
      : config(cfg),
        auth(cfg.auth_url, transport, token_source, cfg.auth_realm, cfg.auth_client_id, cfg.auth_phone_client_id),
        network(cfg.network_url, transport, token_source),
        analytics(cfg.analytics_url, transport, token_source),
        orchestrator(cfg.orchestrator_url, transport, token_source),
        mcp(cfg.mcp_url, transport, token_source),
        trading_control(cfg.trading_control_url, transport, token_source),
        match_engine(cfg.matchengine_ws_url),
        keystore(cfg.keystore_url, transport, token_source),
        hummingbot(cfg.hummingbot_url, transport, token_source),
        wise(cfg.wise_url, transport, token_source),
        airwallex(cfg.airwallex_url, transport, token_source),
        planner(&mcp, &orchestrator, &analytics, &trading_control) {}

  SdkConfig config;
  AuthClient auth;
  NetworkClient network;
  AnalyticsClient analytics;
  OrchestratorClient orchestrator;
  MCPClient mcp;
  TradingControlClient trading_control;
  MatchEngineClient match_engine;
  KeyStoreClient keystore;
  HummingbotClient hummingbot;
  WiseClient wise;
  AirwallexClient airwallex;
  PlannerClient planner;
};

}  // namespace twofinance
