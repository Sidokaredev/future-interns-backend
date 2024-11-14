package models

import "time"

type AssessmentAssignee struct {
	AssessmentId     uint       `gorm:"primaryKey"`
	PipelineId       string     `gorm:"primaryKey"`
	SubmissionStatus string     `gorm:"type:nvarchar(32);not null"`
	SubmissionResult int        `gorm:"type:int"`
	CreatedAt        time.Time  `gorm:"type:datetime;not null"`
	UpdatedAt        *time.Time `gorm:"type:datetime"`
	/* Belong To */
	Pipeline Pipeline
}
