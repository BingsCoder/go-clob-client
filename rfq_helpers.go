package clobclient

import (
	"fmt"
	"net/url"
	"strings"
)

// ParseUnits converts a decimal string to smallest units.
func ParseUnits(value string, decimals int) int64 {
	if strings.Contains(value, ".") {
		parts := strings.SplitN(value, ".", 2)
		integerPart := parts[0]
		decimalPart := parts[1]
		// Truncate or pad decimal part
		if len(decimalPart) > decimals {
			decimalPart = decimalPart[:decimals]
		} else {
			for len(decimalPart) < decimals {
				decimalPart += "0"
			}
		}
		var result int64
		fmt.Sscanf(integerPart+decimalPart, "%d", &result)
		return result
	}
	var intVal int64
	fmt.Sscanf(value, "%d", &intVal)
	pow := int64(1)
	for i := 0; i < decimals; i++ {
		pow *= 10
	}
	return intVal * pow
}

// parseRFQRequestsParams converts GetRFQRequestsParams to URL query parameters.
func parseRFQRequestsParams(params *GetRFQRequestsParams) url.Values {
	if params == nil {
		return nil
	}
	q := url.Values{}
	if params.State != "" {
		q.Set("state", params.State)
	}
	if params.SizeMin != nil {
		q.Set("sizeMin", fmt.Sprintf("%v", *params.SizeMin))
	}
	if params.SizeMax != nil {
		q.Set("sizeMax", fmt.Sprintf("%v", *params.SizeMax))
	}
	if params.SizeUSDCMin != nil {
		q.Set("sizeUsdcMin", fmt.Sprintf("%v", *params.SizeUSDCMin))
	}
	if params.SizeUSDCMax != nil {
		q.Set("sizeUsdcMax", fmt.Sprintf("%v", *params.SizeUSDCMax))
	}
	if params.PriceMin != nil {
		q.Set("priceMin", fmt.Sprintf("%v", *params.PriceMin))
	}
	if params.PriceMax != nil {
		q.Set("priceMax", fmt.Sprintf("%v", *params.PriceMax))
	}
	if params.SortBy != "" {
		q.Set("sortBy", params.SortBy)
	}
	if params.SortDir != "" {
		q.Set("sortDir", params.SortDir)
	}
	if params.Limit != nil {
		q.Set("limit", fmt.Sprintf("%d", *params.Limit))
	}
	if params.Offset != "" {
		q.Set("offset", params.Offset)
	}
	for _, id := range params.RequestIDs {
		q.Add("requestIds", id)
	}
	for _, m := range params.Markets {
		q.Add("markets", m)
	}
	return q
}

// parseRFQQuotesParams converts GetRFQQuotesParams to URL query parameters.
func parseRFQQuotesParams(params *GetRFQQuotesParams) url.Values {
	if params == nil {
		return nil
	}
	q := url.Values{}
	if params.State != "" {
		q.Set("state", params.State)
	}
	if params.SizeMin != nil {
		q.Set("sizeMin", fmt.Sprintf("%v", *params.SizeMin))
	}
	if params.SizeMax != nil {
		q.Set("sizeMax", fmt.Sprintf("%v", *params.SizeMax))
	}
	if params.SizeUSDCMin != nil {
		q.Set("sizeUsdcMin", fmt.Sprintf("%v", *params.SizeUSDCMin))
	}
	if params.SizeUSDCMax != nil {
		q.Set("sizeUsdcMax", fmt.Sprintf("%v", *params.SizeUSDCMax))
	}
	if params.PriceMin != nil {
		q.Set("priceMin", fmt.Sprintf("%v", *params.PriceMin))
	}
	if params.PriceMax != nil {
		q.Set("priceMax", fmt.Sprintf("%v", *params.PriceMax))
	}
	if params.SortBy != "" {
		q.Set("sortBy", params.SortBy)
	}
	if params.SortDir != "" {
		q.Set("sortDir", params.SortDir)
	}
	if params.Limit != nil {
		q.Set("limit", fmt.Sprintf("%d", *params.Limit))
	}
	if params.Offset != "" {
		q.Set("offset", params.Offset)
	}
	for _, id := range params.QuoteIDs {
		q.Add("quoteIds", id)
	}
	for _, id := range params.RequestIDs {
		q.Add("requestIds", id)
	}
	for _, m := range params.Markets {
		q.Add("markets", m)
	}
	return q
}
