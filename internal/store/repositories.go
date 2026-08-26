package store

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"sitepay/internal/domain"
)

func (s *Store) SaveWorker(ctx context.Context, worker domain.Worker) (domain.Worker, error) {
	ctx = contextOrBackground(ctx)
	if err := worker.Validate(); err != nil {
		return domain.Worker{}, err
	}
	if worker.CreatedAt.IsZero() {
		worker.CreatedAt = time.Unix(0, 0).UTC()
	}
	err := s.WithTx(ctx, func(tx *sql.Tx) error {
		id, err := ensureWorkerTx(ctx, tx, worker)
		worker.ID = id
		return err
	})
	return worker, err
}

func (s *Store) FindWorker(ctx context.Context, name, trade string) (domain.Worker, error) {
	ctx = contextOrBackground(ctx)
	row := s.db.QueryRowContext(ctx, `SELECT id, name, trade, created_at FROM workers WHERE name=? AND trade=?`, strings.TrimSpace(name), strings.TrimSpace(trade))
	return scanWorker(row)
}

func (s *Store) SavePolicy(ctx context.Context, policy domain.AllowancePolicy) (domain.AllowancePolicy, error) {
	ctx = contextOrBackground(ctx)
	if err := policy.Validate(); err != nil {
		return domain.AllowancePolicy{}, err
	}
	result, err := s.db.ExecContext(ctx, `INSERT INTO allowance_policies(trade, night_cap, requires_review, active) VALUES(?, ?, ?, ?) ON CONFLICT(trade) DO UPDATE SET night_cap=excluded.night_cap, requires_review=excluded.requires_review, active=excluded.active`, policy.Trade, policy.NightCap, boolInt(policy.RequiresReview), boolInt(policy.Active))
	if err != nil {
		return domain.AllowancePolicy{}, err
	}
	policy.ID, _ = result.LastInsertId()
	if policy.ID == 0 {
		_ = s.db.QueryRowContext(ctx, `SELECT id FROM allowance_policies WHERE trade=?`, policy.Trade).Scan(&policy.ID)
	}
	return policy, nil
}

func (s *Store) ListPolicies(ctx context.Context) ([]domain.AllowancePolicy, error) {
	ctx = contextOrBackground(ctx)
	rows, err := s.db.QueryContext(ctx, `SELECT id, trade, night_cap, requires_review, active FROM allowance_policies ORDER BY trade`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	policies := make([]domain.AllowancePolicy, 0)
	for rows.Next() {
		var policy domain.AllowancePolicy
		var review, active int64
		if err := rows.Scan(&policy.ID, &policy.Trade, &policy.NightCap, &review, &active); err != nil {
			return nil, err
		}
		policy.RequiresReview, policy.Active = intBool(review), intBool(active)
		policies = append(policies, policy)
	}
	return policies, rows.Err()
}

func (s *Store) SaveEntry(ctx context.Context, workerID int64, entry domain.WorkEntry) (domain.WorkEntry, error) {
	ctx = contextOrBackground(ctx)
	if err := entry.Validate(); err != nil {
		return domain.WorkEntry{}, err
	}
	if workerID <= 0 {
		return domain.WorkEntry{}, fmt.Errorf("worker id is required")
	}
	result, err := s.db.ExecContext(ctx, `INSERT INTO work_entries(worker_id, completed_pieces, unit_price, night_allowance, quality_deduction, work_date, imported_line) VALUES(?, ?, ?, ?, ?, ?, ?)`, workerID, entry.CompletedPieces, entry.UnitPrice, entry.NightAllowance, entry.QualityDeduction, entry.WorkDate, entry.ImportedLine)
	if err != nil {
		return domain.WorkEntry{}, err
	}
	entry.ID, err = result.LastInsertId()
	entry.WorkerID = workerID
	return entry, err
}

func (s *Store) SaveStatement(ctx context.Context, statement domain.PayrollStatement) (domain.PayrollStatement, error) {
	ctx = contextOrBackground(ctx)
	if err := statement.Validate(); err != nil {
		return domain.PayrollStatement{}, err
	}
	if statement.CreatedAt.IsZero() {
		statement.CreatedAt = time.Unix(0, 0).UTC()
	}
	var id int64
	err := s.WithTx(ctx, func(tx *sql.Tx) error {
		result, err := tx.ExecContext(ctx, `INSERT INTO payroll_statements(statement_no, worker_id, worker_name, trade, work_date, gross_cents, allowance_cents, deduction_cents, net_cents, status, created_by, created_at) VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, statement.StatementNo, statement.WorkerID, statement.WorkerName, statement.Trade, statement.WorkDate, statement.GrossCents, statement.AllowanceCents, statement.DeductionCents, statement.NetCents, statement.Status, statement.CreatedBy, encodeTime(statement.CreatedAt))
		if err != nil {
			return err
		}
		id, err = result.LastInsertId()
		for _, line := range statement.Lines {
			if err != nil {
				return err
			}
			_, err = tx.ExecContext(ctx, `INSERT INTO payroll_lines(statement_id, entry_id, worker_id, worker_name, trade, pieces, unit_price_cents, gross_cents, night_cents, deduction_cents, net_cents, review_required, review_reason, calculated_at) VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, id, line.EntryID, line.WorkerID, line.WorkerName, line.Trade, line.Pieces, line.UnitPriceCents, line.GrossCents, line.NightCents, line.DeductionCents, line.NetCents, boolInt(line.ReviewRequired), line.ReviewReason, encodeTime(line.CalculatedAt))
		}
		return err
	})
	statement.ID = id
	return statement, err
}

func (s *Store) GetStatement(ctx context.Context, id int64) (domain.PayrollStatement, error) {
	ctx = contextOrBackground(ctx)
	row := s.db.QueryRowContext(ctx, `SELECT id, statement_no, worker_id, worker_name, trade, work_date, gross_cents, allowance_cents, deduction_cents, net_cents, status, created_by, created_at FROM payroll_statements WHERE id=?`, id)
	var statement domain.PayrollStatement
	var created string
	if err := row.Scan(&statement.ID, &statement.StatementNo, &statement.WorkerID, &statement.WorkerName, &statement.Trade, &statement.WorkDate, &statement.GrossCents, &statement.AllowanceCents, &statement.DeductionCents, &statement.NetCents, &statement.Status, &statement.CreatedBy, &created); err != nil {
		if err == sql.ErrNoRows {
			return domain.PayrollStatement{}, domain.ErrNotFound
		}
		return domain.PayrollStatement{}, err
	}
	var err error
	statement.CreatedAt, err = decodeTime(created)
	if err != nil {
		return domain.PayrollStatement{}, err
	}
	statement.Lines, err = s.listLines(ctx, statement.ID)
	return statement, err
}

func (s *Store) ListStatements(ctx context.Context, status string) ([]domain.PayrollStatement, error) {
	ctx = contextOrBackground(ctx)
	query := `SELECT id FROM payroll_statements ORDER BY work_date, statement_no`
	args := []any{}
	if strings.TrimSpace(status) != "" {
		query = `SELECT id FROM payroll_statements WHERE status=? ORDER BY work_date, statement_no`
		args = append(args, status)
	}
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]domain.PayrollStatement, 0)
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		statement, err := s.GetStatement(ctx, id)
		if err != nil {
			return nil, err
		}
		result = append(result, statement)
	}
	return result, rows.Err()
}

func (s *Store) SaveAudit(ctx context.Context, event domain.AuditEvent) (domain.AuditEvent, error) {
	ctx = contextOrBackground(ctx)
	if strings.TrimSpace(event.EntityType) == "" || strings.TrimSpace(event.Action) == "" || strings.TrimSpace(event.Actor) == "" {
		return domain.AuditEvent{}, fmt.Errorf("audit entity, action, and actor are required")
	}
	if event.CreatedAt.IsZero() {
		event.CreatedAt = time.Unix(0, 0).UTC()
	}
	result, err := s.db.ExecContext(ctx, `INSERT INTO audit_events(entity_type, entity_id, action, actor, detail, created_at) VALUES(?, ?, ?, ?, ?, ?)`, event.EntityType, event.EntityID, event.Action, event.Actor, event.Detail, encodeTime(event.CreatedAt))
	if err != nil {
		return domain.AuditEvent{}, err
	}
	event.ID, err = result.LastInsertId()
	return event, err
}

func (s *Store) ListAudits(ctx context.Context, entityType string, entityID int64) ([]domain.AuditEvent, error) {
	ctx = contextOrBackground(ctx)
	rows, err := s.db.QueryContext(ctx, `SELECT id, entity_type, entity_id, action, actor, detail, created_at FROM audit_events WHERE entity_type=? AND entity_id=? ORDER BY id`, entityType, entityID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]domain.AuditEvent, 0)
	for rows.Next() {
		var event domain.AuditEvent
		var created string
		if err := rows.Scan(&event.ID, &event.EntityType, &event.EntityID, &event.Action, &event.Actor, &event.Detail, &created); err != nil {
			return nil, err
		}
		event.CreatedAt, err = decodeTime(created)
		if err != nil {
			return nil, err
		}
		result = append(result, event)
	}
	return result, rows.Err()
}

func (s *Store) SaveSetting(ctx context.Context, setting domain.Setting) error {
	ctx = contextOrBackground(ctx)
	if strings.TrimSpace(setting.Key) == "" {
		return fmt.Errorf("setting key is required")
	}
	if setting.UpdatedAt.IsZero() {
		setting.UpdatedAt = time.Unix(0, 0).UTC()
	}
	_, err := s.db.ExecContext(ctx, `INSERT INTO settings(key, value, updated_at) VALUES(?, ?, ?) ON CONFLICT(key) DO UPDATE SET value=excluded.value, updated_at=excluded.updated_at`, setting.Key, setting.Value, encodeTime(setting.UpdatedAt))
	return err
}

func (s *Store) GetSetting(ctx context.Context, key string) (domain.Setting, error) {
	ctx = contextOrBackground(ctx)
	var setting domain.Setting
	var updated string
	err := s.db.QueryRowContext(ctx, `SELECT key, value, updated_at FROM settings WHERE key=?`, key).Scan(&setting.Key, &setting.Value, &updated)
	if err == sql.ErrNoRows {
		return domain.Setting{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Setting{}, err
	}
	setting.UpdatedAt, err = decodeTime(updated)
	return setting, err
}
