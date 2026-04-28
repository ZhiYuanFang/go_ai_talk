package main

import (
	"os"
	"testing"
)

func TestPrepareHistoryServiceRuntimeSetsDefaults(t *testing.T) {
	oldCfg := os.Getenv("GF_GCFG_FILE")
	oldDB := os.Getenv("GF_DATABASE_DEFAULT_LINK")
	oldHistoryDB := os.Getenv("HISTORY_DB_LINK")
	defer func() {
		_ = os.Setenv("GF_GCFG_FILE", oldCfg)
		_ = os.Setenv("GF_DATABASE_DEFAULT_LINK", oldDB)
		_ = os.Setenv("HISTORY_DB_LINK", oldHistoryDB)
	}()

	_ = os.Unsetenv("GF_GCFG_FILE")
	_ = os.Unsetenv("GF_DATABASE_DEFAULT_LINK")
	_ = os.Setenv("HISTORY_DB_LINK", "mysql:user:password@tcp(127.0.0.1:3306)/history_db")

	prepareHistoryServiceRuntime()

	if got := os.Getenv("GF_GCFG_FILE"); got != "manifest/config/config.history-service.yaml" {
		t.Fatalf("unexpected GF_GCFG_FILE: %s", got)
	}
	if got := os.Getenv("GF_DATABASE_DEFAULT_LINK"); got != "mysql:user:password@tcp(127.0.0.1:3306)/history_db" {
		t.Fatalf("unexpected GF_DATABASE_DEFAULT_LINK: %s", got)
	}
}

