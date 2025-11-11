package config

import (
	"database/sql"
	"fmt"
	"log"

	_ "modernc.org/sqlite"
)

// Database wraps the SQLite database connection
type Database struct {
	DB *sql.DB
}

// NewDatabase initializes and returns a new database connection
func NewDatabase() (*Database, error) {
	// Open SQLite database
	db, err := sql.Open("sqlite", "config.db")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Initialize schema
	if err := initSchema(db); err != nil {
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	log.Println("Database initialized successfully")
	return &Database{DB: db}, nil
}

// Close closes the database connection
func (d *Database) Close() error {
	if d.DB != nil {
		return d.DB.Close()
	}
	return nil
}

// initSchema creates database tables
func initSchema(db *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS users (
		id TEXT PRIMARY KEY,
		username TEXT UNIQUE NOT NULL,
		email TEXT UNIQUE NOT NULL,
		password_hash TEXT NOT NULL,
		role TEXT NOT NULL DEFAULT 'user',
		beta_code TEXT,
		is_active BOOLEAN DEFAULT 1,
		last_login DATETIME,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS beta_codes (
		code TEXT PRIMARY KEY,
		description TEXT,
		max_uses INTEGER DEFAULT 1,
		current_uses INTEGER DEFAULT 0,
		is_active BOOLEAN DEFAULT 1,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		expires_at DATETIME,
		created_by TEXT
	);

	CREATE TABLE IF NOT EXISTS user_configs (
		user_id TEXT PRIMARY KEY,
		ai_provider TEXT DEFAULT 'openai',
		ai_model TEXT DEFAULT 'gpt-4',
		ai_base_url TEXT,
		ai_api_key TEXT,
		default_system_prompt TEXT,
		default_strategy_prompt TEXT,
		exchange_configs TEXT,
		risk_limits TEXT,
		notification_settings TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS traders (
		id TEXT PRIMARY KEY,
		user_id TEXT NOT NULL,
		name TEXT NOT NULL,
		exchange_type TEXT NOT NULL,
		exchange_config TEXT,
		ai_config TEXT,
		strategy_prompt TEXT,
		status TEXT DEFAULT 'stopped',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS decision_records (
		id TEXT PRIMARY KEY,
		trader_id TEXT NOT NULL,
		cycle_number INTEGER NOT NULL,
		timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
		system_prompt TEXT,
		input_prompt TEXT,
		ai_response TEXT,
		decision_json TEXT,
		account_state TEXT,
		position_snapshots TEXT,
		execution_logs TEXT,
		error_message TEXT,
		ai_latency_ms INTEGER,
		FOREIGN KEY (trader_id) REFERENCES traders(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS performance_records (
		id TEXT PRIMARY KEY,
		trader_id TEXT NOT NULL,
		timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
		equity REAL NOT NULL,
		balance REAL NOT NULL,
		unrealized_pnl REAL,
		realized_pnl REAL,
		total_positions INTEGER,
		FOREIGN KEY (trader_id) REFERENCES traders(id) ON DELETE CASCADE
	);

	CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
	CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
	CREATE INDEX IF NOT EXISTS idx_users_role ON users(role);
	CREATE INDEX IF NOT EXISTS idx_beta_codes_is_active ON beta_codes(is_active);
	CREATE INDEX IF NOT EXISTS idx_traders_user_id ON traders(user_id);
	CREATE INDEX IF NOT EXISTS idx_traders_status ON traders(status);
	CREATE INDEX IF NOT EXISTS idx_decision_records_trader_id ON decision_records(trader_id);
	CREATE INDEX IF NOT EXISTS idx_decision_records_timestamp ON decision_records(timestamp);
	CREATE INDEX IF NOT EXISTS idx_performance_records_trader_id ON performance_records(trader_id);
	CREATE INDEX IF NOT EXISTS idx_performance_records_timestamp ON performance_records(timestamp);
	`

	_, err := db.Exec(schema)
	return err
}
