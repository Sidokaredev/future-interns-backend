package models

import "time"

type Role struct {
	// gorm.Model
	ID          uint       `gorm:"type:int;primaryKey;autoIncrement"`
	Name        string     `gorm:"type:nvarchar(256);unique;not null"`
	Description string     `gorm:"type:nvarchar(max);not null"`
	CreatedAt   time.Time  `gorm:"type:datetime;not null"`
	UpdatedAt   *time.Time `gorm:"type:datetime"`
	/* Many2Many */
	Permissions []*Permission `gorm:"many2many:role_permissions"`
}
