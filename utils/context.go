package utils

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// GetUserIDFromContext 从上下文中获取用户ID
func GetUserIDFromContext(c *gin.Context) (uuid.UUID, error) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		return uuid.Nil, fmt.Errorf("user_id not found in context")
	}

	// 将 interface{} 转换为 string
	userIDString, ok := userIDStr.(string)
	if !ok {
		return uuid.Nil, fmt.Errorf("user_id is not a string")
	}

	// 解析 UUID
	userID, err := uuid.Parse(userIDString)
	if err != nil {
		return uuid.Nil, err
	}

	return userID, nil
}
