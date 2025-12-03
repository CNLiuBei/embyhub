package middleware

import (
	"embyhub/internal/service"
	"embyhub/internal/util"
	"strings"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware JWT认证中间件
func AuthMiddleware() gin.HandlerFunc {
	authService := service.NewAuthService()

	return func(c *gin.Context) {
		// 获取Authorization头
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			util.UnauthorizedResponse(c, "未提供认证信息")
			c.Abort()
			return
		}

		// 解析Bearer Token
		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			util.UnauthorizedResponse(c, "认证格式错误")
			c.Abort()
			return
		}

		token := parts[1]

		// 验证Token
		claims, err := authService.ValidateToken(token)
		if err != nil {
			util.UnauthorizedResponse(c, "Token无效或已过期")
			c.Abort()
			return
		}

		// 将用户信息存入上下文
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role_id", claims.RoleID)

		c.Next()
	}
}
