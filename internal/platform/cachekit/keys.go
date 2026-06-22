package cachekit

import (
	"fmt"
	"strings"
)

const (
	DomainVoice   = "voice"
	DomainDevice  = "device"
	DomainHistory = "history"
	DomainGateway = "gateway"
	DomainUCG     = "ucg"
	DomainAI      = "ai"
	DomainSystem  = "platform"
)

// Key 统一生成 Redis 键格式：domain:module:kind:identifier。
func Key(domain, module, kind, identifier string) (string, error) {
	domain = normalizeSegment(domain)
	module = normalizeSegment(module)
	kind = normalizeSegment(kind)
	identifier = normalizeSegment(identifier)
	if domain == "" || module == "" || kind == "" || identifier == "" {
		return "", ErrInvalidKey
	}
	return fmt.Sprintf("%s:%s:%s:%s", domain, module, kind, identifier), nil
}

func normalizeSegment(v string) string {
	v = strings.ToLower(strings.TrimSpace(v))
	// key segment 去空格，避免同义配置产生多份键空间。
	v = strings.ReplaceAll(v, " ", "")
	return v
}

// VersionKey 统一生成读模型版本键：domain:module:version:identifier。
func VersionKey(domain, module, identifier string) (string, error) {
	return Key(domain, module, "version", identifier)
}

// HistoryListKey 设备历史列表键。
func HistoryListKey(deviceNo string) (string, error) {
	return Key(DomainHistory, "record", "list", deviceNo)
}

// HistoryLatestKey 设备历史最新记录键。
func HistoryLatestKey(deviceNo string) (string, error) {
	return Key(DomainHistory, "record", "latest", deviceNo)
}

// HistoryVersionKey 设备历史版本键，用于异步补丁乱序保护。
func HistoryVersionKey(deviceNo string) (string, error) {
	return VersionKey(DomainHistory, "record", deviceNo)
}

// EventOptionsKey 事件选项集合键。
func EventOptionsKey() (string, error) {
	return Key(DomainDevice, "event", "options", "all")
}

// ActionOptionsKey 动作选项集合键。
func ActionOptionsKey() (string, error) {
	return Key(DomainDevice, "action", "options", "all")
}

// UserProfileKey 设备画像键。
func UserProfileKey(deviceNo string) (string, error) {
	return Key(DomainDevice, "user", "profile", deviceNo)
}

// DeviceEventVersionKey 事件元数据版本键。
func DeviceEventVersionKey() (string, error) {
	return VersionKey(DomainDevice, "event", "all")
}

// DeviceActionVersionKey 动作元数据版本键。
func DeviceActionVersionKey() (string, error) {
	return VersionKey(DomainDevice, "action", "all")
}

// UserProfileVersionKey 设备画像版本键。
func UserProfileVersionKey(deviceNo string) (string, error) {
	return VersionKey(DomainDevice, "user", deviceNo)
}

