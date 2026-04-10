package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestGetFreshclamConfigValue(t *testing.T) {
	configDir := t.TempDir()
	configPath := filepath.Join(configDir, "freshclam.conf")
	config := []byte(`# comment
DatabaseDirectory /tmp/clamav-db
UpdateLogFile /tmp/freshclam.log
`)

	if err := os.WriteFile(configPath, config, 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	if got := getFreshclamConfigValue(configPath, "DatabaseDirectory", "/fallback/db"); got != "/tmp/clamav-db" {
		t.Fatalf("getFreshclamConfigValue(DatabaseDirectory) = %q, want %q", got, "/tmp/clamav-db")
	}

	if got := getFreshclamConfigValue(configPath, "MissingSetting", "/fallback/value"); got != "/fallback/value" {
		t.Fatalf("getFreshclamConfigValue(MissingSetting) = %q, want %q", got, "/fallback/value")
	}
}

func TestGetLatestDatabaseUpdateTimestamp(t *testing.T) {
	databaseDir := t.TempDir()

	files := map[string]time.Time{
		"daily.cld":    time.Date(2026, time.April, 10, 9, 0, 0, 0, time.UTC),
		"main.cvd":     time.Date(2026, time.April, 10, 11, 30, 0, 0, time.UTC),
		"notes.txt":    time.Date(2026, time.April, 10, 12, 0, 0, 0, time.UTC),
		"bytecode.cvd": time.Date(2026, time.April, 10, 10, 15, 0, 0, time.UTC),
	}

	for name, modTime := range files {
		path := filepath.Join(databaseDir, name)
		if err := os.WriteFile(path, []byte(name), 0o644); err != nil {
			t.Fatalf("WriteFile(%q) error = %v", name, err)
		}
		if err := os.Chtimes(path, modTime, modTime); err != nil {
			t.Fatalf("Chtimes(%q) error = %v", name, err)
		}
	}

	got := getLatestDatabaseUpdateTimestamp(databaseDir)
	want := float64(files["main.cvd"].Unix())

	if got != want {
		t.Fatalf("getLatestDatabaseUpdateTimestamp() = %v, want %v", got, want)
	}
}

func TestGetLatestFreshclamAttemptTimestamp(t *testing.T) {
	logDir := t.TempDir()
	logPath := filepath.Join(logDir, "clamav.log")
	logContent := []byte(`Fri Apr 10 09:00:00 2026 -> ClamAV update process started
Fri Apr 10 09:00:01 2026 -> daily.cld database is up-to-date (version: 123, sigs: 456, f-level: 90, builder: raynman)
Fri Apr 10 11:15:00 2026 -> ClamAV update process started at Fri Apr 10 11:15:00 2026
`)

	if err := os.WriteFile(logPath, logContent, 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	got := getLatestFreshclamAttemptTimestamp(logPath)
	want := float64(time.Date(2026, time.April, 10, 11, 15, 0, 0, time.Local).Unix())

	if got != want {
		t.Fatalf("getLatestFreshclamAttemptTimestamp() = %v, want %v", got, want)
	}
}
