# 2finance-sdk-client

Unified client workspace for 2Finance SDKs.

## Layout

- `go/`: primary Go SDK implementation.
- `dart/`: Dart unified SDK plus migrated auth and blockchain clients.
- `typescript/`, `javascript/`, `python/`, `php/`, `java/`, `cpp/`: portable SDK implementations with shared config/auth/domain-client shape and package metadata.
- `contracts/`: shared schemas, fixtures, and compatibility test vectors.
- `docs/`: SDK documentation and spec-driven development notes.

See `docs/implementation-matrix.md` for the current language-by-language
coverage, examples, checks, and maturity notes.

Each SDK exposes public metadata for runtime diagnostics and package reporting:
SDK name `2finance-sdk-client` and version `0.1.0`.
Each SDK also exposes the canonical service catalog so applications can inspect
the supported domains and required `TWO_FINANCE_*` environment variables without
instantiating clients.
Configurations can resolve URLs by domain name, including aliases such as
`match_engine`, through the language equivalent of `serviceUrl`.
They can also list configured service URLs through the language equivalent of
`serviceUrls`.
For typed discovery, SDKs expose configured service entries with `name`, `env`,
and `url` through the language equivalent of `configuredServices`.
SDKs can also list missing service URL configuration, returning the service
`name` and required `env` variable through the language equivalent of
`missingServiceURLs`.
The domain operation catalog is represented by public models and lookup helpers
so consumers can find operations like `analytics.balances` without parsing raw
JSON themselves. Operation models can also resolve concrete request targets from
declared path parameters and query parameters, escaping values and ignoring query
keys that are not part of the contract. Service clients can execute resolved
operations directly through the language equivalent of `requestOperation`; Go
also exposes `CallOperation` and `CallResolvedOperation` on the unified client.
For the shortest contract-backed flow, SDKs can resolve and execute by catalog
domain and operation name through helpers such as `requestCatalogOperation` or
Go's `CallCatalogOperation`.

## Standard environment variables

- `TWO_FINANCE_AUTH_URL`
- `TWO_FINANCE_NETWORK_URL`
- `TWO_FINANCE_ANALYTICS_URL`
- `TWO_FINANCE_ORCHESTRATOR_URL`
- `TWO_FINANCE_MCP_URL`
- `TWO_FINANCE_TRADING_CONTROL_URL`
- `TWO_FINANCE_MATCHENGINE_WS_URL`
- `TWO_FINANCE_KEYSTORE_URL`

## Go quick start

```go
client := twofinance.NewFromEnv()
```

The Go SDK exposes domain clients for auth, network, analytics, orchestrator,
MCP, planner, trading control, matchengine, hummingbot, keystore, and providers.

## Blockchain protocol v2

Go and Dart use one signed transaction shape for both Native Go contracts and
EVM execution. A write returns the early business commit immediately; the same
transaction hash can then be queried for its deterministic execution artifact,
block envelope, ordered membership and 3-validator commit certificate.

```go
prepared, err := protocol.PrepareNativeGoTransactionV2(
	protocol.NativeGoTransactionV2Input{
		ChainID: 2,
		From: walletManager.OwnerAddress(),
		To: deployAddress,
		Method: "deploy_contract",
		Payload: map[string]any{"contract_version": "walletV2"},
	},
)
signed, err := walletManager.SignPreparedTransaction(prepared)
submitted, err := client.Network.SubmitSignedTransactionV2(ctx, signed)
proof, err := client.Network.TransactionFinalityV2(
	ctx,
	submitted.Transaction.TransactionHash,
)
```

Use `protocol.PrepareEVMTransactionV2` with `protocol.EVMMessageV2` to route the
same signed envelope to the EVM. Contract-oriented Go callers can use
`client_2finance.NewV2`; Dart callers use `HttpFinanceNetworkTransport`.

## Request options

The SDK service clients share a per-request options shape for query parameters,
custom headers, pagination, idempotent writes, request timeouts, and basic
retries for 429 or 5xx responses. Use the language equivalent of
`RequestOptions` or, in Go, `service.WithQueryParam("symbol", "BTC-USDT")`,
`service.WithPagination(2, 25)`, `service.WithIdempotencyKey("key")`,
`service.WithTimeout(...)`, and `service.WithMaxRetries(1)` to configure a
single call.

## Language status

- Go: primary implementation with domain clients and commit-first blockchain v2.
- Dart: unified client facade plus migrated auth and blockchain packages.
- TypeScript: typed client surface for auth, network, analytics, orchestrator, MCP, trading control, matchengine, keystore, hummingbot, and providers.
- JavaScript: dependency-free Node 18+ client surface mirroring TypeScript.
- Python: dependency-free client surface using the standard library.
- PHP: dependency-free PHP 8.1+ client surface with injectable transport.
- Java: Java 11+ client surface using `java.net.http`, with Maven metadata.
- C++: header-only client surface with injectable HTTP transport and CMake target.

## Local checks

```bash
make test
```

Live blockchain checks use the HTTP v2 endpoint and remain opt-in.

## Examples

Each implementation includes a minimal quickstart under its language folder:

- `go/examples/quickstart/main.go`
- `dart/sdk/example/quickstart.dart`
- `typescript/examples/quickstart.ts`
- `javascript/examples/quickstart.js`
- `python/examples/quickstart.py`
- `php/examples/quickstart.php`
- `java/examples/Quickstart.java`
- `cpp/examples/quickstart.cpp`

Each language also includes a `request_options` or `request-options` example
showing headers, idempotency keys, query parameters, pagination, timeouts, and
basic retries.

Auth examples show client credentials token sources in Go, Dart, TypeScript,
JavaScript, Python, PHP and Java. C++ shows the equivalent injectable token
source pattern for the header-only client.

Error handling examples show how to catch the SDK HTTP error type and inspect
status code and response body in every language.
