package models

import (
	"time"
)

type RolePermission struct {
	RoleId       uint       `gorm:"primaryKey"`
	PermissionId uint       `gorm:"primaryKey"`
	CreatedAt    time.Time  `gorm:"type:datetime;not null"`
	UpdatedAt    *time.Time `gorm:"type:datetime"`
}
