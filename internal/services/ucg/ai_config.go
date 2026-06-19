package ucg

import (
	"context"
	"os"
	"strings"
	"sync"
	"time"

	"hello/internal/services/aimodel"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
)

const aiConfigSingletonID = 1

// defaultVisionModel 默认 vision 模型（DashScope Qwen3-VL-Plus；OpenAI 兼容 endpoint）。
const defaultVisionModel = "qwen3-vl-plus"

// defaultVisionEndpoint DashScope 北京区 OpenAI 兼容 chat completions。
const defaultVisionEndpoint = "https://dashscope.aliyuncs.com/compatible-mode/v1/chat/completions"

// AllowedVisionModels Admin PUT 与 UI 下拉硬编码 allowlist（含智谱与 DashScope）。
var AllowedVisionModels = []string{
	"glm-4.7-flash",
	"glm-4.1v-thinking-flash",
	"glm-4.6v-flash",
	"qwen3-vl-plus",
	"qwen3-vl-plus-2025-09-23",
	"qwen3-vl-flash",
	"qwen3-vl-flash-2025-10-15",
	"qwen-vl-plus",
	"qwen-vl-max",
	"deepseek-ai/deepseek-vl2",
	"deepseek-ai/deepseek-vl2-small",
	"deepseek-ai/deepseek-vl2-tiny",
	"deepseek-vl2",
}

// AllowedProviders Admin 可选 provider。
var AllowedProviders = []string{
	string(aimodel.ProviderZhipu),
	string(aimodel.ProviderDashScope),
}

// RuntimeAIConfig polish 运行时配置（含 provider 与闸门参数）。
type RuntimeAIConfig struct {
	Provider            string
	VisionModel         string
	MaxImagesPerRequest int
	MaxInFlight         int
	MaxWaiters          int
	DashScopeAPIKey     string
	VisionEndpoint      string
	TimeoutSeconds      int
	UpdatedAt           int64
	UpdatedBy           string
}

type aiConfigCache struct {
	cfg       RuntimeAIConfig
	expiresAt time.Time
}

var (
	aiConfigMu    sync.RWMutex
	aiConfigEntry aiConfigCache
)

// LoadAIConfig reads DB singleton id=1 with YAML fallback and ~60s TTL cache.
func LoadAIConfig(ctx context.Context) RuntimeAIConfig {
	aiConfigMu.RLock()
	if time.Now().Before(aiConfigEntry.expiresAt) {
		cfg := aiConfigEntry.cfg
		aiConfigMu.RUnlock()
		return cfg
	}
	aiConfigMu.RUnlock()

	cfg := loadAIConfigFresh(ctx)
	aiConfigMu.Lock()
	aiConfigEntry = aiConfigCache{cfg: cfg, expiresAt: time.Now().Add(60 * time.Second)}
	aiConfigMu.Unlock()
	return cfg
}

// InvalidateAIConfigCache clears the in-memory AI config cache.
func InvalidateAIConfigCache() {
	aiConfigMu.Lock()
	aiConfigEntry = aiConfigCache{}
	aiConfigMu.Unlock()
}

func loadAIConfigFresh(ctx context.Context) RuntimeAIConfig {
	yamlVision := strings.TrimSpace(g.Cfg().MustGet(ctx, "ucg.ai.default_vision_model").String())
	yamlMax := g.Cfg().MustGet(ctx, "ucg.ai.default_max_images_per_request").Int()
	if yamlMax <= 0 {
		yamlMax = 9
	}
	if yamlVision == "" {
		yamlVision = defaultVisionModel
	}
	yamlProvider := strings.TrimSpace(g.Cfg().MustGet(ctx, "ucg.ai.default_provider").String())
	if yamlProvider == "" {
		yamlProvider = string(aimodel.ProviderZhipu)
	}

	cfg := RuntimeAIConfig{
		Provider:            yamlProvider,
		VisionModel:         yamlVision,
		MaxImagesPerRequest: yamlMax,
		MaxInFlight:         1,
		MaxWaiters:          15,
		VisionEndpoint:      strings.TrimSpace(g.Cfg().MustGet(ctx, "ucg.ai.vision_endpoint").String()),
		DashScopeAPIKey:     strings.TrimSpace(g.Cfg().MustGet(ctx, "ucg.ai.dashscope_api_key").String()),
		TimeoutSeconds:      g.Cfg().MustGet(ctx, "ucg.ai.timeout_seconds").Int(),
	}
	if key := strings.TrimSpace(os.Getenv("UCG_DASHSCOPE_API_KEY")); key != "" {
		cfg.DashScopeAPIKey = key
	} else if key := strings.TrimSpace(os.Getenv("UCG_DEEPSEEK_API_KEY")); key != "" {
		cfg.DashScopeAPIKey = key
	}
	if cfg.VisionEndpoint == "" {
		cfg.VisionEndpoint = defaultVisionEndpoint
	}
	if cfg.TimeoutSeconds <= 0 {
		cfg.TimeoutSeconds = 60
	}

	_ = EnsureUcgAIConfigDefaultRow(ctx)
	var row ucgAIConfigLaneRow
	err := g.DB().Model("ucg_ai_config").Ctx(ctx).Where("id", aiConfigSingletonID).Scan(&row)
	if err == nil && row.Id == aiConfigSingletonID {
		if m := strings.TrimSpace(row.VisionModel); m != "" {
			cfg.VisionModel = m
		}
		if row.MaxImagesPerRequest > 0 {
			cfg.MaxImagesPerRequest = row.MaxImagesPerRequest
		}
		if p := strings.TrimSpace(row.Provider); p != "" {
			cfg.Provider = p
		}
		if row.MaxInFlight > 0 {
			cfg.MaxInFlight = row.MaxInFlight
		}
		if row.MaxWaiters >= 0 {
			cfg.MaxWaiters = row.MaxWaiters
		}
		cfg.UpdatedAt = row.UpdatedAt
		cfg.UpdatedBy = row.UpdatedBy
	}
	return cfg
}

// AIConfigDTO Admin GET/PUT 对外字段。
type AIConfigDTO struct {
	Provider            string `json:"provider"`
	VisionModel         string `json:"visionModel"`
	MaxImagesPerRequest int    `json:"maxImagesPerRequest"`
	MaxInFlight         int    `json:"maxInFlight"`
	MaxWaiters          int    `json:"maxWaiters"`
	UpdatedAt           int64  `json:"updatedAt"`
	UpdatedBy           string `json:"updatedBy"`
}

// GetAIConfigForAdmin returns current DB-backed AI config for Admin UI.
func GetAIConfigForAdmin(ctx context.Context) AIConfigDTO {
	_ = EnsureUcgAIConfigDefaultRow(ctx)
	var row ucgAIConfigLaneRow
	_ = g.DB().Model("ucg_ai_config").Ctx(ctx).Where("id", aiConfigSingletonID).Scan(&row)
	if row.Id != aiConfigSingletonID {
		fallback := loadAIConfigFresh(ctx)
		return AIConfigDTO{
			Provider:            fallback.Provider,
			VisionModel:         fallback.VisionModel,
			MaxImagesPerRequest: fallback.MaxImagesPerRequest,
			MaxInFlight:         fallback.MaxInFlight,
			MaxWaiters:          fallback.MaxWaiters,
		}
	}
	return AIConfigDTO{
		Provider:            row.Provider,
		VisionModel:         row.VisionModel,
		MaxImagesPerRequest: row.MaxImagesPerRequest,
		MaxInFlight:         row.MaxInFlight,
		MaxWaiters:          row.MaxWaiters,
		UpdatedAt:           row.UpdatedAt,
		UpdatedBy:           row.UpdatedBy,
	}
}

// UpdateAIConfigForAdmin validates and persists Admin PUT.
func UpdateAIConfigForAdmin(ctx context.Context, provider, visionModel string, maxImages, maxInFlight, maxWaiters int, updatedBy string) error {
	provider = strings.TrimSpace(provider)
	if provider == "" {
		provider = string(aimodel.ProviderZhipu)
	}
	if !isAllowedProvider(provider) {
		return gerror.NewCode(gcode.CodeInvalidParameter, "provider 不在 allowlist")
	}
	visionModel = strings.TrimSpace(visionModel)
	if !isAllowedVisionModelForProvider(provider, visionModel) {
		return gerror.NewCode(gcode.CodeInvalidParameter, "visionModel 不在 allowlist")
	}
	if maxImages <= 0 || maxImages > 9 {
		return gerror.NewCode(gcode.CodeInvalidParameter, "maxImagesPerRequest 须为 1-9")
	}
	if maxInFlight < 1 {
		return gerror.NewCode(gcode.CodeInvalidParameter, "maxInFlight 须 >= 1")
	}
	if maxWaiters < 0 {
		return gerror.NewCode(gcode.CodeInvalidParameter, "maxWaiters 须 >= 0")
	}
	now := time.Now().Unix()
	_, err := g.DB().Model("ucg_ai_config").Ctx(ctx).Data(g.Map{
		"id":                     aiConfigSingletonID,
		"provider":               provider,
		"vision_model":           visionModel,
		"max_images_per_request": maxImages,
		"max_in_flight":          maxInFlight,
		"max_waiters":            maxWaiters,
		"updated_at":             now,
		"updated_by":             strings.TrimSpace(updatedBy),
	}).Save()
	if err != nil {
		return err
	}
	InvalidateAIConfigCache()
	NewUcgPolishProfileStore().InvalidateCache()
	return nil
}

func isAllowedVisionModel(model string) bool {
	return isAllowedVisionModelForProvider("", model)
}

func isAllowedProvider(provider string) bool {
	for _, p := range AllowedProviders {
		if p == provider {
			return true
		}
	}
	return false
}

func isAllowedVisionModelForProvider(provider, model string) bool {
	provider = strings.TrimSpace(provider)
	if provider != "" {
		return aimodel.IsAllowedModel(aimodel.Provider(provider), model)
	}
	for _, m := range AllowedVisionModels {
		if m == model {
			return true
		}
	}
	return false
}

// UcgAdminPassword returns admin password from env UCG_ADMIN_PASSWORD or yaml ucg.admin.password.
func UcgAdminPassword(ctx context.Context) string {
	if v := strings.TrimSpace(os.Getenv("UCG_ADMIN_PASSWORD")); v != "" {
		return v
	}
	return strings.TrimSpace(g.Cfg().MustGet(ctx, "ucg.admin.password").String())
}

// VerifyUcgAdminPassword checks X-Admin-Password header value.
func VerifyUcgAdminPassword(ctx context.Context, password string) bool {
	expected := UcgAdminPassword(ctx)
	if expected == "" {
		return false
	}
	return strings.TrimSpace(password) == expected
}
