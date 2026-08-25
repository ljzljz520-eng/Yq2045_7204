package report

import (
	"encoding/csv"
	"fmt"
	"io"
	"strconv"

	"sitepay/internal/domain"
)

func WriteCSV(writer io.Writer, statements []domain.PayrollStatement, totals domain.ReportTotals) error {
	if writer == nil {
		return fmt.Errorf("report writer is nil")
	}
	out := csv.NewWriter(writer)
	if err := out.Write([]string{"statement_no", "worker_name", "trade", "work_date", "gross", "night_allowance", "quality_deduction", "net", "status", "review"}); err != nil {
		return err
	}
	for _, statement := range statements {
		review := ""
		if statement.Status == domain.StatementReview {
			review = "manager_confirmation"
		}
		if err := out.Write([]string{statement.StatementNo, statement.WorkerName, statement.Trade, statement.WorkDate, domain.FormatCents(statement.GrossCents), domain.FormatCents(statement.AllowanceCents), domain.FormatCents(statement.DeductionCents), domain.FormatCents(statement.NetCents), statement.Status, review}); err != nil {
			return err
		}
	}
	if err := out.Write([]string{"TOTAL", "", "", strconv.Itoa(totals.Statements), domain.FormatCents(totals.GrossCents), domain.FormatCents(totals.NightCents), domain.FormatCents(totals.DeductCents), domain.FormatCents(totals.NetCents), "", strconv.Itoa(totals.Reviews)}); err != nil {
		return err
	}
	out.Flush()
	return out.Error()
}

func Header() []string {
	return []string{"statement_no", "worker_name", "trade", "work_date", "gross", "night_allowance", "quality_deduction", "net", "status", "review"}
}
