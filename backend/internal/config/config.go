package config

import (
	"fmt"
	"net"
	"net/url"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	App App
	Db  DB
	//JWT
	//...
}

type App struct {
	Port int    `env:"APP_PORT" env-default:"8080"`
	Env  string `env:"APP_ENV" env-default:"development"`
}

type DB struct {
	Host     string `env:"DB_HOST" env-default:"127.0.0.1"`
	Port     string `env:"DB_PORT" env-default:"5432"`
	User     string `env:"DB_USER" env-required:"true"`
	Password string `env:"DB_PASSWORD" env-required:"true"`
	Name     string `env:"DB_NAME" env-required:"true"`
	SSLMode  string `env:"DB_SSL_MODE" env-default:"disable"`
}

func (db DB) DSN() string {

	urlValues := url.Values{}

	urlValues.Add("sslmode", db.SSLMode)

	resUrl := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(db.User, db.Password),
		Host:     net.JoinHostPort(db.Host, db.Port),
		Path:     "/" + db.Name,
		RawQuery: urlValues.Encode(),
	}

	return resUrl.String()
}

func LoadConfig() (*Config, error) {
	var cfg Config

	err := cleanenv.ReadEnv(&cfg)
	if err != nil {
		return nil, fmt.Errorf("cannot read config, %w", err)
	}

	return &cfg, nil
}
