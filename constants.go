package clobclient

const (
	// Access levels
	L0 = 0
	L1 = 1
	L2 = 2

	L1AuthUnavailable      = "A private key is needed to interact with this endpoint!"
	L2AuthUnavailable      = "API Credentials are needed to interact with this endpoint!"
	BuilderAuthUnavailable = "Builder API Credentials needed to interact with this endpoint!"

	ZeroAddress = "0x0000000000000000000000000000000000000000"
	Bytes32Zero = "0x0000000000000000000000000000000000000000000000000000000000000000"

	ChainIDAmoy    = 80002
	ChainIDPolygon = 137

	InitialCursor = "MA=="
	EndCursor     = "LTE="

	OrderVersionMismatchError = "order_version_mismatch"

	BuilderFeesBPS = 10000

	CollateralTokenDecimals  = 6
	ConditionalTokenDecimals = 6

	SideBuyStr  = "BUY"
	SideSellStr = "SELL"
)
