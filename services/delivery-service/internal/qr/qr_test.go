package qr

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

const testSecretKey = "test-secret-key-that-is-32-bytes-lon"

func TestGenerator_KeyValidation(t *testing.T) {
	t.Parallel()

	t.Run("rejects key shorter than 32 bytes", func(t *testing.T) {
		t.Parallel()
		_, err := NewGenerator("short-key", 24*time.Hour)
		require.Error(t, err)
		require.Contains(t, err.Error(), "at least 32 bytes")
	})

	t.Run("accepts key of exactly 32 bytes", func(t *testing.T) {
		t.Parallel()
		gen, err := NewGenerator(strings.Repeat("a", 32), 24*time.Hour)
		require.NoError(t, err)
		require.NotNil(t, gen)
	})

	t.Run("accepts key longer than 32 bytes", func(t *testing.T) {
		t.Parallel()
		gen, err := NewGenerator(strings.Repeat("a", 64), 24*time.Hour)
		require.NoError(t, err)
		require.NotNil(t, gen)
	})
}

func TestVerifier_KeyValidation(t *testing.T) {
	t.Parallel()

	t.Run("rejects key shorter than 32 bytes", func(t *testing.T) {
		t.Parallel()
		_, err := NewVerifier("short-key")
		require.Error(t, err)
		require.Contains(t, err.Error(), "at least 32 bytes")
	})

	t.Run("accepts key of exactly 32 bytes", func(t *testing.T) {
		t.Parallel()
		v, err := NewVerifier(strings.Repeat("a", 32))
		require.NoError(t, err)
		require.NotNil(t, v)
	})
}

func TestQR_GenerateAndVerify(t *testing.T) {
	t.Parallel()

	gen, err := NewGenerator(testSecretKey, 24*time.Hour)
	require.NoError(t, err)

	ver, err := NewVerifier(testSecretKey)
	require.NoError(t, err)

	deliveryID := uuid.New()
	orderID := uuid.New()
	userID := uuid.New()
	nonce := uuid.New()

	resp, err := gen.Generate(deliveryID, orderID, userID, nonce)
	require.NoError(t, err)
	require.NotEmpty(t, resp.Token)
	require.False(t, resp.ExpiresAt.IsZero())

	token, err := ver.Verify(resp.Token)
	require.NoError(t, err)
	require.Equal(t, deliveryID, token.DeliveryID)
	require.Equal(t, orderID, token.OrderID)
	require.Equal(t, userID, token.UserID)
	require.Equal(t, nonce, token.Nonce)
	require.True(t, token.ExpiresAt > token.IssuedAt)
}

func TestQR_InvalidSignature_ModifiedPayload(t *testing.T) {
	t.Parallel()

	gen, err := NewGenerator(testSecretKey, 24*time.Hour)
	require.NoError(t, err)

	ver, err := NewVerifier(testSecretKey)
	require.NoError(t, err)

	deliveryID := uuid.New()
	orderID := uuid.New()
	userID := uuid.New()
	nonce := uuid.New()

	resp, err := gen.Generate(deliveryID, orderID, userID, nonce)
	require.NoError(t, err)

	parts := strings.SplitN(resp.Token, ".", 2)
	require.Len(t, parts, 2)

	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	require.NoError(t, err)

	var token map[string]interface{}
	require.NoError(t, json.Unmarshal(payloadBytes, &token))
	token["delivery_id"] = uuid.New().String()
	modifiedPayload, err := json.Marshal(token)
	require.NoError(t, err)

	modifiedToken := base64.RawURLEncoding.EncodeToString(modifiedPayload) + "." + parts[1]

	_, err = ver.Verify(modifiedToken)
	require.ErrorIs(t, err, ErrInvalidSignature)
}

func TestQR_InvalidSignature_ModifiedSignature(t *testing.T) {
	t.Parallel()

	gen, err := NewGenerator(testSecretKey, 24*time.Hour)
	require.NoError(t, err)

	ver, err := NewVerifier(testSecretKey)
	require.NoError(t, err)

	deliveryID := uuid.New()
	orderID := uuid.New()
	userID := uuid.New()
	nonce := uuid.New()

	resp, err := gen.Generate(deliveryID, orderID, userID, nonce)
	require.NoError(t, err)

	parts := strings.SplitN(resp.Token, ".", 2)
	require.Len(t, parts, 2)

	fakeSignature := base64.RawURLEncoding.EncodeToString([]byte("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"))
	modifiedToken := parts[0] + "." + fakeSignature

	_, err = ver.Verify(modifiedToken)
	require.ErrorIs(t, err, ErrInvalidSignature)
}

func TestQR_InvalidSignature_WrongKey(t *testing.T) {
	t.Parallel()

	gen, err := NewGenerator(testSecretKey, 24*time.Hour)
	require.NoError(t, err)

	wrongKeyVerifier, err := NewVerifier(strings.Repeat("b", 32))
	require.NoError(t, err)

	deliveryID := uuid.New()
	orderID := uuid.New()
	userID := uuid.New()
	nonce := uuid.New()

	resp, err := gen.Generate(deliveryID, orderID, userID, nonce)
	require.NoError(t, err)

	_, err = wrongKeyVerifier.Verify(resp.Token)
	require.ErrorIs(t, err, ErrInvalidSignature)
}

func TestQR_ExpiredToken(t *testing.T) {
	t.Parallel()

	pastTime := time.Now().Add(-2 * time.Hour)
	mockClock := &mockClock{now: pastTime}
	gen, err := NewGeneratorWithClock(testSecretKey, 1*time.Second, mockClock)
	require.NoError(t, err)

	ver, err := NewVerifier(testSecretKey)
	require.NoError(t, err)

	deliveryID := uuid.New()
	orderID := uuid.New()
	userID := uuid.New()
	nonce := uuid.New()

	resp, err := gen.Generate(deliveryID, orderID, userID, nonce)
	require.NoError(t, err)

	_, err = ver.Verify(resp.Token)
	require.ErrorIs(t, err, ErrTokenExpired)
}

func TestQR_RevokedNonce_DifferentNonce(t *testing.T) {
	t.Parallel()

	gen, err := NewGenerator(testSecretKey, 24*time.Hour)
	require.NoError(t, err)

	ver, err := NewVerifier(testSecretKey)
	require.NoError(t, err)

	deliveryID := uuid.New()
	orderID := uuid.New()
	userID := uuid.New()
	nonce1 := uuid.New()

	resp, err := gen.Generate(deliveryID, orderID, userID, nonce1)
	require.NoError(t, err)

	token, err := ver.Verify(resp.Token)
	require.NoError(t, err)
	require.Equal(t, nonce1, token.Nonce)

	nonce2 := uuid.New()
	resp2, err := gen.Generate(deliveryID, orderID, userID, nonce2)
	require.NoError(t, err)

	token2, err := ver.Verify(resp2.Token)
	require.NoError(t, err)
	require.Equal(t, nonce2, token2.Nonce)
}

func TestQR_QRForAnotherDelivery(t *testing.T) {
	t.Parallel()

	gen, err := NewGenerator(testSecretKey, 24*time.Hour)
	require.NoError(t, err)

	ver, err := NewVerifier(testSecretKey)
	require.NoError(t, err)

	deliveryID1 := uuid.New()
	deliveryID2 := uuid.New()
	orderID := uuid.New()
	userID := uuid.New()
	nonce := uuid.New()

	resp, err := gen.Generate(deliveryID1, orderID, userID, nonce)
	require.NoError(t, err)

	token, err := ver.Verify(resp.Token)
	require.NoError(t, err)
	require.Equal(t, deliveryID1, token.DeliveryID)
	require.NotEqual(t, deliveryID2, token.DeliveryID)
}

func TestQR_MalformedToken(t *testing.T) {
	t.Parallel()

	ver, err := NewVerifier(testSecretKey)
	require.NoError(t, err)

	tests := []struct {
		name        string
		token       string
		expectedErr error
	}{
		{"empty token", "", ErrInvalidToken},
		{"no dot separator", "justastring", ErrInvalidToken},
		{"invalid base64 payload", "!!!invalid!!!." + base64.RawURLEncoding.EncodeToString([]byte("sig")), ErrInvalidToken},
		{"invalid base64 signature", base64.RawURLEncoding.EncodeToString([]byte("payload")) + ".!!!invalid!!!", ErrInvalidToken},
		{"empty payload", "." + base64.RawURLEncoding.EncodeToString([]byte("sig")), ErrInvalidSignature},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := ver.Verify(tt.token)
			require.ErrorIs(t, err, tt.expectedErr)
		})
	}
}

func TestQR_InvalidJSON(t *testing.T) {
	t.Parallel()

	ver, err := NewVerifier(testSecretKey)
	require.NoError(t, err)

	key := []byte(testSecretKey)

	invalidJSON := []byte(`{"invalid": json`)
	signature := computeHMAC(invalidJSON, key)
	encodedPayload := base64.RawURLEncoding.EncodeToString(invalidJSON)
	encodedSignature := base64.RawURLEncoding.EncodeToString(signature)
	invalidToken := encodedPayload + "." + encodedSignature

	_, err = ver.Verify(invalidToken)
	require.ErrorIs(t, err, ErrInvalidToken)
}

func TestQR_MissingRequiredFields(t *testing.T) {
	t.Parallel()

	ver, err := NewVerifier(testSecretKey)
	require.NoError(t, err)

	key := []byte(testSecretKey)

	tests := []struct {
		name    string
		payload QRToken
	}{
		{
			"missing delivery_id",
			QRToken{OrderID: uuid.New(), UserID: uuid.New(), Nonce: uuid.New(), IssuedAt: time.Now().Unix(), ExpiresAt: time.Now().Add(24 * time.Hour).Unix()},
		},
		{
			"missing user_id",
			QRToken{DeliveryID: uuid.New(), OrderID: uuid.New(), Nonce: uuid.New(), IssuedAt: time.Now().Unix(), ExpiresAt: time.Now().Add(24 * time.Hour).Unix()},
		},
		{
			"missing nonce",
			QRToken{DeliveryID: uuid.New(), OrderID: uuid.New(), UserID: uuid.New(), IssuedAt: time.Now().Unix(), ExpiresAt: time.Now().Add(24 * time.Hour).Unix()},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			payload, err := json.Marshal(tt.payload)
			require.NoError(t, err)

			signature := computeHMAC(payload, key)
			encodedPayload := base64.RawURLEncoding.EncodeToString(payload)
			encodedSignature := base64.RawURLEncoding.EncodeToString(signature)
			token := encodedPayload + "." + encodedSignature

			_, err = ver.Verify(token)
			require.ErrorIs(t, err, ErrInvalidToken)
		})
	}
}

func TestQR_ZeroExpiration(t *testing.T) {
	t.Parallel()

	ver, err := NewVerifier(testSecretKey)
	require.NoError(t, err)

	key := []byte(testSecretKey)

	token := QRToken{
		DeliveryID: uuid.New(),
		OrderID:    uuid.New(),
		UserID:     uuid.New(),
		Nonce:      uuid.New(),
		IssuedAt:   0,
		ExpiresAt:  0,
	}

	payload, err := json.Marshal(token)
	require.NoError(t, err)

	signature := computeHMAC(payload, key)
	encodedPayload := base64.RawURLEncoding.EncodeToString(payload)
	encodedSignature := base64.RawURLEncoding.EncodeToString(signature)
	tokenStr := encodedPayload + "." + encodedSignature

	_, err = ver.Verify(tokenStr)
	require.ErrorIs(t, err, ErrInvalidToken)
}

func TestQR_FutureIssuedAt(t *testing.T) {
	t.Parallel()

	ver, err := NewVerifier(testSecretKey)
	require.NoError(t, err)

	mockClock := &mockClock{now: time.Now()}
	genWithClock, err := NewGeneratorWithClock(testSecretKey, 24*time.Hour, mockClock)
	require.NoError(t, err)

	futureTime := time.Now().Add(2 * time.Hour)
	mockClock.now = futureTime

	deliveryID := uuid.New()
	orderID := uuid.New()
	userID := uuid.New()
	nonce := uuid.New()

	resp, err := genWithClock.Generate(deliveryID, orderID, userID, nonce)
	require.NoError(t, err)

	_, err = ver.Verify(resp.Token)
	require.ErrorIs(t, err, ErrInvalidToken)
}

func TestQR_ReplayProtection_NonceBased(t *testing.T) {
	t.Parallel()

	gen, err := NewGenerator(testSecretKey, 24*time.Hour)
	require.NoError(t, err)

	ver, err := NewVerifier(testSecretKey)
	require.NoError(t, err)

	deliveryID := uuid.New()
	orderID := uuid.New()
	userID := uuid.New()
	nonce := uuid.New()

	resp, err := gen.Generate(deliveryID, orderID, userID, nonce)
	require.NoError(t, err)

	token, err := ver.Verify(resp.Token)
	require.NoError(t, err)
	require.Equal(t, nonce, token.Nonce)

	token2, err := ver.Verify(resp.Token)
	require.NoError(t, err)
	require.Equal(t, token.DeliveryID, token2.DeliveryID)
}

func TestQR_NonceMismatchBetweenTokenAndDB(t *testing.T) {
	t.Parallel()

	gen, err := NewGenerator(testSecretKey, 24*time.Hour)
	require.NoError(t, err)

	deliveryID := uuid.New()
	orderID := uuid.New()
	userID := uuid.New()
	oldNonce := uuid.New()

	resp, err := gen.Generate(deliveryID, orderID, userID, oldNonce)
	require.NoError(t, err)

	newNonce := uuid.New()
	resp2, err := gen.Generate(deliveryID, orderID, userID, newNonce)
	require.NoError(t, err)

	require.NotEqual(t, resp.Token, resp2.Token)
}

func TestQR_DifferentUsersSameDelivery(t *testing.T) {
	t.Parallel()

	gen, err := NewGenerator(testSecretKey, 24*time.Hour)
	require.NoError(t, err)

	ver, err := NewVerifier(testSecretKey)
	require.NoError(t, err)

	deliveryID := uuid.New()
	orderID := uuid.New()
	user1 := uuid.New()
	user2 := uuid.New()
	nonce := uuid.New()

	resp1, err := gen.Generate(deliveryID, orderID, user1, nonce)
	require.NoError(t, err)

	token1, err := ver.Verify(resp1.Token)
	require.NoError(t, err)
	require.Equal(t, user1, token1.UserID)

	nonce2 := uuid.New()
	resp2, err := gen.Generate(deliveryID, orderID, user2, nonce2)
	require.NoError(t, err)

	token2, err := ver.Verify(resp2.Token)
	require.NoError(t, err)
	require.Equal(t, user2, token2.UserID)
	require.NotEqual(t, token1.UserID, token2.UserID)
}

func TestQR_TokenFormat(t *testing.T) {
	t.Parallel()

	gen, err := NewGenerator(testSecretKey, 24*time.Hour)
	require.NoError(t, err)

	deliveryID := uuid.New()
	orderID := uuid.New()
	userID := uuid.New()
	nonce := uuid.New()

	resp, err := gen.Generate(deliveryID, orderID, userID, nonce)
	require.NoError(t, err)

	parts := strings.SplitN(resp.Token, ".", 2)
	require.Len(t, parts, 2)

	_, err = base64.RawURLEncoding.DecodeString(parts[0])
	require.NoError(t, err, "payload should be valid base64url")

	_, err = base64.RawURLEncoding.DecodeString(parts[1])
	require.NoError(t, err, "signature should be valid base64url")
}

func TestQR_SecretKeyNotInToken(t *testing.T) {
	t.Parallel()

	gen, err := NewGenerator(testSecretKey, 24*time.Hour)
	require.NoError(t, err)

	deliveryID := uuid.New()
	orderID := uuid.New()
	userID := uuid.New()
	nonce := uuid.New()

	resp, err := gen.Generate(deliveryID, orderID, userID, nonce)
	require.NoError(t, err)

	require.NotContains(t, resp.Token, testSecretKey, "token must not contain the secret key")
}

func TestQR_GeneratorWithClock(t *testing.T) {
	t.Parallel()

	fixedTime := time.Date(2025, 1, 15, 12, 0, 0, 0, time.UTC)
	mockClock := &mockClock{now: fixedTime}

	gen, err := NewGeneratorWithClock(testSecretKey, 1*time.Hour, mockClock)
	require.NoError(t, err)

	deliveryID := uuid.New()
	orderID := uuid.New()
	userID := uuid.New()
	nonce := uuid.New()

	resp, err := gen.Generate(deliveryID, orderID, userID, nonce)
	require.NoError(t, err)

	expectedExpiry := fixedTime.Add(1 * time.Hour)
	require.WithinDuration(t, expectedExpiry, resp.ExpiresAt, time.Second)
}

type mockClock struct {
	now time.Time
}

func (m *mockClock) Now() time.Time {
	return m.now
}

func TestQR_DoesNotLeakSecretInErrors(t *testing.T) {
	t.Parallel()

	_, err := NewGenerator("short", 24*time.Hour)
	require.Error(t, err)
	require.NotContains(t, err.Error(), "short", "error must not contain the actual secret key value")

	_, err = NewVerifier("short")
	require.Error(t, err)
	require.NotContains(t, err.Error(), "short", "error must not contain the actual secret key value")
}

func TestQR_Verify_ModifiedDeliveryIDInPayload(t *testing.T) {
	t.Parallel()

	gen, err := NewGenerator(testSecretKey, 24*time.Hour)
	require.NoError(t, err)

	ver, err := NewVerifier(testSecretKey)
	require.NoError(t, err)

	deliveryID := uuid.New()
	orderID := uuid.New()
	userID := uuid.New()
	nonce := uuid.New()

	resp, err := gen.Generate(deliveryID, orderID, userID, nonce)
	require.NoError(t, err)

	parts := strings.SplitN(resp.Token, ".", 2)
	require.Len(t, parts, 2)

	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	require.NoError(t, err)

	type tokenJSON struct {
		DeliveryID uuid.UUID `json:"delivery_id"`
		OrderID    uuid.UUID `json:"order_id"`
		UserID     uuid.UUID `json:"user_id"`
		Nonce      uuid.UUID `json:"nonce"`
		IssuedAt   int64     `json:"iat"`
		ExpiresAt  int64     `json:"exp"`
	}

	var token tokenJSON
	require.NoError(t, json.Unmarshal(payloadBytes, &token))

	token.DeliveryID = uuid.New()
	modifiedPayload, err := json.Marshal(token)
	require.NoError(t, err)

	modifiedToken := base64.RawURLEncoding.EncodeToString(modifiedPayload) + "." + parts[1]

	_, err = ver.Verify(modifiedToken)
	require.ErrorIs(t, err, ErrInvalidSignature)
}

func TestQR_Verify_ModifiedExpiry(t *testing.T) {
	t.Parallel()

	gen, err := NewGenerator(testSecretKey, 24*time.Hour)
	require.NoError(t, err)

	ver, err := NewVerifier(testSecretKey)
	require.NoError(t, err)

	deliveryID := uuid.New()
	orderID := uuid.New()
	userID := uuid.New()
	nonce := uuid.New()

	resp, err := gen.Generate(deliveryID, orderID, userID, nonce)
	require.NoError(t, err)

	parts := strings.SplitN(resp.Token, ".", 2)
	require.Len(t, parts, 2)

	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	require.NoError(t, err)

	type tokenJSON struct {
		DeliveryID uuid.UUID `json:"delivery_id"`
		OrderID    uuid.UUID `json:"order_id"`
		UserID     uuid.UUID `json:"user_id"`
		Nonce      uuid.UUID `json:"nonce"`
		IssuedAt   int64     `json:"iat"`
		ExpiresAt  int64     `json:"exp"`
	}

	var token tokenJSON
	require.NoError(t, json.Unmarshal(payloadBytes, &token))

	token.ExpiresAt = time.Now().Add(72 * time.Hour).Unix()
	modifiedPayload, err := json.Marshal(token)
	require.NoError(t, err)

	modifiedToken := base64.RawURLEncoding.EncodeToString(modifiedPayload) + "." + parts[1]

	_, err = ver.Verify(modifiedToken)
	require.ErrorIs(t, err, ErrInvalidSignature, "extending expiry must invalidate signature")
}

func TestQR_InvalidKeyConfiguration(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		key  string
	}{
		{"empty key", ""},
		{"single char key", "a"},
		{"31 byte key", strings.Repeat("a", 31)},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(fmt.Sprintf("generator_%s", tt.name), func(t *testing.T) {
			t.Parallel()
			_, err := NewGenerator(tt.key, 24*time.Hour)
			require.Error(t, err)
		})
		t.Run(fmt.Sprintf("verifier_%s", tt.name), func(t *testing.T) {
			t.Parallel()
			_, err := NewVerifier(tt.key)
			require.Error(t, err)
		})
	}
}
