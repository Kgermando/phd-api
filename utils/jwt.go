package utils

import (
	"time"

	"github.com/golang-jwt/jwt/v4"
)

func secretKey() string {
	return Env("SECRET_KEY")
}

func GenerateJwt(issuer string) (string, error) {

	claims := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:    issuer,
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24)),
	})

	token, err := claims.SignedString([]byte(secretKey()))

	return token, err
}

func VerifyJwt(cookie string) (string, error) {

	claims := &jwt.RegisteredClaims{}
	token, err := jwt.ParseWithClaims(cookie, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(secretKey()), nil
	})

	if err != nil || !token.Valid {
		return "", err
	}

	return claims.Issuer, nil
}
