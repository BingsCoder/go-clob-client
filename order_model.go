package clobclient

import "fmt"

// OrderDataV1 is the input data for building a V1 order.
type OrderDataV1 struct {
	Maker         string
	Taker         string
	TokenID       string
	MakerAmount   string
	TakerAmount   string
	Side          Side
	FeeRateBps    string
	Nonce         string
	SignerAddr    string
	Expiration    string
	SignatureType SignatureTypeV1
}

// OrderV1 is an unsigned V1 order ready for EIP-712 signing.
type OrderV1 struct {
	Salt          string          `json:"salt"`
	Maker         string          `json:"maker"`
	SignerAddr    string          `json:"signer"`
	Taker         string          `json:"taker"`
	TokenID       string          `json:"tokenId"`
	MakerAmount   string          `json:"makerAmount"`
	TakerAmount   string          `json:"takerAmount"`
	Expiration    string          `json:"expiration"`
	Nonce         string          `json:"nonce"`
	FeeRateBps    string          `json:"feeRateBps"`
	Side          Side            `json:"side"`
	SignatureType SignatureTypeV1 `json:"signatureType"`
}

// SignedOrderV1 is a signed V1 order including the EIP-712 signature.
type SignedOrderV1 struct {
	OrderV1
	Signature string `json:"signature"`
}

// OrderToJSONV1 converts a signed V1 order to API payload format.
func OrderToJSONV1(order *SignedOrderV1, owner, orderType string, postOnly, deferExec bool) map[string]interface{} {
	side := SideBuyStr
	if order.Side == SideSell {
		side = SideSellStr
	}
	return map[string]interface{}{
		"order": map[string]interface{}{
			"salt":          mustParseInt(order.Salt),
			"maker":         order.Maker,
			"signer":        order.SignerAddr,
			"taker":         order.Taker,
			"tokenId":       order.TokenID,
			"makerAmount":   order.MakerAmount,
			"takerAmount":   order.TakerAmount,
			"side":          side,
			"expiration":    order.Expiration,
			"nonce":         order.Nonce,
			"feeRateBps":    order.FeeRateBps,
			"signatureType": int(order.SignatureType),
			"signature":     order.Signature,
		},
		"owner":     owner,
		"orderType": orderType,
		"deferExec": deferExec,
		"postOnly":  postOnly,
	}
}

// OrderDataV2 is the input data for building a V2 order.
type OrderDataV2 struct {
	Maker         string
	TokenID       string
	MakerAmount   string
	TakerAmount   string
	Side          Side
	SignerAddr    string
	SignatureType SignatureTypeV2
	Timestamp     string
	Metadata      string
	Builder       string
	Expiration    string
}

// OrderV2 is an unsigned V2 order ready for EIP-712 signing.
type OrderV2 struct {
	Salt        string          `json:"salt"`
	Maker       string          `json:"maker"`
	SignerAddr  string          `json:"signer"`
	TokenID     string          `json:"tokenId"`
	MakerAmount string          `json:"makerAmount"`
	TakerAmount string          `json:"takerAmount"`
	Side        Side            `json:"side"`
	Sig         SignatureTypeV2 `json:"signatureType"`
	Timestamp   string          `json:"timestamp"`
	Metadata    string          `json:"metadata"`
	Builder     string          `json:"builder"`
	Expiration  string          `json:"expiration"`
}

// SignedOrderV2 is a signed V2 order including the EIP-712 signature.
type SignedOrderV2 struct {
	OrderV2
	Signature string `json:"signature"`
}

// OrderToJSONV2 converts a signed V2 order to API payload format.
func OrderToJSONV2(order *SignedOrderV2, owner, orderType string, postOnly, deferExec bool) map[string]interface{} {
	side := SideBuyStr
	if order.Side == SideSell {
		side = SideSellStr
	}
	return map[string]interface{}{
		"order": map[string]interface{}{
			"salt":          mustParseInt(order.Salt),
			"maker":         order.Maker,
			"signer":        order.SignerAddr,
			"tokenId":       order.TokenID,
			"makerAmount":   order.MakerAmount,
			"takerAmount":   order.TakerAmount,
			"side":          side,
			"expiration":    order.Expiration,
			"signatureType": int(order.Sig),
			"timestamp":     order.Timestamp,
			"metadata":      order.Metadata,
			"builder":       order.Builder,
			"signature":     order.Signature,
		},
		"owner":     owner,
		"orderType": orderType,
		"deferExec": deferExec,
		"postOnly":  postOnly,
	}
}

// IsV2Order returns true if the order has a Timestamp field (V2).
func IsV2Order(order interface{}) bool {
	switch order.(type) {
	case *SignedOrderV2:
		return true
	case SignedOrderV2:
		return true
	default:
		return false
	}
}

func mustParseInt(s string) int64 {
	var n int64
	fmt.Sscanf(s, "%d", &n)
	return n
}
