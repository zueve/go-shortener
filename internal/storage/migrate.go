package storage

import (
	"errors"

	"github.com/jmoiron/sqlx"
)

const schemaSqlite3 = `
CREATE TABLE IF NOT EXISTS link (
    id INTEGER PRIMARY KEY,
    user_id VARCHAR(32) NOT NULL,
    origin_url text NOT NULL,
    is_deleted boolean NOT NULL
)`
const schemaPostgres = `
CREATE TABLE IF NOT EXISTS link (
    id SERIAL,
    user_id VARCHAR(32) NOT NULL,
    origin_url text NOT NULL UNIQUE,
	is_deleted boolean NOT NULL,
	constraint cnst_link_origin_url unique (origin_url)
)`

func Migrate(db *sqlx.DB) error {
	var schema string

	switch db.DriverName() {
	case "sqlite3":
		schema = schemaSqlite3
	case "pgx":
		schema = schemaPostgres
	default:
		return errors.New("unsupported driver type")
	}
	_, err := db.Exec(schema)
	return err
}
