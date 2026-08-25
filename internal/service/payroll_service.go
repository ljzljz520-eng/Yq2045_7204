package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"sitepay/internal/domain"
	"sitepay/internal/payroll"
	"sitepay/internal/store"
	"sitepay/internal/validation"
)

type PayrollService struct {
	Store     *store.Store
	Validator validation.Validator
	Book      *payroll.PolicyBook
	Aggregate payroll.Aggregator
}

func NewPayrollService(db *store.Store, policies []domain.AllowancePolicy) (*PayrollService, error) {
	if db == nil {
		return nil, fmt.Errorf("store is required")
	}
	book, err := payroll.NewPolicyBook(policies)
	if err != nil {
		return nil, err
	}
	calculator := payroll.NewCalculator("site").WithClock(func() time.Time { return time.Unix(0, 0).UTC() })
	return &PayrollService{Store: db, Validator: validation.New(), Book: book, Aggregate: payroll.NewAggregator(calculator)}, nil
}

func (s *PayrollService) Generate(ctx context.Context, worker domain.Worker, entry domain.WorkEntry, actor, requestID string) (domain.PayrollStatement, error) {
	ctx = contextOrBackground(ctx)
	if err := s.Validator.ValidateWorker(worker); err != nil {
		return domain.PayrollStatement{}, err
	}
	if err := s.Validator.ValidateEntry(entry); err != nil {
		return domain.PayrollStatement{}, err
	}
	if strings.TrimSpace(actor) == "" {
		return domain.PayrollStatement{}, fmt.Errorf("actor is required")
	}
	policy, err := s.Book.Resolve(worker.Trade)
	if err != nil {
		return domain.PayrollStatement{}, err
	}
	worker, err = s.Store.SaveWorker(ctx, worker)
	if err != nil {
		return domain.PayrollStatement{}, err
	}
	entry, err = s.Store.SaveEntry(ctx, worker.ID, entry)
	if err != nil {
		return domain.PayrollStatement{}, err
	}
	input := domain.PayrollInput{Worker: worker, Entry: entry, Policy: policy, PreparedBy: actor, RequestID: requestID}
	statement, err := s.Aggregate.BuildStatement(input)
	if err != nil {
		return domain.PayrollStatement{}, err
	}
	statement, err = s.Store.SaveStatement(ctx, statement)
	if err != nil {
		return domain.PayrollStatement{}, err
	}
	_, err = s.Store.SaveAudit(ctx, domain.AuditEvent{EntityType: "PayrollStatement", EntityID: statement.ID, Action: domain.AuditCreated, Actor: actor, Detail: statement.StatementNo})
	return statement, err
}

func (s *PayrollService) Approve(ctx context.Context, id int64, actor string) (domain.PayrollStatement, error) {
	ctx = contextOrBackground(ctx)
	if strings.TrimSpace(actor) == "" {
		return domain.PayrollStatement{}, fmt.Errorf("actor is required")
	}
	statement, err := s.Store.GetStatement(ctx, id)
	if err != nil {
		return domain.PayrollStatement{}, err
	}
	updated, err := payroll.Transition(statement, true)
	if err != nil {
		return domain.PayrollStatement{}, err
	}
	if err := s.Store.UpdateStatementStatus(ctx, id, updated.Status); err != nil {
		return domain.PayrollStatement{}, err
	}
	_, err = s.Store.SaveAudit(ctx, domain.AuditEvent{EntityType: "PayrollStatement", EntityID: id, Action: domain.AuditApproved, Actor: actor, Detail: updated.Status})
	return updated, err
}

func (s *PayrollService) List(ctx context.Context, status string) ([]domain.PayrollStatement, error) {
	return s.Store.ListStatements(contextOrBackground(ctx), status)
}

func (s *PayrollService) SeedPolicies(ctx context.Context) error {
	for _, policy := range payroll.DefaultPolicies() {
		if _, err := s.Store.SavePolicy(contextOrBackground(ctx), policy); err != nil {
			return err
		}
	}
	return nil
}

func (s *PayrollService) PolicyBook() *payroll.PolicyBook { return s.Book }

func contextOrBackground(ctx context.Context) context.Context {
	if ctx == nil {
		return context.Background()
	}
	return ctx
}
