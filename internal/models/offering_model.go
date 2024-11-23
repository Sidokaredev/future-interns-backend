package models

import "time"

type Offering struct {
	Id            uint       `gorm:"type:int;primaryKey;autoIncrement"`
	EndOn         time.Time  `gorm:"type:datetime;not null"`
	Status        string     `gorm:"type:nvarchar(16);not null"`
	PipelineId    string     `gorm:"type:nvarchar(256);not null;uniqueIndex:idx_pipeline_vacancy"`
	VacancyId     string     `gorm:"type:nvarchar(256);not null;uniqueIndex:idx_pipeline_vacancy"`
	LoaDocumentId *uint      `gorm:"type:bigint"`
	CreatedAt     time.Time  `gorm:"type:datetime;not null"`
	UpdatedAt     *time.Time `gorm:"type:datetime"`
	/* Belong To */
	LoaDocument Document `gorm:"foreignKey:LoaDocumentId"`
	Pipeline    Pipeline
}
