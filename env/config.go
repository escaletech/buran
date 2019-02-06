package env

import (
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Port       string `default:"3000"`
	BackendURL string `required:"true" split_words:"true"`
	RedisURL   string `default:"redis://localhost" split_words:"true"`
}

func GetConfig() Config {
	var c Config
	envconfig.MustProcess("", &c)
	return c
}
