package migrations

import (
	"database/sql"
	"embed"
	"log"
	"sort"
	"strings"
)

//go:embed *.sql
var migrations embed.FS

// ApplyMigrations applies all SQL migrations in the migrations directory.
func ApplyMigrations(db *sql.DB) error {
	files, err := migrations.ReadDir(".")
	if err != nil {
		return err
	}

	// Sort files to ensure migrations are applied in order
	sort.Slice(files, func(i, j int) bool {
		return files[i].Name() < files[j].Name()
	})

	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".sql") {
			continue
		}

		content, err := migrations.ReadFile(file.Name())
		if err != nil {
			return err
		}

		log.Printf("Applying migration: %s", file.Name())
		_, err = db.Exec(string(content))
		if err != nil {
			return err
		}
	}

	return nil
}
