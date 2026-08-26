package payroll

import (
	"testing"
	"time"

	"sitepay/internal/domain"
)

func TestPayrollKeepsCents(t *testing.T) {
	calculator := NewCalculator("yard").WithClock(func() time.Time { return time.Unix(0, 0).UTC() })
	policy := domain.AllowancePolicy{Trade: "mason", NightCap: 30, RequiresReview: true, Active: true}
	input := domain.PayrollInput{Worker: domain.Worker{ID: 1, Name: "Lin", Trade: "mason"}, Entry: domain.WorkEntry{WorkerID: 1, WorkerName: "Lin", Trade: "mason", CompletedPieces: 10, UnitPrice: 3.25, NightAllowance: 12.75, QualityDeduction: 1.25, WorkDate: "2026-08-25"}, Policy: policy, PreparedBy: "manager", RequestID: "r1"}
	statement, err := NewAggregator(calculator).BuildStatement(input)
	if err != nil {
		t.Fatal(err)
	}
	if statement.Lines[0].NightCents != 1275 {
		t.Fatalf("night allowance cents=%d, want 1275", statement.Lines[0].NightCents)
	}
	if statement.NetCents != 4400 {
		t.Fatalf("net cents=%d, want 4400", statement.NetCents)
	}
}
