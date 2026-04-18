package testutil

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"

	"tech-feed-hub/notification-service/internal/domain"
)

var schemaInit sync.Once

func OpenMySQLForTest(t *testing.T) *sql.DB {
	t.Helper()

	if os.Getenv("NOTIFICATION_SERVICE_RUN_INTEGRATION") != "1" {
		t.Skip("skip mysql integration test; set NOTIFICATION_SERVICE_RUN_INTEGRATION=1 to enable")
	}

	dsn := strings.TrimSpace(os.Getenv("NOTIFICATION_SERVICE_TEST_MYSQL_DSN"))
	if dsn == "" {
		dsn = strings.TrimSpace(os.Getenv("MYSQL_DSN"))
	}
	if dsn == "" {
		dsn = "app_user:app_password@tcp(127.0.0.1:3306)/tech_feed_hub?parseTime=true&multiStatements=true"
	}

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		t.Fatalf("open mysql: %v", err)
	}
	db.SetConnMaxLifetime(time.Minute)
	db.SetMaxOpenConns(5)
	db.SetMaxIdleConns(5)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		t.Skipf("skip mysql integration test; mysql unavailable for dsn %q: %v", dsn, err)
	}

	schemaInit.Do(func() {
		if err := applySchema(ctx, db); err != nil {
			t.Fatalf("apply schema: %v", err)
		}
	})

	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Fatalf("close mysql: %v", err)
		}
	})

	return db
}

func ResetMySQLTables(t *testing.T, db *sql.DB) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	stmts := []string{
		"SET FOREIGN_KEY_CHECKS = 0",
		"TRUNCATE TABLE notifications",
		"TRUNCATE TABLE fetch_jobs",
		"TRUNCATE TABLE articles",
		"TRUNCATE TABLE sources",
		"SET FOREIGN_KEY_CHECKS = 1",
	}
	for _, stmt := range stmts {
		if _, err := db.ExecContext(ctx, stmt); err != nil {
			t.Fatalf("reset tables with %q: %v", stmt, err)
		}
	}
}

func InsertSource(t *testing.T, db *sql.DB, name string) int64 {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := db.ExecContext(ctx, `
		INSERT INTO sources (name, type, fetch_url, fetch_method, interval_minutes, default_category, is_enabled)
		VALUES (?, 'rss', ?, 'rss', 60, 'ops', TRUE)`,
		name,
		"https://example.com/"+strings.ToLower(strings.ReplaceAll(name, " ", "-"))+".xml",
	)
	if err != nil {
		t.Fatalf("insert source: %v", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("source last insert id: %v", err)
	}
	return id
}

func InsertFetchJob(t *testing.T, db *sql.DB, sourceID int64, status string, startedAt time.Time, finishedAt *time.Time) int64 {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := db.ExecContext(ctx, `
		INSERT INTO fetch_jobs (source_id, started_at, finished_at, status, fetched_count, inserted_count, duplicated_count, error_message)
		VALUES (?, ?, ?, ?, 3, 2, 1, ?)`,
		sourceID,
		startedAt,
		finishedAt,
		status,
		nil,
	)
	if err != nil {
		t.Fatalf("insert fetch job: %v", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("fetch job last insert id: %v", err)
	}
	return id
}

func InsertNotification(t *testing.T, db *sql.DB, notification domain.Notification) int64 {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := db.ExecContext(ctx, `
		INSERT INTO notifications (event_id, event_type, level, title, body, source_id, fetch_job_id, is_read, created_at, read_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		notification.EventID,
		notification.EventType,
		notification.Level,
		notification.Title,
		notification.Body,
		notification.SourceID,
		notification.FetchJobID,
		notification.IsRead,
		notification.CreatedAt,
		notification.ReadAt,
	)
	if err != nil {
		t.Fatalf("insert notification: %v", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("notification last insert id: %v", err)
	}
	return id
}

func applySchema(ctx context.Context, db *sql.DB) error {
	schemaPath := filepath.Join(repoRootFromCaller(), "deployments/compose/mysql/init/001_schema.sql")
	content, err := os.ReadFile(schemaPath)
	if err != nil {
		return fmt.Errorf("read schema file: %w", err)
	}

	for _, stmt := range strings.Split(string(content), ";") {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}
		if _, err := db.ExecContext(ctx, stmt); err != nil {
			return fmt.Errorf("exec schema statement: %w", err)
		}
	}
	return nil
}

func repoRootFromCaller() string {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		panic("resolve caller path")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(filename), "..", "..", "..", ".."))
}
