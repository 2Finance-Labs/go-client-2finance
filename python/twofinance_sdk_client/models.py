from __future__ import annotations

from dataclasses import dataclass
from typing import Any
from urllib.parse import quote, urlencode


@dataclass(frozen=True)
class SDKError:
    error: str
    message: str
    code: str
    details: dict[str, Any]

    @classmethod
    def from_dict(cls, payload: dict[str, Any]) -> "SDKError":
        return cls(
            error=str(payload["error"]),
            message=str(payload["message"]),
            code=str(payload["code"]),
            details=dict(payload.get("details") or {}),
        )


@dataclass(frozen=True)
class PaginationResponse:
    items: list[dict[str, Any]]
    limit: int
    cursor: str | None = None
    next_cursor: str | None = None

    @classmethod
    def from_dict(cls, payload: dict[str, Any]) -> "PaginationResponse":
        return cls(
            items=[dict(item) for item in payload.get("items", [])],
            limit=int(payload["limit"]),
            cursor=payload.get("cursor"),
            next_cursor=payload.get("next_cursor"),
        )


@dataclass(frozen=True)
class IdempotencyRecord:
    idempotency_key: str
    operation: str
    scope: str
    request_id: str

    @classmethod
    def from_dict(cls, payload: dict[str, Any]) -> "IdempotencyRecord":
        return cls(
            idempotency_key=str(payload["idempotency_key"]),
            operation=str(payload["operation"]),
            scope=str(payload["scope"]),
            request_id=str(payload["request_id"]),
        )


@dataclass(frozen=True)
class ServiceCatalogEntry:
    name: str
    env: str

    @classmethod
    def from_dict(cls, payload: dict[str, Any]) -> "ServiceCatalogEntry":
        return cls(name=str(payload["name"]), env=str(payload["env"]))


@dataclass(frozen=True)
class ServiceCatalog:
    services: list[ServiceCatalogEntry]

    @classmethod
    def from_dict(cls, payload: dict[str, Any]) -> "ServiceCatalog":
        return cls(
            services=[
                ServiceCatalogEntry.from_dict(dict(item))
                for item in payload.get("services", [])
            ]
        )


@dataclass(frozen=True)
class ConfiguredServiceEntry:
    name: str
    env: str
    url: str


@dataclass(frozen=True)
class MarketDefinition:
    """Language binding for contracts/schemas/market-directory.v1.json."""

    market_id: str
    symbol: str
    engine_id: str
    symbol_id: int
    route_epoch: int
    chain: dict[str, str]
    address: str
    endpoints: dict[str, str]
    status: str
    precision: dict[str, int]
    tick_size: str
    step_size: str
    limits: dict[str, dict[str, str | None]]
    fees: dict[str, str]
    capabilities: dict[str, Any]
    raw: dict[str, Any]

    @classmethod
    def from_dict(cls, payload: dict[str, Any]) -> "MarketDefinition":
        if payload.get("schema") != "2finance.market_definition.v1":
            raise ValueError("unsupported 2Finance market definition schema")
        return cls(
            market_id=str(payload["market_id"]),
            symbol=str(payload["symbol"]),
            engine_id=str(payload["engine_id"]),
            symbol_id=int(payload["symbol_id"]),
            route_epoch=int(payload["route_epoch"]),
            chain=dict(payload["chain"]),
            address=str(payload["address"]),
            endpoints=dict(payload["endpoints"]),
            status=str(payload["status"]),
            precision={key: int(value) for key, value in dict(payload["precision"]).items()},
            tick_size=str(payload["tick_size"]),
            step_size=str(payload["step_size"]),
            limits={key: dict(value) for key, value in dict(payload["limits"]).items()},
            fees={key: str(value) for key, value in dict(payload["fees"]).items()},
            capabilities=dict(payload["capabilities"]),
            raw=dict(payload),
        )


@dataclass(frozen=True)
class MarketDirectory:
    generated_at: str
    markets: list[MarketDefinition]

    @classmethod
    def from_dict(cls, payload: dict[str, Any]) -> "MarketDirectory":
        if payload.get("schema") != "2finance.market_directory.v1":
            raise ValueError("unsupported 2Finance market directory schema")
        return cls(
            generated_at=str(payload["generated_at"]),
            markets=[MarketDefinition.from_dict(dict(item)) for item in payload.get("markets", [])],
        )


@dataclass(frozen=True)
class DomainOperation:
    name: str
    method: str
    path: str
    path_params: list[str]
    query: list[str]
    request_schema: str | None = None
    response_schema: str | None = None
    notes: str | None = None

    @classmethod
    def from_dict(cls, payload: dict[str, Any]) -> "DomainOperation":
        return cls(
            name=str(payload["name"]),
            method=str(payload["method"]),
            path=str(payload["path"]),
            path_params=[str(item) for item in payload.get("path_params", [])],
            query=[str(item) for item in payload.get("query", [])],
            request_schema=payload.get("request_schema"),
            response_schema=payload.get("response_schema"),
            notes=payload.get("notes"),
        )

    def resolve(
        self,
        path_params: dict[str, Any] | None = None,
        query: dict[str, Any] | None = None,
    ) -> "ResolvedOperation":
        path_params = path_params or {}
        query = query or {}
        path = self.path
        for name in self.path_params:
            if name not in path_params:
                raise ValueError(f"2finance: missing operation path parameter {name}")
            path = path.replace("{" + name + "}", quote(str(path_params[name]), safe=""))

        query_pairs = [
            (name, str(query[name]))
            for name in self.query
            if name in query and query[name] is not None
        ]
        if query_pairs:
            separator = "&" if "?" in path else "?"
            path = f"{path}{separator}{urlencode(query_pairs)}"

        return ResolvedOperation(method=self.method.strip().upper(), path=path)


@dataclass(frozen=True)
class ResolvedOperation:
    method: str
    path: str


@dataclass(frozen=True)
class DomainOperationsDomain:
    name: str
    env: str
    transport: str | None
    description: str | None
    operations: list[DomainOperation]

    @classmethod
    def from_dict(cls, payload: dict[str, Any]) -> "DomainOperationsDomain":
        return cls(
            name=str(payload["name"]),
            env=str(payload["env"]),
            transport=payload.get("transport"),
            description=payload.get("description"),
            operations=[
                DomainOperation.from_dict(dict(item))
                for item in payload.get("operations", [])
            ],
        )


@dataclass(frozen=True)
class DomainOperationsCatalog:
    schema: str
    domains: list[DomainOperationsDomain]

    @classmethod
    def from_dict(cls, payload: dict[str, Any]) -> "DomainOperationsCatalog":
        return cls(
            schema=str(payload["schema"]),
            domains=[
                DomainOperationsDomain.from_dict(dict(item))
                for item in payload.get("domains", [])
            ],
        )

    def operation(self, domain_name: str, operation_name: str) -> DomainOperation | None:
        for domain in self.domains:
            if _domain_key(domain.name) != _domain_key(domain_name):
                continue
            for operation in domain.operations:
                if operation.name == operation_name:
                    return operation
            return None
        return None

    def resolve_operation(
        self,
        domain_name: str,
        operation_name: str,
        path_params: dict[str, Any] | None = None,
        query: dict[str, Any] | None = None,
    ) -> ResolvedOperation:
        operation = self.operation(domain_name, operation_name)
        if operation is None:
            raise ValueError(f"2finance: unknown operation {domain_name}.{operation_name}")
        return operation.resolve(path_params=path_params, query=query)


def _domain_key(value: str) -> str:
    return "".join(ch for ch in value.strip().lower() if ch not in "-_ ")


DEFAULT_SERVICE_CATALOG = ServiceCatalog(
    services=[
        ServiceCatalogEntry("auth", "TWO_FINANCE_AUTH_URL"),
        ServiceCatalogEntry("network", "TWO_FINANCE_NETWORK_URL"),
        ServiceCatalogEntry("analytics", "TWO_FINANCE_ANALYTICS_URL"),
        ServiceCatalogEntry("orchestrator", "TWO_FINANCE_ORCHESTRATOR_URL"),
        ServiceCatalogEntry("mcp", "TWO_FINANCE_MCP_URL"),
        ServiceCatalogEntry("planner", "TWO_FINANCE_MCP_URL"),
        ServiceCatalogEntry("tradingcontrol", "TWO_FINANCE_TRADING_CONTROL_URL"),
        ServiceCatalogEntry("matchengine", "TWO_FINANCE_MATCHENGINE_WS_URL"),
        ServiceCatalogEntry("keystore", "TWO_FINANCE_KEYSTORE_URL"),
        ServiceCatalogEntry("hummingbot", "TWO_FINANCE_HUMMINGBOT_URL"),
        ServiceCatalogEntry("wise", "TWO_FINANCE_WISE_URL"),
        ServiceCatalogEntry("airwallex", "TWO_FINANCE_AIRWALLEX_URL"),
    ]
)
