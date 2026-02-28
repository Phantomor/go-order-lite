package middleware

import (
	"go-order-lite/pkg/errno"
	"go-order-lite/pkg/response"
	"net/http"

	"github.com/gin-gonic/gin"
)

func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1 先执行后面的 handler
		c.Next()

		// 2 handler 执行完，如果有错误
		if len(c.Errors) == 0 {
			return
		}

		// 3 取最后一个错误（最重要的那个）
		err := c.Errors.Last().Err

		// 4 翻译成 errno
		if e, ok := err.(*errno.Error); ok {
			c.JSON(http.StatusOK, response.Fail(e))
		} else {
			c.JSON(http.StatusOK, response.Fail(errno.InternalError))
		}
	}
}
