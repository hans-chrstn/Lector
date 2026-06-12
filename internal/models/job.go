package models

import (
	"time"
)

type Job struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	Type      string    `json:"type"`
	Status    string    `json:"status"`
	Progress  int       `json:"progress"`
	Message   string    `json:"message"`
	Payload   string    `json:"payload"`
	Error     string    `json:"error"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
