package appstatus

import (
	"strings"
	"sync"
	"time"
)

// BannerState 进程内维护通知快照；重启后恢复为默认 inactive。
type BannerState struct {
	Active        bool   `json:"active"`
	Title         string `json:"title"`
	Message       string `json:"message"`
	ExpectedEndAt *int64 `json:"expectedEndAt,omitempty"`
	Dismissible   bool   `json:"dismissible"`
	UpdatedAt     int64  `json:"updatedAt"`
}

// ContentKey 客户端「不再提示」与运维页内容指纹均使用 title+\n+message（trim 后）。
func ContentKey(title, message string) string {
	return strings.TrimSpace(title) + "\n" + strings.TrimSpace(message)
}

var (
	bannerMu    sync.RWMutex
	bannerState = BannerState{Active: false}
)

// Snapshot 返回当前 banner 只读副本。
func Snapshot() BannerState {
	bannerMu.RLock()
	defer bannerMu.RUnlock()
	return bannerState
}

// Update 覆盖 banner 状态并刷新 updatedAt（unix 秒）。
func Update(s BannerState) BannerState {
	now := time.Now().Unix()
	s.UpdatedAt = now
	bannerMu.Lock()
	bannerState = s
	bannerMu.Unlock()
	return s
}

// Deactivate 立即关闭通知（active=false，保留文案供运维预览）。
func Deactivate() BannerState {
	bannerMu.Lock()
	defer bannerMu.Unlock()
	bannerState.Active = false
	bannerState.UpdatedAt = time.Now().Unix()
	return bannerState
}
