package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

var (
	dbURL         = "postgres://user:password@localhost:5432/mydb?sslmode=disable"
	migrationsDir = "file://db/migrations"
)

func main() {
	action := flag.String("action", "up", "Migration action: up, down, or version")
	steps := flag.Int("steps", 1, "Number of steps to migrate")
	flag.Parse()

	m, err := migrate.New(migrationsDir, dbURL)
	if err != nil {
		log.Fatalf("Failed to create migrate instance: %v", err)
	}

	switch *action {
	case "up":
		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			log.Fatalf("Failed to apply migrations: %v", err)
		}
		fmt.Println("Migrations applied successfully.")
	case "down":
		if err := m.Steps(-*steps); err != nil && err != migrate.ErrNoChange {
			log.Fatalf("Failed to revert migrations: %v", err)
		}
		fmt.Println("Migrations reverted successfully.")
	case "version":
		version, dirty, err := m.Version()
		if err != nil {
			log.Fatalf("Failed to get migration version: %v", err)
		}
		fmt.Printf("Current migration version: %d, dirty: %v\n", version, dirty)
	default:
		log.Fatalf("Unknown action: %s", *action)
	}
}
