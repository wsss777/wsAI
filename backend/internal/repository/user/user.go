package user

import (
	"context"
	"wsai/backend/internal/model"
	"wsai/backend/utils"

	"gorm.io/gorm"
)

const (
	CodeMsg     = "wsai的验证码如下（为保障信息安全，请勿告诉他人）"
	UserNameMsg = "wsai的账号如下，请保存好，以登陆账号使用"
)

var ctx = context.Background()

func IsExistUser(username string) (bool, *model.User) {
	user, err := GetUserByUsername(username)
	if err == gorm.ErrRecordNotFound || user == nil {
		return false, nil
	}
	return true, user
}

func IsExistUserWithEmail(email string) (bool, *model.User) {
	user, err := GetUserByEmail(email)
	if err == gorm.ErrRecordNotFound || user == nil {
		return false, nil
	}
	return true, user
}
func Register(username, email, password string) (*model.User, bool) {
	if user, err := InsertUser(&model.User{
		Email:    email,
		Name:     username,
		Username: username,
		Password: utils.MD5(password),
	}); err != nil {
		return nil, false
	} else {
		return user, true
	}
}
