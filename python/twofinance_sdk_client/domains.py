from __future__ import annotations

from typing import Any
from urllib.parse import quote

from .service import ServiceClient


def _path(value: str) -> str:
    return quote(value or "", safe="")


class AuthClient(ServiceClient):
    def __init__(
        self,
        base_url: str,
        realm: str = "2finance",
        client_id: str = "2finance-network",
        phone_client_id: str = "2finance-network-phone",
        **kwargs: Any,
    ):
        super().__init__(base_url, **kwargs)
        self.realm = realm
        self.client_id = client_id
        self.phone_client_id = phone_client_id

    def auth_path(self, client_id: str, endpoint: str) -> str:
        return f"/v1/2finance-authenticator/{self.realm}/{client_id}/{endpoint.lstrip('/')}"

    def login(self, username: str, password: str) -> Any:
        return self.post(self.auth_path(self.client_id, "/login"), {"username": username, "password": password})

    def sign_up(self, request: dict[str, Any]) -> Any:
        return self.post(self.auth_path(self.client_id, "/signup"), request)

    def refresh_token(self, refresh_token: str) -> Any:
        return self.post(self.auth_path(self.client_id, "/refresh"), {"refresh_token": refresh_token})

    def phone_login(self, phone_number: str, code: str) -> Any:
        return self.post(
            self.auth_path(self.phone_client_id, "/phone/sms/login"),
            {"phone_number": phone_number, "code": code},
        )

    def jwks(self) -> Any:
        return self.get(self.oidc_path("/protocol/openid-connect/certs"))

    def validate_token(self, token: str) -> Any:
        return self.post(self.oidc_path("/protocol/openid-connect/token/introspect"), {"token": token})

    def oidc_path(self, endpoint: str) -> str:
        return f"/realms/{self.realm}/{endpoint.lstrip('/')}"


class AnalyticsClient(ServiceClient):
    def indicators(self) -> Any:
        return self.get("/analytics/indicators")

    def calculate_technical_analysis(self, request: dict[str, Any]) -> Any:
        return self.post("/analytics/technical-analysis:calculate", request)

    def optimize_portfolio(self, request: dict[str, Any]) -> Any:
        return self.post("/portfolio-manager/optimizer", request)

    def upsert_candles(self, request: dict[str, Any]) -> Any:
        return self.post("/analytics/candles:upsert", request)

    def rankings(self) -> Any:
        return self.get("/portfolio-manager/rankings")

    def balances(self, account_id: str) -> Any:
        return self.get(f"/portfolio-manager/balances/{_path(account_id)}")

    def black_scholes(self, query: str = "") -> Any:
        suffix = f"?{query}" if query else ""
        return self.get(f"/risk-manager/blackscholes{suffix}")

    def staking(self) -> Any:
        return self.get("/staking")


class MCPClient(ServiceClient):
    def __init__(self, base_url: str, **kwargs: Any):
        super().__init__(base_url, **kwargs)
        self._next_id = 1

    def call(self, method: str, params: Any = None) -> Any:
        request_id = self._next_id
        self._next_id += 1
        return self.post("/mcp", {"jsonrpc": "2.0", "id": request_id, "method": method, "params": params})

    def call_tool(self, name: str, arguments: dict[str, Any] | None = None) -> Any:
        return self.call("tools/call", {"name": name, "arguments": arguments or {}})

    def list_tools(self) -> Any:
        return self.call("tools/list")

    def list_prompts(self) -> Any:
        return self.call("prompts/list")

    def list_resources(self) -> Any:
        return self.call("resources/list")

    def read_resource(self, uri: str) -> Any:
        return self.call("resources/read", {"uri": uri})

    def get_prompt(self, name: str, arguments: dict[str, Any] | None = None) -> Any:
        return self.call("prompts/get", {"name": name, "arguments": arguments or {}})

    def conversation_plan(self, arguments: dict[str, Any]) -> Any:
        return self.call_tool("finance_assistant.conversation.plan", arguments)


class OrchestratorClient(ServiceClient):
    def catalog(self) -> Any:
        return self.get("/v1/mcphost/catalog/packages")

    def create_session(self, request: dict[str, Any]) -> Any:
        return self.post("/v1/mcphost/sessions", request)

    def send_message(self, request: dict[str, Any]) -> Any:
        return self.post("/v1/mcphost/messages", request)

    def tools(self) -> Any:
        return self.get("/v1/mcphost/tools")

    def prompts(self) -> Any:
        return self.get("/v1/mcphost/prompts")

    def resources(self) -> Any:
        return self.get("/v1/mcphost/resources")

    def providers(self) -> Any:
        return self.get("/v1/mcphost/providers")

    def approvals(self) -> Any:
        return self.get("/v1/mcphost/approvals")

    def observability(self) -> Any:
        return self.get("/v1/mcphost/observability")

    def delete_session(self, session_id: str) -> Any:
        return self.delete(f"/v1/mcphost/sessions/{_path(session_id)}")


class TradingControlClient(ServiceClient):
    def robots(self) -> Any:
        return self.get("/robots")

    def create_robot(self, request: dict[str, Any]) -> Any:
        return self.post("/robots", request)

    def start_robot(self, robot_id: str) -> Any:
        return self.post(f"/robots/{_path(robot_id)}:start")

    def robot(self, robot_id: str) -> Any:
        return self.get(f"/robots/{_path(robot_id)}")

    def pause_robot(self, robot_id: str) -> Any:
        return self.post(f"/robots/{_path(robot_id)}:pause")

    def resume_robot(self, robot_id: str) -> Any:
        return self.post(f"/robots/{_path(robot_id)}:resume")

    def stop_robot(self, robot_id: str) -> Any:
        return self.post(f"/robots/{_path(robot_id)}:stop")

    def risk_policy(self, robot_id: str) -> Any:
        return self.get(f"/robots/{_path(robot_id)}/risk-policy")

    def set_risk_policy(self, robot_id: str, request: dict[str, Any]) -> Any:
        return self.put(f"/robots/{_path(robot_id)}/risk-policy", request)

    def risk_view(self, robot_id: str) -> Any:
        return self.get(f"/risk-view/{_path(robot_id)}")

    def strategies(self) -> Any:
        return self.get("/strategies")

    def create_strategy(self, request: dict[str, Any]) -> Any:
        return self.post("/strategies", request)

    def directives(self) -> Any:
        return self.get("/directives")

    def create_directive(self, request: dict[str, Any]) -> Any:
        return self.post("/directives", request)

    def audit(self) -> Any:
        return self.get("/audit")

    def activity(self) -> Any:
        return self.get("/activity")

    def mcp_tools(self) -> Any:
        return self.get("/mcp/tools")


class KeyStoreClient(ServiceClient):
    def health(self) -> Any:
        return self.get("/healthz")

    def readiness(self) -> Any:
        return self.get("/readyz")

    def start_keygen(self, request: dict[str, Any]) -> Any:
        return self.post("/keystore/keygen/start", request)

    def start_signing(self, request: dict[str, Any]) -> Any:
        return self.post("/keystore/signing/start", request)

    def keygen_signature(self, request: dict[str, Any]) -> Any:
        return self.post("/keystore/keygen/signature", request)

    def signing_signature(self, request: dict[str, Any]) -> Any:
        return self.post("/keystore/signing/signature", request)

    def start_resharing(self, request: dict[str, Any]) -> Any:
        return self.post("/keystore/resharing/start", request)

    def keys(self, user_public_key: str) -> Any:
        return self.get(f"/keystore/keys/{_path(user_public_key)}")

    def signatures(self, user_public_key: str) -> Any:
        return self.get(f"/keystore/signatures/{_path(user_public_key)}")

    def metrics(self) -> Any:
        return self.get("/keystore/tss/metrics")


class NetworkClient(ServiceClient):
    def virtual_machine(self) -> Any:
        return self.post(
            "/v2/2finance-network/query",
            {"method": "get_blocks", "params": {"page": 1, "limit": 1}},
        )

    def market_candles(self, market: str, query: str = "") -> Any:
        suffix = f"?{query}" if query else ""
        return self.get(f"/v2/2finance-network/markets/{_path(market)}/candles{suffix}")

    def products(self, product_type: str) -> Any:
        return self.get(f"/v2/2finance-network/products/{_path(product_type)}")

    def create_product(self, product_type: str, request: dict[str, Any]) -> Any:
        return self.post(f"/v2/2finance-network/products/{_path(product_type)}", request)

    def bonds(self) -> Any:
        return self.products("bonds")

    def create_bond(self, request: dict[str, Any]) -> Any:
        return self.create_product("bonds", request)

    def loans(self) -> Any:
        return self.products("loans")

    def create_loan(self, request: dict[str, Any]) -> Any:
        return self.create_product("loans", request)

    def swaps(self) -> Any:
        return self.products("swaps")

    def create_swap(self, request: dict[str, Any]) -> Any:
        return self.create_product("swaps", request)

    def staking_products(self) -> Any:
        return self.products("staking")

    def create_staking_product(self, request: dict[str, Any]) -> Any:
        return self.create_product("staking", request)

    def synthetic_assets(self) -> Any:
        return self.products("synthetic-assets")

    def create_synthetic_asset(self, request: dict[str, Any]) -> Any:
        return self.create_product("synthetic-assets", request)

    def liquidity_pools(self) -> Any:
        return self.products("liquidity-pools")

    def create_liquidity_pool(self, request: dict[str, Any]) -> Any:
        return self.create_product("liquidity-pools", request)


class HummingbotClient(ServiceClient):
    def assets(self) -> Any:
        return self.get("/api/v1/assets")

    def symbols(self) -> Any:
        return self.get("/api/v1/symbols")

    def balances(self) -> Any:
        return self.get("/api/v1/balances")

    def connector_config(self, request: dict[str, Any]) -> Any:
        return self.post("/api/v1/connectors/2finance/config", request)


class ProviderClient(ServiceClient):
    pass


class WiseClient(ProviderClient):
    def profiles(self) -> Any:
        return self.get("/v1/profiles")

    def profile(self, profile_id: str) -> Any:
        return self.get(f"/v1/profiles/{_path(profile_id)}")

    def create_quote(self, profile_id: str, request: dict[str, Any]) -> Any:
        return self.post(f"/v3/profiles/{_path(profile_id)}/quotes", request)

    def create_transfer(self, request: dict[str, Any]) -> Any:
        return self.post("/v1/transfers", request)


class AirwallexClient(ProviderClient):
    def accounts(self) -> Any:
        return self.get("/api/v1/accounts")

    def payments(self) -> Any:
        return self.get("/api/v1/payments")

    def create_payment(self, request: dict[str, Any]) -> Any:
        return self.post("/api/v1/payments", request)

    def beneficiaries(self) -> Any:
        return self.get("/api/v1/beneficiaries")

    def create_beneficiary(self, request: dict[str, Any]) -> Any:
        return self.post("/api/v1/beneficiaries", request)


class MatchEngineClient:
    def __init__(self, websocket_url: str):
        self.websocket_url = (websocket_url or "").strip()

    def order_command(self, **command: Any) -> dict[str, Any]:
        payload = {
            "schema": "matchengine.order_command.v2",
            "message_type": "ORDER",
            "operation": "ADD",
        }
        payload.update(command)
        return payload

    def market_data_subscribe(self, **request: Any) -> dict[str, Any]:
        payload = {"schema": "matchengine.market_data_subscribe.v2"}
        payload.update(request)
        return payload

    def send_order(self, sender: Any, **command: Any) -> Any:
        return self._send(sender, self.order_command(**command))

    def subscribe_market_data(self, sender: Any, **request: Any) -> Any:
        return self._send(sender, self.market_data_subscribe(**request))

    @staticmethod
    def _send(sender: Any, message: dict[str, Any]) -> Any:
        if callable(sender):
            return sender(message)
        send = getattr(sender, "send", None)
        if callable(send):
            return send(message)
        raise TypeError("matchengine sender must be callable or expose send(message)")


class PlannerClient:
    def __init__(
        self,
        mcp: MCPClient,
        orchestrator: OrchestratorClient,
        analytics: AnalyticsClient,
        trading_control: TradingControlClient,
    ):
        self.mcp = mcp
        self.orchestrator = orchestrator
        self.analytics = analytics
        self.trading_control = trading_control

    def conversation_plan(self, request: dict[str, Any]) -> Any:
        return self.mcp.conversation_plan(request)

    def orchestrated_plan(self, request: dict[str, Any]) -> Any:
        return self.orchestrator.send_message(request)

    def operational_plan(self, request: dict[str, Any]) -> Any:
        return self.orchestrated_plan(request)

    def trading_plan(self, request: dict[str, Any]) -> Any:
        context = dict(request.get("context") or {})
        if request.get("use_trading") is True or request.get("useTrading") is True:
            try:
                context["trading_robots"] = self.trading_control.robots()
            except Exception:
                # Best-effort enrichment keeps planning usable when trading is unavailable.
                pass
        if request.get("use_analytics") is True or request.get("useAnalytics") is True:
            try:
                context["analytics_indicators"] = self.analytics.indicators()
            except Exception:
                # Best-effort enrichment keeps planning usable when analytics is unavailable.
                pass
        return self.conversation_plan({**request, "context": context})
