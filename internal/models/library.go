package models

import (
	"time"
)

type LibraryPath struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	Path      string    `gorm:"uniqueIndex" json:"path"`
	Pattern   string    `json:"pattern"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}
