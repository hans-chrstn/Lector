package db

import (
	"fmt"
	"log"

	"github.com/user/lector/internal/models"
	"gorm.io/gorm"
)

func RunDataMigration(sourceDB *gorm.DB, targetDB *gorm.DB) error {
	if err := migrateTable(sourceDB, targetDB, "documents", &[]models.Document{}); err != nil {
		return err
	}
	if err := migrateTable(sourceDB, targetDB, "chapters", &[]models.Chapter{}); err != nil {
		return err
	}
	if err := migrateTable(sourceDB, targetDB, "reading_progresses", &[]models.ReadingProgress{}); err != nil {
		return err
	}
	if err := migrateTable(sourceDB, targetDB, "groups", &[]models.Group{}); err != nil {
		return err
	}
	if err := migrateTable(sourceDB, targetDB, "bookmarks", &[]models.Bookmark{}); err != nil {
		return err
	}
	if err := migrateTable(sourceDB, targetDB, "notes", &[]models.Note{}); err != nil {
		return err
	}
	if err := migrateTable(sourceDB, targetDB, "plugins", &[]models.Plugin{}); err != nil {
		return err
	}
	if err := migrateTable(sourceDB, targetDB, "library_paths", &[]models.LibraryPath{}); err != nil {
		return err
	}
	if err := migrateTable(sourceDB, targetDB, "reading_stats", &[]models.ReadingStat{}); err != nil {
		return err
	}
	if err := migrateTable(sourceDB, targetDB, "cache_items", &[]models.CacheItem{}); err != nil {
		return err
	}
	return nil
}

func migrateTable[T any](sourceDB *gorm.DB, targetDB *gorm.DB, tableName string, dest *[]T) error {
	log.Printf("Migrating %s...", tableName)

	var count int64
	sourceDB.Model(new(T)).Count(&count)
	if count == 0 {
		log.Printf("  -> 0 records found. Skipping.")
		return nil
	}

	if err := sourceDB.Find(dest).Error; err != nil {
		return fmt.Errorf("failed to read %s: %v", tableName, err)
	}

	targetDB.Exec("DELETE FROM " + tableName)

	if err := targetDB.CreateInBatches(dest, 100).Error; err != nil {
		return fmt.Errorf("failed to insert %s into target database: %v", tableName, err)
	}

	if tableName != "cache_items" && targetDB.Dialector.Name() == "postgres" {
		query := fmt.Sprintf("SELECT setval(pg_get_serial_sequence('%s', 'id'), COALESCE(max(id), 1), max(id) IS NOT NULL) FROM %s;", tableName, tableName)
		if err := targetDB.Exec(query).Error; err != nil {
			return fmt.Errorf("failed to sync sequence for %s: %v", tableName, err)
		}
	}

	log.Printf("  -> Successfully migrated %d records.", len(*dest))
	return nil
}
records.", len(*dest))
	return nil
}
