package models

import "time"

type AssessmentAssigneeSubmission struct {
	AssessmentId         uint      `gorm:"type:bigint;primaryKey;not null"`
	PipelineId           string    `gorm:"type:nvarchar(256);primaryKey;not null"`
	SubmissionDocumentId uint      `gorm:"type:bigint;not null"`
	CreatedAt            time.Time `gorm:"type:datetime;not null"`
	UpdatedAt            time.Time `gorm:"type:datetime"`
	/* Belong To */
	Pipeline           Pipeline
	SubmissionDocument Document `gorm:"foreignKey:SubmissionDocumentId;references:ID"`
}
