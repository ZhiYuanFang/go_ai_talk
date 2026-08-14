package aimodel

import (
	"context"
	"strings"
	"sync"
	"time"
)

// Profile 单条 lane 的运行时配置（Admin DB > YAML > 代码兜底）。
// FreeProvider/FreeModel：非 premium（非 VIP 且额度用尽）时使用；二者皆空表示 omit，由 Python 自选免费模。
type Profile struct {
	Lane         Lane
	Provider     Provider
	Model        string
	FreeProvider Provider // 额度不足模型供应商；可空
	FreeModel    string   // 额度不足模型名；可空
	MaxInFlight  int
	MaxWaiters   int
	TimeoutSec   int
	UpdatedAt    int64
	UpdatedBy    string
}

// ProfileStore 按进程注入：voice 提供 understanding+clinic，ucg 提供 polish。
type ProfileStore interface {
	Load(ctx context.Context, lane Lane) (Profile, error)
	InvalidateCache()
}

var (
	storeMu      sync.RWMutex
	defaultStore ProfileStore
	profileCache = make(map[Lane]cachedProfile)
	cacheTTL     = 60 * time.Second
)

type cachedProfile struct {
	profile   Profile
	expiresAt time.Time
}

// SetProfileStore 在进程启动时注册本域 lane 配置源（voice-service / ucg-service 各注册一次）。
func SetProfileStore(s ProfileStore) {
	storeMu.Lock()
	defaultStore = s
	profileCache = make(map[Lane]cachedProfile)
	storeMu.Unlock()
}

// InvalidateLaneCache 清空进程内 profile 短 TTL 缓存，并令 ProfileStore 清空本地 cache。
// ProfileStore.InvalidateCache 仅清本地，不得再回调本函数，否则 Admin PUT 会 stack overflow。
func InvalidateLaneCache() {
	storeMu.Lock()
	profileCache = make(map[Lane]cachedProfile)
	storeMu.Unlock()
	if defaultStore != nil {
		defaultStore.InvalidateCache()
	}
}

// LoadProfile 读取 lane 配置（带进程内短 TTL 缓存）。
func LoadProfile(ctx context.Context, lane Lane) (Profile, error) {
	storeMu.RLock()
	entry, ok := profileCache[lane]
	store := defaultStore
	storeMu.RUnlock()
	if ok && time.Now().Before(entry.expiresAt) {
		return entry.profile, nil
	}
	if store == nil {
		return Profile{}, ErrProfileStoreUnset
	}
	p, err := store.Load(ctx, lane)
	if err != nil {
		return Profile{}, err
	}
	normalizeProfile(&p)
	storeMu.Lock()
	profileCache[lane] = cachedProfile{profile: p, expiresAt: time.Now().Add(cacheTTL)}
	storeMu.Unlock()
	return p, nil
}

func normalizeProfile(p *Profile) {
	if p == nil {
		return
	}
	p.Model = NormalizeModel(p.Model)
	if p.MaxInFlight <= 0 {
		p.MaxInFlight = 1
	}
	if p.MaxWaiters < 0 {
		p.MaxWaiters = 0
	}
}

// NormalizeModel 规范化 model 名作为 Redis 闸门键的一部分。
func NormalizeModel(model string) string {
	return strings.ToLower(strings.TrimSpace(model))
}

// LaneProfileDTO Admin API 对外字段。
// FreeProvider/FreeModel 可空；Sim 域可不填；voiceUnderstanding 保存时服务层强制清空 free。
type LaneProfileDTO struct {
	Provider     string `json:"provider"`
	Model        string `json:"model"`
	FreeProvider string `json:"freeProvider"`
	FreeModel    string `json:"freeModel"`
	MaxInFlight  int    `json:"maxInFlight"`
	MaxWaiters   int    `json:"maxWaiters"`
	UpdatedAt    int64  `json:"updatedAt"`
	UpdatedBy    string `json:"updatedBy"`
}

// HasFreeModel 是否配置了有效的额度不足模型（供选模出口判断 omit）。
func (p Profile) HasFreeModel() bool {
	return strings.TrimSpace(string(p.FreeProvider)) != "" && strings.TrimSpace(p.FreeModel) != ""
}

// FreeAsRuntimeProfile 将 free 配置投影为可 Acquire/Invoke 的运行时 Profile（闸门参数沿用正式 lane）。
func (p Profile) FreeAsRuntimeProfile() Profile {
	out := p
	out.Provider = p.FreeProvider
	out.Model = NormalizeModel(p.FreeModel)
	return out
}

// ProviderModels allowlist：provider -> models。
var ProviderModels = map[Provider][]string{
	ProviderZhipu: {
		"glm-4.7-flash",
		"glm-4.1v-thinking-flash",
		"glm-4.6v-flash",
		"cogview-3-flash",
		"cogvideox-flash",
	},
	ProviderDeepSeek: {
		"deepseek-v4-flash",
		"deepseek-v4-pro",
	},
	ProviderDashScope: {
		"qwen3-vl-plus",
		"qwen3-vl-flash",
		"qwen-vl-plus",
		"qwen-vl-max",
	},
}

// IsAllowedModel 校验 provider+model 组合是否在 allowlist 内。
func IsAllowedModel(provider Provider, model string) bool {
	model = NormalizeModel(model)
	for _, m := range ProviderModels[provider] {
		if NormalizeModel(m) == model {
			return true
		}
	}
	return false
}

// DefaultEndpoint 返回 provider 默认 OpenAI 兼容 endpoint。
func DefaultEndpoint(provider Provider) string {
	switch provider {
	case ProviderZhipu:
		return "https://open.bigmodel.cn/api/paas/v4/chat/completions"
	case ProviderDeepSeek:
		return "https://api.deepseek.com/v1/chat/completions"
	case ProviderDashScope:
		return "https://dashscope.aliyuncs.com/compatible-mode/v1/chat/completions"
	default:
		return ""
	}
}
