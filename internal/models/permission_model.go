package models

import "time"

type Permission struct {
	// gorm.Model
	ID          uint       `gorm:"type:int;primaryKey;autoIncrement"`
	Name        string     `gorm:"type:nvarchar(256);unique;not null"`
	Resource    string     `gorm:"type:nvarchar(128);not null"`
	Description string     `gorm:"type:nvarchar(max);not null"`
	CreatedAt   time.Time  `gorm:"type:datetime;not null"`
	UpdatedAt   *time.Time `gorm:"type:datetime"`
}
