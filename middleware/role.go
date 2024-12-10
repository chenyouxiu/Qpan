// middleware/rbac.go
package middleware

import (
	"net/http"

	"Qpan/models"

	"github.com/gin-gonic/gin"
)

func RBAC(requiredPermission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从上下文中获取用户信息
		userInterface, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"code": "401", "msg": "未认证", "data": nil})
			c.Abort()
			return
		}

		user, ok := userInterface.(models.User)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"code": "401", "msg": "用户数据错误", "data": nil})
			c.Abort()
			return
		}

		// 遍历用户的角色和权限，检查是否拥有所需权限
		hasPermission := false
		for _, role := range user.Roles {
			for _, perm := range role.Permissions {
				if perm.Name == requiredPermission {
					hasPermission = true
					break
				}
			}
			if hasPermission {
				break
			}
		}

		if !hasPermission {
			c.JSON(http.StatusForbidden, gin.H{"code": "401", "msg": "权限不足", "data": nil})
			c.Abort()
			return
		}

		c.Next()
	}
}
