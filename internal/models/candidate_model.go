package models

import (
	"time"

	"gorm.io/gorm"
)

type Candidate struct {
	Id                       string         `gorm:"primaryKey;type:nvarchar(256)"`
	Expertise                string         `gorm:"type:nvarchar(128);not null"`
	AboutMe                  string         `gorm:"type:nvarchar(512)" json:"about_me"`
	DateOfBirth              time.Time      `gorm:"type:datetime;not null"`
	BackgroundProfileImageId *uint          `gorm:"type:bigint"`
	ProfileImageId           *uint          `gorm:"type:bigint"`
	CVDocumentId             *uint          `gorm:"type:bigint"`
	UserId                   string         `gorm:"type:nvarchar(256);unique;not null"`
	CreatedAt                time.Time      `gorm:"type:datetime;not null"`
	UpdatedAt                time.Time      `gorm:"type:datetime"`
	DeleteAt                 gorm.DeletedAt `gorm:"type:datetime"`
	/* Belong To */
	User                   *User
	Document               *Document `gorm:"foreignKey:CVDocumentId"`
	BackgroundProfileImage *Image    `gorm:"foreignKey:BackgroundProfileImageId"`
	ProfileImage           *Image    `gorm:"foreignKey:ProfileImageId"`
	/* Has Many */
	Educations       []*Education
	Experiences      []*Experience
	CandidateSocials []*CandidateSocial
	/* Many2Many */
	Socials   []*Social  `gorm:"many2many:candidate_socials"`
	Skills    []*Skill   `gorm:"many2many:candidate_skills"`
	Addresses []*Address `gorm:"many2many:candidate_addresses"`
}
