package repository

import (
	"context"
	"wsai/backend/infra/mysql"
	"wsai/backend/model"
)

const (
	CodeMsg     = "wsai的验证码如下（为保障信息安全，请勿告诉他人）"
	UserNameMsg = "wsai的账号如下，请保存好，以登陆账号使用"
)

var ctx = context.Background()

func IsExistUser(username string) (bool, *model.User) {
	user := &model.User{}
	if err := mysql.DB.Where("username = ?", username).First(user).Error; err != nil {
		return false, nil
	}
	return true, user
}

func IsExistUserWithEmail(email string) (bool, *model.User) {
	user := &model.User{}
	if err := mysql.DB.Where("email = ?", email).First(user).Error; err != nil {
		return false, nil
	}
	return true, user
}

func Register(username, email, password string) (*model.User, bool) {
	user := &model.User{
		Email:    email,
		Name:     username,
		Username: username,
		Password: password,
	}
	if err := mysql.DB.Create(user).Error; err != nil {
		return nil, false
	}
	return user, true
}

func UpdatePassword(userID int64, passwordHash string) error {
	return mysql.DB.Model(&model.User{}).Where("id = ?", userID).Update("password", passwordHash).Error
}
