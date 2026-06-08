package clobclient

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"
)

// ClobClient is the main client for the Polymarket CLOB API.
type ClobClient struct {
	Host           string
	ChainID        int
	UseServerTime  bool
	RetryOnError   bool
	BuilderCfg     *BuilderConfig
	FeeSlippage    float64

	Signer  *Signer
	Creds   *ApiCreds
	Mode    int
	Builder *OrderBuilder

	// RFQ sub-client (set after construction)
	RFQ *RFQClient

	// Caches
	tickSizes         map[string]string
	negRisk           map[string]bool
	feeRates          map[string]int
	feeInfos          map[string]*FeeInfo
	builderFeeRates   map[string]*BuilderFeeRate
	tokenConditionMap map[string]string
	cachedVersion     *int
}

// ClobClientConfig holds all configuration for creating a ClobClient.
type ClobClientConfig struct {
	Host          string
	ChainID       int
	Key           string
	Creds         *ApiCreds
	SignatureType *SignatureTypeV2
	Funder        string
	BuilderConfig *BuilderConfig
	UseServerTime bool
	RetryOnError  bool
	FeeSlippage   float64
}

// NewClobClient creates a new ClobClient.
func NewClobClient(cfg ClobClientConfig) (*ClobClient, error) {
	if err := ValidateFeeSlippage(cfg.FeeSlippage); err != nil {
		return nil, err
	}

	var signer *Signer
	if cfg.Key != "" {
		var err error
		signer, err = NewSigner(cfg.Key, cfg.ChainID)
		if err != nil {
			return nil, fmt.Errorf("failed to create signer: %w", err)
		}
	}

	c := &ClobClient{
		Host:              strings.TrimRight(cfg.Host, "/"),
		ChainID:           cfg.ChainID,
		UseServerTime:     cfg.UseServerTime,
		RetryOnError:      cfg.RetryOnError,
		BuilderCfg:        cfg.BuilderConfig,
		FeeSlippage:       cfg.FeeSlippage,
		Signer:            signer,
		Creds:             cfg.Creds,
		Builder:           NewOrderBuilder(signer, cfg.SignatureType, cfg.Funder),
		tickSizes:         make(map[string]string),
		negRisk:           make(map[string]bool),
		feeRates:          make(map[string]int),
		feeInfos:          make(map[string]*FeeInfo),
		builderFeeRates:   make(map[string]*BuilderFeeRate),
		tokenConditionMap: make(map[string]string),
	}
	c.Mode = c.getClientMode()
	c.RFQ = NewRFQClient(c)
	return c, nil
}

func (c *ClobClient) getClientMode() int {
	if c.Signer == nil {
		return L0
	}
	if c.Creds == nil {
		return L1
	}
	return L2
}

// AssertLevel1Auth checks that L1 auth is available.
func (c *ClobClient) AssertLevel1Auth() error {
	if c.Signer == nil {
		return &PolyException{Msg: L1AuthUnavailable}
	}
	return nil
}

// AssertLevel2Auth checks that L2 auth is available.
func (c *ClobClient) AssertLevel2Auth() error {
	if c.Signer == nil {
		return &PolyException{Msg: L1AuthUnavailable}
	}
	if c.Creds == nil {
		return &PolyException{Msg: L2AuthUnavailable}
	}
	return nil
}

// GetAddress returns the wallet address.
func (c *ClobClient) GetAddress() (string, error) {
	if err := c.AssertLevel1Auth(); err != nil {
		return "", err
	}
	return c.Signer.Address(), nil
}

// SetAPICreds sets the API credentials and updates the client mode.
func (c *ClobClient) SetAPICreds(creds *ApiCreds) {
	c.Creds = creds
	c.Mode = c.getClientMode()
}

func (c *ClobClient) get(endpoint string, headers map[string]string, params map[string]string) (interface{}, error) {
	return httpGet(endpoint, headers, params)
}

func (c *ClobClient) post(endpoint string, headers map[string]string, data interface{}, params map[string]string) (interface{}, error) {
	return httpPost(endpoint, headers, data, params, c.RetryOnError)
}

func (c *ClobClient) del(endpoint string, headers map[string]string, data interface{}, params map[string]string) (interface{}, error) {
	return httpDelete(endpoint, headers, data, params)
}

func (c *ClobClient) getTimestamp() *int64 {
	if !c.UseServerTime {
		return nil
	}
	result, err := httpGet(c.Host+EndpointTime, nil, nil)
	if err != nil {
		return nil
	}
	if m, ok := result.(map[string]interface{}); ok {
		if t, ok := m["time"]; ok {
			ts := toInt64(t)
			return &ts
		}
		if t, ok := m["timestamp"]; ok {
			ts := toInt64(t)
			return &ts
		}
	}
	if s, ok := result.(string); ok {
		ts, _ := strconv.ParseInt(s, 10, 64)
		return &ts
	}
	return nil
}

func (c *ClobClient) l1Headers(nonce int) (map[string]string, error) {
	if err := c.AssertLevel1Auth(); err != nil {
		return nil, err
	}
	return CreateLevel1Headers(c.Signer, nonce, c.getTimestamp())
}

func (c *ClobClient) l2Headers(method, endpoint string, body interface{}, serializedBody string) (map[string]string, error) {
	if err := c.AssertLevel2Auth(); err != nil {
		return nil, err
	}
	args := &RequestArgs{
		Method:         method,
		RequestPath:    endpoint,
		Body:           body,
		SerializedBody: serializedBody,
	}
	return CreateLevel2Headers(c.Signer, c.Creds, args, c.getTimestamp())
}

// GetOk checks server health.
func (c *ClobClient) GetOk() (interface{}, error) {
	return c.get(c.Host+EndpointOK, nil, nil)
}

// PostHeartbeat posts a heartbeat.
func (c *ClobClient) PostHeartbeat(heartbeatID string) (interface{}, error) {
	body := map[string]string{"heartbeat_id": heartbeatID}
	serialized := marshalCompact(body)
	headers, err := c.l2Headers("POST", EndpointPostHeartbeat, body, serialized)
	if err != nil {
		return nil, err
	}
	return c.post(c.Host+EndpointPostHeartbeat, headers, serialized, nil)
}

// GetVersion gets the order version from the server.
func (c *ClobClient) GetVersion() int {
	result, err := c.get(c.Host+EndpointVersion, nil, nil)
	if err != nil {
		return 2
	}
	if m, ok := result.(map[string]interface{}); ok {
		if v, ok := m["version"]; ok {
			n := toInt64(v)
			if n < 0 || n > math.MaxInt32 {
				return 2
			}
			return int(n)
		}
	}
	return 2
}

// GetServerTime returns the server time.
func (c *ClobClient) GetServerTime() (interface{}, error) {
	return c.get(c.Host+EndpointTime, nil, nil)
}

// GetSamplingSimplifiedMarkets returns sampling simplified markets.
func (c *ClobClient) GetSamplingSimplifiedMarkets(nextCursor string) (interface{}, error) {
	if nextCursor == "" {
		nextCursor = InitialCursor
	}
	return c.get(c.Host+EndpointGetSamplingSimplifiedMarkets, nil, map[string]string{"next_cursor": nextCursor})
}

// GetSamplingMarkets returns sampling markets.
func (c *ClobClient) GetSamplingMarkets(nextCursor string) (interface{}, error) {
	if nextCursor == "" {
		nextCursor = InitialCursor
	}
	return c.get(c.Host+EndpointGetSamplingMarkets, nil, map[string]string{"next_cursor": nextCursor})
}

// GetSimplifiedMarkets returns simplified markets.
func (c *ClobClient) GetSimplifiedMarkets(nextCursor string) (interface{}, error) {
	if nextCursor == "" {
		nextCursor = InitialCursor
	}
	return c.get(c.Host+EndpointGetSimplifiedMarkets, nil, map[string]string{"next_cursor": nextCursor})
}

// GetMarkets returns all markets.
func (c *ClobClient) GetMarkets(nextCursor string) (interface{}, error) {
	if nextCursor == "" {
		nextCursor = InitialCursor
	}
	return c.get(c.Host+EndpointGetMarkets, nil, map[string]string{"next_cursor": nextCursor})
}

// GetMarket returns a specific market by condition ID.
func (c *ClobClient) GetMarket(conditionID string) (interface{}, error) {
	return c.get(c.Host+EndpointGetMarket+conditionID, nil, nil)
}

// GetClobMarketInfo fetches and caches market info.
func (c *ClobClient) GetClobMarketInfo(conditionID string) (map[string]interface{}, error) {
	result, err := c.get(c.Host+EndpointGetClobMarket+conditionID, nil, nil)
	if err != nil {
		return nil, err
	}

	m, ok := result.(map[string]interface{})
	if !ok || m["t"] == nil {
		return nil, &PolyException{Msg: fmt.Sprintf("failed to fetch market info for condition id %s", conditionID)}
	}

	tokens, ok := m["t"].([]interface{})
	if !ok {
		return nil, &PolyException{Msg: fmt.Sprintf("failed to fetch market info for condition id %s", conditionID)}
	}

	for _, token := range tokens {
		tMap, ok := token.(map[string]interface{})
		if !ok || tMap == nil {
			continue
		}
		tokenID := fmt.Sprintf("%v", tMap["t"])
		c.tokenConditionMap[tokenID] = conditionID
		c.tickSizes[tokenID] = fmt.Sprintf("%v", m["mts"])
		c.negRisk[tokenID] = toBool(m["nr"])

		fd, _ := m["fd"].(map[string]interface{})
		fi := &FeeInfo{}
		if fd != nil {
			fi.Rate = toFloat64(fd["r"])
			fi.Exponent = toFloat64(fd["e"])
		}
		c.feeInfos[tokenID] = fi
	}

	return m, nil
}

// GetOrderBook returns the order book for a token.
func (c *ClobClient) GetOrderBook(tokenID string) (interface{}, error) {
	return c.get(c.Host+EndpointGetOrderBook, nil, map[string]string{"token_id": tokenID})
}

// GetOrderBooks returns multiple order books.
func (c *ClobClient) GetOrderBooks(params []BookParams) (interface{}, error) {
	payload := bookParamsToJSON(params)
	return c.post(c.Host+EndpointGetOrderBooks, nil, payload, nil)
}

// GetOrderBookHash computes the hash of an order book.
func (c *ClobClient) GetOrderBookHash(ob *OrderBookSummary) string {
	return GenerateOrderBookSummaryHash(ob)
}

// GetTickSize returns the tick size for a token.
func (c *ClobClient) GetTickSize(tokenID string) (TickSize, error) {
	if ts, ok := c.tickSizes[tokenID]; ok {
		return ts, nil
	}

	if condID, ok := c.tokenConditionMap[tokenID]; ok {
		if _, err := c.GetClobMarketInfo(condID); err != nil {
			return "", err
		}
		return c.tickSizes[tokenID], nil
	}

	result, err := c.get(c.Host+EndpointGetTickSize, nil, map[string]string{"token_id": tokenID})
	if err != nil {
		return "", err
	}
	m, ok := result.(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("unexpected response for tick size")
	}
	ts := fmt.Sprintf("%v", m["minimum_tick_size"])
	c.tickSizes[tokenID] = ts
	return ts, nil
}

// GetNegRisk returns whether a token uses negative risk.
func (c *ClobClient) GetNegRisk(tokenID string) (bool, error) {
	if nr, ok := c.negRisk[tokenID]; ok {
		return nr, nil
	}

	if condID, ok := c.tokenConditionMap[tokenID]; ok {
		if _, err := c.GetClobMarketInfo(condID); err != nil {
			return false, err
		}
		return c.negRisk[tokenID], nil
	}

	result, err := c.get(c.Host+EndpointGetNegRisk, nil, map[string]string{"token_id": tokenID})
	if err != nil {
		return false, err
	}
	m, ok := result.(map[string]interface{})
	if !ok {
		return false, fmt.Errorf("unexpected response for neg risk")
	}
	nr := toBool(m["neg_risk"])
	c.negRisk[tokenID] = nr
	return nr, nil
}

// GetFeeRateBps returns the fee rate in basis points for a token.
func (c *ClobClient) GetFeeRateBps(tokenID string) (int, error) {
	if fr, ok := c.feeRates[tokenID]; ok {
		return fr, nil
	}

	result, err := c.get(c.Host+EndpointGetFeeRate, nil, map[string]string{"token_id": tokenID})
	if err != nil {
		return 0, err
	}
	m, ok := result.(map[string]interface{})
	if !ok {
		return 0, nil
	}
	n := toInt64(m["base_fee"])
	if n < 0 || n > math.MaxInt32 {
		c.feeRates[tokenID] = 0
		return 0, nil
	}
	fr := int(n)
	c.feeRates[tokenID] = fr
	return fr, nil
}

// GetFeeExponent returns the fee exponent for a token.
func (c *ClobClient) GetFeeExponent(tokenID string) (float64, error) {
	if fi, ok := c.feeInfos[tokenID]; ok {
		return fi.Exponent, nil
	}
	if err := c.ensureMarketInfoCached(tokenID); err != nil {
		return 0, err
	}
	return c.feeInfos[tokenID].Exponent, nil
}

// GetMidpoint returns the midpoint for a token.
func (c *ClobClient) GetMidpoint(tokenID string) (interface{}, error) {
	return c.get(c.Host+EndpointGetMidpoint, nil, map[string]string{"token_id": tokenID})
}

// GetMidpoints returns midpoints for multiple tokens.
func (c *ClobClient) GetMidpoints(params []BookParams) (interface{}, error) {
	return c.post(c.Host+EndpointGetMidpoints, nil, bookParamsToJSON(params), nil)
}

// GetPrice returns the price for a token and side.
func (c *ClobClient) GetPrice(tokenID, side string) (interface{}, error) {
	return c.get(c.Host+EndpointGetPrice, nil, map[string]string{"token_id": tokenID, "side": side})
}

// GetPrices returns prices for multiple tokens.
func (c *ClobClient) GetPrices(params []BookParams) (interface{}, error) {
	return c.post(c.Host+EndpointGetPrices, nil, bookParamsToJSON(params), nil)
}

// GetSpread returns the spread for a token.
func (c *ClobClient) GetSpread(tokenID string) (interface{}, error) {
	return c.get(c.Host+EndpointGetSpread, nil, map[string]string{"token_id": tokenID})
}

// GetSpreads returns spreads for multiple tokens.
func (c *ClobClient) GetSpreads(params []BookParams) (interface{}, error) {
	return c.post(c.Host+EndpointGetSpreads, nil, bookParamsToJSON(params), nil)
}

// GetLastTradePrice returns the last trade price.
func (c *ClobClient) GetLastTradePrice(tokenID string) (interface{}, error) {
	return c.get(c.Host+EndpointGetLastTradePrice, nil, map[string]string{"token_id": tokenID})
}

// GetLastTradesPrices returns last trade prices for multiple tokens.
func (c *ClobClient) GetLastTradesPrices(params []BookParams) (interface{}, error) {
	return c.post(c.Host+EndpointGetLastTradesPrices, nil, bookParamsToJSON(params), nil)
}

// GetPricesHistory returns price history.
func (c *ClobClient) GetPricesHistory(params *PricesHistoryParams) (interface{}, error) {
	if params.Interval == "" && (params.StartTS == nil || params.EndTS == nil) {
		return nil, fmt.Errorf("get_prices_history requires either interval or both start_ts and end_ts")
	}
	p := make(map[string]string)
	if params.Market != "" {
		p["market"] = params.Market
	}
	if params.StartTS != nil {
		p["startTs"] = fmt.Sprintf("%d", *params.StartTS)
	}
	if params.EndTS != nil {
		p["endTs"] = fmt.Sprintf("%d", *params.EndTS)
	}
	if params.Fidelity != nil {
		p["fidelity"] = fmt.Sprintf("%d", *params.Fidelity)
	}
	if params.Interval != "" {
		p["interval"] = params.Interval
	}
	return c.get(c.Host+EndpointGetPricesHistory, nil, p)
}

// CalculateMarketPrice calculates the market price for an order.
func (c *ClobClient) CalculateMarketPrice(tokenID, side string, amount float64, orderType string) (float64, error) {
	book, err := c.GetOrderBook(tokenID)
	if err != nil {
		return 0, err
	}
	bookMap, ok := book.(map[string]interface{})
	if !ok {
		return 0, &PolyException{Msg: "no orderbook"}
	}

	if side == SideBuyStr {
		asks := toPositionsList(bookMap["asks"])
		if len(asks) == 0 {
			return 0, &PolyException{Msg: "no match"}
		}
		return c.Builder.CalculateBuyMarketPrice(asks, amount, orderType)
	}
	bids := toPositionsList(bookMap["bids"])
	if len(bids) == 0 {
		return 0, &PolyException{Msg: "no match"}
	}
	return c.Builder.CalculateSellMarketPrice(bids, amount, orderType)
}

// GetCurrentRewards returns all current rewards.
func (c *ClobClient) GetCurrentRewards() ([]interface{}, error) {
	var results []interface{}
	nextCursor := InitialCursor
	for nextCursor != EndCursor {
		resp, err := c.get(c.Host+EndpointGetRewardsMarketsCurrent, nil, map[string]string{"next_cursor": nextCursor})
		if err != nil {
			return nil, err
		}
		m := resp.(map[string]interface{})
		nextCursor = fmt.Sprintf("%v", m["next_cursor"])
		if data, ok := m["data"].([]interface{}); ok {
			results = append(results, data...)
		}
	}
	return results, nil
}

// GetRawRewardsForMarket returns raw rewards for a market.
func (c *ClobClient) GetRawRewardsForMarket(conditionID string) ([]interface{}, error) {
	var results []interface{}
	nextCursor := InitialCursor
	for nextCursor != EndCursor {
		resp, err := c.get(c.Host+EndpointGetRewardsMarkets+conditionID, nil, map[string]string{"next_cursor": nextCursor})
		if err != nil {
			return nil, err
		}
		m := resp.(map[string]interface{})
		nextCursor = fmt.Sprintf("%v", m["next_cursor"])
		if data, ok := m["data"].([]interface{}); ok {
			results = append(results, data...)
		}
	}
	return results, nil
}

// CreateAPIKey creates a new API key.
func (c *ClobClient) CreateAPIKey(nonce int) (*ApiCreds, error) {
	headers, err := c.l1Headers(nonce)
	if err != nil {
		return nil, err
	}
	resp, err := c.post(c.Host+EndpointCreateAPIKey, headers, nil, nil)
	if err != nil {
		return nil, err
	}
	m := resp.(map[string]interface{})
	return &ApiCreds{
		APIKey:        fmt.Sprintf("%v", m["apiKey"]),
		APISecret:     fmt.Sprintf("%v", m["secret"]),
		APIPassphrase: fmt.Sprintf("%v", m["passphrase"]),
	}, nil
}

// DeriveAPIKey derives an existing API key.
func (c *ClobClient) DeriveAPIKey(nonce int) (*ApiCreds, error) {
	headers, err := c.l1Headers(nonce)
	if err != nil {
		return nil, err
	}
	resp, err := c.get(c.Host+EndpointDeriveAPIKey, headers, nil)
	if err != nil {
		return nil, err
	}
	m := resp.(map[string]interface{})
	return &ApiCreds{
		APIKey:        fmt.Sprintf("%v", m["apiKey"]),
		APISecret:     fmt.Sprintf("%v", m["secret"]),
		APIPassphrase: fmt.Sprintf("%v", m["passphrase"]),
	}, nil
}

// CreateOrDeriveAPIKey tries to create an API key, falls back to derive.
func (c *ClobClient) CreateOrDeriveAPIKey(nonce int) (*ApiCreds, error) {
	creds, err := c.CreateAPIKey(nonce)
	if err == nil && creds.APIKey != "" {
		return creds, nil
	}
	return c.DeriveAPIKey(nonce)
}

// GetAPIKeys returns all API keys.
func (c *ClobClient) GetAPIKeys() (interface{}, error) {
	headers, err := c.l2Headers("GET", EndpointGetAPIKeys, nil, "")
	if err != nil {
		return nil, err
	}
	return c.get(c.Host+EndpointGetAPIKeys, headers, nil)
}

// GetClosedOnlyMode returns the closed-only mode status.
func (c *ClobClient) GetClosedOnlyMode() (interface{}, error) {
	headers, err := c.l2Headers("GET", EndpointClosedOnly, nil, "")
	if err != nil {
		return nil, err
	}
	return c.get(c.Host+EndpointClosedOnly, headers, nil)
}

// DeleteAPIKey deletes the current API key.
func (c *ClobClient) DeleteAPIKey() (interface{}, error) {
	headers, err := c.l2Headers("DELETE", EndpointDeleteAPIKey, nil, "")
	if err != nil {
		return nil, err
	}
	return c.del(c.Host+EndpointDeleteAPIKey, headers, nil, nil)
}

// GetOrder returns a specific order by ID.
func (c *ClobClient) GetOrder(orderID string) (interface{}, error) {
	endpoint := EndpointGetOrder + orderID
	headers, err := c.l2Headers("GET", endpoint, nil, "")
	if err != nil {
		return nil, err
	}
	return c.get(c.Host+endpoint, headers, nil)
}

// GetOpenOrders returns open orders with optional filters.
func (c *ClobClient) GetOpenOrders(params *OpenOrderParams, onlyFirstPage bool, nextCursor string) ([]interface{}, error) {
	headers, err := c.l2Headers("GET", EndpointOrders, nil, "")
	if err != nil {
		return nil, err
	}

	var results []interface{}
	cursor := nextCursor
	if cursor == "" {
		cursor = InitialCursor
	}
	first := true
	for cursor != EndCursor && (first || !onlyFirstPage) {
		first = false
		p := make(map[string]string)
		if params != nil {
			if params.Market != "" {
				p["market"] = params.Market
			}
			if params.AssetID != "" {
				p["asset_id"] = params.AssetID
			}
			if params.ID != "" {
				p["id"] = params.ID
			}
		}
		p["next_cursor"] = cursor
		resp, err := c.get(c.Host+EndpointOrders, headers, p)
		if err != nil {
			return nil, err
		}
		m := resp.(map[string]interface{})
		cursor = fmt.Sprintf("%v", m["next_cursor"])
		if data, ok := m["data"].([]interface{}); ok {
			results = append(results, data...)
		}
	}
	return results, nil
}

// GetPreMigrationOrders returns pre-migration orders.
func (c *ClobClient) GetPreMigrationOrders(onlyFirstPage bool, nextCursor string) ([]interface{}, error) {
	headers, err := c.l2Headers("GET", EndpointPreMigrationOrders, nil, "")
	if err != nil {
		return nil, err
	}

	var results []interface{}
	cursor := nextCursor
	if cursor == "" {
		cursor = InitialCursor
	}
	first := true
	for cursor != EndCursor && (first || !onlyFirstPage) {
		first = false
		p := map[string]string{"next_cursor": cursor}
		resp, err := c.get(c.Host+EndpointPreMigrationOrders, headers, p)
		if err != nil {
			return nil, err
		}
		m := resp.(map[string]interface{})
		cursor = fmt.Sprintf("%v", m["next_cursor"])
		if data, ok := m["data"].([]interface{}); ok {
			results = append(results, data...)
		}
	}
	return results, nil
}

// GetTrades returns trades with optional filters.
func (c *ClobClient) GetTrades(params *TradeParams, onlyFirstPage bool, nextCursor string) ([]interface{}, error) {
	headers, err := c.l2Headers("GET", EndpointTrades, nil, "")
	if err != nil {
		return nil, err
	}

	var results []interface{}
	cursor := nextCursor
	if cursor == "" {
		cursor = InitialCursor
	}
	first := true
	for cursor != EndCursor && (first || !onlyFirstPage) {
		first = false
		p := make(map[string]string)
		if params != nil {
			if params.Market != "" {
				p["market"] = params.Market
			}
			if params.AssetID != "" {
				p["asset_id"] = params.AssetID
			}
			if params.After != 0 {
				p["after"] = fmt.Sprintf("%d", params.After)
			}
			if params.Before != 0 {
				p["before"] = fmt.Sprintf("%d", params.Before)
			}
			if params.MakerAddress != "" {
				p["maker_address"] = params.MakerAddress
			}
			if params.ID != "" {
				p["id"] = params.ID
			}
		}
		p["next_cursor"] = cursor
		resp, err := c.get(c.Host+EndpointTrades, headers, p)
		if err != nil {
			return nil, err
		}
		m := resp.(map[string]interface{})
		cursor = fmt.Sprintf("%v", m["next_cursor"])
		if data, ok := m["data"].([]interface{}); ok {
			results = append(results, data...)
		}
	}
	return results, nil
}

// GetTradesPaginated returns a single page of trades.
func (c *ClobClient) GetTradesPaginated(params *TradeParams, nextCursor string) (map[string]interface{}, error) {
	headers, err := c.l2Headers("GET", EndpointTrades, nil, "")
	if err != nil {
		return nil, err
	}

	cursor := nextCursor
	if cursor == "" {
		cursor = InitialCursor
	}
	p := make(map[string]string)
	if params != nil {
		if params.Market != "" {
			p["market"] = params.Market
		}
		if params.AssetID != "" {
			p["asset_id"] = params.AssetID
		}
		if params.After != 0 {
			p["after"] = fmt.Sprintf("%d", params.After)
		}
		if params.Before != 0 {
			p["before"] = fmt.Sprintf("%d", params.Before)
		}
		if params.MakerAddress != "" {
			p["maker_address"] = params.MakerAddress
		}
		if params.ID != "" {
			p["id"] = params.ID
		}
	}
	p["next_cursor"] = cursor
	resp, err := c.get(c.Host+EndpointTrades, headers, p)
	if err != nil {
		return nil, err
	}
	m := resp.(map[string]interface{})
	data, _ := m["data"].([]interface{})
	return map[string]interface{}{
		"trades":      data,
		"next_cursor": m["next_cursor"],
		"limit":       m["limit"],
		"count":       m["count"],
	}, nil
}

// GetBuilderTrades returns trades for a builder.
func (c *ClobClient) GetBuilderTrades(params *BuilderTradeParams, nextCursor string) (map[string]interface{}, error) {
	if params.BuilderCode == "" || params.BuilderCode == Bytes32Zero {
		return nil, &PolyException{Msg: "builder_code is required and cannot be zero"}
	}
	cursor := nextCursor
	if cursor == "" {
		cursor = InitialCursor
	}
	p := map[string]string{"builder_code": params.BuilderCode}
	if params.ID != "" {
		p["id"] = params.ID
	}
	if params.MakerAddress != "" {
		p["maker_address"] = params.MakerAddress
	}
	if params.Market != "" {
		p["market"] = params.Market
	}
	if params.AssetID != "" {
		p["asset_id"] = params.AssetID
	}
	if params.Before != "" {
		p["before"] = params.Before
	}
	if params.After != "" {
		p["after"] = params.After
	}
	p["next_cursor"] = cursor
	resp, err := c.get(c.Host+EndpointGetBuilderTrades, nil, p)
	if err != nil {
		return nil, err
	}
	m := resp.(map[string]interface{})
	data, _ := m["data"].([]interface{})
	return map[string]interface{}{
		"trades":      data,
		"next_cursor": m["next_cursor"],
		"limit":       m["limit"],
		"count":       m["count"],
	}, nil
}

// GetNotifications returns notifications.
func (c *ClobClient) GetNotifications() (interface{}, error) {
	headers, err := c.l2Headers("GET", EndpointGetNotifications, nil, "")
	if err != nil {
		return nil, err
	}
	return c.get(c.Host+EndpointGetNotifications, headers, map[string]string{
		"signature_type": fmt.Sprintf("%d", int(c.Builder.SignatureType)),
	})
}

// DropNotifications drops notifications.
func (c *ClobClient) DropNotifications(params *DropNotificationParams) (interface{}, error) {
	headers, err := c.l2Headers("DELETE", EndpointGetNotifications, nil, "")
	if err != nil {
		return nil, err
	}
	p := make(map[string]string)
	if params != nil && len(params.IDs) > 0 {
		p["ids"] = strings.Join(params.IDs, ",")
	}
	return c.del(c.Host+EndpointGetNotifications, headers, nil, p)
}

// GetBalanceAllowance returns balance and allowance info.
func (c *ClobClient) GetBalanceAllowance(params *BalanceAllowanceParams) (interface{}, error) {
	headers, err := c.l2Headers("GET", EndpointGetBalanceAllowance, nil, "")
	if err != nil {
		return nil, err
	}
	p := map[string]string{
		"signature_type": fmt.Sprintf("%d", int(c.Builder.SignatureType)),
	}
	if params != nil {
		if params.AssetType != "" {
			p["asset_type"] = params.AssetType
		}
		if params.TokenID != "" {
			p["token_id"] = params.TokenID
		}
	}
	return c.get(c.Host+EndpointGetBalanceAllowance, headers, p)
}

// UpdateBalanceAllowance updates balance and allowance.
func (c *ClobClient) UpdateBalanceAllowance(params *BalanceAllowanceParams) (interface{}, error) {
	headers, err := c.l2Headers("GET", EndpointUpdateBalanceAllowance, nil, "")
	if err != nil {
		return nil, err
	}
	p := map[string]string{
		"signature_type": fmt.Sprintf("%d", int(c.Builder.SignatureType)),
	}
	if params != nil {
		if params.AssetType != "" {
			p["asset_type"] = params.AssetType
		}
		if params.TokenID != "" {
			p["token_id"] = params.TokenID
		}
	}
	return c.get(c.Host+EndpointUpdateBalanceAllowance, headers, p)
}

// CreateOrder creates a signed order (does not post it).
func (c *ClobClient) CreateOrder(args *OrderArgs, options *PartialCreateOrderOptions) (interface{}, error) {
	if err := c.AssertLevel1Auth(); err != nil {
		return nil, err
	}

	if c.BuilderCfg != nil && c.BuilderCfg.BuilderCode != "" {
		if args.BuilderCode == "" || args.BuilderCode == Bytes32Zero {
			args.BuilderCode = c.BuilderCfg.BuilderCode
		}
	}

	tokenID := args.TokenID
	tickSize, err := c.resolveTickSize(tokenID, options)
	if err != nil {
		return nil, err
	}

	if !PriceValid(args.Price, tickSize) {
		ts, _ := strconv.ParseFloat(tickSize, 64)
		return nil, &PolyException{Msg: fmt.Sprintf("invalid price (%v), min: %v - max: %v", args.Price, ts, 1-ts)}
	}

	price := roundNormal(args.Price, int(RoundingConfig[tickSize].Price))
	version := c.resolveVersion(false)

	size := args.Size
	if version == 2 && (args.Side == SideBuyStr) && args.UserUSDCBalance != nil {
		adjusted, err := c.adjustBuyAmountForBalance(tokenID, size*price, price, *args.UserUSDCBalance, args.BuilderCode)
		if err != nil {
			return nil, err
		}
		size = adjusted / price
	}

	negRisk, err := c.resolveNegRisk(tokenID, options)
	if err != nil {
		return nil, err
	}

	var feeRateBps *int
	if version == 1 {
		fr, err := c.resolveFeeRateBps(tokenID, nil)
		if err != nil {
			return nil, err
		}
		feeRateBps = &fr
	}

	buildArgs := &OrderArgs{
		TokenID:     args.TokenID,
		Price:       price,
		Size:        size,
		Side:        args.Side,
		Expiration:  args.Expiration,
		BuilderCode: args.BuilderCode,
		Metadata:    args.Metadata,
	}

	return c.Builder.BuildOrder(buildArgs, &CreateOrderOptions{TickSize: tickSize, NegRisk: negRisk}, version, feeRateBps)
}

// CreateMarketOrder creates a signed market order.
func (c *ClobClient) CreateMarketOrder(args *MarketOrderArgs, options *PartialCreateOrderOptions) (interface{}, error) {
	if err := c.AssertLevel1Auth(); err != nil {
		return nil, err
	}

	tokenID := args.TokenID
	if err := c.ensureMarketInfoCached(tokenID); err != nil {
		return nil, err
	}

	tickSize, err := c.resolveTickSize(tokenID, options)
	if err != nil {
		return nil, err
	}

	price := args.Price
	if price == 0 {
		orderType := args.OrderType
		if orderType == "" {
			orderType = OrderTypeFOK
		}
		p, err := c.CalculateMarketPrice(tokenID, args.Side, args.Amount, orderType)
		if err != nil {
			return nil, err
		}
		price = p
	}

	if !PriceValid(price, tickSize) {
		ts, _ := strconv.ParseFloat(tickSize, 64)
		return nil, &PolyException{Msg: fmt.Sprintf("invalid price (%v), min: %v - max: %v", price, ts, 1-ts)}
	}

	if c.BuilderCfg != nil && c.BuilderCfg.BuilderCode != "" {
		if args.BuilderCode == "" || args.BuilderCode == Bytes32Zero {
			args.BuilderCode = c.BuilderCfg.BuilderCode
		}
	}

	amount := args.Amount
	if (args.Side == SideBuyStr) && args.UserUSDCBalance > 0 {
		adjusted, err := c.adjustBuyAmountForBalance(tokenID, amount, price, args.UserUSDCBalance, args.BuilderCode)
		if err != nil {
			return nil, err
		}
		amount = adjusted
	}

	negRisk, err := c.resolveNegRisk(tokenID, options)
	if err != nil {
		return nil, err
	}

	version := c.resolveVersion(false)

	var feeRateBps *int
	if version == 1 {
		fr, err := c.resolveFeeRateBps(tokenID, nil)
		if err != nil {
			return nil, err
		}
		feeRateBps = &fr
	}

	buildArgs := &MarketOrderArgs{
		TokenID:     args.TokenID,
		Amount:      amount,
		Side:        args.Side,
		Price:       price,
		OrderType:   args.OrderType,
		BuilderCode: args.BuilderCode,
		Metadata:    args.Metadata,
	}

	return c.Builder.BuildMarketOrder(buildArgs, &CreateOrderOptions{TickSize: tickSize, NegRisk: negRisk}, version, feeRateBps)
}

// CreateAndPostOrder creates and posts an order.
func (c *ClobClient) CreateAndPostOrder(args *OrderArgs, options *PartialCreateOrderOptions, orderType string, postOnly, deferExec bool) (interface{}, error) {
	if orderType == "" {
		orderType = OrderTypeGTC
	}
	return c.retryOnVersionUpdate(func() (interface{}, error) {
		order, err := c.CreateOrder(args, options)
		if err != nil {
			return nil, err
		}
		return c.PostOrder(order, orderType, postOnly, deferExec)
	})
}

// CreateAndPostMarketOrder creates and posts a market order.
func (c *ClobClient) CreateAndPostMarketOrder(args *MarketOrderArgs, options *PartialCreateOrderOptions, orderType string, deferExec bool) (interface{}, error) {
	if orderType == "" {
		orderType = OrderTypeFOK
	}
	return c.retryOnVersionUpdate(func() (interface{}, error) {
		order, err := c.CreateMarketOrder(args, options)
		if err != nil {
			return nil, err
		}
		return c.PostOrder(order, orderType, false, deferExec)
	})
}

// PostOrder posts a signed order to the API.
func (c *ClobClient) PostOrder(order interface{}, orderType string, postOnly, deferExec bool) (interface{}, error) {
	if err := c.AssertLevel2Auth(); err != nil {
		return nil, err
	}
	if postOnly && (orderType == OrderTypeFOK || orderType == OrderTypeFAK) {
		return nil, fmt.Errorf("post_only is not supported for FOK/FAK orders")
	}

	owner := c.Creds.APIKey
	var orderPayload map[string]interface{}
	if IsV2Order(order) {
		o := order.(*SignedOrderV2)
		orderPayload = OrderToJSONV2(o, owner, orderType, postOnly, deferExec)
	} else {
		o := order.(*SignedOrderV1)
		orderPayload = OrderToJSONV1(o, owner, orderType, postOnly, deferExec)
	}

	serialized := marshalCompact(orderPayload)
	headers, err := c.l2Headers("POST", EndpointPostOrder, orderPayload, serialized)
	if err != nil {
		return nil, err
	}

	res, err := c.post(c.Host+EndpointPostOrder, headers, serialized, nil)
	if err != nil {
		return nil, err
	}

	if c.isOrderVersionMismatch(res) {
		c.resolveVersion(true)
	}

	return res, nil
}

// PostOrders posts multiple signed orders to the API.
func (c *ClobClient) PostOrders(args []PostOrdersArgs, postOnly, deferExec bool) (interface{}, error) {
	if err := c.AssertLevel2Auth(); err != nil {
		return nil, err
	}

	owner := c.Creds.APIKey
	var ordersPayload []map[string]interface{}
	for _, arg := range args {
		var payload map[string]interface{}
		if IsV2Order(arg.Order) {
			o := arg.Order.(*SignedOrderV2)
			payload = OrderToJSONV2(o, owner, arg.OrderType, postOnly, deferExec)
		} else {
			o := arg.Order.(*SignedOrderV1)
			payload = OrderToJSONV1(o, owner, arg.OrderType, postOnly, deferExec)
		}
		ordersPayload = append(ordersPayload, payload)
	}

	serialized := marshalCompact(ordersPayload)
	headers, err := c.l2Headers("POST", EndpointPostOrders, ordersPayload, serialized)
	if err != nil {
		return nil, err
	}

	res, err := c.post(c.Host+EndpointPostOrders, headers, serialized, nil)
	if err != nil {
		return nil, err
	}

	if c.isOrderVersionMismatch(res) {
		c.resolveVersion(true)
	}

	return res, nil
}

// CancelOrder cancels a single order.
func (c *ClobClient) CancelOrder(payload *OrderPayload) (interface{}, error) {
	if err := c.AssertLevel2Auth(); err != nil {
		return nil, err
	}
	body := map[string]string{"orderID": payload.OrderID}
	serialized := marshalCompact(body)
	headers, err := c.l2Headers("DELETE", EndpointCancel, body, serialized)
	if err != nil {
		return nil, err
	}
	return c.del(c.Host+EndpointCancel, headers, serialized, nil)
}

// CancelOrders cancels multiple orders by hash.
func (c *ClobClient) CancelOrders(orderHashes []string) (interface{}, error) {
	if err := c.AssertLevel2Auth(); err != nil {
		return nil, err
	}
	serialized := marshalCompact(orderHashes)
	headers, err := c.l2Headers("DELETE", EndpointCancelOrders, orderHashes, serialized)
	if err != nil {
		return nil, err
	}
	return c.del(c.Host+EndpointCancelOrders, headers, serialized, nil)
}

// CancelAll cancels all orders.
func (c *ClobClient) CancelAll() (interface{}, error) {
	if err := c.AssertLevel2Auth(); err != nil {
		return nil, err
	}
	headers, err := c.l2Headers("DELETE", EndpointCancelAll, nil, "")
	if err != nil {
		return nil, err
	}
	return c.del(c.Host+EndpointCancelAll, headers, nil, nil)
}

// CancelMarketOrders cancels all orders for a market.
func (c *ClobClient) CancelMarketOrders(params *OrderMarketCancelParams) (interface{}, error) {
	if err := c.AssertLevel2Auth(); err != nil {
		return nil, err
	}
	body := make(map[string]string)
	if params.Market != "" {
		body["market"] = params.Market
	}
	if params.AssetID != "" {
		body["asset_id"] = params.AssetID
	}
	serialized := marshalCompact(body)
	headers, err := c.l2Headers("DELETE", EndpointCancelMarketOrders, body, serialized)
	if err != nil {
		return nil, err
	}
	return c.del(c.Host+EndpointCancelMarketOrders, headers, serialized, nil)
}

// IsOrderScoring checks if an order is scoring.
func (c *ClobClient) IsOrderScoring(params *OrderScoringParams) (interface{}, error) {
	if err := c.AssertLevel2Auth(); err != nil {
		return nil, err
	}
	headers, err := c.l2Headers("GET", EndpointIsOrderScoring, nil, "")
	if err != nil {
		return nil, err
	}
	p := make(map[string]string)
	if params != nil && params.OrderID != "" {
		p["order_id"] = params.OrderID
	}
	return c.get(c.Host+EndpointIsOrderScoring, headers, p)
}

// AreOrdersScoring checks if multiple orders are scoring.
func (c *ClobClient) AreOrdersScoring(params *OrdersScoringParams) (interface{}, error) {
	if err := c.AssertLevel2Auth(); err != nil {
		return nil, err
	}
	orderIDs := params.OrderIDs
	serialized := marshalCompact(orderIDs)
	headers, err := c.l2Headers("POST", EndpointAreOrdersScoring, orderIDs, serialized)
	if err != nil {
		return nil, err
	}
	return c.post(c.Host+EndpointAreOrdersScoring, headers, serialized, nil)
}

// GetEarningsForUserForDay returns daily earnings.
func (c *ClobClient) GetEarningsForUserForDay(date string) ([]interface{}, error) {
	if err := c.AssertLevel2Auth(); err != nil {
		return nil, err
	}
	headers, err := c.l2Headers("GET", EndpointGetEarningsForUserForDay, nil, "")
	if err != nil {
		return nil, err
	}

	var results []interface{}
	nextCursor := InitialCursor
	for nextCursor != EndCursor {
		p := map[string]string{
			"date":           date,
			"signature_type": fmt.Sprintf("%d", int(c.Builder.SignatureType)),
			"next_cursor":    nextCursor,
		}
		resp, err := c.get(c.Host+EndpointGetEarningsForUserForDay, headers, p)
		if err != nil {
			return nil, err
		}
		m := resp.(map[string]interface{})
		nextCursor = fmt.Sprintf("%v", m["next_cursor"])
		if data, ok := m["data"].([]interface{}); ok {
			results = append(results, data...)
		}
	}
	return results, nil
}

// GetTotalEarningsForUserForDay returns total daily earnings.
func (c *ClobClient) GetTotalEarningsForUserForDay(date string) (interface{}, error) {
	if err := c.AssertLevel2Auth(); err != nil {
		return nil, err
	}
	headers, err := c.l2Headers("GET", EndpointGetTotalEarningsForUserForDay, nil, "")
	if err != nil {
		return nil, err
	}
	p := map[string]string{
		"date":           date,
		"signature_type": fmt.Sprintf("%d", int(c.Builder.SignatureType)),
	}
	return c.get(c.Host+EndpointGetTotalEarningsForUserForDay, headers, p)
}

// GetUserEarningsAndMarketsConfig returns earnings and markets config.
func (c *ClobClient) GetUserEarningsAndMarketsConfig(date, orderBy, position string, noCompetition bool) ([]interface{}, error) {
	if err := c.AssertLevel2Auth(); err != nil {
		return nil, err
	}
	headers, err := c.l2Headers("GET", EndpointGetRewardsEarningsPercentages, nil, "")
	if err != nil {
		return nil, err
	}

	var results []interface{}
	nextCursor := InitialCursor
	for nextCursor != EndCursor {
		p := map[string]string{
			"date":           date,
			"signature_type": fmt.Sprintf("%d", int(c.Builder.SignatureType)),
			"next_cursor":    nextCursor,
			"order_by":       orderBy,
			"position":       position,
			"no_competition": fmt.Sprintf("%v", noCompetition),
		}
		resp, err := c.get(c.Host+EndpointGetRewardsEarningsPercentages, headers, p)
		if err != nil {
			return nil, err
		}
		m := resp.(map[string]interface{})
		nextCursor = fmt.Sprintf("%v", m["next_cursor"])
		if data, ok := m["data"].([]interface{}); ok {
			results = append(results, data...)
		}
	}
	return results, nil
}

// GetRewardPercentages returns reward percentages.
func (c *ClobClient) GetRewardPercentages() (interface{}, error) {
	if err := c.AssertLevel2Auth(); err != nil {
		return nil, err
	}
	headers, err := c.l2Headers("GET", EndpointGetLiquidityRewardPercentages, nil, "")
	if err != nil {
		return nil, err
	}
	return c.get(c.Host+EndpointGetLiquidityRewardPercentages, headers, map[string]string{
		"signature_type": fmt.Sprintf("%d", int(c.Builder.SignatureType)),
	})
}

// CreateBuilderAPIKey creates a builder API key.
func (c *ClobClient) CreateBuilderAPIKey() (interface{}, error) {
	if err := c.AssertLevel2Auth(); err != nil {
		return nil, err
	}
	headers, err := c.l2Headers("POST", EndpointCreateBuilderAPIKey, nil, "")
	if err != nil {
		return nil, err
	}
	return c.post(c.Host+EndpointCreateBuilderAPIKey, headers, nil, nil)
}

// GetBuilderAPIKeys returns builder API keys.
func (c *ClobClient) GetBuilderAPIKeys() (interface{}, error) {
	if err := c.AssertLevel2Auth(); err != nil {
		return nil, err
	}
	headers, err := c.l2Headers("GET", EndpointGetBuilderAPIKeys, nil, "")
	if err != nil {
		return nil, err
	}
	return c.get(c.Host+EndpointGetBuilderAPIKeys, headers, nil)
}

// RevokeBuilderAPIKey revokes a builder API key.
func (c *ClobClient) RevokeBuilderAPIKey() (interface{}, error) {
	if err := c.AssertLevel2Auth(); err != nil {
		return nil, err
	}
	headers, err := c.l2Headers("DELETE", EndpointRevokeBuilderAPIKey, nil, "")
	if err != nil {
		return nil, err
	}
	return c.del(c.Host+EndpointRevokeBuilderAPIKey, headers, nil, nil)
}

// CreateReadonlyAPIKey creates a readonly API key.
func (c *ClobClient) CreateReadonlyAPIKey() (interface{}, error) {
	if err := c.AssertLevel2Auth(); err != nil {
		return nil, err
	}
	headers, err := c.l2Headers("POST", EndpointCreateReadonlyAPIKey, nil, "")
	if err != nil {
		return nil, err
	}
	return c.post(c.Host+EndpointCreateReadonlyAPIKey, headers, nil, nil)
}

// GetReadonlyAPIKeys returns readonly API keys.
func (c *ClobClient) GetReadonlyAPIKeys() (interface{}, error) {
	if err := c.AssertLevel2Auth(); err != nil {
		return nil, err
	}
	headers, err := c.l2Headers("GET", EndpointGetReadonlyAPIKeys, nil, "")
	if err != nil {
		return nil, err
	}
	return c.get(c.Host+EndpointGetReadonlyAPIKeys, headers, nil)
}

// DeleteReadonlyAPIKey deletes a readonly API key.
func (c *ClobClient) DeleteReadonlyAPIKey(key string) (interface{}, error) {
	if err := c.AssertLevel2Auth(); err != nil {
		return nil, err
	}
	body := map[string]string{"key": key}
	serialized := marshalCompact(body)
	headers, err := c.l2Headers("DELETE", EndpointDeleteReadonlyAPIKey, body, serialized)
	if err != nil {
		return nil, err
	}
	return c.del(c.Host+EndpointDeleteReadonlyAPIKey, headers, serialized, nil)
}

// GetMarketTradesEvents returns live trade events for a market.
func (c *ClobClient) GetMarketTradesEvents(conditionID string) (interface{}, error) {
	return c.get(c.Host+EndpointGetMarketTradesEvents+conditionID, nil, nil)
}

// Private helpers

func (c *ClobClient) getBuilderTakerFeeRate(builderCode string) float64 {
	if builderCode == "" || builderCode == Bytes32Zero {
		return 0
	}
	c.ensureBuilderFeeRateCached(builderCode)
	if fr, ok := c.builderFeeRates[builderCode]; ok {
		return fr.Taker
	}
	return 0
}

func (c *ClobClient) adjustBuyAmountForBalance(tokenID string, amount, price, userUSDCBalance float64, builderCode string) (float64, error) {
	if err := c.ensureMarketInfoCached(tokenID); err != nil {
		return 0, err
	}
	builderTakerFeeRate := c.getBuilderTakerFeeRate(builderCode)
	fi := c.feeInfos[tokenID]
	if fi == nil {
		fi = &FeeInfo{}
	}
	return AdjustBuyAmountForFees(amount, price, userUSDCBalance, fi.Rate, fi.Exponent, builderTakerFeeRate, c.FeeSlippage)
}

func (c *ClobClient) resolveTickSize(tokenID string, options *PartialCreateOrderOptions) (TickSize, error) {
	minTickSize, err := c.GetTickSize(tokenID)
	if err != nil {
		return "", err
	}
	if options != nil && options.TickSize != nil {
		if IsTickSizeSmaller(*options.TickSize, minTickSize) {
			return "", &PolyException{Msg: fmt.Sprintf("invalid tick size (%s), minimum for the market is %s", *options.TickSize, minTickSize)}
		}
		return *options.TickSize, nil
	}
	return minTickSize, nil
}

func (c *ClobClient) resolveNegRisk(tokenID string, options *PartialCreateOrderOptions) (bool, error) {
	if options != nil && options.NegRisk != nil {
		return *options.NegRisk, nil
	}
	return c.GetNegRisk(tokenID)
}

func (c *ClobClient) resolveFeeRateBps(tokenID string, userFeeRateBps *int) (int, error) {
	marketFeeRate, err := c.GetFeeRateBps(tokenID)
	if err != nil {
		return 0, err
	}
	if marketFeeRate > 0 && userFeeRateBps != nil && *userFeeRateBps != marketFeeRate {
		return 0, &PolyException{Msg: fmt.Sprintf("invalid user provided fee rate: %d, fee rate for the market must be %d", *userFeeRateBps, marketFeeRate)}
	}
	return marketFeeRate, nil
}

func (c *ClobClient) resolveVersion(forceUpdate bool) int {
	if !forceUpdate && c.cachedVersion != nil {
		return *c.cachedVersion
	}
	v := c.GetVersion()
	c.cachedVersion = &v
	return v
}

func (c *ClobClient) ensureBuilderFeeRateCached(builderCode string) {
	if builderCode == "" || builderCode == Bytes32Zero {
		return
	}
	if _, ok := c.builderFeeRates[builderCode]; ok {
		return
	}
	result, err := c.get(c.Host+EndpointGetBuilderFeeRate+builderCode, nil, nil)
	if err != nil {
		return
	}
	m, ok := result.(map[string]interface{})
	if !ok {
		return
	}
	c.builderFeeRates[builderCode] = &BuilderFeeRate{
		Maker: toFloat64(m["builder_maker_fee_rate_bps"]) / float64(BuilderFeesBPS),
		Taker: toFloat64(m["builder_taker_fee_rate_bps"]) / float64(BuilderFeesBPS),
	}
}

func (c *ClobClient) ensureMarketInfoCached(tokenID string) error {
	if _, ok := c.feeInfos[tokenID]; ok {
		return nil
	}

	if _, ok := c.tokenConditionMap[tokenID]; !ok {
		result, err := c.get(c.Host+EndpointGetMarketByToken+tokenID, nil, nil)
		if err != nil {
			return err
		}
		m, ok := result.(map[string]interface{})
		if !ok || m["condition_id"] == nil {
			return &PolyException{Msg: fmt.Sprintf("failed to resolve condition id for token %s", tokenID)}
		}
		c.tokenConditionMap[tokenID] = fmt.Sprintf("%v", m["condition_id"])
	}

	_, err := c.GetClobMarketInfo(c.tokenConditionMap[tokenID])
	return err
}

func (c *ClobClient) isOrderVersionMismatch(resp interface{}) bool {
	m, ok := resp.(map[string]interface{})
	if !ok {
		return false
	}
	errVal, ok := m["error"]
	if !ok || errVal == nil {
		return false
	}
	var message string
	switch v := errVal.(type) {
	case string:
		message = v
	default:
		b, _ := json.Marshal(v)
		message = string(b)
	}
	return strings.Contains(message, OrderVersionMismatchError)
}

func (c *ClobClient) retryOnVersionUpdate(fn func() (interface{}, error)) (interface{}, error) {
	version := c.resolveVersion(false)
	var result interface{}
	var err error
	for i := 0; i < 2; i++ {
		result, err = fn()
		if err != nil {
			return nil, err
		}
		if version == c.resolveVersion(false) {
			break
		}
	}
	return result, nil
}

// Helper functions

func marshalCompact(v interface{}) string {
	b, _ := json.Marshal(v)
	return string(b)
}

func bookParamsToJSON(params []BookParams) []map[string]interface{} {
	var result []map[string]interface{}
	for _, p := range params {
		item := map[string]interface{}{
			"token_id": p.TokenID,
		}
		if p.Side != "" {
			item["side"] = p.Side
		}
		result = append(result, item)
	}
	return result
}

func toInt64(v interface{}) int64 {
	switch val := v.(type) {
	case float64:
		return int64(val)
	case int:
		return int64(val)
	case int64:
		return val
	case string:
		n, _ := strconv.ParseInt(val, 10, 64)
		return n
	case json.Number:
		n, _ := val.Int64()
		return n
	default:
		return 0
	}
}

func toFloat64(v interface{}) float64 {
	switch val := v.(type) {
	case float64:
		return val
	case int:
		return float64(val)
	case int64:
		return float64(val)
	case string:
		f, _ := strconv.ParseFloat(val, 64)
		return f
	case json.Number:
		f, _ := val.Float64()
		return f
	default:
		return 0
	}
}

func toBool(v interface{}) bool {
	switch val := v.(type) {
	case bool:
		return val
	default:
		return false
	}
}

func toPositionsList(v interface{}) []map[string]interface{} {
	arr, ok := v.([]interface{})
	if !ok {
		return nil
	}
	var result []map[string]interface{}
	for _, item := range arr {
		if m, ok := item.(map[string]interface{}); ok {
			result = append(result, m)
		}
	}
	return result
}
