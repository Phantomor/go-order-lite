package middleware

import (
	"go-order-lite/pkg/jwt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. 取 Authorization

		authHeader := c.GetHeader("Authorization")
		// 2. 判断是否为空 / Bearer
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"msg": "missing token",
			})
			return
		}
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"msg": "invalid token format",
			})
			return
		}
		// 3. 解析 token
		claims, err := jwt.ParseToken(parts[1])
		// 4. 校验失败 → c.Abort()
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"msg": "invalid token",
			})
			return
		}
		// 5. c.Set("user_id", xxx)
		c.Set("user_id", claims.UserID)
		// 6. c.Next()
		c.Next()
	}
}
