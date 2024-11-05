package models

import "time"

type Image struct {
	// gorm.Model
	ID        uint       `gorm:"type:int;primaryKey;autoIncrement"`
	Name      string     `gorm:"type:nvarchar(128);not null"`
	MimeType  string     `gorm:"type:nvarchar(16);not null"`
	Size      int64      `gorm:"type:int;not null"`
	Blob      []byte     `gorm:"type:varbinary(max);not null"`
	CreatedAt time.Time  `gorm:"type:datetime;not null"`
	UpdatedAt *time.Time `gorm:"type:datetime;not null"`
}
