package services

type IStorage interface {
	Get(key string) (string, error)
	Add(url string) string
}

type Service struct {
	storage IStorage
}

func (s *Service) CreateRedirect(key string) string {
	return s.storage.Add(key)
}

func (s *Service) GetUrlByKey(key string) (string, error) {
	return s.storage.Get(key)
}

func New(storage IStorage) Service {
	return Service{
		storage: storage,
	}
}