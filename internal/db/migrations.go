package db

import (
	"database/sql"
	"strings"
)

const migrationCreateTables = `
CREATE TABLE IF NOT EXISTS instances (
  id TEXT PRIMARY KEY,
  type TEXT NOT NULL CHECK(type IN ('radarr','sonarr','plex')),
  name TEXT NOT NULL,
  host TEXT NOT NULL,
  api_key TEXT NOT NULL, -- TODO: encrypt api_key at rest
  path_prefix TEXT NOT NULL,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS media_items (
  id TEXT PRIMARY KEY,
  instance_id TEXT NOT NULL REFERENCES instances(id) ON DELETE CASCADE,
  media_type TEXT NOT NULL,
  title TEXT NOT NULL,
  year INTEGER,
  remote_id TEXT,
  path TEXT NOT NULL,
  images TEXT NOT NULL DEFAULT '[]',
  total_size INTEGER NOT NULL DEFAULT 0,
  total_images INTEGER NOT NULL DEFAULT 0,
  compressed INTEGER NOT NULL DEFAULT 0,
  locked INTEGER DEFAULT NULL,
  scanned_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  UNIQUE(instance_id, path)
);

CREATE TABLE IF NOT EXISTS compression_jobs (
  id TEXT PRIMARY KEY,
  instance_id TEXT NOT NULL REFERENCES instances(id) ON DELETE CASCADE,
  status TEXT NOT NULL DEFAULT 'pending',
  config TEXT NOT NULL DEFAULT '{}',
  total_items INTEGER NOT NULL DEFAULT 0,
  processed_items INTEGER NOT NULL DEFAULT 0,
  total_images INTEGER NOT NULL DEFAULT 0,
  processed_images INTEGER NOT NULL DEFAULT 0,
  saved_bytes INTEGER NOT NULL DEFAULT 0,
  error_count INTEGER NOT NULL DEFAULT 0,
  skip_count INTEGER NOT NULL DEFAULT 0,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  started_at DATETIME,
  completed_at DATETIME
);

CREATE TABLE IF NOT EXISTS logs (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  level TEXT NOT NULL,
  source TEXT NOT NULL,
  instance_id TEXT,
  message TEXT NOT NULL,
  details TEXT,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS compression_results (
  id TEXT PRIMARY KEY,
  job_id TEXT NOT NULL REFERENCES compression_jobs(id) ON DELETE CASCADE,
  media_item_id TEXT NOT NULL REFERENCES media_items(id),
  image_path TEXT NOT NULL,
  role TEXT NOT NULL,
  original_bytes INTEGER NOT NULL,
  new_bytes INTEGER NOT NULL,
  saved_bytes INTEGER NOT NULL,
  status TEXT NOT NULL,
  skip_reason TEXT,
  error TEXT,
  new_width INTEGER,
  new_height INTEGER,
  new_format TEXT,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
`

// RunMigrations executes all database migrations.
func RunMigrations(db *sql.DB) error {
	_, err := db.Exec(migrationCreateTables)
	if err != nil {
		return err
	}

	// Create indexes for hot queries
	indexes := []string{
		`CREATE INDEX IF NOT EXISTS idx_media_items_instance_id ON media_items(instance_id)`,
		`CREATE INDEX IF NOT EXISTS idx_media_items_instance_type ON media_items(instance_id, media_type)`,
		`CREATE INDEX IF NOT EXISTS idx_media_items_compressed ON media_items(instance_id, compressed)`,
		`CREATE INDEX IF NOT EXISTS idx_compression_results_media_item ON compression_results(media_item_id)`,
		`CREATE INDEX IF NOT EXISTS idx_compression_jobs_instance ON compression_jobs(instance_id)`,
		`CREATE INDEX IF NOT EXISTS idx_logs_created_at ON logs(created_at DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_logs_level ON logs(level)`,
	}
	for _, idx := range indexes {
		if _, err := db.Exec(idx); err != nil {
			return err
		}
	}

	// Migrate: if old unique constraint (instance_id, media_type, remote_id) exists,
	// drop it and replace with (instance_id, path).
	if err := migrateConstraint(db); err != nil {
		return err
	}

	// Add settings column to instances table if it doesn't exist
	if err := addColumnIfNotExists(db, "instances", "settings", "TEXT NOT NULL DEFAULT '{}'"); err != nil {
		return err
	}

	// Add new_width, new_height, new_format columns to compression_results table
	columns := map[string]string{
		"new_width":  "INTEGER",
		"new_height": "INTEGER",
		"new_format": "TEXT",
	}
	for colName, colType := range columns {
		if err := addColumnIfNotExists(db, "compression_results", colName, colType); err != nil {
			return err
		}
	}

	// Add original_size column to media_items table (stores pre-compression size)
	if err := addColumnIfNotExists(db, "media_items", "original_size", "INTEGER NOT NULL DEFAULT 0"); err != nil {
		return err
	}

	// Add total_images, processed_images columns to compression_jobs table
	jobColumns := map[string]string{
		"total_images":    "INTEGER NOT NULL DEFAULT 0",
		"processed_images": "INTEGER NOT NULL DEFAULT 0",
	}
	for colName, colType := range jobColumns {
		if err := addColumnIfNotExists(db, "compression_jobs", colName, colType); err != nil {
			return err
		}
	}

	// Add ON DELETE CASCADE to compression_results.media_item_id FK
	if err := migrateCompressionResultsCascade(db); err != nil {
		return err
	}

	return nil
}

func migrateConstraint(db *sql.DB) error {
	// Check current unique constraints on media_items
	rows, err := db.Query(`PRAGMA index_list('media_items')`)
	if err != nil {
		return nil // table might not exist yet
	}
	defer rows.Close()

	hasOldConstraint := false
	hasNewConstraint := false
	for rows.Next() {
		var seq int
		var name string
		var unique bool
		var origin string
		var partial int
		if err := rows.Scan(&seq, &name, &unique, &origin, &partial); err != nil {
			continue
		}
		if name == "media_items_instance_id_media_type_remote_id" {
			hasOldConstraint = true
		}
		if name == "media_items_instance_id_path" {
			hasNewConstraint = true
		}
	}

	if !hasOldConstraint || hasNewConstraint {
		return nil // nothing to migrate
	}

	// Recreate table with new constraint (SQLite can't ALTER constraints)
	_, err = db.Exec(`
		CREATE TABLE media_items_new (
			id TEXT PRIMARY KEY,
			instance_id TEXT NOT NULL REFERENCES instances(id) ON DELETE CASCADE,
			media_type TEXT NOT NULL,
			title TEXT NOT NULL,
			year INTEGER,
			remote_id TEXT,
			path TEXT NOT NULL,
			images TEXT NOT NULL DEFAULT '[]',
			total_size INTEGER NOT NULL DEFAULT 0,
			total_images INTEGER NOT NULL DEFAULT 0,
			compressed INTEGER NOT NULL DEFAULT 0,
			locked INTEGER DEFAULT NULL,
			scanned_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(instance_id, path)
		);
		INSERT OR IGNORE INTO media_items_new
			SELECT id, instance_id, media_type, title, year, remote_id, path, images,
			       total_size, total_images, compressed, locked, scanned_at
			FROM media_items;
		DROP TABLE media_items;
		ALTER TABLE media_items_new RENAME TO media_items;
	`)
	return err
}

// migrateCompressionResultsCascade adds ON DELETE CASCADE to the
// compression_results.media_item_id foreign key by recreating the table.
func migrateCompressionResultsCascade(db *sql.DB) error {
	// Check if the FK already has CASCADE by inspecting the CREATE TABLE SQL
	row := db.QueryRow(`SELECT sql FROM sqlite_master WHERE type='table' AND name='compression_results'`)
	var createSQL string
	if err := row.Scan(&createSQL); err != nil {
		return nil // table might not exist yet
	}
	if strings.Contains(createSQL, "ON DELETE CASCADE") {
		return nil // already migrated
	}

	// Recreate table with ON DELETE CASCADE on the media_item_id FK
	_, err := db.Exec(`
		CREATE TABLE compression_results_new (
			id TEXT PRIMARY KEY,
			job_id TEXT NOT NULL REFERENCES compression_jobs(id) ON DELETE CASCADE,
			media_item_id TEXT NOT NULL REFERENCES media_items(id) ON DELETE CASCADE,
			image_path TEXT NOT NULL,
			role TEXT NOT NULL,
			original_bytes INTEGER NOT NULL,
			new_bytes INTEGER NOT NULL,
			saved_bytes INTEGER NOT NULL,
			status TEXT NOT NULL,
			skip_reason TEXT,
			error TEXT,
			new_width INTEGER,
			new_height INTEGER,
			new_format TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
		INSERT INTO compression_results_new SELECT * FROM compression_results;
		DROP TABLE compression_results;
		ALTER TABLE compression_results_new RENAME TO compression_results;
	`)
	return err
}

// addColumnIfNotExists adds a column to a table if it doesn't already exist.
func addColumnIfNotExists(db *sql.DB, table, column, colType string) error {
	rows, err := db.Query(`PRAGMA table_info(` + table + `)`)
	if err != nil {
		return nil // table might not exist yet
	}
	defer rows.Close()

	for rows.Next() {
		var cid int
		var name, ctype string
		var notnull int
		var dflt sql.NullString
		var pk int
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dflt, &pk); err != nil {
			continue
		}
		if name == column {
			return nil // column already exists
		}
	}

	_, err = db.Exec(`ALTER TABLE ` + table + ` ADD COLUMN ` + column + ` ` + colType)
	return err
}