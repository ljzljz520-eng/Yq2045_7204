package validation

import (
	"fmt"
	"regexp"
	"strings"

	"sitepay/internal/domain"
)

var datePattern = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)

type Validator struct {
	MaxRows       int
	AllowedTrades map[string]bool
}

func New() Validator {
	return Validator{MaxRows: 10000, AllowedTrades: map[string]bool{
		"mason":       true,
		"carpenter":   true,
		"steelworker": true,
		"electrician": true,
		"plumber":     true,
	}}
}

func (v Validator) ValidateWorker(worker domain.Worker) error {
	errors := WorkerErrors(worker)
	if worker.Trade != "" && len(v.AllowedTrades) > 0 && !v.AllowedTrades[strings.ToLower(worker.Trade)] {
		errors.Add("trade", "is not configured for this project")
	}
	if errors.Empty() {
		return nil
	}
	return errors
}

func (v Validator) ValidateEntry(entry domain.WorkEntry) error {
	errors := EntryErrors(entry)
	if entry.WorkDate != "" && !datePattern.MatchString(entry.WorkDate) {
		errors.Add("work_date", "must be a calendar date")
	}
	if errors.Empty() {
		return nil
	}
	return errors
}

func (v Validator) ValidatePolicy(policy domain.AllowancePolicy) error {
	errors := PolicyErrors(policy)
	if errors.Empty() {
		return nil
	}
	return errors
}

func (v Validator) ValidateBatch(entries []domain.WorkEntry) error {
	if len(entries) == 0 {
		return fmt.Errorf("batch cannot be empty")
	}
	if v.MaxRows > 0 && len(entries) > v.MaxRows {
		return fmt.Errorf("batch has %d rows; maximum is %d", len(entries), v.MaxRows)
	}
	for index, entry := range entries {
		if err := v.ValidateEntry(entry); err != nil {
			return fmt.Errorf("row %d: %w", index+1, err)
		}
	}
	return nil
}

func (v Validator) NormalizeTrade(trade string) (string, error) {
	value := strings.ToLower(strings.TrimSpace(trade))
	if value == "" {
		return "", fmt.Errorf("trade is required")
	}
	if len(v.AllowedTrades) > 0 && !v.AllowedTrades[value] {
		return "", fmt.Errorf("trade %q is not configured", trade)
	}
	return value, nil
}

func ValidateRequest(input domain.PayrollInput) error {
	return Merge(WorkerErrors(input.Worker), EntryErrors(input.Entry), PolicyErrors(input.Policy))
}
