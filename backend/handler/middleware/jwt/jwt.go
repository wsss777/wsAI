package jwt

import (
	"fmt"
	"time"
	"wsai/backend/config"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Claims struct {
	Id        int64  `json:"id"`
	Username  string `json:"username"`
	TokenID   string `json:"jti"`
	TokenType string `json:"typ"`
	jwt.RegisteredClaims
}

func GenerateToken(id int64, username string) (string, error) {
	ttl, err := time.ParseDuration(config.C.JWTConfig.AccessTTL)
	if err != nil || ttl <= 0 {
		return "", fmt.Errorf("invalid jwt access_ttl")
	}
	now := time.Now()
	claims := Claims{
		Id:        id,
		Username:  username,
		TokenID:   uuid.NewString(),
		TokenType: "access",
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    config.C.JWTConfig.Issuer,
			Subject:   config.C.JWTConfig.Subject,
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(config.C.JWTConfig.Secret))
}

func GenerateRefreshToken(id int64, username string) (string, error) {
	ttl, err := time.ParseDuration(config.C.JWTConfig.RefreshTTL)
	if err != nil || ttl <= 0 {
		return "", fmt.Errorf("invalid jwt refresh_ttl")
	}
	now := time.Now()
	claims := Claims{Id: id, Username: username, TokenID: uuid.NewString(), TokenType: "refresh", RegisteredClaims: jwt.RegisteredClaims{Issuer: config.C.JWTConfig.Issuer, Subject: config.C.JWTConfig.Subject, ExpiresAt: jwt.NewNumericDate(now.Add(ttl)), IssuedAt: jwt.NewNumericDate(now)}}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(config.C.JWTConfig.Secret))
}

func ParseToken(token string) (*Claims, error) {
	claims := new(Claims)
	t, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (interface{}, error) {
		if t.Method.Alg() != jwt.SigningMethodHS256.Alg() {
			return nil, fmt.Errorf("不存在的签名方法: %v", t.Header["alg"])
		}
		return []byte(config.C.JWTConfig.Secret), nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}), jwt.WithIssuer(config.C.JWTConfig.Issuer), jwt.WithSubject(config.C.JWTConfig.Subject))
	if !t.Valid || err != nil {
		return nil, err
	}
	if claims.Username == "" || claims.Id <= 0 || claims.TokenID == "" || claims.TokenType == "" {
		return nil, fmt.Errorf("jwt claims incomplete")
	}
	return claims, nil
}
