package response

import "go-order-lite/pkg/errno"

type Response struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data,omitempty"`
}

func Success(data interface{}) Response {
	return Response{
		Code: errno.OK.Code,
		Msg:  errno.OK.Msg,
		Data: data,
	}
}

func Fail(err *errno.Error) Response {
	return Response{
		Code: err.Code,
		Msg:  err.Msg,
	}
}
