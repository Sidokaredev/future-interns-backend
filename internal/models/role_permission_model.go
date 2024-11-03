package models

import (
	"time"

	"gorm.io/gorm"
)

type RolePermission struct {
	RoleId       uint      `gorm:"primaryKey"`
	PermissionId uint      `gorm:"primaryKey"`
	CreatedAt    time.Time `gorm:"type:datetime;not null"`
	UpdatedAt    time.Time `gorm:"type:datetime"`
	DeletedAt    gorm.DeletedAt
}
