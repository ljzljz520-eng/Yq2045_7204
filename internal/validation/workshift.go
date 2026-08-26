package validation

import (
	"fmt"
	"time"

	"sitepay/internal/domain"
)

type WorkshiftRules struct {
	EarliestYear int
	LatestYear   int
	MaxPieces    int
	MaxDailyPay  float64
}

func DefaultWorkshiftRules() WorkshiftRules {
	return WorkshiftRules{EarliestYear: 2000, LatestYear: 2100, MaxPieces: 100000, MaxDailyPay: 1000000}
}

func (r WorkshiftRules) ValidateDate(value string) error {
	parsed, err := time.Parse("2006-01-02", value)
	if err != nil {
		return fmt.Errorf("invalid work date %q", value)
	}
	if parsed.Year() < r.EarliestYear || parsed.Year() > r.LatestYear {
		return fmt.Errorf("work date year is outside configured range")
	}
	return nil
}

func (r WorkshiftRules) ValidateEntry(entry domain.WorkEntry) error {
	if err := r.ValidateDate(entry.WorkDate); err != nil {
		return err
	}
	if r.MaxPieces > 0 && entry.CompletedPieces > r.MaxPieces {
		return fmt.Errorf("pieces exceed workshift maximum")
	}
	if r.MaxDailyPay > 0 && float64(entry.CompletedPieces)*entry.UnitPrice+entry.NightAllowance > r.MaxDailyPay {
		return fmt.Errorf("daily pay exceeds workshift maximum")
	}
	return nil
}

func (r WorkshiftRules) ValidateBatch(entries []domain.WorkEntry) []FieldError {
	errors := make([]FieldError, 0)
	for index, entry := range entries {
		if err := r.ValidateEntry(entry); err != nil {
			errors = append(errors, FieldError{Field: fmt.Sprintf("row_%d", index+1), Message: err.Error()})
		}
	}
	return errors
}

func DateWindow(entries []domain.WorkEntry) (time.Time, time.Time, error) {
	if len(entries) == 0 {
		return time.Time{}, time.Time{}, fmt.Errorf("entries are empty")
	}
	var earliest, latest time.Time
	for _, entry := range entries {
		value, err := time.Parse("2006-01-02", entry.WorkDate)
		if err != nil {
			return time.Time{}, time.Time{}, err
		}
		if earliest.IsZero() || value.Before(earliest) {
			earliest = value
		}
		if latest.IsZero() || value.After(latest) {
			latest = value
		}
	}
	return earliest, latest, nil
}

func IsWeekend(date string) bool {
	value, err := time.Parse("2006-01-02", date)
	if err != nil {
		return false
	}
	return value.Weekday() == time.Saturday || value.Weekday() == time.Sunday
}

func WeekendEntries(entries []domain.WorkEntry) []domain.WorkEntry {
	result := make([]domain.WorkEntry, 0)
	for _, entry := range entries {
		if IsWeekend(entry.WorkDate) {
			result = append(result, entry)
		}
	}
	return result
}
