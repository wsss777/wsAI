package session

import (
	"net/http"
	"wsai/backend/internal/common"
	"wsai/backend/internal/common/code"
	"wsai/backend/internal/model"

	"github.com/gin-gonic/gin"
)

type (
	GetUserSessionsResponse struct {
		Sessions []model.SessionInfo `json:"sessions,omitempty"`
		common.Response
	}
)

func GetUserSessionsByUsername(c *gin.Context) {
	res := new(GetUserSessionsResponse)
	username_ := c.GetString("username")

	userSessions, err := session.GetUserSessionByUsername(username_)
	if err != nil {
		c.JSON(http.StatusOK, res.CodeOf(code.CodeServerBusy))
		return
	}
	res.Success()
	res.Sessions = userSessions
	c.JSON(http.StatusOK, res)
}
