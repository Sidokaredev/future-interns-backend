package models

import "time"

type Social struct {
	ID          uint       `gorm:"type:int;primaryKey;autoIncrement"`
	Name        string     `gorm:"type:nvarchar(64);not null"`
	IconImageId int        `gorm:"type:int;not null"`
	CreatedAt   time.Time  `gorm:"type:datetime"`
	UpdatedAt   *time.Time `gorm:"type:datetime"`
	/* Belong To */
	IconImage Image `gorm:"foreignKey:IconImageId"`
}
