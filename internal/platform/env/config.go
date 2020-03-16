package env

import (
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Port          string `default:"3000"`
	BackendURL    string `split_words:"true"`
	RedisURL      string `default:"redis://localhost" split_words:"true"`
	CacheProvider string `default:"memory" split_words:"true"`
	TTL           string `default:"432000"` // five days in seconds
}

func GetConfig() Config {
	var c Config
	envconfig.MustProcess("", &c)
	return c
}
