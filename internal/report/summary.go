package report

import (
	"sort"

	"sitepay/internal/domain"
)

func Totals(statements []domain.PayrollStatement) domain.ReportTotals {
	var result domain.ReportTotals
	result.Statements = len(statements)
	for _, statement := range statements {
		result.GrossCents += statement.GrossCents
		result.NightCents += statement.AllowanceCents
		result.DeductCents += statement.DeductionCents
		result.NetCents += statement.NetCents
		if statement.Status == domain.StatementReview {
			result.Reviews++
		}
	}
	return result
}

func ByWorker(statements []domain.PayrollStatement) map[string]domain.ReportTotals {
	result := make(map[string]domain.ReportTotals)
	for _, statement := range statements {
		totals := result[statement.WorkerName]
		totals.Statements++
		totals.GrossCents += statement.GrossCents
		totals.NightCents += statement.AllowanceCents
		totals.DeductCents += statement.DeductionCents
		totals.NetCents += statement.NetCents
		if statement.Status == domain.StatementReview {
			totals.Reviews++
		}
		result[statement.WorkerName] = totals
	}
	return result
}

func WorkerNames(statements []domain.PayrollStatement) []string {
	seen := make(map[string]bool)
	for _, statement := range statements {
		seen[statement.WorkerName] = true
	}
	result := make([]string, 0, len(seen))
	for name := range seen {
		result = append(result, name)
	}
	sort.Strings(result)
	return result
}

func FilterByDate(statements []domain.PayrollStatement, from, to string) []domain.PayrollStatement {
	result := make([]domain.PayrollStatement, 0, len(statements))
	for _, statement := range statements {
		if from != "" && statement.WorkDate < from {
			continue
		}
		if to != "" && statement.WorkDate > to {
			continue
		}
		result = append(result, statement)
	}
	return result
}
