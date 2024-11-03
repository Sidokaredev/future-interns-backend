package models

import "gorm.io/gorm"

type Permission struct {
	gorm.Model
	Name        string `gorm:"type:nvarchar(256);unique;not null"`
	Action      string `gorm:"type:nvarchar(128);not null"`
	Description string `gorm:"type:nvarchar(max);not null"`
}
