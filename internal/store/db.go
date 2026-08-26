// Package store 负责领域实体的 SQLite 持久化。使用纯 Go 驱动 modernc.org/sqlite，
// 无需 CGO，可在 GOTOOLCHAIN=local + CGO_ENABLED=0 环境下离线构建。
// 所有写操作基于 database/sql，并对关键外键/状态字段做约束，保证重启后可恢复。
package store

import (
	"database/sql"

	_ "modernc.org/sqlite"
)

// Store 封装数据库连接与全部实体的读写方法。
type Store struct {
	db *sql.DB
}

// schema 为建表语句，幂等（IF NOT EXISTS）。
const schema = `
CREATE TABLE IF NOT EXISTS figures (
  id          TEXT PRIMARY KEY,
  title       TEXT NOT NULL,
  status      TEXT NOT NULL,
  source_fp   TEXT NOT NULL DEFAULT '',
  layer_count INTEGER NOT NULL DEFAULT 0,
  created_at  INTEGER NOT NULL,
  updated_at  INTEGER NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_figures_status ON figures(status);

CREATE TABLE IF NOT EXISTS layers (
  id         TEXT PRIMARY KEY,
  figure_id  TEXT NOT NULL,
  name       TEXT NOT NULL,
  layer_type TEXT NOT NULL,
  z_order    INTEGER NOT NULL,
  visible    INTEGER NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_layers_fig ON layers(figure_id);

CREATE TABLE IF NOT EXISTS axes (
  id          TEXT PRIMARY KEY,
  figure_id   TEXT NOT NULL,
  name        TEXT NOT NULL,
  variable    TEXT NOT NULL,
  unit        TEXT NOT NULL DEFAULT '',
  orientation TEXT NOT NULL DEFAULT 'none'
);
CREATE INDEX IF NOT EXISTS idx_axes_fig ON axes(figure_id);

CREATE TABLE IF NOT EXISTS variables (
  id          TEXT PRIMARY KEY,
  figure_id   TEXT NOT NULL,
  name        TEXT NOT NULL,
  unit        TEXT NOT NULL DEFAULT '',
  description TEXT NOT NULL DEFAULT ''
);
CREATE INDEX IF NOT EXISTS idx_variables_fig ON variables(figure_id);

CREATE TABLE IF NOT EXISTS legends (
  id               TEXT PRIMARY KEY,
  figure_id        TEXT NOT NULL,
  channel          TEXT NOT NULL,
  label            TEXT NOT NULL,
  token            TEXT NOT NULL,
  covers_variable  TEXT NOT NULL DEFAULT ''
);
CREATE INDEX IF NOT EXISTS idx_legends_fig ON legends(figure_id);

CREATE TABLE IF NOT EXISTS encodings (
  id         TEXT PRIMARY KEY,
  figure_id  TEXT NOT NULL,
  layer_id   TEXT NOT NULL DEFAULT '',
  variable   TEXT NOT NULL,
  channel    TEXT NOT NULL,
  token      TEXT NOT NULL,
  status     TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_encodings_fig ON encodings(figure_id);

CREATE TABLE IF NOT EXISTS mappings (
  id         TEXT PRIMARY KEY,
  figure_id  TEXT NOT NULL,
  variable   TEXT NOT NULL,
  channel    TEXT NOT NULL,
  decision   TEXT NOT NULL,
  note       TEXT NOT NULL DEFAULT ''
);
CREATE INDEX IF NOT EXISTS idx_mappings_fig ON mappings(figure_id);

CREATE TABLE IF NOT EXISTS exceptions (
  id               TEXT PRIMARY KEY,
  figure_id        TEXT NOT NULL,
  kind             TEXT NOT NULL,
  target_channel   TEXT NOT NULL DEFAULT '',
  target_token     TEXT NOT NULL DEFAULT '',
  target_variable  TEXT NOT NULL DEFAULT '',
  reason           TEXT NOT NULL DEFAULT '',
  created_at       INTEGER NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_exceptions_fig ON exceptions(figure_id);

CREATE TABLE IF NOT EXISTS ambiguities (
  id          TEXT PRIMARY KEY,
  figure_id   TEXT NOT NULL,
  type        TEXT NOT NULL,
  severity    TEXT NOT NULL,
  channel     TEXT NOT NULL DEFAULT '',
  token       TEXT NOT NULL DEFAULT '',
  variables   TEXT NOT NULL DEFAULT '',
  description TEXT NOT NULL DEFAULT '',
  resolved    INTEGER NOT NULL DEFAULT 0,
  exception_id TEXT NOT NULL DEFAULT '',
  created_at  INTEGER NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_ambiguities_fig ON ambiguities(figure_id);

CREATE TABLE IF NOT EXISTS specs (
  id         TEXT PRIMARY KEY,
  figure_id  TEXT NOT NULL,
  version    INTEGER NOT NULL,
  status     TEXT NOT NULL,
  snapshot   TEXT NOT NULL DEFAULT '',
  created_at INTEGER NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_specs_fig ON specs(figure_id);
`

// Open 打开（或创建）位于 path 的 SQLite 数据库，并应用 schema。
// path 可为 ":memory:"（仅用于测试），或磁盘文件路径（支持重启恢复）。
func Open(path string) (*Store, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	// 开启外键与 WAL，提升一致性与并发写入稳定性。
	if _, err := db.Exec("PRAGMA foreign_keys=ON;"); err != nil {
		_ = db.Close()
		return nil, err
	}
	if _, err := db.Exec("PRAGMA journal_mode=WAL;"); err != nil {
		_ = db.Close()
		return nil, err
	}
	if _, err := db.Exec(schema); err != nil {
		_ = db.Close()
		return nil, err
	}
	return &Store{db: db}, nil
}

// DB 暴露底层连接，供需要在事务中组合多实体操作的调用方使用。
func (s *Store) DB() *sql.DB { return s.db }

// Close 关闭数据库连接。
func (s *Store) Close() error { return s.db.Close() }
