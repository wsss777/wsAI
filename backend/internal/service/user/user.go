package user

import (
	"context"
	"time"
	"wsai/backend/internal/common/code"
	"wsai/backend/internal/middleware/jwt"
	"wsai/backend/internal/model"
	"wsai/backend/internal/repository/user"
	"wsai/backend/internal/service/captcha"
	myemail "wsai/backend/internal/service/email"
	"wsai/backend/utils"
)

func Login(username, password string) (string, code.Code) {
	var userInformation *model.User
	var ok bool

	if ok, userInformation = user.IsExistUser(username); !ok {
		return "", code.CodeUserNotExist
	}
	if userInformation.Password != utils.MD5(password) {
		return "", code.CodeInvalidPassword
	}

	token, err := jwt.GenerateToken(userInformation.ID, userInformation.Username)
	if err != nil {
		return "", code.CodeServerBusy
	}
	return token, code.CodeSuccess
}

func LoginWithEmail(email, password string) (string, code.Code) {
	var userInformation *model.User
	var ok bool

	if ok, userInformation = user.IsExistUserWithEmail(email); !ok {
		return "", code.CodeUserNotExist
	}
	if userInformation.Password != utils.MD5(password) {
		return "", code.CodeInvalidPassword
	}

	token, err := jwt.GenerateToken(userInformation.ID, userInformation.Username)
	if err != nil {
		return "", code.CodeServerBusy
	}
	return token, code.CodeSuccess
}

func Register(email, password, captcha_ string) (string, code.Code) {
	var ok bool
	var userInformation *model.User

	if ok, _ = user.IsExistUserWithEmail(email); ok {
		return "", code.CodeUserExist
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if ok, _ := captcha.CheckCaptchaForEmail(ctx, email, captcha_); !ok {
		return "", code.CodeInvalidCaptcha
	}

	username := utils.GetRandomNumbers(11)
	if userInformation, ok = user.Register(username, email, password); !ok {
		return "", code.CodeServerBusy
	}
	if err := myemail.SendCaptcha(email, username, user.UserNameMsg); err != nil {
		return "", code.CodeServerBusy
	}

	token, err := jwt.GenerateToken(userInformation.ID, userInformation.Username)
	if err != nil {
		return "", code.CodeServerBusy
	}
	return token, code.CodeSuccess
}

func SendCaptcha(email_ string) code.Code {
	sendCode := utils.GetRandomNumbers(6)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := captcha.SetCaptchaForEmail(ctx, email_, sendCode); err != nil {
		return code.CodeServerBusy
	}

	if err := myemail.SendCaptcha(email_, sendCode, myemail.CodeMsg); err != nil {
		return code.CodeServerBusy
	}
	return code.CodeSuccess
}

func Logout(token string) code.Code {
	claims, ok := jwt.ParseTokenClaims(token)
	if !ok || claims.ExpiresAt == nil {
		return code.CodeInvalidToken
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := jwt.AddTokenToBlacklist(ctx, token, claims.ExpiresAt.Time); err != nil {
		return code.CodeServerBusy
	}
	return code.CodeSuccess
}
