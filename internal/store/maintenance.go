package store

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

type HealthSnapshot struct {
	Path          string
	Workers       int64
	Entries       int64
	Statements    int64
	Audits        int64
	DatabaseReady bool
	CheckedAt     time.Time
}

func (s *Store) Health(ctx context.Context) (HealthSnapshot, error) {
	ctx = contextOrBackground(ctx)
	if err := s.Ping(ctx); err != nil {
		return HealthSnapshot{}, err
	}
	snapshot := HealthSnapshot{Path: s.Path(), DatabaseReady: true, CheckedAt: time.Unix(0, 0).UTC()}
	var err error
	if snapshot.Workers, err = s.Count(ctx, "workers"); err != nil {
		return HealthSnapshot{}, err
	}
	if snapshot.Entries, err = s.Count(ctx, "work_entries"); err != nil {
		return HealthSnapshot{}, err
	}
	if snapshot.Statements, err = s.Count(ctx, "payroll_statements"); err != nil {
		return HealthSnapshot{}, err
	}
	if snapshot.Audits, err = s.Count(ctx, "audit_events"); err != nil {
		return HealthSnapshot{}, err
	}
	return snapshot, nil
}

func (s *Store) PurgeAudits(ctx context.Context, before time.Time) (int64, error) {
	ctx = contextOrBackground(ctx)
	if before.IsZero() {
		return 0, fmt.Errorf("purge cutoff is required")
	}
	result, err := s.db.ExecContext(ctx, `DELETE FROM audit_events WHERE created_at < ?`, encodeTime(before))
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func (s *Store) Vacuum(ctx context.Context) error {
	ctx = contextOrBackground(ctx)
	if s == nil || s.db == nil {
		return fmt.Errorf("store is closed")
	}
	_, err := s.db.ExecContext(ctx, "VACUUM")
	return err
}

func (s *Store) StatementTotals(ctx context.Context) (int64, int64, int64, error) {
	ctx = contextOrBackground(ctx)
	var gross, allowance, net sql.NullInt64
	err := s.db.QueryRowContext(ctx, `SELECT SUM(gross_cents), SUM(allowance_cents), SUM(net_cents) FROM payroll_statements`).Scan(&gross, &allowance, &net)
	if err != nil {
		return 0, 0, 0, err
	}
	return nullableInt(gross), nullableInt(allowance), nullableInt(net), nil
}

func nullableInt(value sql.NullInt64) int64 {
	if !value.Valid {
		return 0
	}
	return value.Int64
}

func (s *Store) SearchWorkers(ctx context.Context, query string) ([]string, error) {
	ctx = contextOrBackground(ctx)
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, fmt.Errorf("worker search query is required")
	}
	rows, err := s.db.QueryContext(ctx, `SELECT name || ' / ' || trade FROM workers WHERE name LIKE ? OR trade LIKE ? ORDER BY name, trade`, "%"+query+"%", "%"+query+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]string, 0)
	for rows.Next() {
		var value string
		if err := rows.Scan(&value); err != nil {
			return nil, err
		}
		result = append(result, value)
	}
	return result, rows.Err()
}
