package models

import (
	"time"

	"gorm.io/gorm"
)

type Chapter struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"-"`
	UpdatedAt time.Time      `json:"-"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	DocumentID uint   `gorm:"index:idx_doc_url,unique" json:"document_id"`
	Title      string `json:"title"`
	URL        string `gorm:"index:idx_doc_url,unique" json:"url"`
	Content    string `json:"content"`
	Order      int    `gorm:"type:integer;index" json:"order"`
	Status     string `json:"status"`
	IsRead     bool   `gorm:"default:false" json:"is_read"`
}
