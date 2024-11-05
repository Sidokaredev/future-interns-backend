package models

import "time"

type Document struct {
	// gorm.Model
	ID        uint       `gorm:"type:int;primaryKey;autoIncrement"`
	Purpose   string     `gorm:"type:nvarchar(128);not null"`
	Name      string     `gorm:"type:nvarchar(256);not null"`
	MimeType  string     `gorm:"type:nvarchar(32);not null"`
	Size      int64      `gorm:"type:bigint;not null"`
	Blob      []byte     `gorm:"type:varbinary(max);not null"`
	CreatedAt time.Time  `gorm:"type:datetime;not null"`
	UpdatedAt *time.Time `gorm:"type:datetime"`
}
