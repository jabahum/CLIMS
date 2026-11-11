package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	Port      string `mapstructure:"PORT"`
	DBHost    string `mapstructure:"DB_HOST"`
	DBPort    string `mapstructure:"DB_PORT"`
	DBUser    string `mapstructure:"DB_USER"`
	DBPass    string `mapstructure:"DB_PASSWORD"`
	DBName    string `mapstructure:"DB_NAME"`
	JWTSecret string `mapstructure:"JWT_SECRET"` // Add JWT secret
}

func LoadConfig() (cfg Config, err error) {
	// Set default values
	viper.SetDefault("PORT", "8080")
	viper.SetDefault("DB_PORT", "5432")
	viper.SetDefault("DB_HOST", "localhost")

	// Automatically read environment variables
	viper.AutomaticEnv()

	// Unmarshal values into our Config struct
	err = viper.Unmarshal(&cfg)
	return
}

func (cfg *Config) DSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.DBUser, cfg.DBPass, cfg.DBHost, cfg.DBPort, cfg.DBName,
	)
}
