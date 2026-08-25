package validation

import (
	"fmt"
	"strings"

	"sitepay/internal/domain"
)

type FieldError struct {
	Field   string
	Message string
}

func (e FieldError) Error() string { return fmt.Sprintf("%s: %s", e.Field, e.Message) }

type Errors []FieldError

func (e Errors) Error() string {
	parts := make([]string, 0, len(e))
	for _, item := range e {
		parts = append(parts, item.Error())
	}
	return strings.Join(parts, "; ")
}

func (e Errors) Empty() bool { return len(e) == 0 }

func (e *Errors) Add(field, message string) {
	*e = append(*e, FieldError{Field: field, Message: message})
}

func WorkerErrors(worker domain.Worker) Errors {
	var result Errors
	if strings.TrimSpace(worker.Name) == "" {
		result.Add("worker_name", "is required")
	}
	if len([]rune(worker.Name)) > 80 {
		result.Add("worker_name", "must be 80 characters or fewer")
	}
	if strings.TrimSpace(worker.Trade) == "" {
		result.Add("trade", "is required")
	}
	if len([]rune(worker.Trade)) > 80 {
		result.Add("trade", "must be 80 characters or fewer")
	}
	return result
}

func EntryErrors(entry domain.WorkEntry) Errors {
	var result Errors
	if entry.CompletedPieces <= 0 {
		result.Add("completed_pieces", "must be positive")
	}
	if entry.CompletedPieces > 100000 {
		result.Add("completed_pieces", "exceeds daily limit")
	}
	if entry.UnitPrice < 0 {
		result.Add("unit_price", "cannot be negative")
	}
	if entry.NightAllowance < 0 {
		result.Add("night_allowance", "cannot be negative")
	}
	if entry.QualityDeduction < 0 {
		result.Add("quality_deduction", "cannot be negative")
	}
	if strings.TrimSpace(entry.WorkDate) == "" {
		result.Add("work_date", "is required")
	}
	if len(entry.WorkDate) != 10 {
		result.Add("work_date", "must use YYYY-MM-DD")
	}
	return result
}

func PolicyErrors(policy domain.AllowancePolicy) Errors {
	var result Errors
	if strings.TrimSpace(policy.Trade) == "" {
		result.Add("trade", "is required")
	}
	if policy.NightCap < 0 {
		result.Add("night_cap", "cannot be negative")
	}
	if policy.NightCap > 100000 {
		result.Add("night_cap", "is outside configured range")
	}
	return result
}

func Merge(errors ...Errors) Errors {
	var result Errors
	for _, group := range errors {
		result = append(result, group...)
	}
	return result
}
