package services

import "context"

type IStorage interface {
	Get(key string) (string, error)
	Add(url string, userID string) string
	GetAllUserURLs(userID string) map[string]string
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

func (s *Service) CreateRedirect(key string, userID string) string {
	return s.storage.Add(key, userID)
}

func (s *Service) GetURLByKey(key string, userID string) (string, error) {
	return s.storage.Get(key)
}

func (s *Service) GetAllUserURLs(userID string) map[string]string {
	return s.storage.GetAllUserURLs(userID)
}

func (s *Service) Ping(ctx context.Context) error {
	return s.storage.Ping(ctx)
}
