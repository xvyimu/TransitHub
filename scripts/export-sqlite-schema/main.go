// Command export-sqlite-schema runs the same AutoMigrate path as production
// against a temporary SQLite file and prints CREATE TABLE / INDEX DDL.
// Used to draft migrations/main baselines (Phase1 WP-S). Not a runtime dependency.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/xvyimu/TransitHub/common"
	"github.com/xvyimu/TransitHub/model"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	dir, err := os.MkdirTemp("", "th-schema-*")
	if err != nil {
		fatal(err)
	}
	defer os.RemoveAll(dir)

	dbPath := filepath.Join(dir, "schema.db")
	common.SQLitePath = dbPath
	common.SetMainDatabaseType(common.DatabaseTypeSQLite)
	common.SetLogDatabaseType(common.DatabaseTypeSQLite)
	common.IsMasterNode = true

	gdb, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		fatal(err)
	}
	model.DB = gdb
	if err := model.ExportMigrateForSchema(); err != nil {
		fatal(err)
	}

	sqlDB, err := gdb.DB()
	if err != nil {
		fatal(err)
	}
	defer sqlDB.Close()

	rows, err := sqlDB.Query(`SELECT type, name, sql FROM sqlite_master
		WHERE sql IS NOT NULL AND name NOT LIKE 'sqlite_%'
		ORDER BY CASE type WHEN 'table' THEN 0 WHEN 'index' THEN 1 ELSE 2 END, name`)
	if err != nil {
		fatal(err)
	}
	defer rows.Close()

	var b strings.Builder
	b.WriteString("-- Code-generated draft baseline from AutoMigrate (SQLite).\n")
	b.WriteString("-- Phase1 WP-S: review before treating as multi-dialect SSOT.\n")
	b.WriteString("-- golang-migrate will also create schema_migrations; do not include it here.\n\n")

	for rows.Next() {
		var typ, name, ddl string
		if err := rows.Scan(&typ, &name, &ddl); err != nil {
			fatal(err)
		}
		if name == "schema_migrations" {
			continue
		}
		b.WriteString(ddl)
		if !strings.HasSuffix(strings.TrimSpace(ddl), ";") {
			b.WriteString(";")
		}
		b.WriteString("\n\n")
	}
	if err := rows.Err(); err != nil {
		fatal(err)
	}
	fmt.Print(b.String())
}

func fatal(err error) {
	fmt.Fprintf(os.Stderr, "export-sqlite-schema: %v\n", err)
	os.Exit(1)
}
