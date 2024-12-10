package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Folder struct {
	ID        uuid.UUID  `json:"id" gorm:"type:char(36);primary_key"`
	Name      string     `json:"name" gorm:"type:varchar(255)"`
	ParentID  *uuid.UUID `json:"parent_id" gorm:"type:char(36);index:idx_parent_id"`
	UserID    uuid.UUID  `json:"user_id" gorm:"type:char(36);index:idx_user_id"`
	Files     []File     `json:"files" gorm:"foreignKey:FolderID"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

// 添加复合唯一索引：用户ID + 父文件夹ID + 文件夹名称
func (Folder) TableName() string {
	return "folders"
}

func (Folder) Indexes() []string {
	return []string{
		"CREATE UNIQUE INDEX idx_unique_folder_name ON folders (user_id, COALESCE(parent_id, '00000000-0000-0000-0000-000000000000'), name) WHERE deleted_at IS NULL",
	}
}

// BeforeCreate 在创建记录前自动生成 UUID
func (f *Folder) BeforeCreate(tx *gorm.DB) error {
	if f.ID == uuid.Nil {
		f.ID = uuid.New()
	}
	return nil
}
