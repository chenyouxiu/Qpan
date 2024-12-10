package main

import (
	"Qpan/controllers"
	"Qpan/middleware"
	"Qpan/migrations"
	"Qpan/routes"
	"Qpan/utils"
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"os"
)

func main() {
	// 配置数据库连接
	config := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	}

	cm := utils.ConfigInit()
	// 连接数据库
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", cm.Mysql.User, cm.Mysql.Password, cm.Mysql.Host, cm.Mysql.Port, cm.Mysql.Database)
	db, err := gorm.Open(mysql.Open(dsn), config)
	if err != nil {
		log.Fatal("failed to connect database: ", err)
	}

	// 设置全局数据库连接
	controllers.SetDB(db)
	middleware.SetDB(db)
	// 执行数据库迁移
	migrations.Migrate(db)

	// 初始化基础数据
	migrations.InitData(db)

	utils.InitRolesAndPermissions(db)

	// 创建上传目录
	if err := os.MkdirAll("./uploads", 0755); err != nil {
		log.Fatal("failed to create upload directory: ", err)
	}

	// 启动服务器
	r := routes.SetupRouter()
	log.Println("Server is running on :9090")
	if err := r.Run(":" + cm.Server.Port); err != nil {
		log.Fatal("failed to start server: ", err)
	}
}
