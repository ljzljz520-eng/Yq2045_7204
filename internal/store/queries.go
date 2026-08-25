package store

import (
	"context"
	"database/sql"
	"fmt"

	"sitepay/internal/domain"
)

type rowScanner interface{ Scan(...any) error }

func scanWorker(row rowScanner) (domain.Worker, error) {
	var worker domain.Worker
	var created string
	if err := row.Scan(&worker.ID, &worker.Name, &worker.Trade, &created); err != nil {
		if err == sql.ErrNoRows {
			return domain.Worker{}, domain.ErrNotFound
		}
		return domain.Worker{}, err
	}
	var err error
	worker.CreatedAt, err = decodeTime(created)
	return worker, err
}

func (s *Store) listLines(ctx context.Context, statementID int64) ([]domain.PayrollLine, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT entry_id, worker_id, worker_name, trade, pieces, unit_price_cents, gross_cents, night_cents, deduction_cents, net_cents, review_required, review_reason, calculated_at FROM payroll_lines WHERE statement_id=? ORDER BY id`, statementID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	lines := make([]domain.PayrollLine, 0)
	for rows.Next() {
		var line domain.PayrollLine
		var review int64
		var calculated string
		if err := rows.Scan(&line.EntryID, &line.WorkerID, &line.WorkerName, &line.Trade, &line.Pieces, &line.UnitPriceCents, &line.GrossCents, &line.NightCents, &line.DeductionCents, &line.NetCents, &review, &line.ReviewReason, &calculated); err != nil {
			return nil, err
		}
		line.ReviewRequired = intBool(review)
		line.CalculatedAt, err = decodeTime(calculated)
		if err != nil {
			return nil, err
		}
		lines = append(lines, line)
	}
	return lines, rows.Err()
}

func (s *Store) DeleteStatement(ctx context.Context, id int64) error {
	ctx = contextOrBackground(ctx)
	return s.WithTx(ctx, func(tx *sql.Tx) error {
		if _, err := tx.ExecContext(ctx, `DELETE FROM payroll_lines WHERE statement_id=?`, id); err != nil {
			return err
		}
		result, err := tx.ExecContext(ctx, `DELETE FROM payroll_statements WHERE id=?`, id)
		if err == nil {
			count, _ := result.RowsAffected()
			if count == 0 {
				return domain.ErrNotFound
			}
		}
		return err
	})
}

func (s *Store) UpdateStatementStatus(ctx context.Context, id int64, status string) error {
	ctx = contextOrBackground(ctx)
	if !domain.ValidStatementStatus(status) {
		return fmt.Errorf("invalid statement status %q", status)
	}
	result, err := s.db.ExecContext(ctx, `UPDATE payroll_statements SET status=? WHERE id=?`, status, id)
	if err != nil {
		return err
	}
	count, err := result.RowsAffected()
	if err == nil && count == 0 {
		return domain.ErrNotFound
	}
	return err
}

func (s *Store) Count(ctx context.Context, table string) (int64, error) {
	ctx = contextOrBackground(ctx)
	allowed := map[string]bool{"workers": true, "work_entries": true, "payroll_statements": true, "audit_events": true}
	if !allowed[table] {
		return 0, fmt.Errorf("table %q is not countable", table)
	}
	var count int64
	err := s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM "+table).Scan(&count)
	return count, err
}
