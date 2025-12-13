package user

import (
	"context"
	"time"
	"wsai/backend/internal/model"
	"wsai/backend/internal/repository/user"
	"wsai/backend/utils"
	"wsai/backend/utils/common/code"
	myemail "wsai/backend/utils/email"
	"wsai/backend/utils/myjwt"
	"wsai/backend/utils/redis"
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
	//返回token
	token, err := myjwt.GenerateToken(userInformation.ID, userInformation.Username)
	if err != nil {
		return "", code.CodeServerBusy
	}
	return token, code.CodeSuccess
}

func Register(email, password, captcha string) (string, code.Code) {
	var ok bool
	var userInformation *model.User
	//1:先判断用户是否已经存在了
	if ok, _ = user.IsExistUserWithEmail(email); ok {
		return "", code.CodeUserExist
	}
	//2:从redis中验证验证码是否有效
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if ok, _ := redis.CheckCaptchaForEmail(ctx, email, captcha); !ok {
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
	// 6:生成Token
	token, err := myjwt.GenerateToken(userInformation.ID, userInformation.Username)
	if err != nil {
		return "", code.CodeServerBusy
	}
	return token, code.CodeSuccess

}

func SendCaptcha(email_ string) code.Code {
	send_code := utils.GetRandomNumbers(6)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := redis.SetCaptchaForEmail(ctx, email_, send_code); err != nil {
		return code.CodeServerBusy
	}

	if err := myemail.SendCaptcha(email_, send_code, myemail.CodeMsg); err != nil {
		return code.CodeServerBusy
	}
	return code.CodeSuccess
}
