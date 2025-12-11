package user

import (
	"net/http"
	"wsai/backend/internal/service"
	common2 "wsai/backend/utils/common"

	"github.com/gin-gonic/gin"
)

type (
	LoginRequest struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	LoginResponse struct {
		Token string `json:"token"`
		common2.Response
	}
)

func Login(c *gin.Context) {
	req := new(LoginRequest)
	res := new(LoginResponse)
	if err := c.ShouldBindJSON(req); err != nil {
		c.JSON(http.StatusOK, res.CodeOf(common2.CodeInvalidParams))
		return
	}
	token, code_ := user.Login(req.Username, req.Password)
	if code_ != common2.CodeSuccess {
		c.JSON(http.StatusOK, res.CodeOf(code_))
	}
	res.Success()
	res.Token = token
	c.JSON(http.StatusOK, res)
}
