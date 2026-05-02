package db

import (
	"database/sql"
	"fmt"
	"example_api/config"
	"example_api/util"
	"log"

	_ "github.com/lib/pq"
)

// InitDB handles connection, verification, and schema setup
func InitDB() (*sql.DB, error) {
	db, err := sql.Open(config.SQL_DRIVER, config.GetPSQLConnection())
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	if err := ensureTables(db); err != nil {
		return nil, err
	}

	if err := ensureIndexes(db); err != nil {
		return nil, err
	}

	// making sure the user_id=0 is reserved
	db.Exec(`
		INSERT INTO users (id, email, password_hash, auth_type, active)
		VALUES (0, "reserved@localhost", %s, 0, false)
		ON CONFLICT DO NOTHING`,
		util.GenerateToken32(),
	)
	log.Println("Database connection and tables verified.")
	return db, nil
}

// ensureTables creates necessary DB tables on startup
func ensureTables(db *sql.DB) error {
	var err error
	for _, table := range TABLES {
		for table_name, table_schema := range table {
			_, err = db.Exec(fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (%s);", table_name, table_schema))
			if err != nil {
				return fmt.Errorf("DB-init: failed to create table '%s': %v", table_name, err)
			}
		}
	}
	return err
}

// ensureIndexes creates necessary DB indexes on startup
func ensureIndexes(db *sql.DB) error {
	var err error
	for index_desc, index_query := range INDEXES {
		_, err = db.Exec(fmt.Sprintf("CREATE INDEX IF NOT EXISTS %s;", index_query))
		if err != nil {
			return fmt.Errorf("DB-init: failed to create index '%s': %v", index_desc, err)
		}
	}
	return err
}
