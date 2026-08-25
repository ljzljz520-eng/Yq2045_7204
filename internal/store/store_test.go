package store

import (
	"context"
	"path/filepath"
	"testing"

	"sitepay/internal/domain"
)

func TestPersistenceSurvivesReopen(t *testing.T) {
	path := filepath.Join(t.TempDir(), "payroll.db")
	ctx := context.Background()
	db, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	worker, err := db.SaveWorker(ctx, domain.Worker{Name: "Lin", Trade: "mason"})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.SaveSetting(ctx, domain.Setting{Key: "project", Value: "north-yard"}); err != nil {
		t.Fatal(err)
	}
	reopened, err := db.Reopen()
	if err != nil {
		t.Fatal(err)
	}
	defer reopened.Close()
	loaded, err := reopened.FindWorker(ctx, worker.Name, worker.Trade)
	if err != nil || loaded.ID != worker.ID {
		t.Fatalf("loaded worker=%+v err=%v", loaded, err)
	}
	setting, err := reopened.GetSetting(ctx, "project")
	if err != nil || setting.Value != "north-yard" {
		t.Fatalf("setting=%+v err=%v", setting, err)
	}
}

func TestHealthSnapshot(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	snapshot, err := db.Health(context.Background())
	if err != nil || !snapshot.DatabaseReady || snapshot.Workers != 0 {
		t.Fatalf("snapshot=%+v err=%v", snapshot, err)
	}
	if _, err := db.SearchWorkers(context.Background(), ""); err == nil {
		t.Fatal("expected empty search query error")
	}
}
