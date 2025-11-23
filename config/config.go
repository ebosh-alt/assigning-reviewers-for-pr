// Package config loads application configuration.
package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

const envFile = "config/.env"

// NewConfig loads configuration from environment using viper with typed defaults and validation.
func NewConfig() (*Config, error) {
	v := viper.New()
	if envMap, err := godotenv.Read(envFile); err == nil {
		for k, v := range envMap {
			if _, exists := os.LookupEnv(k); !exists {
				_ = os.Setenv(k, v)
			}
		}
	}

	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	setDefaults(v)
	bindEnvs(v)

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("logging.level", "debug")

	v.SetDefault("server.host", "0.0.0.0")
	v.SetDefault("server.port", 8080)
	v.SetDefault("server.shutdown_timeout", 5*time.Second)

	v.SetDefault("http.request_timeout", 3*time.Second)

	v.SetDefault("postgres.host", "localhost")
	v.SetDefault("postgres.port", 5432)
	v.SetDefault("postgres.user", "postgres")
	v.SetDefault("postgres.password", "postgres")
	v.SetDefault("postgres.db_name", "assigning_reviewers_for_pr_db")
	v.SetDefault("postgres.ssl_mode", "disable")
	v.SetDefault("postgres.migrations_dir", "db/migrations")
	v.SetDefault("postgres.migrate_timeout", 10*time.Second)
	v.SetDefault("postgres.query_timeout", 2*time.Second)
	v.SetDefault("postgres.max_conns", 10)
	v.SetDefault("postgres.min_conns", 2)
}

func bindEnvs(v *viper.Viper) {
	keys := []string{
		"logging.level",
		"server.host",
		"server.port",
		"server.shutdown_timeout",
		"http.request_timeout",
		"postgres.host",
		"postgres.port",
		"postgres.user",
		"postgres.password",
		"postgres.db_name",
		"postgres.ssl_mode",
		"postgres.migrations_dir",
		"postgres.migrate_timeout",
		"postgres.query_timeout",
		"postgres.max_conns",
		"postgres.min_conns",
	}

	for _, k := range keys {
		_ = v.BindEnv(k)
	}
}
