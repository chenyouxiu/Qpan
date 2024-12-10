package utils

import (
	"Qpan/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"log"
)

func InitRolesAndPermissions(db *gorm.DB) {
	// 定义权限
	permissions := []models.Permission{
		{Id: uuid.New(), Name: "upload_file"},
		{Id: uuid.New(), Name: "download_file"},
		{Id: uuid.New(), Name: "delete_file"},
		{Id: uuid.New(), Name: "manage_users"},
		{Id: uuid.New(), Name: "move_file"},
		{Id: uuid.New(), Name: "rename_file"},
		{Id: uuid.New(), Name: "batch_move_file"},
		{Id: uuid.New(), Name: "batch_rename_file"},
		{Id: uuid.New(), Name: "search_file"},
		{Id: uuid.New(), Name: "create_share"},
		{Id: uuid.New(), Name: "list_shares"},
	}

	// 创建权限
	for _, perm := range permissions {
		var existing models.Permission
		if err := db.Where("name = ?", perm.Name).First(&existing).Error; err != nil {
			if err := db.Create(&perm).Error; err != nil {
				log.Fatalf("无法创建权限 %s: %v", perm.Name, err)
			}
		}
	}

	// 定义角色及其权限
	roles := map[string][]string{
		"admin": {
			"upload_file",
			"download_file",
			"delete_file",
			"manage_users",
			"move_file",
			"rename_file",
			"batch_move_file",
			"batch_rename_file",
			"search_file",
			"create_share",
			"list_shares",
		},
		"user": {
			"upload_file",
			"download_file",
			"search_file",
			"create_share",
			"list_shares",
			"delete_file",
		},
	}

	// 创建角色并分配权限
	for roleName, permNames := range roles {
		var role models.Role
		if err := db.Where("name = ?", roleName).First(&role).Error; err != nil {
			role = models.Role{
				Id:   uuid.New(),
				Name: roleName,
			}
			if err := db.Create(&role).Error; err != nil {
				log.Fatalf("无法创建角色 %s: %v", roleName, err)
			}
		}

		for _, permName := range permNames {
			var perm models.Permission
			if err := db.Where("name = ?", permName).First(&perm).Error; err != nil {
				log.Printf("权限 %s 未找到，跳过分配", permName)
				continue
			}
			if err := db.Model(&role).Association("Permissions").Append(&perm); err != nil {
				log.Printf("无法为角色 %s 分配权限 %s: %v", roleName, permName, err)
			}
		}
	}
}
