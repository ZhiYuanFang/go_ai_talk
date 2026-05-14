package gatewayapp

import (
	"strconv"
	"strings"
)

// parseThreePartNumericVersion 解析纯数字三段式版本（如 1.0.0）；不接受预发布后缀。
// 用于 App 检查更新：与产品约定「统一 x.y.z」一致，解析失败则交由调用方走字符串兜底。
func parseThreePartNumericVersion(s string) (major, minor, patch int64, ok bool) {
	s = strings.TrimSpace(s)
	parts := strings.Split(s, ".")
	if len(parts) != 3 {
		return 0, 0, 0, false
	}
	var err error
	if major, err = strconv.ParseInt(strings.TrimSpace(parts[0]), 10, 64); err != nil || major < 0 {
		return 0, 0, 0, false
	}
	if minor, err = strconv.ParseInt(strings.TrimSpace(parts[1]), 10, 64); err != nil || minor < 0 {
		return 0, 0, 0, false
	}
	if patch, err = strconv.ParseInt(strings.TrimSpace(parts[2]), 10, 64); err != nil || patch < 0 {
		return 0, 0, 0, false
	}
	return major, minor, patch, true
}

// versionLessThreePart 若 current、latest 均可解析为数字三段式，则按字典序比较 (major,minor,patch)。
func versionLessThreePart(current, latest string) (less bool, ok bool) {
	cm, cmi, cp, ok1 := parseThreePartNumericVersion(current)
	lm, lmi, lp, ok2 := parseThreePartNumericVersion(latest)
	if !ok1 || !ok2 {
		return false, false
	}
	if cm != lm {
		return cm < lm, true
	}
	if cmi != lmi {
		return cmi < lmi, true
	}
	return cp < lp, true
}

// ShouldNeedAppUpdate 是否提示客户端升级：latest 非空且客户端版本严格低于 latest。
// 当两边均为可解析的 x.y.z 时按数值比较，避免库中残留低版本占位行导致「已发 1.0.0 仍提示更新」。
// 任一侧无法按三段式解析时回退为去空格后的字符串不等（兼容历史非规范版本号）。
func ShouldNeedAppUpdate(current, latest string) bool {
	latest = strings.TrimSpace(latest)
	if latest == "" {
		return false
	}
	current = strings.TrimSpace(current)
	if less, ok := versionLessThreePart(current, latest); ok {
		return less
	}
	return current != latest
}
