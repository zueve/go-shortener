package services

import "context"


type Service struct {
	storage StorageExpected
}

func New(storage StorageExpected) Service {
	return Service{
		storage: storage,
	}
}

func (s *Service) CreateRedirect(ctx context.Context, key string, userID string) (string, error) {
	return s.storage.Add(ctx, key, userID)
}

func (s *Service) GetURLByKey(ctx context.Context, key string) (string, error) {
	return s.storage.Get(ctx, key)
}

func (s *Service) GetAllUserURLs(ctx context.Context, userID string) (map[string]string, error) {
	return s.storage.GetAllUserURLs(ctx, userID)
}

func (s *Service) Ping(ctx context.Context) error {
	return s.storage.Ping(ctx)
}

func (s *Service) CreateRedirectByBatch(ctx context.Context, urls []string, userID string) ([]string, error) {
	return s.storage.AddByBatch(ctx, urls, userID)
}
