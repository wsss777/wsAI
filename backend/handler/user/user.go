package user

import (
	"net/http"
	"wsai/backend/response"
	"wsai/backend/response/code"

	"wsai/backend/application/user"

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
	EmailLoginRequest struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
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

// Login 接口文档
// @Summary 用户登录
// @Description 根据用户名和密码登录，返回 JWT 令牌
// @Tags 用户认证
// @Accept json
// @Produce json
// @Param request body LoginRequest true "登录参数"
// @Success 200 {object} LoginResponse "登录成功，返回令牌"
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
		return
	}
	res.Success()
	res.Token = token
	c.JSON(http.StatusOK, res)
}

// EmailLogin 使用邮箱和密码登录。
// @Summary      邮箱登录
// @Description  根据已注册邮箱和密码登录，成功后返回 JWT 令牌。
// @Tags         用户认证
// @Accept       json
// @Produce      json
// @Param        request  body  EmailLoginRequest  true  "邮箱登录参数"
// @Success      200      {object}  LoginResponse     "登录成功，返回令牌"
// @Failure      200      {object}  common.Response   "参数错误、邮箱不存在或密码错误；业务错误使用 HTTP 200 返回"
// @Router       /api/v1/user/email-login [post]
func EmailLogin(c *gin.Context) {
	req := new(EmailLoginRequest)
	res := new(LoginResponse)
	if err := c.ShouldBindJSON(req); err != nil {
		c.JSON(http.StatusOK, res.CodeOf(code.CodeInvalidParams))
		return
	}
	token, code_ := user.LoginWithEmail(req.Email, req.Password)
	if code_ != code.CodeSuccess {
		c.JSON(http.StatusOK, res.CodeOf(code_))
		return
	}
	res.Success()
	res.Token = token
	c.JSON(http.StatusOK, res)
}

// Register 接口文档
// @Summary 用户注册
// @Description 通过邮箱、密码和验证码注册新用户，成功后直接返回登录令牌
// @Tags 用户认证
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "注册参数"
// @Success 200 {object} RegisterResponse "注册成功，返回令牌"
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

// HandleCaptcha 接口文档
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
