package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	var (
		databaseURL string
		email       string
		password    string
		name        string
	)

	flag.StringVar(&databaseURL, "database", "", "Database URL")
	flag.StringVar(&email, "email", "", "Admin email")
	flag.StringVar(&password, "password", "", "Admin password")
	flag.StringVar(&name, "name", "Admin", "Admin name")
	flag.Parse()

	// Fall back to env vars
	if databaseURL == "" {
		databaseURL = os.Getenv("DATABASE_URL")
	}
	if email == "" {
		email = os.Getenv("ADMIN_EMAIL")
	}
	if password == "" {
		password = os.Getenv("ADMIN_PASSWORD")
	}
	if name == "" || name == "Admin" {
		if n := os.Getenv("ADMIN_NAME"); n != "" {
			name = n
		}
	}

	if databaseURL == "" {
		log.Fatal("DATABASE_URL is required")
	}
	if email == "" {
		log.Fatal("ADMIN_EMAIL is required")
	}
	if password == "" {
		log.Fatal("ADMIN_PASSWORD is required")
	}

	// Hash password
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err != nil {
		log.Fatalf("Failed to hash password: %v", err)
	}

	// Connect to database
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, databaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer conn.Close(ctx)

	// Upsert admin user
	_, err = conn.Exec(ctx, `
		INSERT INTO users (email, password_hash, name, role)
		VALUES ($1, $2, $3, 'admin')
		ON CONFLICT (email) DO UPDATE SET
			password_hash = EXCLUDED.password_hash,
			name = EXCLUDED.name,
			role = EXCLUDED.role
	`, email, string(hash), name)

	if err != nil {
		log.Fatalf("Failed to create admin user: %v", err)
	}

	fmt.Printf("Admin user created/updated: %s\n", email)
}
