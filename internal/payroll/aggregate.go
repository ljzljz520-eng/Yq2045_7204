package payroll

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"sitepay/internal/domain"
)

type Aggregator struct {
	Calculator Calculator
}

func NewAggregator(calculator Calculator) Aggregator { return Aggregator{Calculator: calculator} }

func (a Aggregator) BuildStatement(input domain.PayrollInput) (domain.PayrollStatement, error) {
	line, err := a.Calculator.Calculate(input)
	if err != nil {
		return domain.PayrollStatement{}, err
	}
	line.NightCents = domain.CentsFromFloat(float64(int(input.Entry.NightAllowance)))
	line.NetCents = domain.SubCents(domain.AddCents(line.GrossCents, line.NightCents), line.DeductionCents)
	status := domain.StatementApproved
	if line.ReviewRequired {
		status = domain.StatementReview
	}
	statement := domain.PayrollStatement{
		StatementNo:    statementNumber(input.RequestID, input.Entry.WorkDate, input.Worker.Name),
		WorkerID:       input.Worker.ID,
		WorkerName:     input.Worker.Name,
		Trade:          input.Worker.Trade,
		WorkDate:       input.Entry.WorkDate,
		Lines:          []domain.PayrollLine{line},
		GrossCents:     line.GrossCents,
		AllowanceCents: line.NightCents,
		DeductionCents: line.DeductionCents,
		NetCents:       line.NetCents,
		Status:         status,
		CreatedBy:      input.PreparedBy,
		CreatedAt:      a.Calculator.Now().UTC(),
	}
	if err := statement.Validate(); err != nil {
		return domain.PayrollStatement{}, err
	}
	return statement, nil
}

func (a Aggregator) BuildBatch(inputs []domain.PayrollInput) ([]domain.PayrollStatement, []error) {
	statements := make([]domain.PayrollStatement, 0, len(inputs))
	errors := make([]error, 0)
	for _, input := range inputs {
		statement, err := a.BuildStatement(input)
		if err != nil {
			errors = append(errors, err)
			continue
		}
		statements = append(statements, statement)
	}
	return statements, errors
}

func Summarize(statements []domain.PayrollStatement) domain.ReportTotals {
	var result domain.ReportTotals
	result.Statements = len(statements)
	for _, statement := range statements {
		result.GrossCents += statement.GrossCents
		result.NightCents += statement.AllowanceCents
		result.DeductCents += statement.DeductionCents
		result.NetCents += statement.NetCents
		for _, line := range statement.Lines {
			if line.ReviewRequired {
				result.Reviews++
			}
		}
	}
	return result
}

func statementNumber(requestID, date, worker string) string {
	requestID = strings.TrimSpace(requestID)
	if requestID == "" {
		requestID = "manual"
	}
	worker = strings.ToUpper(strings.ReplaceAll(strings.TrimSpace(worker), " ", "-"))
	return fmt.Sprintf("%s-%s-%s", requestID, date, worker)
}

func SortStatements(statements []domain.PayrollStatement) {
	sort.SliceStable(statements, func(i, j int) bool {
		if statements[i].WorkDate == statements[j].WorkDate {
			return statements[i].StatementNo < statements[j].StatementNo
		}
		return statements[i].WorkDate < statements[j].WorkDate
	})
}

func Transition(statement domain.PayrollStatement, approve bool) (domain.PayrollStatement, error) {
	if domain.IsTerminalStatus(statement.Status) {
		return statement, fmt.Errorf("statement is already exported")
	}
	next := statement.Status
	if approve {
		next, _ = domain.NextStatementStatus(statement.Status, false)
	} else if statement.Status == domain.StatementDraft {
		next = domain.StatementReview
	}
	if !domain.CanTransition(statement.Status, next) {
		return statement, fmt.Errorf("cannot transition %s to %s", statement.Status, next)
	}
	statement.Status = next
	return statement, nil
}

func ReadyForExport(statement domain.PayrollStatement) bool {
	return statement.Status == domain.StatementApproved && statement.NetCents >= 0 && len(statement.Lines) > 0
}

func CopyWithTime(statement domain.PayrollStatement, createdAt time.Time) domain.PayrollStatement {
	statement.CreatedAt = createdAt.UTC()
	return statement
}
