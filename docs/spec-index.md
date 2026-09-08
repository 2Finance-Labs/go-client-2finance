# Spec index

This index groups behavior specs by the boundary they protect.

## Exchange

- `market-directory.md`: canonical market identity, one-engine/one-symbol
  routing, trading rules, endpoints, and capabilities.

## SDK implementation

- `implementation-matrix.md`: language-by-language SDK implementation status,
  examples, checks, and maturity notes.

## Protocol

- `protocol.prepared-transaction.signing-contract`
- `protocol.signed-transaction.network-roundtrip`

## Wallet manager

- `wallet-manager.unlock-and-sign-prepared-transaction`

## Client lifecycle

- `lifecycle.send.client-direct-start`
- `lifecycle.send.mcp-sign-submit`

## MCP

- `mcp.finance-tools.registration-contract`

Run all spec validation with:

```bash
make test-harness
```
