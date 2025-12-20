package common

import "github.com/gin-gonic/gin"

type Resp struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

func Error(c *gin.Context, code int, msg string) {
	c.JSON(200, Resp{Code: code, Msg: msg, Data: nil})
}

func Success(c *gin.Context, data interface{}) {
	c.JSON(200, Resp{Code: 200, Msg: "success", Data: data})
}
