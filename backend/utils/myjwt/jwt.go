package myjwt

import (
	"fmt"
	"time"
	"wsai/backend/config"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	Id       int64  `json:"id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

func GenerateToken(id int64, username string) (string, error) {
	claims := Claims{
		Id:       id,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    config.C.JWTConfig.Issuer,
			Subject:   config.C.JWTConfig.Subject,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(config.C.JWTConfig.Secret))
}

func ParseToken(token string) (string, bool) {
	claims := new(Claims)
	t, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("不存在的签名方法: %v", t.Header["alg"])
		}
		return []byte(config.C.JWTConfig.Secret), nil
	})
	if !t.Valid || err != nil {
		return "", false
	}
	return claims.Username, true
}
