package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID        uuid.UUID `json:"id" gorm:"type:char(36);primary_key"`
	Email     string    `json:"email" gorm:"uniqueIndex:idx_email;type:varchar(100)"`
	Username  string    `json:"username" gorm:"type:varchar(100)"`
	Password  string    `json:"password"`
	Files     []File    `json:"files"`
	Roles     []Role    `gorm:"many2many:user_roles;"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}
