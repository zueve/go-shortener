package storage

import "context"

func (c *Storage) AddToDeletingQueue(ctx context.Context, url, userID string) error {
	return c.Delete(ctx, url, userID)
}
