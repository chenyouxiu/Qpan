package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Share struct {
	ID          uuid.UUID `json:"id" gorm:"type:char(36);primary_key"`
	UserID      uuid.UUID `json:"user_id" gorm:"type:char(36);index:idx_user_id"`
	FileID      uuid.UUID `json:"file_id" gorm:"type:char(36);index:idx_file_id"`
	Code        string    `json:"code" gorm:"uniqueIndex:idx_code;type:varchar(32)"`
	ExpireTime  time.Time `json:"expire_time" gorm:"index:idx_expire_time"`
	DownloadNum int       `json:"download_num"`
	MaxDownload int       `json:"max_download"`
	File        File      `json:"file" gorm:"foreignKey:FileID"`
	// User        User      `json:"user" gorm:"foreignKey:UserID"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (s *Share) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}
