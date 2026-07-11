package auth

import (
	"fmt"

	"github.com/golang-jwt/jwt/v5"
)

type Manager struct {
	secret []byte
	issuer string
}

func New(secret, issuer string) *Manager {
	return &Manager{
		secret: []byte(secret),
		issuer: issuer,
	}
}

func (m *Manager) Parse(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(
		tokenString, &Claims{},
		func(t *jwt.Token) (any, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("%w: %v", ErrInvalidToken, t.Header["alg"])
			} 

			return m.secret, nil
		},
		jwt.WithIssuer(m.issuer),
	)

	if err != nil {
		return nil, fmt.Errorf("parse jwt: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	} 

	return claims, nil
}