package db

import (
	"database/sql"
	"fmt"
	"os"
)

func Migrate(conn *sql.DB, schemaPath string) error {
	b, err := os.ReadFile(schemaPath)
	if err != nil {
		return fmt.Errorf("read schema: %w", err)
	}
	if _, err := conn.Exec(string(b)); err != nil {
		return fmt.Errorf("apply schema: %w", err)
	}
	return nil
}
