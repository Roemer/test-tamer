package postgres

import (
	"fmt"
	"log/slog"
)

// DbMigrationConsoleLogger implements [migrate.Logger].
type DbMigrationConsoleLogger struct {
}

func (t DbMigrationConsoleLogger) Printf(format string, v ...interface{}) {
	slog.Info("migration", "message", fmt.Sprintf(format, v...))
}

func (t DbMigrationConsoleLogger) Verbose() bool {
	return true
}
