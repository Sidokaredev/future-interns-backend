package models

import (
	"time"

	"gorm.io/gorm"
)

type Assessment struct {
	Id             uint           `gorm:"type:bigint;primaryKey;autoIncrement"`
	Name           string         `gorm:"type:nvarchar(128);not null"`
	Note           string         `gorm:"type:nvarchar(max);not null"`
	AssessmentLink string         `gorm:"type:nvarchar(max)"`
	StartAt        time.Time      `gorm:"type:datetime;not null"`
	DueDate        time.Time      `gorm:"type:datetime;not null"`
	VacancyId      string         `gorm:"type:nvarchar(256);not null"`
	CreatedAt      time.Time      `gorm:"type:datetime;not null"`
	UpdatedAt      time.Time      `gorm:"type:datetime"`
	DeleteAt       gorm.DeletedAt `gorm:"type:datetime"`
	/* Has Many */
	AssessmentDocuments           []AssessmentDocument           `gorm:"foreignKey:AssessmentId"`
	AssessmentAssignees           []AssessmentAssignee           `gorm:"foreignKey:AssessmentId"`
	AssessmentAssigneeSubmissions []AssessmentAssigneeSubmission `gorm:"foreignKey:AssessmentId"`
}