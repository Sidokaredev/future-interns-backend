package models

import "time"

type Interview struct {
	Id          uint      `gorm:"type:bigint;primaryKey;autoIncrement"`
	Location    string    `gorm:"type:nvarchar(256);not null"`
	LocationURL string    `gorm:"type:nvarchar(max);not null"`
	Status      string    `gorm:"type:nvarchar(16);not null"`
	Result      string    `gorm:"type:nvarchar(32);"`
	PipelineId  string    `gorm:"type:nvarchar(256);not null"`
	CreatedAt   time.Time `gorm:"type:datetime;not null"`
	UpdatedAt   time.Time `gorm:"type:datetime"`
}
