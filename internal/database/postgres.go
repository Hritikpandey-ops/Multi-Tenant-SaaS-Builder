package database

import (
	"context"
	"fmt"
	"time"

	"Desktop/multitenant/saas/internal/config"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Database holds database connections
type Database struct {
	GORM *gorm.DB
	PGX  *pgxpool.Pool
}

// NewDatabase creates a new database connection
func NewDatabase(cfg config.DatabaseConfig, log *zerolog.Logger) (*Database, error) {
	// Create GORM connection
	gormDB, err := newGORMConnection(cfg, log)
	if err != nil {
		return nil, fmt.Errorf("failed to create GORM connection: %w", err)
	}

	// Create pgx connection pool
	pgxPool, err := newPGXPool(cfg, log)
	if err != nil {
		return nil, fmt.Errorf("failed to create pgx pool: %w", err)
	}

	db := &Database{
		GORM: gormDB,
		PGX:  pgxPool,
	}

	// Test connection
	if err := db.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Info().Str("database", cfg.DBName).Msg("Database connected successfully")

	return db, nil
}

// newGORMConnection creates a new GORM database connection
func newGORMConnection(cfg config.DatabaseConfig, log *zerolog.Logger) (*gorm.DB, error) {
	// Configure GORM logger
	gormLogger := newGORMLogger(log)

	// Open connection
	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL()), &gorm.Config{
		Logger: gormLogger,
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	})
	if err != nil {
		return nil, err
	}

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	return db, nil
}

// newGORMLogger creates a GORM logger that uses zerolog
func newGORMLogger(log *zerolog.Logger) logger.Interface {
	return logger.New(
		log,
		logger.Config{
			SlowThreshold:             200 * time.Millisecond,
			LogLevel:                  logger.LogLevel(log.GetLevel()),
			IgnoreRecordNotFoundError: true,
			Colorful:                  false,
		},
	)
}

// newPGXPool creates a new pgx connection pool
func newPGXPool(cfg config.DatabaseConfig, log *zerolog.Logger) (*pgxpool.Pool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	config, err := pgxpool.ParseConfig(cfg.DatabaseURL())
	if err != nil {
		return nil, fmt.Errorf("failed to parse database config: %w", err)
	}

	// Configure pool
	config.MaxConns = 100
	config.MinConns = 10
	config.MaxConnLifetime = time.Hour
	config.MaxConnIdleTime = 10 * time.Minute
	config.HealthCheckPeriod = 1 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	return pool, nil
}

// Ping checks if the database connection is alive
func (d *Database) Ping(ctx context.Context) error {
	// Ping using pgx
	return d.PGX.Ping(ctx)
}

// Close closes the database connections
func (d *Database) Close() error {
	// Close GORM connection
	sqlDB, err := d.GORM.DB()
	if err == nil {
		_ = sqlDB.Close()
	}

	// Close pgx pool
	if d.PGX != nil {
		d.PGX.Close()
	}

	return nil
}

// SetTenantContext sets the tenant context for the current connection
// This enables Row-Level Security policies
func (d *Database) SetTenantContext(ctx context.Context, tenantID string) error {
	// For pgx
	_, err := d.PGX.Exec(ctx, "SELECT set_tenant_context($1)", tenantID)
	return err
}

// WithTenantContext returns a context with tenant isolation set
func (d *Database) WithTenantContext(ctx context.Context, tenantID string) (context.Context, error) {
	// Execute the set_tenant_context function
	if err := d.SetTenantContext(ctx, tenantID); err != nil {
		return nil, fmt.Errorf("failed to set tenant context: %w", err)
	}

	// Store tenant ID in context for reference
	return context.WithValue(ctx, "tenant_id", tenantID), nil
}

// HealthCheck checks the health of the database
func (d *Database) HealthCheck(ctx context.Context) error {
	return d.PGX.Ping(ctx)
}

// Stats returns database connection statistics
func (d *Database) Stats() map[string]interface{} {
	stats := make(map[string]interface{})

	// GORM stats
	if sqlDB, err := d.GORM.DB(); err == nil {
		dbStats := sqlDB.Stats()
		stats["gorm"] = map[string]interface{}{
			"max_open_connections": dbStats.MaxOpenConnections,
			"open_connections":     dbStats.OpenConnections,
			"in_use":               dbStats.InUse,
			"idle":                 dbStats.Idle,
		}
	}

	// pgx stats
	if d.PGX != nil {
		stats["pgx"] = map[string]interface{}{
			"max_conns":     d.PGX.Stat().MaxConns(),
			"total_conns":   d.PGX.Stat().TotalConns(),
			"idle_conns":    d.PGX.Stat().IdleConns(),
			"acquire_count": d.PGX.Stat().AcquireCount(),
		}
	}

	return stats
}
