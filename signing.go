package clobclient

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
)

const (
	ClobDomainName = "ClobAuthDomain"
	ClobVersion    = "1"
	MsgToSign      = "This message attests that I control the given wallet"
)

// BuildHMACSignature creates an HMAC-SHA256 signature for L2 auth.
func BuildHMACSignature(secret string, timestamp int64, method, requestPath string, body interface{}) (string, error) {
	decodedSecret, err := base64.URLEncoding.DecodeString(secret)
	if err != nil {
		// Try with padding
		padded := secret
		if len(padded)%4 != 0 {
			padded += strings.Repeat("=", 4-len(padded)%4)
		}
		decodedSecret, err = base64.URLEncoding.DecodeString(padded)
		if err != nil {
			return "", fmt.Errorf("failed to decode secret: %w", err)
		}
	}

	message := fmt.Sprintf("%d%s%s", timestamp, method, requestPath)
	if body != nil {
		var bodyStr string
		switch v := body.(type) {
		case string:
			bodyStr = v
		default:
			bodyBytes, err := json.Marshal(v)
			if err != nil {
				return "", fmt.Errorf("failed to marshal body: %w", err)
			}
			bodyStr = string(bodyBytes)
		}
		if bodyStr != "" {
			// Replace single quotes with double quotes to match Python behavior
			bodyStr = strings.ReplaceAll(bodyStr, "'", "\"")
			message += bodyStr
		}
	}

	h := hmac.New(sha256.New, decodedSecret)
	h.Write([]byte(message))
	sig := base64.URLEncoding.EncodeToString(h.Sum(nil))
	return sig, nil
}

// SignClobAuthMessage signs an EIP-712 auth message for L1 auth.
func SignClobAuthMessage(signer *Signer, timestamp int64, nonce int) (string, error) {
	typedData := apitypes.TypedData{
		Types: apitypes.Types{
			"EIP712Domain": {
				{Name: "name", Type: "string"},
				{Name: "version", Type: "string"},
				{Name: "chainId", Type: "uint256"},
			},
			"ClobAuth": {
				{Name: "address", Type: "address"},
				{Name: "timestamp", Type: "string"},
				{Name: "nonce", Type: "uint256"},
				{Name: "message", Type: "string"},
			},
		},
		PrimaryType: "ClobAuth",
		Domain: apitypes.TypedDataDomain{
			Name:    ClobDomainName,
			Version: ClobVersion,
			ChainId: math.NewHexOrDecimal256(int64(signer.GetChainID())),
		},
		Message: apitypes.TypedDataMessage{
			"address":   signer.Address(),
			"timestamp": fmt.Sprintf("%d", timestamp),
			"nonce":     math.NewHexOrDecimal256(int64(nonce)),
			"message":   MsgToSign,
		},
	}

	domainSeparator, err := typedData.HashStruct("EIP712Domain", typedData.Domain.Map())
	if err != nil {
		return "", fmt.Errorf("failed to hash domain: %w", err)
	}

	typedDataHash, err := typedData.HashStruct(typedData.PrimaryType, typedData.Message)
	if err != nil {
		return "", fmt.Errorf("failed to hash message: %w", err)
	}

	rawData := []byte(fmt.Sprintf("\x19\x01%s%s", string(domainSeparator), string(typedDataHash)))
	hashBytes := crypto.Keccak256(rawData)
	hashHex := common.Bytes2Hex(hashBytes)

	sig, err := signer.Sign("0x" + hashHex)
	if err != nil {
		return "", err
	}

	return sig, nil
}

// hashEIP712Order computes the EIP-712 hash for a V1 order.
func hashEIP712OrderV1(chainID int, contractAddress string, order *OrderV1) ([]byte, error) {
	typedData := buildOrderV1TypedData(chainID, contractAddress, order)
	return hashTypedData(typedData)
}

// hashEIP712OrderV2 computes the EIP-712 hash for a V2 order.
func hashEIP712OrderV2(chainID int, contractAddress string, order *OrderV2) ([]byte, error) {
	typedData := buildOrderV2TypedData(chainID, contractAddress, order)
	return hashTypedData(typedData)
}

func hashTypedData(typedData apitypes.TypedData) ([]byte, error) {
	domainSeparator, err := typedData.HashStruct("EIP712Domain", typedData.Domain.Map())
	if err != nil {
		return nil, fmt.Errorf("failed to hash domain: %w", err)
	}

	typedDataHash, err := typedData.HashStruct(typedData.PrimaryType, typedData.Message)
	if err != nil {
		return nil, fmt.Errorf("failed to hash message: %w", err)
	}

	rawData := []byte(fmt.Sprintf("\x19\x01%s%s", string(domainSeparator), string(typedDataHash)))
	return crypto.Keccak256(rawData), nil
}

const (
	CTFExchangeV1DomainName    = "Polymarket CTF Exchange"
	CTFExchangeV1DomainVersion = "1"
	CTFExchangeV2DomainName    = "Polymarket CTF Exchange"
	CTFExchangeV2DomainVersion = "2"
)

func buildOrderV1TypedData(chainID int, contractAddress string, order *OrderV1) apitypes.TypedData {
	salt := new(big.Int)
	salt.SetString(order.Salt, 10)
	tokenId := new(big.Int)
	tokenId.SetString(order.TokenID, 10)
	makerAmount := new(big.Int)
	makerAmount.SetString(order.MakerAmount, 10)
	takerAmount := new(big.Int)
	takerAmount.SetString(order.TakerAmount, 10)
	expiration := new(big.Int)
	expiration.SetString(order.Expiration, 10)
	nonceBig := new(big.Int)
	nonceBig.SetString(order.Nonce, 10)
	feeRateBps := new(big.Int)
	feeRateBps.SetString(order.FeeRateBps, 10)

	return apitypes.TypedData{
		Types: apitypes.Types{
			"EIP712Domain": {
				{Name: "name", Type: "string"},
				{Name: "version", Type: "string"},
				{Name: "chainId", Type: "uint256"},
				{Name: "verifyingContract", Type: "address"},
			},
			"Order": {
				{Name: "salt", Type: "uint256"},
				{Name: "maker", Type: "address"},
				{Name: "signer", Type: "address"},
				{Name: "taker", Type: "address"},
				{Name: "tokenId", Type: "uint256"},
				{Name: "makerAmount", Type: "uint256"},
				{Name: "takerAmount", Type: "uint256"},
				{Name: "expiration", Type: "uint256"},
				{Name: "nonce", Type: "uint256"},
				{Name: "feeRateBps", Type: "uint256"},
				{Name: "side", Type: "uint8"},
				{Name: "signatureType", Type: "uint8"},
			},
		},
		PrimaryType: "Order",
		Domain: apitypes.TypedDataDomain{
			Name:              CTFExchangeV1DomainName,
			Version:           CTFExchangeV1DomainVersion,
			ChainId:           math.NewHexOrDecimal256(int64(chainID)),
			VerifyingContract: contractAddress,
		},
		Message: apitypes.TypedDataMessage{
			"salt":          salt.String(),
			"maker":         order.Maker,
			"signer":        order.SignerAddr,
			"taker":         order.Taker,
			"tokenId":       tokenId.String(),
			"makerAmount":   makerAmount.String(),
			"takerAmount":   takerAmount.String(),
			"expiration":    expiration.String(),
			"nonce":         nonceBig.String(),
			"feeRateBps":    feeRateBps.String(),
			"side":          fmt.Sprintf("%d", int(order.Side)),
			"signatureType": fmt.Sprintf("%d", int(order.SignatureType)),
		},
		}
}

func buildOrderV2TypedData(chainID int, contractAddress string, order *OrderV2) apitypes.TypedData {
	salt := new(big.Int)
	salt.SetString(order.Salt, 10)
	tokenId := new(big.Int)
	tokenId.SetString(order.TokenID, 10)
	makerAmount := new(big.Int)
	makerAmount.SetString(order.MakerAmount, 10)
	takerAmount := new(big.Int)
	takerAmount.SetString(order.TakerAmount, 10)
	timestamp := new(big.Int)
	timestamp.SetString(order.Timestamp, 10)

	return apitypes.TypedData{
		Types: apitypes.Types{
			"EIP712Domain": {
				{Name: "name", Type: "string"},
				{Name: "version", Type: "string"},
				{Name: "chainId", Type: "uint256"},
				{Name: "verifyingContract", Type: "address"},
			},
			"Order": {
				{Name: "salt", Type: "uint256"},
				{Name: "maker", Type: "address"},
				{Name: "signer", Type: "address"},
				{Name: "tokenId", Type: "uint256"},
				{Name: "makerAmount", Type: "uint256"},
				{Name: "takerAmount", Type: "uint256"},
				{Name: "side", Type: "uint8"},
				{Name: "signatureType", Type: "uint8"},
				{Name: "timestamp", Type: "uint256"},
				{Name: "metadata", Type: "bytes32"},
				{Name: "builder", Type: "bytes32"},
			},
		},
		PrimaryType: "Order",
		Domain: apitypes.TypedDataDomain{
			Name:              CTFExchangeV2DomainName,
			Version:           CTFExchangeV2DomainVersion,
			ChainId:           math.NewHexOrDecimal256(int64(chainID)),
			VerifyingContract: contractAddress,
		},
		Message: apitypes.TypedDataMessage{
			"salt":          salt.String(),
			"maker":         order.Maker,
			"signer":        order.SignerAddr,
			"tokenId":       tokenId.String(),
			"makerAmount":   makerAmount.String(),
			"takerAmount":   takerAmount.String(),
			"side":          fmt.Sprintf("%d", int(order.Side)),
			"signatureType": fmt.Sprintf("%d", int(order.Sig)),
			"timestamp":     timestamp.String(),
			"metadata":      hexToBytes32(order.Metadata),
			"builder":       hexToBytes32(order.Builder),
		},
		}
}

func hexToBytes32(hexStr string) [32]byte {
	var result [32]byte
	if len(hexStr) >= 2 && hexStr[:2] == "0x" {
		hexStr = hexStr[2:]
	}
	// Pad to 64 hex chars (32 bytes)
	for len(hexStr) < 64 {
		hexStr = "0" + hexStr
	}
	b := common.Hex2Bytes(hexStr)
	copy(result[:], b)
	return result
}
