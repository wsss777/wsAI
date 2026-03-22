package jwt

import (
	"fmt"
	"strings"
	"time"
	"wsai/backend/config"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	Id       int64  `json:"id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

func GenerateToken(id int64, username string) (string, error) {
	accessTTL := 24 * time.Hour
	if config.C != nil && config.C.JWTConfig.AccessTTL != "" {
		if parsedTTL, err := time.ParseDuration(config.C.JWTConfig.AccessTTL); err == nil {
			accessTTL = parsedTTL
		}
	}

	claims := Claims{
		Id:       id,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    config.C.JWTConfig.Issuer,
			Subject:   config.C.JWTConfig.Subject,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(accessTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(config.C.JWTConfig.Secret))
}

func ParseTokenClaims(token string) (*Claims, bool) {
	claims := new(Claims)
	t, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("不支持的签名算法: %v", t.Header["alg"])
		}
		return []byte(config.C.JWTConfig.Secret), nil
	})
	if !t.Valid || err != nil {
		return nil, false
	}
	return claims, true
}

func ParseToken(token string) (string, bool) {
	claims, ok := ParseTokenClaims(token)
	if !ok {
		return "", false
	}
	return claims.Username, true
}

func ExtractToken(c *gin.Context) string {
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
		return strings.TrimPrefix(authHeader, "Bearer ")
	}
	return c.Query("token")
}
