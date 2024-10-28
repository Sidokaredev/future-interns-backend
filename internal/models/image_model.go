package models

import "gorm.io/gorm"

type Image struct {
	gorm.Model
	Name     string `gorm:"type:nvarchar(128);not null"`
	MimeType string `gorm:"type:nvarchar(16);not null"`
	Size     int64  `gorm:"type:int;not null"`
	Blob     []byte `gorm:"type:varbinary(max);not null"`
}
