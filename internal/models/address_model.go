package models

import "time"

type Address struct {
	// gorm.Model
	ID           uint       `gorm:"type:int;primaryKey;autoIncrement"`
	Street       string     `gorm:"type:nvarchar(128);not null"`
	Neighborhood string     `gorm:"type:nvarchar(128);not null"`
	RuralArea    string     `gorm:"type:nvarchar(128);not null"`
	SubDistrict  string     `gorm:"type:nvarchar(128);not null"`
	City         string     `gorm:"type:nvarchar(128);not null"`
	Province     string     `gorm:"type:nvarchar(128);not null"`
	Country      string     `gorm:"type:nvarchar(128);not null"`
	PostalCode   int        `gorm:"type:int;not null"`
	Type         string     `gorm:"type:nvarchar(64);not null"`
	CreatedAt    time.Time  `gorm:"type:datetime;not null"`
	UpdatedAt    *time.Time `gorm:"type:datetime"`
}
