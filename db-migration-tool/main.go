package main

import (
	"database/sql"
	"log"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq" // Import the PostgreSQL driver
)

func main() {
	db, err := sql.Open("postgres", "postgres://user:password@host:port/database_name?sslmode=disable")
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}
	defer db.Close()

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		log.Fatalf("Failed to create the migration driver: %v", err)
	}

	source, err := file.New("file://./migrations")
	if err != nil {
		log.Fatalf("Failed to create the migration source: %v", err)
	}

	m, err := migrate.NewWithInstance("file", source, "postgres", driver)
	if err != nil {
		log.Fatalf("Failed to create the migration instance: %v", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("Failed to apply the migrations: %v", err)
	}

	log.Println("Migrations applied successfully!")
}
