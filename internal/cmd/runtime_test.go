package cmd

import (
	"os"
	"testing"
)

func TestPrepareDomainDBRuntime(t *testing.T) {
	oldDevice := os.Getenv("DEVICE_DB_LINK")
	oldVoice := os.Getenv("VOICE_DB_LINK")
	oldHistory := os.Getenv("HISTORY_DB_LINK")
	oldHistoryRO := os.Getenv("HISTORY_RO_DB_LINK")
	oldGFDevice := os.Getenv("GF_DATABASE_DEVICE_LINK")
	oldGFVoice := os.Getenv("GF_DATABASE_VOICE_LINK")
	oldGFHistory := os.Getenv("GF_DATABASE_HISTORY_LINK")
	oldGFHistoryRO := os.Getenv("GF_DATABASE_HISTORY_RO_LINK")
	defer func() {
		_ = os.Setenv("DEVICE_DB_LINK", oldDevice)
		_ = os.Setenv("VOICE_DB_LINK", oldVoice)
		_ = os.Setenv("HISTORY_DB_LINK", oldHistory)
		_ = os.Setenv("HISTORY_RO_DB_LINK", oldHistoryRO)
		_ = os.Setenv("GF_DATABASE_DEVICE_LINK", oldGFDevice)
		_ = os.Setenv("GF_DATABASE_VOICE_LINK", oldGFVoice)
		_ = os.Setenv("GF_DATABASE_HISTORY_LINK", oldGFHistory)
		_ = os.Setenv("GF_DATABASE_HISTORY_RO_LINK", oldGFHistoryRO)
	}()

	_ = os.Setenv("DEVICE_DB_LINK", "mysql:user:password@tcp(127.0.0.1:3306)/ai_voice_device")
	_ = os.Setenv("VOICE_DB_LINK", "mysql:user:password@tcp(127.0.0.1:3306)/ai_voice_voice")
	_ = os.Setenv("HISTORY_DB_LINK", "mysql:user:password@tcp(127.0.0.1:3306)/ai_voice_history")
	_ = os.Setenv("HISTORY_RO_DB_LINK", "mysql:user:password@tcp(127.0.0.1:3306)/ai_voice_history_ro")
	_ = os.Unsetenv("GF_DATABASE_DEVICE_LINK")
	_ = os.Unsetenv("GF_DATABASE_VOICE_LINK")
	_ = os.Unsetenv("GF_DATABASE_HISTORY_LINK")
	_ = os.Unsetenv("GF_DATABASE_HISTORY_RO_LINK")

	prepareDomainDBRuntime()

	if got := os.Getenv("GF_DATABASE_DEVICE_LINK"); got != "mysql:user:password@tcp(127.0.0.1:3306)/ai_voice_device" {
		t.Fatalf("unexpected GF_DATABASE_DEVICE_LINK: %s", got)
	}
	if got := os.Getenv("GF_DATABASE_VOICE_LINK"); got != "mysql:user:password@tcp(127.0.0.1:3306)/ai_voice_voice" {
		t.Fatalf("unexpected GF_DATABASE_VOICE_LINK: %s", got)
	}
	if got := os.Getenv("GF_DATABASE_HISTORY_LINK"); got != "mysql:user:password@tcp(127.0.0.1:3306)/ai_voice_history" {
		t.Fatalf("unexpected GF_DATABASE_HISTORY_LINK: %s", got)
	}
	if got := os.Getenv("GF_DATABASE_HISTORY_RO_LINK"); got != "mysql:user:password@tcp(127.0.0.1:3306)/ai_voice_history_ro" {
		t.Fatalf("unexpected GF_DATABASE_HISTORY_RO_LINK: %s", got)
	}
}
