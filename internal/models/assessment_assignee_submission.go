package models

import "time"

type AssessmentAssigneeSubmission struct {
	AssessmentId         uint       `gorm:"type:int;primaryKey"`
	PipelineId           string     `gorm:"type:nvarchar(256);primaryKey"`
	SubmissionDocumentId uint       `gorm:"type:int;primaryKey;not null"`
	CreatedAt            time.Time  `gorm:"type:datetime;not null"`
	UpdatedAt            *time.Time `gorm:"type:datetime"`
	/* Belong To */
	Pipeline           *Pipeline
	SubmissionDocument *Document `gorm:"foreignKey:SubmissionDocumentId;references:ID"`
}
