package qr

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

const minKeyLength = 32

type Generator struct {
	secretKey []byte
	tokenTTL  time.Duration
	clock     Clock
}

func NewGenerator(secretKey string, tokenTTL time.Duration) (*Generator, error) {
	key := []byte(secretKey)
	if len(key) < minKeyLength {
		return nil, fmt.Errorf("qr: secret key must be at least %d bytes, got %d", minKeyLength, len(key))
	}
	return &Generator{
		secretKey: key,
		tokenTTL:  tokenTTL,
		clock:     realClock{},
	}, nil
}

func NewGeneratorWithClock(secretKey string, tokenTTL time.Duration, clock Clock) (*Generator, error) {
	key := []byte(secretKey)
	if len(key) < minKeyLength {
		return nil, fmt.Errorf("qr: secret key must be at least %d bytes, got %d", minKeyLength, len(key))
	}
	return &Generator{
		secretKey: key,
		tokenTTL:  tokenTTL,
		clock:     clock,
	}, nil
}

func (g *Generator) Generate(deliveryID, orderID, userID, nonce uuid.UUID) (*QRTokenResponse, error) {
	now := g.clock.Now()
	token := &QRToken{
		DeliveryID: deliveryID,
		OrderID:    orderID,
		UserID:     userID,
		Nonce:      nonce,
		IssuedAt:   now.Unix(),
		ExpiresAt:  now.Add(g.tokenTTL).Unix(),
	}

	payload, err := json.Marshal(token)
	if err != nil {
		return nil, fmt.Errorf("marshal token: %w", err)
	}

	signature := computeHMAC(payload, g.secretKey)

	encodedPayload := base64.RawURLEncoding.EncodeToString(payload)
	encodedSignature := base64.RawURLEncoding.EncodeToString(signature)

	result := encodedPayload + "." + encodedSignature

	return &QRTokenResponse{
		Token:     result,
		ExpiresAt: now.Add(g.tokenTTL),
	}, nil
}

func computeHMAC(payload, key []byte) []byte {
	mac := hmac.New(sha256.New, key)
	mac.Write(payload)
	return mac.Sum(nil)
}
