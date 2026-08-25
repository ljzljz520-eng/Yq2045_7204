package service

import (
	"context"
	"fmt"
	"io"
	"strings"

	"sitepay/internal/domain"
	"sitepay/internal/report"
)

type ReportService struct{ Payroll *PayrollService }

func NewReportService(payrollService *PayrollService) *ReportService {
	return &ReportService{Payroll: payrollService}
}

func (r *ReportService) Export(ctx context.Context, writer io.Writer, format string, status string) (domain.ReportTotals, error) {
	if r == nil || r.Payroll == nil {
		return domain.ReportTotals{}, fmt.Errorf("report service is not configured")
	}
	statements, err := r.Payroll.List(ctx, status)
	if err != nil {
		return domain.ReportTotals{}, err
	}
	if consistent, details := report.Reconcile(statements); !consistent {
		return domain.ReportTotals{}, fmt.Errorf("report reconciliation failed: %s", strings.Join(details, "; "))
	}
	totals := report.Totals(statements)
	switch format {
	case "csv":
		err = report.WriteCSV(writer, statements, totals)
	case "json":
		err = report.WriteJSON(writer, statements, totals)
	default:
		err = fmt.Errorf("unsupported report format %q", format)
	}
	return totals, err
}
