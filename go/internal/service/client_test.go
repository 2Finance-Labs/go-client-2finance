package service

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

type requestOptionsFixture struct {
	Request struct {
		BaseURL        string            `json:"base_url"`
		Path           string            `json:"path"`
		Headers        map[string]string `json:"headers"`
		IdempotencyKey string            `json:"idempotency_key"`
		Query          map[string]string `json:"query"`
		Pagination     struct {
			Page  int `json:"page"`
			Limit int `json:"limit"`
		} `json:"pagination"`
		TimeoutMS  int `json:"timeout_ms"`
		MaxRetries int `json:"max_retries"`
	} `json:"request"`
	Expected struct {
		URL     string            `json:"url"`
		Headers map[string]string `json:"headers"`
	} `json:"expected"`
}

type domainOperationsFixture struct {
	Schema  string `json:"schema"`
	Domains []struct {
		Name       string `json:"name"`
		Transport  string `json:"transport"`
		Operations []struct {
			Name          string   `json:"name"`
			Path          string   `json:"path"`
			PathParams    []string `json:"path_params"`
			RequestSchema string   `json:"request_schema"`
		} `json:"operations"`
	} `json:"domains"`
}

func TestSharedContractFixturesDescribePublicSDKOperations(t *testing.T) {
	fixture := loadDomainOperationsFixture(t)
	if fixture.Schema != "sdk.domain_operations.v2" {
		t.Fatalf("schema = %q", fixture.Schema)
	}
	balances := findContractOperation(t, fixture, "analytics", "balances")
	if balances.Path != "/portfolio-manager/balances/{account_id}" {
		t.Fatalf("analytics balances path = %q", balances.Path)
	}
	if len(balances.PathParams) != 1 || balances.PathParams[0] != "account_id" {
		t.Fatalf("analytics balances path params = %#v", balances.PathParams)
	}
	tradingPlan := findContractOperation(t, fixture, "planner", "trading_plan")
	if tradingPlan.RequestSchema != "planner.trading_plan.request.v1" {
		t.Fatalf("planner trading plan request schema = %q", tradingPlan.RequestSchema)
	}
	assertContractFixtureContains(t, "error.json", `"error": "rate_limited"`)
	assertContractFixtureContains(t, "error.json", `"code": "HTTP_429"`)
	assertContractFixtureContains(t, "pagination.json", `"next_cursor": "cursor-next"`)
	assertContractFixtureContains(t, "idempotency.json", `"idempotency_key": "idem-001"`)
}

func TestDoJSONWithOptionsAddsHeadersAndIdempotencyKey(t *testing.T) {
	fixture := loadRequestOptionsFixture(t)
	var gotTrace string
	var gotIdempotency string
	var gotSymbol string
	var gotPage string
	var gotLimit string
	var gotDeadline bool
	var gotURL string
	client := New(fixture.Request.BaseURL, &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			gotURL = req.URL.String()
			gotTrace = req.Header.Get("X-Trace-ID")
			gotIdempotency = req.Header.Get("Idempotency-Key")
			gotSymbol = req.URL.Query().Get("symbol")
			gotPage = req.URL.Query().Get("page")
			gotLimit = req.URL.Query().Get("limit")
			_, gotDeadline = req.Context().Deadline()
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     make(http.Header),
				Body:       http.NoBody,
				Request:    req,
			}, nil
		}),
	})

	var out json.RawMessage
	if err := client.DoJSONWithOptions(
		context.Background(),
		http.MethodPost,
		fixture.Request.Path,
		map[string]string{"symbol": "BTC-USDT"},
		&out,
		WithHeader("X-Trace-ID", fixture.Request.Headers["X-Trace-ID"]),
		WithIdempotencyKey(" "+fixture.Request.IdempotencyKey+" "),
		WithQueryParam("symbol", fixture.Request.Query["symbol"]),
		WithPagination(fixture.Request.Pagination.Page, fixture.Request.Pagination.Limit),
		WithTimeout(time.Duration(fixture.Request.TimeoutMS)*time.Millisecond),
	); err != nil {
		t.Fatal(err)
	}
	if !sameURLValues(gotURL, fixture.Expected.URL) {
		t.Fatalf("URL = %q", gotURL)
	}
	if gotTrace != fixture.Expected.Headers["X-Trace-ID"] {
		t.Fatalf("X-Trace-ID = %q", gotTrace)
	}
	if gotIdempotency != fixture.Expected.Headers["Idempotency-Key"] {
		t.Fatalf("Idempotency-Key = %q", gotIdempotency)
	}
	if gotSymbol != fixture.Request.Query["symbol"] {
		t.Fatalf("symbol query = %q", gotSymbol)
	}
	if gotPage != "2" {
		t.Fatalf("page query = %q", gotPage)
	}
	if gotLimit != "25" {
		t.Fatalf("limit query = %q", gotLimit)
	}
	if !gotDeadline {
		t.Fatalf("expected request context deadline")
	}
}

func loadDomainOperationsFixture(t *testing.T) domainOperationsFixture {
	t.Helper()
	path := filepath.Join("..", "..", "..", "contracts", "examples", "domain-operations.json")
	payload, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var fixture domainOperationsFixture
	if err := json.Unmarshal(payload, &fixture); err != nil {
		t.Fatal(err)
	}
	return fixture
}

func findContractOperation(t *testing.T, fixture domainOperationsFixture, domainName, operationName string) struct {
	Name          string   `json:"name"`
	Path          string   `json:"path"`
	PathParams    []string `json:"path_params"`
	RequestSchema string   `json:"request_schema"`
} {
	t.Helper()
	for _, domain := range fixture.Domains {
		if domain.Name != domainName {
			continue
		}
		for _, operation := range domain.Operations {
			if operation.Name == operationName {
				return operation
			}
		}
	}
	t.Fatalf("missing contract operation %s.%s", domainName, operationName)
	return struct {
		Name          string   `json:"name"`
		Path          string   `json:"path"`
		PathParams    []string `json:"path_params"`
		RequestSchema string   `json:"request_schema"`
	}{}
}

func assertContractFixtureContains(t *testing.T, name, expected string) {
	t.Helper()
	path := filepath.Join("..", "..", "..", "contracts", "examples", name)
	payload, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(payload), expected) {
		t.Fatalf("%s does not contain %s", name, expected)
	}
}

func loadRequestOptionsFixture(t *testing.T) requestOptionsFixture {
	t.Helper()
	path := filepath.Join("..", "..", "..", "contracts", "examples", "request-options.json")
	payload, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var fixture requestOptionsFixture
	if err := json.Unmarshal(payload, &fixture); err != nil {
		t.Fatal(err)
	}
	return fixture
}

func sameURLValues(left, right string) bool {
	leftURL, err := url.Parse(left)
	if err != nil {
		return false
	}
	rightURL, err := url.Parse(right)
	if err != nil {
		return false
	}
	if leftURL.Scheme != rightURL.Scheme || leftURL.Host != rightURL.Host || leftURL.Path != rightURL.Path {
		return false
	}
	leftQuery := leftURL.Query()
	rightQuery := rightURL.Query()
	if len(leftQuery) != len(rightQuery) {
		return false
	}
	for key, values := range rightQuery {
		if strings.Join(leftQuery[key], ",") != strings.Join(values, ",") {
			return false
		}
	}
	return true
}

func TestDoJSONWithOptionsRetriesRetryableStatus(t *testing.T) {
	attempts := 0
	client := New("https://analytics.example", &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			attempts++
			if attempts == 1 {
				return &http.Response{
					StatusCode: http.StatusInternalServerError,
					Header:     make(http.Header),
					Body:       io.NopCloser(strings.NewReader("temporary")),
					Request:    req,
				}, nil
			}
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     make(http.Header),
				Body:       io.NopCloser(strings.NewReader(`{"ok":true}`)),
				Request:    req,
			}, nil
		}),
	})

	var out map[string]bool
	if err := client.DoJSONWithOptions(
		context.Background(),
		http.MethodGet,
		"/analytics/indicators",
		nil,
		&out,
		WithMaxRetries(1),
	); err != nil {
		t.Fatal(err)
	}
	if attempts != 2 {
		t.Fatalf("attempts = %d", attempts)
	}
	if !out["ok"] {
		t.Fatalf("expected decoded retry response")
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}
