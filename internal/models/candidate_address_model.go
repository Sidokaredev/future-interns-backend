package models

import "time"

type CandidateAddress struct {
	CandidateId string    `gorm:"primaryKey"`
	AddressId   uint      `gorm:"primaryKey"`
	CreatedAt   time.Time `gorm:"type:datetime;not null"`
	UpdatedAt   time.Time `gorm:"type:datetime"`
}
