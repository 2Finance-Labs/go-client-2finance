package matchengine

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

type Client struct {
	WebSocketURL string
	HTTPClient   *http.Client
	Dialer       *websocket.Dialer
}

type OrderCommand struct {
	Schema           string          `json:"schema,omitempty"`
	MessageType      string          `json:"message_type,omitempty"`
	Operation        string          `json:"operation,omitempty"`
	OrderType        string          `json:"order_type,omitempty"`
	ClientOrderID    string          `json:"client_order_id"`
	IdempotencyKey   string          `json:"idempotency_key"`
	WalletID         uint64          `json:"wallet_id"`
	SymbolID         uint64          `json:"symbol_id"`
	OrderID          uint64          `json:"order_id,omitempty"`
	Side             string          `json:"side"`
	Quantity         string          `json:"quantity,omitempty"`
	Price            string          `json:"price,omitempty"`
	StopPrice        string          `json:"stop_price,omitempty"`
	Slippage         string          `json:"slippage,omitempty"`
	MaxVisibleQty    string          `json:"max_visible_quantity,omitempty"`
	TrailingDistance string          `json:"trailing_distance,omitempty"`
	TrailingStep     string          `json:"trailing_step,omitempty"`
	TimeInForce      string          `json:"time_in_force,omitempty"`
	AgentID          string          `json:"agent_id,omitempty"`
	ControllerID     string          `json:"controller_id,omitempty"`
	ExecutorID       string          `json:"executor_id,omitempty"`
	StrategyID       string          `json:"strategy_id,omitempty"`
	ClientTimestamp  time.Time       `json:"client_timestamp,omitempty"`
	Metadata         json.RawMessage `json:"metadata,omitempty"`
}

type ExecutionReport struct {
	Schema         string `json:"schema,omitempty"`
	Version        int    `json:"version,omitempty"`
	MessageType    string `json:"message_type,omitempty"`
	Status         string `json:"status,omitempty"`
	ReasonCode     string `json:"reason_code,omitempty"`
	ErrorCode      int    `json:"error_code,omitempty"`
	ClientOrderID  string `json:"client_order_id,omitempty"`
	IdempotencyKey string `json:"idempotency_key,omitempty"`
	OrderID        uint64 `json:"order_id,omitempty"`
	Timestamp      uint64 `json:"timestamp,omitempty"`
}

type MarketDataSubscribeRequest struct {
	Schema    string          `json:"schema,omitempty"`
	Symbols   []string        `json:"symbols,omitempty"`
	Channels  []string        `json:"channels,omitempty"`
	Interval  string          `json:"interval,omitempty"`
	AccountID string          `json:"account_id,omitempty"`
	Metadata  json.RawMessage `json:"metadata,omitempty"`
}

func New(webSocketURL string, httpClient *http.Client) *Client {
	return &Client{
		WebSocketURL: webSocketURL,
		HTTPClient:   httpClient,
		Dialer:       websocket.DefaultDialer,
	}
}

func NewMarketDataSubscribeRequest(request MarketDataSubscribeRequest) MarketDataSubscribeRequest {
	if request.Schema == "" {
		request.Schema = "matchengine.market_data_subscribe.v2"
	}
	return request
}

func NewOrderCommand(command OrderCommand) OrderCommand {
	if command.Schema == "" {
		command.Schema = "matchengine.order_command.v2"
	}
	if command.MessageType == "" {
		command.MessageType = "ORDER"
	}
	if command.Operation == "" {
		command.Operation = "ADD"
	}
	return command
}

func (c *Client) DialOrderEntry(ctx context.Context, headers http.Header) (*websocket.Conn, *http.Response, error) {
	dialer := c.Dialer
	if dialer == nil {
		dialer = websocket.DefaultDialer
	}
	return dialer.DialContext(ctx, c.WebSocketURL, headers)
}

func (c *Client) DialMarketData(ctx context.Context, headers http.Header) (*websocket.Conn, *http.Response, error) {
	return c.DialOrderEntry(ctx, headers)
}

func (c *Client) SubmitOrder(ctx context.Context, command OrderCommand, headers http.Header) (*ExecutionReport, error) {
	conn, _, err := c.DialOrderEntry(ctx, headers)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	command = NewOrderCommand(command)
	if err := conn.WriteJSON(command); err != nil {
		return nil, err
	}
	var report ExecutionReport
	if err := conn.ReadJSON(&report); err != nil {
		return nil, err
	}
	return &report, nil
}

func (c *Client) SubscribeMarketData(ctx context.Context, request MarketDataSubscribeRequest, headers http.Header) (*websocket.Conn, error) {
	conn, _, err := c.DialMarketData(ctx, headers)
	if err != nil {
		return nil, err
	}
	request = NewMarketDataSubscribeRequest(request)
	if err := conn.WriteJSON(request); err != nil {
		_ = conn.Close()
		return nil, err
	}
	return conn, nil
}
