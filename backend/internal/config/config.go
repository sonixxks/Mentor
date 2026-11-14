package config

import (
	"fmt"
	"net"
	"net/url"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	App App
	DB  DB
	JWT JWT
}

type JWT struct {
	Secret    string        `env:"JWT_SECRET" env-required:"true"`
	AccessTTL time.Duration `env:"JWT_ACCESS_TTL" env-default:"15m"`
	Issuer    string        `env:"JWT_ISSUER" env-default:"mentor-backend"`
}

type App struct {
	Host string `env:"APP_HOST" env-default:"127.0.0.1"`
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
	MaxConns int32  `env:"DB_MAX_CONNS" env-default:"20"`
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
	validateErr := validate(&cfg)
	if validateErr != nil {
		return nil, validateErr
	}
	return &cfg, nil
}

func validate(cfg *Config) error {

	if len(cfg.JWT.Secret) < 32 {
		return fmt.Errorf(
			"JWT_SECRET must be at least 32 bytes, got - %v",
			len(cfg.JWT.Secret))
	}

	if cfg.JWT.AccessTTL <= 0 {
		return fmt.Errorf(
			"JWT_ACCESS_TTL must be > 0, got - %s",
			cfg.JWT.AccessTTL)
	}

	if cfg.JWT.Issuer == "" {
		return fmt.Errorf("JWT_ISSUER cannot be empty")
	}

	if cfg.DB.MaxConns <= 0 {
		return fmt.Errorf("DB_MAX_CONNS must be > 0")
	}
	return nil
}
