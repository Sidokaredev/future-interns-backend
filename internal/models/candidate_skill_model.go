package models

import "time"

type CandidateSkill struct {
	CandidateId string     `gorm:"primaryKey"`
	SkillId     uint       `gorm:"primaryKey"`
	CreatedAt   time.Time  `gorm:"type:datetime;not null"`
	UpdatedAt   *time.Time `gorm:"type:datetime"`
}
