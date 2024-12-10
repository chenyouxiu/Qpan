package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type File struct {
	ID        uuid.UUID  `json:"id" gorm:"type:char(36);primary_key"`
	UserID    uuid.UUID  `json:"user_id" gorm:"type:char(36);index:idx_user_id"`
	FolderID  *uuid.UUID `json:"folder_id" gorm:"type:char(36);index:idx_folder_id"`
	FileName  string     `json:"file_name" gorm:"index:idx_file_name;type:varchar(255)"`
	FilePath  string     `json:"file_path"`
	Size      int64      `json:"size"`
	Hash      string     `json:"hash" gorm:"type:varchar(64)"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

// BeforeCreate 在创建记录前自动生成 UUID
func (f *File) BeforeCreate(tx *gorm.DB) error {
	if f.ID == uuid.Nil {
		f.ID = uuid.New()
	}
	return nil
}
