package report

import (
	"sort"
	"strings"

	"sitepay/internal/domain"
)

type ReviewItem struct {
	StatementID int64
	StatementNo string
	WorkerName  string
	Reason      string
	AmountCents int64
}

type TradeTotals struct {
	Trade       string
	Statements  int
	Pieces      int
	GrossCents  int64
	NetCents    int64
	ReviewCount int
}

func ReviewQueue(statements []domain.PayrollStatement) []ReviewItem {
	result := make([]ReviewItem, 0)
	for _, statement := range statements {
		for _, line := range statement.Lines {
			if !line.ReviewRequired {
				continue
			}
			result = append(result, ReviewItem{StatementID: statement.ID, StatementNo: statement.StatementNo, WorkerName: statement.WorkerName, Reason: line.ReviewReason, AmountCents: line.NightCents})
		}
	}
	sort.SliceStable(result, func(i, j int) bool {
		if result[i].WorkerName == result[j].WorkerName {
			return result[i].StatementNo < result[j].StatementNo
		}
		return result[i].WorkerName < result[j].WorkerName
	})
	return result
}

func TotalsByTrade(statements []domain.PayrollStatement) []TradeTotals {
	groups := make(map[string]TradeTotals)
	for _, statement := range statements {
		key := strings.ToLower(statement.Trade)
		totals := groups[key]
		totals.Trade = statement.Trade
		totals.Statements++
		totals.GrossCents += statement.GrossCents
		totals.NetCents += statement.NetCents
		for _, line := range statement.Lines {
			totals.Pieces += line.Pieces
			if line.ReviewRequired {
				totals.ReviewCount++
			}
		}
		groups[key] = totals
	}
	result := make([]TradeTotals, 0, len(groups))
	for _, totals := range groups {
		result = append(result, totals)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Trade < result[j].Trade })
	return result
}

func Reconcile(statements []domain.PayrollStatement) (bool, []string) {
	errors := make([]string, 0)
	for _, statement := range statements {
		gross := int64(0)
		allowance := int64(0)
		deduction := int64(0)
		net := int64(0)
		for _, line := range statement.Lines {
			gross += line.GrossCents
			allowance += line.NightCents
			deduction += line.DeductionCents
			net += line.NetCents
		}
		if statement.GrossCents != gross {
			errors = append(errors, statement.StatementNo+": gross mismatch")
		}
		if statement.AllowanceCents != allowance {
			errors = append(errors, statement.StatementNo+": allowance mismatch")
		}
		if statement.DeductionCents != deduction {
			errors = append(errors, statement.StatementNo+": deduction mismatch")
		}
		if statement.NetCents != net {
			errors = append(errors, statement.StatementNo+": net mismatch")
		}
	}
	return len(errors) == 0, errors
}

func FilterStatus(statements []domain.PayrollStatement, statuses ...string) []domain.PayrollStatement {
	allowed := make(map[string]bool, len(statuses))
	for _, status := range statuses {
		allowed[status] = true
	}
	result := make([]domain.PayrollStatement, 0)
	for _, statement := range statements {
		if len(allowed) == 0 || allowed[statement.Status] {
			result = append(result, statement)
		}
	}
	return result
}
