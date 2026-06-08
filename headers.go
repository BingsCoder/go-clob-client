package clobclient

import (
	"fmt"
	"time"
)

const (
	HeaderPolyAddress    = "POLY_ADDRESS"
	HeaderPolySignature  = "POLY_SIGNATURE"
	HeaderPolyTimestamp   = "POLY_TIMESTAMP"
	HeaderPolyNonce      = "POLY_NONCE"
	HeaderPolyAPIKey     = "POLY_API_KEY"
	HeaderPolyPassphrase = "POLY_PASSPHRASE"
)

// CreateLevel1Headers creates L1 (wallet signature) auth headers.
func CreateLevel1Headers(signer *Signer, nonce int, timestamp *int64) (map[string]string, error) {
	var ts int64
	if timestamp != nil {
		ts = *timestamp
	} else {
		ts = time.Now().Unix()
	}

	signature, err := SignClobAuthMessage(signer, ts, nonce)
	if err != nil {
		return nil, fmt.Errorf("failed to sign auth message: %w", err)
	}

	return map[string]string{
		HeaderPolyAddress:   signer.Address(),
		HeaderPolySignature: signature,
		HeaderPolyTimestamp: fmt.Sprintf("%d", ts),
		HeaderPolyNonce:     fmt.Sprintf("%d", nonce),
	}, nil
}

// CreateLevel2Headers creates L2 (HMAC) auth headers.
func CreateLevel2Headers(signer *Signer, creds *ApiCreds, requestArgs *RequestArgs, timestamp *int64) (map[string]string, error) {
	var ts int64
	if timestamp != nil {
		ts = *timestamp
	} else {
		ts = time.Now().Unix()
	}

	var bodyForSig interface{}
	if requestArgs.SerializedBody != "" {
		bodyForSig = requestArgs.SerializedBody
	} else {
		bodyForSig = requestArgs.Body
	}

	hmacSig, err := BuildHMACSignature(creds.APISecret, ts, requestArgs.Method, requestArgs.RequestPath, bodyForSig)
	if err != nil {
		return nil, fmt.Errorf("failed to build HMAC signature: %w", err)
	}

	return map[string]string{
		HeaderPolyAddress:    signer.Address(),
		HeaderPolySignature:  hmacSig,
		HeaderPolyTimestamp:  fmt.Sprintf("%d", ts),
		HeaderPolyAPIKey:     creds.APIKey,
		HeaderPolyPassphrase: creds.APIPassphrase,
	}, nil
}
