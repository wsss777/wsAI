package user

import (
	"wsai/backend/internal/common/mysql"
	"wsai/backend/internal/model"
)

func InsertUser(user *model.User) (*model.User, error) {
	err := mysql.DB.Create(user).Error
	return user, err
}

func GetUserByUsername(username string) (*model.User, error) {
	user := &model.User{}
	err := mysql.DB.Where("username = ?", username).First(user).Error
	return user, err
}
func GetUserByEmail(email string) (*model.User, error) {
	user := &model.User{}
	err := mysql.DB.Where("email = ?", email).First(user).Error
	return user, err
}
