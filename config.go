package clobclient

import "fmt"

// ContractConfig holds contract addresses for a specific chain.
type ContractConfig struct {
	Exchange          string
	NegRiskAdapter    string
	NegRiskExchange   string
	Collateral        string
	ConditionalTokens string
	ExchangeV2        string
	NegRiskExchangeV2 string
}

// GetContractConfig returns contract addresses for the given chain ID.
func GetContractConfig(chainID int) (*ContractConfig, error) {
	configs := map[int]*ContractConfig{
		ChainIDPolygon: {
			Exchange:          "0x4bFb41d5B3570DeFd03C39a9A4D8dE6Bd8B8982E",
			NegRiskAdapter:    "0xd91E80cF2E7be2e162c6513ceD06f1dD0dA35296",
			NegRiskExchange:   "0xC5d563A36AE78145C45a50134d48A1215220f80a",
			Collateral:        "0xC011a7E12a19f7B1f670d46F03B03f3342E82DFB",
			ConditionalTokens: "0x4D97DCd97eC945f40cF65F87097ACe5EA0476045",
			ExchangeV2:        "0xE111180000d2663C0091e4f400237545B87B996B",
			NegRiskExchangeV2: "0xe2222d279d744050d28e00520010520000310F59",
		},
		ChainIDAmoy: {
			Exchange:          "0xdFE02Eb6733538f8Ea35D585af8DE5958AD99E40",
			NegRiskAdapter:    "0xd91E80cF2E7be2e162c6513ceD06f1dD0dA35296",
			NegRiskExchange:   "0xC5d563A36AE78145C45a50134d48A1215220f80a",
			Collateral:        "0xC011a7E12a19f7B1f670d46F03B03f3342E82DFB",
			ConditionalTokens: "0x69308FB512518e39F9b16112fA8d994F4e2Bf8bB",
			ExchangeV2:        "0xE111180000d2663C0091e4f400237545B87B996B",
			NegRiskExchangeV2: "0xe2222d279d744050d28e00520010520000310F59",
		},
	}

	config, ok := configs[chainID]
	if !ok {
		return nil, fmt.Errorf("invalid chain_id: %d", chainID)
	}
	return config, nil
}
