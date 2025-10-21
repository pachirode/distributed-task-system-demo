package store

import (
	"context"
	"sync"

	"gorm.io/gorm"
)

var (
	once  sync.Once
	Store *datastore
)

type txKey struct {
}

type IStore interface {
	TX(context.Context, func(ctx context.Context) error) error
	Tasks() ITaskStore
}

type datastore struct {
	db *gorm.DB
}

func NewStore(db *gorm.DB) *datastore {
	once.Do(func() {
		Store = &datastore{
			db: db,
		}
	})

	return Store
}

var _ IStore = (*datastore)(nil)

func (ds *datastore) DB(ctx context.Context) *gorm.DB {
	tx, ok := ctx.Value(txKey{}).(*gorm.DB)
	if ok {
		return tx
	}

	return ds.db
}

func (ds *datastore) TX(ctx context.Context, fn func(ctx context.Context) error) error {
	return ds.db.WithContext(ctx).Transaction(
		func(tx *gorm.DB) error {
			ctx = context.WithValue(ctx, txKey{}, tx)
			return fn(ctx)
		},
	)
}

func (ds *datastore) Tasks() ITaskStore {
	return newTaskMStore(ds)
}
