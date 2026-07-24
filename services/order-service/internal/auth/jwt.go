package auth

import (
	"errors"

	"github.com/golang-jwt/jwt/v5"
)

type Validator struct {
	secret []byte
	issuer string
}

func NewValidator(secret string, issuer string) *Validator {
	return &Validator{
		secret: []byte(secret),
		issuer: issuer,
	}
}

func (v *Validator) Parse(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (any, error) {

			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("unexpected signing method")
			}
			return v.secret, nil
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