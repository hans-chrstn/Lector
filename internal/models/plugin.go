package models

import (
	"time"
)

type Plugin struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	Name      string    `gorm:"uniqueIndex" json:"name"`
	IsEnabled bool      `gorm:"default:true" json:"is_enabled"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}
