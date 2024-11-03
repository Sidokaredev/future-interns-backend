package models

import "gorm.io/gorm"

type Role struct {
	gorm.Model
	Name        string `gorm:"type:nvarchar(256);unique;not null"`
	Description string `gorm:"type:nvarchar(max);not null"`
	/* Many2Many */
	Permissions []*Permission `gorm:"many2many:role_permissions"`
}
