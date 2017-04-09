package main

import (
	"fmt"
	"os"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
)

const twoWeeks = 2 * 7 * 24 * time.Hour

var jwtSecret = []byte(os.Getenv("JWT_SECRET"))

type HakoCustomClaims struct {
	Typ string `json:"typ"`
	Id  string `json:"id"`
	jwt.StandardClaims
}

func createToken(typ string, id string, expireIn time.Duration) (string, error) {
	standardClaims := jwt.StandardClaims{ExpiresAt: time.Now().UTC().Add(expireIn).Unix()}
	claims := HakoCustomClaims{typ, id, standardClaims}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

func createLoginToken(id string) (string, error) {
	return createToken("login", id, 15*time.Minute)
}

func createSessionToken(id string) (string, error) {
	return createToken("session", id, twoWeeks)
}

func validateToken(typ, tokenStr string) (string, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &HakoCustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		if claims, ok := token.Claims.(*HakoCustomClaims); !ok || claims.Typ != typ {
			return nil, fmt.Errorf("Unexpected token type: %s (wanted %s)", claims.Typ, typ)
		}
		return jwtSecret, nil
	})

	if claims, ok := token.Claims.(*HakoCustomClaims); ok && token.Valid {
		return claims.Id, nil
	} else {
		return "", err
	}
}

func validateLoginToken(tokenStr string) (string, error) {
	return validateToken("login", tokenStr)
}

func validateSessionToken(tokenStr string) (string, error) {
	return validateToken("session", tokenStr)
}
