package models

import "time"

type CacheSession struct {
	ID        uint      `gorm:"type:int;primaryKey;autoIncrement"`
	Label     string    `gorm:"type:nvarchar(128);not null"`
	CreatedAt time.Time `gorm:"type:datetimeoffset(7);not null"`
}
