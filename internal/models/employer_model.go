package models

import (
	"time"

	"gorm.io/gorm"
)

type Employer struct {
	Id                       string         `gorm:"type:nvarchar(256);primaryKey"`
	Name                     string         `gorm:"type:nvarchar(128);not null"`
	LegalName                string         `gorm:"type:nvarchar(128);not null"`
	Location                 string         `gorm:"type:nvarchar(256);not null"`
	Founded                  uint           `gorm:"type:int;not null"`
	Founder                  string         `gorm:"type:nvarchar(128);not null"`
	TotalOfEmployee          string         `gorm:"type:nvarchar(64);not null"`
	Website                  string         `gorm:"type:nvarchar(256);not null"`
	Description              string         `gorm:"type:nvarchar(max)"`
	BackgroundProfileImageId *uint          `gorm:"type:bigint"`
	ProfileImageId           *uint          `gorm:"type:bigint"`
	UserId                   string         `gorm:"type:nvarchar(256);unique;not null"`
	CreatedAt                time.Time      `gorm:"type:datetime;not null"`
	UpdatedAt                *time.Time     `gorm:"type:datetime"`
	DeleteAt                 gorm.DeletedAt `gorm:"type:datetime"`
	/* Belong To */
	User                   *User
	BackgroundProfileImage *Image `gorm:"foreignKey:BackgroundProfileImageId"`
	ProfileImage           *Image `gorm:"foreignKey:ProfileImageId"`
	/* Has Many */
	Headquarters    []*Headquarter `gorm:"foreignKey:EmployerId"`
	OfficeImages    []*OfficeImage `gorm:"foreignKey:EmployerId"`
	Vacancies       []*Vacancy
	EmployerSocials []*EmployerSocial // NEW ADDED
	/* Many to Many */
	Socials []*Social `gorm:"many2many:employer_socials"`
}
