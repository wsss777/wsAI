package jwt

import (
	"context"
	"net/http"
	"time"
	"wsai/backend/internal/common"
	"wsai/backend/internal/common/code"
	"wsai/backend/internal/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		res := new(common.Response)

		token := ExtractToken(c)
		if token == "" {
			c.JSON(http.StatusOK, res.CodeOf(code.CodeInvalidToken))
			c.Abort()
			return
		}

		claims, ok := ParseTokenClaims(token)
		if !ok {
			c.JSON(http.StatusOK, res.CodeOf(code.CodeInvalidToken))
			c.Abort()
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		isBlacklisted, err := IsTokenBlacklisted(ctx, token)
		if err != nil && logger.L() != nil {
			logger.L().Warn("检查 Token 黑名单失败", zap.Error(err))
		}
		if isBlacklisted {
			c.JSON(http.StatusOK, res.CodeOf(code.CodeInvalidToken))
			c.Abort()
			return
		}

		c.Set("username", claims.Username)
		c.Set("userID", claims.Id)
		c.Next()
	}
}
