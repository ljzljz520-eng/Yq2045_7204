package service

import (
	"context"
	"path/filepath"
	"testing"

	"sitepay/internal/domain"
	"sitepay/internal/store"
)

func TestPrimaryWorkflow(t *testing.T) {
	db, err := store.Open(filepath.Join(t.TempDir(), "service.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	service, err := NewPayrollService(db, []domain.AllowancePolicy{{Trade: "mason", NightCap: 20, RequiresReview: true, Active: true}})
	if err != nil {
		t.Fatal(err)
	}
	statement, err := service.Generate(context.Background(), domain.Worker{Name: "Lin", Trade: "mason"}, domain.WorkEntry{WorkerName: "Lin", Trade: "mason", CompletedPieces: 4, UnitPrice: 2.5, NightAllowance: 5, WorkDate: "2026-08-25"}, "manager", "req")
	if err != nil {
		t.Fatal(err)
	}
	if statement.ID == 0 || statement.Status != domain.StatementApproved {
		t.Fatalf("statement=%+v", statement)
	}
}
