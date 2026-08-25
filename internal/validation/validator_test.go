package validation

import (
	"testing"

	"sitepay/internal/domain"
)

func TestValidatorRejectsUnknownTrade(t *testing.T) {
	err := New().ValidateWorker(domain.Worker{Name: "Lin", Trade: "unknown"})
	if err == nil {
		t.Fatal("expected trade validation error")
	}
}

func TestValidatorAcceptsEntry(t *testing.T) {
	entry := domain.WorkEntry{WorkerName: "Lin", Trade: "mason", CompletedPieces: 1, UnitPrice: 1, WorkDate: "2026-08-25"}
	if err := New().ValidateEntry(entry); err != nil {
		t.Fatal(err)
	}
}

func TestWorkshiftDateWindow(t *testing.T) {
	entries := []domain.WorkEntry{{WorkDate: "2026-08-22"}, {WorkDate: "2026-08-25"}}
	start, end, err := DateWindow(entries)
	if err != nil || start.Format("2006-01-02") != "2026-08-22" || end.Format("2006-01-02") != "2026-08-25" {
		t.Fatalf("window=%v %v err=%v", start, end, err)
	}
	if !IsWeekend("2026-08-22") {
		t.Fatal("expected weekend")
	}
}
