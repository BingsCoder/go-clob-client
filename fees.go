package clobclient

import (
	"fmt"
	"math"
)

const (
	MinFeeSlippagePercentage = 1
	MaxFeeSlippagePercentage = 100
)

// ValidateFeeSlippage validates the fee slippage parameter.
func ValidateFeeSlippage(feeSlippage float64) error {
	if math.IsNaN(feeSlippage) || math.IsInf(feeSlippage, 0) ||
		feeSlippage < 0 || feeSlippage > MaxFeeSlippagePercentage ||
		(feeSlippage > 0 && feeSlippage < MinFeeSlippagePercentage) {
		return fmt.Errorf("fee_slippage must be 0 or a percentage between %d and %d", MinFeeSlippagePercentage, MaxFeeSlippagePercentage)
	}
	return nil
}

// AdjustBuyAmountForFees adjusts the buy amount accounting for platform and builder fees.
func AdjustBuyAmountForFees(
	amount, price, userUSDCBalance, feeRate, feeExponent, builderTakerFeeRate, feeSlippage float64,
) (float64, error) {
	if err := ValidateFeeSlippage(feeSlippage); err != nil {
		return 0, err
	}

	platformFeeRate := feeRate * math.Pow(price*(1-price), feeExponent)
	effectivePlatformFeeRate := platformFeeRate * (1 + feeSlippage/100)
	feeBaseAmount := math.Min(amount, userUSDCBalance)
	platformFee := (feeBaseAmount / price) * effectivePlatformFeeRate
	builderFee := feeBaseAmount * builderTakerFeeRate
	totalCost := amount + platformFee + builderFee

	if userUSDCBalance <= totalCost {
		result := userUSDCBalance - platformFee - builderFee
		if result < 0 {
			return 0, nil
		}
		return result, nil
	}
	return amount, nil
}
