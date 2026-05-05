package models

import (
	"time"
)

type Bookmark struct {
	ID         uint      `gorm:"primarykey" json:"id"`
	DocumentID uint      `gorm:"index" json:"document_id"`
	ChapterID  uint      `json:"chapter_id"`
	Title      string    `json:"title"`
	CreatedAt  time.Time `json:"created_at"`
}
