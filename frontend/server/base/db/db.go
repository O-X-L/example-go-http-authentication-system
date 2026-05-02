package db

import (
	"database/sql"
	"example_fe/config"

	_ "github.com/lib/pq"
)

func ConnectDB() (*sql.DB, error) {
	db, err := sql.Open(config.SQL_DRIVER, config.GetPSQLConnection())
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}
