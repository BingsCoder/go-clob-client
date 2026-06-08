package clobclient

import (
	"encoding/json"
	"fmt"
	"net/url"
)

// RFQClient handles all RFQ operations.
type RFQClient struct {
	parent *ClobClient
}

// NewRFQClient creates a new RFQ client.
func NewRFQClient(parent *ClobClient) *RFQClient {
	return &RFQClient{parent: parent}
}

func (r *RFQClient) getL2Headers(method, endpoint string, body interface{}, serializedBody string) (map[string]string, error) {
	args := &RequestArgs{
		Method:         method,
		RequestPath:    endpoint,
		Body:           body,
		SerializedBody: serializedBody,
	}
	return CreateLevel2Headers(r.parent.Signer, r.parent.Creds, args, r.parent.getTimestamp())
}

func (r *RFQClient) buildURL(endpoint string) string {
	return r.parent.Host + endpoint
}

// CreateRFQRequest creates and posts an RFQ request.
func (r *RFQClient) CreateRFQRequest(userRequest *RFQUserRequest, options *PartialCreateOrderOptions) (interface{}, error) {
	if err := r.parent.AssertLevel2Auth(); err != nil {
		return nil, err
	}

	tokenID := userRequest.TokenID
	tickSize, err := r.parent.resolveTickSize(tokenID, options)
	if err != nil {
		return nil, err
	}

	rc := RoundingConfig[tickSize]
	roundedPrice := roundNormal(userRequest.Price, int(rc.Price))
	roundedSize := roundDown(userRequest.Size, int(rc.Size))

	sizeDecimals := int(rc.Size)
	amountDecimals := int(rc.Amount)

	roundedSizeStr := fmt.Sprintf("%.*f", sizeDecimals, roundedSize)

	sizeNum := roundedSize
	priceNum := roundedPrice

	userType := int(r.parent.Builder.SignatureType)

	var assetIn, assetOut string
	var amountIn, amountOut int64

	if userRequest.Side == SideBuyStr {
		amountIn = ParseUnits(roundedSizeStr, CollateralTokenDecimals)
		usdcAmountStr := fmt.Sprintf("%.*f", amountDecimals, sizeNum*priceNum)
		amountOut = ParseUnits(usdcAmountStr, CollateralTokenDecimals)
		assetIn = tokenID
		assetOut = "0"
	} else {
		usdcAmountStr := fmt.Sprintf("%.*f", amountDecimals, sizeNum*priceNum)
		amountIn = ParseUnits(usdcAmountStr, CollateralTokenDecimals)
		amountOut = ParseUnits(roundedSizeStr, CollateralTokenDecimals)
		assetIn = "0"
		assetOut = tokenID
	}

	body := map[string]interface{}{
		"assetIn":  assetIn,
		"assetOut": assetOut,
		"amountIn": fmt.Sprintf("%d", amountIn),
		"amountOut": fmt.Sprintf("%d", amountOut),
		"userType": userType,
	}
	serialized := compactJSON(body)
	headers, err := r.getL2Headers("POST", EndpointCreateRFQRequest, body, serialized)
	if err != nil {
		return nil, err
	}
	return httpPost(r.buildURL(EndpointCreateRFQRequest), headers, serialized, nil, false)
}

// CancelRFQRequest cancels an RFQ request.
func (r *RFQClient) CancelRFQRequest(params *CancelRFQRequestParams) (interface{}, error) {
	if err := r.parent.AssertLevel2Auth(); err != nil {
		return nil, err
	}
	body := map[string]string{"requestId": params.RequestID}
	serialized := compactJSON(body)
	headers, err := r.getL2Headers("DELETE", EndpointCancelRFQRequest, body, serialized)
	if err != nil {
		return nil, err
	}
	return httpDelete(r.buildURL(EndpointCancelRFQRequest), headers, serialized, nil)
}

// GetRFQRequests returns RFQ requests with optional filtering.
func (r *RFQClient) GetRFQRequests(params *GetRFQRequestsParams) (interface{}, error) {
	if err := r.parent.AssertLevel2Auth(); err != nil {
		return nil, err
	}
	headers, err := r.getL2Headers("GET", EndpointGetRFQRequests, nil, "")
	if err != nil {
		return nil, err
	}

	reqURL := r.buildURL(EndpointGetRFQRequests)
	queryParams := parseRFQRequestsParams(params)
	if queryParams != nil && len(queryParams) > 0 {
		reqURL = reqURL + "?" + queryParams.Encode()
	}

	return httpGet(reqURL, headers, nil)
}

// CreateRFQQuote creates and posts an RFQ quote.
func (r *RFQClient) CreateRFQQuote(userQuote *RFQUserQuote, options *PartialCreateOrderOptions) (interface{}, error) {
	if err := r.parent.AssertLevel2Auth(); err != nil {
		return nil, err
	}

	tokenID := userQuote.TokenID
	tickSize, err := r.parent.resolveTickSize(tokenID, options)
	if err != nil {
		return nil, err
	}

	rc := RoundingConfig[tickSize]
	roundedPrice := roundNormal(userQuote.Price, int(rc.Price))
	roundedSize := roundDown(userQuote.Size, int(rc.Size))

	sizeDecimals := int(rc.Size)
	amountDecimals := int(rc.Amount)

	roundedSizeStr := fmt.Sprintf("%.*f", sizeDecimals, roundedSize)

	sizeNum := roundedSize
	priceNum := roundedPrice

	userType := int(r.parent.Builder.SignatureType)

	var assetIn, assetOut string
	var amountIn, amountOut int64

	if userQuote.Side == SideBuyStr {
		amountIn = ParseUnits(roundedSizeStr, CollateralTokenDecimals)
		usdcAmountStr := fmt.Sprintf("%.*f", amountDecimals, sizeNum*priceNum)
		amountOut = ParseUnits(usdcAmountStr, CollateralTokenDecimals)
		assetIn = tokenID
		assetOut = "0"
	} else {
		usdcAmountStr := fmt.Sprintf("%.*f", amountDecimals, sizeNum*priceNum)
		amountIn = ParseUnits(usdcAmountStr, CollateralTokenDecimals)
		amountOut = ParseUnits(roundedSizeStr, CollateralTokenDecimals)
		assetIn = "0"
		assetOut = tokenID
	}

	body := map[string]interface{}{
		"requestId": userQuote.RequestID,
		"assetIn":   assetIn,
		"assetOut":  assetOut,
		"amountIn":  fmt.Sprintf("%d", amountIn),
		"amountOut": fmt.Sprintf("%d", amountOut),
		"userType":  userType,
	}
	serialized := compactJSON(body)
	headers, err := r.getL2Headers("POST", EndpointCreateRFQQuote, body, serialized)
	if err != nil {
		return nil, err
	}
	return httpPost(r.buildURL(EndpointCreateRFQQuote), headers, serialized, nil, false)
}

// GetRFQRequesterQuotes returns quotes on requests created by the authenticated user.
func (r *RFQClient) GetRFQRequesterQuotes(params *GetRFQQuotesParams) (interface{}, error) {
	if err := r.parent.AssertLevel2Auth(); err != nil {
		return nil, err
	}
	headers, err := r.getL2Headers("GET", EndpointGetRFQRequesterQuotes, nil, "")
	if err != nil {
		return nil, err
	}

	reqURL := r.buildURL(EndpointGetRFQRequesterQuotes)
	queryParams := parseRFQQuotesParams(params)
	if queryParams != nil && len(queryParams) > 0 {
		reqURL = reqURL + "?" + queryParams.Encode()
	}

	return httpGet(reqURL, headers, nil)
}

// GetRFQQuoterQuotes returns quotes created by the authenticated user.
func (r *RFQClient) GetRFQQuoterQuotes(params *GetRFQQuotesParams) (interface{}, error) {
	if err := r.parent.AssertLevel2Auth(); err != nil {
		return nil, err
	}
	headers, err := r.getL2Headers("GET", EndpointGetRFQQuoterQuotes, nil, "")
	if err != nil {
		return nil, err
	}

	reqURL := r.buildURL(EndpointGetRFQQuoterQuotes)
	queryParams := parseRFQQuotesParams(params)
	if queryParams != nil && len(queryParams) > 0 {
		reqURL = reqURL + "?" + queryParams.Encode()
	}

	return httpGet(reqURL, headers, nil)
}

// GetRFQBestQuote returns the best quote for an RFQ request.
func (r *RFQClient) GetRFQBestQuote(params *GetRFQBestQuoteParams) (interface{}, error) {
	if err := r.parent.AssertLevel2Auth(); err != nil {
		return nil, err
	}
	headers, err := r.getL2Headers("GET", EndpointGetRFQBestQuote, nil, "")
	if err != nil {
		return nil, err
	}

	reqURL := r.buildURL(EndpointGetRFQBestQuote)
	if params != nil && params.RequestID != "" {
		q := url.Values{}
		q.Set("requestId", params.RequestID)
		reqURL = reqURL + "?" + q.Encode()
	}

	return httpGet(reqURL, headers, nil)
}

// CancelRFQQuote cancels an RFQ quote.
func (r *RFQClient) CancelRFQQuote(params *CancelRFQQuoteParams) (interface{}, error) {
	if err := r.parent.AssertLevel2Auth(); err != nil {
		return nil, err
	}
	body := map[string]string{"quoteId": params.QuoteID}
	serialized := compactJSON(body)
	headers, err := r.getL2Headers("DELETE", EndpointCancelRFQQuote, body, serialized)
	if err != nil {
		return nil, err
	}
	return httpDelete(r.buildURL(EndpointCancelRFQQuote), headers, serialized, nil)
}

// AcceptRFQQuote accepts an RFQ quote (requester side).
func (r *RFQClient) AcceptRFQQuote(params *AcceptQuoteParams) (interface{}, error) {
	if err := r.parent.AssertLevel2Auth(); err != nil {
		return nil, err
	}

	resp, err := r.GetRFQRequesterQuotes(&GetRFQQuotesParams{QuoteIDs: []string{params.QuoteID}})
	if err != nil {
		return nil, err
	}

	respMap, ok := resp.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected response type")
	}
	data, ok := respMap["data"].([]interface{})
	if !ok || len(data) == 0 {
		return nil, fmt.Errorf("RFQ quote not found")
	}

	rfqQuote := data[0].(map[string]interface{})
	orderPayload, err := r.getRequestOrderCreationPayload(rfqQuote)
	if err != nil {
		return nil, err
	}

	order, err := r.buildV1Order(&OrderArgsV1{
		TokenID:    orderPayload["token"].(string),
		Price:      orderPayload["price"].(float64),
		Size:       orderPayload["size"].(float64),
		Side:       orderPayload["side"].(string),
		Expiration: params.Expiration,
	})
	if err != nil {
		return nil, fmt.Errorf("error creating order: %w", err)
	}

	acceptPayload := map[string]interface{}{
		"requestId":     params.RequestID,
		"quoteId":       params.QuoteID,
		"owner":         r.parent.Creds.APIKey,
		"salt":          mustParseInt(order.Salt),
		"maker":         order.Maker,
		"signer":        order.SignerAddr,
		"taker":         order.Taker,
		"tokenId":       order.TokenID,
		"makerAmount":   order.MakerAmount,
		"takerAmount":   order.TakerAmount,
		"expiration":    mustParseInt(order.Expiration),
		"nonce":         order.Nonce,
		"feeRateBps":    order.FeeRateBps,
		"side":          orderPayload["side"],
		"signatureType": int(order.SignatureType),
		"signature":     order.Signature,
	}

	serialized := compactJSON(acceptPayload)
	headers, err := r.getL2Headers("POST", EndpointRFQRequestsAccept, acceptPayload, serialized)
	if err != nil {
		return nil, err
	}
	return httpPost(r.buildURL(EndpointRFQRequestsAccept), headers, serialized, nil, false)
}

// ApproveRFQOrder approves an RFQ order (quoter side).
func (r *RFQClient) ApproveRFQOrder(params *ApproveOrderParams) (interface{}, error) {
	if err := r.parent.AssertLevel2Auth(); err != nil {
		return nil, err
	}

	resp, err := r.GetRFQQuoterQuotes(&GetRFQQuotesParams{QuoteIDs: []string{params.QuoteID}})
	if err != nil {
		return nil, err
	}

	respMap, ok := resp.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected response type")
	}
	data, ok := respMap["data"].([]interface{})
	if !ok || len(data) == 0 {
		return nil, fmt.Errorf("RFQ quote not found")
	}

	rfqQuote := data[0].(map[string]interface{})
	side, _ := rfqQuote["side"].(string)
	if side == "" {
		side = SideBuyStr
	}

	var size string
	if side == SideBuyStr {
		size = fmt.Sprintf("%v", rfqQuote["sizeIn"])
	} else {
		size = fmt.Sprintf("%v", rfqQuote["sizeOut"])
	}

	tokenID := fmt.Sprintf("%v", rfqQuote["token"])
	price := toFloat64(rfqQuote["price"])
	sizeF := toFloat64(size)

	order, err := r.buildV1Order(&OrderArgsV1{
		TokenID:    tokenID,
		Price:      price,
		Size:       sizeF,
		Side:       side,
		Expiration: params.Expiration,
	})
	if err != nil {
		return nil, fmt.Errorf("error creating order: %w", err)
	}

	approvePayload := map[string]interface{}{
		"requestId":     params.RequestID,
		"quoteId":       params.QuoteID,
		"owner":         r.parent.Creds.APIKey,
		"salt":          mustParseInt(order.Salt),
		"maker":         order.Maker,
		"signer":        order.SignerAddr,
		"taker":         order.Taker,
		"tokenId":       order.TokenID,
		"makerAmount":   order.MakerAmount,
		"takerAmount":   order.TakerAmount,
		"expiration":    mustParseInt(order.Expiration),
		"nonce":         order.Nonce,
		"feeRateBps":    order.FeeRateBps,
		"side":          side,
		"signatureType": int(order.SignatureType),
		"signature":     order.Signature,
	}

	serialized := compactJSON(approvePayload)
	headers, err := r.getL2Headers("POST", EndpointRFQQuoteApprove, approvePayload, serialized)
	if err != nil {
		return nil, err
	}
	return httpPost(r.buildURL(EndpointRFQQuoteApprove), headers, serialized, nil, false)
}

// RFQConfig returns RFQ configuration.
func (r *RFQClient) RFQConfig() (interface{}, error) {
	if err := r.parent.AssertLevel2Auth(); err != nil {
		return nil, err
	}
	headers, err := r.getL2Headers("GET", EndpointRFQConfig, nil, "")
	if err != nil {
		return nil, err
	}
	return httpGet(r.buildURL(EndpointRFQConfig), headers, nil)
}

// Private helpers

func (r *RFQClient) buildV1Order(args *OrderArgsV1) (*SignedOrderV1, error) {
	tickSize, err := r.parent.resolveTickSize(args.TokenID, nil)
	if err != nil {
		return nil, err
	}
	negRisk, err := r.parent.GetNegRisk(args.TokenID)
	if err != nil {
		return nil, err
	}
	return r.parent.Builder.BuildOrderV1(args, &CreateOrderOptions{TickSize: tickSize, NegRisk: negRisk}, nil)
}

func (r *RFQClient) getRequestOrderCreationPayload(quote map[string]interface{}) (map[string]interface{}, error) {
	rawMatchType, _ := quote["matchType"].(string)
	if rawMatchType == "" {
		rawMatchType = string(MatchTypeComplementary)
	}
	matchType := MatchType(rawMatchType)

	side, _ := quote["side"].(string)
	if side == "" {
		side = SideBuyStr
	}

	switch matchType {
	case MatchTypeComplementary:
		token, _ := quote["token"].(string)
		if token == "" {
			return nil, fmt.Errorf("missing token for COMPLEMENTARY match")
		}
		if side == SideBuyStr {
			side = SideSellStr
		} else {
			side = SideBuyStr
		}
		var size string
		if side == SideBuyStr {
			size = fmt.Sprintf("%v", quote["sizeOut"])
		} else {
			size = fmt.Sprintf("%v", quote["sizeIn"])
		}
		price := toFloat64(quote["price"])
		return map[string]interface{}{
			"token": token,
			"side":  side,
			"size":  toFloat64(size),
			"price": price,
		}, nil

	case MatchTypeMint, MatchTypeMerge:
		token, _ := quote["complement"].(string)
		if token == "" {
			return nil, fmt.Errorf("missing complement token for MINT/MERGE match")
		}
		var size string
		if side == SideBuyStr {
			size = fmt.Sprintf("%v", quote["sizeIn"])
		} else {
			size = fmt.Sprintf("%v", quote["sizeOut"])
		}
		price := toFloat64(quote["price"])
		return map[string]interface{}{
			"token": token,
			"side":  side,
			"size":  toFloat64(size),
			"price": 1 - price,
		}, nil

	default:
		return nil, fmt.Errorf("invalid match type: %s", rawMatchType)
	}
}

func compactJSON(v interface{}) string {
	b, _ := json.Marshal(v)
	return string(b)
}
