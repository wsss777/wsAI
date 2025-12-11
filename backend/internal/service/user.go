package user

import (
	"wsai/backend/internal/model"
	"wsai/backend/internal/repository/user"
	"wsai/backend/utils"
	"wsai/backend/utils/common"
	"wsai/backend/utils/myjwt"
)

func Login(username, password string) (string, common.Code) {
	var userInformation *model.User
	var ok bool
	//检查是否存在
	if ok, userInformation = user.IsExistUser(username); !ok {
		return "", common.CodeUserNotExist
	}
	//判断密码是否正确
	if userInformation.Password != utils.MD5(password) {
		return "", common.CodeInvalidPassword
	}
	token, err := myjwt.GenerateToken(userInformation.ID, userInformation.Username)
	if err != nil {
		return "", common.CodeServerBusy
	}
	return token, common.CodeSuccess
}
