package clobclient

// MatchType represents RFQ match types.
type MatchType string

const (
	MatchTypeComplementary MatchType = "COMPLEMENTARY"
	MatchTypeMint          MatchType = "MINT"
	MatchTypeMerge         MatchType = "MERGE"
)

// RFQUserRequest is simplified user input for creating an RFQ request.
type RFQUserRequest struct {
	TokenID string  `json:"token_id"`
	Price   float64 `json:"price"`
	Side    string  `json:"side"`
	Size    float64 `json:"size"`
}

// RFQUserQuote is simplified user input for creating an RFQ quote.
type RFQUserQuote struct {
	RequestID string  `json:"request_id"`
	TokenID   string  `json:"token_id"`
	Price     float64 `json:"price"`
	Side      string  `json:"side"`
	Size      float64 `json:"size"`
}

// CancelRFQRequestParams holds parameters for canceling an RFQ request.
type CancelRFQRequestParams struct {
	RequestID string `json:"request_id"`
}

// CancelRFQQuoteParams holds parameters for canceling an RFQ quote.
type CancelRFQQuoteParams struct {
	QuoteID string `json:"quote_id"`
}

// AcceptQuoteParams holds parameters for accepting a quote (requester side).
type AcceptQuoteParams struct {
	RequestID  string `json:"request_id"`
	QuoteID    string `json:"quote_id"`
	Expiration int    `json:"expiration"`
}

// ApproveOrderParams holds parameters for approving an order (quoter side).
type ApproveOrderParams struct {
	RequestID  string `json:"request_id"`
	QuoteID    string `json:"quote_id"`
	Expiration int    `json:"expiration"`
}

// GetRFQRequestsParams holds query parameters for fetching RFQ requests.
type GetRFQRequestsParams struct {
	RequestIDs []string `json:"request_ids,omitempty"`
	State      string   `json:"state,omitempty"`
	Markets    []string `json:"markets,omitempty"`
	SizeMin    *float64 `json:"size_min,omitempty"`
	SizeMax    *float64 `json:"size_max,omitempty"`
	SizeUSDCMin *float64 `json:"size_usdc_min,omitempty"`
	SizeUSDCMax *float64 `json:"size_usdc_max,omitempty"`
	PriceMin   *float64 `json:"price_min,omitempty"`
	PriceMax   *float64 `json:"price_max,omitempty"`
	SortBy     string   `json:"sort_by,omitempty"`
	SortDir    string   `json:"sort_dir,omitempty"`
	Limit      *int     `json:"limit,omitempty"`
	Offset     string   `json:"offset,omitempty"`
}

// GetRFQQuotesParams holds query parameters for fetching RFQ quotes.
type GetRFQQuotesParams struct {
	QuoteIDs   []string `json:"quote_ids,omitempty"`
	RequestIDs []string `json:"request_ids,omitempty"`
	State      string   `json:"state,omitempty"`
	Markets    []string `json:"markets,omitempty"`
	SizeMin    *float64 `json:"size_min,omitempty"`
	SizeMax    *float64 `json:"size_max,omitempty"`
	SizeUSDCMin *float64 `json:"size_usdc_min,omitempty"`
	SizeUSDCMax *float64 `json:"size_usdc_max,omitempty"`
	PriceMin   *float64 `json:"price_min,omitempty"`
	PriceMax   *float64 `json:"price_max,omitempty"`
	SortBy     string   `json:"sort_by,omitempty"`
	SortDir    string   `json:"sort_dir,omitempty"`
	Limit      *int     `json:"limit,omitempty"`
	Offset     string   `json:"offset,omitempty"`
}

// GetRFQBestQuoteParams holds parameters for fetching the best quote.
type GetRFQBestQuoteParams struct {
	RequestID string `json:"request_id,omitempty"`
}
