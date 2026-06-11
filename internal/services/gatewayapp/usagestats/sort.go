package usagestats

import "strings"

const (
	// SortByCount 按窗口内调用次数降序（默认）。
	SortByCount = "count"
	// SortByLastAt 按最近成功调用时间降序。
	SortByLastAt = "lastAt"
)

// ParseSortBy 解析排序字段，未知值回退 count。
func ParseSortBy(raw string) string {
	if strings.EqualFold(strings.TrimSpace(raw), SortByLastAt) {
		return SortByLastAt
	}
	return SortByCount
}
