package kafka

import (
	"context"
	"errors"
	"testing"

	"github.com/jochem11/inventory-manager/shared/database"
	"gorm.io/gorm"
)

func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	cfg := database.ConfigFromEnv("shared_kafka_test")
	if err := database.CreateIfMissing(cfg); err != nil {
		t.Skipf("no MySQL: %v", err)
	}
	db, err := database.Connect(cfg)
	if err != nil {
		t.Skipf("no MySQL: %v", err)
	}
	t.Cleanup(func() { db.Exec("DROP DATABASE shared_kafka_test") })
	if err := db.Migrator().DropTable(Models()...); err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(Models()...); err != nil {
		t.Fatal(err)
	}
	return db
}

func TestHandleOnce(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()
	runs := 0
	handle := func(context.Context) error { runs++; return nil }

	for range 3 { // delivered three times
		if err := HandleOnce(ctx, db, "event-1", handle); err != nil {
			t.Fatal(err)
		}
	}
	if runs != 1 {
		t.Errorf("handled %d times, want once", runs)
	}

	// A failed attempt leaves no record behind, so the retry handles it.
	boom := errors.New("boom")
	attempts := 0
	failOnce := func(context.Context) error {
		attempts++
		if attempts == 1 {
			return boom
		}
		return nil
	}
	if err := HandleOnce(ctx, db, "event-2", failOnce); !errors.Is(err, boom) {
		t.Fatalf("first attempt: %v, want boom", err)
	}
	if err := HandleOnce(ctx, db, "event-2", failOnce); err != nil || attempts != 2 {
		t.Errorf("retry: err %v, %d attempts, want 2", err, attempts)
	}
}
