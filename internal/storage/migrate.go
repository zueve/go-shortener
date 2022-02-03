package storage

import (
	"github.com/jmoiron/sqlx"
)

const schema = `
CREATE TABLE IF NOT EXISTS link (
    id SERIAL,
    user_id VARCHAR(32),
    origin_url text
)`

func Migrate(db *sqlx.DB) error {
	_, err := db.Exec(schema)
	return err
}
