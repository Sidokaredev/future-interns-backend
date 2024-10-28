package models

import "time"

type Offering struct {
	Id            uint      `gorm:"type:bigint;primaryKey;autoIncrement"`
	EndOn         time.Time `gorm:"type:datetime;not null"`
	Status        string    `gorm:"type:nvarchar(16);not null"`
	PipelineId    string    `gorm:"type:nvarchar(256);not null"`
	LoADocumentId uint      `gorm:"type:bigint;not null"`
	CreatedAt     time.Time `gorm:"type:datetime;not null"`
	UpdatedAt     time.Time `gorm:"type:datetime"`
	/* Belong To */
	LoADocument Document `gorm:"foreignKey:LoADocumentId"`
}
