package clobclient

// OrderType constants
const (
	OrderTypeGTC = "GTC"
	OrderTypeFOK = "FOK"
	OrderTypeGTD = "GTD"
	OrderTypeFAK = "FAK"
)

// Side represents the order side as an integer.
type Side int

const (
	SideBuy  Side = 0
	SideSell Side = 1
)

// SignatureTypeV1 represents V1 order signature types.
type SignatureTypeV1 int

const (
	SignatureTypeV1EOA        SignatureTypeV1 = 0
	SignatureTypeV1PolyProxy  SignatureTypeV1 = 1
	SignatureTypeV1PolyGnosis SignatureTypeV1 = 2
)

// SignatureTypeV2 represents V2 order signature types.
type SignatureTypeV2 int

const (
	SignatureTypeV2EOA        SignatureTypeV2 = 0
	SignatureTypeV2PolyProxy  SignatureTypeV2 = 1
	SignatureTypeV2PolyGnosis SignatureTypeV2 = 2
	SignatureTypeV2Poly1271   SignatureTypeV2 = 3
)

// ApiCreds holds API credentials for L2 authentication.
type ApiCreds struct {
	APIKey        string `json:"apiKey"`
	APISecret     string `json:"secret"`
	APIPassphrase string `json:"passphrase"`
}

// RequestArgs contains the arguments for building L2 auth headers.
type RequestArgs struct {
	Method         string
	RequestPath    string
	Body           interface{}
	SerializedBody string
}

// BookParams are parameters for order book queries.
type BookParams struct {
	TokenID string `json:"token_id"`
	Side    string `json:"side,omitempty"`
}

// OrderArgs (V2 default) is the input for creating a V2 limit order.
type OrderArgs struct {
	TokenID         string  `json:"token_id"`
	Price           float64 `json:"price"`
	Size            float64 `json:"size"`
	Side            string  `json:"side"`
	Expiration      int     `json:"expiration,omitempty"`
	BuilderCode     string  `json:"builder_code,omitempty"`
	Metadata        string  `json:"metadata,omitempty"`
	UserUSDCBalance *float64 `json:"user_usdc_balance,omitempty"`
}

// OrderArgsV1 is the input for creating a V1 (legacy) limit order.
type OrderArgsV1 struct {
	TokenID     string  `json:"token_id"`
	Price       float64 `json:"price"`
	Size        float64 `json:"size"`
	Side        string  `json:"side"`
	Expiration  int     `json:"expiration,omitempty"`
	FeeRateBps  int     `json:"fee_rate_bps,omitempty"`
	Nonce       int     `json:"nonce,omitempty"`
	Taker       string  `json:"taker,omitempty"`
	BuilderCode string  `json:"builder_code,omitempty"`
}

// MarketOrderArgs (V2 default) is the input for creating a V2 market order.
type MarketOrderArgs struct {
	TokenID         string  `json:"token_id"`
	Amount          float64 `json:"amount"`
	Side            string  `json:"side"`
	Price           float64 `json:"price,omitempty"`
	OrderType       string  `json:"order_type,omitempty"`
	UserUSDCBalance float64 `json:"user_usdc_balance,omitempty"`
	BuilderCode     string  `json:"builder_code,omitempty"`
	Metadata        string  `json:"metadata,omitempty"`
}

// MarketOrderArgsV1 is the input for creating a V1 (legacy) market order.
type MarketOrderArgsV1 struct {
	TokenID     string  `json:"token_id"`
	Amount      float64 `json:"amount"`
	Side        string  `json:"side"`
	Price       float64 `json:"price,omitempty"`
	OrderType   string  `json:"order_type,omitempty"`
	FeeRateBps  int     `json:"fee_rate_bps,omitempty"`
	Nonce       int     `json:"nonce,omitempty"`
	Taker       string  `json:"taker,omitempty"`
	BuilderCode string  `json:"builder_code,omitempty"`
}

// TradeParams holds trade query parameters.
type TradeParams struct {
	ID           string `json:"id,omitempty"`
	MakerAddress string `json:"maker_address,omitempty"`
	Market       string `json:"market,omitempty"`
	AssetID      string `json:"asset_id,omitempty"`
	Before       int    `json:"before,omitempty"`
	After        int    `json:"after,omitempty"`
}

// OpenOrderParams holds open order query parameters.
type OpenOrderParams struct {
	ID      string `json:"id,omitempty"`
	Market  string `json:"market,omitempty"`
	AssetID string `json:"asset_id,omitempty"`
}

// DropNotificationParams holds notification drop parameters.
type DropNotificationParams struct {
	IDs []string `json:"ids,omitempty"`
}

// OrderSummary represents a single price level in the order book.
type OrderSummary struct {
	Price string `json:"price,omitempty"`
	Size  string `json:"size,omitempty"`
}

// OrderBookSummary represents the full order book.
type OrderBookSummary struct {
	Market         string         `json:"market,omitempty"`
	AssetID        string         `json:"asset_id,omitempty"`
	Timestamp      string         `json:"timestamp,omitempty"`
	Bids           []OrderSummary `json:"bids,omitempty"`
	Asks           []OrderSummary `json:"asks,omitempty"`
	MinOrderSize   string         `json:"min_order_size,omitempty"`
	NegRisk        *bool          `json:"neg_risk,omitempty"`
	TickSize       string         `json:"tick_size,omitempty"`
	LastTradePrice string         `json:"last_trade_price,omitempty"`
	Hash           string         `json:"hash,omitempty"`
}

// AssetType constants
const (
	AssetTypeCollateral  = "COLLATERAL"
	AssetTypeConditional = "CONDITIONAL"
)

// BalanceAllowanceParams holds balance/allowance query parameters.
type BalanceAllowanceParams struct {
	AssetType     string `json:"asset_type,omitempty"`
	TokenID       string `json:"token_id,omitempty"`
	SignatureType int    `json:"signature_type,omitempty"`
}

// OrderScoringParams holds order scoring query parameters.
type OrderScoringParams struct {
	OrderID string `json:"orderId"`
}

// OrdersScoringParams holds batch order scoring parameters.
type OrdersScoringParams struct {
	OrderIDs []string `json:"orderIds"`
}

// OrderPayload holds an order ID for cancellation.
type OrderPayload struct {
	OrderID string `json:"orderID"`
}

// TickSize is the string representation of tick size.
type TickSize = string

// CreateOrderOptions holds required options for creating an order.
type CreateOrderOptions struct {
	TickSize TickSize `json:"tick_size"`
	NegRisk  bool     `json:"neg_risk"`
}

// PartialCreateOrderOptions holds optional overrides for creating an order.
type PartialCreateOrderOptions struct {
	TickSize *string `json:"tick_size,omitempty"`
	NegRisk  *bool   `json:"neg_risk,omitempty"`
}

// RoundConfig holds rounding configurations per tick size.
type RoundConfig struct {
	Price  float64
	Size   float64
	Amount float64
}

// BuilderConfig holds builder configuration for fee attribution.
type BuilderConfig struct {
	BuilderAddress string `json:"builder_address,omitempty"`
	BuilderCode    string `json:"builder_code,omitempty"`
}

// FeeInfo holds fee details for a market.
type FeeInfo struct {
	Rate     float64 `json:"rate"`
	Exponent float64 `json:"exponent"`
}

// BuilderFeeRate holds maker/taker fee rates for a builder.
type BuilderFeeRate struct {
	Maker float64 `json:"maker"`
	Taker float64 `json:"taker"`
}

// OrderMarketCancelParams holds parameters for canceling all orders in a market.
type OrderMarketCancelParams struct {
	Market  string `json:"market,omitempty"`
	AssetID string `json:"asset_id,omitempty"`
}

// PriceHistoryInterval constants
const (
	PriceHistoryIntervalMax     = "max"
	PriceHistoryIntervalOneWeek = "1w"
	PriceHistoryIntervalOneDay  = "1d"
	PriceHistoryIntervalSixHour = "6h"
	PriceHistoryIntervalOneHour = "1h"
)

// PricesHistoryParams holds price history query parameters.
type PricesHistoryParams struct {
	Market   string `json:"market,omitempty"`
	StartTS  *int   `json:"start_ts,omitempty"`
	EndTS    *int   `json:"end_ts,omitempty"`
	Fidelity *int   `json:"fidelity,omitempty"`
	Interval string `json:"interval,omitempty"`
}

// EarningsParams holds earnings query parameters.
type EarningsParams struct {
	Date   string `json:"date,omitempty"`
	Market string `json:"market,omitempty"`
}

// BuilderTradeParams holds builder trade query parameters.
type BuilderTradeParams struct {
	BuilderCode  string `json:"builder_code"`
	ID           string `json:"id,omitempty"`
	MakerAddress string `json:"maker_address,omitempty"`
	Market       string `json:"market,omitempty"`
	AssetID      string `json:"asset_id,omitempty"`
	Before       string `json:"before,omitempty"`
	After        string `json:"after,omitempty"`
}

// PostOrdersArgs wraps a signed order with its order type and flags.
type PostOrdersArgs struct {
	Order     interface{} `json:"order"`
	OrderType string      `json:"orderType"`
	DeferExec bool        `json:"deferExec"`
}
