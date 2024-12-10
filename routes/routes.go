package routes

import (
	"Qpan/controllers"
	"Qpan/middleware"
	"github.com/gin-gonic/gin"
	"net/http"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()
	r.Use(Cors())
	// API v1 分组
	v1 := r.Group("/api/v1")
	{
		// 无需认证的公共路由
		public := v1.Group("/public")
		{
			// 用户认证
			public.POST("/register", controllers.Register)
			public.POST("/login", controllers.Login)

			// 公共分享访问
			public.GET("/share/:code", controllers.GetShareInfo)
			public.GET("/share/:code/download", controllers.DownloadSharedFile)

			public.GET("/captcha/", controllers.GetCaptcha)
			public.POST("/captcha/verify", controllers.VerifyCaptcha)
		}

		// 需要认证的路由组
		auth := v1.Group("")
		auth.Use(middleware.JWT())
		{
			// 用户相关路由
			user := auth.Group("/user")
			{
				user.GET("/info", controllers.GetUserDetails)
				user.GET("/storage/stats", controllers.GetFileStats)
			}

			// 文件系统路由组
			fs := auth.Group("/fs")
			{
				// 文件夹操作
				fs.POST("/folder", controllers.CreateFolder)         // 创建文件夹
				fs.GET("/folder/:id", controllers.GetFolderContents) // 获取文件夹内容
				fs.DELETE("/folder/:id", controllers.DeleteFolder)   // 删除文件夹

				// 文件操作
				fs.POST("/upload", middleware.RBAC("upload_file"), controllers.UploadFile)      // 上传文件
				fs.GET("/file/:id", middleware.RBAC("download_file"), controllers.DownloadFile) // 下载文件
				fs.DELETE("/file/:id", middleware.RBAC("delete_file"), controllers.DeleteFile)  // 删除文件
				fs.POST("/file/move", controllers.MoveFile)                                     // 移动文件
				fs.POST("/file/rename", controllers.RenameFile)                                 // 重命名文件
				fs.POST("/file/batch-move", controllers.BatchMoveFiles)                         // 批量移动文件
				fs.POST("/file/batch-rename", controllers.BatchRenameFiles)                     // 批量重命名文件

				// 搜索和统计
				fs.GET("/search", controllers.SearchFiles) // 搜索文件
				fs.GET("/types", controllers.GetFileTypes) // 获取文件类型列表
				fs.GET("/stats", controllers.GetFileStats) // 获取文件统计
			}

			// 分享相关路由组
			share := auth.Group("/share")
			{
				share.POST("/create", middleware.RBAC("create_share"), controllers.CreateShare) // 创建分享
				share.GET("/list", middleware.RBAC("list_shares"), controllers.ListShares)      // 获取分享列表
			}
		}
	}

	return r
}

//func Cors() gin.HandlerFunc {
//	//return func(c *gin.Context) {
//	//	method := c.Request.Method               // 请求方法
//	//origin := c.Request.Header.Get("Origin") // 请求头部
//	//	var headerKeys []string                  // 声明请求头keys
//	//	for k, _ := range c.Request.Header {
//	//		headerKeys = append(headerKeys, k)
//	//	}
//	//	headerStr := strings.Join(headerKeys, ", ")
//	//	if headerStr != "" {
//	//		headerStr = fmt.Sprintf("access-control-allow-origin, access-control-allow-headers, %s", headerStr)
//	//	} else {
//	//		headerStr = "access-control-allow-origin, access-control-allow-headers"
//	//	}
//	//	if origin != "" {
//	//		c.Header("Access-Control-Allow-Origin", origin) // 使用请求来源
//	//		c.Header("Access-Control-Allow-Credentials", "true")
//	//		c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
//	//		c.Header("Access-Control-Allow-Headers", "Authorization, Content-Length, X-CSRF-Token, Token, Content-Type, Accept, Origin, X-Requested-With")
//	//		c.Set("content-type", "application/json") // 设置返回格式是json
//	//	}
//	//	// 放行所有OPTIONS方法
//	//	if method == "OPTIONS" {
//	//		c.JSON(http.StatusOK, "Options Request!")
//	//	}
//	//	// 处理请求
//	//	c.Next()
//	//}
//	return func(c *gin.Context) {
//		origin := c.Request.Header.Get("Origin") // 请求头部
//		method := c.Request.Method
//		if c.Request.Method == "OPTIONS" {
//			c.AbortWithStatus(http.StatusOK)
//			return
//		}
//
//		if origin != "" {
//			c.Header("Access-Control-Allow-Origin", "*")
//			c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, UPDATE")
//			c.Header("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept, Authorization")
//			c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Cache-Control, Content-Language, Content-Type")
//			c.Header("Access-Control-Allow-Credentials", "true")
//		}
//		if method == "OPTIONS" {
//			c.AbortWithStatus(http.StatusNoContent)
//		}
//		c.Next()
//	}
//}

func Cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		if origin != "" {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Credentials", "true")
			c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
			c.Header("Access-Control-Allow-Headers", "Authorization, Content-Length, Content-Type, Accept, Origin, X-Requested-With")
		}

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusOK)
			return
		}
		c.Next()
	}
}
