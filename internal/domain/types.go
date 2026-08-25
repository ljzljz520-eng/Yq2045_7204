package domain

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	ErrInvalidWorker = errors.New("invalid worker")
	ErrInvalidEntry  = errors.New("invalid work entry")
	ErrInvalidPolicy = errors.New("invalid allowance policy")
	ErrNotFound      = errors.New("record not found")
)

type Worker struct {
	ID        int64
	Name      string
	Trade     string
	CreatedAt time.Time
}

func (w Worker) Validate() error {
	if strings.TrimSpace(w.Name) == "" || strings.TrimSpace(w.Trade) == "" {
		return ErrInvalidWorker
	}
	if len([]rune(w.Name)) > 80 || len([]rune(w.Trade)) > 80 {
		return fmt.Errorf("%w: name or trade too long", ErrInvalidWorker)
	}
	return nil
}

type WorkEntry struct {
	ID               int64
	WorkerID         int64
	WorkerName       string
	Trade            string
	CompletedPieces  int
	UnitPrice        float64
	NightAllowance   float64
	QualityDeduction float64
	WorkDate         string
	ImportedLine     int
}

func (e WorkEntry) Validate() error {
	if e.WorkerID < 0 || strings.TrimSpace(e.WorkerName) == "" || strings.TrimSpace(e.Trade) == "" {
		return ErrInvalidEntry
	}
	if e.CompletedPieces <= 0 || e.CompletedPieces > 100000 {
		return fmt.Errorf("%w: pieces must be between 1 and 100000", ErrInvalidEntry)
	}
	if e.UnitPrice < 0 || e.NightAllowance < 0 || e.QualityDeduction < 0 {
		return fmt.Errorf("%w: amounts cannot be negative", ErrInvalidEntry)
	}
	if e.WorkDate == "" {
		return fmt.Errorf("%w: work date is required", ErrInvalidEntry)
	}
	return nil
}

type AllowancePolicy struct {
	ID             int64
	Trade          string
	NightCap       float64
	RequiresReview bool
	Active         bool
}

func (p AllowancePolicy) Validate() error {
	if strings.TrimSpace(p.Trade) == "" || p.NightCap < 0 {
		return ErrInvalidPolicy
	}
	return nil
}

type PayrollInput struct {
	Worker     Worker
	Entry      WorkEntry
	Policy     AllowancePolicy
	PreparedBy string
	RequestID  string
}

func (i PayrollInput) Validate() error {
	if err := i.Worker.Validate(); err != nil {
		return err
	}
	if err := i.Entry.Validate(); err != nil {
		return err
	}
	if err := i.Policy.Validate(); err != nil {
		return err
	}
	if strings.TrimSpace(i.PreparedBy) == "" {
		return errors.New("prepared by is required")
	}
	return nil
}

type PayrollLine struct {
	EntryID        int64
	WorkerID       int64
	WorkerName     string
	Trade          string
	Pieces         int
	UnitPriceCents int64
	GrossCents     int64
	NightCents     int64
	DeductionCents int64
	NetCents       int64
	ReviewRequired bool
	ReviewReason   string
	CalculatedAt   time.Time
}

func (l PayrollLine) Validate() error {
	if l.WorkerID <= 0 || l.Pieces <= 0 || l.UnitPriceCents < 0 || l.GrossCents < 0 || l.NetCents < 0 {
		return errors.New("invalid payroll line")
	}
	if l.ReviewRequired && strings.TrimSpace(l.ReviewReason) == "" {
		return errors.New("review reason is required")
	}
	return nil
}

type PayrollStatement struct {
	ID             int64
	StatementNo    string
	WorkerID       int64
	WorkerName     string
	Trade          string
	WorkDate       string
	Lines          []PayrollLine
	GrossCents     int64
	AllowanceCents int64
	DeductionCents int64
	NetCents       int64
	Status         string
	CreatedBy      string
	CreatedAt      time.Time
}

func (s PayrollStatement) Validate() error {
	if strings.TrimSpace(s.StatementNo) == "" || s.WorkerID <= 0 || len(s.Lines) == 0 {
		return errors.New("invalid payroll statement")
	}
	if s.GrossCents < 0 || s.AllowanceCents < 0 || s.DeductionCents < 0 || s.NetCents < 0 {
		return errors.New("statement totals cannot be negative")
	}
	return nil
}

type ImportIssue struct {
	Line    int
	Column  string
	Value   string
	Message string
}

type ImportResult struct {
	Entries []WorkEntry
	Issues  []ImportIssue
	Total   int
}

func (r ImportResult) Accepted() int { return len(r.Entries) }

func (r ImportResult) Rejected() int {
	seen := make(map[int]struct{})
	for _, issue := range r.Issues {
		seen[issue.Line] = struct{}{}
	}
	return len(seen)
}

type ReportTotals struct {
	Statements  int
	GrossCents  int64
	NightCents  int64
	DeductCents int64
	NetCents    int64
	Reviews     int
}

type AuditEvent struct {
	ID         int64
	EntityType string
	EntityID   int64
	Action     string
	Actor      string
	Detail     string
	CreatedAt  time.Time
}

type Setting struct {
	Key       string
	Value     string
	UpdatedAt time.Time
}

type BatchSummary struct {
	BatchID       string
	StatementIDs  []int64
	Imported      int
	Rejected      int
	ReviewCount   int
	TotalNetCents int64
}

func (b BatchSummary) HasErrors() bool { return b.Rejected > 0 }

func (b BatchSummary) HasReviews() bool { return b.ReviewCount > 0 }
