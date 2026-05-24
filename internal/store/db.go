package store

import (
	"database/sql"

	_ "modernc.org/sqlite"
)

// Store is the central data store backed by SQLite.
type Store struct {
	db *sql.DB
}

// New opens the SQLite database at path and runs migrations.
// Use ":memory:" for an in-memory database (tests).
func New(path string) (*Store, error) {
	db, err := sql.Open("sqlite", path+"?_journal_mode=WAL&_foreign_keys=on")
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1) // SQLite: single writer
	s := &Store{db: db}
	if err := s.migrate(); err != nil {
		db.Close()
		return nil, err
	}
	return s, nil
}

// Close closes the database connection.
func (s *Store) Close() error {
	return s.db.Close()
}

// migrate creates tables if they do not exist.
func (s *Store) migrate() error {
	_, err := s.db.Exec(`
		CREATE TABLE IF NOT EXISTS servers (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			host TEXT NOT NULL,
			port INTEGER DEFAULT 22,
			username TEXT NOT NULL,
			auth_method TEXT NOT NULL DEFAULT 'password',
			credential TEXT NOT NULL,
			os TEXT DEFAULT '',
			arch TEXT DEFAULT '',
			status TEXT DEFAULT 'unknown',
			source TEXT DEFAULT 'fresh',
			tags TEXT DEFAULT '[]',
			last_seen DATETIME,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE IF NOT EXISTS server_backends (
			server_id TEXT REFERENCES servers(id) ON DELETE CASCADE,
			backend_type TEXT NOT NULL,
			version TEXT DEFAULT '',
			status TEXT DEFAULT 'stopped',
			config_path TEXT DEFAULT '',
			api_endpoint TEXT DEFAULT '',
			config_managed BOOLEAN DEFAULT 1,
			installed_at DATETIME,
			PRIMARY KEY (server_id, backend_type)
		);

		CREATE TABLE IF NOT EXISTS chains (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			status TEXT DEFAULT 'draft',
			applied_at DATETIME,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE IF NOT EXISTS chain_nodes (
			chain_id TEXT REFERENCES chains(id) ON DELETE CASCADE,
			server_id TEXT REFERENCES servers(id),
			backend_type TEXT NOT NULL,
			protocol TEXT NOT NULL,
			position INTEGER NOT NULL,
			role TEXT NOT NULL,
			inbound_spec TEXT DEFAULT '{}',
			outbound_spec TEXT DEFAULT '{}',
			inbound_result TEXT DEFAULT '{}',
			outbound_result TEXT DEFAULT '{}',
			PRIMARY KEY (chain_id, position)
		);
	`)
	return err
}
