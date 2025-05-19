package models

import (
	"time"

	"gorm.io/gorm"
)

type Vacancy struct {
	Id              string         `gorm:"type:nvarchar(256);primaryKey"`
	Position        string         `gorm:"type:nvarchar(128);not null"`
	Description     string         `gorm:"type:nvarchar(max);not null"`
	Qualification   string         `gorm:"type:nvarchar(max);not null"`
	Responsibility  string         `gorm:"type:nvarchar(max);not null"`
	LineIndustry    string         `gorm:"type:nvarchar(128);not null"`
	EmployeeType    string         `gorm:"type:nvarchar(32);not null"`
	MinExperience   string         `gorm:"type:nvarchar(32);not null"`
	Salary          uint           `gorm:"type:bigint;not null"`
	WorkArrangement string         `gorm:"type:nvarchar(64);not null"`
	SLA             int32          `gorm:"type:int;not null"`
	IsInactive      bool           `gorm:"type:bit;not null;default:0"`
	EmployerId      string         `gorm:"type:nvarchar(256);not null"`
	CreatedAt       time.Time      `gorm:"type:datetime;not null"`
	UpdatedAt       *time.Time     `gorm:"type:datetime"`
	DeletedAt       gorm.DeletedAt `gorm:"type:datetime"`
	/* Belong To */
	Employer Employer `gorm:"foreignKey:EmployerId"`
	/* Has Many */
	Pipelines   []Pipeline
	Assessments []Assessment
	Interviews  []Interview // new
	Offerings   []Offering  // new
}
