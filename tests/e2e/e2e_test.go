package e2e_test

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"testing"
	"time"

	client2f "github.com/2Finance-Labs/go-client-2finance/client_2finance"
	"gitlab.com/2finance/2finance-network/config"
)

// ----------------------------------------------------------------------------
// Helpers
// ----------------------------------------------------------------------------

func setupClient(t *testing.T) client2f.Client2FinanceNetwork {
	t.Helper()
	env := os.Getenv("APP_ENV")
	if env == "" { env = "prod" }
	config.Load_config(env, "./../../.env")

	emqxHost := fmt.Sprintf("%s://%s:%s", config.EMQX_SCHEME, config.EMQX_HOST, config.EMQX_PORT)

	// <<< important: make the client id unique per test run >>>
	base := config.EMQX_CLIENT_ID
	if base == "" { base = "e2e" }
	id := fmt.Sprintf("%s-%s", base, randSuffix(8))

	c := client2f.New(emqxHost, id, false)
	return c
}

func randSuffix(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)[:n]
}

// unmarshalState decodes an arbitrary state object into a typed struct.
func unmarshalState[T any](t *testing.T, obj any, out *T) {
	t.Helper()
	by, err := json.Marshal(obj)
	if err != nil {
		t.Fatalf("marshal state: %v", err)
	}
	if err := json.Unmarshal(by, out); err != nil {
		t.Fatalf("unmarshal state: %v", err)
	}
}

// amt builds integer string respecting decimals (unscaled * 10^decimals)
func amt(unscaled int64, decimals int) string {
	p := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimals)), nil)
	v := new(big.Int).Mul(big.NewInt(unscaled), p)
	return v.String()
}

func waitUntil(t *testing.T, d time.Duration, pred func() bool) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), d)
	defer cancel()
	tick := time.NewTicker(20 * time.Millisecond)
	defer tick.Stop()
	for {
		select {
		case <-ctx.Done():
			t.Fatalf("timeout waiting for condition")
		case <-tick.C:
			if pred() { return }
		}
	}
}

func genKey(t *testing.T, c client2f.Client2FinanceNetwork) (pub, priv string) {
	pub, priv, err := c.GenerateKeyEd25519()
	if err != nil { t.Fatalf("GenerateKeyEd25519: %v", err) }
	return
}


