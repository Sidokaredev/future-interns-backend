package models

import "gorm.io/gorm"

type Document struct {
	gorm.Model
	Purpose  string `gorm:"type:nvarchar(128);not null"`
	Name     string `gorm:"type:nvarchar(256);not null"`
	MimeType string `gorm:"type:nvarchar(32);not null"`
	Size     int64  `gorm:"type:bigint;not null"`
	Blob     []byte `gorm:"type:varbinary(max);not null"`
}
