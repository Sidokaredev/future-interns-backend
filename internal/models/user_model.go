package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	Id        string         `gorm:"primaryKey;type:nvarchar(256)"`
	Fullname  string         `gorm:"type:nvarchar(128);not null"`
	Email     string         `gorm:"type:nvarchar(64);not null;uniqueIndex"`
	Password  string         `gorm:"type:nvarchar(128);not null"`
	CreatedAt time.Time      `gorm:"type:datetime;not null"`
	UpdatedAt time.Time      `gorm:"type:datetime"`
	DeleteAt  gorm.DeletedAt `gorm:"type:datetime"`
}
