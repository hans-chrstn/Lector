package models

import (
	"time"

	"gorm.io/gorm"
)

type Document struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"-"`
	UpdatedAt time.Time      `json:"-"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	Title       string    `gorm:"index" json:"title"`
	URL         string    `gorm:"uniqueIndex" json:"url"`
	Source      string    `gorm:"index" json:"source"`
	CoverURL    string    `json:"cover_url"`
	Author      string    `json:"author"`
	Studio      string    `json:"studio"`
	Synopsis    string    `json:"synopsis"`
	Genres      string    `json:"genres"`
	Status      string    `json:"status"`
	IsInLibrary bool      `gorm:"default:false" json:"is_in_library"`
	IsArchived  bool      `gorm:"default:false" json:"is_archived"`
	IsLocal     bool      `gorm:"default:false" json:"is_local"`
	LocalPath   string    `json:"local_path"`
	GroupID     uint      `gorm:"index" json:"group_id"`
	Chapters    []Chapter `gorm:"foreignKey:DocumentID" json:"chapters"`

	ReadChapters int `gorm:"-" json:"read_chapters"`
}

type SearchItem struct {
	Title    string `json:"title"`
	URL      string `json:"url"`
	CoverURL string `json:"cover_url"`
	Info     string `json:"info"`
}
