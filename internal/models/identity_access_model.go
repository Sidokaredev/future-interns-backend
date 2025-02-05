package models

import "gorm.io/gorm"

type IdentityAccess struct {
	gorm.Model
	UserId string `gorm:"type:nvarchar(256);not null"`
	RoleId uint   `gorm:"type:bigint;not null"`
	Type   string `gorm:"type:nvarchar(64);not null"`
	/* Belong To */
	User User
	Role Role
}
