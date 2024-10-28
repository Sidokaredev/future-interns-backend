package models

import "time"

type Education struct {
	Id          uint      `gorm:"type:bigint;primaryKey;autoIncrement"`
	University  string    `gorm:"type:nvarchar(256);not null"`
	Address     string    `gorm:"type:nvarchar(256);not null"`
	Major       string    `gorm:"type:nvarchar(128);not null"`
	IsGraduated bool      `gorm:"type:bit;not null"`
	StartAt     time.Time `gorm:"type:datetime;not null"`
	EndAt       time.Time `gorm:"type:datetime"`
	GPA         float32   `gorm:"type:decimal(4, 2);not null"`
	CandidateId string    `gorm:"type:nvarchar(256);not null"`
	CreatedAt   time.Time `gorm:"type:datetime;not null"`
	UpdatedAt   time.Time `gorm:"type:datetime"`
	/* Belong To */
	// UniversityId uint      `gorm:"type:bigint"`
	// University
}
