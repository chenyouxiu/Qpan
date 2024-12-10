package controllers

import (
	"Qpan/models"
	"Qpan/utils"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// CreateFolder 创建文件夹
func CreateFolder(c *gin.Context) {
	var req struct {
		Name     string     `json:"name" binding:"required"`
		ParentID *uuid.UUID `json:"parent_id"`
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

	// 如果指定了父文件夹，检查是否存在且属于当前用户
	if req.ParentID != nil {
		var parentFolder models.Folder
		if err := db.Where("id = ? AND user_id = ?", req.ParentID, userID).First(&parentFolder).Error; err != nil {
			c.JSON(http.StatusOK, utils.Error(404, "父文件夹不存在或无权访问"))
			return
		}
	}

	var existingFolder models.Folder
	query := db.Where("user_id = ? AND name = ?", userID, req.Name)
	if req.ParentID != nil {
		query = query.Where("parent_id = ?", req.ParentID)
	} else {
		query = query.Where("parent_id IS NULL")
	}

	if err := query.First(&existingFolder).Error; err == nil {
		c.JSON(http.StatusOK, utils.Error(409, "该目录下已存在同名文件夹"))
		return
	}

	// 创建新文件夹
	folder := models.Folder{
		Name:     req.Name,
		ParentID: req.ParentID,
		UserID:   userID,
	}

	if err := db.Create(&folder).Error; err != nil {
		c.JSON(http.StatusOK, utils.Error(500, "创建文件夹失败"))
		return
	}

	c.JSON(http.StatusOK, utils.Success(gin.H{
		"folder_id": folder.ID,
		"name":      folder.Name,
		"message":   "文件夹创建成功",
	}))
}

// ListFolders 获取文件夹列表
func ListFolders(c *gin.Context) {
	userIDStr := c.GetString("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusOK, utils.Error(500, "无效的用户ID"))
		return
	}

	parentIDStr := c.Query("parent_id")
	var parentID *uuid.UUID

	if parentIDStr != "" {
		parsed, err := uuid.Parse(parentIDStr)
		if err != nil {
			c.JSON(http.StatusOK, utils.Error(400, "无效的父文件夹ID"))
			return
		}
		parentID = &parsed
	}

	var folders []models.Folder
	query := db.Where("user_id = ?", userID)

	if parentID != nil {
		query = query.Where("parent_id = ?", parentID)
	} else {
		query = query.Where("parent_id IS NULL")
	}

	if err := query.Find(&folders).Error; err != nil {
		c.JSON(http.StatusOK, utils.Error(500, "获取文件夹列表失败"))
		return
	}

	// 获取每个文件夹下的文件和子文件夹数量
	type FolderInfo struct {
		models.Folder
		FileCount      int `json:"file_count"`
		SubFolderCount int `json:"sub_folder_count"`
	}

	var folderInfos []FolderInfo
	for _, folder := range folders {
		var fileCount, subFolderCount int64
		db.Model(&models.File{}).Where("folder_id = ?", folder.ID).Count(&fileCount)
		db.Model(&models.Folder{}).Where("parent_id = ?", folder.ID).Count(&subFolderCount)

		folderInfos = append(folderInfos, FolderInfo{
			Folder:         folder,
			FileCount:      int(fileCount),
			SubFolderCount: int(subFolderCount),
		})
	}

	c.JSON(http.StatusOK, utils.Success(gin.H{
		"folders": folderInfos,
	}))
}

// DeleteFolder 删除文件夹
func DeleteFolder(c *gin.Context) {
	userIDStr := c.GetString("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusOK, utils.Error(500, "无效的用户ID"))
		return
	}

	folderIDStr := c.Param("id")
	folderID, err := uuid.Parse(folderIDStr)
	if err != nil {
		c.JSON(http.StatusOK, utils.Error(400, "无效的文件夹ID"))
		return
	}

	var folder models.Folder
	if err := db.Where("id = ? AND user_id = ?", folderID, userID).First(&folder).Error; err != nil {
		c.JSON(http.StatusOK, utils.Error(404, "文件夹不存在或无权访问"))
		return
	}

	// 开启事务
	tx := db.Begin()

	// 删除文件夹中的文件
	if err := tx.Where("folder_id = ?", folderID).Delete(&models.File{}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusOK, utils.Error(500, "删除文件夹内文件失败"))
		return
	}

	// 删除文件夹
	if err := tx.Delete(&folder).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusOK, utils.Error(500, "删除文件夹失败"))
		return
	}

	tx.Commit()

	c.JSON(http.StatusOK, utils.Success(gin.H{
		"message": "文件夹删除成功",
	}))
}
