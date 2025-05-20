package models

import "time"

type Headquarter struct {
	EmployerId string     `gorm:"primaryKey"`
	AddressId  uint       `gorm:"primaryKey"`
	Name       string     `gorm:"type:nvarchar(256);not null"`
	Type       string     `gorm:"type:nvarchar(128);not null"`
	CreatedAt  time.Time  `gorm:"type:datetime;not null"`
	UpdatedAt  *time.Time `gorm:"type:datetime"`
	/* Belong To */
	Address Address
}
