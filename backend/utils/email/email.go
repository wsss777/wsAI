package email

import (
	"fmt"
	"wsai/backend/config"

	"gopkg.in/gomail.v2"
)

const (
	CodeMsg     = "wsai的验证码如下（为保障信息安全，请勿告诉他人）"
	UserNameMsg = "wsai的账号如下，请保存好，以登陆账号使用"
)

func SendCaptcha(email, code, msg string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", config.C.EmailConfig.Email)
	m.SetHeader("To", email)
	m.SetHeader("Subject", "来自wsai的信息")
	m.SetBody("text/plain", msg+" "+code)

	d := gomail.NewDialer(config.C.EmailConfig.Host, 587, config.C.EmailConfig.Email, config.C.EmailConfig.Authcode)
	if err := d.DialAndSend(m); err != nil {
		fmt.Printf("DialAndSend err %v:\n", err)
		return err
	}
	fmt.Println("Sending Email Success")
	return nil
}
