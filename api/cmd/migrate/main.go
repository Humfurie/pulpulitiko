package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	var migrationsPath string
	var databaseURL string

	flag.StringVar(&migrationsPath, "path", "migrations", "path to migrations folder")
	flag.StringVar(&databaseURL, "database", "", "database connection string")
	flag.Parse()

	if databaseURL == "" {
		databaseURL = os.Getenv("DATABASE_URL")
	}

	if databaseURL == "" {
		log.Fatal("DATABASE_URL is required (via -database flag or environment variable)")
	}

	command := flag.Arg(0)
	if command == "" {
		command = "up"
	}

	m, err := migrate.New(
		fmt.Sprintf("file://%s", migrationsPath),
		databaseURL,
	)
	if err != nil {
		log.Fatalf("Failed to create migrate instance: %v", err)
	}
	defer m.Close()

	switch command {
	case "up":
		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			log.Fatalf("Migration up failed: %v", err)
		}
		fmt.Println("Migrations applied successfully")

	case "down":
		if err := m.Down(); err != nil && err != migrate.ErrNoChange {
			log.Fatalf("Migration down failed: %v", err)
		}
		fmt.Println("Migrations rolled back successfully")

	case "drop":
		if err := m.Drop(); err != nil {
			log.Fatalf("Drop failed: %v", err)
		}
		fmt.Println("Database dropped successfully")

	case "version":
		version, dirty, err := m.Version()
		if err != nil {
			log.Fatalf("Failed to get version: %v", err)
		}
		fmt.Printf("Version: %d, Dirty: %v\n", version, dirty)

	case "force":
		version := flag.Arg(1)
		if version == "" {
			log.Fatal("Version number required for force command")
		}
		var v int
		if _, err := fmt.Sscanf(version, "%d", &v); err != nil {
			log.Fatalf("Invalid version number: %v", err)
		}
		if err := m.Force(v); err != nil {
			log.Fatalf("Force failed: %v", err)
		}
		fmt.Printf("Forced to version %d\n", v)

	default:
		fmt.Println("Usage: migrate [flags] <command>")
		fmt.Println("")
		fmt.Println("Commands:")
		fmt.Println("  up       Apply all pending migrations")
		fmt.Println("  down     Rollback all migrations")
		fmt.Println("  drop     Drop everything in database")
		fmt.Println("  version  Show current migration version")
		fmt.Println("  force N  Force set version to N")
		fmt.Println("")
		fmt.Println("Flags:")
		flag.PrintDefaults()
	}
}
