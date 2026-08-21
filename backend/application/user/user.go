package user

import (
	"context"
	"time"
	"wsai/backend/application/captcha"
	myemail "wsai/backend/application/email"
	"wsai/backend/handler/middleware/jwt"
	user "wsai/backend/infra/mysql/repository"
	"wsai/backend/model"
	"wsai/backend/response/code"
	"wsai/backend/utils"
)

func Login(username, password string) (string, code.Code) {
	var userInformation *model.User
	var ok bool
	//检查是否存在
	if ok, userInformation = user.IsExistUser(username); !ok {
		return "", code.CodeUserNotExist
	}
	//判断密码是否正确
	if userInformation.Password != utils.MD5(password) {
		return "", code.CodeInvalidPassword
	}
	// 返回令牌。
	token, err := jwt.GenerateToken(userInformation.ID, userInformation.Username)
	if err != nil {
		return "", code.CodeServerBusy
	}
	return token, code.CodeSuccess
}

// LoginWithEmail 使用邮箱和密码完成登录，并签发 JWT。
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
	//1:先判断用户是否已经存在了
	if ok, _ = user.IsExistUserWithEmail(email); ok {
		return "", code.CodeUserExist
	}
	// 2：从 Redis 中验证验证码是否有效。
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if ok, _ := captcha.CheckCaptchaForEmail(ctx, email, captcha_); !ok {
		return "", code.CodeInvalidCaptcha
	}
	//3：生成11位的账号
	username := utils.GetRandomNumbers(11)
	//4：注册到数据库中
	if userInformation, ok = user.Register(username, email, password); !ok {
		return "", code.CodeServerBusy
	}
	//5：将账号一并发送到对应邮箱上去，后续需要账号登录
	if err := myemail.SendCaptcha(email, username, user.UserNameMsg); err != nil {
		return "", code.CodeServerBusy
	}
	// 6：生成令牌。
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
