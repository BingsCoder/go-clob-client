package clobclient

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"math/rand"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

// GenerateOrderSalt generates a random salt for an order.
func GenerateOrderSalt() string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	ts := time.Now().UnixMilli()
	salt := int64(r.Float64() * float64(ts))
	return fmt.Sprintf("%d", salt)
}

// ExchangeOrderBuilderV1 builds and signs V1 orders.
type ExchangeOrderBuilderV1 struct {
	ContractAddress string
	ChainID         int
	Signer          *Signer
	GenerateSalt    func() string
}

// NewExchangeOrderBuilderV1 creates a new V1 order builder.
func NewExchangeOrderBuilderV1(contractAddress string, chainID int, signer *Signer) *ExchangeOrderBuilderV1 {
	return &ExchangeOrderBuilderV1{
		ContractAddress: contractAddress,
		ChainID:         chainID,
		Signer:          signer,
		GenerateSalt:    GenerateOrderSalt,
	}
}

// BuildSignedOrder creates and signs a V1 order.
func (b *ExchangeOrderBuilderV1) BuildSignedOrder(data *OrderDataV1) (*SignedOrderV1, error) {
	order, err := b.BuildOrder(data)
	if err != nil {
		return nil, err
	}

	sig, err := b.BuildOrderSignature(order)
	if err != nil {
		return nil, err
	}

	return &SignedOrderV1{
		OrderV1:   *order,
		Signature: sig,
	}, nil
}

// BuildOrder creates an unsigned V1 order from order data.
func (b *ExchangeOrderBuilderV1) BuildOrder(data *OrderDataV1) (*OrderV1, error) {
	signerAddr := data.SignerAddr
	if signerAddr == "" {
		signerAddr = data.Maker
	}

	if !strings.EqualFold(signerAddr, b.Signer.Address()) {
		return nil, fmt.Errorf("signer does not match")
	}

	taker := data.Taker
	if taker == "" {
		taker = ZeroAddress
	}

	expiration := data.Expiration
	if expiration == "" {
		expiration = "0"
	}

	nonce := data.Nonce
	if nonce == "" {
		nonce = "0"
	}

	feeRateBps := data.FeeRateBps
	if feeRateBps == "" {
		feeRateBps = "0"
	}

	sigType := data.SignatureType

	return &OrderV1{
		Salt:          b.GenerateSalt(),
		Maker:         data.Maker,
		SignerAddr:    signerAddr,
		Taker:         taker,
		TokenID:       data.TokenID,
		MakerAmount:   data.MakerAmount,
		TakerAmount:   data.TakerAmount,
		Expiration:    expiration,
		Nonce:         nonce,
		FeeRateBps:    feeRateBps,
		Side:          data.Side,
		SignatureType: sigType,
	}, nil
}

// BuildOrderSignature signs a V1 order using EIP-712.
func (b *ExchangeOrderBuilderV1) BuildOrderSignature(order *OrderV1) (string, error) {
	hash, err := hashEIP712OrderV1(b.ChainID, b.ContractAddress, order)
	if err != nil {
		return "", err
	}

	sig, err := b.Signer.SignHash(hash)
	if err != nil {
		return "", err
	}

	return "0x" + hex.EncodeToString(sig), nil
}

// ExchangeOrderBuilderV2 builds and signs V2 orders.
type ExchangeOrderBuilderV2 struct {
	ContractAddress    string
	ChainID            int
	Signer             *Signer
	GenerateSalt       func() string
	appDomainSeparator []byte
}

// NewExchangeOrderBuilderV2 creates a new V2 order builder.
func NewExchangeOrderBuilderV2(contractAddress string, chainID int, signer *Signer) *ExchangeOrderBuilderV2 {
	// Compute domain separator
	domainTypeHash := crypto.Keccak256([]byte("EIP712Domain(string name,string version,uint256 chainId,address verifyingContract)"))
	nameHash := crypto.Keccak256([]byte(CTFExchangeV2DomainName))
	versionHash := crypto.Keccak256([]byte(CTFExchangeV2DomainVersion))

	chainIDBig := big.NewInt(int64(chainID))
	addr := common.HexToAddress(contractAddress)

	// ABI encode: (bytes32, bytes32, bytes32, uint256, address)
	encoded := make([]byte, 0, 5*32)
	encoded = append(encoded, common.LeftPadBytes(domainTypeHash, 32)...)
	encoded = append(encoded, common.LeftPadBytes(nameHash, 32)...)
	encoded = append(encoded, common.LeftPadBytes(versionHash, 32)...)
	encoded = append(encoded, common.LeftPadBytes(chainIDBig.Bytes(), 32)...)
	encoded = append(encoded, common.LeftPadBytes(addr.Bytes(), 32)...)

	separator := crypto.Keccak256(encoded)

	return &ExchangeOrderBuilderV2{
		ContractAddress:    contractAddress,
		ChainID:            chainID,
		Signer:             signer,
		GenerateSalt:       GenerateOrderSalt,
		appDomainSeparator: separator,
	}
}

// BuildSignedOrder creates and signs a V2 order.
func (b *ExchangeOrderBuilderV2) BuildSignedOrder(data *OrderDataV2) (*SignedOrderV2, error) {
	order, err := b.BuildOrder(data)
	if err != nil {
		return nil, err
	}

	sig, err := b.BuildOrderSignature(order)
	if err != nil {
		return nil, err
	}

	return &SignedOrderV2{
		OrderV2:   *order,
		Signature: sig,
	}, nil
}

// BuildOrder creates an unsigned V2 order from order data.
func (b *ExchangeOrderBuilderV2) BuildOrder(data *OrderDataV2) (*OrderV2, error) {
	signerAddr := data.SignerAddr
	if signerAddr == "" {
		signerAddr = data.Maker
	}

	sigType := data.SignatureType

	if sigType != SignatureTypeV2Poly1271 && !strings.EqualFold(signerAddr, b.Signer.Address()) {
		return nil, fmt.Errorf("signer does not match")
	}

	timestamp := data.Timestamp
	if timestamp == "" {
		timestamp = fmt.Sprintf("%d", time.Now().UnixMilli())
	}

	metadata := data.Metadata
	if metadata == "" {
		metadata = Bytes32Zero
	}

	builder := data.Builder
	if builder == "" {
		builder = Bytes32Zero
	}

	expiration := data.Expiration
	if expiration == "" {
		expiration = "0"
	}

	return &OrderV2{
		Salt:        b.GenerateSalt(),
		Maker:       data.Maker,
		SignerAddr:  signerAddr,
		TokenID:     data.TokenID,
		MakerAmount: data.MakerAmount,
		TakerAmount: data.TakerAmount,
		Side:        data.Side,
		Sig:         sigType,
		Timestamp:   timestamp,
		Metadata:    metadata,
		Builder:     builder,
		Expiration:  expiration,
	}, nil
}

// BuildOrderSignature signs a V2 order using EIP-712.
func (b *ExchangeOrderBuilderV2) BuildOrderSignature(order *OrderV2) (string, error) {
	if order.Sig == SignatureTypeV2Poly1271 {
		return b.buildPoly1271OrderSignature(order)
	}

	hash, err := hashEIP712OrderV2(b.ChainID, b.ContractAddress, order)
	if err != nil {
		return "", err
	}

	sig, err := b.Signer.SignHash(hash)
	if err != nil {
		return "", err
	}

	return "0x" + hex.EncodeToString(sig), nil
}

func (b *ExchangeOrderBuilderV2) buildPoly1271OrderSignature(order *OrderV2) (string, error) {
	orderTypeString := "Order(uint256 salt,address maker,address signer,uint256 tokenId,uint256 makerAmount,uint256 takerAmount,uint8 side,uint8 signatureType,uint256 timestamp,bytes32 metadata,bytes32 builder)"
	soladyTypeString := "TypedDataSign(Order contents,string name,string version,uint256 chainId,address verifyingContract,bytes32 salt)" + orderTypeString

	orderTypeHash := crypto.Keccak256([]byte(orderTypeString))
	soladyTypeHash := crypto.Keccak256([]byte(soladyTypeString))
	depositWalletNameHash := crypto.Keccak256([]byte("DepositWallet"))
	depositWalletVersionHash := crypto.Keccak256([]byte("1"))
	depositWalletDomainSalt := make([]byte, 32)

	// Build contents hash
	salt := new(big.Int)
	salt.SetString(order.Salt, 10)
	maker := common.HexToAddress(order.Maker)
	signerAddr := common.HexToAddress(order.SignerAddr)
	tokenID := new(big.Int)
	tokenID.SetString(order.TokenID, 10)
	makerAmount := new(big.Int)
	makerAmount.SetString(order.MakerAmount, 10)
	takerAmount := new(big.Int)
	takerAmount.SetString(order.TakerAmount, 10)
	timestamp := new(big.Int)
	timestamp.SetString(order.Timestamp, 10)
	metadataBytes := hexToBytes32Slice(order.Metadata)
	builderBytes := hexToBytes32Slice(order.Builder)

	// ABI encode for contents hash
	contentsEncoded := make([]byte, 0, 12*32)
	contentsEncoded = append(contentsEncoded, common.LeftPadBytes(orderTypeHash, 32)...)
	contentsEncoded = append(contentsEncoded, common.LeftPadBytes(salt.Bytes(), 32)...)
	contentsEncoded = append(contentsEncoded, common.LeftPadBytes(maker.Bytes(), 32)...)
	contentsEncoded = append(contentsEncoded, common.LeftPadBytes(signerAddr.Bytes(), 32)...)
	contentsEncoded = append(contentsEncoded, common.LeftPadBytes(tokenID.Bytes(), 32)...)
	contentsEncoded = append(contentsEncoded, common.LeftPadBytes(makerAmount.Bytes(), 32)...)
	contentsEncoded = append(contentsEncoded, common.LeftPadBytes(takerAmount.Bytes(), 32)...)
	contentsEncoded = append(contentsEncoded, common.LeftPadBytes([]byte{byte(order.Side)}, 32)...)
	contentsEncoded = append(contentsEncoded, common.LeftPadBytes([]byte{byte(order.Sig)}, 32)...)
	contentsEncoded = append(contentsEncoded, common.LeftPadBytes(timestamp.Bytes(), 32)...)
	contentsEncoded = append(contentsEncoded, metadataBytes...)
	contentsEncoded = append(contentsEncoded, builderBytes...)

	contentsHash := crypto.Keccak256(contentsEncoded)

	// Build typed data sign struct hash
	chainIDBig := big.NewInt(int64(b.ChainID))
	tdssEncoded := make([]byte, 0, 7*32)
	tdssEncoded = append(tdssEncoded, common.LeftPadBytes(soladyTypeHash, 32)...)
	tdssEncoded = append(tdssEncoded, contentsHash...)
	tdssEncoded = append(tdssEncoded, depositWalletNameHash...)
	tdssEncoded = append(tdssEncoded, depositWalletVersionHash...)
	tdssEncoded = append(tdssEncoded, common.LeftPadBytes(chainIDBig.Bytes(), 32)...)
	tdssEncoded = append(tdssEncoded, common.LeftPadBytes(signerAddr.Bytes(), 32)...)
	tdssEncoded = append(tdssEncoded, depositWalletDomainSalt...)

	typedDataSignStructHash := crypto.Keccak256(tdssEncoded)

	// Final digest
	digestInput := []byte{0x19, 0x01}
	digestInput = append(digestInput, b.appDomainSeparator...)
	digestInput = append(digestInput, typedDataSignStructHash...)
	digest := crypto.Keccak256(digestInput)

	sigBytes, err := crypto.Sign(digest, b.Signer.privateKey)
	if err != nil {
		return "", err
	}
	if sigBytes[64] < 27 {
		sigBytes[64] += 27
	}
	innerSig := hex.EncodeToString(sigBytes)

	contentsType := hex.EncodeToString([]byte(orderTypeString))
	contentsTypeLen := fmt.Sprintf("%04x", len(orderTypeString))

	return "0x" + innerSig + hex.EncodeToString(b.appDomainSeparator) + hex.EncodeToString(contentsHash) + contentsType + contentsTypeLen, nil
}

func hexToBytes32Slice(hexStr string) []byte {
	if len(hexStr) >= 2 && hexStr[:2] == "0x" {
		hexStr = hexStr[2:]
	}
	for len(hexStr) < 64 {
		hexStr = "0" + hexStr
	}
	b, _ := hex.DecodeString(hexStr)
	result := make([]byte, 32)
	copy(result, b)
	return result
}
