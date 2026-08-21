package network

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/2Finance-Labs/2finance-sdk-client/internal/service"
)

type Client struct {
	service *service.Client
}

func New(baseURL string, httpClient *http.Client) *Client {
	return &Client{service: service.New(baseURL, httpClient)}
}

func (c *Client) Get(ctx context.Context, path string, out any) error {
	return c.service.DoJSON(ctx, http.MethodGet, path, nil, out)
}

func (c *Client) Post(ctx context.Context, path string, body any, out any) error {
	return c.service.DoJSON(ctx, http.MethodPost, path, body, out)
}

func (c *Client) VirtualMachine(ctx context.Context) (json.RawMessage, error) {
	var response json.RawMessage
	err := c.Post(ctx, "/v2/2finance-network/query", map[string]any{
		"method": "get_blocks",
		"params": map[string]any{"page": 1, "limit": 1},
	}, &response)
	return response, err
}

func (c *Client) MarketCandles(ctx context.Context, market string, query string) (json.RawMessage, error) {
	var response json.RawMessage
	path := "/v2/2finance-network/markets/" + url.PathEscape(market) + "/candles"
	if query != "" {
		path += "?" + query
	}
	err := c.Get(ctx, path, &response)
	return response, err
}

func (c *Client) CreateProduct(ctx context.Context, productType string, request any) (json.RawMessage, error) {
	var response json.RawMessage
	err := c.Post(ctx, "/v2/2finance-network/products/"+url.PathEscape(productType), request, &response)
	return response, err
}

func (c *Client) Products(ctx context.Context, productType string) (json.RawMessage, error) {
	var response json.RawMessage
	err := c.Get(ctx, "/v2/2finance-network/products/"+url.PathEscape(productType), &response)
	return response, err
}

func (c *Client) Bonds(ctx context.Context) (json.RawMessage, error) {
	return c.Products(ctx, "bonds")
}

func (c *Client) CreateBond(ctx context.Context, request any) (json.RawMessage, error) {
	return c.CreateProduct(ctx, "bonds", request)
}

func (c *Client) Loans(ctx context.Context) (json.RawMessage, error) {
	return c.Products(ctx, "loans")
}

func (c *Client) CreateLoan(ctx context.Context, request any) (json.RawMessage, error) {
	return c.CreateProduct(ctx, "loans", request)
}

func (c *Client) Swaps(ctx context.Context) (json.RawMessage, error) {
	return c.Products(ctx, "swaps")
}

func (c *Client) CreateSwap(ctx context.Context, request any) (json.RawMessage, error) {
	return c.CreateProduct(ctx, "swaps", request)
}

func (c *Client) StakingProducts(ctx context.Context) (json.RawMessage, error) {
	return c.Products(ctx, "staking")
}

func (c *Client) CreateStakingProduct(ctx context.Context, request any) (json.RawMessage, error) {
	return c.CreateProduct(ctx, "staking", request)
}

func (c *Client) SyntheticAssets(ctx context.Context) (json.RawMessage, error) {
	return c.Products(ctx, "synthetic-assets")
}

func (c *Client) CreateSyntheticAsset(ctx context.Context, request any) (json.RawMessage, error) {
	return c.CreateProduct(ctx, "synthetic-assets", request)
}

func (c *Client) LiquidityPools(ctx context.Context) (json.RawMessage, error) {
	return c.Products(ctx, "liquidity-pools")
}

func (c *Client) CreateLiquidityPool(ctx context.Context, request any) (json.RawMessage, error) {
	return c.CreateProduct(ctx, "liquidity-pools", request)
}
