package store

import (
	"context"
	"database/sql"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	_ "modernc.org/sqlite"

	"sitepay/internal/domain"
)

type Store struct {
	db   *sql.DB
	path string
}

func Open(path string) (*Store, error) {
	if strings.TrimSpace(path) == "" {
		return nil, fmt.Errorf("database path is required")
	}
	if path != ":memory:" {
		path = filepath.Clean(path)
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	store := &Store{db: db, path: path}
	if err := store.configure(); err != nil {
		db.Close()
		return nil, err
	}
	if err := store.migrate(context.Background()); err != nil {
		db.Close()
		return nil, err
	}
	return store, nil
}

func (s *Store) configure() error {
	if s == nil || s.db == nil {
		return fmt.Errorf("store is not initialized")
	}
	if _, err := s.db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		return fmt.Errorf("configure sqlite: %w", err)
	}
	return nil
}

func (s *Store) migrate(ctx context.Context) error {
	if s == nil || s.db == nil {
		return fmt.Errorf("store is not initialized")
	}
	if _, err := s.db.ExecContext(ctx, schema); err != nil {
		return fmt.Errorf("migrate schema: %w", err)
	}
	return nil
}

func (s *Store) Close() error {
	if s == nil || s.db == nil {
		return nil
	}
	err := s.db.Close()
	s.db = nil
	return err
}

func (s *Store) Reopen() (*Store, error) {
	if s == nil {
		return nil, fmt.Errorf("store is nil")
	}
	path := s.path
	if err := s.Close(); err != nil {
		return nil, err
	}
	return Open(path)
}

func (s *Store) Path() string {
	if s == nil {
		return ""
	}
	return s.path
}

func (s *Store) Ping(ctx context.Context) error {
	if s == nil || s.db == nil {
		return fmt.Errorf("store is closed")
	}
	return s.db.PingContext(ctx)
}

func (s *Store) WithTx(ctx context.Context, fn func(*sql.Tx) error) error {
	if s == nil || s.db == nil {
		return fmt.Errorf("store is closed")
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	if err := fn(tx); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}

func encodeTime(value time.Time) string { return value.UTC().Format(time.RFC3339Nano) }

func decodeTime(value string) (time.Time, error) {
	result, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		return time.Time{}, fmt.Errorf("parse time %q: %w", value, err)
	}
	return result.UTC(), nil
}

func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}

func intBool(value int64) bool { return value != 0 }

func contextOrBackground(ctx context.Context) context.Context {
	if ctx == nil {
		return context.Background()
	}
	return ctx
}

func ensureWorkerTx(ctx context.Context, tx *sql.Tx, worker domain.Worker) (int64, error) {
	if err := worker.Validate(); err != nil {
		return 0, err
	}
	created := worker.CreatedAt
	if created.IsZero() {
		created = time.Unix(0, 0).UTC()
	}
	result, err := tx.ExecContext(ctx, `INSERT INTO workers(name, trade, created_at) VALUES(?, ?, ?) ON CONFLICT(name, trade) DO UPDATE SET name=excluded.name`, worker.Name, worker.Trade, encodeTime(created))
	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()
	if err != nil || id == 0 {
		row := tx.QueryRowContext(ctx, `SELECT id FROM workers WHERE name=? AND trade=?`, worker.Name, worker.Trade)
		if scanErr := row.Scan(&id); scanErr != nil {
			return 0, scanErr
		}
	}
	return id, nil
}
