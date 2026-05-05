package repository

import (
	"gorm.io/gorm"
)

type Repository[T any] interface {
	FindByID(id uint) (*T, error)
	FindAll() ([]T, error)
	Find(query interface{}, args ...interface{}) ([]T, error)
	FindOne(query interface{}, args ...interface{}) (*T, error)
	Create(entity *T) error
	Update(entity *T) error
	Delete(id uint) error
	DeleteWhere(query interface{}, args ...interface{}) error
}

type baseRepository[T any] struct {
	db *gorm.DB
}

func NewRepository[T any](db *gorm.DB) Repository[T] {
	return &baseRepository[T]{db: db}
}

func (r *baseRepository[T]) FindByID(id uint) (*T, error) {
	var entity T
	err := r.db.First(&entity, id).Error
	if err != nil {
		return nil, err
	}
	return &entity, nil
}

func (r *baseRepository[T]) FindAll() ([]T, error) {
	var entities []T
	err := r.db.Find(&entities).Error
	return entities, err
}

func (r *baseRepository[T]) Find(query interface{}, args ...interface{}) ([]T, error) {
	var entities []T
	err := r.db.Where(query, args...).Find(&entities).Error
	return entities, err
}

func (r *baseRepository[T]) FindOne(query interface{}, args ...interface{}) (*T, error) {
	var entity T
	err := r.db.Where(query, args...).First(&entity).Error
	if err != nil {
		return nil, err
	}
	return &entity, nil
}

func (r *baseRepository[T]) Create(entity *T) error {
	return r.db.Create(entity).Error
}

func (r *baseRepository[T]) Update(entity *T) error {
	return r.db.Save(entity).Error
}

func (r *baseRepository[T]) Delete(id uint) error {
	var entity T
	return r.db.Delete(&entity, id).Error
}

func (r *baseRepository[T]) DeleteWhere(query interface{}, args ...interface{}) error {
	var entity T
	return r.db.Where(query, args...).Delete(&entity).Error
}
