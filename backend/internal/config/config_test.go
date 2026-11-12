package config

import (
	"net/url"
	"testing"
)

func TestEnvironmentConfigAndEscapedDSN(t *testing.T) {
	for key, value := range map[string]string{
		"DB_HOST": "::1", "DB_PORT": "5432", "DB_USER": "user@example",
		"DB_PASSWORD": "p@ss:/?#&", "DB_NAME": "clims", "JWT_SECRET": "test-secret", "DB_SSLMODE": "require",
	} {
		t.Setenv(key, value)
	}
	cfg, err := LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	u, err := url.Parse(cfg.DSN())
	if err != nil {
		t.Fatal(err)
	}
	password, _ := u.User.Password()
	if u.User.Username() != "user@example" || password != "p@ss:/?#&" || u.Host != "[::1]:5432" || u.Query().Get("sslmode") != "require" {
		t.Fatal("DSN did not preserve credentials, IPv6 host, or SSL mode")
	}
}

func TestMissingJWTSecret(t *testing.T) {
	t.Setenv("DB_USER", "test")
	t.Setenv("DB_PASSWORD", "test")
	t.Setenv("DB_NAME", "test")
	t.Setenv("JWT_SECRET", "")
	if _, err := LoadConfig(); err == nil {
		t.Fatal("expected missing JWT_SECRET to fail")
	}
}
