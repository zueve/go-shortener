package storage

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	"github.com/jmoiron/sqlx"
	"github.com/zueve/go-shortener/internal/services"
)

type Row struct {
	ID        string `db:"id"`
	UserID    string `db:"user_id"`
	OriginURL string `db:"origin_url"`
	IsDeleted bool   `db:"is_deleted"`
}

type Storage struct {
	db      *sqlx.DB
	deleter *Deleter
}

func New(db *sqlx.DB, deleteSize int, deleteWorkerCnt int, deletePeriod time.Duration) (*Storage, error) {
	s := &Storage{db: db, deleter: nil}
	deleter, err := NewDeleter(s, deleteSize, deleteWorkerCnt, deletePeriod)
	if err != nil {
		return nil, err
	}
	s.deleter = deleter

	return s, nil
}

func (c *Storage) Shutdown() error {
	defer c.db.Close()
	return c.deleter.Shutdown()
}

func (c *Storage) Ping(ctx context.Context) error {
	return c.db.PingContext(ctx)
}

func (c *Storage) Add(ctx context.Context, url string, userID string) (string, error) {
	query := "INSERT INTO link(user_id, origin_url, is_deleted) VALUES($1, $2, false) returning id"

	var id string
	var pgErr *pgconn.PgError
	err := c.db.GetContext(ctx, &id, query, userID, url)

	if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
		id, err = c.GetURLKey(ctx, url)
		if err != nil {
			return "", err
		}
		return "", services.NewLinkExistError(id, pgErr)
	} else if err != nil {
		return "", err
	}
	return id, nil
}

func (c *Storage) Get(ctx context.Context, key string) (string, error) {
	var row Row
	if err := c.db.GetContext(ctx, &row, "SELECT * FROM link where id=$1", key); err != nil {
		return "", err
	}
	if row.IsDeleted {
		return "", services.ErrRowDeleted
	}
	return row.OriginURL, nil
}

func (c *Storage) GetAllUserURLs(ctx context.Context, userID string) (map[string]string, error) {
	rows := make([]Row, 0)
	query := `
		SELECT id, origin_url, user_id
		FROM link
		WHERE user_id=$1
		AND is_deleted=false
		order by id`
	err := c.db.SelectContext(ctx, &rows, query, userID)
	if err != nil {
		return nil, err
	}

	data := make(map[string]string)
	for i := range rows {
		row := rows[i]
		data[fmt.Sprint(row.ID)] = row.OriginURL
	}

	return data, nil
}

func (c *Storage) AddByBatch(ctx context.Context, urls []string, userID string) ([]string, error) {
	if len(urls) == 0 {
		return make([]string, 0), nil
	}
	rows := make([]map[string]interface{}, len(urls))
	for i := range urls {
		rows[i] = map[string]interface{}{"user_id": userID, "origin_url": urls[i]}
	}

	// Open transaction on batch insert
	tx, err := c.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	query := "INSERT INTO link(user_id, origin_url, is_deleted) VALUES(:user_id, :origin_url, false) returning id"
	result, err := c.db.NamedQueryContext(ctx, query, rows)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}

	// catch result from db
	ids := make([]string, len(urls))
	i := 0
	for result.Next() {
		var id string
		err = result.Scan(&id)
		if err != nil {
			return nil, err
		}
		ids[i] = id
		i = i + 1
	}
	err = result.Err()
	if err != nil {
		return nil, err
	}

	return ids, nil
}

func (c *Storage) GetURLKey(ctx context.Context, originURL string) (string, error) {
	var row Row
	if err := c.db.GetContext(ctx, &row, "SELECT * FROM link where origin_url=$1", originURL); err != nil {
		return "", err
	}
	return row.ID, nil
}

func (c *Storage) AddToDeletingQueue(ctx context.Context, url, userID string) error {
	return c.deleter.Push(Task{URL: url, UserID: userID})
}

func (c *Storage) DeleteByBatch(ctx context.Context, batch []Task) error {
	tx, err := c.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, "UPDATE link SET is_deleted=true WHERE user_id=$1 AND id=$2")
	if err != nil {
		return err
	}

	for _, task := range batch {
		if _, err = stmt.ExecContext(ctx, task.UserID, task.URL); err != nil {
			return err
		}
	}
	return tx.Commit()
}
