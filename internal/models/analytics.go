package models

type ReadingStat struct {
	ID            uint   `gorm:"primarykey" json:"id"`
	Date          string `gorm:"index;unique" json:"date"`
	ReadSeconds   int    `json:"read_seconds"`
	DocumentsRead int    `json:"documents_read"`
	ChaptersRead  int    `json:"chapters_read"`
}
