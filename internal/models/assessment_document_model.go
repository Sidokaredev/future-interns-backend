package models

import "time"

type AssessmentDocument struct {
	AssessmentId uint       `gorm:"type:int;primaryKey"`
	DocumentId   uint       `gorm:"type:int;primaryKey"`
	CreatedAt    time.Time  `gorm:"type:datetime;not null"`
	UpdatedAt    *time.Time `gorm:"type:datetime"`
	/* Belong To */
	// Document Document
}
