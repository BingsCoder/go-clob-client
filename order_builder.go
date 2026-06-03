package clobclient

import (
	"fmt"
	"math"
	"strconv"
	"time"
)

// ROUNDING_CONFIG maps tick sizes to rounding configurations.
var RoundingConfig = map[string]RoundConfig{
	"0.1":    {Price: 1, Size: 2, Amount: 3},
	"0.01":   {Price: 2, Size: 2, Amount: 4},
	"0.001":  {Price: 3, Size: 2, Amount: 5},
	"0.0001": {Price: 4, Size: 2, Amount: 6},
}

// OrderBuilder creates and signs orders.
type OrderBuilder struct {
	Signer        *Signer
	SignatureType SignatureTypeV2
	Funder        string
}

// NewOrderBuilder creates a new OrderBuilder.
func NewOrderBuilder(signer *Signer, signatureType *SignatureTypeV2, funder string) *OrderBuilder {
	sigType := SignatureTypeV2EOA
	if signatureType != nil {
		sigType = *signatureType
	}

	funderAddr := funder
	if funderAddr == "" && signer != nil {
		funderAddr = signer.Address()
	}

	return &OrderBuilder{
		Signer:        signer,
		SignatureType: sigType,
		Funder:        funderAddr,
	}
}

func (ob *OrderBuilder) v2OrderSigner() string {
	if ob.SignatureType == SignatureTypeV2Poly1271 {
		return ob.Funder
	}
	return ob.Signer.Address()
}

// GetOrderAmounts calculates maker/taker amounts for a limit order.
func (ob *OrderBuilder) GetOrderAmounts(side string, size, price float64, rc RoundConfig) (Side, int, int, error) {
	rawPrice := roundNormal(price, int(rc.Price))

	if side == SideBuyStr || side == "BUY" {
		rawTakerAmt := roundDown(size, int(rc.Size))
		rawMakerAmt := rawTakerAmt * rawPrice
		if decimalPlaces(rawMakerAmt) > int(rc.Amount) {
			rawMakerAmt = roundUp(rawMakerAmt, int(rc.Amount)+4)
			if decimalPlaces(rawMakerAmt) > int(rc.Amount) {
				rawMakerAmt = roundDown(rawMakerAmt, int(rc.Amount))
			}
		}
		return SideBuy, toTokenDecimals(rawMakerAmt), toTokenDecimals(rawTakerAmt), nil
	} else if side == SideSellStr || side == "SELL" {
		rawMakerAmt := roundDown(size, int(rc.Size))
		rawTakerAmt := rawMakerAmt * rawPrice
		if decimalPlaces(rawTakerAmt) > int(rc.Amount) {
			rawTakerAmt = roundUp(rawTakerAmt, int(rc.Amount)+4)
			if decimalPlaces(rawTakerAmt) > int(rc.Amount) {
				rawTakerAmt = roundDown(rawTakerAmt, int(rc.Amount))
			}
		}
		return SideSell, toTokenDecimals(rawMakerAmt), toTokenDecimals(rawTakerAmt), nil
	}
	return 0, 0, 0, fmt.Errorf("order side must be 'BUY' or 'SELL'")
}

// GetMarketOrderAmounts calculates maker/taker amounts for a market order.
func (ob *OrderBuilder) GetMarketOrderAmounts(side string, amount, price float64, rc RoundConfig) (Side, int, int, error) {
	rawPrice := roundDown(price, int(rc.Price))

	if side == SideBuyStr || side == "BUY" {
		rawMakerAmt := roundDown(amount, int(rc.Size))
		rawTakerAmt := rawMakerAmt / rawPrice
		if decimalPlaces(rawTakerAmt) > int(rc.Amount) {
			rawTakerAmt = roundUp(rawTakerAmt, int(rc.Amount)+4)
			if decimalPlaces(rawTakerAmt) > int(rc.Amount) {
				rawTakerAmt = roundDown(rawTakerAmt, int(rc.Amount))
			}
		}
		return SideBuy, toTokenDecimals(rawMakerAmt), toTokenDecimals(rawTakerAmt), nil
	} else if side == SideSellStr || side == "SELL" {
		rawMakerAmt := roundDown(amount, int(rc.Size))
		rawTakerAmt := rawMakerAmt * rawPrice
		if decimalPlaces(rawTakerAmt) > int(rc.Amount) {
			rawTakerAmt = roundUp(rawTakerAmt, int(rc.Amount)+4)
			if decimalPlaces(rawTakerAmt) > int(rc.Amount) {
				rawTakerAmt = roundDown(rawTakerAmt, int(rc.Amount))
			}
		}
		return SideSell, toTokenDecimals(rawMakerAmt), toTokenDecimals(rawTakerAmt), nil
	}
	return 0, 0, 0, fmt.Errorf("order side must be 'BUY' or 'SELL'")
}

// BuildOrder creates and signs a limit order.
func (ob *OrderBuilder) BuildOrder(args *OrderArgs, options *CreateOrderOptions, version int, feeRateBps *int) (interface{}, error) {
	rc := RoundingConfig[options.TickSize]
	side, makerAmount, takerAmount, err := ob.GetOrderAmounts(args.Side, args.Size, args.Price, rc)
	if err != nil {
		return nil, err
	}

	config, err := GetContractConfig(ob.Signer.GetChainID())
	if err != nil {
		return nil, err
	}
	ts := fmt.Sprintf("%d", time.Now().UnixMilli())

	if version == 1 {
		if ob.SignatureType == SignatureTypeV2Poly1271 {
			return nil, fmt.Errorf("signature type POLY_1271 is not supported for v1 orders")
		}

		exchangeAddr := config.Exchange
		if options.NegRisk {
			exchangeAddr = config.NegRiskExchange
		}

		resolvedFeeRate := "0"
		if feeRateBps != nil {
			resolvedFeeRate = fmt.Sprintf("%d", *feeRateBps)
		}

		taker := ZeroAddress
		nonce := "0"
		expiration := fmt.Sprintf("%d", args.Expiration)

		data := &OrderDataV1{
			Maker:         ob.Funder,
			Taker:         taker,
			TokenID:       args.TokenID,
			MakerAmount:   fmt.Sprintf("%d", makerAmount),
			TakerAmount:   fmt.Sprintf("%d", takerAmount),
			Side:          side,
			FeeRateBps:    resolvedFeeRate,
			Nonce:         nonce,
			SignerAddr:    ob.Signer.Address(),
			Expiration:    expiration,
			SignatureType: SignatureTypeV1(int(ob.SignatureType)),
		}

		builder := NewExchangeOrderBuilderV1(exchangeAddr, ob.Signer.GetChainID(), ob.Signer)
		return builder.BuildSignedOrder(data)

	} else if version == 2 {
		exchangeAddr := config.ExchangeV2
		if options.NegRisk {
			exchangeAddr = config.NegRiskExchangeV2
		}

		metadata := args.Metadata
		if metadata == "" {
			metadata = Bytes32Zero
		}
		builderCode := args.BuilderCode
		if builderCode == "" {
			builderCode = Bytes32Zero
		}

		data := &OrderDataV2{
			Maker:         ob.Funder,
			TokenID:       args.TokenID,
			MakerAmount:   fmt.Sprintf("%d", makerAmount),
			TakerAmount:   fmt.Sprintf("%d", takerAmount),
			Side:          side,
			SignerAddr:    ob.v2OrderSigner(),
			SignatureType: ob.SignatureType,
			Timestamp:     ts,
			Metadata:      metadata,
			Builder:       builderCode,
			Expiration:    fmt.Sprintf("%d", args.Expiration),
		}

		builder := NewExchangeOrderBuilderV2(exchangeAddr, ob.Signer.GetChainID(), ob.Signer)
		return builder.BuildSignedOrder(data)
	}

	return nil, fmt.Errorf("unsupported order version %d", version)
}

// BuildOrderV1 creates and signs a V1 limit order from V1 args.
func (ob *OrderBuilder) BuildOrderV1(args *OrderArgsV1, options *CreateOrderOptions, feeRateBps *int) (*SignedOrderV1, error) {
	rc := RoundingConfig[options.TickSize]
	side, makerAmount, takerAmount, err := ob.GetOrderAmounts(args.Side, args.Size, args.Price, rc)
	if err != nil {
		return nil, err
	}

	config, err := GetContractConfig(ob.Signer.GetChainID())
	if err != nil {
		return nil, err
	}

	if ob.SignatureType == SignatureTypeV2Poly1271 {
		return nil, fmt.Errorf("signature type POLY_1271 is not supported for v1 orders")
	}

	exchangeAddr := config.Exchange
	if options.NegRisk {
		exchangeAddr = config.NegRiskExchange
	}

	resolvedFeeRate := fmt.Sprintf("%d", args.FeeRateBps)
	if feeRateBps != nil {
		resolvedFeeRate = fmt.Sprintf("%d", *feeRateBps)
	}

	taker := args.Taker
	if taker == "" {
		taker = ZeroAddress
	}

	data := &OrderDataV1{
		Maker:         ob.Funder,
		Taker:         taker,
		TokenID:       args.TokenID,
		MakerAmount:   fmt.Sprintf("%d", makerAmount),
		TakerAmount:   fmt.Sprintf("%d", takerAmount),
		Side:          side,
		FeeRateBps:    resolvedFeeRate,
		Nonce:         fmt.Sprintf("%d", args.Nonce),
		SignerAddr:    ob.Signer.Address(),
		Expiration:    fmt.Sprintf("%d", args.Expiration),
		SignatureType: SignatureTypeV1(int(ob.SignatureType)),
	}

	builder := NewExchangeOrderBuilderV1(exchangeAddr, ob.Signer.GetChainID(), ob.Signer)
	return builder.BuildSignedOrder(data)
}

// BuildMarketOrder creates and signs a market order.
func (ob *OrderBuilder) BuildMarketOrder(args *MarketOrderArgs, options *CreateOrderOptions, version int, feeRateBps *int) (interface{}, error) {
	rc := RoundingConfig[options.TickSize]
	side, makerAmount, takerAmount, err := ob.GetMarketOrderAmounts(args.Side, args.Amount, args.Price, rc)
	if err != nil {
		return nil, err
	}

	config, err := GetContractConfig(ob.Signer.GetChainID())
	if err != nil {
		return nil, err
	}
	ts := fmt.Sprintf("%d", time.Now().UnixMilli())

	if version == 1 {
		if ob.SignatureType == SignatureTypeV2Poly1271 {
			return nil, fmt.Errorf("signature type POLY_1271 is not supported for v1 orders")
		}

		exchangeAddr := config.Exchange
		if options.NegRisk {
			exchangeAddr = config.NegRiskExchange
		}

		resolvedFeeRate := "0"
		if feeRateBps != nil {
			resolvedFeeRate = fmt.Sprintf("%d", *feeRateBps)
		}

		data := &OrderDataV1{
			Maker:         ob.Funder,
			Taker:         ZeroAddress,
			TokenID:       args.TokenID,
			MakerAmount:   fmt.Sprintf("%d", makerAmount),
			TakerAmount:   fmt.Sprintf("%d", takerAmount),
			Side:          side,
			FeeRateBps:    resolvedFeeRate,
			Nonce:         "0",
			SignerAddr:    ob.Signer.Address(),
			Expiration:    "0",
			SignatureType: SignatureTypeV1(int(ob.SignatureType)),
		}

		builder := NewExchangeOrderBuilderV1(exchangeAddr, ob.Signer.GetChainID(), ob.Signer)
		return builder.BuildSignedOrder(data)

	} else if version == 2 {
		exchangeAddr := config.ExchangeV2
		if options.NegRisk {
			exchangeAddr = config.NegRiskExchangeV2
		}

		metadata := args.Metadata
		if metadata == "" {
			metadata = Bytes32Zero
		}
		builderCode := args.BuilderCode
		if builderCode == "" {
			builderCode = Bytes32Zero
		}

		data := &OrderDataV2{
			Maker:         ob.Funder,
			TokenID:       args.TokenID,
			MakerAmount:   fmt.Sprintf("%d", makerAmount),
			TakerAmount:   fmt.Sprintf("%d", takerAmount),
			Side:          side,
			SignerAddr:    ob.v2OrderSigner(),
			SignatureType: ob.SignatureType,
			Timestamp:     ts,
			Metadata:      metadata,
			Builder:       builderCode,
		}

		builder := NewExchangeOrderBuilderV2(exchangeAddr, ob.Signer.GetChainID(), ob.Signer)
		return builder.BuildSignedOrder(data)
	}

	return nil, fmt.Errorf("unsupported order version %d", version)
}

// CalculateBuyMarketPrice calculates the market price for a buy order.
func (ob *OrderBuilder) CalculateBuyMarketPrice(positions []map[string]interface{}, amountToMatch float64, orderType string) (float64, error) {
	if len(positions) == 0 {
		return 0, fmt.Errorf("no match")
	}

	total := 0.0
	for i := len(positions) - 1; i >= 0; i-- {
		p := positions[i]
		size, _ := strconv.ParseFloat(fmt.Sprintf("%v", p["size"]), 64)
		price, _ := strconv.ParseFloat(fmt.Sprintf("%v", p["price"]), 64)
		total += size * price
		if total >= amountToMatch {
			return price, nil
		}
	}

	if orderType == OrderTypeFOK {
		return 0, fmt.Errorf("no match")
	}

	price, _ := strconv.ParseFloat(fmt.Sprintf("%v", positions[0]["price"]), 64)
	return price, nil
}

// CalculateSellMarketPrice calculates the market price for a sell order.
func (ob *OrderBuilder) CalculateSellMarketPrice(positions []map[string]interface{}, amountToMatch float64, orderType string) (float64, error) {
	if len(positions) == 0 {
		return 0, fmt.Errorf("no match")
	}

	total := 0.0
	for i := len(positions) - 1; i >= 0; i-- {
		p := positions[i]
		size, _ := strconv.ParseFloat(fmt.Sprintf("%v", p["size"]), 64)
		total += size
		if total >= amountToMatch {
			price, _ := strconv.ParseFloat(fmt.Sprintf("%v", p["price"]), 64)
			return price, nil
		}
	}

	if orderType == OrderTypeFOK {
		return 0, fmt.Errorf("no match")
	}

	price, _ := strconv.ParseFloat(fmt.Sprintf("%v", positions[0]["price"]), 64)
	return price, nil
}

// Rounding helpers

func roundDown(x float64, sigDigits int) float64 {
	multiplier := math.Pow(10, float64(sigDigits))
	return math.Floor(x*multiplier) / multiplier
}

func roundNormal(x float64, sigDigits int) float64 {
	multiplier := math.Pow(10, float64(sigDigits))
	return math.Round(x*multiplier) / multiplier
}

func roundUp(x float64, sigDigits int) float64 {
	multiplier := math.Pow(10, float64(sigDigits))
	return math.Ceil(x*multiplier) / multiplier
}

func toTokenDecimals(x float64) int {
	f := 1e6 * x
	if decimalPlaces(f) > 0 {
		f = roundNormal(f, 0)
	}
	return int(f)
}

func decimalPlaces(x float64) int {
	s := strconv.FormatFloat(x, 'f', -1, 64)
	i := len(s) - 1
	for i >= 0 && s[i] != '.' {
		i--
	}
	if i < 0 {
		return 0
	}
	return len(s) - i - 1
}
