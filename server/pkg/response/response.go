package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Code int         `json:"code"`
	Data interface{} `json:"data"`
	Msg  string      `json:"msg"`
}

func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code: 0,
		Data: data,
		Msg:  "ok",
	})
}

func Error(c *gin.Context, httpCode int, msg string) {
	c.JSON(httpCode, Response{
		Code: -1,
		Data: nil,
		Msg:  msg,
	})
}

func ErrorWithCode(c *gin.Context, httpCode int, code int, msg string) {
	c.JSON(httpCode, Response{
		Code: code,
		Data: nil,
		Msg:  msg,
	})
}
