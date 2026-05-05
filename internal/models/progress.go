package models

import (
	"time"
)

type ReadingProgress struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"updated_at"`

	DocumentID uint    `gorm:"uniqueIndex" json:"document_id"`
	ChapterID  uint    `json:"chapter_id"`
	ScrollPos  float64 `json:"scroll_pos"`

	ClientUpdatedAt int64 `json:"client_updated_at"`
}
