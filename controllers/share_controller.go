package controllers

import (
	"Qpan/models"
	"Qpan/utils"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func generateShareCode() string {
	code := make([]byte, 8)
	for i := range code {
		code[i] = charset[rand.Intn(len(charset))]
	}
	return string(code)
}

func CreateShare(c *gin.Context) {
	var req struct {
		FileID      string    `json:"file_id" binding:"required"`
		ExpireTime  time.Time `json:"expire_time"`
		Code        string    `json:"code" binding:"min=4"`
		MaxDownload int       `json:"max_download"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, utils.Error(400, "请求参数错误"))
		return
	}

	userIDStr := c.GetString("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusOK, utils.Error(500, "无效的用户ID"))
		return
	}

	fileID, err := uuid.Parse(req.FileID)
	if err != nil {
		c.JSON(http.StatusOK, utils.Error(400, "无效的文件ID"))
		return
	}

	// 检查文件是否存在且属于当前用户
	var file models.File
	if err := db.Where("id = ? AND user_id = ?", fileID, userID).First(&file).Error; err != nil {
		c.JSON(http.StatusOK, utils.Error(404, "文件不存在或无权访问"))
		return
	}
	if req.Code == "" {
		req.Code = generateShareCode()
	}
	share := models.Share{
		UserID:      userID,
		FileID:      fileID,
		Code:        req.Code,
		ExpireTime:  req.ExpireTime,
		MaxDownload: req.MaxDownload,
	}

	if err := db.Create(&share).Error; err != nil {
		c.JSON(http.StatusOK, utils.Error(500, "创建分享失败"))
		return
	}

	c.JSON(http.StatusOK, utils.Success(gin.H{
		"link":       fmt.Sprintf("http://%s/api/v1/public/share/%s/download", models.ServerHost, share.Code),
		"share_code": share.Code,
		"message":    "创建分享成功",
	}))
}

// ListShares 获取分享列表
func ListShares(c *gin.Context) {
	userIDStr := c.GetString("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusOK, utils.Error(500, "无效的用户ID"))
		return
	}
	fmt.Println(userID)
	var shares []models.Share

	if err := db.Preload("File").Where("user_id = ?", userID).Find(&shares).Error; err != nil {
		c.JSON(http.StatusOK, utils.Error(500, "获取分享列表失败"))
		return
	}

	c.JSON(http.StatusOK, utils.Success(gin.H{
		"shares": shares,
	}))
}

// GetShareInfo 获取分享信息
func GetShareInfo(c *gin.Context) {
	code := c.Param("code")
	var share models.Share

	if err := db.Preload("File").Where("code = ?", code).First(&share).Error; err != nil {
		c.JSON(http.StatusOK, utils.Error(404, "分享不存在"))
		return
	}

	// 检查是否过期
	if !share.ExpireTime.IsZero() && share.ExpireTime.Before(time.Now()) {
		c.JSON(http.StatusOK, utils.Error(400, "分享已过期"))
		return
	}

	// 检查下载次数
	if share.MaxDownload > 0 && share.DownloadNum >= share.MaxDownload {
		c.JSON(http.StatusOK, utils.Error(400, "分享已达到最大下载次数"))
		return
	}

	c.JSON(http.StatusOK, utils.Success(gin.H{
		"share": share,
	}))
}

// DownloadSharedFile 下载分享的文件
func DownloadSharedFile(c *gin.Context) {
	code := c.Param("code")
	var share models.Share
	fmt.Println(code)
	if err := db.Preload("File").Where("code = ?", code).First(&share).Error; err != nil {
		c.JSON(http.StatusOK, utils.Error(404, "分享不存在"))
		return
	}

	// 检查是否过期
	if !share.ExpireTime.IsZero() && share.ExpireTime.Before(time.Now()) {
		c.JSON(http.StatusOK, utils.Error(400, "分享已过期"))
		return
	}

	// 检查下载次数
	if share.MaxDownload > 0 && share.DownloadNum >= share.MaxDownload {
		c.JSON(http.StatusOK, utils.Error(400, "分享已达到最大下载次数"))
		return
	}

	// 更新下载次数
	share.DownloadNum++
	db.Save(&share)

	c.Header("Content-Disposition", "attachment; filename="+share.File.FileName)
	c.Header("Content-Type", "application/octet-stream")
	c.File(share.File.FilePath)
}
