#include "twofinance/sdk_client.hpp"

#include <cassert>
#include <fstream>
#include <sstream>
#include <string>

std::string read_file(const std::string& path) {
  std::ifstream input(path);
  assert(input.good());
  std::ostringstream buffer;
  buffer << input.rdbuf();
  return buffer.str();
}

std::string json_string(const std::string& json, const std::string& key) {
  std::string quoted_key = "\"" + key + "\"";
  std::size_t key_pos = json.find(quoted_key);
  assert(key_pos != std::string::npos);
  std::size_t colon_pos = json.find(':', key_pos + quoted_key.size());
  assert(colon_pos != std::string::npos);
  std::size_t value_start = json.find('"', colon_pos + 1);
  assert(value_start != std::string::npos);
  std::size_t value_end = json.find('"', value_start + 1);
  assert(value_end != std::string::npos);
  return json.substr(value_start + 1, value_end - value_start - 1);
}

int json_int(const std::string& json, const std::string& key) {
  std::string quoted_key = "\"" + key + "\"";
  std::size_t key_pos = json.find(quoted_key);
  assert(key_pos != std::string::npos);
  std::size_t colon_pos = json.find(':', key_pos + quoted_key.size());
  assert(colon_pos != std::string::npos);
  std::size_t value_start = json.find_first_of("0123456789", colon_pos + 1);
  assert(value_start != std::string::npos);
  std::size_t value_end = json.find_first_not_of("0123456789", value_start);
  return std::stoi(json.substr(value_start, value_end - value_start));
}

int main() {
  assert(std::string(twofinance::SDK_NAME) == "2finance-sdk-client");
  assert(std::string(twofinance::SDK_VERSION) == "0.1.0");
  assert(twofinance::default_service_catalog().services.size() == 12);
  assert(twofinance::default_service_catalog().services.front().env == "TWO_FINANCE_AUTH_URL");
  twofinance::SdkConfig service_config;
  service_config.analytics_url = "https://analytics.example";
  service_config.matchengine_ws_url = "wss://matchengine.example/ws";
  assert(twofinance::service_url(service_config, "analytics") == "https://analytics.example");
  assert(twofinance::service_url(service_config, "match_engine") == "wss://matchengine.example/ws");
  assert(twofinance::service_urls(service_config).at("matchengine") == "wss://matchengine.example/ws");
  assert(twofinance::configured_services(service_config).at(1).name == "matchengine");
  assert(twofinance::missing_service_urls(service_config).at(0).name == "auth");
  twofinance::SdkConfig env_config =
      twofinance::config_from_env({{"TWO_FINANCE_AUTH_URL", "https://auth.example"},
                                   {"TWO_FINANCE_ANALYTICS_URL", "https://analytics.example"},
                                   {"TWO_FINANCE_MATCHENGINE_WS_URL", "wss://matchengine.example/ws"}});
  assert(env_config.auth_url == "https://auth.example");
  assert(env_config.analytics_url == "https://analytics.example");
  assert(env_config.matchengine_ws_url == "wss://matchengine.example/ws");
  assert(env_config.auth_realm == "2finance");
  assert(twofinance::config_from_environment().auth_realm == "2finance");

  std::string fixture = read_file("../contracts/examples/request-options.json");
  std::string domain_operations = read_file("../contracts/examples/domain-operations.json");
  std::string error_fixture = read_file("../contracts/examples/error.json");
  std::string pagination_fixture = read_file("../contracts/examples/pagination.json");
  std::string idempotency_fixture = read_file("../contracts/examples/idempotency.json");
  std::string fixture_base_url = json_string(fixture, "base_url");
  std::string fixture_path = json_string(fixture, "path");
  std::string fixture_expected_url = json_string(fixture, "url");
  std::string fixture_trace_id = json_string(fixture, "X-Trace-ID");
  std::string fixture_idempotency_key = json_string(fixture, "idempotency_key");
  std::string fixture_symbol = json_string(fixture, "symbol");
  int fixture_page = json_int(fixture, "page");
  int fixture_limit = json_int(fixture, "limit");
  int fixture_timeout_ms = json_int(fixture, "timeout_ms");
  int fixture_max_retries = json_int(fixture, "max_retries");

  assert(domain_operations.find("\"schema\": \"sdk.domain_operations.v2\"") != std::string::npos);
  assert(domain_operations.find("\"name\": \"planner\"") != std::string::npos);
  assert(domain_operations.find("\"name\": \"trading_plan\"") != std::string::npos);
  assert(domain_operations.find("\"path\": \"/portfolio-manager/balances/{account_id}\"") != std::string::npos);
  assert(domain_operations.find("\"path_params\": [\"account_id\"]") != std::string::npos);
  assert(error_fixture.find("\"error\": \"rate_limited\"") != std::string::npos);
  assert(error_fixture.find("\"code\": \"HTTP_429\"") != std::string::npos);
  assert(pagination_fixture.find("\"next_cursor\": \"cursor-next\"") != std::string::npos);
  assert(idempotency_fixture.find("\"idempotency_key\": \"idem-001\"") != std::string::npos);

  assert(twofinance::bearer_authorization("abc") == "Bearer abc");
  assert(twofinance::bearer_authorization(" abc ") == "Bearer abc");
  assert(twofinance::bearer_authorization(" Bearer abc ") == "Bearer abc");
  assert(twofinance::bearer_authorization(" ") == "");
  assert(twofinance::join_url("https://api.example/", "/healthz") == "https://api.example/healthz");
  twofinance::SDKErrorPayload sdk_error{
      "rate_limited", "Too many requests", "HTTP_429", {{"request_id", "req_2finance_001"}}};
  twofinance::PaginationResponse pagination{25, "cursor-current", "cursor-next"};
  twofinance::IdempotencyRecord idempotency{"idem-001", "matchengine.order_command", "client_order_id", "req_2finance_001"};
  twofinance::ServiceCatalog catalog{{{"auth", "TWO_FINANCE_AUTH_URL"}}};
  twofinance::DomainOperationsCatalog operations_catalog{
      "sdk.domain_operations.v2",
      {{"auth",
        "TWO_FINANCE_AUTH_URL",
        "http",
        "User auth and client credentials token flows.",
        {{"login",
          "POST",
          "/v1/2finance-authenticator/{realm}/{client_id}/login",
          {"realm", "client_id"},
          {},
          "auth.login.request.v1",
          "auth.token.response.v1",
          ""}}}}};
  assert(sdk_error.code == "HTTP_429");
  assert(sdk_error.details.at("request_id") == "req_2finance_001");
  assert(pagination.limit == 25);
  assert(pagination.next_cursor == "cursor-next");
  assert(idempotency.idempotency_key == "idem-001");
  assert(catalog.services.front().name == "auth");
  assert(operations_catalog.domains.front().operations.front().request_schema == "auth.login.request.v1");
  assert(twofinance::find_domain_operation(operations_catalog, "auth", "login") != nullptr);
  auto resolved_login = twofinance::resolve_domain_operation(
      operations_catalog.domains.front().operations.front(), {{"realm", "2finance"}, {"client_id", "client/1"}}, {});
  assert(resolved_login.method == "POST");
  assert(resolved_login.path == "/v1/2finance-authenticator/2finance/client%2F1/login");
  auto catalog_resolved_login = twofinance::resolve_catalog_operation(
      operations_catalog, "auth", "login", {{"realm", "2finance"}, {"client_id", "client/1"}}, {});
  assert(catalog_resolved_login.path == resolved_login.path);
  twofinance::DomainOperation risk_operation{"black_scholes",
                                             "get",
                                             "/risk-manager/blackscholes",
                                             {},
                                             {"symbol", "strike", "volatility"},
                                             "",
                                             "",
                                             ""};
  auto resolved_risk = twofinance::resolve_domain_operation(
      risk_operation, {}, {{"symbol", "BTC/USD"}, {"strike", "100000"}, {"ignored", "drop-me"}, {"volatility", "0.5"}});
  assert(resolved_risk.method == "GET");
  assert(resolved_risk.path == "/risk-manager/blackscholes?symbol=BTC%2FUSD&strike=100000&volatility=0.5");
  twofinance::ServiceClient failing_service(
      "https://analytics.example",
      [](const twofinance::HttpRequest&) {
        return twofinance::HttpResponse{429, "rate limited"};
      });
  try {
    failing_service.get("/analytics/indicators");
    assert(false && "expected ServiceError");
  } catch (const twofinance::ServiceError& error) {
    assert(error.method() == "GET");
    assert(error.url() == "https://analytics.example/analytics/indicators");
    assert(error.status_code() == 429);
    assert(error.body() == "rate limited");
  }

  std::string resolved_seen;
  twofinance::ServiceClient resolved_service(
      fixture_base_url,
      [&resolved_seen](const twofinance::HttpRequest& request) {
        resolved_seen = request.method + " " + request.url;
        return twofinance::HttpResponse{200, "{\"ok\":true}"};
      });
  resolved_service.request_operation(twofinance::ResolvedOperation{"GET", "/portfolio-manager/balances/acct%2Fresolved"});
  assert(resolved_seen == "GET https://analytics.example/portfolio-manager/balances/acct%2Fresolved");
  twofinance::DomainOperationsCatalog request_catalog{
      "sdk.domain_operations.v2",
      {{"analytics",
        "TWO_FINANCE_ANALYTICS_URL",
        "http",
        "",
        {{"balances", "GET", "/portfolio-manager/balances/{account_id}", {"account_id"}, {}, "", "", ""}}}}};
  resolved_service.request_catalog_operation(request_catalog, "analytics", "balances", {{"account_id", "acct/1 ok"}});
  assert(resolved_seen == "GET https://analytics.example/portfolio-manager/balances/acct%2F1%20ok");

  twofinance::SdkConfig config;
  config.auth_url = "https://auth.example";
  config.analytics_url = fixture_base_url;
  config.network_url = "https://network.example";
  config.orchestrator_url = "https://orchestrator.example";
  config.mcp_url = "https://mcp.example";
  config.trading_control_url = "https://trading.example";
  config.keystore_url = "https://keys.example";
  config.hummingbot_url = "https://hbot.example";
  config.wise_url = "https://wise.example";
  config.airwallex_url = "https://airwallex.example";
  config.matchengine_ws_url = "wss://matchengine.example/ws";
  bool saw_auth = false;
  bool saw_auth_jwks = false;
  bool saw_auth_validate = false;
  bool saw_plan = false;
  bool saw_trading_plan_context = false;
  bool saw_black_scholes = false;
  bool saw_mcp_tools = false;
  bool saw_operational_plan = false;
  bool saw_orchestrator_delete = false;
  bool saw_trading = false;
  bool saw_trading_pause = false;
  bool saw_trading_resume = false;
  bool saw_trading_stop = false;
  bool saw_trading_set_risk = false;
  bool saw_trading_risk_view = false;
  bool saw_trading_strategies = false;
  bool saw_trading_create_strategy = false;
  bool saw_trading_directives = false;
  bool saw_trading_create_directive = false;
  bool saw_trading_audit = false;
  bool saw_trading_activity = false;
  bool saw_trading_mcp_tools = false;
  bool saw_keystore = false;
  bool saw_keystore_health = false;
  bool saw_keystore_readiness = false;
  bool saw_keystore_metrics = false;
  bool saw_network_vm = false;
  bool saw_network_bonds = false;
  bool saw_network_create_bond = false;
  bool saw_network_loans = false;
  bool saw_network_create_loan = false;
  bool saw_network_swaps = false;
  bool saw_network_create_swap = false;
  bool saw_network_staking = false;
  bool saw_network_create_staking = false;
  bool saw_network_synthetic_assets = false;
  bool saw_network_create_synthetic_asset = false;
  bool saw_network_liquidity_pools = false;
  bool saw_network_create_liquidity_pool = false;
  bool saw_hummingbot_assets = false;
  bool saw_hummingbot_symbols = false;
  bool saw_hummingbot = false;
  bool saw_hummingbot_config = false;
  bool saw_wise = false;
  bool saw_wise_profile = false;
  bool saw_wise_quote = false;
  bool saw_wise_transfer = false;
  bool saw_airwallex = false;
  bool saw_airwallex_accounts = false;
  bool saw_airwallex_payments = false;
  bool saw_airwallex_beneficiaries = false;
  bool saw_airwallex_create_beneficiary = false;
  bool saw_idempotency = false;
  int candle_attempts = 0;
  twofinance::Transport transport = [&](const twofinance::HttpRequest& request) {
    if (request.url == "https://auth.example/realms/2finance/protocol/openid-connect/certs") {
      saw_auth_jwks = request.method == "GET";
    } else if (request.url == "https://auth.example/realms/2finance/protocol/openid-connect/token/introspect") {
      saw_auth_validate = request.method == "POST" && request.body.find("token-1") != std::string::npos;
    } else if (request.url == "https://analytics.example/analytics/indicators") {
      saw_auth = request.headers.at("Authorization") == "Bearer token-123";
      return twofinance::HttpResponse{200, "{\"indicators\":[\"rsi\"]}"};
    } else if (request.url == "https://analytics.example/risk-manager/blackscholes?symbol=BTC") {
      saw_black_scholes = request.method == "GET";
    } else if (request.url == fixture_expected_url) {
      candle_attempts++;
      saw_idempotency = request.headers.at("Idempotency-Key") == fixture_idempotency_key;
      assert(request.headers.at("X-Trace-ID") == fixture_trace_id);
      assert(request.timeout_ms == fixture_timeout_ms);
      if (candle_attempts == 1) {
        return twofinance::HttpResponse{500, "temporary"};
      }
    } else if (request.url == "https://analytics.example/portfolio-manager/balances/acct%2F1") {
      assert(request.method == "GET");
    } else if (request.url == "https://network.example/v2/2finance-network/markets/BTC%2FUSDT/candles?limit=10") {
      assert(request.method == "GET");
    } else if (request.url == "https://network.example/v2/2finance-network/query") {
      saw_network_vm = request.method == "POST" &&
                       request.body.find("get_blocks") != std::string::npos;
    } else if (request.url == "https://network.example/v2/2finance-network/products/bonds" && request.method == "GET") {
      saw_network_bonds = true;
    } else if (request.url == "https://network.example/v2/2finance-network/products/bonds") {
      saw_network_create_bond = request.method == "POST";
    } else if (request.url == "https://network.example/v2/2finance-network/products/loans" && request.method == "GET") {
      saw_network_loans = true;
    } else if (request.url == "https://network.example/v2/2finance-network/products/loans") {
      saw_network_create_loan = request.method == "POST";
    } else if (request.url == "https://network.example/v2/2finance-network/products/swaps" && request.method == "GET") {
      saw_network_swaps = true;
    } else if (request.url == "https://network.example/v2/2finance-network/products/swaps") {
      saw_network_create_swap = request.method == "POST";
    } else if (request.url == "https://network.example/v2/2finance-network/products/staking" && request.method == "GET") {
      saw_network_staking = true;
    } else if (request.url == "https://network.example/v2/2finance-network/products/staking") {
      saw_network_create_staking = request.method == "POST";
    } else if (request.url == "https://network.example/v2/2finance-network/products/synthetic-assets" && request.method == "GET") {
      saw_network_synthetic_assets = true;
    } else if (request.url == "https://network.example/v2/2finance-network/products/synthetic-assets") {
      saw_network_create_synthetic_asset = request.method == "POST";
    } else if (request.url == "https://network.example/v2/2finance-network/products/liquidity-pools" && request.method == "GET") {
      saw_network_liquidity_pools = true;
    } else if (request.url == "https://network.example/v2/2finance-network/products/liquidity-pools") {
      saw_network_create_liquidity_pool = request.method == "POST";
    } else if (request.url == "https://mcp.example/mcp") {
      saw_plan = saw_plan || request.body.find("finance_assistant.conversation.plan") != std::string::npos;
      saw_mcp_tools = saw_mcp_tools || request.body.find("tools/list") != std::string::npos;
      saw_trading_plan_context =
          saw_trading_plan_context ||
          (request.body.find("trading_robots") != std::string::npos &&
           request.body.find("analytics_indicators") != std::string::npos);
    } else if (request.url == "https://orchestrator.example/v1/mcphost/sessions/session%2F1") {
      saw_orchestrator_delete = request.method == "DELETE";
    } else if (request.url == "https://orchestrator.example/v1/mcphost/messages") {
      saw_operational_plan = request.method == "POST";
    } else if (request.url == "https://trading.example/robots") {
      assert(request.method == "GET");
      return twofinance::HttpResponse{200, "{\"robots\":[{\"id\":\"robot-1\"}]}"};
    } else if (request.url == "https://trading.example/robots/robot%2F1/risk-policy") {
      saw_trading = saw_trading || request.method == "GET";
      saw_trading_set_risk = saw_trading_set_risk || request.method == "PUT";
    } else if (request.url == "https://trading.example/robots/robot%2F1:pause") {
      saw_trading_pause = request.method == "POST";
    } else if (request.url == "https://trading.example/robots/robot%2F1:resume") {
      saw_trading_resume = request.method == "POST";
    } else if (request.url == "https://trading.example/robots/robot%2F1:stop") {
      saw_trading_stop = request.method == "POST";
    } else if (request.url == "https://trading.example/risk-view/robot%2F1") {
      saw_trading_risk_view = request.method == "GET";
    } else if (request.url == "https://trading.example/strategies" && request.method == "GET") {
      saw_trading_strategies = true;
    } else if (request.url == "https://trading.example/strategies") {
      saw_trading_create_strategy = request.method == "POST";
    } else if (request.url == "https://trading.example/directives" && request.method == "GET") {
      saw_trading_directives = true;
    } else if (request.url == "https://trading.example/directives") {
      saw_trading_create_directive = request.method == "POST";
    } else if (request.url == "https://trading.example/audit") {
      saw_trading_audit = request.method == "GET";
    } else if (request.url == "https://trading.example/activity") {
      saw_trading_activity = request.method == "GET";
    } else if (request.url == "https://trading.example/mcp/tools") {
      saw_trading_mcp_tools = request.method == "GET";
    } else if (request.url == "https://keys.example/healthz") {
      saw_keystore_health = request.method == "GET";
    } else if (request.url == "https://keys.example/readyz") {
      saw_keystore_readiness = request.method == "GET";
    } else if (request.url == "https://keys.example/keystore/signing/start") {
      saw_keystore = request.method == "POST";
    } else if (request.url == "https://keys.example/keystore/keys/pub%2F1") {
      assert(request.method == "GET");
    } else if (request.url == "https://keys.example/keystore/signatures/pub%2F1") {
      assert(request.method == "GET");
    } else if (request.url == "https://keys.example/keystore/tss/metrics") {
      saw_keystore_metrics = request.method == "GET";
    } else if (request.url == "https://hbot.example/api/v1/assets") {
      saw_hummingbot_assets = request.method == "GET";
    } else if (request.url == "https://hbot.example/api/v1/symbols") {
      saw_hummingbot_symbols = request.method == "GET";
    } else if (request.url == "https://hbot.example/api/v1/balances") {
      saw_hummingbot = request.method == "GET";
    } else if (request.url == "https://hbot.example/api/v1/connectors/2finance/config") {
      saw_hummingbot_config = request.method == "POST";
    } else if (request.url == "https://wise.example/v1/profiles") {
      saw_wise = request.method == "GET";
    } else if (request.url == "https://wise.example/v1/profiles/profile%2F1") {
      saw_wise_profile = request.method == "GET";
    } else if (request.url == "https://wise.example/v3/profiles/profile%2F1/quotes") {
      saw_wise_quote = request.method == "POST";
    } else if (request.url == "https://wise.example/v1/transfers") {
      saw_wise_transfer = request.method == "POST";
    } else if (request.url == "https://airwallex.example/api/v1/accounts") {
      saw_airwallex_accounts = request.method == "GET";
    } else if (request.url == "https://airwallex.example/api/v1/payments" && request.method == "GET") {
      saw_airwallex_payments = true;
    } else if (request.url == "https://airwallex.example/api/v1/payments") {
      saw_airwallex = request.method == "POST";
    } else if (request.url == "https://airwallex.example/api/v1/beneficiaries" && request.method == "GET") {
      saw_airwallex_beneficiaries = true;
    } else if (request.url == "https://airwallex.example/api/v1/beneficiaries") {
      saw_airwallex_create_beneficiary = request.method == "POST";
    } else {
      assert(false && "unexpected request URL");
    }
    return twofinance::HttpResponse{200, "{\"ok\":true}"};
  };
  twofinance::SdkClient client(config, transport, twofinance::static_token_source(" token-123 "));
  client.auth.jwks();
  client.auth.validate_token("token-1");
  auto response = client.analytics.indicators();
  assert(response.body == "{\"indicators\":[\"rsi\"]}");
  assert(saw_auth);
  twofinance::RequestOptions request_options;
  request_options.headers["X-Trace-ID"] = fixture_trace_id;
  request_options.idempotency_key = " " + fixture_idempotency_key + " ";
  request_options.query["symbol"] = fixture_symbol;
  request_options.timeout_ms = fixture_timeout_ms;
  request_options.max_retries = fixture_max_retries;
  request_options.page = fixture_page;
  request_options.limit = fixture_limit;
  client.analytics.post(fixture_path, "{\"symbol\":\"BTC-USDT\"}", request_options);
  assert(saw_idempotency);
  assert(candle_attempts == 2);
  client.analytics.black_scholes("symbol=BTC");
  client.analytics.balances("acct/1");
  client.network.virtual_machine();
  client.network.market_candles("BTC/USDT", "limit=10");
  client.network.bonds();
  client.network.create_bond("{\"symbol\":\"BOND1\"}");
  client.network.loans();
  client.network.create_loan("{\"loan\":\"ln1\"}");
  client.network.swaps();
  client.network.create_swap("{\"pair\":\"BTC-USDT\"}");
  client.network.staking_products();
  client.network.create_staking_product("{\"asset\":\"TWO\"}");
  client.network.synthetic_assets();
  client.network.create_synthetic_asset("{\"asset\":\"sBTC\"}");
  client.network.liquidity_pools();
  client.network.create_liquidity_pool("{\"pool\":\"BTC-USDT\"}");
  client.mcp.list_tools();
  client.planner.conversation_plan("{\"goal\":\"trade plan\"}");
  client.planner.operational_plan("{\"message\":\"operate\"}");
  client.planner.trading_plan("{\"goal\":\"rebalance BTC\"}", true, true);
  client.orchestrator.delete_session("session/1");
  client.trading_control.risk_policy("robot/1");
  client.trading_control.pause_robot("robot/1");
  client.trading_control.resume_robot("robot/1");
  client.trading_control.stop_robot("robot/1");
  client.trading_control.set_risk_policy("robot/1", "{\"max_drawdown\":\"0.1\"}");
  client.trading_control.risk_view("robot/1");
  client.trading_control.strategies();
  client.trading_control.create_strategy("{\"name\":\"mean-reversion\"}");
  client.trading_control.directives();
  client.trading_control.create_directive("{\"action\":\"rebalance\"}");
  client.trading_control.audit();
  client.trading_control.activity();
  client.trading_control.mcp_tools();
  client.keystore.health();
  client.keystore.readiness();
  client.keystore.start_signing("{\"key\":\"k1\"}");
  client.keystore.keys("pub/1");
  client.keystore.signatures("pub/1");
  client.keystore.metrics();
  client.hummingbot.assets();
  client.hummingbot.symbols();
  client.hummingbot.balances();
  client.hummingbot.connector_config("{\"connector\":\"2finance\"}");
  client.wise.profiles();
  client.wise.profile("profile/1");
  client.wise.create_quote("profile/1", "{\"source\":\"USD\"}");
  client.wise.create_transfer("{\"target\":\"BRL\"}");
  client.airwallex.accounts();
  client.airwallex.payments();
  client.airwallex.create_payment("{\"amount\":10}");
  client.airwallex.beneficiaries();
  client.airwallex.create_beneficiary("{\"name\":\"beneficiary\"}");
  assert(saw_mcp_tools);
  assert(saw_auth_jwks);
  assert(saw_auth_validate);
  assert(saw_plan);
  assert(saw_operational_plan);
  assert(saw_trading_plan_context);
  assert(saw_black_scholes);
  assert(saw_orchestrator_delete);
  assert(saw_trading);
  assert(saw_trading_pause);
  assert(saw_trading_resume);
  assert(saw_trading_stop);
  assert(saw_trading_set_risk);
  assert(saw_trading_risk_view);
  assert(saw_trading_strategies);
  assert(saw_trading_create_strategy);
  assert(saw_trading_directives);
  assert(saw_trading_create_directive);
  assert(saw_trading_audit);
  assert(saw_trading_activity);
  assert(saw_trading_mcp_tools);
  assert(saw_keystore_health);
  assert(saw_keystore_readiness);
  assert(saw_keystore);
  assert(saw_keystore_metrics);
  assert(saw_network_vm);
  assert(saw_network_bonds);
  assert(saw_network_create_bond);
  assert(saw_network_loans);
  assert(saw_network_create_loan);
  assert(saw_network_swaps);
  assert(saw_network_create_swap);
  assert(saw_network_staking);
  assert(saw_network_create_staking);
  assert(saw_network_synthetic_assets);
  assert(saw_network_create_synthetic_asset);
  assert(saw_network_liquidity_pools);
  assert(saw_network_create_liquidity_pool);
  assert(saw_hummingbot_assets);
  assert(saw_hummingbot_symbols);
  assert(saw_hummingbot);
  assert(saw_hummingbot_config);
  assert(saw_wise);
  assert(saw_wise_profile);
  assert(saw_wise_quote);
  assert(saw_wise_transfer);
  assert(saw_airwallex_accounts);
  assert(saw_airwallex_payments);
  assert(saw_airwallex);
  assert(saw_airwallex_beneficiaries);
  assert(saw_airwallex_create_beneficiary);
  assert(client.match_engine.websocket_url() == "wss://matchengine.example/ws");
  assert(client.match_engine.order_command("{\"symbol\":\"BTC-USDT\"}").find("matchengine.order_command.v2") != std::string::npos);
  assert(client.match_engine.market_data_subscribe("{\"symbols\":[\"BTC-USDT\"]}").find("matchengine.market_data_subscribe.v2") != std::string::npos);
  std::string last_matchengine_message;
  auto sender = [&](const std::string& message) {
    last_matchengine_message = message;
    return twofinance::HttpResponse{200, "{\"ok\":true}"};
  };
  assert(client.match_engine.send_order(sender, "{\"symbol\":\"BTC-USDT\"}").body == "{\"ok\":true}");
  assert(last_matchengine_message.find("matchengine.order_command.v2") != std::string::npos);
  assert(client.match_engine.subscribe_market_data(sender, "{\"symbols\":[\"BTC-USDT\"]}").body == "{\"ok\":true}");
  assert(last_matchengine_message.find("matchengine.market_data_subscribe.v2") != std::string::npos);
}
