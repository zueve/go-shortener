package config

import (
	"flag"
	"time"

	"github.com/caarlos0/env"
)

type Config struct {
	BaseURL         string        `env:"BASE_URL" envDefault:"http://localhost:8080"`
	ServerAddress   string        `env:"SERVER_ADDRESS" envDefault:":8080"`
	FileStoragePath string        `env:"FILE_STORAGE_PATH" envDefault:"storage.txt"`
	DatabaseDSN     string        `env:"DATABASE_DSN" envDefault:"postgres://user:pass@localhost:5432/db"`
	DeleteSize      int           `env:"DELETE_SIZE" envDefault:"5"`
	DeleteWorkerCnt int           `env:"DELETE_WORKER_CNT" envDefault:"5"`
	DeletePeriod    time.Duration `env:"DELETE_PERIOD" envDefault:"5s"`
}

func NewFromEnvAndCMD() (Config, error) {
	var config Config
	err := env.Parse(&config)
	if err != nil {
		return config, err
	}

	b := flag.String("b", config.BaseURL, "a string")
	s := flag.String("s", config.ServerAddress, "a string")
	f := flag.String("f", config.FileStoragePath, "a string")
	d := flag.String("d", config.DatabaseDSN, "a string")
	flag.Parse()

	config.BaseURL = *b
	config.ServerAddress = *s
	config.FileStoragePath = *f
	config.DatabaseDSN = *d
	return config, nil
}
