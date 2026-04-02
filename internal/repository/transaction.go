package repository

import (
	"context"

	"gorm.io/gorm"
)

type transactionManager struct {
	db *gorm.DB
}

func NewTransactionManager(db *gorm.DB) interface{} {
	return &transactionManager{db: db}
}

func (t *transactionManager) BeginTx(ctx context.Context) (interface{}, error) {
	return t.db.WithContext(ctx).Begin(), nil
}

func (t *transactionManager) Commit(tx interface{}) error {
	return tx.(*gorm.DB).Commit().Error
}

func (t *transactionManager) Rollback(tx interface{}) error {
	return tx.(*gorm.DB).Rollback().Error
}
