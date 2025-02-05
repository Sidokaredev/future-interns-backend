package models

import "time"

type Skill struct {
	ID               uint       `gorm:"primaryKey;type:int;autoIncrement"`
	Name             string     `gorm:"type:nvarchar(256);not null"`
	SkillIconImageId uint       `gorm:"type:bigint;not null"`
	CreatedAt        time.Time  `gorm:"type:datetime;not null"`
	UpdatedAt        *time.Time `gorm:"type:datetime"`
	/* Belong To */
	SkillIconImage Image `gorm:"foreignKey:SkillIconImageId"`
}
