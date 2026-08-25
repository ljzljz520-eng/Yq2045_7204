package sitepay_test

import (
	"bytes"
	"context"
	"path/filepath"
	"strings"
	"testing"

	"sitepay/internal/domain"
	"sitepay/internal/service"
	"sitepay/internal/store"
)

func TestSecondaryWorkflow(t *testing.T) {
	db, err := store.Open(filepath.Join(t.TempDir(), "batch.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	payrollService, err := service.NewPayrollService(db, []domain.AllowancePolicy{{Trade: "mason", NightCap: 10, RequiresReview: true, Active: true}})
	if err != nil {
		t.Fatal(err)
	}
	batch := service.NewBatchService(payrollService)
	input := "worker_name,trade,completed_pieces,unit_price,night_allowance,quality_deduction,work_date\nLin,mason,2,5.00,3.25,0,2026-08-25\n"
	summary, issues, err := batch.Process(context.Background(), strings.NewReader(input), "manager", "batch-1")
	if err != nil || !issues.Empty() || summary.Imported != 1 || len(summary.StatementIDs) != 1 {
		t.Fatalf("summary=%+v issues=%v err=%v", summary, issues, err)
	}
}

func TestTertiaryWorkflow(t *testing.T) {
	db, err := store.Open(filepath.Join(t.TempDir(), "report.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	payrollService, err := service.NewPayrollService(db, []domain.AllowancePolicy{{Trade: "mason", NightCap: 10, RequiresReview: true, Active: true}})
	if err != nil {
		t.Fatal(err)
	}
	statement, err := payrollService.Generate(context.Background(), domain.Worker{Name: "Lin", Trade: "mason"}, domain.WorkEntry{WorkerName: "Lin", Trade: "mason", CompletedPieces: 2, UnitPrice: 5, NightAllowance: 3, WorkDate: "2026-08-25"}, "manager", "report")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := payrollService.Approve(context.Background(), statement.ID, "manager"); err != nil {
		t.Fatal(err)
	}
	reportService := service.NewReportService(payrollService)
	var output bytes.Buffer
	totals, err := reportService.Export(context.Background(), &output, "json", "")
	if err != nil || totals.Statements != 1 || !strings.Contains(output.String(), "statement_no") {
		t.Fatalf("totals=%+v output=%q err=%v", totals, output.String(), err)
	}
}
