package user

import (
	"net/http"
	"time"
	jwtmiddleware "wsai/backend/handler/middleware/jwt"
	"wsai/backend/infra/logger"
	"wsai/backend/response"
	"wsai/backend/response/code"

	"wsai/backend/application/user"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Logout 退出登录并吊销当前会话凭证。
// @Summary      退出登录
// @Description  吊销当前 Access Token、撤销 Refresh Token，并清除 Refresh Cookie。
// @Tags         用户认证
// @Produce      json
// @Security     ApiKeyAuth
// @Success      200  {object}  common.Response  "退出成功"
// @Failure      200  {object}  common.Response  "Token 无效或服务异常"
// @Router       /api/v1/user/logout [post]
func Logout(c *gin.Context) {
	res := new(common.Response)
	token, tokenOK := c.Get("jwt_token")
	claims, claimsOK := c.Get("jwt_claims")
	if !tokenOK || !claimsOK {
		c.JSON(http.StatusOK, res.CodeOf(code.CodeInvalidToken))
		return
	}
	jwtClaims, ok := claims.(*jwtmiddleware.Claims)
	if !ok || jwtClaims.ExpiresAt == nil {
		c.JSON(http.StatusOK, res.CodeOf(code.CodeInvalidToken))
		return
	}
	if err := jwtmiddleware.AddTokenToBlacklist(c.Request.Context(), token.(string), jwtClaims.ExpiresAt.Time); err != nil {
		c.JSON(http.StatusOK, res.CodeOf(code.CodeServerBusy))
		return
	}
	if refreshToken, err := c.Cookie("wsai_refresh_token"); err == nil {
		_ = jwtmiddleware.RevokeRefreshToken(c.Request.Context(), refreshToken)
	}
	c.SetCookie("wsai_refresh_token", "", -1, "/api/v1/user", "", false, true)
	res.Success()
	c.JSON(http.StatusOK, res)
}

// Refresh 使用 HttpOnly Refresh Cookie 轮换令牌。
// @Summary      刷新 Access Token
// @Description  浏览器自动携带 wsai_refresh_token Cookie。服务端验证并一次性消费旧 Refresh Token，返回新的 Access Token 并轮换 Cookie。
// @Tags         用户认证
// @Produce      json
// @Param        Cookie  header  string  false  "HttpOnly Refresh Token Cookie（浏览器自动携带）"
// @Success      200  {object}  LoginResponse  "刷新成功，返回新的 Access Token"
// @Failure      200  {object}  common.Response  "Refresh Token 无效、已过期或服务异常"
// @Router       /api/v1/user/refresh [post]
func Refresh(c *gin.Context) {
	res := new(LoginResponse)
	refreshToken, err := c.Cookie("wsai_refresh_token")
	if err != nil {
		logger.L().Warn("刷新令牌失败：请求未携带 Cookie", zap.String("origin", c.GetHeader("Origin")), zap.Error(err))
		c.JSON(http.StatusOK, res.CodeOf(code.CodeInvalidToken))
		return
	}
	claims, err := jwtmiddleware.ParseToken(refreshToken)
	if err != nil || claims.TokenType != "refresh" {
		logger.L().Warn("刷新令牌失败：Cookie 中的令牌无效", zap.Error(err))
		c.JSON(http.StatusOK, res.CodeOf(code.CodeInvalidToken))
		return
	}
	accessToken, err := jwtmiddleware.GenerateToken(claims.Id, claims.Username)
	if err != nil {
		logger.L().Error("刷新令牌失败：生成 Access Token 失败", zap.String("username", claims.Username), zap.Error(err))
		c.JSON(http.StatusOK, res.CodeOf(code.CodeServerBusy))
		return
	}
	newRefreshToken, refreshExpireAt, err := newRefreshToken(claims.Id, claims.Username)
	if err != nil {
		logger.L().Error("刷新令牌失败：生成新 Refresh Token 失败", zap.String("username", claims.Username), zap.Error(err))
		c.JSON(http.StatusOK, res.CodeOf(code.CodeServerBusy))
		return
	}
	ok, err := jwtmiddleware.RotateRefreshToken(c.Request.Context(), refreshToken, newRefreshToken, refreshExpireAt)
	if err != nil {
		logger.L().Error("刷新令牌失败：Redis 轮换失败", zap.String("username", claims.Username), zap.Error(err))
		c.JSON(http.StatusOK, res.CodeOf(code.CodeServerBusy))
		return
	}
	if !ok {
		logger.L().Warn("刷新令牌失败：Redis 中不存在或已消费", zap.String("username", claims.Username))
		c.JSON(http.StatusOK, res.CodeOf(code.CodeInvalidToken))
		return
	}
	setRefreshCookieValue(c, newRefreshToken, refreshExpireAt)
	logger.L().Info("刷新令牌成功", zap.String("username", claims.Username))
	res.Success()
	res.Token = accessToken
	c.JSON(http.StatusOK, res)
}

func setRefreshCookie(c *gin.Context, id int64, username string) error {
	token, expireAt, err := newRefreshToken(id, username)
	if err != nil {
		return err
	}
	if err := jwtmiddleware.StoreRefreshToken(c.Request.Context(), token, expireAt); err != nil {
		return err
	}
	setRefreshCookieValue(c, token, expireAt)
	return nil
}

func newRefreshToken(id int64, username string) (string, time.Time, error) {
	token, err := jwtmiddleware.GenerateRefreshToken(id, username)
	if err != nil {
		return "", time.Time{}, err
	}
	claims, err := jwtmiddleware.ParseToken(token)
	if err != nil || claims.ExpiresAt == nil {
		if err == nil {
			err = http.ErrNoCookie
		}
		return "", time.Time{}, err
	}
	return token, claims.ExpiresAt.Time, nil
}

func setRefreshCookieValue(c *gin.Context, token string, expireAt time.Time) {
	seconds := int(time.Until(expireAt).Seconds())
	c.SetCookie("wsai_refresh_token", token, seconds, "/api/v1/user", "", false, true)
}

func setRefreshCookieForAccess(c *gin.Context, accessToken string) error {
	claims, err := jwtmiddleware.ParseToken(accessToken)
	if err != nil {
		return err
	}
	return setRefreshCookie(c, claims.Id, claims.Username)
}

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
	if err := setRefreshCookieForAccess(c, token); err != nil {
		c.JSON(http.StatusOK, res.CodeOf(code.CodeServerBusy))
		return
	}
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
	if err := setRefreshCookieForAccess(c, token); err != nil {
		c.JSON(http.StatusOK, res.CodeOf(code.CodeServerBusy))
		return
	}
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
	if err := setRefreshCookieForAccess(c, token); err != nil {
		c.JSON(http.StatusOK, res.CodeOf(code.CodeServerBusy))
		return
	}
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
