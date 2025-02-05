package models

import "time"

type EmployerSocial struct {
	EmployerId string     `gorm:"primaryKey"`
	SocialId   uint       `gorm:"primaryKey"`
	Url        string     `gorm:"type:nvarchar(max);not null"`
	CreatedAt  time.Time  `gorm:"type:datetime;not null"`
	UpdatedAt  *time.Time `gorm:"type:datetime"`
	/* Belong To */
	Social Social
}
