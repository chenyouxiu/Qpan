package controllers

import (
	"Qpan/models"
	"Qpan/utils"
	"github.com/dchest/captcha"
	"net/http"
	"regexp"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var db *gorm.DB

func SetDB(database *gorm.DB) {
	db = database
}

// 验证邮箱格式
func isValidEmail(email string) bool {
	pattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	regex := regexp.MustCompile(pattern)
	return regex.MatchString(email)
}

func Register(c *gin.Context) {
	var req struct {
		Email     string `json:"email" binding:"required"`
		Username  string `json:"username" binding:"required"`
		Password  string `json:"password" binding:"required,min=6"`
		CaptchaID string `json:"captcha_id" binding:"required"`
		Code      string `json:"code" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, utils.Error(400, "请求参数错误"))
		return
	}

	// 验证邮箱格式
	if !isValidEmail(req.Email) {
		c.JSON(http.StatusOK, utils.Error(400, "邮箱格式不正确"))
		return
	}
	if !captcha.VerifyString(req.CaptchaID, req.Code) {
		c.JSON(http.StatusOK, utils.Error(400, "验证码错误"))
		return
	}
	// 检查邮箱是否已存在
	var existingUser models.User
	if err := db.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusOK, utils.Error(400, "该邮箱已被注册"))
		return
	}

	// 创建新用户
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(req.Password), 14)
	user := models.User{
		Email:    req.Email,
		Username: req.Username,
		Password: string(hashedPassword),
	}

	// 开启事务
	tx := db.Begin()

	// 创建用户
	if err := tx.Create(&user).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusOK, utils.Error(500, "用户注册失败，请稍后重试"))
		return
	}

	// 创建用户的根目录
	rootFolder := models.Folder{
		Name:   "根目录",
		UserID: user.ID,
	}

	if err := tx.Create(&rootFolder).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusOK, utils.Error(500, "创建用户根目录失败"))
		return
	}

	var userRole models.Role
	if err := tx.Where("name = ?", "user").First(&userRole).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, utils.Error(500, "默认角色不存在"))
		return
	}

	// 分配 'user' 角色给新用户
	if err := tx.Model(&user).Association("Roles").Append(&userRole); err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, utils.Error(500, "分配默认角色失败"))
		return
	}

	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusOK, utils.Error(500, "用户注册失败，请稍后重试"))
		return
	}

	c.JSON(http.StatusOK, utils.Success(gin.H{
		"id":       user.ID,
		"email":    user.Email,
		"username": user.Username,
		"message":  "注册成功",
	}))
}

func Login(c *gin.Context) {
	var req struct {
		Email    string `json:"email" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, utils.Error(400, "请求参数错误"))
		return
	}

	var user models.User
	if err := db.Where("email = ?", req.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusOK, utils.Error(401, "邮箱或密码不正确"))
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		c.JSON(http.StatusOK, utils.Error(401, "邮箱或密码不正确"))
		return
	}

	// 生成JWT令牌
	token, err := utils.GenerateToken(user.ID, user.Email)
	if err != nil {
		c.JSON(http.StatusOK, utils.Error(500, "生成令牌失败"))
		return
	}

	c.JSON(http.StatusOK, utils.Success(gin.H{
		"token": token,
		"user": gin.H{
			"id":       user.ID,
			"email":    user.Email,
			"username": user.Username,
		},
		"message": "登录成功",
	}))
}

func GetUserDetails(c *gin.Context) {
	userID := c.GetString("user_id") // 从 JWT 中获取的现在是 UUID 字符串
	var user models.User
	var userRootFolder models.Folder
	if err := db.First(&user, "id = ?", userID).Error; err != nil {
		c.JSON(http.StatusOK, utils.Error(404, "未找到用户信息"))
		return
	}

	if err := db.First(&userRootFolder, "user_id = ? AND parent_id is NULL", userID).Error; err != nil {
		c.JSON(http.StatusOK, utils.Error(404, "未找到用户文件夹信息"))
		return
	}

	c.JSON(http.StatusOK, utils.Success(gin.H{
		"id":           user.ID,
		"email":        user.Email,
		"username":     user.Username,
		"rootfolderid": userRootFolder.ID,
		"created_at":   user.CreatedAt,
	}))
}
