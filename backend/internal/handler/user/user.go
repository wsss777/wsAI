package user

import (
	"net/http"
	"wsai/backend/internal/common"
	"wsai/backend/internal/common/code"

	//"wsai/backend/internal/service"
	"wsai/backend/internal/service/user"

	"github.com/gin-gonic/gin"
)

type (
	LoginRequest struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	LoginResponse struct {
		Token string `json:"token"`
		common.Response
	}
	RegisterRequest struct {
		Email    string `json:"email"binding:"required"`
		Captcha  string `json:"captcha"`
		Password string `json:"password"`
	}
	RegisterResponse struct {
		Token string `json:"token,omitempty"`
		common.Response
	}
	CaptchaRequest struct {
		Email string `json:"email"binding:"required"`
	}
	CaptchaResponse struct {
		common.Response
	}
)

// Login godoc
// @Summary 用户登录
// @Description 根据用户名和密码登录，返回 JWT token
// @Tags 用户认证
// @Accept json
// @Produce json
// @Param request body LoginRequest true "登录参数"
// @Success 200 {object} LoginResponse "登录成功，返回 token"
// @Failure 400 {object} common.Response "参数错误"
// @Failure 401 {object} common.Response "用户名或密码错误"
// @Router /api/v1/user/login [post]
func Login(c *gin.Context) {
	req := new(LoginRequest)
	res := new(LoginResponse)
	if err := c.ShouldBindJSON(req); err != nil {
		c.JSON(http.StatusOK, res.CodeOf(code.CodeInvalidParams))
		return
	}
	token, code_ := user.Login(req.Username, req.Password)
	if code_ != code.CodeSuccess {
		c.JSON(http.StatusOK, res.CodeOf(code_))
	}
	res.Success()
	res.Token = token
	c.JSON(http.StatusOK, res)
}

// Register godoc
// @Summary 用户注册
// @Description 通过邮箱、密码和验证码注册新用户，成功后直接返回登录 token
// @Tags 用户认证
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "注册参数"
// @Success 200 {object} RegisterResponse "注册成功，返回 token"
// @Failure 400 {object} common.Response "参数错误或验证码错误"
// @Failure 409 {object} common.Response "用户已存在"
// @Router /api/v1/user/users [post]
func Register(c *gin.Context) {
	req := new(RegisterRequest)
	res := new(RegisterResponse)
	if err := c.ShouldBindJSON(req); err != nil {
		c.JSON(http.StatusOK, res.CodeOf(code.CodeInvalidParams))
		return
	}
	token, code_ := user.Register(req.Email, req.Password, req.Captcha)
	if code_ != code.CodeSuccess {
		c.JSON(http.StatusOK, res.CodeOf(code_))
		return

	}
	res.Success()
	res.Token = token
	c.JSON(http.StatusOK, res)
}

// HandleCaptcha godoc
// @Summary 发送邮箱验证码
// @Description 向指定邮箱发送注册验证码
// @Tags 用户认证
// @Accept json
// @Produce json
// @Param request body CaptchaRequest true "邮箱参数"
// @Success 200 {object} CaptchaResponse "验证码发送成功"
// @Failure 400 {object} common.Response "邮箱格式错误"
// @Failure 429 {object} common.Response "发送过于频繁"
// @Router /api/v1/user/captcha [post]
func HandleCaptcha(c *gin.Context) {
	req := new(CaptchaRequest)
	res := new(CaptchaResponse)
	if err := c.ShouldBindJSON(req); err != nil {
		c.JSON(http.StatusOK, res.CodeOf(code.CodeInvalidParams))
		return
	}

	code_ := user.SendCaptcha(req.Email)
	if code_ != code.CodeSuccess {
		c.JSON(http.StatusOK, res.CodeOf(code_))
		return
	}
	res.Success()
	c.JSON(http.StatusOK, res)
}
