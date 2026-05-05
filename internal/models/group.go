package models

import (
	"time"
)

type Group struct {
	ID        uint       `gorm:"primarykey" json:"id"`
	CreatedAt time.Time  `json:"-"`
	UpdatedAt time.Time  `json:"-"`
	Name      string     `gorm:"uniqueIndex" json:"name"`
	Documents []Document `gorm:"foreignKey:GroupID" json:"documents"`
}
