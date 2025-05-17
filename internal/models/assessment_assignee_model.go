package models

import "time"

type AssessmentAssignee struct {
	AssessmentId     uint       `gorm:"type:int;primaryKey"`
	PipelineId       string     `gorm:"type:nvarchar(256);primaryKey"`
	SubmissionStatus string     `gorm:"type:nvarchar(32);not null"`
	SubmissionResult *int       `gorm:"type:int"`
	CreatedAt        time.Time  `gorm:"type:datetime;not null"`
	UpdatedAt        *time.Time `gorm:"type:datetime"`
	/* Belong To */
	Pipeline   *Pipeline
	Assessment *Assessment
}
