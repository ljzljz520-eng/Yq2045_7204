package payroll

import (
	"fmt"
	"strings"
	"time"

	"sitepay/internal/domain"
)

type Calculator struct {
	Location string
	Now      func() time.Time
}

func NewCalculator(location string) Calculator {
	return Calculator{Location: location, Now: func() time.Time { return time.Unix(0, 0).UTC() }}
}

func (c Calculator) Calculate(input domain.PayrollInput) (domain.PayrollLine, error) {
	if err := input.Validate(); err != nil {
		return domain.PayrollLine{}, err
	}
	if c.Now == nil {
		return domain.PayrollLine{}, fmt.Errorf("calculator clock is not configured")
	}
	unitCents := domain.CentsFromFloat(input.Entry.UnitPrice)
	gross := domain.MultiplyCents(unitCents, input.Entry.CompletedPieces)
	allowance := domain.CentsFromFloat(input.Entry.NightAllowance)
	deduction := domain.CentsFromFloat(input.Entry.QualityDeduction)
	review, reason := AllowanceReview(input.Entry.NightAllowance, input.Policy)
	return domain.PayrollLine{
		EntryID:        input.Entry.ID,
		WorkerID:       input.Worker.ID,
		WorkerName:     input.Worker.Name,
		Trade:          input.Worker.Trade,
		Pieces:         input.Entry.CompletedPieces,
		UnitPriceCents: unitCents,
		GrossCents:     gross,
		NightCents:     allowance,
		DeductionCents: deduction,
		NetCents:       domain.SubCents(domain.AddCents(gross, allowance), deduction),
		ReviewRequired: review,
		ReviewReason:   reason,
		CalculatedAt:   c.Now().UTC(),
	}, nil
}

func (c Calculator) ValidateLocation() error {
	if strings.TrimSpace(c.Location) == "" {
		return fmt.Errorf("location is required")
	}
	return nil
}

func (c Calculator) WithClock(now func() time.Time) Calculator {
	clone := c
	clone.Now = now
	return clone
}
