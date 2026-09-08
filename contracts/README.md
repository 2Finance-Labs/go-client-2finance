# Contracts

Shared schemas and fixtures used to keep Go, Dart, TypeScript, JavaScript,
Python, PHP, Java, and C++ SDK behavior aligned.

- `schemas/domain-operations.v2.json`: canonical operation map for public SDK
  domains, transports, methods, paths, path params, and contract names.
- `schemas/request-options.v2.json`: shared shape for per-call headers,
  idempotency, query params, pagination, timeout, and retry options.
- `schemas/error.v2.json`: common error response shape used by service clients.
- `schemas/pagination.v2.json`: common cursor/limit pagination shape.
- `schemas/idempotency.v2.json`: common idempotency-key fixture shape.
- `schemas/service-catalog.v2.json`: shared SDK domain-to-environment-variable
  catalog shape.
- `schemas/market-directory.v1.json`: canonical market discovery contract,
  including the stable market identity, one-engine/one-symbol route, chain
  address, HTTP/WebSocket endpoints, rules, fees, capabilities, and freshness.
- `schemas/trading-session-authorization.v1.json`,
  `schemas/matchengine-order-command.v3.json`, and
  `schemas/trading-session-revocation.v1.json`: canonical Ed25519 authorization,
  command, cancellation, cancel-all, and revocation envelopes. See
  `../docs/order-authorization.md` for canonical bytes and replay semantics.
- `examples/domain-operations.json`: fixture catalog for auth, network,
  analytics, orchestrator, MCP, planner, trading control, matchengine,
  hummingbot, keystore, and providers. Each language SDK exposes public models
  for this fixture.
- `examples/request-options.json`: canonical request-options fixture used by
  SDK tests and structure validation.
- `examples/error.json`, `examples/pagination.json`, and
  `examples/idempotency.json`: minimal test vectors for cross-language parsing.
- `examples/service-catalog.json`: canonical SDK domain-to-environment-variable
  catalog exposed by each language implementation.
- `examples/market-directory.json`: golden Market Directory response consumed
  unchanged by the Exchange API, frontend, CCXT, and Hummingbot contract tests.
- `examples/matchengine-order-command-v3.json`: cross-language golden Ed25519
  command and signature vector for Dart, Go, and C++.
