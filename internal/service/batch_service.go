package service

import (
	"context"
	"fmt"
	"io"
	"strings"

	"sitepay/internal/domain"
	"sitepay/internal/importer"
)

type BatchService struct {
	Payroll *PayrollService
	Import  importer.CSVImporter
}

func NewBatchService(payrollService *PayrollService) *BatchService {
	return &BatchService{Payroll: payrollService, Import: importer.NewCSVImporter()}
}

func (b *BatchService) Process(ctx context.Context, reader io.Reader, actor, batchID string) (domain.BatchSummary, importer.ErrorReport, error) {
	if b == nil || b.Payroll == nil {
		return domain.BatchSummary{}, importer.ErrorReport{}, fmt.Errorf("batch service is not configured")
	}
	if strings.TrimSpace(actor) == "" {
		return domain.BatchSummary{}, importer.ErrorReport{}, fmt.Errorf("actor is required")
	}
	result, err := b.Import.Read(reader)
	if err != nil {
		return domain.BatchSummary{}, importer.NewErrorReport(result.Issues), err
	}
	summary := domain.BatchSummary{BatchID: batchID, Imported: result.Accepted(), Rejected: result.Rejected()}
	for index, entry := range result.Entries {
		worker := domain.Worker{Name: entry.WorkerName, Trade: entry.Trade}
		statement, err := b.Payroll.Generate(ctx, worker, entry, actor, fmt.Sprintf("%s-%d", batchID, index+1))
		if err != nil {
			summary.Rejected++
			continue
		}
		summary.StatementIDs = append(summary.StatementIDs, statement.ID)
		summary.TotalNetCents += statement.NetCents
		for _, line := range statement.Lines {
			if line.ReviewRequired {
				summary.ReviewCount++
			}
		}
	}
	_, auditErr := b.Payroll.Store.SaveAudit(ctx, domain.AuditEvent{EntityType: "Batch", EntityID: 0, Action: domain.AuditImported, Actor: actor, Detail: fmt.Sprintf("%s accepted=%d rejected=%d", batchID, summary.Imported, summary.Rejected)})
	if auditErr != nil {
		return summary, importer.NewErrorReport(result.Issues), auditErr
	}
	return summary, importer.NewErrorReport(result.Issues), nil
}

func (b *BatchService) ImportResult(reader io.Reader) (domain.ImportResult, importer.ErrorReport, error) {
	result, err := b.Import.Read(reader)
	return result, importer.NewErrorReport(result.Issues), err
}
