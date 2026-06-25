package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	"github.com/flintcraftstudio/k9-trials/internal/store"

	_ "modernc.org/sqlite"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintf(os.Stderr, "Usage: seed <email> <password> [role]\n  role: admin (default), judge, or competitor\n")
		os.Exit(1)
	}

	email := os.Args[1]
	password := os.Args[2]
	role := "admin"
	if len(os.Args) >= 4 {
		role = os.Args[3]
	}
	if role != "admin" && role != "judge" && role != "competitor" {
		fmt.Fprintf(os.Stderr, "Invalid role %q: must be 'admin', 'judge', or 'competitor'\n", role)
		os.Exit(1)
	}

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./data/app.db"
	}

	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating directory: %v\n", err)
		os.Exit(1)
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	st := store.New(db)
	ctx := context.Background()
	id, err := st.CreateUser(ctx, email, password)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating user: %v\n", err)
		os.Exit(1)
	}

	// Every account is a competitor at baseline (implicit, no row). Elevate to
	// judge/admin by granting the capability — that is the source of truth now.
	if role == "admin" || role == "judge" {
		if err := st.GrantCapability(ctx, id, role); err != nil {
			fmt.Fprintf(os.Stderr, "Error granting %s capability: %v\n", role, err)
			os.Exit(1)
		}
	}

	fmt.Printf("Created %s %s (id=%d)\n", role, email, id)
}
