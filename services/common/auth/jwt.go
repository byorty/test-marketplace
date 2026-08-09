package auth

import (
	"crypto/rsa"
	"errors"

	"github.com/golang-jwt/jwt/v5"
)

type Validator struct {
	publicKey *rsa.PublicKey
	issuer    string
}

func NewValidator(publicKey *rsa.PublicKey, issuer string) *Validator {
	return &Validator{
		publicKey: publicKey,
		issuer: issuer,
	}
}

func (v *Validator) Parse(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(
		tokenString,
		&Claims{},
		func(token *jwt.Token) (any, error) {
			if token.Method.Alg() != jwt.SigningMethodRS256.Alg() {
				return nil, errors.New("unexpected signing method")
			}

			return v.publicKey, nil
		},
		jwt.WithIssuer(v.issuer),
	)

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}