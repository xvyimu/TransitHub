// Command dbmigrate applies SQL files under migrations/main using golang-migrate.
// SQLite uses the pure-Go driver (no CGO), suitable for Windows and CI.
// Migration sources use iofs (not file://) so Windows paths work.
//
// Dialect selection: -path points at the parent migrations dir (default
// migrations/main). The per-dialect subdirectory (sqlite/ | mysql/ | postgres/)
// is chosen from the database URL scheme so one baseline set serves all three
// databases without ambiguity. Pass -path directly at a leaf dir to override.
//
// Usage:
//
//	go run ./cmd/dbmigrate -path migrations/main -database sqlite://.tmp/demo.db up
//	go run ./cmd/dbmigrate -path migrations/main -database sqlite://.tmp/demo.db version
//	go run ./cmd/dbmigrate -path migrations/main -database mysql://user:pass@tcp(127.0.0.1:3306)/db up
//	go run ./cmd/dbmigrate -path migrations/main -database postgres://user:pass@127.0.0.1:5432/db?sslmode=disable up
package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/database/sqlite"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

func main() {
	pathFlag := flag.String("path", "migrations/main", "path to migration files")
	dbFlag := flag.String("database", "", "database URL (sqlite://file.db | postgres://... | mysql://...)")
	flag.Parse()

	args := flag.Args()
	if len(args) < 1 {
		fatal(errors.New("usage: dbmigrate -path DIR -database URL up|down [N]|version|force VERSION"))
	}
	cmd := strings.ToLower(args[0])

	dbURL := strings.TrimSpace(*dbFlag)
	if dbURL == "" {
		dbURL = strings.TrimSpace(os.Getenv("MIGRATE_DATABASE_URL"))
	}
	if dbURL == "" {
		fatal(errors.New("-database or MIGRATE_DATABASE_URL is required"))
	}
	dbURL = normalizeDatabaseURL(dbURL)

	absPath, err := filepath.Abs(*pathFlag)
	if err != nil {
		fatal(err)
	}
	absPath, err = resolveDialectPath(absPath, dbURL)
	if err != nil {
		fatal(err)
	}
	info, err := os.Stat(absPath)
	if err != nil {
		fatal(fmt.Errorf("migrations path: %w", err))
	}
	if !info.IsDir() {
		fatal(fmt.Errorf("migrations path is not a directory: %s", absPath))
	}

	sourceDriver, err := iofs.New(os.DirFS(absPath), ".")
	if err != nil {
		fatal(err)
	}

	m, err := migrate.NewWithSourceInstance("iofs", sourceDriver, dbURL)
	if err != nil {
		fatal(err)
	}
	defer m.Close()

	switch cmd {
	case "up":
		if len(args) >= 2 {
			n, err := strconv.Atoi(args[1])
			if err != nil {
				fatal(err)
			}
			err = m.Steps(n)
			if err != nil && !errors.Is(err, migrate.ErrNoChange) {
				fatal(err)
			}
		} else {
			err = m.Up()
			if err != nil && !errors.Is(err, migrate.ErrNoChange) {
				fatal(err)
			}
		}
		printVersion(m)
	case "down":
		steps := 1
		if len(args) >= 2 {
			steps, err = strconv.Atoi(args[1])
			if err != nil {
				fatal(err)
			}
		}
		err = m.Steps(-steps)
		if err != nil && !errors.Is(err, migrate.ErrNoChange) {
			fatal(err)
		}
		printVersion(m)
	case "version":
		printVersion(m)
	case "force":
		if len(args) < 2 {
			fatal(errors.New("force requires VERSION"))
		}
		v, err := strconv.Atoi(args[1])
		if err != nil {
			fatal(err)
		}
		if err := m.Force(v); err != nil {
			fatal(err)
		}
		printVersion(m)
	default:
		fatal(fmt.Errorf("unknown command %q", cmd))
	}
}

func printVersion(m *migrate.Migrate) {
	v, dirty, err := m.Version()
	if errors.Is(err, migrate.ErrNilVersion) {
		fmt.Println("0")
		return
	}
	if err != nil {
		fatal(err)
	}
	if dirty {
		fmt.Printf("%d (dirty)\n", v)
		return
	}
	fmt.Printf("%d\n", v)
}

// resolveDialectPath selects the per-dialect subdirectory under a parent
// migrations directory based on the database URL scheme. If parent already
// contains migration files directly (a leaf dir passed explicitly), or the
// dialect subdirectory does not exist, parent is returned unchanged so the
// caller can point -path at a specific directory.
func resolveDialectPath(parent, dbURL string) (string, error) {
	dialect := dialectFromURL(dbURL)
	if dialect == "" {
		return parent, nil
	}

	// If the parent already holds migration files, treat it as an explicit leaf.
	entries, err := os.ReadDir(parent)
	if err != nil {
		return "", fmt.Errorf("migrations path: %w", err)
	}
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".sql") {
			return parent, nil
		}
	}

	sub := filepath.Join(parent, dialect)
	info, err := os.Stat(sub)
	if err == nil && info.IsDir() {
		return sub, nil
	}
	// No dialect subdir: fall back to parent (backward compatible).
	return parent, nil
}

func dialectFromURL(dbURL string) string {
	scheme := dbURL
	if i := strings.Index(scheme, "://"); i >= 0 {
		scheme = scheme[:i]
	}
	switch strings.ToLower(scheme) {
	case "sqlite", "sqlite3":
		return "sqlite"
	case "mysql":
		return "mysql"
	case "postgres", "postgresql", "pgx", "pgx5":
		return "postgres"
	default:
		return ""
	}
}

func normalizeDatabaseURL(raw string) string {
	if strings.HasPrefix(raw, "sqlite3://") {
		rest := strings.TrimPrefix(raw, "sqlite3://")
		return "sqlite://" + normalizeSQLitePath(rest)
	}
	if strings.HasPrefix(raw, "sqlite://") {
		rest := strings.TrimPrefix(raw, "sqlite://")
		return "sqlite://" + normalizeSQLitePath(rest)
	}
	if !strings.Contains(raw, "://") {
		abs, err := filepath.Abs(raw)
		if err != nil {
			return "sqlite://" + filepath.ToSlash(raw)
		}
		return "sqlite://" + filepath.ToSlash(abs)
	}
	return raw
}

func normalizeSQLitePath(path string) string {
	path = strings.TrimSpace(path)
	// Strip leading slashes before Windows drive letter: ///C:/x → C:/x
	for strings.HasPrefix(path, "/") {
		if len(path) >= 3 && path[2] == ':' {
			path = path[1:]
			break
		}
		if len(path) >= 2 && path[1] != '/' {
			// unix absolute /tmp/x
			break
		}
		path = strings.TrimPrefix(path, "/")
	}
	if path == "" {
		return path
	}
	// Keep relative paths relative (for CI cwd).
	if !filepath.IsAbs(path) && !(len(path) >= 2 && path[1] == ':') {
		return filepath.ToSlash(path)
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return filepath.ToSlash(path)
	}
	return filepath.ToSlash(abs)
}

func fatal(err error) {
	fmt.Fprintf(os.Stderr, "dbmigrate: %v\n", err)
	os.Exit(1)
}
