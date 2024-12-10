// //middleware/jwt.go
package middleware

import (
	"Qpan/models"
	"Qpan/utils"
	"gorm.io/gorm"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

var db *gorm.DB

func SetDB(database *gorm.DB) {
	db = database
}
func JWT() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token == "" {
			c.JSON(http.StatusUnauthorized, utils.Error(401, "请先登录"))
			c.Abort()
			return
		}

		// 去除 Bearer 前缀
		token = strings.TrimPrefix(token, "Bearer ")

		claims, err := utils.ParseToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, utils.Error(401, "无效的令牌"))
			c.Abort()
			return
		}

		// 获取用户信息
		var user models.User
		if err := db.Preload("Roles.Permissions").Where("id = ?", claims.UserID).First(&user).Error; err != nil {
			c.JSON(http.StatusUnauthorized, utils.Error(401, "用户不存在"))
			c.Abort()
			return
		}

		// 将用户信息存储在上下文中
		c.Set("user", user)
		c.Set("user_id", user.ID.String())
		c.Set("email", claims.Email)
		c.Next()
	}
}
