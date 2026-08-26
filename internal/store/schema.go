package store

const schema = `
CREATE TABLE IF NOT EXISTS workers (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  name TEXT NOT NULL,
  trade TEXT NOT NULL,
  created_at TEXT NOT NULL
);
CREATE UNIQUE INDEX IF NOT EXISTS workers_name_trade ON workers(name, trade);
CREATE TABLE IF NOT EXISTS allowance_policies (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  trade TEXT NOT NULL UNIQUE,
  night_cap REAL NOT NULL,
  requires_review INTEGER NOT NULL,
  active INTEGER NOT NULL
);
CREATE TABLE IF NOT EXISTS work_entries (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  worker_id INTEGER NOT NULL,
  completed_pieces INTEGER NOT NULL,
  unit_price REAL NOT NULL,
  night_allowance REAL NOT NULL,
  quality_deduction REAL NOT NULL,
  work_date TEXT NOT NULL,
  imported_line INTEGER NOT NULL DEFAULT 0,
  FOREIGN KEY(worker_id) REFERENCES workers(id)
);
CREATE TABLE IF NOT EXISTS payroll_statements (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  statement_no TEXT NOT NULL UNIQUE,
  worker_id INTEGER NOT NULL,
  worker_name TEXT NOT NULL,
  trade TEXT NOT NULL,
  work_date TEXT NOT NULL,
  gross_cents INTEGER NOT NULL,
  allowance_cents INTEGER NOT NULL,
  deduction_cents INTEGER NOT NULL,
  net_cents INTEGER NOT NULL,
  status TEXT NOT NULL,
  created_by TEXT NOT NULL,
  created_at TEXT NOT NULL,
  FOREIGN KEY(worker_id) REFERENCES workers(id)
);
CREATE TABLE IF NOT EXISTS payroll_lines (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  statement_id INTEGER NOT NULL,
  entry_id INTEGER NOT NULL,
  worker_id INTEGER NOT NULL,
  worker_name TEXT NOT NULL,
  trade TEXT NOT NULL,
  pieces INTEGER NOT NULL,
  unit_price_cents INTEGER NOT NULL,
  gross_cents INTEGER NOT NULL,
  night_cents INTEGER NOT NULL,
  deduction_cents INTEGER NOT NULL,
  net_cents INTEGER NOT NULL,
  review_required INTEGER NOT NULL,
  review_reason TEXT NOT NULL,
  calculated_at TEXT NOT NULL,
  FOREIGN KEY(statement_id) REFERENCES payroll_statements(id)
);
CREATE TABLE IF NOT EXISTS audit_events (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  entity_type TEXT NOT NULL,
  entity_id INTEGER NOT NULL,
  action TEXT NOT NULL,
  actor TEXT NOT NULL,
  detail TEXT NOT NULL,
  created_at TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS settings (
  key TEXT PRIMARY KEY,
  value TEXT NOT NULL,
  updated_at TEXT NOT NULL
);`
