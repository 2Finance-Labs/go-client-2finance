# Canonical order authorization

`matchengine.order_command.v3` is the only wallet/session-key signed order
envelope. The permanent Wallet Mobile key may sign a command directly with
`authorization.mode=WALLET`; normal interactive trading uses a separately
approved, short-lived `SESSION` key.

## Signature input

The signature is Ed25519 over these exact UTF-8 bytes, where `payload` is JSON
with object keys sorted lexicographically at every level, no insignificant
whitespace, integers in base 10, and strings escaped as JSON:

```text
2FINANCE-SIGNED-MESSAGE-V1\n<domain>\n<canonical-payload-json>
```

`contracts/examples/matchengine-order-command-v3.json` contains the valid
golden vector also verified by the Dart, Go, and C++ test suites.

Arrays in `TradingSessionAuthorization` (`allowed_symbols`,
`allowed_actions`, `allowed_order_types`) MUST be sorted, contain no duplicates,
and are therefore deterministic. Monetary fields are minimal unsigned decimal
strings: no sign, exponent, leading zero, or trailing fractional zero. Zero is
`"0"`. Command nonces are positive canonical uint64 strings so JavaScript
cannot round them; session approval and revocation nonces are random 256-bit
lowercase hex strings. Expiration is Unix epoch milliseconds.

The domain MUST match the operation:

| Operation | Domain | Permission |
|---|---|---|
| `ADD` | `2finance.order.add.v1` | `order:submit` |
| `CANCEL` | `2finance.order.cancel.v1` | `order:cancel` |
| `CANCEL_ALL` | `2finance.order.cancel_all.v1` | `order:cancel_all` |
| Session approval | `2finance.trading_session.v1` | Wallet only |
| Session revocation | `2finance.trading_session.revoke.v1` | Wallet only |

An ADD requires BUY/SELL, LIMIT/MARKET, a positive amount, and a positive price
for LIMIT (`"0"` for MARKET). CANCEL and CANCEL_ALL use `NONE` for side/type/TIF
and `"0"` for amount/price. CANCEL requires `order_id`; CANCEL_ALL forbids it.

## Binding and replay rules

The verifier compares every signed route field with the authenticated ticket:
`chain_id`, `account`, `engine_id`, `symbol`, `symbol_id`, and `route_epoch`.
`account` is the lowercase 32-byte Ed25519 public key of the Wallet Mobile root
identity, not an engine-local numeric wallet ID.
The engine-local wallet ID is derived from the ticket and is never accepted
from the client. A session signature must also match `session_id` and
`session_public_key`, its allowed symbols/actions/order types and amount/notional
limits, and both ticket/session expiry.

Before queue admission, the MatchEngine atomically consumes
`(authorization public key, domain, nonce)` and
`(account, idempotency_key)`. A duplicate is rejected. The Network stores
session-approval/revocation nonces durably; MatchEngine replay state must be
checkpointed or rebuilt from its command journal before accepting traffic after
a restart. Revocation makes the session unusable immediately and is published
to every engine named by the session authorization.

`WALLET` is intended for explicit, high-assurance user actions. `SESSION` is the
normal order hot path and can never grant withdrawal or administrative scopes.
Wallet Mobile requires an explicit user-presence/biometric result before every
root-key approval, direct order, or revocation, and release builds refuse a
private key supplied through runtime configuration.
Increasing symbols, actions, limits, or expiry requires a new Wallet Mobile
signature. The Network currently caps a session authorization at 24 hours;
deployments may apply a shorter policy.

For revocation, the Exchange submits the same Wallet-signed envelope to the
Network (durable status and replay nonce) and to every active MatchEngine
connection (immediate in-memory kill switch). Short-lived tickets must be
reissued only while the Network status remains active.
