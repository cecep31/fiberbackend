package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"fiberbackend/config"
	"fiberbackend/pkg/logger"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// NewDatabase creates a new database connection using the provided configuration
func NewDatabase(config *config.Config) *DatabaseWrapper {
	// Create logger for database operations
	dbLog := logger.New(config.ParseLogLevel(), config.LogFormat, false)

	// Configure GORM logger
	var gormLogLevel gormlogger.LogLevel
	if config.Debug {
		gormLogLevel = gormlogger.Info
	} else {
		gormLogLevel = gormlogger.Error
	}

	// Create custom GORM logger that uses our structured logger
	gormConfig := &gorm.Config{
		Logger:      NewGormLogger(dbLog.Logger, gormLogLevel, config.SlowQueryThreshold),
		PrepareStmt: true,
	}

	// Parse database configuration
	pgxConfig, err := pgx.ParseConfig(config.DatabaseURL)
	if err != nil {
		panic(fmt.Errorf("failed to parse database config: %w", err))
	}
	pgxConfig.ConnectTimeout = 10 * time.Second

	// Create database connection with retry logic
	var db *gorm.DB
	var sqldb *sql.DB

	// Retry connection with exponential backoff
	maxRetries := 3
	baseDelay := 1 * time.Second

	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			delay := baseDelay * time.Duration(1<<uint(attempt-1))
			time.Sleep(delay)
			dbLog.Info("retrying database connection",
				logger.Int("attempt", attempt+1),
				logger.Int("max_retries", maxRetries),
			)
		}

		sqldb = stdlib.OpenDB(*pgxConfig)

		// Configure connection pool with better defaults
		maxOpenConns := config.MaxOpenConns
		if maxOpenConns == 0 {
			maxOpenConns = 25
		}

		maxIdleConns := config.MaxIdleConns
		if maxIdleConns == 0 {
			maxIdleConns = 5
		}

		connMaxLifetime := config.ConnMaxLifetime
		if connMaxLifetime == 0 {
			connMaxLifetime = 1 * time.Hour
		}

		sqldb.SetMaxOpenConns(maxOpenConns)
		sqldb.SetMaxIdleConns(maxIdleConns)
		sqldb.SetConnMaxLifetime(connMaxLifetime)
		sqldb.SetConnMaxIdleTime(30 * time.Minute)

		// Create GORM DB instance
		db, err = gorm.Open(postgres.New(postgres.Config{
			Conn: sqldb,
		}), gormConfig)
		if err != nil {
			dbLog.Warn("failed to connect to database",
				logger.Int("attempt", attempt+1),
				logger.Int("max_retries", maxRetries),
				logger.Err(err),
			)
			if sqldb != nil {
				sqldb.Close()
			}
			continue
		}

		// Verify connection
		sqlDB, err := db.DB()
		if err != nil {
			dbLog.Warn("failed to get underlying sql.DB",
				logger.Int("attempt", attempt+1),
				logger.Int("max_retries", maxRetries),
				logger.Err(err),
			)
			continue
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		if err := sqlDB.PingContext(ctx); err != nil {
			cancel()
			dbLog.Warn("failed to ping database",
				logger.Int("attempt", attempt+1),
				logger.Int("max_retries", maxRetries),
				logger.Err(err),
			)
			sqlDB.Close()
			continue
		}
		cancel()

		// Connection successful
		dbLog.Info("successfully connected to database")
		break
	}

	if err != nil {
		panic(fmt.Errorf("failed to connect to database after %d attempts: %w", maxRetries, err))
	}

	return NewDatabaseWrapper(db)
}
