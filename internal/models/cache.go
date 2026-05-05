package models

import (
	"time"
)

type CacheItem struct {
	Key       string `gorm:"primaryKey"`
	Value     []byte
	ExpiresAt time.Time `gorm:"index"`
}
