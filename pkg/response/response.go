package response

import (
	"net/http"

	"cscan/pkg/xerr"

	"github.com/zeromicro/go-zero/rest/httpx"
)

// Response 统一响应结构
type Response struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data,omitempty"`
}

// Success 成功响应
func Success(w http.ResponseWriter, data interface{}) {
	httpx.OkJson(w, &Response{
		Code: xerr.OK,
		Msg:  "success",
		Data: data,
	})
}

// SuccessWithMsg 成功响应带消息
func SuccessWithMsg(w http.ResponseWriter, msg string) {
	httpx.OkJson(w, &Response{
		Code: xerr.OK,
		Msg:  msg,
	})
}

// Error 错误响应
func Error(w http.ResponseWriter, err error) {
	if codeErr, ok := err.(*xerr.CodeError); ok {
		httpx.OkJson(w, &Response{
			Code: codeErr.Code,
			Msg:  codeErr.Msg,
		})
		return
	}
	httpx.OkJson(w, &Response{
		Code: xerr.ServerError,
		Msg:  err.Error(),
	})
}

// ErrorWithCode 指定错误码响应
func ErrorWithCode(w http.ResponseWriter, code int, msg string) {
	if msg == "" {
		msg = xerr.GetMsg(code)
	}
	httpx.OkJson(w, &Response{
		Code: code,
		Msg:  msg,
	})
}

// ParamError 参数错误响应
func ParamError(w http.ResponseWriter, msg string) {
	if msg == "" {
		msg = xerr.GetMsg(xerr.ParamError)
	}
	httpx.OkJson(w, &Response{
		Code: xerr.ParamError,
		Msg:  msg,
	})
}
