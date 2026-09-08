import io
import json
import unittest
from pathlib import Path
from unittest.mock import patch
from urllib.error import HTTPError as UrlHTTPError

from twofinance_sdk_client import (
    ClientCredentialsTokenSource,
    DEFAULT_SERVICE_CATALOG,
    DomainOperationsCatalog,
    IdempotencyRecord,
    MarketDirectory,
    PaginationResponse,
    RequestOptions,
    ResolvedOperation,
    SDK_NAME,
    SDKConfig,
    SDKError,
    ServiceClient,
    ServiceCatalog,
    StaticTokenSource,
    TwoFinanceClient,
    bearer_authorization,
    config_from_env,
    __version__,
)

REQUEST_OPTIONS_FIXTURE = json.loads(
    (Path(__file__).resolve().parents[2] / "contracts/examples/request-options.json").read_text()
)
CONTRACTS_DIR = Path(__file__).resolve().parents[2] / "contracts/examples"


def load_contract_fixture(name):
    return json.loads((CONTRACTS_DIR / name).read_text())


def contract_operation(fixture, domain_name, operation_name):
    domain = next(item for item in fixture["domains"] if item["name"] == domain_name)
    return next(item for item in domain["operations"] if item["name"] == operation_name)


class FakeResponse:
    def __enter__(self):
        return self

    def __exit__(self, exc_type, exc, tb):
        return False

    def read(self):
        return json.dumps({"ok": True}).encode("utf-8")


class BytesResponse(FakeResponse):
    def __init__(self, payload):
        self.payload = payload

    def read(self):
        return json.dumps(self.payload).encode("utf-8")


class SDKTest(unittest.TestCase):
    def test_exposes_sdk_metadata(self):
        self.assertEqual(SDK_NAME, "2finance-sdk-client")
        self.assertEqual(__version__, "0.1.0")
        self.assertEqual(len(DEFAULT_SERVICE_CATALOG.services), 12)
        self.assertEqual(DEFAULT_SERVICE_CATALOG.services[0].env, "TWO_FINANCE_AUTH_URL")

    def test_shared_contract_fixtures_describe_public_sdk_operations(self):
        domains = load_contract_fixture("domain-operations.json")
        error = load_contract_fixture("error.json")
        pagination = load_contract_fixture("pagination.json")
        idempotency = load_contract_fixture("idempotency.json")

        self.assertEqual(domains["schema"], "sdk.domain_operations.v2")
        self.assertEqual(
            contract_operation(domains, "analytics", "balances")["path"],
            "/portfolio-manager/balances/{account_id}",
        )
        self.assertEqual(
            contract_operation(domains, "analytics", "balances")["path_params"],
            ["account_id"],
        )
        self.assertEqual(
            contract_operation(domains, "planner", "trading_plan")["request_schema"],
            "planner.trading_plan.request.v1",
        )
        self.assertEqual(error["error"], "rate_limited")
        self.assertEqual(error["code"], "HTTP_429")
        self.assertEqual(pagination["next_cursor"], "cursor-next")
        self.assertEqual(idempotency["idempotency_key"], "idem-001")

    def test_shared_sdk_models_parse_contract_fixtures(self):
        error = SDKError.from_dict(load_contract_fixture("error.json"))
        pagination = PaginationResponse.from_dict(load_contract_fixture("pagination.json"))
        idempotency = IdempotencyRecord.from_dict(load_contract_fixture("idempotency.json"))
        catalog = ServiceCatalog.from_dict(load_contract_fixture("service-catalog.json"))
        operations = DomainOperationsCatalog.from_dict(load_contract_fixture("domain-operations.json"))
        market_directory = MarketDirectory.from_dict(load_contract_fixture("market-directory.json"))

        self.assertEqual(error.code, "HTTP_429")
        self.assertEqual(error.details["request_id"], "req_2finance_001")
        self.assertEqual(pagination.limit, 25)
        self.assertEqual(pagination.next_cursor, "cursor-next")
        self.assertEqual(idempotency.idempotency_key, "idem-001")
        self.assertEqual(catalog.services[0].name, "auth")
        self.assertEqual(operations.schema, "sdk.domain_operations.v2")
        self.assertEqual(market_directory.markets[0].engine_id, "engine-btc-usdt-01")
        self.assertEqual(market_directory.markets[0].precision["amount"], 8)
        self.assertEqual(operations.domains[0].operations[0].request_schema, "auth.login.request.v1")
        self.assertEqual(
            operations.operation("analytics", "balances").path,
            "/portfolio-manager/balances/{account_id}",
        )
        resolved_balances = operations.operation("analytics", "balances").resolve(
            {"account_id": "acct/1 ok"}
        )
        self.assertEqual(resolved_balances.method, "GET")
        self.assertEqual(
            resolved_balances.path,
            "/portfolio-manager/balances/acct%2F1%20ok",
        )
        self.assertEqual(
            operations.resolve_operation(
                "analytics",
                "balances",
                {"account_id": "acct/1 ok"},
            ).path,
            resolved_balances.path,
        )
        resolved_risk = operations.operation("analytics", "black_scholes").resolve(
            query={
                "symbol": "BTC/USD",
                "strike": 100000,
                "ignored": "drop-me",
                "volatility": 0.5,
            }
        )
        self.assertEqual(
            resolved_risk.path,
            "/risk-manager/blackscholes?symbol=BTC%2FUSD&strike=100000&volatility=0.5",
        )

    def test_network_client_returns_typed_market_directory(self):
        client = TwoFinanceClient(SDKConfig(network_url="https://network.example"))

        with patch(
            "twofinance_sdk_client.service.urlopen",
            lambda request: BytesResponse(load_contract_fixture("market-directory.json")),
        ):
            directory = client.network.market_directory()

        self.assertEqual(directory.markets[0].engine_id, "engine-btc-usdt-01")

    def test_config_from_env(self):
        config = config_from_env(
            {
                "TWO_FINANCE_AUTH_URL": "https://auth.example",
                "TWO_FINANCE_ANALYTICS_URL": "https://analytics.example",
                "TWO_FINANCE_MATCHENGINE_WS_URL": "wss://matchengine.example/ws",
            }
        )
        self.assertEqual(config.auth_url, "https://auth.example")
        self.assertEqual(config.analytics_url, "https://analytics.example")
        self.assertEqual(config.service_url("analytics"), "https://analytics.example")
        self.assertEqual(config.service_url("match_engine"), "wss://matchengine.example/ws")
        self.assertEqual(config.service_urls()["analytics"], "https://analytics.example")
        self.assertEqual(config.service_urls()["matchengine"], "wss://matchengine.example/ws")
        self.assertEqual(len(config.configured_services()), 3)
        self.assertEqual(config.configured_services()[1].name, "analytics")
        self.assertEqual(config.configured_services()[1].url, "https://analytics.example")
        self.assertEqual(len(config.missing_service_urls()), 9)
        self.assertEqual(config.missing_service_urls()[0].name, "network")
        self.assertEqual(config.missing_service_urls()[0].env, "TWO_FINANCE_NETWORK_URL")
        self.assertEqual(config.auth_realm, "2finance")

    def test_bearer_authorization(self):
        self.assertEqual(bearer_authorization("abc"), "Bearer abc")
        self.assertEqual(bearer_authorization("Bearer abc"), "Bearer abc")
        self.assertEqual(bearer_authorization(""), "")

    def test_service_client_injects_bearer(self):
        seen = {}

        def fake_urlopen(request):
            seen["url"] = request.full_url
            seen["authorization"] = request.headers.get("Authorization")
            return FakeResponse()

        client = TwoFinanceClient(
            SDKConfig(analytics_url="https://analytics.example"),
            token_source=StaticTokenSource("token-123"),
        )
        with patch("twofinance_sdk_client.service.urlopen", fake_urlopen):
            result = client.analytics.indicators()

        self.assertEqual(result, {"ok": True})
        self.assertEqual(seen["url"], "https://analytics.example/analytics/indicators")
        self.assertEqual(seen["authorization"], "Bearer token-123")

    def test_service_client_requests_resolved_operation(self):
        seen = {}

        def fake_urlopen(request):
            seen["method"] = request.get_method()
            seen["url"] = request.full_url
            return FakeResponse()

        service = ServiceClient("https://analytics.example")
        with patch("twofinance_sdk_client.service.urlopen", fake_urlopen):
            result = service.request_operation(
                ResolvedOperation("GET", "/portfolio-manager/balances/acct%2Fresolved")
            )

        self.assertEqual(result, {"ok": True})
        self.assertEqual(seen["method"], "GET")
        self.assertEqual(
            seen["url"],
            "https://analytics.example/portfolio-manager/balances/acct%2Fresolved",
        )

        operations = DomainOperationsCatalog.from_dict(load_contract_fixture("domain-operations.json"))
        with patch("twofinance_sdk_client.service.urlopen", fake_urlopen):
            service.request_catalog_operation(
                operations,
                "analytics",
                "balances",
                path_params={"account_id": "acct/1 ok"},
            )
        self.assertEqual(
            seen["url"],
            "https://analytics.example/portfolio-manager/balances/acct%2F1%20ok",
        )

    def test_auth_client_exposes_jwks_and_token_validation_helpers(self):
        seen = []

        def fake_urlopen(request):
            seen.append(f"{request.get_method()} {request.full_url} {request.data.decode('utf-8') if request.data else ''}")
            return FakeResponse()

        client = TwoFinanceClient(SDKConfig(auth_url="https://auth.example"))
        with patch("twofinance_sdk_client.service.urlopen", fake_urlopen):
            client.auth.jwks()
            client.auth.validate_token("token-1")

        self.assertIn("GET https://auth.example/realms/2finance/protocol/openid-connect/certs ", seen)
        self.assertIn(
            'POST https://auth.example/realms/2finance/protocol/openid-connect/token/introspect {"token": "token-1"}',
            seen,
        )

    def test_service_client_applies_request_options_and_idempotency_key(self):
        seen = {}
        fixture = REQUEST_OPTIONS_FIXTURE

        def fake_urlopen(request, timeout=None):
            seen["url"] = request.full_url
            seen["trace"] = request.headers.get("X-trace-id")
            seen["idempotency"] = request.headers.get("Idempotency-key")
            seen["timeout"] = timeout
            return FakeResponse()

        service = ServiceClient(fixture["request"]["base_url"])
        with patch("twofinance_sdk_client.service.urlopen", fake_urlopen):
            service.post(
                fixture["request"]["path"],
                {"symbol": "BTC-USDT"},
                RequestOptions(
                    headers=fixture["request"]["headers"],
                    idempotency_key=f" {fixture['request']['idempotency_key']} ",
                    query=fixture["request"]["query"],
                    timeout=fixture["request"]["timeout_ms"] / 1000,
                    max_retries=fixture["request"]["max_retries"],
                    page=fixture["request"]["pagination"]["page"],
                    limit=fixture["request"]["pagination"]["limit"],
                ),
            )

        self.assertEqual(seen["url"], fixture["expected"]["url"])
        self.assertEqual(seen["trace"], fixture["expected"]["headers"]["X-Trace-ID"])
        self.assertEqual(seen["idempotency"], fixture["expected"]["headers"]["Idempotency-Key"])
        self.assertEqual(seen["timeout"], 0.5)

    def test_service_client_retries_retryable_responses(self):
        attempts = 0

        def fake_urlopen(request):
            nonlocal attempts
            attempts += 1
            if attempts == 1:
                raise UrlHTTPError(request.full_url, 500, "temporary", {}, io.BytesIO(b"temporary"))
            return FakeResponse()

        service = ServiceClient("https://analytics.example")
        with patch("twofinance_sdk_client.service.urlopen", fake_urlopen):
            result = service.get("/analytics/indicators", RequestOptions(max_retries=1))

        self.assertEqual(attempts, 2)
        self.assertEqual(result, {"ok": True})

    def test_client_credentials_token_source_fetches_and_caches(self):
        calls = []

        def fake_urlopen(request):
            calls.append(
                {
                    "url": request.full_url,
                    "body": request.data.decode("utf-8"),
                }
            )

            class TokenResponse:
                def __enter__(self):
                    return self

                def __exit__(self, exc_type, exc, tb):
                    return False

                def read(self):
                    return json.dumps({"access_token": "cc-token", "expires_in": 3600}).encode("utf-8")

            return TokenResponse()

        source = ClientCredentialsTokenSource(
            token_url="https://auth.example/token",
            client_id="client-id",
            client_secret="client-secret",
            scopes=["analytics:read", "mcp:invoke"],
        )
        with patch("twofinance_sdk_client.auth.urlopen", fake_urlopen):
            self.assertEqual(source.token(), "cc-token")
            self.assertEqual(source.token(), "cc-token")

        self.assertEqual(len(calls), 1)
        self.assertEqual(calls[0]["url"], "https://auth.example/token")
        self.assertIn("grant_type=client_credentials", calls[0]["body"])
        self.assertIn("scope=analytics%3Aread+mcp%3Ainvoke", calls[0]["body"])

    def test_matchengine_order_command(self):
        client = TwoFinanceClient(SDKConfig(matchengine_ws_url="wss://matchengine.example/ws"))
        command = client.matchengine.order_command(
            client_order_id="co-1",
            idempotency_key="idem-1",
            symbol="BTC-USDT",
            side="buy",
            type="limit",
            quantity="0.01",
        )

        self.assertEqual(client.matchengine.websocket_url, "wss://matchengine.example/ws")
        self.assertEqual(command["schema"], "matchengine.order_command.v2")
        self.assertEqual(command["symbol"], "BTC-USDT")
        subscription = client.matchengine.market_data_subscribe(symbols=["BTC-USDT"], channels=["book"])
        self.assertEqual(subscription["schema"], "matchengine.market_data_subscribe.v2")
        self.assertEqual(subscription["symbols"], ["BTC-USDT"])
        messages = []

        def sender(message):
            messages.append(message)
            return {"ok": True}

        self.assertEqual(client.matchengine.send_order(sender, **command), {"ok": True})
        self.assertEqual(client.matchengine.subscribe_market_data(sender, **subscription), {"ok": True})
        self.assertEqual(messages[0]["schema"], "matchengine.order_command.v2")
        self.assertEqual(messages[1]["schema"], "matchengine.market_data_subscribe.v2")

    def test_domain_clients_expose_core_service_endpoints(self):
        seen = []

        def fake_urlopen(request):
            seen.append(f"{request.get_method()} {request.full_url}")
            return FakeResponse()

        client = TwoFinanceClient(
            SDKConfig(
                analytics_url="https://analytics.example",
                network_url="https://network.example",
                trading_control_url="https://trading.example",
                keystore_url="https://keys.example",
                hummingbot_url="https://hbot.example",
                wise_url="https://wise.example",
                airwallex_url="https://airwallex.example",
            )
        )
        with patch("twofinance_sdk_client.service.urlopen", fake_urlopen):
            client.analytics.upsert_candles({"symbol": "BTC-USDT"})
            client.analytics.rankings()
            client.analytics.balances("acct/1")
            client.analytics.black_scholes("symbol=BTC")
            client.analytics.staking()
            client.network.market_candles("BTC/USDT", "limit=10")
            client.network.bonds()
            client.network.create_bond({"symbol": "BOND1"})
            client.network.loans()
            client.network.create_loan({"loan": "ln1"})
            client.network.swaps()
            client.network.create_swap({"pair": "BTC-USDT"})
            client.network.staking_products()
            client.network.create_staking_product({"asset": "TWO"})
            client.network.synthetic_assets()
            client.network.create_synthetic_asset({"asset": "sBTC"})
            client.network.liquidity_pools()
            client.network.create_liquidity_pool({"pool": "BTC-USDT"})
            client.trading_control.start_robot("robot/1")
            client.trading_control.pause_robot("robot/1")
            client.trading_control.risk_policy("robot/1")
            client.trading_control.risk_view("robot/1")
            client.trading_control.strategies()
            client.trading_control.create_strategy({"name": "mean-reversion"})
            client.trading_control.directives()
            client.trading_control.create_directive({"action": "rebalance"})
            client.trading_control.audit()
            client.trading_control.activity()
            client.trading_control.mcp_tools()
            client.keystore.health()
            client.keystore.readiness()
            client.keystore.start_signing({"key": "k1"})
            client.keystore.keys("pub/1")
            client.keystore.signatures("pub/1")
            client.keystore.metrics()
            client.hummingbot.balances()
            client.hummingbot.connector_config({"connector": "2finance"})
            client.providers.wise.profiles()
            client.providers.wise.profile("profile/1")
            client.providers.wise.create_quote("profile/1", {"source": "USD"})
            client.providers.wise.create_transfer({"target": "BRL"})
            client.providers.airwallex.accounts()
            client.providers.airwallex.payments()
            client.providers.airwallex.create_payment({"amount": 10})
            client.providers.airwallex.beneficiaries()
            client.providers.airwallex.create_beneficiary({"name": "beneficiary"})

        self.assertIn("POST https://analytics.example/analytics/candles:upsert", seen)
        self.assertIn("GET https://analytics.example/portfolio-manager/rankings", seen)
        self.assertIn("GET https://analytics.example/portfolio-manager/balances/acct%2F1", seen)
        self.assertIn("GET https://analytics.example/risk-manager/blackscholes?symbol=BTC", seen)
        self.assertIn("GET https://analytics.example/staking", seen)
        self.assertIn("GET https://network.example/v2/2finance-network/markets/BTC%2FUSDT/candles?limit=10", seen)
        self.assertIn("GET https://network.example/v2/2finance-network/products/bonds", seen)
        self.assertIn("POST https://network.example/v2/2finance-network/products/bonds", seen)
        self.assertIn("GET https://network.example/v2/2finance-network/products/loans", seen)
        self.assertIn("POST https://network.example/v2/2finance-network/products/loans", seen)
        self.assertIn("GET https://network.example/v2/2finance-network/products/swaps", seen)
        self.assertIn("POST https://network.example/v2/2finance-network/products/swaps", seen)
        self.assertIn("GET https://network.example/v2/2finance-network/products/staking", seen)
        self.assertIn("POST https://network.example/v2/2finance-network/products/staking", seen)
        self.assertIn("GET https://network.example/v2/2finance-network/products/synthetic-assets", seen)
        self.assertIn("POST https://network.example/v2/2finance-network/products/synthetic-assets", seen)
        self.assertIn("GET https://network.example/v2/2finance-network/products/liquidity-pools", seen)
        self.assertIn("POST https://network.example/v2/2finance-network/products/liquidity-pools", seen)
        self.assertIn("POST https://trading.example/robots/robot%2F1:start", seen)
        self.assertIn("POST https://trading.example/robots/robot%2F1:pause", seen)
        self.assertIn("GET https://trading.example/robots/robot%2F1/risk-policy", seen)
        self.assertIn("GET https://trading.example/risk-view/robot%2F1", seen)
        self.assertIn("GET https://trading.example/strategies", seen)
        self.assertIn("POST https://trading.example/strategies", seen)
        self.assertIn("GET https://trading.example/directives", seen)
        self.assertIn("POST https://trading.example/directives", seen)
        self.assertIn("GET https://trading.example/audit", seen)
        self.assertIn("GET https://trading.example/activity", seen)
        self.assertIn("GET https://trading.example/mcp/tools", seen)
        self.assertIn("GET https://keys.example/healthz", seen)
        self.assertIn("GET https://keys.example/readyz", seen)
        self.assertIn("POST https://keys.example/keystore/signing/start", seen)
        self.assertIn("GET https://keys.example/keystore/keys/pub%2F1", seen)
        self.assertIn("GET https://keys.example/keystore/signatures/pub%2F1", seen)
        self.assertIn("GET https://keys.example/keystore/tss/metrics", seen)
        self.assertIn("GET https://hbot.example/api/v1/balances", seen)
        self.assertIn("POST https://hbot.example/api/v1/connectors/2finance/config", seen)
        self.assertIn("GET https://wise.example/v1/profiles", seen)
        self.assertIn("GET https://wise.example/v1/profiles/profile%2F1", seen)
        self.assertIn("POST https://wise.example/v3/profiles/profile%2F1/quotes", seen)
        self.assertIn("POST https://wise.example/v1/transfers", seen)
        self.assertIn("GET https://airwallex.example/api/v1/accounts", seen)
        self.assertIn("GET https://airwallex.example/api/v1/payments", seen)
        self.assertIn("POST https://airwallex.example/api/v1/payments", seen)
        self.assertIn("GET https://airwallex.example/api/v1/beneficiaries", seen)
        self.assertIn("POST https://airwallex.example/api/v1/beneficiaries", seen)

    def test_mcp_and_orchestrator_expose_tools_resources_and_sessions(self):
        seen = []

        def fake_urlopen(request):
            body = request.data.decode("utf-8") if request.data else ""
            seen.append(f"{request.get_method()} {request.full_url} {body}")
            return FakeResponse()

        client = TwoFinanceClient(
            SDKConfig(
                mcp_url="https://mcp.example",
                orchestrator_url="https://orchestrator.example",
            )
        )
        with patch("twofinance_sdk_client.service.urlopen", fake_urlopen):
            client.mcp.list_tools()
            client.mcp.list_prompts()
            client.mcp.read_resource("resource://portfolio")
            client.orchestrator.tools()
            client.orchestrator.resources()
            client.orchestrator.approvals()
            client.orchestrator.delete_session("session/1")
            client.planner.operational_plan({"session_id": "session-1", "message": "operate"})

        self.assertTrue(any('"method": "tools/list"' in item for item in seen))
        self.assertTrue(any('"method": "prompts/list"' in item for item in seen))
        self.assertTrue(any("resource://portfolio" in item for item in seen))
        self.assertIn("GET https://orchestrator.example/v1/mcphost/tools ", seen)
        self.assertIn("GET https://orchestrator.example/v1/mcphost/resources ", seen)
        self.assertIn("GET https://orchestrator.example/v1/mcphost/approvals ", seen)
        self.assertIn("DELETE https://orchestrator.example/v1/mcphost/sessions/session%2F1 ", seen)
        self.assertTrue(any(item.startswith("POST https://orchestrator.example/v1/mcphost/messages ") for item in seen))

    def test_planner_trading_plan_enriches_context(self):
        seen = []

        def fake_urlopen(request):
            body = request.data.decode("utf-8") if request.data else ""
            seen.append(f"{request.get_method()} {request.full_url} {body}")
            if request.full_url == "https://trading.example/robots":
                return BytesResponse({"robots": [{"id": "robot-1"}]})
            if request.full_url == "https://analytics.example/analytics/indicators":
                return BytesResponse({"indicators": ["rsi"]})
            return FakeResponse()

        client = TwoFinanceClient(
            SDKConfig(
                analytics_url="https://analytics.example",
                mcp_url="https://mcp.example",
                trading_control_url="https://trading.example",
            )
        )
        with patch("twofinance_sdk_client.service.urlopen", fake_urlopen):
            client.planner.trading_plan(
                {
                    "goal": "rebalance BTC",
                    "context": {"account_id": "acct-1"},
                    "use_trading": True,
                    "use_analytics": True,
                }
            )

        mcp_calls = [item for item in seen if item.startswith("POST https://mcp.example/mcp ")]
        self.assertEqual(len(mcp_calls), 1)
        self.assertIn("finance_assistant.conversation.plan", mcp_calls[0])
        self.assertIn("trading_robots", mcp_calls[0])
        self.assertIn("analytics_indicators", mcp_calls[0])
        self.assertIn("account_id", mcp_calls[0])


if __name__ == "__main__":
    unittest.main()
