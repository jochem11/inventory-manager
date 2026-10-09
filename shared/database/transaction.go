package database

import (
	"context"

	"gorm.io/gorm"
)

// Transactions travel in the context. Transaction starts one and puts it in
// the ctx it passes on; DB returns it to every query made with that ctx. So a
// service can make several repository calls atomic without the repositories
// knowing:
//
//	err := tx.Transaction(ctx, func(ctx context.Context) error {
//		if err := users.Delete(ctx, id); err != nil {
//			return err // rolls back
//		}
//		return kafka.Enqueue(ctx, db, …) // same transaction
//	})
//
// Repositories must use DB(ctx, r.db) for every query, never r.db directly.

type txKey struct{}

// Transactor runs a function in one database transaction. Services depend on
// this interface; NewTransactor gives the implementation.
type Transactor interface {
	// Transaction runs fn in a transaction: committed when fn returns nil,
	// rolled back when it returns an error or panics. Called inside another
	// transaction, it joins that one, so a function can make its work atomic
	// without knowing whether its caller already did.
	Transaction(ctx context.Context, fn func(ctx context.Context) error) error
}

type transactor struct {
	db *gorm.DB
}

// NewTransactor returns a Transactor for db.
func NewTransactor(db *gorm.DB) Transactor {
	return transactor{db: db}
}

func (t transactor) Transaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return Transaction(ctx, t.db, fn)
}

// Transaction is Transactor.Transaction for code that holds a *gorm.DB.
func Transaction(ctx context.Context, db *gorm.DB, fn func(ctx context.Context) error) error {
	if _, ok := ctx.Value(txKey{}).(*gorm.DB); ok {
		return fn(ctx)
	}
	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(context.WithValue(ctx, txKey{}, tx))
	})
}

// DB returns the database handle for a query: the transaction in ctx when
// there is one, otherwise db. Either way it carries ctx, for cancellation and
// tracing.
func DB(ctx context.Context, db *gorm.DB) *gorm.DB {
	if tx, ok := ctx.Value(txKey{}).(*gorm.DB); ok {
		return tx.WithContext(ctx)
	}
	return db.WithContext(ctx)
}
