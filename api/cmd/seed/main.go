package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"

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

	// Seed roles
	fmt.Println("Seeding roles...")
	if err := seedRoles(ctx, conn); err != nil {
		log.Fatalf("Failed to seed roles: %v", err)
	}
	fmt.Println("Roles seeded successfully")

	// Get admin role ID
	var adminRoleID string
	err = conn.QueryRow(ctx, `SELECT id FROM roles WHERE slug = 'admin'`).Scan(&adminRoleID)
	if err != nil {
		log.Fatalf("Failed to get admin role: %v", err)
	}

	// Upsert admin user with role_id
	_, err = conn.Exec(ctx, `
		INSERT INTO users (email, password_hash, name, role_id)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (email) DO UPDATE SET
			password_hash = EXCLUDED.password_hash,
			name = EXCLUDED.name,
			role_id = EXCLUDED.role_id
	`, email, string(hash), name, adminRoleID)

	if err != nil {
		log.Fatalf("Failed to create admin user: %v", err)
	}

	fmt.Printf("Admin user created/updated: %s\n", email)

	// Also create corresponding author for account page
	slug := generateSlug(name)
	_, err = conn.Exec(ctx, `
		INSERT INTO authors (name, slug, email, role_id)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (email) DO UPDATE SET
			name = EXCLUDED.name,
			role_id = EXCLUDED.role_id
	`, name, slug, email, adminRoleID)

	if err != nil {
		log.Fatalf("Failed to create admin author profile: %v", err)
	}

	fmt.Printf("Admin author profile created/updated: %s\n", email)
}

func seedRoles(ctx context.Context, conn *pgx.Conn) error {
	// Seed the three main roles
	roles := []struct {
		name        string
		slug        string
		description string
		isSystem    bool
	}{
		{"Administrator", "admin", "Full access to all features and settings", true},
		{"Author", "author", "Can manage articles, categories, and tags", true},
		{"User", "user", "Basic user with limited access", true},
	}

	for _, role := range roles {
		_, err := conn.Exec(ctx, `
			INSERT INTO roles (name, slug, description, is_system)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (slug) DO NOTHING
		`, role.name, role.slug, role.description, role.isSystem)
		if err != nil {
			return fmt.Errorf("failed to seed role %s: %w", role.slug, err)
		}
		fmt.Printf("  - Role '%s' seeded\n", role.name)
	}

	return nil
}

func seedCategories(ctx context.Context, conn *pgx.Conn) error {
	categories := []struct {
		name        string
		slug        string
		description string
	}{
		{"National Politics", "national-politics", "News and analysis about national government and politics"},
		{"Local Government", "local-government", "Coverage of local government units and regional political developments"},
		{"Elections", "elections", "Election news, candidates, and voting information"},
		{"Policy", "policy", "Analysis of government policies and their impact"},
		{"Opinion", "opinion", "Editorial opinions and commentary on political matters"},
		{"Investigations", "investigations", "In-depth investigative reports on political matters"},
	}

	for _, cat := range categories {
		_, err := conn.Exec(ctx, `
			INSERT INTO categories (name, slug, description)
			VALUES ($1, $2, $3)
			ON CONFLICT (slug) DO NOTHING
		`, cat.name, cat.slug, cat.description)
		if err != nil {
			return fmt.Errorf("failed to seed category %s: %w", cat.slug, err)
		}
		fmt.Printf("  - Category '%s' seeded\n", cat.name)
	}

	return nil
}

func seedTags(ctx context.Context, conn *pgx.Conn) error {
	tags := []struct {
		name string
		slug string
	}{
		{"Breaking News", "breaking-news"},
		{"Analysis", "analysis"},
		{"Interview", "interview"},
		{"Feature", "feature"},
		{"Editorial", "editorial"},
		{"Fact Check", "fact-check"},
		{"Senate", "senate"},
		{"House of Representatives", "house-of-representatives"},
		{"Supreme Court", "supreme-court"},
		{"Executive", "executive"},
	}

	for _, tag := range tags {
		_, err := conn.Exec(ctx, `
			INSERT INTO tags (name, slug)
			VALUES ($1, $2)
			ON CONFLICT (slug) DO NOTHING
		`, tag.name, tag.slug)
		if err != nil {
			return fmt.Errorf("failed to seed tag %s: %w", tag.slug, err)
		}
		fmt.Printf("  - Tag '%s' seeded\n", tag.name)
	}

	return nil
}

func generateSlug(name string) string {
	slug := strings.ToLower(name)
	slug = strings.ReplaceAll(slug, " ", "-")
	reg := regexp.MustCompile("[^a-z0-9-]")
	slug = reg.ReplaceAllString(slug, "")
	return slug
}
