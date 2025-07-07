package lib

import (
	"github.com/gin-gonic/gin"
)

type APIResponse struct {
	Success bool        `json:"success" default:"true" `
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

type HTTPError struct {
	Success bool   `json:"success" default:"false" `
	Message string `json:"message,omitempty"`
	Code    int    `json:"code,omitempty"`
	Error   error  `json:"error,omitempty"`
}

func RespondError(ctx *gin.Context, code int, message string, err error, errCode int, debug bool) {
	res := HTTPError{
		Success: false,
		Message: message,
		Code:    errCode,
	}
	if debug {
		res.Error = err
	}
	ctx.JSON(code, res)
}

func RespondSuccess(ctx *gin.Context, code int, message string, data interface{}) {
	ctx.JSON(code, APIResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// func RespondSuccessPaginated(ctx *gin.Context, code int, message string, data interface{}, pagination Pagination) {
// 	ctx.JSON(code, APIResponsePaginated{
// 		APIResponse: APIResponse{
// 			Success: true,
// 			Message: message,
// 			Data:    data,
// 		},
// 		Pagination: pagination,
// 	})
// }
