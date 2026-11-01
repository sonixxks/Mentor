package config

import (
	"net/url"
	"testing"
)

func TestDSN(t *testing.T) {
	//nolint:gosec // хардкод учетных данных безопасен для моков в тестах
	db := DB{
		Host:     "127.0.0.1",
		Port:     "5432",
		User:     "D@",
		Password: "p@ss-@ : / # % ?",
		Name:     "n@me",
		SSLMode:  "disable",
	}

	dsn := db.DSN()

	parsedURL, err := url.Parse(dsn)
	if err != nil {
		t.Fatalf("failed to parse DSN: %v", err)
	}

	resPassword, _ := parsedURL.User.Password()
	resUser := parsedURL.User.Username()
	resHost := parsedURL.Hostname()
	resPort := parsedURL.Port()
	resQuery := parsedURL.Query().Get("sslmode")
	resName := parsedURL.Path // url.Path будет содержать "/n@me"

	if resPassword != db.Password {
		t.Errorf("expected password %q, got %q", db.Password, resPassword)
	}
	if resUser != db.User {
		t.Errorf("expected user %q, got %q", db.User, resUser)
	}
	if resHost != db.Host {
		t.Errorf("expected host %q, got %q", db.Host, resHost)
	}
	if resPort != db.Port {
		t.Errorf("expected port %q, got %q", db.Port, resPort)
	}
	if resQuery != db.SSLMode {
		t.Errorf("expected sslmode %q, got %q", db.SSLMode, resQuery)
	}
	if resName != "/"+db.Name {
		t.Errorf("expected db name %q, got %q", "/"+db.Name, resName)
	}
}
