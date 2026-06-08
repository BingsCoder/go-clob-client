package clobclient

import (
	"crypto/ecdsa"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

// Signer wraps an Ethereum private key for signing operations.
type Signer struct {
	privateKey *ecdsa.PrivateKey
	chainID    int
	address    common.Address
}

// NewSigner creates a new Signer from a hex-encoded private key.
func NewSigner(privateKeyHex string, chainID int) (*Signer, error) {
	if len(privateKeyHex) >= 2 && privateKeyHex[:2] == "0x" {
		privateKeyHex = privateKeyHex[2:]
	}

	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		return nil, fmt.Errorf("invalid private key: %w", err)
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("error casting public key to ECDSA")
	}

	address := crypto.PubkeyToAddress(*publicKeyECDSA)

	return &Signer{
		privateKey: privateKey,
		chainID:    chainID,
		address:    address,
	}, nil
}

// Address returns the Ethereum address of the signer.
func (s *Signer) Address() string {
	return s.address.Hex()
}

// GetChainID returns the chain ID.
func (s *Signer) GetChainID() int {
	return s.chainID
}

// Sign signs a message hash (hex-encoded with 0x prefix) and returns the signature hex.
func (s *Signer) Sign(messageHashHex string) (string, error) {
	if len(messageHashHex) >= 2 && messageHashHex[:2] == "0x" {
		messageHashHex = messageHashHex[2:]
	}
	hash := common.HexToHash(messageHashHex)
	sig, err := crypto.Sign(hash.Bytes(), s.privateKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign: %w", err)
	}
	// Adjust V value from 0/1 to 27/28 for Ethereum compatibility
	if sig[64] < 27 {
		sig[64] += 27
	}
	return fmt.Sprintf("0x%x", sig), nil
}

// SignHash signs a raw hash bytes and returns the signature bytes.
func (s *Signer) SignHash(hash []byte) ([]byte, error) {
	sig, err := crypto.Sign(hash, s.privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign: %w", err)
	}
	// Adjust V value
	if sig[64] < 27 {
		sig[64] += 27
	}
	return sig, nil
}
