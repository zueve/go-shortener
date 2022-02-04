package services

import "context"

type IStorage interface {
	Get(ctx context.Context, key string) (string, error)
	Add(ctx context.Context, url string, userID string) (string, error)
	GetAllUserURLs(ctx context.Context, userID string) (map[string]string, error)
	Ping(ctx context.Context) error
}

type Service struct {
	storage IStorage
}

func New(storage IStorage) Service {
	return Service{
		storage: storage,
	}
}

func (s *Service) CreateRedirect(ctx context.Context, key string, userID string) (string, error) {
	return s.storage.Add(ctx, key, userID)
}

func (s *Service) GetURLByKey(ctx context.Context, key string, userID string) (string, error) {
	return s.storage.Get(ctx, key)
}

func (s *Service) GetAllUserURLs(ctx context.Context, userID string) (map[string]string, error) {
	return s.storage.GetAllUserURLs(ctx, userID)
}

func (s *Service) Ping(ctx context.Context) error {
	return s.storage.Ping(ctx)
}
