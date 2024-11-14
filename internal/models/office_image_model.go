package models

import "time"

type OfficeImage struct {
	EmployerId string     `gorm:"primaryKey"`
	ImageId    uint       `gorm:"primaryKey"`
	CreatedAt  time.Time  `gorm:"type:datetime;not null"`
	UpdatedAt  *time.Time `gorm:"type:datetime"`
	/* Belong To */
	Image Image
}
