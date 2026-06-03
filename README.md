# go-clob-client

Go client library for the Polymarket CLOB (Central Limit Order Book) API. This is a Go port of [py-clob-client-v2](https://github.com/Polymarket/py-clob-client-v2).

## Installation

```bash
go get github.com/BingsCoder/go-clob-client
```

## Usage

```go
package main

import (
	"fmt"
	"log"

	clobclient "github.com/BingsCoder/go-clob-client"
)

func main() {
	// Initialize client (no auth)
	client, err := clobclient.NewClobClient(
		"https://clob.polymarket.com",
		nil,    // signer
		nil,    // API credentials
		nil,    // chain ID (defaults to Polygon)
		nil,    // signature type
		nil,    // funder address
	)
	if err != nil {
		log.Fatal(err)
	}

	// Public endpoints
	ok, err := client.GetOk()
	fmt.Println("Server OK:", ok)

	serverTime, _ := client.GetServerTime()
	fmt.Println("Server Time:", serverTime)
}
```

### With Authentication (L1 — Wallet Signing)

```go
// Create a signer from private key
signer, _ := clobclient.NewSigner("0xYOUR_PRIVATE_KEY_HEX")

client, _ := clobclient.NewClobClient(
	"https://clob.polymarket.com",
	signer,
	nil,     // no API creds yet
	nil,
	nil,
	nil,
)

// Derive API key from wallet signature
apiCreds, _ := client.DeriveAPIKey(0)
fmt.Println("API Key:", apiCreds)
```

### With Authentication (L2 — API Key + HMAC)

```go
signer, _ := clobclient.NewSigner("0xYOUR_PRIVATE_KEY_HEX")
creds := &clobclient.ApiCreds{
	APIKey:       "your-api-key",
	APISecret:    "your-api-secret",
	APIPassphrase: "your-api-passphrase",
}

client, _ := clobclient.NewClobClient(
	"https://clob.polymarket.com",
	signer,
	creds,
	nil,
	nil,
	nil,
)

// Get open orders
orders, _ := client.GetOrders(nil)
fmt.Println("Orders:", orders)

// Create a limit order
order, _ := client.CreateOrder(&clobclient.OrderArgs{
	TokenID: "YOUR_TOKEN_ID",
	Price:   0.50,
	Size:    100.0,
	Side:    clobclient.SideBuyStr,
}, nil)
fmt.Println("Order:", order)

// Post the order
resp, _ := client.PostOrder(order, "GTC")
fmt.Println("Posted:", resp)
```

### RFQ (Request for Quote)

```go
// Create RFQ request
rfqResp, _ := client.RFQ.CreateRFQRequest(&clobclient.RFQUserRequest{
	TokenID: "YOUR_TOKEN_ID",
	Price:   0.50,
	Side:    clobclient.SideBuyStr,
	Size:    1000.0,
}, nil)
fmt.Println("RFQ:", rfqResp)

// Get RFQ quotes
quotes, _ := client.RFQ.GetRFQRequesterQuotes(nil)
fmt.Println("Quotes:", quotes)
```

## Features

- **Full API coverage**: All Polymarket CLOB API endpoints
- **EIP-712 signing**: Ethereum wallet-based authentication (L1)
- **HMAC signing**: API key-based authentication (L2)
- **Order management**: Create, cancel, and manage limit/market orders
- **Order book**: Fetch order books, prices, spreads, and midpoints
- **Trade history**: Query trades with flexible filtering
- **Market data**: Markets, events, sampling, rewards
- **RFQ support**: Request for Quote operations
- **Neg-risk support**: Negative risk market handling
- **POLY_1271**: Smart contract wallet signature support (V2 orders)

## API Methods

### Public (No Auth)
- `GetOk()` — Health check
- `GetServerTime()` — Server timestamp
- `GetMarkets(nextCursor)` — List markets
- `GetMarket(conditionID)` — Get specific market
- `GetOrderBook(tokenID)` — Order book
- `GetMidpoint(tokenID)` — Midpoint price
- `GetPrice(tokenID, side)` — Best price
- `GetSpread(tokenID)` — Spread
- `GetTickSize(tokenID)` — Tick size
- `GetPricesHistory(params)` — Price history

### L1 Auth (Wallet Signature)
- `DeriveAPIKey(nonce)` — Derive API key from wallet
- `CreateAPIKey(nonce)` — Create new API key
- `GetAPIKeys()` — List API keys
- `DeleteAPIKey()` — Delete API key

### L2 Auth (API Key)
- `CreateOrder(args, options)` — Create signed order
- `PostOrder(order, orderType)` — Submit order
- `CancelOrder(orderID)` — Cancel order
- `CancelOrders(orderIDs)` — Cancel multiple orders
- `CancelAll()` — Cancel all orders
- `GetOrders(params)` — Get open orders
- `GetTrades(params)` — Get trade history
- `GetBalanceAllowance(params)` — Get balance/allowance
- `UpdateBalanceAllowance(params)` — Update allowance
- `GetNotifications()` — Get notifications
- `DropNotifications()` — Clear notifications

### RFQ
- `RFQ.CreateRFQRequest(request, options)` — Create RFQ request
- `RFQ.CreateRFQQuote(quote, options)` — Create RFQ quote
- `RFQ.AcceptRFQQuote(params)` — Accept quote
- `RFQ.ApproveRFQOrder(params)` — Approve order
- `RFQ.CancelRFQRequest(params)` — Cancel request
- `RFQ.CancelRFQQuote(params)` — Cancel quote
- `RFQ.GetRFQRequests(params)` — List requests
- `RFQ.RFQConfig()` — RFQ configuration

## License

MIT
