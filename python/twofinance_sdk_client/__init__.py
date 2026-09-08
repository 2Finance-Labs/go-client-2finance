"""Unified Python client for 2Finance services."""

SDK_NAME = "2finance-sdk-client"
__version__ = "0.1.0"

from .auth import ClientCredentialsTokenSource, StaticTokenSource, bearer_authorization
from .client import ProvidersClient, TwoFinanceClient
from .config import SDKConfig, config_from_env
from .domains import AirwallexClient, ProviderClient, WiseClient
from .models import (
    DEFAULT_SERVICE_CATALOG,
    ConfiguredServiceEntry,
    DomainOperation,
    DomainOperationsCatalog,
    DomainOperationsDomain,
    IdempotencyRecord,
    MarketDefinition,
    MarketDirectory,
    PaginationResponse,
    ResolvedOperation,
    SDKError,
    ServiceCatalog,
    ServiceCatalogEntry,
)
from .service import HTTPError, RequestOptions, ServiceClient

__all__ = [
    "AirwallexClient",
    "ClientCredentialsTokenSource",
    "ConfiguredServiceEntry",
    "DEFAULT_SERVICE_CATALOG",
    "DomainOperation",
    "DomainOperationsCatalog",
    "DomainOperationsDomain",
    "HTTPError",
    "IdempotencyRecord",
    "MarketDefinition",
    "MarketDirectory",
    "PaginationResponse",
    "SDKConfig",
    "SDKError",
    "RequestOptions",
    "ResolvedOperation",
    "SDK_NAME",
    "ServiceClient",
    "ServiceCatalog",
    "ServiceCatalogEntry",
    "StaticTokenSource",
    "TwoFinanceClient",
    "ProviderClient",
    "ProvidersClient",
    "bearer_authorization",
    "config_from_env",
    "WiseClient",
    "__version__",
]
