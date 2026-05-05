package tests

import (
	"os"
	"testing"
	"time"

	"github.com/user/lector/internal/db"
	"github.com/user/lector/internal/repository"
	"gorm.io/gorm"
)

type TestEntity struct {
	ID        uint `gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
	Name      string
}

func TestGenericRepository(t *testing.T) {
	testDB := "test_generic_repo.db"
	defer os.Remove(testDB)
	db.InitDB(testDB)

	db.DB.AutoMigrate(&TestEntity{})

	repo := repository.NewRepository[TestEntity](db.DB)

	t.Run("Create and Find", func(t *testing.T) {
		entity := &TestEntity{Name: "Test 1"}
		err := repo.Create(entity)
		if err != nil {
			t.Fatalf("Failed to create entity: %v", err)
		}
		if entity.ID == 0 {
			t.Errorf("Expected non-zero ID")
		}

		found, err := repo.FindByID(entity.ID)
		if err != nil {
			t.Fatalf("Failed to find entity: %v", err)
		}
		if found.Name != "Test 1" {
			t.Errorf("Expected 'Test 1', got '%s'", found.Name)
		}
	})

	t.Run("Update", func(t *testing.T) {
		entity := &TestEntity{Name: "Test 2"}
		repo.Create(entity)

		entity.Name = "Updated Test 2"
		err := repo.Update(entity)
		if err != nil {
			t.Fatalf("Failed to update entity: %v", err)
		}

		found, _ := repo.FindByID(entity.ID)
		if found.Name != "Updated Test 2" {
			t.Errorf("Expected 'Updated Test 2', got '%s'", found.Name)
		}
	})

	t.Run("Delete", func(t *testing.T) {
		entity := &TestEntity{Name: "Test 3"}
		repo.Create(entity)

		err := repo.Delete(entity.ID)
		if err != nil {
			t.Fatalf("Failed to delete entity: %v", err)
		}

		_, err = repo.FindByID(entity.ID)
		if err == nil {
			t.Errorf("Expected error when finding deleted entity")
		}
	})
}
