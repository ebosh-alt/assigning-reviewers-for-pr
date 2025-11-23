package config

import (
	"errors"
	"fmt"
	"time"
)

// Config holds application configuration.
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Postgres PostgresConfig `mapstructure:"postgres"`
	HTTP     HTTPConfig     `mapstructure:"http"`
	Logging  LoggingConfig  `mapstructure:"logging"`
}

// Validate ensures required fields are present.
func (c Config) Validate() error {
	if c.Server.Port == 0 {
		return errors.New("server.port is required")
	}
	if c.Postgres.User == "" || c.Postgres.Password == "" || c.Postgres.DBName == "" {
		return errors.New("postgres credentials are required")
	}
	if c.Postgres.Host == "" {
		return errors.New("postgres.host is required")
	}
	return nil
}

// ServerAddr returns host:port for HTTP server binding.
func (c Config) ServerAddr() string {
	return fmt.Sprintf("%s:%d", c.Server.Host, c.Server.Port)
}

// ServerConfig contains HTTP server options.
type ServerConfig struct {
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port"`
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout"`
}

// HTTPConfig contains transport settings.
type HTTPConfig struct {
	RequestTimeout time.Duration `mapstructure:"request_timeout"`
}

// LoggingConfig contains logger preferences.
type LoggingConfig struct {
	Level string `mapstructure:"level"`
}

// PostgresConfig describes database connection parameters.
type PostgresConfig struct {
	Host           string        `mapstructure:"host"`
	Port           int           `mapstructure:"port"`
	User           string        `mapstructure:"user"`
	Password       string        `mapstructure:"password"`
	DBName         string        `mapstructure:"db_name"`
	SSLMode        string        `mapstructure:"ssl_mode"`
	MigrationsDir  string        `mapstructure:"migrations_dir"`
	MigrateTimeout time.Duration `mapstructure:"migrate_timeout"`
	QueryTimeout   time.Duration `mapstructure:"query_timeout"`
	MaxConns       int32         `mapstructure:"max_conns"`
	MinConns       int32         `mapstructure:"min_conns"`
}

// DSN returns a Postgres connection string.
func (p PostgresConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		p.Host, p.Port, p.User, p.Password, p.DBName, p.SSLMode,
	)
}
