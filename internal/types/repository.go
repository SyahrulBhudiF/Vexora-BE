package types

import (
	"fmt"
	"gorm.io/gorm"
)

type TransactionFunc[T any] func(tx *Repository[T]) error

type Repository[T any] struct {
	DB *gorm.DB
}

func (r *Repository[T]) Create(entity *T) error {
	return r.DB.Create(entity).Error
}

func (r *Repository[T]) Update(entity *T) error {
	return r.DB.Save(entity).Error
}

func (r *Repository[T]) Delete(entity *T) error {
	return r.DB.Delete(entity).Error
}

func (r *Repository[T]) CountByUUID(uuid any) (int64, error) {
	var total int64
	err := r.DB.Model(new(T)).Where("uuid = ?", uuid).Count(&total).Error
	return total, err
}

func (r *Repository[T]) FindByUUID(uuid any) error {
	var entity *T
	return r.DB.Where("uuid = ?", uuid).Take(entity).Error
}

func (r *Repository[T]) FindAll(entities *[]T) error {
	return r.DB.Find(entities).Error
}

func (r *Repository[T]) Find(entity *T) error {
	return r.DB.First(entity, entity).Error
}

func (r *Repository[T]) Exists(entity *T) bool {
	return r.DB.First(entity, entity).RowsAffected > 0
}

func (r *Repository[T]) FindByColumnValue(columnName string, value any) ([]T, error) {
	var entities []T
	err := r.DB.Where(fmt.Sprintf("%s = ?", columnName), value).Find(&entities).Error
	if err != nil {
		return nil, err
	}
	return entities, nil
}

func (r *Repository[T]) BeginTx() *Repository[T] {
	return &Repository[T]{
		DB: r.DB.Begin(),
	}
}

func (r *Repository[T]) Commit() error {
	return r.DB.Commit().Error
}

func (r *Repository[T]) Rollback() error {
	return r.DB.Rollback().Error
}

func (r *Repository[T]) Transaction(fn TransactionFunc[T]) error {
	tx := r.BeginTx()

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	if err := fn(tx); err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return fmt.Errorf("transaction error: %v, rollback error: %v", err, rollbackErr)
		}
		return err
	}

	return tx.Commit()
}
