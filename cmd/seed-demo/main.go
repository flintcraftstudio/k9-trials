// seed-demo resets all application data and reseeds the deterministic demo
// world. It opens and migrates the database, then delegates to
// internal/seeddemo, which holds the world definition and the demo logins.
//
//	mage seeddemo            # uses DB_PATH or ./data/app.db
package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/pressly/goose/v3"
	_ "modernc.org/sqlite"

	"github.com/flintcraftstudio/k9-trials/internal/seeddemo"
	"github.com/flintcraftstudio/k9-trials/internal/store"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./data/app.db"
	}
	if err := os.MkdirAll(filepath.Dir(dbPath), 0o755); err != nil {
		return fmt.Errorf("ensure db dir: %w", err)
	}

	conn, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}
	defer conn.Close()
	if _, err := conn.Exec("PRAGMA journal_mode=WAL; PRAGMA foreign_keys=ON;"); err != nil {
		return fmt.Errorf("pragma: %w", err)
	}
	if err := goose.SetDialect("sqlite3"); err != nil {
		return fmt.Errorf("goose dialect: %w", err)
	}
	if err := goose.Up(conn, "migrations"); err != nil {
		return fmt.Errorf("migrations: %w", err)
	}

	summary, err := seeddemo.Run(context.Background(), store.New(conn))
	if err != nil {
		return err
	}
	fmt.Println(summary)
	return nil
}
