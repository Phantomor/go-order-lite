package middleware

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	ContextRequestIDKey = "request_id"
)

func RequestContext() gin.HandlerFunc {
	return func(c *gin.Context) {

		// 1 从 Header 里取 request_id
		requestID := c.GetHeader("X-Request-Id")

		// 2 如果前端没传，就生成一个
		if requestID == "" {
			requestID = uuid.New().String()
		}

		// 3 存到 gin.Context
		c.Set(ContextRequestIDKey, requestID)

		// 4 存到 request.Context
		ctx := context.WithValue(
			c.Request.Context(),
			ContextRequestIDKey,
			requestID,
		)
		c.Request = c.Request.WithContext(ctx)

		// 5 继续执行后续 handler
		c.Next()
	}
}
