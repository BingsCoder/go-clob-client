package clobclient

const (
	EndpointOK      = "/ok"
	EndpointTime    = "/time"
	EndpointVersion = "/version"

	// API Key endpoints
	EndpointCreateAPIKey  = "/auth/api-key"
	EndpointGetAPIKeys    = "/auth/api-keys"
	EndpointDeleteAPIKey  = "/auth/api-key"
	EndpointDeriveAPIKey  = "/auth/derive-api-key"
	EndpointClosedOnly    = "/auth/ban-status/closed-only"

	// Readonly API Key endpoints
	EndpointCreateReadonlyAPIKey = "/auth/readonly-api-key"
	EndpointGetReadonlyAPIKeys   = "/auth/readonly-api-keys"
	EndpointDeleteReadonlyAPIKey = "/auth/readonly-api-key"

	// Builder API Key endpoints
	EndpointCreateBuilderAPIKey = "/auth/builder-api-key"
	EndpointGetBuilderAPIKeys   = "/auth/builder-api-key"
	EndpointRevokeBuilderAPIKey = "/auth/builder-api-key"

	// Live activity
	EndpointGetMarketTradesEvents = "/markets/live-activity/"

	// Markets
	EndpointGetSamplingSimplifiedMarkets = "/sampling-simplified-markets"
	EndpointGetSamplingMarkets           = "/sampling-markets"
	EndpointGetSimplifiedMarkets         = "/simplified-markets"
	EndpointGetMarkets                   = "/markets"
	EndpointGetMarket                    = "/markets/"
	EndpointGetMarketByToken             = "/markets-by-token/"
	EndpointGetClobMarket                = "/clob-markets/"

	// Order Book
	EndpointGetOrderBook  = "/book"
	EndpointGetOrderBooks = "/books"

	// Pricing
	EndpointGetMidpoint         = "/midpoint"
	EndpointGetMidpoints        = "/midpoints"
	EndpointGetPrice            = "/price"
	EndpointGetPrices           = "/prices"
	EndpointGetSpread           = "/spread"
	EndpointGetSpreads          = "/spreads"
	EndpointGetLastTradePrice   = "/last-trade-price"
	EndpointGetLastTradesPrices = "/last-trades-prices"

	// Market parameters
	EndpointGetTickSize = "/tick-size"
	EndpointGetNegRisk  = "/neg-risk"
	EndpointGetFeeRate  = "/fee-rate"

	// Price history
	EndpointGetPricesHistory = "/prices-history"

	// Order endpoints
	EndpointPostOrder          = "/order"
	EndpointPostOrders         = "/orders"
	EndpointCancel             = "/order"
	EndpointCancelOrders       = "/orders"
	EndpointGetOrder           = "/data/order/"
	EndpointCancelAll          = "/cancel-all"
	EndpointCancelMarketOrders = "/cancel-market-orders"
	EndpointOrders             = "/data/orders"
	EndpointPreMigrationOrders = "/data/pre-migration-orders"
	EndpointTrades             = "/data/trades"
	EndpointIsOrderScoring     = "/order-scoring"
	EndpointAreOrdersScoring   = "/orders-scoring"

	// Notifications
	EndpointGetNotifications  = "/notifications"
	EndpointDropNotifications = "/notifications"

	// Balance & Allowance
	EndpointGetBalanceAllowance    = "/balance-allowance"
	EndpointUpdateBalanceAllowance = "/balance-allowance/update"

	// Rewards
	EndpointGetEarningsForUserForDay       = "/rewards/user"
	EndpointGetTotalEarningsForUserForDay  = "/rewards/user/total"
	EndpointGetLiquidityRewardPercentages  = "/rewards/user/percentages"
	EndpointGetRewardsMarketsCurrent       = "/rewards/markets/current"
	EndpointGetRewardsMarkets              = "/rewards/markets/"
	EndpointGetRewardsEarningsPercentages  = "/rewards/user/markets"

	// Builder endpoints
	EndpointPostHeartbeat    = "/v1/heartbeats"
	EndpointGetBuilderTrades = "/builder/trades"
	EndpointGetBuilderFeeRate = "/fees/builder-fees/"

	// RFQ Endpoints
	EndpointCreateRFQRequest     = "/rfq/request"
	EndpointCancelRFQRequest     = "/rfq/request"
	EndpointGetRFQRequests       = "/rfq/data/requests"
	EndpointCreateRFQQuote       = "/rfq/quote"
	EndpointCancelRFQQuote       = "/rfq/quote"
	EndpointGetRFQRequesterQuotes = "/rfq/data/requester/quotes"
	EndpointGetRFQQuoterQuotes   = "/rfq/data/quoter/quotes"
	EndpointGetRFQBestQuote      = "/rfq/data/best-quote"
	EndpointRFQRequestsAccept    = "/rfq/request/accept"
	EndpointRFQQuoteApprove      = "/rfq/quote/approve"
	EndpointRFQConfig            = "/rfq/config"
)
