package database

import (
	"context"
	"errors"
	"testing"

	"gorm.io/gorm"
)

type txThing struct {
	ID   uint
	Name string
}

// newTestDB connects to the MySQL from the DB_* environment, in a database of
// its own, and skips the test when there is no MySQL.
func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	cfg := ConfigFromEnv("shared_database_test")
	if err := CreateIfMissing(cfg); err != nil {
		t.Skipf("no MySQL: %v", err)
	}
	db, err := Connect(cfg)
	if err != nil {
		t.Skipf("no MySQL: %v", err)
	}
	t.Cleanup(func() { db.Exec("DROP DATABASE shared_database_test") })
	if err := db.Migrator().DropTable(&txThing{}); err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&txThing{}); err != nil {
		t.Fatal(err)
	}
	return db
}

func count(t *testing.T, ctx context.Context, db *gorm.DB) int64 {
	t.Helper()
	var n int64
	if err := DB(ctx, db).Model(&txThing{}).Count(&n).Error; err != nil {
		t.Fatal(err)
	}
	return n
}

func TestTransaction(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()
	tx := NewTransactor(db)
	insert := func(ctx context.Context, name string) error {
		return DB(ctx, db).Create(&txThing{Name: name}).Error
	}

	t.Run("commits", func(t *testing.T) {
		err := tx.Transaction(ctx, func(ctx context.Context) error {
			if err := insert(ctx, "a"); err != nil {
				return err
			}
			return insert(ctx, "b")
		})
		if err != nil || count(t, ctx, db) != 2 {
			t.Fatalf("err %v, %d rows, want 2", err, count(t, ctx, db))
		}
	})

	t.Run("rolls back on an error", func(t *testing.T) {
		boom := errors.New("boom")
		err := tx.Transaction(ctx, func(ctx context.Context) error {
			if err := insert(ctx, "c"); err != nil {
				return err
			}
			return boom
		})
		if !errors.Is(err, boom) || count(t, ctx, db) != 2 {
			t.Fatalf("err %v, %d rows, want boom and 2", err, count(t, ctx, db))
		}
	})

	t.Run("a nested transaction joins the outer one", func(t *testing.T) {
		boom := errors.New("boom")
		err := tx.Transaction(ctx, func(ctx context.Context) error {
			// Committed by itself, this would survive the outer rollback.
			if err := Transaction(ctx, db, func(ctx context.Context) error { return insert(ctx, "d") }); err != nil {
				return err
			}
			return boom
		})
		if !errors.Is(err, boom) || count(t, ctx, db) != 2 {
			t.Fatalf("err %v, %d rows, want boom and 2", err, count(t, ctx, db))
		}
	})

	t.Run("queries without the ctx don't see uncommitted rows", func(t *testing.T) {
		err := tx.Transaction(ctx, func(txCtx context.Context) error {
			if err := insert(txCtx, "e"); err != nil {
				return err
			}
			if inside, outside := count(t, txCtx, db), count(t, ctx, db); inside != 3 || outside != 2 {
				t.Errorf("inside %d, outside %d; want 3 and 2", inside, outside)
			}
			return nil
		})
		if err != nil || count(t, ctx, db) != 3 {
			t.Fatalf("err %v, %d rows, want 3", err, count(t, ctx, db))
		}
	})
}

type modelThing struct {
	Model
	Name string
	SoftDelete
}

func TestModel(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()
	if err := db.AutoMigrate(&modelThing{}); err != nil {
		t.Fatal(err)
	}

	thing := &modelThing{Name: "new"}
	if err := DB(ctx, db).Create(thing).Error; err != nil {
		t.Fatal(err)
	}
	if len(thing.ID) != 27 || thing.CreatedAt.IsZero() {
		t.Errorf("created %+v: want a 27-char id and a created_at", thing.Model)
	}

	given := &modelThing{Model: Model{ID: NewID()}, Name: "given"}
	want := given.ID
	if err := DB(ctx, db).Create(given).Error; err != nil || given.ID != want {
		t.Errorf("a given id: err %v, id %q, want %q", err, given.ID, want)
	}

	if err := DB(ctx, db).Delete(thing).Error; err != nil {
		t.Fatal(err)
	}
	var left []modelThing
	if err := DB(ctx, db).Find(&left).Error; err != nil || len(left) != 1 || left[0].Name != "given" {
		t.Errorf("after a soft delete: err %v, %d rows, want only \"given\"", err, len(left))
	}
}

func TestTranslateError(t *testing.T) {
	errNotFound := errors.New("not found")
	errTaken := errors.New("taken")
	other := errors.New("other")
	tests := []struct {
		name      string
		err       error
		duplicate []error
		want      error
	}{
		{"missing row", gorm.ErrRecordNotFound, nil, errNotFound},
		{"duplicate key, with a duplicate error", gorm.ErrDuplicatedKey, []error{errTaken}, errTaken},
		{"duplicate key, table without unique columns", gorm.ErrDuplicatedKey, nil, gorm.ErrDuplicatedKey},
		{"anything else", other, []error{errTaken}, other},
	}
	for _, tt := range tests {
		if got := TranslateError(tt.err, errNotFound, tt.duplicate...); got != tt.want {
			t.Errorf("%s: got %v, want %v", tt.name, got, tt.want)
		}
	}
}
