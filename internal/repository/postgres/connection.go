// Package postgres implements the repository against PostgreSQL.
package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"assigning-reviewers-for-pr/config"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
	"go.uber.org/zap"
)

// Postgres wraps a pgx pool and configuration.
type Postgres struct {
	baseCtx context.Context
	log     *zap.SugaredLogger
	db      *pgxpool.Pool
	cfg     config.PostgresConfig
}

// New creates a Postgres repository instance.
func New(ctx context.Context, log *zap.SugaredLogger, cfg *config.Config) *Postgres {
	return &Postgres{
		baseCtx: ctx,
		log:     log.Named("repo.postgres"),
		cfg:     cfg.Postgres,
	}
}

// OnStart establishes connection pool and applies migrations.
func (p *Postgres) OnStart(_ context.Context) error {
	poolCfg, err := pgxpool.ParseConfig(p.cfg.DSN())
	if err != nil {
		return fmt.Errorf("parse pool config: %w", err)
	}
	poolCfg.MaxConns = p.cfg.MaxConns
	poolCfg.MinConns = p.cfg.MinConns

	connectCtx, cancelConnect := context.WithTimeout(p.baseCtx, p.cfg.QueryTimeout)
	defer cancelConnect()

	pool, err := pgxpool.NewWithConfig(connectCtx, poolCfg)
	if err != nil {
		return fmt.Errorf("open pool: %w", err)
	}
	if err := pool.Ping(connectCtx); err != nil {
		return fmt.Errorf("ping pool: %w", err)
	}

	sqlDB, err := sql.Open("postgres", p.cfg.DSN())
	if err != nil {
		return fmt.Errorf("open sql: %w", err)
	}
	defer func() { _ = sqlDB.Close() }()

	migrateCtx, cancelMigrate := context.WithTimeout(p.baseCtx, p.cfg.MigrateTimeout)
	defer cancelMigrate()

	if err := goose.UpContext(migrateCtx, sqlDB, p.cfg.MigrationsDir); err != nil {
		return fmt.Errorf("migrate: %w", err)
	}
	if _, err := goose.EnsureDBVersion(sqlDB); err != nil {
		return fmt.Errorf("migrate version: %w", err)
	}

	p.db = pool
	p.log.Infow("postgres ready", "host", p.cfg.Host, "port", p.cfg.Port)
	return nil
}

// OnStop closes pool connections.
func (p *Postgres) OnStop(_ context.Context) error {
	if p.db != nil {
		p.db.Close()
	}
	return nil
}
