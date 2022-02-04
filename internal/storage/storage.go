package storage

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
)

type Row struct {
	ID        string `db:"id"`
	UserID    string `db:"user_id"`
	OriginURL string `db:"origin_url"`
}

type Storage struct {
	db *sqlx.DB
}

func New(db *sqlx.DB) (*Storage, error) {
	return &Storage{db: db}, nil
}

func (c *Storage) Ping(ctx context.Context) error {
	return c.db.PingContext(ctx)
}

func (c *Storage) Add(url string, userID string) string {
	query := "INSERT INTO link(user_id, origin_url) VALUES($1, $2) returning id"

	var id string
	err := c.db.Get(&id, query, userID, url)
	if err != nil {
		fmt.Println(query)
		panic(err)
	}
	return fmt.Sprint(id)
}

func (c *Storage) Get(key string) (string, error) {
	var row Row
	if err := c.db.Get(&row, "SELECT * FROM link where id=$1", key); err != nil {
		return "", err
	}
	return row.OriginURL, nil
}

func (c *Storage) GetAllUserURLs(userID string) map[string]string {
	rows := make([]Row, 0)
	err := c.db.Select(&rows, "SELECT id, origin_url, user_id FROM link WHERE user_id=$1 order by id", userID)
	if err != nil {
		panic(err)
	}

	data := make(map[string]string)
	for i := range rows {
		row := rows[i]
		data[fmt.Sprint(row.ID)] = row.OriginURL
	}

	return data
}
