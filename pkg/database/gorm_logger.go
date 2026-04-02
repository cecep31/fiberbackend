package database

import (
	"context"
	"log/slog"
	"time"

	gormlogger "gorm.io/gorm/logger"
)

// GormLogger wraps slog.Logger to implement GORM's logger.Interface
type GormLogger struct {
	log                  *slog.Logger
	level                gormlogger.LogLevel
	slowThreshold        time.Duration
	ignoreRecordNotFound bool
}

// NewGormLogger creates a new GORM logger that uses slog
func NewGormLogger(log *slog.Logger, level gormlogger.LogLevel, slowThreshold time.Duration) *GormLogger {
	return &GormLogger{
		log:                  log,
		level:                level,
		slowThreshold:        slowThreshold,
		ignoreRecordNotFound: true,
	}
}

// LogMode sets the log level for GORM
func (l *GormLogger) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	newLogger := *l
	newLogger.level = level
	return &newLogger
}

// Info logs info messages
func (l *GormLogger) Info(ctx context.Context, msg string, args ...any) {
	if l.level <= gormlogger.Info {
		l.log.InfoContext(ctx, msg, args...)
	}
}

// Warn logs warning messages
func (l *GormLogger) Warn(ctx context.Context, msg string, args ...any) {
	if l.level <= gormlogger.Warn {
		l.log.WarnContext(ctx, msg, args...)
	}
}

// Error logs error messages
func (l *GormLogger) Error(ctx context.Context, msg string, args ...any) {
	if l.level <= gormlogger.Error {
		l.log.ErrorContext(ctx, msg, args...)
	}
}

// Trace logs SQL queries and execution time
func (l *GormLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	if l.level <= gormlogger.Silent {
		return
	}

	elapsed := time.Since(begin)

	// Get the SQL query and rows affected
	sql, rows := fc()

	// Check for slow query
	if l.slowThreshold > 0 && elapsed > l.slowThreshold {
		l.log.WarnContext(ctx, "slow query detected",
			slog.String("sql", sql),
			slog.Duration("duration", elapsed),
			slog.Int64("rows", rows),
		)
		return
	}

	// Log error if exists
	if err != nil {
		if l.ignoreRecordNotFound && err.Error() == "record not found" {
			return
		}
		l.log.ErrorContext(ctx, "query error",
			slog.String("sql", sql),
			slog.Duration("duration", elapsed),
			slog.Any("error", err),
		)
		return
	}

	// Log successful query
	l.log.DebugContext(ctx, "query executed",
		slog.String("sql", sql),
		slog.Duration("duration", elapsed),
		slog.Int64("rows", rows),
	)
}
