package qr

import (
	"crypto/hmac"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

var (
	ErrInvalidToken     = errors.New("invalid QR token")
	ErrTokenExpired     = errors.New("QR token expired")
	ErrInvalidSignature = errors.New("invalid QR token signature")
	ErrTokenRevoked     = errors.New("QR token has been revoked")
	ErrInvalidKeyConfig = errors.New("qr: secret key must be at least 32 bytes")
)

type Verifier struct {
	secretKey []byte
	clock     Clock
}

func NewVerifier(secretKey string) (*Verifier, error) {
	key := []byte(secretKey)
	if len(key) < minKeyLength {
		return nil, fmt.Errorf("qr: secret key must be at least %d bytes, got %d", minKeyLength, len(key))
	}
	return &Verifier{
		secretKey: key,
		clock:     realClock{},
	}, nil
}

func NewVerifierWithClock(secretKey string, clock Clock) (*Verifier, error) {
	key := []byte(secretKey)
	if len(key) < minKeyLength {
		return nil, fmt.Errorf("qr: secret key must be at least %d bytes, got %d", minKeyLength, len(key))
	}
	return &Verifier{
		secretKey: key,
		clock:     clock,
	}, nil
}

func (v *Verifier) Verify(tokenString string) (*QRToken, error) {
	parts := strings.SplitN(tokenString, ".", 2)
	if len(parts) != 2 {
		return nil, ErrInvalidToken
	}

	encodedPayload := parts[0]
	encodedSignature := parts[1]

	payload, err := base64.RawURLEncoding.DecodeString(encodedPayload)
	if err != nil {
		return nil, ErrInvalidToken
	}

	expectedSignature := computeHMAC(payload, v.secretKey)

	signature, err := base64.RawURLEncoding.DecodeString(encodedSignature)
	if err != nil {
		return nil, ErrInvalidToken
	}

	if !hmac.Equal(signature, expectedSignature) {
		return nil, ErrInvalidSignature
	}

	var token QRToken
	if err := json.Unmarshal(payload, &token); err != nil {
		return nil, ErrInvalidToken
	}

	if token.ExpiresAt == 0 || token.IssuedAt == 0 {
		return nil, ErrInvalidToken
	}

	if token.DeliveryID == uuid.Nil || token.UserID == uuid.Nil || token.Nonce == uuid.Nil {
		return nil, ErrInvalidToken
	}

	now := v.clock.Now().Unix()
	if token.ExpiresAt < now {
		return nil, ErrTokenExpired
	}

	if token.IssuedAt > now {
		return nil, ErrInvalidToken
	}

	return &token, nil
}
