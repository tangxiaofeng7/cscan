package xerr

import "fmt"

// CodeError 业务错误
type CodeError struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

func (e *CodeError) Error() string {
	return fmt.Sprintf("code: %d, msg: %s", e.Code, e.Msg)
}

// NewCodeError 创建错误码错误
func NewCodeError(code int) *CodeError {
	return &CodeError{
		Code: code,
		Msg:  GetMsg(code),
	}
}

// NewCodeErrorMsg 创建自定义消息错误
func NewCodeErrorMsg(code int, msg string) *CodeError {
	return &CodeError{
		Code: code,
		Msg:  msg,
	}
}

// NewParamError 参数错误
func NewParamError(msg string) *CodeError {
	if msg == "" {
		msg = GetMsg(ParamError)
	}
	return &CodeError{Code: ParamError, Msg: msg}
}

// NewServerError 服务器错误
func NewServerError(msg string) *CodeError {
	if msg == "" {
		msg = GetMsg(ServerError)
	}
	return &CodeError{Code: ServerError, Msg: msg}
}

// NewNotFoundError 资源不存在错误
func NewNotFoundError(msg string) *CodeError {
	if msg == "" {
		msg = GetMsg(NotFound)
	}
	return &CodeError{Code: NotFound, Msg: msg}
}
