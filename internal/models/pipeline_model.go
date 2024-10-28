package models

import (
	"time"

	"gorm.io/gorm"
)

type Pipeline struct {
	Id          string         `gorm:"type:nvarchar(256);primaryKey;"`
	CandidateId string         `gorm:"type:nvarchar(256);not null"`
	VacancyId   string         `gorm:"type:nvarchar(256);not null"`
	Stage       string         `gorm:"type:nvarchar(64);not null"`
	Status      string         `gorm:"type:nvarchar(32);not null"`
	CreatedAt   time.Time      `gorm:"type:datetime;not null"`
	UpdatedAt   time.Time      `gorm:"type:datetime"`
	DeleteAt    gorm.DeletedAt `gorm:"type:datetime"`
	/* Belong To */
	Candidate Candidate
	/* Has Many */
	Interviews []Interview
	Offerings  []Offering
}
