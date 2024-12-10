//rbac.go

package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Role struct {
	Id          uuid.UUID      `gorm:"column:id;primary_key;type:char(36)"`
	Name        string         `gorm:"column:name;type:varchar(255);not null"`
	Permissions []Permission   `gorm:"many2many:role_permissions"`
	CreatedAt   time.Time      `gorm:"column:created_at;type:timestamp;not null;default:current_timestamp"`
	UpdatedAt   time.Time      `gorm:"column:updated_at;type:timestamp;not null;default:current_timestamp"`
	DeletedAt   gorm.DeletedAt `gorm:"column:deleted_at;type:timestamp"`
}

type Permission struct {
	Id        uuid.UUID `gorm:"column:id;primary_key;type:char(36)"`
	Name      string    `gorm:"column:name;type:varchar(255);not null"`
	CreatedAt time.Time `gorm:"column:created_at;type:timestamp;not null;default:current_timestamp"`
	UpdatedAt time.Time `gorm:"column:updated_at;type:timestamp;not null;default:current_timestamp"`
}

type UserRole struct {
	UserId uuid.UUID `gorm:"column:user_id;type:char(36);not null"`
	RoleId uuid.UUID `gorm:"column:role_id;type:char(36);not null"`
}

func (r *Role) BeforeCreate(tx *gorm.DB) (err error) {
	if r.Id == uuid.Nil {
		r.Id = uuid.New()
	}
	return nil
}

func (p *Permission) BeforeCreate(tx *gorm.DB) (err error) {
	if p.Id == uuid.Nil {
		p.Id = uuid.New()
	}
	return nil
}
