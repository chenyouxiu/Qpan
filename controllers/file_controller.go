package controllers

import (
	"Qpan/models"
	"Qpan/utils"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UploadFile 上传文件
func UploadFile(c *gin.Context) {
	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusOK, utils.Error(500, "无效的用户ID"))
		return
	}

	// 获取文件夹ID
	folderIDStr := c.PostForm("folder_id")
	var folderID *uuid.UUID

	if folderIDStr != "" {
		parsed, err := uuid.Parse(folderIDStr)
		if err != nil {
			c.JSON(http.StatusOK, utils.Error(400, "无效的文件夹ID"))
			return
		}
		folderID = &parsed

		// 验证文件夹是否存在且属于当前用户
		var folder models.Folder
		if err := db.Where("id = ? AND user_id = ?", folderID, userID).First(&folder).Error; err != nil {
			c.JSON(http.StatusOK, utils.Error(404, "文件夹不存在或无权访问"))
			return
		}
	} else {
		// 如果没有指定文件夹ID，获取用户的根目录
		var rootFolder models.Folder
		if err := db.Where("user_id = ? AND parent_id IS NULL", userID).First(&rootFolder).Error; err != nil {
			c.JSON(http.StatusOK, utils.Error(500, "未找到用户根目录"))
			return
		}
		folderID = &rootFolder.ID
	}

	formFile, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusOK, utils.Error(400, "请选择要上传的文件"))
		return
	}

	var MyFile models.File
	if err := db.Where("user_id = ? AND file_name = ? AND folder_id = ?", userID, formFile.Filename, folderID).First(&MyFile).Error; err == nil {
		// 文件已存在
		c.JSON(http.StatusOK, utils.Error(400, "当前文件夹下已存在同名文件"))
		return
	}

	fileHash := sha256.New()
	fileContent, err := formFile.Open()
	if err != nil {
		c.JSON(http.StatusOK, utils.Error(500, "文件读取失败"))
		return
	}
	defer fileContent.Close()

	if _, err := io.Copy(fileHash, fileContent); err != nil {
		c.JSON(http.StatusOK, utils.Error(500, "文件读取失败"))
		return
	}
	hashStr := hex.EncodeToString(fileHash.Sum(nil))

	// 开启事务
	tx := db.Begin()

	// 检查文件是否已存在
	var existingFile models.File
	if err := tx.Where("hash = ?", hashStr).First(&existingFile).Error; err == nil {
		// 文件已存在，直接创建新的引用记录
		fmt.Println("hashStr", hashStr)
		newFile := models.File{
			UserID:   userID,
			FolderID: folderID,
			FileName: formFile.Filename,
			FilePath: existingFile.FilePath, // 使用已存在文件的路径
			Size:     existingFile.Size,
			Hash:     hashStr,
		}

		// 使用事务创建新记录
		if err := tx.Create(&newFile).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusOK, utils.Error(500, "文件信息保存失败"))
			return
		}

		tx.Commit()
		c.JSON(http.StatusOK, utils.Success(gin.H{
			"file_id":   newFile.ID,
			"file_name": newFile.FileName,
			"size":      newFile.Size,
			"message":   "文件上传成功",
			"reused":    true,
		}))
		return
	}

	// 文件不存在，保存新文件
	// 使用 UUID 字符串和时间戳创建唯一文件名
	ext := filepath.Ext(formFile.Filename)
	newFileName := userID.String() + "_" + strconv.FormatInt(time.Now().Unix(), 10) + ext
	dst := filepath.Join("uploads", newFileName)

	// 重新打开文件用于保存
	fileContent.Seek(0, 0)
	dstFile, err := os.Create(dst)
	if err != nil {
		tx.Rollback()
		c.JSON(http.StatusOK, utils.Error(500, "创建文件失败"))
		return
	}
	defer dstFile.Close()

	// 复制文件内容
	if _, err := io.Copy(dstFile, fileContent); err != nil {
		tx.Rollback()
		os.Remove(dst)
		c.JSON(http.StatusOK, utils.Error(500, "保存文件失败"))
		return
	}

	// 创建新文件记录
	newFile := models.File{
		UserID:   userID,
		FolderID: folderID,
		FileName: formFile.Filename,
		FilePath: dst,
		Size:     formFile.Size,
		Hash:     hashStr,
	}

	if err := tx.Create(&newFile).Error; err != nil {
		tx.Rollback()
		os.Remove(dst)
		c.JSON(http.StatusOK, utils.Error(500, "文件信息保存失败"))
		return
	}

	tx.Commit()
	c.JSON(http.StatusOK, utils.Success(gin.H{
		"file_id":   newFile.ID,
		"file_name": newFile.FileName,
		"size":      newFile.Size,
		"message":   "文件上传成功",
		"reused":    false,
	}))
}

// ListFiles 获取文件列表
func ListFiles(c *gin.Context) {
	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusOK, utils.Error(500, "无效的用户ID"))
		return
	}

	var files []models.File

	if err := db.Where("user_id = ?", userID).Find(&files).Error; err != nil {
		c.JSON(http.StatusOK, utils.Error(500, "获取文件列表失败"))
		return
	}

	c.JSON(http.StatusOK, utils.Success(gin.H{
		"files": files,
	}))
}

// DownloadFile 下载文件
func DownloadFile(c *gin.Context) {
	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusOK, utils.Error(500, "无效的用户ID"))
		return
	}

	fileID := c.Param("id")

	var file models.File
	if err := db.Where("id = ? AND user_id = ?", fileID, userID).First(&file).Error; err != nil {
		c.JSON(http.StatusOK, utils.Error(404, "文件不存在或无权访问"))
		return
	}

	// 检查文件是否存在
	if _, err := os.Stat(file.FilePath); os.IsNotExist(err) {
		c.JSON(http.StatusOK, utils.Error(404, "文件不存在"))
		return
	}

	// 设置下载文件名
	c.Header("Content-Disposition", "attachment; filename="+file.FileName)
	c.Header("Content-Type", "application/octet-stream")
	c.File(file.FilePath)
}

// DeleteFile 删除文件
func DeleteFile(c *gin.Context) {
	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusOK, utils.Error(500, "无效的用户ID"))
		return
	}

	fileID := c.Param("id")

	var file models.File
	if err := db.Where("id = ? AND user_id = ?", fileID, userID).First(&file).Error; err != nil {
		c.JSON(http.StatusOK, utils.Error(404, "文件不存在或无权访问"))
		return
	}

	// 删除物理文件
	//if err := os.Remove(file.FilePath); err != nil {
	//	c.JSON(http.StatusOK, utils.Error(500, "文件删除失败"))
	//	return
	//}

	// 删除数据库记录
	if err := db.Delete(&file).Error; err != nil {
		c.JSON(http.StatusOK, utils.Error(500, "文件信息删除失败"))
		return
	}

	c.JSON(http.StatusOK, utils.Success(gin.H{
		"message": "文件删除成功",
	}))
}

// SearchFiles 搜索文件
func SearchFiles(c *gin.Context) {
	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusOK, utils.Error(500, "无效的用户ID"))
		return
	}

	keyword := c.Query("keyword")
	fileType := c.Query("type") // 可选参数：按文件类型筛选

	if keyword == "" {
		c.JSON(http.StatusOK, utils.Error(400, "搜索关键词不能为空"))
		return
	}

	query := db.Where("user_id = ?", userID).
		Where("file_name LIKE ?", "%"+keyword+"%")

	// 如果指定了文件类型，添加文件类型过滤
	if fileType != "" {
		query = query.Where("file_name LIKE ?", "%."+fileType)
	}

	var files []models.File
	if err := query.Find(&files).Error; err != nil {
		c.JSON(http.StatusOK, utils.Error(500, "搜索文件失败"))
		return
	}

	// 获取文件所在文件夹信息
	type FileInfo struct {
		models.File
		FolderName string `json:"folder_name"`
	}

	var fileInfos []FileInfo
	for _, file := range files {
		var folderName string
		if file.FolderID != nil {
			var folder models.Folder
			if err := db.First(&folder, file.FolderID).Error; err == nil {
				folderName = folder.Name
			}
		}

		fileInfos = append(fileInfos, FileInfo{
			File:       file,
			FolderName: folderName,
		})
	}

	c.JSON(http.StatusOK, utils.Success(gin.H{
		"files": fileInfos,
		"total": len(fileInfos),
	}))
}

// GetFileTypes 获取所有文件类型
func GetFileTypes(c *gin.Context) {
	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusOK, utils.Error(500, "无效的用户ID"))
		return
	}

	var fileTypes []string
	err = db.Model(&models.File{}).
		Where("user_id = ?", userID).
		Select("DISTINCT SUBSTRING_INDEX(file_name, '.', -1) as type").
		Pluck("type", &fileTypes).Error

	if err != nil {
		c.JSON(http.StatusOK, utils.Error(500, "获取文件类型失败"))
		return
	}

	c.JSON(http.StatusOK, utils.Success(gin.H{
		"types": fileTypes,
	}))
}

// GetFileStats 获取文件统计信息
func GetFileStats(c *gin.Context) {
	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusOK, utils.Error(500, "无效的用户ID"))
		return
	}

	// 获取总文件数和总大小
	var totalCount int64
	var totalSize int64
	db.Model(&models.File{}).Where("user_id = ?", userID).Count(&totalCount)
	db.Model(&models.File{}).Where("user_id = ?", userID).Select("COALESCE(SUM(size), 0)").Scan(&totalSize)

	// 获取各类型文件数量
	type TypeStat struct {
		Type  string `json:"type"`
		Count int64  `json:"count"`
		Size  int64  `json:"size"`
	}

	var typeStats []TypeStat
	db.Model(&models.File{}).
		Where("user_id = ?", userID).
		Select("SUBSTRING_INDEX(file_name, '.', -1) as type, COUNT(*) as count, SUM(size) as size").
		Group("type").
		Scan(&typeStats)

	c.JSON(http.StatusOK, utils.Success(gin.H{
		"total_files": totalCount,
		"total_size":  totalSize,
		"type_stats":  typeStats,
	}))
}

// MoveFile 移动文件到指定文件夹
func MoveFile(c *gin.Context) {
	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusOK, utils.Error(500, "无效的用户ID"))
		return
	}

	var req struct {
		FileID   string  `json:"file_id" binding:"required"`
		FolderID *string `json:"folder_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, utils.Error(400, "请求参数错误"))
		return
	}

	// 检查文件是否存在且属于当前用户
	var file models.File
	if err := db.Where("id = ? AND user_id = ?", req.FileID, userID).First(&file).Error; err != nil {
		c.JSON(http.StatusOK, utils.Error(404, "文件不存在或无权访问"))
		return
	}

	// 如果指定了目标文件夹，检查文件夹是否存在且属于当前用户
	if req.FolderID != nil {
		var folder models.Folder
		if err := db.Where("id = ? AND user_id = ?", req.FolderID, userID).First(&folder).Error; err != nil {
			c.JSON(http.StatusOK, utils.Error(404, "目标文件夹不存在或无权访问"))
			return
		}
	}

	// 检查目标文件夹中是否已存在同名文件
	var existingFile models.File
	if err := db.Where("folder_id = ? AND file_name = ? AND id != ?",
		req.FolderID, file.FileName, file.ID).First(&existingFile).Error; err == nil {
		c.JSON(http.StatusOK, utils.Error(409, "目标文件夹中已存在同名文件"))
		return
	}

	// 更新文件的文件夹ID
	if err := db.Model(&file).Update("folder_id", req.FolderID).Error; err != nil {
		c.JSON(http.StatusOK, utils.Error(500, "移动文件失败"))
		return
	}

	c.JSON(http.StatusOK, utils.Success(gin.H{
		"message": "文件移动成功",
	}))
}

// BatchMoveFiles 批量移动文件
func BatchMoveFiles(c *gin.Context) {
	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusOK, utils.Error(500, "无效的用户ID"))
		return
	}

	var req struct {
		FileIDs  []uint `json:"file_ids" binding:"required"`
		FolderID *uint  `json:"folder_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, utils.Error(400, "请求参数错误"))
		return
	}

	// 检查目标文件夹是否存在且属于当前用户
	if req.FolderID != nil {
		var folder models.Folder
		if err := db.Where("id = ? AND user_id = ?", req.FolderID, userID).First(&folder).Error; err != nil {
			c.JSON(http.StatusOK, utils.Error(404, "目标文件夹不存在或无权访问"))
			return
		}
	}

	// 开启事务
	tx := db.Begin()

	// 检查并移动每个文件
	for _, fileID := range req.FileIDs {
		var file models.File
		if err := tx.Where("id = ? AND user_id = ?", fileID, userID).First(&file).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusOK, utils.Error(404, "部分文件不存在或无权访问"))
			return
		}

		// 检查目标文件夹中是否已存在同名文件
		var existingFile models.File
		if err := tx.Where("folder_id = ? AND file_name = ? AND id != ?",
			req.FolderID, file.FileName, file.ID).First(&existingFile).Error; err == nil {
			tx.Rollback()
			c.JSON(http.StatusOK, utils.Error(409, "目标文件夹中存在同名文件: "+file.FileName))
			return
		}

		// 更新文件的文件夹ID
		if err := tx.Model(&file).Update("folder_id", req.FolderID).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusOK, utils.Error(500, "移动文件失败"))
			return
		}
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusOK, utils.Error(500, "移动文件失败"))
		return
	}

	c.JSON(http.StatusOK, utils.Success(gin.H{
		"message": "文件批量移动成功",
	}))
}

// RenameFile 重命名文件
func RenameFile(c *gin.Context) {
	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusOK, utils.Error(500, "无效的用户ID"))
		return
	}

	var req struct {
		FileID      string `json:"file_id" binding:"required"`
		NewFileName string `json:"new_file_name" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, utils.Error(400, "请求参数错误"))
		return
	}

	// 检查文件是否存在且属于当前用户
	var file models.File
	if err := db.Where("id = ? AND user_id = ?", req.FileID, userID).First(&file).Error; err != nil {
		c.JSON(http.StatusOK, utils.Error(404, "文件不存在或无权访问"))
		return
	}

	// 检查新文件名是否合法
	if req.NewFileName == "" {
		c.JSON(http.StatusOK, utils.Error(400, "文件名不能为空"))
		return
	}

	// 检查同一文件夹下是否存在同名文件
	var existingFile models.File
	if err := db.Where("folder_id = ? AND file_name = ? AND id != ?",
		file.FolderID, req.NewFileName, file.ID).First(&existingFile).Error; err == nil {
		c.JSON(http.StatusOK, utils.Error(409, "该文件夹下已存在同名文件"))
		return
	}

	// 获取原文件的扩展名
	oldExt := filepath.Ext(file.FileName)
	newExt := filepath.Ext(req.NewFileName)

	// 如果新文件名没有包含扩展名，则使用原文件的扩展名
	if newExt == "" {
		req.NewFileName = req.NewFileName + oldExt
	}

	// // 生成新的文件路径
	// newFilePath := filepath.Join(filepath.Dir(file.FilePath), req.NewFileName)

	// // 重命名物理文件
	// if err := os.Rename(file.FilePath, newFilePath); err != nil {
	// 	c.JSON(http.StatusOK, utils.Error(500, "文件重命名失败"))
	// 	return
	// }

	// 更新数据库中的文件信息
	updates := map[string]interface{}{
		"file_name": req.NewFileName,
		// "file_path": newFilePath,
	}

	if err := db.Model(&file).Updates(updates).Error; err != nil {
		// 如果数据库更新失败，尝试恢复文件名
		// os.Rename(newFilePath, file.FilePath)
		c.JSON(http.StatusOK, utils.Error(500, "文件重命名失败"))
		return
	}

	c.JSON(http.StatusOK, utils.Success(gin.H{
		"file_id":       file.ID,
		"new_file_name": req.NewFileName,
		"message":       "文件重命名成功",
	}))
}

// BatchRenameFiles 批量重命名文件
func BatchRenameFiles(c *gin.Context) {
	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusOK, utils.Error(500, "无效的用户ID"))
		return
	}

	var req struct {
		Files []struct {
			FileID      uint   `json:"file_id" binding:"required"`
			NewFileName string `json:"new_file_name" binding:"required"`
		} `json:"files" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, utils.Error(400, "请求参数错误"))
		return
	}

	// 开启事务
	tx := db.Begin()

	for _, fileReq := range req.Files {
		var file models.File
		if err := tx.Where("id = ? AND user_id = ?", fileReq.FileID, userID).First(&file).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusOK, utils.Error(404, "部分文件不存在或无权访问"))
			return
		}

		// 检查新文件名是否合法
		if fileReq.NewFileName == "" {
			tx.Rollback()
			c.JSON(http.StatusOK, utils.Error(400, "文件名不能为空"))
			return
		}

		// 检查同一文件夹下是否存在同名文件
		var existingFile models.File
		if err := tx.Where("folder_id = ? AND file_name = ? AND id != ?",
			file.FolderID, fileReq.NewFileName, file.ID).First(&existingFile).Error; err == nil {
			tx.Rollback()
			c.JSON(http.StatusOK, utils.Error(409, "文件夹中已存在同名文件: "+fileReq.NewFileName))
			return
		}

		// 获取原文件的扩展名
		oldExt := filepath.Ext(file.FileName)
		newExt := filepath.Ext(fileReq.NewFileName)

		// 如果新文件名没有包含扩展名，则使用原文件的扩展名
		if newExt == "" {
			fileReq.NewFileName = fileReq.NewFileName + oldExt
		}

		// 生成新的文件路径
		newFilePath := filepath.Join(filepath.Dir(file.FilePath), fileReq.NewFileName)

		// 重命名物理文件
		if err := os.Rename(file.FilePath, newFilePath); err != nil {
			tx.Rollback()
			c.JSON(http.StatusOK, utils.Error(500, "文件重命名失败"))
			return
		}

		// 更新数据库中的文件信息
		updates := map[string]interface{}{
			"file_name": fileReq.NewFileName,
			"file_path": newFilePath,
		}

		if err := tx.Model(&file).Updates(updates).Error; err != nil {
			// 如果数据库更新失败，尝试恢复文件名
			os.Rename(newFilePath, file.FilePath)
			tx.Rollback()
			c.JSON(http.StatusOK, utils.Error(500, "文件重命名失败"))
			return
		}
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusOK, utils.Error(500, "文件重命名失败"))
		return
	}

	c.JSON(http.StatusOK, utils.Success(gin.H{
		"message": "文件批量重命名成功",
	}))
}

// 添加一个新的结构体来表示文件夹路径
type BreadcrumbItem struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

// 获取文件夹的完整路径
func getFolderPath(db *gorm.DB, folderID uuid.UUID) ([]BreadcrumbItem, error) {
	var path []BreadcrumbItem
	currentID := folderID

	for {
		var folder models.Folder
		if err := db.First(&folder, currentID).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				break
			}
			return nil, err
		}

		// 将当前文件夹添加到路径开头
		path = append([]BreadcrumbItem{{ID: folder.ID, Name: folder.Name}}, path...)

		// 如果没有父文件夹，则结束
		if folder.ParentID == nil {
			break
		}
		currentID = *folder.ParentID
	}

	return path, nil
}

// GetFolderContents 获取指定文件夹内容
func GetFolderContents(c *gin.Context) {
	userID, err := utils.GetUserIDFromContext(c)
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

	// 验证文件夹所有权
	var currentFolder models.Folder
	if err := db.Where("id = ? AND user_id = ?", folderID, userID).First(&currentFolder).Error; err != nil {
		c.JSON(http.StatusOK, utils.Error(404, "文件夹不存在或无权访问"))
		return
	}

	// 获取面包屑导航路径
	breadcrumbs, err := getFolderPath(db, folderID)
	if err != nil {
		c.JSON(http.StatusOK, utils.Error(500, "获取文件夹路径失败"))
		return
	}

	// 获取子文件夹
	var folders []models.Folder
	if err := db.Where("parent_id = ? AND user_id = ?", folderID, userID).Find(&folders).Error; err != nil {
		c.JSON(http.StatusOK, utils.Error(500, "获取子文件夹失败"))
		return
	}

	// 获取当前文件夹中的文件
	var files []models.File
	if err := db.Where("folder_id = ? AND user_id = ?", folderID, userID).Find(&files).Error; err != nil {
		c.JSON(http.StatusOK, utils.Error(500, "获取文件列表失败"))
		return
	}

	// 获取文件夹详细信息
	type FolderInfo struct {
		models.Folder
		FileCount      int       `json:"file_count"`
		SubFolderCount int       `json:"sub_folder_count"`
		UpdatedAt      time.Time `json:"updated_at"`
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
			UpdatedAt:      folder.UpdatedAt,
		})
	}

	// 按更新时间排序
	sort.Slice(folderInfos, func(i, j int) bool {
		return folderInfos[i].UpdatedAt.After(folderInfos[j].UpdatedAt)
	})
	sort.Slice(files, func(i, j int) bool {
		return files[i].UpdatedAt.After(files[j].UpdatedAt)
	})

	c.JSON(http.StatusOK, utils.Success(gin.H{
		"current_folder": currentFolder,
		"breadcrumbs":    breadcrumbs,
		"folders":        folderInfos,
		"files":          files,
		"total": gin.H{
			"folders": len(folderInfos),
			"files":   len(files),
		},
	}))
}
