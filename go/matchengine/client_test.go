package matchengine

import (
	"encoding/json"
	"testing"
)

func TestNewOrderCommandBuildsMatchEngineV2WireShape(t *testing.T) {
	command := NewOrderCommand(OrderCommand{
		OrderType:      "LIMIT",
		ClientOrderID:  "go-sdk-1",
		IdempotencyKey: "go-sdk-1-add",
		WalletID:       1,
		SymbolID:       1,
		Side:           "BUY",
		Quantity:       "0.010000",
		Price:          "50000.000000",
	})

	encoded, err := json.Marshal(command)
	if err != nil {
		t.Fatal(err)
	}
	var wire map[string]any
	if err := json.Unmarshal(encoded, &wire); err != nil {
		t.Fatal(err)
	}
	for field, expected := range map[string]any{
		"schema":          "matchengine.order_command.v2",
		"message_type":    "ORDER",
		"operation":       "ADD",
		"order_type":      "LIMIT",
		"client_order_id": "go-sdk-1",
		"idempotency_key": "go-sdk-1-add",
		"wallet_id":       float64(1),
		"symbol_id":       float64(1),
		"side":            "BUY",
		"quantity":        "0.010000",
		"price":           "50000.000000",
	} {
		if wire[field] != expected {
			t.Fatalf("unexpected %s: got %#v want %#v; payload=%s", field, wire[field], expected, encoded)
		}
	}
}
