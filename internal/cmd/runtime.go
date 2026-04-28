package cmd

import (
	"os"
	"strings"
)

func prepareDomainDBRuntime() {
	applyDBLinkEnv("DEVICE_DB_LINK", "GF_DATABASE_DEVICE_LINK")
	applyDBLinkEnv("VOICE_DB_LINK", "GF_DATABASE_VOICE_LINK")
	applyDBLinkEnv("HISTORY_DB_LINK", "GF_DATABASE_HISTORY_LINK")
	applyDBLinkEnv("HISTORY_RO_DB_LINK", "GF_DATABASE_HISTORY_RO_LINK")
}

func applyDBLinkEnv(srcEnv, targetEnv string) {
	if link := strings.TrimSpace(os.Getenv(srcEnv)); link != "" {
		_ = os.Setenv(targetEnv, link)
	}
}
