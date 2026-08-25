package jwt

import (
	"net/http"
	"strings"
	"wsai/backend/response"
	"wsai/backend/response/code"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		res := new(common.Response)

		token, ok := bearerToken(c.GetHeader("Authorization"))
		if !ok {
			c.JSON(http.StatusOK, res.CodeOf(code.CodeInvalidToken))
			c.Abort()
			return
		}

		claims, err := ParseToken(token)
		if err != nil {
			c.JSON(http.StatusOK, res.CodeOf(code.CodeInvalidToken))
			c.Abort()
			return
		}
		blacklisted, err := IsTokenBlacklisted(c.Request.Context(), token)
		if err != nil {
			c.JSON(http.StatusOK, res.CodeOf(code.CodeServerBusy))
			c.Abort()
			return
		}
		if blacklisted {
			c.JSON(http.StatusOK, res.CodeOf(code.CodeInvalidToken))
			c.Abort()
			return
		}
		c.Set("username", claims.Username)
		c.Set("jwt_token", token)
		c.Set("jwt_claims", claims)
		c.Next()

	}
}

func bearerToken(header string) (string, bool) {
	parts := strings.Fields(header)
	returnValue := ""
	if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") && parts[1] != "" {
		returnValue = parts[1]
	}
	return returnValue, returnValue != ""
}
