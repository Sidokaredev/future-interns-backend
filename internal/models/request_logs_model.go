package models

import "time"

type RequestLog struct {
	ID                  uint         `gorm:"type:int;primaryKey;autoIncrement"`
	CacheHit            int          `gorm:"type:int;not null"`
	CacheMiss           int          `gorm:"type:int;not null"`
	ResponseTime        uint64       `gorm:"type:bigint;not null"`
	MemoryUsage         float64      `gorm:"type:float;not null"`
	CPUUsage            float64      `gorm:"type:float;not null"`
	ResourceUtilization float64      `gorm:"type:float;not null"`
	CacheType           string       `gorm:"type:nvarchar(128);not null"`
	CreatedAt           time.Time    `gorm:"type:datetimeoffset(7);not null"`
	CacheSessionID      int          `gorm:"type:int;not null"`
	CacheSession        CacheSession `gorm:"foreignKey:CacheSessionID;"`
}
