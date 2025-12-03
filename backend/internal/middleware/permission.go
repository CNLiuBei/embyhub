package middleware

import (
	"embyhub/internal/dao"
	"embyhub/internal/util"
	"fmt"
	"time"

	"embyhub/pkg/redis"

	"github.com/gin-gonic/gin"
)

// PermissionMiddleware 权限校验中间件
func PermissionMiddleware(requiredPermission string) gin.HandlerFunc {
	permissionDAO := dao.NewPermissionDAO()

	return func(c *gin.Context) {
		// 获取用户角色ID
		roleID, exists := c.Get("role_id")
		if !exists {
			util.ForbiddenResponse(c, "无法获取用户角色信息")
			c.Abort()
			return
		}

		roleIDInt := roleID.(int)

		// 先从缓存获取权限列表
		cacheKey := fmt.Sprintf("emby_ums:role:perms:%d", roleIDInt)
		var permissions []string

		// 尝试从Redis获取
		// 简化处理：直接从数据库查询
		permissions, err := permissionDAO.GetPermissionKeysByRoleID(roleIDInt)
		if err != nil {
			util.InternalErrorResponse(c, "获取权限信息失败")
			c.Abort()
			return
		}

		// 缓存权限列表（这里简化处理）
		redis.Set(cacheKey, "cached", 2*time.Hour)

		// 检查是否有所需权限
		hasPermission := false
		for _, perm := range permissions {
			if perm == requiredPermission {
				hasPermission = true
				break
			}
		}

		if !hasPermission {
			util.ForbiddenResponse(c, "权限不足")
			c.Abort()
			return
		}

		c.Next()
	}
}
