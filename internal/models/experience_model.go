package models

import "time"

type Experience struct {
	ID                   uint       `gorm:"type:int;primaryKey;autoIncrement"`
	CompanyName          string     `gorm:"type:nvarchar(128);not null"`
	Position             string     `gorm:"type:nvarchar(128);not null"`
	Type                 string     `gorm:"type:nvarchar(32);not null"`
	LocationAddress      string     `gorm:"type:nvarchar(256);not null"`
	IsCurrent            bool       `gorm:"type:bit;not null"`
	StartAt              time.Time  `gorm:"type:datetime;not null"`
	EndAt                time.Time  `gorm:"type:datetime"`
	Description          string     `gorm:"type:nvarchar(max)"`
	AttachmentDocumentId uint       `gorm:"type:bigint"`
	CandidateId          string     `gorm:"type:nvarchar(256);not null"`
	CreatedAt            time.Time  `gorm:"type:datetime;not null"`
	UpdatedAt            *time.Time `gorm:"type:datetime"`
	/* Belong To */
	AttachmentDocument *Document `gorm:"foreignKey:AttachmentDocumentId"`
	// CompanyId            uint      `gorm:"type:bigint"`
	// Company Soon
}
