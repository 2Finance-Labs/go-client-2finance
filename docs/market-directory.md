# Market Directory V1

The canonical contract is
`contracts/schemas/market-directory.v1.json`; the golden response is
`contracts/examples/market-directory.json`. JSON Schema is the source of truth
shared by the Network backend, SDK bindings, Exchange frontend, CCXT, and the
native Hummingbot connector.

## Identity and ownership

- `market_id` is the stable logical identity and survives an engine migration.
- `engine_id` owns new commands for the current route.
- `symbol_id` is local to an engine. Its operational key is
  `(engine_id, symbol_id)`, never `symbol_id` alone.
- `route_epoch` is monotonic. Clients must send the epoch used to discover the
  route and re-resolve a rejected stale route.
- One MatchEngine process controls exactly one symbol. The chain contract,
  Market Directory builder, frontend parser, CCXT adapter, Hummingbot connector,
  and MatchEngine admission policy all reject an ambiguous one-to-many route.

## Runtime ownership

The MatchEngine V2 registry on the 2Finance Network is the durable source for
engine ownership and symbol identity. Engine `metadata` stores canonical route
properties (`market_id`, `route_epoch`, chain/address, endpoints, precision,
limits, fees, capabilities, environment, and protocol version). The Exchange
API overlays health and freshness when producing:

- `GET /api/v2/exchange/markets`
- `GET /api/v2/exchange/market-directory`
- `GET /api/v2/exchange/markets/{base}/{quote}`

`MATCH_ENGINE_REGISTRY_JSON` remains accepted only as a local bootstrap overlay.
Deployments should migrate to `MARKET_DIRECTORY_JSON` and ultimately write the
same metadata through the Network registry. Neither variable is a new durable
source of truth.

Recommended engine metadata shape:

```json
{
  "market_id": "BTC/USDT",
  "route_epoch": 7,
  "chain": {"id": "2finance-mainnet", "network": "2finance", "address": "2f1enginebtc"},
  "endpoints": {
    "http": "https://api.2finance.com/api/v2/exchange",
    "websocket": "wss://ws.2finance.com/engines/engine-btc-usdt-01"
  },
  "precision": {"amount": 8, "price": 2, "base": 8, "quote": 6},
  "tick_size": "0.01",
  "step_size": "0.00000001",
  "limits": {
    "amount": {"min": "0.00001", "max": "100"},
    "notional": {"min": "5", "max": "1000000"}
  },
  "fees": {"maker": "0.001", "taker": "0.002", "asset_policy": "received"},
  "capabilities": {
    "order_types": ["LIMIT", "MARKET"],
    "time_in_force": ["GTC", "IOC", "FOK"],
    "channels": ["BOOK", "TRADE", "TICKER"],
    "operations": ["ORDER_SUBMIT", "ORDER_CANCEL"],
    "snapshot": true,
    "recovery": true
  }
}
```

## Compatibility decisions

`/markets` keeps the existing `markets` array and CCXT-compatible `id`, `base`,
`quote`, `precision`, `limits`, `type`, `spot`, and `active` fields. It adds the
versioned directory envelope and route metadata. Decimal values are strings;
precision fields and IDs are integers. `limits.cost` aliases
`limits.notional` for CCXT while `limits.notional` remains the canonical name.

HTTP snapshot/recovery paths may be same-origin relative URLs. The WebSocket
endpoint is a stable public edge URL or alias and must never expose pod or
cluster-internal DNS. A route whose status is not `active` is discoverable for
recovery/cancellation policy but cannot accept new orders.
