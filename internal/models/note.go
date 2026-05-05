package models

import (
	"time"
)

type Note struct {
	ID         uint      `gorm:"primarykey" json:"id"`
	DocumentID uint      `gorm:"index" json:"document_id"`
	ChapterID  uint      `json:"chapter_id"`
	Content    string    `json:"content"`
	Quote      string    `json:"quote"`
	CreatedAt  time.Time `json:"created_at"`
}
