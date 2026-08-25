package report

import (
	"bytes"
	"strings"
	"testing"

	"sitepay/internal/domain"
)

func TestReportExportsCSV(t *testing.T) {
	statement := domain.PayrollStatement{StatementNo: "r1", WorkerName: "Lin", Trade: "mason", WorkDate: "2026-08-25", GrossCents: 1000, AllowanceCents: 275, DeductionCents: 25, NetCents: 1250, Status: domain.StatementApproved}
	var output bytes.Buffer
	if err := WriteCSV(&output, []domain.PayrollStatement{statement}, Totals([]domain.PayrollStatement{statement})); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), "1,000") && !strings.Contains(output.String(), "10.00") {
		t.Fatalf("unexpected csv %q", output.String())
	}
}
