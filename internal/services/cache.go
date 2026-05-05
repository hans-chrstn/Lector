package services

import (
	"encoding/json"
	"github.com/user/lector/internal/db"
	"github.com/user/lector/internal/models"
	"gorm.io/gorm"
	"time"
)

func SetCache(key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	item := models.CacheItem{
		Key:       key,
		Value:     data,
		ExpiresAt: time.Now().Add(ttl),
	}

	return db.DB.Save(&item).Error
}

func GetCache(key string, target interface{}) (bool, error) {
	var item models.CacheItem
	err := db.DB.Where("key = ? AND expires_at > ?", key, time.Now()).First(&item).Error
	if err == gorm.ErrRecordNotFound {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	err = json.Unmarshal(item.Value, target)
	return true, err
}

func CleanCache() {
	db.DB.Where("expires_at < ?", time.Now()).Delete(&models.CacheItem{})
}
