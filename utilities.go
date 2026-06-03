package clobclient

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"strconv"
)

// ParseRawOrderBookSummary converts a raw map to an OrderBookSummary.
func ParseRawOrderBookSummary(raw map[string]interface{}) *OrderBookSummary {
	obs := &OrderBookSummary{}

	if v, ok := raw["market"].(string); ok {
		obs.Market = v
	}
	if v, ok := raw["asset_id"].(string); ok {
		obs.AssetID = v
	}
	if v, ok := raw["timestamp"].(string); ok {
		obs.Timestamp = v
	}
	if v, ok := raw["last_trade_price"].(string); ok {
		obs.LastTradePrice = v
	}
	if v, ok := raw["min_order_size"].(string); ok {
		obs.MinOrderSize = v
	}
	if v, ok := raw["neg_risk"].(bool); ok {
		obs.NegRisk = &v
	}
	if v, ok := raw["tick_size"].(string); ok {
		obs.TickSize = v
	}
	if v, ok := raw["hash"].(string); ok {
		obs.Hash = v
	}

	if bids, ok := raw["bids"].([]interface{}); ok {
		for _, b := range bids {
			if bMap, ok := b.(map[string]interface{}); ok {
				obs.Bids = append(obs.Bids, OrderSummary{
					Size:  fmt.Sprintf("%v", bMap["size"]),
					Price: fmt.Sprintf("%v", bMap["price"]),
				})
			}
		}
	}

	if asks, ok := raw["asks"].([]interface{}); ok {
		for _, a := range asks {
			if aMap, ok := a.(map[string]interface{}); ok {
				obs.Asks = append(obs.Asks, OrderSummary{
					Size:  fmt.Sprintf("%v", aMap["size"]),
					Price: fmt.Sprintf("%v", aMap["price"]),
				})
			}
		}
	}

	return obs
}

// GenerateOrderBookSummaryHash computes the server-compatible SHA1 hash of an order book.
func GenerateOrderBookSummaryHash(ob *OrderBookSummary) string {
	bids := make([]map[string]string, 0)
	for _, b := range ob.Bids {
		bids = append(bids, map[string]string{"price": b.Price, "size": b.Size})
	}
	asks := make([]map[string]string, 0)
	for _, a := range ob.Asks {
		asks = append(asks, map[string]string{"price": a.Price, "size": a.Size})
	}

	payload := map[string]interface{}{
		"market":           ob.Market,
		"asset_id":         ob.AssetID,
		"timestamp":        ob.Timestamp,
		"hash":             "",
		"bids":             bids,
		"asks":             asks,
		"min_order_size":   ob.MinOrderSize,
		"tick_size":        ob.TickSize,
		"neg_risk":         ob.NegRisk,
		"last_trade_price": ob.LastTradePrice,
	}

	serialized, _ := json.Marshal(payload)
	h := sha1.Sum(serialized)
	hash := fmt.Sprintf("%x", h)
	ob.Hash = hash
	return hash
}

// IsTickSizeSmaller returns true if a < b as tick sizes.
func IsTickSizeSmaller(a, b TickSize) bool {
	aF, _ := strconv.ParseFloat(a, 64)
	bF, _ := strconv.ParseFloat(b, 64)
	return aF < bF
}

// PriceValid returns true if the price is valid for the given tick size.
func PriceValid(price float64, tickSize TickSize) bool {
	ts, _ := strconv.ParseFloat(tickSize, 64)
	return ts <= price && price <= 1-ts
}
