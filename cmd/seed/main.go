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
		fmt.Fprintf(os.Stderr, "Usage: seed <email> <password> [role]\n  role: admin (default) or judge\n")
		os.Exit(1)
	}

	email := os.Args[1]
	password := os.Args[2]
	role := "admin"
	if len(os.Args) >= 4 {
		role = os.Args[3]
	}
	if role != "admin" && role != "judge" {
		fmt.Fprintf(os.Stderr, "Invalid role %q: must be 'admin' or 'judge'\n", role)
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
	id, err := st.CreateUser(context.Background(), email, password, role)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating user: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Created %s %s (id=%d)\n", role, email, id)
}
