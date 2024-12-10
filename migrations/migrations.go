package migrations

import (
	"Qpan/models"
	"log"

	"gorm.io/gorm"
)

// Migrate 执行数据库迁移
func Migrate(db *gorm.DB) {
	log.Println("开始执行数据库迁移...")

	// 删除现有的外键约束
	dropForeignKeys(db)

	// 按照依赖关系顺序执行迁移
	// 1. 先创建用户表（其他表都依赖于用户表）
	if err := db.AutoMigrate(&models.User{}); err != nil {
		log.Fatal("用户表迁移失败: ", err)
	}

	// 2. 创建文件夹表（文件表依赖于文件夹表）
	if err := db.AutoMigrate(&models.Folder{}); err != nil {
		log.Fatal("文件夹表迁移失败: ", err)
	}

	// 3. 创建文件表
	if err := db.AutoMigrate(&models.File{}); err != nil {
		log.Fatal("文件表迁移失败: ", err)
	}

	// 4. 最后创建分享表（依赖于文件表）
	if err := db.AutoMigrate(&models.Share{}); err != nil {
		log.Fatal("分享表迁移失败: ", err)
	}

	if err := db.AutoMigrate(&models.Role{}, &models.Permission{}, &models.UserRole{}); err != nil {
		log.Fatal("权限表迁移失败: ", err)
	}

	// 添加新的外键约束
	addForeignKeys(db)

	log.Println("数据库迁移完成")
}

// dropForeignKeys 删除现有的外键约束
func dropForeignKeys(db *gorm.DB) {
	// 删除 files 表的外键
	db.Exec("ALTER TABLE files DROP FOREIGN KEY IF EXISTS fk_users_files")
	db.Exec("ALTER TABLE files DROP FOREIGN KEY IF EXISTS fk_folders_files")

	// 删除 folders 表的外键
	db.Exec("ALTER TABLE folders DROP FOREIGN KEY IF EXISTS fk_users_folders")
	db.Exec("ALTER TABLE folders DROP FOREIGN KEY IF EXISTS fk_folders_folders")

	// 删除 shares 表的外键
	db.Exec("ALTER TABLE shares DROP FOREIGN KEY IF EXISTS fk_users_shares")
	db.Exec("ALTER TABLE shares DROP FOREIGN KEY IF EXISTS fk_files_shares")
}

// addForeignKeys 添加新的外键约束
func addForeignKeys(db *gorm.DB) {
	// 添加 files 表的外键
	db.Exec(`ALTER TABLE files 
		ADD CONSTRAINT fk_users_files FOREIGN KEY (user_id) REFERENCES users(id),
		ADD CONSTRAINT fk_folders_files FOREIGN KEY (folder_id) REFERENCES folders(id)`)

	// 添加 folders 表的外键
	db.Exec(`ALTER TABLE folders 
		ADD CONSTRAINT fk_users_folders FOREIGN KEY (user_id) REFERENCES users(id),
		ADD CONSTRAINT fk_folders_folders FOREIGN KEY (parent_id) REFERENCES folders(id)`)

	// 添加 shares 表的外键
	db.Exec(`ALTER TABLE shares 
		ADD CONSTRAINT fk_users_shares FOREIGN KEY (user_id) REFERENCES users(id),
		ADD CONSTRAINT fk_files_shares FOREIGN KEY (file_id) REFERENCES files(id)`)

	// 添加文件夹名称唯一性约束
	db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS idx_unique_folder_name 
		ON folders (user_id, COALESCE(parent_id, '00000000-0000-0000-0000-000000000000'), name) 
		WHERE deleted_at IS NULL`)
}

func InitRoles(db *gorm.DB) {
	var adminRole models.Role
	if err := db.Where("name = ?", "admin").First(&adminRole).Error; err != nil {
		adminRole = models.Role{
			Name: "admin",
		}

		if err := db.Create(&adminRole).Error; err != nil {
			log.Fatal("创建管理员角色失败: ", err)
		}
		log.Println("已创建管理员角色")
	}

	var uploadPermission models.Permission
	if err := db.Where("name = ?", "upload").First(&uploadPermission).Error; err != nil {
		uploadPermission = models.Permission{
			Name: "upload",
		}

		if err := db.Create(&uploadPermission).Error; err != nil {
			log.Fatal("创建上传权限失败: ", err)
		}
		log.Println("已创建上传权限")
	}

	var downloadPermission models.Permission
	if err := db.Where("name = ?", "download").First(&downloadPermission).Error; err != nil {
		downloadPermission = models.Permission{
			Name: "download",
		}

		if err := db.Create(&downloadPermission).Error; err != nil {
			log.Fatal("创建下载权限失败: ", err)
		}
		log.Println("已创建下载权限")
	}

	if err := db.Model(&adminRole).Association("Permissions").Append(&uploadPermission); err != nil {
		log.Fatal("添加上传权限到管理员角色失败: ", err)
	}
	log.Println("已将上传权限添加到管理员角色")

	if err := db.Model(&adminRole).Association("Permissions").Append(&downloadPermission); err != nil {
		log.Fatal("添加下载权限到管理员角色失败: ", err)
	}
	log.Println("已将下载权限添加到管理员角色")

	var adminUser models.User
	if err := db.Where("email = ?", "admin@qpan.com").First(&adminUser).Error; err != nil {
		adminUser = models.User{
			Email:    "admin@qpan.com",
			Username: "admin",
		}
	}

	if err := db.Model(&adminUser).Association("Roles").Append(&adminRole); err != nil {
		log.Fatal("添加管理员角色到管理员用户失败: ", err)
	}
	log.Println("已将管理员角色添加到管理员用户")
}

// InitData 初始化基础数据
func InitData(db *gorm.DB) {
	log.Println("开始初始化基础数据...")

	// 检查是否已存在管理员用户
	var adminCount int64
	db.Model(&models.User{}).Where("email = ?", "admin@qpan.com").Count(&adminCount)

	if adminCount == 0 {
		// 创建管理员用户
		admin := models.User{
			Email:    "admin@qpan.com",
			Username: "admin",
			Password: "$2a$14$rnbMqMPPH.yYCgZRfHAyX.CbhGBzPlvQSKtqEW9Sc0kGxXQXhfiGe", // 密码: admin123
		}
		if err := db.Create(&admin).Error; err != nil {
			log.Fatal("创建管理员用户失败: ", err)
		}
		log.Println("已创建管理员用户")
	}
	InitRoles(db)
	log.Println("基础数据初始化完成")
}
