package services

type IStorage interface {
	Get(key string) (string, error)
	Add(url string, userID string) string
	GetAllUserURLs(userID string) map[string]string
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
